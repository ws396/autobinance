package output

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/ws396/autobinance/modules/strategies"
	"github.com/xuri/excelize/v2"
)

// Looking a bit too prematurely convoluted, but let's try to do this the "right" way :^)
// https://github.com/AlexanderGrom/go-patterns/blob/master/Creational/FactoryMethod/factory_method.go

type action string

const (
	Txt   action = "txt"
	Excel action = "excel"
)

type Creator interface {
	CreateWriter(action action) Writer
}

type Writer interface {
	WriteToLog(map[string]string)
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

func (p *TxtWriter) WriteToLog(data map[string]string) {
	f, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_APPEND, 0644)
	if errors.Is(err, os.ErrNotExist) {
		f, err = os.Create("log.txt")
		if err != nil {
			log.Fatalln(err)
			return
		}
	}
	defer f.Close()

	message := fmt.Sprint(
		"----------", "\n",
	)
	for _, v := range *strategies.Datakeys[data["Strategy"]] {
		message += fmt.Sprint(v, ": ", data[v], "\n")
	}
	_, err = f.WriteString(message)
	if err != nil {
		log.Fatalf("Got error while writing to a file. Err: %s", err.Error())
	}
}

type ExcelWriter struct {
}

func (p *ExcelWriter) WriteToLog(data map[string]string) {
	f, err := excelize.OpenFile("log.xlsx")
	if errors.Is(err, os.ErrNotExist) {
		f = excelize.NewFile()

		/*
			for strategy := range strategies.Datakeys {
				i := 0
				f.NewSheet(strategy)
				for _, v := range *strategies.Datakeys[strategy] {
					f.SetCellValue(strategy, string(rune('A'+i))+"1", v)
					i++
				}
			}
		*/

		f.NewSheet(data["Strategy"])

		i := 0
		for _, v := range *strategies.Datakeys[data["Strategy"]] { // There is no need to pass dataKeys
			f.SetCellValue(data["Strategy"], string(rune('A'+i))+"1", v)
			f.SetCellValue(data["Strategy"], string(rune('A'+i))+"2", data[v])
			i++
		}

		f.DeleteSheet("Sheet1") // ?

		if err := f.SaveAs("log.xlsx"); err != nil {
			log.Panicln(err)
			return
		}

		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Panicln(err)
		}
	}()

	rows, _ := f.GetRows(data["Strategy"])
	lastRow := fmt.Sprint(len(rows) + 1)

	if f.GetSheetIndex(data["Strategy"]) == -1 {
		f.NewSheet(data["Strategy"])

		i := 0
		for _, v := range *strategies.Datakeys[data["Strategy"]] { // There is no need to pass dataKeys
			f.SetCellValue(data["Strategy"], string(rune('A'+i))+"1", v)
			f.SetCellValue(data["Strategy"], string(rune('A'+i))+"2", data[v])
			i++
		}

		f.Save()

		return
	}

	i := 0
	for _, v := range *strategies.Datakeys[data["Strategy"]] {
		f.SetCellValue(data["Strategy"], string(rune('A'+i))+lastRow, data[v])
		i++
	}

	f.Save()
}
