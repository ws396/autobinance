package trader

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"github.com/ws396/autobinance/internal/binancew"
	"github.com/ws396/autobinance/internal/globals"
	"github.com/ws396/autobinance/internal/output"
	"github.com/ws396/autobinance/internal/storage"
	"github.com/ws396/autobinance/internal/strategies"
	"github.com/ws396/autobinance/internal/techanext"
	"gorm.io/driver/postgres"
)

type Trader struct {
	TradingRunning bool
	StopTrading    chan bool
	StorageClient  storage.StorageClient
	ExchangeClient binancew.ExchangeClient
	Settings       map[string]storage.Setting
	TickerChan     <-chan time.Time
}

func SetupTrader() (*Trader, error) {
	ticker := time.NewTicker(globals.Durations[globals.Timeframe] / 6)
	apiKey := os.Getenv("API_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	var exchangeClient binancew.ExchangeClient
	if globals.SimulationMode {
		exchangeClient = binancew.NewExtClientSim(apiKey, secretKey)
	} else {
		exchangeClient = binancew.NewExtClient(apiKey, secretKey)
	}

	dialect := postgres.New(postgres.Config{
		DSN: fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
			os.Getenv("PGSQL_HOST"),
			os.Getenv("PGSQL_PORT"),
			os.Getenv("PGSQL_DB"),
			os.Getenv("PGSQL_USER"),
			os.Getenv("PGSQL_PASS"),
		),
	})
	storageClient, err := storage.NewGORMClient(dialect)
	if err != nil {
		return nil, err
	}

	//storageClient := storage.NewInMemoryClient()
	storageClient.AutoMigrateAll()
	s, err := storageClient.GetAllSettings()
	if err != nil {
		return nil, err
	}

	var keys []string
	for k := range strategies.StrategiesInfo {
		keys = append(keys, k)
	}

	s["available_strategies"], err = storageClient.UpdateSetting(
		s["available_strategies"].Name,
		strings.Join(keys, ","),
	)
	if err != nil {
		return nil, err
	}

	return &Trader{
		StopTrading:    make(chan bool),
		StorageClient:  storageClient,
		ExchangeClient: exchangeClient,
		Settings:       s,
		TickerChan:     ticker.C,
	}, nil
}

func (t *Trader) StartTradingSession(w output.Writer, errChan chan error) {
	go func() {
		if t.TradingRunning {
			errChan <- globals.ErrTradingAlreadyRunning
			return
		}
		if len(t.Settings["selected_strategies"].ValueArr) == 0 {
			errChan <- globals.ErrStrategiesNotFound
			return
		}
		if len(t.Settings["selected_symbols"].ValueArr) == 0 {
			errChan <- globals.ErrSymbolsNotFound
			return
		}

		t.TradingRunning = true
		chanSize := len(t.Settings["selected_strategies"].ValueArr) * len(t.Settings["selected_symbols"].ValueArr)
		ordersChan := make(chan *storage.Order)

		for {
			select {
			case <-t.TickerChan:
				for _, symbol := range t.Settings["selected_symbols"].ValueArr {
					go func(symbol string) {
						klines, err := t.ExchangeClient.GetKlines(symbol, globals.Timeframe)
						if err != nil {
							errChan <- err
							return
						}

						series := techanext.GetSeries(klines, globals.Durations[globals.Timeframe])
						for _, strategy := range t.Settings["selected_strategies"].ValueArr {
							go func(strategy string) {
								order, err := t.Trade(strategy, symbol, series)
								if err != nil {
									errChan <- err
									return
								}

								ordersChan <- order
							}(strategy)
						}
					}(symbol)
				}

				var orders []*storage.Order
				for i := 0; i < chanSize; i++ {
					data := <-ordersChan

					if data != nil {
						orders = append(orders, data)
					}
				}

				err := w.WriteToLog(orders)
				if err != nil {
					errChan <- err
					return
				}
			case <-t.StopTrading:
				return
			}
		}
	}()
}

func (t *Trader) StopTradingSession() error {
	if t.TradingRunning {
		t.TradingRunning = false
		t.StopTrading <- true
	} else {
		return globals.ErrTradingNotRunning
	}

	return nil
}

func (t *Trader) Trade(strategy, symbol string, series *techan.TimeSeries) (*storage.Order, error) {
	decision, indicators := strategies.StrategiesInfo[strategy].Handler(series)
	order := &storage.Order{
		Strategy:   strategy,
		Symbol:     symbol,
		Decision:   decision,
		Quantity:   0,
		Price:      0,
		Indicators: indicators,
		Timeframe:  globals.Timeframe,
		Successful: false,
		CreatedAt:  time.Now(),
	}

	if decision == globals.Hold {
		return order, nil
	}

	foundOrder, err := t.StorageClient.GetLastOrder(strategy, symbol)
	if err != nil && !errors.Is(err, globals.ErrOrderNotFound) {
		return nil, err
	}

	if foundOrder != nil {
		if (foundOrder.Decision == globals.Sell || foundOrder.Decision == "") && decision == globals.Sell {
			//return nil, errors.New("err: no recent buy has been done on this symbol to initiate sell")
			return order, nil
		} else if foundOrder.Decision == globals.Buy && decision == globals.Buy {
			//return nil, errors.New("err: this position is already bought")
			return order, nil
		}
	} else if decision == globals.Sell {
		return order, nil
	}

	assetPrice := series.LastCandle().ClosePrice
	var quantity big.Decimal
	switch decision {
	case globals.Buy:
		quantity = big.NewDecimal(globals.BuyAmount).Div(assetPrice)
	case globals.Sell:
		quantity = big.NewDecimal(foundOrder.Quantity)
	}

	orderPrice := assetPrice.Mul(quantity)
	order.Quantity = quantity.Float()
	order.Price = orderPrice.Float()
	order.Successful = true

	_, err = t.ExchangeClient.CreateOrder(symbol, quantity.String(), orderPrice.String(), binance.SideType(decision))
	if err != nil {
		return nil, err
	}

	err = t.StorageClient.StoreOrder(order)
	if err != nil {
		return nil, err
	}

	return order, nil
}
