package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2" // Might be better to eventually get rid of this dependency here
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
	"github.com/ws396/autobinance/modules/analysis"
	"github.com/ws396/autobinance/modules/binancew-sim"
	"github.com/ws396/autobinance/modules/db"
	"github.com/ws396/autobinance/modules/globals"
	"github.com/ws396/autobinance/modules/orders"
	"github.com/ws396/autobinance/modules/output"
	"github.com/ws396/autobinance/modules/settings"
	"github.com/ws396/autobinance/modules/techanext"
	"github.com/ws396/autobinance/modules/util"
	"gorm.io/gorm"
)

var (
	buyAmount float64 = 50
	client    *binancew.ClientExt
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

	apiKey := os.Getenv("API_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	client = binancew.NewExtClient(apiKey, secretKey)

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

type errMsg error

// All view-related data goes through model
type model struct {
	settings       settings.Settings
	quitting       bool
	choice         string
	textInput      textinput.Model
	width          int
	info           string
	err            error
	tradingRunning bool
	stopTrading    chan bool
	ticker         *time.Ticker
}

func initialModel() model {
	ti := textinput.New()
	ti.Focus()
	//ti.CharLimit = 156
	ti.Width = 80

	// Pre-load all settings here?
	settings, err := settings.GetSettings()

	// Feels a bit wrong to put it here. But then again, it IS directly related to the main business logic of this app
	ticker := time.NewTicker(time.Duration(globals.Timeframe) * time.Minute / 6)

	return model{
		settings:    settings,
		choice:      "root",
		textInput:   ti,
		err:         err,
		stopTrading: make(chan bool),
		ticker:      ticker,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyEnter:
			m.Logic()
			m.textInput.Reset()
		}
	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		//m.height = msg.Height
	}

	m.textInput, cmd = m.textInput.Update(msg)

	return m, cmd
}

func (m *model) Logic() {
	if m.textInput.Value() == "\\q" && m.choice != "root" {
		m.choice = "root"
		return
	}

	// Could also use something like linked list structures here?
	// Last choice after-render logic
	switch m.choice {
	case "2":
		settings.Update(m.settings.SelectedStrategies.Name, m.textInput.Value())
		m.settings.SelectedStrategies.Value = m.textInput.Value()
	case "3":
		settings.Update(m.settings.SelectedSymbols.Name, m.textInput.Value())
		m.settings.SelectedSymbols.Value = m.textInput.Value()
	}

	// Choice transition (Might also want to check here if previous step actually needs a transition)
	if m.err == nil {
		switch m.choice {
		case "root":
			m.choice = m.textInput.Value()
		default:
			m.choice = "root"
		}
	}

	// New choice pre-render logic
	switch m.choice {
	case "1": // Keep in mind that with current approach the unsold assets will remain unsold when the app stops
		excelWriter := output.NewWriterCreator().CreateWriter(output.Excel)
		m.startTradingSession(excelWriter)
	case "2":
		m.settings.SelectedStrategies.Value = settings.Find(m.settings.SelectedStrategies.Name)
	case "3":
		m.settings.SelectedSymbols.Value = settings.Find(m.settings.SelectedSymbols.Name)
	case "4":
		klines, err := client.GetKlines("LTCBTC", globals.Timeframe)
		if err != nil {
			log.Panicln(err)
			break
		}

		util.WriteToLogMisc(klines)
	case "5":
		util.WriteToLogMisc(client.GetAccount())
	case "6":
		filesToRemove := []string{
			output.Filename + ".txt",
			output.Filename + ".xlsx",
			"log_gorm.txt",
			"log_misc.txt",
		}
		for _, path := range filesToRemove {
			os.Remove(path)
		}
	case "7":
		db.Client.Migrator().DropTable(&analysis.Analysis{})
		db.Client.Migrator().DropTable(&settings.Setting{})
		db.Client.Migrator().DropTable(&orders.Order{})

		analysis.AutoMigrateAnalyses()
		settings.AutoMigrateSettings()
		orders.AutoMigrateOrders()
	case "8":
		if m.tradingRunning {
			m.tradingRunning = false
			m.stopTrading <- true
		}
	}
}

func (m *model) startTradingSession(writer output.Writer) {
	if m.tradingRunning {
		m.err = errors.New("err: the trading is already running")
		return
	}

	selectedStrategies := strings.Split(m.settings.SelectedStrategies.Value, ",")
	selectedSymbols := strings.Split(m.settings.SelectedSymbols.Value, ",")
	m.info = "Trading started (you can still do other actions)"
	m.tradingRunning = true

	go func() {
		for {
			select {
			case <-m.stopTrading:
				return
			case <-m.ticker.C:
				dataChannel := make(chan *orders.Order, len(selectedStrategies)*len(selectedSymbols))
				for _, symbol := range selectedSymbols {
					go func(symbol string) {
						klines, err := client.GetKlines(symbol, globals.Timeframe)
						if err != nil {
							m.err = err
							return
						}

						series := techanext.GetSeries(klines)
						for _, strategy := range selectedStrategies {
							go func(strategy string) {
								decision, indicators := globals.StrategiesInfo[strategy].Handler(series)

								// Might want to consider a different approach with a clearer logic for these DB calls
								var foundOrder orders.Order
								r := db.Client.Table("orders").Last(&foundOrder, "strategy = ? AND symbol = ?", strategy, symbol)
								if r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
									m.err = r.Error
									return
								}

								// I guess I assume that this table will not contain "Holds"
								if foundOrder.Decision == globals.Sell && decision == globals.Sell {
									m.err = errors.New("err: no recent buy has been done on this symbol to initiate sell")
									return
								} else if foundOrder.Decision == globals.Buy && decision == globals.Buy {
									m.err = errors.New("err: this position is already bought")
									return
								}

								var quantity float64
								// Should depend on found order?
								switch decision {
								case globals.Buy, globals.Hold:
									quantity = buyAmount / series.LastCandle().ClosePrice.Float()
								case globals.Sell:
									quantity = foundOrder.Quantity
								}

								price := series.LastCandle().ClosePrice.String()
								// Getting the actual order from this might be useful, but keep in mind that GORM might not like the []*Fill field
								// Could also just selectively add fields from this to orders.Order below instead
								client.CreateOrder(symbol, fmt.Sprintf("%f", quantity), price, binance.SideType(decision))

								order := &orders.Order{
									Strategy:   strategy,
									Symbol:     symbol,
									Decision:   decision,
									Quantity:   quantity,
									Price:      series.LastCandle().ClosePrice.Float() * quantity,
									Indicators: indicators,
									Time:       time.Now(),
								}
								util.WriteToLogMisc(order)

								//if decision != globals.Hold {
								dataChannel <- order
								//}

								if decision != globals.Hold {
									r := db.Client.Table("orders").Create(order)
									if r.Error != nil {
										m.err = r.Error
									}
								}

								// Maybe need to group this up with channels like writing to log below
								err = analysis.UpdateOrCreateAnalysis(order)
								if err != nil {
									m.err = err
								}
							}(strategy)
						}
					}(symbol)
				}

				writer.WriteToLog(dataChannel)
			}
		}
	}()
}

// The main view, which just calls the appropriate sub-view
func (m model) View() string {
	var errMsg string
	if m.err != nil {
		errMsg = m.err.Error()
	}

	return wordwrap.String(indent.String("\n"+chosenView(m)+"\n\n"+m.textInput.View()+"\n\n"+m.info+"\n\n"+errMsg, 4), m.width)
}

func chosenView(m model) string {
	var msg string

	switch m.choice {
	case "root":
		tradingStatus := "OFF"
		if m.tradingRunning {
			tradingStatus = "ON"
		}

		msg = fmt.Sprint(
			"THE GO-BINANCE WRAPPER IS IN SIMULATION MODE", "\n",
			"AUTOBINANCE", "\n",
			"Trading status: ", tradingStatus, "\n\n",
			"1) Start trading session", "\n",
			"2) Set strategies", "\n",
			"3) Set trade symbols", "\n",
			"4) Check klines", "\n",
			"5) Check account", "\n",
			"6) Clear logs and trade history", "\n",
			"7) Recreate tables", "\n",
			"8) Quit trading session",
		)
	case "1":
		msg = fmt.Sprint("Trading session started.")
	case "2":
		msg = fmt.Sprint("Currently selected strategies: ", m.settings.SelectedStrategies.Value)
	case "3":
		msg = fmt.Sprint("Currently selected symbols: ", m.settings.SelectedSymbols.Value)
	case "4", "5":
		msg = fmt.Sprint("Output in log_misc.txt")
	case "6":
		msg = fmt.Sprint("Logs cleared")
	case "7":
		msg = fmt.Sprint("All tables have been recreated")
	case "8":
		msg = fmt.Sprint("Trading session stopped.")
	default:
		msg = fmt.Sprint("Invalid choice (type \\q to go back to root)")
	}

	return msg
}
