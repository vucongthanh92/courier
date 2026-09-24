package entities

import (
	"strconv"
	"time"

	"github.com/vucongthanh92/courier/payment-gateway/helper/constants"
)

type ProviderEvent struct {
	ID              uint64     `gorm:"column:id;primaryKey"`
	Provider        string     `gorm:"column:provider"`
	ProviderEventID string     `gorm:"column:provider_event_id"`
	Payload         []byte     `gorm:"column:payload;type:jsonb"`
	SignatureValid  bool       `gorm:"column:signature_valid"`
	Status          string     `gorm:"column:status"`
	ReceivedAt      time.Time  `gorm:"column:received_at"`
	ProcessedAt     *time.Time `gorm:"column:processed_at"`
	ErrorCode       *string    `gorm:"column:error_code"`
}

func (ProviderEvent) TableName() string {
	return `"payment-gateway".provider_events`
}

func (e *ProviderEvent) Initialize(
	eventID uint64,
	provider string,
	providerEventID int64,
	payload []byte,
) {
	e.ID = eventID
	e.Provider = provider
	e.ProviderEventID = strconv.FormatInt(providerEventID, 10)
	e.Payload = payload
	e.Status = constants.SepayEventStatusReceived
}
