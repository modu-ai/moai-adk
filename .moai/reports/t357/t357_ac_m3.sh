#!/usr/bin/env bash
# t357_ac_m3.sh — AC-AST-001-05 / -06 / -07 / -08 / -09 / -10 judgment runner.
#
# The command blocks in acceptance.md are reproduced here VERBATIM in predicate
# and in verdict text. The reason this file exists at all is mechanical: the
# worktree-session guard refuses a compound shell payload it cannot statically
# track (a function, a heredoc, a `test A && test B` chain), so the blocks
# cannot be pasted into the session's shell as written. They run identically
# from a file.
#
# It is a JUDGMENT runner, not a summary: every criterion prints the numbers it
# compared and then compares them itself, so no verdict rests on the operator
# having read the numbers correctly. Sibling of t357_ac_precheck.sh (M1) and
# t357_ac_state3.sh.
#
# Usage: bash .moai/reports/t357/t357_ac_m3.sh [project-root]

set -u
cd "${1:-.}" || exit 2

P=.moai/specs/SPEC-ARTIFACT-STATELESS-001/progress.md
R=.moai/reports/t357

echo "HEAD=$(git rev-parse --short HEAD)"

# (C) base-value extraction — the 3-guard form. A missing slot dies loudly here
#     rather than letting ""..HEAD collapse to HEAD..HEAD and PASS vacuously.
read_sha() {
  local v
  v=$(sed -n "s/^| $1 | \`\([0-9a-f]\{7,\}\)\` |.*/\1/p" "$P")
  [ -n "$v" ] || { echo "FAIL — 「$1」 슬롯이 비어 있거나 7자리 미만이다" >&2; return 1; }
  git rev-parse --verify "$v^{commit}" >/dev/null 2>&1 \
    || { echo "FAIL — 「$1」 = $v 가 이 트리에서 커밋으로 해석되지 않는다" >&2; return 1; }
  echo "$v"
}
read_num() {
  local v
  v=$(sed -n "s/^| $1 | \`\([0-9]\{1,\}\)\` |.*/\1/p" "$P")
  [ -n "$v" ] || { echo "FAIL — 「$1」 슬롯이 비어 있거나 숫자가 아니다" >&2; return 1; }
  echo "$v"
}

# (B) D1 residual count — whole corpus, frontmatter block only.
count_d1() {
  local n=0 f
  for d in .moai/specs/SPEC-*/; do
    for a in plan acceptance design research; do
      f="${d}${a}.md"; [ -f "$f" ] || continue
      awk 'NR==1&&/^---/{p=1;next} p&&/^---/{exit} p' "$f" \
        | grep -qE '^status:[[:space:]]' && n=$((n+1))
    done
  done
  echo "$n"
}

BASE_M3=$(read_sha 'M3 착수 직전') || { echo "AC-07/-08/-10 FAIL — 기준 SHA 없음"; exit 1; }
N=$(read_num 'M3 착수 시 D1 baseline N') || { echo "AC-07 FAIL — baseline N 없음"; exit 1; }
echo "BASE_M3=$BASE_M3  baseline_N=$N"

echo "--- AC-AST-001-06 (D1 residual across the whole corpus) ---"
D=$(count_d1)
echo "D1 remaining (whole corpus) = $D"
[ "$D" -eq 0 ] && echo "AC-06 PASS" || echo "AC-06 FAIL — residual $D"

echo "--- AC-AST-001-07 (D1 not D2: only status lines removed, none missed) ---"
git diff "$BASE_M3"..HEAD -- '.moai/specs/*/plan.md' '.moai/specs/*/acceptance.md' \
  '.moai/specs/*/design.md' '.moai/specs/*/research.md' \
  | grep '^-' | grep -v '^---' > "$R/t357_removed.txt"
TOTAL=$(wc -l < "$R/t357_removed.txt" | tr -d ' ')
NONSTATUS=$(grep -cvE '^-status:' "$R/t357_removed.txt" | tr -d ' ')
echo "removed lines total:      $TOTAL   (baseline N = $N)"
echo "removed non-status lines: $NONSTATUS"
grep -vE '^-status:' "$R/t357_removed.txt" | head -20
[ "$NONSTATUS" -eq 0 ] && [ "$TOTAL" -eq "$N" ] \
  && echo "AC-07 PASS" \
  || echo "AC-07 FAIL — non-status=$NONSTATUS (must be 0), total=$TOTAL vs N=$N (must be equal)"

echo "--- AC-AST-001-08 (spec.md / progress.md untouched) ---"
CLEANED=$(git diff --name-only "$BASE_M3"..HEAD -- '.moai/specs/*/plan.md' \
  '.moai/specs/*/acceptance.md' '.moai/specs/*/design.md' '.moai/specs/*/research.md' \
  | grep -v 'SPEC-ARTIFACT-STATELESS-001' | wc -l | tr -d ' ')
[ "$CLEANED" -gt 0 ] || { echo "AC-08 FAIL — $BASE_M3..HEAD 에 정리 대상 편집이 0건 (M3가 아직 착지하지 않았다)"; exit 1; }
echo "cleanup-target files in range = $CLEANED"
git diff --name-only "$BASE_M3"..HEAD -- .moai/specs \
  | grep -E '/(spec|progress)\.md$' \
  | grep -v 'SPEC-ARTIFACT-STATELESS-001' \
  && echo "AC-08 FAIL" || echo "no spec.md/progress.md touched: AC-08 PASS"

echo "--- AC-AST-001-09 (rule registered AND residual 0, together) ---"
RULE=$(grep -c 'ArtifactStatusFieldForbidden' internal/spec/lint.go)
echo "rule_matches=$RULE d1_remaining=$D"
{ [ "$RULE" -ge 1 ] && [ "$D" -eq 0 ]; } \
  && echo "AC-09 PASS" \
  || echo "AC-09 FAIL — 둘이 갈라졌다면 era 예외가 필수가 된다"

echo "--- AC-AST-001-10 (out-of-scope axes untouched) ---"
TIER=$(git diff "$BASE_M3"..HEAD | grep -cE '^[+-]tier:' | tr -d ' ')
ADDED=$(git diff "$BASE_M3"..HEAD -- '.moai/specs/*/plan.md' '.moai/specs/*/acceptance.md' \
  '.moai/specs/*/design.md' '.moai/specs/*/research.md' | grep -cE '^\+status:' | tr -d ' ')
echo "tier edits=$TIER  status additions=$ADDED"
{ [ "$TIER" -eq 0 ] && [ "$ADDED" -eq 0 ]; } \
  && echo "AC-10 PASS" \
  || echo "AC-10 FAIL — tier=$TIER status-added=$ADDED (both must be 0)"
