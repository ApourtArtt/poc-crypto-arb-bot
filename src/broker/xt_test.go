package broker_test

import (
	"context"
	"testing"

	"github.com/ArbitrageCoin/crypto-sdk/src/broker"
	"github.com/ArbitrageCoin/crypto-sdk/src/database/database"
)

func TestXTGetOrderBooks(t *testing.T) {
	xt, _ := broker.NewXT(broker.Config{})

	_, err := xt.GetOrderBooks(context.Background(), database.SelectExchangeTickersRow{
		Base:  "btc",
		Quote: "usdt",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestXTGetTickersInfo(t *testing.T) {
	xt, _ := broker.NewXT(broker.Config{})

	_, err := xt.GetTickersInformation(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestXTGetBalance(t *testing.T) {
	xt, _ := broker.NewXT(broker.Config{
		InternalName: "XT",
		Key:          "...",
		Secret:       "...",
	})

	_, err := xt.GetBalance(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}
