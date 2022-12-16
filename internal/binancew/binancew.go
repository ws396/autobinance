package binancew

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
)

var (
	once    sync.Once
	symbols []string
)

type ExchangeClient interface {
	CreateOrder(input, quantity, price string, orderType binance.SideType) (*binance.CreateOrderResponse, error)
	GetOrders(symbol string) ([]*binance.Order, error)
	GetKlines(symbol, timeframe string) ([]*binance.Kline, error)
	GetKlinesByPeriod(symbol, timeframe string, start, end time.Time) ([]*binance.Kline, error)
	GetAccount() (*binance.Account, error)
	GetCurrencies(symbol ...string) ([]binance.Balance, error)
	GetAllSymbols() []string
}

type ClientExt struct {
	*binance.Client
}

func init() {
	binance.UseTestnet = true
}

func NewExtClient(apiKey, secretKey string) ExchangeClient {
	return &ClientExt{binance.NewClient(apiKey, secretKey)}
}

func (client *ClientExt) CreateOrder(input, quantity, price string, orderType binance.SideType) (*binance.CreateOrderResponse, error) {
	order, err := client.NewCreateOrderService().
		Symbol(input).
		Side(orderType).
		Type(binance.OrderTypeLimit). // Use market order instead?
		TimeInForce(binance.TimeInForceTypeIOC).
		Quantity(quantity).
		Price(price).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (client *ClientExt) GetOrders(symbol string) ([]*binance.Order, error) {
	orders, err := client.NewListOrdersService().Symbol(symbol).
		Do(context.Background(), binance.WithRecvWindow(10000))
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (client *ClientExt) GetKlines(symbol, timeframe string) ([]*binance.Kline, error) {
	klines, err := client.NewKlinesService().Symbol(symbol).
		Interval(timeframe).Do(context.Background())
	if err != nil {
		return nil, err
	}

	return klines, nil
}

func (client *ClientExt) GetKlinesByPeriod(symbol, timeframe string, start, end time.Time) ([]*binance.Kline, error) {
	klines, err := client.NewKlinesService().Symbol(symbol).
		Interval(timeframe).StartTime(start.Unix() * 1000).
		EndTime(end.Unix() * 1000).Do(context.Background())
	if err != nil {
		return nil, err
	}

	return klines, nil
}

func (client *ClientExt) GetAccount() (*binance.Account, error) {
	account, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (client *ClientExt) GetCurrencies(symbol ...string) ([]binance.Balance, error) {
	result := []binance.Balance{}
	account, err := client.GetAccount()
	if err != nil {
		return nil, err
	}

	for i := range account.Balances {
		for _, v := range symbol {
			if account.Balances[i].Asset == v {
				result = append(result, account.Balances[i])
			}
		}
	}

	return result, nil
}

func (client *ClientExt) GetAllSymbols() []string {
	once.Do(func() {
		info, err := client.NewExchangeInfoService().Do(context.Background())
		if err != nil {
			log.Panicln(err)
		}

		for _, v := range info.Symbols {
			symbols = append(symbols, v.Symbol)
		}
	})

	return symbols
}
