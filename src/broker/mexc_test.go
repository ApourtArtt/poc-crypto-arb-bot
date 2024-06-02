package broker_test

import (
	"context"
	"testing"

	"github.com/ArbitrageCoin/crypto-sdk/src/broker"
)

func TestMEXCGetTickersInformation(t *testing.T) {
	mexc, _ := broker.NewMEXC(broker.Config{})

	mexc.GetTickersInformation(context.Background())
}

func TestMEXCGetBalance(t *testing.T) {
	mexc, _ := broker.NewMEXC(broker.Config{
		InternalName: "MEXC",
		Key:          "...",
		Secret:       "...",
	})

	_, err := mexc.GetBalance(context.Background())
	if err != nil {
		panic(err)
	}
}

func TestMEXCRefreshExchangeInformation(t *testing.T) {
	mexc, _ := broker.NewMEXC(broker.Config{
		InternalName: "MEXC",
		Key:          "...",
		Secret:       "...",
	})

	err := mexc.RefreshExchangeInformation(context.Background())
	if err != nil {
		panic(err)
	}
}
