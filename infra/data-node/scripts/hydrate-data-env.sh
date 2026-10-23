#!/bin/bash
set -euo pipefail

readonly parameter_prefix="/courier/dev/data"
readonly destination="/etc/courier/data.env"

get_parameter() {
  aws ssm get-parameter \
    --region ap-southeast-2 \
    --name "${parameter_prefix}/$1" \
    --with-decryption \
    --query 'Parameter.Value' \
    --output text
}

token=$(curl --fail --silent --show-error --request PUT \
  --header 'X-aws-ec2-metadata-token-ttl-seconds: 21600' \
  http://169.254.169.254/latest/api/token)
data_private_ip=$(curl --fail --silent --show-error \
  --header "X-aws-ec2-metadata-token: ${token}" \
  http://169.254.169.254/latest/meta-data/local-ipv4)

install --owner root --group root --mode 0600 /dev/null "${destination}"
{
  printf 'POSTGRES_ADMIN_PASSWORD=%q\n' "$(get_parameter postgres-admin-password)"
  printf 'COURIER_APP_PASSWORD=%q\n' "$(get_parameter courier-app-password)"
  printf 'REDIS_PASSWORD=%q\n' "$(get_parameter redis-password)"
  printf 'QDRANT_API_KEY=%q\n' "$(get_parameter qdrant-api-key)"
  printf 'DATA_PRIVATE_IP=%q\n' "${data_private_ip}"
} >"${destination}"
