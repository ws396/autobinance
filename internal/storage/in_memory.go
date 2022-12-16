package storage

import (
	"strings"
	"sync"

	"github.com/ws396/autobinance/internal/globals"
)

type InMemoryClient struct {
	orders   []Order
	settings map[string]Setting
	analyses []Analysis
	lock     sync.RWMutex
}

func NewInMemoryClient() *InMemoryClient {
	return &InMemoryClient{
		[]Order{},
		map[string]Setting{},
		[]Analysis{},
		sync.RWMutex{},
	}
}

func (c *InMemoryClient) AutoMigrateAll() {
	c.AutoMigrateOrders()
	c.AutoMigrateSettings()
	c.AutoMigrateAnalyses()
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
		c.StoreSetting(v, "")
	}
}

func (c *InMemoryClient) AutoMigrateAnalyses() {
}

func (c *InMemoryClient) DropAll() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.orders = []Order{}
	c.settings = map[string]Setting{}
	c.analyses = []Analysis{}
}

func (c *InMemoryClient) GetAllSettings() (map[string]Setting, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.settings, nil
}

func (c *InMemoryClient) GetSetting(name string) (Setting, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	s := c.settings[name]

	return s, nil
}

func (c *InMemoryClient) UpdateSetting(name, value string) (Setting, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	s := c.settings[name]
	s.Value = value
	c.settings[name] = s
	s.ValueArr = strings.Split(value, ",")

	return s, nil
}

func (c *InMemoryClient) StoreSetting(name, value string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	s := Setting{name, value, []string{}}
	c.settings[name] = s

	return nil
}

func (c *InMemoryClient) GetAllOrders() ([]Order, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.orders, nil
}

func (c *InMemoryClient) GetLastOrder(strategy, symbol string) (*Order, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	for i := len(c.orders) - 1; i >= 0; i-- {
		if c.orders[i].Strategy == strategy && c.orders[i].Symbol == symbol {
			return &c.orders[i], nil
		}
	}

	return nil, globals.ErrOrderNotFound
}

func (c *InMemoryClient) StoreOrder(order *Order) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.orders = append(c.orders, *order)

	return nil
}

func (c *InMemoryClient) StoreAnalyses(analyses map[string]Analysis) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, v := range analyses {
		c.analyses = append(c.analyses, v)
	}

	return nil
}
