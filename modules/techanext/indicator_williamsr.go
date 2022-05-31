package techanext

// I've just found out, that this indicator is literally a Stochastic Oscillator * -1 lol.

import (
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
)

type williamsRIndicator struct {
	series *techan.TimeSeries
	window int
}

func NewWilliamsRIndicator(series *techan.TimeSeries, window int) techan.Indicator {
	return williamsRIndicator{
		series: series,
		window: window,
	}
}

func (wri williamsRIndicator) Calculate(index int) big.Decimal {
	closingPrices := techan.NewClosePriceIndicator(wri.series).Calculate(index)
	highestHigh := techan.NewMaximumValueIndicator(techan.NewHighPriceIndicator(wri.series), wri.window).Calculate(index)
	lowestLow := techan.NewMinimumValueIndicator(techan.NewLowPriceIndicator(wri.series), wri.window).Calculate(index)

	if highestHigh.EQ(lowestLow) {
		return big.NewDecimal(0)
	}

	result := highestHigh.Sub(closingPrices).
		Div(highestHigh.Sub(lowestLow)).
		Mul(big.NewDecimal(-100))

	return result
}
