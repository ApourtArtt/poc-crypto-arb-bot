package broker_test

import (
	"context"
	"testing"

	"github.com/ArbitrageCoin/crypto-sdk/src/broker"
)

func TestBitrueGetTickersInformation(t *testing.T) {
	bitrue, _ := broker.NewBitrue(broker.Config{})

	bitrue.GetTickersInformation(context.Background())
}

func TestBitrueGetBalance(t *testing.T) {
	bitrue, _ := broker.NewBitrue(broker.Config{
		InternalName: "Bitrue",
		Key:          "...",
		Secret:       "...",
	})

	bitrue.GetBalance(context.Background())
}
