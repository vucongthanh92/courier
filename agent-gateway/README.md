# agent-gateway

`agent-gateway` owns Courier's AI assistant integration. It prepares conversation context, manages vector memory in Qdrant, calls AI providers, and emits assistant responses back to `chat-service`.

`chat-service` remains the source of truth for conversation messages and websocket fanout.

See `docs/codegraph.md` for the package-level flow and target assistant event flow.

## Local Qdrant

Start Qdrant only:

```sh
make qdrant-up
```

Qdrant endpoints:

- REST: `http://localhost:6333`
- gRPC: `localhost:6334`
- Readiness: `http://localhost:6333/readyz`

Run agent-gateway against local Qdrant:

```sh
make run-local
```

The default local command reads:

```sh
./config/local/config.yaml
```

Healthcheck:

```sh
curl http://localhost:5010/healthz
```

Run Qdrant and agent-gateway together:

```sh
docker compose up --build
```

## Default Memory Collection

The service bootstraps this collection on startup:

- name: `courier_agent_memory`
- vector size: `1536`
- distance: `Cosine`

The default vector size matches `text-embedding-3-small`. If the embedding model changes, update `QDRANT_VECTOR_SIZE` to match the embedding dimensions.

## Helper Packages

Shared helper code follows the same shape as other Courier services:

- `helper/constants`: system constants such as service names, Kafka topics, Qdrant defaults, role names, and HTTP paths.
- `helper/utils`: small reusable helpers such as timeout context and correlation ID generation.

## Vendor Dependencies

Install or refresh vendored dependencies from inside `agent-gateway`:

```sh
cd agent-gateway
go mod tidy
go mod vendor
```

Use vendored dependencies explicitly when needed:

```sh
go test -mod=vendor ./...
go build -mod=vendor .
```
