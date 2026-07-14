package constants

import "time"

const (
	Time_Cache_5_minutes  = 5 * time.Minute
	Time_Cache_1_day      = 24 * time.Hour
	Time_Cache_5_seconds  = 5 * time.Second
	Time_Cache_15_minutes = 15 * time.Minute
)

const (
	InvalidValue       = "InvalidValue"
	InvalidLength      = "InvalidLength"
	InvalidEmailFormat = "InvalidEmailFormat"
)

const (
	Yaml               = "yaml"
	Gzip               = "gzip"
	Redis              = "redis"
	ReadDatabase       = "read-database"
	WriteDatabase      = "write-database"
	GoroutineThreshold = "goroutine-threshold"
	Kafka              = "kafka"
)

const (
	GithubProvider       = "github"
	GithubAccessTokenURL = "https://github.com/login/oauth/access_token"

	GoogleProvider = "google"
)

const (
	ConversationTypeDirect = "direct"
	ConversationTypeGroup  = "group"
)

const (
	ConversationMemberRoleOwner  = "owner"
	ConversationMemberRoleAdmin  = "admin"
	ConversationMemberRoleMember = "member"
)

const (
	ConversationMemberStatusActive  = "active"
	ConversationMemberStatusLeft    = "left"
	ConversationMemberStatusRemoved = "removed"
)

const (
	MessageTypeText   = "text"
	MessageTypeSystem = "system"
)
