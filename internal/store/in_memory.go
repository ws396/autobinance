package store

import (
	"strings"

	"github.com/ws396/autobinance/internal/globals"
)

// It's not needed right now, but I do need to add mutex here
type InMemoryClient struct {
	orders   []Order
	settings map[string]Setting
}

func NewInMemoryClient() *InMemoryClient {
	return &InMemoryClient{
		[]Order{},
		map[string]Setting{},
	}
}

func (c *InMemoryClient) AutoMigrateAll() {
	c.AutoMigrateOrders()
	c.AutoMigrateSettings()
}

func (c *InMemoryClient) AutoMigrateOrders() {
}

func (c *InMemoryClient) AutoMigrateSettings() {
	fields := []string{
		"selected_symbols",
		"selected_strategies",
		"available_strategies",
	}
	for _, v := range fields {
		c.CreateSetting(v, "")
	}
}

func (c *InMemoryClient) DropAll() {
	c.orders = []Order{}
	c.settings = map[string]Setting{}
}

func (c *InMemoryClient) GetAllSettings() (map[string]Setting, error) {
	return c.settings, nil
}

func (c *InMemoryClient) GetSetting(name string) Setting {
	s := c.settings[name]
	return s
}

func (c *InMemoryClient) UpdateSetting(name, value string) Setting {
	s := c.settings[name]
	s.Value = value
	c.settings[name] = s
	s.ValueArr = strings.Split(value, ",")

	return s
}

func (c *InMemoryClient) CreateSetting(name, value string) {
	s := Setting{name, value, []string{}}
	c.settings[name] = s
}

func (c *InMemoryClient) GetAllOrders() ([]Order, error) {
	return c.orders, nil
}

func (c *InMemoryClient) GetLastOrder(strategy, symbol string) (*Order, error) {
	for i := len(c.orders) - 1; i >= 0; i-- {
		if c.orders[i].Strategy == strategy && c.orders[i].Symbol == symbol {
			return &c.orders[i], nil
		}
	}

	return nil, globals.ErrOrderNotFound
}

func (c *InMemoryClient) CreateOrder(order *Order) error {
	c.orders = append(c.orders, *order)

	return nil
}
