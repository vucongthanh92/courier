package entities

import (
	"strconv"
	"time"
)

type LedgerJournal struct {
	ID             uint64    `gorm:"column:id;primaryKey"`
	ReferenceType  string    `gorm:"column:reference_type"`
	ReferenceID    string    `gorm:"column:reference_id"`
	SourceType     string    `gorm:"column:source_type"`
	SourceProvider *string   `gorm:"column:source_provider"`
	Status         string    `gorm:"column:status"`
	ReversalOfID   *uint64   `gorm:"column:reversal_of_id"`
	Narrative      string    `gorm:"column:narrative"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	PostedAt       time.Time `gorm:"column:posted_at"`
}

func (LedgerJournal) TableName() string {
	return `"payment-gateway".ledger_journals`
}

func (e *LedgerJournal) Initialize(
	journalID uint64,
	referenceType string,
	referenceID int64,
	sourceType string,
	sourceProvider *string,
	status string,
	narrative *string,
) {
	e.ID = journalID
	e.ReferenceType = referenceType
	e.ReferenceID = strconv.FormatInt(referenceID, 10)
	e.SourceType = sourceType
	e.SourceProvider = sourceProvider
	e.Status = status

	if narrative != nil {
		e.Narrative = "SePay wallet top-up " + *narrative
	} else {
		e.Narrative = "cannot find narrative"
	}
}
