---
id: SPEC-ASKUSER-ENFORCE-001
version: "1.0.0"
status: draft
created: "2026-04-25"
updated: "2026-04-25"
author: GOOS
priority: high
issue_number: 0
depends_on: []
---

# SPEC-ASKUSER-ENFORCE-001 Task Decomposition

> 작업 분해, REQ/AC 추적성, Phase 매핑, TRUST 5 gate
> 최종 갱신: 2026-04-25
> Task 개수: **18**
> Phase 매핑: plan.md Phase A → H

---

## 범례

- **Owner-role**: `implementer` (구현), `tester` (검증), `reviewer` (검토). 본 SPEC은 코드 추가가 없으므로 implementer = 문서 편집자, tester = grep/diff 검증자.
- **Isolation**: 본 SPEC은 모두 `.claude/`, `.moai/specs/`, `internal/template/templates/`, memory 영역만 수정 — worktree 불필요.
- **Blocks**: 후속 task ID.
- **REQ**: 매핑되는 EARS 요구사항.
- **AC**: 매핑되는 acceptance criteria.
- **DoD**: Definition of Done — TRUST 5 gate 포함.

---

## Phase A: askuser-protocol.md canonical reference 작성

### T-01: askuser-protocol.md §1 Channel Monopoly 작성

- **Phase**: A
- **Owner-role**: implementer
- **REQ**: REQ-AUE-001, REQ-AUE-007
- **AC**: AC-AUE-001, AC-AUE-007, AC-AUE-008-5
- **의존성**: 없음 (root)
- **Blocks**: T-08 (다른 파일 cross-reference 대상)
- **설명**: `.claude/rules/moai/core/askuser-protocol.md` 신규 파일을 생성하고 §1 Channel Monopoly 섹션 작성. 본 섹션은 "AskUserQuestion is the only user-facing question channel" 규범과 free-form 우회 금지 규범을 명시한다.
- **DoD**:
    - [ ] 파일 존재 확인 — `test -f .claude/rules/moai/core/askuser-protocol.md`
    - [ ] §1 헤더 존재 — `grep -E '^## Channel Monopoly' askuser-protocol.md` ≥ 1
    - [ ] AC-AUE-001 grep 통과
    - [ ] TRUST 5 Readable: 한국어/영어 혼용 시 일관 어휘 사용

### T-02: askuser-protocol.md §2 ToolSearch Preload Procedure 작성

- **Phase**: A
- **Owner-role**: implementer
- **REQ**: REQ-AUE-002
- **AC**: AC-AUE-002, AC-AUE-008-5
- **의존성**: T-01
- **Blocks**: T-04, T-05, T-06, T-07
- **설명**: §2 섹션 작성. `ToolSearch(query: "select:AskUserQuestion")` 사전 호출 절차를 명시. deferred tool 일반화 표현("any deferred tool requires ToolSearch select preload before invocation")을 함께 명시.
- **DoD**:
    - [ ] 마커 존재 — `grep 'ToolSearch.*select:AskUserQuestion' askuser-protocol.md` ≥ 1
    - [ ] 일반화 표현 존재 — `grep -E 'deferred tool.*(ToolSearch|preload)' askuser-protocol.md` ≥ 1
    - [ ] AC-AUE-002 grep 통과

### T-03: askuser-protocol.md §3-§7 잔여 섹션 작성

- **Phase**: A
- **Owner-role**: implementer
- **REQ**: REQ-AUE-003, REQ-AUE-004, REQ-AUE-005, REQ-AUE-006, REQ-AUE-007
- **AC**: AC-AUE-003, AC-AUE-004, AC-AUE-005, AC-AUE-006, AC-AUE-007, AC-AUE-008-5
- **의존성**: T-02
- **Blocks**: T-08, T-09, T-10, T-11
- **설명**: §3 Socratic Interview Structure (4Q/4O/`(권장)`/conversation_language), §4 Option Description Standards (description 필수 항목 + bias prevention), §5 Orchestrator–Subagent Boundary (양방향 codify + blocker report 형식), §6 Ambiguity Triggers and Exceptions (4 trigger + 5 exception, CLAUDE.md §7 Rule 5 / §8 단일 출처 명시), §7 Free-form Circumvention Prohibition (Other 옵션 안내) 작성.
- **DoD**:
    - [ ] 7개 섹션 헤더 모두 존재 — `grep -cE '^## (Channel Monopoly|ToolSearch Preload|Socratic Interview|Option Description|Orchestrator.*Subagent|Ambiguity Triggers|Free-form)' askuser-protocol.md` ≥ 7
    - [ ] AC-AUE-003 ~ AC-AUE-007 grep 모두 통과
    - [ ] AC-AUE-008-5 통과
    - [ ] TRUST 5 Unified: 7개 섹션의 헤더 어휘 / 들여쓰기 / 코드 블록 스타일 일관

---

## Phase B: CLAUDE.md §7 Rule 5 / §8 보강

### T-04: CLAUDE.md §7 Rule 5 Discovery process에 ToolSearch step 삽입

- **Phase**: B
- **Owner-role**: implementer
- **REQ**: REQ-AUE-002, REQ-AUE-006
- **AC**: AC-AUE-002, AC-AUE-006, AC-AUE-008-1
- **의존성**: T-02
- **Blocks**: T-12 (template mirror)
- **설명**: §7 Rule 5 "Discovery process" 7-step 시퀀스에서 "Conduct Socratic interview via AskUserQuestion" 직전에 새 step "First, preload AskUserQuestion schema via `ToolSearch(query: \"select:AskUserQuestion\")` (deferred tool prerequisite)" 삽입. step 번호 재정렬.
- **DoD**:
    - [ ] `grep 'ToolSearch.*select:AskUserQuestion' CLAUDE.md` ≥ 1
    - [ ] 기존 7-step 본문 보존 (additive only)
    - [ ] AC-AUE-008-1 통과

### T-05: CLAUDE.md §8 User Interaction Architecture 보강

- **Phase**: B
- **Owner-role**: implementer
- **REQ**: REQ-AUE-001, REQ-AUE-002, REQ-AUE-008
- **AC**: AC-AUE-001, AC-AUE-002, AC-AUE-008-1
- **의존성**: T-03
- **Blocks**: T-12
- **설명**: §8 "AskUserQuestion is the ONLY User Question Channel" 단락 끝에 ToolSearch precondition 항목 추가 + askuser-protocol.md cross-reference. "Canonical reference: see `.claude/rules/moai/core/askuser-protocol.md` for full procedure."
- **DoD**:
    - [ ] `grep 'askuser-protocol.md' CLAUDE.md` ≥ 1
    - [ ] AC-AUE-001 통과
    - [ ] AC-AUE-008-1 통과

---

## Phase C: moai-constitution.md MoAI Orchestrator 단락 강화

### T-06: moai-constitution.md MoAI Orchestrator Rules 추가

- **Phase**: C
- **Owner-role**: implementer
- **REQ**: REQ-AUE-001, REQ-AUE-002, REQ-AUE-008
- **AC**: AC-AUE-008-2
- **의존성**: T-03
- **Blocks**: T-13 (template mirror)
- **설명**: `MoAI Orchestrator` 단락 Rules 리스트 끝에 다음 2개 항목 추가:
    - `- [HARD] AskUserQuestion is a deferred tool — invoke \`ToolSearch(query: "select:AskUserQuestion")\` immediately before each AskUserQuestion call`
    - `- Canonical reference: \`.claude/rules/moai/core/askuser-protocol.md\``
- **DoD**:
    - [ ] `grep 'ToolSearch' moai-constitution.md` ≥ 1
    - [ ] `grep 'askuser-protocol' moai-constitution.md` ≥ 1
    - [ ] AC-AUE-008-2 통과

---

## Phase D: agent-common-protocol.md User Interaction Boundary 양방향 codify

### T-07: agent-common-protocol.md Orchestrator 의무 + Blocker Report 형식 추가

- **Phase**: D
- **Owner-role**: implementer
- **REQ**: REQ-AUE-001, REQ-AUE-002, REQ-AUE-005, REQ-AUE-008
- **AC**: AC-AUE-005, AC-AUE-008-3
- **의존성**: T-03
- **Blocks**: T-14 (template mirror)
- **설명**: `User Interaction Boundary` 단락에 다음 추가:
    - 신규 단락 "Orchestrator obligations": "Orchestrator MUST preload AskUserQuestion via `ToolSearch(query: \"select:AskUserQuestion\")` before each call."
    - 신규 단락 "Blocker report format": "If subagent requires user input, return a structured report containing `## Missing Inputs` section that lists each missing parameter with type, expected values, and rationale."
    - 신규 단락 "Re-delegation procedure": "On receiving a blocker report, the orchestrator runs an AskUserQuestion round, then injects the user's responses into a fresh subagent prompt and re-delegates."
    - cross-reference: "see `.claude/rules/moai/core/askuser-protocol.md`"
- **DoD**:
    - [ ] `grep 'missing inputs' agent-common-protocol.md` ≥ 1 (case insensitive)
    - [ ] `grep -E 'orchestrator.*ToolSearch' agent-common-protocol.md` ≥ 1
    - [ ] `grep 'askuser-protocol' agent-common-protocol.md` ≥ 1
    - [ ] AC-AUE-005, AC-AUE-008-3 통과

---

## Phase E: output-styles/moai/moai.md §3 / §10 갱신

### T-08: output-styles §3 Stage 1 Clarify 행동 시퀀스에 ToolSearch step 삽입

- **Phase**: E
- **Owner-role**: implementer
- **REQ**: REQ-AUE-002, REQ-AUE-003
- **AC**: AC-AUE-002, AC-AUE-003, AC-AUE-008-4
- **의존성**: T-03
- **Blocks**: T-15 (template mirror)
- **설명**: §3 Stage 1 "Clarify" 행동 시퀀스에서 "Ask via AskUserQuestion" step 직전에 "First: ToolSearch(query: \"select:AskUserQuestion\")" step 삽입. Socratic interview 구조 제약(4Q/4O/`(권장)`)을 짧게 인용 + askuser-protocol.md §3 cross-reference.
- **DoD**:
    - [ ] `grep 'ToolSearch.*select' output-styles/moai/moai.md` ≥ 1
    - [ ] `grep -E '4 questions|4개 질문' output-styles/moai/moai.md` ≥ 1
    - [ ] AC-AUE-008-4 통과

### T-09: output-styles §10 Output Rules에 free-form anti-pattern 추가

- **Phase**: E
- **Owner-role**: implementer
- **REQ**: REQ-AUE-001, REQ-AUE-007
- **AC**: AC-AUE-001, AC-AUE-007, AC-AUE-008-4
- **의존성**: T-03
- **Blocks**: T-15
- **설명**: §10 Output Rules 끝에 신규 anti-pattern 항목: "Free-form interrogative prose in response body is **prohibited** as a question channel. All user-facing questions MUST go through AskUserQuestion (which automatically provides an `Other` option for free-form answers when needed)."
- **DoD**:
    - [ ] `grep -E 'free-form.*(prose|interrogative).*(prohibited|금지)' output-styles/moai/moai.md` ≥ 1
    - [ ] AC-AUE-007 통과

---

## Phase F: Template 미러 동기화 + make build

### T-10: Template 미러 5개 동기화

- **Phase**: F
- **Owner-role**: implementer
- **REQ**: REQ-AUE-009
- **AC**: AC-AUE-009-1
- **의존성**: T-04, T-05, T-06, T-07, T-08, T-09
- **Blocks**: T-11
- **설명**: 5개 핵심 파일 → `internal/template/templates/*` 미러로 byte-for-byte 동기화:
    1. `CLAUDE.md` → `internal/template/templates/CLAUDE.md`
    2. `.claude/rules/moai/core/moai-constitution.md` → `internal/template/templates/.claude/rules/moai/core/moai-constitution.md`
    3. `.claude/rules/moai/core/agent-common-protocol.md` → `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`
    4. `.claude/rules/moai/core/askuser-protocol.md` → `internal/template/templates/.claude/rules/moai/core/askuser-protocol.md` (신규 디렉토리 / 파일)
    5. `.claude/output-styles/moai/moai.md` → `internal/template/templates/.claude/output-styles/moai/moai.md`
- **DoD**:
    - [ ] 5개 `diff -q` 명령 모두 exit 0
    - [ ] AC-AUE-009-1 통과

### T-11: make build 실행 + embedded.go 검증

- **Phase**: F
- **Owner-role**: implementer
- **REQ**: REQ-AUE-009
- **AC**: AC-AUE-009-2, AC-AUE-009-3
- **의존성**: T-10
- **Blocks**: T-12
- **설명**: `make build` 실행 → `internal/template/embedded.go` 갱신 확인. `git diff --name-only` 출력에 embedded.go 포함 여부 검증. `go build ./...` 실행으로 회귀 부재 확인.
- **DoD**:
    - [ ] `make build` exit 0
    - [ ] `git diff --name-only internal/template/embedded.go` 출력에 해당 파일 포함
    - [ ] `go build ./...` exit 0
    - [ ] AC-AUE-009-2, AC-AUE-009-3 통과
    - [ ] TRUST 5 Tested: 빌드 회귀 부재 검증

---

## Phase G: Memory Dead Lesson SUPERSEDED 처리

### T-12: feedback_askuserquestion_enforcement.md SUPERSEDED 마커 추가

- **Phase**: G
- **Owner-role**: implementer
- **REQ**: (memory graduation)
- **AC**: AC-AUE-010
- **의존성**: T-03 (cross-reference 대상 존재 후)
- **Blocks**: T-13
- **설명**: `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_askuserquestion_enforcement.md` 본문 상단(frontmatter 직후)에 `> [SUPERSEDED by SPEC-ASKUSER-ENFORCE-001]` 줄 추가, 본문 끝에 `> Graduated to canonical guidance: see .claude/rules/moai/core/askuser-protocol.md` 줄 추가. 원본 본문은 변경 금지 — additive only (moai-constitution.md Lessons Protocol).
- **DoD**:
    - [ ] `grep '\[SUPERSEDED by SPEC-ASKUSER-ENFORCE-001\]' feedback_askuserquestion_enforcement.md` ≥ 1
    - [ ] `grep 'askuser-protocol.md' feedback_askuserquestion_enforcement.md` ≥ 1
    - [ ] 원본 lesson 본문 byte-for-byte 변경 부재 (diff 검증)
    - [ ] AC-AUE-010 통과

### T-13: MEMORY.md 인덱스 항목에 GRADUATED 표시 추가

- **Phase**: G
- **Owner-role**: implementer
- **REQ**: (memory graduation)
- **AC**: AC-AUE-011
- **의존성**: T-12
- **Blocks**: T-14 (integration verification)
- **설명**: `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md`의 "AskUserQuestion Enforcement" 항목 hook 텍스트 갱신:
    - 변경 전: `Deferred tool preload 의무 + 산문 질문 anti-pattern + 재발 방지 프로토콜 (v3.4.0 CLAUDE.local.md §19)`
    - 변경 후: `[GRADUATED to SPEC-ASKUSER-ENFORCE-001] Deferred tool preload + Socratic interview — see .claude/rules/moai/core/askuser-protocol.md`
- **DoD**:
    - [ ] `grep 'GRADUATED.*SPEC-ASKUSER-ENFORCE-001' MEMORY.md` ≥ 1
    - [ ] AC-AUE-011 통과

---

## Phase H: 통합 검증 + Lint + Self-Audit

### T-14: AC-AUE-012 통합 검증 시퀀스 실행

- **Phase**: H
- **Owner-role**: tester
- **REQ**: 모든 REQ
- **AC**: AC-AUE-012 (전체 SPEC end-to-end)
- **의존성**: T-13
- **Blocks**: T-15
- **설명**: acceptance.md Part E의 bash 검증 시퀀스를 실행하여 전체 정합성을 한 번에 확인. 어느 한 줄이라도 실패하면 재작업.
- **DoD**:
    - [ ] AC-AUE-012의 bash 시퀀스 전체 exit 0
    - [ ] `go build ./...` exit 0
    - [ ] `go vet ./...` exit 0
    - [ ] TRUST 5 Tested: 통합 검증 모두 통과

### T-15: golangci-lint 회귀 부재 검증

- **Phase**: H
- **Owner-role**: tester
- **REQ**: REQ-AUE-009
- **AC**: AC-AUE-009-2 (회귀 부재)
- **의존성**: T-14
- **Blocks**: T-16
- **설명**: `golangci-lint run ./...` 실행. 본 SPEC은 코드를 추가하지 않으므로 회귀 부재 확인이 목적.
- **DoD**:
    - [ ] `golangci-lint run ./...` exit 0
    - [ ] 신규 lint 경고 0건

### T-16: plan-auditor 8 criteria self-audit

- **Phase**: H
- **Owner-role**: reviewer
- **REQ**: (self-dogfood)
- **AC**: (self-audit)
- **의존성**: T-15
- **Blocks**: T-17
- **설명**: plan.md 하단 "Self-Audit Summary" 섹션 8개 criterion을 자가 검증. 각 criterion에 대한 PASS 근거가 명확한지 reviewer 관점에서 재확인. 결과를 plan.md에 commit한 상태 유지.
- **DoD**:
    - [ ] plan.md "Self-Audit Summary" 표 8행 모두 PASS
    - [ ] 각 행의 근거가 spec.md 또는 acceptance.md의 구체 섹션 / AC ID로 연결됨

### T-17: SPEC HISTORY 갱신 + status 변경 검토

- **Phase**: H
- **Owner-role**: implementer
- **REQ**: (SPEC lifecycle)
- **AC**: (lifecycle)
- **의존성**: T-16
- **Blocks**: T-18
- **설명**: spec.md / acceptance.md / plan.md / tasks.md 4개 파일의 frontmatter에서 status 검토. 모든 AC 통과 시점에 status를 `draft` → `approved` 후보로 표기 (실제 승인은 사용자 결정).
- **DoD**:
    - [ ] HISTORY 섹션에 본 SPEC의 v1.0.0 entry 존재 (spec.md §6)
    - [ ] 4개 파일 frontmatter의 `version` / `created` / `updated` 일관

### T-18: 최종 SPEC 디렉토리 정합성 검증

- **Phase**: H
- **Owner-role**: tester
- **REQ**: (SPEC structure)
- **AC**: (SPEC structure)
- **의존성**: T-17
- **Blocks**: 없음 (terminal task)
- **설명**: `.moai/specs/SPEC-ASKUSER-ENFORCE-001/` 디렉토리에 4개 파일 (spec.md, plan.md, acceptance.md, tasks.md)이 모두 존재하고, 각 파일의 frontmatter `id`가 `SPEC-ASKUSER-ENFORCE-001`로 일치하는지 확인.
- **DoD**:
    - [ ] `ls .moai/specs/SPEC-ASKUSER-ENFORCE-001/ | wc -l` = 4
    - [ ] `grep -h '^id:' .moai/specs/SPEC-ASKUSER-ENFORCE-001/*.md | sort -u | wc -l` = 1
    - [ ] TRUST 5 Trackable: 4 파일 일관 추적성

---

## Parallel Group 요약

본 SPEC은 단일 PR 내 직렬 작업이 적합하다 (모든 task가 5개 핵심 파일을 다루며 각 파일의 변경이 cross-reference로 묶여 있음). 그러나 다음 그룹은 안전하게 병렬화 가능하다:

| Group | Tasks | 사전 조건 | 비고 |
|-------|-------|-----------|------|
| G0 (직렬) | T-01, T-02, T-03 | 없음 | Phase A 단일 파일 작업 — 직렬 |
| G1 (병렬) | T-04, T-05, T-06, T-07, T-08, T-09 | T-03 완료 | Phase B–E 4개 파일 병렬 가능 (각기 다른 파일) |
| G2 (직렬) | T-10, T-11 | G1 완료 | template mirror → make build 직렬 |
| G3 (병렬) | T-12, T-13 | T-03 완료 | memory 영역 작업 — G1과 병렬 가능 |
| G4 (직렬) | T-14, T-15, T-16, T-17, T-18 | G2, G3 완료 | 통합 검증 직렬 |

병렬 실행 시: G1과 G3을 동시에 실행하여 wall-clock 시간 단축 가능. 단, 본 SPEC의 task 작업량이 작으므로 직렬 실행이 단순성 측면에서 더 권장된다.

---

## REQ → Task → AC 추적성 매트릭스

| REQ | Task | AC |
|-----|------|----|
| REQ-AUE-001 | T-01, T-05, T-09 | AC-AUE-001, AC-AUE-008-1, AC-AUE-008-4 |
| REQ-AUE-002 | T-02, T-04, T-05, T-06, T-08 | AC-AUE-002, AC-AUE-008-1, AC-AUE-008-2, AC-AUE-008-4, AC-AUE-008-5 |
| REQ-AUE-003 | T-03, T-08 | AC-AUE-003, AC-AUE-008-4, AC-AUE-008-5 |
| REQ-AUE-004 | T-03 | AC-AUE-004, AC-AUE-008-5 |
| REQ-AUE-005 | T-03, T-07 | AC-AUE-005, AC-AUE-008-3 |
| REQ-AUE-006 | T-03, T-04 | AC-AUE-006, AC-AUE-008-1 |
| REQ-AUE-007 | T-01, T-03, T-09 | AC-AUE-007, AC-AUE-008-4, AC-AUE-008-5 |
| REQ-AUE-008 | T-04, T-05, T-06, T-07, T-08, T-09 | AC-AUE-008-1 ~ AC-AUE-008-5 |
| REQ-AUE-009 | T-10, T-11 | AC-AUE-009-1, AC-AUE-009-2, AC-AUE-009-3 |
| (graduation) | T-12, T-13 | AC-AUE-010, AC-AUE-011 |
| (integration) | T-14, T-15, T-16, T-17, T-18 | AC-AUE-012 |

총 9 REQ + 2 보조 영역 → 18 task → 20 AC. 모든 REQ가 최소 1개 task에 매핑되며, 모든 task가 최소 1개 AC를 충족한다.

---

## TRUST 5 Gate Summary

| Pillar | 적용 Task | 검증 방법 |
|--------|-----------|-----------|
| Tested | T-11, T-14, T-15 | `go build`, `go vet`, `golangci-lint`, AC bash 시퀀스 |
| Readable | T-01 ~ T-09 | DoD에 한국어/영어 일관 어휘 명시, canonical reference 단일 출처 |
| Unified | T-10, T-11 | byte-for-byte mirror 동기화, embedded.go diff 검증 |
| Secured | (해당 없음) | 본 SPEC은 사용자 입력 / 인증 / 데이터 처리 변경 없음 |
| Trackable | T-12, T-13, T-17, T-18 | SPEC ID 단일 추적, memory graduation 마커, frontmatter 일관성 |
