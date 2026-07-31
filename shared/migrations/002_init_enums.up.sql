-- up
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_type t
        JOIN pg_namespace n ON n.oid = t.typnamespace
        WHERE t.typname = 'user_status_enum'
          AND n.nspname = 'user-service'
    ) THEN
        CREATE TYPE "user-service".user_status_enum AS ENUM (
            'pending',
            'verified',
            'suspended',
            'deleted'
        );
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_type t
        JOIN pg_namespace n ON n.oid = t.typnamespace
        WHERE t.typname = 'identity_provider_enum'
          AND n.nspname = 'user-service'
    ) THEN
        CREATE TYPE "user-service".identity_provider_enum AS ENUM (
            'google',
            'github',
            'facebook',
            'twitter',
            'linkedin',
            'apple',
            'microsoft',
            'gitlab',
            'bitbucket'
        );
    END IF;
END $$;
