# ZarishLog Stabilization Report

> **Status:** Historical record. This document captures a completed stabilization effort from 2026-09-09 and is maintained for reference only.

**Date:** 2026-09-09
**Branch:** `chore/stabilize-platform`
**Commit:** `57e4469`

## Executive summary

The repository was audited against the supplied GitHub, Supabase, and Vercel references. The application repository is a Docker Compose/Terraform-based Go API plus Next.js PWA monorepo; Supabase and Vercel are not currently wired into the application runtime. A reviewable stabilization branch was created and pushed to GitHub. The frontend quality gates now pass locally, and the CI workflow runs linting, type-checking, unit tests, and a production build instead of skipping lint and tests.

The platform work is not fully deployable yet because the supplied Vercel team has no projects, the Supabase project has no public tables or migration history, and the intended relationship between the repository-managed PostgreSQL/Keycloak stack and Supabase has not been decided. The exposed GitHub personal access token must be revoked or rotated by the owner immediately.

## Changes delivered

| Area | Change | Evidence |
|---|---|---|
| Package management | Pinned pnpm to `11.24.0`; made native build-script approvals explicit for `esbuild` and `sharp` | Root `package.json`, `pnpm-workspace.yaml`, lockfile |
| Frontend type safety | Removed the unsupported Vitest `fakeTimers.doNotFake` option | `apps/web/vitest.config.ts`; type-check passes |
| Frontend linting | Added ESLint 9 flat config and migrated from deprecated interactive `next lint` to `eslint .` | `apps/web/eslint.config.mjs`, `apps/web/package.json` |
| Test coverage | Added offline helper tests and an online-reconnect hook regression test | `apps/web/lib/offline.test.ts`, `apps/web/hooks/useOnlineStatus.test.tsx` |
| Hook correctness | Fixed the missing `syncNow` effect dependency and initialized the callback before the effect | `apps/web/hooks/useOnlineStatus.ts`; lint passes without warnings |
| CI | Added least-privilege permissions, concurrency cancellation, frontend lint/test gates, and exact pnpm version | `.github/workflows/ci.yml` |
| Dependency monitoring | Added weekly Dependabot checks for GitHub Actions, npm/pnpm, and Go modules | `.github/dependabot.yml` |
| Documentation | Added platform integration/runbook documentation and linked it from the README | `docs/PLATFORM_INTEGRATION.md`, `README.md` |

## Verification evidence

The following commands passed after the changes:

```text
pnpm install --frozen-lockfile --ignore-scripts
pnpm --filter @zarishlog/web lint
pnpm --filter @zarishlog/web typecheck
pnpm --filter @zarishlog/web test
pnpm --filter @zarishlog/web build
git diff --check
```

The frontend test suite reports **2 test files and 4 tests passing**. The Next.js production build completed successfully and generated the expected routes. A staged-diff scan found no GitHub PAT, Supabase secret key, or private-key material in the committed change.

The Go checks could not be run in the sandbox because the Go toolchain is not installed there. GitHub Actions remains configured to run `go vet`, race-enabled tests, and the API build on a runner with the declared Go version.

## Connected-service findings

### Supabase

The active project is `ZarishLog`, ref `rnhwbmnttxjypcpvebfz`, in `ap-southeast-1`, and reports `ACTIVE_HEALTHY`. The public table inventory is empty and the migration history is empty, while the repository contains six PostgreSQL migrations and extensive seed data. The audit therefore made no schema changes and did not attempt to port the local database model into Supabase.

Supabase security advisors reported that `public.rls_auto_enable()` is a `SECURITY DEFINER` function executable by both `anon` and `authenticated` roles through PostgREST. This should be revoked, changed, or removed only after the owner confirms whether the function is intentionally exposed.

### Vercel

The supplied Vercel team was queried successfully, but its project inventory is empty. No deployment, domain, environment-variable, preview, or production configuration could therefore be validated or changed. A Vercel project must be created or linked by the owner with an approved root directory and environment-variable contract before deployment automation can be completed.

### GitHub

The stabilization branch was pushed successfully. GitHub reported **58 dependency vulnerabilities on the default branch: 1 critical, 39 high, and 18 moderate**. Dependabot monitoring was added to the branch, but the vulnerability backlog requires a separate dependency/security remediation pass with compatibility testing.

## Required owner actions

1. Revoke or rotate the GitHub PAT that was supplied in plaintext. Do not reuse it in CI, files, or chat.
2. Decide whether Supabase is intended to replace or supplement the repository-managed PostgreSQL/Keycloak architecture.
3. Confirm the intended Vercel project, root directory, domain, preview/production environment variables, and promotion policy.
4. Review the Supabase `SECURITY DEFINER` advisory and approve the least-privilege remediation.
5. Review and merge the branch through the normal pull-request process after the Go CI job passes on GitHub Actions.
6. Triage the reported GitHub dependency vulnerabilities, beginning with the critical and high-severity items.

## Deliberately not performed

No production deployment was triggered. No Supabase DDL, data, function privileges, or migration was changed. No Vercel project or domain was created. No credentials were copied into the repository. No merge to `main` was performed.
