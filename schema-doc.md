# Courier Schema Documentation

-- users: thông tin cơ bản
CREATE TABLE users (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email           CITEXT UNIQUE NOT NULL,
  email_verified  BOOLEAN NOT NULL DEFAULT FALSE,
  phone           TEXT,
  display_name    TEXT,
  avatar_url      TEXT,
  status          TEXT NOT NULL DEFAULT 'active', -- active|locked|deleted
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- auth_credentials: chỉ cho local/email-password (không lưu plaintext)
CREATE TABLE auth_credentials (
  user_id     UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  password_hash TEXT NOT NULL,
  password_algo TEXT NOT NULL, -- argon2id, scrypt, bcrypt
  mfa_enabled  BOOLEAN NOT NULL DEFAULT FALSE,
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- identities: mapping cho social/SSO providers
CREATE TABLE identities (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider     TEXT NOT NULL,  -- google|facebook|github|oidc:company
  provider_uid TEXT NOT NULL,  -- sub/subject từ IdP
  email_at_auth CITEXT,
  access_token  TEXT,          -- nếu cần call API provider (mã hóa ở app)
  refresh_token TEXT,          -- (mã hóa)
  expires_at    TIMESTAMPTZ,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(provider, provider_uid)
);

-- refresh_tokens: rotate on every use (one-time)
CREATE TABLE refresh_tokens (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash   TEXT NOT NULL, -- hash SHA-256 của refresh token
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at   TIMESTAMPTZ NOT NULL,
  revoked_at   TIMESTAMPTZ
);

-- audit_logs: theo dõi hoạt động
CREATE TABLE audit_logs (
  id          BIGSERIAL PRIMARY KEY,
  user_id     UUID,
  action      TEXT NOT NULL, -- signup | signin | change_pwd | update_profile | link_identity
  ip          INET,
  ua          TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- email_verifications: OTP/email token cho signup/verify email
CREATE TABLE email_verifications (
  id           BIGSERIAL PRIMARY KEY,
  user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  email        CITEXT NOT NULL,
  token_hash   TEXT NOT NULL,  -- hash của OTP/token gửi mail
  expires_at   TIMESTAMPTZ NOT NULL,
  verified_at  TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(user_id, email)
);

-- outbox: event sourcing nhẹ để publish sau commit
CREATE TABLE outbox (
  id             BIGINT PRIMARY KEY,
  aggregate_type TEXT NOT NULL,  -- vd: 'User'
  aggregate_id   TEXT NOT NULL,  -- vd: user id
  event_type     TEXT NOT NULL,  -- vd: USER_CREATED
  payload        JSONB NOT NULL, -- dữ liệu sự kiện
  published_at   TIMESTAMPTZ,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indices & ràng buộc gợi ý
-- users: CREATE UNIQUE INDEX idx_users_email_ci ON users (email);
-- email_verifications: INDEX (user_id), CHECK(expires_at > now() at insert), token_hash được hash để không lộ token gốc;
-- outbox: INDEX (published_at NULLS FIRST) cho worker quét nhanh; optional UNIQUE (aggregate_type, aggregate_id, event_type) nếu muốn chống trùng;
-- identities: UNIQUE(provider, provider_uid) đã có, thêm INDEX(user_id);
-- audit_logs: INDEX(user_id, created_at DESC).



# ERD (ASCII)

+-------------+        1 ── 1      +------------------+
|  users      |<--------------------| auth_credentials |
+-------------+                     +------------------+
| id (UUID) PK|                     | user_id (PK, FK) |
| email (uniq)|                     | password_hash    |
| status      |                     | mfa_enabled      |
| ...         |                     | ...              |
+------+------+                     +------------------+
       | 1                                   
       | 1 ── *                     +------------------+
       +---------------------------> | identities       |
                                     +------------------+
                                     | id (UUID) PK     |
                                     | user_id (FK)     |
                                     | provider, uid    |
                                     | ...              |
                                     +------------------+

       | 1 ── *                     +------------------+
       +---------------------------> | refresh_tokens   |
                                     +------------------+

       | 1 ── *                     +------------------+
       +---------------------------> | email_verif.     |
                                     +------------------+

       | 1 ── *                     +------------------+
       +---------------------------> | password_resets  |
                                     +------------------+

       | 1 ── 0..1                  +------------------+
       +---------------------------> | mfa_totp         |
                                     +------------------+

       | 1 ── *                     +------------------+
       +---------------------------> | audit_logs       |
                                     +------------------+

                                     +------------------+
                                     | jwk_keys         |
                                     +------------------+




# Design Notes & Best Practices
- Email (CITEXT): Case-insensitive matching, set to UNIQUE.
- Data Deletion:
    users.deleted_at is soft-delete (preserves history). For sensitive PII requiring "true deletion," you can: pseudoonymize and hard-delete linked tables (DELETE CASCADE enabled).
- Tokens: Always save the SHA-256 (hex) hash of refresh/email-verify/reset tokens; do not save raw tokens.
- Rotate refresh token: Use parent_id/replaced_by_id to trace the string and detect reuse (revoke the entire branch).
- MFA: secret_enc and recovery_codes_hash[] — recovery code checks by hashed, replaces with a trace (e.g., timestamp) after use, or removes the element from the array.
- Identities: UNIQUE(provider, provider_uid) ensures a social account is associated with only one user.
- JWKS: holds >=2 keys: current and retired (to verify the old token). When rotating: add a new current key, change the status of the old key to retired, and revoke the access token after the TTL.
- Conditional index: added for "unused/currently active" cases.
- IP/User-Agent: uses INET data for IP addresses to query subnets and statistics.
- JSONB metadata: allows for flexible profile expansion (dob, locale, timezone, etc.) without early migration.


# Extensions (optional, add later)
- roles, permissions, user_roles, role_permissions: if you need RBAC.
- session: if you want to actively track sessions (different from refresh token).
- auth_providers: dynamic configuration (OIDC issuer, client_id/secret) if not using env.


## 1. Users
- Meaning: The "root" user profile – the central point of all relationships.
- Contains: email, phone, display_name, avatar_url, status (active/locked/deleted), metadata (extended JSONB), timestamps.
- Used for:
    Creating new users (sign-up), reading profiles, updating personal information.
    Serves as a foreign key for most other tables (credentials, identities, tokens, logs, etc.).
- Note:
    email is unique (CITEXT to be case-insensitive).
    metadata allows for extension (locale, timezone, dob, etc.) without needing migration.

## 2. Auth_credentials
- Meaning: Internal login information (email/password).
- Contains: password_hash, password_algo (argon2id/bcrypt/scrypt), mfa_enabled, password_updated_at, password_version.
- Used for:
    Signing in with email/password, changing password.
    Checking if MFA is enabled.
- Note:
    Never save plaintext passwords; only secure hashes.
    When changing passwords: update password_updated_at, the old refresh can be revoked.

## 3. Identities
- Meaning: Links Social/SSO accounts to internal users.
- Contains: provider (google/facebook/github/apple/oidc), provider_uid (sub/subject), email_at_auth, scopes, encrypted token (*_enc) + expires_at.
- Used for:
    Logging in via OIDC/OAuth2.
    “Link account” (when a user is already logged in internally, add a Google account…).
- Note:
    UNIQUE (provider, provider_uid) ensures that one social account is linked to only one user.
    If you need to call the provider's API (Google/Facebook), store the token in encrypted form (application-level/KMS), not plaintext.

## 4. refresh_tokens
- Meaning: Manages long-term login sessions (refresh tokens) + prevents “token reuse”.
- Contains: token_hash (SHA-256 hex), rotation string parent_id/replaced_by_id, expires_at, revoked_at, ip, user_agent.
- Used for:
    Issuing pairs (access, refresh), rotate-on-use: each refresh creates a new token, the old token is marked as replaced.
    Detects attacks (if the old token is reused ⇒ revoke the branch).
- Note:
    Only stores the hash of the refresh token (for security).
    Conditional index for valid tokens.

## 5. email_verifications
- Meaning: Verifys email using a one-time code.
- Contains: email, token_hash, expires_at, used_at.
- Used for:
    After signing up or changing email, send a verification link. When clicked, compare token_hash and expires_at; if valid, set users.email_verified = true.
- Note:
    Only store the verification code hash. It is recommended to periodically clean up records that have been used_at or have expired.

## 6. password_resets
- Meaning: Password reset process.
- Contains: token_hash, requested_at, expires_at, used_at, ip, user_agent.
- Used for:
    Sending an email containing a reset link. The user opens the link → verifies the token → sets a new password → sets used_at, records audit_logs.
- Note:
    Only the hash is saved. The reset token should have a short expiration date (e.g., 15–30 minutes).
    Logs IP/UA to detect abuse.

## 7. mfa_totp
- Meaning: Configures MFA TOTP (Google Authenticator/Authenticator apps).
- Contains: secret_enc (TOTP encryption key), issuer, label, recovery_codes_hash[], created_at, last_used_at.
- Used for:
    Enabling/disabling MFA, verifying OTP codes, using recovery codes when a device is lost.
- Note:
    Recovery codes save a hash for each code; when finished, delete/destroy the corresponding element.
    Do not save the secret TOTP plaintext; it is always encrypted in the app or KMS.

## 8. audit_logs
- Meaning: Security logs & important behavior logs (audit trail).
- Contains: action (enum: signup/signin/refresh/change_pwd/update_profile/link_identity/…), user_id (can be null), IP, user_agent, meta (JSONB).
- Used for:
    Incident investigation, compliance, login anomaly detection.
    Statistical analysis of behavior (time, location, device).
- Note:
    Increases in size over time ⇒ consider monthly partitioning for querying/cleaning.
    Logs important behaviors (password changes, enabling/disabling MFA, password reset…).

## 9. jwk_keys
- Meaning: Manages JWT and JWKS signing keys (for other services to verify).
- Contains: kid, alg, kty, public_pem, private_pem_enc, status (current/retired/revoked), timestamps.
- Used for:
    Signing access tokens (RS256/ES256…), key rotation.
    Exporting JWKS via the /.well-known/jwks.json endpoint for services to verify by kid.
- Note:
    Always keep at least two keys: current + retired (to verify old tokens).
    The private key stores encryption (KMS/app-level), not plaintext.



# Outbox chart — ensuring “at-least-once delivery” (Transactional Outbox Pattern)
- Practical Problem
- Let's say in the SignUp use case, after creating a user in the database, you want to:
    Send a welcome email via a message broker (Kafka, RabbitMQ, etc.),
    Or notify another service (Analytics, Notification, etc.).
- If you perform two separate actions:
    tx.Commit() — successfully saves the user.
    publisher.Publish("user.created") — calls the API to Kafka (and... network error!)
- The user already exists in the database, but the event hasn't been sent.
- This results in a "lost event" → other services won't know the new user has been created.
- Solution: Transactional Outbox:
    Instead of sending the event directly, you save the event to the outbox table within the same transaction as the data change.
- Operation Flow:
    BEGIN TRANSACTION
        INSERT INTO users (...)
        INERT INTO outbox (aggregate_type='user', aggregate_id='123', event_type='created', payload='...')
    COMMIT
- Both records are committed at the same time.
- You ensure that if the user is successfully saved, the event also exists (even if not yet published).
- Then, a background worker (dispatcher) reads the outbox records that haven't been published_at, sends them to the broker, and updates published_at = now() upon success.

    ____Column________________Meaning
    ____id____________________Snowflake ID
    ____aggregate_type________Entity type (user, order, payment…)
    ____aggregate_id__________Entity ID
    ____event_type____________Action (created, updated, deleted)
    ____payload JSON__________containing event data
    ____created_at____________When written to the DB
    ____published_at__________When the worker has finished publishing

## Background worker (dispatcher)
- Runs periodically (e.g., every 1–5 seconds),
- Retrieves the lines WHERE published_at IS NULL,
- Publish to Kafka/NATS, then update published_at=now().

## Benefits:
- Events are not lost even if the app crashes immediately after committing.
- Integrity is maintained between the database and message broker (atomicity).
- Easier to control retry, audit events, and replays if needed.

## Note:
- Ensure idempotent is published in the worker (republishing with the same event should not cause errors).
- You can partition or use TTL for the outbox table to avoid bloat.


# Idempotency chart — ensuring “at-most-once effect”

## The practical problem
- Suppose the client (mobile app/frontend) calls the API:
    POST /v1/auth/sign-up
- But the network is weak → the user clicks "Register" twice, or automatically retrys (HTTP 5xx → resend request).
- If not controlled, you will:
    Create two users with the same email address, or Send two verification emails consecutively, or Cause a race condition.

## Solution: Idempotency Key
- Client sends header:
    Idempotency-Key: a83d-abc-123
- Server will:
    Check the idempotency table to see if this key has already been processed.
    If not: Create a new record: status=pending, save the request_hash.
    Process the normal business logic.
    If successful, save the JSON response to the response, set status=succeeded.
- If it already exists:
    Compare the request_hash to ensure the client sent the same payload.
    Return the old result (response) without re-executing.

## Table Structure
    ____Column________________Meaning
    ____key___________________Idempotency - Key sent by client (PK)
    ____request_sig___________SHA256 hash of request body
    ____response______________JSON of response (stores byte[])
    ____status________________pending / succeeded / failed
    ____created_at____________When the request is first received
    ____expires_at____________When the record expires (TTL, e.g., 24h)

`
Client sends:
POST /signup
Idempotency-Key: abc123
Body: {"email": "alice@example.com"}

Server:
- Search for record key=abc123 → not found
- INSERT key=abc123, status=pending
- Process signup
- After success: UPDATE record.status='succeeded', response=<JSON>
Resend:
- Find key=abc123, request_sig matches
- Return saved response
`

## Benefits:
- Avoid creating duplicate users/payments/orders when retrying.
- Ensure the API is "idempotent" (truly RESTful).
- Reduces load and side-effect errors.

## Note:
- TTL: periodically delete old records (e.g., expires_at < now()).
- Hash body (request_sig) to ensure the same payload.
- If you need faster caching, you can buffer Idempotency in Redis (but the database is still the safest source for consistency).
