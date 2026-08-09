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
	defaultChatEventsTopic = "courier.chat.events.v1"
)

type ChatEventPublisher struct {
	writer *kafkago.Writer
	source string
}

func InitChatEventPublisher(cfg *config.AppConfig) interfaces.ChatEventPublisherI {
	brokers := []string{"localhost:9092"}
	topic := defaultChatEventsTopic
	source := "chat-service"
	if cfg != nil {
		if strings.TrimSpace(cfg.ServiceName) != "" {
			source = cfg.ServiceName
		}
		if cfg.Kafka != nil && len(cfg.Kafka.Brokers) > 0 {
			brokers = cfg.Kafka.Brokers
		}
		if cfg.Kafka != nil && strings.TrimSpace(cfg.Kafka.Topics.ChatEvents) != "" {
			topic = cfg.Kafka.Topics.ChatEvents
		}
	}

	return &ChatEventPublisher{
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafkago.Hash{},
			RequiredAcks: kafkago.RequireAll,
		},
		source: source,
	}
}

func (p *ChatEventPublisher) PublishConversationCreated(ctx context.Context, payload models.ConversationCreatedPayload) error {
	eventID := "conversation-created-" + strconv.FormatUint(payload.ConversationID, 10) + "-" + utils.RandString(12)
	envelope := models.IntegrationEventEnvelope{
		EventID:       eventID,
		EventType:     models.ConversationCreatedEventType,
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
