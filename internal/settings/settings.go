package settings

import (
	"errors"
	"log"

	"github.com/ws396/autobinance/internal/db"
	"gorm.io/gorm"
)

type Setting struct {
	ID    uint   `json:"id" gorm:"primary_key;auto_increment"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Settings struct {
	SelectedSymbols     Setting
	SelectedStrategies  Setting
	AvailableSymbols    Setting
	AvailableStrategies Setting
}

func AutoMigrateSettings() {
	db.Client.AutoMigrate(&Setting{})

	var foundSetting Setting
	r := db.Client.First(&foundSetting)
	if errors.Is(r.Error, gorm.ErrRecordNotFound) {
		fields := []string{
			"selected_symbols",
			"selected_strategies",
			//"available_symbols",
			"available_strategies",
		}
		for _, v := range fields {
			Create(v, "")
		}
	}
}

func GetSettings() (*Settings, error) {
	var foundSettings []Setting
	r := db.Client.Find(&foundSettings)
	if r.Error != nil {
		return nil, r.Error
	}

	m := map[string]Setting{}
	for _, v := range foundSettings {
		m[v.Name] = v
	}

	return &Settings{
		SelectedSymbols:    Setting{m["selected_symbols"].ID, m["selected_symbols"].Name, m["selected_symbols"].Value},
		SelectedStrategies: Setting{m["selected_strategies"].ID, m["selected_strategies"].Name, m["selected_strategies"].Value},
		//AvailableSymbols:    Setting{m["available_symbols"].ID, m["available_symbols"].Name, m["available_symbols"].Value},
		AvailableStrategies: Setting{m["available_strategies"].ID, m["available_strategies"].Name, m["available_strategies"].Value},
	}, nil
}

func Find(name string) string {
	var foundSetting Setting
	r := db.Client.First(&foundSetting, "name = ?", name)
	if r.Error != nil {
		log.Panicln(r.Error)
	}

	return foundSetting.Value
}

func Update(name, value string) {
	var foundSetting Setting
	r := db.Client.First(&foundSetting, "name = ?", name)
	if r.Error != nil {
		log.Panicln(r.Error)
	}

	foundSetting.Value = value
	r = db.Client.Save(&foundSetting)
	if r.Error != nil {
		log.Panicln(r.Error)
	}
}

func Create(name, value string) {
	r := db.Client.Create(&Setting{
		Name:  name,
		Value: value,
	})
	if r.Error != nil {
		log.Panicln(r.Error)
	}
}
