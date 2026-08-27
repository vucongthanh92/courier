package entities

import "time"

type IdempotencyKey struct {
	ID             uint64    `gorm:"column:id;primaryKey"`
	Scope          string    `gorm:"column:scope"`
	IdempotencyKey string    `gorm:"column:idempotency_key"`
	RequestHash    string    `gorm:"column:request_hash"`
	ResponseStatus *int16    `gorm:"column:response_status"`
	ResponseBody   []byte    `gorm:"column:response_body;type:jsonb"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	ExpiresAt      time.Time `gorm:"column:expires_at"`
}

func (IdempotencyKey) TableName() string { return `"payment-gateway".idempotency_keys` }

type OutboxEvent struct {
	ID            uint64     `gorm:"column:id;primaryKey"`
	AggregateType string     `gorm:"column:aggregate_type"`
	AggregateID   string     `gorm:"column:aggregate_id"`
	EventType     string     `gorm:"column:event_type"`
	Payload       []byte     `gorm:"column:payload;type:jsonb"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	PublishedAt   *time.Time `gorm:"column:published_at"`
	Attempts      int        `gorm:"column:attempts"`
	LastError     *string    `gorm:"column:last_error"`
}

func (OutboxEvent) TableName() string { return `"payment-gateway".outbox_events` }

type AuditLog struct {
	ID           uint64    `gorm:"column:id;primaryKey"`
	ActorType    string    `gorm:"column:actor_type"`
	ActorID      *string   `gorm:"column:actor_id"`
	Action       string    `gorm:"column:action"`
	ResourceType string    `gorm:"column:resource_type"`
	ResourceID   string    `gorm:"column:resource_id"`
	IP           *string   `gorm:"column:ip"`
	Metadata     []byte    `gorm:"column:metadata;type:jsonb"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (AuditLog) TableName() string { return `"payment-gateway".audit_logs` }

type ReconciliationRun struct {
	ID          uint64     `gorm:"column:id;primaryKey"`
	Provider    string     `gorm:"column:provider"`
	PeriodStart time.Time  `gorm:"column:period_start"`
	PeriodEnd   time.Time  `gorm:"column:period_end"`
	Status      string     `gorm:"column:status"`
	StartedAt   time.Time  `gorm:"column:started_at"`
	CompletedAt *time.Time `gorm:"column:completed_at"`
	Summary     []byte     `gorm:"column:summary;type:jsonb"`
}

func (ReconciliationRun) TableName() string { return `"payment-gateway".reconciliation_runs` }

type ReconciliationItem struct {
	ID                    uint64    `gorm:"column:id;primaryKey"`
	ReconciliationRunID   uint64    `gorm:"column:reconciliation_run_id"`
	ProviderTransactionID *string   `gorm:"column:provider_transaction_id"`
	TopUpIntentID         *uint64   `gorm:"column:topup_intent_id"`
	Status                string    `gorm:"column:status"`
	Detail                []byte    `gorm:"column:detail;type:jsonb"`
	CreatedAt             time.Time `gorm:"column:created_at"`
}

func (ReconciliationItem) TableName() string { return `"payment-gateway".reconciliation_items` }
