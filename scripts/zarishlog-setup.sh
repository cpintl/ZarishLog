#!/usr/bin/env bash
# ============================================================================
# ZarishLog — Sandbox Bootstrap Script
# ============================================================================
# Auto-detects local environment, installs missing prerequisites,
# configures the development sandbox, and validates the setup.
#
# Usage:
#   chmod +x scripts/zarishlog-setup.sh
#   ./scripts/zarishlog-setup.sh [--install-go] [--install-node] [--install-docker] [--yes]
#
# Options:
#   --yes            Auto-approve all installations (non-interactive)
#   --install-go     Install/upgrade Go even if present
#   --install-node   Install/upgrade Node.js even if present
#   --install-docker Install Docker if missing
#   --check-only     Only check prerequisites, don't install anything
#   --help           Show this help message
# ============================================================================

set -euo pipefail

# ─── Version Pinning ─────────────────────────────────────────────────────
# All versions are pinned here. Update these when upgrading the stack.
readonly GO_VERSION="1.26.4"
readonly GO_DOWNLOAD_BASE="https://go.dev/dl"
readonly NODE_MAJOR="22"
readonly PNPM_VERSION="11"
readonly GOLANGCI_LINT_VERSION="1.64.2"
readonly SQLC_VERSION="1.27.0"
readonly GOFUMPT_VERSION="0.7.0"
readonly DOCKER_COMPOSE_VERSION="2.32.0"
readonly POSTGRES_CLIENT_VERSION="16"

# ─── Color Output ────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# ─── State ───────────────────────────────────────────────────────────────
INSTALL_GO=false
INSTALL_NODE=false
INSTALL_DOCKER=false
AUTO_YES=false
CHECK_ONLY=false
HAS_CHANGES=false
PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

# ─── Help ────────────────────────────────────────────────────────────────
show_help() {
  sed -n '2,20p' "$0"
  exit 0
}

echo "Bootstrap complete. Use 'bash scripts/sandbox-start.sh' to start the local sandbox."

# Add a convenience flag for non-interactive auto mode
if [[ "${1:-}" == "--auto" || "${1:-}" == "--yes" ]]; then
  echo "Auto mode: installing optional tools non-interactively (if missing)."
  # Minimal safe checks: docker, docker compose, psql
  if ! command -v docker &>/dev/null; then
    echo "docker not found. Please install Docker and re-run this script in auto mode."
  fi
  if ! command -v docker-compose &>/dev/null && ! docker compose version &>/dev/null; then
    echo "docker compose not found. Please install Docker Compose."
  fi
  if ! command -v psql &>/dev/null; then
    echo "psql client not found. Installing postgresql-client is recommended." 
  fi
fi

# ─── Parse arguments ─────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --yes) AUTO_YES=true; shift ;;
    --install-go) INSTALL_GO=true; shift ;;
    --install-node) INSTALL_NODE=true; shift ;;
    --install-docker) INSTALL_DOCKER=true; shift ;;
    --check-only) CHECK_ONLY=true; shift ;;
    --help|-h) show_help ;;
    *) echo "Unknown option: $1"; show_help ;;
  esac
done

# ─── Utility Functions ───────────────────────────────────────────────────

log_info()  { echo -e "${BLUE}ℹ${NC} $*"; }
log_ok()    { echo -e "${GREEN}✓${NC} $*"; }
log_warn()  { echo -e "${YELLOW}⚠${NC} $*"; }
log_error() { echo -e "${RED}✗${NC} $*"; }
log_bold()  { echo -e "${BOLD}$*${NC}"; }
log_step()  { echo; echo -e "${CYAN}==>${NC} ${BOLD}$*${NC}"; }
log_section() { echo; echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"; echo -e "${BOLD}  $*${NC}"; echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"; }

confirm() {
  if [[ "$AUTO_YES" == "true" ]]; then
    return 0
  fi
  local prompt="$1"
  local default="${2:-n}"
  local yn
  read -p "$(echo -e "${YELLOW}?${NC} $prompt [y/N] ")" yn
  case "$yn" in
    [Yy]*) return 0 ;;
    *) return 1 ;;
  esac
}

run_cmd() {
  local desc="$1"
  shift
  echo -e "  ${BLUE}→${NC} $desc..."
  "$@" 2>&1 | sed 's/^/    /' || {
    log_error "Command failed: $*"
    return 1
  }
}

# ─── OS Detection ────────────────────────────────────────────────────────
detect_os() {
  log_section "System Detection"

  OS="$(uname -s)"
  ARCH="$(uname -m)"
  
  # Detect distribution codename (Linux Mint needs Ubuntu codename for Docker repos)
  OS_CODENAME="$(lsb_release -cs 2>/dev/null || echo 'unknown')"
  UBUNTU_CODENAME="$(grep -oP 'UBUNTU_CODENAME=\K.*' /etc/os-release 2>/dev/null || echo '')"
  DOCKER_CODENAME="${UBUNTU_CODENAME:-$OS_CODENAME}"

  case "$OS" in
    Linux)
      if command -v apt-get &>/dev/null; then
        PKG_MANAGER="apt"
      elif command -v dnf &>/dev/null; then
        PKG_MANAGER="dnf"
      elif command -v yum &>/dev/null; then
        PKG_MANAGER="yum"
      elif command -v pacman &>/dev/null; then
        PKG_MANAGER="pacman"
      elif command -v zypper &>/dev/null; then
        PKG_MANAGER="zypper"
      else
        PKG_MANAGER="unknown"
      fi
      ;;
    Darwin)
      PKG_MANAGER="brew"
      ;;
    *)
      log_error "Unsupported OS: $OS"
      exit 1
      ;;
  esac

  log_ok "OS: ${BOLD}$OS${NC} | Arch: ${BOLD}$ARCH${NC} | Package Manager: ${BOLD}$PKG_MANAGER${NC}"
}

# ─── Prerequisite Checks ─────────────────────────────────────────────────

check_tool() {
  local name="$1"
  local binary="$2"
  local version_cmd="$3"
  local required_version="$4"

  if command -v "$binary" &>/dev/null; then
    local version
    version="$(eval "$version_cmd" 2>/dev/null || echo "unknown")"
    log_ok "${name}: ${version} (required: ${required_version})"
    return 0
  else
    log_warn "${name}: NOT FOUND (required: ${required_version})"
    return 1
  fi
}

check_prerequisites() {
  log_section "Prerequisite Check"

  local all_present=true

  # Core prerequisites
  check_tool "Docker" "docker" "docker --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+'" "${DOCKER_COMPOSE_VERSION%%.*}.x" || all_present=false
  # Docker Compose is a CLI plugin, not a standalone binary — check via docker subcommand
  if docker compose version &>/dev/null; then
    local dc_version
    dc_version="$(docker compose version --short 2>/dev/null | grep -oP '^\d+\.\d+\.\d+')"
    log_ok "Docker Compose: ${dc_version} (required: ${DOCKER_COMPOSE_VERSION})"
  else
    log_warn "Docker Compose: NOT FOUND (required: ${DOCKER_COMPOSE_VERSION})"
    all_present=false
  fi
  check_tool "Go" "go" "go version 2>/dev/null | grep -oP 'go\S+' | tr -d 'go'" "${GO_VERSION}" || all_present=false
  check_tool "Node.js" "node" "node --version 2>/dev/null | tr -d 'v'" "${NODE_MAJOR}.x" || all_present=false
  check_tool "pnpm" "pnpm" "pnpm --version 2>/dev/null" "${PNPM_VERSION}.x" || all_present=false
  check_tool "psql" "psql" "psql --version 2>/dev/null | grep -oP '\d+\.\d+' | head -1" "${POSTGRES_CLIENT_VERSION}.x" || all_present=false

  # Optional but recommended
  log_info ""
  log_info "Optional tools (recommended):"

  check_tool "golangci-lint" "golangci-lint" "golangci-lint version 2>/dev/null | grep -oP '\d+\.\d+\.\d+'" "${GOLANGCI_LINT_VERSION}" || true
  check_tool "sqlc" "sqlc" "sqlc version 2>/dev/null | grep -oP '\d+\.\d+\.\d+'" "${SQLC_VERSION}" || true
  check_tool "gofumpt" "gofumpt" "gofumpt --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+'" "${GOFUMPT_VERSION}" || true

  # Docker group membership
  if command -v docker &>/dev/null; then
    if groups | grep -q docker; then
      log_ok "Docker: user is in docker group"
    else
      log_warn "Docker: user NOT in docker group (need sudo for docker commands)"
    fi
  fi

  # Git
  if command -v git &>/dev/null; then
    log_ok "Git: $(git --version)"
    local git_name
    git_name="$(git config user.name 2>/dev/null || echo 'NOT SET')"
    local git_email
    git_email="$(git config user.email 2>/dev/null || echo 'NOT SET')"
    log_info "  Git user: ${git_name} <${git_email}>"
  else
    log_warn "Git: NOT FOUND"
    all_present=false
  fi

  if [[ "$CHECK_ONLY" == "true" ]]; then
    echo
    if [[ "$all_present" == "true" ]]; then
      log_bold "All core prerequisites satisfied!"
    else
      log_warn "Some prerequisites are missing. Run without --check-only to install."
    fi
    exit 0
  fi

  echo
  return 0
}

# ─── Docker Setup ────────────────────────────────────────────────────────

install_docker() {
  log_section "Docker Setup"

  if command -v docker &>/dev/null && docker compose version &>/dev/null; then
    log_ok "Docker already installed"
    return 0
  fi

  if [[ "$INSTALL_DOCKER" != "true" ]] && ! confirm "Install Docker and Docker Compose?"; then
    log_warn "Skipping Docker installation"
    return 0
  fi

  case "$OS" in
    Linux)
      case "$PKG_MANAGER" in
        apt)
          run_cmd "Updating package list" sudo apt-get update -qq
          run_cmd "Installing Docker dependencies" sudo apt-get install -y -qq ca-certificates curl gnupg lsb-release
          sudo mkdir -p /etc/apt/keyrings
          curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg 2>/dev/null
          echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu ${DOCKER_CODENAME} stable" | sudo tee /etc/apt/sources.list.d/docker.list >/dev/null
          sudo apt-get update -qq
          run_cmd "Installing Docker" sudo apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-compose-plugin
          ;;
        dnf|yum)
          run_cmd "Installing Docker" sudo dnf install -y dnf-plugins-core
          sudo dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo
          sudo dnf install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
          sudo systemctl enable docker
          sudo systemctl start docker
          ;;
        pacman)
          run_cmd "Installing Docker" sudo pacman -S --noconfirm docker docker-compose
          sudo systemctl enable docker
          sudo systemctl start docker
          ;;
        *)
          log_error "Unsupported package manager for automatic Docker installation"
          log_info "Please install Docker manually: https://docs.docker.com/engine/install/"
          return 1
          ;;
      esac
      ;;
    Darwin)
      if ! command -v brew &>/dev/null; then
        log_info "Installing Homebrew first..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
      fi
      brew install --cask docker
      log_info "Docker installed. Please open Docker.app manually to complete setup."
      ;;
  esac

  # Add user to docker group
  if ! groups | grep -q docker; then
    sudo usermod -aG docker "$USER"
    log_warn "User added to docker group. Log out and back in for this to take effect."
    log_info "Alternatively, run: 'newgrp docker' to use docker immediately."
  fi

  log_ok "Docker installed: $(docker --version 2>/dev/null || echo 'restart shell to verify')"
  HAS_CHANGES=true
}

# ─── Go Installation ─────────────────────────────────────────────────────

install_go() {
  log_section "Go ${GO_VERSION} Setup"

  if command -v go &>/dev/null; then
    local current_version
    current_version="$(go version | grep -oP 'go\K\S+')"
    if [[ "$current_version" == "$GO_VERSION" ]]; then
      log_ok "Go ${GO_VERSION} already installed"
      return 0
    fi
    if [[ "$INSTALL_GO" != "true" ]] && ! confirm "Upgrade Go from ${current_version} to ${GO_VERSION}?"; then
      log_warn "Skipping Go upgrade"
      return 0
    fi
  else
    if [[ "$INSTALL_GO" != "true" ]] && ! confirm "Install Go ${GO_VERSION}?"; then
      log_warn "Skipping Go installation"
      return 0
    fi
  fi

  local go_tarball="go${GO_VERSION}.linux-amd64.tar.gz"
  case "$ARCH" in
    aarch64|arm64) go_tarball="go${GO_VERSION}.linux-arm64.tar.gz" ;;
    x86_64|amd64) go_tarball="go${GO_VERSION}.linux-amd64.tar.gz" ;;
  esac
  if [[ "$OS" == "Darwin" ]]; then
    case "$ARCH" in
      arm64) go_tarball="go${GO_VERSION}.darwin-arm64.tar.gz" ;;
      x86_64) go_tarball="go${GO_VERSION}.darwin-amd64.tar.gz" ;;
    esac
  fi

  # Remove old Go
  if [[ -d /usr/local/go ]]; then
    run_cmd "Removing old Go installation" sudo rm -rf /usr/local/go
  fi

  run_cmd "Downloading Go ${GO_VERSION}" \
    curl -fsSL "${GO_DOWNLOAD_BASE}/${go_tarball}" -o "/tmp/${go_tarball}"

  run_cmd "Extracting Go" sudo tar -C /usr/local -xzf "/tmp/${go_tarball}"
  rm -f "/tmp/${go_tarball}"

  # Add to PATH if not present (prepend so /usr/local/go/bin takes priority)
  if ! grep -q '/usr/local/go/bin' "$HOME/.profile" 2>/dev/null; then
    echo 'export PATH=/usr/local/go/bin:$PATH' >> "$HOME/.profile"
    log_info "Added Go to PATH in ~/.profile"
  fi
  export PATH="/usr/local/go/bin:$PATH"

  log_ok "Go ${GO_VERSION} installed successfully"
  HAS_CHANGES=true
}

# ─── Go Tools ────────────────────────────────────────────────────────────

install_go_tools() {
  log_section "Go Tools"

  if ! command -v go &>/dev/null; then
    log_warn "Go not found, skipping Go tools"
    return 0
  fi

  # golangci-lint
  if command -v golangci-lint &>/dev/null && [[ "$INSTALL_GO" != "true" ]]; then
    log_ok "golangci-lint already installed"
  else
    if command -v golangci-lint &>/dev/null; then
      confirm "Install/upgrade golangci-lint?" || { log_info "Skipping golangci-lint"; }
    else
      confirm "Install golangci-lint (linter)?" || { log_info "Skipping golangci-lint"; }
    fi
    if command -v golangci-lint &>/dev/null; then
      log_info "golangci-lint already available, skipping install"
    else
      log_info "Installing golangci-lint..."
      curl -fsSL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" "v${GOLANGCI_LINT_VERSION}" 2>&1 | sed 's/^/    /' && \
        log_ok "golangci-lint ${GOLANGCI_LINT_VERSION} installed" || \
        log_info "Skipping golangci-lint (install failed)"
    fi
  fi

  # sqlc
  if ! command -v sqlc &>/dev/null || [[ "$INSTALL_GO" == "true" ]]; then
    if command -v sqlc &>/dev/null; then
      if ! confirm "Install/upgrade sqlc?"; then
        log_info "Skipping sqlc"
      else
        run_cmd "Installing sqlc" go install "github.com/sqlc-dev/sqlc/cmd/sqlc@v${SQLC_VERSION}"
        log_ok "sqlc ${SQLC_VERSION} installed"
      fi
    else
      confirm "Install sqlc (SQL code generator)?" && \
        run_cmd "Installing sqlc" go install "github.com/sqlc-dev/sqlc/cmd/sqlc@v${SQLC_VERSION}" && \
        log_ok "sqlc ${SQLC_VERSION} installed" || \
        log_info "Skipping sqlc"
    fi
  else
    log_ok "sqlc already installed"
  fi

  # gofumpt
  if ! command -v gofumpt &>/dev/null || [[ "$INSTALL_GO" == "true" ]]; then
    confirm "Install gofumpt (Go formatter)?" && \
      run_cmd "Installing gofumpt" go install "mvdan.cc/gofumpt@v${GOFUMPT_VERSION}" && \
      log_ok "gofumpt ${GOFUMPT_VERSION} installed" || \
      log_info "Skipping gofumpt"
  else
    log_ok "gofumpt already installed"
  fi

  HAS_CHANGES=true
}

# ─── Node.js Setup via pnpm ─────────────────────────────────────────────

install_node() {
  log_section "Node.js ${NODE_MAJOR}.x LTS + pnpm ${PNPM_VERSION} Setup"

  if command -v node &>/dev/null; then
    local current_major
    current_major="$(node --version | tr -d 'v' | cut -d. -f1)"
    if [[ "$current_major" == "$NODE_MAJOR" ]]; then
      log_ok "Node.js ${NODE_MAJOR}.x already installed ($(node --version))"
    elif [[ "$INSTALL_NODE" != "true" ]] && ! confirm "Install Node.js ${NODE_MAJOR}.x LTS (current: $(node --version))?"; then
      log_warn "Skipping Node.js installation"
      return 0
    else
      install_node_via_pnpm
    fi
  else
    if [[ "$INSTALL_NODE" != "true" ]] && ! confirm "Install Node.js ${NODE_MAJOR}.x LTS?"; then
      log_warn "Skipping Node.js installation"
      return 0
    fi
    install_node_via_pnpm
  fi

  # Ensure pnpm is available
  if command -v pnpm &>/dev/null; then
    local pnpm_ver
    pnpm_ver="$(pnpm --version)"
    log_ok "pnpm ${pnpm_ver} already available"
  else
    log_info "Installing pnpm via corepack..."
    corepack enable 2>/dev/null || npm install -g corepack
    corepack prepare "pnpm@${PNPM_VERSION}" --activate
    log_ok "pnpm ${PNPM_VERSION} installed: $(pnpm --version)"
  fi

  HAS_CHANGES=true
}

install_node_via_pnpm() {
  # Install Node.js using pnpm's built-in env management
  if ! command -v pnpm &>/dev/null; then
    # Bootstrap pnpm if needed
    npm install -g pnpm 2>/dev/null || true
  fi

  if command -v pnpm &>/dev/null; then
    # Ensure pnpm's global bin is in PATH
    eval "$(pnpm setup 2>/dev/null)" || true
    run_cmd "Installing Node.js ${NODE_MAJOR}.x via pnpm" \
      pnpm env use --global "${NODE_MAJOR}" 2>/dev/null || \
      pnpm install -g "node@${NODE_MAJOR}" 2>/dev/null || true
  fi

  # Fallback to nvm
  if ! command -v node &>/dev/null || [[ "$(node --version | tr -d 'v' | cut -d. -f1)" != "$NODE_MAJOR" ]]; then
    log_info "Using nvm as fallback..."
    export NVM_DIR="$HOME/.nvm"
    if [[ ! -d "$NVM_DIR" ]]; then
      curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.2/install.sh | bash
    fi
    [[ -s "$NVM_DIR/nvm.sh" ]] && source "$NVM_DIR/nvm.sh"
    nvm install "${NODE_MAJOR}"
    nvm alias default "${NODE_MAJOR}"
    nvm use default
  fi

  log_ok "Node.js $(node --version) installed"
}

# ─── PostgreSQL Client ──────────────────────────────────────────────────

install_postgres_client() {
  log_section "PostgreSQL Client"

  if command -v psql &>/dev/null; then
    log_ok "psql already installed: $(psql --version)"
    return 0
  fi

  if ! confirm "Install PostgreSQL client (psql)?"; then
    log_warn "Skipping psql installation"
    return 0
  fi

  case "$PKG_MANAGER" in
    apt)
      sudo sh -c "echo 'deb http://apt.postgresql.org/pub/repos/apt ${DOCKER_CODENAME}-pgdg main' > /etc/apt/sources.list.d/pgdg.list"
      curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo gpg --dearmor -o /etc/apt/trusted.gpg.d/postgresql.gpg
      sudo apt-get update -qq
      sudo apt-get install -y -qq "postgresql-client-${POSTGRES_CLIENT_VERSION}"
      ;;
    dnf|yum)
      sudo dnf install -y "postgresql${POSTGRES_CLIENT_VERSION/./}-server"
      ;;
    brew)
      brew install postgresql@"${POSTGRES_CLIENT_VERSION}"
      ;;
    *)
      log_warn "Automatic psql installation not supported for your package manager"
      log_info "Install manually: https://www.postgresql.org/download/"
      return 1
      ;;
  esac

  log_ok "PostgreSQL client $(psql --version) installed"
  HAS_CHANGES=true
}

# ─── Git Configuration ──────────────────────────────────────────────────

setup_git() {
  log_section "Git Configuration"

  if ! command -v git &>/dev/null; then
    log_error "Git not found. Please install git manually."
    return 1
  fi

  # Check git user config
  local git_name git_email
  git_name="$(git config user.name 2>/dev/null || echo '')"
  git_email="$(git config user.email 2>/dev/null || echo '')"

  if [[ -z "$git_name" ]]; then
    read -p "$(echo -e "${YELLOW}?${NC} Enter your Git user name: ")" git_name
    git config user.name "$git_name"
    log_ok "Git user.name set to: $git_name"
  fi

  if [[ -z "$git_email" ]]; then
    read -p "$(echo -e "${YELLOW}?${NC} Enter your Git email: ")" git_email
    git config user.email "$git_email"
    log_ok "Git user.email set to: $git_email"
  fi

  # Set up git hooks directory
  git config core.hooksPath .githooks 2>/dev/null || true
  mkdir -p "${PROJECT_DIR}/.githooks"

  # Create pre-commit hook for Go vet
  cat > "${PROJECT_DIR}/.githooks/pre-commit" << 'HOOK'
#!/usr/bin/env bash
set -euo pipefail

echo "Running pre-commit checks..."

# Go vet
if command -v go &>/dev/null && [[ -f apps/api/go.mod ]]; then
  echo "  → Go vet..."
  cd apps/api
  go vet ./... 2>&1 | sed 's/^/    /' || exit 1
  cd "$OLDPWD"
fi

# Check for merge conflicts
if grep -r "^<<<<<<< " --include="*.go" --include="*.ts" --include="*.tsx" --include="*.md" --include="*.sql" . 2>/dev/null; then
  echo "ERROR: Merge conflict markers found!"
  exit 1
fi

echo "  ✓ All checks passed"
HOOK
  chmod +x "${PROJECT_DIR}/.githooks/pre-commit"
  log_ok "Git pre-commit hook installed"

  log_ok "Git configured: ${git_name:-$(git config user.name)} <${git_email:-$(git config user.email)}>"
}

# ─── Project Environment Setup ─────────────────────────────────────────

setup_project_env() {
  log_section "Project Environment Setup"

  # Create .env from .env.example if not exists
  if [[ ! -f "${PROJECT_DIR}/.env" ]]; then
    if [[ -f "${PROJECT_DIR}/.env.example" ]]; then
      cp "${PROJECT_DIR}/.env.example" "${PROJECT_DIR}/.env"
      log_ok ".env created from .env.example"
    else
      log_warn ".env.example not found; skipping .env creation"
    fi
  else
    log_ok ".env already exists"
  fi

  # Ensure data directories exist
  mkdir -p "${PROJECT_DIR}/data/postgres" "${PROJECT_DIR}/data/redis" "${PROJECT_DIR}/data/minio" "${PROJECT_DIR}/data/meilisearch" 2>/dev/null || true

  # Install frontend dependencies
  if [[ -f "${PROJECT_DIR}/apps/web/package.json" ]]; then
    log_info "Installing frontend dependencies..."
    if command -v pnpm &>/dev/null; then
      cd "${PROJECT_DIR}/apps/web" && pnpm install --frozen-lockfile 2>/dev/null || pnpm install 2>/dev/null || log_warn "pnpm install failed (will retry later)"
      cd "${PROJECT_DIR}"
      log_ok "Frontend dependencies installed"
    else
      log_warn "pnpm not available; skip frontend deps"
    fi
  fi

  log_ok "Project environment setup complete"
  HAS_CHANGES=true
}

# ─── Docker Environment Setup ───────────────────────────────────────────

setup_docker_env() {
  log_section "Docker Environment"

  if ! command -v docker &>/dev/null; then
    log_warn "Docker not available; skip Docker environment setup"
    return 0
  fi

  # Check if Docker daemon is running
  if docker info &>/dev/null; then
    log_ok "Docker daemon is running"
  else
    log_warn "Docker daemon is not running. Start it manually:"
    log_info "  sudo systemctl start docker   (Linux)"
    log_info "  open -a Docker                 (macOS)"
    if confirm "Attempt to start Docker daemon?"; then
      sudo systemctl start docker 2>/dev/null || log_warn "Could not start Docker automatically"
    fi
  fi

  # Pull required images
  if docker info &>/dev/null; then
    log_info "Pulling Docker images (this may take a while)..."
    local images=(
      "postgres:18-alpine"
      "redis:8-alpine"
      "minio/minio:latest"
      "quay.io/keycloak/keycloak:26.0"
      "getmeili/meilisearch:latest"
    )
    for img in "${images[@]}"; do
      if [[ "$(docker images -q "$img" 2>/dev/null)" == "" ]]; then
        log_info "  Pulling ${img}..."
        docker pull "$img" 2>/dev/null || log_warn "  Failed to pull ${img}"
      fi
    done
    log_ok "Docker images pulled"
  fi
}

# ─── VS Code Setup ──────────────────────────────────────────────────────

setup_vscode() {
  log_section "VS Code Configuration"

  if command -v code &>/dev/null || command -v codium &>/dev/null; then
    local code_cmd="code"
    command -v codium &>/dev/null && code_cmd="codium"

    log_ok "VS Code detected: $($code_cmd --version 2>/dev/null | head -1)"

    # Install extensions (non-blocking)
    log_info "Installing recommended extensions..."
    local extensions=(
      "golang.go"
      "esbenp.prettier-vscode"
      "dbaeumer.vscode-eslint"
      "bradlc.vscode-tailwindcss"
      "mtxr.sqltools"
      "mtxr.sqltools-driver-pg"
      "tamasfe.even-better-toml"
      "streetsidesoftware.code-spell-checker"
      "ms-azuretools.vscode-docker"
      "redhat.vscode-yaml"
      "yzhang.markdown-all-in-one"
      "pkief.material-icon-theme"
      "github.vscode-github-actions"
    )

    for ext in "${extensions[@]}"; do
      $code_cmd --install-extension "$ext" --force 2>/dev/null || true
    done
    log_ok "VS Code extensions installed"
  else
    log_warn "VS Code CLI not found. Install VS Code from: https://code.visualstudio.com/"
    log_info "Extensions are listed in .vscode/extensions.json for manual installation"
  fi
}

# ─── Summary Reports ────────────────────────────────────────────────────

print_summary() {
  log_section "Setup Summary"

  echo ""
  echo -e "  ${BOLD}ZarishLog Development Sandbox${NC}"
  echo ""
  echo -e "  ${BOLD}Project:${NC}   ${PROJECT_DIR}"
  echo ""

  # Tool versions
  local tools=(
    "Go:$(go version 2>/dev/null | grep -oP 'go\S+' || echo 'NOT INSTALLED')"
    "Node.js:$(node --version 2>/dev/null || echo 'NOT INSTALLED')"
    "pnpm:$(pnpm --version 2>/dev/null || echo 'NOT INSTALLED')"
    "Docker:$(docker --version 2>/dev/null || echo 'NOT INSTALLED')"
    "Docker Compose:$(docker compose version --short 2>/dev/null | grep -oP '^\d+\.\d+\.\d+' || echo 'NOT INSTALLED')"
    "psql:$(psql --version 2>/dev/null || echo 'NOT INSTALLED')"
    "golangci-lint:$(golangci-lint version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo 'NOT INSTALLED')"
    "sqlc:$(sqlc version 2>/dev/null || echo 'NOT INSTALLED')"
    "gofumpt:$(gofumpt --version 2>/dev/null || echo 'NOT INSTALLED')"
  )

  for tool in "${tools[@]}"; do
    local name="${tool%%:*}"
    local ver="${tool#*:}"
    if [[ "$ver" == "NOT INSTALLED" ]]; then
      echo -e "  ${YELLOW}⚠${NC} ${name}: ${RED}${ver}${NC}"
    else
      echo -e "  ${GREEN}✓${NC} ${name}: ${ver}"
    fi
  done

  echo ""
  echo -e "  ${BOLD}Next Steps:${NC}"
  echo ""
  echo -e "  1. ${BLUE}make docker-up${NC}     — Start all infrastructure services"
  echo -e "  2. ${BLUE}make db-migrate${NC}   — Run database migrations"
  echo -e "  3. ${BLUE}make db-seed${NC}      — Seed master data"
  echo -e "  4. ${BLUE}make dev${NC}          — Start development servers"
  echo -e "  Or: ${BLUE}bash scripts/sandbox-start.sh${NC} — Start full sandbox (recommended for new users)"
  echo ""
  echo -e "  Or open VS Code and use the built-in Tasks (Ctrl+Shift+P → Tasks: Run Task)"
  echo ""
}

# ─── Main Execution ──────────────────────────────────────────────────────

main() {
  echo ""
  echo -e "${BOLD}╔═══════════════════════════════════════════════════════════════╗${NC}"
  echo -e "${BOLD}║            ZarishLog Development Sandbox Setup              ║${NC}"
  echo -e "${BOLD}║        Version: ${GO_VERSION} · Node: ${NODE_MAJOR}.x · pnpm: ${PNPM_VERSION}.x        ║${NC}"
  echo -e "${BOLD}╚═══════════════════════════════════════════════════════════════╝${NC}"
  echo ""

  detect_os
  check_prerequisites

  # Sequential setup steps
  install_docker
  install_go
  install_go_tools
  install_postgres_client
  install_node
  setup_git
  setup_project_env
  setup_docker_env
  setup_vscode

  print_summary

  if [[ "$HAS_CHANGES" == "true" ]]; then
    log_info "Some components were installed. Open a new terminal for PATH changes to take effect."
  fi

  log_bold "Setup complete!"
}

main "$@"
