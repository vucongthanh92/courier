package entities

import "time"

type RefreshToken struct {
	ID           int64      `gorm:"primaryKey;autoIncrement:true;check:id>0" json:"id"`
	UserID       int64      `gorm:"index;not null" json:"user_id"`
	User         User       `gorm:"constraint:OnDelete:CASCADE"`
	TokenHash    string     `gorm:"type:text;uniqueIndex;not null" json:"-"` // SHA-256 hex
	ParentID     *int64     `gorm:"" json:"parent_id,omitempty"`
	ReplacedByID *int64     `gorm:"" json:"replaced_by_id,omitempty"`
	UserAgent    *string    `gorm:"type:text" json:"user_agent,omitempty"`
	IP           *string    `gorm:"type:inet" json:"ip,omitempty"`
	CreatedAt    time.Time  `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	ExpiresAt    time.Time  `gorm:"type:timestamptz;not null" json:"expires_at"`
	RevokedAt    *time.Time `gorm:"type:timestamptz" json:"revoked_at,omitempty"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
