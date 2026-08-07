package kafka

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/vucongthanh92/courier/chat-service/config"
	"github.com/vucongthanh92/courier/chat-service/helper/utils"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/chat-service/internal/domain/models"
)

const (
	defaultAssistantRequestedTopic = "courier.assistant.requested.v1"
	assistantRequestedEventType    = "assistant.requested"
)

type AssistantEventPublisher struct {
	writer *kafkago.Writer
	source string
}

func InitAssistantEventPublisher(cfg *config.AppConfig) interfaces.AssistantEventPublisherI {
	brokers := []string{"localhost:9092"}
	topic := defaultAssistantRequestedTopic
	source := "chat-service"
	if cfg != nil {
		if strings.TrimSpace(cfg.ServiceName) != "" {
			source = cfg.ServiceName
		}
		if cfg.Kafka != nil && cfg.Kafka.Brokers != nil && len(cfg.Kafka.Brokers) > 0 {
			brokers = cfg.Kafka.Brokers
		}
		if cfg.Kafka != nil && strings.TrimSpace(cfg.Kafka.Topics.AssistantRequested) != "" {
			topic = cfg.Kafka.Topics.AssistantRequested
		}
	}
	return &AssistantEventPublisher{
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafkago.Hash{},
			RequiredAcks: kafkago.RequireAll,
		},
		source: source,
	}
}

func (p *AssistantEventPublisher) PublishAssistantRequested(ctx context.Context, payload models.AssistantRequestedPayload) error {
	if payload.CorrelationID == "" {
		payload.CorrelationID = utils.RandString(24)
	}
	envelope := models.IntegrationEventEnvelope{
		EventID:       payload.CorrelationID,
		EventType:     assistantRequestedEventType,
		EventVersion:  1,
		OccurredAt:    time.Now().UTC(),
		Source:        p.source,
		AggregateType: "conversation",
		AggregateID:   strconv.FormatUint(payload.ConversationID, 10),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	envelope.Payload = body
	message, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(envelope.AggregateID),
		Value: message,
		Time:  envelope.OccurredAt,
	})
}
