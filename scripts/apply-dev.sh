#!/usr/bin/env bash
# Recompila el binario en local y lo inyecta en el contenedor app (sin rebuild de imagen).
set -euo pipefail

cd "$(dirname "$0")/.."

container="$(docker compose ps -q app)"
if [ -z "$container" ]; then
  echo "El contenedor app no está en marcha. Ejecuta: docker compose up -d" >&2
  exit 1
fi

CGO_ENABLED=0 go build -o /tmp/klowna-server ./cmd/server
docker cp /tmp/klowna-server "$container":/usr/local/bin/server
docker restart "$container"

echo "Cambios aplicados en el contenedor app."
