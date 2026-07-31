#!/usr/bin/env bash
set -euo pipefail

BOOTSTRAP_SERVER="${BOOTSTRAP_SERVER:-kafka:29092}"
TOPICS_FILE="${TOPICS_FILE:-/etc/kafka/topics.json}"
KAFKA_TOPICS_BIN="${KAFKA_TOPICS_BIN:-/opt/kafka/bin/kafka-topics.sh}"

if [[ ! -f "${TOPICS_FILE}" ]]; then
  echo "topics file not found: ${TOPICS_FILE}" >&2
  exit 1
fi

awk '
  /"name"[[:space:]]*:/ {
    name = $0
    sub(/^.*"name"[[:space:]]*:[[:space:]]*"/, "", name)
    sub(/".*$/, "", name)
  }
  /"partitions"[[:space:]]*:/ {
    partitions = $0
    sub(/^.*"partitions"[[:space:]]*:[[:space:]]*/, "", partitions)
    sub(/[^0-9].*$/, "", partitions)
  }
  /"replicationFactor"[[:space:]]*:/ {
    replication = $0
    sub(/^.*"replicationFactor"[[:space:]]*:[[:space:]]*/, "", replication)
    sub(/[^0-9].*$/, "", replication)
  }
  name != "" && partitions != "" && replication != "" {
    print name ":" partitions ":" replication
    name = ""
    partitions = ""
    replication = ""
  }
' "${TOPICS_FILE}" | while IFS=":" read -r topic partitions replication_factor; do
  if [[ -z "${topic}" || -z "${partitions}" || -z "${replication_factor}" ]]; then
    echo "invalid topic definition in ${TOPICS_FILE}" >&2
    exit 1
  fi

  "${KAFKA_TOPICS_BIN}" \
    --bootstrap-server "${BOOTSTRAP_SERVER}" \
    --create \
    --if-not-exists \
    --topic "${topic}" \
    --partitions "${partitions}" \
    --replication-factor "${replication_factor}"
done
