package binancew

import (
	"context"
	"fmt"

	"github.com/adshao/go-binance/v2"
)

type ClientExt struct {
	*binance.Client
	placedOrders []bool
}

func init() {
	binance.UseTestnet = true
}

func NewExtClient(apiKey, secretKey string) *ClientExt {
	return &ClientExt{binance.NewClient(apiKey, secretKey), []bool{}}
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

func (client *ClientExt) GetKlines(symbol string, timeframe int) ([]*binance.Kline, error) {
	klines, err := client.NewKlinesService().Symbol(symbol).
		Interval(fmt.Sprint(timeframe) + "m").Do(context.Background())
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
