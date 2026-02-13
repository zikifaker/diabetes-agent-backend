package model

import "time"

type SystemMessage struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `gorm:"index:idx_email_created" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserEmail string    `gorm:"not null;index:idx_email_created" json:"user_email"`
	Title     string    `gorm:"not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	IsRead    bool      `gorm:"not null" json:"is_read"`
}

func (SystemMessage) TableName() string {
	return "system_message"
}
