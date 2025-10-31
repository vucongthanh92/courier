package entities

import "time"

type RefreshToken struct {
	ID           uint64     `gorm:"column:id;primaryKey;autoIncrement:true;check:id>0" json:"id"`
	UserID       uint64     `gorm:"column:user_id;index;not null" json:"user_id"`
	TokenHash    string     `gorm:"column:token_hash;type:text;uniqueIndex;not null" json:"-"`
	ParentID     *int64     `gorm:"column:parent_id" json:"parent_id,omitempty"`
	ReplacedByID *int64     `gorm:"column:replaced_by_id" json:"replaced_by_id,omitempty"`
	UserAgent    *string    `gorm:"column:user_agent;type:text" json:"user_agent,omitempty"`
	IP           *string    `gorm:"column:ip;type:inet" json:"ip,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at;type:timestamptz;autoCreateTime" json:"created_at"`
	ExpiresAt    time.Time  `gorm:"column:expires_at;type:timestamptz;not null" json:"expires_at"`
	RevokedAt    *time.Time `gorm:"column:revoked_at;type:timestamptz" json:"revoked_at,omitempty"`
}

func (RefreshToken) TableName() string {
	return `"user-service".refresh_tokens`
}
