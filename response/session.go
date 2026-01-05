package response

import (
	"diabetes-agent-backend/model"
	"time"
)

type SessionResponse struct {
	SessionID string `json:"session_id"`
	Title     string `json:"title"`
}

type GetSessionsResponse struct {
	Sessions []SessionResponse `json:"sessions"`
}

type MessageResponse struct {
	CreatedAt       time.Time              `json:"created_at"`
	Role            string                 `json:"role"`
	Content         string                 `json:"content"`
	ImmediateSteps  string                 `json:"immediate_steps"`
	ToolCallResults []model.ToolCallResult `json:"tool_call_results"`
	UploadedFiles   []string               `json:"uploaded_files"`
}

type GetSessionMessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
}
