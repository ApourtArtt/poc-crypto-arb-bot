package mexcsdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/shopspring/decimal"
)

type PostOrderResponse struct {
	Symbol       string          `json:"symbol"`
	OrderID      string          `json:"orderId"`
	OrderListID  int64           `json:"orderListId"`
	Price        decimal.Decimal `json:"price"`
	OrigQty      decimal.Decimal `json:"origQty"`
	Type         string          `json:"type"`
	Side         string          `json:"side"`
	TransactTime int64           `json:"transactTime"`
}

type OrderSide string

const (
	BUY  OrderSide = "BUY"
	SELL OrderSide = "SELL"
)

type OrderType string

const (
	LIMIT               OrderType = "LIMIT"
	MARKET              OrderType = "MARKET"
	LIMIT_MAKER         OrderType = "LIMIT_MAKER"
	IMMEDIATE_OR_CANCEL OrderType = "IMMEDIATE_OR_CANCEL"
	FILL_OR_KILL        OrderType = "FILL_OR_KILL"
)

type Order struct {
	Symbol   string          `json:"symbol"`             // Symbol of the order (mandatory)
	Side     OrderSide       `json:"side"`               // Side of the order (BUY or SELL) (mandatory)
	Type     OrderType       `json:"type"`               // Type of the order (LIMIT, MARKET, etc.) (mandatory)
	Quantity decimal.Decimal `json:"quantity,omitempty"` // Quantity of the order (optional)
	//QuoteOrderQty    decimal.Decimal `json:"quoteOrderQty,omitempty"`     // Quote order quantity (optional)
	Price decimal.Decimal `json:"price,omitempty"` // Price of the order (optional)
	//NewClientOrderId string          `json:"newClientOrderId,omitempty"`  // New client order ID (optional)
	//RecvWindow       int64           `json:"recvWindow,omitempty"`        // Receive window, max 60000 (optional)
	//Timestamp        int64           `json:"timestamp"`                   // Timestamp of the order (mandatory)
}

func PostOrder(apiKey, secretKey string, order Order) (PostOrderResponse, error) {
	baseUrl := "https://api.mexc.com/api/v3/order"
	client := &http.Client{}

	values := url.Values{}
	values.Set("symbol", order.Symbol)
	values.Set("side", string(order.Side))
	values.Set("type", string(order.Type))
	values.Set("quantity", order.Quantity.String())
	values.Set("price", order.Price.String())

	finalUrl := signQuery(baseUrl, values.Encode(), secretKey)
	req, _ := http.NewRequest("POST", finalUrl, nil)
	req.Header.Set("X-MEXC-APIKEY", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return PostOrderResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return PostOrderResponse{}, err
	}

	if resp.StatusCode >= 400 {
		return PostOrderResponse{}, fmt.Errorf("error: %v:%v (%v)", resp.StatusCode, resp.Status, body)
	}

	var res PostOrderResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return PostOrderResponse{}, err
	}

	return res, nil
}
