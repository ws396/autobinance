package globals

import "github.com/sdcoffey/techan"

const (
	Timeframe int    = 1
	Buy       string = "BUY"
	Sell      string = "SELL"
	Hold      string = "HOLD"
)

// Could also place everything below in strategies.go
var (
	StrategiesInfo = map[string]StrategyInfo{}
)

type StrategyInfo struct {
	Handler  func(*techan.TimeSeries) (string, map[string]string)
	Datakeys []string
}

// Add error handling?
func AddStrategyInfo(strategy string, handler func(*techan.TimeSeries) (string, map[string]string), datakeys []string) {
	datakeys = append(datakeys, "Current price",
		"Time",
		"Symbol",
		"Decision",
		"Strategy",
	)

	StrategiesInfo[strategy] = StrategyInfo{handler, datakeys}
}
