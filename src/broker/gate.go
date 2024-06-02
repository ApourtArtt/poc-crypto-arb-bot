package broker

import (
	"context"
	"fmt"
	"strings"

	"github.com/ArbitrageCoin/crypto-sdk/src/coin"
	"github.com/ArbitrageCoin/crypto-sdk/src/database/database"
	"github.com/antihax/optional"
	"github.com/gateio/gateapi-go/v6"
	"github.com/shopspring/decimal"
)

type Gate struct {
	config Config

	coins           CoinsMap
	exchangeCoins   ExchangeCoinsMap
	exchangeTickers ExchangeTickersMap

	tickersStatus map[string]coin.TickerStatus
	accountStatus AccountStatus
}

func NewGate(config Config) (IBroker, error) {
	return &Gate{
		config:        config,
		tickersStatus: make(map[string]coin.TickerStatus),
		accountStatus: NewAccountStatus(false),
	}, nil
}

func (b Gate) GetBrokerName() string { return b.config.InternalName }

func (b Gate) GetTickersInformation(ctx context.Context) (map[coin.TickerPair]CoinAllInfo, error) {
	client := gateapi.NewAPIClient(gateapi.NewConfiguration())
	tickers, _, err := client.SpotApi.ListTickers(ctx, nil)
	if err != nil {
		return nil, err
	}

	tickersInfo := make(map[coin.TickerPair]CoinAllInfo)
	for _, ticker := range tickers {
		exchangeTicker, ok := b.exchangeTickers[ticker.CurrencyPair]
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

		bid, err := decimal.NewFromString(ticker.HighestBid)
		if err != nil {
			continue
		}
		if bid.Equal(decimal.Zero) {
			continue
		}
		ask, err := decimal.NewFromString(ticker.LowestAsk)
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

func (b Gate) GetOrderBooks(ctx context.Context, ticker database.SelectExchangeTickersRow) (coin.OrderBook, error) {
	currencyPair := strings.ToUpper(ticker.Base) + "_" + strings.ToUpper(ticker.Quote)

	client := gateapi.NewAPIClient(gateapi.NewConfiguration())
	orders, _, err := client.SpotApi.ListOrderBook(ctx, currencyPair, &gateapi.ListOrderBookOpts{
		Limit: optional.NewInt32(50),
	})
	if err != nil {
		return coin.OrderBook{}, err
	}

	orderbook := coin.OrderBook{
		Bids: make([]coin.Offer, len(orders.Bids)),
		Asks: make([]coin.Offer, len(orders.Asks)),
	}

	for i, bid := range orders.Bids {
		price, err := decimal.NewFromString(bid[0])
		if err != nil {
			return orderbook, err
		}
		quantity, err := decimal.NewFromString(bid[1])
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
		price, err := decimal.NewFromString(ask[0])
		if err != nil {
			return orderbook, err
		}
		quantity, err := decimal.NewFromString(ask[1])
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

func (b Gate) GetBalance(ctx context.Context) (map[coin.CoinBaseStr]coin.Balance, error) {
	config := gateapi.NewConfiguration()
	config.Key = b.config.Key
	config.Secret = b.config.Secret
	client := gateapi.NewAPIClient(config)
	coins, _, err := client.SpotApi.ListSpotAccounts(ctx, nil)
	if err != nil {
		return nil, err
	}

	if len(coins) == 0 {
		return nil, nil
	}

	balance := make(map[coin.CoinBaseStr]coin.Balance)

	for _, c := range coins {
		qty, err := decimal.NewFromString(c.Available)
		if err != nil {
			continue
		}
		if qty.Equal(decimal.Zero) {
			continue
		}
		balance[c.Currency] = coin.Balance{
			Quantity: qty,
			//DepositEnabled:  false,
			//WithdrawEnabled: false,
		}
	}

	/*coinInfos, _, err := client.SpotApi.ListCurrencies(ctx)
	if err != nil {
		return nil, err
	}

	for _, ci := range coinInfos {
		coinBase := ci.Currency
		// coinBase can be "USDT", but it can also be "USDT_BTC", "USDT_ALGO", etc
		// with the second part that represents the blockchain network on which it runs on
		// USDT might be blocked on BTC but available on ALGO.
		// Check all chains and if one chain makes withdrawal available, then it is.
		parts := strings.Split(ci.Currency, "_")
		if len(parts) == 2 && parts[1] == ci.Chain {
			coinBase = parts[0]
		}

		b, ok := balance[coinBase]
		if ok {
			b.DepositEnabled = b.DepositEnabled || !ci.DepositDisabled
			b.WithdrawEnabled = b.WithdrawEnabled || (!ci.WithdrawDisabled && !ci.WithdrawDelayed)
			balance[coinBase] = b
		}
	}*/

	return balance, nil
}

func (b *Gate) RefreshCoinsInformation(coins CoinsMap, exchangeCoins ExchangeCoinsMap, exchangeTickers ExchangeTickersMap) {
	b.coins = coins
	b.exchangeCoins = exchangeCoins
	b.exchangeTickers = exchangeTickers
}

func (b *Gate) RefreshExchangeInformation(ctx context.Context) error {
	config := gateapi.NewConfiguration()
	config.Key = b.config.Key
	config.Secret = b.config.Secret
	client := gateapi.NewAPIClient(config)

	respAccount, _, err := client.AccountApi.GetAccountDetail(ctx)
	if err != nil {
		return err
	}

	b.accountStatus.CanDeposit = respAccount.Key.Mode == 1
	b.accountStatus.CanTrade = respAccount.Key.Mode == 1
	b.accountStatus.CanUseSpot = respAccount.Key.Mode == 1
	b.accountStatus.CanWithdraw = respAccount.Key.Mode == 1

	respTickers, _, err := client.SpotApi.ListCurrencyPairs(ctx)
	if err != nil {
		return err
	}

	// untradable: cannot be bought or sold - buyable: can be bought - sellable: can be sold - tradable: can be bought or sold
	for _, ticker := range respTickers {
		tradableStatus := strings.ToUpper(ticker.TradeStatus)
		b.tickersStatus[ticker.Id] = coin.TickerStatus{
			IsEnabled:            tradableStatus != "UNTRADABLE",
			IsLimitOrderAllowed:  tradableStatus != "UNTRADABLE",
			IsSpotTradingAllowed: tradableStatus != "UNTRADABLE",
			IsBuyable:            tradableStatus == "BUYABLE" || tradableStatus == "TRADABLE",
			IsSellable:           tradableStatus == "SELLABLE" || tradableStatus == "TRADABLE",
		}
	}

	// **CAREFULL**
	// https://api.gateio.ws/api/v4/spot/currencies is bugged
	// Compared to https://api.gateio.ws/api/v4/wallet/currency_chains?currency=USDT
	// USDT has "BNB" on the first link, but not on the second
	// And on the website, we indeed cannot use BNB,
	// So using the first one would lead to lose of funds.
	// Solution for this exchange => Do the second request for each coin instead of
	// caching the first one. More time-greedy but safer on the long run...

	return nil
}

func (b *Gate) Buy(ctx context.Context, ticker database.SelectExchangeTickersRow, maxPrice, quoteQuantity decimal.Decimal) error {
	currencyPair := strings.ToUpper(ticker.Base) + "_" + strings.ToUpper(ticker.Quote)

	config := gateapi.NewConfiguration()
	config.Key = b.config.Key
	config.Secret = b.config.Secret
	client := gateapi.NewAPIClient(config)

	quoteQuantityReal := quoteQuantity.Div(maxPrice)

	_, _, err := client.SpotApi.CreateOrder(ctx, gateapi.Order{
		Account:      "spot",
		CurrencyPair: currencyPair,
		Side:         "buy",
		Type:         "limit",
		Amount:       quoteQuantityReal.String(),
		Price:        maxPrice.String(),
		TimeInForce:  "fok",
	})
	if err != nil {
		return err
	}

	return nil
}

func (b *Gate) Sell(ctx context.Context, ticker database.SelectExchangeTickersRow, minPrice, quoteQuantity decimal.Decimal) error {
	currencyPair := strings.ToUpper(ticker.Base) + "_" + strings.ToUpper(ticker.Quote)

	config := gateapi.NewConfiguration()
	config.Key = b.config.Key
	config.Secret = b.config.Secret
	client := gateapi.NewAPIClient(config)

	quoteQuantityReal := quoteQuantity.Div(minPrice)

	sellOrder, _, err := client.SpotApi.CreateOrder(ctx, gateapi.Order{
		Account:      "spot",
		CurrencyPair: currencyPair,
		Side:         "sell",
		Type:         "limit",
		Amount:       quoteQuantityReal.String(),
		Price:        minPrice.String(),
		TimeInForce:  "ioc",
	})
	if err != nil {
		return err
	}

	if strings.ToUpper(sellOrder.Status) != "CLOSED" {
		return fmt.Errorf("the sell order has not been filled")
	}

	return nil
}

func (b Gate) CanBuyAndWithdraw(ctx context.Context, ticker database.SelectExchangeTickersRow) error {
	currencyPair := strings.ToUpper(ticker.Base) + "_" + strings.ToUpper(ticker.Quote)
	tickerStatus, ok := b.tickersStatus[currencyPair]
	if !ok {
		return fmt.Errorf("%v is not in tickerStatus", currencyPair)
	}

	if !tickerStatus.CanBeBought() {
		return fmt.Errorf("%v cannot be bought", currencyPair)
	}

	config := gateapi.NewConfiguration()
	config.Key = b.config.Key
	config.Secret = b.config.Secret
	client := gateapi.NewAPIClient(config)
	networks, _, err := client.WalletApi.ListCurrencyChains(ctx, ticker.Base)
	if err != nil {
		return err
	}

	for _, network := range networks {
		if network.IsDisabled == 0 && network.IsWithdrawDisabled == 0 {
			return nil
		}
	}

	return nil
}

func (b Gate) CanDepositAndSell(ctx context.Context, ticker database.SelectExchangeTickersRow) error {
	currencyPair := strings.ToUpper(ticker.Base) + "_" + strings.ToUpper(ticker.Quote)
	tickerStatus, ok := b.tickersStatus[currencyPair]
	if !ok {
		return fmt.Errorf("%v is not in tickerStatus", currencyPair)
	}

	if !tickerStatus.CanBeSold() {
		return fmt.Errorf("%v cannot be sold", currencyPair)
	}

	config := gateapi.NewConfiguration()
	config.Key = b.config.Key
	config.Secret = b.config.Secret
	client := gateapi.NewAPIClient(config)
	networks, _, err := client.WalletApi.ListCurrencyChains(ctx, ticker.Base)
	if err != nil {
		return err
	}

	for _, network := range networks {
		if network.IsDisabled == 0 && network.IsDepositDisabled == 0 {
			return nil
		}
	}

	return nil
}
