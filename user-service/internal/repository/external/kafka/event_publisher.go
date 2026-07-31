package kafka

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/vucongthanh92/courier/user-service/config"
	"github.com/vucongthanh92/courier/user-service/internal/domain/entities"
)

type IntegrationEventEnvelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	EventVersion  int             `json:"event_version"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Source        string          `json:"source"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	Payload       json.RawMessage `json:"payload"`
}

type EventPublisher struct {
	writer *kafkago.Writer
	source string
}

func InitEventPublisher(cfg *config.AppConfig) *EventPublisher {
	brokers := []string{}
	if cfg != nil && cfg.Kafka != nil && cfg.Kafka.Config != nil && len(cfg.Kafka.Config.Brokers) > 0 {
		brokers = cfg.Kafka.Config.Brokers
	}
	source := "user-service"
	if cfg != nil && strings.TrimSpace(cfg.ServiceName) != "" {
		source = cfg.ServiceName
	}
	return &EventPublisher{
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(brokers...),
			Topic:        cfg.Kafka.Topics.EmailVerified.TopicName,
			Balancer:     &kafkago.Hash{},
			RequiredAcks: kafkago.RequireAll,
		},
		source: source,
	}
}

func (p *EventPublisher) PublishOutbox(ctx context.Context, event *entities.Outbox) error {
	versionedType := event.EventType
	eventType := strings.TrimSuffix(versionedType, ".v1")
	envelope := IntegrationEventEnvelope{
		EventID:       strconv.FormatUint(event.ID, 10),
		EventType:     eventType,
		EventVersion:  1,
		OccurredAt:    event.CreatedAt,
		Source:        p.source,
		AggregateType: event.AggregateType,
		AggregateID:   event.AggregateID,
		Payload:       json.RawMessage(event.Payload),
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(event.AggregateID),
		Value: body,
		Time:  event.CreatedAt,
	})
}

func (p *EventPublisher) Close() error {
	return p.writer.Close()
}
