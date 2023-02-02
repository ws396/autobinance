package trader

import (
	"reflect"
	"testing"
	"time"

	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"github.com/ws396/autobinance/internal/binancew"
	"github.com/ws396/autobinance/internal/globals"
	"github.com/ws396/autobinance/internal/storage"
)

func TestTrade(t *testing.T) {
	series := getMockSeries()
	trader, err := setupMockTrader()
	if err != nil {
		t.Errorf("failed to setup mock trader, %v", err)
	}

	t.Run("successfully orders buy", func(t *testing.T) {
		got, err := trader.Trade(
			trader.Settings["selected_strategies"].ValueArr[0],
			trader.Settings["selected_symbols"].ValueArr[0],
			series,
		)
		if err != nil {
			t.Errorf("failed to attempt trade, %v", err)
		}

		want := &storage.Order{
			ID:         0,
			Strategy:   "example",
			Symbol:     "LTCBTC",
			Decision:   globals.Buy,
			Quantity:   5,
			Price:      10,
			Indicators: map[string]string{"SMA0": "5", "SMA1": "5.5"},
			Timeframe:  "1m",
			Successful: true,
			CreatedAt:  got.CreatedAt,
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("created wrong order, got %v want %v", got, want)
		}
	})

	t.Run("successfully holds based on current position", func(t *testing.T) {
		period := techan.NewTimePeriod(time.Unix(52*60, 0), time.Minute)
		candle := techan.NewCandle(period)
		candle.OpenPrice = big.NewDecimal(10)
		candle.ClosePrice = big.NewDecimal(10)
		candle.MaxPrice = big.NewDecimal(10)
		candle.MinPrice = big.NewDecimal(10)
		candle.Volume = big.NewDecimal(50)
		series.AddCandle(candle)

		got, err := trader.Trade(
			trader.Settings["selected_strategies"].ValueArr[0],
			trader.Settings["selected_symbols"].ValueArr[0],
			series,
		)
		if err != nil {
			t.Errorf("failed to attempt trade, %v", err)
		}

		want := &storage.Order{
			ID:         0,
			Strategy:   "example",
			Symbol:     "LTCBTC",
			Decision:   globals.Buy,
			Quantity:   0,
			Price:      0,
			Indicators: map[string]string{"SMA0": "5", "SMA1": "6"},
			Timeframe:  "1m",
			Successful: false,
			CreatedAt:  got.CreatedAt,
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("created wrong order, got %v want %v", got, want)
		}
	})

	t.Run("successfully holds based on strategy decision", func(t *testing.T) {
		period := techan.NewTimePeriod(time.Unix(53*60, 0), time.Minute)
		candle := techan.NewCandle(period)
		candle.OpenPrice = big.NewDecimal(5)
		candle.ClosePrice = big.NewDecimal(5)
		candle.MaxPrice = big.NewDecimal(5)
		candle.MinPrice = big.NewDecimal(5)
		candle.Volume = big.NewDecimal(50)
		series.AddCandle(candle)

		got, err := trader.Trade(
			trader.Settings["selected_strategies"].ValueArr[0],
			trader.Settings["selected_symbols"].ValueArr[0],
			series,
		)
		if err != nil {
			t.Errorf("failed to attempt trade, %v", err)
		}

		want := &storage.Order{
			ID:         0,
			Strategy:   "example",
			Symbol:     "LTCBTC",
			Decision:   globals.Hold,
			Quantity:   0,
			Price:      0,
			Indicators: map[string]string{"SMA0": "5.5", "SMA1": "6"},
			Timeframe:  "1m",
			Successful: false,
			CreatedAt:  got.CreatedAt,
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("created wrong order, got %v want %v", got, want)
		}
	})

	t.Run("successfully orders sell", func(t *testing.T) {
		period := techan.NewTimePeriod(time.Unix(54*60, 0), time.Minute)
		candle := techan.NewCandle(period)
		candle.OpenPrice = big.NewDecimal(1)
		candle.ClosePrice = big.NewDecimal(1)
		candle.MaxPrice = big.NewDecimal(1)
		candle.MinPrice = big.NewDecimal(1)
		candle.Volume = big.NewDecimal(1)
		series.AddCandle(candle)

		got, err := trader.Trade(
			trader.Settings["selected_strategies"].ValueArr[0],
			trader.Settings["selected_symbols"].ValueArr[0],
			series,
		)
		if err != nil {
			t.Errorf("failed to attempt trade, %v", err)
		}

		want := &storage.Order{
			ID:         0,
			Strategy:   "example",
			Symbol:     "LTCBTC",
			Decision:   globals.Sell,
			Quantity:   5,
			Price:      1,
			Indicators: map[string]string{"SMA0": "6", "SMA1": "5.6"},
			Timeframe:  "1m",
			Successful: true,
			CreatedAt:  got.CreatedAt,
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("created wrong order, got %v want %v", got, want)
		}
	})
}

func BenchmarkTrade(b *testing.B) {
	series := getMockSeries()
	trader, err := setupMockTrader()
	if err != nil {
		b.Errorf("failed to setup mock trader, %v", err)
	}

	b.Run("orders buy 10000 times", func(b *testing.B) {
		for i := 0; i < 10000; i++ {
			_, err := trader.Trade(
				trader.Settings["selected_strategies"].ValueArr[0],
				trader.Settings["selected_symbols"].ValueArr[0],
				series,
			)
			if err != nil {
				b.Errorf("failed to attempt trade, %v", err)
			}
		}
	})
}

func setupMockTrader() (*Trader, error) {
	storageClient := storage.NewInMemoryClient()
	exchangeClient := binancew.NewExtClientSim("", "")
	settings := map[string]storage.Setting{
		"selected_symbols": {
			Name:     "selected_symbols",
			Value:    "LTCBTC",
			ValueArr: []string{"LTCBTC"},
		},
		"selected_strategies": {
			Name:     "selected_strategies",
			Value:    "example",
			ValueArr: []string{"example"},
		},
	}
	tickerChan := make(chan time.Time)

	return &Trader{
		StorageClient:  storageClient,
		ExchangeClient: exchangeClient,
		Settings:       settings,
		TickerChan:     tickerChan,
	}, nil
}

func getMockSeries() *techan.TimeSeries {
	series := techan.NewTimeSeries()

	for i := 0; i < 50; i++ {
		period := techan.NewTimePeriod(time.Unix(int64(i)*60, 0), time.Minute)
		candle := techan.NewCandle(period)
		candle.OpenPrice = big.NewFromInt(5)
		candle.ClosePrice = big.NewFromInt(5)
		candle.MaxPrice = big.NewFromInt(5)
		candle.MinPrice = big.NewFromInt(5)
		candle.Volume = big.NewFromInt(5)
		series.AddCandle(candle)
	}

	period := techan.NewTimePeriod(time.Unix(51*60, 0), time.Minute)
	candle := techan.NewCandle(period)
	candle.OpenPrice = big.NewDecimal(10)
	candle.ClosePrice = big.NewDecimal(10)
	candle.MaxPrice = big.NewDecimal(10)
	candle.MinPrice = big.NewDecimal(10)
	candle.Volume = big.NewDecimal(50)
	series.AddCandle(candle)

	return series
}
