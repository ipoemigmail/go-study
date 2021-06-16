package strategy

import (
	"ipoemi/go-upbit/trade"
	"ipoemi/go-upbit/upbit"

	"github.com/shopspring/decimal"
)

var DEFAULT_SELL_RATE_LOSE decimal.Decimal = decimal.NewFromFloat(0.95)
var DEFAULT_SELL_RATE_WIN decimal.Decimal = decimal.NewFromFloat(1.025)

type SellStrategy interface {
	CheckMarket(marketName string) bool
}

type DefaultSellStrategy struct {
	history map[string][]upbit.MarketTicker
	wallet  trade.Wallet
}

func NewDefaultSellStrategy(history map[string][]upbit.MarketTicker, wallet trade.Wallet) *DefaultSellStrategy {
	return &DefaultSellStrategy{
		history,
		wallet,
	}
}

func (s *DefaultSellStrategy) CheckMarket(marketName string) bool {
	historySize := len(s.history[marketName])
	var result bool
	if historySize > 0 {
		last := s.history[marketName][historySize-1]
		if s.wallet.TickerMap[last.MarketName] != nil {
			buyItem := s.wallet.TickerMap[last.MarketName].MarketTicker
			if last.TradePrice.Cmp(buyItem.TradePrice.Mul(DEFAULT_SELL_RATE_LOSE)) < 0 {
				result = true
			} else if last.TradePrice.Cmp(buyItem.TradePrice.Mul(DEFAULT_SELL_RATE_WIN)) >= 0 {
				result = false
			}
		}
	}
	return result
}

var RSI_SELL_THRESHOLD float64 = 0.7

type RsiSellStrategy struct {
	history       map[string][]upbit.MarketTicker
	wallet        trade.Wallet
	minuteCandles map[string][]upbit.MarketCandle
}

func NewRsiSellStrategy(history map[string][]upbit.MarketTicker, wallet trade.Wallet, minuteCandles map[string][]upbit.MarketCandle) *RsiSellStrategy {
	return &RsiSellStrategy{
		history,
		wallet,
		minuteCandles,
	}
}

func (s *RsiSellStrategy) CheckMarket(marketName string) bool {
	result := false
	if s.wallet.TickerMap[marketName] != nil {
		candles := s.minuteCandles[marketName]
		candlePrices := make([]decimal.Decimal, len(candles))
		for _, candle := range candles {
			candlePrices = append(candlePrices, candle.TradePrice)
		}
		prevRsi := trade.GetRsi(candlePrices[:])
		lastPrice := s.history[marketName][len(s.history[marketName])-1].TradePrice
		curRsi := trade.GetRsi(append(candlePrices[1:], lastPrice))
		if prevRsi > RSI_SELL_THRESHOLD && curRsi <= RSI_SELL_THRESHOLD {
			result = true
		}
	}
	return result
}
