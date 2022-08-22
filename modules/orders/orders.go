package orders

import (
	"time"

	"github.com/ws396/autobinance/modules/db"
)

// Btw, it might be more accurate to call this trades, not orders, because I don't think I'll be verifying the order anyway.

type Order struct {
	ID         uint      `json:"id" gorm:"primary_key;auto_increment"`
	Strategy   string    `json:"strategy"`
	Symbol     string    `json:"symbol"`
	Decision   string    `json:"decision"`
	Quantity   float64   `json:"quantity"`
	Price      float64   `json:"price"`
	Indicators string    `json:"indicators"` // Serialized json of all used indicators
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func AutoMigrateOrders() {
	db.Client.AutoMigrate(&Order{})
}