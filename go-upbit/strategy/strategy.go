package strategy

import (
	"fmt"
	"ipoemi/go-upbit/upbit"
	"time"

	"github.com/shopspring/decimal"
)

var DETECTED_RATE decimal.Decimal = decimal.NewFromFloat(0.05)

type BuyStrategy interface {
	CheckMarket(ms []upbit.MarketTicker) bool
}

type FivePercentDecStrategy struct{}

func (s *FivePercentDecStrategy) CheckMarket(ms []upbit.MarketTicker) bool {
	last := ms[len(ms)-1]
	min := ms[0]
	for _, v := range ms {
		if v.TradePrice.Cmp(min.TradePrice) < 0 {
			min = v
		}
	}
	lastPrice := last.TradePrice
	minPrice := min.TradePrice
	if minPrice.Mul(DETECTED_RATE).Cmp(lastPrice.Sub(minPrice)) < 0 {
		now := time.Now()
		fmt.Printf("[%v] FivePercetDecStrategy - Buy / %s, last: %v, min: %v\n", now, last.Market, lastPrice, minPrice)
		return true
	} else {
		return false
	}
}
