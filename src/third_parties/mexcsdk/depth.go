package mexcsdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/shopspring/decimal"
)

type ResponseGetDepth struct {
	LastUpdateID int64               `json:"lastUpdateId"`
	Bids         [][]decimal.Decimal `json:"bids"`
	Asks         [][]decimal.Decimal `json:"asks"`
	Timestamp    int64               `json:"timestamp"`
}

func GetDepth(symbol string) (ResponseGetDepth, error) {
	url := "https://api.mexc.com/api/v3/depth?symbol=" + symbol
	resp, err := http.Get(url)
	if err != nil {
		return ResponseGetDepth{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return ResponseGetDepth{}, fmt.Errorf("error: %v:%v (%v)", resp.StatusCode, resp.Status, resp.Body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ResponseGetDepth{}, err
	}

	var res ResponseGetDepth
	if err := json.Unmarshal(body, &res); err != nil {
		return ResponseGetDepth{}, err
	}

	return res, nil
}
