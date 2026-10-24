#!/usr/bin/env bash
# Installs the upstream Kubernetes AWS credential provider for K3s nodes.
# It uses the EC2 instance role to request short-lived ECR credentials; it
# never writes registry credentials or AWS keys to disk.

set -euo pipefail

ECR_CREDENTIAL_PROVIDER_VERSION="${ECR_CREDENTIAL_PROVIDER_VERSION:-v1.37.0}"
K3S_CREDENTIAL_PROVIDER_DIR="${K3S_CREDENTIAL_PROVIDER_DIR:-/var/lib/rancher/credentialprovider}"
K3S_BIN_DIR="$K3S_CREDENTIAL_PROVIDER_DIR/bin"
K3S_CONFIG="$K3S_CREDENTIAL_PROVIDER_DIR/config.yaml"

sudo dnf install -y golang
sudo install -d -m 0755 "$K3S_BIN_DIR"

sudo env GOBIN="$K3S_BIN_DIR" \
  go install "k8s.io/cloud-provider-aws/cmd/ecr-credential-provider@$ECR_CREDENTIAL_PROVIDER_VERSION"

sudo tee "$K3S_CONFIG" >/dev/null <<'EOF'
apiVersion: kubelet.config.k8s.io/v1
kind: CredentialProviderConfig
providers:
  - name: ecr-credential-provider
    matchImages:
      - "*.dkr.ecr.*.amazonaws.com"
      - "*.dkr.ecr.*.amazonaws.com.cn"
      - "*.dkr.ecr-fips.*.amazonaws.com"
    defaultCacheDuration: "12h"
    apiVersion: credentialprovider.kubelet.k8s.io/v1
EOF
sudo chmod 0644 "$K3S_CONFIG"

sudo systemctl restart k3s
until sudo /usr/local/bin/kubectl get node >/dev/null 2>&1; do
  sleep 5
done

sudo /usr/local/bin/kubectl get node
