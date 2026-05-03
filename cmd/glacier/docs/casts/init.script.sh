#!/usr/bin/env bash
# init.script.sh
#
# Cast script for the glacier init command page.
# Shows the --yes (non-interactive) form scaffolding a new project.
#
# Rendered to site/public/casts/init.cast at site-build time via agg.
# Run from a temporary directory so the scaffold does not pollute the repo.
#
# Usage:
#   asciinema rec --command="bash /path/to/init.script.sh" \
#     site/public/casts/init.cast
#
# Requirements:
#   - glacier binary on PATH

set -euo pipefail

# Work in a temp directory so the scaffold is self-contained.
TMPDIR_CAST="$(mktemp -d)"
trap 'rm -rf "$TMPDIR_CAST"' EXIT

cd "$TMPDIR_CAST"

sleep 0.3

# Scaffold a new project using all defaults (non-interactive --yes form).
echo '$ glacier init my-app --yes'
glacier init my-app --yes

sleep 0.5

# Show the generated layout so viewers can see what shipped.
echo ''
echo '$ find my-app -type f | sort'
find my-app -type f | sort

sleep 0.5
