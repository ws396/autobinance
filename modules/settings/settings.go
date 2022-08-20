package settings

import (
	"log"

	"github.com/ws396/autobinance/modules/db"
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
	if r.RecordNotFound() {
		fields := []string{
			"selected_symbols",
			"selected_strategies",
		}
		for _, v := range fields {
			Create(v, "")
		}
	} // These likely need to have more error checks.

	/*
			if err = db.AutoMigrate(&User{}); err == nil && db.Migrator().HasTable(&User{}) {
		    if err := db.First(&User{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		        //Insert seed data
		    }
		}
	*/
	/*
		// Reflect is considered bad practice. Guess I'll just repeat the keys manually.
			v := reflect.ValueOf(Settings{})
			for i := 0; i < v.NumField(); i++ {
				field := v.Type().Field(i).Name // Convert to snake case?
				Create(field, "")
			}
	*/
}

func GetSettingsOutline() Settings {
	return Settings{
		Setting{0, "selected_symbols", ""},
		Setting{0, "selected_strategies", ""},
	}
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
