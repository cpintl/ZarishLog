# ZarishLog — Development Sandbox Setup Guide

> **Versions pinned:** Go 1.26.4 · Node.js 22.x LTS · pnpm 11.x · PostgreSQL 18.4 · Keycloak 26.6 · Redis 8 · MinIO latest · Meilisearch latest

---

## 0. Quick Start (3 commands)

```bash
git clone https://github.com/cpintl/zarishlog.git
cd zarishlog
bash scripts/zarishlog-setup.sh --yes
```

The bootstrap script auto-detects your machine, installs missing prerequisites, configures Git, pulls Docker images, installs VS Code extensions, and sets up the project environment.

---

## 1. Manual Prerequisites Installation

If you prefer to install tools manually or the bootstrap script doesn't support your OS, follow below:

### 1.1 Essential Tools

| Tool | Version | Install Command (Linux) | Install Command (macOS) |
|------|---------|------------------------|-------------------------|
| **Go** | `1.26.4` | [Download](https://go.dev/dl/go1.26.4.linux-amd64.tar.gz) + extract to `/usr/local/go` | `brew install go@1.26` |
| **Node.js** | `22.x LTS` | `curl -fsSL https://deb.nodesource.com/setup_22.x \| sudo -E bash - && sudo apt install -y nodejs` | `brew install node@22` |
| **pnpm** | `11.x` | `corepack enable && corepack prepare pnpm@11 --activate` | `corepack enable && corepack prepare pnpm@11 --activate` |
| **Docker** | Latest CE | [Docker Desktop for Linux](https://docs.docker.com/engine/install/) | [Docker Desktop for Mac](https://docs.docker.com/desktop/mac/install/) |
| **Docker Compose** | `v2.32+` | Included with Docker Desktop | Included with Docker Desktop |
| **psql** | `18` | `sudo apt install postgresql-client-18` | `brew install postgresql@18` |

### 1.2 Optional Go Tools

```bash
# golangci-lint (linter)
curl -fsSL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.64.2

# sqlc (type-safe SQL code generator)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0

# gofumpt (stricter Go formatter)
go install mvdan.cc/gofumpt@v0.7.0
```

---

## 2. Environment Setup

### 2.1 Configure Environment

```bash
cp .env.example .env
# Edit .env if needed (defaults work for local dev)
```

### 2.2 Start Infrastructure Services

```bash
make docker-up
# Or manually:
docker compose up -d
```

This starts:
| Service | Port | Purpose |
|---------|------|---------|
| **PostgreSQL 18** | `5432` | Primary database |
| **Redis 8** | `6379` | Cache + job queue |
| **MinIO** | `9000` (API), `9001` (Console) | Object storage |
| **Keycloak 26** | `8080` | Auth (OIDC/OAuth2) |
| **Meilisearch** | `7700` | Full-text search |

### 2.3 Run Database Migrations

```bash
make db-migrate
# This runs: psql -h localhost -U zarishlog -d zarishlog -f packages/data-models/sql/migrations/001_initial_schema.sql
```

### 2.4 Seed Master Data

```bash
make db-seed
# Loads: roles, permissions, UoMs, sample organization, products
```

### 2.5 Install Frontend Dependencies

```bash
cd apps/web && pnpm install && cd ../..
```

---

## 3. Start Development Servers

### Option A: Split Terminals (Recommended)

```bash
# Terminal 1 — Go API Server
cd apps/api && go run ./cmd/api
# → http://localhost:8080/api/v1/health

# Terminal 2 — Next.js Frontend
cd apps/web && pnpm dev
# → http://localhost:3000
```

### Option B: VS Code Tasks

Open VS Code, press `Ctrl+Shift+P`, select **Tasks: Run Task**, then choose:

- **Docker: Start All Services**
- **DB: Run Migrations**
- **Go: Test All**
- **Web: Dev Server**

### Option C: Make

```bash
make dev    # Starts both API and Web
```

---

## 4. Verify Setup

### 4.1 Health Check

```bash
curl http://localhost:8080/api/v1/health
# Expected: {"status":"healthy","db":"connected"}
```

### 4.2 List Products

```bash
curl http://localhost:8080/api/v1/products
# Expected: {"data":[...]}
```

### 4.3 Frontend

Visit http://localhost:3000 — you should see the ZarishLog landing page.
Visit http://localhost:3000/products — you should see the seeded product catalogue.

### 4.4 Run Tests

```bash
# Go tests
cd apps/api && go test ./... -v -race -count=1

# Frontend tests
cd apps/web && pnpm test

# Run all via make
make test
```

---

## 5. Access Management Consoles

| Service | URL | Credentials |
|---------|-----|-------------|
| **Keycloak Admin** | http://localhost:8080/admin | `admin` / `zarishlog_dev_password` |
| **MinIO Console** | http://localhost:9001 | `zarishlog` / `zarishlog_dev_password` |
| **Meilisearch** | http://localhost:7700 | Key: `zarishlog_search_key` |

---

## 6. Configuration Files

| File | Purpose |
|------|---------|
| `.env` | Environment variables (local overrides) |
| `.env.example` | Template with all variables documented |
| `docker-compose.yml` | Infrastructure service definitions |
| `apps/api/sqlc.yaml` | SQL code generation config |
| `Makefile` | Common development tasks |
| `go.work` | Go workspace (multi-module) |
| `.vscode/settings.json` | VS Code settings (Go, SQL, formatting) |
| `.vscode/tasks.json` | VS Code build/test/deploy tasks |
| `.vscode/launch.json` | Debug configurations |
| `.vscode/extensions.json` | Recommended extensions |

---

## 7. Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `psql: connection refused` | Docker not running | `make docker-up` |
| `go: command not found` | Go not in PATH | Add `export PATH=$PATH:/usr/local/go/bin` to `~/.profile` |
| `pnpm: command not found` | corepack not enabled | `corepack enable && corepack prepare pnpm@11 --activate` |
| Docker permission denied | User not in docker group | `sudo usermod -aG docker $USER && newgrp docker` |
| Go build fails | Missing dependencies | `cd apps/api && go mod tidy` |
| Database migration fails | PostgreSQL not ready | Wait 5s after `docker compose up -d` and retry |
| Port already in use | Conflict with existing service | Change port in `.env` and `docker-compose.yml` |
