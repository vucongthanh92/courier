package config

type AppConfig struct {
	ServiceName string            `mapstructure:"serviceName"`
	Development bool              `mapstructure:"development"`
	Logger      *LoggerConfig     `mapstructure:"logger"`
	Http        *HttpConfig       `mapstructure:"http"`
	GRPC        *GrpcConfig       `mapstructure:"grpc"`
	Database    *DatabaseConfig   `mapstructure:"database"`
	Tracing     *TracingConfig    `mapstructure:"tracing"`
	Redis       *RedisConfig      `mapstructure:"redis"`
	Heathcheck  *HeathcheckConfig `mapstructure:"heathcheck"`
	Metrics     *MetricsConfig    `mapstructure:"metrics"`
	CronJob     *CronJob          `mapstructure:"cronjob"`
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

type HeathcheckConfig struct {
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
