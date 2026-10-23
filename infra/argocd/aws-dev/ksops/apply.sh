#!/usr/bin/env bash

set -euo pipefail

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

kubectl apply -f "${SCRIPT_DIR}/plugin-configmap.yaml"
kubectl patch deployment argocd-repo-server \
  --namespace argocd \
  --type strategic \
  --patch-file "${SCRIPT_DIR}/repo-server-ksops-patch.yaml"
kubectl rollout status deployment/argocd-repo-server --namespace argocd --timeout=5m
