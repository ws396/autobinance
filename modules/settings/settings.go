package settings

import (
	"fmt"
	"log"

	"github.com/ws396/autobinance/modules/db"
)

type Setting struct {
	ID    uint   `json:"id" gorm:"primary_key;auto_increment"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func AutoMigrateSettings() {
	db.Client.AutoMigrate(&Setting{})
}

func ScanUpdateOrCreate(name string) {
	var input string
	var foundSetting Setting
	r := db.Client.Table("settings").First(&foundSetting, "name = ?", name)
	if r.Error != nil && !r.RecordNotFound() {
		log.Panicln(r.Error)
	}
	if !r.RecordNotFound() {
		fmt.Printf("Current value of %s: %s \n", name, foundSetting.Value)
	}

	fmt.Printf("Set new value of %s: \n", name)
	fmt.Scanln(&input)
	if input == "\\back" {
		return
	} // Should use a return like this everywhere. Need to do something about it after I find a suitable package for the cmd menus overhaul.

	if r.RecordNotFound() {
		r = db.Client.Table("settings").Create(&Setting{
			Name:  name,
			Value: input,
		})
		if r.Error != nil {
			log.Panicln(r.Error)
		}
	} else {
		foundSetting.Value = input
		r = db.Client.Table("settings").Save(&foundSetting)
		if r.Error != nil {
			log.Panicln(r.Error)
		}
	}
}
