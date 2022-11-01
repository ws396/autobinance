package util

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

func ToSnakeCase(str string) string {
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")

	return strings.ToLower(snake)
}

func WriteToLogMisc(data ...interface{}) {
	j, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		log.Panicln(err)
	}

	f, err := OpenOrCreateFile("log_misc.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	_, err = f.WriteString(time.Now().Format("02-01-2006 15:04:05") + "\n" + string(j) + "\n")
	if err != nil {
		log.Fatalf("Got error while writing to a file. Err: %s", err.Error())
	}
}

func OpenOrCreateFile(name string) (*os.File, error) {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND, 0644)
	if errors.Is(err, os.ErrNotExist) {
		f, err = os.Create(name)
		if err != nil {
			return nil, err
		}
	}

	return f, err
}
