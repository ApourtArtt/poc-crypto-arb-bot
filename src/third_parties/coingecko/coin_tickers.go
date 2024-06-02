package coingecko

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Market struct {
	Name                string `json:"name"`
	Identifier          string `json:"identifier"`
	HasTradingIncentive bool   `json:"has_trading_incentive"`
}

type Converted struct {
	Btc float64 `json:"btc"`
	Eth float64 `json:"eth"`
	Usd float64 `json:"usd"`
}

type Ticker struct {
	Base                   string      `json:"base"`
	Target                 string      `json:"target"`
	Market                 Market      `json:"market"`
	Last                   float64     `json:"last"`
	Volume                 float64     `json:"volume"`
	ConvertedLast          Converted   `json:"converted_last"`
	ConvertedVolume        Converted   `json:"converted_volume"`
	TrustScore             string      `json:"trust_score"`
	BidAskSpreadPercentage float64     `json:"bid_ask_spread_percentage"`
	Timestamp              time.Time   `json:"timestamp"`
	LastTradedAt           time.Time   `json:"last_traded_at"`
	LastFetchAt            time.Time   `json:"last_fetch_at"`
	IsAnomaly              bool        `json:"is_anomaly"`
	IsStale                bool        `json:"is_stale"`
	TradeURL               string      `json:"trade_url"`
	TokenInfoURL           interface{} `json:"token_info_url"` // Since it's sometimes null
	CoinID                 string      `json:"coin_id"`
	TargetCoinID           string      `json:"target_coin_id"`
}

type CryptoData struct {
	Name    string   `json:"name"`
	Tickers []Ticker `json:"tickers"`
}

func GetCoinTickers(apiKey, coinID string) (CryptoData, error) {
	url := "https://api.coingecko.com/api/v3/coins/" + coinID + "/tickers?x-cg-pro-api-key=" + apiKey

	resp, err := http.Get(url)
	if err != nil {
		return CryptoData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return CryptoData{}, fmt.Errorf("Error: %v:%v (%v)", resp.StatusCode, resp.Status, resp.Body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CryptoData{}, err
	}

	var ticker CryptoData
	if err := json.Unmarshal(body, &ticker); err != nil {
		return CryptoData{}, err
	}

	return ticker, nil
}
