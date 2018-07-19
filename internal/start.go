package internal

import (
	"github.com/kyokan/drawbridge/internal/eth"
	"math/big"
	"strconv"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"github.com/kyokan/drawbridge/internal/api"
	"github.com/kyokan/drawbridge/internal/logger"
	"github.com/kyokan/drawbridge/internal/p2p"
	"github.com/kyokan/drawbridge/pkg"
	"github.com/kyokan/drawbridge/internal/db"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/roasbeef/btcd/btcec"
	"crypto/ecdsa"
	"github.com/kyokan/drawbridge/internal/wallet"
	"github.com/kyokan/drawbridge/internal/channel"
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

	client, err := eth.NewClient(km, stringFlag("rpc-url"), stringFlag("contract-address"))

	if err != nil {
		log.Panicw("failed to instantiate ETH client", "err", err.Error())
	}

	chainHashes, err := pkg.NewChainHashes()

	if err != nil {
		log.Panicw("failed to generate chain hashes", "err", err.Error())
	}

	database, err := db.NewDB(databaseUrl)

	if err != nil {
		log.Panicw("failed to open database connection", "err", err.Error())
	}

	err = database.Connect()

	if err != nil {
		log.Panicw("failed to connect to the database", "err", err.Error())
	}

	config := &pkg.Config{
		ChainHashes:    chainHashes,
		P2PAddr:        stringFlag("p2p-ip"),
		P2PPort:        stringFlag("p2p-port"),
		BootstrapPeers: viper.GetStringSlice("bootstrap-peers"),
		SigningPubkey:  km.PublicKey(),
	}

	peerBook := p2p.NewPeerBook()

	cMgr := channel.NewFundingManager(peerBook, database, km, client, config)

	reactor := p2p.NewReactor([]p2p.MsgHandler{
		cMgr,
	})

	node, err := p2p.NewNode(reactor, peerBook, config)

	container := &api.ServiceContainer{
		FundingService: api.NewFundingService(client, cMgr),
	}

	if err != nil {
		log.Panicw("failed to create node", "err", err.Error())
	}

	go reactor.Run()

	chainsaw := eth.NewChainsaw(client, database)

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
