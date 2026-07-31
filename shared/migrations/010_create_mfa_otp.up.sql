-- up
CREATE TABLE IF NOT EXISTS "user-service".mfa_otp (
    id                   BIGSERIAL PRIMARY KEY CHECK (id > 0),
    user_id              BIGINT NOT NULL,
    secret_enc           BYTEA NOT NULL,
    issuer               TEXT NOT NULL,
    label                TEXT NOT NULL,
    recovery_codes_hash  TEXT[] NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at         TIMESTAMPTZ,

    CONSTRAINT fk_mfa_otp_user
        FOREIGN KEY (user_id)
        REFERENCES "user-service".users(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_mfa_otp_user_id
        UNIQUE (user_id)
);
