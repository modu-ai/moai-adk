# Test Scenarios — SPEC-V3R4-WORKFLOW-SPLIT-001

## Overview

본 문서는 Wave별 run-phase 진입 시 실행할 test scenarios + regression test list + Out-of-scope follow-up 후보를 정의한다.

테스트 실행 순서: **Wave 0 → Wave 1 → Wave 2 → Wave 3 → Wave 4** sequentially. 각 Wave 머지 후 다음 Wave 진입.

---

## Wave 0 — Preparation Infrastructure Tests

### Scenario W0-S1: Audit script baseline 검증 (현재 main HEAD)

- **Given** main HEAD `7a118e6b2`, audit script `scripts/audit-workflow-split.sh` 작성됨
- **When** `bash scripts/audit-workflow-split.sh` 실행
- **Then** exit code 0, output 마지막 줄에 `✓ All references valid (0 checked, 0 broken)` (split 전이므로 sub-skill 0개)

### Scenario W0-S2: LOC ceiling Go test baseline (의도된 FAIL)

- **Given** main HEAD, 4 workflow은 1073/1203/1076/932 LOC 상태
- **When** `go test -run TestEntryRouterLOCCeiling ./internal/skills/...` 실행
- **Then** Test FAIL, output 4개 entry router의 ceiling 위반 보고
- **Note**: 이는 의도된 baseline failure — split 후 PASS로 전환됨을 검증할 fixture

### Scenario W0-S3: Template mirror dir scaffold

- **Given** `internal/template/templates/.claude/skills/moai/workflows/` 디렉토리 존재
- **When** `mkdir -p internal/template/templates/.claude/skills/moai/workflows/{run,sync,project,plan}` 실행
- **Then** 4개 빈 디렉토리 생성, `git status -uall` 에서 untracked 인식 (gitkeep 없이도 OK — Wave별 sub-skill 추가 시 자동 add)

### Scenario W0-S4: SKILL.md hash baseline 캡처

- **Given** main HEAD `7a118e6b2`
- **When** `git rev-parse HEAD:.claude/skills/moai/SKILL.md > /tmp/skill-md-baseline.hash` 실행
- **Then** hash 파일 생성, 후속 Wave에서 비교 기준으로 사용

---

## Wave 1 — run.md split (3 sub-skills + entry router)

### Happy Path

#### Scenario W1-S1: Sub-skill 3개 생성 + LOC ceiling PASS

- **Given** Wave 1 T1.1 LOC verification 완료 (run.md 1073 LOC, phase boundaries identified)
- **When** T1.2 실행 후 `workflows/run/{context-loading,phase-execution,mode-orchestration}.md` 3개 생성
- **Then** `wc -l workflows/run/*.md` 결과 모두 ≤500 LOC AND `go test TestSubSkillLOCCeiling` PASS

#### Scenario W1-S2: Entry router refactor 후 ≤200 LOC

- **Given** Wave 1 T1.3 완료
- **When** `wc -l workflows/run.md` 실행
- **Then** 결과 ≤200 LOC AND frontmatter `user-invocable: true` 보존 AND 본문에 phase map + invocation flow 포함

#### Scenario W1-S3: Template mirror + make build

- **Given** Wave 1 T1.4 + T1.5 완료
- **When** `find` 비교 + `make build` + `git diff --stat internal/template/embedded.go`
- **Then** local과 template 파일 목록 100% 일치 AND embedded.go diff non-empty

#### Scenario W1-S4: `/moai run` regression smoke test

- **Given** Wave 1 머지 직전 (PR open, CI green)
- **When** `/moai run SPEC-DUMMY --dry-run` 실행 (dummy SPEC, agent spawn skipped)
- **Then** phase execution trace가 split 전 baseline (T0 시점 캡처)과 일치 (phase 순서, AskUserQuestion 호출, plan-auditor delegation, evaluator-active gate 모두 동일)

### Edge Cases

#### Scenario W1-EC1: `phase-execution.md` 가 ≤500 LOC 초과 (predicted 745 LOC)

- **Given** Wave 1 T1.1 LOC verification 결과 `phase-execution.md` 추정 745 LOC
- **When** Wave 1 T1.2 실행 전 추가 sub-split 결정
- **Then** `phase-execution.md` 를 2개로 분할 (`phase-execution.md` + `task-decomposition.md`), 둘 다 ≤500 LOC AND Wave 1 sub-skill 총 개수 3 → 4

#### Scenario W1-EC2: 내부 anchor link 깨짐

- **Given** original `run.md` 에 `#phase-3-1` 등 anchor 존재
- **When** sub-skill 분할 후 cross-ref audit 실행
- **Then** broken anchor 발견 시 entry router에 anchor map 명시 + sub-skill 내 anchor 보존

#### Scenario W1-EC3: SKILL.md 의도치 않은 변경 발견

- **Given** Wave 1 PR open 상태
- **When** `git diff main HEAD -- .claude/skills/moai/SKILL.md` 실행
- **Then** 결과 empty 기대; non-empty 발견 시 즉시 revert + AC-WFSP-003 hash 재검증

---

## Wave 2 — sync.md split (3 sub-skills + entry router)

### Happy Path

#### Scenario W2-S1: HUMAN GATE 5개 모두 보존

- **Given** Wave 2 sub-skill 생성 완료 (quality-gates.md + doc-execution.md + delivery.md)
- **When** grep으로 "HUMAN GATE" 패턴 검색
- **Then** 5개 GATE 모두 발견 (Phase 0 HUMAN GATE 1/2/3 in quality-gates.md, Phase 1 HUMAN GATE in doc-execution.md, Phase 3 stash GATE in delivery.md)
- **And** entry router에 GATE map 명시 (어느 sub-skill의 어느 GATE인지)

#### Scenario W2-S2: `/moai sync` regression smoke test

- **Given** Wave 2 머지 직전
- **When** `/moai sync SPEC-DUMMY --dry-run` 실행
- **Then** Phase 0 → Phase 0.5 → Phase 0.7 → Phase 1 → Phase 2 → Phase 3 sequence 보존 AND 5 HUMAN GATE trigger 모두 일치

### Edge Cases

#### Scenario W2-EC1: `quality-gates.md` borderline 540 LOC

- **Given** Wave 2 T2.1 LOC verification 결과 540 LOC
- **When** sub-split 결정
- **Then** Phase 0~0.4 (~320 LOC) + Phase 0.5~0.7 (~220 LOC) 로 분할, 둘 다 ≤500 LOC

#### Scenario W2-EC2: Wave 1 lesson learned 반영

- **Given** Wave 1 머지 완료 + post-merge 회고
- **When** Wave 2 T2.1 진입
- **Then** Wave 1에서 발견된 LOC borderline 패턴, anchor 깨짐 패턴, template mirror 검증 절차를 Wave 2 작업 시작 전 체크리스트로 적용

---

## Wave 3 — project.md split (4 sub-skills + entry router)

### Happy Path

#### Scenario W3-S1: 4 sub-skill 생성 + sequential cross-ref

- **Given** Wave 3 T3.2 완료
- **When** sub-skill 4개 (mode-detection / codebase-analysis / doc-generation / meta-harness) 생성
- **Then** 모두 ≤500 LOC AND cross-ref audit: meta-harness → doc-generation → codebase-analysis 호출 패턴 검증 (broken 0)

#### Scenario W3-S2: Phase 1.5 3-round AskUserQuestion 한 sub-skill 내 유지

- **Given** `project/codebase-analysis.md` 작성
- **When** grep으로 "AskUserQuestion" + "Round 1/2/3" 검색
- **Then** 3 round 모두 동일 sub-skill (codebase-analysis.md) 내 위치 (split 시 round 분리 금지)

#### Scenario W3-S3: `/moai project` regression smoke test

- **Given** Wave 3 머지 직전
- **When** `/moai project --explain --dry-run` 실행
- **Then** Phase 0 → 0.3 → 1 → 1.5 (3 rounds) → 2 → 3 → 3.1 → 3.3 → 3.5 → 3.7 → 4.1a → 4 → 5 → 6 sequence 보존

### Edge Cases

#### Scenario W3-EC1: `doc-generation.md` 470 LOC + `--from-brain` flag cross-ref

- **Given** `doc-generation.md` 의 `--from-brain` 처리 로직이 `mode-detection.md` 와 cross-ref
- **When** cross-ref audit
- **Then** entry router에 명시적 sub-skill map 표기 + audit script PASS

---

## Wave 4 — plan.md split (3 sub-skills + entry router, self-referential)

### Critical Pre-condition

#### Scenario W4-PC1: Open plan-PR 0건 확인

- **Given** Wave 4 진입 직전
- **When** `gh pr list --state open --label 'type:feature' --search 'plan/'` 실행
- **Then** 결과 0건 — Wave 4 작업 중 plan workflow 자체 변경이 다른 active plan-PR과 충돌하지 않도록 보장
- **Mitigation**: 만약 open plan-PR 발견 시 Wave 4 진입 보류, 해당 PR 머지 후 진입

### Happy Path

#### Scenario W4-S1: clarity-interview.md 475 LOC + Annotation Cycle 보존

- **Given** Wave 4 T4.2 완료
- **When** `wc -l workflows/plan/clarity-interview.md` + grep "Annotation Cycle"
- **Then** ≤500 LOC AND Annotation Cycle 한 sub-skill 내 유지 AND Decision Point 1 trigger 일치

#### Scenario W4-S2: `/moai plan` regression smoke test (self-referential)

- **Given** Wave 4 머지 직전, 본 SPEC scope plan workflow 자체가 split된 상태
- **When** `/moai plan "dummy test feature" --dry-run` 실행 (실제 SPEC 작성은 stop)
- **Then** Phase 1A → 0.3 → 0.3.1 → 0.4 → 0.5 → 1.25 → 1B → 1.5 → 2 → 2.3 → 2.5 → 3 → 3.5 → 3.6 sequence 보존
- **And** Decision Point 1/2/3.5 trigger 동일 AND plan-auditor delegation 보존

### Edge Cases

#### Scenario W4-EC1: clarity-interview.md ≤500 LOC borderline

- **Given** Wave 4 T4.1 LOC verification 결과 borderline (예: 510 LOC)
- **When** 추가 sub-split 결정
- **Then** `plan/annotation-cycle.md` 별도 분리, 4 sub-skill로 증가 (acceptable)

#### Scenario W4-EC2: Wave 4 진행 중 다른 SPEC plan-PR 등장

- **Given** Wave 4 PR open 상태에서 새로운 SPEC plan-PR이 다른 작업자에 의해 생성됨
- **When** 충돌 가능성 확인
- **Then** Wave 4 PR 우선 머지 (이미 in-progress) AND 새 plan-PR 작업자에게 Wave 4 머지 후 rebase 권장 안내

---

## Cross-Wave Regression Test (Wave 4 머지 후)

### Scenario CW-S1: 4 slash command 모두 invocation 정상

- **Given** Wave 4 머지 완료, 4 workflow 모두 split된 상태
- **When** 4 slash command 각 1회 dry-run 실행:
  ```bash
  /moai plan "test feature" --dry-run
  /moai run SPEC-DUMMY --dry-run
  /moai sync SPEC-DUMMY --dry-run
  /moai project --explain --dry-run
  ```
- **Then** 4건 모두 phase execution trace가 split 전 baseline과 일치

### Scenario CW-S2: `moai init` 사용자 프로젝트 회귀 방지

- **Given** Wave 4 머지 완료
- **When** `moai init /tmp/test-post-split-$(date +%s)` 실행
- **Then** 사용자 프로젝트 `.claude/skills/moai/workflows/{run,sync,project,plan}/` 하위에 각각 sub-skill 디렉토리 + 파일 배포 확인
- **And** 사용자 프로젝트에서 `/moai plan "test"` 실행 가능 (entry router → sub-skill 로딩 정상)

### Scenario CW-S3: Total token-load reduction 측정

- **Given** Wave 4 완료 후
- **When** 4 workflow의 active context token 측정 (manual or 자동화)
- **Then** 4-workflow aggregate: ~42K tokens → ~10K tokens = ~76% 감소 달성 (Goal #1)

### Scenario CW-S4: SKILL.md hash 최종 검증

- **Given** Wave 4 머지 완료
- **When** `git rev-parse HEAD:.claude/skills/moai/SKILL.md` 와 T0.4 baseline 비교
- **Then** hash 동일 (Intent Router 무변경 100% 보장)

---

## Out of Scope (Follow-up SPECs)

본 SPEC에서 제외되며, 별도 SPEC 또는 chore PR로 처리해야 할 항목:

### OoS-1: F-015 docs drift chore

- **항목**: hooks-system.md (또는 관련 문서) 에 UserPromptExpansion / PostToolBatch / mcp_tool 5 events 누락
- **출처**: Workflow Audit 2026-05-16 finding F-015
- **처리 방식**: 별도 chore PR (SPEC 없음, 단순 docs sync)
- **이유**: Bundle F (workflow split) 범위 밖, docs drift는 별개 issue

### OoS-2: SPEC-V3R4-LLM-REVIEW-CI-001 (Claude/GLM/Gemini/Codex CLI not found in CI runner)

- **항목**: CI runner에 LLM CLI 4종 미설치로 인한 review workflow 실패
- **출처**: Workflow Audit 별도 finding
- **처리 방식**: 별도 SPEC `SPEC-V3R4-LLM-REVIEW-CI-001` 작성 후 plan → run → sync lifecycle
- **이유**: 본 SPEC은 markdown reorganization, CI runner 환경 변경과 무관

### OoS-3: SPEC-V3R4-WORKFLOW-SPLIT-001-DOCS-FOLLOWUP (조건부 생성)

- **항목**: docs-site 4-locale에 workflow skill 직접 reference 발견 시 sync 작업
- **조건**: Wave별 PR 머지 직전 grep 결과가 ≥1 matches 일 때 발동
- **현재 상태**: pre-write grep 결과 0건, 따라서 본 follow-up SPEC 미생성 (현재 시점)
- **재검증 시기**: Wave 1/2/3/4 각 PR 머지 직전
- **처리 방식**: 발동 시 별도 SPEC 작성, CLAUDE.local.md §17 docs-site 4-locale sync 규칙 준수

### OoS-4: Sub-skill 추가 분할 (5+ split)

- **항목**: 본 SPEC은 13 sub-skill까지 분할 (최대 15 if edge case). 만약 미래 사용 추이상 (예: phase-execution.md 가 1500 LOC로 성장) 추가 분할 필요 시
- **처리 방식**: 별도 SPEC `SPEC-V3R4-WORKFLOW-SPLIT-002` (또는 -003 ...)
- **트리거**: `TestSubSkillLOCCeiling` 가 미래 PR에서 FAIL 시 자동 식별

### OoS-5: SPEC.md Intent Router 동시 분할

- **항목**: 만약 미래에 SKILL.md (Intent Router) 자체도 500 LOC 초과로 split 필요 시
- **처리 방식**: 별도 SPEC. **본 SPEC scope 내에서는 SKILL.md 무변경 강제** (AC-WFSP-003)
- **이유**: SKILL.md 변경은 모든 workflow 진입점에 영향, 본 SPEC과 isolated scope 필요

---

## Test Execution Order Summary

```
Wave 0 (prep): T0.1~T0.4 → W0-S1~W0-S4 검증
   ↓ PR merged
Wave 1 (run): T1.1~T1.8 → W1-S1~W1-S4 + W1-EC1~W1-EC3 검증
   ↓ PR merged + lesson learned
Wave 2 (sync): T2.1~T2.8 → W2-S1~W2-S2 + W2-EC1~W2-EC2 검증
   ↓ PR merged
Wave 3 (project): T3.1~T3.8 → W3-S1~W3-S3 + W3-EC1 검증
   ↓ PR merged
Wave 4 (plan, self-referential): T4.1~T4.8 → W4-PC1 + W4-S1~W4-S2 + W4-EC1~W4-EC2 검증
   ↓ PR merged
Cross-Wave Final: CW-S1~CW-S4 검증
   ↓ DoD 체크리스트 (acceptance.md § Definition of Done)
SPEC status: completed (별도 sync-PR)
```

---

## Acceptance Test Tooling

### Bash audit script (T0.1)

- **파일**: `scripts/audit-workflow-split.sh`
- **언어**: Bash (CI 호환)
- **외부 의존**: grep, find, awk
- **호출**: `bash scripts/audit-workflow-split.sh [--strict]`
- **출력**: stdout에 검증 결과, exit 0/1

### Go test (T0.2)

- **파일**: `internal/skills/workflow_split_test.go`
- **테스트 함수**:
  - `TestSubSkillLOCCeiling`: 모든 `workflows/{name}/*.md` ≤500 LOC
  - `TestEntryRouterLOCCeiling`: 4 entry router ≤200 LOC
  - `TestTemplateMirrorParity`: local과 template 디렉토리 비교
  - `TestSKILLmdUnchanged`: SKILL.md hash baseline 비교
- **호출**: `go test ./internal/skills/...`

### Slash command dry-run (regression)

- **메커니즘**: 각 slash command가 `--dry-run` flag 지원 (현재 구현 여부 확인 필요)
- **만약 미지원**: phase execution log를 stdout에 출력하는 trace mode 추가 (별도 Wave 0 task)
- **대안**: agent spawn 직전 단계까지만 진행, phase sequence log를 stderr에 캡처

---

## References

- spec.md § EARS Requirements (REQ-WFSP-001~005)
- plan.md § Wave-by-Wave Tasks (T0.1~T4.8)
- design.md § Architectural Decisions (AD-001~006)
- acceptance.md § Quality Gates (AC-WFSP-001~008)
- `.moai/research/workflow-audit-2026-05-16.md` Bundle F
- lessons #9 wave-split
