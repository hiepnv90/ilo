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
	ChainID          int64     `yaml:"chain_id"`
	NodeRPC          string    `yaml:"node_rpc"`
	GasPriceEndpoint string    `yaml:"gas_price_endpoint"`
	KeystoreDir      string    `yaml:"keystore_dir"`
	RouterAddress    string    `yaml:"router_address"`
	InputToken       string    `yaml:"input_token"`
	OutputToken      string    `yaml:"output_token"`
	FeeTier          int64     `yaml:"fee_tier"`
	GasTipMultiplier float64   `yaml:"gas_tip_multiplier"`
	StartTime        time.Time `yaml:"start_time"`
	GasLimit         int64     `yaml:"gas_limit"`
	MinReturnAmount  *big.Int  `yaml:"min_return_amount"`
	Weth             string    `yaml:"weth"`
	Accounts         []Account `yaml:"accounts"`
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
