-- up
CREATE TABLE IF NOT EXISTS "user-service".identities (
    id                 BIGSERIAL PRIMARY KEY CHECK (id > 0),
    user_id            BIGINT NOT NULL,
    provider           "user-service".identity_provider_enum NOT NULL,
    provider_uid       TEXT NOT NULL,
    email_at_auth      CITEXT,
    scopes             TEXT[],
    access_token_enc   BYTEA,
    refresh_token_enc  BYTEA,
    expires_at         TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ,

    CONSTRAINT fk_identities_user
        FOREIGN KEY (user_id)
        REFERENCES "user-service".users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_identities_user_id
    ON "user-service".identities (user_id);

CREATE INDEX IF NOT EXISTS idx_identities_provider
    ON "user-service".identities (provider);

CREATE UNIQUE INDEX IF NOT EXISTS uq_identities_provider_uid_active
    ON "user-service".identities (provider, provider_uid)
    WHERE deleted_at IS NULL;
