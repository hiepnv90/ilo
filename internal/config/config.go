package config

import (
	"fmt"
	"math/big"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Account struct {
	Address         string   `yaml:"address"`
	Passphrase      string   `yaml:"passphrase"`
	InputAmount     *big.Int `yaml:"amount"`
	Recipient       string   `yaml:"recipient"`
	MaxGasFee       *big.Int `yaml:"max_gas_fee"`
	MinReturnAmount *big.Int `yaml:"min_return_amount"`

	PrivKey string `yaml:"priv_key"` // optional, set this empty to use keystore
}

type Config struct {
	ChainID            int64     `yaml:"chain_id"`
	NodeRPC            string    `yaml:"node_rpc"`
	GasPriceEndpoint   string    `yaml:"gas_price_endpoint"`
	KrystalAPIEndpoint string    `yaml:"krystal_api_endpoint"`
	KeystoreDir        string    `yaml:"keystore_dir"`
	InputToken         string    `yaml:"input_token"`
	OutputToken        string    `yaml:"output_token"`
	SlippageBPS        int       `yaml:"slippage_bps"`
	PlatformWallet     string    `yaml:"platform_wallet"`
	GasTipMultiplier   float64   `yaml:"gas_tip_multiplier"`
	StartTime          time.Time `yaml:"start_time"`
	GasLimit           int64     `yaml:"gas_limit"`
	MinReturnAmount    *big.Int  `yaml:"min_return_amount"`
	Accounts           []Account `yaml:"accounts"`
}

func LoadFromFile(fpath string) (Config, error) {
	var cfg Config

	f, err := os.Open(fpath)
	if err != nil {
		return Config{}, fmt.Errorf("open config file: %w", err)
	}

	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}
