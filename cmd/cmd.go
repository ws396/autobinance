package cmd

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
	"github.com/sdcoffey/techan"
	"github.com/ws396/autobinance/internal/analysis"
	"github.com/ws396/autobinance/internal/binancew"
	"github.com/ws396/autobinance/internal/db"
	"github.com/ws396/autobinance/internal/download"
	"github.com/ws396/autobinance/internal/globals"
	"github.com/ws396/autobinance/internal/orders"
	"github.com/ws396/autobinance/internal/output"
	"github.com/ws396/autobinance/internal/settings"
	"github.com/ws396/autobinance/internal/strategies"
	"github.com/ws396/autobinance/internal/techanext"
	"github.com/ws396/autobinance/internal/util"
	"gorm.io/gorm"
)

var (
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
)

type errMsg error

// All view-related data goes through Autobinance
type Autobinance struct {
	quitting       bool
	choice         string
	textInput      textinput.Model
	width          int
	info           string
	err            error
	help           string
	tradingRunning bool
	stopTrading    chan bool
	Client         binancew.ExchangeClient
	Settings       *settings.Settings
	TickerChan     <-chan time.Time
}

func InitialModel() Autobinance {
	ti := textinput.New()
	ti.Focus()
	//ti.CharLimit = 156
	ti.Width = 80

	// Feels a bit wrong to put it here. But then again, it IS directly related to the main business logic of this app
	ticker := time.NewTicker(time.Duration(globals.Timeframe) * time.Minute / 6)

	apiKey := os.Getenv("API_KEY")
	secretKey := os.Getenv("SECRET_KEY")

	var client binancew.ExchangeClient
	if globals.SimulationMode {
		client = binancew.NewExtClientSim(apiKey, secretKey)
	} else {
		client = binancew.NewExtClient(apiKey, secretKey)
	}

	s, err := settings.GetSettings()
	if err != nil {
		log.Println(err)
	}

	var keys []string
	for k := range strategies.StrategiesInfo {
		keys = append(keys, k)
	}

	s.AvailableStrategies = settings.Update(s.AvailableStrategies.Name, strings.Join(keys, ","))

	return Autobinance{
		choice:      "root",
		textInput:   ti,
		err:         err,
		help:        "\\q - back to root",
		stopTrading: make(chan bool),
		Client:      client,
		Settings:    s,
		TickerChan:  ticker.C,
	}
}

func (m Autobinance) Init() tea.Cmd {
	return textinput.Blink
}

func (m Autobinance) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *Autobinance) Logic() {
	// Need to review the error flow
	m.err = nil
	m.info = ""

	if m.textInput.Value() == "\\q" && m.choice != "root" {
		m.choice = "root"
		return
	}

	// Could also use something like linked list structures here?
	// Last choice after-render logic
	switch m.choice {
	case "2":
		selectedStrategies := strings.Split(m.textInput.Value(), ",")

		for _, v := range selectedStrategies {
			if !util.Contains(m.Settings.AvailableStrategies.ValueArr, v) {
				m.err = globals.ErrWrongStrategyName
				return
			}
		}

		m.Settings.SelectedStrategies = settings.Update(m.Settings.SelectedStrategies.Name, m.textInput.Value())
	case "3":
		symbols := m.Client.GetAllSymbols()
		selectedSymbols := strings.Split(m.textInput.Value(), ",")

		for _, v := range selectedSymbols {
			if !util.Contains(symbols, v) {
				m.err = globals.ErrWrongSymbol
				return
			}
		}

		m.Settings.SelectedSymbols = settings.Update(m.Settings.SelectedSymbols.Name, m.textInput.Value())
	case "8":
		if len(m.Settings.SelectedSymbols.Value) == 0 {
			m.err = globals.ErrSymbolsNotFound
			return
		}

		start, end, err := util.ExtractTimepoints(m.textInput.Value())
		if err != nil {
			m.err = err
			return
		}

		// Visualize progress bar for this?
		for _, s := range m.Settings.SelectedSymbols.ValueArr {
			err := download.KlinesCSVFromZips(s, globals.Timeframe, start, end)
			if err != nil {
				m.err = err
				return
			}
		}

		m.info = "Testdata downloaded"
	case "9":
		m.Backtest()
	}

	// Choice transition (Might also want to check here if previous step actually needs a transition)
	switch m.choice {
	case "root":
		m.choice = m.textInput.Value()
	default:
		m.choice = "root"
	}

	// New choice pre-render logic
	switch m.choice {
	case "1":
		w := output.NewWriterCreator().CreateWriter(output.Excel)
		m.StartTradingSession(w)
	case "2":
		m.Settings.SelectedStrategies = settings.Find(m.Settings.SelectedStrategies.Name)
	case "3":
		m.Settings.SelectedSymbols = settings.Find(m.Settings.SelectedSymbols.Name)
	case "4":
		var foundAnalyses []analysis.Analysis
		r := db.Client.Find(&foundAnalyses)
		if r.Error != nil {
			m.err = r.Error
			break
		}

		util.WriteToLogMisc(foundAnalyses)
	case "5":
		util.WriteToLogMisc(m.Client.GetCurrencies())
	case "6":
		filesToRemove := []string{
			output.Filename + ".txt",
			output.Filename + ".xlsx",
			"log_gorm.txt",
			"log_misc.txt",
			"log_error.txt",
		}
		for _, path := range filesToRemove {
			os.Remove(path)
		}
	case "7":
		// Confirmation would be nice...
		db.Client.Migrator().DropTable(&analysis.Analysis{})
		db.Client.Migrator().DropTable(&settings.Setting{})
		db.Client.Migrator().DropTable(&orders.Order{})

		analysis.AutoMigrateAnalyses()
		settings.AutoMigrateSettings()
		orders.AutoMigrateOrders()
	case "8":
	case "9":
	case "10":
		m.StopTradingSession()
	}
}

func (m *Autobinance) StartTradingSession(w output.Writer) {
	if m.tradingRunning {
		m.err = globals.ErrTradingAlreadyRunning
		m.choice = "root"
		return
	}
	if len(m.Settings.SelectedStrategies.ValueArr) == 0 {
		m.err = globals.ErrStrategiesNotFound
		m.choice = "root"
		return
	}
	if len(m.Settings.SelectedSymbols.ValueArr) == 0 {
		m.err = globals.ErrSymbolsNotFound
		m.choice = "root"
		return
	}

	m.tradingRunning = true

	go func() {
		for {
			select {
			case <-m.stopTrading:
				return
			case <-m.TickerChan:
				chanSize := len(m.Settings.SelectedStrategies.ValueArr) * len(m.Settings.SelectedSymbols.ValueArr)
				dataChannel := make(chan *orders.Order, chanSize)
				for _, symbol := range m.Settings.SelectedSymbols.ValueArr {
					go func(symbol string) {
						klines, err := m.Client.GetKlines(symbol, globals.Timeframe)
						if err != nil {
							m.err = err
							return
						}

						series := techanext.GetSeries(klines)
						for _, strategy := range m.Settings.SelectedStrategies.ValueArr {
							go func(strategy string) {
								order, err := m.Trade(strategy, symbol, series)
								if err != nil {
									log.Println(err)
								}

								dataChannel <- order
							}(strategy)
						}
					}(symbol)
				}

				var orders []*orders.Order
				for i := 0; i < cap(dataChannel); i++ {
					data := <-dataChannel

					if data != nil && data.Decision != globals.Hold {
						orders = append(orders, data)
					}
				}

				w.WriteToLog(orders)
			}
		}
	}()
}

func (m *Autobinance) StopTradingSession() {
	if m.tradingRunning {
		m.tradingRunning = false
		m.stopTrading <- true
	} else {
		m.choice = "root"
		m.info = "Trading is not running"
	}
}

func (m *Autobinance) Trade(strategy, symbol string, series *techan.TimeSeries) (*orders.Order, error) {
	decision, indicators := strategies.StrategiesInfo[strategy].Handler(series)

	// Might want to consider a different approach with a clearer logic for these DB calls
	var foundOrder orders.Order
	r := db.Client.Last(&foundOrder, "strategy = ? AND symbol = ?", strategy, symbol)
	if r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
		return nil, r.Error
	}

	// I guess I assume that this table will not contain "Holds"
	if (foundOrder.Decision == globals.Sell || foundOrder.Decision == "") && decision == globals.Sell {
		//return nil, errors.New("err: no recent buy has been done on this symbol to initiate sell")
		return nil, nil
	} else if foundOrder.Decision == globals.Buy && decision == globals.Buy {
		//return nil, errors.New("err: this position is already bought")
		return nil, nil
	}

	var quantity float64
	// Should depend on found order?
	switch decision {
	case globals.Buy, globals.Hold:
		quantity = globals.BuyAmount / series.LastCandle().ClosePrice.Float()
	case globals.Sell:
		quantity = foundOrder.Quantity
	}

	price := series.LastCandle().ClosePrice.String()
	// Getting the actual order from this might be useful, but keep in mind that GORM might not like the []*Fill field
	// Could also just selectively add fields from this to orders.Order below instead
	_, err := m.Client.CreateOrder(symbol, fmt.Sprintf("%f", quantity), price, binance.SideType(decision))
	if err != nil {
		return nil, err
	}

	order := &orders.Order{
		Strategy:   strategy,
		Symbol:     symbol,
		Decision:   decision,
		Quantity:   quantity,
		Price:      series.LastCandle().ClosePrice.Float() * quantity,
		Indicators: indicators,
		Time:       time.Now(),
	}

	if decision != globals.Hold {
		r := db.Client.Create(order)
		if r.Error != nil {
			return nil, r.Error
		}
	}

	// Maybe need to group this up somehow?
	err = analysis.UpdateOrCreateAnalysis(order)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (m *Autobinance) Backtest() {
	if !globals.SimulationMode {
		m.info = "Only available in simulation mode"
		return
	}
	if len(m.Settings.SelectedSymbols.ValueArr) == 0 {
		m.err = globals.ErrSymbolsNotFound
	}

	klinesFeed := map[string][]*binance.Kline{}

	start, end, err := util.ExtractTimepoints(m.textInput.Value())
	if err != nil {
		m.err = err
		return
	}

	for _, s := range m.Settings.SelectedSymbols.ValueArr {
		path := fmt.Sprintf(
			"internal/testdata/%s_%dm_%s_%s.csv",
			s,
			globals.Timeframe,
			start.Format("02-01-2006"),
			end.Format("02-01-2006"),
		)
		f, err := os.Open(path)
		if err != nil {
			m.err = err
			return
		}
		defer f.Close()

		reader := csv.NewReader(f)
		records, err := reader.ReadAll()
		if err != nil {
			m.err = err
			return
		}

		klines := []*binance.Kline{}
		for _, record := range records {
			openTime, _ := strconv.ParseInt(record[0], 10, 64)
			closeTime, _ := strconv.ParseInt(record[6], 10, 64)
			tradeNum, _ := strconv.ParseInt(record[8], 10, 64)

			kline := &binance.Kline{
				OpenTime:                 openTime,
				Open:                     record[1],
				High:                     record[2],
				Low:                      record[3],
				Close:                    record[4],
				Volume:                   record[5],
				CloseTime:                closeTime,
				QuoteAssetVolume:         record[7],
				TradeNum:                 tradeNum,
				TakerBuyBaseAssetVolume:  record[9],
				TakerBuyQuoteAssetVolume: record[10],
			}

			klines = append(klines, kline)
		}

		klinesFeed[s] = klines
	}

	batchLimit := 60
	client := binancew.NewClientBacktest(start, end, klinesFeed, batchLimit)
	tickerChan := make(chan time.Time)
	btModel := Autobinance{
		Client:     client,
		Settings:   m.Settings,
		TickerChan: tickerChan,
	}

	w := output.NewWriterCreator().CreateWriter(output.Stub)
	btModel.StartTradingSession(w)

	klinesLen := len(klinesFeed[m.Settings.SelectedSymbols.ValueArr[0]])
	for i := 0; i < klinesLen-batchLimit; i++ {
		tickerChan <- time.Now()
		binancew.BacktestIndex++
	}

	binancew.BacktestIndex = 0
	m.info = "Backtesting successful. Check analyses to see results."
}

func (m *Autobinance) QuitApp() (tea.Model, tea.Cmd) {
	m.quitting = true

	return m, tea.Quit
}

// The main view, which just calls the appropriate sub-view
func (m Autobinance) View() string {
	var errMsg string
	if m.err != nil {
		errMsg = m.err.Error()
	}

	err := errStyle.Render(errMsg)
	help := helpStyle.Render(m.help)
	border := borderStyle.Render(" ───────────────────────────────────────────")

	return wordwrap.String(
		indent.String(
			"\n"+chosenView(m)+"\n\n"+m.textInput.View()+"\n\n"+border+"\n"+m.info+"\n"+err+"\n"+help,
			4,
		),
		m.width,
	)
}

func chosenView(m Autobinance) string {
	var msg string

	switch m.choice {
	case "root":
		tradingStatus := "OFF"
		if m.tradingRunning {
			tradingStatus = "ON"
		}

		simulationStatus := ""
		if globals.SimulationMode {
			simulationStatus = "THE GO-BINANCE WRAPPER IS IN SIMULATION MODE\n"
		}

		msg = fmt.Sprint(
			simulationStatus,
			"AUTOBINANCE", "\n",
			"Trading status: ", tradingStatus, "\n\n",
			"1) Start trading session", "\n",
			"2) Set strategies", "\n",
			"3) Set trade symbols", "\n",
			"4) Write analyses to log", "\n",
			"5) Check account", "\n",
			"6) Clear logs and trade history", "\n",
			"7) Recreate tables", "\n",
			"8) Download testdata", "\n",
			"9) Run backtest", "\n",
			"10) Quit trading session",
		)
	// Most of these don't really need to be a separate view btw. Use model.info more.
	case "1":
		msg = fmt.Sprint("Trading session started (you can still do other actions).")
	case "2":
		msg = fmt.Sprint(
			"Available strategies: ", m.Settings.AvailableStrategies.Value, "\n",
			"Currently selected strategies: ", m.Settings.SelectedStrategies.Value, "\n",
			"Enter new strategy set (ex. example):",
		)
	case "3":
		msg = fmt.Sprint(
			"Currently selected symbols: ", m.Settings.SelectedSymbols.Value, "\n",
			"Enter new symbols set (ex. LTCBTC,ETHBTC):",
		)
	case "4", "5":
		msg = fmt.Sprint("Output in log_misc.txt")
	case "6":
		msg = fmt.Sprint("Logs cleared")
	case "7":
		msg = fmt.Sprint("All tables have been recreated")
	case "8":
		msg = fmt.Sprint(
			"Testdata will be downloaded for next symbols:", "\n",
			m.Settings.SelectedSymbols.Value, "\n",
			"Enter the desired period of time (ex. 01-02-2021 30-03-2021):",
		)
	case "9":
		msg = fmt.Sprint(
			"Backtesting will be done for next strategies-symbols:", "\n",
			m.Settings.SelectedStrategies.Value, "\n",
			m.Settings.SelectedSymbols.Value, "\n",
			"Enter the period for backtesting (ex. 01-02-2021 30-03-2021):",
		)
	case "10":
		msg = fmt.Sprint("Trading session stopped.")
	default:
		msg = fmt.Sprint("Invalid choice")
	}

	return msg
}
