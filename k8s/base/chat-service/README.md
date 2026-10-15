# chat-service Kubernetes configuration

Create the `chat-service-config` Secret manually in the target namespace with
one `config.yaml` key. Its complete rendered configuration must use in-cluster
addresses such as `user-service:6001`, `redis:6379`, and `kafka:9092`; never
commit its values. The deployment mounts it at `/app/config/dev/config.yaml`.

The health server exposes `/live` and `/ready` on port `7002`. Readiness
requires both read and write database connections to respond.
