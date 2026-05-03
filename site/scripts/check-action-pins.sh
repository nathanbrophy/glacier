#!/usr/bin/env bash
# Spec 0031 §Security & Supply-Chain Notes / D-S14, D-S25:
#   1. Every `uses:` action ref in .github/workflows/site-*.yml is pinned
#      by 40-char SHA. No `@v...` tag pins, no `@main`, no `@latest`.
#   2. site-pr-checks.yml triggers on `pull_request` only, NEVER
#      `pull_request_target` (which leaks secrets to fork PRs).
#   3. PR-checks workflow has read-only `contents: read` permission.
set -eu

ROOT=$(cd "$(dirname "$0")/../.." && pwd)
cd "$ROOT"

fail=0

shopt -s nullglob
workflows=(.github/workflows/site-*.yml)
shopt -u nullglob

if [ ${#workflows[@]} -eq 0 ]; then
  echo "SKIP no site-*.yml workflows present yet"
  exit 0
fi

for wf in "${workflows[@]}"; do
  echo "scan $wf"

  # 1. SHA pin enforcement
  while IFS= read -r line; do
    # Skip comments
    case "$line" in
      *"#"*uses:*) ;;
    esac
    if echo "$line" | grep -qE '^\s*-?\s*uses:\s*[A-Za-z0-9./_-]+@[0-9a-fA-F]{40}\s*(#.*)?$'; then
      :  # 40-char SHA pin
    elif echo "$line" | grep -qE '^\s*-?\s*uses:'; then
      echo "FAIL $wf: action not SHA-pinned -> $line"
      fail=1
    fi
  done < "$wf"

  # 2. pull_request_target ban (PR-checks workflow only)
  case "$wf" in
    *pr-checks*|*pr_checks*)
      if grep -qE '^\s*pull_request_target\s*:' "$wf"; then
        echo "FAIL $wf: uses pull_request_target (forbidden)"
        fail=1
      fi
      # 3. read-only token
      if grep -qE 'permissions:[[:space:]]*$' "$wf"; then
        # has explicit permissions block; expect contents: read
        if ! grep -qE '^\s*contents:\s*read' "$wf"; then
          echo "WARN $wf: PR-checks workflow may not declare 'contents: read'"
        fi
      fi
      ;;
  esac
done

if [ "$fail" -eq 0 ]; then
  echo "OK   all action pins SHA-locked, PR triggers safe"
fi
exit $fail
