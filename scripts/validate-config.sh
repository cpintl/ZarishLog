#!/usr/bin/env bash
# ============================================================================
# ZarishLog — Configuration Validator
# ============================================================================
# Validates all configuration files (CSV, JSON, Markdown) for correctness.
# Run this after editing config files to catch errors before loading.
#
# Usage:
#   ./scripts/validate-config.sh           # Validate all configs
#   ./scripts/validate-config.sh --csv     # CSV files only
#   ./scripts/validate-config.sh --json    # JSON files only
#   ./scripts/validate-config.sh --fix     # Auto-fix common issues
# ============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CHECK_CSV=false
CHECK_JSON=false
CHECK_ALL=true
AUTO_FIX=false
ERRORS=0
WARNINGS=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --csv) CHECK_CSV=true; CHECK_ALL=false; shift ;;
    --json) CHECK_JSON=true; CHECK_ALL=false; shift ;;
    --fix) AUTO_FIX=true; shift ;;
    --help) echo "Usage: $0 [--csv|--json] [--fix]"; exit 0 ;;
    *) echo "Unknown: $1"; exit 1 ;;
  esac
done

if [[ "$CHECK_ALL" == "true" ]]; then
  CHECK_CSV=true
  CHECK_JSON=true
fi

report_error()   { echo -e "  ${RED}✗${NC} ERROR: $1"; ERRORS=$((ERRORS+1)); }
report_warning() { echo -e "  ${YELLOW}⚠${NC} WARN: $1"; WARNINGS=$((WARNINGS+1)); }
report_pass()    { echo -e "  ${GREEN}✓${NC} $1"; }
report_info()    { echo -e "  ${BLUE}ℹ${NC} $1"; }

echo ""
echo -e "${BOLD}╔═══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║           ZarishLog Configuration Validator                   ║${NC}"
echo -e "${BOLD}╚═══════════════════════════════════════════════════════════════╝${NC}"
echo ""

# ─── Validate CSV Files ─────────────────────────────────────────────────

validate_csv() {
  local file="$1"
  local name="$2"
  local required_headers="$3"
  
  if [[ ! -f "$file" ]]; then
    report_error "${name}: File not found: $file"
    return
  fi
  
  # Check file is not empty
  if [[ ! -s "$file" ]]; then
    report_error "${name}: File is empty"
    return
  fi
  
  # Read headers
  local headers
  headers="$(head -1 "$file")"
  
  # Check required headers
  local missing=""
  IFS=',' read -ra REQUIRED <<< "$required_headers"
  for header in "${REQUIRED[@]}"; do
    if ! echo "$headers" | grep -q "$header"; then
      missing="$missing $header"
    fi
  done
  
  if [[ -n "$missing" ]]; then
    report_error "${name}: Missing required columns:${missing}"
  else
    report_pass "${name}: Required columns present"
  fi
  
  # Count rows
  local rows
  rows="$(tail -n +2 "$file" | grep -c -v '^\s*$' || true)"
  report_info "${name}: ${rows} data rows"
  
  # Check for duplicate SKUs (products only)
  if echo "$file" | grep -q "product"; then
    local sku_col=1
    if echo "$headers" | grep -qi "sku"; then
      sku_col="$(echo "$headers" | tr ',' '\n' | grep -n -i "sku" | cut -d: -f1)"
      local dupes
      dupes="$(tail -n +2 "$file" | cut -d, -f"$sku_col" | sort | uniq -d | grep -v '^\s*$' || true)"
      if [[ -n "$dupes" ]]; then
        report_warning "${name}: Duplicate SKUs found:"
        echo "$dupes" | sed 's/^/      - /'
      fi
    fi
  fi
}

# ─── Validate JSON Files ────────────────────────────────────────────────

validate_json() {
  local file="$1"
  local name="$2"
  
  if [[ ! -f "$file" ]]; then
    report_error "${name}: File not found: $file"
    return
  fi
  
  # Validate JSON syntax
  if jq empty "$file" 2>/dev/null; then
    report_pass "${name}: Valid JSON"
  else
    local error
    error="$(jq empty "$file" 2>&1 || true)"
    report_error "${name}: Invalid JSON: $error"
    return
  fi
  
  # File size
  local size
  size="$(stat -c%s "$file" 2>/dev/null || stat -f%z "$file" 2>/dev/null)"
  report_info "${name}: $(numfmt --to=iec $size 2>/dev/null || echo "${size}B")"
}

# ─── Validate Environment ──────────────────────────────────────────────

validate_env() {
  local file="${PROJECT_DIR}/.env"
  
  if [[ ! -f "$file" ]]; then
    report_warning ".env file not found (create from .env.example)"
    return
  fi
  
  # Check required vars
  local required_vars=("DATABASE_URL" "API_PORT" "OIDC_ISSUER")
  for var in "${required_vars[@]}"; do
    if ! grep -q "^${var}=" "$file" 2>/dev/null; then
      report_warning ".env: Missing required variable: $var"
    fi
  done
  
  report_pass ".env: Present with $(grep -c '=' "$file" 2>/dev/null || echo 0) variables"
}

# ─── Main Validation ────────────────────────────────────────────────────

echo -e "${BOLD}Validating configuration files...${NC}"
echo ""

# CSV files
if [[ "$CHECK_CSV" == "true" ]]; then
  echo -e "${BOLD}── CSV Files ──${NC}"
  echo ""
  
  validate_csv "${PROJECT_DIR}/config/metadata/master_product_list.csv" "Product Catalogue" "sku,name,category_name,uom_abbreviation,item_type"
  validate_csv "${PROJECT_DIR}/config/metadata/organization.csv" "Organization Hierarchy" "name,code,level"
  validate_csv "${PROJECT_DIR}/config/metadata/programs.csv" "Programs" "code,name"
  validate_csv "${PROJECT_DIR}/config/metadata/uom.csv" "Units of Measure" "name,abbreviation,category"
  
  echo ""
fi

# JSON files
if [[ "$CHECK_JSON" == "true" ]]; then
  echo -e "${BOLD}── JSON Files ──${NC}"
  echo ""
  
  validate_json "${PROJECT_DIR}/config/location/warehouse.json" "Warehouse Configuration"
  validate_json "${PROJECT_DIR}/config/templates/goods_receipt_form.json" "GRN Form Template"
  validate_json "${PROJECT_DIR}/config/templates/stock_issue_form.json" "Stock Issue Form Template"
  
  echo ""
fi

# Environment
echo -e "${BOLD}── Environment ──${NC}"
echo ""
validate_env
echo ""

# ─── Summary ────────────────────────────────────────────────────────────

echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

if [[ "$ERRORS" -eq 0 && "$WARNINGS" -eq 0 ]]; then
  echo -e "  ${GREEN}✓ All configuration files valid!${NC}"
elif [[ "$ERRORS" -eq 0 ]]; then
  echo -e "  ${YELLOW}✓ Valid with ${WARNINGS} warning(s)${NC}"
  echo -e "  ${BLUE}ℹ Warnings are non-critical but should be reviewed${NC}"
else
  echo -e "  ${RED}✗ ${ERRORS} error(s) and ${WARNINGS} warning(s) found${NC}"
  echo -e "  ${BLUE}ℹ Fix errors before running 'make db-seed'${NC}"
  exit 1
fi

echo ""
echo -e "  ${BLUE}Next step:${NC} make db-seed (to load configuration into database)"
echo ""
