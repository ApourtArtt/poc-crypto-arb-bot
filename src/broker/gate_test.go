package broker_test

import (
	"context"
	"testing"

	"github.com/ArbitrageCoin/crypto-sdk/src/broker"
	"github.com/ArbitrageCoin/crypto-sdk/src/database/database"
	"github.com/shopspring/decimal"
)

func TestGateGetTickersInformation(t *testing.T) {
	gate, _ := broker.NewGate(broker.Config{})

	gate.GetTickersInformation(context.Background())
}

func TestGateGetBalance(t *testing.T) {
	gate, _ := broker.NewGate(broker.Config{
		InternalName: "Gate",
		Key:          "...",
		Secret:       "...",
	})

	gate.GetBalance(context.Background())
}

func TestGateBuy(t *testing.T) {
	gate, _ := broker.NewGate(broker.Config{
		InternalName: "Gate",
		Key:          "...",
		Secret:       "...",
	})

	gate.Buy(context.Background(), database.SelectExchangeTickersRow{
		Base:  "TAO",
		Quote: "USDT",
	}, decimal.NewFromFloat(390.), decimal.NewFromFloat(5.))
}

func TestGateSell(t *testing.T) {
	gate, _ := broker.NewGate(broker.Config{
		InternalName: "Gate",
		Key:          "...",
		Secret:       "...",
	})

	price := decimal.NewFromFloat(385.)
	toSell := price.Mul(decimal.NewFromFloat(0.0127871))
	gate.Sell(context.Background(), database.SelectExchangeTickersRow{
		Base:  "TAO",
		Quote: "USDT",
	}, price, toSell)
}

func TestGateRefreshExchangeInformation(t *testing.T) {
	gate, _ := broker.NewGate(broker.Config{
		InternalName: "Gate",
		Key:          "...",
		Secret:       "...",
	})

	gate.RefreshExchangeInformation(context.Background())
}
