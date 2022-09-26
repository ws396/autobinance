package strategies

import (
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"github.com/ws396/autobinance/modules/techanext"
)

type buyRuleShortened struct {
	EMA50        techan.Indicator
	MACDH        techan.Indicator
	WilliamsR    techan.Indicator
	WilliamsREMA techan.Indicator
	series       *techan.TimeSeries
}

func (r buyRuleShortened) IsSatisfied() bool {
	l := len(r.series.Candles)

	a0 := r.EMA50.Calculate(l - 2)
	a1 := r.EMA50.Calculate(l - 1)
	if !(r.series.LastCandle().ClosePrice.GT(a1) && a1.GT(a0)) {
		return false
	}

	b0 := r.MACDH.Calculate(l - 2)
	b1 := r.MACDH.Calculate(l - 1)
	if !(b0.GT(big.NewDecimal(0)) && b1.LT(big.NewDecimal(0))) {
		return false
	}

	c0 := r.WilliamsR.Calculate(l - 2)
	c1 := r.WilliamsR.Calculate(l - 1)
	d0 := r.WilliamsREMA.Calculate(l - 2)
	d1 := r.WilliamsREMA.Calculate(l - 1)
	if !(c0.GT(d0) && c1.LT(d1)) {
		return false
	}

	return true
}

type sellRuleShortened struct {
	EMA50        techan.Indicator
	MACDH        techan.Indicator
	WilliamsR    techan.Indicator
	WilliamsREMA techan.Indicator
	series       *techan.TimeSeries
}

func (r sellRuleShortened) IsSatisfied() bool {
	l := len(r.series.Candles)

	a0 := r.EMA50.Calculate(l - 2)
	a1 := r.EMA50.Calculate(l - 1)
	if !(r.series.LastCandle().ClosePrice.LT(a1) && a1.LT(a0)) {
		return false
	}

	b0 := r.MACDH.Calculate(l - 2)
	b1 := r.MACDH.Calculate(l - 1)
	if !(b0.LT(big.NewDecimal(0)) && b1.GT(big.NewDecimal(0))) {
		return false
	}

	c0 := r.WilliamsR.Calculate(l - 2)
	c1 := r.WilliamsR.Calculate(l - 1)
	d0 := r.WilliamsREMA.Calculate(l - 2)
	d1 := r.WilliamsREMA.Calculate(l - 1)
	if !(c0.LT(d0) && c1.GT(d1) && c1.LT(big.NewDecimal(-15))) {
		return false
	}

	return true
}

func StrategyShortened(series *techan.TimeSeries) (string, map[string]string) {
	closePrices := techan.NewClosePriceIndicator(series)
	EMA50 := techan.NewEMAIndicator(closePrices, 50)
	MACDH := techan.NewMACDHistogramIndicator(techan.NewMACDIndicator(closePrices, 12, 26), 9)
	WilliamsR := techanext.NewWilliamsRIndicator(series, 21)
	WilliamsREMA := techan.NewEMAIndicator(WilliamsR, 13)

	//record := techan.NewTradingRecord() // Currently not checking if position is new or not

	buyRule := buyRuleShortened{EMA50, MACDH, WilliamsR, WilliamsREMA, series}
	sellRule := sellRuleShortened{EMA50, MACDH, WilliamsR, WilliamsREMA, series}

	result := "Hold"
	if buyRule.IsSatisfied() {
		result = "Buy"
	} else if sellRule.IsSatisfied() {
		result = "Sell"
	}

	indicators := map[string]string{
		"EMA0":          EMA50.Calculate(len(series.Candles) - 2).String(),
		"EMA1":          EMA50.Calculate(len(series.Candles) - 1).String(),
		"MACDH0":        MACDH.Calculate(len(series.Candles) - 2).String(),
		"MACDH1":        MACDH.Calculate(len(series.Candles) - 1).String(),
		"WilliamsR0":    WilliamsR.Calculate(len(series.Candles) - 2).String(),
		"WilliamsR1":    WilliamsR.Calculate(len(series.Candles) - 1).String(),
		"WilliamsREMA0": WilliamsREMA.Calculate(len(series.Candles) - 2).String(),
		"WilliamsREMA1": WilliamsREMA.Calculate(len(series.Candles) - 1).String(),
	}

	return result, indicators
}
