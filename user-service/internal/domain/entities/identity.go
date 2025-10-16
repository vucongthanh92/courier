package entities

import "time"

type Identity struct {
	ID              int64      `gorm:"primaryKey;autoIncrement:true;check:id>0" json:"id"`
	UserID          int64      `gorm:"index;not null" json:"user_id"`
	User            User       `gorm:"constraint:OnDelete:CASCADE"`
	Provider        string     `gorm:"type:identity_provider_enum;not null;index" json:"provider"`
	ProviderUID     string     `gorm:"type:text;not null" json:"provider_uid"`
	EmailAtAuth     *string    `gorm:"type:citext" json:"email_at_auth,omitempty"`
	Scopes          []string   `gorm:"type:text[]" json:"scopes"`
	AccessTokenEnc  []byte     `gorm:"type:bytea" json:"-"`
	RefreshTokenEnc []byte     `gorm:"type:bytea" json:"-"`
	ExpiresAt       *time.Time `gorm:"type:timestamptz" json:"expires_at,omitempty"`
	CreatedAt       time.Time  `gorm:"type:timestamptz;autoCreateTime" json:"created_at"`
}

func (Identity) TableName() string {
	return "identities"
}
