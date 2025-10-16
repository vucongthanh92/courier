package entities

import "time"

type MFAOTP struct {
	ID                int64      `gorm:"primaryKey;autoIncrement:true;check:id>0" json:"id"`
	UserID            int64      `gorm:"primaryKey" json:"user_id"`
	User              User       `gorm:"constraint:OnDelete:CASCADE"`
	SecretEnc         []byte     `gorm:"type:bytea;not null" json:"-"`
	Issuer            string     `gorm:"type:text;not null" json:"issuer"`
	Label             string     `gorm:"type:text;not null" json:"label"`
	RecoveryCodesHash []string   `gorm:"type:text[];not null" json:"-"`
	CreatedAt         time.Time  `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
	LastUsedAt        *time.Time `gorm:"type:timestamptz" json:"last_used_at,omitempty"`
}

func (MFAOTP) TableName() string { return "mfa_otp" }
