#!/usr/bin/env bash
set -euo pipefail

# Wrapper: start sandbox and open browser tabs for main services
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

bash scripts/sandbox-start.sh

# Give services a moment to bind ports
sleep 2

open_cmd=""
if command -v xdg-open &>/dev/null; then
  open_cmd="xdg-open"
elif command -v gnome-open &>/dev/null; then
  open_cmd="gnome-open"
elif command -v open &>/dev/null; then
  open_cmd="open"
fi

urls=(
  "http://localhost:3000"
  "http://localhost:8080"
  "http://localhost:9001"
  "http://localhost:7700"
)

if [[ -n "$open_cmd" ]]; then
  echo "Opening browser tabs..."
  for u in "${urls[@]}"; do
    echo "  -> $u"
    $open_cmd "$u" || true
  done
else
  echo "No GUI opener found. Open these URLs in your browser:"
  for u in "${urls[@]}"; do echo "  $u"; done
fi

echo "Done."
