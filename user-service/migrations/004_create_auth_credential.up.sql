-- up
CREATE TABLE IF NOT EXISTS "user-service".auth_credential (
    id                BIGSERIAL PRIMARY KEY CHECK (id > 0),
    user_id           BIGINT NOT NULL,
    password_hash     TEXT,
    password_algo     TEXT,
    mfa_enabled       BOOLEAN NOT NULL DEFAULT FALSE,
    password_version  SMALLINT NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_auth_credential_user
        FOREIGN KEY (user_id)
        REFERENCES "user-service".users(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_auth_credential_user_id
        UNIQUE (user_id),

    CONSTRAINT chk_auth_credential_password_algo
        CHECK (
            password_algo IS NULL
            OR password_algo IN ('sha256', 'bcrypt', 'scrypt')
        )
);

CREATE INDEX IF NOT EXISTS idx_auth_credential_user_id
    ON "user-service".auth_credential (user_id);
