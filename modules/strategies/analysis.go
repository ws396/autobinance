package strategies

import (
	"time"

	"github.com/ws396/autobinance/modules/db"
)

//"github.com/go-playground/validator/v10"

type Analysis struct {
	ID              uint          `json:"id" gorm:"primary_key;auto_increment"`
	StrategyName    string        `json:"strategyName" validate:"required"`
	Symbol          string        `json:"symbol" validate:"required"`
	Buys            uint          `json:"buys"`
	Sells           uint          `json:"sells"`
	SuccessfulSells uint          `json:"successfulSells"`
	ProfitUSD       float64       `json:"profitUSD"`
	SuccessRate     float32       `json:"successRate"` // Should be calculated from: Sells for a better price than their buys / Total sells
	ActiveTime      time.Duration `json:"activeTime"`  // Should be updated with each trade attempt for each position. Feels a bit too performance heavy? Could rather count the trades in DB on demand
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

func AutoMigrateAnalyses() {
	db.Client.AutoMigrate(&Analysis{})
}
