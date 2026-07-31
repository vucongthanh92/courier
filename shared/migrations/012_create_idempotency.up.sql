-- up
CREATE TABLE IF NOT EXISTS "user-service".idempotency (
    key          TEXT PRIMARY KEY,
    request_sig  TEXT NOT NULL,
    response     BYTEA NOT NULL,
    status       TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_idempotency_expires_at
    ON "user-service".idempotency (expires_at);
