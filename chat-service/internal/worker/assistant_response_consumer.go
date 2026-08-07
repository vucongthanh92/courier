package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/vucongthanh92/courier/chat-service/config"
	"github.com/vucongthanh92/courier/chat-service/helper/constants"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
	"github.com/vucongthanh92/go-base-utils/logger"
	"go.uber.org/zap"
)

const (
	defaultAssistantRespondedTopic = "courier.assistant.responded.v1"
	assistantRespondedEventType    = "assistant.responded"
)

type AssistantResponseConsumer struct {
	reader         *kafkago.Reader
	messageService interfaces.MessageServiceI
	logger         logger.Logger
}

func InitAssistantResponseConsumer(
	cfg *config.AppConfig,
	messageService interfaces.MessageServiceI,
	log logger.Logger,
) *AssistantResponseConsumer {
	brokers := []string{defaultBroker}
	groupID := defaultGroup
	topic := defaultAssistantRespondedTopic
	if cfg != nil && cfg.Kafka != nil {
		if len(cfg.Kafka.Brokers) > 0 {
			brokers = cfg.Kafka.Brokers
		}
		if strings.TrimSpace(cfg.Kafka.GroupID) != "" {
			groupID = cfg.Kafka.GroupID
		}
		if strings.TrimSpace(cfg.Kafka.Topics.AssistantResponded) != "" {
			topic = cfg.Kafka.Topics.AssistantResponded
		}
	}

	return &AssistantResponseConsumer{
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:        brokers,
			GroupID:        groupID + "-assistant-response",
			Topic:          topic,
			CommitInterval: 0,
			MinBytes:       1,
			MaxBytes:       10e6,
		}),
		messageService: messageService,
		logger:         log,
	}
}

func (c *AssistantResponseConsumer) Start(ctx context.Context) error {
	c.logger.Info("starting Kafka assistant response consumer")
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			c.logger.Error("fetch Kafka assistant response failed", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		if handleErr := c.handleMessage(ctx, msg.Value); handleErr != nil {
			c.logger.Error("handle Kafka assistant response failed", zap.Any("error", handleErr), zap.String("topic", msg.Topic), zap.Int("partition", msg.Partition), zap.Int64("offset", msg.Offset))
			continue
		}
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("commit Kafka assistant response failed", zap.Error(err), zap.String("topic", msg.Topic), zap.Int("partition", msg.Partition), zap.Int64("offset", msg.Offset))
		}
	}
}

func (c *AssistantResponseConsumer) handleMessage(ctx context.Context, body []byte) error {
	var envelope models.IntegrationEventEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return err
	}
	if envelope.EventType != assistantRespondedEventType || envelope.EventVersion != 1 {
		return nil
	}
	var payload models.AssistantRespondedPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return err
	}
	parts := payload.MessageParts
	if len(parts) == 0 && strings.TrimSpace(payload.Body) != "" {
		parts = []models.AssistantMessagePart{{Body: payload.Body, Index: 1, Total: 1}}
	}
	for _, part := range parts {
		metadata := map[string]any{
			"event_id":              envelope.EventID,
			"event_type":            "assistant.responded.v1",
			"correlation_id":        payload.CorrelationID,
			"triggering_message_id": payload.TriggeringMessageID,
			"part_index":            part.Index,
			"part_total":            part.Total,
		}
		for key, value := range payload.Metadata {
			metadata[key] = value
		}
		for key, value := range part.Metadata {
			metadata[key] = value
		}
		_, _, errBuilder := c.messageService.CreateSystemMessage(ctx, &models.CreateSystemMessageRequest{
			ConversationID: payload.ConversationID,
			Type:           constants.MessageTypeText,
			Body:           part.Body,
			Metadata:       metadata,
		})
		if errBuilder != nil {
			return fmt.Errorf("create assistant system message failed: %v", errBuilder)
		}
	}
	return nil
}

func (c *AssistantResponseConsumer) Close() error {
	return c.reader.Close()
}
