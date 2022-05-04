package techanext

// Taken from https://github.com/sdcoffey/techan/pull/37/commits/b00fdf455d24569caef9784369f293413662e7a1

import (
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
)

type williamsRIndicator struct {
	series    *techan.TimeSeries
	timeframe int
}

func NewWilliamsRIndicator(series *techan.TimeSeries, timeframe int) techan.Indicator {
	return williamsRIndicator{
		series:    series,
		timeframe: timeframe,
	}
}

func (wri williamsRIndicator) Calculate(index int) big.Decimal {
	/*
		var lowPrices, highPrices, closingPrices []float64
		start := len(wri.series.Candles) - wri.timeframe //- 1

		for i := 0; i < wri.timeframe; i++ {
			lowPrices = append(lowPrices, wri.series.Candles[start+i].MinPrice.Float())
			highPrices = append(highPrices, wri.series.Candles[start+i].MaxPrice.Float())
			closingPrices = append(closingPrices, wri.series.Candles[start+i].ClosePrice.Float())
		}

		highestHigh := indicator.Max(wri.timeframe, highPrices)
		lowestLow := indicator.Min(wri.timeframe, lowPrices)

		result := make([]float64, len(closingPrices))

		for i := 0; i < len(closingPrices); i++ {
			result[i] = (highestHigh[i] - closingPrices[i]) / (highestHigh[i] - lowestLow[i]) * float64(-100)
		}
	*/
	closingPrices := techan.NewClosePriceIndicator(wri.series)
	highestHigh := techan.NewMaximumValueIndicator(techan.NewHighPriceIndicator(wri.series), wri.timeframe)
	lowestLow := techan.NewMinimumValueIndicator(techan.NewLowPriceIndicator(wri.series), wri.timeframe)

	result := highestHigh.Calculate(index).Sub(closingPrices.Calculate(index)).
		Div(highestHigh.Calculate(index).Sub(lowestLow.Calculate(index))).Mul(big.NewDecimal(-100)) // Oh dear

	return result
}
