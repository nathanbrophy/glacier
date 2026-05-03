#!/usr/bin/env bash
# Spec 0001 D11 / spec 0031 §Migration Amendment B: no superlatives in
# committed prose under site/. Scans .md, .vue, .ts source files.
set -eu
cd "$(dirname "$0")/.."

PATTERN='blazing|revolutionary|best-in-class|amazing|seamless'

# shellcheck disable=SC2155
matches=$(grep -rEni "$PATTERN" . \
  --include='*.md' --include='*.vue' --include='*.ts' \
  --exclude-dir='node_modules' \
  --exclude-dir='.vitepress/dist' \
  --exclude-dir='.vitepress/cache' \
  2>/dev/null || true)

if [ -n "$matches" ]; then
  echo "FAIL banned superlative(s) found in site/ source:"
  echo "$matches"
  exit 1
fi
echo "OK   no banned superlatives in site/ source"
