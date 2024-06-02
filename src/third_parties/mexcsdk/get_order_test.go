package mexcsdk_test

import (
	"fmt"
	"testing"

	"github.com/ArbitrageCoin/crypto-sdk/src/third_parties/mexcsdk"
)

func TestGetOrder(t *testing.T) {
	resp, err := mexcsdk.GetOrder(
		"...", "...", mexcsdk.GetOrderParams{
			Symbol:  "USDCUSDT",
			OrderId: "...",
		})

	fmt.Println(resp)
	if err != nil {
		panic(err)
	}
}
