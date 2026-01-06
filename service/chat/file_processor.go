package chat

import (
	"context"
	"diabetes-agent-backend/config"
	"diabetes-agent-backend/request"
	ossauth "diabetes-agent-backend/service/oss-auth"
	"diabetes-agent-backend/utils"
	"fmt"
	"log/slog"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

const modelNameVLM = "qwen3-vl-flash"

var (
	httpClient = utils.DefaultHTTPClient()

	imageExtensions = []string{".png", ".jpg", ".jpeg", ".gif", ".webp"}
	docExtensions   = []string{".pdf", ".txt"}
)

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

func handleChatFiles(ctx context.Context, req request.ChatRequest, email string) string {
	var images, docs []string
	for _, fileName := range req.UploadedFiles {
		url, err := ossauth.GeneratePresignedURL(request.OSSAuthRequest{
			Namespace: ossauth.OSSKeyPrefixUpload,
			Email:     email,
			SessionID: req.SessionID,
			FileName:  fileName,
		})
		if err != nil {
			slog.Error("failed to generate presigned url", "err", err)
			continue
		}

		switch {
		case supportImage(fileName):
			images = append(images, url)
		case supportDoc(fileName):
			docs = append(docs, url)
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

// 从 OSS 下载图片并生成图片摘要
func handleImages(ctx context.Context, urls []string) (string, error) {
	vlm, err := openai.New(
		openai.WithModel(modelNameVLM),
		openai.WithToken(config.Cfg.Model.APIKey),
		openai.WithBaseURL(BaseURL),
		openai.WithHTTPClient(httpClient),
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

func handleDocs(ctx context.Context, urls []string) (string, error) {
	return "", nil
}
