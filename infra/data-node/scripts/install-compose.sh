#!/bin/bash
set -euo pipefail

readonly compose_version="v5.6.0"
readonly compose_asset="docker-compose-linux-aarch64"
readonly download_url="https://github.com/docker/compose/releases/download/${compose_version}"
readonly temporary_directory=$(mktemp -d)
trap 'rm -rf "${temporary_directory}"' EXIT

curl --fail --location --silent --show-error \
  "${download_url}/${compose_asset}" \
  --output "${temporary_directory}/${compose_asset}"
curl --fail --location --silent --show-error \
  "${download_url}/checksums.txt" \
  --output "${temporary_directory}/checksums.txt"

expected_checksum=$(awk -v asset="${compose_asset}" '$2 ~ ("\\*?" asset "$") { print $1 }' "${temporary_directory}/checksums.txt")
test -n "${expected_checksum}"
printf '%s  %s\n' "${expected_checksum}" "${temporary_directory}/${compose_asset}" | sha256sum --check --status

install --owner root --group root --mode 0755 \
  "${temporary_directory}/${compose_asset}" /usr/local/bin/docker-compose
/usr/local/bin/docker-compose version
