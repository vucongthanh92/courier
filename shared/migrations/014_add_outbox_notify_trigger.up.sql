-- up
CREATE OR REPLACE FUNCTION notify_outbox_event()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('outbox_events', NEW.id::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS outbox_notify_trigger ON "cron-service".outbox;

CREATE TRIGGER outbox_notify_trigger
AFTER INSERT ON "cron-service".outbox
FOR EACH ROW
EXECUTE FUNCTION notify_outbox_event();
