package main

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"

	"ipoemi/go-upbit/upbit"
)

const BUFFER_SIZE int = 60 * 10

func CheckError(err error, msg string) {
	if err != nil {
		panic(fmt.Errorf("%s : Cause By\n%s", msg, err))
	}
}

func CheckMarket(ms []upbit.MarketTicker) {
	if len(ms) >= BUFFER_SIZE {
		last := ms[len(ms)-1]
		sort.Slice(ms, func(i, j int) bool {
			p1 := big.Float(ms[i].TradePrice)
			p2 := big.Float(ms[j].TradePrice)
			return p1.Cmp(&p2) < 0
		})
		min := ms[0]
		lastPrice := big.Float(last.TradePrice)
		minPrice := big.Float(min.TradePrice)
		diff := big.NewFloat(0)
		thresholdPrice := big.NewFloat(0)
		thresholdPrice.Mul(&lastPrice, big.NewFloat(0.1))
		diff.Sub(&lastPrice, &minPrice)
		if thresholdPrice.Cmp(diff) < 0 {
			fmt.Printf("%s, last: %v, min: %v\n", last.Market, lastPrice, minPrice)
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
