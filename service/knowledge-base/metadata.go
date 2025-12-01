package knowledgebase

import (
	"diabetes-agent-backend/dao"
	"diabetes-agent-backend/request"
	"fmt"
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
