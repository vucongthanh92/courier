package entities

import "time"

type PasswordReset struct {
	ID          int64      `gorm:"primaryKey;autoIncrement:true;check:id>0" json:"id"`
	UserID      int64      `gorm:"index;not null" json:"user_id"`
	User        User       `gorm:"constraint:OnDelete:CASCADE"`
	TokenHash   string     `gorm:"type:text;not null" json:"-"`
	RequestedAt time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"requested_at"`
	ExpiresAt   time.Time  `gorm:"type:timestamptz;not null" json:"expires_at"`
	UsedAt      *time.Time `gorm:"type:timestamptz" json:"used_at,omitempty"`
	IP          *string    `gorm:"type:inet" json:"ip,omitempty"`
	UserAgent   *string    `gorm:"type:text" json:"user_agent,omitempty"`
}

func (PasswordReset) TableName() string {
	return "password_resets"
}
