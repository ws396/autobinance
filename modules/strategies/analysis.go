package strategies

// Use separate package maybe?

import (
	"log"
	"time"

	"github.com/ws396/autobinance/modules/db"
	"github.com/ws396/autobinance/modules/orders"
)

var Timeframe *int

type Analysis struct {
	ID              uint          `json:"id" gorm:"primary_key;auto_increment"`
	Strategy        string        `json:"strategy" validate:"required"`
	Symbol          string        `json:"symbol" validate:"required"`
	Buys            uint          `json:"buys"`
	Sells           uint          `json:"sells"`
	SuccessfulSells uint          `json:"successfulSells"`
	ProfitUSD       float64       `json:"profitUSD"`
	SuccessRate     float64       `json:"successRate"` // Should be calculated from: Sells for a better price than their buys / Total sells
	ActiveTime      time.Duration `json:"activeTime"`  // Should be updated with each trade attempt for each position. Feels a bit too performance heavy? Could rather count the trades in DB on demand
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

func AutoMigrateAnalyses() {
	db.Client.AutoMigrate(&Analysis{})
}

func UpdateOrCreateAnalysis(strategy, symbol, decision string, price float64) {
	var foundAnalysis Analysis
	r := db.Client.Table("analyses").First(&foundAnalysis, "strategy = ? AND symbol = ?", strategy, symbol)
	if r.Error != nil && !r.RecordNotFound() {
		log.Panicln(r.Error)
	}
	if r.RecordNotFound() {
		if decision == "Buy" {
			CreateAnalysis(strategy, symbol, price)
		} else {
			log.Panicln("err: the first order type of symbol should always be buy")
		}
	} else {
		if decision == "Buy" {
			foundAnalysis.Buys += 1
			foundAnalysis.ProfitUSD -= price
		} else if decision == "Sell" {
			foundAnalysis.Sells += 1

			var foundOrder orders.Order
			r := db.Client.Table("orders").First(&foundOrder, "strategy = ? AND symbol = ? AND decision = ?", strategy, symbol, "Buy")
			if r.Error != nil && !r.RecordNotFound() {
				log.Panicln(r.Error)
				return
			}

			if foundOrder.Price < price {
				foundAnalysis.SuccessfulSells += 1
			}

			foundAnalysis.ProfitUSD += price
			foundAnalysis.SuccessRate = float64(foundAnalysis.SuccessfulSells / foundAnalysis.Sells)
		}

		foundAnalysis.ActiveTime += time.Duration(*Timeframe) * time.Minute

		db.Client.Table("analyses").Save(&foundAnalysis)
	}
}

func CreateAnalysis(strategy, symbol string, price float64) {
	r := db.Client.Table("analyses").Create(&Analysis{
		Strategy:        strategy,
		Symbol:          symbol,
		Buys:            1, // Should be safe to assume this?
		Sells:           0,
		SuccessfulSells: 0,
		ProfitUSD:       -price, // And this
		SuccessRate:     0,
		ActiveTime:      0,
	})
	if r.Error != nil {
		log.Panicln(r.Error)
	}
}
