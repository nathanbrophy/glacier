#!/usr/bin/env bash
# Spec 0031 §Test Matrix: every /docs/packages/<name>.md must link to
# the matching specs/<NNNN-name>.md source spec via a "View source spec"
# link. Verifies presence + correct spec ID.
set -eu
cd "$(dirname "$0")/.."

# Map package name -> expected spec ID.
declare -A EXPECTED=(
  [option]=0003 [errs]=0004 [log]=0005 [assert]=0006
  [concur]=0007 [fluent]=0008 [conf]=0009 [fixture]=0010
  [cli]=0011 [mock]=0012 [httpmock]=0013 [httpc]=0015
  [term]=0016 [obs]=0017 [cache]=0033
)

fail=0
for pkg in "${!EXPECTED[@]}"; do
  spec_id="${EXPECTED[$pkg]}"
  page="docs/packages/${pkg}.md"

  if [ ! -f "$page" ]; then
    echo "FAIL $page not present"
    fail=1
    continue
  fi

  expected="specs/${spec_id}-${pkg}.md"
  if grep -q "$expected" "$page"; then
    echo "OK   $page links to $expected"
  else
    echo "FAIL $page missing link to $expected"
    fail=1
  fi
done

exit $fail
