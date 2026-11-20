#!/usr/bin/env bash
set -euo pipefail

cid=$(docker run -d --rm -p 127.0.0.1::5432 -e POSTGRES_USER=ddd -e POSTGRES_PASSWORD=ddd postgres:18)
trap 'docker rm -f "$cid" >/dev/null' EXIT

for _ in {1..60}; do docker exec "$cid" pg_isready -qh 127.0.0.1 && break; sleep 0.5; done

port=$(docker port "$cid" 5432); port=${port##*:}
DATABASE_URL="postgres://ddd:ddd@127.0.0.1:$port/ddd?sslmode=disable" go run ./cmd/server migrate

# Drop lines that change between runs (the \restrict key and the server version).
docker exec "$cid" pg_dump -U ddd --schema-only --no-owner --no-privileges \
  | sed -E '/^\\(un)?restrict /d; /^-- Dumped (from|by) /d' > db/schema.sql
