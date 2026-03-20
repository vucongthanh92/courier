DROP TRIGGER IF EXISTS outbox_notify_trigger ON "cron-service".outbox;
DROP FUNCTION IF EXISTS notify_outbox_event();


-- down
DROP INDEX CONCURRENTLY IF EXISTS idx_refresh_tokens_user_agent_active;