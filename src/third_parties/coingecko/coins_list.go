package coingecko

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Coin struct {
	Id     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

func GetCoinsList(apiKey string) ([]Coin, error) {
	url := "https://api.coingecko.com/api/v3/coins/list?x-cg-pro-api-key=" + apiKey

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Error: %v:%v (%v)", resp.StatusCode, resp.Status, resp.Body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var coins []Coin
	if err := json.Unmarshal(body, &coins); err != nil {
		return nil, err
	}

	return coins, nil

}
