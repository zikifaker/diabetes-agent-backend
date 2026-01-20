package dao

import (
	"diabetes-agent-backend/model"
	"diabetes-agent-backend/response"
	"errors"

	"gorm.io/gorm"
)

func GetKnowledgeMetadataByEmail(email string) ([]response.MetadataResponse, error) {
	var fileMetadata []response.MetadataResponse
	err := DB.Table("knowledge_metadata").
		Select("file_name, file_type, file_size").
		Where("user_email = ?", email).
		Order("created_at DESC").
		Find(&fileMetadata).Error
	return fileMetadata, err
}

func GetKnowledgeMetadataByEmailAndFileName(email, fileName string) (*model.KnowledgeMetadata, error) {
	var fileMetadata model.KnowledgeMetadata
	if err := DB.Where("user_email = ? AND file_name = ?", email, fileName).
		First(&fileMetadata).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &fileMetadata, nil
}

func DeleteKnowledgeMetadataByEmailAndFileName(email, fileName string) error {
	return DB.Where("user_email = ? AND file_name = ?", email, fileName).
		Delete(&model.KnowledgeMetadata{}).Error
}

func UpdateKnowledgeMetadataStatus(email, fileName string, status model.Status) error {
	return DB.Model(&model.KnowledgeMetadata{}).
		Where("user_email = ? AND file_name = ?", email, fileName).
		Update("status", status).Error
}

func SearchKnowledgeMetadataByFullText(email, query string) ([]response.MetadataResponse, error) {
	var fileMetadata []response.MetadataResponse

	// 使用全文索引做左右模糊匹配
	err := DB.Table("knowledge_metadata").
		Select("file_name, file_type, file_size").
		Where("user_email = ? AND MATCH(file_name) AGAINST(? IN BOOLEAN MODE)", email, "*"+query+"*").
		Order("created_at DESC").
		Find(&fileMetadata).Error

	return fileMetadata, err
}
