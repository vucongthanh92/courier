-- Create function to notify when new outbox event is inserted
CREATE OR REPLACE FUNCTION notify_outbox_event()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('outbox_events', NEW.id::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Create schema if not exists for cron-service
-- Create trigger on outbox table
CREATE TRIGGER outbox_notify_trigger
AFTER INSERT ON "cron-service".outbox
FOR EACH ROW
EXECUTE FUNCTION notify_outbox_event();


-- Create index on outbox table for event_type and created_at to optimize querying pending events
-- Create unique index on refresh_tokens for user_id and user_agent where revoked_at is null to prevent multiple active refresh tokens for same user and device
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_refresh_tokens_user_device
ON "user-service".refresh_tokens (user_id, user_agent)
WHERE revoked_at IS NULL;
