package analysis

import (
	"errors"
	"time"

	"github.com/ws396/autobinance/modules/db"
	"github.com/ws396/autobinance/modules/globals"
	"github.com/ws396/autobinance/modules/orders"
	"gorm.io/gorm"
)

type Analysis struct {
	ID              uint          `json:"id" gorm:"primary_key;auto_increment"`
	Strategy        string        `json:"strategy" validate:"required"`
	Symbol          string        `json:"symbol" validate:"required"`
	Buys            uint          `json:"buys"`
	Sells           uint          `json:"sells"`
	SuccessfulSells uint          `json:"successfulSells"`
	ProfitUSD       float64       `json:"profitUSD"`
	SuccessRate     float64       `json:"successRate"`
	Timeframe       int           `json:"timeframe"`
	ActiveTime      time.Duration `json:"activeTime"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

func AutoMigrateAnalyses() {
	db.Client.AutoMigrate(&Analysis{})
}

func UpdateOrCreateAnalysis(order *orders.Order) error {
	var foundAnalysis Analysis
	r := db.Client.Table("analyses").First(&foundAnalysis, "strategy = ? AND symbol = ?", order.Strategy, order.Symbol)
	if errors.Is(r.Error, gorm.ErrRecordNotFound) {
		if order.Decision == globals.Buy {
			err := createAnalysis(order.Strategy, order.Symbol, order.Price)
			if err != nil {
				return err
			}
		} else if order.Decision == globals.Sell {
			return errors.New("err: the first order type of symbol should always be buy")
		}
	} else if r.Error != nil {
		return r.Error
	} else {
		err := updateAnalysis(order, &foundAnalysis)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateAnalysis(order *orders.Order, foundAnalysis *Analysis) error {
	if order.Decision == globals.Buy {
		foundAnalysis.Buys += 1
		foundAnalysis.ProfitUSD -= order.Price
	} else if order.Decision == globals.Sell {
		var foundOrder orders.Order
		r := db.Client.Table("orders").Last(&foundOrder, "strategy = ? AND symbol = ? AND decision = ?", order.Strategy, order.Symbol, globals.Buy)
		if r.Error != nil {
			return r.Error
		}

		if foundOrder.Price < order.Price {
			foundAnalysis.SuccessfulSells += 1
		}

		foundAnalysis.ProfitUSD += order.Price
		foundAnalysis.Sells += 1
		foundAnalysis.SuccessRate = float64(foundAnalysis.SuccessfulSells) / float64(foundAnalysis.Sells)
	}

	// Or calculate on demand. Could also just do UpdatedAt - CreatedAt
	foundAnalysis.ActiveTime += time.Duration(globals.Timeframe) * time.Minute

	db.Client.Table("analyses").Save(foundAnalysis)

	return nil
}

func createAnalysis(strategy, symbol string, price float64) error {
	r := db.Client.Table("analyses").Create(&Analysis{
		Strategy:        strategy,
		Symbol:          symbol,
		Buys:            1, // Should be safe to assume this?
		Sells:           0,
		SuccessfulSells: 0,
		ProfitUSD:       -price, // And this
		SuccessRate:     0,
		Timeframe:       globals.Timeframe,
		ActiveTime:      0,
	})
	if r.Error != nil {
		return r.Error
	}

	return nil
}
