# Agent Gateway Implementation Plan

## Decisions

- Service name: `agent-gateway`.
- Memory/context store: Qdrant.
- `agent-gateway` owns AI context preparation and vector memory management for now.
- `chat-service` remains the source of truth for persisted chat messages.
- Integration boundary between `chat-service` and `agent-gateway`: Kafka events.
- MVP generation model target: `gpt-5.5`; verify the exact API model id with the provided API key before production use.
- Embedding model: `text-embedding-3-small` with Qdrant vector size `1536`.
- Assistant replies use `sender_id = 0` when `chat-service` persists the response.
- AI provider failures should produce a user-visible assistant error message for the MVP.
- Web/current-data is enabled for MVP so the assistant can fetch fresh information when needed.
- Guardrails are enabled for MVP to block sensitive, secret-seeking, or prohibited requests before provider execution.

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

Current implementation:

- `chat-service` publishes `courier.assistant.requested.v1` when a normal user message is created inside the `system/assistant` conversation.
- `agent-gateway` consumes assistant request events and publishes `courier.assistant.responded.v1`.
- `chat-service` consumes assistant response events and creates system messages, preserving the existing websocket fanout.

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

Current implementation:

- `agent-gateway` creates embeddings through the OpenAI embeddings endpoint.
- User and assistant memories are upserted into Qdrant.
- Relevant memories are retrieved from Qdrant by conversation id.
- Recent chat-service message history is not included yet because the current event payload only carries the triggering message. Add either a chat-service history API/internal event payload extension in a follow-up if richer short-term context is required.

## Phase 4: AI Provider Integration

- Add OpenAI provider client.
- Configure model, embedding model, request timeout, and API key via environment.
- Map AI provider output to assistant response event.
- Add retries with bounded timeout.
- Log usage metadata without logging secrets.

Current implementation:

- `internal/provider/openai.Client` calls OpenAI `/v1/responses` for generation and `/v1/embeddings` for vector creation.
- Web/current-data is enabled through the Responses API web search tool when `safety.webSearchEnabled = true`.
- Long AI responses are split into message parts capped by `memory.maxMessageRunes`, defaulting to the current chat-service message limit of 4000 runes.
- Provider failures are converted to a user-visible assistant error response.

## Web Search / Global Current Data Decision

There are two different answer modes:

- Model + conversation memory only: the AI answers from the model's trained knowledge, the current user message, recent conversation context, summaries, and Qdrant memories.
- Web/current-data enabled: the AI can call a search or browsing tool/API before answering questions that require fresh information.

Examples that need web/current-data:

- "Tin mới nhất hôm nay về..."
- "Giá hiện tại của..."
- "Version mới nhất của thư viện..."
- "Lịch thi đấu tuần này..."
- "Luật/quy định hiện tại..."

Examples that usually do not need web/current-data:

- Explain a concept.
- Summarize conversation history.
- Help draft text.
- Answer based on information the user already provided.
- Reason over Courier app context and stored assistant memory.

MVP should enable web/current-data retrieval for questions that require fresh information. The assistant should still avoid web search when the answer can be produced safely from conversation context, Qdrant memory, or stable model knowledge.

## Safety / Guardrail Decision

`agent-gateway` should run a safety check before building the final AI request. The guardrail should decide one of:

- `allow`: continue normally.
- `block`: publish an assistant response with a user-visible refusal/error message.
- `review`: optional future state for human/admin review.

Blocked MVP categories:

- requests for secrets, credentials, API keys, tokens, private keys, passwords, or bypass instructions
- illegal behavior or evasion
- self-harm instructions
- sexual content that should not be handled by the assistant
- hate, harassment, or abusive targeting
- violent instructions
- cyber abuse, exploit instructions, credential theft, malware, or phishing
- privacy-invasive requests or attempts to expose private user data

Storage decision:

- Keep deterministic block categories and high-priority rules in config/code.
- Use Qdrant for semantic safety examples, policy snippets, and retrieval-assisted classification.
- Do not rely on Qdrant alone for safety because vector search is approximate and can miss exact prohibited patterns.
- Do not store actual secrets in Qdrant; store only redacted examples or policy metadata.

Current implementation:

- `internal/safety.Guardrail` implements the deterministic MVP guardrail.
- `gateway.Service.EvaluateSafety` exposes the guardrail to orchestration code.
- `POST /v1/safety/evaluate` exists for local/manual guardrail testing.
- Qdrant-backed semantic policy retrieval is intentionally deferred until embeddings and provider calls are wired.

## Phase 5: Reliability

- Make consumers idempotent by `event_id` and/or `correlation_id`.
- Avoid duplicate assistant responses for the same `triggering_message_id`.
- Add retry/DLQ topic policy.
- Add failure response strategy:
  - create a user-visible failed assistant message for MVP
  - keep retry/DLQ metadata for later operations and debugging

## Phase 6: UX Enhancements

- Add pending/generating state if needed.
- Add citations/source metadata if retrieval or web search is enabled.
- Add streaming only after final-message MVP is stable.

## Confirmed Config

- OpenAI generation model target: `gpt-5.5`.
- OpenAI API key will be provided during provider implementation.
- Embedding model: `text-embedding-3-small`.
- Qdrant vector size: `1536`.
- Assistant reply sender: `sender_id = 0`.
- AI provider failures should create a user-visible assistant error message.
- Web/current-data retrieval is enabled for MVP.
- Safety guardrails are enabled for MVP.
- Kafka topic names:
  - `courier.assistant.requested.v1`
  - `courier.assistant.responded.v1`
  - `courier.assistant.events.retry.v1`
  - `courier.assistant.events.dlq.v1`

## Config To Confirm

- Whether Qdrant local should require an API key.
- Maximum recent messages to include in each context package.
- Whether global/current web search is required in the first implementation.

— Nội dung được viết/cập nhật bởi AI (OpenAI Codex).
