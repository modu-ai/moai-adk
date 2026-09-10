#!/bin/bash
# State-3 verification: run the VERBATIM AC-08 and AC-10 bodies from acceptance.md
# against a filled slot table. Expects the guards to pass and the real checks to run.
cd "$1" || exit 1
P=.moai/specs/SPEC-ARTIFACT-STATELESS-001/progress.md
read_sha() {
  local v
  v=$(sed -n "s/^| $1 | \`\([0-9a-f]\{7,\}\)\` |.*/\1/p" "$P")
  [ -n "$v" ] || { echo "FAIL — 「$1」 슬롯이 비어 있거나 7자리 미만이다" >&2; return 1; }
  git rev-parse --verify "$v^{commit}" >/dev/null 2>&1 \
    || { echo "FAIL — 「$1」 = $v 가 이 트리에서 커밋으로 해석되지 않는다" >&2; return 1; }
  echo "$v"
}

echo "=== AC-08 body ==="
BASE_M3=$(read_sha 'M3 착수 직전') || { echo "AC-08 FAIL — 기준 SHA 없음"; exit 1; }
echo "BASE_M3=$BASE_M3"
CHANGED=$(git diff --name-only "$BASE_M3"..HEAD -- '.moai/specs/*/plan.md' '.moai/specs/*/acceptance.md' '.moai/specs/*/design.md' '.moai/specs/*/research.md' | grep -v 'SPEC-ARTIFACT-STATELESS-001' | wc -l | tr -d ' ')
[ "$CHANGED" -gt 0 ] || { echo "AC-08 FAIL — $BASE_M3..HEAD 사이에 .moai/specs 변경이 0건 (판정할 대상이 없다)"; exit 1; }
echo "changed files under .moai/specs = $CHANGED"
git diff --name-only "$BASE_M3"..HEAD -- .moai/specs \
  | grep -E '/(spec|progress)\.md$' \
  | grep -v 'SPEC-ARTIFACT-STATELESS-001' \
  && echo "AC-08 FAIL" || echo "no spec.md/progress.md touched: AC-08 PASS"

echo
echo "=== AC-10 body ==="
BASE_SPEC=$(read_sha 'SPEC 착수 직전') || { echo "AC-10 FAIL — 기준 SHA 없음"; exit 1; }
echo "BASE_SPEC=$BASE_SPEC"
CHANGED=$(git diff --name-only "$BASE_SPEC"..HEAD -- '.moai/specs/*/plan.md' '.moai/specs/*/acceptance.md' '.moai/specs/*/design.md' '.moai/specs/*/research.md' | grep -v 'SPEC-ARTIFACT-STATELESS-001' | wc -l | tr -d ' ')
[ "$CHANGED" -gt 0 ] || { echo "AC-10 FAIL — $BASE_SPEC..HEAD 사이에 .moai/specs 변경이 0건 (판정할 대상이 없다)"; exit 1; }
echo "changed files under .moai/specs = $CHANGED"
TIER=$(git diff "$BASE_SPEC"..HEAD -- .moai/specs | grep -E '^[+-]tier:' \
  | grep -v 'SPEC-ARTIFACT-STATELESS-001' | wc -l | tr -d ' ')
ADDS=$(git diff "$BASE_SPEC"..HEAD -- '.moai/specs/*/plan.md' '.moai/specs/*/acceptance.md' \
  '.moai/specs/*/design.md' '.moai/specs/*/research.md' | grep -cE '^\+status:' | tr -d ' ')
echo "tier edits: $TIER"
echo "status additions in non-spec artifacts: $ADDS"
[ "$TIER" -eq 0 ] && [ "$ADDS" -eq 0 ] && echo "AC-10 PASS" || echo "AC-10 FAIL"
