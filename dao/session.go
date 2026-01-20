package dao

import (
	"diabetes-agent-backend/model"
	"diabetes-agent-backend/response"
)

func GetSessionsByEmail(email string) ([]response.SessionResponse, error) {
	var sessions []response.SessionResponse
	err := DB.Table("chat_session").
		Select("session_id, title").
		Where("user_email = ?", email).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

func DeleteSession(email, sessionID string) error {
	// 删除会话
	err := DB.Where("user_email = ? AND session_id = ?", email, sessionID).
		Delete(&model.Session{}).Error
	if err != nil {
		return err
	}

	// 删除会话内的对话记录
	err = DB.Where("session_id = ?", sessionID).
		Delete(&[]model.Message{}).Error
	if err != nil {
		return err
	}

	return nil
}

func GetMessagesBySessionID(sessionID string) ([]response.MessageResponse, error) {
	var messages []response.MessageResponse
	err := DB.Table("chat_message").
		Select("created_at, role, content, intermediate_steps, tool_call_results, uploaded_files").
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

func GetMessageByID(messageID uint) (*model.Message, error) {
	var message model.Message
	err := DB.Where("id = ?", messageID).
		First(&message).Error
	return &message, err
}

func UpdateSessionTitle(email, sessionID, title string) error {
	return DB.Model(&model.Session{}).
		Where("user_email = ? AND session_id = ?", email, sessionID).
		Update("title", title).Error
}
