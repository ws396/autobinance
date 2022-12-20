package download

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ws396/autobinance/internal/globals"
)

// Limited to 500 candles per request. This approach requires batch processing.
// Could possibly use timestamps as correlation ids for concurrent usage?
/*
func KlinesCSVFromAPI(filename, symbol string, timeframe uint, start, end time.Time) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(f)

	err = writer.Write([]string{
		"opentime", "open", "close", "low", "high", "volume",
	})
	if err != nil {
		return err
	}

	client := binancew.NewExtClientSim("", "")
	klines, err := client.GetKlinesByPeriod(symbol, timeframe, start, end)
	if err != nil {
		return err
	}

	for _, kline := range klines {
		data := []string{
			time.Unix(0, kline.OpenTime*int64(time.Millisecond)).Format("02-01-2006, 15:04:05"),
			kline.Open,
			kline.Close,
			kline.Low,
			kline.High,
			kline.Volume,
		}

		err := writer.Write(data)
		if err != nil {
			return err
		}
	}

	writer.Flush()

	return nil
}
*/

// Works better with full months.
//
// Endpoint formats:
//
// <base_url>/data/spot/monthly/klines/<symbol_in_uppercase>/<interval>/<symbol_in_uppercase>-<interval>-<year>-<month>.zip
// <base_url>/data/spot/daily/klines/<symbol_in_uppercase>/<interval>/<symbol_in_uppercase>-<interval>-<year>-<month>-<day>.zip
func KlinesCSVFromZips(symbols []string, timeframe string, start, end time.Time) error {
	wg := &sync.WaitGroup{}
	errChan := make(chan error)
	errChanOuter := make(chan error)

	for _, symbol := range symbols {
		wg.Add(1)
		go func(symbol string) {
			var filename, url string

			filepaths := []string{}
			urls := []string{}
			path := "internal/testdata/"
			timepoint := start

			for !timepoint.After(end) {
				if timepoint.Day() != 1 || timepoint.Month() == end.Month() {
					filename = fmt.Sprintf("%s-%s-%s.zip",
						symbol,
						timeframe,
						timepoint.Format("2006-01-02"),
					)

					url = fmt.Sprintf(
						"https://data.binance.vision/data/spot/daily/klines/%s/%s/%s",
						symbol,
						timeframe,
						filename,
					)

					timepoint = timepoint.AddDate(0, 0, 1)
				} else {
					filename = fmt.Sprintf("%s-%s-%s.zip",
						symbol,
						timeframe,
						timepoint.Format("2006-01"),
					)

					url = fmt.Sprintf(
						"https://data.binance.vision/data/spot/monthly/klines/%s/%s/%s",
						symbol,
						timeframe,
						filename,
					)

					timepoint = timepoint.AddDate(0, 1, 0)
				}

				filepaths = append(filepaths, path+filename)
				urls = append(urls, url)
			}

			for i := 0; i < len(filepaths) && i < len(urls); i++ {
				go func(f, url string) {
					errChan <- downloadFile(f, url)
				}(filepaths[i], urls[i])
			}
			for i := 0; i < len(filepaths) && i < len(urls); i++ {
				if err := <-errChan; err != nil {
					errChanOuter <- err
				}
			}

			wg.Done()
			wg.Wait()

			// May be better to do this separately
			err := extractKlinesToCSV(
				filepaths,
				fmt.Sprintf(
					"%s%s_%s_%s_%s.csv",
					path,
					symbol,
					timeframe,
					start.Format("02-01-2006"),
					end.Format("02-01-2006"),
				),
			)
			if err != nil {
				errChanOuter <- err
			}

			for _, filepath := range filepaths {
				err = os.Remove(filepath)
				if err != nil {
					errChanOuter <- err
				}
			}

			errChanOuter <- nil
		}(symbol)
	}

	for i := 0; i < len(symbols); i++ {
		if err := <-errChanOuter; err != nil {
			return err
		}
	}

	return nil
}

func extractKlinesToCSV(srcs []string, dest string) error {
	allRecords := [][]string{}

	for _, src := range srcs {
		r, err := zip.OpenReader(src)
		if err != nil {
			return err
		}
		defer r.Close()

		for _, f := range r.File {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			reader := csv.NewReader(rc)
			records, err := reader.ReadAll()
			if err != nil {
				return err
			}

			allRecords = append(allRecords, records...)
		}
	}

	os.MkdirAll(filepath.Dir(dest), 0755)

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	for _, candle := range allRecords {
		err := w.Write(candle)
		if err != nil {
			return err
		}
	}

	w.Flush()

	return nil
}

func downloadFile(dest string, url string) error {
	os.MkdirAll(filepath.Dir(dest), 0755)

	// Create the file
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return globals.ErrCouldNotDownloadFile
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
