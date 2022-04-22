package strategies

import (
	"fmt"

	"github.com/sdcoffey/techan"
)

func StrategyExample(series *techan.TimeSeries) bool {
	closePrices := techan.NewClosePriceIndicator(series)
	movingAverage := techan.NewEMAIndicator(closePrices, 10) // Windows depicts amount of periods taken

	// record trades on this object
	record := techan.NewTradingRecord()

	entryConstant := techan.NewConstantIndicator(30)
	exitConstant := techan.NewConstantIndicator(10)

	entryRule := techan.And(
		techan.NewCrossUpIndicatorRule(entryConstant, movingAverage),
		techan.PositionNewRule{},
	) // Is satisfied when the price ema moves above 30 and the current position is new

	exitRule := techan.And(
		techan.NewCrossDownIndicatorRule(movingAverage, exitConstant),
		techan.PositionOpenRule{},
	) // Is satisfied when the price ema moves below 10 and the current position is open

	strategy := techan.RuleStrategy{
		UnstablePeriod: 10,
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}

	fmt.Println(movingAverage.Calculate(len(series.Candles) - 1)) // Calculate argument is index of series array
	return strategy.ShouldEnter(20, record)                       // Index should be > UnstablePeriod for getting true!
}
