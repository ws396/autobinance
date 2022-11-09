package main

import (
	"log"

	// Might be better to eventually get rid of this dependency here
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/ws396/autobinance/modules/analysis"
	"github.com/ws396/autobinance/modules/cmd"
	"github.com/ws396/autobinance/modules/db"
	"github.com/ws396/autobinance/modules/orders"
	"github.com/ws396/autobinance/modules/settings"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Panicln("Error loading .env file")
	}

	db.ConnectDB()
	analysis.AutoMigrateAnalyses()
	settings.AutoMigrateSettings()
	orders.AutoMigrateOrders()

	p := tea.NewProgram(cmd.InitialModel(), tea.WithAltScreen())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
