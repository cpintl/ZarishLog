#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$ROOT_DIR"

echo "Resetting ZarishLog sandbox (this will remove volumes and local data)..."

docker compose down -v || true

echo "Removing local data directories..."
rm -rf data/postgres data/redis data/minio data/meilisearch || true
mkdir -p data/postgres data/redis data/minio data/meilisearch

echo "Bringing up minimal DB services..."
docker compose up -d postgres redis

echo "Waiting for PostgreSQL to be ready..."
container="$(docker compose ps -q postgres)"
for i in {1..60}; do
  status="$(docker inspect --format='{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$container" 2>/dev/null || true)"
  echo "  - postgres status: ${status:-unknown}"
  if [[ "$status" == "healthy" || "$status" == "running" ]]; then
    break
  fi
  sleep 2
done

echo "Re-applying migrations and seed..."
for f in packages/data-models/sql/migrations/*.sql; do
  echo "  -> Applying $(basename "$f")"
  docker compose exec -T postgres psql -U zarishlog -d zarishlog -f - < "$f" || true
done

if [[ -f packages/data-models/sql/seed.sql ]]; then
  docker compose exec -T postgres psql -U zarishlog -d zarishlog -f - < packages/data-models/sql/seed.sql || true
fi

echo "Reset complete. Use 'scripts/sandbox-start.sh' to start full stack services." 
