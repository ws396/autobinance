package download

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ws396/autobinance/internal/globals"
)

func TestDownload(t *testing.T) {
	t.Run("connects to backtest data provider", func(t *testing.T) {
		res, _ := http.Get(globals.BacktestDataBaseURL) // Add timeout?

		got := res.StatusCode
		want := 200

		if got != want {
			t.Errorf("could not connect to backtest data provider, got %v want %v", got, want)
		}
	})
}

func TestKlinesCSVFromZips(t *testing.T) {
	rootpath, _ := os.Getwd() // Might want to deal with paths some other way
	rootpath += "/../../"
	globals.BacktestDataDir = ""
	globals.BacktestDataBaseURL = "/"
	symbols := []string{
		"BTCBUSD",
		"LTCBTC",
	}
	timeframe := "1m"
	start := time.Date(2022, 12, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2022, 12, 21, 0, 0, 0, 0, time.UTC)

	t.Run("downloads and extracts klines to csv from zip files", func(t *testing.T) {
		mux := http.NewServeMux()

		for _, s := range symbols {
			filenames, urls := generateFilepathsAndURLs(s, timeframe, start, end)

			for i := 0; i < len(filenames) && i < len(urls); i++ {
				url := urls[i]
				filename := filenames[i]
				mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
					filepath := rootpath + globals.TestDataDir + "/test_" + filename
					data, err := ioutil.ReadFile(filepath)
					if err != nil {
						t.Errorf("failed to read file %s", err)
					}

					http.ServeContent(w, r, filepath, time.Now(), bytes.NewReader(data))
				})
			}
		}

		ts := httptest.NewServer(mux)
		defer ts.Close()

		globals.BacktestDataDir = t.TempDir() + "/"
		globals.BacktestDataBaseURL = ts.URL + "/"
		err := KlinesCSVFromZips(symbols, timeframe, start, end)
		if err != nil {
			t.Errorf("failed to generate csv %s", err)
		}

		for _, s := range symbols {
			filename := filepath.Base(getCSVPath(s, timeframe, start, end))
			filepath1 := globals.BacktestDataDir + filename
			filepath2 := rootpath + globals.TestDataDir + "/test_" + filename
			file1, err := os.Open(filepath1)
			if err != nil {
				t.Error(err)
			}
			file2, err := os.Open(filepath2)
			if err != nil {
				t.Error(err)
			}

			fileScanner1 := bufio.NewScanner(file1)
			fileScanner2 := bufio.NewScanner(file2)

			for fileScanner1.Scan() && fileScanner2.Scan() {
				got := fileScanner1.Text()
				want := fileScanner2.Text()

				if got != want {
					t.Errorf("wrong csv file generated on symbol %s, got %v want %v", s, got, want)
				}
			}

			if fileScanner1.Scan() || fileScanner2.Scan() {
				t.Errorf("wrong amount of rows generated on symbol %s", s)
			}
		}
	})
}
