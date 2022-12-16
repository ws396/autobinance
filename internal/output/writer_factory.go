package output

import (
	"errors"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/ws396/autobinance/internal/binancew"
	"github.com/ws396/autobinance/internal/globals"
	"github.com/ws396/autobinance/internal/storage"
	"github.com/ws396/autobinance/internal/strategies"
	"github.com/xuri/excelize/v2"
)

type action string

const (
	Txt      action = "txt"
	Excel    action = "excel"
	Stub     action = "stub"
	Filename string = "trade_attempts"
)

type Creator interface {
	CreateWriter(action action) (Writer, error)
}

type Writer interface {
	WriteToLog([]*storage.Order) error
}

type ConcreteWriterCreator struct {
}

func NewWriterCreator() Creator {
	return &ConcreteWriterCreator{}
}

func (p *ConcreteWriterCreator) CreateWriter(action action) (Writer, error) {
	var w Writer

	switch action {
	case Txt:
		w = &TxtWriter{}
	case Excel:
		w = &ExcelWriter{}
	case Stub:
		w = &StubWriter{}
	default:
		return nil, globals.ErrWriterNotFound
	}

	return w, nil
}

type TxtWriter struct {
}

func (p *TxtWriter) WriteToLog(orders []*storage.Order) error {
	if len(orders) == 0 {
		return globals.ErrEmptyOrderList
	}

	f, err := os.OpenFile(Filename+".txt", os.O_WRONLY|os.O_APPEND, 0644)
	if errors.Is(err, os.ErrNotExist) {
		f, err = os.Create(Filename + ".txt")
		if err != nil {
			return err
		}
	}
	defer f.Close()

	message := fmt.Sprint(
		"----------", "\n",
	)

	for _, data := range orders {
		dataMap := orderToMap(data)

		for _, v := range strategies.StrategiesInfo[data.Strategy].Datakeys {
			message += fmt.Sprint(v, ": ", dataMap[v], "\n")
		}
	}

	_, err = f.WriteString(message)
	if err != nil {
		return err
	}

	return nil
}

type ExcelWriter struct {
}

func (p *ExcelWriter) WriteToLog(orders []*storage.Order) error {
	if len(orders) == 0 {
		return globals.ErrEmptyOrderList
	}
	f, err := excelize.OpenFile(Filename + ".xlsx")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			f = excelize.NewFile()

			for k, v := range strategies.StrategiesInfo {
				f.NewSheet(k)

				pos := 0
				for _, v2 := range v.Datakeys {
					f.SetCellValue(k, string(rune('A'+pos))+"1", v2)
					pos++
				}
			}

			f.DeleteSheet("Sheet1")

			if err := f.SaveAs(Filename + ".xlsx"); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer f.Close()

	for _, data := range orders {
		rows, _ := f.GetRows(data.Strategy)
		lastRow := fmt.Sprint(len(rows) + 1)
		dataMap := orderToMap(data)
		pos := 0

		for _, v := range strategies.StrategiesInfo[data.Strategy].Datakeys {
			f.SetCellValue(data.Strategy, string(rune('A'+pos))+lastRow, dataMap[v])
			pos++
		}
	}

	f.Save()

	return nil
}

type StubWriter struct {
}

func (p *StubWriter) WriteToLog(orders []*storage.Order) error {
	// Should probably be somewhere else...
	atomic.AddInt64(&binancew.BacktestIndex, 1)

	return nil
}

func orderToMap(data *storage.Order) map[string]string {
	dataMap := data.Indicators
	dataMap["Current price"] = fmt.Sprintf("%v", data.Price)
	dataMap["Created at"] = data.CreatedAt.Format("02-01-2006 15:04:05")
	dataMap["Symbol"] = data.Symbol
	dataMap["Decision"] = data.Decision
	dataMap["Strategy"] = data.Strategy
	dataMap["Successful"] = fmt.Sprint(data.Successful)

	return dataMap
}
