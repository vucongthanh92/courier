# User-service dev flow with kind + Argo CD

## Goal

- Build a Docker image for `user-service` on every merge to `main`.
- Push the image to a registry.
- Let Argo CD sync Kubernetes manifests from Git into the dev cluster.
- Use `kind` as the local Kubernetes cluster for development.

## Flow

1. Merge a change into `main`.
2. GitHub Actions runs `user-service` tests.
3. GitHub Actions builds the Docker image.
4. GitHub Actions pushes the image tag to the registry.
5. Argo CD watches the Git repo and syncs the `k8s/overlays/dev/user-service` path.
6. `kind` pulls or loads the image and rolls out the new deployment.

## Local cluster bootstrap

```bash
make kind-create
make kind-apply-argocd
make kind-apply-user-service
```

## Local image update

If you want to test the image locally without pushing it:

```bash
make kind-load-user-service
kubectl rollout restart deployment/user-service -n courier-dev
```
