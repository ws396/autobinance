package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	fmt.Println("1) Get kline-based prognosis\n2) Check klines\n3) Check account\n4) Get EMA prognosis\n5) List trades")

	var input string
	for {
		fmt.Println("Your choice: ")
		fmt.Scanln(&input)

		switch input {
		/*
			case "1":
				fmt.Println("Select the coin symbols (ex. LTCBTC): ")
				fmt.Scanln(&input)

				klines := client.getKlines(input)
				fmt.Println(klinePrognosis(klines))
		*/
		case "2":
			showInConsole(client.getKlines("LTCBTC"))
		case "3":
			showInConsole(client.getAccount())
		case "4":
			fmt.Println("Select the coin symbols (ex. LTCBTC): ")
			fmt.Scanln(&input)

			klines := client.getKlines(input)
			indicator, candleCount := BasicEma(klines)
			fmt.Println(StrategyExample(indicator, candleCount))
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

	/*
		port := os.Getenv("PORT")
		if port == "" {
			port = "3000"
		}

		mux := http.NewServeMux()

		mux.HandleFunc("/", indexHandler)
		http.ListenAndServe(":"+port, mux)
	*/
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<h1>Hello World!</h1>"))
}

// BasicEma is an example of how to create a basic Exponential moving average indicator
// based on the close prices of a timeseries from your exchange of choice.
func BasicEma(klines []*binance.Kline) (techan.Indicator, int) {
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

	closePrices := techan.NewClosePriceIndicator(series)
	movingAverage := techan.NewEMAIndicator(closePrices, 10) // Windows depicts amount of periods taken

	return movingAverage, len(series.Candles)
}

/*
func StrategyOne(indicator techan.Indicator, candleCount int) bool {

}
*/

func StrategyExample(indicator techan.Indicator, candleCount int) bool {
	// record trades on this object
	record := techan.NewTradingRecord()

	entryConstant := techan.NewConstantIndicator(30)
	exitConstant := techan.NewConstantIndicator(10)

	entryRule := techan.And(
		techan.NewCrossUpIndicatorRule(entryConstant, indicator),
		techan.PositionNewRule{},
	) // Is satisfied when the price ema moves above 30 and the current position is new

	exitRule := techan.And(
		techan.NewCrossDownIndicatorRule(indicator, exitConstant),
		techan.PositionOpenRule{},
	) // Is satisfied when the price ema moves below 10 and the current position is open

	strategy := techan.RuleStrategy{
		UnstablePeriod: 10,
		EntryRule:      entryRule,
		ExitRule:       exitRule,
	}

	fmt.Println(indicator.Calculate(candleCount - 1)) // Calculate argument is index of series array
	return strategy.ShouldEnter(20, record)           // Index should be > UnstablePeriod for getting true!
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

func (client binanceClientExt) getKlines(symbol string) []*binance.Kline {
	klines, err := client.NewKlinesService().Symbol(symbol).
		Interval("5m").Do(context.Background())
	if err != nil {
		fmt.Println(err)
		//return
	}

	return klines
}

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
