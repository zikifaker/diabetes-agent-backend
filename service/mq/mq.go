package mq

import (
	"context"
	"diabetes-agent-backend/config"
	"diabetes-agent-backend/service/knowledge-base/etl"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"github.com/avast/retry-go/v4"
)

const (
	TopicKnowledgeBase = "topic_knowledge_base"

	TagETL = "tag_etl"

	consumerGroup = "cg_knowledge_base_etl"

	sendAttempts = 3
)

var (
	// 生产者实例
	producerInstance rocketmq.Producer

	// 消费者实例
	consumerInstance rocketmq.PushConsumer

	// 消息处理器表
	handlers = make(map[string]MessageHandler)
)

type MessageHandler func(context.Context, *primitive.MessageExt) error

type Message struct {
	Topic   string
	Tag     string
	Payload any
}

func init() {
	// 设置RocketMQ客户端（使用rlog）的日志级别
	rlog.SetLogLevel("warn")

	var err error
	consumerInstance, err = rocketmq.NewPushConsumer(
		consumer.WithNameServer(config.Cfg.MQ.NameServer),
		consumer.WithGroupName(consumerGroup),
		consumer.WithConsumerModel(consumer.Clustering),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create consumer: %v", err))
	}

	producerInstance, err = rocketmq.NewProducer(
		producer.WithNameServer(config.Cfg.MQ.NameServer),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create producer: %v", err))
	}
}

func Run() error {
	// 注册消息处理器
	if err := registerHandler(TopicKnowledgeBase, TagETL, etl.HandleETLMessage); err != nil {
		return fmt.Errorf("failed to register handler: %v", err)
	}

	if err := producerInstance.Start(); err != nil {
		return fmt.Errorf("failed to start producer: %v", err)
	}

	if err := consumerInstance.Start(); err != nil {
		return fmt.Errorf("failed to start consumer: %v", err)
	}
	return nil
}

// registerHandler 注册消息处理器
func registerHandler(topic string, tag string, handler MessageHandler) error {
	handlers[topic] = handler

	selector := consumer.MessageSelector{}
	if tag != "" {
		selector = consumer.MessageSelector{
			Type:       consumer.TAG,
			Expression: tag,
		}
	}

	err := consumerInstance.Subscribe(topic, selector, func(ctx context.Context, messages ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range messages {
			h := handlers[msg.Topic]
			if h == nil {
				slog.Warn("No message handler found for topic", "topic", msg.Topic)
				continue
			}

			if err := h(ctx, msg); err != nil {
				slog.Error("Failed to process message",
					"topic", msg.Topic,
					"msg_id", msg.MsgId,
					"error", err)
				return consumer.ConsumeRetryLater, err
			}
		}
		return consumer.ConsumeSuccess, nil
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %v", topic, err)
	}

	return nil
}

// SendMessage 向MQ发送消息
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
		retry.Attempts(sendAttempts),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			slog.Warn("Retrying to send message",
				"attempt", n+1,
				"topic", msg.Topic,
				"err", err)
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to send message to topic %s after retries: %v", msg.Topic, err)
	}

	return nil
}

// Shutdown 关闭MQ服务
func Shutdown() {
	if producerInstance != nil {
		producerInstance.Shutdown()
	}
	if consumerInstance != nil {
		consumerInstance.Shutdown()
	}
}
