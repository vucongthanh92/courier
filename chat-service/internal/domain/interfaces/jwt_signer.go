package interfaces

import (
	"context"
	"time"

	errHandler "github.com/vucongthanh92/courier/chat-service/helper/error_handler"
)

type JWTSignerI interface {
	SignAccessToken(user User, now time.Time, ttl time.Duration) (string, *errHandler.ErrorBuilder)
}

type JWKQueryRepoI interface {
	GetActiveKey(ctx context.Context) (JWKKey, *errHandler.ErrorBuilder)
}

type JWKPublicKeyProviderI interface {
	GetPublicKey(ctx context.Context, kid string) (kidOut string, publicPEM string, alg string, err *errHandler.ErrorBuilder)
}

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
