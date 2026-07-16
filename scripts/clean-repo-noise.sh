#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

ARCH_DIR=".archive/cleanup-$(date +%F-%H%M%S)"
mkdir -p "$ARCH_DIR"

auto=false
if [[ "${1:-}" == "--auto" || "${1:-}" == "-y" ]]; then
  auto=true
fi

# Find large files > 5MB (adjustable)
mapfile -t files < <(find . -type f -size +5M -not -path "./.git/*" -not -path "./node_modules/*" -not -path "./.archive/*" -print)

if [[ ${#files[@]} -eq 0 ]]; then
  echo "No large files found to archive."
  exit 0
fi

echo "Found ${#files[@]} large files (>$((5))M):"
for f in "${files[@]}"; do
  ls -lh "$f"
done

if [[ "$auto" == "false" ]]; then
  read -p "Archive these files to ${ARCH_DIR}? [y/N] " yn
  case "$yn" in
    [Yy]*) : ;;
    *) echo "Aborted."; exit 1 ;;
  esac
fi

for f in "${files[@]}"; do
  dest="$ARCH_DIR/$(echo "$f" | sed 's#\./##' | tr '/ ' '_' )"
  mkdir -p "$(dirname "$dest")" || true
  mv "$f" "$dest" || true
  echo "Moved $f -> $dest"
done

# Ensure .archive is in .gitignore
if ! grep -q "^\.archive/" .gitignore 2>/dev/null; then
  echo ".archive/" >> .gitignore
  git add .gitignore || true
  git commit -m "chore: add .archive to .gitignore" || true
fi

echo "Archived files to ${ARCH_DIR}" 
