package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"ipoemi/go-upbit/trade"
	"ipoemi/go-upbit/upbit"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/shopspring/decimal"
)

const BUFFER_SIZE int = 60 * 3

var LOSE_RATE = decimal.NewFromFloat(0.05)
var ONE_DECIMAL = decimal.NewFromFloat(1.0)
var BUY_UNIT = decimal.NewFromFloat(100_000)
var WALLET_START = decimal.NewFromFloat(1_000_000)

var CurrencyPrinter = message.NewPrinter(language.English)

type History map[string][]upbit.MarketTicker

func (h History) Summary() string {
	var maxTs int64 = 0
	for _, v := range h {
		if len(v) > 0 {
			last := v[len(v)-1]
			if last.Timestamp > maxTs {
				maxTs = last.Timestamp
			}
		}
	}
	tm := time.Unix(maxTs/1000, 0)
	return fmt.Sprintf("%d MarketTicker, Last Update: %s", len(h), tm)
}

func (h History) AllSummary() string {
	resultBuilder := make([]string, 0)
	for _, v := range h {
		if len(v) > 0 {
			last := v[len(v)-1]
			tm := time.Unix(last.Timestamp/1000, 0).Format(time.RFC3339)
			resultBuilder = append(resultBuilder, fmt.Sprintf("%v: %v (%s)", last.Market, last.TradePrice, tm))
		}
	}
	return strings.Join(resultBuilder, "\n")
}

type MinMax map[string]*struct {
	Min upbit.MarketTicker
	Max upbit.MarketTicker
}

func (h *History) GetLastMarketTickers() []upbit.MarketTicker {
	result := make([]upbit.MarketTicker, 0)
	for _, v := range *h {
		if len(v) > 0 {
			result = append(result, v[len(v)-1])
		}
	}
	return result
}

func GetMarketsToBuy(mss [][]upbit.MarketTicker, wallet *trade.Wallet, strategies []trade.BuyStrategy) []upbit.MarketTicker {
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

func SendBuyMessagesToWallet(wallet *trade.Wallet, history History, ms []upbit.MarketTicker) {
	for _, v := range ms {
		if wallet.Amount.Cmp(BUY_UNIT) >= 0 {
			<-wallet.Buy(v, BUY_UNIT.Div(v.TradePrice))
		}
	}
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

func ToMap(ms []upbit.MarketTicker) map[string]*upbit.MarketTicker {
	result := make(map[string]*upbit.MarketTicker)
	for _, m := range ms {
		m1 := m
		result[m.Market] = &m1
	}
	return result
}

func GetMarketTickerToSell(wallet trade.Wallet, history History, minMax MinMax) []upbit.MarketTicker {
	result := make([]upbit.MarketTicker, 0)
	lastMarketMap := ToMap(history.GetLastMarketTickers())
	for k := range wallet.TickerMap {
		maxPrice := minMax[k].Max.TradePrice
		lastPrice := lastMarketMap[k].TradePrice
		if maxPrice.Mul(ONE_DECIMAL.Sub(LOSE_RATE)).Cmp(lastPrice) > 0 {
			result = append(result, *lastMarketMap[k])
		}
	}
	return result
}

func UpdateMinMaxForBuying(minMax *MinMax, wallet trade.Wallet, history History) {
	lastMarketTickerMap := ToMap(history.GetLastMarketTickers())
	for k := range wallet.TickerMap {
		if (*minMax)[k] == nil {
			(*minMax)[k] = &struct {
				Min upbit.MarketTicker
				Max upbit.MarketTicker
			}{
				Min: *lastMarketTickerMap[k],
				Max: *lastMarketTickerMap[k],
			}
		} else {
			item := (*minMax)[k]
			maxPrice := item.Max.TradePrice
			minPrice := item.Min.TradePrice
			lastPrice := lastMarketTickerMap[k].TradePrice
			if lastPrice.Cmp(maxPrice) > 0 {
				(*minMax)[k].Max = *lastMarketTickerMap[k]
			}
			if lastPrice.Cmp(minPrice) < 0 {
				(*minMax)[k].Min = *lastMarketTickerMap[k]
			}
		}
	}
}

func UpdateMinMaxForSelling(minMax *MinMax, sellMarketTickers []upbit.MarketTicker) {
	for _, v := range sellMarketTickers {
		if (*minMax)[v.Market] != nil {
			delete((*minMax), v.Market)
		}
	}
}

func SendSellMessageToWallet(wallet *trade.Wallet, ms []upbit.MarketTicker) {
	for _, m := range ms {
		<-wallet.Sell(m)
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
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					time.Sleep(1 * time.Second)
					continue
				}
				marketIDs := make([]string, 0)
				for _, m := range list {
					if strings.HasPrefix(m.Market, "KRW") {
						marketIDs = append(marketIDs, m.Market)
					}
				}
				list2, err := upbit.GetMarketTickerList(marketIDs)
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					time.Sleep(1 * time.Second)
					continue
				}
				result <- list2
			case <-ctx.Done():
				end = true
			}
		}
	}()
	return result
}

func GetValues(m map[string][]upbit.MarketTicker) [][]upbit.MarketTicker {
	result := make([][]upbit.MarketTicker, 0)
	for _, v := range m {
		result = append(result, v)
	}
	return result
}

func Run(ctx context.Context, wg *sync.WaitGroup, wallet *trade.Wallet, history *History, minMax *MinMax) {
	defer wg.Done()
	tickerStream := GetTickerStream(ctx)

	strategies := make([]trade.BuyStrategy, 0)
	strategies = append(strategies, &trade.FivePercentDecStrategy{})
	wallet.ProcessMessage(ctx)

	for tickers := range tickerStream {
		*history = UpdateHistory(*history, tickers)
		marketsToBuy := GetMarketsToBuy(GetValues(*history), wallet, strategies)
		SendBuyMessagesToWallet(wallet, *history, marketsToBuy)
		UpdateMinMaxForBuying(minMax, *wallet, *history)
		marketsToSell := GetMarketTickerToSell(*wallet, *history, *minMax)
		SendSellMessageToWallet(wallet, marketsToSell)
		UpdateMinMaxForSelling(minMax, marketsToSell)
	}
}

func ShowHelp() {
	cmds := []string{
		" h (history): show history summary",
		" ah (allhistory): show all history",
		" w (wallet): show wallet",
		" ?: show all commands",
	}
	fmt.Println(strings.Join(cmds, "\n"))
}

func main() {
	fmt.Fprintf(os.Stderr, "start go-upbit\n")
	ctx := context.Background()
	wallet := trade.NewWallet(WALLET_START)
	history := make(History)
	minMax := make(MinMax)
	wg := new(sync.WaitGroup)
	go Run(ctx, wg, wallet, &history, &minMax)

	for {
		reader := bufio.NewReader(os.Stdin)
		println()
		fmt.Print("Enter CMD (Help: ?): ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.Replace(cmd, "\n", "", -1)

		switch cmd {
		case "w", "wallet":
			println()
			lastMarketMap := ToMap(history.GetLastMarketTickers())
			fmt.Println(wallet.Summary(lastMarketMap))
		case "h", "history":
			println()
			fmt.Println(history.Summary())
		case "ah", "allhistory":
			println()
			fmt.Println(history.AllSummary())
		case "?":
			println()
			ShowHelp()
		case "":
		default:
			println()
			CurrencyPrinter.Printf("'%s' is Wrong Command\n", cmd)
			println()
			ShowHelp()
		}
	}
}
