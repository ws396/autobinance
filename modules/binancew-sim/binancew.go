package binancew

import (
	"context"
	"fmt"

	"github.com/adshao/go-binance/v2"
)

type ClientExt struct {
	*binance.Client
}

func NewExtClient(apiKey, secretKey string) *ClientExt {
	return &ClientExt{binance.NewClient("", "")}
}

func (client *ClientExt) CreateOrder(symbol, quantity, price string, orderType binance.SideType) (*binance.CreateOrderResponse, error) {
	return nil, nil
}

func (client *ClientExt) GetOrders(symbol string) ([]*binance.Order, error) {
	return nil, nil
}

func (client *ClientExt) GetKlines(symbol string, timeframe int) ([]*binance.Kline, error) {
	klines, err := client.NewKlinesService().Symbol(symbol).
		Interval(fmt.Sprint(timeframe) + "m").Do(context.Background())
	if err != nil {
		return nil, err
	}

	return klines, nil
}

func (client *ClientExt) GetAccount() (*binance.Account, error) {
	return nil, nil
}

func (client *ClientExt) GetCurrencies(symbol ...string) ([]binance.Balance, error) {
	return nil, nil
}
