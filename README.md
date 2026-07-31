# Courier System Documentation

# Run Database Migration

`migrate \
  -path migrations \
  -database "postgresql://dev:dev@127.0.0.1:5432/dev_db?sslmode=disable" \
  up
`

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
