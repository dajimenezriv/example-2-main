#!/usr/bin/env bash
# Regenerates db/schema.sql: applies migrations/*.up.sql to a throwaway Postgres and dumps the schema.
# Throwaway DB so the result never depends on the state of the local dev DB.
# Applied with psql, same as the test DB (WithInitScripts in adapters/db/main_test.go).
set -euo pipefail

PG_IMAGE="${PG_IMAGE:-postgres:16}" # keep in sync with docker-compose.yaml
OUT="${1:-db/schema.sql}"

cid="$(docker run -d --rm -e POSTGRES_USER=ddd -e POSTGRES_PASSWORD=ddd -e POSTGRES_DB=ddd "$PG_IMAGE")"
trap 'docker rm -f "$cid" >/dev/null' EXIT

# TCP check: the image's init phase only listens on the unix socket.
for _ in $(seq 60); do
  docker exec "$cid" pg_isready -q -h 127.0.0.1 -U ddd && break
  sleep 0.5
done
docker exec "$cid" pg_isready -q -h 127.0.0.1 -U ddd

for f in migrations/*.up.sql; do
  docker exec -i "$cid" psql -q -v ON_ERROR_STOP=1 -U ddd -d ddd <"$f"
done

# Strip lines that change between runs/patch versions (random \restrict key, server version).
docker exec "$cid" pg_dump -U ddd -d ddd --schema-only --no-owner --no-privileges \
  | sed -E '/^\\(un)?restrict /d; /^-- Dumped (from|by) /d' >"$OUT"

echo "Wrote $OUT"
