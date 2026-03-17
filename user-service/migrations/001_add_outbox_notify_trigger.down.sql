DROP TRIGGER IF EXISTS outbox_notify_trigger ON "cron-service".outbox;
DROP FUNCTION IF EXISTS notify_outbox_event();
