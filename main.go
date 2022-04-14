package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/joho/godotenv"
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
)

type binanceClientExt struct {
	*binance.Client
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	var (
		apiKey    = os.Getenv("API_KEY")
		secretKey = os.Getenv("SECRET_KEY")
	)
	binance.UseTestnet = true
	client := &binanceClientExt{binance.NewClient(apiKey, secretKey)}

	fmt.Println("1) THE Strategy\n2) Check klines\n3) Check account\n4) Get EMA prognosis\n5) List trades")

	var input string
	for {
		fmt.Println("Your choice: ")
		fmt.Scanln(&input)

		switch input {
		case "1":
			fmt.Println("Select the coin symbols (ex. LTCBTC): ")
			fmt.Scanln(&input)

			klines, err := client.getKlines(input)
			if err != nil {
				fmt.Println(err)
				break
			}

			series := getSeries(klines)
			fmt.Println(StrategyOne(series))
		case "2":
			klines, err := client.getKlines("LTCBTC")
			if err != nil {
				fmt.Println(err)
				break
			}

			showInConsole(klines)
		case "3":
			showInConsole(client.getAccount())
		case "4":
			fmt.Println("Select the coin symbols (ex. LTCBTC): ")
			fmt.Scanln(&input)

			klines, err := client.getKlines(input)
			if err != nil {
				fmt.Println(err)
				break
			}

			series := getSeries(klines)
			fmt.Println(StrategyExample(series))
		case "5":
			prices, err := client.NewListPricesService().Do(context.Background())
			if err != nil {
				fmt.Println(err)
				return
			}

			showInConsole(prices)
		default:
			fmt.Println("Invalid choice.")
		}
	}
}

func getSeries(klines []*binance.Kline) *techan.TimeSeries {
	series := techan.NewTimeSeries()

	for _, data := range klines {
		period := techan.NewTimePeriod(time.UnixMilli(data.OpenTime), time.Minute*5)

		candle := techan.NewCandle(period)
		candle.OpenPrice = big.NewFromString(data.Open)
		candle.ClosePrice = big.NewFromString(data.Close)
		candle.MaxPrice = big.NewFromString(data.High)
		candle.MinPrice = big.NewFromString(data.Low)
		candle.Volume = big.NewFromString(data.Volume)

		series.AddCandle(candle)
		showInConsole(data)
	}

	return series
}

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
	if c.GTE(r.series.LastCandle().ClosePrice) {
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
	if a1.LT(big.NewFromInt(0)) || a1.GTE(a0) {
		return false
	}

	b0 := r.RSI.Calculate(len - 2)
	b1 := r.RSI.Calculate(len - 1)
	if b1.LT(big.NewFromInt(66)) || b1.GTE(b0) {
		return false
	}

	c := r.upperBB.Calculate(len - 1)
	if c.LTE(r.series.LastCandle().ClosePrice) {
		return false
	}

	return true
}

func StrategyOne(series *techan.TimeSeries) string {
	closePrices := techan.NewClosePriceIndicator(series)

	MACD := techan.NewMACDIndicator(closePrices, 12, 26)
	RSI := techan.NewRelativeStrengthIndexIndicator(closePrices, 14)
	upperBB := techan.NewBollingerUpperBandIndicator(closePrices, 20, 2)
	lowerBB := techan.NewBollingerLowerBandIndicator(closePrices, 20, 2)

	fmt.Println(MACD.Calculate(len(series.Candles) - 2))
	fmt.Println(MACD.Calculate(len(series.Candles) - 1))
	fmt.Println(RSI.Calculate(len(series.Candles) - 2))
	fmt.Println(RSI.Calculate(len(series.Candles) - 1))
	fmt.Println(upperBB.Calculate(len(series.Candles) - 1))
	fmt.Println(lowerBB.Calculate(len(series.Candles) - 1))

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

	return result
}

func StrategyExample(series *techan.TimeSeries) bool {
	closePrices := techan.NewClosePriceIndicator(series)
	movingAverage := techan.NewEMAIndicator(closePrices, 10) // Windows depicts amount of periods taken

	// record trades on this object
	record := techan.NewTradingRecord()

	entryConstant := techan.NewConstantIndicator(30)
	exitConstant := techan.NewConstantIndicator(10)

	entryRule := techan.And(
		techan.NewCrossUpIndicatorRule(entryConstant, movingAverage),
		techan.PositionNewRule{},
	) // Is satisfied when the price ema moves above 30 and the current position is new

	exitRule := techan.And(
		techan.NewCrossDownIndicatorRule(movingAverage, exitConstant),
		techan.PositionOpenRule{},
	) // Is satisfied when the price ema moves below 10 and the current position is open

	strategy := techan.RuleStrategy{
		UnstablePeriod: 10,
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}

	fmt.Println(movingAverage.Calculate(len(series.Candles) - 1)) // Calculate argument is index of series array
	return strategy.ShouldEnter(20, record)                       // Index should be > UnstablePeriod for getting true!
}

func (client binanceClientExt) getPrices() []*binance.SymbolPrice {
	prices, err := client.NewListPricesService().Symbol("LTCBTC").
		Do(context.Background())
	if err != nil {
		fmt.Println(err)
		//return
	}

	return prices
}

func (client binanceClientExt) getOrders() []*binance.Order {
	orders, err := client.NewListOrdersService().Symbol("LTCBTC").
		Do(context.Background(), binance.WithRecvWindow(10000))
	if err != nil {
		fmt.Println(err)
		//return
	}

	return orders
}

func (client binanceClientExt) getKlines(symbol string) ([]*binance.Kline, error) {
	klines, err := client.NewKlinesService().Symbol(symbol).
		Interval("5m").Do(context.Background())
	if err != nil {
		return nil, err
	}

	return klines, nil
}

/**/
func (client binanceClientExt) getAccount() *binance.Account {
	account, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		fmt.Println(err)
		//return
	}

	return account
}

func showInConsole(data interface{}) {
	j, err := json.MarshalIndent(data, "", "    🐱") // 🐱🐱🐱🐱🐱🐱🐱🐱!!
	if err != nil {
		fmt.Println(err)
		//return
	}

	fmt.Println(string(j))
}
