package techanext

// Taken from https://github.com/sdcoffey/techan/pull/37/commits/b00fdf455d24569caef9784369f293413662e7a1

import (
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
)

type stochasticRSIIndicator struct {
	curRSI techan.Indicator
	minRSI techan.Indicator
	maxRSI techan.Indicator
}

// NewStochasticRSIIndicator returns a derivative Indicator which returns the stochastic RSI indicator for the given
// RSI window.
// https://www.investopedia.com/terms/s/stochrsi.asp
func NewStochasticRSIIndicator(indicator techan.Indicator, window int) techan.Indicator {
	rsiIndicator := techan.NewRelativeStrengthIndexIndicator(indicator, window)
	return stochasticRSIIndicator{
		curRSI: rsiIndicator,
		minRSI: techan.NewMinimumValueIndicator(rsiIndicator, window),
		maxRSI: techan.NewMaximumValueIndicator(rsiIndicator, window),
	}
}

func (sri stochasticRSIIndicator) Calculate(index int) big.Decimal {
	curRSI := sri.curRSI.Calculate(index)
	minRSI := sri.minRSI.Calculate(index)
	maxRSI := sri.maxRSI.Calculate(index)

	if minRSI.EQ(maxRSI) {
		return big.NewDecimal(100)
	}

	return curRSI.Sub(minRSI).Div(maxRSI.Sub(minRSI)).Mul(big.NewDecimal(100))
}

type stochRSIKIndicator struct {
	stochasticRSI techan.Indicator
	window        int
}

// NewFastStochasticRSIIndicator returns a derivative Indicator which returns the fast stochastic RSI indicator (%K)
// for the given stochastic window.
func NewFastStochasticRSIIndicator(stochasticRSI techan.Indicator, window int) techan.Indicator {
	return stochRSIKIndicator{stochasticRSI, window}
}

func (k stochRSIKIndicator) Calculate(index int) big.Decimal {
	return techan.NewSimpleMovingAverage(k.stochasticRSI, k.window).Calculate(index)
}

type stochRSIDIndicator struct {
	fastStochasticRSI techan.Indicator
	window            int
}

// NewSlowStochasticRSIIndicator returns a derivative Indicator which returns the slow stochastic RSI indicator (%D)
// for the given stochastic window.
func NewSlowStochasticRSIIndicator(fastStochasticRSI techan.Indicator, window int) techan.Indicator {
	return stochRSIDIndicator{fastStochasticRSI, window}
}

func (d stochRSIDIndicator) Calculate(index int) big.Decimal {
	return techan.NewSimpleMovingAverage(d.fastStochasticRSI, d.window).Calculate(index)
}
