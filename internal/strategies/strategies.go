package strategies

import (
	"github.com/sdcoffey/techan"
	"github.com/ws396/autobinance/internal/globals"
)

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

func RunStrategy(strategy string, series *techan.TimeSeries) (string, map[string]string, error) {
	if _, ok := StrategiesInfo[strategy]; !ok {
		return "", nil, globals.ErrWrongStrategyName
	}

	decision, indicators := StrategiesInfo[strategy].Handler(series)

	return decision, indicators, nil
}
