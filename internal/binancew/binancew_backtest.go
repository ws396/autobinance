package binancew

import (
	"time"

	"github.com/adshao/go-binance/v2"
)

var BacktestIndex = 0

type BacktestClient struct {
	ExchangeClient
	Start      time.Time
	End        time.Time
	KlinesFeed map[string][]*binance.Kline
	BatchLimit int
	Index      int
}

func NewClientBacktest(start, end time.Time, klinesFeed map[string][]*binance.Kline, batchLimit int) ExchangeClient {
	return &BacktestClient{NewExtClientSim("", ""), start, end, klinesFeed, batchLimit, 0}
}

func (bc *BacktestClient) GetKlines(symbol string, timeframe uint) ([]*binance.Kline, error) {
	klines := []*binance.Kline{}
	for i := BacktestIndex; i < bc.BatchLimit+BacktestIndex; i++ {
		klines = append(klines, bc.KlinesFeed[symbol][i])
	}

	return klines, nil
}
