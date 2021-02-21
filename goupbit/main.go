package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"math/big"

	//"encoding/json"
	"github.com/goccy/go-json"
)

// BaseURL is
const BaseURL = "https://api.upbit.com/v1"

// Market is
type Market struct {
	MarketWarning string `json:"market_warning"`
	Market        string `json:"market"`
	KoreanName    string `json:"korean_name"`
	EnglishName   string `json:"english_name"`
}

// MarketTicker is
type MarketTicker struct {
	Market             string    `json:"market"`
	TradeDate          string    `json:"trade_date"`
	TradeTime          string    `json:"trade_time"`
	TradeDateKst       string    `json:"trade_date_kst"`
	TradeTimeKst       string    `json:"trade_time_kst"`
	TradeTimestamp     int64     `json:"trade_timestamp"`
	OpeningPrice       big.Float `json:"opening_price"`
	HighPrice          big.Float `json:"high_price"`
	LowPrice           big.Float `json:"low_price"`
	TradePrice         big.Float `json:"trade_price"`
	PrevClosingPrice   big.Float `json:"prev_closing_price"`
	Change             string    `json:"change"`
	ChangePrice        big.Float `json:"change_price"`
	ChangeRate         float64   `json:"change_rate"`
	SignedChangePrice  big.Float `json:"signed_change_price"`
	SignedChangeRate   float64   `json:"signed_change_rate"`
	TradeVolume        big.Float `json:"trade_volume"`
	AccTradePrice      big.Float `json:"acc_trade_price"`
	AccTradePrice24H   big.Float `json:"acc_trade_price_24h"`
	AccTradeVolume     big.Float `json:"acc_trade_volume"`
	AccTradeVolume24H  big.Float `json:"acc_trade_volume_24h"`
	Highest52WeekPrice big.Float `json:"highest_52_week_price"`
	Highest52WeekDate  string    `json:"highest_52_week_date"`
	Lowest52WeekPrice  big.Float `json:"lowest_52_week_price"`
	Lowest52WeekDate   string    `json:"lowest_52_week_date"`
	Timestamp          int64     `json:"timestamp"`
}

// InternalError is
type InternalError struct {
	Msg   string
	Cause error
}

// NewInternalError is
func NewInternalError(msg string, cause error) InternalError {
	return InternalError{Msg: msg, Cause: cause}
}

func (e InternalError) Error() string {
	return fmt.Sprintf("%s\nCause By %s", e.Msg, e.Cause.Error())
}

// GetMarketList is
func GetMarketList() ([]Market, error) {
	response, err := http.Get(fmt.Sprintf("%s/market/all?isDetails=true", BaseURL))
	if err != nil {
		return nil, NewInternalError("GetMarketList Http Error", err)
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, NewInternalError("GetMarketList Http Error", err)
	}
	defer response.Body.Close()
	var r []Market
	err = json.Unmarshal(contents, &r)
	if err != nil {
		return nil, NewInternalError("GetMarketList Invalid Json Error", err)
	}
	return r, nil
}

// GetMarketTickerList is
func GetMarketTickerList(marketIDs []string) ([]MarketTicker, error) {
	marketsParam := strings.Join(marketIDs, ",")
	response, err := http.Get(fmt.Sprintf("%s/ticker?markets=%s", BaseURL, marketsParam))
	if err != nil {
		return nil, NewInternalError("GetMarketTickerList Http Error", err)
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, NewInternalError("GetMarketTickerList Http Error", err)
	}
	defer response.Body.Close()
	var r []MarketTicker
	err = json.Unmarshal(contents, &r)
	if err != nil {
		return nil, NewInternalError("GetMarketTickerList Invalid Json Error", err)
	}
	return r, nil
}

func checkError(err error, msg string) {
	if err != nil {
		panic(fmt.Errorf("%s : Cause By\n%s", msg, err))
	}
}

func main() {
	ticker := time.NewTicker(time.Second * 1)
	for {
		<-ticker.C
		list, err := GetMarketList()
		checkError(err, "Application Error")
		marketIDs := make([]string, 0)
		for _, m := range list {
			//fmt.Println(m)
			if strings.HasPrefix(m.Market, "KRW") {
				marketIDs = append(marketIDs, m.Market)
			}
		}
		list2, err := GetMarketTickerList(marketIDs)
		checkError(err, "Application Error")
		fmt.Printf("cnt: %d", len(list2))
		//for _, t := range list2 {
		//	fmt.Println(t)
		//}
	}
}
