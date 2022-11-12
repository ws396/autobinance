package main_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ws396/autobinance/modules/binancew-sim"
	"github.com/ws396/autobinance/modules/cmd"
	"github.com/ws396/autobinance/modules/db"
	"github.com/ws396/autobinance/modules/orders"
	"github.com/ws396/autobinance/modules/settings"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Remove channels entirely?
type mockWriter struct {
	dataChan chan []*orders.Order
}

func (p *mockWriter) WriteToLog(orders []*orders.Order) {
	p.dataChan <- orders
}

func TestTrade(t *testing.T) {
	t.Run("successfully starts trading session and attempts one trade", func(t *testing.T) {
		/*
			// Something like this with an actual mocked DB based on migrations is also an option I guess
			dbMock, err := sql.Open("pgx", "postgres://username:password@localhost:5432/test613463")
			if err != nil {
				log.Println(err)
			}
		*/

		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer dbMock.Close()

		// I really don't like the idea of writing raw SQL expectaions to ORM queries, but I'll stick to it for now
		mock.ExpectQuery(
			regexp.QuoteMeta(`SELECT * FROM "orders" WHERE strategy = $1 AND symbol = $2 ORDER BY "orders"."id" DESC LIMIT 1`)).
			WithArgs("example", "LTCBTC").
			WillReturnError(gorm.ErrRecordNotFound)

		mock.ExpectBegin()
		mock.ExpectQuery(
			regexp.QuoteMeta(`INSERT INTO "orders" ("strategy","symbol","decision","quantity","price","indicators","time") 
			VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`)).
			WithArgs("example", "LTCBTC", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectCommit()

		mock.ExpectQuery(
			regexp.QuoteMeta(`SELECT * FROM "analyses" WHERE strategy = $1 AND symbol = $2 ORDER BY "analyses"."id" LIMIT 1`)).
			WithArgs("example", "LTCBTC").
			WillReturnError(gorm.ErrRecordNotFound)

		mock.ExpectBegin()
		mock.ExpectQuery(
			regexp.QuoteMeta(`INSERT INTO "analyses" ("strategy","symbol","buys","sells","successful_sells","profit_usd","success_rate","timeframe","active_time","created_at","updated_at") 
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "id"`)).
			WithArgs("example", "LTCBTC", 1, 0, 0, sqlmock.AnyArg(), float64(0), sqlmock.AnyArg(), 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectCommit()

		dialector := postgres.New(postgres.Config{
			DSN:                  "sqlmock_db_0",
			DriverName:           "postgres",
			Conn:                 dbMock,
			PreferSimpleProtocol: true,
		})
		database, err := gorm.Open(dialector, &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			log.Panicln(err)
		}
		db.Client = database

		client := binancew.NewExtClient("", "")
		tickerChan := make(chan time.Time, 1)
		model := cmd.Autobinance{
			Client: client,
			Settings: settings.Settings{
				SelectedSymbols:    settings.Setting{ID: 0, Name: "selected_symbols", Value: "LTCBTC"},
				SelectedStrategies: settings.Setting{ID: 0, Name: "selected_strategies", Value: "example"},
			},
			TickerChan: tickerChan,
		}
		w := &mockWriter{
			dataChan: make(chan []*orders.Order, 1),
		}

		model.StartTradingSession(w)
		tickerChan <- time.Now()
		data := <-w.dataChan

		fmt.Println(data)
		got := data[0].Symbol
		want := "LTCBTC"

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})
}
