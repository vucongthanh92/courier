-- up
CREATE TABLE IF NOT EXISTS "user-service".audit_logs (
    id          BIGSERIAL PRIMARY KEY CHECK (id > 0),
    user_id     BIGINT,
    action      VARCHAR(50),
    ip          VARCHAR(100),
    user_agent  TEXT,
    metadata    JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_audit_logs_user
        FOREIGN KEY (user_id)
        REFERENCES "user-service".users(id)
        ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id
    ON "user-service".audit_logs (user_id);

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at
    ON "user-service".audit_logs (created_at);
