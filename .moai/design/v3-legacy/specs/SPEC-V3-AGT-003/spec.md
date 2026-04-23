---
id: SPEC-V3-AGT-003
title: "Fork Subagent Primitive (simplified — inherit-prompt variant)"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P1 High
phase: "v3.0.0 — Phase 6a Tier 2 Strategic"
module: "internal/cli/deps.go, internal/core/agent/fork/"
dependencies:
  - SPEC-V3-AGT-001
  - SPEC-V3-AGT-002
related_gap:
  - gm#67
related_theme: "Theme 4 — Agent Frontmatter Expansion"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "agent, fork, subagent, inherit-prompt, v3"
---

# SPEC-V3-AGT-003: Fork Subagent Primitive (simplified — inherit-prompt variant)

## HISTORY

| Version | Date       | Author | Description                                                    |
|---------|------------|--------|----------------------------------------------------------------|
| 0.1.0   | 2026-04-22 | GOOS   | Initial draft (Wave 4); simplified scope for v3.0              |

---

## 1. Goal (목적)

CC의 fork subagent primitive(findings-wave1-agent-team.md §4.4, §8.1, forkSubagent.ts:60-71)를 v3.0에서 **simplified scope**로 moai에 도입한다. Fork subagent란 `Agent()` 호출 시 `subagent_type`을 생략하면 child가 parent의 system prompt를 상속하는 메커니즘이다. v3.0은 **inherit-prompt variant**만 구현하며, **cache-identical prefix sharing**(CC의 `buildForkedMessages`)은 v3.1로 연기된다(master §9 Open Question #5 Recommended default).

### 1.1 배경

Fork subagent는 CC에서 다음 세 가지 특성을 가진다(findings-wave1-agent-team.md §4.4, §5.2 step 5):

1. **System prompt inheritance**: child가 parent의 `getSystemPrompt()` 결과를 그대로 사용 (빌드 시간 절감)
2. **Conversation context inheritance**: child가 parent의 `promptMessages[]`를 상속 (task 전환 시간 절감)
3. **Cache-identical prefix** (`buildForkedMessages`): Anthropic API의 cache_control prefix가 parent와 child 사이 byte-identical → prompt cache 적중률 대폭 상승 (가장 큰 이득)

v3.0 scope 축소 근거(master §9 Open Question #5):
- Pros(simplified): 빠른 출시, recursion risk 낮음, 신규 인프라 최소
- Cons(simplified): cache-identical prefix 미구현으로 CC의 key win 미달성
- Recommended default: Simplified (inherit system prompt) in v3.0; full cache-identical prefix in v3.1

v3.0 초점은 (1) + (2)만 구현. (3)은 별도 SPEC-V3-AGT-004(v3.1 예정)로 분리.

### 1.2 Non-Goals

- **Cache-identical prefix sharing**: `buildForkedMessages` 포팅은 v3.1로 defer
- **Recursion > depth 2**: parent → fork → leaf는 허용, leaf가 또 fork하는 것은 금지
- **Remote fork**: `isolation: remote`(Anthropic CCR) 동작 포팅 불가(Non-Goal)
- **Fork with different tools pool**: child는 parent tools 상속만; 별도 tools 지정은 v3.1+
- **Parallel fork optimization**: 병렬 fork spawn은 일반 `Agent()` 병렬 spawn과 동일 동작

---

## 2. Scope (범위)

### 2.1 In Scope

- `internal/core/agent/fork/dispatcher.go`: `Agent()` 호출 시 `subagent_type`이 비어 있으면 fork 경로 분기
- `internal/core/agent/fork/inherit.go`: parent의 system prompt와 conversation history를 child에 복사
- `internal/core/agent/fork/depth_guard.go`: recursion depth cap 2 enforcement (parent → fork → leaf)
- `internal/cli/deps.go` 에이전트 spawn dispatcher 확장: fork path 라우팅
- `.claude/rules/moai/workflow/fork-subagent.md` 신설: fork 사용 가이드
- `moai doctor agent --validate`: fork depth 위반 탐지

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Cache-identical API prefix (v3.1 — 별도 SPEC)
- Fork with tool pool override
- Fork with different permissionMode (parent의 permissionMode가 그대로 상속되며, `bubble`로 강제되지 않음 — 단, bubble을 권장)
- Recursion depth > 2 (hard cap으로 거부)
- Remote fork (Anthropic CCR)
- Fork metrics dashboard
- Fork replay mechanism
- Parent abort의 fork child 연쇄 abort 신호 세밀 제어(기본 signal propagation만)

---

## 3. Environment (환경)

- 런타임: moai-adk-go v3.0.0+ (Go 1.23+)
- Claude Code v2.1.111+
- 의존: SPEC-V3-AGT-001(frontmatter), SPEC-V3-AGT-002(`bubble` mode 권장)
- 영향 디렉터리: `internal/core/agent/fork/`, `internal/cli/deps.go`, `.claude/rules/moai/workflow/`
- OS 동등성: macOS / Linux / Windows
- 참고: CC 구현 위치 — `tools/AgentTool/forkSubagent.ts` (findings-wave1-agent-team.md §4.4)

---

## 4. Assumptions (가정)

- Parent agent의 system prompt는 immutable이다(runtime 중 변경 없음).
- Parent conversation history는 fork 시점의 snapshot으로 복사되며, 이후 parent의 turn은 child에 전달되지 않는다(v3.0은 one-shot fork).
- Recursion depth 2는 대부분의 실사용 케이스에 충분하다(CC 동일 정책).
- Fork child는 별도 `agentId`를 가지며 parent와 별개로 스케줄링된다.
- `permissionMode: bubble`이 지정되지 않은 fork child는 parent의 permissionMode를 상속한다.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous (항시)

**REQ-AGT-003-001 (Ubiquitous) — fork 분기 조건**
The spawn dispatcher **shall** route an `Agent()` invocation to the fork path when the `subagent_type` parameter is omitted OR explicitly set to `"fork"`.

**REQ-AGT-003-002 (Ubiquitous) — system prompt 상속**
The fork path **shall** copy the parent agent's rendered system prompt (result of `getSystemPrompt()` equivalent) into the child agent's initial system prompt slot.

**REQ-AGT-003-003 (Ubiquitous) — conversation history 상속**
The fork path **shall** copy the parent agent's conversation history (`promptMessages[]` snapshot at fork time) into the child agent's initial conversation history.

**REQ-AGT-003-004 (Ubiquitous) — depth cap**
The fork path **shall** enforce a maximum recursion depth of 2: root (depth 0) → first fork (depth 1) → leaf (depth 2). A fourth-level fork attempt **shall** be rejected.

**REQ-AGT-003-005 (Ubiquitous) — child agentId 분리**
The fork path **shall** assign the child a new unique `agentId` distinct from the parent's `agentId`.

**REQ-AGT-003-006 (Ubiquitous) — cache-identical 미구현 명시**
The fork path **shall** document in `.claude/rules/moai/workflow/fork-subagent.md` that v3.0 does NOT provide cache-identical API prefix; child's first API call may re-process the inherited prompt.

### 5.2 Event-Driven (이벤트 기반)

**REQ-AGT-003-007 (Event-Driven) — fork spawn 로깅**
**When** the fork path is taken, the spawn dispatcher **shall** log the fork event to `.moai/logs/agent-fork.log` with fields `{parent_agent_id, child_agent_id, fork_depth, parent_type, timestamp}`.

**REQ-AGT-003-008 (Event-Driven) — depth 초과 rejection**
**When** a fork request would result in depth > 2, the spawn dispatcher **shall** reject the spawn with error `AGT_FORK_DEPTH_EXCEEDED` and include the current chain's `agentId` list in the error payload.

**REQ-AGT-003-009 (Event-Driven) — parent abort propagation**
**When** the parent agent is aborted (SIGTERM, user abort, or `TaskStop`), the fork child **shall** receive an abort signal and terminate within 5 seconds OR be forcibly killed.

### 5.3 State-Driven (상태 기반)

**REQ-AGT-003-010 (State-Driven) — inherited permissionMode**
**While** the fork child is executing AND `permissionMode: bubble` is not explicitly set on the child, the child **shall** use the parent's current `permissionMode` for its tool calls.

**REQ-AGT-003-011 (State-Driven) — snapshot isolation**
**While** the fork child is executing, subsequent changes to the parent's conversation history **shall not** propagate to the child.

### 5.4 Optional (선택)

**REQ-AGT-003-012 (Optional) — explicit fork type**
**Where** the user passes `subagent_type: "fork"` explicitly (instead of omitting), the spawn dispatcher **shall** produce identical behavior to the implicit omit path.

**REQ-AGT-003-013 (Optional) — fork with bubble permissionMode**
**Where** the fork child is declared with `permissionMode: bubble`, the permission forwarding logic (SPEC-V3-AGT-002 REQ-AGT-002-005) **shall** route permission prompts to the parent's parent (the root session), not to the immediate parent.

### 5.5 Unwanted Behavior

**REQ-AGT-003-014 (Unwanted Behavior) — nested remote fork**
**If** the fork path is invoked with `isolation: "remote"`, **then** the spawn dispatcher **shall** reject the spawn with error `AGT_FORK_REMOTE_UNSUPPORTED` and **shall not** silently fall back.

**REQ-AGT-003-015 (Unwanted Behavior) — parent 사라짐**
**If** the parent agent terminates before the fork child's system prompt is copied, **then** the spawn dispatcher **shall** abort the fork with error `AGT_FORK_PARENT_GONE` and clean up any partially-initialized child state.

**REQ-AGT-003-016 (Unwanted Behavior) — cache-identical 오해 방지**
**If** a user reports "fork not reducing token cost", **then** `moai doctor agent --validate` **shall** surface a diagnostic note explaining v3.0's inherit-prompt (not cache-identical) semantics and reference SPEC-V3-AGT-004 (planned v3.1) for cache-identical support.

---

## 6. Acceptance Criteria (수용 기준 요약)

상세 Given-When-Then 시나리오는 `acceptance.md`에서 정의한다.

핵심 기준:

- **AC-AGT-003-01**: `Agent()` 호출 시 `subagent_type` 생략 → fork path 진입 확인 (로그 검증)
- **AC-AGT-003-02**: Fork child의 초기 system prompt === parent의 system prompt (bit-identical)
- **AC-AGT-003-03**: Fork child의 conversation history === parent의 promptMessages[] snapshot
- **AC-AGT-003-04**: depth 3 fork 시도 → `AGT_FORK_DEPTH_EXCEEDED` 반환
- **AC-AGT-003-05**: Fork child의 agentId가 parent와 다름 (unique UUID)
- **AC-AGT-003-06**: Parent SIGTERM → child 5초 이내 종료 확인
- **AC-AGT-003-07**: Fork 이후 parent의 새 turn이 child에 전달되지 않음 (snapshot isolation)
- **AC-AGT-003-08**: `isolation: "remote"` + fork → `AGT_FORK_REMOTE_UNSUPPORTED` 반환
- **AC-AGT-003-09**: `go test ./internal/core/agent/fork/...` 전체 통과 + coverage ≥ 85%

---

## 7. Constraints (제약)

- [HARD] Depth cap 2는 v3.0에서 고정. 완화 불가.
- [HARD] Cache-identical prefix 구현 금지(v3.1 scope; 명확한 분리로 v3.0 risk 관리).
- [HARD] Remote fork 금지(Non-Goal; Anthropic 내부 CCR 의존).
- [HARD] Fork child는 parent와 동일 프로세스 내에서 실행(별도 subprocess/worktree spawn은 일반 `Agent()` 경로로 처리).
- [HARD] Snapshot isolation: fork 후 parent 변경이 child로 전파되지 않음(one-shot fork).
- [HARD] `AGT_FORK_*` 에러 코드는 `internal/core/agent/fork/errors.go`에 const로 정의.
- [HARD] 16개 언어 중립성 유지: fork mechanism은 사용자 프로젝트 언어에 무관.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크                                                              | 영향     | 완화                                                                                                     |
|---------------------------------------------------------------------|----------|----------------------------------------------------------------------------------------------------------|
| Fork recursion으로 모든 depth capped 해제 시도                       | High     | HARD depth cap 2; `moai doctor agent --validate` 탐지; test coverage 필수 (Master §7 R-006)              |
| Cache-identical 미구현으로 token cost 기대치 미달                    | Medium   | REQ-AGT-003-006, 016에 명시; SPEC-V3-AGT-004(v3.1) 링크; CHANGELOG에 limitation 기록                     |
| Parent 종료 후 child orphan 생성                                     | Medium   | REQ-AGT-003-009 signal propagation; 5초 forcible kill                                                    |
| Snapshot timing race (parent가 fork 직전에 turn 추가)                | Low      | atomic snapshot via `sync.Mutex`; fork 시점의 `promptMessages[]` deep copy                               |
| Fork 이후 parent가 동일 tool 호출로 dup side effect                  | Low      | fork child와 parent는 독립적 tool context; 동일 file 편집은 worktree isolation 권장(별도 SPEC)            |
| `bubble` 모드 부재 시 permission prompt 사용자 혼란                  | Low      | 문서에 `bubble` 사용 강력 권장; `moai doctor`로 경고                                                     |
| `subagent_type: "fork"` 와 omit path 간 semantic drift               | Low      | REQ-AGT-003-012 동일 경로 강제; 단일 구현 함수                                                            |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3-AGT-001** (Agent frontmatter v2 bundle): `permissionMode`, `isolation` 필드가 먼저 정의되어야 함.
- **SPEC-V3-AGT-002** (permissionMode extensions): `bubble` mode가 먼저 구현되어야 fork + bubble 조합이 가능.

### 9.2 Blocks

- (v3.1 예정) SPEC-V3-AGT-004 — Cache-identical fork prefix.

### 9.3 Related

- **SPEC-V3-TEAM-001** (Team mailbox): fork는 team mode와 무관하지만, team mode에서 fork 사용 시 mailbox 동작은 일반 agent와 동일.
- **SPEC-HOOKS-008** (sibling writer; Hook type:agent): hook type:agent의 agent spawn 경로에서도 fork가 가능해야 하지만, 본 SPEC v3.0에서는 명시적 call path만 지원(hook level 자동 fork는 Non-Goal).

---

## 10. Traceability (추적성)

- 총 REQ 개수: 16 (Ubiquitous 6, Event-Driven 3, State-Driven 2, Optional 2, Unwanted Behavior 3)
- 예상 AC 개수: 9
- 관련 Wave 1 근거:
  - findings-wave1-agent-team.md §4.4 (fork agent built-in; forkSubagent.ts:60-71)
  - findings-wave1-agent-team.md §5.2 step 5 (dispatch tree — "Resolve subagent_type (or FORK_AGENT if fork gate is on + subagent_type omitted)")
  - findings-wave1-agent-team.md §8.1 (Gap: "`fork` subagent primitive — Omit subagent_type → child inherits parent's full conversation context + system prompt (byte-identical for cache sharing). moai has nothing equivalent")
  - master-v3 §9 Open Question #5 (Recommended default: simplified variant in v3.0)
- 구현 시 각 소스 파일에 `@SPEC:SPEC-V3-AGT-003:REQ-AGT-003-NNN` 주석 부착
- 코드 구현 예상 경로:
  - `internal/core/agent/fork/dispatcher.go` (REQ-AGT-003-001, 002, 003, 005, 012)
  - `internal/core/agent/fork/inherit.go` (REQ-AGT-003-002, 003, 011)
  - `internal/core/agent/fork/depth_guard.go` (REQ-AGT-003-004, 008)
  - `internal/core/agent/fork/abort.go` (REQ-AGT-003-009, 015)
  - `internal/core/agent/fork/errors.go` (AGT_FORK_* 상수)
  - `internal/core/agent/fork/dispatcher_test.go` (AC-AGT-003-01..08)
  - `.claude/rules/moai/workflow/fork-subagent.md` (REQ-AGT-003-006 문서)

---

End of SPEC.
