package entities

import "time"

type JWKKey struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement:true"`
	Kid        string    `gorm:"column:kid;uniqueIndex;not null"`
	Alg        string    `gorm:"column:alg;type:text;not null"` // ví dụ "RS256"
	Kty        string    `gorm:"column:kty;type:text;not null"` // ví dụ "sig"
	PublicPEM  string    `gorm:"column:public_pem;type:text;not null"`
	PrivatePEM string    `gorm:"column:private_pem;type:text;not null"`
	Active     bool      `gorm:"column:active;not null;default:false"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
	RotatedAt  time.Time `gorm:"column:rotated_at;"`
	ExpiresAt  time.Time `gorm:"column:expires_at;"`
}

func (JWKKey) TableName() string {
	return `"user-service".jwk_key`
}
