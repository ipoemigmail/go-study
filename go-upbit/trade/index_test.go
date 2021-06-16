package trade

import (
	"math"
	"testing"

	"github.com/shopspring/decimal"
)

const float64EqualityThreshold = 0.001

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

func TestGetRsi(t *testing.T) {
	v := []decimal.Decimal{
		decimal.NewFromFloat32(503),
		decimal.NewFromFloat32(502),
		decimal.NewFromFloat32(503),
		decimal.NewFromFloat32(503),
		decimal.NewFromFloat32(503),
		decimal.NewFromFloat32(504),
		decimal.NewFromFloat32(503),
		decimal.NewFromFloat32(505),
		decimal.NewFromFloat32(504),
		decimal.NewFromFloat32(505),
		decimal.NewFromFloat32(503),
		decimal.NewFromFloat32(503),
		decimal.NewFromFloat32(503),
		decimal.NewFromFloat32(504),
		decimal.NewFromFloat32(504),
	}
	expect := 0.683
	result := GetRsi(v)
	if !almostEqual(result, expect) {
		t.Errorf("expect %v but return %v", expect, result)
	}
}
