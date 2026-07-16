#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="$ROOT_DIR/.sandbox"
mkdir -p "$LOG_DIR"

compose_cmd() {
  if docker compose version >/dev/null 2>&1; then
    docker compose "$@"
  elif command -v docker-compose >/dev/null 2>&1; then
    docker-compose "$@"
  else
    echo "Docker Compose is required but was not found." >&2
    exit 1
  fi
}

ensure_env() {
  if [[ ! -f "$ROOT_DIR/.env" ]]; then
    if [[ -f "$ROOT_DIR/.env.example" ]]; then
      cp "$ROOT_DIR/.env.example" "$ROOT_DIR/.env"
    fi
  fi
}

wait_for_postgres() {
  echo "Waiting for PostgreSQL to accept connections..."
  local attempts=0
  while (( attempts < 60 )); do
    if compose_cmd exec -T postgres pg_isready -U zarishlog >/dev/null 2>&1; then
      echo "✓ PostgreSQL is ready"
      return 0
    fi
    attempts=$((attempts + 1))
    sleep 2
  done
  echo "PostgreSQL did not become ready in time." >&2
  return 1
}

run_sql_file() {
  local target_file="$1"
  local attempts=0
  while (( attempts < 3 )); do
    if PGPASSWORD=zarishlog_dev_password psql -h localhost -U zarishlog -d zarishlog -f "$target_file" >/dev/null 2>&1; then
      return 0
    fi
    attempts=$((attempts + 1))
    sleep 5
  done
  echo "Failed to apply $target_file" >&2
  return 1
}

run_migrations() {
  echo "Running database migrations..."
  if ! command -v psql >/dev/null 2>&1; then
    echo "psql is required to apply migrations." >&2
    return 1
  fi

  local migration_dir="$ROOT_DIR/packages/data-models/sql/migrations"
  for migration in "$migration_dir"/*.sql; do
    [[ -f "$migration" ]] || continue
    echo "  Applying $(basename "$migration")..."
    run_sql_file "$migration"
  done
  echo "✓ Migrations applied"
}

run_seed() {
  echo "Seeding master data..."
  if ! command -v psql >/dev/null 2>&1; then
    echo "psql is required to seed the database." >&2
    return 1
  fi

  local seed_file="$ROOT_DIR/packages/data-models/sql/seed.sql"
  run_sql_file "$seed_file"
  echo "✓ Seed data loaded"
}

install_web_dependencies() {
  if [[ -d "$ROOT_DIR/apps/web/node_modules" ]]; then
    return 0
  fi

  echo "Installing web dependencies..."
  (cd "$ROOT_DIR/apps/web" && pnpm install --frozen-lockfile)
}

wait_for_http() {
  local url="$1"
  local attempts=0
  while (( attempts < 30 )); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    attempts=$((attempts + 1))
    sleep 2
  done
  return 1
}

start_api() {
  local pid_file="$LOG_DIR/api.pid"
  if [[ -f "$pid_file" ]] && kill -0 "$(cat "$pid_file")" 2>/dev/null; then
    echo "API server already running"
    return 0
  fi

  echo "Starting API server..."
  (
    cd "$ROOT_DIR/apps/api"
    nohup go run -ldflags="-X main.Version=dev -X main.CommitHash=local -X main.BuildTime=$(date -u '+%Y-%m-%dT%H:%M:%SZ')" ./cmd/api >"$LOG_DIR/api.log" 2>&1 &
    echo $! >"$pid_file"
  )
}

start_web() {
  local pid_file="$LOG_DIR/web.pid"
  if [[ -f "$pid_file" ]] && kill -0 "$(cat "$pid_file")" 2>/dev/null; then
    echo "Web server already running"
    return 0
  fi

  echo "Starting web server..."
  (
    cd "$ROOT_DIR/apps/web"
    nohup pnpm dev >"$LOG_DIR/web.log" 2>&1 &
    echo $! >"$pid_file"
  )
}

start() {
  ensure_env
  echo "Starting sandbox services..."
  compose_cmd up -d postgres redis minio keycloak meilisearch
  wait_for_postgres
  run_migrations
  run_seed
  install_web_dependencies
  start_api
  start_web

  wait_for_http "http://localhost:8080/api/v1/health" >/dev/null 2>&1 || true
  wait_for_http "http://localhost:3000" >/dev/null 2>&1 || true

  echo ""
  echo "Sandbox is ready."
  echo "  API:       http://localhost:8080"
  echo "  Web:       http://localhost:3000"
  echo "  MinIO:     http://localhost:9001"
  echo "  Keycloak:  http://localhost:8080/admin"
  echo "  Meilisearch: http://localhost:7700"
}

stop() {
  local pid_file
  for pid_file in "$LOG_DIR/api.pid" "$LOG_DIR/web.pid"; do
    if [[ -f "$pid_file" ]]; then
      local pid
      pid="$(cat "$pid_file")"
      if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
        kill "$pid" 2>/dev/null || true
      fi
      rm -f "$pid_file"
    fi
  done
  compose_cmd down >/dev/null 2>&1 || true
  echo "Sandbox services stopped."
}

reset() {
  echo "Resetting sandbox data..."
  compose_cmd down -v >/dev/null 2>&1 || true
  rm -rf "$ROOT_DIR/data/postgres" "$ROOT_DIR/data/redis" "$ROOT_DIR/data/minio" "$ROOT_DIR/data/meilisearch" 2>/dev/null || true
  start
}

health() {
  echo "Container status:"
  compose_cmd ps
  echo ""
  echo "Service URLs:"
  echo "  API:       http://localhost:8080"
  echo "  Web:       http://localhost:3000"
  echo "  MinIO:     http://localhost:9001"
  echo "  Keycloak:  http://localhost:8080/admin"
  echo "  Meilisearch: http://localhost:7700"
}

status() {
  local api_pid="$(cat "$LOG_DIR/api.pid" 2>/dev/null || echo 'n/a')"
  local web_pid="$(cat "$LOG_DIR/web.pid" 2>/dev/null || echo 'n/a')"
  echo "API pid: $api_pid"
  echo "Web pid: $web_pid"
  echo "Postgres: $(compose_cmd exec -T postgres pg_isready -U zarishlog >/dev/null 2>&1 && echo ready || echo not-ready)"
}

case "${1:-start}" in
  start)
    start
    ;;
  stop)
    stop
    ;;
  reset)
    reset
    ;;
  health)
    health
    ;;
  status)
    status
    ;;
  migrate)
    ensure_env
    compose_cmd up -d postgres redis >/dev/null
    wait_for_postgres
    run_migrations
    ;;
  seed)
    ensure_env
    compose_cmd up -d postgres >/dev/null
    wait_for_postgres
    run_seed
    ;;
  *)
    echo "Usage: $0 {start|stop|reset|health|status|migrate|seed}" >&2
    exit 1
    ;;
esac
