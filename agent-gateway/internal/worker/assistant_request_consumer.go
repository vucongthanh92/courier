package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
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
	log.Printf("configuring Kafka assistant request consumer: brokers=%v group_id=%s topic=%s", cfg.Kafka.Brokers, cfg.Kafka.GroupID, cfg.Kafka.AssistantRequestedTopic)
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
	log.Println("starting Kafka assistant request consumer")
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			log.Printf("fetch Kafka assistant request failed: %v", err)
			time.Sleep(time.Second)
			continue
		}

		if err := c.handleMessage(ctx, msg.Value); err != nil {
			log.Printf("handle Kafka assistant request failed: topic=%s partition=%d offset=%d error=%v", msg.Topic, msg.Partition, msg.Offset, err)
			continue
		}
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("commit Kafka assistant request failed: topic=%s partition=%d offset=%d error=%v", msg.Topic, msg.Partition, msg.Offset, err)
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
		log.Printf("skip unsupported assistant event: type=%s version=%d", envelope.EventType, envelope.EventVersion)
		return nil
	}
	var payload models.AssistantRequestedPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return err
	}
	if payload.CorrelationID == "" {
		payload.CorrelationID = envelope.EventID
	}
	log.Printf("received assistant request: conversation_id=%d triggering_message_id=%d correlation_id=%s", payload.ConversationID, payload.TriggeringMessageID, payload.CorrelationID)
	response, err := c.gateway.ProcessAssistantRequest(ctx, payload)
	if err != nil {
		return err
	}
	if err := c.publisher.PublishAssistantResponded(ctx, response); err != nil {
		return err
	}
	log.Printf("published assistant response: conversation_id=%d triggering_message_id=%d correlation_id=%s parts=%d", response.ConversationID, response.TriggeringMessageID, response.CorrelationID, len(response.MessageParts))
	return nil
}

func (c *AssistantRequestConsumer) Close() error {
	return c.reader.Close()
}
