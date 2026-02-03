package processor

import (
	"context"
	"diabetes-agent-server/config"
	"diabetes-agent-server/constants"
	"diabetes-agent-server/model"
	knowledgebase "diabetes-agent-server/service/knowledge-base"
	"diabetes-agent-server/utils"
	"fmt"

	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/textsplitter"
)

const (
	embeddingModelName = "text-embedding-v4"
	chunkSize          = 4000
	chunkOverlap       = 400
	embeddingBatchSize = 10
	vectorDim          = 1024

	CollectionName = "knowledge_doc"
)

// ETLProcessor 知识文件 ETL 处理器
type ETLProcessor interface {
	// 判断是否支持传入的文件类型
	CanProcess(fileType model.FileType) bool

	// 执行 ETL 流程
	ExecuteETLPipeline(ctx context.Context, object []byte, objectName string) error

	// 删除向量存储
	DeleteVectorStore(ctx context.Context, objectName string) error
}

// BaseETLProcessor 基础 ETL处理器，提供删除向量存储的默认实现
type BaseETLProcessor struct {
	TextSplitter textsplitter.TextSplitter
	Embedder     embeddings.Embedder
	MilvusClient *milvusclient.Client
}

var _ ETLProcessor = &BaseETLProcessor{}

func NewBaseETLProcessor(textSplitter textsplitter.TextSplitter) (*BaseETLProcessor, error) {
	client, err := openai.New(
		openai.WithEmbeddingModel(embeddingModelName),
		openai.WithToken(config.Cfg.Model.APIKey),
		openai.WithBaseURL(constants.BaseURL),
		openai.WithHTTPClient(utils.GlobalHTTPClient),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder client: %v", err)
	}

	embedder, err := embeddings.NewEmbedder(client,
		embeddings.WithBatchSize(embeddingBatchSize),
		embeddings.WithStripNewLines(false),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %v", err)
	}

	milvusConfig := milvusclient.ClientConfig{
		Address: config.Cfg.Milvus.Endpoint,
		APIKey:  config.Cfg.Milvus.APIKey,
	}

	milvusClient, err := milvusclient.New(context.Background(), &milvusConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create milvus client: %v", err)
	}
	return &BaseETLProcessor{
		TextSplitter: textSplitter,
		Embedder:     embedder,
		MilvusClient: milvusClient,
	}, nil
}

func (p *BaseETLProcessor) CanProcess(fileType model.FileType) bool {
	return false
}

func (p *BaseETLProcessor) ExecuteETLPipeline(ctx context.Context, object []byte, objectName string) error {
	return nil
}

func (p *BaseETLProcessor) DeleteVectorStore(ctx context.Context, objectName string) error {
	userEmail, fileName, err := knowledgebase.ParseObjectName(objectName)
	if err != nil {
		return fmt.Errorf("error parsing object name: %v", err)
	}

	expression := fmt.Sprintf("user_email == '%s' and title == '%s'", userEmail, fileName)
	deleteOption := milvusclient.NewDeleteOption(CollectionName).WithExpr(expression)

	_, err = p.MilvusClient.Delete(ctx, deleteOption)
	if err != nil {
		return fmt.Errorf("error deleting document chunks: %v", err)
	}

	return nil
}

type Metadata struct {
	objectName string
}

// 增加 milvus 元数据列
func addMetadataColumns(columns []column.Column, recordNum int, metadata *Metadata) ([]column.Column, error) {
	userEmail, title, err := knowledgebase.ParseObjectName(metadata.objectName)
	if err != nil {
		return nil, err
	}

	titles := make([]string, recordNum)
	userEmails := make([]string, recordNum)
	for i := range recordNum {
		titles[i] = title
		userEmails[i] = userEmail
	}

	columns = append(columns, column.NewColumnVarChar("title", titles))
	columns = append(columns, column.NewColumnVarChar("user_email", userEmails))

	return columns, nil
}
