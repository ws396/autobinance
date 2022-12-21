package util

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ws396/autobinance/internal/globals"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitZapLogger() {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	/* 	fileEncoder := zapcore.NewJSONEncoder(config)
	   	logFile, _ := os.OpenFile("log.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) */
	fileEncoder := zapcore.NewConsoleEncoder(config)
	logFile, _ := os.OpenFile("log_error.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	writer := zapcore.AddSync(logFile)
	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewCore(fileEncoder, writer, defaultLogLevel)
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

func ToSnakeCase(str string) string {
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")

	return strings.ToLower(snake)
}

func WriteToLogMisc(data ...interface{}) error {
	j, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	f, err := os.OpenFile("log_misc.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(
		time.Now().Format("02-01-2006 15:04:05") + "\n" + string(j) + "\n",
	)
	if err != nil {
		return err
	}

	return nil
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

func ExtractTimepoints(s string) (time.Time, time.Time, error) {
	timePeriod := strings.Split(s, " ")
	if len(timePeriod) != 2 {
		return time.Time{}, time.Time{}, globals.ErrWrongArgumentAmount
	}
	start, err := time.Parse("02-01-2006", timePeriod[0])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := time.Parse("02-01-2006", timePeriod[1])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	if end.Before(start) {
		return time.Time{}, time.Time{}, globals.ErrWrongDateOrder
	}

	return start, end, nil
}
