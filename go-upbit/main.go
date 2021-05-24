package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"ipoemi/go-upbit/strategy"
	"ipoemi/go-upbit/upbit"

	"github.com/shopspring/decimal"
)

const BUFFER_SIZE int = 60 * 10

var LOSE_RATE decimal.Decimal = decimal.NewFromFloat(0.05)
var ONE_DECIMAL decimal.Decimal = decimal.NewFromFloat(1.0)

type BuyItem struct {
	MarketTicker upbit.MarketTicker
	BuyPrice     decimal.Decimal
}

func NewBuyItem(m upbit.MarketTicker) *BuyItem {
	return &BuyItem{
		MarketTicker: m,
		BuyPrice:     m.TradePrice,
	}
}

type History map[string][]upbit.MarketTicker
type Wallet struct {
	Amount    decimal.Decimal
	TickerMap map[string]*BuyItem
}

func NewWallet() *Wallet {
	return &Wallet{
		Amount:    decimal.NewFromFloat(0),
		TickerMap: make(map[string]*BuyItem),
	}
}

func CheckError(err error, msg string) {
	if err != nil {
		panic(fmt.Errorf("%s : Cause By\n%s", msg, err))
	}
}

func GetMarketsToWallet(mss [][]upbit.MarketTicker, wallet Wallet, strategies []strategy.BuyStrategy) []upbit.MarketTicker {
	chs := make([]chan *upbit.MarketTicker, 0)
	for _, ms := range mss {
		if len(ms) > BUFFER_SIZE && wallet.TickerMap[ms[0].Market] == nil {
			ch := make(chan *upbit.MarketTicker)
			chs = append(chs, ch)
			go func(ms []upbit.MarketTicker, c chan *upbit.MarketTicker) {
				var result *upbit.MarketTicker = nil
				for _, s := range strategies {
					if s.CheckMarket(ms) {
						result = &ms[len(ms)-1]
						break
					}
				}
				c <- result
			}(ms, ch)
		}
	}
	result := make([]upbit.MarketTicker, 0, BUFFER_SIZE)
	for _, ch := range chs {
		m := <-ch
		if m != nil {
			result = append(result, *m)
		}
	}
	return result
}

func UpdateWallet(wallet Wallet, ms []upbit.MarketTicker) Wallet {
	result := wallet
	for _, v := range ms {
		result.TickerMap[v.Market] = NewBuyItem(v)
	}
	return result
}

func GetValues(m map[string][]upbit.MarketTicker) [][]upbit.MarketTicker {
	result := make([][]upbit.MarketTicker, 0)
	for _, v := range m {
		result = append(result, v)
	}
	return result
}

func UpdateHistory(history History, tickers []upbit.MarketTicker) History {
	result := history
	for _, ticker := range tickers {
		historyOfTicker := result[ticker.Market]
		if historyOfTicker == nil {
			historyOfTicker = make([]upbit.MarketTicker, 0, BUFFER_SIZE)
		}
		if len(historyOfTicker) > BUFFER_SIZE {
			historyOfTicker = historyOfTicker[1:]
		}
		historyOfTicker = append(historyOfTicker, ticker)
		result[ticker.Market] = historyOfTicker
	}
	return result
}

func GetLastMarketMap(history History) map[string]*upbit.MarketTicker {
	result := make(map[string]*upbit.MarketTicker)
	for k, v := range history {
		if len(v) > 0 {
			result[k] = &v[len(v)-1]
		}
	}
	return result
}

func ProcessWallet(wallet Wallet, history History) Wallet {
	result := wallet
	lastMap := GetLastMarketMap(history)
	toSell := make([]BuyItem, 0)
	for k, v := range result.TickerMap {
		checkedPrice := v.MarketTicker.TradePrice
		lastPrice := lastMap[k].TradePrice
		if checkedPrice.Mul(ONE_DECIMAL.Sub(LOSE_RATE)).Cmp(lastPrice) > 0 {
			toSell = append(toSell, *v)
		} else if lastPrice.Cmp(checkedPrice) > 0 {
			v.MarketTicker = *lastMap[k]
		}
	}
	for _, m := range toSell {
		buyingPrice := m.BuyPrice
		lastPrice := lastMap[m.MarketTicker.Market].TradePrice
		now := time.Now()
		result.Amount = result.Amount.Add(lastPrice.Sub(buyingPrice))
		fmt.Printf("[%v] Sell / %s, sell: %v, buy: %v, wallet: %v\n", now, m.MarketTicker.Market, lastPrice, buyingPrice, result.Amount)
		delete(result.TickerMap, m.MarketTicker.Market)
	}
	return result
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
	fmt.Fprintf(os.Stderr, "start go-upbit\n")
	ctx := context.Background()
	tickerStream := GetTickerStream(ctx)

	strategies := make([]strategy.BuyStrategy, 0)
	strategies = append(strategies, &strategy.FivePercentDecStrategy{})

	history := make(History)
	wallet := *NewWallet()
	for {
		tickers := <-tickerStream
		history = UpdateHistory(history, tickers)
		targets := GetMarketsToWallet(GetValues(history), wallet, strategies)
		wallet = UpdateWallet(wallet, targets)
		wallet = ProcessWallet(wallet, history)
	}
}
