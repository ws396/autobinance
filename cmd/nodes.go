package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ws396/autobinance/internal/analysis"
	"github.com/ws396/autobinance/internal/backtest"
	"github.com/ws396/autobinance/internal/download"
	"github.com/ws396/autobinance/internal/globals"
	"github.com/ws396/autobinance/internal/output"
	"github.com/ws396/autobinance/internal/util"
)

type ViewNode struct {
	view   func(*CLI) string
	action func(*CLI) *ViewNode
}

var (
	root   *ViewNode
	root_1 *ViewNode
	root_2 *ViewNode
	root_3 *ViewNode
	root_4 *ViewNode
	root_5 *ViewNode
	root_6 *ViewNode
	root_7 *ViewNode
	root_8 *ViewNode
	root_9 *ViewNode
	//root_10 *ViewNode
)

func init() {
	root = &ViewNode{
		view: func(m *CLI) string {
			tradingStatus := "OFF"
			if m.T.TradingRunning {
				tradingStatus = "ON"
			}

			simulationStatus := ""
			if globals.SimulationMode {
				simulationStatus = "THE GO-BINANCE WRAPPER IS IN SIMULATION MODE\n"
			}

			msg := fmt.Sprint(
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
			return msg
		},
		action: func(m *CLI) *ViewNode {
			switch m.textInput.Value() {
			case "1":
				w, err := output.NewWriterCreator().CreateWriter(output.Excel)
				if err != nil {
					m.HandleError(err)
					return nil
				}

				errChan := make(chan error)
				m.T.StartTradingSession(w, errChan)

				go func() {
					if err := <-errChan; err != nil {
						errInner := m.T.StopTradingSession()
						if errInner != nil {
							m.HandleError(errInner)
						}

						m.HandleError(err)
					}
				}()

				return root_1
			case "2":
				var err error
				m.T.Settings["selected_strategies"], err = m.T.StorageClient.GetSetting(m.T.Settings["selected_strategies"].Name)
				if err != nil {
					m.HandleError(err)
					return nil
				}

				return root_2
			case "3":
				var err error
				m.T.Settings["selected_symbols"], err = m.T.StorageClient.GetSetting(m.T.Settings["selected_symbols"].Name)
				if err != nil {
					m.HandleError(err)
					return nil
				}

				return root_3
			case "4":
				foundOrders, err := m.T.StorageClient.GetAllOrders()
				if err != nil {
					m.HandleError(err)
					return nil
				}

				if len(foundOrders) == 0 {
					m.HandleError(globals.ErrEmptyOrderList)
					return nil
				}

				analyses := analysis.CreateAnalyses(foundOrders, foundOrders[0].CreatedAt, foundOrders[len(foundOrders)-1].CreatedAt)
				err = m.T.StorageClient.StoreAnalyses(analyses)
				if err != nil {
					m.HandleError(err)
					return nil
				}

				util.WriteToLogMisc(analyses)
				return root_4
			case "5":
				util.WriteToLogMisc(m.T.ExchangeClient.GetCurrencies())
				return root_5
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
				return root_6
			case "7":
				// Confirmation would be nice...
				m.T.StorageClient.DropAll()
				m.T.StorageClient.AutoMigrateAll()
				return root_7
			case "8":
				return root_8
			case "9":
				return root_9
			case "10":
				err := m.T.StopTradingSession()
				if err != nil {
					m.HandleError(err)
					return nil
				}

				return root
			default:
				m.info = "Invalid choice"
			}

			return nil
		},
	}

	root_1 = &ViewNode{
		view: func(m *CLI) string {
			return fmt.Sprint(
				"Trading session started (press Enter to go back to root).",
			)
		},
		action: func(m *CLI) *ViewNode {
			return root
		},
	}

	root_2 = &ViewNode{
		view: func(m *CLI) string {
			return fmt.Sprint(
				"Available strategies: ", m.T.Settings["available_strategies"].Value, "\n",
				"Currently selected strategies: ", m.T.Settings["selected_strategies"].Value, "\n",
				"Enter new strategy set (ex. example):",
			)
		},
		action: func(m *CLI) *ViewNode {
			selectedStrategies := strings.Split(m.textInput.Value(), ",")

			for _, v := range selectedStrategies {
				if !util.Contains(m.T.Settings["available_strategies"].ValueArr, v) {
					m.err = globals.ErrWrongStrategyName
					return nil
				}
			}

			var err error
			m.T.Settings["selected_strategies"], err = m.T.StorageClient.UpdateSetting(m.T.Settings["selected_strategies"].Name, m.textInput.Value())
			if err != nil {
				m.HandleError(err)
				return nil
			}

			return root
		},
	}

	root_3 = &ViewNode{
		view: func(m *CLI) string {
			return fmt.Sprint(
				"Currently selected symbols: ", m.T.Settings["selected_symbols"].Value, "\n",
				"Enter new symbols set (ex. LTCBTC,ETHBTC):",
			)
		},
		action: func(m *CLI) *ViewNode {
			allSymbols := m.T.ExchangeClient.GetAllSymbols()
			selectedSymbols := strings.Split(m.textInput.Value(), ",")

			for _, v := range selectedSymbols {
				if !util.Contains(allSymbols, v) {
					m.err = globals.ErrWrongSymbol
					return nil
				}
			}

			var err error
			m.T.Settings["selected_symbols"], err = m.T.StorageClient.UpdateSetting(m.T.Settings["selected_symbols"].Name, m.textInput.Value())
			if err != nil {
				m.HandleError(err)
				return nil
			}
			return root
		},
	}

	root_4 = &ViewNode{
		view: func(m *CLI) string {
			return fmt.Sprint("Output in log_misc.txt")
		},
		action: func(m *CLI) *ViewNode {
			return root
		},
	}

	root_5 = &ViewNode{
		view: func(m *CLI) string {
			return fmt.Sprint("Output in log_misc.txt")
		},
		action: func(m *CLI) *ViewNode {
			return root
		},
	}

	root_6 = &ViewNode{
		view: func(m *CLI) string {
			return fmt.Sprint("Logs cleared")
		},
		action: func(m *CLI) *ViewNode {
			return root
		},
	}

	root_7 = &ViewNode{
		view: func(m *CLI) string {
			return fmt.Sprint("All tables have been recreated")
		},
		action: func(m *CLI) *ViewNode {
			return root
		},
	}

	root_8 = &ViewNode{
		view: func(m *CLI) string {
			return fmt.Sprint(
				"Testdata will be downloaded for next symbols:", "\n",
				m.T.Settings["selected_symbols"].Value, "\n",
				"Enter the desired period of time (ex. 01-02-2021 30-03-2021):",
			)
		},
		action: func(m *CLI) *ViewNode {
			if len(m.T.Settings["selected_symbols"].Value) == 0 {
				m.err = globals.ErrSymbolsNotFound
				return nil
			}

			start, end, err := util.ExtractTimepoints(m.textInput.Value())
			if err != nil {
				m.HandleError(err)
				return nil
			}

			// Visualize progress bar for this?

			err = download.KlinesCSVFromZips(m.T.Settings["selected_symbols"].ValueArr, globals.Timeframe, start, end)
			if err != nil {
				m.HandleError(err)
				return nil
			}

			m.info = "Testdata downloaded"

			return root
		},
	}

	root_9 = &ViewNode{
		view: func(m *CLI) string {
			return fmt.Sprint(
				"Backtesting will be done for next strategies-symbols:", "\n",
				m.T.Settings["selected_strategies"].Value, "\n",
				m.T.Settings["selected_symbols"].Value, "\n",
				"Enter the period for backtesting (ex. 01-02-2021 30-03-2021):",
			)
		},
		action: func(m *CLI) *ViewNode {
			analyses, err := backtest.Backtest(m.textInput.Value(), m.T.Settings)
			if err != nil {
				m.HandleError(err)
				return nil
			}

			err = m.T.StorageClient.StoreAnalyses(analyses)
			if err != nil {
				m.HandleError(err)
				return nil
			}

			m.info = "Backtesting successful. Analyses written to storage and log_misc."

			return root
		},
	}

	/*
		root_10 = &ViewNode{
			view: func(m *CLI) string {
				return fmt.Sprint("Trading session stopped.")
			},
			action: func(m *CLI) *ViewNode {
				return root
			},
		}
	*/
}
