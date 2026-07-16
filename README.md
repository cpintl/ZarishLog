# ZarishLog

## Unified Humanitarian Logistics, Supply Chain & Asset Management Platform

[![CI Pipeline](https://github.com/cpintl/zarishlog/actions/workflows/ci.yml/badge.svg)](https://github.com/cpintl/zarishlog/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Stack: Go + Next.js](https://img.shields.io/badge/Stack-Go%201.26%20%7C%20Next.js%2015%20%7C%20PostgreSQL%2018-blue)](https://github.com/cpintl/zarishlog)

ZarishLog is an **open-source, offline-first, multi-tenant** platform unifying warehouse management (WMS), inventory management (IMS), procurement, quality assurance, distribution, and fixed-asset tracking for humanitarian and development organizations operating across multi-level (L1 Global → L2 Country Office → L3 Project Office → L4 Program Site) structures.

> Companion docs: [`SETUP.md`](./SETUP.md) · [`CONFIGURE.md`](./CONFIGURE.md) · [`BLUEPRINT.md`](./docs/BLUEPRINT.md) · [`ARCHITECTURE.md`](./docs/ARCHITECTURE.md) · [`STATUS.md`](./docs/STATUS.md) · [`MAINTAINERS.md`](./MAINTAINERS.md) · [`GLOSSARY.md`](./config/reference_data/GLOSSARY.md) · [`PRODUCT_REQUIREMENTS_DOCUMENT.md`](./docs/PRODUCT_REQUIREMENTS_DOCUMENT.md)

---

## 1. Why ZarishLog

Humanitarian organizations run inventory across disconnected spreadsheets, paper stock cards, and heavyweight commercial ERPs their field offices can't afford or run offline. This creates:

- **Fragmented data** — same item, warehouse, or program named differently across sheets
- **No real-time visibility** across L1–L4 levels
- **Weak compliance** — FEFO, expiry, batch, QA rules enforced manually or not at all
- **No offline capability** — field/camp locations frequently lose connectivity

ZarishLog solves this with one canonical master catalogue, one multi-tenant data model, and an offline-first stack.

## 2. Core Objectives

| Objective         | What it means in practice                                                    |
| ----------------- | ---------------------------------------------------------------------------- |
| **Unification**   | One platform for medical, non-medical, consumable, and fixed-asset inventory |
| **Visibility**    | Real-time stock, location, and movement across all L1–L4 levels              |
| **Compliance**    | FEFO/FIFO enforcement, expiry management, QA workflows, full audit trail     |
| **Efficiency**    | Automated GRN, issue, transfer, adjustment, and AMC/FMC-based replenishment  |
| **Integration**   | Clean APIs into Finance, Procurement, and MEAL systems                       |
| **Resilience**    | Works fully offline at field/camp level and syncs when connectivity returns  |
| **Zero/low cost** | Built entirely on open-source software                                       |

## 3. Tech Stack

| Layer              | Choice                                         |
| ------------------ | ---------------------------------------------- |
| **Backend**        | Go 1.26 + Gin 1.12 (REST API)                  |
| **Frontend**       | Next.js 15 + React 19 PWA (Workbox + Dexie.js) |
| **Mobile**         | Expo/React Native                              |
| **Database**       | PostgreSQL 18 (sqlc + sqlx, RLS multi-tenant)  |
| **Auth**           | Keycloak 26 (OIDC/OAuth2)                      |
| **Search**         | Meilisearch                                    |
| **Analytics**      | Metabase (CE)                                  |
| **Object Storage** | MinIO (S3-compatible)                          |
| **Infrastructure** | Docker Compose → k3s, Terraform                |
| **CI/CD**          | GitHub Actions                                 |

## 4. Monorepo Structure

```
zarishlog/
├── apps/
│   ├── api/              # Go + Gin REST API (backend)
│   ├── web/              # Next.js 15 PWA (frontend)
│   └── mobile/           # Expo/React Native (field ops)
├── packages/
│   ├── data-models/      # SQL migrations, sqlc queries, Go types
│   ├── business-logic/   # Shared Go business rules (FEFO, AMC, etc.)
│   └── ui/               # Shared React components (design system)
├── infrastructure/
│   ├── docker/           # Dockerfiles
│   ├── terraform/        # IaC modules
│   └── kubernetes/       # k3s manifests (future)
├── docs/                 # Architecture, PRD, Blueprint docs
├── config/               # CSV metadata, templates, location data
├── .github/workflows/    # CI/CD pipelines
└── scripts/              # Build, seed, utility scripts
```

## 5. Getting Started

### Quick Start (one-click sandbox)

```bash
./scripts/zarishlog-setup.sh --check-only   # verify prerequisites
./scripts/sandbox.sh start                  # start Docker, DB, API, and Web
```

Use the VS Code tasks named "Sandbox: Start", "Sandbox: Stop", and "Sandbox: Reset" for the same flow from the editor.

### Manual Steps

```bash
git clone https://github.com/cpintl/zarishlog.git
cd zarishlog
cp .env.example .env
make docker-up      # starts postgres, redis, minio, keycloak, meilisearch
make db-migrate     # creates all tables
make db-seed        # loads master catalogue
make dev            # starts API (:8080) + Web (:3000)
```

### One-click sandbox (recommended)

You can start a local sandbox with the provided scripts or via VS Code Tasks.

Terminal commands:

```bash
# Start full sandbox (pull images, wait for DB, run migrations + seed)
bash scripts/sandbox-start.sh

# Stop all sandbox services
bash scripts/sandbox-stop.sh

# Reset local data (drops volumes/local data and re-seeds)
bash scripts/sandbox-reset.sh

# Quick health check
bash scripts/sandbox-health.sh
```

VS Code: open the Command Palette (Ctrl+Shift+P) → Tasks: Run Task → choose `Sandbox: Start` / `Sandbox: Stop` / `Sandbox: Reset` / `Sandbox: Health`.
See [`SETUP.md`](./SETUP.md) for detailed setup instructions. A 1912-product master catalogue CSV is included at `config/metadata/master_product_catalogue.csv` (65+ categories, 3 major domains: Drugs, Assets, Humanitarian/Technical).

## 6. Status

Sandbox environment is scaffolded. Master Catalogue designed (1912 products across 65+ categories in expanded CSV catalogue). Phase 0 (Foundation) complete, Phase 1 (Database) complete, Phase 2 (API Core) complete, Phases 3–5 (Catalogue, Warehouse, Stock) partially built. See [`STATUS.md`](./docs/STATUS.md) for detailed phase tracking.

## 7. License

**MIT** for code, **CC-BY-4.0** for documentation/catalogue data.
