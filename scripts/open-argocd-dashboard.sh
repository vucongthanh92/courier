#!/usr/bin/env bash
# Opens the private Courier Argo CD dashboard on localhost through AWS SSM.
#
# Requirements: AWS CLI credentials for the Courier account and the
# Session Manager plugin. The script creates/repairs the remote systemd
# port-forward service before it starts the local SSM tunnel.

set -euo pipefail

AWS_PROFILE="${AWS_PROFILE:-courier-terraform}"
AWS_REGION="${AWS_REGION:-ap-southeast-2}"
INSTANCE_ID="${INSTANCE_ID:-i-06109722166dede79}"
LOCAL_PORT="${LOCAL_PORT:-8080}"
REMOTE_PORT=8080
SERVICE_NAME="argocd-dashboard-port-forward.service"

for command in aws session-manager-plugin base64; do
  if ! command -v "$command" >/dev/null 2>&1; then
    echo "Missing required command: $command" >&2
    exit 1
  fi
done

aws sts get-caller-identity \
  --profile "$AWS_PROFILE" \
  --region "$AWS_REGION" >/dev/null

systemd_unit='[Unit]
Description=Courier Argo CD dashboard port forward
After=k3s.service
Requires=k3s.service

[Service]
Type=simple
ExecStart=/usr/local/bin/kubectl --kubeconfig /etc/rancher/k3s/k3s.yaml -n argocd port-forward service/argocd-server 8080:443 --address 127.0.0.1
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
'

unit_base64="$(printf '%s' "$systemd_unit" | base64 | tr -d '\n')"
remote_command="sudo systemctl stop $SERVICE_NAME 2>/dev/null || true; echo $unit_base64 | base64 -d | sudo tee /etc/systemd/system/$SERVICE_NAME >/dev/null; sudo systemctl daemon-reload; sudo systemctl enable --now $SERVICE_NAME; sudo systemctl is-active $SERVICE_NAME; for attempt in \$(seq 1 20); do if sudo ss -ltn | grep -F 127.0.0.1:$REMOTE_PORT; then exit 0; fi; sleep 1; done; sudo systemctl status $SERVICE_NAME --no-pager; exit 1"
remote_command_base64="$(printf '%s' "$remote_command" | base64 | tr -d '\n')"

echo "Configuring the private Argo CD port forward on $INSTANCE_ID..."
command_id="$(aws ssm send-command \
  --instance-ids "$INSTANCE_ID" \
  --document-name AWS-RunShellScript \
  --parameters "commands=echo $remote_command_base64 | base64 -d | bash" \
  --region "$AWS_REGION" \
  --profile "$AWS_PROFILE" \
  --query 'Command.CommandId' \
  --output text)"

if ! aws ssm wait command-executed \
  --command-id "$command_id" \
  --instance-id "$INSTANCE_ID" \
  --region "$AWS_REGION" \
  --profile "$AWS_PROFILE"; then
  aws ssm get-command-invocation \
    --command-id "$command_id" \
    --instance-id "$INSTANCE_ID" \
    --region "$AWS_REGION" \
    --profile "$AWS_PROFILE" \
    --query '{status:Status,stdout:StandardOutputContent,stderr:StandardErrorContent}' \
    --output json >&2 || true
  exit 1
fi

echo "Argo CD is ready. Open https://localhost:$LOCAL_PORT"

if [[ "${CONFIGURE_ONLY:-0}" == "1" ]]; then
  echo "CONFIGURE_ONLY=1: remote service verified; no local tunnel was opened."
  exit 0
fi

echo "Keep this terminal open; press Ctrl+C to close the local tunnel."

exec aws ssm start-session \
  --target "$INSTANCE_ID" \
  --document-name AWS-StartPortForwardingSession \
  --parameters "{\"portNumber\":[\"$REMOTE_PORT\"],\"localPortNumber\":[\"$LOCAL_PORT\"]}" \
  --region "$AWS_REGION" \
  --profile "$AWS_PROFILE"
