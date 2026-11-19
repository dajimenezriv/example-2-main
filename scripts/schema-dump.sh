#!/usr/bin/env bash
# Regenerates db/schema.sql: applies migrations/ to a throwaway Postgres and dumps the schema.
# Throwaway DB so the result never depends on the state of the local dev DB.
# Applied with `server migrate`, the same runner the deploy pipeline and the test DB use.
set -euo pipefail

PG_IMAGE="${PG_IMAGE:-postgres:16}" # keep in sync with docker-compose.yaml
OUT="${1:-db/schema.sql}"

cid="$(docker run -d --rm -p 127.0.0.1::5432 -e POSTGRES_USER=ddd -e POSTGRES_PASSWORD=ddd -e POSTGRES_DB=ddd "$PG_IMAGE")"
trap 'docker rm -f "$cid" >/dev/null' EXIT

# TCP check: the image's init phase only listens on the unix socket.
for _ in $(seq 60); do
  docker exec "$cid" pg_isready -q -h 127.0.0.1 -U ddd && break
  sleep 0.5
done
docker exec "$cid" pg_isready -q -h 127.0.0.1 -U ddd

port="$(docker port "$cid" 5432/tcp | head -1 | sed 's/.*://')"
# Exported env wins over .env (godotenv doesn't override), so the local DB is never touched.
DATABASE_URL="postgres://ddd:ddd@127.0.0.1:${port}/ddd?sslmode=disable" go run ./cmd/server migrate

# Strip lines that change between runs/patch versions (random \restrict key, server version).
docker exec "$cid" pg_dump -U ddd -d ddd --schema-only --no-owner --no-privileges \
  | sed -E '/^\\(un)?restrict /d; /^-- Dumped (from|by) /d' >"$OUT"

echo "Wrote $OUT"
