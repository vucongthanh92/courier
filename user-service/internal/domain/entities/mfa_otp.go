package entities

import "time"

type MFAOTP struct {
	ID                uint64     `gorm:"column:id;primaryKey;autoIncrement:true;check:id>0" json:"id"`
	UserID            uint64     `gorm:"column:user_id;primaryKey" json:"user_id"`
	SecretEnc         []byte     `gorm:"column:secret_enc;type:bytea;not null" json:"-"`
	Issuer            string     `gorm:"column:issuer;type:text;not null" json:"issuer"`
	Label             string     `gorm:"column:label;type:text;not null" json:"label"`
	RecoveryCodesHash []string   `gorm:"column:recovery_codes_hash;type:text[];not null" json:"-"`
	CreatedAt         time.Time  `gorm:"column:created_at;type:timestamptz;autoCreateTime" json:"created_at"`
	LastUsedAt        *time.Time `gorm:"column:last_used_at;type:timestamptz" json:"last_used_at,omitempty"`
}

func (MFAOTP) TableName() string {
	return `"user-service".mfa_otp`
}
