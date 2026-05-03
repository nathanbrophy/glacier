#!/usr/bin/env bash
# Spec 0031 §Migration Amendment A: each companion sprite SVG must be
# <= 4096 bytes gzipped.
set -eu
cd "$(dirname "$0")/.."

LIMIT=4096
fail=0

for state in idle wave thinking; do
  file="public/mascot/companion-${state}.svg"
  if [ ! -f "$file" ]; then
    echo "SKIP $file (not yet authored)"
    continue
  fi
  size=$(gzip -c "$file" | wc -c | tr -d ' ')
  if [ "$size" -gt "$LIMIT" ]; then
    echo "FAIL $file gzipped ${size}B (> ${LIMIT}B budget)"
    fail=1
  else
    echo "OK   $file gzipped ${size}B"
  fi
done

exit $fail
