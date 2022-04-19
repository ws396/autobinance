package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/adshao/go-binance/v2" // Might be better to eventually get rid of this dependency here
	"github.com/joho/godotenv"
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"github.com/ws396/autobinance/modules/binancew-sim"
)

var (
	timeframe         = 3
	buyAmount float64 = 100
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	var (
		apiKey          = os.Getenv("API_KEY")
		secretKey       = os.Getenv("SECRET_KEY")
		input           string
		strategyRunning = false
	)

	//binance.UseTestnet = true // Interestingly enough, this also applies to all other modules, where this package is used
	client := binancew.NewExtClient(apiKey, secretKey)

	fmt.Print(
		"1) Start strategy execution", "\n",
		"2) Check klines", "\n",
		"3) Check account", "\n",
		"4) Get EMA prognosis", "\n",
		"5) List trades", "\n",
	)

	for {
		fmt.Println("Your choice: ")
		fmt.Scanln(&input)

		switch input {
		case "1": // Keep in mind that with current approach the unsold assets will remain unsold with app relaunch
			if strategyRunning {
				fmt.Println("err: the strategy is already running")
				break
			}

			fmt.Println("Select the coin symbols for Strategy One (ex. BNBBUSD): ")
			fmt.Scanln(&input)

			exchangeInfo, err := client.NewExchangeInfoService().Symbol(input).Do(context.Background())
			if err != nil {
				fmt.Println(err)
				return
			}
			showJSON(exchangeInfo.Symbols[0].LotSizeFilter().MinQuantity)

			ticker := time.NewTicker(time.Duration(timeframe) * time.Minute)
			strategyRunning = true

			go func(input string) {
				for {
					select {
					case <-ticker.C:
						klines, err := client.GetKlines(input, timeframe)
						if err != nil {
							fmt.Println(err)
							break
						}

						series := getSeries(klines)
						decision := StrategyOne(series)

						switch decision {
						case "Buy":
							quantity := fmt.Sprintf("%f", buyAmount/series.LastCandle().ClosePrice.Float())
							price := series.LastCandle().ClosePrice.String()
							order := client.CreateOrder(input, quantity, price, binance.SideTypeBuy)
							fmt.Println(order)

							showJSON(client.GetCurrencies("BNB", "BUSD"))
						case "Sell":
							quantity := fmt.Sprintf("%f", buyAmount/series.LastCandle().ClosePrice.Float())
							price := series.LastCandle().ClosePrice.String()
							order := client.CreateOrder(input, quantity, price, binance.SideTypeSell)
							fmt.Println(order)

							showJSON(client.GetCurrencies("BNB", "BUSD"))
						}
					}
				}
			}(input)
		case "2":
			klines, err := client.GetKlines("LTCBTC", timeframe)
			if err != nil {
				fmt.Println(err)
				break
			}

			showJSON(klines)
		case "3":
			showJSON(client.GetAccount())
		case "4":
			fmt.Println("Select the coin symbols (ex. LTCBTC): ")
			fmt.Scanln(&input)

			klines, err := client.GetKlines(input, timeframe)
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

			showJSON(prices)
		default:
			fmt.Println("Invalid choice.")
		}
	}
}

func getSeries(klines []*binance.Kline) *techan.TimeSeries {
	series := techan.NewTimeSeries()

	for _, data := range klines {
		period := techan.NewTimePeriod(time.UnixMilli(data.OpenTime), time.Duration(timeframe)*time.Minute)

		candle := techan.NewCandle(period)
		candle.OpenPrice = big.NewFromString(data.Open)
		candle.ClosePrice = big.NewFromString(data.Close)
		candle.MaxPrice = big.NewFromString(data.High)
		candle.MinPrice = big.NewFromString(data.Low)
		candle.Volume = big.NewFromString(data.Volume)

		series.AddCandle(candle)
		//showJSON(data)
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
	if a1.LT(big.NewFromInt(0)) || a1.GTE(a0) {
		return false
	}

	b0 := r.RSI.Calculate(len - 2)
	b1 := r.RSI.Calculate(len - 1)
	if b1.LT(big.NewFromInt(66)) || b1.GTE(b0) {
		return false
	}

	c := r.upperBB.Calculate(len - 1)
	if c.GTE(r.series.LastCandle().ClosePrice) {
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

	message := fmt.Sprint(
		"----------", "\n",
		"MACD0: ", MACD.Calculate(len(series.Candles)-2), "\n",
		"MACD1: ", MACD.Calculate(len(series.Candles)-1), "\n",
		"RSI0: ", RSI.Calculate(len(series.Candles)-2), "\n",
		"RSI1: ", RSI.Calculate(len(series.Candles)-1), "\n",
		"upperBB: ", upperBB.Calculate(len(series.Candles)-1), "\n",
		"lowerBB: ", lowerBB.Calculate(len(series.Candles)-1), "\n",
		"Current price: ", series.LastCandle().ClosePrice.String(), "\n",
		"Time: ", time.Now().Format("02-01-2006 15:04:05"), "\n",
		result, "\n",
	)

	fmt.Println(message)

	file, err := os.OpenFile("./log.txt", os.O_WRONLY|os.O_APPEND, 0644)
	if errors.Is(err, os.ErrNotExist) {
		file, err = os.Create("./log.txt")
		if err != nil {
			log.Fatal(err)
		}
	}
	defer file.Close()

	//writer := bufio.NewWriter(file)
	_, err = file.WriteString(message)
	if err != nil {
		log.Fatalf("Got error while writing to a file. Err: %s", err.Error())
	}

	//writer.Flush()

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

func showJSON(data interface{}) {
	j, err := json.MarshalIndent(data, "", "    🐱") // 🐱🐱🐱🐱🐱🐱🐱🐱!!
	if err != nil {
		fmt.Println(err)
		//return
	}

	fmt.Println(string(j))
}
