package globals

const (
	Timeframe int    = 1
	Buy       string = "BUY"
	Sell      string = "SELL"
	Hold      string = "HOLD"
)

var (
	Datakeys = map[string][]string{}
)

// Add error handling?
func AddStrategyDatakeys(strategy string, datakeys []string) {
	Datakeys[strategy] = datakeys
	Datakeys[strategy] = append(Datakeys[strategy], "Current price",
		"Time",
		"Symbol",
		"Decision",
		"Strategy",
	)
}
