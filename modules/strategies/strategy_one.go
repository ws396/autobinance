package strategies

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"github.com/ws396/autobinance/modules/db"
	"github.com/ws396/autobinance/modules/orders"
	"github.com/ws396/autobinance/modules/output"
)

var DataKeysOne = []string{
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
} // I don't really like this, but I'm not sure if there is a better way :(

type buyRuleOne struct {
	MACD    techan.Indicator
	RSI     techan.Indicator
	lowerBB techan.Indicator
	series  *techan.TimeSeries
}

func (r buyRuleOne) IsSatisfied(index int, record *techan.TradingRecord) bool {
	len := len(r.series.Candles)

	a0 := r.MACD.Calculate(len - 2)
	a1 := r.MACD.Calculate(len - 1)
	if a1.GT(big.NewFromInt(0)) || a1.LTE(a0) {
		return false
	}

	b0 := r.RSI.Calculate(len - 2)
	b1 := r.RSI.Calculate(len - 1)
	if b1.GT(big.NewFromInt(33)) || b1.LTE(b0) {
		return false
	}

	c := r.lowerBB.Calculate(len - 1)
	if c.LTE(r.series.LastCandle().ClosePrice) {
		return false
	}

	return true
}

type sellRuleOne struct {
	MACD    techan.Indicator
	RSI     techan.Indicator
	upperBB techan.Indicator
	series  *techan.TimeSeries
}

func (r sellRuleOne) IsSatisfied(index int, record *techan.TradingRecord) bool {
	len := len(r.series.Candles)

	a0 := r.MACD.Calculate(len - 2)
	a1 := r.MACD.Calculate(len - 1)
	if a1.LT(big.NewFromInt(0)) || a1.GT(a0) {
		return false
	}

	b0 := r.RSI.Calculate(len - 2)
	b1 := r.RSI.Calculate(len - 1)
	if b1.LT(big.NewFromInt(66)) || b1.GT(b0) {
		return false
	}

	c := r.upperBB.Calculate(len - 1)
	if c.GT(r.series.LastCandle().ClosePrice) {
		return false
	}

	return true
}

func StrategyOne(symbol string, series *techan.TimeSeries, placedOrders *map[string]bool) string {
	closePrices := techan.NewClosePriceIndicator(series)

	MACD := techan.NewMACDIndicator(closePrices, 12, 26)
	RSI := techan.NewRelativeStrengthIndexIndicator(closePrices, 14)
	upperBB := techan.NewBollingerUpperBandIndicator(closePrices, 20, 1.7)
	lowerBB := techan.NewBollingerLowerBandIndicator(closePrices, 20, 1.8)

	record := techan.NewTradingRecord()

	entryRule := techan.And(
		buyRuleOne{MACD, RSI, lowerBB, series},
		techan.PositionNewRule{},
	)

	exitRule := techan.And(
		sellRuleOne{MACD, RSI, upperBB, series},
		techan.PositionOpenRule{},
	)

	strategy := techan.RuleStrategy{
		UnstablePeriod: 10,
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}

	result := "Hold"
	if strategy.ShouldEnter(20, record) {
		result = "Buy"
	} else if strategy.ShouldExit(20, record) {
		result = "Sell"
	}

	if !(*placedOrders)[symbol] && result == "Sell" {
		log.Println("err: no buy has been done on this symbol to initiate sell")
		return ""
	} else if (*placedOrders)[symbol] && result == "Buy" {
		log.Println("err: this position is already bought")
		return ""
	}

	data := map[string]string{
		"MACD0":         MACD.Calculate(len(series.Candles) - 2).String(),
		"MACD1":         MACD.Calculate(len(series.Candles) - 1).String(),
		"RSI0":          RSI.Calculate(len(series.Candles) - 2).String(),
		"RSI1":          RSI.Calculate(len(series.Candles) - 1).String(),
		"upperBB":       upperBB.Calculate(len(series.Candles) - 1).String(),
		"lowerBB":       lowerBB.Calculate(len(series.Candles) - 1).String(),
		"Current price": series.LastCandle().ClosePrice.String(),
		"Time":          time.Now().Format("02-01-2006 15:04:05"),
		"Symbol":        symbol,
		"Decision":      result,
		"Strategy":      "one",
	}

	message := fmt.Sprint(
		"----------", "\n",
	)
	for _, v := range DataKeysOne {
		message += fmt.Sprint(v, ": ", data[v], "\n")
	}

	fmt.Println(message)

	indicators, err := json.Marshal(map[string]string{
		"MACD0":   data["MACD0"],
		"MACD1":   data["MACD1"],
		"RSI0":    data["RSI0"],
		"RSI1":    data["RSI1"],
		"upperBB": data["upperBB"],
		"lowerBB": data["lowerBB"],
	})
	if err != nil {
		log.Panicln(err)
	}

	r := db.Client.Table("orders").Create(&orders.Order{
		Symbol:     data["Symbol"],
		Strategy:   "one",
		Decision:   data["Decision"],
		Price:      series.LastCandle().ClosePrice.Float(),
		Indicators: string(indicators),
	})
	if r.Error != nil {
		log.Panicln(r.Error)
	}

	if result != "Hold" {
		writerType := output.Txt
		writer := output.NewWriterCreator().CreateWriter(writerType)
		writer.WriteToLog(data, &DataKeysOne)
	}

	return result
}
