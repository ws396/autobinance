package cmd

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sdcoffey/big"
	"gorm.io/driver/postgres"

	"github.com/adshao/go-binance/v2"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
	"github.com/sdcoffey/techan"
	"github.com/ws396/autobinance/internal/analysis"
	"github.com/ws396/autobinance/internal/binancew"
	"github.com/ws396/autobinance/internal/download"
	"github.com/ws396/autobinance/internal/globals"
	"github.com/ws396/autobinance/internal/output"
	"github.com/ws396/autobinance/internal/storage"
	"github.com/ws396/autobinance/internal/strategies"
	"github.com/ws396/autobinance/internal/techanext"
	"github.com/ws396/autobinance/internal/util"
)

var (
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
)

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
	StorageClient  storage.StorageClient
	ExchangeClient binancew.ExchangeClient
	Settings       map[string]storage.Setting
	TickerChan     <-chan time.Time
}

func InitialModel() (*Autobinance, error) {
	ti := textinput.New()
	ti.Focus()
	ti.Width = 80

	ticker := time.NewTicker(globals.Durations[globals.Timeframe] / 6)
	apiKey := os.Getenv("API_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	var exchangeClient binancew.ExchangeClient
	if globals.SimulationMode {
		exchangeClient = binancew.NewExtClientSim(apiKey, secretKey)
	} else {
		exchangeClient = binancew.NewExtClient(apiKey, secretKey)
	}

	dialect := postgres.New(postgres.Config{
		DSN: fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
			os.Getenv("PGSQL_HOST"),
			os.Getenv("PGSQL_PORT"),
			os.Getenv("PGSQL_DB"),
			os.Getenv("PGSQL_USER"),
			os.Getenv("PGSQL_PASS"),
		),
	})
	storageClient, err := storage.NewGORMClient(dialect)
	if err != nil {
		return nil, err
	}

	//storageClient := storage.NewInMemoryClient()
	storageClient.AutoMigrateAll()

	s, err := storageClient.GetAllSettings()
	if err != nil {
		return nil, err
	}

	var keys []string
	for k := range strategies.StrategiesInfo {
		keys = append(keys, k)
	}
	s["available_strategies"], err = storageClient.UpdateSetting(s["available_strategies"].Name, strings.Join(keys, ","))
	if err != nil {
		return nil, err
	}

	return &Autobinance{
		choice:         "root",
		textInput:      ti,
		err:            err,
		help:           "\\q - back to root",
		stopTrading:    make(chan bool),
		StorageClient:  storageClient,
		ExchangeClient: exchangeClient,
		Settings:       s,
		TickerChan:     ticker.C,
	}, nil
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}

	m.textInput, cmd = m.textInput.Update(msg)

	return m, cmd
}

func (m *Autobinance) Logic() {
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
			if !util.Contains(m.Settings["available_strategies"].ValueArr, v) {
				m.err = globals.ErrWrongStrategyName
				return
			}
		}

		var err error
		m.Settings["selected_strategies"], err = m.StorageClient.UpdateSetting(m.Settings["selected_strategies"].Name, m.textInput.Value())
		if err != nil {
			m.HandleError(err)
			return
		}
	case "3":
		allSymbols := m.ExchangeClient.GetAllSymbols()
		selectedSymbols := strings.Split(m.textInput.Value(), ",")

		for _, v := range selectedSymbols {
			if !util.Contains(allSymbols, v) {
				m.err = globals.ErrWrongSymbol
				return
			}
		}

		var err error
		m.Settings["selected_symbols"], err = m.StorageClient.UpdateSetting(m.Settings["selected_symbols"].Name, m.textInput.Value())
		if err != nil {
			m.HandleError(err)
			return
		}
	case "8":
		if len(m.Settings["selected_symbols"].Value) == 0 {
			m.err = globals.ErrSymbolsNotFound
			return
		}

		start, end, err := util.ExtractTimepoints(m.textInput.Value())
		if err != nil {
			m.HandleError(err)
			return
		}

		// Visualize progress bar for this?

		err = download.KlinesCSVFromZips(m.Settings["selected_symbols"].ValueArr, globals.Timeframe, start, end)
		if err != nil {
			m.HandleError(err)
			return
		}

		m.info = "Testdata downloaded"
	case "9":
		m.Backtest()
	}

	// Choice transition
	switch m.choice {
	case "root":
		m.choice = m.textInput.Value()
	default:
		m.choice = "root"
	}

	// Need a sensible way to return these to root. Either way, rework of this logic is inevitable.
	// New choice pre-render logic
	switch m.choice {
	case "1":
		w, err := output.NewWriterCreator().CreateWriter(output.Excel)
		if err != nil {
			m.HandleError(err)
			return
		}

		err = m.StartTradingSession(w)
		if err != nil {
			m.HandleError(err)
			return
		}
	case "2":
		var err error
		m.Settings["selected_strategies"], err = m.StorageClient.GetSetting(m.Settings["selected_strategies"].Name)
		if err != nil {
			m.HandleError(err)
			return
		}
	case "3":
		var err error
		m.Settings["selected_symbols"], err = m.StorageClient.GetSetting(m.Settings["selected_symbols"].Name)
		if err != nil {
			m.HandleError(err)
			return
		}
	case "4":
		foundOrders, err := m.StorageClient.GetAllOrders()
		if err != nil {
			m.HandleError(err)
			return
		}

		if len(foundOrders) == 0 {
			m.HandleError(globals.ErrEmptyOrderList)
			return
		}

		analyses := analysis.CreateAnalyses(foundOrders, foundOrders[0].CreatedAt, foundOrders[len(foundOrders)-1].CreatedAt)

		err = m.StorageClient.StoreAnalyses(analyses)
		if err != nil {
			m.HandleError(err)
			return
		}

		util.WriteToLogMisc(analyses)
	case "5":
		util.WriteToLogMisc(m.ExchangeClient.GetCurrencies())
	case "6":
		filesToRemove := []string{
			output.Filename + ".txt",
			"log_gorm.txt",
			"log_misc.txt",
			"log_error.txt",
		}
		for _, path := range filesToRemove {
			os.Truncate(path, 0)
		}

		os.Remove(output.Filename + ".xlsx")
	case "7":
		// Confirmation would be nice...
		m.StorageClient.DropAll()
		m.StorageClient.AutoMigrateAll()
	case "10":
		m.StopTradingSession()
	}
}

func (m *Autobinance) HandleError(err error) {
	util.Logger.Error(err.Error())
	m.err = err
}

func (m *Autobinance) HandleFatal(err error) {
	util.Logger.Fatal(err.Error())
}

func (m *Autobinance) StartTradingSession(w output.Writer) error {
	if m.tradingRunning {
		return globals.ErrTradingAlreadyRunning
	}
	if len(m.Settings["selected_strategies"].ValueArr) == 0 {
		return globals.ErrStrategiesNotFound
	}
	if len(m.Settings["selected_symbols"].ValueArr) == 0 {
		return globals.ErrSymbolsNotFound
	}

	m.tradingRunning = true
	chanSize := len(m.Settings["selected_strategies"].ValueArr) * len(m.Settings["selected_symbols"].ValueArr)
	errs := make(chan error)
	ordersChan := make(chan *storage.Order)

	go func() {
		for {
			select {
			case <-m.TickerChan:
				for _, symbol := range m.Settings["selected_symbols"].ValueArr {
					go func(symbol string) {
						klines, err := m.ExchangeClient.GetKlines(symbol, globals.Timeframe)
						if err != nil {
							errs <- err
						}

						series := techanext.GetSeries(klines, globals.Durations[globals.Timeframe])
						for _, strategy := range m.Settings["selected_strategies"].ValueArr {
							go func(strategy string) {
								order, err := m.Trade(strategy, symbol, series)
								if err != nil {
									errs <- err
								}

								ordersChan <- order
							}(strategy)
						}
					}(symbol)
				}

				var orders []*storage.Order
				for i := 0; i < chanSize; i++ {
					data := <-ordersChan

					if data != nil {
						orders = append(orders, data)
					}
				}

				err := w.WriteToLog(orders)
				if err != nil {
					errs <- err
				}
			case <-m.stopTrading:
				return
			case err := <-errs:
				m.HandleError(err)
				return
			}
		}
	}()

	return nil
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

func (m *Autobinance) Trade(strategy, symbol string, series *techan.TimeSeries) (*storage.Order, error) {
	decision, indicators := strategies.StrategiesInfo[strategy].Handler(series)
	order := &storage.Order{
		Strategy:   strategy,
		Symbol:     symbol,
		Decision:   decision,
		Quantity:   0,
		Price:      0,
		Indicators: indicators,
		Timeframe:  globals.Timeframe,
		Successful: false,
		CreatedAt:  time.Now(),
	}

	if decision == globals.Hold {
		return order, nil
	}

	foundOrder, err := m.StorageClient.GetLastOrder(strategy, symbol)
	if err != nil && !errors.Is(err, globals.ErrOrderNotFound) {
		return nil, err
	}

	if foundOrder != nil {
		if (foundOrder.Decision == globals.Sell || foundOrder.Decision == "") && decision == globals.Sell {
			//return nil, errors.New("err: no recent buy has been done on this symbol to initiate sell")
			return order, nil
		} else if foundOrder.Decision == globals.Buy && decision == globals.Buy {
			//return nil, errors.New("err: this position is already bought")
			return order, nil
		}
	} else if decision == globals.Sell {
		return order, nil
	}

	assetPrice := series.LastCandle().ClosePrice
	var quantity big.Decimal
	switch decision {
	case globals.Buy:
		quantity = big.NewDecimal(globals.BuyAmount).Div(assetPrice)
	case globals.Sell:
		quantity = big.NewDecimal(foundOrder.Quantity)
	}

	orderPrice := assetPrice.Mul(quantity)
	order.Quantity = quantity.Float()
	order.Price = orderPrice.Float()
	order.Successful = true

	_, err = m.ExchangeClient.CreateOrder(symbol, quantity.String(), orderPrice.String(), binance.SideType(decision))
	if err != nil {
		return nil, err
	}

	err = m.StorageClient.StoreOrder(order)
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
	if len(m.Settings["selected_symbols"].ValueArr) == 0 {
		m.err = globals.ErrSymbolsNotFound
		return
	}

	klinesFeed := map[string][]*binance.Kline{}

	start, end, err := util.ExtractTimepoints(m.textInput.Value())
	if err != nil {
		m.HandleError(err)
		return
	}

	for _, s := range m.Settings["selected_symbols"].ValueArr {
		path := fmt.Sprintf(
			"internal/testdata/%s_%s_%s_%s.csv",
			s,
			globals.Timeframe,
			start.Format("02-01-2006"),
			end.Format("02-01-2006"),
		)
		f, err := os.Open(path)
		if err != nil {
			m.HandleError(err)
			return
		}
		defer f.Close()

		reader := csv.NewReader(f)
		records, err := reader.ReadAll()
		if err != nil {
			m.HandleError(err)
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
	btExchangeClient := binancew.NewClientBacktest(start, end, klinesFeed, batchLimit)
	btStorageClient := storage.NewInMemoryClient()
	tickerChan := make(chan time.Time)
	btModel := Autobinance{
		StorageClient:  btStorageClient,
		ExchangeClient: btExchangeClient,
		Settings:       m.Settings,
		TickerChan:     tickerChan,
	}

	w, err := output.NewWriterCreator().CreateWriter(output.Stub)
	if err != nil {
		m.HandleError(err)
		return
	}

	err = btModel.StartTradingSession(w)
	if err != nil {
		m.HandleError(err)
		return
	}

	klinesLen := len(klinesFeed[m.Settings["selected_symbols"].ValueArr[0]])
	for i := 0; i < klinesLen-batchLimit; i++ {
		tickerChan <- time.Now()
	}

	// I'm not sure if there's much meaning in DRYing this.
	foundOrders, err := btStorageClient.GetAllOrders()
	if err != nil {
		m.HandleError(err)
		return
	}

	analyses := analysis.CreateAnalyses(foundOrders, start, end)

	err = m.StorageClient.StoreAnalyses(analyses)
	if err != nil {
		m.HandleError(err)
		return
	}

	util.WriteToLogMisc(analyses)

	binancew.BacktestIndex = 0
	m.info = "Backtesting successful. Analyses written to storage and log_misc."
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
		msg = fmt.Sprint("Trading session started (press Enter to go back to root).")
	case "2":
		msg = fmt.Sprint(
			"Available strategies: ", m.Settings["available_strategies"].Value, "\n",
			"Currently selected strategies: ", m.Settings["selected_strategies"].Value, "\n",
			"Enter new strategy set (ex. example):",
		)
	case "3":
		msg = fmt.Sprint(
			"Currently selected symbols: ", m.Settings["selected_symbols"].Value, "\n",
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
			m.Settings["selected_symbols"].Value, "\n",
			"Enter the desired period of time (ex. 01-02-2021 30-03-2021):",
		)
	case "9":
		msg = fmt.Sprint(
			"Backtesting will be done for next strategies-symbols:", "\n",
			m.Settings["selected_strategies"].Value, "\n",
			m.Settings["selected_symbols"].Value, "\n",
			"Enter the period for backtesting (ex. 01-02-2021 30-03-2021):",
		)
	case "10":
		msg = fmt.Sprint("Trading session stopped.")
	default:
		msg = fmt.Sprint("Invalid choice")
	}

	return msg
}
