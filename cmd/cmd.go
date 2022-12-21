package cmd

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
	"github.com/ws396/autobinance/internal/trader"
	"github.com/ws396/autobinance/internal/util"
)

var (
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
)

type CLI struct {
	node      *ViewNode
	textInput textinput.Model
	width     int
	info      string
	err       error
	help      string
	quitting  bool
	T         *trader.Trader
}

func InitialModel() (*CLI, error) {
	ti := textinput.New()
	ti.Focus()
	ti.Width = 80

	t, err := trader.SetupTrader()
	if err != nil {
		return nil, err
	}

	return &CLI{
		node:      root,
		textInput: ti,
		help:      "\\q - back to root",
		T:         t,
	}, nil
}

func (cli CLI) Init() tea.Cmd {
	return textinput.Blink
}

func (cli *CLI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return cli.QuitApp()
		case tea.KeyEnter:
			cli.Logic()
			cli.textInput.Reset()
		}
	case tea.WindowSizeMsg:
		cli.width = msg.Width
	}

	cli.textInput, cmd = cli.textInput.Update(msg)

	return cli, cmd
}

func (cli *CLI) clearMessages() {
	cli.err = nil
	cli.info = ""
}

func (cli *CLI) Logic() {
	cli.clearMessages()

	if cli.textInput.Value() == "\\q" {
		cli.node = root
		return
	}

	if nextNode := cli.node.action(cli); nextNode != nil {
		cli.node = nextNode
	}
}

func (cli *CLI) HandleError(err error) {
	util.Logger.Error(err.Error())
	cli.err = err
}

func (cli *CLI) HandleFatal(err error) {
	util.Logger.Fatal(err.Error())
}

func (cli *CLI) QuitApp() (tea.Model, tea.Cmd) {
	cli.quitting = true
	return cli, tea.Quit
}

// The main view, which just calls the appropriate sub-view
func (cli *CLI) View() string {
	var errMsg string
	if cli.err != nil {
		errMsg = cli.err.Error()
	}

	err := errStyle.Render(errMsg)
	help := helpStyle.Render(cli.help)
	border := borderStyle.Render(" ───────────────────────────────────────────")

	return wordwrap.String(
		indent.String(
			"\n"+cli.node.view(cli)+"\n\n"+
				cli.textInput.View()+"\n\n"+
				border+"\n"+
				cli.info+"\n"+
				err+"\n"+
				help,
			4,
		),
		cli.width,
	)
}
