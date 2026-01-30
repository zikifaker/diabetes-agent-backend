package model

import "time"

type HealthReport struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt  time.Time `gorm:"not null" json:"updated_at"`
	UserEmail  string    `gorm:"not null;index:idx_email_start_at" json:"user_email"`
	StartAt    time.Time `gorm:"not null;index:idx_email_start_at" json:"start_at"`
	EndAt      time.Time `gorm:"not null" json:"end_at"`
	ReportType string    `gorm:"not null;type:enum('weekly','monthly','yearly')" json:"report_type"`
	Content    string    `gorm:"type:text" json:"content"`
}

func (HealthReport) TableName() string {
	return "health_report"
}
