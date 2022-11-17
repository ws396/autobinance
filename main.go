package main

import (
	"log"
	"os"

	// Might be better to eventually get rid of this dependency here
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/ws396/autobinance/cmd"
	"github.com/ws396/autobinance/internal/analysis"
	"github.com/ws396/autobinance/internal/db"
	"github.com/ws396/autobinance/internal/orders"
	"github.com/ws396/autobinance/internal/settings"
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

	f, err := os.OpenFile("log_error.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Could also serve the CLI. This would also open the opportunity to containerize it properly
	p := tea.NewProgram(cmd.InitialModel(), tea.WithAltScreen())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
