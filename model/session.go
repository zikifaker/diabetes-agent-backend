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

// Message 聊天消息
// 建立联合索引 (session_id, created_at)
type Message struct {
	ID                uint               `gorm:"primarykey" json:"id"`
	CreatedAt         time.Time          `gorm:"index:idx_session_created" json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
	SessionID         string             `gorm:"not null;index:idx_session_created" json:"session_id"`
	Role              string             `gorm:"not null" json:"role"`
	Content           string             `gorm:"type:text" json:"content"`
	Summary           string             `gorm:"type:text" json:"summary"`
	IntermediateSteps InterMediateSteps  `gorm:"foreignKey:MessageID"`
	ToolCallResults   ToolCallResults    `gorm:"foreignKey:MessageID"`
	Files             []ChatUploadedFile `gorm:"foreignKey:MessageID"`
}

func (Message) TableName() string {
	return "chat_message"
}

// InterMediateSteps Agent 的思考步骤
type InterMediateSteps struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	MessageID uint      `gorm:"not null;index:idx_message" json:"message_id"`
	SessionID string    `gorm:"not null" json:"session_id"`
	Content   string    `gorm:"type:text" json:"content"`
}

func (InterMediateSteps) TableName() string {
	return "chat_intermediate_steps"
}

// ToolCallResults Agent 的工具调用结果
type ToolCallResults struct {
	ID        uint             `gorm:"primarykey" json:"id"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	MessageID uint             `gorm:"not null;index:idx_message" json:"message_id"`
	SessionID string           `gorm:"not null" json:"session_id"`
	Content   []ToolCallResult `gorm:"type:json;serializer:json" json:"content"`
}

type ToolCallResult struct {
	Name   string   `json:"name"`
	Result []string `json:"result"`
}

func (ToolCallResults) TableName() string {
	return "chat_tool_call_results"
}

// ChatUploadedFiles 聊天文件元数据
type ChatUploadedFile struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	MessageID uint      `gorm:"not null;index:idx_message" json:"message_id"`
	SessionID string    `gorm:"not null" json:"session_id"`
	FileName  string    `json:"file_name"`
}

func (ChatUploadedFile) TableName() string {
	return "chat_uploaded_file"
}
