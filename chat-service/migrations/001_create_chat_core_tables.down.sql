DROP TRIGGER IF EXISTS trg_sync_conversation_last_message ON "chat-service".messages;
DROP FUNCTION IF EXISTS "chat-service".sync_conversation_last_message();

DROP TRIGGER IF EXISTS trg_validate_message_sender_membership ON "chat-service".messages;
DROP FUNCTION IF EXISTS "chat-service".validate_message_sender_membership();

DROP TRIGGER IF EXISTS trg_validate_member_last_read_message ON "chat-service".conversation_members;
DROP FUNCTION IF EXISTS "chat-service".validate_member_last_read_message();

DROP TRIGGER IF EXISTS trg_validate_direct_membership_limit ON "chat-service".conversation_members;
DROP FUNCTION IF EXISTS "chat-service".validate_direct_membership_limit();

ALTER TABLE IF EXISTS "chat-service".conversation_members
    DROP CONSTRAINT IF EXISTS fk_conversation_members_last_read_message;

ALTER TABLE IF EXISTS "chat-service".conversations
    DROP CONSTRAINT IF EXISTS fk_conversations_last_message;

DROP TABLE IF EXISTS "chat-service".messages;
DROP TABLE IF EXISTS "chat-service".conversation_members;
DROP TABLE IF EXISTS "chat-service".conversations;

DROP TYPE IF EXISTS "chat-service".message_type_enum;
DROP TYPE IF EXISTS "chat-service".conversation_member_status_enum;
DROP TYPE IF EXISTS "chat-service".conversation_member_role_enum;
DROP TYPE IF EXISTS "chat-service".conversation_type_enum;

DROP SCHEMA IF EXISTS "chat-service";
