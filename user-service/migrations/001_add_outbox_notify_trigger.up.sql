-- Create function to notify when new outbox event is inserted
CREATE OR REPLACE FUNCTION notify_outbox_event()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('outbox_events', NEW.id::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger on outbox table
CREATE TRIGGER outbox_notify_trigger
AFTER INSERT ON "cron-service".outbox
FOR EACH ROW
EXECUTE FUNCTION notify_outbox_event();
