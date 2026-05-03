#!/usr/bin/env bash
# Verify built dist/ uses /glacier/ as the base path consistently.
# Runs after `npm run build`; harmlessly skips if dist/ doesn't exist.
set -eu
cd "$(dirname "$0")/.."

DIST=".vitepress/dist"
if [ ! -d "$DIST" ]; then
  echo "SKIP $DIST not present (run npm run build first)"
  exit 0
fi

# Find <a href="/..."> or src="/..." that don't start with /glacier/.
# Allow: external https?:, anchors (#...), data URIs, and the /glacier/ prefix.
fail=0
matches=$(grep -rEoh 'href="/[^"]*"|src="/[^"]*"' "$DIST" --include='*.html' 2>/dev/null \
  | sort -u \
  | grep -vE '^(href|src)="/glacier/' || true)

if [ -n "$matches" ]; then
  echo "FAIL bare-/ references found in dist (should be /glacier/):"
  echo "$matches"
  fail=1
fi

if [ "$fail" -eq 0 ]; then
  echo "OK   all internal hrefs/srcs use /glacier/ base"
fi
exit $fail
