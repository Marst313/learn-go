package models

import "time"

type Reminder struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"not null;index"`
	User        User      `json:"-" gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description"`
	Date        time.Time `json:"date" gorm:"type:date;not null"`
	Time        time.Time `json:"time" gorm:"type:time;not null"`
	Category    string    `json:"category" gorm:"type:varchar(50);not null;default:'personal'"`
	Priority    string    `json:"priority" gorm:"type:varchar(20);not null;default:'medium'"`
	Recurring   bool      `json:"recurring" gorm:" null;default:false"`
	IsCompleted bool      `json:"is_completed" gorm:"default:false;not null;index"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type ReminderResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	IsCompleted bool      `json:"is_completed"`
	Category    string    `json:"category"`
	Priority    string    `json:"priority"`
	Time        time.Time `json:"time"`
	Recurring   bool      `json:"recurring"`
	Date        time.Time `json:"date"`
}

type ReminderToggleComplete struct {
	IsCompleted bool `json:"is_completed" gorm:"not null"`
}
