package main

import "github.com/ilyakaznacheev/cleanenv"

type ConfigVars struct {
	GrpcNodeURL         string `yaml:"GrpcNodeUrl" env:"GRPCNODEURL" env-default:"n06.testnet.vega.xyz:3002"`
	SentryEnabled       bool   `yaml:"SentryEnabled" env:"SENTRY-ENABLED" env-default:"false"`
	SentryDsn           string `yaml:"SentryDsn" env:"SENTRY-DSN" env-default:""`
	PrometheusEnabled   bool   `yaml:"PrometheusEnabled" env:"PROMETHEUS-ENABLED" env-default:"false"`
	PrometheusPort      int    `yaml:"PrometheusPort" env:"PROMETHEUS-PORT" env-default:"2112"`
	VegaEventsBatchSize int64  `yaml:"VegaEventsBatchSize" env:"BATCH-SIZE" env-default:"5000"`
	Debug               bool   `yaml:"Debug" env:"DEBUG" env-default:"false"`
	MarketName          string `yaml:"MarketName" env:"MARKETNAME" env-default:""`
	Strategy            string `yaml:"Strategy" env:"STRATEGY" env-default:""`
	CandlesBacklog      int    `yaml:"CandlesBacklog" env:"CANDLESBACKLOG" env-default:""`
	TransactionQuantity int    `yaml:"TransactionQuantity" env:"TRANSACTIONQUANTITY" env-default:""`
	WalletServerURL     string `yaml:"WalletServerURL" env:"WALLETSERVER_URL" env-default:"localhost:1789"`
	WalletName          string `yaml:"WalletName" env:"WALLETNAME" env-default:""`
	WalletPassphrase    string `yaml:"WalletPassphrase" env:"WALLETPASSPHRASE" env-default:""`
	WalletPubKey        string `yaml:"WalletPubKey" env:"WALLETPUBKEY", env-default:""`
}

// ReadConfig import config struct from yaml file
func ReadConfig(path string) (ConfigVars, error) {
	var cfg ConfigVars
	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
