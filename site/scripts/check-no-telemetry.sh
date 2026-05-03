#!/usr/bin/env bash
# Spec 0031 §Security / D-S27: no telemetry, no analytics, no cookies,
# no fingerprinting.
#
# Patterns are anchored to function-call or object-access syntax to avoid
# false positives from minified JS bundle variable names (e.g. mermaid
# legitimately ships the substring "amplitude" inside waveform math).
# We scan:
#   - All built HTML in dist/ (catches injected <script src=...>)
#   - All source markdown / Vue / TypeScript under site/ (catches
#     deliberate telemetry author error)
# We deliberately do NOT scan minified .js bundles in dist/assets/.
set -eu
cd "$(dirname "$0")/.."

# Anchored patterns: function call (foo(), foo.bar) or URL hostnames.
PATTERN='\bsendBeacon\(|\bgtag\(|\b_gaq\.|\b_ga\(|\bfbq\(|googletagmanager\.|\bmixpanel\.|\bamplitude\.[a-zA-Z_]|segment\.com|\bhj\(|clarity\.ms|\bposthog\.|plausible\.io|plausible\.com|\bfathom-client\b|dataLayer\.push'

DIST=".vitepress/dist"
fail=0

# 1. Built HTML
if [ -d "$DIST" ]; then
  matches=$(grep -rEn "$PATTERN" "$DIST" --include='*.html' 2>/dev/null || true)
  if [ -n "$matches" ]; then
    echo "FAIL telemetry identifier(s) in built HTML:"
    echo "$matches"
    fail=1
  fi
else
  echo "SKIP $DIST not present (run npm run build first)"
fi

# 2. Source files
matches=$(grep -rEn "$PATTERN" . \
  --include='*.md' --include='*.vue' --include='*.ts' \
  --exclude-dir='node_modules' \
  --exclude-dir='.vitepress/dist' \
  --exclude-dir='.vitepress/cache' \
  2>/dev/null || true)
if [ -n "$matches" ]; then
  echo "FAIL telemetry identifier(s) in source:"
  echo "$matches"
  fail=1
fi

if [ "$fail" -eq 0 ]; then
  echo "OK   no telemetry identifiers in HTML or source"
fi
exit $fail
