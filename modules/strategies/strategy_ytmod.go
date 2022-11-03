package strategies

import (
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"github.com/ws396/autobinance/modules/globals"
	"github.com/ws396/autobinance/modules/techanext"
)

func init() {
	globals.AddStrategyDatakeys("ytmod", []string{
		"HMA0",
		"HMA1",
		"MACDH0",
		"MACDH1",
		"stochRSI0",
		"stochRSI1",
		"stochRSIEMA0",
		"stochRSIEMA1",
		"upperBB",
		"lowerBB",
	})
}

type buyRuleYTMod struct {
	HMA50       techan.Indicator
	MACDH       techan.Indicator
	stochRSI    techan.Indicator
	stochRSISMA techan.Indicator
	lowerBB     techan.Indicator
	closePrices techan.Indicator
	series      *techan.TimeSeries
}

func (r buyRuleYTMod) IsSatisfied(index int, record *techan.TradingRecord) bool {
	l := len(r.series.Candles)
	closePrice := r.series.LastCandle().ClosePrice

	a0 := r.HMA50.Calculate(l - 2)
	a1 := r.HMA50.Calculate(l - 1)
	if !(closePrice.GT(a1) && a1.GT(a0)) {
		return false
	}

	/*
		b0 := r.MACDH.Calculate(l - 4)
		b1 := r.MACDH.Calculate(l - 1)
		if !(b0.LT(big.NewDecimal(0)) && b1.GT(big.NewDecimal(0))) {
			return false
		}
		// Shouldn't affect buy?
	*/

	c0 := r.stochRSI.Calculate(l - 3)
	c1 := r.stochRSI.Calculate(l - 1)
	d0 := r.stochRSISMA.Calculate(l - 3)
	d1 := r.stochRSISMA.Calculate(l - 1)
	if !(c0.GT(d0) && c1.LT(d1)) {
		return false
	}

	// Check if close price was below the lowerBB before the last candle (e0) and check if close price upcrossed the lowerBB (e1)
	e0 := false
	for i := 2; i < 7; i++ {
		if r.lowerBB.Calculate(l - i).GT(r.closePrices.Calculate(l - i)) {
			e0 = true
			break
		}
	}
	e1 := r.lowerBB.Calculate(l - 1).LT(r.closePrices.Calculate(l - 1))
	if !(e0 && e1) {
		return false
	}

	return true
}

type sellRuleYTMod struct {
	HMA50       techan.Indicator
	MACDH       techan.Indicator
	stochRSI    techan.Indicator
	stochRSISMA techan.Indicator
	closePrices techan.Indicator
	series      *techan.TimeSeries
}

func (r sellRuleYTMod) IsSatisfied(index int, record *techan.TradingRecord) bool {
	l := len(r.series.Candles)
	closePrice := r.series.LastCandle().ClosePrice

	a0 := r.HMA50.Calculate(l - 2)
	a1 := r.HMA50.Calculate(l - 1)
	if !(closePrice.LT(a1) && a1.LT(a0)) {
		return false
	}

	b0 := r.MACDH.Calculate(l - 2)
	b1 := r.MACDH.Calculate(l - 1)
	e0 := r.closePrices.Calculate(l - 2)
	e1 := r.closePrices.Calculate(l - 1)
	if !(b1.GT(b0) && e1.GT(e0)) {
		return false
	}
	// Divergence?

	c0 := r.stochRSI.Calculate(l - 3)
	c1 := r.stochRSI.Calculate(l - 1)
	d0 := r.stochRSISMA.Calculate(l - 3)
	d1 := r.stochRSISMA.Calculate(l - 1)
	if !(c0.LT(d0) && c1.GT(d1) && c1.LT(big.NewDecimal(-15))) {
		return false
	}

	return true
}

func StrategyYTMod(series *techan.TimeSeries) (string, map[string]string) {
	closePrices := techan.NewClosePriceIndicator(series)

	//check2 := techanext.NewHMAIndicator(closePrices, 9)
	//fmt.Println(check2.Calculate(len(series.Candles) - 2))

	HMA50 := techanext.NewHMAIndicator(closePrices, 50)
	// Above EMA + upward slope - buy, below EMA + downward slope - sell (Not sure if I like this, but okay)

	MACDH := techan.NewMACDHistogramIndicator(techan.NewMACDIndicator(closePrices, 12, 26), 9)
	// Negative->posiive for buy, positive->negative for sell

	stochRSI := techanext.NewFastStochasticRSIIndicator(techanext.NewStochasticRSIIndicator(closePrices, 14), 3)

	stochRSISMA := techanext.NewSlowStochasticRSIIndicator(stochRSI, 3)
	// WR downcross through EMA and WR + EMA downcross -20 line - sell, WR upcross through EMA (and WR + EMA upcross -80 line?) - buy
	// Need to check 3 or 4 candles here, because it is acceptable for these conditions to not be fulfilled simultaneously.
	// Although I think I can simply check if the WR through EMA cross happened above -20 or below -80? Shouldn't make a huge difference.

	upperBB := techan.NewBollingerUpperBandIndicator(closePrices, 20, 1.5)
	lowerBB := techan.NewBollingerLowerBandIndicator(closePrices, 20, 1.8)

	record := techan.NewTradingRecord()

	entryRule := techan.And(
		buyRuleYTMod{HMA50, MACDH, stochRSI, stochRSISMA, lowerBB, closePrices, series},
		techan.PositionNewRule{},
	)

	exitRule := techan.And(
		sellRuleYTMod{HMA50, MACDH, stochRSI, stochRSISMA, closePrices, series},
		techan.PositionOpenRule{},
	)

	strategy := techan.RuleStrategy{
		UnstablePeriod: 10,
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}

	result := globals.Hold
	if strategy.ShouldEnter(20, record) {
		result = globals.Buy
	} else if strategy.ShouldExit(20, record) {
		result = globals.Sell
	}

	indicators := map[string]string{
		"HMA0":         HMA50.Calculate(len(series.Candles) - 2).String(),
		"HMA1":         HMA50.Calculate(len(series.Candles) - 1).String(),
		"MACDH0":       MACDH.Calculate(len(series.Candles) - 2).String(),
		"MACDH1":       MACDH.Calculate(len(series.Candles) - 1).String(),
		"stochRSI0":    stochRSI.Calculate(len(series.Candles) - 2).String(),
		"stochRSI1":    stochRSI.Calculate(len(series.Candles) - 1).String(),
		"stochRSIEMA0": stochRSISMA.Calculate(len(series.Candles) - 2).String(),
		"stochRSIEMA1": stochRSISMA.Calculate(len(series.Candles) - 1).String(),
		"upperBB":      upperBB.Calculate(len(series.Candles) - 1).String(),
		"lowerBB":      lowerBB.Calculate(len(series.Candles) - 1).String(),
	}

	return result, indicators
}
