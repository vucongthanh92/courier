-- Drop trigger
DROP TRIGGER IF EXISTS outbox_notify_trigger ON outbox;

-- Drop function
DROP FUNCTION IF EXISTS notify_outbox_event();
