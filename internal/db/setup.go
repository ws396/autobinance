package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ws396/autobinance/internal/util"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Client *gorm.DB

func ConnectDB() {
	f, err := util.OpenOrCreateFile("log_gorm.txt")
	if err != nil {
		log.Fatalln(err)
	}

	newLogger := logger.New(
		log.New(f, "\r\n", log.LstdFlags), // Could put something related to model.err in here?
		logger.Config{
			SlowThreshold:             time.Second,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
			LogLevel:                  logger.Info,
		},
	)

	database, err := gorm.Open(postgres.New(postgres.Config{
		DSN: fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
			os.Getenv("PGSQL_HOST"),
			os.Getenv("PGSQL_PORT"),
			os.Getenv("PGSQL_DB"),
			os.Getenv("PGSQL_USER"),
			os.Getenv("PGSQL_PASS"),
		),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage. By default pgx automatically uses the extended protocol
	}), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Panicln(err)
	}

	Client = database
}
