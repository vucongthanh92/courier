# Courier System Documentation

# Run Database Migration

`migrate \
  -path shared/migrations \
  -database "postgresql://dev:dev@127.0.0.1:5432/dev_db?sslmode=disable" \
  up
`

All database migrations now live in `shared/migrations` because local Courier services share the same PostgreSQL database and `schema_migrations` table. Do not run service-local migration folders for `user-service` or `chat-service`.

# Event-Driven System Conversations

Courier uses Kafka as the primary event bus for cross-service integration events.

Local Kafka stack:

```bash
docker compose -f event-bus/kafka/docker-compose.yaml up -d
```

Kafka endpoints:

- Kafka broker: `localhost:9092`
- Kafka UI: `http://localhost:8081`
- Topic definitions: `event-bus/kafka/topics.json`

Topic initialization is handled by the one-shot `kafka-init` service. When adding a topic, update `event-bus/kafka/topics.json`, then run:

```bash
docker compose -f event-bus/kafka/docker-compose.yaml up kafka-init
```

## Verified User Flow

When a user verifies their email successfully in `user-service`, the service writes a transactional outbox record with event type `user.email_verified.v1`. The outbox worker publishes this event to Kafka topic `courier.user.events.v1`.

`chat-service` runs a Kafka consumer in consumer group `chat-service`. After consuming `user.email_verified.v1`, it idempotently provisions two system conversations for the verified user:

- `notification`
- `assistant`

Each system conversation:

- has `type = system`
- has `created_by = user_id`
- has one active member, the user
- has a deterministic `direct_key` based on `system:{user_id}:{name}`
- uses Snowflake-format IDs matching `chat-service` ID generation

Backend-generated system messages use `sender_id = 0`. Database triggers allow `sender_id = 0` only in `system` conversations. Normal user messages still require the sender to be an active conversation member.

Existing verified users are backfilled by migration `016_chat_service_core_and_system_conversations` with empty `notification` and `assistant` conversations.

# Agent Gateway And Qdrant Memory

Courier uses `agent-gateway` as the service boundary for AI assistant work. The first implementation keeps two responsibilities in this service:

- build and send conversation context to the AI provider
- manage assistant memory and vector data in Qdrant

`chat-service` remains the source of truth for chat messages. When a user sends a message to the `system` conversation named `assistant`, `chat-service` publishes an assistant request event to Kafka. `agent-gateway` prepares context, uses Qdrant for memory lookup/storage, calls the AI provider, then publishes an assistant response event back to `chat-service`. `chat-service` inserts the assistant reply into its database, so `conversa-app` receives the reply through the existing WebSocket flow.

Assistant Kafka topics:

- Request topic: `courier.assistant.requested.v1`
- Response topic: `courier.assistant.responded.v1`
- Retry topic: `courier.assistant.events.retry.v1`
- Dead-letter topic: `courier.assistant.events.dlq.v1`

Run local Qdrant:

```bash
cd agent-gateway
make qdrant-up
```

Qdrant local endpoints:

- REST API: `http://localhost:6333`
- Dashboard: `http://localhost:6333/dashboard`
- Readiness: `http://localhost:6333/readyz`
- gRPC: `localhost:6334`

Run `agent-gateway` against local Qdrant:

```bash
cd agent-gateway
make run-local
curl http://localhost:5010/healthz
```

Run Qdrant and `agent-gateway` together:

```bash
cd agent-gateway
docker compose up --build
```

For full assistant flow testing, start Kafka and initialize all topics from the repository root:

```bash
docker compose -f event-bus/kafka/docker-compose.yaml up -d
docker compose -f event-bus/kafka/docker-compose.yaml up kafka-init
```

Set `OPENAI_API_KEY` before running `agent-gateway` when testing real AI responses.

`agent-gateway` follows the same config folder style as the other Courier services:

- `agent-gateway/config/local/config.yaml`
- `agent-gateway/config/dev/config.yaml`
- `agent-gateway/config/prd/config.yaml`

Important config defaults:

- Qdrant collection: `courier_agent_memory`
- Qdrant vector size: `1536`
- Qdrant distance: `Cosine`
- Generation model target: `gpt-5.5`
- Embedding model: `text-embedding-3-small`

Before production use, verify the exact OpenAI API model id with the provided API key. OpenAI API model availability is account/project dependent and can be checked through the Models API.

# Case Study: WebSocket Realtime Delivery With Redis Pub/Sub

This pattern was introduced for `chat-service` and `conversa-app` so active conversation members can receive newly created messages without reloading or polling REST APIs.

## When To Use

Use this structure when a Courier service needs realtime client delivery and may run as multiple service instances:

- chat messages
- user/session presence
- typing indicators
- notification counters
- account/security events that should appear in active clients

## Architecture Pattern

```text
Client
  opens 1 WebSocket connection per user session/device/tab

HTTP API
  authenticates normal REST requests
  persists business state
  publishes a domain event after successful write

Redis Pub/Sub
  fans out events across all service instances

api/ws hub
  keeps only local active WebSocket connections
  subscribes to Redis events
  routes events to connected target users

REST APIs
  remain the source of truth for initial load, pagination, and reconnect recovery
```

## Backend Shape

Recommended package layout:

```text
<service>/internal/api/http
<service>/internal/api/grpc
<service>/internal/api/ws
<service>/internal/domain/interfaces/ws.go
<service>/internal/domain/models/event.go
<service>/internal/repository/external/redis/ws_pubsub.go
```

Keep responsibilities separated:

- `api/ws`: WebSocket hub, connection registry, ping/pong, local delivery.
- `api/http`: REST routes plus a small WebSocket upgrade handler if the service uses Gin routing.
- `domain/interfaces`: `WsPublisherI` and `WsSubscriberI`.
- `domain/models`: stable realtime event envelopes.
- `repository/external/redis`: Redis Pub/Sub publish/subscribe implementation.
- usecase layer: publish the event only after the business write succeeds.

### Naming And Route Convention

Use `Ws` for names owned by the WebSocket transport or its delivery infrastructure:

- `WsHandler`, `WsPublisherI`, `WsSubscriberI`
- `InitWsHandler`, `InitWsPublisher`, `InitWsSubscriber`
- `ws.go`, `ws_pubsub.go`, `wsHandler`, `wsHub`

Keep domain event names transport-independent, such as `MessageCreatedEvent`, because the same event may later be consumed by WebSocket, push notification, or another adapter.

For Gin routes, separate the shared API origin from authentication groups:

```go
origin := router.Group("/api")

v1NonAuth := origin.Group("/v1")
v1NonAuth.GET("/ws", wsHandler.VerifyAndConnect)

v1HasAuth := origin.Group("/v1")
v1HasAuth.Use(authMiddleWare)
```

Use the same naming and route grouping when WebSocket support is added to another Courier service, including `user-service`.

## Client Shape

Recommended client behavior:

- Open one WebSocket connection after login.
- Pass the access token during connection, for browser clients commonly as `?access_token=...`.
- Render small/full realtime payloads immediately when the event is enough for UI.
- Deduplicate by stable entity ID, such as `message.id`.
- Use REST APIs for initial loading and reconnect recovery.
- Reconnect with backoff when the socket closes.

## Chat-Service Implementation Notes

`chat-service` uses this pattern for `message.created`:

```text
POST /api/v1/conversation/:id/messages/create
  -> validate sender and active membership
  -> persist message
  -> invalidate latest-message cache
  -> publish Redis event: chat:events:message.created
  -> every instance receives event
  -> api/ws hub sends event to active member connections
```

The WebSocket endpoint is:

```text
GET /api/v1/ws?access_token=<JWT>
```

The event payload includes the full message for MVP rendering:

```json
{
  "type": "message.created",
  "conversation_id": "10",
  "recipient_user_ids": [20, 21],
  "message": {
    "id": "100",
    "conversation_id": "10",
    "sender_id": "20",
    "type": "text",
    "body": "hello",
    "metadata": {},
    "created_at": "2026-07-30T10:00:00Z",
    "updated_at": "2026-07-30T10:00:00Z"
  },
  "event_at": "2026-07-30T10:00:00Z"
}
```

For future heavy payloads such as attachments, prefer sending metadata or previews over WebSocket and use REST sync for the full payload.
