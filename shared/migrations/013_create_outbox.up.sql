-- up
CREATE TABLE IF NOT EXISTS "cron-service".outbox (
    id              BIGINT PRIMARY KEY CHECK (id > 0),
    aggregate_type  TEXT NOT NULL,
    aggregate_id    TEXT NOT NULL,
    event_type      TEXT NOT NULL,
    payload         BYTEA NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_outbox_event_type_created_at
    ON "cron-service".outbox (event_type, created_at);

CREATE INDEX IF NOT EXISTS idx_outbox_published_at_created_at
    ON "cron-service".outbox (published_at, created_at);
