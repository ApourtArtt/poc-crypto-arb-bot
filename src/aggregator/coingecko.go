package aggregator

import "github.com/ArbitrageCoin/crypto-sdk/src/third_parties/coingecko"

type Config struct {
	Key string
}

type CoinGecko struct {
	config Config
}

func NewCoinGecko(config Config) (*CoinGecko, error) {
	return &CoinGecko{
		config: config,
	}, nil
}

func (a CoinGecko) GetCoins() (map[string]coingecko.Coin, error) {
	coins, err := coingecko.GetCoinsList(a.config.Key)
	if err != nil {
		return nil, err
	}

	out := make(map[string]coingecko.Coin)
	for _, coin := range coins {
		out[coin.Id] = coin
	}

	return out, nil
}

func (a CoinGecko) GetCoinInfo(coin string) (coingecko.CryptoData, error) {
	ticker, err := coingecko.GetCoinTickers(a.config.Key, coin)
	if err != nil {
		return coingecko.CryptoData{}, err
	}

	return ticker, nil
}
