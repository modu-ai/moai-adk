# Acceptance Criteria: SPEC-V3R6-SKILL-CONSOLIDATE-001

본 acceptance.md 는 `/moai run` 완료 시점에 binary contract 로 검증되는 9개 AC 와 REQ↔AC traceability matrix 를 정의한다. 모든 AC 는 `PASS` 또는 `FAIL` 로 평가되며 (subjective scoring 없음), grep / wc / diff / test exit code 같은 reproducible command 로 verification 가능하다.

## AC Binary Matrix

### AC-SC-001 (BLOCKING) — Source skill removal complete

**REQ**: REQ-SC-001 (Source skill removal)

**Statement**: `/moai run` 완료 시점에 7개 old skill 디렉토리가 local `.claude/skills/` + template `internal/template/templates/.claude/skills/` 양쪽 모두에서 존재하지 않는다.

**Verification command**:

```bash
for s in moai-workflow-ci-watch moai-workflow-ci-autofix moai-workflow-design-import moai-workflow-design-context moai-harness-hook-ci moai-harness-workflow moai-harness-quality; do
  for prefix in ".claude/skills" "internal/template/templates/.claude/skills"; do
    if [ -d "$prefix/$s" ]; then echo "FAIL: $prefix/$s still exists"; fi
  done
done
echo "AC-SC-001 done"
```

**Expected output**: `AC-SC-001 done` (no `FAIL:` lines).

**Pass criterion**: Zero `FAIL:` output lines. All 14 paths absent.

---

### AC-SC-002 (BLOCKING) — Unified skill creation complete

**REQ**: REQ-SC-002 (Unified skill creation), REQ-SC-008 (Skill body minimum contract)

**Statement**: 3개 new skill 디렉토리가 local + template 양쪽에 존재하고, 각 SKILL.md 가 valid YAML frontmatter 와 4개 필수 섹션을 포함한다.

**Verification command**:

```bash
for s in moai-workflow-ci-loop moai-workflow-design moai-harness-patterns; do
  for prefix in ".claude/skills" "internal/template/templates/.claude/skills"; do
    f="$prefix/$s/SKILL.md"
    if [ ! -f "$f" ]; then echo "FAIL: $f missing"; continue; fi
    head -1 "$f" | grep -q "^---$" || echo "FAIL: $f missing YAML frontmatter"
    grep -q "^name:" "$f" || echo "FAIL: $f missing name field"
    grep -q "^description:" "$f" || echo "FAIL: $f missing description field"
    grep -qE "^## (Quick Reference|.*Quick Reference)" "$f" || echo "WARN: $f missing Quick Reference section"
    grep -qE "^## (Implementation Guide|Implementation)" "$f" || echo "WARN: $f missing Implementation Guide section"
    grep -qE "^## Works Well With" "$f" || echo "WARN: $f missing Works Well With section"
    grep -qi "absorbed from\|absorbs:" "$f" || echo "WARN: $f missing absorbed-from footer"
  done
done
echo "AC-SC-002 done"
```

**Expected output**: `AC-SC-002 done` with zero `FAIL:` lines. `WARN:` lines acceptable but should be 0 for ideal compliance.

**Pass criterion**: Zero `FAIL:` lines (6 files exist + frontmatter valid). `WARN:` count ≤ 3 acceptable (REQ-SC-008 is HARD on structure but ## heading exact text may vary).

---

### AC-SC-003 (BLOCKING) — Word count budget respected

**REQ**: REQ-SC-003 (Word-count budget)

**Statement**: 각 통합 SKILL.md 의 `wc -w` 측정값이 1,500 words 이하이고, 3개 합계가 3,800 words 이하이다.

**Verification command**:

```bash
total=0
fail=0
for s in moai-workflow-ci-loop moai-workflow-design moai-harness-patterns; do
  f=".claude/skills/$s/SKILL.md"
  if [ ! -f "$f" ]; then echo "FAIL: $f missing"; fail=$((fail+1)); continue; fi
  w=$(wc -w < "$f" | tr -d ' ')
  echo "  $s: $w words"
  total=$((total+w))
  if [ "$w" -gt 1500 ]; then echo "FAIL: $s exceeds 1500w (got $w)"; fail=$((fail+1)); fi
done
echo "  total: $total words (budget 3800)"
if [ "$total" -gt 3800 ]; then echo "FAIL: total exceeds 3800w (got $total)"; fail=$((fail+1)); fi
[ "$fail" -eq 0 ] && echo "AC-SC-003 PASS" || echo "AC-SC-003 FAIL ($fail failures)"
```

**Expected output**: `AC-SC-003 PASS`. Token saving estimate: baseline 6,914w − total ≥ 3,114w ≈ -9.3K tokens.

**Pass criterion**: Each file ≤ 1,500 words AND total ≤ 3,800 words.

---

### AC-SC-004 (BLOCKING) — Template-First mirror parity

**REQ**: REQ-SC-004 (Template-First mirror parity)

**Statement**: 3 pair 의 SKILL.md (local ↔ template) 가 byte-identical 하다.

**Verification command**:

```bash
fail=0
for s in moai-workflow-ci-loop moai-workflow-design moai-harness-patterns; do
  local_f=".claude/skills/$s/SKILL.md"
  tmpl_f="internal/template/templates/.claude/skills/$s/SKILL.md"
  if diff -q "$local_f" "$tmpl_f" >/dev/null 2>&1; then
    echo "  $s: PARITY"
  else
    echo "FAIL: $s diff detected"
    diff "$local_f" "$tmpl_f" | head -10
    fail=$((fail+1))
  fi
done
[ "$fail" -eq 0 ] && echo "AC-SC-004 PASS" || echo "AC-SC-004 FAIL ($fail diff)"
```

**Expected output**: 3 lines `PARITY` + `AC-SC-004 PASS`.

**Pass criterion**: `diff -q` exit 0 for all 3 pair. CLAUDE.local.md §2 Template-First Rule preserved.

---

### AC-SC-005 (BLOCKING) — Catalog SSOT consistency

**REQ**: REQ-SC-005 (Catalog SSOT consistency)

**Statement**: `internal/template/catalog.yaml` 에서 7 old entry 가 제거되고 3 new entry 가 추가되었으며, manifest hash format 검증이 통과한다.

**Verification command**:

```bash
fail=0
old_count=$(grep -c "moai-workflow-ci-watch\|moai-workflow-ci-autofix\|moai-workflow-design-import\|moai-workflow-design-context\|moai-harness-hook-ci\|moai-harness-workflow\|moai-harness-quality" internal/template/catalog.yaml 2>/dev/null)
new_count=$(grep -c "moai-workflow-ci-loop\|moai-workflow-design\|moai-harness-patterns" internal/template/catalog.yaml 2>/dev/null)
echo "  old refs in catalog: $old_count (expected 0)"
echo "  new refs in catalog: $new_count (expected ≥ 3)"
[ "$old_count" -ne 0 ] && { echo "FAIL: $old_count old refs remain in catalog.yaml"; fail=$((fail+1)); }
[ "$new_count" -lt 3 ] && { echo "FAIL: only $new_count new refs (need ≥ 3)"; fail=$((fail+1)); }
# Manifest hash format
if go test ./internal/template/... -run TestManifestHashFormat -count=1 2>&1 | tail -3 | grep -q "PASS\|ok "; then
  echo "  TestManifestHashFormat: PASS"
else
  echo "FAIL: TestManifestHashFormat failed"
  fail=$((fail+1))
fi
[ "$fail" -eq 0 ] && echo "AC-SC-005 PASS" || echo "AC-SC-005 FAIL ($fail)"
```

**Expected output**: `AC-SC-005 PASS`.

**Pass criterion**: 0 old refs + ≥ 3 new refs in catalog.yaml + TestManifestHashFormat PASS.

---

### AC-SC-006 (BLOCKING) — Cross-reference rename completeness

**REQ**: REQ-SC-006 (Cross-reference rename completeness)

**Statement**: `grep` 으로 7 old skill name 검색 시 0 matches (stale worktrees 제외, 본 SPEC 디렉토리 제외, HISTORY entry 제외, agent-memory 제외).

**Verification command**:

```bash
matches=$(grep -rln "moai-workflow-ci-watch\|moai-workflow-ci-autofix\|moai-workflow-design-import\|moai-workflow-design-context\|moai-harness-hook-ci\|moai-harness-workflow\|moai-harness-quality" \
  .claude/ CLAUDE.md internal/template/templates/.claude/ internal/design/dtcg/ 2>/dev/null \
  | grep -v ".claude/worktrees/" \
  | grep -v ".claude/agent-memory/" \
  | grep -v ".moai/specs/SPEC-V3R6-SKILL-CONSOLIDATE-001/" \
  | wc -l | tr -d ' ')
echo "  cross-ref matches: $matches (expected 0)"
if [ "$matches" -eq 0 ]; then
  echo "AC-SC-006 PASS"
else
  echo "FAIL: $matches files still reference old skill names:"
  grep -rln "moai-workflow-ci-watch\|moai-workflow-ci-autofix\|moai-workflow-design-import\|moai-workflow-design-context\|moai-harness-hook-ci\|moai-harness-workflow\|moai-harness-quality" \
    .claude/ CLAUDE.md internal/template/templates/.claude/ internal/design/dtcg/ 2>/dev/null \
    | grep -v ".claude/worktrees/" \
    | grep -v ".claude/agent-memory/" \
    | grep -v ".moai/specs/SPEC-V3R6-SKILL-CONSOLIDATE-001/"
  echo "AC-SC-006 FAIL"
fi
```

**Expected output**: `AC-SC-006 PASS`.

**Pass criterion**: 0 cross-ref matches in scope. Stale worktrees + agent-memory + 본 SPEC 디렉토리는 명시적 제외 (정보 보존 + 자기 참조 회피).

---

### AC-SC-007 (BLOCKING) — Test suite regression preservation

**REQ**: REQ-SC-007 (Test-suite regression preservation)

**Statement**: 다음 4개 test 가 통합 후 PASS 한다 (catalog.yaml + skill registry 정합성).

**Verification command**:

```bash
fail=0
for test in TestAllSkillsInCatalog TestAllAgentsInCatalog TestManifestHashFormat TestRuleTemplateMirrorDrift; do
  output=$(go test ./internal/template/... -run "^$test\$" -count=1 -timeout=60s 2>&1 | tail -5)
  if echo "$output" | grep -q "^PASS\|^ok "; then
    echo "  $test: PASS"
  else
    echo "FAIL: $test"
    echo "$output" | head -3
    fail=$((fail+1))
  fi
done
[ "$fail" -eq 0 ] && echo "AC-SC-007 PASS" || echo "AC-SC-007 FAIL ($fail)"
```

**Expected output**: `AC-SC-007 PASS`.

**Pass criterion**: All 4 tests PASS. NEW regression 0건; pre-existing baseline failure 별도 식별 (e.g., `TestImplementationSkillsContainPipelineRejectionSentinel` 같은 known baseline 은 본 SPEC scope 외, 별도 SPEC `SPEC-V3R6-CI-BASELINE-CLEANUP-001` 후보).

---

### AC-SC-008 (BLOCKING) — Cross-platform build

**REQ**: 모든 REQ (간접적, build sanity)

**Statement**: Linux + Windows cross-compile 가 통과한다.

**Verification command**:

```bash
fail=0
if go build ./... 2>&1 | tail -3; then echo "  linux build: PASS"; else echo "FAIL: linux build"; fail=$((fail+1)); fi
if GOOS=windows GOARCH=amd64 go build ./... 2>&1 | tail -3; then echo "  windows build: PASS"; else echo "FAIL: windows build"; fail=$((fail+1)); fi
[ "$fail" -eq 0 ] && echo "AC-SC-008 PASS" || echo "AC-SC-008 FAIL ($fail)"
```

**Expected output**: `AC-SC-008 PASS`.

**Pass criterion**: Both builds exit 0. Cross-platform parity preserved.

---

### AC-SC-009 (NON-BLOCKING, advisory) — Skill trigger keyword 비충돌 명시

**REQ**: REQ-SC-009 (Skill trigger keyword 비충돌)

**Statement**: 각 통합 SKILL.md frontmatter `description` 또는 본문에 sibling skill 과의 의미적 충돌 회피 의도가 명시되어 있다 (best-effort guidance, manual review).

**Verification command** (advisory):

```bash
for s in moai-workflow-ci-loop moai-workflow-design moai-harness-patterns; do
  f=".claude/skills/$s/SKILL.md"
  desc=$(grep "^description:" "$f" | head -1)
  echo "  $s description: $desc"
done
echo "AC-SC-009 (advisory) — manual review recommended"
```

**Expected output**: 3 description lines visible. No machine-binary criterion (Best-effort guidance only).

**Pass criterion**: AC-SC-009 PASS by default unless reviewer finds explicit conflict. Failure 발견 시 docs-only follow-up edit.

---

## REQ↔AC Traceability Matrix

| REQ ID | REQ Statement (summary) | AC Coverage | Pass criterion |
|--------|-------------------------|-------------|----------------|
| REQ-SC-001 | Source skill removal (7 dirs × 2 mirrors absent) | AC-SC-001 | Zero `FAIL:` lines |
| REQ-SC-002 | Unified skill creation (3 dirs × 2 mirrors present + valid YAML FM) | AC-SC-002 | Zero `FAIL:` lines (WARN ≤ 3) |
| REQ-SC-003 | Word-count budget (each ≤ 1,500w, total ≤ 3,800w) | AC-SC-003 | Each + total budget met |
| REQ-SC-004 | Template-First mirror parity (byte-identical) | AC-SC-004 | `diff -q` exit 0 × 3 |
| REQ-SC-005 | Catalog SSOT consistency (7 removed + 3 added + hash OK) | AC-SC-005 | 0 old + ≥3 new + TestManifestHashFormat PASS |
| REQ-SC-006 | Cross-ref rename completeness (0 grep matches in scope) | AC-SC-006 | 0 matches |
| REQ-SC-007 | Test suite regression preservation | AC-SC-007 | 4 tests PASS |
| REQ-SC-008 | Skill body minimum contract (FM + 3 sections + absorbed-from) | AC-SC-002 (overlap) | Zero `FAIL:` (WARN ≤ 3) |
| REQ-SC-009 | Trigger keyword 비충돌 명시 (advisory) | AC-SC-009 | Manual review |
| (build sanity) | Cross-platform build (Linux + Windows) | AC-SC-008 | Both build exit 0 |

**Coverage**: 9 REQ → 9 AC (1:1 with REQ-SC-008 overlapping AC-SC-002 structure check). Full traceability.

---

## Edge Cases

### EC-SC-001 — Source skill 디렉토리에 SKILL.md 외 파일 (modules/, examples.md 등) 존재 시

**시나리오**: `moai-workflow-design-import/` 에 `examples.md` 또는 `modules/figma.md` 같은 추가 파일이 존재한다면, `rm -rf` 명령이 SKILL.md 외 컨텐츠도 함께 제거 → 정보 손실 위험.

**Treatment**:

- Pre-flight Check §G step 2 에 추가: `find .claude/skills/<old>/ -type f -not -name SKILL.md` 으로 추가 파일 발견 시 보고
- 발견 시: 통합 SKILL.md 의 footer 에 "[absorbed from <old-skill> including <extra-files>]" 명시 + 내용 본문에 흡수
- 발견 못한 경우 (대부분): SKILL.md 만 존재로 가정하고 `rm -rf` 정상 진행
- 검증: AC-SC-001 후 `ls .claude/skills/<old>/ 2>/dev/null` no output

### EC-SC-002 — Cross-reference grep 이 본문의 *변경 이력 (HISTORY) entry* 를 매치할 경우

**시나리오**: 통합 skill SKILL.md 의 footer "absorbed from moai-workflow-ci-watch" 표기가 AC-SC-006 grep 에 매치 → False positive FAIL.

**Treatment**:

- AC-SC-006 의 grep 검색 범위 의도 명시: `.moai/specs/SPEC-V3R6-SKILL-CONSOLIDATE-001/` 제외
- 통합 skill 자체의 footer absorption 표기는 `.claude/skills/moai-{workflow-ci-loop,workflow-design,harness-patterns}/SKILL.md` 에 위치하므로 검색 범위에 포함됨 → **이 경우 false positive 아님**. 통합 skill 자체는 absorbed-from footer 를 명시할 의무 (REQ-SC-008) — grep 매치가 발생하면 그 행은 expected.
- Resolution: AC-SC-006 verification 의 grep 결과를 line 단위로 검토. footer "absorbed from" 라인은 expected; 다른 위치 매치만 FAIL 로 처리. acceptance.md AC-SC-006 verification command 의 보강:

```bash
matches=$(grep -rln "moai-workflow-ci-watch\|moai-workflow-ci-autofix\|moai-workflow-design-import\|moai-workflow-design-context\|moai-harness-hook-ci\|moai-harness-workflow\|moai-harness-quality" \
  .claude/ CLAUDE.md internal/template/templates/.claude/ internal/design/dtcg/ 2>/dev/null \
  | grep -v ".claude/worktrees/" \
  | grep -v ".claude/agent-memory/" \
  | grep -v ".moai/specs/SPEC-V3R6-SKILL-CONSOLIDATE-001/" \
  | xargs -I {} sh -c 'grep -l "moai-workflow-ci-watch\|moai-workflow-ci-autofix\|moai-workflow-design-import\|moai-workflow-design-context\|moai-harness-hook-ci\|moai-harness-workflow\|moai-harness-quality" "{}" | xargs grep -L "absorbed from"' \
  | wc -l | tr -d ' ')
```

위 보강은 *"absorbed from" 라인이 있는 파일은 그 매치만 제외* 하지 않고 *파일 전체 제외* — 보수적. 더 정밀한 line-level 검증은 run-phase 작성자가 manual review.

**Simpler resolution**: 통합 SKILL.md 의 absorbed-from footer 를 *주석 형태* (HTML comment `<!-- absorbed from ... -->` 또는 markdown blockquote) 로 작성하여 grep target keyword 가 plain text 본문에 노출되지 않게. 그러면 AC-SC-006 grep 이 자연스럽게 매치 안 함.

### EC-SC-003 — `internal/design/dtcg/frozen_guard_test.go` 가 design skill name 을 hardcoded 로 reference

**시나리오**: Go test 코드 내 string literal `"moai-workflow-design-import"` 사용 → grep 매치 + Go build 통과하나 의미적으로 stale reference.

**Treatment**:

- Pre-flight Check §G step 5 grep 결과에 해당 파일 포함 확인됨
- M3 에서 명시적 처리: `Read` 로 line 위치 확인 → `Edit` 으로 새 이름 (`moai-workflow-design`) 치환
- 검증: `go test ./internal/design/dtcg/... -run TestFrozenGuard` PASS

---

## Quality Gate Summary

| Gate | Criterion | Status (filled by run-phase) |
|------|-----------|-------------------------------|
| AC-SC-001 | 14 paths absent | _TBD_ |
| AC-SC-002 | 6 files present + valid FM | _TBD_ |
| AC-SC-003 | Each ≤ 1,500w, total ≤ 3,800w | _TBD_ |
| AC-SC-004 | 3 diff -q pairs identical | _TBD_ |
| AC-SC-005 | Catalog SSOT + hash OK | _TBD_ |
| AC-SC-006 | 0 cross-ref matches in scope | _TBD_ |
| AC-SC-007 | 4 tests PASS | _TBD_ |
| AC-SC-008 | Both builds exit 0 | _TBD_ |
| AC-SC-009 | (advisory) Manual review | _TBD_ |
| **BLOCKING AC PASS rate** | 8/8 required for `status: implemented` | _TBD_ |

## Definition of Done

본 SPEC 가 `status: implemented` (v0.2.0) 로 전환되는 조건:

1. AC-SC-001 ~ AC-SC-008 (8 BLOCKING) 모두 PASS
2. AC-SC-009 advisory PASS (manual review by run-phase 작성자 또는 orchestrator)
3. NEW regression 0건 (pre-existing baseline failure 는 별도 SPEC scope, 본 SPEC 미해당)
4. PRESERVE list (plan.md §6) 위반 0건 — `git status --short` 비교 결과
5. `make build` 통과 + `go test ./...` aggregate 결과에 본 SPEC 도입으로 인한 NEW FAIL 0건
6. progress.md 작성 + spec.md frontmatter `status: implemented`, `version: 0.2.0`, `updated: <run-phase 완료일>` 갱신

**Stretch goal** (non-blocking):

- 통합 SKILL.md 의 word count 가 *목표* (1,200 / 1,400 / 1,200) 에 근접 — *상한* (1,500 × 3) 보다 더 타이트한 절감
- absorbed-from footer 가 HTML comment 또는 blockquote 형태로 grep target keyword 미노출 (EC-SC-002 resolution)
- 3 통합 skill 의 `## Works Well With` 섹션이 향후 SPEC-V3R6-SKILL-COMPRESS-001 의 5 압축 대상과 cross-ref 일관성 유지
