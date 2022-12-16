package storage

import "time"

type Order struct {
	ID         uint              `json:"id" gorm:"primary_key;auto_increment"`
	Strategy   string            `json:"strategy"`
	Symbol     string            `json:"symbol"`
	Decision   string            `json:"decision"`
	Quantity   float64           `json:"quantity"`
	Price      float64           `json:"price"`
	Indicators map[string]string `json:"indicators" gorm:"serializer:json"`
	Timeframe  string            `json:"timeframe"`
	Successful bool              `json:"successful"`
	CreatedAt  time.Time         `json:"createdAt"`
}

type Setting struct {
	Name     string   `json:"name" gorm:"primary_key"`
	Value    string   `json:"value"`
	ValueArr []string `json:"valueArr" gorm:"-"`
}

type Analysis struct {
	ID              uint      `json:"id" gorm:"primary_key;auto_increment"`
	Strategy        string    `json:"strategy" validate:"required"`
	Symbol          string    `json:"symbol" validate:"required"`
	Buys            uint      `json:"buys"`
	Sells           uint      `json:"sells"`
	SuccessfulSells uint      `json:"successfulSells"`
	ProfitUSD       float64   `json:"profitUSD"`
	SuccessRate     float64   `json:"successRate"`
	Timeframe       string    `json:"timeframe"`
	Start           time.Time `json:"start"`
	End             time.Time `json:"end"`
	CreatedAt       time.Time `json:"createdAt"`
	//UpdatedAt       time.Time `json:"updatedAt"`
}

type StorageClient interface {
	AutoMigrateAll()
	AutoMigrateOrders()
	AutoMigrateSettings()
	AutoMigrateAnalyses()
	DropAll()
	GetAllSettings() (map[string]Setting, error)
	GetSetting(name string) (Setting, error)
	UpdateSetting(name, value string) (Setting, error)
	StoreSetting(name, value string) error
	GetAllOrders() ([]Order, error)
	GetLastOrder(strategy, symbol string) (*Order, error)
	StoreOrder(order *Order) error
	StoreAnalyses(analyses map[string]Analysis) error
}
