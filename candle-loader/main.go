package main

import (
	"context"
	"fmt"
	"ipoemi/candle-loader/trade/candle"
	"ipoemi/candle-loader/upbit"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var CurrencyPrinter = message.NewPrinter(language.Korean)

func GetLastMinuteCandle(ctx context.Context, maretId string, unit int, coll *mongo.Collection) (candle.MinuteCandle, error) {
	var lastCandle candle.MinuteCandle
	result, err := coll.Find(ctx, bson.M{"market_id": maretId, "unit": unit}, options.Find().SetSort(bson.M{"trade_timestamp": -1}).SetLimit(1))
	if err != nil {
		return lastCandle, errors.Wrap(err, "GetMinuteLastData")
	}
	defer result.Close(ctx)
	if result.Next(ctx) {
		result.Decode(&lastCandle)
	}
	return lastCandle, nil
}

func GetLastDayCandle(ctx context.Context, maretId string, coll *mongo.Collection) (candle.DayCandle, error) {
	var lastCandle candle.DayCandle
	result, err := coll.Find(ctx, bson.M{"market_id": maretId}, options.Find().SetSort(bson.M{"trade_timestamp": -1}).SetLimit(1))
	if err != nil {
		return lastCandle, errors.Wrap(err, "GetDayLastData")
	}
	defer result.Close(ctx)
	if result.Next(ctx) {
		result.Decode(&lastCandle)
	}
	return lastCandle, nil
}

var A float64 = 5_000_000

func Calc(ctx context.Context, ch chan candle.Candle) (earn float64, min float64, max float64) {
	var prevPrev candle.Candle = nil
	var prev candle.Candle = nil
	for cur := range ch {
		if prevPrev != nil && prev != nil {
			std := ((prevPrev.GetHighPrice() - prevPrev.GetLowPrice()) * 0.5) + prev.GetOpeningPrice()
			if prev.GetHighPrice() > std {
				CurrencyPrinter.Printf(
					"[%v] high: %.2f, std: %.2f, opening: %.2f, earn: %.2f\n",
					time.Unix(cur.GetTradeTimestamp(), 0).Format(time.RFC3339),
					prev.GetHighPrice(),
					std,
					cur.GetOpeningPrice(),
					cur.GetOpeningPrice()-std,
				)
				factor := A / std
				feeFactor := 0.0005
				buyAmount := A
				saleAmount := cur.GetOpeningPrice() * factor
				fee := (buyAmount + saleAmount) * feeFactor
				earn += (saleAmount - buyAmount) - fee
				if earn < min {
					min = earn
				}
				if earn > max {
					max = earn
				}
			}
		}
		prevPrev = prev
		tmp := cur
		prev = tmp
	}
	return
}

func LoadMinuteCandles(ctx context.Context, target string, unit int, collection *mongo.Collection) {
	lastCandle, err := GetLastMinuteCandle(ctx, target, unit, collection)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", lastCandle)
	t := time.Now()
	for {
		candles, err := upbit.GetMarketMinuteCandleList(ctx, target, unit, &t, 200)
		if err != nil {
			panic(err)
		}
		tradeCandles := make([]candle.MinuteCandle, 0)
		for _, c := range candles {
			c1, err := candle.NewFromUpbitMinuteCandle(c)
			if err != nil {
				panic(err)
			}
			tradeCandles = append(tradeCandles, *c1)
		}
		for _, c := range tradeCandles {
			_, err := collection.ReplaceOne(ctx, bson.M{
				"market_id":       c.MarketId,
				"trade_timestamp": c.TradeTimestamp,
			}, c, options.Replace().SetUpsert(true))
			if err != nil {
				panic(err)
			}
		}
		now := time.Now()
		if len(tradeCandles) < 200 || tradeCandles[0].TradeTimestamp <= lastCandle.TradeTimestamp {
			fmt.Printf("[%v] %v market done \n", now.Format(time.RFC3339), target)
			break
		} else {
			t = time.Unix(tradeCandles[0].TradeTimestamp, 0)
			fmt.Printf("[%v] %v market %v updated, until %v \n", now.Format(time.RFC3339), target, len(tradeCandles), t.Format(time.RFC3339))
		}
	}
}

func LoadDayCandles(ctx context.Context, target string, collection *mongo.Collection) {
	lastCandle, err := GetLastDayCandle(ctx, target, collection)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", lastCandle)
	t := time.Now()
	for {
		candles, err := upbit.GetMarketDayCandleList(ctx, target, &t, 200)
		if err != nil {
			panic(err)
		}
		tradeCandles := make([]candle.DayCandle, 0)
		for _, c := range candles {
			c1, err := candle.NewFromUpbitDayCandle(c)
			if err != nil {
				panic(err)
			}
			tradeCandles = append(tradeCandles, *c1)
		}
		for _, c := range tradeCandles {
			_, err := collection.ReplaceOne(ctx, bson.M{
				"market_id":       c.MarketId,
				"trade_timestamp": c.TradeTimestamp,
			}, c, options.Replace().SetUpsert(true))
			if err != nil {
				panic(err)
			}
		}
		now := time.Now()
		if len(tradeCandles) < 200 || tradeCandles[0].TradeTimestamp <= lastCandle.TradeTimestamp {
			fmt.Printf("[%v] %v market done \n", now.Format(time.RFC3339), target)
			break
		} else {
			t = time.Unix(tradeCandles[0].TradeTimestamp, 0)
			fmt.Printf("[%v] %v market %v updated, until %v \n", now.Format(time.RFC3339), target, len(tradeCandles), t.Format(time.RFC3339))
		}
	}
}

func Take(ch chan candle.Candle, num int) chan candle.Candle {
	result := make(chan candle.Candle)
	go func() {
		defer close(result)
		i := 0
		for c := range ch {
			if i < num {
				result <- c
				i++
			} else {
				break
			}
		}
	}()
	return result
}

func DropWhile(ch chan candle.Candle, f func(c candle.Candle) bool) chan candle.Candle {
	result := make(chan candle.Candle)
	go func() {
		defer close(result)
		ok := false
		for c := range ch {
			if !f(c) {
				ok = true
			}
			if ok {
				result <- c
			}
		}
	}()
	return result
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://admin:admin@localhost:27017"))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)
	//collection := client.Database("upbit").Collection("candles_minutely")
	//target := "KRW-ETH"
	//unit := 240
	//LoadMinuteCandles(ctx, target, unit, collection)
	//all, err := collection.Find(ctx, bson.M{"market_id": target, "unit": unit})
	collection := client.Database("upbit").Collection("candles_daily")
	target := "KRW-ETH"
	LoadDayCandles(ctx, target, collection)
	all, err := collection.Find(ctx, bson.M{"market_id": target})
	if err != nil {
		panic(err)
	}
	ch := make(chan candle.Candle)
	go func() {
		defer close(ch)
		var cur candle.DayCandle
		for all.Next(ctx) {
			err = all.Decode(&cur)
			if err != nil {
				panic(err)
			}
			ch <- cur
		}
	}()
	tt, _ := time.Parse(time.RFC3339, "2021-06-01T00:00:00+09:00")
	fmt.Println(tt)
	ts := tt.Unix()
	earn, min, max := Calc(ctx, DropWhile(ch, func(c candle.Candle) bool {
		return c.GetTradeTimestamp() < ts
	}))
	CurrencyPrinter.Printf("total earn: %.2f, min: %.2f, max: %.2f\n", earn, min, max)
}
