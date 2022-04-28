package strategies

import (
	"fmt"
	"time"

	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"github.com/ws396/autobinance/modules/output"
	"github.com/ws396/autobinance/modules/techanext"
)

type buyRuleTwo struct {
	MACD     techan.Indicator
	stochRSI techan.Indicator
	lowerBB  techan.Indicator
	series   *techan.TimeSeries
}

func (r buyRuleTwo) IsSatisfied(index int, record *techan.TradingRecord) bool {
	len := len(r.series.Candles)

	//a0 := r.MACD.Calculate(len - 2)
	a1 := r.MACD.Calculate(len - 1)
	if a1.GT(big.NewFromInt(0)) /*|| a1.LTE(a0)*/ {
		return false
	}

	b0 := r.stochRSI.Calculate(len - 2)
	b1 := r.stochRSI.Calculate(len - 1)
	if b1.GT(big.NewFromInt(20)) || b1.LTE(b0) {
		return false
	}

	c := r.lowerBB.Calculate(len - 1)
	if c.LTE(r.series.LastCandle().ClosePrice) {
		return false
	}

	return true
}

type sellRuleTwo struct {
	MACD     techan.Indicator
	stochRSI techan.Indicator
	upperBB  techan.Indicator
	series   *techan.TimeSeries
}

func (r sellRuleTwo) IsSatisfied(index int, record *techan.TradingRecord) bool {
	len := len(r.series.Candles)

	//a0 := r.MACD.Calculate(len - 2)
	a1 := r.MACD.Calculate(len - 1)
	if a1.LT(big.NewFromInt(0)) /*|| a1.GT(a0)*/ {
		return false
	}

	b0 := r.stochRSI.Calculate(len - 2)
	b1 := r.stochRSI.Calculate(len - 1)
	if b1.LT(big.NewFromInt(80)) || b1.GT(b0) {
		return false
	}

	c := r.upperBB.Calculate(len - 1)
	if c.GT(r.series.LastCandle().ClosePrice) {
		return false
	}

	return true
}

func StrategyTwo(symbol string, series *techan.TimeSeries, placedOrders *map[string]bool) string {
	closePrices := techan.NewClosePriceIndicator(series)

	MACD := techan.NewMACDIndicator(closePrices, 12, 26)
	stochRSI := techanext.NewFastStochasticRSIIndicator(techanext.NewStochasticRSIIndicator(closePrices, 14), 3)
	upperBB := techan.NewBollingerUpperBandIndicator(closePrices, 20, 1.5)
	lowerBB := techan.NewBollingerLowerBandIndicator(closePrices, 20, 1.8)

	record := techan.NewTradingRecord()

	entryRule := techan.And(
		buyRuleTwo{MACD, stochRSI, lowerBB, series},
		techan.PositionNewRule{},
	)

	exitRule := techan.And(
		sellRuleTwo{MACD, stochRSI, upperBB, series},
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
		fmt.Println("err: no buy has been done on this symbol to initiate sell")
		return ""
	} else if (*placedOrders)[symbol] && result == "Buy" {
		fmt.Println("err: this position is already bought")
		return ""
	}

	data := map[string]string{
		"MACD0":         MACD.Calculate(len(series.Candles) - 2).String(),
		"MACD1":         MACD.Calculate(len(series.Candles) - 1).String(),
		"stochRSI0":     stochRSI.Calculate(len(series.Candles) - 2).String(),
		"stochRSI1":     stochRSI.Calculate(len(series.Candles) - 1).String(),
		"upperBB":       upperBB.Calculate(len(series.Candles) - 1).String(),
		"lowerBB":       lowerBB.Calculate(len(series.Candles) - 1).String(),
		"Current price": series.LastCandle().ClosePrice.String(),
		"Time":          time.Now().Format("02-01-2006 15:04:05"),
		"Symbol":        symbol,
		"Decision":      result,
		"Strategy":      "two",
	}

	message := fmt.Sprint(
		"----------", "\n",
	)
	for k, v := range data {
		message += fmt.Sprint(k, ": ", v, "\n")
	}

	fmt.Println(message)

	if result != "Hold" {
		writerType := output.Txt
		writer := output.NewWriterCreator().CreateWriter(writerType)
		writer.WriteToLog(data)
	}

	return result
}
