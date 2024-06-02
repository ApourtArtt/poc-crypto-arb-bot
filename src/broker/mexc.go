package broker

import (
	"context"
	"fmt"
	"strings"

	"github.com/ArbitrageCoin/crypto-sdk/src/coin"
	"github.com/ArbitrageCoin/crypto-sdk/src/database/database"
	"github.com/ArbitrageCoin/crypto-sdk/src/third_parties/mexcsdk"
	"github.com/shopspring/decimal"
)

type MEXC struct {
	config Config

	coins           CoinsMap
	exchangeCoins   ExchangeCoinsMap
	exchangeTickers ExchangeTickersMap

	tickersStatus map[string]coin.TickerStatus
	accountStatus AccountStatus
}

func NewMEXC(config Config) (IBroker, error) {
	return &MEXC{
		config:        config,
		tickersStatus: make(map[string]coin.TickerStatus),
		accountStatus: NewAccountStatus(false),
	}, nil
}

func (b MEXC) GetBrokerName() string { return b.config.InternalName }

func (b MEXC) GetTickersInformation(ctx context.Context) (map[coin.TickerPair]CoinAllInfo, error) {
	tickers, err := mexcsdk.GetBookTickers()
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

func (b MEXC) GetOrderBooks(ctx context.Context, ticker database.SelectExchangeTickersRow) (coin.OrderBook, error) {
	base := strings.ToUpper(ticker.Base)
	quote := strings.ToUpper(ticker.Quote)

	orders, err := mexcsdk.GetDepth(base + quote)
	if err != nil {
		return coin.OrderBook{}, err
	}

	orderbook := coin.OrderBook{
		Bids: make([]coin.Offer, len(orders.Bids)),
		Asks: make([]coin.Offer, len(orders.Asks)),
	}

	for i, bid := range orders.Bids {
		if !bid[1].GreaterThan(decimal.Zero) {
			break
		}
		orderbook.Bids[i] = coin.Offer{
			Price:    bid[0],
			Quantity: bid[1],
		}
	}

	for i, ask := range orders.Asks {
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

func (b MEXC) GetBalance(ctx context.Context) (map[coin.CoinBaseStr]coin.Balance, error) {
	coins, err := mexcsdk.GetBalance(b.config.Key, b.config.Secret)
	if err != nil {
		return nil, err
	}

	if len(coins.Balances) == 0 {
		return nil, nil
	}

	balance := make(map[coin.CoinBaseStr]coin.Balance)

	for _, c := range coins.Balances {
		if c.Free.Equal(decimal.Zero) {
			continue
		}
		balance[c.Asset] = coin.Balance{
			Quantity: c.Free,
			//DepositEnabled:  true,
			//WithdrawEnabled: true,
		}
	}

	return balance, nil
}

func (b *MEXC) RefreshCoinsInformation(coins CoinsMap, exchangeCoins ExchangeCoinsMap, exchangeTickers ExchangeTickersMap) {
	b.coins = coins
	b.exchangeCoins = exchangeCoins
	b.exchangeTickers = exchangeTickers
}

func (b *MEXC) RefreshExchangeInformation(ctx context.Context) error {
	respExchange, err := mexcsdk.GetExchangeInfo(ctx)
	if err != nil {
		return err
	}

	for _, ticker := range respExchange.Symbols {
		isLimitOrderAllowed := false
		for _, orderAllowed := range ticker.OrderTypes {
			if strings.ToUpper(orderAllowed) == "LIMIT" {
				isLimitOrderAllowed = true
				break
			}
		}

		isSpotAllowed := false
		for _, permission := range ticker.Permissions {
			if strings.ToUpper(permission) == "SPOT" {
				isSpotAllowed = true
				break
			}
		}

		b.tickersStatus[ticker.Symbol] = coin.TickerStatus{
			IsEnabled:            strings.ToUpper(ticker.Status) == "ENABLED",
			IsLimitOrderAllowed:  isLimitOrderAllowed,
			IsSpotTradingAllowed: ticker.IsSpotTradingAllowed && isSpotAllowed,
			IsBuyable:            strings.ToUpper(ticker.Status) == "ENABLED",
			IsSellable:           strings.ToUpper(ticker.Status) == "ENABLED",
		}
	}

	respAccount, err := mexcsdk.GetBalance(b.config.Key, b.config.Secret)
	if err != nil {
		return err
	}

	b.accountStatus.CanDeposit = respAccount.CanDeposit
	b.accountStatus.CanTrade = respAccount.CanTrade
	b.accountStatus.CanWithdraw = respAccount.CanWithdraw
	for _, permission := range respAccount.Permissions {
		if strings.ToUpper(permission) == "SPOT" {
			b.accountStatus.CanUseSpot = true
			break
		}
	}
	if strings.ToUpper(respAccount.AccountType) != "SPOT" {
		b.accountStatus.CanUseSpot = false
	}

	return nil
}

func (b *MEXC) Buy(ctx context.Context, ticker database.SelectExchangeTickersRow, maxPrice, quoteQuantity decimal.Decimal) error {
	symbol := strings.ToUpper(ticker.Base) + strings.ToUpper(ticker.Quote)

	postResp, err := mexcsdk.PostOrder(b.config.Key, b.config.Secret, mexcsdk.Order{
		Symbol:   symbol,
		Side:     mexcsdk.BUY,
		Type:     mexcsdk.FILL_OR_KILL,
		Quantity: quoteQuantity,
		Price:    maxPrice,
	})
	if err != nil {
		return err
	}

	getResp, err := mexcsdk.GetOrder(b.config.Key, b.config.Secret, mexcsdk.GetOrderParams{
		Symbol:  symbol,
		OrderId: postResp.OrderID,
	})
	if err != nil {
		return err
	}

	if strings.ToUpper(getResp.Status) != "FILLED" {
		return fmt.Errorf("the buy order has not been filled")
	}

	return nil
}

func (b *MEXC) Sell(ctx context.Context, ticker database.SelectExchangeTickersRow, minPrice, quoteQuantity decimal.Decimal) error {
	symbol := strings.ToUpper(ticker.Base) + strings.ToUpper(ticker.Quote)

	postResp, err := mexcsdk.PostOrder(b.config.Key, b.config.Secret, mexcsdk.Order{
		Symbol:   symbol,
		Side:     mexcsdk.SELL,
		Type:     mexcsdk.IMMEDIATE_OR_CANCEL,
		Quantity: quoteQuantity,
		Price:    minPrice,
	})
	if err != nil {
		return err
	}

	getResp, err := mexcsdk.GetOrder(b.config.Key, b.config.Secret, mexcsdk.GetOrderParams{
		Symbol:  symbol,
		OrderId: postResp.OrderID,
	})
	if err != nil {
		return err
	}

	if strings.ToUpper(getResp.Status) != "FILLED" {
		return fmt.Errorf("the sell order has not been filled")
	}

	return nil
}

func (b MEXC) CanBuyAndWithdraw(ctx context.Context, ticker database.SelectExchangeTickersRow) error {
	symbol := strings.ToUpper(ticker.Base) + strings.ToUpper(ticker.Quote)
	tickerStatus, ok := b.tickersStatus[symbol]
	if !ok {
		return fmt.Errorf("%v is not in tickerStatus", symbol)
	}

	if !tickerStatus.CanBeBought() {
		return fmt.Errorf("%v cannot be bought", symbol)
	}

	coinsNetwork, err := mexcsdk.GetAllDeposit(b.config.Key, b.config.Secret)
	if err != nil {
		return err
	}

	for _, coin := range coinsNetwork {
		if coin.Coin == ticker.Base {
			for _, network := range coin.NetworkList {
				if network.WithdrawEnable {
					return nil
				}
			}
		}
	}

	return fmt.Errorf("no network available")
}

func (b MEXC) CanDepositAndSell(ctx context.Context, ticker database.SelectExchangeTickersRow) error {
	symbol := strings.ToUpper(ticker.Base) + strings.ToUpper(ticker.Quote)
	tickerStatus, ok := b.tickersStatus[symbol]
	if !ok {
		return fmt.Errorf("%v is not in tickerStatus", symbol)
	}

	if !tickerStatus.CanBeSold() {
		return fmt.Errorf("%v cannot be sold", symbol)
	}

	coinsNetwork, err := mexcsdk.GetAllDeposit(b.config.Key, b.config.Secret)
	if err != nil {
		return err
	}

	for _, coin := range coinsNetwork {
		if coin.Coin == ticker.Base {
			for _, network := range coin.NetworkList {
				if network.DepositEnable {
					return nil
				}
			}
		}
	}

	return fmt.Errorf("no network available")
}
