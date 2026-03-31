package config

// AppConfig holds the entire configuration for the application, including service name,
// This struct serves as a centralized configuration object that can be loaded from a configuration file (e.g., YAML) or environment variables,
// allowing for easy management and access to all application settings throughout the codebase.
type AppConfig struct {
	ServiceName    string                `mapstructure:"serviceName"`
	Development    bool                  `mapstructure:"development"`
	Logger         *LoggerConfig         `mapstructure:"logger"`
	Http           *HttpConfig           `mapstructure:"http"`
	GRPC           *GrpcConfig           `mapstructure:"grpc"`
	Database       *DatabaseConfig       `mapstructure:"database"`
	Tracing        *TracingConfig        `mapstructure:"tracing"`
	Kafka          *KafkaConfig          `mapstructure:"kafka"`
	Redis          *RedisConfig          `mapstructure:"redis"`
	Heathcheck     *HeathcheckConfig     `mapstructure:"heathcheck"`
	Metrics        *MetricsConfig        `mapstructure:"metrics"`
	KakaoMap       *KakaoMapConfig       `mapstructure:"kakaomap"`
	Authenticate   *Authenticate         `mapstructure:"authenticate"`
	Client         *GrpcClientConfig     `mapstructure:"client"`
	S3             *S3Config             `mapstructure:"s3"`
	CronJob        *CronJob              `mapstructure:"cronjob"`
	PaymentService *PaymentServiceConfig `mapstructure:"paymentService"`
	SlackService   *SlackConfig          `mapstructure:"slackService"`
	Email          *EmailConfig          `mapstructure:"email"`
	Loki           *LokiConfig           `mapstructure:"loki"`
	OAuth          *OAuthConfig          `mapstructure:"oauth"`
}

// GrpcClientConfig holds the configuration for gRPC clients that this service will call,
// including the addresses of the user service, driver service, and common service.
// This configuration is essential for setting up gRPC clients within the application,
// allowing it to communicate with other services in a microservices architecture.
type GrpcClientConfig struct {
	UserService   string `mapstructure:"userService"`
	DriverService string `mapstructure:"driverService"`
	CommonService string `mapstructure:"commonService"`
}

// LoggerConfig holds the configuration for the application's logging system,
// including log level, development mode, and log encoder settings.
type LoggerConfig struct {
	LogLevel string `mapstructure:"level"`
	DevMode  bool   `mapstructure:"devMode"`
	Encoder  string `mapstructure:"encoder"`
}

// HttpConfig holds the configuration for the HTTP server, including port,
// development mode, shutdown timeout, CORS settings such as allowed origins and resources.
type HttpConfig struct {
	Port            string   `mapstructure:"port"`
	Development     bool     `mapstructure:"development"`
	ShutdownTimeout int      `mapstructure:"shutdownTimeout"`
	AllowOrigins    []string `mapstructure:"allowOrigins"`
	Resources       []string `mapstructure:"resources"`
}

// GrpcConfig holds the configuration for gRPC server,
// including port, development mode, and connection settings such as max connection idle time, timeout, and max connection age.
// This configuration is essential for setting up and managing the gRPC server within the application,
// ensuring proper handling of client connections and efficient resource management.
type GrpcConfig struct {
	Port              string `mapstructure:"port"`
	Development       bool   `mapstructure:"development"`
	MaxConnectionIdle int    `mapstructure:"maxConnectionIdle"`
	Timeout           int    `mapstructure:"timeout"`
	MaxConnectionAge  int    `mapstructure:"maxConnectionAge"`
	Time              int    `mapstructure:"time"`
}

// DatabaseConfig holds the configuration for database connections,
// including separate configurations for read and write databases.
type DatabaseConfig struct {
	ReadDbCfg  *ReadDbConfig  `mapstructure:"readDb"`
	WriteDbCfg *WriteDbConfig `mapstructure:"writeDb"`
}

// ReadDbConfig holds the configuration for the read database connection,
type ReadDbConfig struct {
	DbType            string `mapstructure:"dbType"`
	ConnectionString  string `mapstructure:"connectionString"`
	MigrationFilePath string `mapstructure:"migrationFilePath"`
	MaxIdleConns      int    `mapstructure:"maxIdleConns"`
	MaxOpenConns      int    `mapstructure:"maxOpenConns"`
	ConnMaxLifetime   int    `mapstructure:"connMaxLifetime"`
}

// WriteDbConfig holds the configuration for the write database connection,
// including database type, connection string, migration file path, and connection pool settings.
type WriteDbConfig struct {
	DbType            string `mapstructure:"dbType"`
	ConnectionString  string `mapstructure:"connectionString"`
	MigrationFilePath string `mapstructure:"migrationFilePath"`
	MaxIdleConns      int    `mapstructure:"maxIdleConns"`
	MaxOpenConns      int    `mapstructure:"maxOpenConns"`
	ConnMaxLifetime   int    `mapstructure:"connMaxLifetime"`
}

// TracingConfig holds the configuration for distributed tracing, including service name, host and port for the tracing agent,
type TracingConfig struct {
	ServiceName string `mapstructure:"serviceName"`
	HostPort    string `mapstructure:"hostPort"`
	Enable      bool   `mapstructure:"enable"`
	LogSpans    bool   `mapstructure:"logSpans"`
}

// KafkaConfig holds the configuration for Kafka integration,
// including broker addresses, consumer group ID, topic settings, and authentication credentials.
type KafkaConfig struct {
	Config *KafkaConfigDetail `mapstructure:"config"`
	Topics *KafkaTopics       `mapstructure:"topics"`
	Dialer *DialerConfig      `mapstructure:"dialer"`
}

// DialerConfig holds the configuration for Kafka dialer, including authentication credentials such as username and password.
type DialerConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// KafkaConfigDetail holds the configuration details for connecting to a Kafka cluster, including broker addresses, consumer group ID,
// topic initialization settings, and the number of worker goroutines for processing Kafka messages.
// This configuration is essential for setting up Kafka producers and consumers within the application,
// enabling efficient message handling and processing in an event-driven architecture.
type KafkaConfigDetail struct {
	Brokers    []string `mapstructure:"brokers"`
	GroupID    string   `mapstructure:"groupID"`
	InitTopics bool     `mapstructure:"initTopics"`
	NumWorker  int      `mapstructure:"numWorker"`
}

// KafkaTopics holds the configuration for Kafka topics used in the application, including topic names and settings for each topic.
// This configuration allows the application to manage and interact with Kafka topics effectively,
// ensuring proper message handling and processing within the application's event-driven architecture.
type KafkaTopics struct {
	SubmitOrder KafkaTopicConfig `mapstructure:"submitOrder"`
}

// KafkaTopicConfig holds the configuration for a Kafka topic,
// including the topic name, number of partitions, and replication factor.
type KafkaTopicConfig struct {
	TopicName         string `mapstructure:"topicName"`
	NumPartitions     int    `mapstructure:"numPartitions"`
	ReplicationFactor int    `mapstructure:"replicationFactor"`
}

// RedisConfig holds the configuration for Redis connection, including addresses, authentication, and connection pool settings.
type RedisConfig struct {
	Addrs    []string `mapstructure:"addrs"`
	Password string   `mapstructure:"password"`
	PoolSize int      `mapstructure:"poolSize"`
	Username string   `mapstructure:"username"`
	DB       int      `mapstructure:"db"`
}

// HeathcheckConfig holds the configuration for health checks, including the interval for performing health checks,
type HeathcheckConfig struct {
	Interval           int    `mapstructure:"interval"`
	Port               string `mapstructure:"port"`
	GoroutineThreshold int    `mapstructure:"goroutineThreshold"`
}

// MetricsConfig holds the configuration for application metrics, such as Prometheus integration settings.
// This configuration allows the application to expose metrics for monitoring and observability purposes,
// enabling integration with tools like Prometheus to collect and analyze performance data.
type MetricsConfig struct {
	PrometheusPath string `mapstructure:"prometheusPath"`
	PrometheusPort string `mapstructure:"prometheusPort"`
}

// KakaoMapConfig holds the configuration for Kakao Map API integration, including API keys and endpoint URLs.
type KakaoMapConfig struct {
	AppRestKey          string `mapstructure:"appRestKey"`
	MobilityApiEndpoint string `mapstructure:"mobilityApiEndpoint"`
	Coord2regioncode    string `mapstructure:"coord2regioncode"`
	Priority            string `mapstructure:"priority"`
	TimeChange          string `mapstructure:"timeChange"`
	Address             string `mapstructure:"address"`
}

// Authenticate holds the configuration for authentication,
// such as client URL for email verification links or other auth-related endpoints.
type Authenticate struct {
	ClientURL string `mapstructure:"clientURI"`
}

// S3Config holds the configuration for S3 storage, including the path and reconciliation settings.
// This configuration is used for managing file storage and retrieval in the application, allowing for integration with S3-compatible storage services.
type S3Config struct {
	Path    string `mapstructure:"path"`
	Reconci string `mapstructure:"reconci"`
}

// CronJob holds the configuration for scheduled tasks or cron jobs in the application
type CronJob struct {
	Disable     bool   `mapstructure:"disable"`
	DispatchSms uint64 `mapstructure:"dispatchSms"`
	Stat        string `mapstructure:"stat"`
	FILELOC5    string `mapstructure:"FILELOC5"`
}

// Encryption holds the encryption configuration for the application
type Encryption struct {
	Salt string `mapstructure:"salt"`
}

// PaymentServiceConfig holds the configuration for the payment service integration
type PaymentServiceConfig struct {
	Url                string `mapstructure:"url"`
	MakePaymentUrl     string `mapstructure:"makePaymentUrl"`
	CheckPaymentStatus string `mapstructure:"checkPaymentStatus"`
	CancelPaymentUrl   string `mapstructure:"cancelPaymentUrl"`
}

// SlackConfig holds the configuration for Slack integration
type SlackConfig struct {
	UrlSlackWebhook string `mapstructure:"urlSlackWebhook"`
	Channel         string `mapstructure:"channel"`
	Username        string `mapstructure:"username"`
}

// EmailConfig holds the configuration for email sending
type EmailConfig struct {
	Enabled   bool        `mapstructure:"enabled"`
	Provider  string      `mapstructure:"provider"`
	From      string      `mapstructure:"from"`
	VerifyURL string      `mapstructure:"verifyUrl"`
	SMTP      *SMTPConfig `mapstructure:"smtp"`
}

// SMTPConfig holds the configuration for SMTP email provider
type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// LokiConfig holds the configuration for Loki logging, including URL, environment, service name, and batching settings.
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

// OAuthConfig holds the configuration for third-party OAuth providers,
// including client IDs and secrets for Google and GitHub.
type OAuthConfig struct {
	Google struct {
		ClientID string `yaml:"client_id"`
	} `yaml:"google"`
	Github struct {
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
		APIBase      string `yaml:"api_base"`
	} `yaml:"github"`
}
