# ZarishLog — Local Setup Guide

Follow these steps in order. Everything here is free.

## Prerequisites
- Node.js 20+ (`nvm install` will pick up `.nvmrc`)
- pnpm 9+ (`npm i -g pnpm`)
- Docker Desktop (or Docker Engine on Linux)

## 1. Clone and install

```bash
git clone https://github.com/<your-org>/zarishlog.git
cd zarishlog
cp .env.example .env
pnpm install
```

## 2. Start infrastructure

```bash
docker compose up -d
```

This starts Postgres (5432), Redis (6379), MinIO (9000/9001 console), and Keycloak (8080). Give Keycloak ~30 seconds to finish starting on first run.

## 3. Set up the database

```bash
pnpm db:generate     # generates the Prisma client from schema.prisma
pnpm db:migrate       # creates all tables
pnpm db:seed          # loads roles, reference data, org hierarchy, and the product catalogue
```

You should see console output like:
```
Seeded 9 roles and 40 permissions.
Seeded 9 UoMs and 6 product categories.
Seeded organization hierarchy: 1 org, 3 org levels, 7 programs, 1 warehouse, 5 locations.
Imported 18 products, skipped 0 malformed rows.
Seed complete.
```

> The 18 imported products come from the sample `config/metadata/master_product_list.csv`. Replace this file with the full 22,000+ item export when available, double-check the column mapping at the top of `packages/data-models/seed/seed.ts`, and re-run `pnpm db:seed` (it's idempotent — safe to re-run).

## 4. Run the app

```bash
pnpm dev
```

This starts both the API and web app via Turborepo:
- API: http://localhost:4000 (Swagger docs at `/docs`)
- Web: http://localhost:3000

Visit http://localhost:3000/products — you should see the seeded catalogue rendered in a table. This is the Phase 1/2 milestone from `docs/BLUEPRINT.md`: seeded database → API → UI, working end to end.

## 5. Inspect the database (optional)

```bash
pnpm db:studio
```

Opens Prisma Studio, a free GUI browser for the database, at http://localhost:5555.

## Troubleshooting

| Symptom | Fix |
|---|---|
| `pnpm db:migrate` fails to connect | Confirm `docker compose ps` shows `postgres` as healthy; wait a few seconds after `docker compose up -d` |
| Products page shows "No products found" | Confirm `pnpm db:seed` ran successfully and the API is running on port 4000 |
| Keycloak admin console unreachable | It can take 20–40s to boot on first start; check `docker compose logs keycloak` |
| Port already in use | Change `WEB_PORT`/`API_PORT` in `.env` |

## Next steps
See `docs/BLUEPRINT.md` for the full phased plan — Phase 2 (Core API & Business Logic) and Phase 3 (Web Console) build directly on this foundation.
