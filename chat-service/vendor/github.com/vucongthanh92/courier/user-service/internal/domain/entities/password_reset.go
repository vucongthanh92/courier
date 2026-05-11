package entities

import "time"

type PasswordReset struct {
	ID          uint64     `gorm:"column:id;primaryKey;autoIncrement:true;check:id>0" json:"id"`
	UserID      uint64     `gorm:"column:user_id;index;not null" json:"user_id"`
	TokenHash   string     `gorm:"column:token_hash;type:text;not null" json:"-"`
	RequestedAt time.Time  `gorm:"column:requested_at;type:timestamptz;not null;default:now()" json:"requested_at"`
	ExpiresAt   time.Time  `gorm:"column:expires_at;type:timestamptz;not null" json:"expires_at"`
	UsedAt      *time.Time `gorm:"column:used_at;type:timestamptz" json:"used_at,omitempty"`
	IP          *string    `gorm:"column:ip;type:inet" json:"ip,omitempty"`
	UserAgent   *string    `gorm:"column:user_agent;type:text" json:"user_agent,omitempty"`
}

func (PasswordReset) TableName() string {
	return `"user-service".password_resets`
}
