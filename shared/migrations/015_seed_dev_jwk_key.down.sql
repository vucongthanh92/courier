-- down
DELETE FROM "user-service".jwk_key
WHERE kid = 'user-service-dev-rs256-20260507';
