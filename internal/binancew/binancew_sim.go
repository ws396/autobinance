package binancew

import (
	"time"

	"github.com/adshao/go-binance/v2"
)

var (
	refClient = NewExtClient("", "")
)

type ClientExtSim struct {
	*binance.Client
}

func NewExtClientSim(apiKey, secretKey string) ExchangeClient {
	return &ClientExtSim{binance.NewClient("", "")}
}

func (client *ClientExtSim) CreateOrder(symbol, quantity, price string, orderType binance.SideType) (*binance.CreateOrderResponse, error) {
	return nil, nil
}

func (client *ClientExtSim) GetOrders(symbol string) ([]*binance.Order, error) {
	return nil, nil
}

func (client *ClientExtSim) GetKlines(symbol string, timeframe uint) ([]*binance.Kline, error) {
	return refClient.GetKlines(symbol, timeframe)
}

func (client *ClientExtSim) GetKlinesByPeriod(symbol string, timeframe uint, start, end time.Time) ([]*binance.Kline, error) {
	return refClient.GetKlinesByPeriod(symbol, timeframe, start, end)
}

func (client *ClientExtSim) GetAccount() (*binance.Account, error) {
	return nil, nil
}

func (client *ClientExtSim) GetCurrencies(symbol ...string) ([]binance.Balance, error) {
	return nil, nil
}

func (client *ClientExtSim) GetAllSymbols() []string {
	return refClient.GetAllSymbols()
}
