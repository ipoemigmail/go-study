package trade

import (
	"context"
	"ipoemi/go-upbit/upbit"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var CurrencyPrinter = message.NewPrinter(language.English)
var FeeRate = decimal.NewFromFloat(0.05 / 100)

type BuyItem struct {
	MarketTicker upbit.MarketTicker
	Quantity     decimal.Decimal
}

func NewBuyItem(m upbit.MarketTicker, q decimal.Decimal) *BuyItem {
	return &BuyItem{
		MarketTicker: m,
		Quantity:     q,
	}
}

type Wallet struct {
	Amount       decimal.Decimal
	TickerMap    map[string]*BuyItem
	MessageQueue chan interface{}
}

type WalletBuyMessage struct {
	MarketTicker upbit.MarketTicker
	Quantity     decimal.Decimal
	Sender       *chan error
}

type WalletSellMessage struct {
	MarketTicker upbit.MarketTicker
	Sender       *chan error
}

func NewWallet(ctx context.Context, amount decimal.Decimal) *Wallet {
	w := &Wallet{
		Amount:       amount,
		TickerMap:    make(map[string]*BuyItem),
		MessageQueue: make(chan interface{}),
	}

	go func() {
		done := false
		for !done {
			select {
			case m := <-w.MessageQueue:
				switch v := m.(type) {
				case WalletBuyMessage:
					w.TickerMap[v.MarketTicker.MarketName] = NewBuyItem(v.MarketTicker, v.Quantity)
					feePrice := v.MarketTicker.TradePrice.Mul(v.Quantity).Mul(FeeRate)
					w.Amount = w.Amount.Sub(v.MarketTicker.TradePrice.Mul(v.Quantity)).Sub(feePrice)
					*v.Sender <- nil
				case WalletSellMessage:
					buyItem := w.TickerMap[v.MarketTicker.MarketName]
					lastPrice := v.MarketTicker.TradePrice
					earnPrice := buyItem.Quantity.Mul(lastPrice)
					feePrice := earnPrice.Mul(FeeRate)
					w.Amount = w.Amount.Add(earnPrice).Sub(feePrice)
					delete(w.TickerMap, v.MarketTicker.MarketName)
					*v.Sender <- nil
				}
			case <-ctx.Done():
				done = true
			}
		}
	}()
	return w
}

func (w *Wallet) Buy(m upbit.MarketTicker, quantity decimal.Decimal) chan error {
	ch := make(chan error, 1)
	w.MessageQueue <- WalletBuyMessage{MarketTicker: m, Quantity: quantity, Sender: &ch}
	return ch
}

func (w *Wallet) Sell(m upbit.MarketTicker) chan error {
	ch := make(chan error, 1)
	w.MessageQueue <- WalletSellMessage{MarketTicker: m, Sender: &ch}
	return ch
}

func (w Wallet) AllAmount(lastMarketMap map[string]*upbit.MarketTicker) decimal.Decimal {
	tickerAmount := decimal.Zero
	for _, v := range w.TickerMap {
		tickerAmount = tickerAmount.Add(v.Quantity.Mul(lastMarketMap[v.MarketTicker.MarketName].TradePrice))
	}
	return w.Amount.Add(tickerAmount)
}

func (w Wallet) Summary(lastMarketMap map[string]*upbit.MarketTicker) string {
	resultAmount := decimal.Zero
	resultBuilder := make([]string, 0)
	now := time.Now().Format(time.RFC3339)
	resultBuilder = append(resultBuilder, "==================================================================")
	resultBuilder = append(resultBuilder, CurrencyPrinter.Sprintf(" Wallet Summary [%s]", now))
	resultBuilder = append(resultBuilder, "==================================================================")
	resultAmount = resultAmount.Add(w.Amount)
	fAmount, _ := w.Amount.Float64()
	resultBuilder = append(resultBuilder, CurrencyPrinter.Sprintf("- Amount: %f", fAmount))
	resultBuilderForItem := make([]string, 0)
	for _, v := range w.TickerMap {
		buyPrice := v.MarketTicker.TradePrice
		fBuyPrice, _ := buyPrice.Float64()
		marketAmount := v.Quantity.Mul(lastMarketMap[v.MarketTicker.MarketName].TradePrice)
		resultAmount = resultAmount.Add(marketAmount)
		fLastPrice, _ := lastMarketMap[v.MarketTicker.MarketName].TradePrice.Float64()
		fMarketAmount, _ := marketAmount.Float64()
		resultBuilderForItem = append(resultBuilderForItem, CurrencyPrinter.Sprintf("- %s Amount: %f (Buy: %f, Cur: %f)", v.MarketTicker.MarketName, fMarketAmount, fBuyPrice, fLastPrice))
	}
	sort.Slice(resultBuilderForItem, func(i, j int) bool {
		return resultBuilderForItem[i] < resultBuilderForItem[j]
	})
	resultBuilder = append(resultBuilder, resultBuilderForItem...)
	resultBuilder = append(resultBuilder, "==================================================================")
	fResultAmount, _ := resultAmount.Float64()
	resultBuilder = append(resultBuilder, CurrencyPrinter.Sprintf(" Total Amout: %f", fResultAmount))
	return strings.Join(resultBuilder, "\n")
}
