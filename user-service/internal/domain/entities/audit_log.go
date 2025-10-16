package entities

import (
	"time"

	"gorm.io/datatypes"
)

type AuditLog struct {
	ID        int64             `gorm:"primaryKey;autoIncrement:true;check:id>0" json:"id"`
	UserID    *int64            `gorm:"index" json:"user_id,omitempty"`
	User      *User             `gorm:"constraint:OnDelete:SET NULL"`
	Action    string            `gorm:"type:audit_action_enum;not null;index" json:"action"`
	IP        *string           `gorm:"type:inet" json:"ip,omitempty"`
	UserAgent *string           `gorm:"type:text" json:"user_agent,omitempty"`
	Meta      datatypes.JSONMap `gorm:"type:jsonb;not null;default:'{}'" json:"meta"`
	CreatedAt time.Time         `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_logs" }
