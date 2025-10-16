package entities

import "time"

type EmailVerification struct {
	ID        int64      `gorm:"primaryKey;autoIncrement:true;check:id>0" json:"id"`
	UserID    int64      `gorm:"index;not null" json:"user_id"`
	User      User       `gorm:"constraint:OnDelete:CASCADE"`
	Email     string     `gorm:"type:citext;not null" json:"email"`
	TokenHash string     `gorm:"type:text;not null" json:"-"`
	CreatedAt time.Time  `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	ExpiresAt time.Time  `gorm:"type:timestamptz;not null" json:"expires_at"`
	UsedAt    *time.Time `gorm:"type:timestamptz" json:"used_at,omitempty"`
}

func (EmailVerification) TableName() string {
	return "email_verifications"
}
