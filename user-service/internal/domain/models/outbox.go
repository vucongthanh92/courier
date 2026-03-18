package models

type CreateOutboxRequest struct {
	AggregateType string `json:"aggregate_type"`
	AggregateID   string `json:"aggregate_id"`
	EventType     string `json:"event_type"`
	Payload       []byte `json:"payload"`
}
