package mexcsdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/shopspring/decimal"
)

type NetworkGetAllDeposit struct {
	Coin                    string          `json:"coin"`
	DepositDesc             *string         `json:"depositDesc"`
	DepositEnable           bool            `json:"depositEnable"`
	MinConfirm              int             `json:"minConfirm"`
	Name                    string          `json:"Name"`
	Network                 string          `json:"network"`
	NetWork                 string          `json:"netWork"` // What the fuck, notice the 'W'
	WithdrawEnable          bool            `json:"withdrawEnable"`
	WithdrawFee             decimal.Decimal `json:"withdrawFee"`
	WithdrawIntegerMultiple interface{}     `json:"withdrawIntegerMultiple"`
	WithdrawMax             decimal.Decimal `json:"withdrawMax"`
	WithdrawMin             decimal.Decimal `json:"withdrawMin"`
	SameAddress             bool            `json:"sameAddress"`
	Contract                *string         `json:"contract"`
	WithdrawTips            *string         `json:"withdrawTips"`
	DepositTips             *string         `json:"depositTips"`
}

// GetAllDepositResponse represents the information for a specific coin, including its network list.
type GetAllDepositResponse struct {
	Coin        string                 `json:"coin"`
	Name        string                 `json:"Name"`
	NetworkList []NetworkGetAllDeposit `json:"networkList"`
}

func GetAllDeposit(apiKey, secretKey string) ([]GetAllDepositResponse, error) {
	url := "https://api.mexc.com/api/v3/capital/config/getall"

	client := &http.Client{}
	finalUrl := sign(url, "", secretKey)
	req, _ := http.NewRequest("GET", finalUrl, nil)
	req.Header.Set("X-MEXC-APIKEY", apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("error: %v:%v (%v)", resp.StatusCode, resp.Status, body)
	}

	if err != nil {
		return nil, err
	}

	var res []GetAllDepositResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return res, nil
}
