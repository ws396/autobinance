package store

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ws396/autobinance/internal/globals"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type GORMClient struct {
	*gorm.DB
}

func NewGORMClient(dialect gorm.Dialector) *GORMClient {
	f, err := os.OpenFile("log_gorm.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	newLogger := logger.New(
		log.New(f, "\r\n", log.LstdFlags), // Could put something related to model.err in here?
		logger.Config{
			SlowThreshold:             time.Second,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
			LogLevel:                  logger.Info,
		},
	)

	config := &gorm.Config{
		Logger: newLogger,
	}

	if globals.SimulationMode {
		config.NamingStrategy = schema.NamingStrategy{
			TablePrefix: "sim_",
		}
	}

	db, err := gorm.Open(dialect, config)
	if err != nil {
		log.Panicln(err)
	}

	return &GORMClient{db}
}

func (c *GORMClient) AutoMigrateAll() {
	c.AutoMigrateOrders()
	c.AutoMigrateSettings()
}

func (c *GORMClient) AutoMigrateOrders() {
	c.AutoMigrate(&Order{})
}

func (c *GORMClient) AutoMigrateSettings() {
	c.AutoMigrate(&Setting{})

	var foundSetting Setting
	r := c.First(&foundSetting)
	if errors.Is(r.Error, gorm.ErrRecordNotFound) {
		fields := []string{
			"selected_symbols",
			"selected_strategies",
			//"available_symbols",
			"available_strategies",
		}
		for _, v := range fields {
			c.CreateSetting(v, "")
		}
	}
}

func (c *GORMClient) GetAllSettings() (*Settings, error) {
	var foundSettings []Setting
	r := c.Find(&foundSettings)
	if r.Error != nil {
		return nil, r.Error
	}

	m := map[string]Setting{}
	for _, v := range foundSettings {
		m[v.Name] = v
	}

	return &Settings{
		SelectedSymbols: Setting{
			m["selected_symbols"].ID,
			m["selected_symbols"].Name,
			m["selected_symbols"].Value,
			strings.Split(m["selected_symbols"].Value, ","),
		},
		SelectedStrategies: Setting{
			m["selected_strategies"].ID,
			m["selected_strategies"].Name,
			m["selected_strategies"].Value,
			strings.Split(m["selected_strategies"].Value, ","),
		},
		AvailableStrategies: Setting{
			m["available_strategies"].ID,
			m["available_strategies"].Name,
			m["available_strategies"].Value,
			strings.Split(m["available_strategies"].Value, ","),
		},
	}, nil
}

func (c *GORMClient) GetSetting(name string) Setting {
	var foundSetting Setting
	r := c.First(&foundSetting, "name = ?", name)
	if r.Error != nil {
		log.Panicln(r.Error)
	}

	foundSetting.ValueArr = strings.Split(foundSetting.Value, ",")

	return foundSetting
}

func (c *GORMClient) UpdateSetting(name, value string) Setting {
	var foundSetting Setting
	r := c.First(&foundSetting, "name = ?", name)
	if r.Error != nil {
		log.Panicln(r.Error)
	}

	foundSetting.Value = value
	foundSetting.ValueArr = strings.Split(value, ",")
	r = c.Save(&foundSetting)
	if r.Error != nil {
		log.Panicln(r.Error)
	}

	return foundSetting
}

func (c *GORMClient) CreateSetting(name, value string) {
	r := c.Create(&Setting{
		Name:  name,
		Value: value,
	})
	if r.Error != nil {
		log.Panicln(r.Error)
	}
}

func (c *GORMClient) GetAllOrders() ([]Order, error) {
	var foundOrders []Order
	r := c.Find(&foundOrders)
	if r.Error != nil {
		return nil, r.Error
	}

	return foundOrders, nil
}

func (c *GORMClient) GetLastOrder(strategy, symbol string) (*Order, error) {
	var foundOrder Order
	r := c.Last(&foundOrder, "strategy = ? AND symbol = ?", strategy, symbol)
	if r.Error != nil {
		return nil, r.Error
	}

	return &foundOrder, nil
}

func (c *GORMClient) CreateOrder(order *Order) error {
	r := c.Create(order)
	if r.Error != nil {
		return r.Error
	}

	return nil
}
