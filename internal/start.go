package internal

import (
	"github.com/kyokan/drawbridge/internal/eth"
	"math/big"
	"strconv"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"github.com/kyokan/drawbridge/internal/api"
	"github.com/kyokan/drawbridge/internal/logger"
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

	go (func() {
		api.Start(container, stringFlag("listen-ip"), stringFlag("listen-port"))
	})()

	log.Info("started")

	select {}
}

func stringFlag(name string) (string) {
	return viper.GetString(name)
}