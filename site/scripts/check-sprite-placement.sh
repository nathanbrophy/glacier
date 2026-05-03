#!/usr/bin/env bash
# Spec 0031 §Migration Amendment A: <MascotSprite> may only appear in:
#   - footer (theme/components/Footer.vue or Layout.vue)
#   - 404 page (404.md or theme/components/NotFound.vue)
#   - sidebar empty state (no current file; allowed when added)
#   - scroll-to-top button (theme/components/ScrollTop.vue or similar)
#   - page-level accent placements approved by the user:
#     sdk/index.md, why.md, docs/index.md
#
# Any occurrence outside these allowed files fails the check.
set -eu
cd "$(dirname "$0")/.."

ALLOWED='^(\.vitepress/theme/(Layout|components/(MascotSprite|Footer|NotFound|ScrollTop|EmptySidebar))\.vue|404\.md|sdk/index\.md|why\.md|docs/index\.md)$'

# shellcheck disable=SC2155
matches=$(grep -rln '<MascotSprite' . \
  --include='*.vue' --include='*.md' \
  --exclude-dir='node_modules' \
  --exclude-dir='.vitepress/dist' \
  --exclude-dir='.vitepress/cache' \
  2>/dev/null | sed 's|^\./||' || true)

fail=0
if [ -z "$matches" ]; then
  echo "OK   no <MascotSprite> usages found yet"
  exit 0
fi

while IFS= read -r f; do
  [ -z "$f" ] && continue
  if echo "$f" | grep -qE "$ALLOWED"; then
    echo "OK   $f"
  else
    echo "FAIL $f - <MascotSprite> outside allowed placements"
    fail=1
  fi
done <<EOF
$matches
EOF

exit $fail
