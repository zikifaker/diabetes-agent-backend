package mq

import (
	"context"
	"log/slog"
	"strings"

	"github.com/apache/rocketmq-client-go/v2"
	c "github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

type TopicRouter struct {
	handlers map[string]MessageHandler
}

type MessageDispatcher struct {
	routes map[string]*TopicRouter
}

type Message struct {
	Topic   string
	Tag     string
	Payload any
}

type MessageHandler func(context.Context, *primitive.MessageExt) error

func NewMessageDispatcher() *MessageDispatcher {
	return &MessageDispatcher{
		routes: make(map[string]*TopicRouter),
	}
}

// Register 注册消息处理器
func (d *MessageDispatcher) Register(topic string, tag string, handler MessageHandler) {
	if _, ok := d.routes[topic]; !ok {
		d.routes[topic] = &TopicRouter{
			handlers: make(map[string]MessageHandler),
		}
	}
	d.routes[topic].handlers[tag] = handler
}

// Bind 绑定消费者和处理器并启动订阅
func (d *MessageDispatcher) Bind(consumer rocketmq.PushConsumer) error {
	for topic, router := range d.routes {
		tags := make([]string, 0, len(router.handlers))
		for tag := range router.handlers {
			tags = append(tags, tag)
		}

		selector := c.MessageSelector{
			Type:       c.TAG,
			Expression: strings.Join(tags, "||"),
		}

		currRouter := router
		err := consumer.Subscribe(topic, selector, func(ctx context.Context, msgs ...*primitive.MessageExt) (c.ConsumeResult, error) {
			for _, msg := range msgs {
				h, ok := currRouter.handlers[msg.GetTags()]
				if !ok {
					slog.Warn("No handler for tag",
						"msg_id", msg.MsgId,
						"topic", msg.Topic,
						"tags", msg.GetTags(),
					)
					continue
				}

				if err := h(ctx, msg); err != nil {
					slog.Error("handle message failed",
						"msg_id", msg.MsgId,
						"topic", msg.Topic,
						"tags", msg.GetTags(),
						"err", err,
					)
					return c.ConsumeRetryLater, err
				}
			}
			return c.ConsumeSuccess, nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
