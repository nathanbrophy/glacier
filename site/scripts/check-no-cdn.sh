#!/usr/bin/env bash
# Spec 0031 §Security: zero CDN runtime dependencies in built output.
# Greps the built dist/ for any of the well-known CDN origins.
set -eu
cd "$(dirname "$0")/.."

DIST=".vitepress/dist"
if [ ! -d "$DIST" ]; then
  echo "SKIP $DIST not present (run npm run build first)"
  exit 0
fi

PATTERN='fonts\.googleapis\.com|fonts\.gstatic\.com|unpkg\.com|jsdelivr\.net|cdn\.jsdelivr\.net|cdnjs\.cloudflare\.com|fonts\.bunny\.net|polyfill\.io'

# shellcheck disable=SC2155
matches=$(grep -rEn "$PATTERN" "$DIST" 2>/dev/null || true)

if [ -n "$matches" ]; then
  echo "FAIL CDN reference(s) found in built output:"
  echo "$matches"
  exit 1
fi
echo "OK   no CDN references in dist/"
