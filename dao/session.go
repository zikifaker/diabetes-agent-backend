package dao

import (
	"diabetes-agent-server/model"
	"diabetes-agent-server/response"

	"github.com/tmc/langchaingo/llms"
	"gorm.io/gorm"
)

func GetSessionsByEmail(email string) ([]response.SessionResponse, error) {
	var sessions []response.SessionResponse
	err := DB.Model(&model.Session{}).
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

	// 删除关联的记录
	err = DB.Where("session_id = ?", sessionID).
		Delete(&[]model.Message{}).Error
	if err != nil {
		return err
	}

	err = DB.Where("session_id = ?", sessionID).
		Delete(&[]model.InterMediateSteps{}).Error
	if err != nil {
		return err
	}

	err = DB.Where("session_id = ?", sessionID).
		Delete(&[]model.ToolCallResults{}).Error
	if err != nil {
		return err
	}

	err = DB.Where("session_id = ?", sessionID).
		Delete(&[]model.ChatUploadedFile{}).Error
	if err != nil {
		return err
	}

	return nil
}

func GetMessagesBySessionID(sessionID string) ([]response.MessageResponse, error) {
	var messages []model.Message

	// 预加载与 Message 表关联的记录
	err := DB.Model(&model.Message{}).
		Preload("IntermediateSteps", func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN chat_message ON chat_message.id = chat_intermediate_steps.message_id").
				Where("chat_message.role = ?", llms.ChatMessageTypeAI)
		}).
		Preload("ToolCallResults", func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN chat_message ON chat_message.id = chat_tool_call_results.message_id").
				Where("chat_message.role = ?", llms.ChatMessageTypeAI)
		}).
		Preload("Files", func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN chat_message ON chat_message.id = chat_uploaded_file.message_id").
				Where("chat_message.role = ?", llms.ChatMessageTypeHuman)
		}).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	messageResponse := make([]response.MessageResponse, 0, len(messages))
	for _, msg := range messages {
		fileNames := make([]string, 0)
		for _, f := range msg.Files {
			fileNames = append(fileNames, f.FileName)
		}

		resp := response.MessageResponse{
			CreatedAt:         msg.CreatedAt,
			Role:              msg.Role,
			Content:           msg.Content,
			IntermediateSteps: msg.IntermediateSteps.Content,
			ToolCallResults:   msg.ToolCallResults.Content,
			UploadedFiles:     fileNames,
		}
		messageResponse = append(messageResponse, resp)
	}

	return messageResponse, nil
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

func SaveChatUploadedFiles(fileNames []string, messageID uint, sessionID string) error {
	for _, fileName := range fileNames {
		if err := DB.Create(&model.ChatUploadedFile{
			MessageID: messageID,
			SessionID: sessionID,
			FileName:  fileName,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
