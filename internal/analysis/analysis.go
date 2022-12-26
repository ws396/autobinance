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
		if !o.Successful {
			continue
		}

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

			if a.SuccessfulSells != 0 {
				a.SuccessRate = float64(a.SuccessfulSells) /
					float64(a.Sells) * 100
			} else {
				a.SuccessRate = 0
			}
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
