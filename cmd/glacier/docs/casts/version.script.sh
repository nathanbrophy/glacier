#!/usr/bin/env bash
# version.script.sh
#
# Cast script for the glacier version command page.
# Shows: default output, then --check with a newer release available.
#
# Rendered to site/public/casts/version.cast at site-build time via agg.
#
# Usage:
#   asciinema rec --command="bash version.script.sh" site/public/casts/version.cast
#
# Requirements:
#   - glacier binary on PATH
#   - Network access for --check (or a pre-seeded cache at
#     <UserCacheDir>/glacier/versioncheck.json)

set -euo pipefail

sleep 0.3

# Show plain version output.
echo '$ glacier version'
glacier version

sleep 1.0

# Show --check output. The cast is recorded against a live binary;
# if the version check is stale-cached, the (stale) annotation will appear.
echo '$ glacier version --check'
glacier version --check

sleep 0.5
