package summarization

import (
	"context"
	"diabetes-agent-server/config"
	"diabetes-agent-server/dao"
	"diabetes-agent-server/model"
	"diabetes-agent-server/service/chat"
	"diabetes-agent-server/utils"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	"gorm.io/gorm"
)

const (
	modelName       = "deepseek-v3"
	updateBatchSize = 1

	// 生成消息摘要的最小消息长度
	summaryThreshold = 2500
)

//go:embed prompts/summarization.txt
var summaryPrompt string

var (
	updates = make([]*model.Message, 0, updateBatchSize)
	mu      sync.Mutex
)

type Message struct {
	MsgIDs []uint `json:"msg_ids"`
}

func HandleSummarizationMessage(ctx context.Context, msg *primitive.MessageExt) error {
	var summarizationMessage Message
	if err := json.Unmarshal(msg.Body, &summarizationMessage); err != nil {
		return fmt.Errorf("failed to unmarshal message body: %v", err)
	}

	for _, msgID := range summarizationMessage.MsgIDs {
		msg, err := dao.GetMessageByID(msgID)
		if err != nil {
			slog.Error("Failed to get message",
				"msg_id", msgID,
				"err", err,
			)
			continue
		}

		if len(msg.Content) < summaryThreshold {
			continue
		}

		summary, err := generateSummary(ctx, msg.Role, msg.Content)
		if err != nil {
			slog.Error("Failed to summarize message",
				"msg_id", msgID,
				"err", err,
			)
			continue
		}

		msg.Summary = summary

		mu.Lock()
		updates = append(updates, msg)
		canFlush := len(updates) >= updateBatchSize
		mu.Unlock()

		if canFlush {
			mu.Lock()
			toFlush := make([]*model.Message, len(updates))
			copy(toFlush, updates)
			updates = updates[:0]
			mu.Unlock()

			// 批量更新消息摘要
			if err := flushBatchUpdates(toFlush); err != nil {
				slog.Error("Failed to flush batch updates", "err", err)
			}
		}
	}

	return nil
}

func generateSummary(ctx context.Context, role, content string) (string, error) {
	template := prompts.NewPromptTemplate(summaryPrompt, []string{"role", "content"})
	prompt, err := template.Format(map[string]any{
		"role":    role,
		"content": content,
	})
	if err != nil {
		return "", fmt.Errorf("failed to format prompt: %v", err)
	}

	llm, err := openai.New(
		openai.WithModel(modelName),
		openai.WithToken(config.Cfg.Model.APIKey),
		openai.WithBaseURL(chat.BaseURL),
		openai.WithHTTPClient(utils.GlobalHTTPClient),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create llm client: %v", err)
	}

	res, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		return "", fmt.Errorf("error calling llm: %w", err)
	}

	return res, nil
}

func flushBatchUpdates(updates []*model.Message) error {
	if len(updates) == 0 {
		return nil
	}

	err := dao.DB.Transaction(func(tx *gorm.DB) error {
		for _, msg := range updates {
			if err := tx.Model(&model.Message{}).
				Where("id = ?", msg.ID).
				Update("summary", msg.Summary).Error; err != nil {
				return fmt.Errorf("failed to update message %d: %v", msg.ID, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to update messages batch: %v", err)
	}

	return nil
}
