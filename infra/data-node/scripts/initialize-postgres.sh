#!/bin/sh
set -eu

: "${COURIER_APP_USER:?COURIER_APP_USER is required}"
: "${COURIER_APP_PASSWORD:?COURIER_APP_PASSWORD is required}"
: "${COURIER_DATABASES:?COURIER_DATABASES is required}"

psql --username "$POSTGRES_USER" --dbname postgres \
  --set=app_user="$COURIER_APP_USER" \
  --set=app_password="$COURIER_APP_PASSWORD" <<'SQL'
SELECT format('CREATE ROLE %I LOGIN PASSWORD %L', :'app_user', :'app_password')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'app_user')
\gexec
SQL

old_ifs=$IFS
IFS=','
set -- $COURIER_DATABASES
IFS=$old_ifs

for database in "$@"; do
  psql --username "$POSTGRES_USER" --dbname postgres \
    --set=app_user="$COURIER_APP_USER" \
    --set=database="$database" <<'SQL'
SELECT format('CREATE DATABASE %I OWNER %I', :'database', :'app_user')
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = :'database')
\gexec
SQL
done
