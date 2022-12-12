package main_test

import (
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ws396/autobinance/cmd"
	"github.com/ws396/autobinance/internal/binancew"
	"github.com/ws396/autobinance/internal/store"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Remove channels entirely?
type mockWriter struct {
	dataChan chan []*store.Order
}

func (p *mockWriter) WriteToLog(orders []*store.Order) {
	p.dataChan <- orders
	log.Println("uh")
}

func TestTrade(t *testing.T) {
	t.Run("successfully starts trading session and attempts one trade", func(t *testing.T) {
		storeClient := &store.GORMClient{setupMockStore(t, mockExpect)}

		exchangeClient := binancew.NewExtClientSim("", "")
		tickerChan := make(chan time.Time)
		model := cmd.Autobinance{
			StoreClient:    storeClient,
			ExchangeClient: exchangeClient,
			Settings: &store.Settings{
				SelectedSymbols:    store.Setting{ID: 0, Name: "selected_symbols", Value: "LTCBTC", ValueArr: []string{"LTCBTC"}},
				SelectedStrategies: store.Setting{ID: 0, Name: "selected_strategies", Value: "example", ValueArr: []string{"example"}},
			},
			TickerChan: tickerChan,
		}
		w := &mockWriter{
			dataChan: make(chan []*store.Order),
		}

		model.StartTradingSession(w)

		tickerChan <- time.Now()
		data := <-w.dataChan
		//log.Println(data)
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
		regexp.QuoteMeta(`SELECT * FROM "orders" WHERE strategy = $1 AND symbol = $2 ORDER BY "orders"."id" DESC LIMIT 1`)).
		WithArgs("example", "LTCBTC").
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectQuery(
		regexp.QuoteMeta(`INSERT INTO "orders" ("strategy","symbol","decision","quantity","price","indicators","timeframe","created_at") 
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "id"`)).
		WithArgs("example", "LTCBTC", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectCommit()
}

// Will likely end up being used in several other test files
func setupMockStore(t *testing.T, expectHandler func(sqlmock.Sqlmock)) *gorm.DB {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	//defer dbMock.Close()

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
		log.Panicln(err)
	}

	return database
}
