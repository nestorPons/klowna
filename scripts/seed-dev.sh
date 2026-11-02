#!/usr/bin/env bash
# Aplica scripts/dev-seed.sql contra la base de datos de Docker Compose.
set -euo pipefail

cd "$(dirname "$0")/.."

set -a
[ -f .env ] && . ./.env
set +a

DB_NAME="${DB_NAME:-klowna}"
DB_USER="${DB_USER:-klowna}"
DB_PASSWORD="${DB_PASSWORD:-klownapass}"

docker compose exec -T db mariadb -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < scripts/dev-seed.sql

echo "Seed de desarrollo aplicado en la base '${DB_NAME}'."
