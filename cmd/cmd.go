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

func (m CLI) Init() tea.Cmd {
	return textinput.Blink
}

func (m CLI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m.QuitApp()
		case tea.KeyEnter:
			m.Logic()
			m.textInput.Reset()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}

	m.textInput, cmd = m.textInput.Update(msg)

	return m, cmd
}

func (m *CLI) clearMessages() {
	m.err = nil
	m.info = ""
}

func (m *CLI) Logic() {
	m.clearMessages()

	if m.textInput.Value() == "\\q" {
		m.node = root
		return
	}

	if nextNode := m.node.action(m); nextNode != nil {
		m.node = nextNode
	}
}

func (m *CLI) HandleError(err error) {
	util.Logger.Error(err.Error())
	m.err = err
}

func (m *CLI) HandleFatal(err error) {
	util.Logger.Fatal(err.Error())
}

func (m *CLI) QuitApp() (tea.Model, tea.Cmd) {
	m.quitting = true

	return m, tea.Quit
}

// The main view, which just calls the appropriate sub-view
func (m CLI) View() string {
	var errMsg string
	if m.err != nil {
		errMsg = m.err.Error()
	}

	err := errStyle.Render(errMsg)
	help := helpStyle.Render(m.help)
	border := borderStyle.Render(" ───────────────────────────────────────────")

	return wordwrap.String(
		indent.String(
			"\n"+m.node.view(&m)+"\n\n"+
				m.textInput.View()+"\n\n"+
				border+"\n"+
				m.info+"\n"+
				err+"\n"+
				help,
			4,
		),
		m.width,
	)
}
