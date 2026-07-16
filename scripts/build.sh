#!/usr/bin/env bash
# ============================================================================
# ZarishLog — Build & Publish Pipeline
# ============================================================================
# Builds all artifacts: Go binary, Docker images, frontend static files.
# Supports local builds, CI builds, and Docker image publishing.
#
# Usage:
#   ./scripts/build.sh                    # Build everything locally
#   ./scripts/build.sh --go               # Build Go binary only
#   ./scripts/build.sh --docker           # Build Docker images
#   ./scripts/build.sh --publish          # Build + push to registry
#   ./scripts/build.sh --version v0.2.0   # Tag build with version
# ============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BUILD_GO=false
BUILD_FRONTEND=false
BUILD_DOCKER=false
PUBLISH=false
BUILD_ALL=true
VERSION="$(date '+%Y%m%d')-dev"
COMMIT_HASH="$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
DOCKER_REGISTRY="ghcr.io/cpintl"
BUILD_TIME="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --go) BUILD_GO=true; BUILD_ALL=false; shift ;;
    --frontend) BUILD_FRONTEND=true; BUILD_ALL=false; shift ;;
    --docker) BUILD_DOCKER=true; BUILD_ALL=false; shift ;;
    --publish) PUBLISH=true; shift ;;
    --version) VERSION="$2"; shift 2 ;;
    --registry) DOCKER_REGISTRY="$2"; shift 2 ;;
    --help) echo "Usage: $0 [--go|--frontend|--docker] [--publish] [--version X.Y.Z] [--registry url]"; exit 0 ;;
    *) echo "Unknown: $1"; exit 1 ;;
  esac
done

if [[ "$BUILD_ALL" == "true" ]]; then
  BUILD_GO=true
  BUILD_FRONTEND=true
  BUILD_DOCKER=true
fi

echo ""
echo -e "${BOLD}╔═══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║           ZarishLog Build Pipeline                           ║${NC}"
echo -e "${BOLD}╚═══════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  Version:    ${BOLD}${VERSION}${NC} (${COMMIT_HASH})"
echo -e "  Build time: ${BUILD_TIME}"
echo ""

# ─── Build Go Binary ────────────────────────────────────────────────────

build_go() {
  echo -e "${CYAN}── Building Go API Binary ──${NC}"
  echo ""

  local output="${PROJECT_DIR}/apps/api/bin/api"
  mkdir -p "${PROJECT_DIR}/apps/api/bin"

  local ldflags="\
    -X main.Version=${VERSION} \
    -X main.CommitHash=${COMMIT_HASH} \
    -X main.BuildTime=${BUILD_TIME} \
    -s -w"

  echo -e "  ${BLUE}→${NC} Building Go binary..."
  cd "${PROJECT_DIR}/apps/api"
  
  CGO_ENABLED=0 go build \
    -ldflags="$ldflags" \
    -o "$output" \
    ./cmd/api

  echo -e "  ${GREEN}✓${NC} Binary built: ${output}"
  ls -lh "$output" | awk '{print "    Size: " $5}'

  cd "${PROJECT_DIR}"
  echo ""
}

# ─── Build Frontend ─────────────────────────────────────────────────────

build_frontend() {
  echo -e "${CYAN}── Building Frontend ──${NC}"
  echo ""

  if [[ ! -d "${PROJECT_DIR}/apps/web" ]]; then
    echo -e "  ${RED}✗${NC} apps/web not found"
    return 1
  fi

  cd "${PROJECT_DIR}/apps/web"

  if [[ ! -d "node_modules" ]]; then
    echo -e "  ${BLUE}→${NC} Installing dependencies..."
    pnpm install --frozen-lockfile 2>/dev/null || pnpm install
  fi

  echo -e "  ${BLUE}→${NC} Building Next.js..."
  NEXT_PUBLIC_VERSION="${VERSION}" \
  NEXT_PUBLIC_BUILD_TIME="${BUILD_TIME}" \
  NEXT_PUBLIC_COMMIT_HASH="${COMMIT_HASH}" \
  pnpm build

  echo -e "  ${GREEN}✓${NC} Frontend built"
  echo ""

  cd "${PROJECT_DIR}"
}

# ─── Build Docker Images ────────────────────────────────────────────────

build_docker() {
  echo -e "${CYAN}── Building Docker Images ──${NC}"
  echo ""

  if ! command -v docker &>/dev/null; then
    echo -e "  ${RED}✗${NC} Docker not available"
    return 1
  fi

  if ! docker info &>/dev/null 2>&1; then
    echo -e "  ${RED}✗${NC} Docker daemon not running"
    return 1
  fi

  # API image
  echo -e "  ${BLUE}→${NC} Building API Docker image..."
  docker build \
    -f "${PROJECT_DIR}/infrastructure/docker/Dockerfile.api" \
    -t "${DOCKER_REGISTRY}/zarishlog-api:${VERSION}" \
    -t "${DOCKER_REGISTRY}/zarishlog-api:latest" \
    "${PROJECT_DIR}"
  echo -e "  ${GREEN}✓${NC} API image built"

  # Web image
  echo -e "  ${BLUE}→${NC} Building Web Docker image..."
  docker build \
    -f "${PROJECT_DIR}/infrastructure/docker/Dockerfile.web" \
    -t "${DOCKER_REGISTRY}/zarishlog-web:${VERSION}" \
    -t "${DOCKER_REGISTRY}/zarishlog-web:latest" \
    "${PROJECT_DIR}"
  echo -e "  ${GREEN}✓${NC} Web image built"

  # Publish
  if [[ "$PUBLISH" == "true" ]]; then
    echo ""
    echo -e "  ${BLUE}→${NC} Publishing images to ${DOCKER_REGISTRY}..."

    docker push "${DOCKER_REGISTRY}/zarishlog-api:${VERSION}"
    docker push "${DOCKER_REGISTRY}/zarishlog-api:latest"
    
    docker push "${DOCKER_REGISTRY}/zarishlog-web:${VERSION}"
    docker push "${DOCKER_REGISTRY}/zarishlog-web:latest"

    echo -e "  ${GREEN}✓${NC} Images published"
  fi

  echo ""
}

# ─── Summary ────────────────────────────────────────────────────────────

print_summary() {
  echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
  echo ""
  echo -e "  ${GREEN}✓${NC} Build complete: ${BOLD}${VERSION}${NC} (${COMMIT_HASH})"
  echo ""
  
  if [[ "$BUILD_GO" == "true" ]]; then
    echo -e "    API binary: apps/api/bin/api"
  fi
  if [[ "$BUILD_DOCKER" == "true" ]]; then
    echo -e "    API image:  ${DOCKER_REGISTRY}/zarishlog-api:${VERSION}"
    echo -e "    Web image:  ${DOCKER_REGISTRY}/zarishlog-web:${VERSION}"
  fi
  
  echo ""
  if [[ "$PUBLISH" == "true" ]]; then
    echo -e "  ${GREEN}✓${NC} Published to ${DOCKER_REGISTRY}"
  fi
  echo ""
}

# ─── Main ────────────────────────────────────────────────────────────────

if [[ "$BUILD_GO" == "true" ]]; then build_go; fi
if [[ "$BUILD_FRONTEND" == "true" ]]; then build_frontend; fi
if [[ "$BUILD_DOCKER" == "true" ]]; then build_docker; fi

print_summary
