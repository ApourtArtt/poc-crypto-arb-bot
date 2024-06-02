package broker

import (
	"context"

	"github.com/ArbitrageCoin/crypto-sdk/src/coin"
	"github.com/ArbitrageCoin/crypto-sdk/src/database/database"
	"github.com/ArbitrageCoin/crypto-sdk/src/third_parties/bitruesdk"
	"github.com/shopspring/decimal"
)

type Bitrue struct {
	config Config

	coins           CoinsMap
	exchangeCoins   ExchangeCoinsMap
	exchangeTickers ExchangeTickersMap
}

func NewBitrue(config Config) (IBroker, error) {
	return &Bitrue{
		config: config,
	}, nil
}

func (b Bitrue) GetBrokerName() string { return b.config.InternalName }

func (b Bitrue) GetTickersInformation(ctx context.Context) (map[coin.TickerPair]CoinAllInfo, error) {
	tickers, err := bitruesdk.GetTickersInformation()
	if err != nil {
		return nil, err
	}

	tickersInfo := make(map[coin.TickerPair]CoinAllInfo)
	for symbol, ticker := range tickers.Data {
		// Note: ticker.Symbol is always "", so use symbol instead
		exchangeTicker, ok := b.exchangeTickers[symbol]
		if !ok {
			continue
		}

		exchangeCoinBase, ok := b.exchangeCoins[exchangeTicker.BaseExchCoinID]
		if !ok {
			continue
		}
		exchangeCoinQuote, ok := b.exchangeCoins[exchangeTicker.QuoteExchCoinID]
		if !ok {
			continue
		}

		baseID := exchangeCoinBase.CoinID
		quoteID := exchangeCoinQuote.CoinID

		if ticker.HighestBid.Equal(decimal.Zero) {
			continue
		}
		if ticker.LowestAsk.Equal(decimal.Zero) {
			continue
		}

		tickersInfo[coin.TickerPair{
			Base:  baseID,
			Quote: quoteID,
		}] = CoinAllInfo{
			Values: coin.TickerValues{
				HighestBid: ticker.HighestBid,
				LowestAsk:  ticker.LowestAsk,
			},
			ExchangeCoinBase:  exchangeCoinBase,
			ExchangeCoinQuote: exchangeCoinQuote,
			ExchangeTicker:    exchangeTicker,
		}
	}

	return tickersInfo, nil
}

func (b Bitrue) GetOrderBooks(ctx context.Context, ticker database.SelectExchangeTickersRow) (coin.OrderBook, error) {
	return coin.OrderBook{}, nil
}

func (b Bitrue) GetBalance(ctx context.Context) (map[coin.CoinBaseStr]coin.Balance, error) {
	coins, err := bitruesdk.GetBalance(b.config.Key, b.config.Secret)
	if err != nil {
		return nil, err
	}

	balance := make(map[coin.CoinBaseStr]coin.Balance)

	for _, c := range coins.Balances {
		if c.Free.Equal(decimal.Zero) {
			continue
		}
		balance[c.Asset] = coin.Balance{
			Quantity: c.Free,
			// coins.CanDeposit and coins.CanWithdraw are always false...
			//DepositEnabled:  true, //coins.CanDeposit,
			//WithdrawEnabled: true, //coins.CanWithdraw,
		}
	}

	return balance, nil
}

func (b *Bitrue) RefreshCoinsInformation(coins CoinsMap, exchangeCoins ExchangeCoinsMap, exchangeTickers ExchangeTickersMap) {
	b.coins = coins
	b.exchangeCoins = exchangeCoins
	b.exchangeTickers = exchangeTickers
}

func (b *Bitrue) RefreshExchangeInformation(ctx context.Context) error {
	return nil
}

func (b *Bitrue) Buy(ctx context.Context, ticker database.SelectExchangeTickersRow, maxPrice, quoteQuantity decimal.Decimal) error {
	return nil
}

func (b *Bitrue) Sell(ctx context.Context, ticker database.SelectExchangeTickersRow, minPrice, quoteQuantity decimal.Decimal) error {
	return nil
}

func (b Bitrue) CanBuyAndWithdraw(ctx context.Context, ticker database.SelectExchangeTickersRow) error {
	return nil
}

func (b Bitrue) CanDepositAndSell(ctx context.Context, ticker database.SelectExchangeTickersRow) error {
	return nil
}
