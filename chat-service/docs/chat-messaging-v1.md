# Chat Messaging V1

## Muc tieu

Xay dung `chat-service` cho phep user nhan tin text 1-1 va nhom, tach rieng khoi `user-service` nhung van tai su dung bang `"user-service".users` lam nguon user chinh.

## Pham vi V1

- Ho tro `direct conversation` giua 2 user.
- Ho tro `group conversation` voi nhieu user.
- Ho tro gui `text message`.
- Theo doi thanh vien, vai tro co ban, trang thai roi nhom.
- Luu `last_message`, `last_read_message` de phuc vu inbox va unread count o cac buoc sau.

## Database da duoc thiet ke

### 1. `chat-service.conversations`

- Dai dien cho 1 cuoc tro chuyen.
- `type`: `direct` hoac `group`.
- `direct_key`: khoa duy nhat cho chat 1-1, giup tranh tao 2 room direct cho cung 1 cap user.
- `last_message_id`, `last_message_at`: toi uu danh sach inbox.

### 2. `chat-service.conversation_members`

- Quan ly user trong cuoc tro chuyen.
- `role`: `owner`, `admin`, `member`.
- `status`: `active`, `left`, `removed`.
- `last_read_message_id`, `last_read_at`: phuc vu unread va read state.

### 3. `chat-service.messages`

- Luu message text.
- `client_message_id`: idempotency key theo conversation de tranh gui trung tu mobile/web retry.
- `reply_to_message_id`: de san cho thread/reply sau nay.

### 4. Trigger / rang buoc

- Khong cho `direct conversation` co qua 2 thanh vien active.
- Khong cho user gui message neu khong con la member active.
- Tu dong dong bo `conversations.last_message_id` va `last_message_at` khi co tin nhan moi.

## User Story backlog

### Story 1: Tao cuoc tro chuyen 1-1

As a user, I want tao room chat voi mot user khac so that toi co the gui text message rieng tu.

Acceptance criteria:

- Neu room direct da ton tai giua 2 user thi tra ve room cu.
- Neu chua ton tai thi tao `conversation` type `direct`.
- He thong tao dung 2 `conversation_members` active.
- User khong the tao room direct voi chinh minh.

### Story 2: Tao nhom chat

As a user, I want tao group chat so that nhieu user co the trao doi trong cung mot room.

Acceptance criteria:

- Nguoi tao nhom tro thanh `owner`.
- Group phai co `name`.
- Co it nhat 2 thanh vien active khi tao nhom.
- Moi user chi xuat hien 1 lan trong cung group.

### Story 3: Gui text message

As a user, I want gui text message vao room so that cac thanh vien khac co the doc duoc.

Acceptance criteria:

- Chi member active moi gui duoc.
- Message body khong duoc rong.
- Sau khi gui, `last_message_id` va `last_message_at` duoc cap nhat.
- Retry cung `client_message_id` khong tao duplicate message.

### Story 4: Doc danh sach hoi thoai

As a user, I want xem danh sach hoi thoai so that toi biet room nao co tin nhan moi.

Acceptance criteria:

- Danh sach sap xep theo `last_message_at` giam dan.
- Co du thong tin room, message cuoi, so thanh vien, read state co ban.
- Chi tra ve room ma user la member active.

### Story 5: Danh dau da doc

As a user, I want cap nhat message da doc so that he thong tinh unread count chinh xac.

Acceptance criteria:

- Update `last_read_message_id`, `last_read_at` theo member.
- Khong cho danh dau message thuoc room khac.
- Unread count duoc tinh tu `last_read_message_id`.

## Cac buoc ky thuat tiep theo

1. Tao `chat-service` skeleton theo convention cua `user-service`: `main.go`, `startup/`, `internal/`, `config/`, `migrations/`.
2. Them entity/domain model cho `Conversation`, `ConversationMember`, `Message`.
3. Xay repository cho create conversation, add member, send message, list inbox, list messages.
4. Them usecase transaction:
   - tao direct conversation
   - tao group conversation
   - send text message
   - mark as read
5. Thiet ke API/gRPC:
   - `POST /conversations/direct`
   - `POST /conversations/groups`
   - `GET /conversations`
   - `GET /conversations/{id}/messages`
   - `POST /conversations/{id}/messages`
   - `POST /conversations/{id}/read`
6. Them authN/authZ: lay `user_id` tu access token, validate member truoc moi thao tac.
7. Them outbox event cho:
   - `chat.conversation.created`
   - `chat.message.sent`
   - `chat.message.read`
8. Ket noi realtime o phase sau qua WebSocket/Kafka/NATS, nhung V1 co the chay polling API truoc.
9. Viet test:
   - repository integration test cho constraint quan trong
   - usecase test cho direct/group/message flow
   - regression test cho duplicate direct room va duplicate client message

## Ghi chu kien truc

- Nen de `user-service` tiep tuc la source of truth cho user profile.
- `chat-service` chi nen luu cac du lieu can cho messaging, tranh copy toan bo user profile.
- Neu sau nay can tim kiem nhanh, co the bo sung bang projection hoac cache inbox/unread trong Redis.
