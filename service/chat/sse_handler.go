package chat

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/callbacks"
)

const (
	// Agent输出缓冲区大小阈值
	prefixBufferMaxKeep = 10

	finalAnswerPrefix   = "AI:"
	eventImmediateSteps = "immediate_steps"
	eventFinalAnswer    = "final_answer"
)

// GinSSEHandler 基于Gin的回调处理器，使用SSE发送Agent的输出内容
type GinSSEHandler struct {
	callbacks.SimpleHandler
	Ctx         *gin.Context
	ChatHistory *MySQLChatMessageHistory
	Session     string

	// 缓冲区，用于跨 chunk 识别最终答案的前缀
	prefixBuffer *strings.Builder

	// Agent输出中是否包含最终答案
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
		prefixBuffer:          &strings.Builder{},
		hasFinalAnswer:        false,
		immediateStepsBuilder: &strings.Builder{},
	}
}

func (h *GinSSEHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	text := string(chunk)

	if h.hasFinalAnswer {
		h.Ctx.SSEvent(eventFinalAnswer, text)
		h.Ctx.Writer.Flush()
		return
	}

	// 处于思考阶段
	h.prefixBuffer.WriteString(text)
	bufferStr := h.prefixBuffer.String()

	if idx := strings.Index(bufferStr, finalAnswerPrefix); idx != -1 {
		// 前缀之前为思考内容
		before := bufferStr[:idx]
		if len(before) > 0 {
			h.immediateStepsBuilder.WriteString(before)
			h.Ctx.SSEvent(eventImmediateSteps, before)
		}

		// 前缀之后为最终答案
		after := bufferStr[idx+len(finalAnswerPrefix):]
		if len(after) > 0 {
			h.Ctx.SSEvent(eventFinalAnswer, "\n"+after)
		}

		h.hasFinalAnswer = true

		h.prefixBuffer.Reset()
	} else {
		// 保留最后 prefixBufferMaxKeep 个字符, 防止缓冲区过大
		if h.prefixBuffer.Len() > prefixBufferMaxKeep {
			flushLen := h.prefixBuffer.Len() - prefixBufferMaxKeep
			flushText := bufferStr[:flushLen]

			h.immediateStepsBuilder.WriteString(flushText)
			h.Ctx.SSEvent(eventImmediateSteps, flushText)

			h.prefixBuffer.Reset()
			h.prefixBuffer.WriteString(bufferStr[flushLen:])
		}
	}

	h.Ctx.Writer.Flush()
}

func (h *GinSSEHandler) SaveAgentSteps(ctx context.Context) error {
	return h.ChatHistory.SetImmediateSteps(ctx, h.immediateStepsBuilder.String())
}
