# Chat Service

`chat-service` owns conversations, membership, messages, realtime delivery, and chat-domain integration events.

## Local Development

```bash
make run-local
make test
```

Local API base URL:

```text
http://localhost:5002/api/v1
```

## Create Conversation

`conversa-app` creates conversations with:

```text
POST /api/v1/conversation/create
Authorization: Bearer <access_token>
```

Request shape:

```json
{
  "name": "optional custom name",
  "member_user_ids": ["124397457160273920"]
}
```

Important behavior:

- The authenticated creator is added automatically.
- `member_user_ids` are accepted as string IDs to avoid JavaScript precision loss.
- Conversation type is inferred by selected member count:
  - one selected user creates a `direct` conversation
  - two or more selected users create a `group` conversation
- If `name` is provided, it is used for both direct and group conversations.
- If `name` is omitted, the name is generated from participant display names.
- Existing direct conversations are still detected and returned as a duplicate error.

## Conversation Created Event

After a conversation is created, `chat-service` publishes a Kafka event to:

```text
courier.chat.events.v1
```

Event contract:

```text
event-bus/contracts/conversation.created.v1.json
```

Publishing is intentionally non-blocking after the conversation database write. If Kafka is unavailable, the conversation remains created and the failure is logged instead of returning a false failure to the user.

## Notification System Messages

`chat-service` also consumes `conversation.created.v1` in consumer group `chat-service`.

For every member in the new conversation, it creates a system message in that member's `notification` system conversation:

```text
Bạn đã được thêm vào một cuộc trò chuyện mới
```

The message is inserted through the same message creation path used by normal create-message flows, so realtime WebSocket delivery is preserved for active clients.
