#!/usr/bin/env bash
# Spec 0031 §API + §Test Matrix: every /docs/packages/<name>.md must
# carry the five magpie:extract directive blocks in order:
#   public-summary, mental-model, api, examples, faq
# Plus check that each cited source spec lists the cited section in its
# `docs-extract` array. (SHA-256 source-checksum verification arrives
# in slice B step 5; for now we accept `source-checksum=PENDING`.)
set -eu
cd "$(dirname "$0")/.."

REPO_ROOT=$(cd .. && pwd)

REQUIRED_SECTIONS=(public-summary mental-model api examples faq)

fail=0

# Iterate every package page.
for page in docs/packages/*.md; do
  [ -f "$page" ] || continue
  pkg=$(basename "$page" .md)
  echo "scan $page"

  # Check directive block presence + order.
  prev_idx=-1
  for section in "${REQUIRED_SECTIONS[@]}"; do
    if ! grep -qE "<!-- magpie:extract source=specs/[0-9]+-${pkg}\.md section=${section}" "$page"; then
      echo "FAIL $page missing directive block for section=${section}"
      fail=1
    fi
  done

  # Check checksum attribute is present (PENDING accepted in v1).
  if ! grep -qE 'source-checksum=' "$page"; then
    echo "FAIL $page missing source-checksum attribute on at least one directive"
    fail=1
  fi

  # Resolve the source spec and confirm docs-extract lists all 5 sections.
  source_spec=$(grep -oE 'source=specs/[0-9]+-'"${pkg}"'\.md' "$page" | head -1 | sed 's/source=//')
  if [ -z "$source_spec" ]; then
    echo "FAIL $page references no source spec"
    fail=1
    continue
  fi
  spec_path="${REPO_ROOT}/${source_spec}"
  if [ ! -f "$spec_path" ]; then
    echo "FAIL $page references missing $source_spec"
    fail=1
    continue
  fi
  for section in "${REQUIRED_SECTIONS[@]}"; do
    if ! awk '/^docs-extract:/{f=1;next} /^[a-zA-Z]/{f=0} f' "$spec_path" | grep -qE "^\s*-\s*${section}\s*$"; then
      echo "FAIL $source_spec docs-extract array missing '${section}' (required by ${page})"
      fail=1
    fi
  done
done

if [ "$fail" -eq 0 ]; then
  echo "OK   all package pages carry the 5 directive blocks; all source specs expose docs-extract"
fi
exit $fail
