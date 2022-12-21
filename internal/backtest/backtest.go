package backtest

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/ws396/autobinance/internal/analysis"
	"github.com/ws396/autobinance/internal/binancew"
	"github.com/ws396/autobinance/internal/globals"
	"github.com/ws396/autobinance/internal/output"
	"github.com/ws396/autobinance/internal/storage"
	"github.com/ws396/autobinance/internal/trader"
	"github.com/ws396/autobinance/internal/util"
)

func Backtest(input string, settings map[string]storage.Setting) (map[string]storage.Analysis, error) {
	if !globals.SimulationMode {
		return nil, globals.ErrNotInSimulationMode
	}
	if len(settings["selected_symbols"].ValueArr) == 0 {
		return nil, globals.ErrSymbolsNotFound
	}

	klinesFeed := map[string][]*binance.Kline{}
	start, end, err := util.ExtractTimepoints(input)
	if err != nil {
		return nil, err
	}

	for _, s := range settings["selected_symbols"].ValueArr {
		path := fmt.Sprintf(
			"internal/testdata/%s_%s_%s_%s.csv",
			s,
			globals.Timeframe,
			start.Format("02-01-2006"),
			end.Format("02-01-2006"),
		)
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		reader := csv.NewReader(f)
		records, err := reader.ReadAll()
		if err != nil {
			return nil, err
		}

		klines := []*binance.Kline{}
		for _, record := range records {
			openTime, _ := strconv.ParseInt(record[0], 10, 64)
			closeTime, _ := strconv.ParseInt(record[6], 10, 64)
			tradeNum, _ := strconv.ParseInt(record[8], 10, 64)
			kline := &binance.Kline{
				OpenTime:                 openTime,
				Open:                     record[1],
				High:                     record[2],
				Low:                      record[3],
				Close:                    record[4],
				Volume:                   record[5],
				CloseTime:                closeTime,
				QuoteAssetVolume:         record[7],
				TradeNum:                 tradeNum,
				TakerBuyBaseAssetVolume:  record[9],
				TakerBuyQuoteAssetVolume: record[10],
			}
			klines = append(klines, kline)
		}

		klinesFeed[s] = klines
	}

	batchLimit := 60
	btExchangeClient := binancew.NewClientBacktest(start, end, klinesFeed, batchLimit)
	btStorageClient := storage.NewInMemoryClient()
	tickerChan := make(chan time.Time)
	btTrader := trader.Trader{
		StorageClient:  btStorageClient,
		ExchangeClient: btExchangeClient,
		Settings:       settings,
		TickerChan:     tickerChan,
	}
	w, err := output.NewWriterCreator().CreateWriter(output.Stub)
	if err != nil {
		return nil, err
	}

	errChan := make(chan error)
	btTrader.StartTradingSession(w, errChan)
	klinesLen := len(klinesFeed[settings["selected_symbols"].ValueArr[0]])
	for i := 0; i < klinesLen-batchLimit; i++ {
		tickerChan <- time.Now()
		if err := <-errChan; err != nil {
			btTrader.StopTradingSession()
			return nil, err
		}
	}

	foundOrders, err := btStorageClient.GetAllOrders()
	if err != nil {
		return nil, err
	}

	binancew.BacktestIndex = 0
	analyses := analysis.CreateAnalyses(foundOrders, start, end)

	return analyses, nil
}
