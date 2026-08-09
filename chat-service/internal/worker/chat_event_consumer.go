package worker

import (
	"context"
	"errors"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/vucongthanh92/courier/chat-service/config"
	conversationUc "github.com/vucongthanh92/courier/chat-service/internal/usecase/conversation"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
)

const defaultChatEventsTopic = "courier.chat.events.v1"

type ChatEventConsumer struct {
	reader  *kafkago.Reader
	handler *conversationUc.ChatEventHandler
	logger  logger.Logger
}

func InitChatEventConsumer(
	cfg *config.AppConfig,
	handler *conversationUc.ChatEventHandler,
	log logger.Logger,
) *ChatEventConsumer {
	brokers := []string{defaultBroker}
	groupID := defaultGroup
	topic := defaultChatEventsTopic
	if cfg != nil && cfg.Kafka != nil {
		if len(cfg.Kafka.Brokers) > 0 {
			brokers = cfg.Kafka.Brokers
		}
		if strings.TrimSpace(cfg.Kafka.GroupID) != "" {
			groupID = cfg.Kafka.GroupID
		}
		if strings.TrimSpace(cfg.Kafka.Topics.ChatEvents) != "" {
			topic = cfg.Kafka.Topics.ChatEvents
		}
	}

	return &ChatEventConsumer{
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:        brokers,
			GroupID:        groupID + "-chat-events",
			Topic:          topic,
			CommitInterval: 0,
			MinBytes:       1,
			MaxBytes:       10e6,
		}),
		handler: handler,
		logger:  log,
	}
}

func (c *ChatEventConsumer) Start(ctx context.Context) error {
	c.logger.Info("starting Kafka chat event consumer")
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			c.logger.Error("fetch Kafka chat event failed", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		if handleErr := c.handler.HandleConversationCreated(ctx, msg.Value); handleErr != nil {
			c.logger.Error("handle Kafka chat event failed", zap.Any("error", handleErr), zap.String("topic", msg.Topic), zap.Int("partition", msg.Partition), zap.Int64("offset", msg.Offset))
			continue
		}
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("commit Kafka chat event failed", zap.Error(err), zap.String("topic", msg.Topic), zap.Int("partition", msg.Partition), zap.Int64("offset", msg.Offset))
		}
	}
}

func (c *ChatEventConsumer) Close() error {
	return c.reader.Close()
}
