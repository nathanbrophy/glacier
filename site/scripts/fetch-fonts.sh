#!/usr/bin/env bash
# fetch-fonts.sh - download the three vendored font families to
# site/public/fonts/. Run once to populate the public asset dir.
#
# Sources are upstream GitHub releases. Each family ships SIL OFL 1.1;
# the script writes OFL.txt alongside the WOFF2s so the license travels
# with the binaries (per spec 0031 §Architecture / Visual system /
# Typography).
#
# Re-running is idempotent (overwrites existing files).
#
# Usage:
#   bash site/scripts/fetch-fonts.sh
set -eu

cd "$(dirname "$0")/.."

OUT="public/fonts"
mkdir -p "$OUT"

# Pinned upstream tags:
#   Inter           v4.0   (rsms/inter)
#   Space Grotesk   master (floriankarsten/space-grotesk - no semver tags)
#   JetBrains Mono  v2.304 (JetBrains/JetBrainsMono)
INTER_TAG="v4.0"
SPACE_TAG="master"
JBM_TAG="v2.304"

INTER_BASE="https://raw.githubusercontent.com/rsms/inter/${INTER_TAG}/docs/font-files"
SPACE_BASE="https://raw.githubusercontent.com/floriankarsten/space-grotesk/${SPACE_TAG}/fonts/woff2/static"
JBM_BASE="https://raw.githubusercontent.com/JetBrains/JetBrainsMono/${JBM_TAG}/fonts/webfonts"

fetch() {
  local url="$1"
  local out="$2"
  echo "fetch $out"
  curl -fsSL --retry 3 -o "$out" "$url"
}

# Inter (Regular 400, Medium 500, SemiBold 600)
fetch "${INTER_BASE}/Inter-Regular.woff2"  "${OUT}/Inter-400.woff2"
fetch "${INTER_BASE}/Inter-Medium.woff2"   "${OUT}/Inter-500.woff2"
fetch "${INTER_BASE}/Inter-SemiBold.woff2" "${OUT}/Inter-600.woff2"
fetch "https://raw.githubusercontent.com/rsms/inter/${INTER_TAG}/LICENSE.txt" \
      "${OUT}/Inter-OFL.txt"

# Space Grotesk (Medium 500, Bold 700)
fetch "${SPACE_BASE}/SpaceGrotesk-Medium.woff2" "${OUT}/SpaceGrotesk-500.woff2"
fetch "${SPACE_BASE}/SpaceGrotesk-Bold.woff2"   "${OUT}/SpaceGrotesk-700.woff2"
fetch "https://raw.githubusercontent.com/floriankarsten/space-grotesk/${SPACE_TAG}/OFL.txt" \
      "${OUT}/SpaceGrotesk-OFL.txt"

# JetBrains Mono (Regular 400, Bold 700)
fetch "${JBM_BASE}/JetBrainsMono-Regular.woff2" "${OUT}/JetBrainsMono-400.woff2"
fetch "${JBM_BASE}/JetBrainsMono-Bold.woff2"    "${OUT}/JetBrainsMono-700.woff2"
fetch "https://raw.githubusercontent.com/JetBrains/JetBrainsMono/${JBM_TAG}/OFL.txt" \
      "${OUT}/JetBrainsMono-OFL.txt"

echo
echo "OK   vendored fonts in ${OUT}:"
ls -lh "${OUT}/" | tail -n +2
