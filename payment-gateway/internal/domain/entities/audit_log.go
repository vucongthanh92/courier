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

func NewAuditLog(
	id uint64,
	actorType string,
	actorID *string,
	action string,
	resourceType string,
	resourceID string,
	ip *string,
	metadata []byte,
) *AuditLog {
	return &AuditLog{
		ID:           id,
		ActorType:    actorType,
		ActorID:      actorID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IP:           ip,
		Metadata:     metadata,
		CreatedAt:    time.Now().UTC(),
	}
}
