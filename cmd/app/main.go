package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/KyberNetwork/tradinglib/pkg/convert"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"github.com/hiepnv90/ilo/internal/blockchain"
	"github.com/hiepnv90/ilo/internal/config"
	"github.com/hiepnv90/ilo/internal/gasprice"
)

const (
	flagNameConfig = "config"

	gasMultiplierBPS    = 12_000 // 1.2
	gweiDecimals        = 9
	maxGasLimit         = 20_000_000
	defaultDeadlineTime = 24 * time.Second

	eth = "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
)

func main() {
	app := cli.NewApp()
	app.Action = runApp
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    flagNameConfig,
			EnvVars: []string{"CONFIG"},
			Value:   "config.yaml",
			Usage:   "Path to configuration file",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln("App exit with error:", err)
	}

	log.Println("App exit successfully")
}

func runApp(c *cli.Context) error {
	configFile := c.String(flagNameConfig)

	log.Println("Load config from file:", configFile)
	cfg, err := config.LoadFromFile(configFile)
	if err != nil {
		log.Println("Fail to load config from file:", err)
		return err
	}

	keystore := keystore.NewKeyStore(cfg.KeystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)

	return makeTrades(cfg, keystore)
}

func makeTrades(cfg config.Config, keystore *keystore.KeyStore) error {
	delay := time.Until(cfg.StartTime)
	if delay > 0 {
		log.Printf("Wait %v before starting to make trades\n", delay)
		time.Sleep(delay)
	}

	ethClient, err := ethclient.Dial(cfg.NodeRPC)
	if err != nil {
		log.Println("Fail to create ethclient:", err)
		return err
	}

	metamaskGasPricer, err := gasprice.NewMetamaskGasPricer(cfg.GasPriceEndpoint, nil)
	if err != nil {
		log.Println("Fail to create metamask gas pricer:", err)
		return err
	}
	cacheGasPricer := gasprice.NewCacheGasPricer(metamaskGasPricer, time.Second)

	var gasLimit uint64
	if cfg.GasLimit > 0 {
		gasLimit = uint64(cfg.GasLimit)
		if gasLimit > maxGasLimit {
			gasLimit = maxGasLimit
		}
	}

	g, _ := errgroup.WithContext(context.Background())
	for _, acc := range cfg.Accounts {
		acc := acc
		g.Go(func() error {
			err = makeTrade(
				ethClient, cacheGasPricer, keystore, big.NewInt(cfg.ChainID), acc,
				strings.ToLower(cfg.InputToken), strings.ToLower(cfg.OutputToken),
				cfg.GasTipMultiplier, gasLimit, cfg.MinReturnAmount, big.NewInt(cfg.FeeTier),
				cfg.RouterAddress, strings.ToLower(cfg.Weth),
			)
			if err != nil {
				log.Printf("Fail to make trade: account=%+v err=%v", acc, err)
				return err
			}

			log.Printf("Successfully make trade: %+v", acc)
			return nil
		})
	}

	return g.Wait()
}

func makeTrade(
	ethClient *ethclient.Client,
	gasPricer gasprice.GasPricer,
	keystore *keystore.KeyStore,
	chainID *big.Int,
	account config.Account,
	inputToken string,
	outputToken string,
	gasTipMultiplier float64,
	gasLimit uint64,
	minReturnAmount *big.Int,
	feeTier *big.Int,
	routerAddress string,
	weth string,
) error {
	// create a context with timeout 30s
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	accountAddress := common.HexToAddress(account.Address)
	tokenIn := toTokenAddress(inputToken, weth)
	tokenOut := toTokenAddress(outputToken, weth)
	if account.MinReturnAmount != nil {
		minReturnAmount = account.MinReturnAmount
	} else if minReturnAmount == nil {
		minReturnAmount = big.NewInt(0)
	}

	var priv *ecdsa.PrivateKey
	var err error
	if account.PrivKey != "" {
		priv, err = crypto.HexToECDSA(account.PrivKey)
		if err != nil {
			log.Println("invalid private key")
			return err
		}
		publicKeyECDSA, ok := priv.Public().(*ecdsa.PublicKey)
		if !ok {
			return errors.New("failed to get public key")
		}
		tmpAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
		account.Address = tmpAddress.Hex()
	}

	recipient := accountAddress
	if account.Recipient != "" {
		recipient = common.HexToAddress(account.Recipient)
	}

	encodedData, err := blockchain.EncodeSwap02(
		tokenIn, tokenOut, recipient, account.InputAmount, minReturnAmount, feeTier)
	if err != nil {
		log.Println("Fail to encode swap:", err)
		return err
	}

	toAddr := common.HexToAddress(routerAddress)
	msg := ethereum.CallMsg{
		From: accountAddress,
		To:   &toAddr,
		Data: encodedData,
	}
	if isEth(inputToken) {
		msg.Value = account.InputAmount
	}

	if gasLimit == 0 {
		gasLimit, err = ethClient.EstimateGas(context.Background(), msg)
		if err != nil {
			log.Printf("Fail to estimate gas: from=%v to=%v data=%s error=%v",
				msg.From, msg.To, hexutil.Encode(msg.Data), err)
			return err
		}

		gasLimit = gasLimit * gasMultiplierBPS / 10_000
	}

	maxGasPriceGwei, gasTipCapGwei, err := gasPricer.GasPrice(ctx)
	if err != nil {
		log.Printf("Fail to get gas price: error=%v", err)
		return err
	}
	maxGasPrice := gasPriceWithCap(gasLimit, maxGasPriceGwei, account.MaxGasFee)
	gasTipCap := convert.MustFloatToWei(gasTipMultiplier*gasTipCapGwei, gweiDecimals)

	nonce, err := ethClient.NonceAt(ctx, accountAddress, nil)
	if err != nil {
		log.Printf("Fail to get nonce: error=%v", err)
		return err
	}

	tx := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: maxGasPrice,
		Gas:       gasLimit,
		To:        msg.To,
		Data:      msg.Data,
		Value:     msg.Value,
	}

	var signedTx *types.Transaction
	if priv != nil {
		// TO sign transaction using privKey
		signer := types.LatestSignerForChainID(chainID)
		signedTx, err = types.SignTx(types.NewTx(tx), signer, priv)
		if err != nil {
			logTx := *tx
			logTx.Data = nil
			log.Printf("Fail to sign transaction use privKey: tx=%+v data=%s error=%v",
				logTx, hexutil.Encode(tx.Data), err)
			return err
		}
	} else {
		signedTx, err = keystore.SignTxWithPassphrase(
			accounts.Account{Address: accountAddress}, account.Passphrase, types.NewTx(tx), chainID)
		if err != nil {
			logTx := *tx
			logTx.Data = nil
			log.Printf("Fail to sign transaction: tx=%+v data=%s error=%v",
				logTx, hexutil.Encode(tx.Data), err)
			return err
		}
	}

	log.Printf("Submit transaction: inputAmount=%v transactionHash=%v", account.InputAmount, signedTx.Hash())
	err = ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		log.Printf("Fail to submit transaction: sender=%v error=%v", getSender(chainID, signedTx), err)
		return err
	}

	// Wait for transaction to be mined
	receipt, err := waitForTransactionReceipt(ctx, ethClient, signedTx.Hash(), defaultDeadlineTime)
	if err != nil {
		log.Printf("Fail to get transaction receipt: transactionHash=%v error=%v", signedTx.Hash(), err)
		return err
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		log.Printf("Transaction failed: transactionHash=%v status=%v", signedTx.Hash(), receipt.Status)
		return errors.New("transaction failed")
	}

	log.Printf("Successfully submit transaction: inputAmount=%v transactionHash=%v", account.InputAmount, signedTx.Hash())

	return nil
}

func waitForTransactionReceipt(ctx context.Context, ethClient *ethclient.Client, txHash common.Hash, timeout time.Duration) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		receipt, err := ethClient.TransactionReceipt(ctx, txHash)
		if err == nil {
			return receipt, nil
		}
		if err != ethereum.NotFound {
			return nil, fmt.Errorf("error fetching receipt: %v", err)
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("transaction not mined within %v", timeout)
		case <-ticker.C:
			continue
		}
	}
}

func getSender(chainID *big.Int, tx *types.Transaction) common.Address {
	signer := types.LatestSignerForChainID(chainID)
	sender, err := types.Sender(signer, tx)
	if err != nil {
		return common.Address{}
	}

	return sender
}

func isEth(token string) bool {
	return token == eth
}

func toTokenAddress(token string, weth string) common.Address {
	if isEth(token) {
		return common.HexToAddress(weth)
	}

	return common.HexToAddress(token)
}

func gasPriceWithCap(gasLimit uint64, maxGasPriceGwei float64, maxGasFee *big.Int) *big.Int {
	if maxGasFee == nil {
		return convert.MustFloatToWei(maxGasPriceGwei, gweiDecimals)
	}

	return new(big.Int).Div(maxGasFee, new(big.Int).SetUint64(gasLimit))
}
