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
	"github.com/ws396/autobinance/modules/db"
	"github.com/ws396/autobinance/modules/settings"
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

	db.ConnectDB()
	strategies.AutoMigrateAnalyses()
	settings.AutoMigrateSettings()

	var (
		apiKey              = os.Getenv("API_KEY")
		secretKey           = os.Getenv("SECRET_KEY")
		input               string
		strategyRunning     = false
		availableStrategies = map[string]func(string, *techan.TimeSeries, *map[string]bool) string{
			"one": strategies.StrategyOne,
			"two": strategies.StrategyTwo,
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

			fmt.Println("Select the coin symbols for Strategy One (ex. BNBBUSD,BTCBUSD): ")
			fmt.Scanln(&input)
			symbols := strings.Split(input, ",")
			for i := range symbols {
				symbols[i] = strings.Trim(symbols[i], " ") // fmt.Scanln doesn't work well with whitespaces anyway (use bufio.Scanner?)
				client.PlacedOrders[symbols[i]] = false
			}

			var foundStrategies settings.Setting
			r := db.Client.Table("settings").First(&foundStrategies, "name = selected_strategies")
			if r.Error != nil {
				fmt.Println(r.Error)
			}
			if r.RecordNotFound() {
				fmt.Println("err: no selected strategies could be found")
				break
			}

			selectedStrategies := strings.Split(foundStrategies.Value, ",")

			ticker := time.NewTicker(time.Duration(timeframe) * time.Minute)
			fmt.Println("Strategy execution started (you can still do other actions)")
			strategyRunning = true

			go func(symbols []string) {
				for {
					select {
					case <-ticker.C:
						for _, strategy := range selectedStrategies {
							for _, symbol := range symbols {
								klines, err := client.GetKlines(symbol, timeframe)
								if err != nil {
									fmt.Println(err)
									break
								}

								series := getSeries(klines)
								//decision := strategies.StrategyOne(symbol, series, &client.PlacedOrders)
								//decision := strategies.StrategyTwo(symbol, series, &client.PlacedOrders)
								decision := availableStrategies[strategy](symbol, series, &client.PlacedOrders)

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

								// Should be after order is made...or should it? It's the analysis that matters after all, not the fact of the order.
								var foundAnalysis strategies.Analysis
								r := db.Client.Table("analyses").First(&foundAnalysis, "strategyName = ? AND symbol = ?", strategy, symbol)
								if r.Error != nil {
									fmt.Println(r.Error)
								}
								if r.RowsAffected == 0 {
									r = db.Client.Table("analyses").Create(strategies.Analysis{
										StrategyName:    strategy,
										Symbol:          symbol,
										Buys:            0,
										Sells:           0,
										SuccessfulSells: 0,
										ProfitUSD:       0,
										SuccessRate:     0,
										ActiveTime:      0,
									})
									if r.Error != nil {
										fmt.Println(r.Error)
									}
								} else {
									if decision == "Buy" {
										foundAnalysis.Buys += 1
									} else if decision == "Sell" {
										foundAnalysis.Sells += 1
									}

									db.Client.Table("analyses").Save(&foundAnalysis)
								}
							}
						}
					}
				}
			}(symbols)
		case "2": // Could these two cases be somehow reasonably unified?
			var foundSetting settings.Setting
			r := db.Client.Table("settings").FirstOrCreate(&foundSetting, "name = ?", "selected_strategies")
			if r.Error != nil {
				fmt.Println(r.Error)
			}

			fmt.Println("Currently selected strategies: ", foundSetting.Value)
			fmt.Println("Set the strategies to execute: ")
			fmt.Scanln(&input)

			if !r.RecordNotFound() {
				foundSetting.Value = input
				db.Client.Table("settings").Save(&foundSetting)
			}
		case "3":
			var foundSetting settings.Setting
			r := db.Client.Table("settings").FirstOrCreate(&foundSetting, "name = selected_symbols")
			if r.Error != nil {
				fmt.Println(r.Error)
			}

			fmt.Println("Currently selected symbols: ", foundSetting.Value)
			fmt.Println("Set the symbols to trade on: ")
			fmt.Scanln(&input)

			if !r.RecordNotFound() {
				foundSetting.Value = input
				db.Client.Table("settings").Save(&foundSetting)
			}
		case "4":
			klines, err := client.GetKlines("LTCBTC", timeframe)
			if err != nil {
				fmt.Println(err)
				break
			}

			util.ShowJSON(klines)
		case "5":
			util.ShowJSON(client.GetAccount())
		case "6":
			prices, err := client.NewListPricesService().Do(context.Background())
			if err != nil {
				fmt.Println(err)
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
		//util.ShowJSON(data)
	}

	return series
}
