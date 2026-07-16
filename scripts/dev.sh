#!/usr/bin/env bash
set -euo pipefail

echo "=== ZarishLog Development Environment ==="

# Check prerequisites
command -v go >/dev/null 2>&1 || { echo "Error: Go is not installed"; exit 1; }
command -v pnpm >/dev/null 2>&1 || { echo "Error: pnpm is not installed"; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "Error: Docker is not installed"; exit 1; }

echo "✓ Go $(go version | grep -oP 'go\S+')"
echo "✓ pnpm $(pnpm --version)"
echo "✓ Docker $(docker --version | grep -oP '\d+\.\d+\.\d+')"

# Start infrastructure
echo ""
echo "Starting infrastructure services..."
docker compose up -d postgres redis minio keycloak 2>/dev/null || docker-compose up -d postgres redis minio keycloak

# Wait for PostgreSQL
echo "Waiting for PostgreSQL..."
for i in {1..30}; do
  if pg_isready -h localhost -U zarishlog >/dev/null 2>&1; then
    echo "✓ PostgreSQL ready"
    break
  fi
  sleep 1
done

# Install dependencies
echo ""
echo "Installing dependencies..."
cd apps/web && pnpm install
cd ../..

# Run migrations
echo ""
echo "Running migrations..."
for f in packages/data-models/sql/migrations/*.sql; do
  echo "  Applying $f ..."
  psql -h localhost -U zarishlog -d zarishlog -f "$f" 2>/dev/null || \
    PGPASSWORD=zarishlog_dev_password psql -h localhost -U zarishlog -d zarishlog -f "$f"
done

# Seed data
echo ""
echo "Seeding data..."
psql -h localhost -U zarishlog -d zarishlog -f packages/data-models/sql/seed.sql 2>/dev/null || \
  PGPASSWORD=zarishlog_dev_password psql -h localhost -U zarishlog -d zarishlog -f packages/data-models/sql/seed.sql

echo ""
echo "=== Development environment ready ==="
echo "  API:  http://localhost:8080"
echo "  Web:  http://localhost:3000"
echo "  MinIO: http://localhost:9001"
echo "  Keycloak: http://localhost:8081"
echo ""
echo "Run: cd apps/api && go run ./cmd/api   (API server)"
echo "     cd apps/web && pnpm dev           (Web server)"
