package chat

import (
	"context"
	"diabetes-agent-backend/config"
	"diabetes-agent-backend/model"
	"diabetes-agent-backend/request"
	knowledgebase "diabetes-agent-backend/service/knowledge-base"
	"diabetes-agent-backend/utils"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	mcpadapter "github.com/i2y/langchaingo-mcp-adapter"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/tools"
)

const (
	BaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"

	methodToolCompleted = "tool_completed"
)

// Agent 对话服务的 HTTP 客户端，配置 300s 超时时间处理流式输出
var httpClient = utils.NewHTTPClient(
	utils.WithTimeout(300 * time.Second),
)

var (
	//go:embed prompts/conversational_format_instructions.txt
	conversationalFormatInstructions string

	//go:embed prompts/conversational_prefix.txt
	conversationalPrefix string

	//go:embed prompts/conversational_suffix.txt
	conversationalSuffix string
)

type Agent struct {
	// Agent 执行器
	Executor *agents.Executor

	// MCP 客户端
	MCPClient *client.Client

	// 聊天记录存储
	ChatHistory *MySQLChatMessageHistory

	// SSE 回调处理器，推送输出结果
	SSEHandler *GinSSEHandler
}

func NewAgent(req request.ChatRequest, c *gin.Context) (*Agent, error) {
	llm, err := openai.New(
		openai.WithModel(req.AgentConfig.Model),
		openai.WithToken(config.Cfg.Model.APIKey),
		openai.WithBaseURL(BaseURL),
		openai.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create llm client: %v", err)
	}

	mcpClient, err := createMCPClient(c)
	if err != nil {
		return nil, fmt.Errorf("failed to create mcp client: %v", err)
	}

	ctx := context.Background()
	if err := mcpClient.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to init connection to the mcp server: %v", err)
	}

	mcpTools, err := getMCPTools(mcpClient, req.AgentConfig.Tools)
	if err != nil {
		slog.Error("Failed to get mcp tools", "err", err)
	}

	sseHandler := NewGinSSEHandler(c, req.SessionID)
	registerMCPNotificationHandler(ctx, mcpClient, sseHandler)

	a := agents.NewConversationalAgent(llm, mcpTools,
		agents.WithCallbacksHandler(sseHandler),
		agents.WithPromptPrefix(conversationalPrefix),
		agents.WithPromptFormatInstructions(conversationalFormatInstructions),
		agents.WithPromptSuffix(conversationalSuffix),
	)

	chatHistory := NewMySQLChatMessageHistory(req.SessionID)
	memory := memory.NewConversationBuffer(
		memory.WithChatHistory(chatHistory),
	)

	executor := agents.NewExecutor(
		a,
		agents.WithMemory(memory),
		agents.WithMaxIterations(req.AgentConfig.MaxIterations),
	)

	return &Agent{
		Executor:    executor,
		MCPClient:   mcpClient,
		ChatHistory: chatHistory,
		SSEHandler:  sseHandler,
	}, nil
}

func (a *Agent) Call(ctx context.Context, req request.ChatRequest, c *gin.Context) error {
	// 引入上传文件和知识库检索结果
	if len(req.UploadedFiles) > 0 || req.EnableKnowledgeBaseRetrieval {
		req.Query = a.buildUserContext(ctx, req, c)
	}

	// 存储聊天信息的上下文，避免请求被取消时保存失败
	saveCtx := context.Background()

	// 若 chains.Run 成功执行，会自动存储问答对
	_, err := chains.Run(ctx, a.Executor, req.Query)
	if err != nil {
		switch {
		// 若抛出 ErrUnableToParseOutput，从错误信息中提取回答后推送，持久化问答对
		case errors.Is(err, agents.ErrUnableToParseOutput):
			slog.Warn("Failed to parse agent output, missing prefix 'AI:'")

			answer := strings.TrimPrefix(err.Error(), agents.ErrUnableToParseOutput.Error()+":")
			utils.SendSSEMessage(c, utils.EventFinalAnswer, answer)

			if err := a.saveConversation(saveCtx, req.Query, answer); err != nil {
				slog.Error("Failed to save agent final answer", "err", err)
			}

		// 若抛出 context.Canceled，持久化问答对
		case errors.Is(err, context.Canceled):
			slog.Warn("Client canceled")

			answer := a.SSEHandler.FinalAnswer.String()
			if err := a.saveConversation(saveCtx, req.Query, answer); err != nil {
				slog.Error("Failed to save agent final answer", "err", err)
			}

		default:
			return err
		}
	}

	if err := a.ChatHistory.UpdateFields(saveCtx, llms.ChatMessageTypeHuman, &model.Message{
		UploadedFiles: req.UploadedFiles,
	}); err != nil {
		slog.Error("Failed to update user message", "err", err)
	}

	if err := a.ChatHistory.UpdateFields(saveCtx, llms.ChatMessageTypeAI, &model.Message{
		IntermediateSteps: a.SSEHandler.IntermediateSteps.String(),
		ToolCallResults:   a.SSEHandler.ToolCallResults,
	}); err != nil {
		slog.Error("Failed to update agent message", "err", err)
	}

	return nil
}

// SaveConversation 存储问答对
func (a *Agent) saveConversation(ctx context.Context, query, answer string) error {
	err := a.ChatHistory.AddUserMessage(ctx, query)
	if err != nil {
		return err
	}

	err = a.ChatHistory.AddAIMessage(ctx, answer)
	if err != nil {
		return err
	}

	return nil
}

func (a *Agent) Close() error {
	if a.MCPClient != nil {
		return a.MCPClient.Close()
	}
	return nil
}

func createMCPClient(c *gin.Context) (*client.Client, error) {
	mcpServerPath := fmt.Sprintf("http://%s:%s/mcp", config.Cfg.MCP.Host, config.Cfg.MCP.Port)
	mcpClient, err := client.NewStreamableHttpClient(mcpServerPath,
		transport.WithHTTPBasicClient(utils.GlobalHTTPClient),
		transport.WithHTTPHeaders(map[string]string{
			"Authorization": c.GetHeader("Authorization"),
		}),
		transport.WithContinuousListening(),
	)
	if err != nil {
		return nil, err
	}
	return mcpClient, nil
}

// 返回用户选择的工具
func getMCPTools(mcpClient *client.Client, toolNames []string) ([]tools.Tool, error) {
	if len(toolNames) == 0 {
		return nil, nil
	}

	// 初始化与 MCP 服务端的连接
	mcpAdapter, err := mcpadapter.New(mcpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create mcp adapter: %v", err)
	}

	mcpTools, err := mcpAdapter.Tools()
	if err != nil {
		return nil, fmt.Errorf("failed to get mcp tools: %v", err)
	}

	toolMap := make(map[string]bool)
	for _, name := range toolNames {
		toolMap[name] = true
	}

	var filteredTools []tools.Tool
	for _, tool := range mcpTools {
		if toolMap[tool.Name()] {
			filteredTools = append(filteredTools, tool)
		}
	}

	return filteredTools, nil
}

// 注册通知处理方法，接收 MCP 服务端推送的工具调用结果
func registerMCPNotificationHandler(ctx context.Context, mcpClient *client.Client, sseHandler *GinSSEHandler) {
	mcpClient.OnNotification(func(notification mcp.JSONRPCNotification) {
		if notification.Method != methodToolCompleted {
			return
		}

		toolName, ok := notification.Params.AdditionalFields["tool"].(string)
		if !ok {
			slog.Error("Invalid tool name type")
			return
		}

		results, ok := notification.Params.AdditionalFields["result"].([]any)
		if !ok {
			slog.Error("Invalid tool call result type")
			return
		}

		textContent := []string{}
		for _, res := range results {
			if content, ok := res.(map[string]any); ok {
				switch contentType := content["type"].(string); contentType {
				case "text":
					textContent = append(textContent, content["text"].(string))
				}
			}
		}

		toolCallResult := model.ToolCallResult{
			Name:   toolName,
			Result: textContent,
		}

		sseHandler.HandleToolCallResult(ctx, toolCallResult)
	})
}

func (a *Agent) buildUserContext(ctx context.Context, req request.ChatRequest, c *gin.Context) string {
	email := c.GetString("email")

	var userContext strings.Builder
	userContext.WriteString("User Question:\n")
	userContext.WriteString(req.Query + "\n\n")
	userContext.WriteString("User Context:\n")

	if len(req.UploadedFiles) > 0 {
		utils.SendSSEMessage(c, utils.EventFileParseStart, nil)

		content := handleChatFiles(ctx, req, email)
		userContext.WriteString("Uploaded Files:\n")
		userContext.WriteString(content + "\n\n")

		utils.SendSSEMessage(c, utils.EventFileParseDone, nil)
	}

	if req.EnableKnowledgeBaseRetrieval {
		utils.SendSSEMessage(c, utils.EventKBRetrievalStart, nil)

		docs := knowledgebase.RetrieveSimilarDocuments(ctx, req.Query, email)
		docsJSON, _ := json.Marshal(docs)
		userContext.WriteString("Knowledge Base Retrieve Results:\n")
		userContext.WriteString(string(docsJSON) + "\n\n")

		utils.SendSSEMessage(c, utils.EventKBRetrievalChunkNum, len(docs))
		utils.SendSSEMessage(c, utils.EventKBRetrievalDone, nil)
	}

	return userContext.String()
}
