# Changelog

All notable changes to ZarishLog are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] — 2026-09-13

First stable release ("Version 01"). The platform is feature-complete for
offline-first, multi-tenant humanitarian logistics.

### Added

- **API** (`apps/api`, Go 1.26 / Gin / sqlc / PostgreSQL): multi-tenant REST
  API under `/api/v1` with Row-Level Security, OIDC (Keycloak) authorization,
  role-based access control (12 roles), audit trail, and domain enums
  validated through `internal/validator`.
- **Web** (`apps/web`, Next.js 16 / React 19 / Tailwind CSS 4): PWA with
  Workbox service worker, offline-first IndexedDB sync layer (Dexie), product
  catalogue, stock, quality, distribution workflows, and a health/status page.
- **Shared rules** (`packages/business-logic`): FEFO issue logic and Average
  Monthly Consumption (AMC) calculation as pure, tested Go packages.
- **Data models** (`packages/data-models`): 7 sequential migrations, sqlc
  query layer, and reference seeds in `config/`.
- **Reference data**: expanded 1,912-product catalogue (pharmaceuticals,
  medical supplies, equipment), organization hierarchy (CPI, CPI-BD),
  departments, programs, units of measure, warehouse layout, and glossary.
- **Infrastructure**: Docker Compose stack (PostgreSQL 18, Redis 8, MinIO,
  Keycloak 26.7, Meilisearch), production-grade Dockerfiles for API and web,
  CI pipeline (backend vet/test/build; frontend lint/typecheck/test/build).
- **Governance**: MIT license, dependency governance via Dependabot.

### Fixed

- Web app status page reads the API health envelope correctly (`data.status`).
- Removed dead test-runner page that referenced a non-existent API route.
- `fake-indexeddb` polyfill moved from production code to the Vitest setup.
- `Dockerfile.web` rebuilt for Node 24 + pnpm 12 + workspace lockfile layout.
- Sync queue now caps automatic retries at 5 attempts.

### Changed

- Unified all package/app versions to `1.0.0` (previously `0.2.0` mixture).
- Docs rewritten for non-technical readers; `docs/archive/` retired.

[1.0.0]: https://github.com/zarishlog/zarishlog/releases/tag/v1.0.0