#!/usr/bin/env bash
# test.script.sh
#
# Cast script for the glacier test command page.
# Shows: a focused test run with one failing test, the live panel,
# the summary block, and the isolation hint.
#
# Rendered to site/public/casts/test.cast at site-build time via agg.
# Run from the repo root.
#
# Usage:
#   asciinema rec --command="bash cmd/glacier/docs/casts/test.script.sh" \
#     site/public/casts/test.cast
#
# Requirements:
#   - glacier binary on PATH
#   - Run from the repo root (the Glacier module directory)

set -euo pipefail

sleep 0.3

# Run the conf package tests. In the cast environment TestLayerConflict_File_vs_Env
# is intentionally broken to demonstrate the failure block and isolation hint.
# In a clean repo this will show a clean pass; that is also a valid cast.
echo '$ glacier test ./conf/ ./cli/...'
glacier test ./conf/ ./cli/... || true

sleep 0.5
