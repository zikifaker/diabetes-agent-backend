package mq

import (
	"context"
	"diabetes-agent-server/config"
	"diabetes-agent-server/service/chat"
	"diabetes-agent-server/service/knowledge-base/etl"
	"diabetes-agent-server/service/summarization"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/apache/rocketmq-client-go/v2"
	c "github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"github.com/avast/retry-go/v4"
)

const (
	TopicKnowledgeBase = "topic_knowledge_base"
	TagETL             = "tag_etl"
	TagDelete          = "tag_delete"

	TopicAgentChat         = "topic_agent_chat"
	TagCompressContext     = "tag_compress_context"
	TagDeleteUploadedFiles = "tag_delete_uploaded_files"

	consumerGroupKnowledgeBase = "cg_knowledge_base"
	consumerGroupAgentChat     = "cg_agent_chat"
	maxReconsumeTimes          = 5
	consumeGoroutineNums       = 10

	sendMessageAttempts = 3
)

var (
	// 全局生产者
	producerInstance rocketmq.Producer

	// 知识库业务消费者
	consumerKnowledgeBase rocketmq.PushConsumer

	// Agent 聊天业务消费者
	consumerAgentChat rocketmq.PushConsumer
)

func init() {
	// 设置 RocketMQ 客户端（使用 rlog）的日志级别
	rlog.SetLogLevel("error")

	var err error
	producerInstance, err = rocketmq.NewProducer(
		producer.WithNameServer(config.Cfg.MQ.NameServer),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create producer: %v", err))
	}

	consumerKnowledgeBase, err = rocketmq.NewPushConsumer(
		c.WithNameServer(config.Cfg.MQ.NameServer),
		c.WithGroupName(consumerGroupKnowledgeBase),
		c.WithConsumerModel(c.Clustering),
		c.WithConsumeFromWhere(c.ConsumeFromLastOffset),
		c.WithMaxReconsumeTimes(maxReconsumeTimes),
		c.WithConsumeGoroutineNums(consumeGoroutineNums),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create knowledge base consumer: %v", err))
	}

	consumerAgentChat, err = rocketmq.NewPushConsumer(
		c.WithNameServer(config.Cfg.MQ.NameServer),
		c.WithGroupName(consumerGroupAgentChat),
		c.WithConsumerModel(c.Clustering),
		c.WithConsumeFromWhere(c.ConsumeFromLastOffset),
		c.WithMaxReconsumeTimes(maxReconsumeTimes),
		c.WithConsumeGoroutineNums(consumeGoroutineNums),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create agent context consumer: %v", err))
	}

	knowledgeBaseDispatcher := NewMessageDispatcher()
	knowledgeBaseDispatcher.Register(TopicKnowledgeBase, TagETL, etl.HandleETLMessage)
	knowledgeBaseDispatcher.Register(TopicKnowledgeBase, TagDelete, etl.HandleDeleteMessage)

	if err := knowledgeBaseDispatcher.Bind(consumerKnowledgeBase); err != nil {
		panic(fmt.Sprintf("Failed to bind dispatcher to knowledge base consumer: %v", err))
	}

	agentChatDispatcher := NewMessageDispatcher()
	agentChatDispatcher.Register(TopicAgentChat, TagCompressContext, summarization.HandleSummarizationMessage)
	agentChatDispatcher.Register(TopicAgentChat, TagDeleteUploadedFiles, chat.HandleDeleteUploadedFilesMessage)

	if err := agentChatDispatcher.Bind(consumerAgentChat); err != nil {
		panic(fmt.Sprintf("Failed to bind dispatcher to agent chat consumer: %v", err))
	}
}

func Run() error {
	if err := producerInstance.Start(); err != nil {
		return fmt.Errorf("failed to start producer: %v", err)
	}
	if err := consumerKnowledgeBase.Start(); err != nil {
		return fmt.Errorf("failed to start knowledge base consumer: %v", err)
	}
	if err := consumerAgentChat.Start(); err != nil {
		return fmt.Errorf("failed to start agent chat consumer: %v", err)
	}
	return nil
}

func SendMessage(ctx context.Context, message *Message) error {
	payloadJSON, err := json.Marshal(message.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	msg := primitive.NewMessage(message.Topic, payloadJSON)
	if message.Tag != "" {
		msg = msg.WithTag(message.Tag)
	}

	err = retry.Do(
		func() error {
			_, err := producerInstance.SendSync(ctx, msg)
			return err
		},
		retry.Attempts(sendMessageAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			slog.Warn("Retrying to send message",
				"attempt", n+1,
				"topic", msg.Topic,
				"err", err,
			)
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to send message to topic %s after retries: %v", msg.Topic, err)
	}

	return nil
}

func Shutdown() {
	if producerInstance != nil {
		producerInstance.Shutdown()
	}
	if consumerKnowledgeBase != nil {
		consumerKnowledgeBase.Shutdown()
	}
	if consumerAgentChat != nil {
		consumerAgentChat.Shutdown()
	}
}
