package entities

import "time"

type AuthCredential struct {
	ID              uint64    `gorm:"column:id;primaryKey;autoIncrement:true;check:id>0" json:"id"`
	UserID          uint64    `gorm:"column:user_id" json:"user_id"`
	PasswordHash    string    `gorm:"column:password_hash;type:text;not null" json:"-"`
	PasswordAlgo    string    `gorm:"column:password_algo;type:text;not null;check:password_algo IN ('argon2id','bcrypt','scrypt')" json:"-"`
	MFAEnabled      bool      `gorm:"column:mfa_enabled;not null;default:false" json:"mfa_enabled"`
	PasswordVersion int16     `gorm:"column:password_version;not null;default:1" json:"password_version"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;type:timestamptz;autoUpdateTime" json:"updated_at"`
}

func (AuthCredential) TableName() string {
	return `"user-service".auth_credential`
}
