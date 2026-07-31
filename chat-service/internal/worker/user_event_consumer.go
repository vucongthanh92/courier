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

const (
	defaultBroker = "localhost:9092"
	defaultGroup  = "chat-service"
	defaultTopic  = "courier.user.events.v1"
)

type UserEventConsumer struct {
	reader  *kafkago.Reader
	handler *conversationUc.UserEventHandler
	logger  logger.Logger
}

func InitUserEventConsumer(
	cfg *config.AppConfig,
	handler *conversationUc.UserEventHandler,
	log logger.Logger,
) *UserEventConsumer {
	brokers := []string{defaultBroker}
	groupID := defaultGroup
	topic := defaultTopic
	if cfg != nil && cfg.Kafka != nil {
		if len(cfg.Kafka.Brokers) > 0 {
			brokers = cfg.Kafka.Brokers
		}
		if strings.TrimSpace(cfg.Kafka.GroupID) != "" {
			groupID = cfg.Kafka.GroupID
		}
		if strings.TrimSpace(cfg.Kafka.Topics.UserEvents) != "" {
			topic = cfg.Kafka.Topics.UserEvents
		}
	}

	return &UserEventConsumer{
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:        brokers,
			GroupID:        groupID,
			Topic:          topic,
			CommitInterval: 0,
			MinBytes:       1,
			MaxBytes:       10e6,
		}),
		handler: handler,
		logger:  log,
	}
}

func (c *UserEventConsumer) Start(ctx context.Context) error {
	c.logger.Info("starting Kafka user event consumer")
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			c.logger.Error("fetch Kafka message failed", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		if handleErr := c.handler.HandleUserEmailVerified(ctx, msg.Value); handleErr != nil {
			c.logger.Error("handle Kafka user event failed", zap.Any("error", handleErr), zap.String("topic", msg.Topic), zap.Int("partition", msg.Partition), zap.Int64("offset", msg.Offset))
			continue
		}
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("commit Kafka message failed", zap.Error(err), zap.String("topic", msg.Topic), zap.Int("partition", msg.Partition), zap.Int64("offset", msg.Offset))
		}
	}
}

func (c *UserEventConsumer) Close() error {
	return c.reader.Close()
}
