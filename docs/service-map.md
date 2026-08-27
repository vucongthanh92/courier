# Service Map

This document is the quick entry point for understanding the courier repository before changing code.

## Repository Layout

| Path | Purpose |
| --- | --- |
| `user-service/` | Owns users, identities, credentials, JWT issuing, JWK keys, and user-facing auth flows. |
| `chat-service/` | Owns conversations, members, chat APIs, JWT verification, and calls user-service through gRPC. |
| `payment-gateway/` | Owns wallet balances, ledger records, top-up intents, and payment-provider integrations. |
| `shared/` | Holds shared contracts and generated code, currently focused on gRPC protobuf contracts. |
| `infra/` | Local infrastructure manifests such as kind and Argo CD setup. |
| `.github/workflows/` | GitHub Actions workflows. |
| `argocd/` | Argo CD application manifests. |
| `docs/` | Project documentation and operating guides. |

## user-service

| Area | Path |
| --- | --- |
| Entrypoint | `user-service/main.go` |
| Dependency injection | `user-service/internal/wire.go`, `user-service/internal/wire_gen.go` |
| HTTP API | `user-service/internal/api/http/` |
| gRPC API | `user-service/internal/api/grpc/` |
| Usecases | `user-service/internal/usecase/` |
| Repositories | `user-service/internal/repository/` |
| Domain models and entities | `user-service/internal/domain/` |
| Migrations | `user-service/migrations/` |
| Config | `user-service/config/<env>/config.yaml` |
| Docker | `user-service/Dockerfile` |
| Commands | `user-service/Makefile` |

Primary responsibilities:

- user lifecycle and account status
- OAuth identity flows
- credentials and refresh tokens
- JWT signing
- JWK key lookup and public key provider gRPC endpoint

## chat-service

| Area | Path |
| --- | --- |
| Entrypoint | `chat-service/main.go` |
| Dependency injection | `chat-service/internal/wire.go`, `chat-service/internal/wire_gen.go` |
| HTTP API | `chat-service/internal/api/http/` |
| Usecases | `chat-service/internal/usecase/` |
| Repositories | `chat-service/internal/repository/` |
| External gRPC clients | `chat-service/internal/repository/external/user_grpc/` |
| Domain models and entities | `chat-service/internal/domain/` |
| Migrations | `chat-service/migrations/` |
| Config | `chat-service/config/<env>/config.yaml` |
| Commands | `chat-service/Makefile` |

Primary responsibilities:

- conversation creation
- conversation members
- JWT verification using public keys from user-service
- user status checks through user-service gRPC before creating conversations

## payment-gateway

| Area | Path |
| --- | --- |
| Entrypoint | `payment-gateway/main.go` |
| HTTP API | `payment-gateway/internal/api/http/` |
| Usecases | `payment-gateway/internal/usecase/` |
| Provider adapters | `payment-gateway/internal/repository/external/` |
| Domain models and entities | `payment-gateway/internal/domain/` |
| Migrations | `shared/migrations/017_create_payment_gateway_core.*.sql` |
| Config | `payment-gateway/config/<env>/config.yaml` |
| Commands | `payment-gateway/Makefile` |

Primary responsibilities:

- one VND wallet per user, linked only by `user_id`
- immutable double-entry ledger and materialized wallet balance projection
- SePay Sandbox hosted-checkout top-up initiation
- future signed IPN processing, provider reconciliation, and system notifications

## shared

| Area | Path |
| --- | --- |
| Go module | `shared/go.mod` |
| user-service JWK proto | `shared/grpc/user-service/jwk/proto/jwk.proto` |
| user-service JWK generated code | `shared/grpc/user-service/jwk/gen/` |
| user status proto | `shared/grpc/user-service/user_status/proto/user_status.proto` |
| user status generated code | `shared/grpc/user-service/user_status/gen/` |

Use `shared` for contracts or resources that multiple services import. Avoid putting service business logic here.

## Common Change Areas

| Task | Start Reading |
| --- | --- |
| Add or change HTTP endpoint | Service route, handler, request/response model, usecase, repository interface |
| Add or change database schema | Service migrations, entity, repository query/command |
| Add or change gRPC contract | `docs/grpc-contracts.md`, `shared/grpc/.../proto`, server registration, client usage |
| Change JWT verification | `user-service/internal/api/grpc/`, `chat-service/internal/api/http/middleware/`, `chat-service/internal/repository/external/user_grpc/` |
| Change local deployment | `docs/dev-commands.md`, root `Makefile`, `infra/`, `argocd/` |
