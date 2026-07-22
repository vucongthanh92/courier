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
