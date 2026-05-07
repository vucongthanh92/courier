-- up
CREATE TABLE IF NOT EXISTS "user-service".email_verifications (
    id          BIGSERIAL PRIMARY KEY CHECK (id > 0),
    user_id     BIGINT NOT NULL,
    email       CITEXT NOT NULL,
    token_hash  TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,

    CONSTRAINT fk_email_verifications_user
        FOREIGN KEY (user_id)
        REFERENCES "user-service".users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_email_verifications_user_id
    ON "user-service".email_verifications (user_id);

CREATE INDEX IF NOT EXISTS idx_email_verifications_email
    ON "user-service".email_verifications (email);

CREATE INDEX IF NOT EXISTS idx_email_verifications_expires_at
    ON "user-service".email_verifications (expires_at);

CREATE UNIQUE INDEX IF NOT EXISTS uq_email_verifications_email_active
    ON "user-service".email_verifications (email)
    WHERE used_at IS NULL;
