package main

import (
	"github.com/spf13/cobra"
	"github.com/kyokan/drawbridge/internal"
	"github.com/spf13/viper"
	"fmt"
	"os"
	"go.uber.org/zap"
	"github.com/kyokan/drawbridge/internal/logger"
)

var configFile string

var rootCmd *cobra.Command

var log *zap.SugaredLogger

func init() {
	log = logger.Logger.Named("cli")

	cobra.OnInitialize(initConfig)

	rootCmd = &cobra.Command{
		Use: "drawbridge",
		Short: "runs Lightning payment channels on Ethereum",
		Run: func(cmd *cobra.Command, args []string) {
			internal.Start()
		},
	}

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file")
	rootCmd.PersistentFlags().String("rpc-url", "", "URL to a running Ethereum RPC node")
	rootCmd.PersistentFlags().String("contract-address", "", "address of the payment channel smart contract")
	rootCmd.PersistentFlags().String("chain-id", "", "target chain ID")
	rootCmd.PersistentFlags().String("private-key", "", "your wallet's private key")
	rootCmd.PersistentFlags().String("listen-ip", "127.0.0.1", "IP address to listen for RPC requests on")
	rootCmd.PersistentFlags().String("listen-port", "8080", "port to listen for RPC requests on")
	viper.BindPFlag("rpc-url", rootCmd.PersistentFlags().Lookup("rpc-url"))
	viper.BindPFlag("contract-address", rootCmd.PersistentFlags().Lookup("contract-address"))
	viper.BindPFlag("chain-id", rootCmd.PersistentFlags().Lookup("chain-id"))
	viper.BindPFlag("private-key", rootCmd.PersistentFlags().Lookup("private-key"))
	viper.SetDefault("listen-ip", "127.0.0.1")
	viper.SetDefault("listen-port", "8080")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	if configFile == "" {
		log.Info("no config file argument found")

		return
	}

	log.Infow("reading in config", "configFile", configFile)

	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		log.Panicw("failed to read in config file", "err", err.Error())
	}
}