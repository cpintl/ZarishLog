#!/usr/bin/env bash
# ============================================================================
# ZarishLog — Interactive Test Runner
# ============================================================================
# Runs all test suites with visual output, coverage reports, and status badges.
#
# Usage:
#   ./scripts/test.sh              # Run all tests
#   ./scripts/test.sh --go         # Go tests only
#   ./scripts/test.sh --frontend   # Frontend tests only
#   ./scripts/test.sh --watch      # Watch mode (file change detection)
#   ./scripts/test.sh --coverage   # With coverage report
#   ./scripts/test.sh --ci         # CI mode (strict, no watch)
# ============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
RUN_GO=false
RUN_FRONTEND=false
RUN_ALL=true
WATCH=false
COVERAGE=false
CI_MODE=false
EXIT_CODE=0

# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --go) RUN_GO=true; RUN_ALL=false; shift ;;
    --frontend) RUN_FRONTEND=true; RUN_ALL=false; shift ;;
    --watch) WATCH=true; shift ;;
    --coverage) COVERAGE=true; shift ;;
    --ci) CI_MODE=true; shift ;;
    --help) echo "Usage: $0 [--go|--frontend] [--watch] [--coverage] [--ci]"; exit 0 ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

if [[ "$RUN_ALL" == "true" ]]; then
  RUN_GO=true
  RUN_FRONTEND=true
fi

# ─── Header ──────────────────────────────────────────────────────────────

print_header() {
  echo ""
  echo -e "${BOLD}╔═══════════════════════════════════════════════════════════════╗${NC}"
  echo -e "${BOLD}║            ZarishLog Test Runner                             ║${NC}"
  echo -e "${BOLD}╚═══════════════════════════════════════════════════════════════╝${NC}"
  echo ""
  echo -e "  Date:    $(date '+%Y-%m-%d %H:%M:%S')"
  echo -e "  Project: ${PROJECT_DIR}"
  echo -e "  Mode:    $([[ "$CI_MODE" == "true" ]] && echo "CI (strict)" || echo "Interactive")"
  echo ""
}

# ─── Section Header ──────────────────────────────────────────────────────

section() {
  echo ""
  echo -e "${CYAN}┌─────────────────────────────────────────────────────────────┐${NC}"
  echo -e "${CYAN}│${NC}  ${BOLD}$1${NC}"
  echo -e "${CYAN}└─────────────────────────────────────────────────────────────┘${NC}"
  echo ""
}

# ─── Results ──────────────────────────────────────────────────────────────

pass() {
  echo -e "  ${GREEN}✓${NC} $1"
}

fail() {
  echo -e "  ${RED}✗${NC} $1"
  EXIT_CODE=1
}

info() {
  echo -e "  ${BLUE}ℹ${NC} $1"
}

# ─── Environment Check ───────────────────────────────────────────────────

check_env() {
  section "Environment Check"

  local status=0
  
  # Go
  if command -v go &>/dev/null; then
    pass "Go: $(go version | grep -oP 'go\S+')"
  else
    fail "Go: NOT INSTALLED"
    status=1
  fi

  # Node
  if command -v node &>/dev/null; then
    pass "Node.js: $(node --version)"
  else
    fail "Node.js: NOT INSTALLED"
    status=1
  fi

  # pnpm
  if command -v pnpm &>/dev/null; then
    pass "pnpm: $(pnpm --version)"
  else
    fail "pnpm: NOT INSTALLED"
    status=1
  fi

  # Docker
  if command -v docker &>/dev/null; then
    if docker info &>/dev/null 2>&1; then
      pass "Docker: running"
    else
      fail "Docker: installed but not running"
      status=1
    fi
  else
    fail "Docker: NOT INSTALLED"
    status=1
  fi

  # Database
  if command -v psql &>/dev/null; then
    if pg_isready -h localhost -U zarishlog &>/dev/null 2>&1; then
      pass "PostgreSQL: connected"
    else
      info "PostgreSQL: client installed but server may not be running"
    fi
  else
    info "psql: not installed (required for DB tests)"
  fi

  return $status
}

# ─── Go Tests ────────────────────────────────────────────────────────────

run_go_tests() {
  section "Go Tests"

  if [[ ! -d "${PROJECT_DIR}/apps/api" ]]; then
    fail "apps/api directory not found"
    return 1
  fi

  cd "${PROJECT_DIR}/apps/api"

  local extra_args=""
  [[ "$COVERAGE" == "true" ]] && extra_args="-coverprofile=coverage.out -covermode=atomic"
  [[ "$CI_MODE" == "true" ]] && extra_args="$extra_args -v"
  [[ "$WATCH" == "true" ]] && extra_args="$extra_args -count=1"

  # Download dependencies
  info "Downloading Go dependencies..."
  go mod tidy 2>&1 | sed 's/^/    /' || true

  # Vet
  info "Running go vet..."
  go vet ./... 2>&1 | sed 's/^/    /' || { fail "go vet failed"; }

  # Business logic tests (FEFO, AMC)
  info "Running business logic tests..."
  cd "${PROJECT_DIR}/packages/business-logic"
  if go test ./... -v -count=1 2>&1 | sed 's/^/    /'; then
    pass "Business logic tests passed"
  else
    fail "Business logic tests failed"
  fi

  # API tests
  info "Running API tests..."
  cd "${PROJECT_DIR}/apps/api"
  if go test ./... $extra_args -count=1 -timeout=60s 2>&1 | sed 's/^/    /'; then
    pass "All API tests passed"
  else
    fail "Some API tests failed"
  fi

  # Coverage report
  if [[ "$COVERAGE" == "true" ]] && [[ -f coverage.out ]]; then
    info "Coverage report:"
    go tool cover -func=coverage.out 2>/dev/null | tail -1 | sed 's/^/    /'
    # Generate HTML report
    go tool cover -html=coverage.out -o coverage.html 2>/dev/null
    pass "Coverage report: ${PROJECT_DIR}/apps/api/coverage.html"
  fi

  cd "${PROJECT_DIR}"
}

# ─── Frontend Tests ─────────────────────────────────────────────────────

run_frontend_tests() {
  section "Frontend Tests"

  if [[ ! -d "${PROJECT_DIR}/apps/web" ]]; then
    fail "apps/web directory not found"
    return 1
  fi

  cd "${PROJECT_DIR}/apps/web"

  if [[ ! -d "node_modules" ]]; then
    info "Installing frontend dependencies..."
    pnpm install 2>&1 | sed 's/^/    /' || true
  fi

  # Lint
  info "Running ESLint..."
  if pnpm lint 2>&1 | sed 's/^/    /'; then
    pass "ESLint passed"
  else
    fail "ESLint found issues"
  fi

  # TypeScript type check
  info "Running TypeScript type check..."
  if pnpm typecheck 2>&1 | sed 's/^/    /'; then
    pass "TypeScript type check passed"
  else
    fail "TypeScript type check failed"
  fi

  # Unit tests
  if [[ -f "vitest.config.ts" ]] || [[ -f "vitest.config.js" ]] || grep -q '"test"' package.json 2>/dev/null; then
    info "Running unit tests..."
    local test_cmd="pnpm test"
    [[ "$WATCH" == "true" ]] && test_cmd="pnpm test -- --watch"
    [[ "$COVERAGE" == "true" ]] && test_cmd="$test_cmd -- --coverage"
    
    if $test_cmd 2>&1 | sed 's/^/    /'; then
      pass "All frontend tests passed"
    else
      fail "Some frontend tests failed"
    fi
  else
    info "No test runner configured for frontend (skipping)"
  fi

  # Build test
  info "Running build check..."
  if pnpm build 2>&1 | tail -5 | sed 's/^/    /'; then
    pass "Build successful"
  else
    fail "Build failed"
  fi

  cd "${PROJECT_DIR}"
}

# ─── Integration Tests ──────────────────────────────────────────────────

run_integration_tests() {
  section "Integration Tests"

  local api_url="http://localhost:8080"

  # Check if API is running
  if curl -sf "${api_url}/api/v1/health" >/dev/null 2>&1; then
    pass "API server is running at ${api_url}"
  else
    info "API server not running. Start with: cd apps/api && go run ./cmd/api"
    info "Skipping integration tests"
    return 0
  fi

  # Test health endpoint
  info "Testing health endpoint..."
  local health
  health="$(curl -sf "${api_url}/api/v1/health" 2>/dev/null || echo '{"status":"failed"}')"
  if echo "$health" | grep -q '"healthy"'; then
    pass "Health check: healthy"
  else
    fail "Health check failed: $health"
  fi

  # Test products endpoint
  info "Testing products endpoint..."
  local products
  products="$(curl -sf "${api_url}/api/v1/products" 2>/dev/null || echo '')"
  if echo "$products" | grep -q '"data"'; then
    local count
    count="$(echo "$products" | grep -o '"sku"' | wc -l)"
    pass "Products endpoint: ${count} products loaded"
  else
    fail "Products endpoint failed"
  fi
}

# ─── Summary ─────────────────────────────────────────────────────────────

print_summary() {
  echo ""
  echo -e "${BOLD}╔═══════════════════════════════════════════════════════════════╗${NC}"
  if [[ "$EXIT_CODE" -eq 0 ]]; then
    echo -e "${BOLD}║  ${GREEN}ALL TESTS PASSED${NC}${BOLD}                                                    ║${NC}"
  else
    echo -e "${BOLD}║  ${RED}SOME TESTS FAILED${NC}${BOLD}                                                  ║${NC}"
  fi
  echo -e "${BOLD}╚═══════════════════════════════════════════════════════════════╝${NC}"
  echo ""
  exit $EXIT_CODE
}

# ─── Main ────────────────────────────────────────────────────────────────

main() {
  print_header
  
  check_env || true  # Don't exit on env check failures
  
  if [[ "$RUN_GO" == "true" ]]; then
    run_go_tests
  fi
  
  if [[ "$RUN_FRONTEND" == "true" ]]; then
    run_frontend_tests
  fi
  
  run_integration_tests
  
  print_summary
}

main
