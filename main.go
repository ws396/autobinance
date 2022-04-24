package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2" // Might be better to eventually get rid of this dependency here
	"github.com/joho/godotenv"
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"github.com/ws396/autobinance/modules/binancew-sim"
	"github.com/ws396/autobinance/modules/output"
	"github.com/ws396/autobinance/modules/strategies"
	"github.com/ws396/autobinance/modules/util"
)

var (
	timeframe         = 3
	buyAmount float64 = 50
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
		"6) Test things", "\n",
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

			fmt.Println("Select the coin symbols for Strategy One (ex. BNBBUSD, BTCBUSD): ")
			fmt.Scanln(&input)
			symbols := strings.Split(input, ",")
			for i := range symbols {
				symbols[i] = strings.Trim(symbols[i], " ") // fmt.Scanln doesn't work well with whitespaces anyway (use bufio.Scanner?)
				client.PlacedOrders[symbols[i]] = false
			}

			ticker := time.NewTicker(time.Duration(timeframe) * time.Minute)
			fmt.Println("Strategy execution started (you can still do other actions)")
			strategyRunning = true

			go func(symbols []string) {
				for {
					select {
					case <-ticker.C:
						for _, symbol := range symbols {
							klines, err := client.GetKlines(symbol, timeframe)
							if err != nil {
								fmt.Println(err)
								break
							}

							series := getSeries(klines)
							//decision := strategies.StrategyOne(symbol, series, &client.PlacedOrders)
							decision := strategies.StrategyTwo(symbol, series, &client.PlacedOrders)

							switch decision {
							case "Buy":
								quantity := fmt.Sprintf("%f", buyAmount/series.LastCandle().ClosePrice.Float())
								price := series.LastCandle().ClosePrice.String()
								order := client.CreateOrder(symbol, quantity, price, binance.SideTypeBuy)
								fmt.Println(order)

								util.ShowJSON(client.GetCurrencies("BNB", "BUSD"))
							case "Sell":
								quantity := fmt.Sprintf("%f", buyAmount/series.LastCandle().ClosePrice.Float())
								price := series.LastCandle().ClosePrice.String()
								order := client.CreateOrder(symbol, quantity, price, binance.SideTypeSell)
								fmt.Println(order)

								util.ShowJSON(client.GetCurrencies("BNB", "BUSD"))
							}
						}
					}
				}
			}(symbols)
		case "2":
			klines, err := client.GetKlines("LTCBTC", timeframe)
			if err != nil {
				fmt.Println(err)
				break
			}

			util.ShowJSON(klines)
		case "3":
			util.ShowJSON(client.GetAccount())
		case "4":
			fmt.Println("Select the coin symbols (ex. LTCBTC): ")
			fmt.Scanln(&input)

			klines, err := client.GetKlines(input, timeframe)
			if err != nil {
				fmt.Println(err)
				break
			}

			series := getSeries(klines)
			fmt.Println(strategies.StrategyExample(series))
		case "5":
			prices, err := client.NewListPricesService().Do(context.Background())
			if err != nil {
				fmt.Println(err)
				return
			}

			util.ShowJSON(prices)
		case "6":
			writer := output.NewWriterCreator().CreateWriter(output.Excel)
			writer.WriteToLog(map[string]string{
				"MACD0": "MACD051235",
				"MACD1": "MACD512351231",
			})
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
		//util.ShowJSON(data)
	}

	return series
}
