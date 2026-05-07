-- up
CREATE TABLE IF NOT EXISTS "user-service".jwk_key (
    id           BIGSERIAL PRIMARY KEY,
    kid          TEXT NOT NULL,
    alg          TEXT NOT NULL,
    kty          TEXT NOT NULL,
    public_pem   TEXT NOT NULL,
    private_pem  TEXT NOT NULL,
    active       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rotated_at   TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_jwk_key_kid
    ON "user-service".jwk_key (kid);

CREATE UNIQUE INDEX IF NOT EXISTS uq_jwk_key_one_active
    ON "user-service".jwk_key ((active))
    WHERE active = TRUE;
