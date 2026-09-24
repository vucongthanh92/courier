package constants

import "time"

// Time constants for caching and expiration
const (
	Time_Cache_5_minutes  = 5 * time.Minute
	Time_Cache_1_day      = 24 * time.Hour
	Time_Cache_5_seconds  = 5 * time.Second
	Time_Cache_15_minutes = 15 * time.Minute
	Time_Cache_30_seconds = 30 * time.Second
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

// constants for provider
const (
	GithubProvider       = "github"
	GithubAccessTokenURL = "https://github.com/login/oauth/access_token"

	GoogleProvider = "google"
	SePayProvider  = "sepay"
)

// ws channels
const (
	MessageCreatedChannel = "chat:events:message.created"
)

// ws configs
const (
	WsConfigWriteWait      = 10 * time.Second
	WsConfigPongWait       = 60 * time.Second
	WsConfigPingPeriod     = 45 * time.Second
	WsConfigSendBufferSize = 32
)

// Constants for Currency
const (
	CurrencyVND = "VND"
	CurrencyUSD = "USD"
)

// wallet status
const (
	WalletStatusActive   = "active"
	WalletStatusInactive = "inactive"
)
const ()

// Constants for Idempotency
const (
	IdempotencyScope = "wallet-topup"
)

// provider event status
const (
	SepayEventStatusReceived = "received"
)

// Constants for TopUpIntent status
const (
	TopupIntentStatusSucceeded = "succeeded"
	TopupIntentStatusFailed    = "failed"
	TopupIntentStatusPending   = "pending"
)

// Constants for ReferenceType
const (
	SePayRefTypeBankTransaction = "sepay_bank_transaction"
)

// Constants for TransferType
const (
	TransferTypeIn  = "in"
	TransferTypeOut = "out"
)

const (
	SePayPaymentCodePrefix = "COUR"
	SePayPaymentCodeLength = 8
)
