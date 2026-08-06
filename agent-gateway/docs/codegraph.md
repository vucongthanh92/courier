# Agent Gateway Codegraph

This document explains how requests and future assistant events flow through `agent-gateway`.

## Current Runtime Flow

The current implementation is a service skeleton that starts an HTTP server, checks Qdrant readiness, and ensures the memory collection exists.

```mermaid
flowchart LR
    Main["main.go"] --> Config["config.Load"]
    Config --> YAML["config/<env>/config.yaml"]
    Config --> Env["env overrides"]
    Main --> QdrantClient["repository/qdrant.Client"]
    Main --> GatewayService["internal/gateway.Service"]
    GatewayService --> Bootstrap["Bootstrap"]
    Bootstrap --> QdrantReady["Qdrant /readyz"]
    Bootstrap --> EnsureCollection["Ensure Qdrant collection"]
    Main --> HTTPServer["net/http server"]
    HTTPServer --> Handler["internal/gateway.HTTPHandler"]
    Handler --> Health["GET /healthz"]
    Health --> GatewayService
    GatewayService --> QdrantReady
```

## Package Responsibilities

```mermaid
flowchart TB
    ConfigPkg["config"] --> ConfigFiles["config/local, dev, prd"]
    HelperConstants["helper/constants"] --> Defaults["service names, topics, defaults, paths"]
    HelperUtils["helper/utils"] --> SmallHelpers["correlation IDs, timeout contexts"]
    GatewayPkg["internal/gateway"] --> Usecase["assistant gateway orchestration"]
    QdrantRepo["internal/repository/qdrant"] --> Qdrant["Qdrant REST API"]
    Models["internal/domain/models"] --> Contracts["event and memory DTOs"]

    GatewayPkg --> HelperUtils
    GatewayPkg --> HelperConstants
    GatewayPkg --> QdrantRepo
    GatewayPkg --> Models
    ConfigPkg --> HelperConstants
```

## Target Assistant Event Flow

This is the agreed feature flow once Kafka and OpenAI provider integration are added.

```mermaid
sequenceDiagram
    participant User
    participant App as conversa-app
    participant Chat as chat-service
    participant Kafka as Kafka
    participant Agent as agent-gateway
    participant Qdrant
    participant AI as AI Provider

    User->>App: Send message to assistant conversation
    App->>Chat: POST create message
    Chat->>Chat: Persist user message
    Chat->>Chat: Detect type=system and name=assistant
    Chat->>Kafka: Publish courier.assistant.requested.v1
    Kafka->>Agent: Consume assistant request
    Agent->>Qdrant: Upsert user message memory
    Agent->>Qdrant: Search relevant memories
    Agent->>Agent: Build context package
    Agent->>AI: Send context and user request
    AI-->>Agent: Assistant answer
    Agent->>Qdrant: Upsert assistant answer memory
    Agent->>Kafka: Publish courier.assistant.responded.v1
    Kafka->>Chat: Consume assistant response
    Chat->>Chat: Insert assistant/system message
    Chat-->>App: Existing websocket message.created event
```

## Target Internal Layer Flow

When the assistant request consumer is implemented, the internal flow should look like this:

```mermaid
flowchart LR
    KafkaConsumer["Kafka consumer"] --> EventModel["AssistantRequestedPayload"]
    EventModel --> GatewayService["gateway.Service"]
    GatewayService --> Idempotency["idempotency check"]
    GatewayService --> EmbeddingProvider["embedding provider"]
    EmbeddingProvider --> UserVector["user message vector"]
    GatewayService --> QdrantUpsert["Qdrant upsert memory"]
    GatewayService --> QdrantSearch["Qdrant search relevant memories"]
    GatewayService --> ContextBuilder["context builder"]
    ContextBuilder --> SystemPrompt["system instructions"]
    ContextBuilder --> RecentMessages["recent messages"]
    ContextBuilder --> RelevantMemory["relevant Qdrant memory"]
    ContextBuilder --> CurrentMessage["current user message"]
    ContextBuilder --> AIProvider["AI provider client"]
    AIProvider --> ResponseModel["AssistantRespondedPayload"]
    ResponseModel --> KafkaProducer["Kafka producer"]
```

## Context Package Shape

`agent-gateway` should avoid sending the entire conversation history to the AI provider. It should send a compact context package:

```mermaid
flowchart TB
    ContextPackage["ContextPackage"] --> Instructions["system instructions"]
    ContextPackage --> Summary["conversation summary"]
    ContextPackage --> Recent["recent messages"]
    ContextPackage --> Relevant["relevant Qdrant memories"]
    ContextPackage --> Current["current user message"]
```

The current DTO lives in `internal/domain/models.ContextPackage`.

## Ownership Boundaries

- `chat-service` owns canonical chat message storage and websocket fanout.
- `agent-gateway` owns assistant context preparation, vector memory, provider calls, and assistant response events.
- Qdrant stores AI memory vectors and metadata, not canonical chat messages.
- Kafka is the boundary between `chat-service` and `agent-gateway`.

— Nội dung được viết/cập nhật bởi AI (OpenAI Codex).
