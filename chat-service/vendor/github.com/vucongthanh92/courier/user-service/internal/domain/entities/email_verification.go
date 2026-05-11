package entities

import "time"

type EmailVerification struct {
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement:true;check:id>0" json:"id"`
	UserID    uint64     `gorm:"column:user_id;index;not null" json:"user_id"`
	Email     string     `gorm:"column:email;type:citext;not null" json:"email"`
	TokenHash string     `gorm:"column:token_hash;type:text" json:"-"`
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamptz;autoCreateTime" json:"created_at"`
	ExpiresAt time.Time  `gorm:"column:expires_at;type:timestamptz;not null" json:"expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at;type:timestamptz" json:"used_at,omitempty"`
}

func (EmailVerification) TableName() string {
	return `"user-service".email_verifications`
}
