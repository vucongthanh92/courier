package constants

import "time"

const (
	ServiceNameAgentGateway = "agent-gateway"
)

// Kafka topics for assistant events
const (
	AssistantRequestedTopic = "courier.assistant.requested.v1"
	AssistantRespondedTopic = "courier.assistant.responded.v1"
	AssistantRetryTopic     = "courier.assistant.events.retry.v1"
	AssistantDLQTopic       = "courier.assistant.events.dlq.v1"
)

// Event types for assistant events
const (
	AssistantRequestedEventType = "assistant.requested"
	AssistantRespondedEventType = "assistant.responded"
)

// Qdrant constants for memory storage
const (
	QdrantDefaultCollection = "courier_agent_memory"
	QdrantDefaultDistance   = "Cosine"
	QdrantDefaultVectorSize = 1536
	QdrantDefaultURL        = "http://localhost:6333"
)

// OpenAI constants for embedding and model selection
const (
	OpenAIEmbeddingModelTextEmbedding3Small = "text-embedding-3-small"
	OpenAIDefaultModel                      = "gpt-5.6"
)

// Memory roles for the assistant's memory system
const (
	MemoryRoleUser      = "user"
	MemoryRoleAssistant = "assistant"
	MemoryRoleSystem    = "system"
)

// HTTP paths for the service
const (
	HealthzPath               = "/healthz"
	AssistantInstructionsPath = "/v1/assistant/instructions"
)

// Default values for various configurations
const (
	DefaultHTTPPort              = ":5010"
	DefaultShutdownTimeout       = 10 * time.Second
	DefaultQdrantTimeout         = 5 * time.Second
	DefaultOpenAIRequestTimeout  = 45 * time.Second
	DefaultOpenAIMaxRecentTurns  = 20
	DefaultOpenAIMaxMemoryChunks = 8
	DefaultSummaryEveryMessages  = 20
	DefaultKafkaBroker           = "localhost:9092"
	DefaultReadHeaderTimeout     = 5 * time.Second
)

// AssistantSystemInstructions is the system instructions for the assistant. It is used to guide the assistant's behavior and responses.
const (
	AssistantSystemInstructions = `You are Courier's assistant. Help the user with clear, accurate, and useful answers. Use the provided conversation context, say when you are uncertain, and do not invent sources.`
)
