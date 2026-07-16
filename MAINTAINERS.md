# ZarishLog — Maintainers Guide

## Release Process

### 1. Prepare Release

```bash
# Update version
export VERSION="v0.3.0"

# Run full test suite
./scripts/test.sh --ci

# Validate all configs
./scripts/validate-config.sh

# Build all artifacts
./scripts/build.sh --version "${VERSION}"
```

### 2. Tag and Release

```bash
git tag -a "${VERSION}" -m "Release ${VERSION}"
git push origin "${VERSION}"
```

### 3. Publish Docker Images

```bash
./scripts/build.sh --docker --publish --version "${VERSION}"
```

### 4. Verify Deployment

```bash
# Health check
curl https://api.zarishlog.org/api/v1/health

# Smoke test
curl https://api.zarishlog.org/api/v1/products | jq '.data | length'
```

## CI/CD Pipeline

See `.github/workflows/ci.yml` for the automated pipeline:

1. **PR opened** → `go vet`, `go build`, `go test`, `pnpm build` (Next.js)
2. **Merge to main** → (future: Docker build + push to GHCR)
3. **Tag push** → (future: Auto-deploy to staging)
4. **Manual approval** → (future: Deploy to production)

## Adding a New Module

### Backend (Go)

```bash
# Create handler
touch apps/api/internal/handler/newmodule.go

# Add model
# Edit apps/api/internal/model/ (add struct with json/db/validate tags)

# Add routes
# Edit apps/api/cmd/api/main.go (add routes with middleware.RequireRole)

# Add SQL queries
# Create packages/data-models/sql/queries/newmodule.sql

# Run sqlc to generate code
cd apps/api && sqlc generate

# Use internal/response/ for JSON responses and internal/validator/ for validation
```

### Frontend (Next.js)

```bash
# Create page
touch apps/web/app/newmodule/page.tsx

# Create API client
touch apps/web/lib/api/newmodule.ts

# Add tests
touch apps/web/__tests__/newmodule.test.ts
```

## Database Migrations

```bash
# Create new migration
export TIMESTAMP=$(date -u +%Y%m%d_%H%M%S)
touch "packages/data-models/sql/migrations/${TIMESTAMP}_description.sql"

# Apply migration
make db-migrate

# Rollback (if needed)
psql -h localhost -U zarishlog -d zarishlog -c "DROP TABLE IF EXISTS new_table CASCADE;"
```

## Monitoring

- **Health:** `GET /api/v1/health`
- **Metrics:** Prometheus endpoint at `/api/v1/metrics` (future)
- **Logs:** Gin default logger (structured JSON logging via zerolog planned for future)

## Security

- **Secrets:** Never commit `.env`. Use GitHub Secrets for CI.
- **Dependencies:** Run `go mod verify` and `pnpm audit` regularly.
- **Vulnerabilities:** Monitor GitHub Dependabot alerts.
- **Auth:** All API endpoints (except health) require JWT via Keycloak.
