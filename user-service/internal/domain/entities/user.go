package entities

import (
	"time"

	"gorm.io/datatypes"
)

type User struct {
	ID            int64             `gorm:"primaryKey;autoIncrement:false;check:id>0" json:"id"`
	Email         string            `gorm:"type:citext;uniqueIndex" json:"email"`
	EmailVerified bool              `gorm:"not null;default:false"  json:"email_verified"`
	Phone         *string           `gorm:"type:text"               json:"phone,omitempty"`
	PhoneVerified bool              `gorm:"not null;default:false"  json:"phone_verified"`
	DisplayName   *string           `gorm:"type:text"               json:"display_name,omitempty"`
	AvatarURL     *string           `gorm:"type:text"               json:"avatar_url,omitempty"`
	Status        string            `gorm:"type:user_status_enum;not null;default:'active'" json:"status"`
	Metadata      datatypes.JSONMap `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	CreatedAt     time.Time         `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time         `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
	DeletedAt     *time.Time        `gorm:"type:timestamptz" json:"deleted_at,omitempty"`
}

func (u *User) TableName() string {
	return "users"
}
