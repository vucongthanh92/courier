package config

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
	DaService      *DaService            `mapstructure:"daservice"`
	CronJob        *CronJob              `mapstructure:"cronjob"`
	FeatureFlag    *FeatureFlag          `mapstructure:"featureFlag"`
	PaymentService *PaymentServiceConfig `mapstructure:"paymentService"`
	SlackService   *SlackConfig          `mapstructure:"slackService"`
}

type GrpcClientConfig struct {
	UserService   string `mapstructure:"userService"`
	DriverService string `mapstructure:"driverService"`
	CommonService string `mapstructure:"commonService"`
}

type LoggerConfig struct {
	LogLevel string `mapstructure:"level"`
	DevMode  bool   `mapstructure:"devMode"`
	Encoder  string `mapstructure:"encoder"`
}

type HttpConfig struct {
	Port            string `mapstructure:"port"`
	Development     bool   `mapstructure:"development"`
	ShutdownTimeout int    `mapstructure:"shutdownTimeout"`

	AllowOrigins []string `mapstructure:"allowOrigins"`
	Resources    []string `mapstructure:"resources"`
}

type GrpcConfig struct {
	Port              string `mapstructure:"port"`
	Development       bool   `mapstructure:"development"`
	MaxConnectionIdle int    `mapstructure:"maxConnectionIdle"`
	Timeout           int    `mapstructure:"timeout"`
	MaxConnectionAge  int    `mapstructure:"maxConnectionAge"`
	Time              int    `mapstructure:"time"`
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

type KafkaConfig struct {
	Config *KafkaConfigDetail `mapstructure:"config"`
	Topics *KafkaTopics       `mapstructure:"topics"`
	Dialer *DialerConfig      `mapstructure:"dialer"`
}

type DialerConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type KafkaConfigDetail struct {
	Brokers    []string `mapstructure:"brokers"`
	GroupID    string   `mapstructure:"groupID"`
	InitTopics bool     `mapstructure:"initTopics"`
	NumWorker  int      `mapstructure:"numWorker"`
}

type KafkaTopics struct {
	SubmitOrder KafkaTopicConfig `mapstructure:"submitOrder"`
}

type KafkaTopicConfig struct {
	TopicName         string `mapstructure:"topicName"`
	NumPartitions     int    `mapstructure:"numPartitions"`
	ReplicationFactor int    `mapstructure:"replicationFactor"`
}

type RedisConfig struct {
	Addrs    []string `mapstructure:"addrs"`
	Password string   `mapstructure:"password"`
	PoolSize int      `mapstructure:"poolSize"`
	Username string   `mapstructure:"username"`
	DB       int      `mapstructure:"db"`
}

type HeathcheckConfig struct {
	Interval           int    `mapstructure:"interval"`
	Port               string `mapstructure:"port"`
	GoroutineThreshold int    `mapstructure:"goroutineThreshold"`
}

type MetricsConfig struct {
	PrometheusPath string `mapstructure:"prometheusPath"`
	PrometheusPort string `mapstructure:"prometheusPort"`
}

type KakaoMapConfig struct {
	AppRestKey          string `mapstructure:"appRestKey"`
	MobilityApiEndpoint string `mapstructure:"mobilityApiEndpoint"`
	Coord2regioncode    string `mapstructure:"coord2regioncode"`
	Priority            string `mapstructure:"priority"`
	TimeChange          string `mapstructure:"timeChange"`
	Address             string `mapstructure:"address"`
}

type Authenticate struct {
	ClientURL string `mapstructure:"clientURI"`
}

type S3Config struct {
	Path    string `mapstructure:"path"`
	Reconci string `mapstructure:"reconci"`
}

type DaService struct {
	WebApi                  string `mapstructure:"webApi"`
	Notification            string `mapstructure:"notification"`
	Webhook                 string `mapstructure:"webhook"`
	UpdateNotification      string `mapstructure:"updateNotification"`
	CancelOrderNotification string `mapstructure:"cancelOrderNotification"`
	SubmitOrderNotification string `mapstructure:"submitOrderNotification"`
	GetListCoupons          string `mapstructure:"getListCoupons"`
	SubmitTipOrderAssigned  string `mapstructure:"submitTipOrderAssigned"`
}

type CronJob struct {
	Disable     bool   `mapstructure:"disable"`
	DispatchSms uint64 `mapstructure:"dispatchSms"`
	Stat        string `mapstructure:"stat"`
	FILELOC5    string `mapstructure:"FILELOC5"`
}

type Encryption struct {
	Salt string `mapstructure:"salt"`
}

type FeatureFlag struct {
	Payment string `mapstructure:"payment"`
}

type PaymentServiceConfig struct {
	Url                string `mapstructure:"url"`
	MakePaymentUrl     string `mapstructure:"makePaymentUrl"`
	CheckPaymentStatus string `mapstructure:"checkPaymentStatus"`
	CancelPaymentUrl   string `mapstructure:"cancelPaymentUrl"`
}

type SlackConfig struct {
	UrlSlackWebhook string `mapstructure:"urlSlackWebhook"`
	Channel         string `mapstructure:"channel"`
	Username        string `mapstructure:"username"`
}
