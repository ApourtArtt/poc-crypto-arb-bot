package mexcsdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SymbolGetExchangeInfo struct {
	Symbol                     string   `json:"symbol"`
	Status                     string   `json:"status"`
	BaseAsset                  string   `json:"baseAsset"`
	BaseAssetPrecision         int      `json:"baseAssetPrecision"`
	QuoteAsset                 string   `json:"quoteAsset"`
	QuotePrecision             int      `json:"quotePrecision"`
	QuoteAssetPrecision        int      `json:"quoteAssetPrecision"`
	BaseCommissionPrecision    int      `json:"baseCommissionPrecision"`
	QuoteCommissionPrecision   int      `json:"quoteCommissionPrecision"`
	OrderTypes                 []string `json:"orderTypes"`
	IsSpotTradingAllowed       bool     `json:"isSpotTradingAllowed"`
	IsMarginTradingAllowed     bool     `json:"isMarginTradingAllowed"`
	QuoteAmountPrecision       string   `json:"quoteAmountPrecision"`
	BaseSizePrecision          string   `json:"baseSizePrecision"`
	Permissions                []string `json:"permissions"`
	Filters                    []string `json:"filters"`
	MaxQuoteAmount             string   `json:"maxQuoteAmount"`
	MakerCommission            string   `json:"makerCommission"`
	TakerCommission            string   `json:"takerCommission"`
	QuoteAmountPrecisionMarket string   `json:"quoteAmountPrecisionMarket"`
	MaxQuoteAmountMarket       string   `json:"maxQuoteAmountMarket"`
	FullName                   string   `json:"fullName"`
}

type ResponseGetExchangeInfo struct {
	Timezone        string                  `json:"timezone"`
	ServerTime      int                     `json:"serverTime"`
	RateLimits      []string                `json:"rateLimits"`
	ExchangeFilters []string                `json:"exchangeFilters"`
	Symbols         []SymbolGetExchangeInfo `json:"symbols"`
}

func GetExchangeInfo(ctx context.Context) (ResponseGetExchangeInfo, error) {
	url := "https://api.mexc.com/api/v3/exchangeInfo"
	resp, err := http.Get(url)
	if err != nil {
		return ResponseGetExchangeInfo{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return ResponseGetExchangeInfo{}, fmt.Errorf("error: %v:%v (%v)", resp.StatusCode, resp.Status, body)
	}

	if err != nil {
		return ResponseGetExchangeInfo{}, err
	}

	var res ResponseGetExchangeInfo
	if err := json.Unmarshal(body, &res); err != nil {
		return ResponseGetExchangeInfo{}, err
	}

	return res, nil
}
