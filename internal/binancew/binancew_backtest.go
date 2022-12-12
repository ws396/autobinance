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
}

func NewClientBacktest(start, end time.Time, klinesFeed map[string][]*binance.Kline, batchLimit int) ExchangeClient {
	return &BacktestClient{NewExtClientSim("", ""), start, end, klinesFeed, batchLimit}
}

func (bc *BacktestClient) GetKlines(symbol string, timeframe uint) ([]*binance.Kline, error) {
	bi := BacktestIndex 
	klines := []*binance.Kline{}
	for i := bi; i < bc.BatchLimit+bi; i++ {
		klines = append(klines, bc.KlinesFeed[symbol][i])
	}

	return klines, nil
}
