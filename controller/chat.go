package controller

import (
	"context"
	"diabetes-agent-server/request"
	"diabetes-agent-server/service/chat"
	"diabetes-agent-server/service/mq"
	"diabetes-agent-server/service/summarization"
	"diabetes-agent-server/utils"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func AgentChat(c *gin.Context) {
	utils.SetSSEHeaders(c)

	var req request.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(ErrParseRequest.Error(), "err", err)
		utils.SendSSEMessage(c, utils.EventError, ErrParseRequest)
		utils.SendSSEMessage(c, utils.EventDone, nil)
		return
	}

	agent, err := chat.NewAgent(req, c)
	if err != nil {
		slog.Error(ErrCreateAgent.Error(), "err", err)
		utils.SendSSEMessage(c, utils.EventError, ErrCreateAgent)
		utils.SendSSEMessage(c, utils.EventDone, nil)
		return
	}
	defer agent.Close()

	// 监听客户端的取消信号
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	go func() {
		<-c.Done()
		cancel()
	}()

	if err := agent.Call(ctx, req, c); err != nil {
		slog.Error(ErrCallAgent.Error(), "err", err)
		utils.SendSSEMessage(c, utils.EventError, ErrCallAgent)
		utils.SendSSEMessage(c, utils.EventDone, nil)
		return
	}

	utils.SendSSEMessage(c, utils.EventDone, nil)

	mq.SendMessage(ctx, &mq.Message{
		Topic: mq.TopicAgentChat,
		Tag:   mq.TagCompressContext,
		Payload: summarization.Message{
			MsgIDs: []uint{
				agent.ChatHistory.UserMessageID,
				agent.ChatHistory.AgentMessageID,
			},
		},
	})
}
