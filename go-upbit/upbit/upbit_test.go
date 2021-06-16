package upbit

import (
	"context"
	"testing"
	"time"
)

func TestGetMarketCandleList(t *testing.T) {
	to := time.Now()
	_, err := GetMinuteMarketCandleList(context.TODO(), "KRW-BTC", 1, &to, 10)
	if err != nil {
		panic(err)
	}
}
