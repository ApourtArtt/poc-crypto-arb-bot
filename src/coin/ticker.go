package coin

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TickerPair struct {
	Base  uuid.UUID
	Quote uuid.UUID
}

// ASKS
// PRICE	QUANTITY
// 1.02	5
// 1.01	5
// 1.00	5 <- LOWEST ASK
//
// BIDS
// PRICE	QUANTITY
// 0.97	5 <- HIGHEST BID
// 0.96	5
// 0.95	5
type TickerValues struct {
	HighestBid decimal.Decimal
	LowestAsk  decimal.Decimal
}

type TickerStatus struct {
	IsEnabled            bool
	IsLimitOrderAllowed  bool
	IsSpotTradingAllowed bool
	IsBuyable            bool
	IsSellable           bool
}

func (ts TickerStatus) CanTrade() bool {
	return ts.IsEnabled && ts.IsLimitOrderAllowed && ts.IsSpotTradingAllowed
}

func (ts TickerStatus) CanBeBought() bool {
	return ts.CanTrade() && ts.IsBuyable
}

func (ts TickerStatus) CanBeSold() bool {
	return ts.CanTrade() && ts.IsSellable
}
