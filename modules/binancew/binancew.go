package binancew

import (
	"context"
	"fmt"
	"log"

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

func (client *ClientExt) GetPrices() []*binance.SymbolPrice {
	prices, err := client.NewListPricesService().Symbol("LTCBTC").
		Do(context.Background())
	if err != nil {
		log.Panicln(err)
		//return
	}

	return prices
}

func (client *ClientExt) CreateOrder(input, quantity, price string, orderType binance.SideType) *binance.CreateOrderResponse {
	order, err := client.NewCreateOrderService().
		Symbol(input).
		Side(orderType).
		Type(binance.OrderTypeLimit). // Use market order instead?
		TimeInForce(binance.TimeInForceTypeIOC).
		Quantity(quantity).
		Price(price).
		Do(context.Background())
	if err != nil {
		log.Panicln(err)
		return nil
	}

	return order
}

func (client *ClientExt) GetOrders() []*binance.Order {
	orders, err := client.NewListOrdersService().Symbol("LTCBTC").
		Do(context.Background(), binance.WithRecvWindow(10000))
	if err != nil {
		log.Panicln(err)
		//return
	}

	return orders
}

func (client *ClientExt) GetKlines(symbol string, timeframe int) ([]*binance.Kline, error) {
	klines, err := client.NewKlinesService().Symbol(symbol).
		Interval(fmt.Sprint(timeframe) + "m").Do(context.Background())
	if err != nil {
		return nil, err
	}

	return klines, nil
}

func (client *ClientExt) GetAccount() *binance.Account {
	account, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		log.Panicln(err)
		//return
	}

	return account
}

func (client *ClientExt) GetCurrencies(symbol ...string) []binance.Balance {
	result := []binance.Balance{}
	balances := client.GetAccount().Balances
	for i := range balances {
		for _, v := range symbol {
			if balances[i].Asset == v {
				result = append(result, balances[i])
			}
		}
	}

	return result
}
