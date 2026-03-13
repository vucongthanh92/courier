# Repository Guidelines

## Project Structure & Module Organization
- Service code lives in `user-service/` (Go 1.24). Entry: `main.go`; DI wiring: `internal/wire.go`.
- Layers: `internal/api` (Gin handlers, Swagger annotations), `internal/usecase` (business rules), `internal/repository` (DB/Redis adapters), `internal/domain` (entities/DTOs), `internal/worker` (cron/async jobs), `startup/` (bootstrap), `helper/` (shared helpers).
- Configuration per environment in `config/<env>/config.yaml`. Database migrations in `migrations/`; SQL/data assets in `resources/` and `database/`. Docker and compose files sit at the `user-service/` root.

## Build, Test, and Development Commands
- `make install` — install local tooling (`migrate`, `swag`, `wire`, protoc plugins).
- `make run-local` — run API with `config/local/config.yaml`; escape-analysis logging stays enabled for tuning.
- `make compose` / `make compose-dev` — bring up the full stack for local or dev using the corresponding compose file.
- `make swagger` — regenerate OpenAPI docs. `make wire` — refresh DI code. `make proto-gen` — regenerate gRPC stubs.
- `make test` — run `go test -v ./...`; required before PRs.
- `make build-image` — build the `courier/user-service` container image.

## Coding Style & Naming Conventions
- Format with `gofmt`/`goimports`; tabs, no unused imports.
- Package/dir names lower_snake; exported Go names UpperCamel (e.g., `UserService`). Keep handler DTOs in `internal/domain`; pass `context.Context`.
- Prefer structured logging via `zap` with typed fields; validate incoming data using `validator/v10`.
- After adding providers or bindings, update `wire.go` and run `make wire`.

## Testing Guidelines
- Use Go `testing` plus `stretchr/testify` (vendored). Place `*_test.go` beside the code, favor table-driven cases.
- Mock repositories via interfaces; store deterministic fixtures in `testdata/` when they grow.
- Cover business logic and validation paths; add regression tests for bug fixes and run `make test` before pushes.

## Commit & Pull Request Guidelines
- Commit messages: short, present-tense summaries with optional scope prefix (`api: add signup validation`); keep under 72 chars.
- PRs should include: passing `make test`; regenerated Swagger if endpoints change; migrations when schemas change; concise description linking issues (`Fixes #123`); notes on config/env impacts; screenshots only when user-facing docs or flows change.

## Security & Configuration Tips
- Never commit secrets; load credentials via env vars referenced by `config/<env>/config.yaml` or compose files.
- Document new external dependencies and defaults in config; validate inbound data at handler boundaries.
