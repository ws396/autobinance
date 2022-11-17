package util

import (
	"encoding/json"
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

	f, err := os.OpenFile("log_misc.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	_, err = f.WriteString(time.Now().Format("02-01-2006 15:04:05") + "\n" + string(j) + "\n")
	if err != nil {
		log.Fatalf("Got error while writing to a file. Err: %s", err.Error())
	}
}

func IsRunningInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err != nil {
		return false
	}

	return true
}

func Contains[T comparable](slice []T, el T) bool {
	for _, v := range slice {
		if v == el {
			return true
		}
	}

	return false
}
