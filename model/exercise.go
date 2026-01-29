package model

import (
	"time"
)

type ExerciseRecord struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null" json:"updated_at"`
	UserEmail   string    `gorm:"not null;index:idx_email_start_at" json:"user_email"`
	Type        string    `gorm:"not null;type:enum('aerobic', 'strength', 'flexibility', 'other')" json:"type"`
	Name        string    `json:"name"`
	Intensity   string    `gorm:"not null;type:enum('low', 'medium', 'high')" json:"intensity"`
	StartAt     time.Time `gorm:"index:idx_email_start_at" json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	Duration    int       `json:"duration"`
	PreGlucose  float32   `json:"pre_glucose"`
	PostGlucose float32   `json:"post_glucose"`
	Notes       string    `json:"notes"`
}

func (ExerciseRecord) TableName() string {
	return "exercise_record"
}
