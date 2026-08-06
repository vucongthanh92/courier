package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vucongthanh92/courier/agent-gateway/helper/constants"
)

type AppConfig struct {
	ServiceName string
	Development bool
	HTTP        HTTPConfig
	Qdrant      QdrantConfig
	OpenAI      OpenAIConfig
	Kafka       KafkaConfig
	Memory      MemoryConfig
}

type HTTPConfig struct {
	Port            string
	Development     bool
	ShutdownTimeout time.Duration
}

type QdrantConfig struct {
	URL            string
	APIKey         string
	CollectionName string
	VectorSize     uint64
	Distance       string
	Timeout        time.Duration
}

type OpenAIConfig struct {
	APIKey          string
	Model           string
	EmbeddingModel  string
	RequestTimeout  time.Duration
	MaxRecentTurns  int
	MaxMemoryChunks int
}

type KafkaConfig struct {
	Brokers                 []string
	GroupID                 string
	AssistantRequestedTopic string
	AssistantRespondedTopic string
}

type MemoryConfig struct {
	SummaryEveryMessages int
}

func Load(configPath string) (AppConfig, error) {
	cfg := defaultConfig()
	if configPath == "" {
		configPath = getEnv("CONFIG_FILE", "./config/local/config.yaml")
	}
	if err := loadYAML(configPath, &cfg); err != nil {
		return AppConfig{}, err
	}
	applyEnvOverrides(&cfg)
	return cfg, nil
}

func defaultConfig() AppConfig {
	return AppConfig{
		ServiceName: constants.ServiceNameAgentGateway,
		Development: true,
		HTTP: HTTPConfig{
			Port:            constants.DefaultHTTPPort,
			Development:     true,
			ShutdownTimeout: constants.DefaultShutdownTimeout,
		},
		Qdrant: QdrantConfig{
			URL:            constants.QdrantDefaultURL,
			CollectionName: constants.QdrantDefaultCollection,
			VectorSize:     constants.QdrantDefaultVectorSize,
			Distance:       constants.QdrantDefaultDistance,
			Timeout:        constants.DefaultQdrantTimeout,
		},
		OpenAI: OpenAIConfig{
			Model:           constants.OpenAIDefaultModel,
			EmbeddingModel:  constants.OpenAIEmbeddingModelTextEmbedding3Small,
			RequestTimeout:  constants.DefaultOpenAIRequestTimeout,
			MaxRecentTurns:  constants.DefaultOpenAIMaxRecentTurns,
			MaxMemoryChunks: constants.DefaultOpenAIMaxMemoryChunks,
		},
		Kafka: KafkaConfig{
			Brokers:                 []string{constants.DefaultKafkaBroker},
			GroupID:                 constants.ServiceNameAgentGateway,
			AssistantRequestedTopic: constants.AssistantRequestedTopic,
			AssistantRespondedTopic: constants.AssistantRespondedTopic,
		},
		Memory: MemoryConfig{
			SummaryEveryMessages: constants.DefaultSummaryEveryMessages,
		},
	}
}

func loadYAML(path string, cfg *AppConfig) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open config file %s: %w", path, err)
	}
	defer file.Close()

	values := map[string]string{}
	var section string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		raw := stripComment(scanner.Text())
		if strings.TrimSpace(raw) == "" {
			continue
		}
		indent := len(raw) - len(strings.TrimLeft(raw, " "))
		line := strings.TrimSpace(raw)
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if indent == 0 && value == "" {
			section = key
			continue
		}
		fullKey := key
		if indent > 0 && section != "" {
			fullKey = section + "." + key
		}
		values[fullKey] = trimValue(value)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read config file %s: %w", path, err)
	}

	applyYAMLValues(values, cfg)
	return nil
}

func applyYAMLValues(values map[string]string, cfg *AppConfig) {
	cfg.ServiceName = stringValue(values, "serviceName", cfg.ServiceName)
	cfg.Development = boolValue(values, "development", cfg.Development)
	cfg.HTTP.Port = stringValue(values, "http.port", cfg.HTTP.Port)
	cfg.HTTP.Development = boolValue(values, "http.development", cfg.HTTP.Development)
	cfg.HTTP.ShutdownTimeout = secondsValue(values, "http.shutdownTimeout", cfg.HTTP.ShutdownTimeout)

	cfg.Qdrant.URL = stringValue(values, "qdrant.url", cfg.Qdrant.URL)
	cfg.Qdrant.APIKey = stringValue(values, "qdrant.apiKey", cfg.Qdrant.APIKey)
	cfg.Qdrant.CollectionName = stringValue(values, "qdrant.collectionName", cfg.Qdrant.CollectionName)
	cfg.Qdrant.VectorSize = uint64Value(values, "qdrant.vectorSize", cfg.Qdrant.VectorSize)
	cfg.Qdrant.Distance = stringValue(values, "qdrant.distance", cfg.Qdrant.Distance)
	cfg.Qdrant.Timeout = durationValue(values, "qdrant.timeout", cfg.Qdrant.Timeout)

	cfg.OpenAI.APIKey = stringValue(values, "openai.apiKey", cfg.OpenAI.APIKey)
	cfg.OpenAI.Model = stringValue(values, "openai.model", cfg.OpenAI.Model)
	cfg.OpenAI.EmbeddingModel = stringValue(values, "openai.embeddingModel", cfg.OpenAI.EmbeddingModel)
	cfg.OpenAI.RequestTimeout = durationValue(values, "openai.requestTimeout", cfg.OpenAI.RequestTimeout)
	cfg.OpenAI.MaxRecentTurns = intValue(values, "openai.maxRecentTurns", cfg.OpenAI.MaxRecentTurns)
	cfg.OpenAI.MaxMemoryChunks = intValue(values, "openai.maxMemoryChunks", cfg.OpenAI.MaxMemoryChunks)

	cfg.Kafka.Brokers = stringSliceValue(values, "kafka.brokers", cfg.Kafka.Brokers)
	cfg.Kafka.GroupID = stringValue(values, "kafka.groupID", cfg.Kafka.GroupID)
	cfg.Kafka.AssistantRequestedTopic = stringValue(values, "kafka.assistantRequestedTopic", cfg.Kafka.AssistantRequestedTopic)
	cfg.Kafka.AssistantRespondedTopic = stringValue(values, "kafka.assistantRespondedTopic", cfg.Kafka.AssistantRespondedTopic)

	cfg.Memory.SummaryEveryMessages = intValue(values, "memory.summaryEveryMessages", cfg.Memory.SummaryEveryMessages)
}

func applyEnvOverrides(cfg *AppConfig) {
	cfg.ServiceName = getEnv("SERVICE_NAME", cfg.ServiceName)
	cfg.HTTP.Port = getEnv("HTTP_PORT", cfg.HTTP.Port)
	cfg.Qdrant.CollectionName = getEnv("QDRANT_COLLECTION", cfg.Qdrant.CollectionName)
	cfg.Qdrant.Distance = getEnv("QDRANT_DISTANCE", cfg.Qdrant.Distance)
	cfg.Qdrant.VectorSize = getUint64Env("QDRANT_VECTOR_SIZE", cfg.Qdrant.VectorSize)
	cfg.Qdrant.URL = getEnv("QDRANT_URL", cfg.Qdrant.URL)
	cfg.Qdrant.APIKey = getEnv("QDRANT_API_KEY", cfg.Qdrant.APIKey)
	cfg.OpenAI.APIKey = getEnv("OPENAI_API_KEY", cfg.OpenAI.APIKey)
	cfg.OpenAI.Model = getEnv("OPENAI_MODEL", cfg.OpenAI.Model)
	cfg.OpenAI.EmbeddingModel = getEnv("OPENAI_EMBEDDING_MODEL", cfg.OpenAI.EmbeddingModel)
	cfg.Kafka.Brokers = getCSVEnv("KAFKA_BROKERS", cfg.Kafka.Brokers)
	cfg.Kafka.GroupID = getEnv("KAFKA_GROUP_ID", cfg.Kafka.GroupID)
	cfg.Kafka.AssistantRequestedTopic = getEnv("KAFKA_ASSISTANT_REQUESTED_TOPIC", cfg.Kafka.AssistantRequestedTopic)
	cfg.Kafka.AssistantRespondedTopic = getEnv("KAFKA_ASSISTANT_RESPONDED_TOPIC", cfg.Kafka.AssistantRespondedTopic)
}

func stripComment(line string) string {
	if index := strings.Index(line, "#"); index >= 0 {
		return line[:index]
	}
	return line
}

func trimValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)
	return strings.Trim(value, `'`)
}

func stringValue(values map[string]string, key string, fallback string) string {
	if value := strings.TrimSpace(values[key]); value != "" {
		return value
	}
	return fallback
}

func boolValue(values map[string]string, key string, fallback bool) bool {
	value := strings.TrimSpace(values[key])
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func intValue(values map[string]string, key string, fallback int) int {
	value := strings.TrimSpace(values[key])
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func uint64Value(values map[string]string, key string, fallback uint64) uint64 {
	value := strings.TrimSpace(values[key])
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func secondsValue(values map[string]string, key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(values[key])
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return time.Duration(parsed) * time.Second
}

func durationValue(values map[string]string, key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(values[key])
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err == nil {
		return parsed
	}
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}

func stringSliceValue(values map[string]string, key string, fallback []string) []string {
	value := strings.TrimSpace(values[key])
	if value == "" {
		return fallback
	}
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if item := trimValue(part); item != "" {
			result = append(result, item)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getCSVEnv(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if item := strings.TrimSpace(part); item != "" {
			result = append(result, item)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

func getUint64Env(key string, fallback uint64) uint64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
