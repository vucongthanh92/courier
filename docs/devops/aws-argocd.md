# Argo CD on AWS development K3s

This runbook installs the non-HA Argo CD control plane on the Courier
Application EC2. It is intentionally private: `argocd-server` remains a
Kubernetes `ClusterIP` Service, and no AWS security-group rule is added.

The pinned version is `v3.5.6`. Upgrade it only by changing
`infra/argocd/aws-dev/kustomization.yaml` and `bootstrap.sh` in the same
reviewed pull request.

## Prerequisites

- Application EC2 is running and visible as `Online` in Systems Manager.
- AWS CLI uses the `courier-terraform` profile and `ap-southeast-2` region.
- K3s system pods are healthy.

Set the instance ID once for the commands below:

```bash
export AWS_PROFILE=courier-terraform
export AWS_REGION=ap-southeast-2
export APPLICATION_INSTANCE_ID=i-06109722166dede79
```

The instance ID is stable across stop/start cycles. Its public IP is not.

## Install or upgrade the pinned release

Run this through SSM instead of opening SSH. The command downloads the pinned
official manifest directly on the Application EC2 and applies CRDs server-side.

```bash
aws ssm send-command \
  --region "$AWS_REGION" \
  --instance-ids "$APPLICATION_INSTANCE_ID" \
  --document-name AWS-RunShellScript \
  --comment 'Install pinned Argo CD for Courier AWS dev' \
  --parameters 'commands=["curl --fail --location --show-error --silent https://raw.githubusercontent.com/argoproj/argo-cd/v3.5.6/manifests/install.yaml | sudo /usr/local/bin/kubectl apply --server-side --force-conflicts -n argocd -f -","sudo /usr/local/bin/kubectl rollout status deployment/argocd-server -n argocd --timeout=5m","sudo /usr/local/bin/kubectl rollout status deployment/argocd-repo-server -n argocd --timeout=5m","sudo /usr/local/bin/kubectl rollout status statefulset/argocd-application-controller -n argocd --timeout=5m","sudo /usr/local/bin/kubectl get pods -n argocd"]'
```

Record the returned command ID, then read its result:

```bash
aws ssm get-command-invocation \
  --region "$AWS_REGION" \
  --command-id '<command-id>' \
  --instance-id "$APPLICATION_INSTANCE_ID"
```

## Open the dashboard privately

First start a temporary loopback-only port-forward process on the node:

```bash
aws ssm send-command \
  --region "$AWS_REGION" \
  --instance-ids "$APPLICATION_INSTANCE_ID" \
  --document-name AWS-RunShellScript \
  --comment 'Create private temporary Argo CD port forward' \
  --parameters 'commands=["sudo pkill -f \"kubectl.*argocd-server.*8080:443\" || true","sudo nohup /usr/local/bin/kubectl port-forward --address 127.0.0.1 -n argocd svc/argocd-server 8080:443 >/var/log/argocd-port-forward.log 2>&1 &","sleep 2","sudo ss -ltnp | grep 127.0.0.1:8080"]'
```

In a separate local terminal, tunnel that loopback port over SSM:

```bash
aws ssm start-session \
  --target "$APPLICATION_INSTANCE_ID" \
  --document-name AWS-StartPortForwardingSession \
  --parameters '{"portNumber":["8080"],"localPortNumber":["8080"]}' \
  --region "$AWS_REGION"
```

Open <https://localhost:8080> and accept the initial self-signed certificate
warning. The username is `admin`.

Retrieve the one-time bootstrap password only in a trusted terminal. Treat the
command output as a secret; do not paste it into an issue, pull request, or
repository file:

```bash
aws ssm send-command \
  --region "$AWS_REGION" \
  --instance-ids "$APPLICATION_INSTANCE_ID" \
  --document-name AWS-RunShellScript \
  --comment 'Read Argo CD initial administrator password' \
  --parameters 'commands=["sudo /usr/local/bin/kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath=\"{.data.password}\" | base64 --decode; echo"]'
```

After the first login, rotate the `admin` password and configure repository
credentials before creating application resources. Do not give Argo CD an AWS
security-group ingress rule or expose its service through Traefik.

## Verify the private boundary

```bash
sudo /usr/local/bin/kubectl get svc -n argocd argocd-server
```

Expected service type is `ClusterIP`, never `LoadBalancer` or `NodePort`.
