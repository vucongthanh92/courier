-- up
CREATE TABLE IF NOT EXISTS "user-service".users (
    id              BIGINT PRIMARY KEY CHECK (id > 0),
    email           CITEXT NOT NULL,
    email_verified  BOOLEAN NOT NULL DEFAULT FALSE,
    phone_number    VARCHAR(50),
    phone_verified  BOOLEAN NOT NULL DEFAULT FALSE,
    display_name    VARCHAR(255),
    avatar_url      TEXT,
    status          "user-service".user_status_enum NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_email
    ON "user-service".users (email);

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_phone_number
    ON "user-service".users (phone_number)
    WHERE phone_number IS NOT NULL AND phone_number <> '';

CREATE INDEX IF NOT EXISTS idx_users_deleted_at
    ON "user-service".users (deleted_at);
