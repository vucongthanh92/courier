#!/usr/bin/env bash
set -euo pipefail

readonly bootstrap_server="${BOOTSTRAP_SERVER:-kafka:9092}"
readonly topics_file="${TOPICS_FILE:-/etc/courier/topics.json}"
readonly kafka_topics_bin="/opt/kafka/bin/kafka-topics.sh"

test -f "${topics_file}"

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
    replication_factor = $0
    sub(/^.*"replicationFactor"[[:space:]]*:[[:space:]]*/, "", replication_factor)
    sub(/[^0-9].*$/, "", replication_factor)
  }
  name != "" && partitions != "" && replication_factor != "" {
    print name ":" partitions ":" replication_factor
    name = ""
    partitions = ""
    replication_factor = ""
  }
' "${topics_file}" | while IFS=: read -r topic partitions replication_factor; do
  "${kafka_topics_bin}" \
    --bootstrap-server "${bootstrap_server}" \
    --create \
    --if-not-exists \
    --topic "${topic}" \
    --partitions "${partitions}" \
    --replication-factor "${replication_factor}"
done
