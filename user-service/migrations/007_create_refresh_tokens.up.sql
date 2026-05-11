-- up
CREATE TABLE IF NOT EXISTS "user-service".refresh_tokens (
    id              BIGSERIAL PRIMARY KEY CHECK (id > 0),
    user_id         BIGINT NOT NULL,
    token_hash      TEXT NOT NULL,
    parent_id       BIGINT,
    replaced_by_id  BIGINT,
    user_agent      TEXT NOT NULL DEFAULT 'unknown',
    ip              VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,

    CONSTRAINT fk_refresh_tokens_user
        FOREIGN KEY (user_id)
        REFERENCES "user-service".users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_refresh_tokens_parent
        FOREIGN KEY (parent_id)
        REFERENCES "user-service".refresh_tokens(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_refresh_tokens_replaced_by
        FOREIGN KEY (replaced_by_id)
        REFERENCES "user-service".refresh_tokens(id)
        ON DELETE SET NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_refresh_tokens_token_hash
    ON "user-service".refresh_tokens (token_hash);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id
    ON "user-service".refresh_tokens (user_id);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at
    ON "user-service".refresh_tokens (expires_at);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_revoked_at
    ON "user-service".refresh_tokens (revoked_at);

CREATE UNIQUE INDEX IF NOT EXISTS idx_refresh_tokens_user_device
    ON "user-service".refresh_tokens (user_id, user_agent)
    WHERE revoked_at IS NULL;
