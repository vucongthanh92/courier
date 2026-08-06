# Agent Gateway Implementation Plan

## Decisions

- Service name: `agent-gateway`.
- Memory/context store: Qdrant.
- `agent-gateway` owns AI context preparation and vector memory management for now.
- `chat-service` remains the source of truth for persisted chat messages.
- Integration boundary between `chat-service` and `agent-gateway`: Kafka events.

## Target Flow

1. User sends a message to a conversation.
2. `chat-service` persists the user message.
3. If the conversation has `type = system` and `name = assistant`, `chat-service` publishes `courier.assistant.requested.v1`.
4. `agent-gateway` consumes the request event.
5. `agent-gateway` stores/updates vector memory in Qdrant.
6. `agent-gateway` builds an AI context package from:
   - fixed system instructions
   - recent assistant conversation messages
   - relevant Qdrant memories
   - optional long-running summary
   - current user message
7. `agent-gateway` calls the configured AI provider.
8. `agent-gateway` publishes `courier.assistant.responded.v1`.
9. `chat-service` consumes the response event and creates an assistant/system message in the original conversation.
10. The existing chat-service websocket path sends the new message to `conversa-app`.

## Phase 1: Local Infrastructure and Service Skeleton

- Create `agent-gateway` as a standalone Go service.
- Add local Docker Compose for Qdrant.
- Add Qdrant config defaults.
- Add Qdrant readiness check.
- Ensure the default memory collection exists on startup.
- Add health endpoint for local verification.

## Phase 2: Event Contracts

- Add assistant requested/responded event payloads.
- Add Kafka topic definitions under `event-bus/kafka/topics.json`.
- Add chat-service producer trigger after user message creation.
- Add agent-gateway consumer for assistant requested events.
- Add chat-service consumer for assistant responded events.

## Phase 3: Memory and Context Pipeline

- Generate embeddings for user and assistant messages.
- Upsert memory points into Qdrant.
- Search relevant memories by current message embedding.
- Build context package with budget limits:
  - recent turns
  - relevant memory chunks
  - summary
  - current message
- Store correlation metadata for debugging and idempotency.

## Phase 4: AI Provider Integration

- Add OpenAI provider client.
- Configure model, embedding model, request timeout, and API key via environment.
- Map AI provider output to assistant response event.
- Add retries with bounded timeout.
- Log usage metadata without logging secrets.

## Phase 5: Reliability

- Make consumers idempotent by `event_id` and/or `correlation_id`.
- Avoid duplicate assistant responses for the same `triggering_message_id`.
- Add retry/DLQ topic policy.
- Add failure response strategy:
  - either create a user-visible failed assistant message
  - or keep failure internal and retry

## Phase 6: UX Enhancements

- Add pending/generating state if needed.
- Add citations/source metadata if retrieval or web search is enabled.
- Add streaming only after final-message MVP is stable.

## Config To Confirm

- OpenAI model name for text generation.
- OpenAI embedding model and vector size.
- Whether Qdrant local should require an API key.
- Kafka topic names.
- Whether assistant replies should use `sender_id = 0` or a virtual assistant user id.
- Maximum recent messages to include in each context package.
- Whether global/current web search is required in the first implementation.

— Nội dung được viết/cập nhật bởi AI (OpenAI Codex).
