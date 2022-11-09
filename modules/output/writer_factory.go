package output

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/ws396/autobinance/modules/orders"
	"github.com/ws396/autobinance/modules/strategies"
	"github.com/xuri/excelize/v2"
)

// Based on:
// https://github.com/AlexanderGrom/go-patterns/blob/master/Creational/FactoryMethod/factory_method.go

type action string

const (
	Txt      action = "txt"
	Excel    action = "excel"
	Filename string = "trade_history"
)

type Creator interface {
	CreateWriter(action action) Writer
}

type Writer interface {
	WriteToLog(chan *orders.Order)
}

type ConcreteWriterCreator struct {
}

func NewWriterCreator() Creator {
	return &ConcreteWriterCreator{}
}

func (p *ConcreteWriterCreator) CreateWriter(action action) Writer {
	var product Writer

	switch action {
	case Txt:
		product = &TxtWriter{}
	case Excel:
		product = &ExcelWriter{}
	default:
		log.Fatalln("Unknown Action")
	}

	return product
}

type TxtWriter struct {
}

func (p *TxtWriter) WriteToLog(ch chan *orders.Order) {
	f, err := os.OpenFile(Filename+".txt", os.O_WRONLY|os.O_APPEND, 0644)
	if errors.Is(err, os.ErrNotExist) {
		f, err = os.Create(Filename + ".txt")
		if err != nil {
			log.Fatalln(err)
			return
		}
	}
	defer f.Close()

	message := fmt.Sprint(
		"----------", "\n",
	)

	for i := 0; i < cap(ch); i++ {
		data := <-ch
		dataMap := orderToMap(data)

		for _, v := range strategies.StrategiesInfo[data.Strategy].Datakeys {
			message += fmt.Sprint(v, ": ", dataMap[v], "\n")
		}
	}

	_, err = f.WriteString(message)
	if err != nil {
		log.Fatalf("Got error while writing to a file. Err: %s", err.Error())
	}
}

type ExcelWriter struct {
}

func (p *ExcelWriter) WriteToLog(ch chan *orders.Order) {
	f, err := excelize.OpenFile(Filename + ".xlsx")
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
			log.Panicln(err)
			return
		}
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Panicln(err)
		}
	}()

	for i := 0; i < cap(ch); i++ {
		data := <-ch
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
}

func orderToMap(data *orders.Order) map[string]string {
	dataMap := data.Indicators
	dataMap["Current price"] = fmt.Sprintf("%v", data.Price)
	dataMap["Time"] = data.Time.Format("02-01-2006 15:04:05")
	dataMap["Symbol"] = data.Symbol
	dataMap["Decision"] = data.Decision
	dataMap["Strategy"] = data.Strategy

	return dataMap
}
