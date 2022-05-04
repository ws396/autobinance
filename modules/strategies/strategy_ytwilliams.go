package strategies

import (
	"fmt"
	"log"

	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"github.com/ws396/autobinance/modules/techanext"
)

type buyRuleYTWilliams struct {
	MACD    techan.Indicator
	RSI     techan.Indicator
	lowerBB techan.Indicator
	series  *techan.TimeSeries
}

func (r buyRuleYTWilliams) IsSatisfied(index int, record *techan.TradingRecord) bool {
	// cinar/indicator
	/*
		var lowSlice, highSlice, closingSlice []float64
		williamsRlen := 21
		start := len(r.series.Candles) - williamsRlen //- 1

		for i := 0; i < williamsRlen; i++ {
			lowSlice = append(lowSlice, r.series.Candles[start+i].MinPrice.Float())
			highSlice = append(highSlice, r.series.Candles[start+i].MaxPrice.Float())
			closingSlice = append(closingSlice, r.series.Candles[start+i].ClosePrice.Float())
		}

		d := indicator.WilliamsR(lowSlice, highSlice, closingSlice) // This function actually has a hardcoded 14 period for some reason......
	*/
	d := techanext.NewWilliamsRIndicator(r.series, 21)

	l := len(r.series.Candles)
	fmt.Println(d.Calculate(l - 1))

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
	if c.LTE(r.series.LastCandle().ClosePrice) {
		return false
	}

	return true
}

type sellRuleYTWilliams struct {
	MACD    techan.Indicator
	RSI     techan.Indicator
	upperBB techan.Indicator
	series  *techan.TimeSeries
}

func (r sellRuleYTWilliams) IsSatisfied(index int, record *techan.TradingRecord) bool {
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
	if c.GT(r.series.LastCandle().ClosePrice) {
		return false
	}

	return true
}

func StrategyYTWilliams(symbol string, series *techan.TimeSeries, placedOrders *map[string]bool) (string, map[string]string) {
	closePrices := techan.NewClosePriceIndicator(series)

	MACD := techan.NewMACDIndicator(closePrices, 12, 26)
	RSI := techan.NewRelativeStrengthIndexIndicator(closePrices, 14)
	upperBB := techan.NewBollingerUpperBandIndicator(closePrices, 20, 1.7)
	lowerBB := techan.NewBollingerLowerBandIndicator(closePrices, 20, 1.8)

	record := techan.NewTradingRecord()

	entryRule := techan.And(
		buyRuleYTWilliams{MACD, RSI, lowerBB, series},
		techan.PositionNewRule{},
	)

	exitRule := techan.And(
		sellRuleYTWilliams{MACD, RSI, upperBB, series},
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

	if !(*placedOrders)[symbol] && result == "Sell" {
		log.Println("err: no buy has been done on this symbol to initiate sell")
		return "", nil
	} else if (*placedOrders)[symbol] && result == "Buy" {
		log.Println("err: this position is already bought")
		return "", nil
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
