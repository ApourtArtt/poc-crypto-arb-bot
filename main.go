package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ArbitrageCoin/crypto-sdk/src/aggregator"
	"github.com/ArbitrageCoin/crypto-sdk/src/broker"
	"github.com/ArbitrageCoin/crypto-sdk/src/coin"
	"github.com/ArbitrageCoin/crypto-sdk/src/database/database"
	"github.com/ArbitrageCoin/crypto-sdk/src/third_parties/coingecko"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type config struct {
	Brokers map[string]broker.Config
}

func getAllCoinsInfo(exchanges map[string]database.Exchange) (broker.CoinsMap, map[string]broker.ExchangeCoinsMap, map[string]broker.ExchangeTickersMap) {
	db, err := database.NewDatabase("postgres", "postgres", "postgres")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	coins, err := db.Queries.SelectAllCoins(context.Background())
	if err != nil {
		panic(err)
	}

	coinsMap := make(map[uuid.UUID]database.Coin)
	for _, coin := range coins {
		coinsMap[coin.ID] = coin
	}

	exchangeCoins, err := db.Queries.SelectExchangeCoins(context.Background())
	if err != nil {
		panic(err)
	}

	exchangeToExchangeCoins := make(map[string]broker.ExchangeCoinsMap)
	for _, exchange := range exchanges {
		exchangeToExchangeCoins[exchange.Name] = make(broker.ExchangeCoinsMap)
	}

	for _, ec := range exchangeCoins {
		if !ec.ExchangeName.Valid {
			panic(fmt.Errorf("a row is not valid: %v", ec))
		}
		if ec.ExchangeName.String == "Binance" && strings.ToLower(ec.Name) == "ethereum" {
			fmt.Println("")
		}
		exchangeToExchangeCoins[ec.ExchangeName.String][ec.ID] = ec
	}

	tickers, err := db.Queries.SelectExchangeTickers(context.Background())
	if err != nil {
		panic(err)
	}

	tickerToExchangeTickers := make(map[string]broker.ExchangeTickersMap)
	for _, exchange := range exchanges {
		tickerToExchangeTickers[exchange.Name] = make(broker.ExchangeTickersMap)
	}
	for _, ticker := range tickers {
		tickerStr := fmt.Sprintf(exchanges[ticker.ExchangeName].TickerFormat, ticker.Base, ticker.Quote)
		tickerToExchangeTickers[ticker.ExchangeName][tickerStr] = ticker
	}

	// map[coins.id]coin
	// map[exchange.Name]map[exchange_coins.id]exchange_coin
	// map[exchange.Name]map[ticker_generated%s%s]exchangeticker
	// BTCUSDT

	// map[coins.id]coin, map[exchange.name]map[exchange_coin.id]exchange_coin, map[exchange.name]map[struct{exchange_ticker.]

	return coinsMap, exchangeToExchangeCoins, tickerToExchangeTickers
}

func loadBrokers() map[string]broker.IBroker {
	brokers := make(map[string]broker.IBroker)

	data, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	var c config
	if err := json.Unmarshal(data, &c); err != nil {
		panic(err)
	}

	binance, _ := broker.NewBinance(c.Brokers["Binance"])
	brokers[binance.GetBrokerName()] = binance

	mexc, _ := broker.NewMEXC(c.Brokers["MEXC"])
	brokers[mexc.GetBrokerName()] = mexc

	gate, _ := broker.NewGate(c.Brokers["Gate"])
	brokers[gate.GetBrokerName()] = gate

	xt, _ := broker.NewXT(c.Brokers["XT"])
	brokers[xt.GetBrokerName()] = xt

	bitrue, _ := broker.NewBitrue(c.Brokers["Bitrue"])
	brokers[bitrue.GetBrokerName()] = bitrue

	return brokers
}

var (
	brokers                               = loadBrokers()
	exchanges                             = getExchanges()
	coins, exchangeCoins, exchangeTickers = getAllCoinsInfo(exchanges)
)

func main() {
	//mergeNetworks()
	//populateDb()
	getOpportunities()
}

type arbitrageResult struct {
	quantityToBuy decimal.Decimal
	usdForBuying  decimal.Decimal
	usdForSelling decimal.Decimal
}

func calculateArbitrage(asks, bids coin.OrderBook) arbitrageResult {
	results := arbitrageResult{
		quantityToBuy: decimal.Zero,
		usdForBuying:  decimal.Zero,
		usdForSelling: decimal.Zero,
	}

	minProfitabilityPercent, _ := decimal.NewFromString("1.1")
	maximumToSpend, _ := decimal.NewFromString("1000")

	askIndex := 0
out:
	for _, bid := range bids.Bids {
		for askIndex < len(asks.Asks) {
			ask := asks.Asks[askIndex]

			// If the buyer is offering a price lower than the price we can buy it, taking into account the minimum profitability, then stop
			if bid.Price.GreaterThan(ask.Price.Mul(minProfitabilityPercent)) {
				break out
			}

			canSpend := maximumToSpend.Sub(results.usdForBuying)

			// If there are not enough quantity in the current bid for the slot we want to buy
			if bid.Quantity.LessThan(ask.Quantity) {
				// We take the quantity of the bid, but the price of the ask, as the bid price is
				// the price we well it to. The price we pay for 1 of quantity is ask.Price
				askTotalPrice := ask.Price.Mul(bid.Quantity)
				// If we cannot buy the whole bid
				if askTotalPrice.GreaterThan(canSpend) {
					qty := canSpend.Div(ask.Price)
					results.quantityToBuy = results.quantityToBuy.Add(qty)
					results.usdForBuying = results.usdForBuying.Add(ask.Price.Mul(qty))
					results.usdForSelling = results.usdForSelling.Add(bid.Price.Mul(qty))
					break out
				}

				// If we can buy the whole bid, buy it and check the next one
				results.quantityToBuy = results.quantityToBuy.Add(bid.Quantity)
				results.usdForBuying = results.usdForBuying.Add(ask.Price.Mul(bid.Quantity))
				results.usdForSelling = results.usdForSelling.Add(bid.Price.Mul(bid.Quantity))
				asks.Asks[askIndex].Quantity = ask.Quantity.Sub(bid.Quantity)
				break // Move to another bid
			}

			askTotalPrice := ask.Price.Mul(ask.Quantity)
			//If we cannot buy the entire offer, buy a chunk and stop
			if askTotalPrice.GreaterThanOrEqual(canSpend) {
				qty := canSpend.Div(ask.Price)
				results.usdForBuying = results.usdForBuying.Add(qty.Mul(ask.Price))
				results.usdForSelling = results.usdForSelling.Add(qty.Mul(bid.Price))
				results.quantityToBuy = results.quantityToBuy.Add(qty)
				break out
			}

			// Otherwise just buy the entire ask and move to another ask
			results.quantityToBuy = results.quantityToBuy.Add(ask.Quantity)
			results.usdForBuying = results.usdForBuying.Add(askTotalPrice)
			results.usdForSelling = results.usdForSelling.Add(ask.Quantity.Mul(bid.Price))
			bid.Quantity = bid.Quantity.Sub(ask.Quantity)
			askIndex++
		}
	}

	return results
}

type analyzeResult struct {
	ExchangeBuy  string
	ExchangeSell string
	Ticker       coin.TickerPair
	Results      arbitrageResult
}

func realAnalyze(tickerPair coin.TickerPair, exchanges map[string]broker.CoinAllInfo) analyzeResult {
	minProfit, _ := decimal.NewFromString("1.1")
	for brokerName, tickerValues := range exchanges {
		for brokerName2, tickerValues2 := range exchanges {
			if brokerName == brokerName2 {
				continue
			}

			if tickerValues.Values.HighestBid.GreaterThan(tickerValues2.Values.LowestAsk.Mul(minProfit)) {
				ob1, err := brokers[brokerName].GetOrderBooks(context.Background(), tickerValues.ExchangeTicker)
				if err != nil {
					panic(err)
				}
				ob2, err := brokers[brokerName2].GetOrderBooks(context.Background(), tickerValues2.ExchangeTicker)
				if err != nil {
					panic(err)
				}

				results := calculateArbitrage(ob1, ob2)
				if results.quantityToBuy == decimal.Zero || results.usdForBuying == decimal.Zero || results.usdForSelling == decimal.Zero {
					return analyzeResult{}
				}

				return analyzeResult{
					ExchangeBuy:  brokerName,
					ExchangeSell: brokerName2,
					Results:      results,
					Ticker:       tickerPair,
				}
			}
		}
	}
	return analyzeResult{}
}

func analyze(tickers map[coin.TickerPair]map[string]broker.CoinAllInfo) {
	ch := make(chan analyzeResult)

	go func() {
		wg := sync.WaitGroup{}
		wg.Add(len(tickers))

		for tickerPair, exchanges := range tickers {
			if len(exchanges) < 2 {
				wg.Done()
				continue
			}

			go func(exchanges map[string]broker.CoinAllInfo, tickerPair coin.TickerPair) {
				defer wg.Done()
				res := realAnalyze(tickerPair, exchanges)
				if res.ExchangeBuy == "" || res.ExchangeSell == "" {
					return
				}

				res.Ticker = tickerPair
				ch <- res
			}(exchanges, tickerPair)
		}

		wg.Wait()
		close(ch)
	}()

	var results []analyzeResult
	for res := range ch {
		results = append(results, res)
	}

	for _, res := range results {
		ticker := tickers[res.Ticker]
		tickerBuy := ticker[res.ExchangeBuy]
		tickerSell := ticker[res.ExchangeSell]
		if err := brokers[res.ExchangeBuy].CanBuyAndWithdraw(context.Background(), tickerBuy.ExchangeTicker); err != nil {
			continue
		}
		if err := brokers[res.ExchangeSell].CanDepositAndSell(context.Background(), tickerSell.ExchangeTicker); err != nil {
			continue
		}

		fmt.Println(tickerBuy.ExchangeCoinBase.Base, "_", tickerBuy.ExchangeCoinQuote.Base, "Buying from", res.ExchangeBuy, "Selling on", res.ExchangeSell, res.Results.quantityToBuy.String(), res.Results.usdForBuying.String(), res.Results.usdForSelling.String())
		fmt.Println()
	}
}

func getOpportunities() {
	for exchangeName, _ := range exchanges {
		brokers[exchangeName].RefreshCoinsInformation(coins, exchangeCoins[exchangeName], exchangeTickers[exchangeName])
		if err := brokers[exchangeName].RefreshExchangeInformation(context.Background()); err != nil {
			panic(err)
		}
	}

	for {
		allTickers := make(map[string]map[coin.TickerPair]broker.CoinAllInfo)
		for _, b := range brokers {
			answTickers, err := b.GetTickersInformation(context.Background())
			if err != nil {
				panic(err)
			}
			allTickers[b.GetBrokerName()] = answTickers
		}

		allTickersSorted := make(map[coin.TickerPair]map[string]broker.CoinAllInfo)
		for brokerName, tickers := range allTickers {
			for tickerPair, tickerValues := range tickers {
				tickerValuesSorted, ok := allTickersSorted[tickerPair]
				if !ok {
					tickerValuesSorted = make(map[string]broker.CoinAllInfo)
				}
				tickerValuesSorted[brokerName] = tickerValues
				allTickersSorted[tickerPair] = tickerValuesSorted
			}
		}

		analyze(allTickersSorted)

		time.Sleep(1 * time.Minute)
	}
}

func getExchangeNameFromAlias(exchName string) string {
	exchNameLower := strings.ToLower(exchName)
	if exchNameLower == "binance" {
		return "Binance"
	} else if exchNameLower == "gate.io" {
		return "Gate"
	} else if exchNameLower == "bitrue" {
		return "Bitrue"
	} else if exchNameLower == "mexc" {
		return "MEXC"
	} else if exchNameLower == "xt.com" {
		return "XT"
	}
	return ""
}

func getExchanges() map[string]database.Exchange {
	db, err := database.NewDatabase("postgres", "postgres", "postgres")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	exchangesArr, err := db.Queries.SelectExchanges(context.Background())
	if err != nil {
		panic(err)
	}
	exchanges := make(map[string]database.Exchange)
	for _, exchange := range exchangesArr {
		exchanges[exchange.Name] = exchange
	}
	return exchanges
}

func populateDb() {
	exchanges := getExchanges()
	db, err := database.NewDatabase("postgres", "postgres", "postgres")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	entries, err := os.ReadDir("coingecko-data/coins/")
	if err != nil {
		panic(err)
	}

	coinCGIDToName := make(map[string]string)

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		data, err := os.ReadFile("coingecko-data/coins/" + e.Name())
		if err != nil {
			panic(err)
		}

		var coin coingecko.CryptoData
		if err := json.Unmarshal(data, &coin); err != nil {
			panic(err)
		}

		coinCGID := e.Name()[:len(e.Name())-5]
		if _, ok := coinCGIDToName[coinCGID]; ok {
			panic(fmt.Errorf(""))
		}

		coinCGIDToName[coinCGID] = coin.Name
	}

	uniqueExchanges := make(map[string]struct{})
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		data, err := os.ReadFile("coingecko-data/coins/" + e.Name())
		if err != nil {
			panic(err)
		}

		var coin coingecko.CryptoData
		if err := json.Unmarshal(data, &coin); err != nil {
			panic(err)
		}

		var coinUUID uuid.UUID
		exchangesAdded := make(map[uuid.UUID]struct{})
		added := false
		for i, ticker := range coin.Tickers {
			if strings.ToLower(ticker.Target) == "try" {
				continue
			}
			if ticker.TrustScore == "" {
				continue
			}
			exchName := getExchangeNameFromAlias(ticker.Market.Name)
			if exchName == "" {
				continue
			}
			exchangeID := exchanges[exchName].ID
			if _, ok := exchangesAdded[exchangeID]; ok {
				continue
			}
			exchangesAdded[exchangeID] = struct{}{}

			coinCGID := e.Name()[:len(e.Name())-5]

			realBase := ticker.Base
			if coinCGID != ticker.CoinID {
				realBase = ticker.Target
			}

			if !added {
				added = true
				coinUUID, err = db.Queries.InsertCoin(context.Background(), database.InsertCoinParams{
					Name: coin.Name,
					Base: realBase,
				})
				if err != nil {
					panic(err)
				}
			}

			if strings.ToLower(coin.Name) == "tether" && strings.ToLower(realBase) == "usdc" {
				fmt.Println(i)
			}
			if err := db.Queries.InsertCoinExchange(context.Background(), database.InsertCoinExchangeParams{
				CoinID:     coinUUID,
				ExchangeID: exchangeID,
				Name:       coin.Name,
				Base:       realBase,
			}); err != nil {
				panic(err)
			}
		}
	}
	fmt.Println(uniqueExchanges)

	// do it once all coins and exchange_coins were populated, because
	// if we get for example "BTC_USDT", we will need to add BTC, ok
	// But we will need to retrieve the ID of USDT, which would not be
	// added yet.
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		data, err := os.ReadFile("coingecko-data/coins/" + e.Name())
		if err != nil {
			panic(err)
		}

		var coin coingecko.CryptoData
		if err := json.Unmarshal(data, &coin); err != nil {
			panic(err)
		}

		for _, ticker := range coin.Tickers {
			if strings.ToLower(ticker.Target) == "try" {
				continue
			}
			if ticker.TrustScore == "" {
				continue
			}
			exchName := getExchangeNameFromAlias(ticker.Market.Name)
			if exchName == "" {
				continue
			}
			exchangeID := exchanges[exchName].ID

			coinCGID := e.Name()[:len(e.Name())-5]

			realBase := ticker.Base
			realTarget := ticker.Target
			if coinCGID != ticker.CoinID {
				realBase = ticker.Target
				realTarget = ticker.Base
			}

			baseUUID, err := db.Queries.SelectExchangeCoinIDFromBase(context.Background(), database.SelectExchangeCoinIDFromBaseParams{
				Base:       realBase,
				ExchangeID: exchangeID,
			})
			if err != nil {
				fmt.Println(ticker, err)
				continue
			}
			quoteUUID, err := db.Queries.SelectExchangeCoinIDFromBase(context.Background(), database.SelectExchangeCoinIDFromBaseParams{
				Base:       realTarget,
				ExchangeID: exchangeID,
			})
			if err != nil {
				fmt.Println(ticker, err)
				continue
			}

			if err := db.Queries.InsertTicker(context.Background(), database.InsertTickerParams{
				ExchangeID:      exchangeID,
				BaseExchCoinID:  baseUUID,
				QuoteExchCoinID: quoteUUID,
			}); err != nil {
				fmt.Println("insert:", ticker, err)
				continue
			}
		}
	}
}

func getBalance() {
	brokers := loadBrokers()
	for _, broker := range brokers {
		wallet, err := broker.GetBalance(context.Background())
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(broker.GetBrokerName(), wallet)
	}
}

func coingeckoExtract() {
	cg, _ := aggregator.NewCoinGecko(aggregator.Config{
		Key: "...",
	})

	data, err := os.ReadFile("coingecko-data/coingecko-coins.json")
	if err != nil {
		coins, err := cg.GetCoins()
		if err != nil {
			panic(err)
		}

		var coinsArr []coingecko.Coin
		for _, coin := range coins {
			coinsArr = append(coinsArr, coin)
		}

		data, err := json.Marshal(coinsArr)
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile("coingecko-data/coingecko-coins.json", data, 0777); err != nil {
			panic(err)
		}
	}

	var coinsArr []coingecko.Coin
	if err := json.Unmarshal(data, &coinsArr); err != nil {
		panic(err)
	}

	for _, coin := range coinsArr {
		_, err := os.ReadFile("coingecko-data/coins/" + coin.Id + ".json")
		if err == nil {
			continue // if file already exist, skip it
		}
		ticker, err := cg.GetCoinInfo(coin.Id)
		if err != nil {
			fmt.Println(err)
			time.Sleep(55 * time.Second)
			continue
		}
		data, err := json.Marshal(ticker)
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile("coingecko-data/coins/"+coin.Id+".json", data, 0777); err != nil {
			panic(err)
		}
		fmt.Println(coin.Name)
		time.Sleep(12000 * time.Millisecond)
	}
}

func getOrderBook() {
	var taoUUID uuid.UUID
	// hardcoded UUID for poc
	if err := taoUUID.Scan("48a51b13-bed7-462e-9ffc-5e827fab35c0"); err != nil {
		panic(err)
	}

	db, err := database.NewDatabase("postgres", "postgres", "postgres")
	if err != nil {
		panic(err)
	}

	rows, err := db.Queries.SelectExchangeCoinFromCoinID(context.Background(), taoUUID)
	if err != nil {
		panic(err)
	}

	fmt.Println(rows)

	brokers := loadBrokers()

	for _, row := range rows {
		broker, ok := brokers[row.Name]
		if !ok {
			log.Fatalf("oupsie, %s not in the brokers", row.Name)
		}

		orderbook, err := broker.GetOrderBooks(context.Background(), database.SelectExchangeTickersRow{
			Base:  row.Base,
			Quote: "usdt",
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(row.Name, ":", orderbook)
	}
}

func mergeNetworks() {
	type network struct {
		Code          string
		Name          string
		AddressRegexp string
		MemoRegexp    string
	}

	var mexc_networks []network
	var binance_networks []network

	mexc_data, _ := os.ReadFile("blockchain-network-data/networks_mexc.json")
	binance_data, _ := os.ReadFile("blockchain-network-data/networks_binance.json")

	json.Unmarshal(mexc_data, &mexc_networks)
	json.Unmarshal(binance_data, &binance_networks)

	type has struct {
		Binance network
		Mexc    network
	}

	mapHas := make(map[string]*has)

	for _, network := range mexc_networks {
		mapHas[network.Code] = &has{}
	}
	for _, network := range binance_networks {
		mapHas[network.Code] = &has{}
	}

	for _, network := range mexc_networks {
		mapHas[network.Code].Mexc = network
	}
	for _, network := range binance_networks {
		mapHas[network.Code].Binance = network
	}

	data, _ := json.MarshalIndent(mapHas, "", "\t")

	os.WriteFile("networks-compiled.json", data, 0777)
}
