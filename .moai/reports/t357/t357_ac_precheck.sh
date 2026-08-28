#!/bin/bash
# t357: pre-implementation FAIL evidence for the rewritten AC-AST-001-01 / -02.
# Runs the EXACT commands the ACs carry, against the unmodified tree.
# Expected on an unmodified tree: every line prints FAIL / 0-match.
cd "$1" || exit 1
F=.claude/rules/moai/development/spec-frontmatter-schema.md
SEC=$(awk '/^### Artifact Statelessness/{f=1;next} f&&/^##/{exit} f' "$F")

echo "HEAD=$(git rev-parse --short HEAD)"
echo "section bytes = ${#SEC}"

echo "--- AC-AST-001-01 (three separate statements) ---"
printf '%s\n' "$SEC" | grep -qF 'binds `spec.md` only' \
  && echo "S1 spec.md-only:      PASS" || echo "S1 spec.md-only:      FAIL"

s2=1
for a in plan.md acceptance.md design.md research.md; do
  printf '%s\n' "$SEC" | grep -qF "$a" || s2=0
done
printf '%s\n' "$SEC" | grep -qE '`status:`' || s2=0
[ "$s2" -eq 1 ] && echo "S2 four-artifacts+status: PASS" || echo "S2 four-artifacts+status: FAIL"

printf '%s\n' "$SEC" | grep -qF 'Tier-independent' \
  && echo "S3 Tier-independent:  PASS" || echo "S3 Tier-independent:  FAIL"

echo "--- AC-AST-001-02 (permission present, blanket prohibition absent) ---"
printf '%s\n' "$SEC" | grep -qF 'Frontmatter itself is permitted' \
  && echo "P  permission stated: PASS" || echo "P  permission stated: FAIL"
n=$(printf '%s\n' "$SEC" | grep -cE 'MUST NOT (carry|have|contain) (a |any )?(YAML )?frontmatter')
echo "N  blanket-prohibition matches = $n (must be 0, AND P must be PASS)"

echo "--- AC-AST-001-11 (template mirror byte-identity) ---"
if diff -q "$F" "internal/template/templates/$F" >/dev/null 2>&1; then
  echo "mirror identical: yes"
else
  echo "mirror identical: NO"
fi
printf '%s\n' "$SEC" | grep -qF 'Artifact Statelessness' && echo "anchor in local: PASS" || echo "anchor in local: FAIL (section empty)"
grep -qF '### Artifact Statelessness' "internal/template/templates/$F" \
  && echo "anchor in mirror: PASS" || echo "anchor in mirror: FAIL"
