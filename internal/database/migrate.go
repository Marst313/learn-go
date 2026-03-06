package database

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // postgres driver
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dbURL string, migrationPath string) error {
	log.Println("Running database migrations...")

	m, err := migrate.New(
		"file://"+migrationPath,
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("Could not create migrate instance :%w", err)
	}

	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			log.Println("Database already up to date")
			return nil
		}
		return fmt.Errorf("Could not run migrations : %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}
