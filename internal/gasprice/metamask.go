package gasprice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type MetamaskGasPricer struct {
	client  *http.Client
	baseURL *url.URL
}

func NewMetamaskGasPricer(baseURL string, httpClient *http.Client) (*MetamaskGasPricer, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &MetamaskGasPricer{
		client:  httpClient,
		baseURL: u,
	}, nil
}

func (p *MetamaskGasPricer) GasPrice(ctx context.Context) (float64, float64, error) {
	u := p.baseURL.JoinPath("suggestedGasFees")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return 0, 0, fmt.Errorf("new request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("%w: %s", errors.New("request failed"), resp.Status)
	}

	var res SuggestedGasFeesResp
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, 0, fmt.Errorf("decode response: %w", err)
	}

	gasPriceGwei, err := strconv.ParseFloat(res.High.SuggestedMaxFeePerGas, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("decode gas price: %w", err)
	}

	tipCapGwei, err := strconv.ParseFloat(res.High.SuggestedMaxPriorityFeePerGas, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("decode max priority gas: %w", err)
	}

	return gasPriceGwei, tipCapGwei, nil
}

type GasFee struct {
	SuggestedMaxPriorityFeePerGas string `json:"suggestedMaxPriorityFeePerGas"`
	SuggestedMaxFeePerGas         string `json:"suggestedMaxFeePerGas"`
	MinWaitTimeEstimate           int64  `json:"minWaitTimeEstimate"`
	MaxWaitTimeEstimate           int64  `json:"maxWaitTimeEstimate"`
}

type SuggestedGasFeesResp struct {
	Low                        GasFee   `json:"low"`
	Medium                     GasFee   `json:"medium"`
	High                       GasFee   `json:"high"`
	EstimatedBaseFee           string   `json:"estimatedBaseFee"`
	NetworkCongestion          float64  `json:"networkCongestion"`
	LatestPriorityFeeRange     []string `json:"latestPriorityFeeRange"`
	HistoricalPriorityFeeRange []string `json:"historicalPriorityFeeRange"`
	HistoricalBaseFeeRange     []string `json:"historicalBaseFeeRange"`
	PriorityFeeTrend           string   `json:"priorityFeeTrend"`
	BaseFeeTrend               string   `json:"baseFeeTrend"`
}
