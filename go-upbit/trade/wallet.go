package trade

import (
	"context"
	"ipoemi/go-upbit/upbit"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var CurrencyPrinter = message.NewPrinter(language.English)

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

func NewWallet(amount decimal.Decimal) *Wallet {
	return &Wallet{
		Amount:       amount,
		TickerMap:    make(map[string]*BuyItem),
		MessageQueue: make(chan interface{}),
	}
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
		tickerAmount = tickerAmount.Add(v.Quantity.Mul(lastMarketMap[v.MarketTicker.Market].TradePrice))
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
	for _, v := range w.TickerMap {
		marketAmount := v.Quantity.Mul(lastMarketMap[v.MarketTicker.Market].TradePrice)
		resultAmount = resultAmount.Add(marketAmount)
		fMarketAmount, _ := marketAmount.Float64()
		resultBuilder = append(resultBuilder, CurrencyPrinter.Sprintf("- %s Amount: %f", v.MarketTicker.Market, fMarketAmount))
	}
	resultBuilder = append(resultBuilder, "==================================================================")
	fResultAmount, _ := resultAmount.Float64()
	resultBuilder = append(resultBuilder, CurrencyPrinter.Sprintf(" Total Amout: %f", fResultAmount))
	return strings.Join(resultBuilder, "\n")
}

func (w *Wallet) ProcessMessage(ctx context.Context) {
	go func() {
		done := false
		for !done {
			select {
			case m := <-w.MessageQueue:
				switch v := m.(type) {
				case WalletBuyMessage:
					w.TickerMap[v.MarketTicker.Market] = NewBuyItem(v.MarketTicker, v.Quantity)
					w.Amount = w.Amount.Sub(v.MarketTicker.TradePrice.Mul(v.Quantity))

					fTradePrice, _ := v.MarketTicker.TradePrice.Float64()
					now := time.Now().Format(time.RFC3339)
					CurrencyPrinter.Printf("[%s] Buy / %s, Last: %f\n", now, v.MarketTicker.Market, fTradePrice)

					*v.Sender <- nil

				case WalletSellMessage:
					buyItem := w.TickerMap[v.MarketTicker.Market]
					buyingPrice := buyItem.MarketTicker.TradePrice
					lastPrice := v.MarketTicker.TradePrice
					earnPrice := buyItem.Quantity.Mul(lastPrice)
					w.Amount = w.Amount.Add(earnPrice)
					delete(w.TickerMap, v.MarketTicker.Market)

					changePrice := buyItem.Quantity.Mul(lastPrice.Sub(buyingPrice))
					fLastPrice, _ := lastPrice.Float64()
					fBuyingPrice, _ := buyingPrice.Float64()
					fChangePrice, _ := changePrice.Float64()
					now := time.Now().Format(time.RFC3339)
					CurrencyPrinter.Printf("[%s] Sell / %s, Sell: %f, Buy: %f (%f)\n", now, v.MarketTicker.Market, fLastPrice, fBuyingPrice, fChangePrice)

					*v.Sender <- nil
				}
			case <-ctx.Done():
				done = true
			}
		}
	}()
}
