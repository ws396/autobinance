package globals

import (
	"errors"
	"time"
)

const (
	BuyAmount float64 = 50
	Timeframe string  = "1m"
	Buy       string  = "BUY"
	Sell      string  = "SELL"
	Hold      string  = "HOLD"
)

var (
	SimulationMode bool = true

	BacktestDataBaseURL string = "https://data.binance.vision/"
	BacktestDataDir     string = "internal/backtest/data/"
	TestDataDir         string = "internal/testutil/data/"

	Durations = map[string]time.Duration{
		"1s":  time.Second,
		"1m":  time.Minute,
		"3m":  3 * time.Minute,
		"5m":  5 * time.Minute,
		"15m": 15 * time.Minute,
		"30m": 30 * time.Minute,
		"1h":  time.Hour,
		"2h":  2 * time.Hour,
		"4h":  4 * time.Hour,
		"6h":  6 * time.Hour,
		"8h":  8 * time.Hour,
		"12h": 12 * time.Hour,
		"1d":  24 * time.Hour,
	}

	ErrCouldNotDownloadFile  = errors.New("err: could not download file")
	ErrEmptyOrderList        = errors.New("err: order list is empty")
	ErrNotInSimulationMode   = errors.New("err: only available in simulation mode")
	ErrOrderNotFound         = errors.New("err: order not found")
	ErrStrategiesNotFound    = errors.New("err: no selected strategies found")
	ErrSymbolsNotFound       = errors.New("err: no selected symbols found")
	ErrTradingAlreadyRunning = errors.New("err: trading is already running")
	ErrTradingNotRunning     = errors.New("err: trading is not running")
	ErrWriterNotFound        = errors.New("err: writer not found")
	ErrWrongArgumentAmount   = errors.New("err: wrong amount of arguments")
	ErrWrongDateOrder        = errors.New("err: expected second date to be later than first")
	ErrWrongStrategyName     = errors.New("err: entered wrong strategy names")
	ErrWrongSymbol           = errors.New("err: entered wrong symbols")
)
