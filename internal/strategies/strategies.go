package strategies

import "github.com/sdcoffey/techan"

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
		"Created at",
		"Symbol",
		"Decision",
		"Strategy",
		"Successful",
	)

	StrategiesInfo[strategy] = StrategyInfo{handler, datakeys}
}
