package settings

import "github.com/ws396/autobinance/modules/db"

type Setting struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func AutoMigrateSettings() {
	db.Client.AutoMigrate(&Setting{})
}
