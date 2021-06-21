package upbit

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/goccy/go-json"
	"golang.org/x/time/rate"
)

// BaseURL is
const BaseURL = "https://api.upbit.com/v1"

const SECOND_LIMIT = 5
const MINUTE_LIMIT = 500

var rateLimiterPerSecond map[string]*rate.Limiter = map[string]*rate.Limiter{
	"ticker":  rate.NewLimiter(rate.Every(time.Second/time.Duration(SECOND_LIMIT)), SECOND_LIMIT),
	"market":  rate.NewLimiter(rate.Every(time.Second/time.Duration(SECOND_LIMIT)), SECOND_LIMIT),
	"candles": rate.NewLimiter(rate.Every(time.Second/time.Duration(SECOND_LIMIT)), SECOND_LIMIT),
}

var rateLimiterPerMinute map[string]*rate.Limiter = map[string]*rate.Limiter{
	"ticker":  rate.NewLimiter(rate.Every(time.Minute/time.Duration(MINUTE_LIMIT)), MINUTE_LIMIT),
	"market":  rate.NewLimiter(rate.Every(time.Minute/time.Duration(MINUTE_LIMIT)), MINUTE_LIMIT),
	"candles": rate.NewLimiter(rate.Every(time.Minute/time.Duration(MINUTE_LIMIT)), MINUTE_LIMIT),
}

// ErrorResponse is
type ErrorResponse struct {
	Error struct {
		Name    int    `json:"name"`
		Message string `json:"message"`
	} `json:"error"`
}

// Market is
type Market struct {
	MarketWarning string `json:"market_warning"`
	MarketName    string `json:"market"`
	KoreanName    string `json:"korean_name"`
	EnglishName   string `json:"english_name"`
}

// MarketTicker is
type MarketTicker struct {
	MarketName         string          `json:"market"`
	TradeDate          string          `json:"trade_date"`
	TradeTime          string          `json:"trade_time"`
	TradeDateKst       string          `json:"trade_date_kst"`
	TradeTimeKst       string          `json:"trade_time_kst"`
	TradeTimestamp     int64           `json:"trade_timestamp"`
	OpeningPrice       decimal.Decimal `json:"opening_price"`
	HighPrice          decimal.Decimal `json:"high_price"`
	LowPrice           decimal.Decimal `json:"low_price"`
	TradePrice         decimal.Decimal `json:"trade_price"`
	PrevClosingPrice   decimal.Decimal `json:"prev_closing_price"`
	Change             string          `json:"change"`
	ChangePrice        decimal.Decimal `json:"change_price"`
	ChangeRate         float64         `json:"change_rate"`
	SignedChangePrice  decimal.Decimal `json:"signed_change_price"`
	SignedChangeRate   float64         `json:"signed_change_rate"`
	TradeVolume        decimal.Decimal `json:"trade_volume"`
	AccTradePrice      decimal.Decimal `json:"acc_trade_price"`
	AccTradePrice24H   decimal.Decimal `json:"acc_trade_price_24h"`
	AccTradeVolume     decimal.Decimal `json:"acc_trade_volume"`
	AccTradeVolume24H  decimal.Decimal `json:"acc_trade_volume_24h"`
	Highest52WeekPrice decimal.Decimal `json:"highest_52_week_price"`
	Highest52WeekDate  string          `json:"highest_52_week_date"`
	Lowest52WeekPrice  decimal.Decimal `json:"lowest_52_week_price"`
	Lowest52WeekDate   string          `json:"lowest_52_week_date"`
	Timestamp          int64           `json:"timestamp"`
}

// MarketCandle is
type MarketCandle struct {
	MarketName           string          `json:"market"`
	CandleDateTimeUtc    string          `json:"candle_date_time_utc"`
	CandleDateTimeKst    string          `json:"candle_date_time_kst"`
	OpeningPrice         decimal.Decimal `json:"opening_price"`
	HighPrice            decimal.Decimal `json:"high_price"`
	LowPrice             decimal.Decimal `json:"low_price"`
	TradePrice           decimal.Decimal `json:"trade_price"`
	Timestamp            int64           `json:"timestamp"`
	CandleAccTradePrice  float64         `json:"candle_acc_trade_price"`
	CandleAccTradeVolume float64         `json:"candle_acc_trade_volume"`
	Unit                 int             `json:"unit"`
}

// InternalError is
type InternalError struct {
	Msg   string
	Stack []byte
	Cause error
}

// NewInternalError is
func NewInternalError(msg string, cause error) *InternalError {
	return &InternalError{Msg: msg, Stack: debug.Stack(), Cause: cause}
}

func (e InternalError) Error() string {
	return fmt.Sprintf("%s: %v\n%s", e.Msg, e.Cause.Error(), e.Stack)
}

// GetMarketList is
func GetMarketList(ctx context.Context) ([]Market, error) {
	rateLimiterPerSecond["market"].Wait(ctx)
	rateLimiterPerMinute["market"].Wait(ctx)
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/market/all?isDetails=true", BaseURL), nil)
	if err != nil {
		panic(err)
	}
	response, err := http.DefaultClient.Do(req)
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
		var er ErrorResponse
		err2 := json.Unmarshal(contents, &er)
		if err2 != nil {
			return nil, NewInternalError("GetMarketList Invalid Json Error", fmt.Errorf(string(contents)))
		}
		return nil, NewInternalError(er.Error.Message, fmt.Errorf(""))
	}
	return r, nil
}

// GetMarketTickerList is
func GetMarketTickerList(ctx context.Context, marketIDs []string) ([]MarketTicker, error) {
	rateLimiterPerSecond["ticker"].Wait(ctx)
	rateLimiterPerMinute["ticker"].Wait(ctx)
	marketsParam := strings.Join(marketIDs, ",")
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/ticker?markets=%s", BaseURL, marketsParam), nil)
	if err != nil {
		panic(err)
	}
	response, err := http.DefaultClient.Do(req)
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
	if err != nil {
		var er ErrorResponse
		err2 := json.Unmarshal(contents, &er)
		if err2 != nil {
			return nil, NewInternalError("GetMarketTickerList Invalid Json Error", fmt.Errorf(string(contents)))
		}
		return nil, NewInternalError(er.Error.Message, fmt.Errorf(""))
	}
	return r, nil
}

var lastLimit string

// GetMinuteMarketCandleList is
func GetMinuteMarketCandleList(ctx context.Context, marketId string, unit int, to *time.Time, count int) ([]MarketCandle, error) {
	rateLimiterPerSecond["candles"].Wait(ctx)
	rateLimiterPerMinute["candles"].Wait(ctx)
	var url string
	if to == nil {
		url = fmt.Sprintf("%s/candles/minutes/%v?market=%s&count=%v", BaseURL, unit, marketId, count)
	} else {
		timeFormat := to.UTC().Format(time.RFC3339)
		url = fmt.Sprintf("%s/candles/minutes/%v?market=%s&to=%v&count=%v", BaseURL, unit, marketId, timeFormat, count)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		panic(err)
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, NewInternalError("GetMarketCandleList Http Error", err)
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, NewInternalError("GetMarketCandleList Http Error", err)
	}
	defer response.Body.Close()
	for k, v := range response.Header {
		if strings.Contains(k, "Remaining") {
			//fmt.Printf("[%v] k=%v, v=%v\n", time.Now().Format(time.RFC3339), k, v)
			lastLimit = fmt.Sprintf("[%v] k=%v, v=%v\n", time.Now().Format(time.RFC3339), k, v)
		}
	}
	var r []MarketCandle
	err = json.Unmarshal(contents, &r)
	if err != nil {
		fmt.Print(lastLimit)
		var er ErrorResponse
		err2 := json.Unmarshal(contents, &er)
		if err2 != nil {
			return nil, NewInternalError("GetMarketCandleList Invalid Json Error", fmt.Errorf(string(contents)))
		}
		return nil, NewInternalError(er.Error.Message, fmt.Errorf(""))
	}
	sort.Slice(r, func(i, j int) bool {
		return r[i].CandleDateTimeKst < r[j].CandleDateTimeKst
	})
	return r, nil
}
