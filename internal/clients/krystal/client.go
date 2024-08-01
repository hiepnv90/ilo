package krystal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:    u,
		httpClient: httpClient,
	}, nil
}

func (c *Client) GetAllRates(
	srcToken string, dstToken string, srcAmount *big.Int, platformWallet string, userAddress string,
) (RatesResponse, error) {
	u := c.baseURL.JoinPath("swap", "allRates")
	q := url.Values{}
	q.Set("src", srcToken)
	q.Set("dest", dstToken)
	q.Set("srcAmount", srcAmount.String())
	q.Set("platformWallet", platformWallet)
	q.Set("userAddress", userAddress)
	u.RawQuery = q.Encode()

	resp, err := c.httpClient.Get(u.String())
	if err != nil {
		return RatesResponse{}, fmt.Errorf("request rates: %w", err)
	}
	defer resp.Body.Close()

	var ratesResp RatesResponse
	if err = decodeResponse(resp, &ratesResp); err != nil {
		return RatesResponse{}, fmt.Errorf("decode response: %w", err)
	}

	return ratesResp, nil
}

func (c *Client) BuildTx(
	srcToken, dstToken string, srcAmount, minDestAmount *big.Int,
	platformWallet, userAddress, hint string, gasPrice *big.Int, nonce uint64,
	skipBalanceCheck bool,
) (BuildTxResponse, error) {
	u := c.baseURL.JoinPath("swap", "buildTx")
	q := url.Values{}
	q.Set("src", srcToken)
	q.Set("dest", dstToken)
	q.Set("srcAmount", srcAmount.String())
	q.Set("minDestAmount", minDestAmount.String())
	q.Set("platformWallet", platformWallet)
	q.Set("userAddress", userAddress)
	q.Set("hint", hint)
	q.Set("gasPrice", bigIntToString(gasPrice))
	q.Set("nonce", strconv.FormatUint(nonce, 10))
	if skipBalanceCheck {
		q.Set("skipBalanceCheck", "true")
	}
	u.RawQuery = q.Encode()

	resp, err := c.httpClient.Get(u.String())
	if err != nil {
		return BuildTxResponse{}, fmt.Errorf("request buildTx: %w", err)
	}
	defer resp.Body.Close()

	var buildTxResp BuildTxResponse
	if err = decodeResponse(resp, &buildTxResp); err != nil {
		return BuildTxResponse{}, fmt.Errorf("decode response: %w", err)
	}

	return buildTxResp, nil
}

type TokenPrice struct {
	Address  common.Address `json:"address"`
	USDPrice float64        `json:"usdPrice"`
}

type Transaction struct {
	From  common.Address `json:"from"`
	To    common.Address `json:"to"`
	Value hexutil.Big    `json:"value"`
	Data  hexutil.Bytes  `json:"data"`
}

type Rate struct {
	Amount       string      `json:"amount"`
	Rate         string      `json:"rate"`
	PriceImpact  int         `json:"priceImpact"`
	Platform     string      `json:"platform"`
	Hint         string      `json:"hint"`
	EstimatedGas uint64      `json:"estimatedGas"`
	TxObject     Transaction `json:"txObject"`
}

type RatesResponse struct {
	Timestamp int64        `json:"timestamp"`
	Prices    []TokenPrice `json:"prices"`
	Rates     []Rate       `json:"rates"`
}

type BuildTxResponse struct {
	Timestamp      int64       `json:"timestamp"`
	TxObject       Transaction `json:"txObject"`
	UsedDefaultGas bool        `json:"usedDefaultGas"`
}

func decodeResponse(resp *http.Response, o interface{}) error {
	if resp.StatusCode/100 != 2 {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return errors.New(string(data))
	}

	return json.NewDecoder(resp.Body).Decode(o)
}

func bigIntToString(i *big.Int) string {
	if i == nil {
		return ""
	}

	return i.String()
}
