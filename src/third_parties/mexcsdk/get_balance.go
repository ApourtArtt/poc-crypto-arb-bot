package mexcsdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/shopspring/decimal"
)

type BalanceGetBalance struct {
	Asset  string          `json:"asset"`
	Free   decimal.Decimal `json:"free"`
	Locked decimal.Decimal `json:"locked"`
}

type ResponseGetBalance struct {
	MakerCommission  *int                `json:"makerCommission"`
	TakerCommission  *int                `json:"takerCommission"`
	BuyerCommission  *int                `json:"buyerCommission"`
	SellerCommission *int                `json:"sellerCommission"`
	CanTrade         bool                `json:"canTrade"`
	CanWithdraw      bool                `json:"canWithdraw"`
	CanDeposit       bool                `json:"canDeposit"`
	UpdateTime       *int                `json:"updateTime"`
	AccountType      string              `json:"accountType"`
	Balances         []BalanceGetBalance `json:"balances"`
	Permissions      []string            `json:"permissions"`
}

func GetBalance(apiKey, secretKey string) (ResponseGetBalance, error) {
	url := "https://api.mexc.com/api/v3/account"

	finalUrl := sign(url, "", secretKey)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", finalUrl, nil)
	req.Header.Set("X-MEXC-APIKEY", apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return ResponseGetBalance{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return ResponseGetBalance{}, fmt.Errorf("error: %v:%v (%v)", resp.StatusCode, resp.Status, body)
	}

	if err != nil {
		return ResponseGetBalance{}, err
	}

	var res ResponseGetBalance
	if err := json.Unmarshal(body, &res); err != nil {
		return ResponseGetBalance{}, err
	}

	return res, nil
}
