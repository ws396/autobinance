package binancew

import (
	"context"
	"fmt"
	"log"

	"github.com/adshao/go-binance/v2"
)

type ClientExt struct {
	*binance.Client
}

func NewExtClient(apiKey, secretKey string) *ClientExt {
	return &ClientExt{binance.NewClient("", "")}
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

func (client *ClientExt) CreateOrder(symbol, quantity, price string, orderType binance.SideType) string {
	/*
		q, err := strconv.ParseFloat(quantity, 64)
		if err != nil {
			log.Panicln(err)
			return ""
		}

		p, err := strconv.ParseFloat(price, 64)
		if err != nil {
			log.Panicln(err)
			return ""
		}

		if orderType == binance.SideTypeBuy {
			client.walletBUSD -= q * p
			client.walletOther[symbol] += q
			client.PlacedOrders[symbol] = true
		} else if orderType == binance.SideTypeSell {
			client.walletBUSD += q * p
			client.walletOther[symbol] -= q
			client.PlacedOrders[symbol] = false
		}

		return fmt.Sprint("BUSD: ", client.walletBUSD, "\n", symbol, ": ", client.walletOther[symbol])
	*/
	return "Order successful (check analyses or orders table)"
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
	/*
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
	*/

	return []binance.Balance{0: {Asset: "Unavailable in simulation mode"}} // Probably need to put something different here?
}
