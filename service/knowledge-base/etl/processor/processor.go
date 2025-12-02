package processor

import (
	"context"
	"strings"

	"github.com/milvus-io/milvus/client/v2/column"
)

// ETLProcessor 知识文件ETL处理器
type ETLProcessor interface {
	// 判断是否支持传入的文件类型
	CanProcess(fileType string) bool

	// 执行ETL流程
	ExecuteETLPipeline(ctx context.Context, object []byte, objectName string) error
}

type Metadata struct {
	objectName string
}

// 增加milvus元数据列
func addMetadataColumns(columns []column.Column, recordNum int, metadata *Metadata) []column.Column {
	pathSegments := strings.Split(metadata.objectName, "/")
	userEmail := pathSegments[0]
	title := pathSegments[len(pathSegments)-1]

	titles := make([]string, recordNum)
	userEmails := make([]string, recordNum)
	for i := range recordNum {
		titles[i] = title
		userEmails[i] = userEmail
	}

	columns = append(columns, column.NewColumnVarChar("title", titles))
	columns = append(columns, column.NewColumnVarChar("user_email", userEmails))

	return columns
}
