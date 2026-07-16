# ZarishLog
## Unified Humanitarian Logistics, Supply Chain & Asset Management Platform

[![CI Pipeline](https://github.com/cpintl/zarishlog/actions/workflows/ci-test-pipeline.yml/badge.svg)](https://github.com/cpintl/zarishlog/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Stack: Offline-First PWA](https://img.shields.io/badge/Stack-Next.js%20%7C%20Supabase%20%7C%20PouchDB-blue)](https://github.com/cpintl)

ZarishLog is an open-source, offline-first, multi-tenant platform that unifies warehouse management (WMS), inventory management (IMS), procurement, quality assurance, distribution, and fixed-asset tracking for humanitarian and development organizations operating across multi-level (L1 Global → L2 Country Office → L3 Project Office/Central Warehouse → L4 Program Site/Sub-Warehouse) structures.
 
It is the successor and unification of two earlier working names in this project's history — **ZarishStock** (WMS-focused) and the original **ZarishLog** blueprint (full logistics + asset scope) — merged into one system and one set of documents.
 
> Companion documents: [`PRODUCT_REQUIREMENTS_DOCUMENT.md`](./PRODUCT_REQUIREMENTS_DOCUMENT.md) · [`BLUEPRINT.md`](./BLUEPRINT.md) · [`ARCHITECTURE.md`](./ARCHITECTURE.md)
 
---
 
## 1. Why ZarishLog
 
Humanitarian organizations run inventory today across disconnected spreadsheets (`MASTER ASSETS INVENTORY.xlsx`), paper stock cards, and heavyweight commercial ERPs their field offices can't afford or run offline. This creates four recurring failures the source material for this project documents directly:
 
- **Fragmented data** — the same item, warehouse, or program is named differently across sheets and offices, with no canonical ID.
- **No real-time visibility** across L1–L4 levels — a country office cannot see project-office stock until someone emails a spreadsheet.
- **Weak compliance** — FEFO, expiry, batch, and QA rules are enforced manually or not at all.
- **No offline capability** — field/camp locations frequently lose connectivity, and most commercial WMS tools assume a permanent connection.
ZarishLog solves this with one canonical master catalogue, one multi-tenant data model, and an offline-first application that keeps working when the network doesn't.
 
## 2. Core Objectives
 
| Objective | What it means in practice |
|---|---|
| **Unification** | One platform for medical, non-medical, consumable, and fixed-asset inventory |
| **Visibility** | Real-time stock, location, and movement across all L1–L4 levels |
| **Compliance** | FEFO/FIFO enforcement, expiry management, QA workflows, full audit trail |
| **Efficiency** | Automated GRN, issue, transfer, adjustment, and AMC/FMC-based replenishment |
| **Integration** | Clean APIs into Finance, Procurement, and MEAL systems |
| **Resilience** | Works fully offline at field/camp level and syncs when connectivity returns |
| **Zero/low cost** | Built entirely on open-source software and always-free or generous free-tier cloud services |
 
## 3. Tech Stack at a Glance
 
| Layer | Choice | Why |
|---|---|---|
| Monorepo | **Turborepo** (npm/pnpm workspaces) | Fast incremental builds, shared packages across web/mobile/api |
| Frontend (Web/PWA) | **Next.js 15** (React 19) + Tailwind CSS + shadcn/ui | PWA-capable, huge ecosystem, free hosting on Vercel/Cloudflare Pages |
| Mobile | **React Native (Expo)** | Shares business logic/types with the web app; one codebase for Android + iOS |
| Backend API | **NestJS** (Node.js/TypeScript) | Structured, DI-based, scales with team size better than ad-hoc API routes; still fine to prototype with Next.js route handlers early and graduate later |
| Database | **PostgreSQL 17** (self-hosted or Supabase/Neon free tier) | Relational integrity for a 22,000+ item catalogue and audit-critical stock ledger; row-level security for multi-tenancy |
| ORM | **Prisma** or **Drizzle** | Type-safe schema + migrations shared between NestJS and scripts |
| Auth | **Keycloak** (self-host) or **Supabase Auth** (managed free tier) | OIDC/JWT, RBAC, MFA, SSO-ready |
| Offline sync | **RxDB** or **PowerSync** + **PGlite**/SQLite (mobile) | CRDT/log-based sync between local device DB and Postgres; proven offline-first pattern |
| Object storage | **MinIO** (self-host, S3-compatible) or Cloudflare R2 free tier | Photos, scanned GRNs, batch certificates |
| Background jobs | **BullMQ** + Redis | AMC calculation, expiry alerts, notification dispatch |
| CI/CD | **GitHub Actions** | Free for public/most private repos; GUI-driven approvals via GitHub Environments |
| IaC | **Terraform** | Declarative infra for self-host, hybrid, or cloud (GCP/AWS/Azure free tiers) |
| Container/orchestration | **Docker Compose** (single VPS) → **k3s** (lightweight Kubernetes) at scale | Keeps ops simple until real scale is needed |
| Reverse proxy/TLS | **Caddy** or **Traefik** | Automatic HTTPS, zero-config free certs |
| Monitoring | **Beszel** / **Uptime Kuma** + **OpenObserve** | Lightweight, self-hosted, free |
| BI/Reporting | **Metabase** (free CE) | Point-and-click dashboards over Postgres, no separate BI build needed |
 
Full rationale for each choice is in [`ARCHITECTURE.md`](./ARCHITECTURE.md).
 
## 4. Monorepo Structure
 
```
zarishlog/
├── apps/
│   ├── web/              # Next.js PWA (staff console, dashboards, forms)
│   ├── mobile/            # Expo/React Native (scanning, field ops, offline)
│   └── api/                # NestJS backend (REST/GraphQL, business logic)
├── packages/
│   ├── ui/                 # Shared React components (design system)
│   ├── data-models/        # Shared TypeScript types + Zod/Prisma schemas
│   ├── business-logic/     # FEFO sort, AMC/FMC calc, reorder point, QR/barcode gen
│   └── config/             # ESLint, TS config, Tailwind config
├── config/
│   ├── metadata/            # Canonical CSV/JSON: UoM, categories, statuses, roles
│   └── templates/           # GRN, SRF, transfer, adjustment, QA form definitions
├── infrastructure/
│   ├── terraform/           # Cloud/self-host provisioning modules
│   └── kubernetes/          # k3s manifests (optional, phase 3+)
├── docs/
│   ├── PRODUCT_REQUIREMENTS_DOCUMENT.md
│   ├── BLUEPRINT.md
│   └── ARCHITECTURE.md
└── .github/workflows/       # CI/CD pipelines
```
 
## 5. Getting Started (Local, Free)
 
```bash
git clone https://github.com/<your-org>/zarishlog.git
cd zarishlog
cp .env.example .env          # fill in local secrets
docker compose up -d           # postgres, redis, minio, keycloak
pnpm install
pnpm db:migrate && pnpm db:seed   # loads master catalogue from config/metadata
pnpm dev                        # runs web + api together via Turborepo
```
 
Everything above runs on a free-tier VPS or even a laptop — no paid service is required to develop or demo the full system.
 
## 6. Status
 
This repository currently ships four foundational documents (README, PRD, Blueprint, Architecture) synthesizing all prior ZarishStock/ZarishLog research and catalogue work into one coherent, buildable plan. Implementation follows the phased roadmap in `BLUEPRINT.md`, starting with the database schema and master catalogue import.
 
## 7. License
 
Recommended: **MIT** for code, **CC-BY-4.0** for documentation/catalogue data — keeps the project maximally reusable by other humanitarian organizations, consistent with the "re-usable piece by piece" requirement.

## 8. Contributing
 
Non-coder-led, AI-paired development is the primary workflow for this project. See `BLUEPRINT.md` Phase 0 for how specifications are turned into working code with Claude/AI assistance without requiring the maintainer to hand-write code.


