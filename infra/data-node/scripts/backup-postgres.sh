#!/bin/bash
set -euo pipefail

readonly environment_file="/etc/courier/data.env"
readonly target_file="/etc/courier/backup-target"

test -r "${environment_file}"
test -r "${target_file}"

set -a
source "${environment_file}"
set +a

backup_target=$(<"${target_file}")
timestamp=$(date --utc +%Y-%m-%dT%H-%M-%SZ)
temporary_file=$(mktemp /data/backups/postgres-${timestamp}.sql.gz.XXXXXX)
trap 'rm -f "${temporary_file}"' EXIT

umask 077
docker exec -e "PGPASSWORD=${POSTGRES_ADMIN_PASSWORD}" courier-postgres \
  pg_dumpall --username courier_admin --clean --if-exists | gzip --stdout >"${temporary_file}"

aws s3 cp "${temporary_file}" "${backup_target}/postgres/${timestamp}.sql.gz" \
  --sse AES256 \
  --only-show-errors
