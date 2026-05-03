#!/usr/bin/env bash
# generate.script.sh
#
# Cast script for the glacier generate command page.
# Shows: a full generate run, then --check with one stale file.
#
# Rendered to site/public/casts/generate.cast at site-build time via agg.
# Run from the repo root so ./... resolves against the Glacier module itself.
#
# Usage:
#   asciinema rec --command="bash cmd/glacier/docs/casts/generate.script.sh" \
#     site/public/casts/generate.cast
#
# Requirements:
#   - glacier binary on PATH
#   - Run from the repo root (the Glacier module directory)

set -euo pipefail

sleep 0.3

# Run all generators over the repo.
echo '$ glacier generate ./...'
glacier generate ./...

sleep 1.5

# Run drift check. The cast environment has one intentionally-stale file
# to illustrate the --check output with a diff. In a clean repo this
# will show "in sync" for all generators; that is also a valid cast.
echo '$ glacier generate --check ./...'
glacier generate --check ./... || true

sleep 0.5
