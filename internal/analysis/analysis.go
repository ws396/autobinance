package analysis

import (
	"strings"
	"time"

	"github.com/ws396/autobinance/internal/globals"
	"github.com/ws396/autobinance/internal/storage"
)

func CreateAnalyses(orders []storage.Order, start, end time.Time) map[string]storage.Analysis {
	analyses := map[string]storage.Analysis{}
	lastBuyPrices := map[string]float64{}
	for _, o := range orders {
		k := o.Strategy + "_" + o.Symbol
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
		ss := strings.Split(k, "_")
		a.Strategy = ss[0]
		a.Symbol = ss[1]
		a.Start = start
		a.End = end
		a.CreatedAt = t
		a.Timeframe = orders[0].Timeframe
		analyses[k] = a
	}

	return analyses
}
