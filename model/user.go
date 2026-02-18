package model

import (
	"time"
)

type User struct {
	ID                             uint      `gorm:"primarykey" json:"id"`
	CreatedAt                      time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt                      time.Time `gorm:"not null" json:"updated_at"`
	Email                          string    `gorm:"uniqueIndex;not null" json:"email"`
	Password                       string    `gorm:"not null" json:"-"`
	Avatar                         string    `gorm:"not null" json:"avatar"`
	EnableWeeklyReportNotification bool      `gorm:"not null" json:"enable_weekly_report_notification"`
}

func (User) TableName() string {
	return "user"
}
