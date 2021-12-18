package candle

import (
	"ipoemi/candle-loader/upbit"
	"time"

	"github.com/pkg/errors"
)

type Candle interface {
	GetMarketId() string
	GetMarketName() string
	GetTradeTimestamp() int64
	GetOpeningPrice() float64
	GetHighPrice() float64
	GetLowPrice() float64
	GetTradePrice() float64
}

type MinuteCandle struct {
	MarketId       string  `json:"market_id" bson:"market_id"`
	Unit           int     `json:"unit" bson:"unit"`
	MarketName     string  `json:"market_name" bson:"market_name"`
	TradeTimestamp int64   `json:"trade_timestamp" bson:"trade_timestamp"`
	OpeningPrice   float64 `json:"opening_price" bson:"opening_price"`
	HighPrice      float64 `json:"high_price" bson:"high_price"`
	LowPrice       float64 `json:"low_price" bson:"low_price"`
	TradePrice     float64 `json:"trade_price" bson:"trade_price"`
}

func (c MinuteCandle) GetMarketId() string      { return c.MarketId }
func (c MinuteCandle) GetMarketName() string    { return c.MarketName }
func (c MinuteCandle) GetTradeTimestamp() int64 { return c.TradeTimestamp }
func (c MinuteCandle) GetOpeningPrice() float64 { return c.OpeningPrice }
func (c MinuteCandle) GetHighPrice() float64    { return c.HighPrice }
func (c MinuteCandle) GetLowPrice() float64     { return c.LowPrice }
func (c MinuteCandle) GetTradePrice() float64   { return c.TradePrice }

type DayCandle struct {
	MarketId       string  `json:"market_id" bson:"market_id"`
	MarketName     string  `json:"market_name" bson:"market_name"`
	TradeTimestamp int64   `json:"trade_timestamp" bson:"trade_timestamp"`
	OpeningPrice   float64 `json:"opening_price" bson:"opening_price"`
	HighPrice      float64 `json:"high_price" bson:"high_price"`
	LowPrice       float64 `json:"low_price" bson:"low_price"`
	TradePrice     float64 `json:"trade_price" bson:"trade_price"`
}

func (c DayCandle) GetMarketId() string      { return c.MarketId }
func (c DayCandle) GetMarketName() string    { return c.MarketName }
func (c DayCandle) GetTradeTimestamp() int64 { return c.TradeTimestamp }
func (c DayCandle) GetOpeningPrice() float64 { return c.OpeningPrice }
func (c DayCandle) GetHighPrice() float64    { return c.HighPrice }
func (c DayCandle) GetLowPrice() float64     { return c.LowPrice }
func (c DayCandle) GetTradePrice() float64   { return c.TradePrice }

type TimeParseError interface {
	error
}

func NewFromUpbitMinuteCandle(c upbit.MarketCandle) (*MinuteCandle, error) {
	t, err := time.Parse("2006-01-02T15:04:05", c.CandleDateTimeUtc)
	if err != nil {
		return nil, TimeParseError(errors.Wrapf(err, "Invalid time format: %v", c.CandleDateTimeUtc))
	}
	openPrice, _ := c.OpeningPrice.Float64()
	highPrice, _ := c.HighPrice.Float64()
	lowPrice, _ := c.LowPrice.Float64()
	tradePrice, _ := c.TradePrice.Float64()
	return &MinuteCandle{
		MarketId:       c.MarketName,
		MarketName:     c.MarketName,
		TradeTimestamp: t.Unix(),
		OpeningPrice:   openPrice,
		HighPrice:      highPrice,
		LowPrice:       lowPrice,
		TradePrice:     tradePrice,
		Unit:           *c.Unit,
	}, nil
}

func NewFromUpbitDayCandle(c upbit.MarketCandle) (*DayCandle, error) {
	t, err := time.Parse("2006-01-02T15:04:05", c.CandleDateTimeUtc)
	if err != nil {
		return nil, TimeParseError(errors.Wrapf(err, "Invalid time format: %v", c.CandleDateTimeUtc))
	}
	openPrice, _ := c.OpeningPrice.Float64()
	highPrice, _ := c.HighPrice.Float64()
	lowPrice, _ := c.LowPrice.Float64()
	tradePrice, _ := c.TradePrice.Float64()
	return &DayCandle{
		MarketId:       c.MarketName,
		MarketName:     c.MarketName,
		TradeTimestamp: t.Unix(),
		OpeningPrice:   openPrice,
		HighPrice:      highPrice,
		LowPrice:       lowPrice,
		TradePrice:     tradePrice,
	}, nil
}
