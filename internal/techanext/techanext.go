package techanext

import (
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
)

func GetSeries(klines []*binance.Kline, timeframe time.Duration) *techan.TimeSeries {
	series := techan.NewTimeSeries()

	for _, data := range klines {
		period := techan.NewTimePeriod(time.UnixMilli(data.OpenTime), timeframe)

		candle := techan.NewCandle(period)
		candle.OpenPrice = big.NewFromString(data.Open)
		candle.ClosePrice = big.NewFromString(data.Close)
		candle.MaxPrice = big.NewFromString(data.High)
		candle.MinPrice = big.NewFromString(data.Low)
		candle.Volume = big.NewFromString(data.Volume)

		series.AddCandle(candle)
	}

	return series
}
