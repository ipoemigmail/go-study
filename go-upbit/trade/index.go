package trade

import "github.com/shopspring/decimal"

func GetRsi(tradePrices []decimal.Decimal) float64 {
	if len(tradePrices) == 0 {
		return -1
	}

	size := len(tradePrices) - 1
	//ONE := decimal.NewFromInt(1)

	upSum := decimal.Zero
	downSum := decimal.Zero
	factor := decimal.NewFromInt(1).Sub(decimal.NewFromFloat(1.0 / 14.0))
	factorList := make([]decimal.Decimal, size)
	factorList[size-1] = decimal.NewFromInt(1)
	factorSum := factorList[size-1]
	for i := size - 2; i >= 0; i-- {
		factorList[i] = factorList[i+1].Mul(factor)
		factorSum = factorSum.Add(factorList[i])
	}

	for i := range tradePrices[:len(tradePrices)-1] {
		diff := tradePrices[i+1].Sub(tradePrices[i])
		if diff.IsPositive() {
			upSum = upSum.Add(diff.Mul(factorList[i]))
		} else {
			downSum = downSum.Add(diff.Abs().Mul(factorList[i]))
		}
	}

	if upSum.IsZero() && downSum.IsZero() {
		return -1
	}

	au := upSum.Div(factorSum)
	ad := downSum.Div(factorSum)

	result, _ := au.Div(au.Add(ad)).Float64()
	return result
}
