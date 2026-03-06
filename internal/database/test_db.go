package database

import (
	"log"
	"os"

	_ "github.com/lib/pq"
)

func ConnectTestDB() {
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		log.Println("[TEST] TEST_DATABASE_URL not set — DB tests will be skipped")
		DB = nil
		return
	}

	if err := Connect(testDBURL); err != nil {
		log.Printf("[TEST] Failed to connect to test DB: %v — DB tests will be skipped", err)
		DB = nil
		return
	}

	log.Println("[TEST] Connected to test database")
}

func CleanupTestDB() {
	if DB == nil {
		return
	}

	queries := []string{
		`DELETE FROM reminders WHERE user_id IN (
			SELECT id FROM users WHERE email LIKE '%@test.example.com'
		)`,
		`DELETE FROM users WHERE email LIKE '%@test.example.com'`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			log.Printf("[TEST] CleanupTestDB warning: %v", err)
		}
	}

	log.Println("[TEST] Test data cleaned up")
}
