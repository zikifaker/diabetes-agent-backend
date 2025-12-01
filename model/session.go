package model

import (
	"time"
)

const DefaultSessionTitle = "新会话"

type Session struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserEmail string    `gorm:"not null;index" json:"user_email"`
	SessionID string    `gorm:"not null" json:"session_id"`
	Title     string    `json:"title"`
}

func (Session) TableName() string {
	return "chat_session"
}

// Message 建立联合索引 (session_id, created_at)
type Message struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time `gorm:"index:idx_session_created" json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	SessionID      string    `gorm:"not null;index:idx_session_created" json:"session_id"`
	Role           string    `gorm:"not null" json:"role"`
	Content        string    `gorm:"type:text" json:"content"`
	ImmediateSteps string    `gorm:"type:text" json:"immediate_steps"`
	Summary        string    `gorm:"type:text" json:"summary"`
}

func (Message) TableName() string {
	return "chat_message"
}
