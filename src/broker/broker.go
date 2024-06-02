package broker

import (
	"context"

	"github.com/ArbitrageCoin/crypto-sdk/src/coin"
	"github.com/ArbitrageCoin/crypto-sdk/src/database/database"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CoinsMap = map[uuid.UUID]database.Coin
type ExchangeCoinsMap = map[uuid.UUID]database.SelectExchangeCoinsRow
type ExchangeTickersMap = map[string]database.SelectExchangeTickersRow

type CoinAllInfo struct {
	Values            coin.TickerValues
	ExchangeCoinBase  database.SelectExchangeCoinsRow
	ExchangeCoinQuote database.SelectExchangeCoinsRow
	ExchangeTicker    database.SelectExchangeTickersRow
}

type IBroker interface {
	// GetBrokerName returns the value that is configured in config.json
	GetBrokerName() string

	// GetTickersInformation returns the best bid and ask for each coin at once
	GetTickersInformation(ctx context.Context) (map[coin.TickerPair]CoinAllInfo, error)
	// GetDepositInformation returns the available network for a specific coin. Only coin.Base is used
	//GetDepositInformation(ctx context.Context, coin database.SelectExchangeTickersRow) error

	// GetOrderBooks returns the list of the best bid*s* and ask*s* for a specific ticker
	GetOrderBooks(ctx context.Context, ticker database.SelectExchangeTickersRow) (coin.OrderBook, error)
	// GetBalance retrieve the entire balance for the Spot account, for each tokens
	GetBalance(ctx context.Context) (map[coin.CoinBaseStr]coin.Balance, error)

	// RefreshCoinsInformation sets the data for each coins, from the three databases: coins, exchange_coins, exchange_tickers
	RefreshCoinsInformation(coins CoinsMap, exchangeCoins ExchangeCoinsMap, exchangeTickers ExchangeTickersMap)
	// RefreshExchangeInformation refreshes status about the account and the state of the different coins/tickers (whether they are enabled, etc)
	RefreshExchangeInformation(ctx context.Context) error

	// Buy sets a FOK buy order for the specified ticker, buying the equivalent of quoteQuantity, at a maximum price of maxPrice
	Buy(ctx context.Context, ticker database.SelectExchangeTickersRow, maxPrice, quoteQuantity decimal.Decimal) error
	// Sell sets a IOC sell order for the specified ticker, selling the equivalent of quoteQuantity, at a minimum price of minPrice
	Sell(ctx context.Context, ticker database.SelectExchangeTickersRow, minPrice, quoteQuantity decimal.Decimal) error

	CanBuyAndWithdraw(ctx context.Context, ticker database.SelectExchangeTickersRow) error
	CanDepositAndSell(ctx context.Context, ticker database.SelectExchangeTickersRow) error
}
