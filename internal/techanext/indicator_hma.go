package techanext

import (
	"math"

	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
)

type HMAIndicator struct {
	indicator techan.Indicator
	window    int
}

func NewHMAIndicator(indicator techan.Indicator, window int) techan.Indicator {
	return HMAIndicator{
		indicator: indicator,
		window:    window,
	}
}

func (hma HMAIndicator) Calculate(index int) big.Decimal {
	wma1 := NewWMAIndicator(hma.indicator, hma.window/2)
	wma2 := NewWMAIndicator(hma.indicator, hma.window)

	result := NewWMAIndicator(
		newInnerSubHMAIndicator(wma1, wma2),
		int(math.Floor(math.Sqrt(float64(hma.window)))),
	).Calculate(index)
	//ta.wma(2*ta.wma(src, length/2)-ta.wma(src, length), math.floor(math.sqrt(length)))

	return result
}

type innerSubHMAIndicator struct {
	wma1 techan.Indicator
	wma2 techan.Indicator
}

func newInnerSubHMAIndicator(wma1, wma2 techan.Indicator) techan.Indicator {
	return innerSubHMAIndicator{
		wma1: wma1,
		wma2: wma2,
	}
}

func (ish innerSubHMAIndicator) Calculate(index int) big.Decimal {
	result := big.NewDecimal(2).Mul(ish.wma1.Calculate(index)).Sub(ish.wma2.Calculate(index))

	return result
}
