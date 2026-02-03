package model

import "time"

type HealthWeeklyReport struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt  time.Time `gorm:"not null" json:"updated_at"`
	UserEmail  string    `gorm:"not null;index:idx_email_start_at" json:"user_email"`
	StartAt    time.Time `gorm:"not null;index:idx_email_start_at" json:"start_at"`
	EndAt      time.Time `gorm:"not null" json:"end_at"`
	ObjectName string    `gorm:"not null" json:"object_name"`
}

func (HealthWeeklyReport) TableName() string {
	return "health_weekly_report"
}
