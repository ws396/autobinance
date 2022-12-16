package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/ws396/autobinance/cmd"
	"github.com/ws396/autobinance/internal/util"
)

func main() {
	/*
		f, err := os.OpenFile("log_error.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Panicln(err)
		}
		defer f.Close()

		log.SetOutput(f)
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	*/
	util.InitZapLogger()

	err := godotenv.Load()
	if err != nil {
		util.Logger.Fatal(err.Error())
	}

	// Could also serve the CLI. This would also open the opportunity to containerize it properly
	m, err := cmd.InitialModel()
	if err != nil {
		util.Logger.Fatal(err.Error())
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		util.Logger.Fatal(err.Error())
	}
}
