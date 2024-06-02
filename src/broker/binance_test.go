package broker_test

import (
	"context"
	"testing"
	"time"

	"github.com/ArbitrageCoin/crypto-sdk/src/broker"
	"github.com/ArbitrageCoin/crypto-sdk/src/database/database"
	"github.com/shopspring/decimal"
)

func TestBinanceGetTickersInformation(t *testing.T) {
	binance, _ := broker.NewBinance(broker.Config{})

	binance.GetTickersInformation(context.Background())
}

func TestBinanceGetDepositInformation(t *testing.T) {
	binance, _ := broker.NewBinance(broker.Config{
		Key:    "...",
		Secret: "...",
	})

	binance.RefreshCoinsInformation(nil, nil, nil)
}

func TestBinanceGetBalance(t *testing.T) {
	binance, _ := broker.NewBinance(broker.Config{
		Key:    "...",
		Secret: "...",
	})

	_, err := binance.GetBalance(context.Background())
	if err != nil {
		panic(err)
	}
}

func TestBinanceBuy(t *testing.T) {
	binance, _ := broker.NewBinance(broker.Config{
		Key:            "...",
		Secret:         "...",
		RetryTimerHTTP: 1000 * time.Millisecond,
	})

	err := binance.Buy(context.Background(), database.SelectExchangeTickersRow{
		Base:  "TAO",
		Quote: "USDT",
	}, decimal.NewFromFloat(400.), decimal.NewFromFloat(5.))

	if err != nil {
		panic(err)
	}
}

func TestBinanceSell(t *testing.T) {
	binance, _ := broker.NewBinance(broker.Config{
		Key:            "...",
		Secret:         "...",
		RetryTimerHTTP: 1000 * time.Millisecond,
	})

	err := binance.Sell(context.Background(), database.SelectExchangeTickersRow{
		Base:  "TAO",
		Quote: "USDT",
	}, decimal.NewFromFloat(400.), decimal.NewFromFloat(5.))

	if err != nil {
		panic(err)
	}
}
