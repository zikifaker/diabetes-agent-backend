package chat

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/callbacks"
)

// GinSSEHandler 基于Gin的回调处理器，使用SSE发送Agent的输出内容
type GinSSEHandler struct {
	callbacks.SimpleHandler
	Ctx         *gin.Context
	ChatHistory *MySQLChatMessageHistory
	Session     string

	// Agent输出中是否包含final answer
	hasFinalAnswer bool

	// 存储Agent的思考步骤
	immediateStepsBuilder *strings.Builder
}

var _ callbacks.Handler = &GinSSEHandler{}

func NewGinSSEHandler(ctx *gin.Context, chatHistory *MySQLChatMessageHistory, session string) *GinSSEHandler {
	return &GinSSEHandler{
		Ctx:                   ctx,
		ChatHistory:           chatHistory,
		Session:               session,
		hasFinalAnswer:        false,
		immediateStepsBuilder: &strings.Builder{},
	}
}

func (h *GinSSEHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	text := string(chunk)

	if h.hasFinalAnswer {
		h.Ctx.SSEvent("final_answer", text)
	} else if idx := strings.Index(text, "AI:"); idx != -1 {
		finalText := text[idx+3:]
		h.Ctx.SSEvent("final_answer", "\n"+finalText)
		h.hasFinalAnswer = true
	} else {
		h.immediateStepsBuilder.WriteString(text)
		h.Ctx.SSEvent("immediate_steps", text)
	}

	h.Ctx.Writer.Flush()
}

func (h *GinSSEHandler) SaveAgentSteps(ctx context.Context) error {
	return h.ChatHistory.SetImmediateSteps(ctx, h.immediateStepsBuilder.String())
}
