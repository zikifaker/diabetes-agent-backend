package response

import (
	"diabetes-agent-backend/model"
	"time"
)

type SessionResponse struct {
	SessionID string `json:"session_id"`
	Title     string `json:"title"`
}

type MessageResponse struct {
	CreatedAt         time.Time              `json:"created_at"`
	Role              string                 `json:"role"`
	Content           string                 `json:"content"`
	IntermediateSteps string                 `json:"intermediate_steps"`
	ToolCallResults   []model.ToolCallResult `gorm:"type:json;serializer:json" json:"tool_call_results"`
	UploadedFiles     []string               `gorm:"type:json;serializer:json" json:"uploaded_files"`
}
