package testutil

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"github.com/stretchr/testify/assert"
)

var CandleIndex int
var MockedTimeSeries = MockTimeSeriesFl(
	64.75, 63.79, 63.73,
	63.73, 63.55, 63.19,
	63.91, 63.85, 62.95,
	63.37, 61.33, 61.51)

func dump(indicator techan.Indicator) (values []float64) {
	precision := 4.0
	m := math.Pow(10, precision)

	defer func() {
		recover()
	}()

	var index int
	for {
		values = append(values, math.Round(indicator.Calculate(index).Float()*m)/m)
		index++
	}
}

func IndicatorEquals(t *testing.T, expected []float64, indicator techan.Indicator) {
	actualValues := dump(indicator)
	assert.EqualValues(t, expected, actualValues)
}

func MockTimeSeries(values ...string) *techan.TimeSeries {
	ts := techan.NewTimeSeries()
	for _, val := range values {
		candle := techan.NewCandle(techan.NewTimePeriod(time.Unix(int64(CandleIndex), 0), time.Second))
		candle.OpenPrice = big.NewFromString(val)
		candle.ClosePrice = big.NewFromString(val)
		candle.MaxPrice = big.NewFromString(val).Add(big.ONE)
		candle.MinPrice = big.NewFromString(val).Sub(big.ONE)
		candle.Volume = big.NewFromString(val)

		ts.AddCandle(candle)

		CandleIndex++
	}

	return ts
}

func MockTimeSeriesFl(values ...float64) *techan.TimeSeries {
	strVals := make([]string, len(values))

	for i, val := range values {
		strVals[i] = fmt.Sprint(val)
	}

	return MockTimeSeries(strVals...)
}
