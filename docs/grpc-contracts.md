# gRPC Contracts

The courier project keeps cross-service gRPC contracts in the `shared` module. Services should import generated code from `shared` instead of copying generated files between service modules.

## Current Contracts

| Contract | Proto | Go package | Server | Client |
| --- | --- | --- | --- | --- |
| JWK public key provider | `shared/grpc/user-service/jwk/proto/jwk.proto` | `github.com/vucongthanh92/courier/shared/grpc/user-service/jwk/gen` | `user-service` | `chat-service` |
| User status check | `shared/grpc/user-service/user_status/proto/user_status.proto` | `github.com/vucongthanh92/courier/shared/grpc/user-service/user_status/gen` | `user-service` | `chat-service` |

## Package Convention

Each proto should set `go_package` to the shared generated path.

Example:

```proto
option go_package = "github.com/vucongthanh92/courier/shared/grpc/user-service/user_status/gen;userstatuspb";
```

The value before `;` is the Go import path. The value after `;` is the Go package name used inside generated files.

## Generate Flow

1. Edit the proto under `shared/grpc/<service>/<contract>/proto/`.
2. Generate Go code into `shared/grpc/<service>/<contract>/gen/`.
3. Run `go mod vendor` in each service that imports the shared contract.
4. Regenerate Wire when constructor dependencies change.
5. Build affected services.

Example for `user_status`:

```bash
cd shared
mkdir -p grpc/user-service/user_status/gen
PATH=$HOME/go/bin:$PATH protoc \
  --go_out=grpc/user-service/user_status/gen \
  --go_opt=paths=source_relative \
  --go-grpc_out=grpc/user-service/user_status/gen \
  --go-grpc_opt=paths=source_relative,require_unimplemented_servers=false \
  -I grpc/user-service/user_status/proto \
  grpc/user-service/user_status/proto/user_status.proto
```

Then sync consumers:

```bash
cd ../user-service
go mod vendor

cd ../chat-service
go mod vendor
```

## Server Pattern

`user-service` registers gRPC services in:

- `user-service/internal/api/grpc/grpc_server.go`

The current service methods are implemented in:

- `user-service/internal/api/grpc/grpc_usecase.go`

This file can host multiple gRPC usecase methods when the logic is small and closely related to user-service as a provider.

## Client Pattern

`chat-service` calls user-service gRPC through:

- `chat-service/internal/repository/external/user_grpc/client.go`

The client reuses one connection to user-service and creates typed clients for each gRPC service contract.

## Rules

- Do not expose private keys through gRPC.
- Do not query another service database directly when a gRPC contract exists.
- Keep proto contracts backward compatible when possible.
- Regenerate generated files from proto. Do not manually edit generated files.
- After changing shared contracts, vendor sync the services that import them.
