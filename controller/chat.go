package controller

import (
	"context"
	"diabetes-agent-backend/request"
	"diabetes-agent-backend/service/chat"
	"diabetes-agent-backend/service/summarization"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func AgentChat(c *gin.Context) {
	setSSEHeader(c)

	var req request.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Agent Chat", "err", ErrParseRequest)
		c.SSEvent("error", ErrParseRequest.Error())
		c.Writer.Flush()
		return
	}

	agent, err := chat.NewAgent(c, req)
	if err != nil {
		slog.Error("Agent Chat", "err", ErrCreateAgent)
		c.SSEvent("error", ErrCallAgent.Error())
		c.Writer.Flush()
		return
	}
	defer agent.Close()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// 监听客户端是否断开连接
	go func() {
		<-c.Done()
		cancel()
	}()

	result, err := agent.Call(ctx, req)
	if err != nil {
		slog.Error("Agent Chat", "err", ErrCallAgent)
		c.SSEvent("error", result)
		c.Writer.Flush()
		return
	}

	c.SSEvent("done", "")
	c.Writer.Flush()

	if err := agent.SaveAgentSteps(ctx); err != nil {
		slog.Error("Agent Chat", "err", ErrSaveAgentSteps)
	}

	// 注册对话摘要生成任务
	summaryTask := summarization.SummaryTask{
		MessageIDs: []uint{
			agent.ChatHistory.UserMessageID,
			agent.ChatHistory.AgentMessageID,
		},
	}
	summarization.SummarizerInstance.RegisterSummaryTask(summaryTask)
}

func setSSEHeader(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
}
