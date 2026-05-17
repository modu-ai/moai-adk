# Acceptance Criteria — SPEC-V3R4-WORKFLOW-SPLIT-001

## Quality Gates Overview

본 SPEC은 8개 acceptance criteria로 구성되며, 모두 binary PASS/FAIL 검증 가능. Wave별로 점진 적용되어 Wave 4 머지 시점에 모두 PASS 상태여야 한다.

| AC ID | 제목 | 검증 방식 | Wave 적용 시점 |
|-------|------|----------|--------------|
| AC-WFSP-001 | Sub-skill LOC ceiling | Go test (`TestSubSkillLOCCeiling`) | Wave 1-4 (점진) |
| AC-WFSP-002 | Entry router LOC ceiling | Go test (`TestEntryRouterLOCCeiling`) | Wave 1-4 (점진) |
| AC-WFSP-003 | Intent Router unchanged | hash 비교 | Wave 0-4 (전체) |
| AC-WFSP-004 | Slash command regression 0 | dry-run trace diff | Wave 1-4 (해당 workflow) |
| AC-WFSP-005 | Cross-reference integrity | bash audit script | Wave 1-4 (점진) |
| AC-WFSP-006 | Spec lint clean | `moai spec lint --strict` | 매 PR |
| AC-WFSP-007 | docs-site impact documented | 본 design.md/spec.md 자체 | 본 plan-PR |
| AC-WFSP-008 | Template synchronization | `TestTemplateMirrorParity` | Wave 1-4 |

---

## AC-WFSP-001 — Sub-skill LOC Ceiling

### Statement

각 phase-scoped sub-skill의 LOC는 500 이하를 유지한다. 위반 시 추가 sub-split 강제.

### Given/When/Then

**Scenario 1: Wave 1 run.md split**

- GIVEN main HEAD 직후 `workflows/run/` 디렉토리에 3 sub-skill 생성됨
- WHEN `wc -l workflows/run/*.md` 실행
- THEN 모든 결과 ≤500 LOC AND Go test `TestSubSkillLOCCeiling` PASS

**Scenario 2: Wave 2 sync.md split — borderline case**

- GIVEN `workflows/sync/quality-gates.md` 추정 540 LOC
- WHEN Wave 2 T2.1 LOC verification 단계 실행
- THEN measured LOC > 500이면 `quality-gates.md` 를 2개로 추가 분할 (예: `quality-gates-phase0.md` + `quality-gates-phase0-7.md`)
- AND 분할 후 양쪽 모두 ≤500 LOC AND Go test PASS

### Verification

```bash
# Go test
go test -run TestSubSkillLOCCeiling ./internal/skills/...

# Manual verification
find .claude/skills/moai/workflows -name '*.md' -mindepth 2 \
  -exec wc -l {} \; | awk '$1 > 500 { print "VIOLATION:", $0; exit 1 }'
```

### Acceptance: Binary

- PASS: 모든 sub-skill ≤500 LOC
- FAIL: 1개 이상 위반 → Wave PR block

---

## AC-WFSP-002 — Entry Router LOC Ceiling

### Statement

각 entry router skill (`workflows/run.md`, `sync.md`, `project.md`, `plan.md`)은 ≤200 LOC.

### Given/When/Then

**Scenario 1: Entry router refactor**

- GIVEN Wave 1 T1.3에서 `workflows/run.md` 1073 LOC → router로 재작성
- WHEN `wc -l workflows/run.md` 실행
- THEN 결과 ≤200 LOC AND router 본문에 phase map + invocation flow 포함 AND `user-invocable: true` frontmatter 보존

**Scenario 2: Entry router 무변경 (다른 Wave 진행 중)**

- GIVEN Wave 1 진행 중, sync.md/project.md/plan.md 는 아직 split 되지 않음
- WHEN `wc -l workflows/sync.md workflows/project.md workflows/plan.md` 실행
- THEN 결과는 원본 LOC (1203/1076/932)로 unchanged — Wave별 점진 적용이므로 정상
- AND CI Go test는 skip-annotated 또는 conditional assertion으로 처리

### Verification

```bash
go test -run TestEntryRouterLOCCeiling ./internal/skills/...

# Manual
for f in .claude/skills/moai/workflows/{run,sync,project,plan}.md; do
  loc=$(wc -l < "$f")
  if [ "$loc" -gt 200 ]; then
    # 해당 Wave가 머지된 후에만 enforce
    echo "$f: $loc LOC"
  fi
done
```

### Acceptance: Binary

- Wave N 머지 후: 해당 entry router ≤200 LOC
- Wave 4 완료 후: 4개 entry router 모두 ≤200 LOC

---

## AC-WFSP-003 — Intent Router Unchanged (SKILL.md byte-for-byte)

### Statement

`.claude/skills/moai/SKILL.md` 는 본 SPEC 전체 lifecycle 동안 byte-for-byte 무변경 유지.

### Given/When/Then

**Scenario 1: Plan-PR 머지 시점 baseline 캡처**

- GIVEN main HEAD `7a118e6b2` 에서 SKILL.md hash 측정
- WHEN `git rev-parse HEAD:.claude/skills/moai/SKILL.md` 실행
- THEN hash 값 `<H_BASELINE>` 기록 (T0.4)

**Scenario 2: Wave N PR 머지 직전 검증**

- GIVEN Wave N PR이 머지 후보 상태
- WHEN `git diff main HEAD -- .claude/skills/moai/SKILL.md` 실행
- THEN 결과 empty (0 lines changed)
- AND `git rev-parse HEAD:.claude/skills/moai/SKILL.md` == `<H_BASELINE>`

### Verification

```bash
# Baseline (Wave 0)
git rev-parse main:.claude/skills/moai/SKILL.md > /tmp/skill-md-baseline.hash

# Per Wave
current=$(git rev-parse HEAD:.claude/skills/moai/SKILL.md)
baseline=$(cat /tmp/skill-md-baseline.hash)
[ "$current" = "$baseline" ] || { echo "VIOLATION: SKILL.md modified"; exit 1; }
```

### Acceptance: Binary

- PASS: hash 동일 (Intent Router 무변경)
- FAIL: hash 다름 → 즉시 revert + Wave PR block

---

## AC-WFSP-004 — Slash Command Regression 0

### Statement

`/moai plan`, `/moai run`, `/moai sync`, `/moai project` invocation은 split 전과 동일한 phase execution trace 생성.

### Given/When/Then

**Scenario 1: Wave 1 후 `/moai run` regression**

- GIVEN Wave 1 머지 직전 (PR open 상태)
- WHEN `/moai run SPEC-DUMMY` dry-run 실행 (agent spawn skipped, phase log only)
- THEN phase trace 출력은 다음 sequence 포함:
  ```
  Phase 0: Context Loading [from workflows/run/context-loading.md]
  Phase 1: Execution [from workflows/run/phase-execution.md]
  ...
  ```
- AND split 전 baseline trace와 비교 시 phase 순서, 산출물, AskUserQuestion 호출 패턴 100% 일치

**Scenario 2: Wave 4 후 `/moai plan` regression (self-referential)**

- GIVEN Wave 4 머지 직전, plan workflow 자체가 split된 상태
- WHEN `/moai plan "dummy test feature"` dry-run 실행
- THEN phase trace는 Phase 1A → 0.3 → 0.4 → 0.5 → 1.25 → 1B → 1.5 → 2 → 2.3 → 2.5 → 3 → 3.5 → 3.6 sequence 보존
- AND Annotation Cycle, Decision Point 1/2/3.5 trigger 동일

### Verification

```bash
# Pre-Wave baseline (main HEAD `7a118e6b2` 시점에 캡처)
/moai run SPEC-DUMMY --dry-run > /tmp/run-trace-baseline.log

# Post-Wave
/moai run SPEC-DUMMY --dry-run > /tmp/run-trace-after.log

# Diff
diff /tmp/run-trace-baseline.log /tmp/run-trace-after.log
# Allowed differences: file path mentions (workflows/run.md → workflows/run/context-loading.md)
# Disallowed: missing phase, reordered phase, missing AskUserQuestion call
```

### Acceptance: Binary

- PASS: phase 순서 + 산출물 + interaction pattern 100% 일치
- FAIL: 1건 이상 missing/reordered phase

---

## AC-WFSP-005 — Cross-Reference Integrity

### Statement

모든 sub-skill 간 reference link는 실제 파일 경로로 resolve. 깨진 link 0건.

### Given/When/Then

**Scenario 1: Wave 1 cross-ref audit**

- GIVEN Wave 1 sub-skills 3개 + entry router 생성됨
- WHEN `bash scripts/audit-workflow-split.sh` 실행
- THEN 모든 `Read workflows/...` 패턴이 실제 파일에 resolve AND output 마지막 줄에 `✓ All references valid (N checked, 0 broken)`

**Scenario 2: 의도된 broken link 도입 (fixture test)**

- GIVEN audit script test fixture (가상의 broken ref)
- WHEN audit script 실행
- THEN script exits with code 1 AND error message `BROKEN REF: <file>:<line>` 표시

### Verification

```bash
bash scripts/audit-workflow-split.sh
echo "Exit: $?"  # 0 = PASS

# Manual cross-check
grep -rn "Read workflows/" .claude/skills/moai/workflows/ | while IFS=: read -r file lineno text; do
  # extract target path from "Read workflows/..." pattern
  target=$(echo "$text" | sed -nE 's/.*Read (workflows\/[^[:space:]]+).*/\1/p')
  if [ -n "$target" ] && [ ! -f ".claude/skills/moai/$target" ]; then
    echo "BROKEN: $file:$lineno → $target"
  fi
done
```

### Acceptance: Binary

- PASS: 0 broken refs
- FAIL: 1+ broken refs → Wave PR block

---

## AC-WFSP-006 — Spec Lint Clean

### Statement

`moai spec lint --strict` 가 본 SPEC 디렉토리 + main HEAD 전체에 대해 `✓ No findings` 보고.

### Given/When/Then

**Scenario 1: Plan-PR 머지 직전 lint**

- GIVEN 본 plan-PR (spec.md + plan.md + design.md + acceptance.md + scenarios.md) 작성 완료
- WHEN `moai spec lint --strict` 실행 (main HEAD 기준)
- THEN exit code 0 AND output `✓ No findings`

**Scenario 2: Wave 1-4 PR 머지 직전 lint**

- GIVEN Wave N PR 작성 완료 (sub-skill 추가 + entry router refactor + template mirror)
- WHEN `moai spec lint --strict` 실행
- THEN exit code 0 AND `✓ No findings`
- AND 본 SPEC frontmatter 12-field canonical schema 준수

### Verification

```bash
moai spec lint --strict
# exit 0 expected
# Output should contain "✓ No findings"
```

### Acceptance: Binary

- PASS: 0 findings
- FAIL: 1+ findings → 수정 후 재실행

---

## AC-WFSP-007 — docs-site Impact Documented

### Statement

spec.md 또는 design.md 에 docs-site (4-locale) impact analysis 섹션 명시. 만약 reference 발견 시 follow-up SPEC 후보 식별.

### Given/When/Then

**Scenario 1: Pre-write grep 결과 0건 (현재 상태)**

- GIVEN main HEAD `7a118e6b2` 에서 `grep -rln "workflows/run\|workflows/sync\|workflows/project\|workflows/plan" docs-site/` 실행
- WHEN 결과 0 matches
- THEN design.md §"docs-site Impact Analysis" 섹션에 "0 matches, follow-up 불필요" 기록 (이미 완료됨)

**Scenario 2: Run-phase 재검증 (Wave별)**

- GIVEN Wave N PR 머지 직전
- WHEN 동일 grep 재실행
- THEN 0건 유지면 정상; ≥1건이면 `SPEC-V3R4-WORKFLOW-SPLIT-001-DOCS-FOLLOWUP` 별도 SPEC 생성 (scenarios.md "Out of Scope")

### Verification

```bash
grep -rln "workflows/run\|workflows/sync\|workflows/project\|workflows/plan" docs-site/ | wc -l
# expected: 0
```

### Acceptance: Binary

- PASS: design.md § docs-site Impact Analysis 섹션 존재 AND grep 결과 0건 (현재 상태) 기록
- FAIL: 섹션 누락 또는 ≥1 matches 시 follow-up SPEC 미생성

---

## AC-WFSP-008 — Template Synchronization

### Statement

`internal/template/templates/.claude/skills/moai/workflows/` 가 local과 1:1 동기화. `make build` 실행으로 `embedded.go` 재생성.

### Given/When/Then

**Scenario 1: Wave 1 template mirror**

- GIVEN Wave 1 T1.4 실행 후 (template mirror commit)
- WHEN `find .claude/skills/moai/workflows -name '*.md'` 와 `find internal/template/templates/.claude/skills/moai/workflows -name '*.md'` 비교
- THEN 파일 목록 100% 일치 AND 각 파일 내용 byte-identical

**Scenario 2: make build 후 embedded.go diff**

- GIVEN Wave N T*.5 `make build` 실행
- WHEN `git diff --stat internal/template/embedded.go` 실행
- THEN diff 결과 non-empty (재생성됨) AND `git status` 에서 staged 상태

**Scenario 3: moai init 사용자 프로젝트 회귀 방지**

- GIVEN Wave 4 머지 완료
- WHEN `moai init /tmp/test-post-split-$(date +%s)` 실행
- THEN 사용자 프로젝트 `/tmp/test-post-split-*/...claude/skills/moai/workflows/{run,sync,project,plan}/*.md` 전체 sub-skill 배포 확인

### Verification

```bash
# Parity check
diff <(find .claude/skills/moai/workflows -name '*.md' | sort) \
     <(find internal/template/templates/.claude/skills/moai/workflows -name '*.md' | \
       sed 's|internal/template/templates/||' | sort)

# embedded.go diff
make build
git diff --stat internal/template/embedded.go
# expected: non-empty

# Post-merge user project test
TMP=$(mktemp -d)
moai init "$TMP/test-project"
ls -la "$TMP/test-project/.claude/skills/moai/workflows/run/"
# expected: 3 sub-skill files
```

### Acceptance: Binary

- PASS: parity + embedded.go regenerated + moai init test 모두 OK
- FAIL: 1건 이상 누락 → Wave PR block

---

## Performance Gates

본 SPEC은 markdown reorganization이므로 runtime performance 직접 영향 없음. 단 token-load reduction을 비기능적 목표로 추적.

| Metric | Baseline | Target | Verification |
|--------|----------|--------|-------------|
| `/moai run` workflow context tokens | ~10,000 | ≤2,500 | manual measurement (Wave 1 후) |
| `/moai sync` workflow context tokens | ~12,000 | ≤2,500 | Wave 2 후 |
| `/moai project` workflow context tokens | ~10,500 | ≤2,500 | Wave 3 후 |
| `/moai plan` workflow context tokens | ~9,000 | ≤2,500 | Wave 4 후 |

**Total token-load reduction (Wave 4 완료 후)**: ~42K tokens → ~10K tokens = **~76% 감소** (4-workflow aggregate).

---

## Edge Cases

### EC-1: Sub-skill phase content가 future maintenance에서 >500 LOC로 성장

- 발생 시기: 본 SPEC 머지 후 6개월 ~ 1년 후
- 탐지: 미래 PR에서 `TestSubSkillLOCCeiling` FAIL
- 대응: 별도 SPEC `SPEC-V3R4-WORKFLOW-SPLIT-002` 로 추가 sub-split
- 자동화: CI test가 영구 enforce하므로 silent drift 불가

### EC-2: Template sync race during partial Wave merge

- 발생 시기: Wave 2 PR open 상태에서 Wave 1 hotfix가 동일 디렉토리 건드림
- 탐지: Wave 2 PR rebase 시 conflict
- 대응: Wave별 PR은 순차 진행, 동시 open 금지 (plan.md Implementation Sequence)
- 방지: PR 머지 후 다음 Wave 진입 (gates: CI green + admin merge 확인)

### EC-3: SKILL.md Intent Router 의도치 않은 변경

- 발생 시기: Wave PR에서 typo fix 같은 부수 변경
- 탐지: AC-WFSP-003 hash 비교 FAIL
- 대응: 즉시 revert SKILL.md 변경, 본 SPEC scope 외 작업으로 별도 PR 분리

### EC-4: `phase-execution.md` 또는 `quality-gates.md` 가 ≤500 LOC 실패

- 발생 시기: Wave 1 T1.1 또는 Wave 2 T2.1 LOC verification 단계
- 대응: 추가 sub-split 결정. plan.md Risk Table에 명시된 보조 sub-skill 생성:
  - Wave 1: `run/task-decomposition.md` 추가 (4 sub-skill)
  - Wave 2: `sync/quality-gates-phase0.md` + `sync/quality-gates-phase0-7.md` 분할 (4 sub-skill)
- 영향: sub-skill 총 개수 13 → 14 or 15 (acceptable)

---

## Definition of Done (Wave 4 머지 완료 후)

- [ ] AC-WFSP-001~008 모두 PASS
- [ ] 4 entry router 모두 ≤200 LOC
- [ ] 13 (또는 추가 분할 시 14-15) sub-skill 모두 ≤500 LOC
- [ ] SKILL.md hash 무변경 (T0.4 baseline 일치)
- [ ] `moai spec lint --strict` `✓ No findings`
- [ ] 4 slash command regression test PASS
- [ ] Cross-reference audit 0 broken
- [ ] Template parity 100% (local ↔ template)
- [ ] `moai init /tmp/test-X` 사용자 프로젝트에 sub-skill 디렉토리 배포 확인
- [ ] CI all-green: golangci-lint clean, Go test all PASS, multi-OS (ubuntu/macos/windows) build PASS
- [ ] 본 SPEC frontmatter `status: completed` 로 업데이트 (별도 sync-PR)
