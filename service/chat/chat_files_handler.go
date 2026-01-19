package chat

import (
	"context"
	"diabetes-agent-backend/config"
	"diabetes-agent-backend/request"
	ossauth "diabetes-agent-backend/service/oss-auth"
	"diabetes-agent-backend/utils"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

const modelNameVLM = "qwen3-vl-flash"

var (
	imageExtensions = []string{".png", ".jpg", ".jpeg", ".gif", ".webp"}
	docExtensions   = []string{".doc", ".docx", ".pdf", ".xls", ".xlsx", ".txt", ".md"}
)

type DeleteUploadedFilesMessage struct {
	Email     string `json:"email"`
	SessionID string `json:"session_id"`
}

func handleChatFiles(ctx context.Context, req request.ChatRequest, email string) string {
	var images, docs []string
	for _, fileName := range req.UploadedFiles {
		ossAuthReq := request.OSSAuthRequest{
			Namespace: ossauth.OSSKeyPrefixUpload,
			Email:     email,
			SessionID: req.SessionID,
			FileName:  fileName,
		}

		switch {
		case supportImage(fileName):
			url, err := ossauth.GeneratePresignedURL(ossAuthReq)
			if err != nil {
				slog.Error("failed to generate presigned url",
					"file_name", fileName,
					"err", err,
				)
				continue
			}
			images = append(images, url)

		case supportDoc(fileName):
			objectName, err := ossauth.GenerateKey(ossAuthReq)
			if err != nil {
				slog.Error("failed to generate oss key",
					"file_name", fileName,
					"err", err,
				)
				continue
			}
			docs = append(docs, objectName)

		default:
			slog.Warn("unsupported file type", "file_name", fileName)
		}
	}

	var uploadedFilesContext strings.Builder

	if len(images) > 0 {
		content, err := handleImages(ctx, images)
		if err != nil {
			slog.Error("failed to handle uploaded images", "err", err)
		}
		uploadedFilesContext.WriteString("images:\n")
		uploadedFilesContext.WriteString(content + "\n\n")
	}

	if len(docs) > 0 {
		content, err := handleDocs(ctx, docs)
		if err != nil {
			slog.Error("failed to handle uploaded docs", "err", err)
		}
		uploadedFilesContext.WriteString("docs:\n")
		uploadedFilesContext.WriteString(content + "\n\n")
	}

	return uploadedFilesContext.String()
}

// 调用视觉理解模型生成图片的内容摘要
func handleImages(ctx context.Context, urls []string) (string, error) {
	vlm, err := openai.New(
		openai.WithModel(modelNameVLM),
		openai.WithToken(config.Cfg.Model.APIKey),
		openai.WithBaseURL(BaseURL),
		openai.WithHTTPClient(utils.GlobalHTTPClient),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create vlm client: %w", err)
	}

	parts := []llms.ContentPart{}
	for _, url := range urls {
		parts = append(parts, llms.ImageURLPart(url))
	}

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Summarize the content of the images briefly."),
			},
		},
	}

	result, err := vlm.GenerateContent(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("error generating content: %w", err)
	}

	return result.Choices[0].Content, nil
}

func handleDocs(ctx context.Context, objectNames []string) (string, error) {
	cfg := &oss.Config{
		Region: oss.Ptr(config.Cfg.OSS.Region),
		CredentialsProvider: credentials.NewStaticCredentialsProvider(
			config.Cfg.OSS.AccessKeyID,
			config.Cfg.OSS.AccessKeySecret,
		),
		HttpClient: utils.GlobalHTTPClient,
	}
	client := oss.NewClient(cfg)

	var content strings.Builder
	for _, objectName := range objectNames {
		request := &oss.GetObjectRequest{
			Bucket: oss.Ptr(config.Cfg.OSS.BucketName),
			Key:    oss.Ptr(objectName),
		}

		resp, err := client.GetObject(ctx, request)
		if err != nil {
			slog.Error("failed to get object",
				"object_name", objectName,
				"err", err,
			)
			continue
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("failed to read object",
				"object_name", objectName,
				"err", err,
			)
		}
		content.WriteString(objectName + ":\n")
		content.WriteString(string(data))
		content.WriteString("\n\n")

		resp.Body.Close()
	}

	return content.String(), nil
}

func supportImage(fileName string) bool {
	for _, ext := range imageExtensions {
		if strings.HasSuffix(fileName, ext) {
			return true
		}
	}
	return false
}

func supportDoc(fileName string) bool {
	for _, ext := range docExtensions {
		if strings.HasSuffix(fileName, ext) {
			return true
		}
	}
	return false
}

func HandleDeleteUploadedFilesMessage(ctx context.Context, msg *primitive.MessageExt) error {
	var message DeleteUploadedFilesMessage
	if err := json.Unmarshal(msg.Body, &message); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	cfg := &oss.Config{
		Region: oss.Ptr(config.Cfg.OSS.Region),
		CredentialsProvider: credentials.NewStaticCredentialsProvider(
			config.Cfg.OSS.AccessKeyID,
			config.Cfg.OSS.AccessKeySecret,
		),
		HttpClient: utils.GlobalHTTPClient,
	}
	client := oss.NewClient(cfg)

	// 构造对象前缀，过滤出当前会话的文件
	prefix := strings.Join([]string{ossauth.OSSKeyPrefixUpload, message.Email, message.SessionID}, "/")
	result, err := client.ListObjectsV2(ctx, &oss.ListObjectsV2Request{
		Bucket: oss.Ptr(config.Cfg.OSS.BucketName),
		Prefix: oss.Ptr(prefix),
	})
	if err != nil {
		return fmt.Errorf("failed to list objects: %w", err)
	}

	if result.KeyCount == 0 {
		slog.Warn("no objects found")
		return nil
	}

	deleteObjects := make([]oss.DeleteObject, 0, result.KeyCount)
	for i := 0; i < result.KeyCount; i++ {
		deleteObjects = append(deleteObjects, oss.DeleteObject{
			Key: result.Contents[i].Key,
		})
	}

	_, err = client.DeleteMultipleObjects(ctx, &oss.DeleteMultipleObjectsRequest{
		Bucket:  oss.Ptr(config.Cfg.OSS.BucketName),
		Objects: deleteObjects,
	})
	if err != nil {
		return fmt.Errorf("failed to delete objects: %w", err)
	}

	slog.Info("deleted uploaded files",
		"object_prefix", prefix,
		"count", result.KeyCount,
	)

	return nil
}
