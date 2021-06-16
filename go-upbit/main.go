package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"ipoemi/go-upbit/trade"
	"ipoemi/go-upbit/trade/strategy"
	"ipoemi/go-upbit/upbit"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/shopspring/decimal"
)

const BUFFER_SIZE int = 60 * 1

var LOSE_RATE = decimal.NewFromFloat(0.025)
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
			resultBuilder = append(resultBuilder, fmt.Sprintf("%v: %v (%s)", last.MarketName, last.TradePrice, tm))
		}
	}
	return strings.Join(resultBuilder, "\n")
}

type MinuteCandles map[string][]upbit.MarketCandle

func (m MinuteCandles) Summary() string {
	resultBuilder := make([]string, 0)
	for _, v := range m {
		if len(v) > 0 {
			last := v[len(v)-1]
			tm := time.Unix(last.Timestamp/1000, 0).Format(time.RFC3339)
			resultBuilder = append(resultBuilder, fmt.Sprintf("%v: %v (%s)", last.MarketName, last.TradePrice, tm))
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

func GetValues(m map[string][]upbit.MarketTicker) [][]upbit.MarketTicker {
	result := make([][]upbit.MarketTicker, 0)
	for _, v := range m {
		result = append(result, v)
	}
	return result
}

func ToMap(ms []upbit.MarketTicker) map[string]*upbit.MarketTicker {
	result := make(map[string]*upbit.MarketTicker)
	for _, m := range ms {
		m1 := m
		result[m.MarketName] = &m1
	}
	return result
}

func GetLast(tickers []upbit.MarketTicker) *upbit.MarketTicker {
	if len(tickers) > 0 {
		return &tickers[len(tickers)-1]
	} else {
		return nil
	}
}

func GetMarketsToBuy(history History, wallet *trade.Wallet, strategies []strategy.BuyStrategy) []upbit.MarketTicker {
	chs := make([]chan *upbit.MarketTicker, 0)
	result := make([]upbit.MarketTicker, 0, len(history))
	if wallet.Amount.Cmp(BUY_UNIT) < 0 {
		return result
	}
	for marketName := range history {
		if len(history[marketName]) > 2 {
			ch := make(chan *upbit.MarketTicker)
			chs = append(chs, ch)
			go func(marketName string, c chan *upbit.MarketTicker) {
				defer close(c)
				var result *upbit.MarketTicker = nil
				for _, s := range strategies {
					if s.CheckMarket(marketName) {
						result = GetLast(history[marketName])
						break
					}
				}
				c <- result
			}(marketName, ch)
		}
	}
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
			fTradePrice, _ := v.TradePrice.Float64()
			now := time.Now().Format(time.RFC3339)
			CurrencyPrinter.Printf("[%s] Buy / %s, Last: %f\n", now, v.MarketName, fTradePrice)
			<-wallet.Buy(v, BUY_UNIT.Div(v.TradePrice))
		} else {
			break
		}
	}
}

func UpdateHistory(history History, tickers []upbit.MarketTicker) History {
	result := history
	for _, ticker := range tickers {
		historyOfTicker := result[ticker.MarketName]
		if historyOfTicker == nil {
			historyOfTicker = make([]upbit.MarketTicker, 0, BUFFER_SIZE)
		}
		if len(historyOfTicker) > BUFFER_SIZE {
			historyOfTicker = historyOfTicker[1:]
		}
		historyOfTicker = append(historyOfTicker, ticker)
		result[ticker.MarketName] = historyOfTicker
	}
	return result
}

func GetMarketTickerToSell(wallet trade.Wallet, history History, minMax MinMax, strategies []strategy.SellStrategy) []upbit.MarketTicker {
	chs := make([]chan *upbit.MarketTicker, 0)
	for marketName := range history {
		ch := make(chan *upbit.MarketTicker)
		chs = append(chs, ch)
		go func(marketName string, c chan *upbit.MarketTicker) {
			defer close(c)
			var result *upbit.MarketTicker = nil
			for _, s := range strategies {
				if s.CheckMarket(marketName) {
					result = GetLast(history[marketName])
					break
				}
			}
			c <- result
		}(marketName, ch)
	}
	result := make([]upbit.MarketTicker, 0, len(history))
	for _, ch := range chs {
		m := <-ch
		if m != nil {
			buyItem := wallet.TickerMap[m.MarketName]
			buyingPrice := buyItem.MarketTicker.TradePrice
			lastPrice := history[m.MarketName][len(history[m.MarketName])-1].TradePrice
			changePrice := buyItem.Quantity.Mul(lastPrice.Sub(buyingPrice))
			fLastPrice, _ := lastPrice.Float64()
			fBuyingPrice, _ := buyingPrice.Float64()
			fChangePrice, _ := changePrice.Float64()
			now := time.Now().Format(time.RFC3339)
			CurrencyPrinter.Printf("[%s] Sell / %s, Sell: %f, Buy: %f (%f)\n", now, m.MarketName, fLastPrice, fBuyingPrice, fChangePrice)

			result = append(result, *m)
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
		if (*minMax)[v.MarketName] != nil {
			delete((*minMax), v.MarketName)
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

	process := func() error {
		ctx1, cancel1 := context.WithTimeout(ctx, time.Second*1)
		list, err := upbit.GetMarketList(ctx1)
		defer cancel1()
		if err != nil {
			return err
		}
		marketIds := make([]string, 0)
		for _, m := range list {
			if strings.HasPrefix(m.MarketName, "KRW") {
				marketIds = append(marketIds, m.MarketName)
			}
		}
		ctx2, cancel2 := context.WithTimeout(ctx, time.Second*1)
		defer cancel2()
		list2, err := upbit.GetMarketTickerList(ctx2, marketIds)
		if err != nil {
			return err
		}
		select {
		case result <- list2:
		case <-ctx.Done():
		}
		return nil
	}

	go func() {
		defer close(result)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				err := process()
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					time.Sleep(1 * time.Second)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return result
}

func GetCandleStream(ctx context.Context) chan []upbit.MarketCandle {
	result := make(chan []upbit.MarketCandle)
	ticker := time.NewTicker(time.Second * 30)

	process := func() error {
		ctx1, cancel1 := context.WithTimeout(ctx, time.Second*1)
		list, err := upbit.GetMarketList(ctx1)
		defer cancel1()
		if err != nil {
			return err
		}
		marketIds := make([]string, 0)
		for _, m := range list {
			if strings.HasPrefix(m.MarketName, "KRW") {
				marketIds = append(marketIds, m.MarketName)
			}
		}
		for _, marketId := range marketIds {
			ctx2, cancel2 := context.WithTimeout(ctx, time.Second*2)
			list2, err := upbit.GetMinuteMarketCandleList(ctx2, marketId, 1, nil, 30)
			defer cancel2()
			if err != nil {
				return err
			}
			select {
			case result <- list2:
			case <-ctx.Done():
			}
		}
		return nil
	}

	go func() {
		defer close(result)
		defer ticker.Stop()
		process()
		for {
			select {
			case <-ticker.C:
				err := process()
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					time.Sleep(15 * time.Second)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return result
}

func Run(ctx context.Context, wallet *trade.Wallet, history *History, minuteCandles *MinuteCandles, minMax *MinMax) {
	tickerStream := GetTickerStream(ctx)
	candleStream := GetCandleStream(ctx)

	buyStrategies := make([]strategy.BuyStrategy, 1)
	sellStrategies := make([]strategy.SellStrategy, 1)
	for {
		select {
		case candles := <-candleStream:
			(*minuteCandles)[candles[0].MarketName] = candles
		case tickers := <-tickerStream:
			buyStrategies[0] = strategy.NewRsiBuyStrategy(*history, *wallet, *minuteCandles)
			sellStrategies[0] = strategy.NewRsiSellStrategy(*history, *wallet, *minuteCandles)
			*history = UpdateHistory(*history, tickers)
			marketsToBuy := GetMarketsToBuy(*history, wallet, buyStrategies)
			SendBuyMessagesToWallet(wallet, *history, marketsToBuy)
			UpdateMinMaxForBuying(minMax, *wallet, *history)
			marketsToSell := GetMarketTickerToSell(*wallet, *history, *minMax, sellStrategies)
			SendSellMessageToWallet(wallet, marketsToSell)
			UpdateMinMaxForSelling(minMax, marketsToSell)
		case <-ctx.Done():
			return
		}
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wallet := trade.NewWallet(ctx, WALLET_START)
	history := make(History)
	minuteCandles := make(MinuteCandles)
	minMax := make(MinMax)
	go Run(ctx, wallet, &history, &minuteCandles, &minMax)

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
		case "m", "minuteCandles":
			println()
			fmt.Println(minuteCandles.Summary())
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
