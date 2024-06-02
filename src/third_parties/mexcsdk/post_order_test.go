package mexcsdk_test

import (
	"fmt"
	"testing"

	"github.com/ArbitrageCoin/crypto-sdk/src/third_parties/mexcsdk"
	"github.com/shopspring/decimal"
)

func TestPostOrder(t *testing.T) {
	resp, err := mexcsdk.PostOrder(
		"...", "...", mexcsdk.Order{
			Symbol:   "USDCUSDT",
			Side:     mexcsdk.BUY,
			Type:     mexcsdk.IMMEDIATE_OR_CANCEL,
			Quantity: decimal.NewFromFloat(6),
			Price:    decimal.NewFromFloat(0.9),
		})

	fmt.Println(resp)
	if err != nil {
		panic(err)
	}
}
