package model

import (
	"time"
)

const DefaultSessionTitle = "新会话"

type Session struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserEmail string    `gorm:"not null;index:idx_email" json:"user_email"`
	SessionID string    `gorm:"not null" json:"session_id"`
	Title     string    `json:"title"`
}

func (Session) TableName() string {
	return "chat_session"
}

// Message 存储聊天记录
// 建立联合索引 (session_id, created_at)
// ToolCallResults 和 UploadedFiles 必须在 gorm 标签中指定 json 序列化器
type Message struct {
	ID                uint             `gorm:"primarykey" json:"id"`
	CreatedAt         time.Time        `gorm:"index:idx_session_created" json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
	SessionID         string           `gorm:"not null;index:idx_session_created" json:"session_id"`
	Role              string           `gorm:"not null" json:"role"`
	Content           string           `gorm:"type:text" json:"content"`
	Summary           string           `gorm:"type:text" json:"summary"`
	IntermediateSteps string           `gorm:"type:text" json:"intermediate_steps"`
	ToolCallResults   []ToolCallResult `gorm:"type:json;serializer:json" json:"tool_call_results"`
	UploadedFiles     []string         `gorm:"type:json;serializer:json" json:"uploaded_files"`
}

type ToolCallResult struct {
	Name string `json:"name"`

	// 每次工具调用返回一组结果
	Result []string `json:"result"`
}

func (Message) TableName() string {
	return "chat_message"
}
