package etl

import (
	"context"
	"diabetes-agent-backend/config"
	"diabetes-agent-backend/model"
	"diabetes-agent-backend/service/knowledge-base/etl/processor"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

var (
	// 知识文件ETL处理器注册表
	etlProcessorRegistry []processor.ETLProcessor

	// 全局HTTP客户端，访问OSS时复用
	httpClient *http.Client
)

type ETLMessage struct {
	FileType   model.FileType `json:"file_type"`
	ObjectName string         `json:"object_name"`
}

type DeleteMessage struct {
	FileType   model.FileType `json:"file_type"`
	ObjectName string         `json:"object_name"`
}

func init() {
	pdfProcessor, err := processor.NewPDFETLProcessor()
	if err != nil {
		panic(fmt.Sprintf("error creating PDFETLProcessor: %v", err))
	}

	markdownProcessor, err := processor.NewMarkdownETLProcessor()
	if err != nil {
		panic(fmt.Sprintf("error creating MarkdownETLProcessor: %v", err))
	}

	etlProcessorRegistry = []processor.ETLProcessor{
		pdfProcessor,
		markdownProcessor,
	}

	httpClient = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

func HandleETLMessage(ctx context.Context, msg *primitive.MessageExt) error {
	var etlMessage ETLMessage
	if err := json.Unmarshal(msg.Body, &etlMessage); err != nil {
		return fmt.Errorf("failed to unmarshal message body: %v", err)
	}

	object, err := getObjectFromOSS(ctx, &etlMessage)
	if err != nil {
		return fmt.Errorf("failed to get object from oss: %v", err)
	}

	slog.Debug("get object from oss successfully", "object_name", etlMessage.ObjectName)

	// 查找匹配文件类型的处理器，执行ETL流程
	foundProcessor := false
	for _, processor := range etlProcessorRegistry {
		if processor.CanProcess(etlMessage.FileType) {
			foundProcessor = true
			if err := processor.ExecuteETLPipeline(ctx, object, etlMessage.ObjectName); err != nil {
				return fmt.Errorf("failed to execute ETL pipeline: %v", err)
			}
			slog.Info("ETL pipeline executed successfully", "msg_id", msg.MsgId)
			return nil
		}
	}

	if !foundProcessor {
		return fmt.Errorf("no processor found for file type: %s", etlMessage.FileType)
	}

	return nil
}

func HandleDeleteMessage(ctx context.Context, msg *primitive.MessageExt) error {
	var deleteMessage DeleteMessage
	if err := json.Unmarshal(msg.Body, &deleteMessage); err != nil {
		return fmt.Errorf("failed to unmarshal message body: %v", err)
	}

	if err := deleteObjectFromOSS(ctx, &deleteMessage); err != nil {
		return fmt.Errorf("failed to delete object from oss: %v", err)
	}

	foundProcessor := false
	for _, processor := range etlProcessorRegistry {
		if processor.CanProcess(deleteMessage.FileType) {
			foundProcessor = true
			if err := processor.DeleteVectorStore(ctx, deleteMessage.ObjectName); err != nil {
				return fmt.Errorf("failed to delete vector store: %v", err)
			}
			slog.Info("vector store deleted successfully", "msg_id", msg.MsgId)
			return nil
		}
	}

	if !foundProcessor {
		return fmt.Errorf("no processor found for file type: %s", deleteMessage.FileType)
	}

	return nil
}

func getObjectFromOSS(ctx context.Context, etlMessage *ETLMessage) ([]byte, error) {
	cfg := &oss.Config{
		Region: oss.Ptr(config.Cfg.OSS.Region),
		CredentialsProvider: credentials.NewStaticCredentialsProvider(
			config.Cfg.OSS.AccessKeyID,
			config.Cfg.OSS.AccessKeySecret,
		),
		HttpClient: httpClient,
	}
	client := oss.NewClient(cfg)

	result, err := client.GetObject(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(config.Cfg.OSS.BucketName),
		Key:    oss.Ptr(etlMessage.ObjectName),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %v", err)
	}

	return data, nil
}

func deleteObjectFromOSS(ctx context.Context, deleteMessage *DeleteMessage) error {
	cfg := &oss.Config{
		Region: oss.Ptr(config.Cfg.OSS.Region),
		CredentialsProvider: credentials.NewStaticCredentialsProvider(
			config.Cfg.OSS.AccessKeyID,
			config.Cfg.OSS.AccessKeySecret,
		),
		HttpClient: httpClient,
	}
	client := oss.NewClient(cfg)

	_, err := client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(config.Cfg.OSS.BucketName),
		Key:    oss.Ptr(deleteMessage.ObjectName),
	})
	if err != nil {
		return err
	}

	return nil
}
