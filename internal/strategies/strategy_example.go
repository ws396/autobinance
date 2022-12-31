package strategies

import (
	"github.com/sdcoffey/techan"
	"github.com/ws396/autobinance/internal/globals"
)

func init() {
	AddStrategyInfo("example", StrategyExample, []string{
		"SMA0",
		"SMA1",
	})
}

type buyRuleExample struct {
	SMA10  techan.Indicator
	series *techan.TimeSeries
}

func (r buyRuleExample) IsSatisfied() bool {
	l := len(r.series.Candles)

	a0 := r.SMA10.Calculate(l - 3)
	a1 := r.SMA10.Calculate(l - 1)
	if !(r.series.LastCandle().ClosePrice.GT(a1) && a1.GT(a0)) {
		return false
	}

	return true
}

type sellRuleExample struct {
	SMA10  techan.Indicator
	series *techan.TimeSeries
}

func (r sellRuleExample) IsSatisfied() bool {
	l := len(r.series.Candles)

	a0 := r.SMA10.Calculate(l - 3)
	a1 := r.SMA10.Calculate(l - 1)
	if !(r.series.LastCandle().ClosePrice.LT(a1) && a1.LT(a0)) {
		return false
	}

	return true
}

func StrategyExample(series *techan.TimeSeries) (string, map[string]string) {
	closePrices := techan.NewClosePriceIndicator(series)
	SMA10 := techan.NewSimpleMovingAverage(closePrices, 10)

	buyRule := buyRuleExample{SMA10, series}
	sellRule := sellRuleExample{SMA10, series}

	result := globals.Hold
	if buyRule.IsSatisfied() {
		result = globals.Buy
	} else if sellRule.IsSatisfied() {
		result = globals.Sell
	}

	indicators := map[string]string{
		"SMA0": SMA10.Calculate(len(series.Candles) - 3).String(),
		"SMA1": SMA10.Calculate(len(series.Candles) - 1).String(),
	}

	return result, indicators
}
