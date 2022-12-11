package globals

import "errors"

const (
	BuyAmount float64 = 50
	Timeframe uint    = 1
	Buy       string  = "BUY"
	Sell      string  = "SELL"
	Hold      string  = "HOLD"
)

var (
	SimulationMode = true

	ErrSymbolsNotFound       = errors.New("err: no selected symbols found")
	ErrStrategiesNotFound    = errors.New("err: no selected strategies found")
	ErrWrongStrategyName     = errors.New("err: entered wrong strategy names")
	ErrWrongSymbol           = errors.New("err: entered wrong symbols")
	ErrWrongArgumentAmount   = errors.New("err: wrong amount of arguments")
	ErrWrongDateOrder        = errors.New("err: expected second date to be later than first")
	ErrTradingAlreadyRunning = errors.New("err: the trading is already running")
	ErrCouldNotDownloadFile  = errors.New("err: could not download file")
	ErrNotInSimulationMode   = errors.New("err: only available in simulation mode")
)
