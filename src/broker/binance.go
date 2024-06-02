package broker

import (
	"context"
	"fmt"
	"strings"

	"github.com/ArbitrageCoin/crypto-sdk/src/coin"
	"github.com/ArbitrageCoin/crypto-sdk/src/database/database"
	binance_connector "github.com/binance/binance-connector-go"
	"github.com/shopspring/decimal"
)

type Binance struct {
	config Config

	coins           CoinsMap
	exchangeCoins   ExchangeCoinsMap
	exchangeTickers ExchangeTickersMap
}

func NewBinance(config Config) (IBroker, error) {
	return &Binance{
		config: config,
	}, nil
}

func (b Binance) GetBrokerName() string { return b.config.InternalName }

func (b Binance) GetTickersInformation(ctx context.Context) (map[coin.TickerPair]CoinAllInfo, error) {
	client := binance_connector.NewClient("", "")
	tickers, err := client.NewTickerBookTickerService().Do(ctx)
	if err != nil {
		return nil, err
	}

	tickersInfo := make(map[coin.TickerPair]CoinAllInfo)
	for _, ticker := range tickers {
		exchangeTicker, ok := b.exchangeTickers[ticker.Symbol]
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

		bid, err := decimal.NewFromString(ticker.BidPrice)
		if err != nil {
			continue
		}
		if bid.Equal(decimal.Zero) {
			continue
		}
		ask, err := decimal.NewFromString(ticker.AskPrice)
		if err != nil {
			continue
		}
		if ask.Equal(decimal.Zero) {
			continue
		}

		tickersInfo[coin.TickerPair{
			Base:  baseID,
			Quote: quoteID,
		}] = CoinAllInfo{
			Values: coin.TickerValues{
				HighestBid: bid,
				LowestAsk:  ask,
			},
			ExchangeCoinBase:  exchangeCoinBase,
			ExchangeCoinQuote: exchangeCoinQuote,
			ExchangeTicker:    exchangeTicker,
		}
	}

	return tickersInfo, nil
}

func (b Binance) GetOrderBooks(ctx context.Context, ticker database.SelectExchangeTickersRow) (coin.OrderBook, error) {
	base := strings.ToUpper(ticker.Base)
	quote := strings.ToUpper(ticker.Quote)

	client := binance_connector.NewClient(b.config.Key, b.config.Secret)
	client.TimeOffset = 1000
	orders, err := client.
		NewOrderBookService().
		Symbol(base + quote).
		Do(ctx)

	if err != nil {
		return coin.OrderBook{}, err
	}

	orderbook := coin.OrderBook{
		Bids: make([]coin.Offer, len(orders.Bids)),
		Asks: make([]coin.Offer, len(orders.Asks)),
	}

	for i, bid := range orders.Bids {
		price, err := decimal.NewFromString(bid[0].String())
		if err != nil {
			return orderbook, err
		}
		quantity, err := decimal.NewFromString(bid[1].String())
		if err != nil {
			return orderbook, err
		}
		if quantity.Equal(decimal.Zero) {
			break
		}
		orderbook.Bids[i] = coin.Offer{
			Price:    price,
			Quantity: quantity,
		}
	}

	for i, ask := range orders.Asks {
		price, err := decimal.NewFromString(ask[0].String())
		if err != nil {
			return orderbook, err
		}
		quantity, err := decimal.NewFromString(ask[1].String())
		if err != nil {
			return orderbook, err
		}
		if !quantity.GreaterThan(decimal.Zero) {
			break
		}
		orderbook.Asks[i] = coin.Offer{
			Price:    price,
			Quantity: quantity,
		}
	}

	orderbook.SortAsks()
	orderbook.SortBids()

	return orderbook, nil
}

func (b Binance) GetBalance(ctx context.Context) (map[coin.CoinBaseStr]coin.Balance, error) {
	client := binance_connector.NewClient(b.config.Key, b.config.Secret)
	client.TimeOffset = 1000
	coins, err := client.NewGetAllCoinsInfoService().Do(ctx)
	if err != nil {
		return nil, err
	}

	balance := make(map[coin.CoinBaseStr]coin.Balance)

	for _, c := range coins {
		qty, err := decimal.NewFromString(c.Free)
		if err != nil {
			continue
		}
		if qty.Equal(decimal.Zero) {
			continue
		}

		balance[c.Coin] = coin.Balance{
			Quantity: qty,
			//WithdrawEnabled: c.WithdrawAllEnable,
			//DepositEnabled:  c.DepositAllEnable,
		}
	}

	return balance, nil
}

func (b *Binance) RefreshCoinsInformation(coins CoinsMap, exchangeCoins ExchangeCoinsMap, exchangeTickers ExchangeTickersMap) {
	b.coins = coins
	b.exchangeCoins = exchangeCoins
	b.exchangeTickers = exchangeTickers
}

func (b *Binance) RefreshExchangeInformation(ctx context.Context) error {
	return nil
}

func (b *Binance) Buy(ctx context.Context, ticker database.SelectExchangeTickersRow, maxPrice, quoteQuantity decimal.Decimal) error {
	base := strings.ToUpper(ticker.Base)
	quote := strings.ToUpper(ticker.Quote)
	tickerStr := base + quote

	client := binance_connector.NewClient(b.config.Key, b.config.Secret)

	quoteQuantityReal := quoteQuantity.Div(maxPrice).InexactFloat64()

	newOrder, err := client.NewCreateOrderService().Symbol(tickerStr).
		Side("BUY").Type("LIMIT").TimeInForce("IOC").
		Price(maxPrice.InexactFloat64()).Quantity(quoteQuantityReal).
		Do(ctx)
	if err != nil {
		return err
	}

	newOrderTyped, ok := newOrder.(*binance_connector.CreateOrderResponseFULL)
	if !ok {
		return fmt.Errorf("newOrderTyped is not a binance_connector.CreateOrderResponseFULL")
	}

	if strings.ToUpper(newOrderTyped.Status) != "FILLED" {
		return fmt.Errorf("the buy order has not been filled")
	}

	return nil
}

func (b *Binance) Sell(ctx context.Context, ticker database.SelectExchangeTickersRow, minPrice, quoteQuantity decimal.Decimal) error {
	base := strings.ToUpper(ticker.Base)
	quote := strings.ToUpper(ticker.Quote)
	tickerStr := base + quote

	client := binance_connector.NewClient(b.config.Key, b.config.Secret)

	quoteQuantityReal := quoteQuantity.Div(minPrice).InexactFloat64()

	newOrder, err := client.NewCreateOrderService().Symbol(tickerStr).
		Side("SELL").Type("LIMIT").TimeInForce("IOC").
		Price(minPrice.InexactFloat64()).Quantity(quoteQuantityReal).
		Do(ctx)
	if err != nil {
		return err
	}

	newOrderTyped, ok := newOrder.(*binance_connector.CreateOrderResponseFULL)
	if !ok {
		return fmt.Errorf("newOrderTyped is not a binance_connector.CreateOrderResponseFULL: %v", newOrder)
	}

	if strings.ToUpper(newOrderTyped.Status) != "FILLED" {
		return fmt.Errorf("the sell order has not been filled")
	}
	return nil
}

func (b Binance) CanBuyAndWithdraw(ctx context.Context, ticker database.SelectExchangeTickersRow) error {
	return nil
}

func (b Binance) CanDepositAndSell(ctx context.Context, ticker database.SelectExchangeTickersRow) error {
	return nil
}
