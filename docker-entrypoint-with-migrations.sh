#!/usr/bin/env bash
set -e

: "${DB_HOST:=localhost}"
: "${DB_PORT:=5432}"
: "${DB_USER:=${POSTGRES_USER:-postgres}}"
: "${DB_PASS:=${POSTGRES_PASSWORD:-}}"
: "${DB_NAME:=${POSTGRES_DB:-postgres}}"

docker-entrypoint.sh postgres &

echo "⏳ Waiting for Postgres on ${DB_HOST}:${DB_PORT}..."
until pg_isready -U "${DB_USER}" -h "${DB_HOST}" -p "${DB_PORT}"; do
  sleep 1
done

echo "🚀 Running migrations on ${DB_HOST}:${DB_PORT}"
if ! migrator up; then
  echo "❌ Migrations failed!"
  exit 1
fi
echo "✅ Migrations applied"

wait
