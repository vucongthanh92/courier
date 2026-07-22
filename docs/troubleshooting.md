# Troubleshooting

This document records recurring local development issues and the known fixes.

## `make migrate-file`: `migrate: No such file or directory`

Cause: the `migrate` CLI is not installed or is not on `PATH`.

Fix:

```bash
cd user-service
make install
```

Or install `golang-migrate` manually and ensure the binary is on `PATH`.

## `make swagger`: `swag: No such file or directory`

Cause: `swag` is not installed or not on `PATH`.

Fix:

```bash
cd user-service
make install
```

## `make wire`: `wire: No such file or directory`

Cause: `wire` is not installed or not on `PATH`.

Fix:

```bash
go install github.com/google/wire/cmd/wire@latest
```

If using older Go versions or project-specific constraints, install the version expected by the service.

## `make proto-gen`: `protoc: command not found`

Cause: the Protocol Buffers compiler is not installed.

Fix on macOS:

```bash
brew install protobuf
```

Also ensure these plugins are installed:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Protobuf generated package panic at startup

Symptom: startup panic in `google.golang.org/protobuf/internal/filedesc` or generated `*.pb.go`.

Common cause: generated files in `shared` and vendored copies inside services are out of sync.

Fix:

1. Regenerate the proto contract in `shared`.
2. Run `go mod vendor` in every service importing that shared contract.
3. Build the affected services.

## `cannot find module providing package ... import lookup disabled by -mod=vendor`

Cause: service is building in vendor mode, but `vendor/` does not contain the imported package.

Fix:

```bash
go mod tidy
go mod vendor
```

For shared contracts, make sure the service `go.mod` has:

```go
require github.com/vucongthanh92/courier/shared v0.0.0

replace github.com/vucongthanh92/courier/shared => ../shared
```

## Git `index.lock` blocks `git add`

Symptom:

```text
fatal: Unable to create '.git/index.lock': File exists.
```

Cause: a Git process was interrupted or another tool is refreshing Git state.

Fix if no Git process is running:

```bash
rm -f .git/index.lock
```

If it returns immediately, check IDE Git integration or another terminal running Git commands.

## Migration dirty database version

Symptom:

```text
error: Dirty database version <n>. Fix and force version.
```

Cause: a migration failed partway through and left the migration state dirty.

Fix after correcting the failed SQL:

```bash
migrate \
  -path migrations \
  -database "postgresql://dev:dev@127.0.0.1:5432/dev_db?sslmode=disable" \
  force <version>
```

Then rerun:

```bash
migrate \
  -path migrations \
  -database "postgresql://dev:dev@127.0.0.1:5432/dev_db?sslmode=disable" \
  up
```

## Go build cache permission issue

Symptom:

```text
open ~/Library/Caches/go-build/...: operation not permitted
```

Fix:

```bash
GOCACHE=/Users/vucongthanh/Documents/workspace/golang/courier/.cache/go-build go build ./...
```

## GitHub CLI works in terminal but not in Codex sandbox

Cause: Codex sandbox may not access macOS keyring the same way as the normal terminal.

Fix: run `gh` actions requiring keyring access outside the sandbox with approval. Inside the sandbox, `gh auth status` may report an invalid token even when the user terminal is authenticated.
