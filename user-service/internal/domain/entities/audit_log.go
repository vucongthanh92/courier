package entities

import (
	"time"

	"gorm.io/datatypes"
)

type AuditLog struct {
	ID        uint64            `gorm:"column:id;primaryKey;autoIncrement:true;check:id>0" json:"id"`
	UserID    uint64            `gorm:"column:user_id;index" json:"user_id"`
	Action    string            `gorm:"column:action;type:varchar(50);" json:"action"`
	IP        string            `gorm:"column:ip;type:inet" json:"ip"`
	UserAgent string            `gorm:"column:user_agent;type:text" json:"user_agent"`
	Meta      datatypes.JSONMap `gorm:"column:meta;type:jsonb;not null;default:'{}'" json:"meta"`
	CreatedAt time.Time         `gorm:"column:created_at;type:timestamptz;autoCreateTime" json:"created_at"`
}

func (AuditLog) TableName() string {
	return `"user-service".audit_logs`
}
