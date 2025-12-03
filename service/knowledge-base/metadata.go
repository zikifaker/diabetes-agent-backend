package knowledgebase

import (
	"diabetes-agent-backend/dao"
	"diabetes-agent-backend/model"
	"diabetes-agent-backend/request"
	"fmt"
	"log/slog"
	"strings"
)

func UploadKnowledgeMetadata(req request.UploadKnowledgeMetadataRequest, email string) error {
	// 检查文件是否已经上传过
	exists, err := dao.GetKnowledgeMetadataByEmailAndFileName(email, req.FileName)
	if err != nil {
		return fmt.Errorf("failed to get knowledge metadata: %v", err)
	}
	if exists != nil {
		return fmt.Errorf("file already exists")
	}

	err = dao.SaveKnowledgeMetadata(req, email)
	if err != nil {
		return fmt.Errorf("failed to save knowledge metadata: %v", err)
	}

	return nil
}

func UpdateKnowledgeMetadataStatus(objectName string, status model.Status) error {
	pathSegments := strings.Split(objectName, "/")
	if len(pathSegments) < 2 {
		return fmt.Errorf("invalid object name: %s", objectName)
	}

	userEmail := pathSegments[0]
	fileName := pathSegments[len(pathSegments)-1]

	err := dao.UpdateKnowledgeMetadataStatus(userEmail, fileName, status)
	if err != nil {
		slog.Error("failed to update knowledge metadata", "err", err)
		return err
	}

	return nil
}
