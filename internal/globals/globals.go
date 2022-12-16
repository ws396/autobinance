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
	SimulationMode = true

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

	ErrSymbolsNotFound       = errors.New("err: no selected symbols found")
	ErrStrategiesNotFound    = errors.New("err: no selected strategies found")
	ErrWrongStrategyName     = errors.New("err: entered wrong strategy names")
	ErrWrongSymbol           = errors.New("err: entered wrong symbols")
	ErrWrongArgumentAmount   = errors.New("err: wrong amount of arguments")
	ErrWrongDateOrder        = errors.New("err: expected second date to be later than first")
	ErrTradingAlreadyRunning = errors.New("err: the trading is already running")
	ErrCouldNotDownloadFile  = errors.New("err: could not download file")
	ErrNotInSimulationMode   = errors.New("err: only available in simulation mode")
	ErrOrderNotFound         = errors.New("err: order not found")
	ErrWriterNotFound        = errors.New("err: writer not found")
	ErrEmptyOrderList        = errors.New("err: order list is empty")
)
