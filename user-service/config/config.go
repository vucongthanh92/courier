package config

import (
	"github.com/vucongthanh92/go-base-utils/models"
)

// AppConfig holds the entire configuration for the application, including service name,
// This struct serves as a centralized configuration object that can be loaded from a configuration file (e.g., YAML) or environment variables,
// allowing for easy management and access to all application settings throughout the codebase.
type AppConfig struct {
	ServiceName    string                       `mapstructure:"serviceName"`
	Development    bool                         `mapstructure:"development"`
	Logger         *models.LoggerConfig         `mapstructure:"logger"`
	Http           *models.HttpConfig           `mapstructure:"http"`
	GRPC           *models.GrpcConfig           `mapstructure:"grpc"`
	Database       *models.DatabaseConfig       `mapstructure:"database"`
	Tracing        *models.TracingConfig        `mapstructure:"tracing"`
	Kafka          *models.KafkaConfig          `mapstructure:"kafka"`
	Redis          *models.RedisConfig          `mapstructure:"redis"`
	Heathcheck     *models.HeathcheckConfig     `mapstructure:"heathcheck"`
	Metrics        *models.MetricsConfig        `mapstructure:"metrics"`
	KakaoMap       *models.KakaoMapConfig       `mapstructure:"kakaomap"`
	Authenticate   *models.Authenticate         `mapstructure:"authenticate"`
	Client         *models.GrpcClientConfig     `mapstructure:"client"`
	S3             *models.S3Config             `mapstructure:"s3"`
	CronJob        *models.CronJob              `mapstructure:"cronjob"`
	PaymentService *models.PaymentServiceConfig `mapstructure:"paymentService"`
	SlackService   *models.SlackConfig          `mapstructure:"slackService"`
	Email          *models.EmailConfig          `mapstructure:"email"`
	Loki           *models.LokiConfig           `mapstructure:"loki"`
	OAuth          *models.OAuthConfig          `mapstructure:"oauth"`
	SSO            *models.SSOConfig            `mapstructure:"sso"`
}

// Please refer to the configuration models in `base-utils` before creating a new configuration.
// If you need to add a new configuration, please create a new model in `base-utils` and then reference it here.
