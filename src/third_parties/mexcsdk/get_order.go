package mexcsdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/shopspring/decimal"
)

type GetOrderParams struct {
	Symbol  string `json:"symbol"`
	OrderId string `json:"orderId,omitempty"`
}

type GetOrderResult struct {
	Symbol              string          `json:"symbol"`
	OrderId             string          `json:"orderId"`
	OrderListId         int64           `json:"orderListId"`
	ClientOrderId       string          `json:"clientOrderId"`
	Price               decimal.Decimal `json:"price"`
	OrigQty             decimal.Decimal `json:"origQty"`
	ExecutedQty         decimal.Decimal `json:"executedQty"`
	CummulativeQuoteQty decimal.Decimal `json:"cummulativeQuoteQty"`
	Status              string          `json:"status"`
	TimeInForce         string          `json:"timeInForce"`
	Type                OrderType       `json:"type"`
	Side                OrderSide       `json:"side"`
	StopPrice           decimal.Decimal `json:"stopPrice"`
	Time                int64           `json:"time"`
	UpdateTime          int64           `json:"updateTime"`
	IsWorking           bool            `json:"isWorking"`
	OrigQuoteOrderQty   decimal.Decimal `json:"origQuoteOrderQty"`
}

func GetOrder(apiKey, secretKey string, order GetOrderParams) (GetOrderResult, error) {
	baseUrl := "https://api.mexc.com/api/v3/order"
	client := &http.Client{}

	values := url.Values{}
	values.Set("symbol", order.Symbol)
	values.Set("orderId", order.OrderId)

	finalUrl := signQuery(baseUrl, values.Encode(), secretKey)
	req, _ := http.NewRequest("GET", finalUrl, nil)
	req.Header.Set("X-MEXC-APIKEY", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return GetOrderResult{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GetOrderResult{}, err
	}

	if resp.StatusCode >= 400 {
		return GetOrderResult{}, fmt.Errorf("error: %v:%v (%v)", resp.StatusCode, resp.Status, body)
	}

	var res GetOrderResult
	if err := json.Unmarshal(body, &res); err != nil {
		return GetOrderResult{}, err
	}

	return res, nil
}
