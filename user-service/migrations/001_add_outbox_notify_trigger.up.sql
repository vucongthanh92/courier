-- Create function to notify when new outbox event is inserted
CREATE OR REPLACE FUNCTION notify_outbox_event()
RETURNS TRIGGER AS $$
BEGIN
    -- Send notification with event ID as payload
    PERFORM pg_notify('outbox_events', NEW.id::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger on outbox table
CREATE TRIGGER outbox_notify_trigger
AFTER INSERT ON outbox
FOR EACH ROW
EXECUTE FUNCTION notify_outbox_event();
