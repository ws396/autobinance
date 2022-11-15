package techanext

import (
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
)

type WMAIndicator struct {
	indicator techan.Indicator
	window    int
}

func NewWMAIndicator(indicator techan.Indicator, window int) techan.Indicator {
	return WMAIndicator{
		indicator: indicator,
		window:    window,
	}
}

func (wma WMAIndicator) Calculate(index int) big.Decimal {
	norm, sum := big.Decimal(big.ZERO), big.Decimal(big.ZERO)

	for i := wma.window - 1; i >= 0; i-- {
		weight := big.NewFromInt((wma.window - i) * wma.window)
		norm = norm.Add(weight)
		sum = sum.Add(wma.indicator.Calculate(index - i).Mul(weight))
	}

	result := sum.Div(norm)

	return result
}
