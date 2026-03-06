package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Marst/reminder-app/internal/database"
	"github.com/Marst/reminder-app/internal/models"
	"github.com/Marst/reminder-app/internal/utils"
)

func GetReminders(ctx context.Context, userId int) (*[]models.ReminderResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := database.DB.QueryContext(ctx, `SELECT id, title, description, category, is_completed, priority, recurring, time, date
	FROM reminders
	WHERE user_id = $1`,
		userId)

	if err != nil {
		return nil, errors.New("Failed to fetch reminder")
	}

	defer rows.Close()

	reminders := []models.ReminderResponse{}

	for rows.Next() {
		var r models.ReminderResponse
		err := rows.Scan(
			&r.ID,
			&r.Title,
			&r.Description,
			&r.Category,
			&r.IsCompleted,
			&r.Priority,
			&r.Recurring,
			&r.Time,
			&r.Date,
		)

		if err != nil {
			fmt.Println(err)
			return nil, errors.New("Failed to scan reminders")
		}

		reminders = append(reminders, r)
	}

	return &reminders, nil

}

func NewReminder(ctx context.Context, req *models.Reminder, userID int) (*models.ReminderResponse, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := database.DB.QueryRowContext(ctx, `INSERT INTO reminders (title, description, category, is_completed, priority, recurring, time, date, user_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id, title, description, category, is_completed, priority, recurring, time, date, created_at, updated_at, user_id`,
		&req.Title,
		&req.Description,
		&req.Category,
		&req.IsCompleted,
		&req.Priority,
		&req.Recurring,
		&req.Time,
		&req.Date,
		userID,
	).Scan(
		&req.ID,
		&req.Title,
		&req.Description,
		&req.Category,
		&req.IsCompleted,
		&req.Priority,
		&req.Recurring,
		&req.Time,
		&req.Date,
		&req.CreatedAt,
		&req.UpdatedAt,
		&req.UserID,
	)

	if err != nil {
		_, msg := utils.MapReminderCreateError(err)
		return nil, errors.New(msg)
	}

	resp := models.ReminderResponse{
		Title:       req.Title,
		Description: req.Description,
		IsCompleted: req.IsCompleted,
		Category:    req.Category,
		Priority:    req.Priority,
		Time:        req.Time,
		Recurring:   req.Recurring,
		Date:        req.Date,
	}

	return &resp, nil
}

func UpdateReminder(ctx context.Context, id int64, req *models.Reminder) (*models.Reminder, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var updated models.Reminder

	err := database.DB.QueryRowContext(ctx, `UPDATE reminders
	SET title = $1,
		description = $2,
		date = $3,
		time = $4,
		category = $5,
		priority = $6,
		recurring = $7,
		is_completed = $8
	WHERE id = $9
	RETURNING id, title, description, date, time, category, priority, recurring, is_completed, updated_at

	`, &req.Title,
		&req.Description,
		&req.Date,
		&req.Time,
		&req.Category,
		&req.Priority,
		&req.Recurring,
		&req.IsCompleted,
		id).Scan(&updated.ID,
		&updated.Title,
		&updated.Description,
		&updated.Date,
		&updated.Time,
		&updated.Category,
		&updated.Priority,
		&updated.Recurring,
		&updated.IsCompleted,
		&updated.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("reminder not found")
		}
		return nil, err
	}

	return &updated, nil
}

func DeleteReminder(ctx context.Context, id int64) (*models.Reminder, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var deleted models.Reminder

	err := database.DB.QueryRowContext(ctx, `DELETE FROM reminders 
	WHERE id = $1
	RETURNING id, title, description, date, time, category, priority, recurring, is_completed, updated_at
	`, id).Scan(&deleted.ID,
		&deleted.Title,
		&deleted.Description,
		&deleted.Date,
		&deleted.Time,
		&deleted.Category,
		&deleted.Priority,
		&deleted.Recurring,
		&deleted.IsCompleted,
		&deleted.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Reminder not found")
		}
	}

	return &deleted, nil
}

func UpdateCompleteReminder(ctx context.Context, id int64, isCompleted bool) (*models.Reminder, error) {

	var reminder models.Reminder

	err := database.DB.QueryRowContext(ctx, `UPDATE reminders
	SET is_completed = $2
	WHERE id = $1
	RETURNING id
	`, id, isCompleted).Scan(&reminder.ID)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &reminder, nil
}
