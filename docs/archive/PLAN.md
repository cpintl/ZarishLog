# ZarishLog — Complete Setup, Development & DevOps Plan

> **For non-coders and vibe-coders.** Every section explains *what to click* or *what to type*, not how the code works underneath.

---

## Table of Contents

1. [What This Project Is](#1-what-this-project-is)
2. [Complete Project Inventory](#2-complete-project-inventory)
3. [Prerequisites — What You Need Installed](#3-prerequisites--what-you-need-installed)
4. [Environment Setup](#4-environment-setup)
5. [Starting the Platform (3 Ways)](#5-starting-the-platform-3-ways)
6. [Browser-Based Navigation — Every URL](#6-browser-based-navigation--every-url)
7. [VS Code — Full GUI DevOps Setup](#7-vs-code--full-gui-devops-setup)
8. [OpenCode — AI-Assisted DevOps](#8-opencode--ai-assisted-devops)
9. [Day-to-Day Development Workflows](#9-day-to-day-development-workflows)
10. [Validation & Testing](#10-validation--testing)
11. [Database Management](#11-database-management)
12. [Building for Production](#12-building-for-production)
13. [Troubleshooting](#13-troubleshooting)
14. [CI/CD — What Happens on GitHub](#14-cicd--what-happens-on-github)

---

## 1. What This Project Is

ZarishLog is an **open-source humanitarian logistics platform** that helps organizations manage warehouses, inventory, procurement, quality assurance, and asset tracking. Think of it as a free, offline-capable alternative to commercial ERP systems — built for field offices that may lose internet connectivity.

**The tech stack (you don't need to understand this, but it helps when Googling):**

| Layer | Technology | What it does |
|-------|-----------|--------------|
| Backend | Go + Gin | Handles API requests, business logic |
| Frontend | Next.js + React | The web interface you see in the browser |
| Database | PostgreSQL | Stores all data (products, warehouses, stock) |
| Search | Meilisearch | Fast product search (like Google for your catalogue) |
| Auth | Keycloak | Login, user accounts, roles & permissions |
| Storage | MinIO | Stores uploaded files (photos, documents) |
| Cache | Redis | Speeds up repeated queries |
| AI Tools | OpenCode | AI assistant that can read your code and database |

**The 5 services that run locally (all inside Docker containers):**

| Service | Port | GUI URL | Login |
|---------|------|---------|-------|
| PostgreSQL (database) | 5432 | VS Code SQLTools sidebar | `zarishlog` / `zarishlog_dev_password` |
| Redis (cache) | 6379 | None (background service) | N/A |
| MinIO (file storage) | 9000 / 9001 | http://localhost:9001 | `zarishlog` / `zarishlog_dev_password` |
| Keycloak (auth) | 8080 | http://localhost:8080/admin | `admin` / `zarishlog_dev_password` |
| Meilisearch (search) | 7700 | http://localhost:7700 | Key: `zarishlog_search_key` |

---

## 2. Complete Project Inventory

### File Counts by Type

| Extension | Count | Where |
|-----------|-------|-------|
| `.go` | 67 | `apps/api/` (backend) + `packages/business-logic/` |
| `.json` | 49 | Config files, templates, package manifests |
| `.csv` | 29 | Product catalogues, org data, import templates |
| `.sql` | 25 | Database migrations + queries + seeds |
| `.md` | 18 | Documentation |
| `.sh` | 12 | Scripts in `scripts/` |
| `.ts` / `.tsx` | 16 | Frontend code in `apps/web/` |
| `.yml` / `.yaml` | 5 | CI/CD, Docker, workspace config |
| `.tf` | 3 | Infrastructure-as-code (Terraform) |
| **Total** | **~247** | Excluding `.git/` and `node_modules/` |

### Directory Map

```
ZarishLog/                        (project root, 5.7 MB)
├── apps/                         (1.7 MB — your applications)
│   ├── api/                      (Go REST API — the backend)
│   │   ├── cmd/api/main.go       (entry point — where the server starts)
│   │   └── internal/             (all backend code: handlers, middleware, DB)
│   ├── web/                      (Next.js PWA — the frontend)
│   │   ├── app/                  (pages you see in the browser)
│   │   ├── components/           (reusable UI pieces)
│   │   ├── hooks/                (React utilities)
│   │   └── lib/                  (API client, offline/sync logic)
│   └── mobile/                   (Expo/React Native — scaffold only, no code yet)
├── packages/                     (1.2 MB — shared code)
│   ├── data-models/              (SQL migrations, queries, seeds)
│   │   ├── sql/migrations/       (6 migration files — database schema)
│   │   ├── sql/queries/          (16 query files — sqlc input)
│   │   └── sql/seed.sql          (seed data — sample products, orgs, etc.)
│   ├── business-logic/           (Go business rules: FEFO, AMC)
│   └── ui/                       (shared React components — placeholder)
├── config/                       (1.5 MB — configuration data)
│   ├── metadata/                 (CSV catalogues, org hierarchy)
│   ├── templates/                (29 JSON forms + 25 CSV templates)
│   ├── location/                 (warehouse.json)
│   └── reference_data/           (GLOSSARY.md)
├── infrastructure/               (92 KB — deployment configs)
│   ├── docker/                   (Dockerfiles + Keycloak realm)
│   └── terraform/                (AWS/cloud IaC)
├── scripts/                      (196 KB — 12 shell scripts)
├── docs/                         (156 KB — architecture, status, PRD)
├── data/                         (empty — runtime data dirs for Docker volumes)
├── .github/                      (CI/CD + issue templates)
├── .vscode/                      (VS Code settings, tasks, launch configs)
├── .opencode/                    (OpenCode AI tool config)
├── .githooks/                    (pre-commit hook)
├── Makefile                      (240 lines — developer entry point)
├── docker-compose.yml            (5 services)
├── go.work                       (Go workspace: api + business-logic)
├── package.json                  (pnpm workspace root)
└── pnpm-workspace.yaml           (defines: apps/web, apps/mobile, packages/ui)
```

### Key Config Files

| File | Purpose |
|------|---------|
| `Makefile` | All developer commands (`make dev`, `make test`, etc.) |
| `docker-compose.yml` | Defines the 5 Docker services |
| `.env` | Your local environment variables (gitignored) |
| `.env.example` | Template — copy to `.env` |
| `go.work` | Links Go modules: `apps/api` + `packages/business-logic` |
| `apps/api/sqlc.yaml` | Configures sqlc code generation from SQL |
| `apps/web/package.json` | Frontend dependencies and scripts |
| `pnpm-workspace.yaml` | Defines pnpm workspaces |
| `.prettierrc` | Code formatting rules |
| `.nvmrc` | Pins Node.js to version 22 |
| `tsconfig.json` | TypeScript compiler settings |

---

## 3. Prerequisites — What You Need Installed

### Option A: Automatic Install (Recommended)

Run one command and it installs everything:

```bash
make setup
```

This runs `scripts/zarishlog-setup.sh --yes` which detects your OS, installs missing tools, configures Git, pulls Docker images, and sets up VS Code extensions.

### Option B: Check What You Already Have

```bash
bash scripts/zarishlog-setup.sh --check-only
```

This prints a report of what's installed and what's missing — **no changes are made**.

### Option C: Manual Install Checklist

Check each item. If you don't have it, click the link to install:

- [ ] **Docker Desktop** (includes Docker Compose) — [Install](https://docs.docker.com/desktop/)
  - After install, make sure Docker is running (whale icon in system tray)
  - On Linux, add your user to the docker group: `sudo usermod -aG docker $USER && newgrp docker`

- [ ] **Go 1.26.4** — [Install](https://go.dev/dl/)
  - Linux: `curl -fsSL https://go.dev/dl/go1.26.4.linux-amd64.tar.gz | sudo tar -C /usr/local -xzf -`
  - macOS: `brew install go@1.26`
  - Verify: `go version`

- [ ] **Node.js 22.x LTS** — [Install](https://nodejs.org/)
  - Via pnpm: `corepack enable && corepack prepare pnpm@11 --activate`
  - Verify: `node --version` and `pnpm --version`

- [ ] **PostgreSQL client (psql)** — for database commands
  - Linux: `sudo apt install postgresql-client-18`
  - macOS: `brew install postgresql@18`
  - Verify: `psql --version`

- [ ] **VS Code** — [Install](https://code.visualstudio.com/)
  - After opening the project, VS Code will prompt you to install 22 recommended extensions — **click Install All**

- [ ] **OpenCode CLI** — for AI-assisted development
  - Follow instructions at [opencode.ai](https://opencode.ai)

### Post-Install: Copy Environment File

```bash
cp .env.example .env
```

The `.env` file contains all the passwords and settings for local development. It's already pre-configured with safe defaults. **Never commit this file to git** (it's already in `.gitignore`).

---

## 4. Environment Setup

### What's in `.env` (all pre-configured for local dev)

| Variable | Default Value | What It Does |
|----------|---------------|--------------|
| `DATABASE_URL` | `postgresql://zarishlog:zarishlog_dev_password@localhost:5432/zarishlog` | Database connection string |
| `DATABASE_PASSWORD` | `zarishlog_dev_password` | Database password |
| `REDIS_URL` | `redis://localhost:6379` | Cache connection |
| `S3_ENDPOINT` | `http://localhost:9000` | MinIO file storage |
| `S3_ACCESS_KEY` | `zarishlog` | MinIO login |
| `S3_SECRET_KEY` | `zarishlog_dev_password` | MinIO password |
| `OIDC_ISSUER` | `http://localhost:8080/realms/zarishlog` | Keycloak auth server |
| `API_PORT` | `8080` | Backend API port |
| `WEB_PORT` | `3000` | Frontend port |
| `MEILISEARCH_URL` | `http://localhost:7700` | Search engine URL |
| `MEILISEARCH_API_KEY` | `zarishlog_search_key` | Search engine key |
| `DEFAULT_TIMEZONE` | `Asia/Dhaka` | Default timezone for the app |

**You don't need to change anything** — the defaults work out of the box for local development.

---

## 5. Starting the Platform (3 Ways)

### Way 1: Easiest — GUI with Browser Tabs (Recommended for First-Timers)

```bash
bash scripts/sandbox-start-gui.sh
```

**What happens:**
1. Starts all 5 Docker containers (PostgreSQL, Redis, MinIO, Keycloak, Meilisearch)
2. Waits for the database to be ready
3. Runs database migrations (creates all tables)
4. Loads sample data (products, warehouses, users)
5. **Opens 4 browser tabs automatically** with all service UIs

**What you'll see in your browser:**
- Tab 1: http://localhost:3000 — ZarishLog web app
- Tab 2: http://localhost:8080 — Keycloak (will show a login page)
- Tab 3: http://localhost:9001 — MinIO Console (file storage)
- Tab 4: http://localhost:7700 — Meilisearch (search dashboard)

### Way 2: VS Code Tasks (No Terminal Needed)

1. Open the project in VS Code
2. Press `Ctrl+Shift+P` (or `Cmd+Shift+P` on Mac)
3. Type `Tasks: Run Task` and select it
4. Choose **"Sandbox: Start"**

**Or use the keyboard shortcut:** `Ctrl+Shift+B` (this is bound to Sandbox: Start)

**To stop:** Same steps, choose **"Sandbox: Stop"**

**To reset everything (delete all data and start fresh):** Choose **"Sandbox: Reset"**

### Way 3: Terminal Commands

```bash
# Start everything
make sandbox-start

# Check if everything is running
bash scripts/sandbox-health.sh

# Stop everything
make sandbox-stop

# Reset (delete all data and start fresh)
make sandbox-reset
```

### What "Start" Does Behind the Scenes

1. `docker compose up -d` — starts 5 containers
2. Waits for PostgreSQL to be healthy (up to 2 minutes)
3. Runs 6 SQL migration files in order (creates 76 database tables)
4. Loads seed data (sample products, orgs, roles, warehouses)
5. Starts the Go API server on port 8080
6. Starts the Next.js dev server on port 3000

---

## 6. Browser-Based Navigation — Every URL

### ZarishLog Web App — http://localhost:3000

The main application. This is where you interact with:
- Product catalogue
- Warehouse management
- Stock levels and movements
- Quality assurance inspections
- Asset tracking
- Distribution records

### API Health Check — http://localhost:8080/api/v1/health

Returns `{"status":"healthy","db":"connected"}` if the backend is running and connected to the database. Use this to verify the API is up.

### Keycloak Admin Console — http://localhost:8080/admin

**Login:** `admin` / `zarishlog_dev_password`

Manage users, roles, and permissions:
- View/edit user accounts
- Configure authentication settings
- Manage role assignments (R01 Global Admin through R12 Auditor)

### MinIO Console — http://localhost:9001

**Login:** `zarishlog` / `zarishlog_dev_password`

Manage file storage:
- Upload/download files (photos, scanned documents, certificates)
- Create/manage buckets
- Browse stored objects

### Meilisearch Dashboard — http://localhost:7700

**API Key:** `zarishlog_search_key`

Monitor the search engine:
- View indexed documents
- Test search queries
- Check index health

### VS Code SQLTools — Browse Database in Sidebar

No browser needed — this runs inside VS Code:
1. Open VS Code
2. Click the **SQLTools** icon in the left sidebar (database icon)
3. Expand **ZarishLog Local** connection
4. Browse tables, run queries, view data

**Pre-configured connection:**
- Host: `localhost:5432`
- Database: `zarishlog`
- Username: `zarishlog`
- Password: `zarishlog_dev_password`

### Quick Health Check (All Services)

```bash
bash scripts/sandbox-health.sh
```

Output shows OK or DOWN for each service:
- Postgres: OK
- API (http://localhost:8080): OK
- Web (http://localhost:3000): OK
- MinIO Console (http://localhost:9001): OK
- Meilisearch (http://localhost:7700): OK

---

## 7. VS Code — Full GUI DevOps Setup

### 22 Recommended Extensions

When you open the project, VS Code will prompt you to install these. **Click "Install All":**

| Extension | What It Does (Plain English) |
|-----------|------------------------------|
| **Go** (`golang.go`) | Syntax highlighting, autocomplete, and debugging for Go code |
| **Prettier** (`esbenp.prettier-vscode`) | Automatically formats your code when you save |
| **ESLint** (`dbaeumer.vscode-eslint`) | Finds and fixes JavaScript/TypeScript errors |
| **Tailwind CSS** (`bradlc.vscode-tailwindcss`) | Autocomplete for Tailwind CSS classes |
| **SQLTools** (`mtxr.sqltools`) | Browse and query the database from the sidebar |
| **SQLTools PostgreSQL** (`mtxr.sqltools-driver-pg`) | PostgreSQL driver for SQLTools |
| **Even Better TOML** (`tamasfe.even-better-toml`) | Syntax highlighting for TOML config files |
| **Better Comments** (`aaron-bond.better-comments`) | Color-codes comments (TODO, WARNING, etc.) |
| **Spell Checker** (`streetsidesoftware.code-spell-checker`) | Catches typos in code and comments |
| **GitLens** (`eamodio.gitlens`) | See who wrote each line of code and when |
| **Git Graph** (`mhutchie.git-graph`) | Visual git history graph |
| **Conventional Commits** (`vivaxy.vscode-conventional-commits`) | Helps write standardized commit messages |
| **Docker** (`ms-azuretools.vscode-docker`) | Manage Docker containers from VS Code |
| **YAML** (`redhat.vscode-yaml`) | Syntax highlighting for YAML files |
| **Markdown All in One** (`yzhang.markdown-all-in-one`) | Better markdown editing |
| **Markdown Mermaid** (`bierner.markdown-mermaid`) | Render diagrams in markdown |
| **Markdown Lint** (`davidanson.vscode-markdownlint`) | Check markdown formatting |
| **Material Icon Theme** (`pkief.material-icon-theme`) | Better file icons in the sidebar |
| **ENV** (`irongeek.vscode-env`) | Syntax highlighting for `.env` files |
| **GitHub Actions** (`github.vscode-github-actions`) | Syntax highlighting for CI/CD workflows |
| **GitHub Copilot** (`github.copilot`) | AI code completion |
| **GitHub Copilot Chat** (`github.copilot-chat`) | AI chat assistant in VS Code |

### 15 Tasks (Ctrl+Shift+P → Tasks: Run Task)

#### Sandbox Lifecycle (start/stop the whole platform)
| Task | What It Does |
|------|--------------|
| **Sandbox: Start** | Starts all services + API + Web servers (default build task) |
| **Sandbox: Stop** | Stops all services and servers |
| **Sandbox: Reset** | Deletes all data and starts fresh |
| **Sandbox: Health** | Checks if all services are running |
| **Sandbox: Open Services** | Prints all browser URLs |

#### Docker Lifecycle
| Task | What It Does |
|------|--------------|
| **Docker: Start All Services** | Starts Docker containers only (no API/Web) |
| **Docker: Stop All Services** | Stops Docker containers |
| **Docker: Restart PostgreSQL** | Restarts just the database |

#### Database
| Task | What It Does |
|------|--------------|
| **DB: Run Migrations** | Creates/updates database tables |
| **DB: Seed Data** | Loads sample products, orgs, users |

#### Go (Backend)
| Task | What It Does |
|------|--------------|
| **Go: Generate sqlc Queries** | Regenerates Go code from SQL queries |
| **Go: Lint** | Checks Go code for errors |
| **Go: Test All** | Runs all Go tests |

#### Web (Frontend)
| Task | What It Does |
|------|--------------|
| **Web: Dev Server** | Starts Next.js dev server |
| **Web: Lint** | Checks frontend code for errors |
| **Web: Typecheck** | Checks TypeScript types |

### 9 Debug Configurations (F5 or Run Menu)

| Config | What It Does |
|--------|--------------|
| **API Server (Go)** | Starts the Go API with a debugger — set breakpoints, step through code |
| **Web (Next.js)** | Starts the Next.js dev server |
| **Go Test (Package)** | Debugs tests in the currently open file |
| **Go Test All** | Debugs all Go tests |
| **Attach to Chrome** | Connects debugger to Chrome for frontend debugging |
| **Sandbox: Start** | Starts the sandbox from the debugger |
| **Sandbox: Stop** | Stops the sandbox from the debugger |
| **Sandbox: Reset** | Resets the sandbox from the debugger |

### Format-on-Save (Automatic)

When you save any file, VS Code automatically:
- **Go files:** Formats with `gofumpt` (stricter `gofmt`)
- **JS/TS/CSS/JSON files:** Formats with Prettier
- **SQL files:** Formats with SQLTools
- **All files:** Runs ESLint auto-fix for JS/TS

You don't need to do anything — just save and the code formats itself.

### MCP Server (GitHub Integration)

VS Code has a GitHub MCP server configured (`.vscode/mcp.json`) that enables AI tools to interact with GitHub:
- Read issues and pull requests
- Search code
- Create PRs

Requires a GitHub Personal Access Token (prompted on first use).

---

## 8. OpenCode — AI-Assisted DevOps

### What is OpenCode?

OpenCode is an AI coding assistant that can read your project, query your database, manage Docker containers, and interact with GitHub — all through natural language commands.

### How It Knows About Your Project

OpenCode reads two key files:

1. **`AGENTS.md`** (project root) — Tells the AI:
   - The project structure and architecture
   - What commands to run
   - Coding conventions and patterns
   - What NOT to do (guardrails)

2. **`.opencode/mcp.json`** — Defines 4 tool servers the AI can use

### The 4 MCP Tool Servers

| Tool | What the AI Can Do |
|------|-------------------|
| **PostgreSQL** | Query your database directly — "show me all products", "how many warehouses are there?" |
| **GitHub** | Read issues, create PRs, search code — "what issues are open?", "create a PR for this change" |
| **Docker** | Manage containers — "restart postgres", "show running containers" |
| **Filesystem** | Read and write project files — "read the main.go file", "create a new component" |

### Skills Available (30+)

OpenCode has access to specialized skills for web research and automation:

| Category | Skills |
|----------|--------|
| **Web Research** | firecrawl-search, firecrawl-scrape, firecrawl-deep-research |
| **Documentation** | firecrawl-knowledge-base, firecrawl-knowledge-ingest |
| **Competitive Intel** | firecrawl-competitive-intel, firecrawl-market-research |
| **Lead Generation** | firecrawl-lead-gen, firecrawl-lead-research |
| **SEO** | firecrawl-seo-audit |
| **QA Testing** | firecrawl-qa |
| **Website Design** | firecrawl-website-design-clone |
| **Monitoring** | firecrawl-monitor |
| **Shopping** | firecrawl-shop |

### Custom Commands

You can define custom slash commands in `.opencode/commands/`:

```markdown
---
description: Run all tests and show failures
agent: build
---

Run the full test suite. If any tests fail, show the failure messages and suggest fixes.
```

Save as `.opencode/commands/test.md` and run with `/test` in OpenCode.

### Plugins

The project uses `@opencode-ai/plugin` (v1.18.30) for custom tooling. Plugins are stored in `.opencode/plugins/` and can:
- Add custom tools
- Hook into shell execution
- Inject environment variables
- Protect sensitive files (like `.env`)

### Permissions

OpenCode can be configured to ask before certain actions:

```json
{
  "permission": {
    "edit": "ask",
    "bash": "ask"
  }
}
```

This means the AI will ask your approval before editing files or running commands.

---

## 9. Day-to-Day Development Workflows

### Making a Code Change

**The golden rule:** Read → Edit → Validate → Commit

1. **Read** the relevant code and docs first
2. **Edit** the file(s) — VS Code auto-formats on save
3. **Validate** your change (see Section 10)
4. **Commit** when tests pass

### Adding a New API Endpoint

```bash
# 1. Create handler file
touch apps/api/internal/handler/newmodule.go

# 2. Add model struct (in apps/api/internal/model/)

# 3. Add SQL query (in packages/data-models/sql/queries/newmodule.sql)

# 4. Regenerate Go code from SQL
cd apps/api && sqlc generate

# 5. Add routes (in apps/api/cmd/api/main.go)

# 6. Test
cd apps/api && go test ./...
```

**VS Code shortcut:** Ctrl+Shift+P → Tasks: Run Task → "Go: Generate sqlc Queries"

### Changing the Database Schema

```bash
# 1. Create migration file
#命名: packages/data-models/sql/migrations/YYYYMMDD_HHMMSS_description.sql

# 2. Apply migration
make db-migrate

# 3. Update SQL queries if needed

# 4. Regenerate Go code
cd apps/api && sqlc generate

# 5. Test
cd apps/api && go test ./...
```

### Changing the Frontend

1. Edit files in `apps/web/` (pages, components, hooks, lib)
2. Save — VS Code auto-formats
3. Check for errors: Ctrl+Shift+P → Tasks → "Web: Typecheck"
4. Run tests: Ctrl+Shift+P → Tasks → "Web: Typecheck" (or `pnpm test` in terminal)

### Working with the Database Directly

**GUI method (VS Code):**
1. Click the SQLTools icon in the left sidebar
2. Expand "ZarishLog Local"
3. Browse tables, double-click to view data
4. Click "New SQL" to write queries
5. Press Ctrl+Enter to run

**Terminal method:**
```bash
# Connect to database
psql -h localhost -U zarishlog -d zarishlog

# Or use the Makefile shortcut
make db-connect
```

---

## 10. Validation & Testing

### Quick Check (After Any Change)

```bash
make lint
```

This runs:
- `golangci-lint` (or `go vet` if golangci-lint isn't installed) for Go code
- `pnpm lint` for frontend code

### Full Validation (Before Committing)

```bash
make lint && make test
```

This runs all linters and all tests.

### What CI Runs (GitHub Actions Pipeline)

When you push to GitHub, two parallel jobs run:

**Backend Job (Go):**
1. Start PostgreSQL container
2. Run all 6 migrations
3. Seed database
4. `go vet ./...` — static analysis
5. `go test ./... -v -race` — all tests with race detector
6. `go build ./cmd/api` — verify it compiles

**Frontend Job (Next.js):**
1. Install dependencies (`pnpm install --frozen-lockfile`)
2. `pnpm lint` — ESLint
3. `pnpm typecheck` — TypeScript type checking
4. `pnpm test` — Vitest unit tests
5. `pnpm build` — production build

### Single Test Commands

**Run one Go test file:**
```bash
cd apps/api && go test -run TestFunctionName ./...
```

**Run tests for one Go package:**
```bash
cd apps/api && go test ./internal/handler/...
```

**Run Go tests without race detector (faster):**
```bash
cd apps/api && go test -short ./...
```

**Run frontend tests:**
```bash
cd apps/web && pnpm test
```

**Run frontend tests with verbose output:**
```bash
cd apps/web && pnpm test --reporter=verbose
```

**Generate Go coverage report:**
```bash
make test-coverage
# Opens: apps/api/coverage.html
```

### Validation Order (Recommended)

When making changes, validate in this order:

1. **Lint** — `make lint` (catches style issues)
2. **Typecheck** — `pnpm typecheck` (catches type errors)
3. **Test** — `make test` (catches logic errors)
4. **Build** — `make build` (catches compilation errors)

---

## 11. Database Management

### Browsing Data (GUI)

**VS Code SQLTools (Recommended):**
1. Click the database icon in the left sidebar
2. Expand "ZarishLog Local"
3. Browse 76 tables organized by domain:
   - Organization & Users (14 tables)
   - Master Data & Products (8 tables)
   - Procurement & Suppliers (4 tables)
   - Warehouse & Locations (4 tables)
   - Stock & Batches (4 tables)
   - Goods Receipt & Issue (9 tables)
   - Quality Assurance (5 tables)
   - Distribution (3 tables)
   - Returns & Disposal (5 tables)
   - Physical Count (3 tables)
   - Asset Management (5 tables)
   - Replenishment & Forecasting (3 tables)
   - Alerts & Notifications (3 tables)
   - Audit & Compliance (2 tables)
   - Offline Sync (2 tables)
   - Reports & Scheduling (2 tables)

**MinIO Console (File Storage):**
- URL: http://localhost:9001
- Login: `zarishlog` / `zarishlog_dev_password`
- Browse uploaded files, photos, documents

### Database Commands

```bash
# Apply all migrations (creates/updates tables)
make db-migrate

# Load sample data
make db-seed

# Full reset (drops everything and starts fresh)
make db-reset

# Connect to database in terminal
make db-connect
```

### Migration Naming Convention

New migration files must follow this pattern:
```
packages/data-models/sql/migrations/YYYYMMDD_HHMMSS_description.sql
```

Example: `20260910_143000_add_barcode_column.sql`

Migrations run in filename order (alphabetical), so the timestamp prefix ensures correct ordering.

### What's in the Database

| Category | Count | Examples |
|----------|-------|----------|
| Tables | 76 | products, warehouses, stock_levels, users, roles |
| RLS Policies | 39 | Row-level security on all tenant tables |
| Custom Enums | 8 | movement_type, stock_status, warehouse_type |
| Database Functions | 7 | uuid_generate_v7, app.current_org_id |
| Indexes | 52 | Performance indexes on key columns |

---

## 12. Building for Production

### Build Everything

```bash
make build
```

This builds:
1. Go API binary → `apps/api/bin/api`
2. Next.js frontend → `apps/web/.next/`

### Build Docker Images

```bash
make build-docker
```

Creates:
- `zarishlog-api:version` — Go API container
- `zarishlog-web:version` — Next.js frontend container

### Publish to Container Registry

```bash
make publish
```

Pushes Docker images to GitHub Container Registry (GHCR).

### Create a Release

```bash
make release
```

This:
1. Creates a git tag with the version
2. Pushes the tag to GitHub
3. Builds and publishes Docker images

### Build Script Options

```bash
# Build only Go
bash scripts/build.sh --go

# Build only frontend
bash scripts/build.sh --frontend

# Build Docker images
bash scripts/build.sh --docker

# Build and publish
bash scripts/build.sh --docker --publish --version v1.0.0
```

---

## 13. Troubleshooting

### Common Issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| `psql: connection refused` | Docker not running | Start Docker Desktop, then `make docker-up` |
| `go: command not found` | Go not in PATH | Add `export PATH=$PATH:/usr/local/go/bin` to `~/.profile` |
| `pnpm: command not found` | corepack not enabled | `corepack enable && corepack prepare pnpm@11 --activate` |
| Docker permission denied | User not in docker group | `sudo usermod -aG docker $USER && newgrp docker` |
| Go build fails | Missing dependencies | `cd apps/api && go mod tidy` |
| Database migration fails | PostgreSQL not ready | Wait 5s after `docker compose up -d` and retry |
| Port already in use | Another service on that port | Change port in `.env` and `docker-compose.yml` |
| `node_modules` missing | Dependencies not installed | `cd apps/web && pnpm install` |
| API returns 401 | Not authenticated | Most endpoints require JWT auth via Keycloak |
| Frontend shows blank page | API not running | Start API: `cd apps/api && go run ./cmd/api` |

### Nuclear Option: Reset Everything

```bash
make sandbox-reset
```

This:
1. Stops all Docker containers
2. Deletes all Docker volumes (database data, files)
3. Deletes local data directories
4. Starts fresh with migrations + seed data

### Check What's Running

```bash
# Quick health check
bash scripts/sandbox-health.sh

# See Docker containers
docker compose ps

# See running processes
ps aux | grep -E "(go run|pnpm dev|next)"
```

### View Logs

```bash
# All Docker services
docker compose logs -f

# Just PostgreSQL
docker compose logs -f postgres

# Just API (if running via sandbox.sh)
cat .sandbox/api.log

# Just Web (if running via sandbox.sh)
cat .sandbox/web.log
```

---

## 14. CI/CD — What Happens on GitHub

### When You Push or Create a PR

GitHub Actions automatically runs two parallel jobs:

```
Push/PR to main
    ├── Backend (Go) Job
    │   ├── Start PostgreSQL container
    │   ├── Run 6 migrations
    │   ├── Seed database
    │   ├── go vet ./...          (static analysis)
    │   ├── go test -race ./...   (all tests)
    │   └── go build ./cmd/api    (compile check)
    │
    └── Frontend (Next.js) Job
        ├── Install dependencies
        ├── pnpm lint              (ESLint)
        ├── pnpm typecheck         (TypeScript)
        ├── pnpm test              (Vitest)
        └── pnpm build             (production build)
```

### Dependabot (Automatic Dependency Updates)

Every week, Dependabot checks for updates to:
- GitHub Actions workflows
- npm packages (root)
- Go modules (`apps/api` and `packages/business-logic`)

It creates PRs automatically when updates are available.

### Issue Templates

Two templates are available when creating GitHub issues:
- **Bug Report** — structured format for reporting bugs
- **Feature Request** — structured format for suggesting features

---

## Quick Reference Card

### Essential Commands

| What You Want | Command |
|---------------|---------|
| Start everything (GUI) | `bash scripts/sandbox-start-gui.sh` |
| Start everything (VS Code) | Ctrl+Shift+P → "Sandbox: Start" |
| Start everything (terminal) | `make sandbox-start` |
| Stop everything | `make sandbox-stop` |
| Reset everything | `make sandbox-reset` |
| Check health | `bash scripts/sandbox-health.sh` |
| Run all tests | `make test` |
| Lint all code | `make lint` |
| Build for production | `make build` |
| Open database browser | VS Code sidebar → SQLTools icon |
| Connect to database | `make db-connect` |

### Service URLs

| Service | URL | Login |
|---------|-----|-------|
| ZarishLog App | http://localhost:3000 | — |
| API Health | http://localhost:8080/api/v1/health | — |
| Keycloak Admin | http://localhost:8080/admin | `admin` / `zarishlog_dev_password` |
| MinIO Console | http://localhost:9001 | `zarishlog` / `zarishlog_dev_password` |
| Meilisearch | http://localhost:7700 | Key: `zarishlog_search_key` |

### VS Code Shortcuts

| Shortcut | What It Does |
|----------|--------------|
| `Ctrl+Shift+P` | Open command palette |
| `Ctrl+Shift+B` | Start sandbox (default build task) |
| `F5` | Start debugging (API Server) |
| `Ctrl+S` | Save + auto-format |
| `Ctrl+Shift+L` | Format document |

### Three Difficulty Levels

| Level | How to Start | How to Test | How to Debug |
|-------|-------------|-------------|--------------|
| **GUI** | `sandbox-start-gui.sh` | Browser health check | Check browser console |
| **VS Code** | Tasks: Sandbox: Start | Tasks: Go/Web: Test | F5: API Server debugger |
| **Terminal** | `make sandbox-start` | `make test` | `go test -v -race ./...` |
