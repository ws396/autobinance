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
	Timeframe  uint              `json:"timeframe"`
	CreatedAt  time.Time         `json:"createdAt"`
}

type Setting struct {
	ID       uint     `json:"id" gorm:"primary_key;auto_increment"`
	Name     string   `json:"name"`
	Value    string   `json:"value"`
	ValueArr []string `json:"valueArr" gorm:"-"`
}

type Settings struct {
	SelectedSymbols     Setting
	SelectedStrategies  Setting
	AvailableStrategies Setting
}
