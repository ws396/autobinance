package output

import (
	"errors"
	"fmt"
	"log"
	"os"

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
			log.Fatal(err)
			return
		}
	}
	defer f.Close()

	message := fmt.Sprint(
		"----------", "\n",
	)
	for k, v := range data {
		message += fmt.Sprint(k, ": ", v, "\n")
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

		i := 0
		for k := range data {
			f.SetCellValue("Sheet1", string(rune('A'+i))+"1", k)
			i++
		}

		if err := f.SaveAs("log.xlsx"); err != nil {
			fmt.Println(err)
			return
		}

		f, err = excelize.OpenFile("log.xlsx")
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	rows, _ := f.GetRows("Sheet1")
	lastRow := fmt.Sprint(len(rows) + 1)

	i := 0
	for _, v := range rows[0] {
		f.SetCellValue("Sheet1", string(rune('A'+i))+lastRow, data[v])
		i++
	}

	f.Save()
}
