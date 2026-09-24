package entities

import "time"

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

func (ReconciliationRun) TableName() string {
	return `"payment-gateway".reconciliation_runs`
}
