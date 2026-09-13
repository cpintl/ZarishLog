# ZarishLog

**Offline-first, multi-tenant humanitarian logistics and asset management platform**

[![CI Pipeline](https://github.com/cpintl/zarishlog/actions/workflows/ci.yml/badge.svg)](https://github.com/cpintl/zarishlog/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Version 1.0.0](https://img.shields.io/badge/Version-1.0.0-brightgreen)](https://github.com/cpintl/zarishlog)

---

## What ZarishLog is

ZarishLog is a complete software platform that helps humanitarian and development organizations track **everything they manage** — medicines, medical supplies, food, equipment, office assets, vehicles — across every level of their organization, **even without an internet connection**.

Think of it as a unified replacement for the spreadsheets, paper stock cards, and disjointed inventory systems that field offices typically use. It brings all of that into one place where every warehouse officer, department head, and auditor works from the **same data**.

---

## Who it is for

ZarishLog is built for organizations working in humanitarian relief and development — NGOs, government agencies, UN bodies, and development partners — particularly those operating in settings where:

- Internet connectivity is unreliable or unavailable
- Inventory spans multiple locations: central warehouses, field warehouses, health posts, offices
- Staff rotate frequently and need a system that is easy to learn
- Regulatory and donor reporting require accurate, auditable stock records

It is designed for the **CPI** operational model — organizations working across four levels: global, country office, project/department office, and field/program site.

---

## The problems it solves

| Problem | How ZarishLog addresses it |
| --- | --- |
| The same item is named differently in different spreadsheets | One standardized **product catalogue** (1,912 items across 65+ categories, with unique SKU codes for every item) |
| No one can see real-time stock across all locations | A single source of truth with **role-based access** — each person sees what they are authorized to see, in real time |
| Expiry dates and quality checks are tracked on paper | Built-in **FEFO** (first-expiry, first-out) issue logic, expiry alerts, QA inspection workflows |
| Field offices lose internet access | Full **offline mode**: the web app works entirely in the browser with local storage; changes sync automatically when connectivity returns |
| Tracking items across organizations, countries, and departments is chaotic | **Multi-tenant** data isolation — each organization has its own secure data space, enforced at the database level |
| Manual stock counts are slow and error-prone | Guided **physical count** workflows with variance tracking |
| Fixed assets (laptops, vehicles, furniture) live in separate spreadsheets | Integrated **asset management** with transfer tracking and depreciation awareness |

---

## Core capabilities

| Area | What it does |
| --- | --- |
| **Product catalogue** | A master list of all items the organization manages — medicines, supplies, equipment — with SKU codes, categories, batch/expiry tracking flags, and sourcing information |
| **Organization hierarchy** | Maps the organizational structure from global level down to individual health posts, with departments, programs, and warehouses tied to each level |
| **Warehouse management** | Physical warehouse layout, storage zones, bin locations, temperature monitoring readiness |
| **Stock management** | Goods receipt, stock issuance, stock-to-stock transfers, and stock adjustment — all with audit trails |
| **Quality assurance** | Inspection workflows for received goods, shelf-life review, quarantine management |
| **Distribution** | Program-based distribution tracking to beneficiaries or sub-warehouses |
| **Asset tracking** | Fixed assets (equipment, furniture, vehicles) with transfer and status tracking |
| **Reporting** | Stock reports, expiry alerts, AMC (average monthly consumption) calculations, dashboard views |
| **Access control** | 12 distinct roles with permission boundaries — from field staff who can only create distributions to auditors with read-only access across all stock |

---

## How the technology works (non-technical explanation)

ZarishLog is built from several connected pieces of software, each with a specific job:

### The server (backend)

**What it is:** The "brain" of the system — a program written in [Go](https://go.dev/) that runs on a server and handles all data processing, calculations, and security.

**What it does:** When someone in the field taps "Receive Stock" on their screen, the server validates that the action is allowed, records it correctly in the database, updates stock balances, and logs who did it and when. It also enforces rules — for example, it won't let you issue stock that doesn't exist.

**Where it lives:** `apps/api/`

---

### The web interface (frontend)

**What it is:** The screens and buttons that people actually interact with — built with [Next.js](https://nextjs.org/) and [React](https://react.dev/).

**What it does:** Presents stock data, forms for receiving and issuing stock, reports, dashboards, and administrative settings. Works as a **Progressive Web App (PWA)** — meaning it can be "installed" on a laptop or phone and will work without internet by storing data locally in the browser until the connection comes back.

**Where it lives:** `apps/web/`

---

### The database

**What it is:** A [PostgreSQL](https://www.postgresql.org/) database — a battle-tested, open-source system for storing structured data reliably.

**What it does:** Stores all product information, stock records, transactions, user accounts, and organizational data. Uses **Row-Level Security** so that a user in one organization can never see data from another organization.

**What it contains:** 100 data tables organized across 17 domains, 205 typed database queries, 7 sequential schema migrations, and a seed dataset of 1,912 products.

---

### Shared business rules

**What it is:** A standalone Go package containing the core logistics formulas.

**What it does:** Calculates average monthly consumption (AMC) for stock replenishment planning, and determines which batch to issue first based on expiry date (FEFO). These rules are tested independently and used by the server.

**Where it lives:** `packages/business-logic/`

---

### Authentication

**What it is:** [Keycloak](https://www.keycloak.org/) — an open-source identity and access management system.

**What it does:** Manages user logins, password policies, and role assignments. Connects to ZarishLog's 12-role permission model so each person only sees and does what their role allows.

---

### Object storage

**What it is:** [MinIO](https://min.io/) — an open-source file storage system compatible with Amazon S3.

**What it does:** Stores uploaded files — documents, photos of received goods, inspection records, and reports.

---

### Search

**What it is:** [Meilisearch](https://www.meilisearch.com/) — a fast, typo-tolerant search engine.

**What it does:** Powers the product search so users can find items quickly even with partial or misspelled names.

---

## Repository structure

The code lives in one repository (a **monorepo**) organized like this:

```
zarishlog/
├── apps/
│   ├── api/              Server and API code (Go)
│   ├── web/              User interface (Next.js)
│   └── mobile/           Mobile app (React Native — scaffold only, not yet built)
├── packages/
│   ├── data-models/      Database migrations, queries, shared type definitions
│   ├── business-logic/   Core formulas (FEFO issue logic, AMC calculations)
│   └── ui/               Shared interface components (scaffold only)
├── config/
│   ├── metadata/         Product catalogue, organization hierarchy, roles, programs, units
│   ├── location/         Warehouse definitions (zones, bins, temperature requirements)
│   ├── templates/        Form templates for goods receipt, stock issue, etc.
│   └── reference_data/   Glossary of logistics terms
├── infrastructure/
│   ├── docker/           Files for building deployment images
│   └── keycloak/         Authentication server configuration
├── docs/                 Architecture decisions, status reports, roadmaps
├── scripts/              Setup and development helper scripts
└── .github/workflows/    Automated testing pipeline (runs on every change)
```

---

## Quick start (for developers)

These commands set up a fully working local copy of the platform on your machine. You will need **Docker**, **Go**, **Node.js 24**, and **pnpm 12** installed.

### Automated setup

```bash
git clone https://github.com/cpintl/zarishlog.git
cd zarishlog
./scripts/zarishlog-setup.sh --check-only   # verify everything is installed
```

### Start the sandbox

```bash
bash scripts/sandbox-start.sh     # starts database, API server, and web app
bash scripts/sandbox-health.sh    # verify everything is running
```

### Stop or reset

```bash
bash scripts/sandbox-stop.sh      # shut everything down
bash scripts/sandbox-reset.sh     # wipe local data and start fresh
```

### Manual alternative

```bash
cp .env.example .env              # create your local configuration
make docker-up                    # start database + supporting services
make db-migrate                   # create all database tables
make db-seed                      # load product catalogue and sample data
make dev                          # start API (port 8080) and Web (port 3000)
```

Once running:
- **Web app:** http://localhost:3000
- **API:** http://localhost:8080
- **MinIO (file storage):** http://localhost:9001
- **Keycloak (user management):** http://localhost:8180

See [`SETUP.md`](./SETUP.md) for full prerequisites and troubleshooting.

---

## Development workflow

```bash
# Run tests
cd apps/api && go test ./...                   # backend
cd apps/web && pnpm test                       # frontend

# Type checking and linting
cd apps/web && pnpm typecheck && pnpm lint

# Validate configuration files
./scripts/validate-config.sh
```

See [`CONTRIBUTING.md`](./CONTRIBUTING.md) for branch naming, commit conventions, and code review process.

---

## Configuration reference

| File | Purpose |
| --- | --- |
| `config/metadata/master_product_catalogue.csv` | 1,912 products with SKU codes, categories, batch/expiry flags |
| `config/metadata/organization.csv` | Organizational hierarchy (CPI → CPI-BD → CPI-BD-CXB → CPI-CXB-UKH) |
| `config/metadata/roles.md` | 12 roles (R01–R12) with level and purpose |
| `config/metadata/departments.csv` | Department structure (HPP, HOP, HSS) |
| `config/metadata/programs.csv` | Program codes (Health & Nutrition, WASH, Livelihood) |
| `config/location/warehouse.json` | CPI Bangladesh site directory (L2 country office → L4 program site offices) |
| `config/reference_data/GLOSSARY.md` | Logistics terminology used throughout the system |

See [`CONFIGURE.md`](./CONFIGURE.md) for detailed configuration instructions.

---

## Project status

ZarishLog is at **Version 1.0.0** ("Version 01"). All 12 planned phases of core development are complete:

| Phase | Status |
| --- | --- |
| Phase 0 — Foundation (CI, Docker, project structure) | Complete |
| Phase 1 — Database & data models (100 tables, 7 migrations, 205 queries) | Complete |
| Phase 2 — API core (REST endpoints, auth, audit, validation) | Complete |
| Phase 3 — Product catalogue module | Complete |
| Phase 4 — Warehouse & location module | Complete |
| Phase 5 — Stock & inventory module | Complete |
| Phase 6 — Quality assurance module | Complete |
| Phase 7 — Distribution & asset management | Complete |
| Phase 8 — Replenishment & forecasting | Complete |
| Phase 9 — User & access management | Complete |
| Phase 10 — Offline-first & PWA | Scaffolded (service worker ready, offline sync working) |
| Phase 11 — Reporting & analytics | Scaffolded |

Ongoing work: Keycloak UI integration, production deployment pipeline (Terraform), mobile app, operational monitoring, and expanded test coverage.

See [`docs/STATUS.md`](./docs/STATUS.md) for the detailed current status.

---

## Documentation

| Document | Who it is for |
| --- | --- |
| [`SETUP.md`](./SETUP.md) | Developers — prerequisites, local environment setup |
| [`CONFIGURE.md`](./CONFIGURE.md) | System administrators — adding products, configuring warehouses, managing roles |
| [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md) | Developers and architects — design decisions, data model, security model |
| [`docs/BLUEPRINT.md`](./docs/BLUEPRINT.md) | Anyone — full development roadmap and feature plan |
| [`docs/STATUS.md`](./docs/STATUS.md) | Anyone — current completion status and what remains |
| [`AGENTS.md`](./AGENTS.md) | AI coding assistants — how to work in this codebase correctly |
| [`MAINTAINERS.md`](./MAINTAINERS.md) | Contributors — governance, release process, code ownership |
| [`config/reference_data/GLOSSARY.md`](./config/reference_data/GLOSSARY.md) | Anyone — definitions of all logistics terms used in the system |
| [`CHANGELOG.md`](./CHANGELOG.md) | Anyone — history of changes between versions |

---

## License

**MIT** — see [`LICENSE`](./LICENSE) for the full text.

The codebase is open-source and free to use, modify, and deploy. Documentation and catalogue data are provided as-is.