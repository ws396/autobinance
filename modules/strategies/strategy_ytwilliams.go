package strategies

import (
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"github.com/ws396/autobinance/modules/techanext"
)

type buyRuleYTWilliams struct {
	EMA50        techan.Indicator
	MACDH        techan.Indicator
	WilliamsR    techan.Indicator
	WilliamsREMA techan.Indicator
	series       *techan.TimeSeries
}

func (r buyRuleYTWilliams) IsSatisfied(index int, record *techan.TradingRecord) bool {
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

type sellRuleYTWilliams struct {
	EMA50        techan.Indicator
	MACDH        techan.Indicator
	WilliamsR    techan.Indicator
	WilliamsREMA techan.Indicator
	series       *techan.TimeSeries
}

func (r sellRuleYTWilliams) IsSatisfied(index int, record *techan.TradingRecord) bool {
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

func StrategyYTWilliams(symbol string, series *techan.TimeSeries) (string, map[string]string) {
	closePrices := techan.NewClosePriceIndicator(series)

	EMA50 := techan.NewEMAIndicator(closePrices, 50)
	// Above EMA + upward slope - buy, below EMA + downward slope - sell (Not sure if I like this, but okay)

	MACDH := techan.NewMACDHistogramIndicator(techan.NewMACDIndicator(closePrices, 12, 26), 9)
	// Negative->posiive for buy, positive->negative for sell

	WilliamsR := techanext.NewWilliamsRIndicator(series, 21)

	WilliamsREMA := techan.NewEMAIndicator(WilliamsR, 13)
	// WR downcross through EMA and WR + EMA downcross -20 line - sell, WR upcross through EMA (and WR + EMA upcross -80 line?) - buy
	// Need to check 3 or 4 candles here, because it is acceptable for these conditions to not be fulfilled simultaneously.
	// Although I think I can simply check if the WR through EMA cross happened above -20 or below -80? Shouldn't make a huge difference.

	record := techan.NewTradingRecord()

	entryRule := techan.And(
		buyRuleYTWilliams{EMA50, MACDH, WilliamsR, WilliamsREMA, series},
		techan.PositionNewRule{},
	)

	exitRule := techan.And(
		sellRuleYTWilliams{EMA50, MACDH, WilliamsR, WilliamsREMA, series},
		techan.PositionOpenRule{},
	)

	strategy := techan.RuleStrategy{
		UnstablePeriod: 10,
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}

	result := "Hold"
	if strategy.ShouldEnter(20, record) {
		result = "Buy"
	} else if strategy.ShouldExit(20, record) {
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
