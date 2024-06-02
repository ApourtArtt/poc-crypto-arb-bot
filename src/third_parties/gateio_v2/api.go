package gateio_v2

import (
	"encoding/json"
	"io"
	"net/http"
)

// https://github.com/gateio/rest-v2/blob/master/go/gateioApi.go
// https://www.gate.io/fr/api2

type DataGetMarketInfo struct {
	No          int         `json:"no"`
	Symbol      string      `json:"symbol"`
	Name        string      `json:"name"`
	NameEn      string      `json:"name_en"`
	NameCn      string      `json:"name_cn"`
	Pair        string      `json:"pair"`
	Rate        string      `json:"rate"`
	VolA        string      `json:"vol_a"`
	VolB        string      `json:"vol_b"`
	CurrA       string      `json:"curr_a"`
	CurrB       string      `json:"curr_b"`
	CurrSuffix  string      `json:"curr_suffix"`
	RatePercent string      `json:"rate_percent"`
	Trend       string      `json:"trend"`
	Supply      interface{} `json:"supply"`    // Wtf. Can be int(0) or string(value)
	MarketCap   interface{} `json:"marketcap"` // Same
	Lq          string      `json:"lq"`
	PRate       int         `json:"p_rate"`
	High        string      `json:"high"`
	Low         string      `json:"low"`
}

type ResponseGetMarketInfo struct {
	Result string              `json:"result"`
	Data   []DataGetMarketInfo `json:"data"`
}

func GetMarketInfo() (*ResponseGetMarketInfo, error) {
	url := "https://data.gateapi.io/api2/1/marketlist"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := ResponseGetMarketInfo{}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
