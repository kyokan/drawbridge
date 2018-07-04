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
)

var log *zap.SugaredLogger

func init() {
	log = logger.Logger.Named("start")
}

func Start() {
	keyHex := stringFlag("private-key")
	chainIdFlag := stringFlag("chain-id")
	chainId, err := strconv.Atoi(chainIdFlag)

	if err != nil {
		log.Panicw("mal-formed chain id argument", "err", err.Error())
	}

	km, err := eth.NewKeyManager(keyHex, big.NewInt(int64(chainId)))

	if err != nil {
		log.Panicw("failed to instantiate key manager", "err", err.Error())
	}

	client, err := eth.NewClient(km, stringFlag("rpc-url"), stringFlag("contract-address"))

	if err != nil {
		log.Panicw("failed to instantiate ETH client", "err", err.Error())
	}

	container := &api.ServiceContainer{
		FundingService: api.NewFundingService(client),
	}

	reactor := p2p.NewReactor()

	go reactor.Run()

	go (func() {
		api.Start(container, stringFlag("rpc-ip"), stringFlag("rpc-port"))
	})()

	go(func() {
		p2p.StartNode(reactor, stringFlag("p2p-ip"), stringFlag("p2p-port"), viper.GetStringSlice("bootstrap-peers"))
	})()

	log.Info("started")

	select {}
}

func stringFlag(name string) (string) {
	return viper.GetString(name)
}