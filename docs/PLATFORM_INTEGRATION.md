# ZarishLog Platform Integration and Operations

**Last verified:** 2026-09-09

This document records the platform state observed during the repository stabilization work. It distinguishes the repository’s current self-hosted architecture from the separately supplied Supabase and Vercel accounts so that deployment work does not silently assume an integration that has not been configured.

## Current architecture boundary

The repository currently implements a Go API, a Next.js 16 PWA, PostgreSQL migrations and seed data, Keycloak/OIDC configuration, Redis, MinIO-compatible object storage, Meilisearch, Docker Compose, and Terraform scaffolding. The application source does not currently reference the Supabase client or Supabase URL, and the repository does not contain a `supabase/` directory. Supabase therefore remains an external project inventory item, not an application runtime dependency.

| Area | Verified state | Operational consequence |
|---|---|---|
| GitHub | `cpintl/ZarishLog`, default branch `main`; repository is clean at the inspected revision before local changes | Changes should land through a short-lived branch and pull request |
| Supabase | Project `ZarishLog`, ref `rnhwbmnttxjypcpvebfz`, region `ap-southeast-1`, status `ACTIVE_HEALTHY`; public table inventory and migration history are empty | Do not apply the repository’s PostgreSQL migrations to this project until the owner confirms Supabase is the intended database target |
| Supabase security | The `public.rls_auto_enable()` function is `SECURITY DEFINER` and executable by anonymous and authenticated roles | Revoke public execution or move the function out of the exposed API surface after confirming whether it is needed |
| Vercel | Team `ZarishLogDevOps` was queried successfully, but it currently has no projects | There is no Vercel project or deployment target to configure yet |
| CI | GitHub Actions already tests the Go backend against PostgreSQL and builds the frontend; frontend lint and test gates were previously disabled or absent | CI now runs explicit lint, type-check, unit-test, and build steps |
| Secrets | A GitHub PAT was supplied in plaintext in project instructions | Revoke or rotate that token immediately; do not copy it into repository files or CI settings |

## Required environment separation

Local development uses the `.env.example` contract for Docker Compose services. Preview and production must use managed environment variables with separate values and scopes. Only public browser-safe values may use a `NEXT_PUBLIC_` prefix. Database URLs, OIDC client secrets, object-storage secret keys, Meilisearch keys, and SMTP credentials must remain server-side.

The repository’s current local stack uses PostgreSQL, Redis, MinIO, Keycloak, and Meilisearch. A Supabase URL or publishable key should not be added to the application merely because the account exists; that would create a second, undocumented data path and could bypass the existing tenant/RLS design.

## CI quality gates

Pull requests and pushes to `main` run the following gates:

1. The Go job provisions PostgreSQL, applies the repository migrations, seeds test data, runs `go vet`, runs race-enabled tests, and builds the API.
2. The frontend job installs the locked pnpm workspace, runs ESLint, runs TypeScript checking, runs Vitest, and builds the Next.js PWA.
3. Workflow permissions are limited to `contents: read`, and duplicate runs for the same ref are cancelled.

The production deployment step remains intentionally absent because there is no verified Vercel project or alternative production target connected to this repository.

## Supabase decision gate

Before any database mutation, answer all of the following:

| Question | Required answer |
|---|---|
| Is Supabase replacing the repository’s PostgreSQL/Keycloak architecture? | Explicit owner decision |
| Should the six repository migrations be ported and applied to Supabase? | Explicit migration plan and compatibility review |
| Should authentication remain Keycloak or move to Supabase Auth? | Explicit identity decision |
| Should the existing `SECURITY DEFINER` function remain callable through PostgREST? | Security owner approval or revocation change |
| Is production data currently present in Supabase? | Verified backup and data classification |

Until those answers are recorded, the safe action is read-only inspection and no schema change.

## Vercel onboarding runbook

When the team owner is ready to connect a project, create or import the repository as a Vercel project, set the correct root directory (`apps/web` if deploying only the web app), pin Node 24 LTS (or 26 LTS when ready to upgrade) and pnpm 12.0.0, configure preview and production environment variables separately, and run the same lint/type-check/test/build gates before enabling production promotion. The deployment should expose a health or status route that can be checked after each release.

A Vercel project should not be invented or created automatically during this audit because the team currently has no projects and the intended root, domain, environment values, and production approval path are not yet confirmed.

## Incident and rollback procedure

For a failed application release, stop promotion, identify the first failing CI or deployment check, and roll back to the last verified commit or Vercel deployment. For database changes, restore from the approved backup or apply a tested down/forward migration according to the migration record; do not reset a shared production database. For secret exposure, revoke the credential first, then inspect history and deployment logs, replace the secret in managed stores, and redeploy affected services.

## Outstanding owner actions

The repository work can be reviewed independently, but platform completion remains blocked by four owner actions: rotate the exposed GitHub PAT, decide whether Supabase is an intended runtime database, create or identify the intended Vercel project, and provide the approved preview/production environment-variable contract through secure secret management rather than chat or committed files.
