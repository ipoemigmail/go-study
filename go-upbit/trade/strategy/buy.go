package strategy

import (
	"fmt"
	"ipoemi/go-upbit/trade"
	"ipoemi/go-upbit/upbit"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var DEFAULT_BUY_RATE decimal.Decimal = decimal.NewFromFloat(0.025)

type BuyStrategy interface {
	CheckMarket(marketName string) bool
}

type DefaultBuyStrategy struct {
	history map[string][]upbit.MarketTicker
	wallet  trade.Wallet
}

func NewDefaultBuyStrategy(history map[string][]upbit.MarketTicker, wallet trade.Wallet) *DefaultBuyStrategy {
	return &DefaultBuyStrategy{
		history,
		wallet,
	}
}

func (s *DefaultBuyStrategy) CheckMarket(marketName string) bool {
	historySize := len(s.history[marketName])
	last := s.history[marketName][historySize-1]
	result := false
	if s.wallet.TickerMap[last.MarketName] == nil {
		min := last
		last1MinuteHistory := s.history[marketName][historySize-60 : historySize]
		for _, v := range last1MinuteHistory {
			if v.TradePrice.Cmp(min.TradePrice) < 0 {
				min = v
			}
		}
		lastPrice := last.TradePrice
		minPrice := min.TradePrice
		if minPrice.Mul(DEFAULT_BUY_RATE).Cmp(lastPrice.Sub(minPrice)) > 0 {
			result = true
		}
	}
	return result
}

var RSI_BUY_THRESHOLD float64 = 0.3

type RsiBuyStrategy struct {
	history       map[string][]upbit.MarketTicker
	wallet        trade.Wallet
	minuteCandles map[string][]upbit.MarketCandle
}

func NewRsiBuyStrategy(history map[string][]upbit.MarketTicker, wallet trade.Wallet, minuteCandles map[string][]upbit.MarketCandle) *RsiBuyStrategy {
	return &RsiBuyStrategy{
		history,
		wallet,
		minuteCandles,
	}
}

func (s *RsiBuyStrategy) CheckMarket(marketName string) bool {
	result := false
	if s.wallet.TickerMap[marketName] == nil && len(s.minuteCandles[marketName]) > 0 {
		candles := s.minuteCandles[marketName]
		candlePrices := make([]decimal.Decimal, len(candles))
		for i := range candles {
			candlePrices[i] = candles[i].TradePrice
		}
		lastPrice := s.history[marketName][len(s.history[marketName])-1].TradePrice
		if strings.HasPrefix(time.Now().UTC().Format(time.RFC3339), candles[len(candles)-1].CandleDateTimeUtc) {
			candlePrices[len(candles)-1] = lastPrice
		} else {
			candlePrices = append(candlePrices[1:], lastPrice)
		}
		prevRsi := trade.GetRsi(candlePrices[:len(candles)-1])
		curRsi := trade.GetRsi(candlePrices[1:])
		if prevRsi < RSI_BUY_THRESHOLD && curRsi >= RSI_BUY_THRESHOLD {
			fmt.Printf("%v: %f => %f\n", marketName, prevRsi, curRsi)
			result = true
		}
	}
	return result
}
