package entities

import "time"

type AuthCredential struct {
	ID                int64     `gorm:"primaryKey;autoIncrement:true;check:id>0" json:"id"`
	UserID            int64     `gorm:"primaryKey" json:"user_id"`
	User              User      `gorm:"constraint:OnDelete:CASCADE"`
	PasswordHash      string    `gorm:"type:text;not null" json:"-"`
	PasswordAlgo      string    `gorm:"type:text;not null;check:password_algo IN ('argon2id','bcrypt','scrypt')" json:"-"`
	MFAEnabled        bool      `gorm:"not null;default:false" json:"mfa_enabled"`
	PasswordUpdatedAt time.Time `gorm:"type:timestamptz;not null;default:now()" json:"password_updated_at"`
	PasswordVersion   int16     `gorm:"not null;default:1" json:"password_version"`
	UpdatedAt         time.Time `gorm:"type:timestamptz;autoUpdateTime" json:"updated_at"`
}

func (AuthCredential) TableName() string {
	return "auth_credentials"
}
