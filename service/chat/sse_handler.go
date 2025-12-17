package chat

import (
	"context"
	"diabetes-agent-backend/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/callbacks"
)

const (
	// Agent 输出缓冲区大小阈值
	prefixBufferMaxKeep = 10

	// 最终答案的前缀
	finalAnswerPrefix = "AI:"

	eventImmediateSteps = "immediate_steps"
	eventFinalAnswer    = "final_answer"
	eventToolCallResult = "tool_call_result"
)

// GinSSEHandler 基于 Gin 的回调处理器，使用 SSE 发送 Agent 的输出内容
type GinSSEHandler struct {
	callbacks.SimpleHandler

	Ctx     *gin.Context
	Session string

	// Agent 输出中是否包含最终答案
	hasFinalAnswer bool

	// 缓冲区，用于跨 chunk 识别最终答案的前缀
	prefixBuffer *strings.Builder

	// 存储 Agent 的思考步骤
	immediateStepsBuilder *strings.Builder
}

var _ callbacks.Handler = &GinSSEHandler{}

func NewGinSSEHandler(ctx *gin.Context, session string) *GinSSEHandler {
	return &GinSSEHandler{
		Ctx:                   ctx,
		Session:               session,
		hasFinalAnswer:        false,
		prefixBuffer:          &strings.Builder{},
		immediateStepsBuilder: &strings.Builder{},
	}
}

func (h *GinSSEHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	text := string(chunk)

	if h.hasFinalAnswer {
		utils.SendSSEMessage(h.Ctx, eventFinalAnswer, text)
		return
	}

	h.prefixBuffer.WriteString(text)
	bufferStr := h.prefixBuffer.String()

	if idx := strings.Index(bufferStr, finalAnswerPrefix); idx != -1 {
		// 前缀前为思考内容
		before := bufferStr[:idx]
		if len(before) > 0 {
			h.immediateStepsBuilder.WriteString(before)
			utils.SendSSEMessage(h.Ctx, eventImmediateSteps, before)
		}

		// 前缀后为最终答案
		after := bufferStr[idx+len(finalAnswerPrefix):]
		if len(after) > 0 {
			utils.SendSSEMessage(h.Ctx, eventFinalAnswer, after)
		}

		h.hasFinalAnswer = true
		h.prefixBuffer.Reset()
	} else {
		// 保留最后 prefixBufferMaxKeep 个 rune，防止缓冲区过大
		if h.prefixBuffer.Len() > 0 {
			runes := []rune(bufferStr)
			if len(runes) > prefixBufferMaxKeep {
				flushRunes := runes[:len(runes)-prefixBufferMaxKeep]
				flushText := string(flushRunes)
				h.immediateStepsBuilder.WriteString(flushText)
				utils.SendSSEMessage(h.Ctx, eventImmediateSteps, flushText)

				remaining := string(runes[len(runes)-prefixBufferMaxKeep:])
				h.prefixBuffer.Reset()
				h.prefixBuffer.WriteString(remaining)
			}
		}
	}
}

func (h *GinSSEHandler) HandleToolEnd(ctx context.Context, result string) {
	utils.SendSSEMessage(h.Ctx, eventToolCallResult, result)
}

func (h *GinSSEHandler) GetImmediateSteps() string {
	return h.immediateStepsBuilder.String()
}
