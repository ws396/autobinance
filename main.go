package main

import (
	"context"
	"encoding/json"
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
	"github.com/sdcoffey/techan"
	"github.com/ws396/autobinance/modules/binancew-sim"
	"github.com/ws396/autobinance/modules/db"
	"github.com/ws396/autobinance/modules/globals"
	"github.com/ws396/autobinance/modules/orders"
	"github.com/ws396/autobinance/modules/output"
	"github.com/ws396/autobinance/modules/settings"
	"github.com/ws396/autobinance/modules/strategies"
	"github.com/ws396/autobinance/modules/techanext"
	"github.com/ws396/autobinance/modules/util"
)

var (
	//timeframe           int     = 1
	buyAmount           float64 = 50
	strategyRunning             = false
	availableStrategies         = map[string]func(string, *techan.TimeSeries) (string, map[string]string){
		"one":        strategies.StrategyOne,
		"two":        strategies.StrategyTwo,
		"ytwilliams": strategies.StrategyYTWilliams,
		"ytmod":      strategies.StrategyYTMod,
	}
	client *binancew.ClientExt
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Panicln("Error loading .env file")
	}

	db.ConnectDB()
	strategies.AutoMigrateAnalyses()
	settings.AutoMigrateSettings()
	orders.AutoMigrateOrders()

	//strategies.Timeframe = &timeframe

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
	settings  settings.Settings // Non-map version of settings looks kinda ugly, but it's nice to have a specified struct with no magical strings
	quitting  bool
	choice    string
	textInput textinput.Model
	width     int
	//height    int
	info string
	err  error
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Pikachu"
	ti.Focus()
	//ti.CharLimit = 156
	ti.Width = 80
	// Pre-load all options here?

	return model{
		settings:  settings.GetSettingsOutline(),
		choice:    "root",
		textInput: ti,
		err:       nil,
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

	// Last choice after-render logic
	switch m.choice {
	case "2":
		settings.Update(m.settings.SelectedStrategies.Name, m.textInput.Value())
	case "3":
		settings.Update(m.settings.SelectedSymbols.Name, m.textInput.Value())
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
	case "1": // Keep in mind that with current approach the unsold assets will remain unsold with app relaunch
		if strategyRunning {
			m.err = errors.New("err: the strategy is already running")
			return
		}

		var foundStrategies settings.Setting
		r := db.Client.Table("settings").First(&foundStrategies, "name = ?", "selected_strategies")
		if r.Error != nil && !r.RecordNotFound() { // Should be reverse order with elseif
			m.err = errors.New(r.Error.Error()) // Might want to leave panics in places like these :)
		}
		if r.RecordNotFound() {
			m.err = errors.New("err: please specify the strategies first")
		}

		selectedStrategies := strings.Split(foundStrategies.Value, ",")

		var foundSymbols settings.Setting
		r = db.Client.Table("settings").First(&foundSymbols, "name = ?", "selected_symbols")
		if r.Error != nil && !r.RecordNotFound() {
			m.err = errors.New(r.Error.Error())
		}
		if r.RecordNotFound() {
			m.err = errors.New("err: please specify the settings first")
		}

		selectedSymbols := strings.Split(foundSymbols.Value, ",")

		ticker := time.NewTicker(time.Duration(globals.Timeframe) * time.Minute / 6) // Let's try doing these twice per timeframe
		m.info = "Strategy execution started (you can still do other actions)"
		strategyRunning = true

		for { // Wrap in goroutine?
			<-ticker.C
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
							decision, indicators := availableStrategies[strategy](symbol, series)

							var foundOrder orders.Order
							r := db.Client.Table("orders").First(&foundOrder, "strategy = ? AND symbol = ? AND decision = ?", strategy, symbol, "Buy")
							if r.Error != nil && !r.RecordNotFound() {
								m.err = errors.New(r.Error.Error())
								return
							}

							if r.RecordNotFound() && decision == "Sell" {
								m.err = errors.New("err: no buy has been done on this symbol to initiate sell")
								return
							} else if !r.RecordNotFound() && decision == "Buy" {
								m.err = errors.New("err: this position is already bought")
								return
							}

							price := series.LastCandle().ClosePrice.String()

							var quantity float64

							switch decision {
							case "Buy":
								quantity := fmt.Sprintf("%f", buyAmount/series.LastCandle().ClosePrice.Float())
								order := client.CreateOrder(symbol, quantity, price, binance.SideTypeBuy)
								util.ShowJSON(order)
							case "Sell":
								quantity := fmt.Sprint(foundOrder.Quantity)
								order := client.CreateOrder(symbol, quantity, price, binance.SideTypeSell)
								util.ShowJSON(order)
							}

							indicatorsJSON, err := json.Marshal(indicators)
							if err != nil {
								m.err = err
							}

							order := &orders.Order{
								Symbol:     symbol,
								Strategy:   strategy,
								Decision:   decision,
								Quantity:   quantity,
								Price:      series.LastCandle().ClosePrice.Float() * quantity,
								Indicators: string(indicatorsJSON),
							}
							if decision != "Hold" {
								r := db.Client.Table("orders").Create(order)
								if r.Error != nil {
									m.err = errors.New(r.Error.Error())
								}
							}

							data := indicators
							data["Current price"] = series.LastCandle().ClosePrice.String()
							data["Time"] = time.Now().Format("02-01-2006 15:04:05")
							data["Symbol"] = symbol
							data["Decision"] = decision
							data["Strategy"] = strategy

							//if decision != "Hold" {
								
							writerType := output.Excel
							writer := output.NewWriterCreator().CreateWriter(writerType)
							writer.WriteToLog(data) // NOT THREADSAFE!
							//}

							strategies.UpdateOrCreateAnalysis(order)
						}(strategy)
					}
				}(symbol)
			}
		}
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

		util.ShowJSON(klines)
	case "5":
		util.ShowJSON(client.GetAccount())
	case "6":
		prices, err := client.NewListPricesService().Do(context.Background())
		if err != nil {
			log.Panicln(err)
			return
		}

		util.ShowJSON(prices)
	}

}

// The main view, which just calls the appropriate sub-view
func (m model) View() string {
	var err string
	if m.err != nil {
		err = m.err.Error()
	}

	return wordwrap.String(indent.String("\n"+chosenView(m)+"\n\n"+m.textInput.View()+"\n\n"+err, 4), m.width)
}

func chosenView(m model) string {
	var msg string

	switch m.choice {
	case "root":
		msg = fmt.Sprint(
			"THE GO-BINANCE WRAPPER IS IN SIMULATION MODE", "\n",
			"BUBBLETEA SUPER AUTOBINANCE TRADER WOW", "\n\n",
			"1) Start trading execution", "\n",
			"2) Set strategies", "\n",
			"3) Set trade symbols", "\n",
			"4) Check klines", "\n",
			"5) Check account", "\n",
			"6) List trades",
		)
	case "1":
		msg = fmt.Sprint("Strategy execution started.")
	case "2":
		msg = fmt.Sprint("Currently selected strategies: ", m.settings.SelectedStrategies.Value)
	case "3":
		msg = fmt.Sprint("Currently selected symbols: ", m.settings.SelectedSymbols.Value)
	case "4", "5", "6":
		msg = fmt.Sprint("Output in log_misc.txt")
	default:
		msg = fmt.Sprint("Invalid choice (type \\q to go back to root)")
	}

	return msg
}
