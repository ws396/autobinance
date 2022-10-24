package strategies

import (
	"github.com/sdcoffey/techan"
)

type buyRuleExample struct {
	EMA50  techan.Indicator
	series *techan.TimeSeries
}

func (r buyRuleExample) IsSatisfied() bool {
	l := len(r.series.Candles)

	a0 := r.EMA50.Calculate(l - 2)
	a1 := r.EMA50.Calculate(l - 1)
	if !(r.series.LastCandle().ClosePrice.GT(a1) && a1.GT(a0)) {
		return false
	}

	return true
}

type sellRuleExample struct {
	EMA50  techan.Indicator
	series *techan.TimeSeries
}

func (r sellRuleExample) IsSatisfied() bool {
	l := len(r.series.Candles)

	a0 := r.EMA50.Calculate(l - 2)
	a1 := r.EMA50.Calculate(l - 1)
	if !(r.series.LastCandle().ClosePrice.LT(a1) && a1.LT(a0)) {
		return false
	}

	return true
}

func StrategyExample(series *techan.TimeSeries) (string, map[string]string) {
	closePrices := techan.NewClosePriceIndicator(series)
	EMA50 := techan.NewEMAIndicator(closePrices, 50)

	//record := techan.NewTradingRecord() // Currently not checking if position is new or not

	buyRule := buyRuleExample{EMA50, series}
	sellRule := sellRuleExample{EMA50, series}

	result := "Hold"
	if buyRule.IsSatisfied() {
		result = "Buy"
	} else if sellRule.IsSatisfied() {
		result = "Sell"
	}

	indicators := map[string]string{
		"EMA0": EMA50.Calculate(len(series.Candles) - 2).String(),
		"EMA1": EMA50.Calculate(len(series.Candles) - 1).String(),
	}

	return result, indicators
}
