package store

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

/*
type Settings struct {
	SelectedSymbols     Setting
	SelectedStrategies  Setting
	AvailableStrategies Setting
}
*/

type StoreClient interface {
	AutoMigrateAll()
	AutoMigrateOrders()
	AutoMigrateSettings()
	DropAll()
	GetAllSettings() (map[string]Setting, error)
	GetSetting(name string) Setting
	UpdateSetting(name, value string) Setting
	CreateSetting(name, value string)
	GetAllOrders() ([]Order, error)
	GetLastOrder(strategy, symbol string) (*Order, error)
	CreateOrder(order *Order) error
}
