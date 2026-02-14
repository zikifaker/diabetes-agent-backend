package dao

import (
	"diabetes-agent-server/model"
	"diabetes-agent-server/response"
)

const pageSize = 10

func GetSystemMessages(email string, page int) (*response.GetSystemMessagesResponse, error) {
	var msgs response.GetSystemMessagesResponse
	err := DB.Model(&model.SystemMessage{}).
		Select("id, created_at, title, content, is_read").
		Where("user_email = ?", email).
		Order("created_at DESC").
		Count(&msgs.Total).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&msgs.Messages).Error
	return &msgs, err
}

func GetSystemMessageById(id string) (*model.SystemMessage, error) {
	var message model.SystemMessage
	err := DB.Where("id = ?", id).
		First(&message).Error
	return &message, err
}

func UpdateSystemMessageAsRead(id string) error {
	return DB.Model(&model.SystemMessage{}).
		Where("id = ?", id).
		Update("is_read", true).Error
}

func DeleteSystemMessage(id string) (*model.SystemMessage, error) {
	var message model.SystemMessage

	err := DB.Where("id = ? ", id).First(&message).Error
	if err != nil {
		return nil, err
	}

	err = DB.Where("id = ?", id).Delete(&model.SystemMessage{}).Error
	if err != nil {
		return nil, err
	}

	return &message, nil
}
