package entities

import "time"

type Identity struct {
	ID              uint64     `gorm:"column:id;primaryKey;autoIncrement:true;check:id>0" json:"id"`
	UserID          uint64     `gorm:"column:user_id;index;not null" json:"user_id"`
	Provider        string     `gorm:"column:provider;type:identity_provider_enum;not null;index" json:"provider"`
	ProviderUID     string     `gorm:"column:provider_uid;type:text;not null" json:"provider_uid"`
	EmailAtAuth     *string    `gorm:"column:email_at_auth;type:citext" json:"email_at_auth"`
	Scopes          []string   `gorm:"column:scopes;type:text[]" json:"scopes"`
	AccessTokenEnc  []byte     `gorm:"column:access_token_enc;type:bytea" json:"-"`
	RefreshTokenEnc []byte     `gorm:"column:refresh_token_enc;type:bytea" json:"-"`
	ExpiresAt       *time.Time `gorm:"column:expires_at;type:timestamptz" json:"expires_at"`
	CreatedAt       time.Time  `gorm:"column:created_at;type:timestamptz;autoCreateTime" json:"created_at"`
}

func (Identity) TableName() string {
	return `"user-service".identities`
}
