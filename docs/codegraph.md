# Codegraph

This document gives a lightweight map of how code flows through courier services.

## Service Layer Pattern

Most services follow this shape:

```mermaid
flowchart LR
    Route["Route"] --> Handler["HTTP Handler"]
    Handler --> Model["Request / Response Models"]
    Handler --> Usecase["Usecase"]
    Usecase --> Interface["Domain Interfaces"]
    Interface --> Repo["Repository Implementation"]
    Repo --> DB["Database"]
    Usecase --> External["External Client"]
```

## user-service

```mermaid
flowchart LR
    HTTP["HTTP API"] --> AuthHandler["Auth / Identity / Credential Handlers"]
    AuthHandler --> AuthUsecase["Auth / Identity / Credential Usecases"]
    AuthUsecase --> UserRepo["User Repository"]
    AuthUsecase --> CredentialRepo["Credential Repository"]
    AuthUsecase --> IdentityRepo["Identity Repository"]
    AuthUsecase --> TokenUsecase["Token Usecase"]
    TokenUsecase --> JWTSigner["JWT Signer"]
    JWTSigner --> JWKRepo["JWK Repository"]
    JWKRepo --> UserDB["user-service DB"]
    AuthUsecase --> Outbox["Outbox"]
    Outbox --> Worker["Outbox Worker"]

    GRPC["gRPC Server"] --> GrpcUsecase["GrpcUsecase"]
    GrpcUsecase --> JWKRepo
    GrpcUsecase --> JWKCache["Redis JWK Cache"]
    GrpcUsecase --> UserRepo
```

Important files:

- `user-service/internal/api/http/v1/`
- `user-service/internal/usecase/`
- `user-service/internal/repository/`
- `user-service/internal/api/grpc/grpc_server.go`
- `user-service/internal/api/grpc/grpc_usecase.go`

## chat-service

```mermaid
flowchart LR
    Route["Routes"] --> ConversationHandler["Conversation Handler"]
    ConversationHandler --> ConversationUsecase["Conversation Usecase"]
    ConversationUsecase --> ConversationRepo["Conversation Repository"]
    ConversationUsecase --> MemberRepo["Member Repository"]
    ConversationRepo --> ChatDB["chat-service DB"]
    MemberRepo --> ChatDB

    HTTPMiddleware["JWT Middleware"] --> UserGrpcClient["user_grpc Client"]
    ConversationUsecase --> UserGrpcClient
    UserGrpcClient --> UserServiceGrpc["user-service gRPC"]
    HTTPMiddleware --> JWKCache["Redis JWK Cache"]
```

Important files:

- `chat-service/internal/api/http/v1/conversation_handler.go`
- `chat-service/internal/usecase/conversation/conversation_uc.go`
- `chat-service/internal/repository/persistent/`
- `chat-service/internal/repository/external/user_grpc/client.go`
- `chat-service/internal/api/http/middleware/auth_middleware.go`

## JWT Public Key Verification Flow

```mermaid
sequenceDiagram
    participant Client
    participant Chat as chat-service
    participant Redis as Redis
    participant User as user-service gRPC
    participant DB as user-service DB

    Client->>Chat: Request with JWT
    Chat->>Chat: Parse JWT header and read kid
    Chat->>Redis: Get public key by kid
    alt Cache hit
        Redis-->>Chat: PublicPEM
    else Cache miss
        Chat->>User: GetPublicKey(kid)
        User->>Redis: Get public key by kid
        alt user-service cache hit
            Redis-->>User: PublicPEM
        else user-service cache miss
            User->>DB: Query jwk_key by kid
            DB-->>User: PublicPEM
            User->>Redis: Cache PublicPEM
        end
        User-->>Chat: PublicPEM
        Chat->>Redis: Cache PublicPEM
    end
    Chat->>Chat: Verify JWT
```

## Create Conversation Flow

```mermaid
sequenceDiagram
    participant Client
    participant Chat as chat-service
    participant User as user-service gRPC
    participant DB as chat-service DB

    Client->>Chat: POST /api/v1/conversation/create
    Chat->>Chat: Extract creator id from JWT claims
    Chat->>Chat: Normalize member_user_ids
    Chat->>Chat: Validate conversation type and member count
    Chat->>User: CheckUsersStatus(member_user_ids)
    User-->>Chat: all_verified + invalid_user_ids
    alt all users verified
        Chat->>DB: Create conversation
        Chat->>DB: Create conversation members
        DB-->>Chat: Created rows
        Chat-->>Client: Conversation response
    else invalid user status
        Chat-->>Client: 400 invalid_user_status
    end
```

## Cross-Service Contract Rule

If code crosses service boundaries, start from `docs/grpc-contracts.md` and `shared/grpc/`. Avoid direct database access across service ownership boundaries.

## payment-gateway (planned)

`payment-gateway` will own payment attempts, wallet accounting and the
immutable ledger. It will not read another service database: `user-service`
remains the authority for identity and user eligibility; `chat-service` owns
delivery of notification messages.

```mermaid
flowchart LR
    Client["Courier Client"] -->|"JWT + Idempotency-Key"| Routes["HTTP v1 routes"]
    Routes --> TopUpHandler["Top-up Handler"]
    Routes --> WalletHandler["Wallet / History Handler"]
    Routes --> IPNHandler["SePay IPN Handler"]

    TopUpHandler --> TopUpUC["TopUp Usecase"]
    WalletHandler --> WalletUC["Wallet Usecase"]
    IPNHandler --> WebhookUC["Webhook Usecase"]

    TopUpUC --> Gateway["PaymentGateway port"]
    Gateway --> SePay["SePayProvider"]
    SePay -->|"signed checkout form"| SePayCheckout["SePay hosted checkout"]
    SePayCheckout -->|"IPN"| IPNHandler

    TopUpUC --> TopUpRepo["Top-up / Idempotency repositories"]
    WalletUC --> WalletRepo["Wallet / Ledger read repositories"]
    WebhookUC --> LedgerUC["Ledger posting usecase"]
    WebhookUC --> ProviderRepo["Provider event repository"]
    LedgerUC --> LedgerRepo["Journal / Entry repository"]
    LedgerUC --> BalanceRepo["Wallet balance projection repository"]
    LedgerUC --> OutboxRepo["Outbox repository"]

    TopUpRepo --> PaymentDB[("payment-gateway schema")]
    WalletRepo --> PaymentDB
    ProviderRepo --> PaymentDB
    LedgerRepo --> PaymentDB
    BalanceRepo --> PaymentDB
    OutboxRepo --> PaymentDB

    OutboxWorker["Outbox publisher worker"] --> OutboxRepo
    OutboxWorker --> EventBus["courier payment events"]
    EventBus --> ChatConsumer["chat-service payment consumer"]
    ChatConsumer --> NotifyConversation["system / notification conversation"]

    AuthMiddleware["JWT middleware"] --> UserJWK["user-service public JWK / cache"]
    TopUpUC --> UserEligibility["user-service status contract"]
```

### Top-up and credit flow

```mermaid
sequenceDiagram
    participant C as Client
    participant PG as payment-gateway
    participant DB as payment-gateway DB
    participant S as SePay Gateway
    participant E as Event Bus
    participant CS as chat-service

    C->>PG: POST /wallet/top-ups + Idempotency-Key
    PG->>PG: validate JWT, eligible user, limits and VND
    PG->>DB: create pending topup_intent + idempotency record
    PG->>PG: generate unique provider_invoice_number and SePay signature
    PG-->>C: checkout action + signed fields + expiry
    C->>S: POST checkout form
    S-->>C: redirect result (display only)
    S->>PG: POST /webhooks/sepay (IPN)
    PG->>PG: authenticate IPN and validate amount/status/invoice
    PG->>DB: lock intent + balance; persist provider event/transaction
    PG->>DB: post balanced journal + update projection + insert outbox
    PG-->>S: HTTP 200
    PG->>E: publish payment.wallet_credited.v1
    E->>CS: consume event idempotently
    CS->>CS: create system message in notification conversation
```

### Database ownership map

| Aggregate | Tables | Write owner | Notes |
| --- | --- | --- | --- |
| Wallet | `wallets`, `wallet_balances` | Wallet/Ledger usecases | Balance is a locked, rebuildable projection. |
| Accounting | `ledger_accounts`, `ledger_journals`, `ledger_entries` | Ledger posting usecase only | Journal/entries are immutable and balanced. |
| Payment attempt | `topup_intents`, `idempotency_keys` | TopUp usecase | One provider invoice per intent; client replay is safe. |
| Provider evidence | `provider_events`, `provider_transactions` | Webhook/Reconciliation usecases | Deduplicate before ledger posting. |
| Integration | `outbox_events` | Same DB transaction as ledger posting | Ensures eventual notification event publication. |
| Operations | `reconciliation_runs`, `reconciliation_items`, `audit_logs` | Workers/admin operations | Never changes accounting entries directly. |

### Planned primary paths

| Concern | Primary path |
| --- | --- |
| HTTP API | `payment-gateway/internal/api/http/v1/` |
| JWT and request safeguards | `payment-gateway/internal/api/http/middleware/` |
| Top-up, wallet and ledger rules | `payment-gateway/internal/usecase/{topup,wallet,ledger,webhook}/` |
| Provider port and normalized types | `payment-gateway/internal/domain/interfaces/payment_gateway.go` |
| SePay adapter | `payment-gateway/internal/repository/external/sepay/` |
| PostgreSQL persistence | `payment-gateway/internal/repository/persistent/` |
| Event publication/reconciliation/expiry | `payment-gateway/internal/worker/` |
| Shared cross-service contracts | `shared/grpc/` and `event-bus/contracts/` |
