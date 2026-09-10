#!/bin/bash
# t357: pre-implementation evidence for the rewritten ACs.
# Runs the EXACT commands the ACs carry, against the unmodified tree.
#
# v2 (iter-2 N1/N2/R1): extractor stops at level<=3 headings; AC-07/-08/-10 now
# read their base values through the 3-guard read_sha/read_num, so an empty slot
# fails LOUDLY instead of printing PASS.
#
# Expected on an unmodified tree with empty slots: every line FAIL / rc=1.
cd "$1" || exit 1
F=.claude/rules/moai/development/spec-frontmatter-schema.md
P=.moai/specs/SPEC-ARTIFACT-STATELESS-001/progress.md
SEC=$(awk '/^### Artifact Statelessness/{f=1;next} f&&/^#{1,3} /{exit} f' "$F")

echo "HEAD=$(git rev-parse --short HEAD)"
echo "section bytes = ${#SEC}"

neg_hit() {
  printf '%s\n' "$SEC" | grep -F "$1" \
    | grep -qiE '\b(not|never|no longer|rejected|incorrect|false)\b|아니|않'
}

echo "--- AC-AST-001-01 (three separate statements + negation proximity) ---"
if printf '%s\n' "$SEC" | grep -qF 'binds `spec.md` only'; then
  neg_hit 'binds `spec.md` only' && echo "S1 FAIL — negated" || echo "S1 PASS"
else echo "S1 FAIL"; fi
s2=1
for a in plan.md acceptance.md design.md research.md; do
  printf '%s\n' "$SEC" | grep -qF "$a" || s2=0
done
printf '%s\n' "$SEC" | grep -qE '`status:`' || s2=0
[ "$s2" -eq 1 ] && echo "S2 PASS" || echo "S2 FAIL"
if printf '%s\n' "$SEC" | grep -qF 'Tier-independent'; then
  neg_hit 'Tier-independent' && echo "S3 FAIL — negated" || echo "S3 PASS"
else echo "S3 FAIL"; fi

echo "--- AC-AST-001-02 (permission present + not negated, no blanket prohibition) ---"
PP=0; printf '%s\n' "$SEC" | grep -qF 'Frontmatter itself is permitted' && PP=1
printf '%s\n' "$SEC" | grep -F 'Frontmatter itself is permitted' \
  | grep -qiE '\b(not|never|no longer|rejected|incorrect|false)\b|아니|않' && PP=0
NN=$(printf '%s\n' "$SEC" | grep -cE 'MUST NOT (carry|have|contain) (a |any )?(YAML )?frontmatter')
echo "permission=$PP blanket_prohibition=$NN"
[ "$PP" -eq 1 ] && [ "$NN" -eq 0 ] && echo "AC-02 PASS" || echo "AC-02 FAIL"

echo "--- AC-AST-001-11 (template mirror) ---"
diff -q "$F" "internal/template/templates/$F" >/dev/null 2>&1 \
  && echo "mirror identical: PASS" || echo "mirror identical: FAIL"
grep -qF '### Artifact Statelessness' "$F" && echo "anchor local:  PASS" || echo "anchor local:  FAIL"
grep -qF '### Artifact Statelessness' "internal/template/templates/$F" \
  && echo "anchor mirror: PASS" || echo "anchor mirror: FAIL"

echo "--- 3-guard base-value extraction (AC-07 / -08 / -10) ---"
read_sha() {
  local v
  v=$(sed -n "s/^| $1 | \`\([0-9a-f]\{7,\}\)\` |.*/\1/p" "$P")
  [ -n "$v" ] || { echo "FAIL — 「$1」 slot empty or shorter than 7" >&2; return 1; }
  git rev-parse --verify "$v^{commit}" >/dev/null 2>&1 \
    || { echo "FAIL — 「$1」 = $v does not resolve to a commit in this tree" >&2; return 1; }
  echo "$v"
}
read_num() {
  local v
  v=$(sed -n "s/^| $1 | \`\([0-9]\{1,\}\)\` |.*/\1/p" "$P")
  [ -n "$v" ] || { echo "FAIL — 「$1」 slot empty or not a number" >&2; return 1; }
  echo "$v"
}

B1=$(read_sha 'SPEC 착수 직전') && echo "BASE_SPEC=$B1 (rc=0)" || echo "AC-10 FAIL — no base SHA (rc=$?)"
B2=$(read_sha 'M3 착수 직전')   && echo "BASE_M3=$B2 (rc=0)"   || echo "AC-07/-08 FAIL — no base SHA (rc=$?)"
B3=$(read_num 'M3 착수 시 D1 baseline N') && echo "baseline_N=$B3 (rc=0)" || echo "AC-07 FAIL — no baseline N (rc=$?)"
