package main_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ws396/autobinance/internal/binancew"
	"github.com/ws396/autobinance/internal/storage"
	"github.com/ws396/autobinance/internal/trader"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type mockWriter struct {
	dataChan chan []*storage.Order
}

func (p *mockWriter) WriteToLog(orders []*storage.Order) error {
	p.dataChan <- orders

	return nil
}

func TestTrade(t *testing.T) {
	t.Run("successfully starts trading session and attempts one trade", func(t *testing.T) {
		storageClient := &storage.GORMClient{setupMockStorage(t, mockExpect)}
		exchangeClient := binancew.NewExtClientSim("", "")
		tickerChan := make(chan time.Time)
		trader := trader.Trader{
			StorageClient:  storageClient,
			ExchangeClient: exchangeClient,
			Settings: map[string]storage.Setting{
				"selected_symbols": {
					Name:     "selected_symbols",
					Value:    "LTCBTC",
					ValueArr: []string{"LTCBTC"},
				},
				"selected_strategies": {
					Name:     "selected_strategies",
					Value:    "example",
					ValueArr: []string{"example"},
				},
			},
			TickerChan: tickerChan,
		}
		w := &mockWriter{
			dataChan: make(chan []*storage.Order),
		}
		errChan := make(chan error)
		trader.StartTradingSession(w, errChan)
		tickerChan <- time.Now()
		data := <-w.dataChan
		got := data[0].Symbol
		want := "LTCBTC"

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})
}

func mockExpect(mock sqlmock.Sqlmock) {
	// I really don't like the idea of writing raw SQL expectaions to ORM queries, but I'll stick to it for now
	mock.ExpectQuery(
		regexp.QuoteMeta(
			`SELECT * FROM "orders" 
			WHERE strategy = $1 AND symbol = $2 
			ORDER BY "orders"."id" DESC LIMIT 1`,
		),
	).
		WithArgs("example", "LTCBTC").
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(
		regexp.QuoteMeta(
			`INSERT INTO "orders" 
			("strategy","symbol","decision","quantity","price","indicators","timeframe","successful","created_at") 
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "id"`,
		),
	).
		WithArgs(
			"example",
			"LTCBTC",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectCommit()
}

// This will only be used in internal/storage/gorm_test.go. Everything else needs to be tested through in-memory storage.
func setupMockStorage(t *testing.T, expectHandler func(sqlmock.Sqlmock)) *gorm.DB {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	expectHandler(mock)

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
		t.Fatal(err)
	}

	return database
}
