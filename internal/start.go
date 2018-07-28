package internal

import (
	"math/big"
	"strconv"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"github.com/kyokan/drawbridge/internal/api"
	"github.com/kyokan/drawbridge/internal/logger"
	"github.com/kyokan/drawbridge/internal/p2p"
	"github.com/kyokan/drawbridge/internal/db"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/btcsuite/btcd/btcec"
	"crypto/ecdsa"
	"github.com/kyokan/drawbridge/internal/wallet"
	"github.com/kyokan/drawbridge/internal/ethclient"
	dwcrypto "github.com/kyokan/drawbridge/pkg/crypto"
	"github.com/kyokan/drawbridge/internal/protocol"
	"github.com/kyokan/drawbridge/internal/lndclient"
	"golang.org/x/net/context"
)

var log *zap.SugaredLogger

func init() {
	log = logger.Logger.Named("start")
}

func Start() {
	keyHex := stringFlag("private-key")
	identityKeyHex := stringFlag("identity-private-key")
	databaseUrl := stringFlag("database-url")
	chainIdFlag := stringFlag("chain-id")

	identityKey, err := crypto.HexToECDSA(identityKeyHex)
	if err != nil {
		log.Panicw("invalid identity key", "err", err.Error())
	}

	chainId, err := strconv.Atoi(chainIdFlag)
	if err != nil {
		log.Panicw("mal-formed chain id argument", "err", err.Error())
	}

	km, err := wallet.NewKeyManager(keyHex, big.NewInt(int64(chainId)))
	if err != nil {
		log.Panicw("failed to instantiate key manager", "err", err.Error())
	}

	ethClient, err := ethclient.NewClient(km, stringFlag("eth-rpc-url"), stringFlag("contract-address"))
	if err != nil {
		log.Panicw("failed to instantiate ETH client", "err", err.Error())
	}

	lndClientConfig := &lndclient.LNDClientConfig{
		Host:         stringFlag("lnd-host"),
		Port:         stringFlag("lnd-port"),
		CertFile:     stringFlag("lnd-cert-file"),
		MacaroonFile: stringFlag("lnd-macaroon-file"),
		Context:      context.TODO(),
	}
	lndClient, err := lndclient.NewClient(lndClientConfig)
	if err != nil {
		log.Panicw("failed to connect to lnd", "err", err.Error())
	}

	database, err := db.NewDB(databaseUrl)
	if err != nil {
		log.Panicw("failed to open database connection", "err", err.Error())
	}

	err = database.Connect()
	if err != nil {
		log.Panicw("failed to connect to the database", "err", err.Error())
	}

	info, err := lndClient.GetInfo()
	if err != nil {
		log.Panicw("failed to connect to lnd", "err", err.Error())
	}

	peerBook := p2p.NewPeerBook()

	chanHandler := protocol.NewChannelHandler(
		peerBook,
		km,
		ethClient,
		database,
	)

	reactor := p2p.NewReactor([]p2p.MsgHandler{
		&protocol.PingPongHandler{},
		protocol.NewHandshakeHandler(lndClient),
		chanHandler,
	})

	lndIdentity, err := dwcrypto.PublicFromCompressedHex("0x" + info.IdentityPubkey)
	if err != nil {
		log.Panicw("failed to parse identity key from lnd", "err", err.Error())
	}

	node, err := p2p.NewNode(&p2p.NodeConfig{
		Reactor:        reactor,
		PeerBook:       peerBook,
		P2PAddr:        stringFlag("p2p-ip"),
		P2PPort:        stringFlag("p2p-port"),
		BootstrapPeers: viper.GetStringSlice("bootstrap-peers"),
		LNDIdentity:    lndIdentity,
		LNDHost:        lndClientConfig.Host,
	})

	container := &api.ServiceContainer{
		FundingService: api.NewFundingService(ethClient, chanHandler),
	}

	if err != nil {
		log.Panicw("failed to create node", "err", err.Error())
	}

	go reactor.Run()

	chainsaw := ethclient.NewChainsaw(ethClient, database)

	go (func() {
		chainsaw.Start()
	})()

	go (func() {
		api.Start(container, stringFlag("rpc-ip"), stringFlag("rpc-port"))
	})()

	go (func() {
		if err := node.Start(convKey(identityKey)); err != nil {
			log.Panicw("failed to start node", "err", err.Error())
		}
	})()

	log.Info("started")

	select {}
}

func stringFlag(name string) (string) {
	return viper.GetString(name)
}

func convKey(key *ecdsa.PrivateKey) *btcec.PrivateKey {
	return (*btcec.PrivateKey)(key)
}
