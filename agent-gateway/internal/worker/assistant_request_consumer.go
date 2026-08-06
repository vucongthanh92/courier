package worker

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/vucongthanh92/courier/agent-gateway/config"
	"github.com/vucongthanh92/courier/agent-gateway/helper/constants"
	"github.com/vucongthanh92/courier/agent-gateway/internal/domain/models"
	"github.com/vucongthanh92/courier/agent-gateway/internal/gateway"
	"github.com/vucongthanh92/courier/agent-gateway/internal/repository/external/kafka"
)

type AssistantRequestConsumer struct {
	reader    *kafkago.Reader
	gateway   *gateway.Service
	publisher *kafka.Publisher
}

func NewAssistantRequestConsumer(cfg config.AppConfig, gateway *gateway.Service, publisher *kafka.Publisher) *AssistantRequestConsumer {
	return &AssistantRequestConsumer{
		reader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:        cfg.Kafka.Brokers,
			GroupID:        cfg.Kafka.GroupID,
			Topic:          cfg.Kafka.AssistantRequestedTopic,
			CommitInterval: 0,
			MinBytes:       1,
			MaxBytes:       10e6,
		}),
		gateway:   gateway,
		publisher: publisher,
	}
}

func (c *AssistantRequestConsumer) Start(ctx context.Context) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			time.Sleep(time.Second)
			continue
		}

		if err := c.handleMessage(ctx, msg.Value); err != nil {
			continue
		}
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			continue
		}
	}
}

func (c *AssistantRequestConsumer) handleMessage(ctx context.Context, body []byte) error {
	var envelope models.EventEnvelope[json.RawMessage]
	if err := json.Unmarshal(body, &envelope); err != nil {
		return err
	}
	if envelope.EventType != constants.AssistantRequestedEventType || envelope.EventVersion != 1 {
		return nil
	}
	var payload models.AssistantRequestedPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return err
	}
	if payload.CorrelationID == "" {
		payload.CorrelationID = envelope.EventID
	}
	response, err := c.gateway.ProcessAssistantRequest(ctx, payload)
	if err != nil {
		return err
	}
	return c.publisher.PublishAssistantResponded(ctx, response)
}

func (c *AssistantRequestConsumer) Close() error {
	return c.reader.Close()
}
