# user-service Kubernetes configuration

The Deployment expects a manually provisioned Kubernetes Secret named
`user-service-config` in the target namespace. It must contain one key:

```text
config.yaml
```

The key's value is the complete runtime configuration mounted at
`/app/config/dev/config.yaml`. Do not commit its content: it includes database,
Redis, OAuth, and other sensitive values. The non-secret `APP_ENV=dev` setting
remains in the ConfigMap.

The health server exposes `/live` and `/ready` on port `7001`. `/live` checks
only process health; `/ready` also checks database connectivity.
