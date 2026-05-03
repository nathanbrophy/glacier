#!/usr/bin/env bash
# Spec 0031 §Security / D-S26: forbid <script>, <foreignObject>, and
# external xlink:href / href in any committed .svg under site/.
set -eu
cd "$(dirname "$0")/.."

PATTERN='<script|<foreignObject|xlink:href[[:space:]]*=[[:space:]]*"http|href[[:space:]]*=[[:space:]]*"http'

# shellcheck disable=SC2155
matches=$(grep -rEn "$PATTERN" . \
  --include='*.svg' \
  --exclude-dir='node_modules' \
  --exclude-dir='.vitepress/dist' \
  --exclude-dir='.vitepress/cache' \
  2>/dev/null || true)

if [ -n "$matches" ]; then
  echo "FAIL forbidden SVG construct(s) found:"
  echo "$matches"
  exit 1
fi
echo "OK   all committed SVGs are safe"
