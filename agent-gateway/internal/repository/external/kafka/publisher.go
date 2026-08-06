package kafka

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/vucongthanh92/courier/agent-gateway/config"
	"github.com/vucongthanh92/courier/agent-gateway/helper/constants"
	"github.com/vucongthanh92/courier/agent-gateway/helper/utils"
	"github.com/vucongthanh92/courier/agent-gateway/internal/domain/models"
)

type Publisher struct {
	writer *kafkago.Writer
	source string
}

func NewPublisher(cfg config.AppConfig) *Publisher {
	return &Publisher{
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(cfg.Kafka.Brokers...),
			Topic:        cfg.Kafka.AssistantRespondedTopic,
			Balancer:     &kafkago.Hash{},
			RequiredAcks: kafkago.RequireAll,
		},
		source: cfg.ServiceName,
	}
}

func (p *Publisher) PublishAssistantResponded(ctx context.Context, payload models.AssistantRespondedPayload) error {
	eventID := payload.CorrelationID
	if eventID == "" {
		eventID = utils.NewCorrelationID()
	}
	envelope := models.EventEnvelope[models.AssistantRespondedPayload]{
		EventID:       eventID,
		EventType:     constants.AssistantRespondedEventType,
		EventVersion:  1,
		OccurredAt:    time.Now().UTC(),
		Source:        p.source,
		AggregateType: "conversation",
		AggregateID:   strconv.FormatUint(payload.ConversationID, 10),
		Payload:       payload,
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(envelope.AggregateID),
		Value: body,
		Time:  envelope.OccurredAt,
	})
}

func (p *Publisher) Close() error {
	return p.writer.Close()
}
