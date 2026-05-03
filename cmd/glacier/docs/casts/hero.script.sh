#!/usr/bin/env bash
# hero.script.sh
#
# Cast script for the Glacier SDK hero animation.
# Drives: glacier vibe --seed=0 --duration=10s
#
# Rendered to site/public/casts/vibe.cast and site/public/casts/vibe.svg
# at site-build time via agg. Do not commit .cast or .svg here.
#
# Usage:
#   asciinema rec --command="bash hero.script.sh" site/public/casts/vibe.cast
#
# Requirements:
#   - glacier binary on PATH (go install github.com/nathanbrophy/glacier/cmd/glacier@latest)
#   - Terminal width >= 100 columns for the full wordmark to render correctly

set -euo pipefail

# Give the recorder a moment to settle before starting.
sleep 0.5

# Run the vibes animation for exactly 10 seconds with deterministic tip order.
# --seed=0 ensures the tip rotation is reproducible across re-renders.
# --duration=10s exits cleanly after 10 seconds; no key press needed.
glacier vibe --seed=0 --duration=10s
