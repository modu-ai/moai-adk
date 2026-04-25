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

# SPEC-ASKUSER-ENFORCE-001 Acceptance Criteria

> Given-When-Then 인수 조건 + 추적성 + 검증 증빙 매핑
> 최종 갱신: 2026-04-25
> AC 개수: **20** (REQ-AUE-001~009 1차 매핑 9 + 파일별 grep evidence 5 + 빌드 검증 3 + memory graduation 2 + integration 1)
> 출처: SPEC §3 EARS 요구사항 + acceptance 단계에서 파생된 file-level / build-level 증빙

---

## 읽는 방법

- 각 AC는 고유 ID(`AC-AUE-NN`)와 REQ 추적성을 명시한다.
- "검증 방법"은 자동화 가능한 명령(grep/regex/build) 또는 수동 검토 절차를 명시한다.
- "Evidence"는 검증 결과를 어디에서 확인할 수 있는지를 명시한다 — 파일 경로, 라인 범위, 빌드 산출물.
- 모든 AC는 Given-When-Then-Evidence 형식을 따른다.

---

## Part A: REQ → AC 1차 매핑

### AC-AUE-001 — AskUserQuestion-only 채널 명문화 (REQ-AUE-001)

- **Given**: orchestrator가 사용자에게 의사결정을 요청해야 하는 상황 (Stage 1 Clarify trigger 충족 또는 분기 결정).
- **When**: 5개 핵심 지침 파일 중 어느 것이라도 다시 로드되어 orchestrator가 규범을 참조한다.
- **Then**: 모든 5개 파일이 "AskUserQuestion is the only user-facing question channel" 또는 동등 의미 문장을 명시적으로 포함한다. 자유 서술 의문문 사용은 anti-pattern으로 명시된다.
- **Evidence**:
    - 파일: `CLAUDE.md` §8, `.claude/rules/moai/core/moai-constitution.md` `MoAI Orchestrator` 단락, `.claude/rules/moai/core/agent-common-protocol.md` `User Interaction Boundary` 단락, `.claude/output-styles/moai/moai.md` §3 / §10, `.claude/rules/moai/core/askuser-protocol.md` 본문
    - 검증: `grep -rE "AskUserQuestion.*(only|단 하나|유일).*question" CLAUDE.md .claude/rules/moai/core/ .claude/output-styles/moai/` ≥ 5 hits

### AC-AUE-002 — ToolSearch 사전 로드 절차 명문화 (REQ-AUE-002)

- **Given**: orchestrator가 `AskUserQuestion` 호출을 결정한 시점.
- **When**: orchestrator가 핵심 지침 파일을 참조한다.
- **Then**: 5개 파일 중 최소 3개(askuser-protocol.md, output-styles, CLAUDE.md §8)가 `ToolSearch(query: "select:AskUserQuestion")` 사전 호출을 호출 직전 의무로 명시한다. 일반화 표현("deferred tool requires ToolSearch select preload")이 적어도 1회 등장한다.
- **Evidence**:
    - 검증: `grep -rE 'ToolSearch.*select:.*AskUserQuestion' CLAUDE.md .claude/rules/moai/core/ .claude/output-styles/moai/` ≥ 3 hits
    - 일반화 검증: `grep -rE 'deferred tool.*(ToolSearch|select|preload)' .claude/rules/moai/core/askuser-protocol.md` ≥ 1 hit

### AC-AUE-003 — Socratic Interview 구조 제약 명문화 (REQ-AUE-003)

- **Given**: Stage 1 Clarify trigger 충족 상태.
- **When**: orchestrator가 Socratic interview 라운드를 구성한다.
- **Then**: 신규 `askuser-protocol.md`와 `output-styles/moai/moai.md` §3 Stage 1이 다음 4개 제약을 모두 명시한다 — (a) 라운드당 최대 4개 질문, (b) 질문당 최대 4개 옵션, (c) 첫 옵션 라벨에 `(권장)` 또는 `(Recommended)` 접미, (d) 모든 텍스트는 사용자의 `conversation_language`로 작성.
- **Evidence**:
    - 검증: `grep -E '(4개 질문|≤ ?4 questions|max 4 questions)' .claude/rules/moai/core/askuser-protocol.md .claude/output-styles/moai/moai.md` ≥ 2 hits
    - 검증: `grep -E '(권장|Recommended)' .claude/rules/moai/core/askuser-protocol.md .claude/output-styles/moai/moai.md` ≥ 2 hits
    - 검증: `grep -E 'conversation_language' .claude/rules/moai/core/askuser-protocol.md` ≥ 1 hit

### AC-AUE-004 — Option description 의무 명문화 (REQ-AUE-004)

- **Given**: AskUserQuestion 라운드에서 옵션을 구성하는 시점.
- **When**: orchestrator가 옵션 description을 작성한다.
- **Then**: `askuser-protocol.md`가 description 필수 포함 항목을 명시한다 — (a) 즉각적 결과, (b) 부수 효과/리스크, (c) 가능한 경우 정량 정보. bias prevention 항목으로 "추천/폄하 어조 금지, 추천 신호는 라벨 접미만"이 명시된다.
- **Evidence**:
    - 검증: `grep -E '(description|설명).*(implication|trade-off|함의|결과|리스크)' .claude/rules/moai/core/askuser-protocol.md` ≥ 1 hit
    - 검증: `grep -E '(bias|편향).*prevention' .claude/rules/moai/core/askuser-protocol.md` ≥ 1 hit

### AC-AUE-005 — Subagent 시도 거부 + Blocker Report 변환 (REQ-AUE-005)

- **Given**: subagent가 사용자 질문이 필요한 상황에 도달.
- **When**: subagent가 prompt body를 구성한다.
- **Then**: `agent-common-protocol.md`의 `User Interaction Boundary` 단락이 양방향(orchestrator 의무 + subagent 금지)으로 codify되어 있고, subagent의 올바른 응답 형식("missing inputs" 섹션을 포함한 blocker report)이 명시된다. orchestrator의 후속 행동(blocker 수신 → AskUserQuestion 라운드 → 결과를 subagent prompt에 주입하여 재위임)도 명시된다.
- **Evidence**:
    - 검증: `grep -E '(blocker|missing inputs).*(report|section)' .claude/rules/moai/core/agent-common-protocol.md` ≥ 1 hit
    - 검증: `grep -E 'orchestrator.*(re-?delegate|재위임|prompt.*주입)' .claude/rules/moai/core/agent-common-protocol.md .claude/rules/moai/core/askuser-protocol.md` ≥ 1 hit

### AC-AUE-006 — Ambiguity Trigger 4종 단일 출처 정합 (REQ-AUE-006)

- **Given**: orchestrator가 ambiguity trigger를 평가한다.
- **When**: `CLAUDE.md` §7 Rule 5와 §8 `Ambiguity Triggers` 두 섹션이 trigger 정의를 표현한다.
- **Then**: 두 섹션이 동일 4종(pronoun without referent, multi-interpretable verb, unclear boundary, conflict with existing state)을 동일 어휘로 표현하며, 동일 면제 조건 5종(single-line typo, reproduction이 제공된 bug fix, 경로가 지정된 file read, 모든 인자가 제공된 command, 동일 세션 연속 작업)을 명시한다.
- **Evidence**:
    - 검증: `CLAUDE.md` §7 Rule 5와 §8 두 섹션을 비교한 diff(텍스트 대조)에서 trigger 4종이 의미적으로 동일함을 확인 (수동 review)
    - 자동 검증: 4 trigger 키워드가 두 섹션에 모두 등장하는지 grep — `grep -c 'pronoun\|multi-interpretable\|unclear boundar\|conflict.*state' CLAUDE.md` ≥ 8 hits (4 trigger × 2 섹션)

### AC-AUE-007 — Free-form 우회 금지 + Other 옵션 안내 (REQ-AUE-007)

- **Given**: 사용자가 자유 서술 답변을 요구하거나 그렇게 보이는 상황.
- **When**: orchestrator가 행동을 결정한다.
- **Then**: `askuser-protocol.md`가 (a) 자유 서술 의문문 우회 금지, (b) AskUserQuestion이 자동 "Other" 옵션을 제공한다는 사실, (c) "Other" 선택 후 자유 서술 제출이 가능하다는 절차를 명시한다.
- **Evidence**:
    - 검증: `grep -E '(자유 서술|free-form|freeform).*(금지|prohibited|circumvent)' .claude/rules/moai/core/askuser-protocol.md` ≥ 1 hit
    - 검증: `grep -E '"Other"|Other 옵션' .claude/rules/moai/core/askuser-protocol.md` ≥ 1 hit

### AC-AUE-008 — 5 핵심 지침 파일 동시 명문화 (REQ-AUE-008)

- **Given**: 5개 핵심 지침 파일 — CLAUDE.md, moai-constitution.md, agent-common-protocol.md, output-styles/moai/moai.md, askuser-protocol.md(신규).
- **When**: orchestrator가 어느 한 파일에서 AskUserQuestion 규범을 찾는다.
- **Then**: 5개 파일 모두 동일 사실에 대해 동일 어휘로 작성되어 있고, askuser-protocol.md는 canonical reference로서 다른 4개 파일에서 cross-reference된다.
- **Evidence**: 별도 sub-AC AC-AUE-008-1 ~ AC-AUE-008-5 (Part B) 참조

### AC-AUE-009 — Template 미러 동기화 + 빌드 산출물 검증 (REQ-AUE-009)

- **Given**: 본 SPEC PR 직전 상태.
- **When**: 5 핵심 파일 중 하나라도 수정되어 PR이 구성된다.
- **Then**: `internal/template/templates/*` 미러 5개가 동시에 동기화되며, `make build` 실행 후 `internal/template/embedded.go`의 diff가 위 5개 파일 변경분만을 반영한다.
- **Evidence**: 별도 sub-AC AC-AUE-009-1 ~ AC-AUE-009-3 (Part C) 참조

---

## Part B: 파일별 Grep Evidence (REQ-AUE-008 sub-AC)

### AC-AUE-008-1 — CLAUDE.md 명문화

- **Given**: 본 SPEC 적용 후 `CLAUDE.md` 상태.
- **When**: §7 Rule 5 또는 §8 섹션을 검색한다.
- **Then**: 다음 마커 문자열이 모두 등장한다:
    - `AskUserQuestion is the only user-facing question channel` (또는 한국어 동등 표현)
    - `ToolSearch(query: "select:AskUserQuestion")`
    - `Stage 1 Clarify` 또는 `Socratic interview`
- **Evidence**:
    - 검증: `grep -c 'AskUserQuestion' CLAUDE.md` ≥ 5
    - 검증: `grep 'ToolSearch.*select:AskUserQuestion' CLAUDE.md` ≥ 1
    - 검증: `grep -E 'Socratic interview|Stage 1 Clarify' CLAUDE.md` ≥ 1

### AC-AUE-008-2 — moai-constitution.md 명문화

- **Given**: 본 SPEC 적용 후 `.claude/rules/moai/core/moai-constitution.md` 상태.
- **When**: `MoAI Orchestrator` 단락을 검색한다.
- **Then**: 다음 마커가 등장한다:
    - "AskUserQuestion is used ONLY by MoAI orchestrator" (기존 문장 유지) + ToolSearch preload 문장 신규 추가
    - askuser-protocol.md로의 cross-reference (예: "see `.claude/rules/moai/core/askuser-protocol.md`")
- **Evidence**:
    - 검증: `grep 'AskUserQuestion' .claude/rules/moai/core/moai-constitution.md` ≥ 2
    - 검증: `grep 'ToolSearch' .claude/rules/moai/core/moai-constitution.md` ≥ 1
    - 검증: `grep 'askuser-protocol' .claude/rules/moai/core/moai-constitution.md` ≥ 1

### AC-AUE-008-3 — agent-common-protocol.md 명문화

- **Given**: 본 SPEC 적용 후 `.claude/rules/moai/core/agent-common-protocol.md` 상태.
- **When**: `User Interaction Boundary` 단락을 검색한다.
- **Then**: 다음 마커가 등장한다:
    - 기존 "Subagents MUST NOT prompt the user" 유지
    - 신규: "Orchestrator MUST preload AskUserQuestion via ToolSearch before invoking" (또는 동등 한국어 표현)
    - 신규: blocker report 형식 ("missing inputs" 섹션) 명시
    - askuser-protocol.md로의 cross-reference
- **Evidence**:
    - 검증: `grep -c 'Subagents MUST NOT prompt the user' .claude/rules/moai/core/agent-common-protocol.md` ≥ 1
    - 검증: `grep -E 'orchestrator.*(MUST|반드시).*ToolSearch' .claude/rules/moai/core/agent-common-protocol.md` ≥ 1
    - 검증: `grep 'missing inputs' .claude/rules/moai/core/agent-common-protocol.md` ≥ 1
    - 검증: `grep 'askuser-protocol' .claude/rules/moai/core/agent-common-protocol.md` ≥ 1

### AC-AUE-008-4 — output-styles/moai/moai.md 명문화

- **Given**: 본 SPEC 적용 후 `.claude/output-styles/moai/moai.md` 상태.
- **When**: §3 Stage 1 또는 §10 Output Rules를 검색한다.
- **Then**: §3 Stage 1의 행동 시퀀스에 "Step 1: ToolSearch select → Step 2: AskUserQuestion 구성"이 명시되며, §10 Output Rules에 "free-form interrogative prose 금지"가 anti-pattern으로 명시된다.
- **Evidence**:
    - 검증: `grep -c 'AskUserQuestion' .claude/output-styles/moai/moai.md` ≥ 3
    - 검증: `grep 'ToolSearch.*select' .claude/output-styles/moai/moai.md` ≥ 1
    - 검증: `grep -E 'free-form.*(prose|interrogative).*(금지|prohibited|anti-pattern)' .claude/output-styles/moai/moai.md` ≥ 1

### AC-AUE-008-5 — askuser-protocol.md 신규 canonical reference 존재

- **Given**: 본 SPEC 적용 후 `.claude/rules/moai/core/askuser-protocol.md` 파일이 존재.
- **When**: 파일 본문을 검색한다.
- **Then**: 다음 섹션이 모두 존재한다:
    - `## Channel Monopoly` (또는 동등) — REQ-AUE-001 본문
    - `## ToolSearch Preload Procedure` — REQ-AUE-002 본문
    - `## Socratic Interview Structure` — REQ-AUE-003 본문
    - `## Option Description Standards` — REQ-AUE-004 본문
    - `## Orchestrator–Subagent Boundary` — REQ-AUE-005 본문
    - `## Ambiguity Triggers and Exceptions` — REQ-AUE-006 본문
    - `## Free-form Circumvention Prohibition` — REQ-AUE-007 본문
- **Evidence**:
    - 검증: `test -f .claude/rules/moai/core/askuser-protocol.md && wc -l .claude/rules/moai/core/askuser-protocol.md` ≥ 80 lines
    - 검증: 7개 섹션 헤더 grep — `grep -cE '^## (Channel Monopoly|ToolSearch Preload|Socratic Interview|Option Description|Orchestrator.*Subagent|Ambiguity Triggers|Free-form)' .claude/rules/moai/core/askuser-protocol.md` ≥ 7

---

## Part C: 빌드 산출물 검증 (REQ-AUE-009 sub-AC)

### AC-AUE-009-1 — Template 미러 5개 동기화

- **Given**: 5 핵심 파일 변경분.
- **When**: PR diff를 검사한다.
- **Then**: `internal/template/templates/` 하위 미러 5개가 동시에 변경되었으며, 미러 본문은 원본과 byte-for-byte 일치한다.
- **Evidence**:
    - 검증: `diff CLAUDE.md internal/template/templates/CLAUDE.md` (exit 0)
    - 검증: `diff .claude/rules/moai/core/moai-constitution.md internal/template/templates/.claude/rules/moai/core/moai-constitution.md` (exit 0)
    - 검증: `diff .claude/rules/moai/core/agent-common-protocol.md internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` (exit 0)
    - 검증: `diff .claude/rules/moai/core/askuser-protocol.md internal/template/templates/.claude/rules/moai/core/askuser-protocol.md` (exit 0)
    - 검증: `diff .claude/output-styles/moai/moai.md internal/template/templates/.claude/output-styles/moai/moai.md` (exit 0)

### AC-AUE-009-2 — make build 후 embedded.go 갱신

- **Given**: 5개 미러 파일 변경 적용 상태.
- **When**: `make build` 실행.
- **Then**: 빌드는 zero error로 완료되며, `internal/template/embedded.go`가 변경된다.
- **Evidence**:
    - 검증: `make build` exit 0
    - 검증: `git diff --name-only internal/template/embedded.go` 출력에 해당 파일 포함
    - 검증: `go build ./...` exit 0 (회귀 부재)

### AC-AUE-009-3 — embedded.go diff가 5 파일에 한정

- **Given**: `make build` 직후.
- **When**: PR diff를 검사한다.
- **Then**: `embedded.go`의 변경분이 5개 미러 파일의 변경분만을 반영하며, 다른 템플릿 파일의 임베디드 표현이 변경되지 않는다.
- **Evidence**:
    - 검증: `git diff internal/template/embedded.go | grep -E '^[+-].*templates/(CLAUDE\.md|\.claude/rules/moai/core/(moai-constitution|agent-common-protocol|askuser-protocol)\.md|\.claude/output-styles/moai/moai\.md)'` ≥ 1 hit
    - 검증: 다른 임베디드 경로의 diff 부재 — 수동 확인

---

## Part D: Memory Graduation 검증

### AC-AUE-010 — Memory dead lesson SUPERSEDED 처리

- **Given**: `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_askuserquestion_enforcement.md` 파일 존재.
- **When**: 본 SPEC 적용 후 파일을 읽는다.
- **Then**: 파일 본문 상단에 `[SUPERSEDED by SPEC-ASKUSER-ENFORCE-001]` 마커가 추가되어 있고, body 끝에 "Graduated to canonical guidance: see `.claude/rules/moai/core/askuser-protocol.md`" cross-reference가 추가된다. 원본 lesson 내용은 변경되지 않는다 (additive only — moai-constitution.md Lessons Protocol 준수).
- **Evidence**:
    - 검증: `grep '\[SUPERSEDED by SPEC-ASKUSER-ENFORCE-001\]' ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_askuserquestion_enforcement.md` ≥ 1
    - 검증: `grep 'askuser-protocol.md' ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_askuserquestion_enforcement.md` ≥ 1

### AC-AUE-011 — Memory MEMORY.md 인덱스 갱신

- **Given**: `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md` 파일 존재.
- **When**: 본 SPEC 적용 후 인덱스를 읽는다.
- **Then**: AskUserQuestion Enforcement 항목에 `[GRADUATED to SPEC-ASKUSER-ENFORCE-001]` 표시가 추가되며, hook 텍스트가 "v3.4.0 CLAUDE.local.md §19" 대신 canonical reference 경로로 갱신된다.
- **Evidence**:
    - 검증: `grep 'GRADUATED.*SPEC-ASKUSER-ENFORCE-001' ~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md` ≥ 1

---

## Part E: 통합 검증

### AC-AUE-012 — End-to-End 정합성 단일 명령 검증

- **Given**: 본 SPEC PR 머지 직전 상태.
- **When**: 다음 통합 검증 스크립트(임시 명령 묶음)를 실행한다 — 별도 신규 코드 없이 bash one-liner로 표현 가능:
    ```bash
    set -euo pipefail
    test -f .claude/rules/moai/core/askuser-protocol.md
    grep -q 'ToolSearch.*select:AskUserQuestion' CLAUDE.md
    grep -q 'ToolSearch' .claude/rules/moai/core/moai-constitution.md
    grep -q 'orchestrator.*ToolSearch' .claude/rules/moai/core/agent-common-protocol.md
    grep -q 'ToolSearch.*select' .claude/output-styles/moai/moai.md
    diff -q CLAUDE.md internal/template/templates/CLAUDE.md
    diff -q .claude/rules/moai/core/askuser-protocol.md \
            internal/template/templates/.claude/rules/moai/core/askuser-protocol.md
    ```
- **Then**: 모든 명령이 exit 0으로 종료된다.
- **Evidence**:
    - 검증: 위 스크립트 전체 실행 exit 0
    - 검증: 본 명령 묶음의 어느 한 줄이라도 실패하면 본 SPEC PR은 acceptance 미충족으로 간주된다.

---

## REQ → AC 추적성 매트릭스

| REQ ID | AC IDs |
|--------|--------|
| REQ-AUE-001 | AC-AUE-001, AC-AUE-008-1 ~ AC-AUE-008-5 |
| REQ-AUE-002 | AC-AUE-002, AC-AUE-008-1, AC-AUE-008-2, AC-AUE-008-3, AC-AUE-008-4, AC-AUE-008-5 |
| REQ-AUE-003 | AC-AUE-003, AC-AUE-008-4, AC-AUE-008-5 |
| REQ-AUE-004 | AC-AUE-004, AC-AUE-008-5 |
| REQ-AUE-005 | AC-AUE-005, AC-AUE-008-3 |
| REQ-AUE-006 | AC-AUE-006, AC-AUE-008-1 |
| REQ-AUE-007 | AC-AUE-007, AC-AUE-008-5 |
| REQ-AUE-008 | AC-AUE-008-1 ~ AC-AUE-008-5 |
| REQ-AUE-009 | AC-AUE-009-1, AC-AUE-009-2, AC-AUE-009-3 |
| (graduation) | AC-AUE-010, AC-AUE-011 |
| (integration) | AC-AUE-012 |

총 9 REQ → 20 AC. 모든 REQ가 최소 1개 AC로 매핑되며, AC-AUE-012는 전체 SPEC의 end-to-end 검증을 담당한다.
