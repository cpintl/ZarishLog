# Contributing

Thanks for contributing to ZarishLog! This guide explains how to make changes
that are easy to review and safe to ship.

## Repository layout

- `apps/api` — Go REST API (Gin, sqlc, PostgreSQL).
- `apps/web` — Next.js 16 PWA frontend.
- `apps/mobile` — Expo/React Native client (scaffold only).
- `packages/data-models` — SQL migrations, queries, shared schema artifacts.
- `packages/business-logic` — pure Go business rules (FEFO, AMC).
- `packages/ui` — shared UI components (scaffold only).
- `config/` — reference data (catalogue, organization, roles, programs, UoM).
- `infrastructure/` — Docker, Keycloak realm, CI definitions.

## Ground rules

1. Read the affected module and its docs before editing.
2. Match existing patterns — do not introduce new frameworks or layers.
3. Prefer small, focused changes over broad refactors.
4. If you change SQL or schema, update the migration, regenerate with
   `sqlc generate` (run from `apps/api`), and commit generated files.
5. Run the smallest relevant validation command after every change.
6. Never commit secrets. `.env` is gitignored on purpose.

## Validation

Always run the matching checks before opening a PR:

```bash
# Backend
cd apps/api && go vet ./... && go test ./... -short
cd packages/business-logic && go test ./...

# Frontend
cd apps/web && pnpm lint && pnpm typecheck && pnpm test && pnpm build
```

From the repo root you can shorten this to `make lint` and `make test`.

## Git workflow

- Branch from `main`: `git checkout -b <topic>` (e.g. `feat/sync-queue`).
- Commit in small logical units with a short imperative subject line, e.g.
  `fix(web): unwrap status health envelope`.
- Enable the pre-commit hook: `git config core.hooksPath .githooks`.
- Keep `docs/STATUS.md` and this guide accurate when you change scope.

## Docs for non-coders

User-facing documentation must avoid jargon. If you change an architecture
decision, update the relevant section in `docs/` — or, for a new public
interface, open a pull request that documents it at the same time.