package strategies

var Datakeys = map[string]*[]string{
	"one": {
		"MACD0",
		"MACD1",
		"RSI0",
		"RSI1",
		"upperBB",
		"lowerBB",
		"Current price",
		"Time",
		"Symbol",
		"Decision",
		"Strategy",
	},
	"two": {
		"MACD0",
		"MACD1",
		"stochRSI0",
		"stochRSI1",
		"upperBB",
		"lowerBB",
		"Current price",
		"Time",
		"Symbol",
		"Decision",
		"Strategy",
	},
	"ytwilliams": {
		"EMA0",
		"EMA1",
		"MACDH0",
		"MACDH1",
		"WilliamsR0",
		"WilliamsR1",
		"WilliamsREMA0",
		"WilliamsREMA1",
		"Current price",
		"Time",
		"Symbol",
		"Decision",
		"Strategy",
	},
}

// Guess I'll keep them here? Also need to make double sure if these pointers even make sense 🐱!
