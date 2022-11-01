package strategies

import (
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"github.com/ws396/autobinance/modules/globals"
)

type buyRuleOne struct {
	MACD    techan.Indicator
	RSI     techan.Indicator
	lowerBB techan.Indicator
	series  *techan.TimeSeries
}

func (r buyRuleOne) IsSatisfied(index int, record *techan.TradingRecord) bool {
	l := len(r.series.Candles)

	a0 := r.MACD.Calculate(l - 2)
	a1 := r.MACD.Calculate(l - 1)
	if a1.GT(big.NewFromInt(0)) || a1.LTE(a0) {
		return false
	}

	b0 := r.RSI.Calculate(l - 2)
	b1 := r.RSI.Calculate(l - 1)
	if b1.GT(big.NewFromInt(33)) || b1.LTE(b0) {
		return false
	}

	c := r.lowerBB.Calculate(l - 1)
	if c.LTE(r.series.LastCandle().ClosePrice) { //nolint:gosimple // Do I want it like this?
		return false
	}

	return true
}

type sellRuleOne struct {
	MACD    techan.Indicator
	RSI     techan.Indicator
	upperBB techan.Indicator
	series  *techan.TimeSeries
}

func (r sellRuleOne) IsSatisfied(index int, record *techan.TradingRecord) bool {
	l := len(r.series.Candles)

	a0 := r.MACD.Calculate(l - 2)
	a1 := r.MACD.Calculate(l - 1)
	if a1.LT(big.NewFromInt(0)) || a1.GT(a0) {
		return false
	}

	b0 := r.RSI.Calculate(l - 2)
	b1 := r.RSI.Calculate(l - 1)
	if b1.LT(big.NewFromInt(66)) || b1.GT(b0) {
		return false
	}

	c := r.upperBB.Calculate(l - 1)
	if c.GT(r.series.LastCandle().ClosePrice) { //nolint:gosimple
		return false
	}

	return true
}

func StrategyOne(series *techan.TimeSeries) (string, map[string]string) {
	closePrices := techan.NewClosePriceIndicator(series)

	MACD := techan.NewMACDIndicator(closePrices, 12, 26)
	RSI := techan.NewRelativeStrengthIndexIndicator(closePrices, 14)
	upperBB := techan.NewBollingerUpperBandIndicator(closePrices, 20, 1.7)
	lowerBB := techan.NewBollingerLowerBandIndicator(closePrices, 20, 1.8)

	record := techan.NewTradingRecord()

	entryRule := techan.And(
		buyRuleOne{MACD, RSI, lowerBB, series},
		techan.PositionNewRule{},
	)

	exitRule := techan.And(
		sellRuleOne{MACD, RSI, upperBB, series},
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
		"MACD0":   MACD.Calculate(len(series.Candles) - 2).String(),
		"MACD1":   MACD.Calculate(len(series.Candles) - 1).String(),
		"RSI0":    RSI.Calculate(len(series.Candles) - 2).String(),
		"RSI1":    RSI.Calculate(len(series.Candles) - 1).String(),
		"upperBB": upperBB.Calculate(len(series.Candles) - 1).String(),
		"lowerBB": lowerBB.Calculate(len(series.Candles) - 1).String(),
	}

	return result, indicators
}
