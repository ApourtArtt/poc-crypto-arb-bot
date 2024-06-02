package mexcsdk_test

import (
	"context"
	"testing"

	"github.com/ArbitrageCoin/crypto-sdk/src/third_parties/mexcsdk"
)

func TestGetExchangeInfo(t *testing.T) {
	mexcsdk.GetExchangeInfo(context.Background())
}
