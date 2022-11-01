package settings

import (
	"errors"
	"log"

	"github.com/ws396/autobinance/modules/db"
	"gorm.io/gorm"
)

type Setting struct {
	ID    uint   `json:"id" gorm:"primary_key;auto_increment"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Settings struct {
	SelectedSymbols    Setting
	SelectedStrategies Setting
}

func AutoMigrateSettings() {
	db.Client.AutoMigrate(&Setting{})

	var foundSetting Setting
	r := db.Client.Table("settings").First(&foundSetting)
	if errors.Is(r.Error, gorm.ErrRecordNotFound) {
		fields := []string{
			"selected_symbols",
			"selected_strategies",
		}
		for _, v := range fields {
			Create(v, "")
		}
	}
}

func GetSettings() (Settings, error) { // Return pointer maybe?
	var foundStrategies Setting
	r := db.Client.Table("settings").First(&foundStrategies, "name = ?", "selected_strategies")
	if errors.Is(r.Error, gorm.ErrRecordNotFound) {
		return Settings{}, errors.New("err: please specify the strategies first")
	} else if r.Error != nil {
		return Settings{}, r.Error
	}

	var foundSymbols Setting
	r = db.Client.Table("settings").First(&foundSymbols, "name = ?", "selected_symbols")
	if errors.Is(r.Error, gorm.ErrRecordNotFound) {
		return Settings{}, errors.New("err: please specify the settings first")
	} else if r.Error != nil {
		return Settings{}, r.Error
	} // Should really merge these two

	return Settings{
		Setting{0, "selected_symbols", foundSymbols.Value},
		Setting{0, "selected_strategies", foundStrategies.Value},
	}, nil
}

func Find(name string) string {
	var foundSetting Setting
	r := db.Client.Table("settings").First(&foundSetting, "name = ?", name)
	if r.Error != nil {
		log.Panicln(r.Error)
	}

	return foundSetting.Value
}

func Update(name, value string) {
	var foundSetting Setting
	r := db.Client.Table("settings").First(&foundSetting, "name = ?", name)
	if r.Error != nil {
		log.Panicln(r.Error)
	}

	foundSetting.Value = value
	r = db.Client.Table("settings").Save(&foundSetting)
	if r.Error != nil {
		log.Panicln(r.Error)
	}
}

func Create(name, value string) {
	r := db.Client.Table("settings").Create(&Setting{
		Name:  name,
		Value: value,
	})
	if r.Error != nil {
		log.Panicln(r.Error)
	}
}
