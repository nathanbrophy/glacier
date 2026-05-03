#!/usr/bin/env bash
# check-site.sh - orchestrator for every static check this site enforces.
# Runs each scripts/check-*.sh in turn; non-zero exit on any failure.
#
# Run from /site:  bash scripts/check-site.sh
# Run from repo root: bash site/scripts/check-site.sh

set -eu
cd "$(dirname "$0")/.."

CHECKS=(
  "check-action-pins.sh"
  "check-base-path.sh"
  "check-extraction-directives.sh"
  "check-no-cdn.sh"
  "check-no-telemetry.sh"
  "check-spec-links.sh"
  "check-sprite-budget.sh"
  "check-sprite-placement.sh"
  "check-svg-safety.sh"
  "check-voice.sh"
)

PASS=0
FAIL=0
FAILED=()

echo "=== glacier site checks ==="
for c in "${CHECKS[@]}"; do
  printf -- "--- %s ---\n" "$c"
  if bash "scripts/$c"; then
    PASS=$((PASS + 1))
  else
    FAIL=$((FAIL + 1))
    FAILED+=("$c")
  fi
  echo
done

echo "=== summary: ${PASS} pass, ${FAIL} fail ==="
if [ "$FAIL" -gt 0 ]; then
  echo "failed checks:"
  for f in "${FAILED[@]}"; do
    echo "  - $f"
  done
  exit 1
fi
