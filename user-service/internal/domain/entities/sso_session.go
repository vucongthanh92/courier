package entities

import "time"

type SsoSession struct {
	ID          uint64     `gorm:"column:id;primaryKey;type:bigint;check:id>0" json:"id"`
	UserID      uint64     `gorm:"column:user_id;index;not null" json:"user_id"`
	SessionHash string     `gorm:"column:session_hash;type:text;uniqueIndex;not null" json:"-"`
	UserAgent   *string    `gorm:"column:user_agent;type:text" json:"user_agent,omitempty"`
	IP          *string    `gorm:"column:ip;type:inet" json:"ip,omitempty"`
	AuthTime    time.Time  `gorm:"column:auth_time;type:timestamptz;not null" json:"auth_time"`
	CreatedAt   time.Time  `gorm:"column:created_at;type:timestamptz;autoCreateTime" json:"created_at"`
	ExpiresAt   time.Time  `gorm:"column:expires_at;type:timestamptz;not null" json:"expires_at"`
	RevokedAt   *time.Time `gorm:"column:revoked_at;type:timestamptz" json:"revoked_at,omitempty"`
}

func (SsoSession) TableName() string {
	return `"user-service".sso_sessions`
}

func (e *SsoSession) Initialize(sessionID uint64, sessionToken string, UserAgent *string, ip *string, sessionTTLMinutes int, userID uint64) {
	now := time.Now()
	sessionTTL := time.Duration(sessionTTLMinutes) * time.Minute

	e.ID = sessionID
	e.UserID = userID
	e.UserAgent = UserAgent
	e.IP = ip
	e.SessionHash = sessionToken
	e.AuthTime = now.UTC()
	e.ExpiresAt = now.Add(sessionTTL)
}
