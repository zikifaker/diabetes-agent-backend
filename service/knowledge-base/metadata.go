package knowledgebase

import (
	"diabetes-agent-backend/dao"
	"diabetes-agent-backend/model"
	"fmt"
	"log/slog"
	"strings"
)

func UploadKnowledgeMetadata(metadata model.KnowledgeMetadata) error {
	// 检查文件是否已经上传过
	exists, err := dao.GetKnowledgeMetadataByEmailAndFileName(metadata.UserEmail, metadata.FileName)
	if err != nil {
		return fmt.Errorf("failed to get knowledge metadata: %v", err)
	}
	if exists != nil {
		return fmt.Errorf("file already exists")
	}

	if err := dao.DB.Create(&metadata).Error; err != nil {
		return fmt.Errorf("failed to save knowledge metadata: %v", err)
	}

	return nil
}

func UpdateKnowledgeMetadataStatus(objectName string, status model.Status) error {
	userEmail, fileName, err := ParseObjectName(objectName)
	if err != nil {
		return err
	}

	err = dao.UpdateKnowledgeMetadataStatus(userEmail, fileName, status)
	if err != nil {
		slog.Error("failed to update knowledge metadata", "err", err)
		return err
	}

	return nil
}

// ParseObjectName 解析 objectName，提取 email 和 fileName
func ParseObjectName(objectName string) (string, string, error) {
	pathSegments := strings.Split(objectName, "/")
	if len(pathSegments) < 3 {
		return "", "", fmt.Errorf("invalid object name: %s", objectName)
	}
	return pathSegments[1], pathSegments[len(pathSegments)-1], nil
}
