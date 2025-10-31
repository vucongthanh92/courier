package entities

import (
	"time"
)

type User struct {
	ID            uint64     `gorm:"column:id;primaryKey;type:bigint;check:id>0" json:"id"`
	Email         string     `gorm:"column:email;type:citext;uniqueIndex" json:"email"`
	EmailVerified bool       `gorm:"column:email_verified;not null;default:false" json:"email_verified"`
	PhoneNumber   string     `gorm:"column:phone_number;type:varchar(50);" json:"phone_number"`
	PhoneVerified bool       `gorm:"column:phone_verified;not null;default:false" json:"phone_verified"`
	DisplayName   string     `gorm:"column:display_name;type:varchar(255)" json:"display_name,omitempty"`
	AvatarURL     string     `gorm:"column:avatar_url;type:text" json:"avatar_url,omitempty"`
	Status        string     `gorm:"column:status;type:user_status_enum;not null;default:'active'" json:"status"`
	CreatedAt     time.Time  `gorm:"column:created_at;type:timestamptz;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;type:timestamptz;autoUpdateTime" json:"updated_at"`
	DeletedAt     *time.Time `gorm:"column:deleted_at;type:timestamptz" json:"deleted_at,omitempty"`
}

func (u *User) TableName() string {
	return `"user-service".users`
}
