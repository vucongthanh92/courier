#!/usr/bin/env bash

set -euo pipefail

readonly ARGOCD_VERSION="v3.5.6"
readonly ARGOCD_MANIFEST_URL="https://raw.githubusercontent.com/argoproj/argo-cd/${ARGOCD_VERSION}/manifests/install.yaml"

kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
curl --fail --location --show-error --silent "${ARGOCD_MANIFEST_URL}" |
  kubectl apply --server-side --force-conflicts -n argocd -f -

for deployment in argocd-server argocd-repo-server; do
  kubectl rollout status "deployment/${deployment}" -n argocd --timeout=5m
done
kubectl rollout status statefulset/argocd-application-controller -n argocd --timeout=5m
