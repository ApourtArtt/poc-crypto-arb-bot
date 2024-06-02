package broker

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ArbitrageCoin/crypto-sdk/src/coin"
	"github.com/ArbitrageCoin/crypto-sdk/src/database/database"
	"github.com/ArbitrageCoin/crypto-sdk/src/third_parties/xt_com"
	"github.com/shopspring/decimal"
)

type XT struct {
	config Config

	coins           CoinsMap
	exchangeCoins   ExchangeCoinsMap
	exchangeTickers ExchangeTickersMap
}

func NewXT(config Config) (IBroker, error) {
	return &XT{
		config: config,
	}, nil
}

func (b XT) GetBrokerName() string { return b.config.InternalName }

func (b XT) GetTickersInformation(ctx context.Context) (map[coin.TickerPair]CoinAllInfo, error) {
	client := xt_com.PublicHttpAPI{}
	resp := client.GetFullTicker(nil)
	var tickers xt_com.ResponseGetFullTicker
	if err := json.Unmarshal([]byte(resp.Data), &tickers); err != nil {
		return nil, err
	}

	tickersInfo := make(map[coin.TickerPair]CoinAllInfo)
	for _, ticker := range tickers.Result {
		exchangeTicker, ok := b.exchangeTickers[strings.ToUpper(ticker.Symbol)]
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

		if ticker.BidPrice.Equal(decimal.Zero) {
			continue
		}
		if ticker.AskPrice.Equal(decimal.Zero) {
			continue
		}

		tickersInfo[coin.TickerPair{
			Base:  baseID,
			Quote: quoteID,
		}] = CoinAllInfo{
			Values: coin.TickerValues{
				HighestBid: ticker.BidPrice,
				LowestAsk:  ticker.AskPrice,
			},
			ExchangeCoinBase:  exchangeCoinBase,
			ExchangeCoinQuote: exchangeCoinQuote,
			ExchangeTicker:    exchangeTicker,
		}
	}

	return tickersInfo, nil
}

func (b XT) GetOrderBooks(ctx context.Context, ticker database.SelectExchangeTickersRow) (coin.OrderBook, error) {
	base := strings.ToLower(ticker.Base)
	quote := strings.ToLower(ticker.Quote)

	client := xt_com.PublicHttpAPI{}
	resp := client.GetDepth(map[string]interface{}{
		"symbol": base + "_" + quote,
	})
	var orders xt_com.ResponseGetDepth
	if err := json.Unmarshal([]byte(resp.Data), &orders); err != nil {
		return coin.OrderBook{}, err
	}

	orderbook := coin.OrderBook{
		Bids: make([]coin.Offer, len(orders.Result.Bids)),
		Asks: make([]coin.Offer, len(orders.Result.Asks)),
	}

	for i, bid := range orders.Result.Bids {
		if !bid[1].GreaterThan(decimal.Zero) {
			break
		}
		orderbook.Bids[i] = coin.Offer{
			Price:    bid[0],
			Quantity: bid[1],
		}
	}

	for i, ask := range orders.Result.Asks {
		if !ask[1].GreaterThan(decimal.Zero) {
			break
		}
		orderbook.Asks[i] = coin.Offer{
			Price:    ask[0],
			Quantity: ask[1],
		}
	}

	orderbook.SortAsks()
	orderbook.SortBids()

	return orderbook, nil
}

func (b XT) GetBalance(ctx context.Context) (map[coin.CoinBaseStr]coin.Balance, error) {
	client := xt_com.SignedHttpAPI{
		Accesskey: b.config.Key,
		Secretkey: b.config.Secret,
	}
	resp := client.GetBalance(nil)
	var coins xt_com.ResultGetBalance
	if err := json.Unmarshal([]byte(resp.Data), &coins); err != nil {
		return nil, err
	}

	if len(coins.Assets) == 0 {
		return nil, nil
	}

	balance := make(map[coin.CoinBaseStr]coin.Balance)

	for _, c := range coins.Assets {
		if c.AvailableAmount.Equal(decimal.Zero) {
			continue
		}
		balance[strings.ToUpper(c.Currency)] = coin.Balance{
			Quantity: c.AvailableAmount,
			//DepositEnabled:  false,
			//WithdrawEnabled: false,
		}
	}

	/*data := xt_com.PublicHttpAPI{}.GetCoinsInfo()
	var coinInfo xt_com.ResponseGetCoinsInfo
	err := json.Unmarshal([]byte(data.Data), &coinInfo)
	if err != nil {
		return nil, err
	}

	for _, ci := range coinInfo.Result.Currencies {
		b, ok := balance[strings.ToUpper(ci.Currency)]
		if !ok {
			continue
		}

		b.DepositEnabled = ci.DepositStatus == 1
		b.WithdrawEnabled = ci.WithdrawStatus == 1
	}*/

	return balance, nil
}

func (b *XT) RefreshCoinsInformation(coins CoinsMap, exchangeCoins ExchangeCoinsMap, exchangeTickers ExchangeTickersMap) {
	b.coins = coins
	b.exchangeCoins = exchangeCoins
	b.exchangeTickers = exchangeTickers
}

func (b *XT) RefreshExchangeInformation(ctx context.Context) error {
	return nil
}

func (b *XT) Buy(ctx context.Context, ticker database.SelectExchangeTickersRow, maxPrice, quoteQuantity decimal.Decimal) error {
	return nil
}

func (b *XT) Sell(ctx context.Context, ticker database.SelectExchangeTickersRow, minPrice, quoteQuantity decimal.Decimal) error {
	return nil
}

func (b XT) CanBuyAndWithdraw(ctx context.Context, ticker database.SelectExchangeTickersRow) error {
	return nil
}

func (b XT) CanDepositAndSell(ctx context.Context, ticker database.SelectExchangeTickersRow) error {
	return nil
}
