package main

import (
	"context"
	"encoding/json"
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
	"github.com/ws396/autobinance/modules/db"
	"github.com/ws396/autobinance/modules/orders"
	"github.com/ws396/autobinance/modules/output"
	"github.com/ws396/autobinance/modules/settings"
	"github.com/ws396/autobinance/modules/strategies"
	"github.com/ws396/autobinance/modules/util"
)

var (
	timeframe int     = 1
	buyAmount float64 = 50
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Panicln("Error loading .env file")
	}

	db.ConnectDB()
	strategies.AutoMigrateAnalyses()
	settings.AutoMigrateSettings()
	orders.AutoMigrateOrders()

	strategies.Timeframe = &timeframe

	var (
		apiKey              = os.Getenv("API_KEY")
		secretKey           = os.Getenv("SECRET_KEY")
		input               string
		strategyRunning     = false
		availableStrategies = map[string]func(string, *techan.TimeSeries) (string, map[string]string){
			"one":        strategies.StrategyOne,
			"two":        strategies.StrategyTwo,
			"ytwilliams": strategies.StrategyYTWilliams,
		}
	)

	//binance.UseTestnet = true // Interestingly enough, this also applies to all other modules, where this package is used
	client := binancew.NewExtClient(apiKey, secretKey)

	fmt.Print(
		"1) Start trading execution", "\n",
		"2) Set strategies", "\n",
		"3) Set trade symbols", "\n",
		"4) Check klines", "\n",
		"5) Check account", "\n",
		"6) List trades", "\n",
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

			var foundStrategies settings.Setting
			r := db.Client.Table("settings").First(&foundStrategies, "name = ?", "selected_strategies")
			if r.Error != nil && !r.RecordNotFound() {
				log.Panicln(r.Error)
			}
			if r.RecordNotFound() {
				fmt.Println("Please specify the strategies first")
			}

			selectedStrategies := strings.Split(foundStrategies.Value, ",")

			var foundSymbols settings.Setting
			r = db.Client.Table("settings").First(&foundSymbols, "name = ?", "selected_symbols")
			if r.Error != nil && !r.RecordNotFound() {
				log.Panicln(r.Error)
			}
			if r.RecordNotFound() {
				fmt.Println("Please specify the settings first")
			}

			selectedSymbols := strings.Split(foundSymbols.Value, ",")

			ticker := time.NewTicker(time.Duration(timeframe) * time.Minute)
			fmt.Println("Strategy execution started (you can still do other actions)")
			strategyRunning = true

			go func(selectedSymbols []string) {
				for {
					select {
					case <-ticker.C:
					strategiesLoop:
						for _, strategy := range selectedStrategies {
							for _, symbol := range selectedSymbols {
								klines, err := client.GetKlines(symbol, timeframe)
								if err != nil {
									log.Println(err)
									break strategiesLoop
								}

								series := getSeries(klines)
								decision, indicators := availableStrategies[strategy](symbol, series)

								var foundOrder orders.Order
								r := db.Client.Table("orders").First(&foundOrder, "strategy = ? AND symbol = ? AND decision = ?", strategy, symbol, "Buy")
								if r.Error != nil && !r.RecordNotFound() {
									log.Panicln(r.Error)
									return
								}

								if r.RecordNotFound() && decision == "Sell" {
									log.Println("err: no buy has been done on this symbol to initiate sell")
									continue
								} else if !r.RecordNotFound() && decision == "Buy" {
									log.Println("err: this position is already bought")
									continue
								} // I get the feeling I can somehow improve this logic. Readability though?

								price := series.LastCandle().ClosePrice.String()

								var quantity float64

								switch decision {
								case "Buy":
									quantity := fmt.Sprintf("%f", buyAmount/series.LastCandle().ClosePrice.Float())
									order := client.CreateOrder(symbol, quantity, price, binance.SideTypeBuy)
									fmt.Println(order)

									util.ShowJSON(client.GetCurrencies("BNB", "BUSD"))
								case "Sell":
									quantity := fmt.Sprint(foundOrder.Quantity)
									order := client.CreateOrder(symbol, quantity, price, binance.SideTypeSell)
									fmt.Println(order)

									util.ShowJSON(client.GetCurrencies("BNB", "BUSD"))
								}

								indicatorsJSON, err := json.Marshal(indicators)
								if err != nil {
									log.Panicln(err)
								}

								//var order *orders.Order
								order := &orders.Order{
									Symbol:     symbol,
									Strategy:   strategy,
									Decision:   decision,
									Quantity:   quantity,
									Price:      series.LastCandle().ClosePrice.Float() * quantity,
									Indicators: string(indicatorsJSON),
								}
								if decision != "Hold" {
									r := db.Client.Table("orders").Create(order)
									if r.Error != nil {
										log.Panicln(r.Error)
									}
								}

								data := indicators
								data["Current price"] = series.LastCandle().ClosePrice.String()
								data["Time"] = time.Now().Format("02-01-2006 15:04:05")
								data["Symbol"] = symbol
								data["Decision"] = decision
								data["Strategy"] = strategy

								message := fmt.Sprint(
									"----------", "\n",
								)
								for _, v := range *strategies.Datakeys[strategy] {
									message += fmt.Sprint(v, ": ", data[v], "\n")
								}

								fmt.Println(message)

								if decision != "Hold" {
									writerType := output.Excel
									writer := output.NewWriterCreator().CreateWriter(writerType)
									writer.WriteToLog(data, strategies.Datakeys[strategy])
								}

								strategies.UpdateOrCreateAnalysis(order)
							}
						}
					}
				}
			}(selectedSymbols)
		case "2":
			settings.ScanUpdateOrCreate("selected_strategies")
		case "3":
			settings.ScanUpdateOrCreate("selected_symbols")
		case "4":
			klines, err := client.GetKlines("LTCBTC", timeframe)
			if err != nil {
				log.Panicln(err)
				break
			}

			util.ShowJSON(klines)
		case "5":
			util.ShowJSON(client.GetAccount())
		case "6":
			prices, err := client.NewListPricesService().Do(context.Background())
			if err != nil {
				log.Panicln(err)
				return
			}

			util.ShowJSON(prices)
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
	}

	return series
}
