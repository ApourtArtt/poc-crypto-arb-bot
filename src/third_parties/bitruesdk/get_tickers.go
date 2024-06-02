package bitruesdk

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/shopspring/decimal"
)

type DataGetTickersInformation struct {
	Symbol        string          `json:"symbol"`
	ID            int             `json:"id"`
	Last          string          `json:"last"`
	LowestAsk     decimal.Decimal `json:"lowestAsk"`
	HighestBid    decimal.Decimal `json:"highestBid"`
	PercentChange string          `json:"percentChange"`
	BaseVolume    string          `json:"baseVolume"`
	QuoteVolume   string          `json:"quoteVolume"`
	IsFrozen      string          `json:"isFrozen"`
	High24hr      string          `json:"high24hr"`
	Low24hr       string          `json:"low24hr"`
}

type ResponseGetTickersInformation struct {
	Code string                               `json:"code"`
	Msg  string                               `json:"msg"`
	Data map[string]DataGetTickersInformation `json:"data"`
}

func GetTickersInformation() (*ResponseGetTickersInformation, error) {
	url := "https://www.bitrue.com/kline-api/public.json?command=returnTicker"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := ResponseGetTickersInformation{}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
