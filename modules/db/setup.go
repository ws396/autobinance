package db

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var Client *gorm.DB

func ConnectDB() {
	database, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		os.Getenv("PGSQL_HOST"),
		os.Getenv("PGSQL_PORT"),
		os.Getenv("PGSQL_DB"),
		os.Getenv("PGSQL_USER"),
		os.Getenv("PGSQL_PASS"),
	))
	if err != nil {
		panic(err)
	}

	Client = database
}
