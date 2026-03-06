package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var DB *sql.DB

func Connect(databaseURL string) error {
	var err error

	DB, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return err
	}

	// Test the connection
	if err = DB.Ping(); err != nil {
		return err
	}

	log.Println("Database connected")

	return nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
