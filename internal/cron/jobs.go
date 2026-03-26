package cron

import (
	"log"

	"github.com/Marst/reminder-app/internal/database"
)

func (s *Scheduler) ResetDailyReminders() {
	log.Println("Starting : Reset Daily reminders")

	query := `
	UPDATE reminders 
	SET 
		is_completed = false, 
		date = CURRENT_DATE + INTERVAL '1 day' ,
		updated_at = CURRENT_TIMESTAMP
	WHERE 
		recurring = true 
		AND date < CURRENT_DATE
	`

	result, err := database.DB.Exec(query)

	if err != nil {
		log.Printf("Error resetting reminders: %v", err)
		return
	}
	rowsAffected, err := result.RowsAffected()

	if err != nil {
		log.Printf("Could not get rows affected: %v", err)
		return
	}

	log.Printf("✅ Success: Reset %d daily reminders", rowsAffected)
}
