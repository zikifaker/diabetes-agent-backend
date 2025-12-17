package controller

import (
	"context"
	"diabetes-agent-backend/request"
	"diabetes-agent-backend/service/chat"
	"diabetes-agent-backend/service/summarization"
	"diabetes-agent-backend/utils"
	"errors"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/agents"
)

func AgentChat(c *gin.Context) {
	utils.SetSSEHeaders(c)

	var req request.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(ErrParseRequest.Error(), "err", err)
		utils.SendSSEMessage(c, "error", ErrParseRequest.Error())
		return
	}

	agent, err := chat.NewAgent(c, req)
	if err != nil {
		slog.Error(ErrCreateAgent.Error(), "err", err)
		utils.SendSSEMessage(c, "error", ErrCreateAgent.Error())
		return
	}
	defer agent.Close()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	go func() {
		<-c.Done()
		cancel()
	}()

	if err := agent.Call(ctx, req); err != nil {
		targetErr := agents.ErrUnableToParseOutput

		// 若发生解析 Agent 输出错误，提取最终答案，进行推送和持久化
		if errors.Is(err, targetErr) {
			slog.Warn(targetErr.Error())

			result := strings.TrimPrefix(err.Error(), targetErr.Error()+":")
			utils.SendSSEMessage(c, "final_answer", result)
			utils.SendSSEMessage(c, "done", "")

			agent.SaveFinalAnswer(ctx, result)
		} else {
			slog.Error(ErrCallAgent.Error(), "err", err)

			utils.SendSSEMessage(c, "error", ErrCallAgent.Error())
			utils.SendSSEMessage(c, "done", "")
			return
		}
	}

	if err := agent.SaveAgentSteps(ctx); err != nil {
		slog.Error(ErrSaveAgentSteps.Error(), "err", err)
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
