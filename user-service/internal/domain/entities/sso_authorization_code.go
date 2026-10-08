package entities

import "time"

type SsoAuthorizationCode struct {
	ID                  uint64     `gorm:"column:id;primaryKey;type:bigint;check:id>0" json:"id"`
	CodeHash            string     `gorm:"column:code_hash;type:text;uniqueIndex;not null" json:"-"`
	UserID              uint64     `gorm:"column:user_id;index;not null" json:"user_id"`
	ClientID            string     `gorm:"column:client_id;type:text;not null" json:"client_id"`
	RedirectURI         string     `gorm:"column:redirect_uri;type:text;not null" json:"redirect_uri"`
	Scope               string     `gorm:"column:scope;type:text;not null" json:"scope"`
	State               *string    `gorm:"column:state;type:text" json:"state,omitempty"`
	Nonce               *string    `gorm:"column:nonce;type:text" json:"nonce,omitempty"`
	CodeChallenge       string     `gorm:"column:code_challenge;type:text;not null" json:"code_challenge"`
	CodeChallengeMethod string     `gorm:"column:code_challenge_method;type:text;not null" json:"code_challenge_method"`
	SsoSessionID        *uint64    `gorm:"column:sso_session_id" json:"sso_session_id,omitempty"`
	CreatedAt           time.Time  `gorm:"column:created_at;type:timestamptz;autoCreateTime" json:"created_at"`
	ExpiresAt           time.Time  `gorm:"column:expires_at;type:timestamptz;not null" json:"expires_at"`
	ConsumedAt          *time.Time `gorm:"column:consumed_at;type:timestamptz" json:"consumed_at,omitempty"`
}

func (SsoAuthorizationCode) TableName() string {
	return `"user-service".sso_authorization_codes`
}
