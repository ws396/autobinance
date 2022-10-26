package main

import (
	"testing"
	"time"

	"github.com/ws396/autobinance/modules/binancew-sim"
	"github.com/ws396/autobinance/modules/orders"
	"github.com/ws396/autobinance/modules/settings"
)

type stubWriter struct {
	data *orders.Order
}

func (p *stubWriter) WriteToLog(ch chan *orders.Order) {
	/* 	for i := 0; i < cap(ch); i++ {
		data := <-ch
	} */ // Maybe should also check if it returned only one message through channel

	p.data = <-ch
}

func TestTrade(t *testing.T) {
	t.Run("successfully starts trading session and attempts one trade", func(t *testing.T) {
		client = binancew.NewExtClient("", "") // This prooobably shouldn't be done like this

		model := model{
			settings: settings.Settings{
				SelectedSymbols:    settings.Setting{ID: 0, Name: "selected_symbols", Value: "LTCBTC"},
				SelectedStrategies: settings.Setting{ID: 0, Name: "selected_strategies", Value: "example"},
			},
			ticker: time.NewTicker(time.Millisecond),
		}

		w := &stubWriter{}
		model.startTradingSession(w)

		time.Sleep(12 * time.Second)
		got := w.data.Symbol
		want := "LTCBTC"

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})
}
