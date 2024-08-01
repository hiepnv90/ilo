package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/KyberNetwork/tradinglib/pkg/convert"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"

	"github.com/hiepnv90/ilo/internal/clients/krystal"
	"github.com/hiepnv90/ilo/internal/config"
	"github.com/hiepnv90/ilo/internal/gasprice"
)

const (
	flagNameConfig = "config"

	gasMultiplierBPS = 12000 // 1.2
	gweiDecimals     = 9
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

	krystalClient, err := krystal.NewClient(cfg.KrystalAPIEndpoint, nil)
	if err != nil {
		log.Println("Fail to create krystal client:", err)
		return err
	}

	var wg sync.WaitGroup
	for _, acc := range cfg.Accounts {
		wg.Add(1)
		go func(account config.Account) {
			defer wg.Done()

			err = makeTrade(
				ethClient, cacheGasPricer, keystore, krystalClient,
				big.NewInt(cfg.ChainID), account, cfg.InputToken, cfg.OutputToken,
				cfg.PlatformWallet, cfg.SlippageBPS, cfg.GasTipMultiplier,
			)
			if err != nil {
				log.Printf("Fail to make trade: account=%+v err=%v", account, err)
			} else {
				log.Printf("Successfully make trade: %+v", account)
			}
		}(acc)
	}

	wg.Wait()

	return nil
}

func makeTrade(
	ethClient *ethclient.Client,
	gasPricer gasprice.GasPricer,
	keystore *keystore.KeyStore,
	krystalClient *krystal.Client,
	chainID *big.Int,
	account config.Account,
	inputToken string,
	outputToken string,
	platformWallet string,
	slippageBPS int,
	gasTipMultiplier float64,
) error {
	// create a context with timeout 30s
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ratesResp, err := krystalClient.GetAllRates(
		inputToken, outputToken, account.InputAmount, platformWallet, account.Address)
	if err != nil {
		log.Printf("Fail to get route: error=%v", err)
		return err
	}

	if len(ratesResp.Rates) == 0 {
		return errors.New("no rate return")
	}
	rate := ratesResp.Rates[0]

	accountAddress := common.HexToAddress(account.Address)
	nonce, err := ethClient.PendingNonceAt(ctx, accountAddress)
	if err != nil {
		log.Printf("Fail to get nonce: error=%v", err)
		return err
	}

	log.Printf("wallet %v current nonce: %d", accountAddress, nonce)

	minDestAmount := applySlippage(rate.Amount, slippageBPS)
	buildTxResp, err := krystalClient.BuildTx(
		inputToken, outputToken, account.InputAmount, minDestAmount, platformWallet,
		account.Address, rate.Hint, nil, nonce,
	)
	if err != nil {
		log.Printf("Fail to build tx: nonce=%d error=%v", nonce, err)
		return err
	}

	msg := ethereum.CallMsg{
		From:  buildTxResp.TxObject.From,
		To:    &buildTxResp.TxObject.To,
		Value: buildTxResp.TxObject.Value.ToInt(),
		Data:  []byte(buildTxResp.TxObject.Data),
	}

	estimatedGas, err := ethClient.EstimateGas(ctx, msg)
	if err != nil {
		log.Printf("Fail to estimate gas: error=%v", err)
		return err
	}

	maxGasPriceGwei, gasTipCapGwei, err := gasPricer.GasPrice(ctx)
	if err != nil {
		log.Printf("Fail to get gas price: error=%v", err)
		return err
	}
	maxGasPrice := convert.MustFloatToWei(maxGasPriceGwei, gweiDecimals)
	gasTipCap := convert.MustFloatToWei(gasTipMultiplier*gasTipCapGwei, gweiDecimals)
	gas := estimatedGas * gasMultiplierBPS / 10000

	tx := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: maxGasPrice,
		Gas:       gas,
		To:        msg.To,
		Data:      msg.Data,
		Value:     msg.Value,
	}
	signedTx, err := keystore.SignTxWithPassphrase(
		accounts.Account{Address: accountAddress}, account.Passphrase, types.NewTx(tx), chainID)
	if err != nil {
		fmt.Printf("Fail to sign transaction: tx=%+v error=%v", buildTxResp.TxObject, err)
		return err
	}

	log.Printf("Submit transaction: inputAmount=%v transactionHash=%v tx=%+v", account.InputAmount, signedTx.Hash(), buildTxResp.TxObject)
	err = ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		log.Printf("Fail to submit transaction: sender=%v error=%v", getSender(chainID, signedTx), err)
		return err
	}

	// Wait for transaction to be mined
	receipt, err := ethClient.TransactionReceipt(ctx, signedTx.Hash())
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

func getSender(chainID *big.Int, tx *types.Transaction) common.Address {
	signer := types.LatestSignerForChainID(chainID)
	sender, err := types.Sender(signer, tx)
	if err != nil {
		return common.Address{}
	}

	return sender
}

func applySlippage(amount string, slippageBPS int) *big.Int {
	if amount == "" {
		return nil
	}

	amountBig, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		panic(fmt.Errorf("invalid amount: %s", amount))
	}

	const maxBPS = 10000
	if slippageBPS > maxBPS {
		return big.NewInt(0)
	}

	minAmount := new(big.Int).Mul(amountBig, big.NewInt(int64(maxBPS-slippageBPS)))
	minAmount.Div(minAmount, big.NewInt(maxBPS))

	return minAmount
}
