package mexcsdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/shopspring/decimal"
)

type ResponseGetBookTickers struct {
	Symbol   string          `json:"symbol"`
	BidPrice decimal.Decimal `json:"bidPrice"`
	BidQty   *string         `json:"bidQty"`
	AskPrice decimal.Decimal `json:"askPrice"`
	AskQty   string          `json:"askQty"`
}

func GetBookTickers() ([]ResponseGetBookTickers, error) {
	url := "https://api.mexc.com/api/v3/ticker/bookTicker"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("error: %v:%v (%v)", resp.StatusCode, resp.Status, resp.Body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res []ResponseGetBookTickers
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return res, nil
}
