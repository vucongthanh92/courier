package constants

import "time"

const (
	Time_Cache_5_minutes = 5 * time.Minute
	Time_Cache_1_day     = 24 * time.Hour
	Time_Cache_5_seconds = 5 * time.Second
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

// OAuth2 provider constants
const (
	GithubProvider       = "github"
	GithubAccessTokenURL = "https://github.com/login/oauth/access_token"

	GoogleProvider = "google"
)

// JWK cache key prefix
const (
	JwkCacheKeyPrefix = "auth:jwk:kid:"
	DenylistKeyPrefix = "deny:jti:"
)
