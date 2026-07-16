#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Starting ZarishLog sandbox..."

cd "$ROOT_DIR"

if ! command -v docker &>/dev/null; then
  echo "Error: docker not found. Please install Docker and try again." >&2
  exit 1
fi

echo "Bringing up infrastructure with docker compose..."
docker compose up -d

echo "Waiting for PostgreSQL to report healthy status..."
container="$(docker compose ps -q postgres)"
if [[ -z "$container" ]]; then
  echo "Postgres container not found via docker compose ps -q postgres" >&2
  exit 1
fi

for i in {1..60}; do
  status="$(docker inspect --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$container" 2>/dev/null || true)"
  echo "  - postgres status: ${status:-unknown}"
  if [[ "$status" == "healthy" || "$status" == "running" ]]; then
    echo "Postgres is ready"
    break
  fi
  sleep 2
done

echo "Applying SQL migrations..."
for f in packages/data-models/sql/migrations/*.sql; do
  echo "  -> Applying $(basename "$f")"
  if command -v psql &>/dev/null; then
    psql -h localhost -U zarishlog -d zarishlog -f "$f" || PGPASSWORD=zarishlog_dev_password psql -h localhost -U zarishlog -d zarishlog -f "$f"
  else
    # Fallback: pipe file into psql inside postgres container
    docker compose exec -T postgres psql -U zarishlog -d zarishlog -f - < "$f" || true
  fi
done

echo "Seeding data..."
if [[ -f packages/data-models/sql/seed.sql ]]; then
  if command -v psql &>/dev/null; then
    psql -h localhost -U zarishlog -d zarishlog -f packages/data-models/sql/seed.sql || PGPASSWORD=zarishlog_dev_password psql -h localhost -U zarishlog -d zarishlog -f packages/data-models/sql/seed.sql
  else
    docker compose exec -T postgres psql -U zarishlog -d zarishlog -f - < packages/data-models/sql/seed.sql || true
  fi
fi

echo "Sandbox started. Services:"
echo "  API:    http://localhost:8080"
echo "  Web:    http://localhost:3000"
echo "  MinIO:  http://localhost:9001 (console)"
echo "  Keycloak: http://localhost:8080"
echo "  Meilisearch: http://localhost:7700"

echo "Done. Use 'make dev' or open VS Code tasks to run servers." 
