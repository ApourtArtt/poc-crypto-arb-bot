package mexcsdk_test

import (
	"fmt"
	"testing"

	"github.com/ArbitrageCoin/crypto-sdk/src/third_parties/mexcsdk"
)

func TestGetBalance(t *testing.T) {
	resp, err := mexcsdk.GetBalance("...", "...")

	fmt.Println(resp)
	if err != nil {
		panic(err)
	}
}
