package analysis

import (
	"time"

	"github.com/ws396/autobinance/internal/globals"
	"github.com/ws396/autobinance/internal/store"
)

/*
type Analysis struct {
	ID              uint      `json:"id" gorm:"primary_key;auto_increment"`
	Strategy        string    `json:"strategy" validate:"required"`
	Symbol          string    `json:"symbol" validate:"required"`
	Buys            uint      `json:"buys"`
	Sells           uint      `json:"sells"`
	SuccessfulSells uint      `json:"successfulSells"`
	ProfitUSD       float64   `json:"profitUSD"`
	SuccessRate     float64   `json:"successRate"`
	Timeframe       uint      `json:"timeframe"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func AutoMigrateAnalyses() {
	store.Client.AutoMigrate(&Analysis{})
}
*/

type Analysis struct {
	Buys            uint      `json:"buys"`
	Sells           uint      `json:"sells"`
	SuccessfulSells uint      `json:"successfulSells"`
	ProfitUSD       float64   `json:"profitUSD"`
	SuccessRate     float64   `json:"successRate"`
	Timeframe       uint      `json:"timeframe"`
	CreatedAt       time.Time `json:"createdAt"`
}

func CreateAnalysis(orders []store.Order) map[string]Analysis {
	analyses := map[string]Analysis{}
	lastBuyPrices := map[string]float64{}
	for _, o := range orders {
		k := o.Strategy + o.Symbol
		a := analyses[k]

		if o.Decision == globals.Buy {
			a.Buys += 1
			a.ProfitUSD -= o.Price
			lastBuyPrices[k] = o.Price
		} else if o.Decision == globals.Sell {
			if lastBuyPrices[k] < o.Price {
				a.SuccessfulSells += 1
			}

			a.ProfitUSD += o.Price
			a.Sells += 1
			a.SuccessRate = float64(analyses[k].SuccessfulSells) / float64(analyses[k].Sells)
		}

		analyses[k] = a
	}

	t := time.Now()
	for k, a := range analyses {
		a.CreatedAt = t
		a.Timeframe = orders[0].Timeframe
		analyses[k] = a
	}

	return analyses
}
