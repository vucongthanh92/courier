package entities

import "time"

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

func (AuditLog) TableName() string {
	return `"payment-gateway".audit_logs`
}
