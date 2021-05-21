package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"ipoemi/go-upbit/upbit"

	"github.com/shopspring/decimal"
)

const BUFFER_SIZE int = 60 * 10
const DETECTED_RATE float64 = 0.05

func CheckError(err error, msg string) {
	if err != nil {
		panic(fmt.Errorf("%s : Cause By\n%s", msg, err))
	}
}

func CheckMarket(ms []upbit.MarketTicker) {
	if len(ms) >= BUFFER_SIZE {
		last := ms[len(ms)-1]
		sort.Slice(ms, func(i, j int) bool {
			p1 := ms[i].TradePrice
			p2 := ms[j].TradePrice
			return p1.Cmp(p2) < 0
		})
		min := ms[0]
		lastPrice := last.TradePrice
		minPrice := min.TradePrice
		now := time.Now()
		if lastPrice.Mul(decimal.NewFromFloat(DETECTED_RATE)).Cmp(lastPrice.Sub(minPrice)) < 0 {
			fmt.Printf("[%v] %s, last: %v, min: %v\n", now, last.Market, lastPrice, minPrice)
		}
	}
}

func GetTickerStream(ctx context.Context) chan []upbit.MarketTicker {
	result := make(chan []upbit.MarketTicker)
	ticker := time.NewTicker(time.Second * 1)
	go func() {
		end := false
		for !end {
			select {
			case <-ticker.C:
				list, err := upbit.GetMarketList()
				CheckError(err, "Application Error")
				marketIDs := make([]string, 0)
				for _, m := range list {
					//fmt.Println(m)
					if strings.HasPrefix(m.Market, "KRW") {
						marketIDs = append(marketIDs, m.Market)
					}
				}
				list2, err := upbit.GetMarketTickerList(marketIDs)
				CheckError(err, "Application Error")
				result <- list2
			case <-ctx.Done():
				end = true
			}
		}
	}()
	return result
}

func main() {
	//ctx, _ := context.WithCancel(context.Background())
	ctx := context.Background()
	tickerStream := GetTickerStream(ctx)
	buffer := make(map[string][]upbit.MarketTicker)
	for {
		tickers := <-tickerStream
		wg := new(sync.WaitGroup)
		for _, ticker := range tickers {
			b := buffer[ticker.Market]
			if b == nil {
				b = make([]upbit.MarketTicker, 0, BUFFER_SIZE)
			}
			if len(b) > BUFFER_SIZE {
				b = b[1:]
			}
			b = append(b, ticker)
			buffer[ticker.Market] = b
			wg.Add(1)
			go func() {
				CheckMarket(b)
				defer wg.Done()
			}()
		}
		wg.Wait()
	}
}
