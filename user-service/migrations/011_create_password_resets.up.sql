-- up
CREATE TABLE IF NOT EXISTS "user-service".password_resets (
    id            BIGSERIAL PRIMARY KEY CHECK (id > 0),
    user_id       BIGINT NOT NULL,
    token_hash    TEXT NOT NULL,
    requested_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at    TIMESTAMPTZ NOT NULL,
    used_at       TIMESTAMPTZ,
    ip            INET,
    user_agent    TEXT,

    CONSTRAINT fk_password_resets_user
        FOREIGN KEY (user_id)
        REFERENCES "user-service".users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_password_resets_user_id
    ON "user-service".password_resets (user_id);

CREATE INDEX IF NOT EXISTS idx_password_resets_expires_at
    ON "user-service".password_resets (expires_at);
