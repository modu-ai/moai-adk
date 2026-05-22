# SPEC-V3R6-SKILL-COMPRESS-001 Acceptance Criteria

## Acceptance Criteria (Binary PASS/FAIL)

### AC-SCM-001 — testing skill ≤ 2,000w
**REQ**: REQ-SCM-001
**Verification command**:
```bash
wc -w .claude/skills/moai-workflow-testing/SKILL.md
```
**PASS condition**: Output value (number) ≤ 2000.
**FAIL condition**: > 2000.
**Baseline (pre-compression)**: 3153.

### AC-SCM-002 — spec skill ≤ 1,700w
**REQ**: REQ-SCM-002
**Verification command**:
```bash
wc -w .claude/skills/moai-workflow-spec/SKILL.md
```
**PASS condition**: Output value ≤ 1700.
**Baseline**: 2394.

### AC-SCM-003 — project skill ≤ 1,400w
**REQ**: REQ-SCM-003
**Verification command**:
```bash
wc -w .claude/skills/moai-workflow-project/SKILL.md
```
**PASS condition**: Output value ≤ 1400.
**Baseline**: 2068.

### AC-SCM-004 — design-handoff skill ≤ 1,600w
**REQ**: REQ-SCM-004
**Verification command**:
```bash
wc -w .claude/skills/moai-domain-design-handoff/SKILL.md
```
**PASS condition**: Output value ≤ 1600.
**Baseline**: 2039.

### AC-SCM-005 — meta-harness skill ≤ 1,600w
**REQ**: REQ-SCM-005
**Verification command**:
```bash
wc -w .claude/skills/moai-meta-harness/SKILL.md
```
**PASS condition**: Output value ≤ 1600.
**Baseline**: 2010.

### AC-SCM-006 — Aggregate ≤ 8,200w
**REQ**: REQ-SCM-006
**Verification command**:
```bash
total=0
for s in moai-workflow-testing moai-workflow-spec moai-workflow-project moai-domain-design-handoff moai-meta-harness; do
  w=$(wc -w < .claude/skills/$s/SKILL.md)
  total=$((total + w))
done
echo "TOTAL: $total"
```
**PASS condition**: TOTAL ≤ 8200 (target 7300 + 900 aggregate 여유).
**Baseline TOTAL**: 11664.
**Expected**: TOTAL is near 7300, MUST be ≤ 8200.

### AC-SCM-007 — Trigger keyword preservation (5 skills × N keywords)
**REQ**: REQ-SCM-009
**Verification command**:
```bash
for s in moai-workflow-testing moai-workflow-spec moai-workflow-project moai-domain-design-handoff moai-meta-harness; do
  # Extract keywords from frontmatter
  kws=$(awk '/^  keywords:/,/^  [a-z]+:/' .claude/skills/$s/SKILL.md | grep -oE '"[^"]+"' | tr -d '"')
  for kw in $kws; do
    # Check keyword appears in body (after second ---)
    body_lines=$(awk '/^---$/{c++; next} c>=2' .claude/skills/$s/SKILL.md)
    if ! echo "$body_lines" | grep -iq "$kw"; then
      echo "MISSING in $s: $kw"
      exit 1
    fi
  done
done
echo "All keywords preserved"
```
**PASS condition**: Final output `All keywords preserved`. Zero `MISSING in ...` lines.
**Notes**: Either exact keyword match OR documented synonym map (in HISTORY) is allowed per REQ-SCM-009. If synonym substitution is intentional, body MUST contain the synonym + commentary establishing equivalence.

### AC-SCM-008 — Template mirror byte-identical (5 SKILL.md + new Level 3 files)
**REQ**: REQ-SCM-010
**Verification command**:
```bash
fail=0
for s in moai-workflow-testing moai-workflow-spec moai-workflow-project moai-domain-design-handoff moai-meta-harness; do
  diff -q .claude/skills/$s/SKILL.md internal/template/templates/.claude/skills/$s/SKILL.md || fail=1
  # Verify new Level 3 references too
  if [ -d .claude/skills/$s/references/ ]; then
    diff -rq .claude/skills/$s/references/ internal/template/templates/.claude/skills/$s/references/ 2>&1 | grep -v "^$" && fail=1
  fi
done
[ $fail -eq 0 ] && echo "All mirrors identical" || echo "MIRROR DRIFT detected"
```
**PASS condition**: Output `All mirrors identical`. Zero `MIRROR DRIFT` lines.

### AC-SCM-009 — Cross-reference integrity (Level 3 link targets exist)
**REQ**: REQ-SCM-011
**Verification command**:
```bash
for s in moai-workflow-testing moai-workflow-spec moai-workflow-project moai-domain-design-handoff moai-meta-harness; do
  # Extract markdown links pointing to references/ in body
  awk '/^---$/{c++; next} c>=2' .claude/skills/$s/SKILL.md | \
    grep -oE 'references/[a-zA-Z0-9_/-]+\.md' | sort -u | while read link; do
    target=".claude/skills/$s/$link"
    if [ ! -f "$target" ]; then
      echo "DANGLING in $s: $link"
    fi
  done
done | tee /tmp/dangling-check.txt
[ ! -s /tmp/dangling-check.txt ] && echo "All references resolve" || (cat /tmp/dangling-check.txt; exit 1)
```
**PASS condition**: Output `All references resolve`. Zero `DANGLING in ...` lines.

### AC-SCM-010 — Catalog hash regeneration consistency
**REQ**: REQ-SCM-012
**Verification command**:
```bash
# Pre-condition: gen-catalog-hashes.go must run successfully
go run gen-catalog-hashes.go --all
# Run catalog hash format test
go test ./internal/template/... -run TestManifestHashFormat -v
# Run catalog completeness test
go test ./internal/template/... -run TestAllSkillsInCatalog -v
```
**PASS condition**: Both tests exit 0. `gen-catalog-hashes.go --all` exits 0.

### AC-SCM-011 — Frontmatter byte-identity (Level 1 preservation)
**REQ**: REQ-SCM-008
**Verification command**:
```bash
# Compare current frontmatter against pre-compression baseline
# Baseline is captured in M1 Pre-flight Step 6 to /tmp/trigger-<skill>-before.txt
for s in moai-workflow-testing moai-workflow-spec moai-workflow-project moai-domain-design-handoff moai-meta-harness; do
  awk '/^triggers:/,/^---$/' .claude/skills/$s/SKILL.md > /tmp/trigger-$s-after.txt
  if ! diff -q /tmp/trigger-$s-before.txt /tmp/trigger-$s-after.txt > /dev/null; then
    echo "FRONTMATTER TRIGGER DRIFT in $s"
    diff /tmp/trigger-$s-before.txt /tmp/trigger-$s-after.txt
    exit 1
  fi
done
echo "All triggers byte-identical"
```
**PASS condition**: Output `All triggers byte-identical`. Zero diff for `triggers:` block.
**Notes**: Full frontmatter (name / description / allowed-tools / metadata) byte-identity strongly recommended; AC-SCM-011 minimum bar is `triggers:` block only since that is the runtime-critical metadata.

### AC-SCM-012 — Cross-platform build + template tests PASS
**REQ**: Implicit (CLAUDE.local.md §6 cross-platform discipline)
**Verification command**:
```bash
go build ./... && \
  GOOS=windows GOARCH=amd64 go build ./... && \
  go test ./internal/template/...
```
**PASS condition**: All three commands exit 0.
**Notes**: Pre-existing baseline failures (per CATALOG-SSOT-001 lessons) are allowed if NEW=0 from this SPEC. Run-phase MUST distinguish NEW vs baseline.

## REQ ↔ AC Traceability Matrix

| REQ | AC | Verification Mode |
|---|---|---|
| REQ-SCM-001 (testing ≤ 2,000w) | AC-SCM-001 | `wc -w` numeric |
| REQ-SCM-002 (spec ≤ 1,700w) | AC-SCM-002 | `wc -w` numeric |
| REQ-SCM-003 (project ≤ 1,400w) | AC-SCM-003 | `wc -w` numeric |
| REQ-SCM-004 (design-handoff ≤ 1,600w) | AC-SCM-004 | `wc -w` numeric |
| REQ-SCM-005 (meta-harness ≤ 1,600w) | AC-SCM-005 | `wc -w` numeric |
| REQ-SCM-006 (aggregate ≤ 8,200w) | AC-SCM-006 | shell loop sum |
| REQ-SCM-007 (removed → Level 3 or justified) | (implicit in M2-M6 plan + AC-SCM-009) | Triage matrix in progress.md |
| REQ-SCM-008 (frontmatter triggers unchanged) | AC-SCM-011 | `diff` |
| REQ-SCM-009 (keyword/synonym preserved in body) | AC-SCM-007 | shell grep loop |
| REQ-SCM-010 (template mirror byte-identical) | AC-SCM-008 | `diff -q` / `diff -rq` |
| REQ-SCM-011 (cross-reference integrity) | AC-SCM-009 | shell grep + `test -f` |
| REQ-SCM-012 (catalog hash + tests) | AC-SCM-010 | `go test` |

**Coverage**: 12/12 REQs traced to AC (100%). REQ-SCM-007 is verified indirectly through plan.md Triage matrix (M2-M6 each milestone produces explicit MOVE_TO_LEVEL_3 / REMOVE_with_justification entries in progress.md).

## Edge Cases

### EC-SCM-001 — Trigger keyword가 multi-word phrase일 때 (예: "code review", "user story")
**Scenario**: REQ-SCM-009 보존 의무. multi-word phrase는 단어 단위 grep으로 분리 시 false positive 가능.
**Handling**:
- AC-SCM-007의 grep은 `grep -iq "$kw"` (단일 quoted string으로 phrase 검색)
- 단일 단어로 분리하지 않음
- 매칭 시 case-insensitive
- phrase 자체가 body에 등장하지 않더라도 동등한 다른 표현이 있으면 manual override (HISTORY에 명시) 가능

### EC-SCM-002 — Level 3 references/ 디렉토리가 디자인-handoff / meta-harness에 신규 생성될 때 template mirror 동기화
**Scenario**: 기존 reference/ 디렉토리 부재 → 신규 mkdir 필요. local + template 양쪽 모두 mkdir + file write 의무.
**Handling**:
- M5/M6에서 mkdir -p `.claude/skills/<skill>/references/` + mkdir -p `internal/template/templates/.claude/skills/<skill>/references/` 둘 다 실행
- AC-SCM-008의 `diff -rq` 디렉토리 비교가 자동 검증

### EC-SCM-003 — 압축 후 word count이 target보다 너무 낮을 때 (under-compression)
**Scenario**: REQ-SCM-007 의무 (removed → Level 3 or justified). 만약 testing skill을 1,500w (target 1,800w보다 -300w)까지 압축했다면, 200w-300w 의 essential workflow logic이 제거된 것일 수 있음 — silent regression.
**Handling**:
- Plan §1.0 Triage step에서 KEEP_INLINE 표시된 섹션은 모두 retain
- Word count target은 UPPER bound (≤), 하한선은 강제하지 않음 — 단 -10% 이상 cut 시 reviewer가 justification 요구 (CLAUDE.md §7 Rule 5 post-review)
- Aggregate ≤ 8,200w와 함께 individual target 별 ±200w 여유 부여로 over-compression 완화

## Quality Gates

| Gate | Threshold | Tool |
|---|---|---|
| Word count (per skill) | ≤ target | `wc -w` (AC-SCM-001..005) |
| Aggregate word count | ≤ 8,200w | shell sum (AC-SCM-006) |
| Trigger preservation | 0 missing | grep loop (AC-SCM-007) |
| Template mirror | 0 drift | `diff -rq` (AC-SCM-008) |
| Cross-reference | 0 dangling | `test -f` (AC-SCM-009) |
| Catalog hash | tests PASS | `go test TestManifestHashFormat / TestAllSkillsInCatalog` (AC-SCM-010) |
| Frontmatter trigger byte-identity | 0 diff | `diff` (AC-SCM-011) |
| Cross-platform build | exit 0 | `go build` Linux + Windows (AC-SCM-012) |

## Definition of Done

본 SPEC가 **implemented** 상태로 전환되기 위한 조건:

1. AC-SCM-001 ~ AC-SCM-012 **모두** PASS (12/12).
2. progress.md에 다음 항목 기록:
   - M1 baseline 측정 결과 (5 skill 실제 word count + git log clean 확인)
   - M2-M6 각 milestone의 Triage matrix (KEEP_INLINE / MOVE_TO_LEVEL_3 / REMOVE 분류 + justification)
   - M7 catalog hash diff (변경된 hash entries enumeration)
   - M8 cross-platform build 결과 + 잔존 baseline failures 분류 (NEW vs pre-existing)
3. SPEC frontmatter `status: draft` → `implemented`, `version: 0.1.0` → `0.2.0`.
4. HISTORY에 v0.2.0 row 추가 (implementation summary + AC matrix 결과).
5. 무관 untracked / modified 파일 일체 보존 (PRESERVE 의무).
6. Wave 1 batch sync 또는 단독 PR 결정 (orchestrator AskUserQuestion 별도 진행).

## Coverage 측정 (TRUST 5 craft)

Skill body 압축은 Go test coverage 직접 영향 없음 (테스트 대상 코드 변경 부재). 단:

- `go test ./internal/template/... -cover` 결과가 baseline ±0.5pp 이내 (AC-SCM-010 + AC-SCM-012)
- Skill consumer (manager-develop, manager-quality 등) 동작은 본 SPEC scope 외 — 통합 테스트는 사용자 fixture 의존이므로 본 SPEC에서 verify 불가
- 향후 SPEC `SPEC-V3R6-SKILL-INTEGRATION-VERIFY-001` (가칭) 후보: LLM-based semantic equivalence test
