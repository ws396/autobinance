package main

import (
	"github.com/joho/godotenv"
	"github.com/ws396/autobinance/internal/server"
	"github.com/ws396/autobinance/internal/util"
)

func main() {
	util.InitZapLogger()

	err := godotenv.Load()
	if err != nil {
		util.Logger.Fatal(err.Error())
	}

	/* 	m, err := cmd.InitialModel()
	   	if err != nil {
	   		util.Logger.Fatal(err.Error())
	   	}

	   	p := tea.NewProgram(m, tea.WithAltScreen())
	   	//p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithoutSignalHandler())
	   	//p := tea.NewProgram(m)

	   	if _, err := p.Run(); err != nil {
	   		util.Logger.Fatal(err.Error())
	   	} */

	server.Start()
}

/* func signalsHandler() {
	sigCh := make(chan os.Signal, 1)
	   	signal.Notify(sigCh,
	   		syscall.SIGINT,
	   		syscall.SIGKILL,
	   		syscall.SIGTERM,
	   		syscall.SIGQUIT,
	   	)

	   	var input string
	   	_, err := fmt.Scan(&input)
	   	if err != nil {
	   		util.Logger.Fatal(err.Error())
	   	}

	   	fmt.Println(input)

	   	s := <-sigCh
	   	util.Logger.Error(s.String())
} */
