package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// AppConfig mirrors the common service configuration shape in chat-service.
type AppConfig struct {
	ServiceName string             `mapstructure:"serviceName"`
	Development bool               `mapstructure:"development"`
	Logger      *LoggerConfig      `mapstructure:"logger"`
	Http        *HttpConfig        `mapstructure:"http"`
	GRPC        *GrpcConfig        `mapstructure:"grpc"`
	Client      *ClientConfig      `mapstructure:"client"`
	Database    *DatabaseConfig    `mapstructure:"database"`
	Tracing     *TracingConfig     `mapstructure:"tracing"`
	Redis       *RedisConfig       `mapstructure:"redis"`
	Kafka       *KafkaConfig       `mapstructure:"kafka"`
	Healthcheck *HealthcheckConfig `mapstructure:"healthcheck"`
	Metrics     *MetricsConfig     `mapstructure:"metrics"`
	CronJob     *CronJob           `mapstructure:"cronjob"`
	Loki        *LokiConfig        `mapstructure:"loki"`
	SePay       SePayConfig        `mapstructure:"sepay"`
}

type LoggerConfig struct {
	LogLevel string `mapstructure:"level"`
	DevMode  bool   `mapstructure:"devMode"`
	Encoder  string `mapstructure:"encoder"`
}
type HttpConfig struct {
	Port            string   `mapstructure:"port"`
	Development     bool     `mapstructure:"development"`
	ShutdownTimeout int      `mapstructure:"shutdownTimeout"`
	AllowOrigins    []string `mapstructure:"allowOrigins"`
	Resources       []string `mapstructure:"resources"`
}
type GrpcConfig struct {
	Port              string `mapstructure:"port"`
	Development       bool   `mapstructure:"development"`
	MaxConnectionIdle int    `mapstructure:"maxConnectionIdle"`
	Timeout           int    `mapstructure:"timeout"`
	MaxConnectionAge  int    `mapstructure:"maxConnectionAge"`
	Time              int    `mapstructure:"time"`
}
type ClientConfig struct {
	UserService   string `mapstructure:"userService"`
	CommonService string `mapstructure:"commonService"`
}

type DatabaseConfig struct {
	ReadDbCfg  *ReadDbConfig  `mapstructure:"readDb"`
	WriteDbCfg *WriteDbConfig `mapstructure:"writeDb"`
}
type ReadDbConfig struct {
	DbType            string `mapstructure:"dbType"`
	ConnectionString  string `mapstructure:"connectionString"`
	MigrationFilePath string `mapstructure:"migrationFilePath"`
	MaxIdleConns      int    `mapstructure:"maxIdleConns"`
	MaxOpenConns      int    `mapstructure:"maxOpenConns"`
	ConnMaxLifetime   int    `mapstructure:"connMaxLifetime"`
}
type WriteDbConfig struct {
	DbType            string `mapstructure:"dbType"`
	ConnectionString  string `mapstructure:"connectionString"`
	MigrationFilePath string `mapstructure:"migrationFilePath"`
	MaxIdleConns      int    `mapstructure:"maxIdleConns"`
	MaxOpenConns      int    `mapstructure:"maxOpenConns"`
	ConnMaxLifetime   int    `mapstructure:"connMaxLifetime"`
}
type TracingConfig struct {
	ServiceName string `mapstructure:"serviceName"`
	HostPort    string `mapstructure:"hostPort"`
	Enable      bool   `mapstructure:"enable"`
	LogSpans    bool   `mapstructure:"logSpans"`
}
type RedisConfig struct {
	Addrs    []string `mapstructure:"addrs"`
	Password string   `mapstructure:"password"`
	PoolSize int      `mapstructure:"poolSize"`
	Username string   `mapstructure:"username"`
	DB       int      `mapstructure:"db"`
}
type KafkaConfig struct {
	Brokers []string          `mapstructure:"brokers"`
	GroupID string            `mapstructure:"groupID"`
	Topics  KafkaTopicsConfig `mapstructure:"topics"`
}
type KafkaTopicsConfig struct {
	UserEvents string `mapstructure:"userEvents"`
	ChatEvents string `mapstructure:"chatEvents"`
}
type HealthcheckConfig struct {
	Interval           int    `mapstructure:"interval"`
	Port               string `mapstructure:"port"`
	GoroutineThreshold int    `mapstructure:"goroutineThreshold"`
}
type MetricsConfig struct {
	PrometheusPath string `mapstructure:"prometheusPath"`
	PrometheusPort string `mapstructure:"prometheusPort"`
}
type CronJob struct {
	Disable bool `mapstructure:"disable"`
}
type LokiConfig struct {
	URL       string `mapstructure:"url"`
	Env       string `mapstructure:"env"`
	Service   string `mapstructure:"service"`
	BatchSize int    `mapstructure:"batchSize"`
	FlushMs   int    `mapstructure:"flushMs"`
	MaxQueue  int    `mapstructure:"maxQueue"`
	Retry     int    `mapstructure:"retry"`
	TimeoutMs int    `mapstructure:"timeoutMs"`
}

type SePayConfig struct {
	Environment             string   `mapstructure:"environment"`
	MerchantID              string   `mapstructure:"merchantId"`
	SecretKey               string   `mapstructure:"secretKey"`
	CheckoutURL             string   `mapstructure:"checkoutUrl"`
	SuccessURL              string   `mapstructure:"successUrl"`
	ErrorURL                string   `mapstructure:"errorUrl"`
	CancelURL               string   `mapstructure:"cancelUrl"`
	WebhookSecret           string   `mapstructure:"webhookSecret"`
	ReceivingAccountNumbers []string `mapstructure:"receivingAccountNumbers"`
	TimestampToleranceSec   int64    `mapstructure:"timestampToleranceSec"`
}

func Load(path string) (*AppConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvPrefix("PAYMENT_GATEWAY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	if cfg.Database != nil {
		if cfg.Database.ReadDbCfg != nil {
			cfg.Database.ReadDbCfg.ConnectionString = os.ExpandEnv(cfg.Database.ReadDbCfg.ConnectionString)
		}
		if cfg.Database.WriteDbCfg != nil {
			cfg.Database.WriteDbCfg.ConnectionString = os.ExpandEnv(cfg.Database.WriteDbCfg.ConnectionString)
		}
	}
	cfg.SePay.MerchantID = os.ExpandEnv(cfg.SePay.MerchantID)
	cfg.SePay.SecretKey = os.ExpandEnv(cfg.SePay.SecretKey)
	cfg.SePay.CheckoutURL = os.ExpandEnv(cfg.SePay.CheckoutURL)
	cfg.SePay.SuccessURL = os.ExpandEnv(cfg.SePay.SuccessURL)
	cfg.SePay.ErrorURL = os.ExpandEnv(cfg.SePay.ErrorURL)
	cfg.SePay.CancelURL = os.ExpandEnv(cfg.SePay.CancelURL)
	cfg.SePay.WebhookSecret = os.ExpandEnv(cfg.SePay.WebhookSecret)
	for i := range cfg.SePay.ReceivingAccountNumbers {
		cfg.SePay.ReceivingAccountNumbers[i] = os.ExpandEnv(cfg.SePay.ReceivingAccountNumbers[i])
	}
	if cfg.SePay.TimestampToleranceSec == 0 {
		cfg.SePay.TimestampToleranceSec = 300
	}
	if cfg.ServiceName == "" || cfg.Http == nil || cfg.Http.Port == "" || cfg.Database == nil || cfg.Database.ReadDbCfg == nil || cfg.Database.WriteDbCfg == nil || cfg.Database.ReadDbCfg.ConnectionString == "" || cfg.Database.WriteDbCfg.ConnectionString == "" {
		return nil, fmt.Errorf("serviceName, http.port, database.readDb and database.writeDb are required")
	}
	if cfg.SePay.Environment != "sandbox" {
		return nil, fmt.Errorf("only sepay sandbox is allowed in this phase")
	}
	return &cfg, nil
}
