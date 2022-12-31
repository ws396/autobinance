package analysis_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/ws396/autobinance/internal/analysis"
	"github.com/ws396/autobinance/internal/globals"
	"github.com/ws396/autobinance/internal/storage"
)

func TestAnalysis(t *testing.T) {
	t.Run("performs analysis on 3 orders", func(t *testing.T) {
		orders := []storage.Order{
			{
				ID:         1,
				Strategy:   "-",
				Symbol:     "-",
				Decision:   globals.Buy,
				Quantity:   10,
				Price:      5,
				Indicators: map[string]string{},
				Timeframe:  "1m",
				Successful: true,
				CreatedAt:  time.Now(),
			},
			{
				ID:         2,
				Strategy:   "-",
				Symbol:     "-",
				Decision:   globals.Hold,
				Quantity:   10,
				Price:      5,
				Indicators: map[string]string{},
				Timeframe:  "1m",
				Successful: false,
				CreatedAt:  time.Now(),
			},
			{
				ID:         3,
				Strategy:   "-",
				Symbol:     "-",
				Decision:   globals.Sell,
				Quantity:   10,
				Price:      5.15,
				Indicators: map[string]string{},
				Timeframe:  "1m",
				Successful: true,
				CreatedAt:  time.Now(),
			},
		}

		stubTime := time.Unix(1600000000, 0)
		got := analysis.CreateAnalyses(orders, time.Unix(1600000000, 0), time.Unix(1600000000, 0))

		for k, a := range got {
			a.CreatedAt = stubTime
			got[k] = a
		}

		want := map[string]storage.Analysis{
			"-_-": {
				ID:              0,
				Strategy:        "-",
				Symbol:          "-",
				Buys:            1,
				Sells:           1,
				SuccessfulSells: 1,
				ProfitUSD:       1.5,
				SuccessRate:     100,
				Timeframe:       "1m",
				Start:           stubTime,
				End:             stubTime,
				CreatedAt:       stubTime,
			},
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("created wrong analysis, got %v want %v", got, want)
		}
	})
}
