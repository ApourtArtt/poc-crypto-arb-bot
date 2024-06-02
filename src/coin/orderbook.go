package coin

import (
	"sort"

	"github.com/shopspring/decimal"
)

type OrderBook struct {
	Bids []Offer
	Asks []Offer
}

func (ob *OrderBook) Equals(other *OrderBook) bool {
	// Compare Bids
	if len(ob.Bids) != len(other.Bids) {
		return false
	}
	for i := range ob.Bids {
		if ob.Bids[i] != other.Bids[i] {
			return false
		}
	}

	// Compare Asks
	if len(ob.Asks) != len(other.Asks) {
		return false
	}
	for i := range ob.Asks {
		if ob.Asks[i] != other.Asks[i] {
			return false
		}
	}

	return true
}

func (ob *OrderBook) SortBids() {
	sort.Slice(ob.Bids, func(i, j int) bool {
		return ob.Bids[i].Price.GreaterThan(ob.Bids[j].Price)
	})
}

func (ob *OrderBook) SortAsks() {
	sort.Slice(ob.Asks, func(i, j int) bool {
		return ob.Asks[i].Price.LessThan(ob.Asks[j].Price)
	})
}

type Offer struct {
	Price    decimal.Decimal
	Quantity decimal.Decimal
}

func (o Offer) TotalPrice() decimal.Decimal {
	return o.Price.Mul(o.Quantity)
}
