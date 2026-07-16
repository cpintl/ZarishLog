#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo "Checking service health..."

echo -n "Postgres: " && docker compose ps postgres --quiet >/dev/null 2>&1 && docker compose exec -T postgres pg_isready -U zarishlog >/dev/null 2>&1 && echo "OK" || echo "DOWN"

echo -n "API (http://localhost:8080): "
if curl -sS --max-time 2 http://localhost:8080/api/v1/health >/dev/null 2>&1; then
  echo "OK"
else
  echo "DOWN"
fi

echo -n "Web (http://localhost:3000): "
if curl -sS --max-time 2 http://localhost:3000 >/dev/null 2>&1; then
  echo "OK"
else
  echo "DOWN"
fi

echo -n "MinIO Console (http://localhost:9001): " && curl -sS --max-time 2 http://localhost:9001 >/dev/null 2>&1 && echo "OK" || echo "DOWN"
echo -n "Meilisearch (http://localhost:7700): " && curl -sS --max-time 2 http://localhost:7700 >/dev/null 2>&1 && echo "OK" || echo "DOWN"
