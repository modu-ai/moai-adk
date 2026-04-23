---
id: SPEC-V3-AGT-002
title: "Agent permissionMode Extensions (bubble, auto)"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P1 High
phase: "v3.0.0 — Phase 3 Agent Runtime v2"
module: "internal/config/schema/, internal/core/agent/permission/"
dependencies:
  - SPEC-V3-AGT-001
  - SPEC-V3-SCH-001
related_gap:
  - gm#64
  - gm#65
related_theme: "Theme 4 — Agent Frontmatter Expansion"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "agent, permission, bubble, auto, fork, v3"
---

# SPEC-V3-AGT-002: Agent permissionMode Extensions (bubble, auto)

## HISTORY

| Version | Date       | Author | Description                          |
|---------|------------|--------|--------------------------------------|
| 0.1.0   | 2026-04-22 | GOOS   | Initial draft (Wave 4 SPEC writer)   |

---

## 1. Goal (목적)

moai의 agent `permissionMode` enum을 Claude Code와 동등하게 확장하여 `bubble`과 `auto` 두 모드를 추가한다. `bubble`은 SPEC-V3-AGT-003의 fork subagent가 parent로 permission prompt를 surface하는 데 필요하고, `auto`는 classifier-based auto-approval 경로를 제공한다. 현 moai는 `default`, `plan`, `acceptEdits`, `bypassPermissions` 4개만 지원하며(findings-wave1-moai-current.md §6.3), CC는 총 6개를 지원한다(findings-wave1-agent-team.md §4.3).

### 1.1 배경

CC의 `PERMISSION_MODES`(findings-wave1-agent-team.md §4.3, PermissionMode.ts)는 6개 모드를 정의한다:

| Mode                | 동작                                                  |
|---------------------|-------------------------------------------------------|
| `default`           | 비-auto-approved tool call마다 permission 요청      |
| `acceptEdits`       | 파일 편집 자동 승인                                   |
| `bypassPermissions` | 모든 permission prompt 스킵 (위험)                    |
| `plan`              | Plan mode — 쓰기 전 `ExitPlanMode` 필수               |
| `auto`              | Classifier 기반 auto-approval (CC는 `tengu_auto_mode` gated) |
| `bubble`            | Permission prompt를 parent로 surface (fork subagent 용) |

moai는 현재 앞의 4개만 지원. `bubble`은 SPEC-V3-AGT-003(Fork subagent)의 핵심 전제이며, `auto`는 `moai run` 자동화 시 유용하다. 본 SPEC은 두 모드를 안전하게 추가하되, `auto`는 experimental flag 뒤에 gate한다(CC와 동일 정책).

### 1.2 Non-Goals

- Classifier 모델 훈련 및 배포 (`auto` 모드의 classifier는 v3.0에서 simple rule-based prototype, Sonnet-class LLM 호출은 v3.1 이후)
- `auto` 모드의 기본 활성화 (experimental; feature flag 필요)
- 기존 4개 모드 동작 변경
- `bubble` 모드 없이 fork subagent 구현 (SPEC-V3-AGT-003의 책임)

---

## 2. Scope (범위)

### 2.1 In Scope

- `internal/config/schema/agent.go`의 `PermissionMode` validator enum 확장: `{default, plan, acceptEdits, bypassPermissions, bubble, auto}`
- `internal/core/agent/permission/bubble.go`: `bubble` 모드 구현 — permission prompt를 parent session의 `AskUserQuestion` 채널로 forwarding
- `internal/core/agent/permission/auto.go`: `auto` 모드 prototype — rule-based classifier (safe/dangerous/ambiguous 3-tier)
- `.moai/config/sections/quality.yaml`의 `permission.auto_mode.enabled` feature flag 정의 (default `false`)
- `moai doctor agent --validate`에 permissionMode 유효성 검사 추가
- 기존 `.claude/agents/*.md` 22개의 permissionMode 필드가 v3 enum에서 0 error로 검증되는지 확인

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- LLM-based classifier for `auto` mode (v3.1+로 deferred; master §3.4 Recommended default for auto)
- `bubble` 모드의 UI 변형(permission prompt 형식은 기존 `AskUserQuestion` 준수)
- `auto` 모드의 대규모 텔레메트리 수집
- classifier 학습 데이터 수집 프레임워크
- `permissionMode` inheritance across fork chain (SPEC-V3-AGT-003 범위)
- `bypassPermissions` 모드 강화 또는 약화
- 사용자 커스텀 permissionMode 플러그인(플러그인 시스템은 SPEC-V3-PLG-001 범위)

---

## 3. Environment (환경)

- 런타임: moai-adk-go v3.0.0+ (Go 1.23+)
- Claude Code v2.1.111+
- 의존: SPEC-V3-AGT-001(프론트매터 확장), SPEC-V3-SCH-001(validator/v10)
- 영향 디렉터리: `internal/config/schema/`, `internal/core/agent/permission/`, `.claude/agents/`
- 영향 에이전트 수: 0 초기(opt-in). fork subagent primitive 도입(SPEC-V3-AGT-003) 이후 1개 이상.
- OS 동등성: macOS / Linux / Windows

---

## 4. Assumptions (가정)

- SPEC-V3-AGT-001이 먼저 머지되어 `PermissionMode` enum이 기존 4개 값으로 정의되어 있다.
- `auto` 모드 prototype의 rule-based classifier는 Wave 4 구현자가 정의하는 safelist(`Read`, `Grep`, `Glob`, `Bash(go test)`)와 denylist(`Bash(rm *)`, `Bash(git push --force)`, network-touching tools)로 충분하다.
- `bubble` 모드는 기존 `AskUserQuestion` 채널을 재사용하며 별도 transport 구축은 불필요하다.
- fork subagent(SPEC-V3-AGT-003)가 없는 v3.0 초기에는 `bubble` 모드 사용 에이전트가 0개이므로 regression risk는 무시할 만하다.
- `auto` 모드는 v3.0에서 experimental이며 default disabled다(CC의 `tengu_auto_mode` gate와 동일 정책).

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous (항시)

**REQ-AGT-002-001 (Ubiquitous) — enum 확장**
The `PermissionMode` validator tag in `internal/config/schema/agent.go` **shall** be updated to `oneof=default plan acceptEdits bypassPermissions bubble auto`.

**REQ-AGT-002-002 (Ubiquitous) — classifier location**
The `auto` mode classifier **shall** be implemented in `internal/core/agent/permission/auto.go` and **shall not** depend on external LLM calls in v3.0.

**REQ-AGT-002-003 (Ubiquitous) — safelist/denylist 상수화**
The `auto` mode safelist and denylist **shall** be declared as package-level constants in `internal/core/agent/permission/auto.go` with supporting comments citing Master design §3.4 scope-reduction rationale.

**REQ-AGT-002-004 (Ubiquitous) — feature flag 기본값**
The `.moai/config/sections/quality.yaml` field `permission.auto_mode.enabled` **shall** default to `false`.

### 5.2 Event-Driven (이벤트 기반)

**REQ-AGT-002-005 (Event-Driven) — bubble mode forwarding**
**When** an agent spawned with `permissionMode: bubble` encounters a tool call requiring permission, the permission handler **shall** forward the permission decision request to the parent session's `AskUserQuestion` channel, await the user response, and then apply the decision to the agent's tool call.

**REQ-AGT-002-006 (Event-Driven) — auto mode safe classification**
**When** an agent with `permissionMode: auto` attempts a tool call matching the safelist (`Read`, `Grep`, `Glob`, `Bash(go test)`, `Bash(go build)`, etc.), the permission handler **shall** approve the call without user prompt and log the auto-approval to `.moai/logs/agent-autoapprove.log`.

**REQ-AGT-002-007 (Event-Driven) — auto mode dangerous classification**
**When** an agent with `permissionMode: auto` attempts a tool call matching the denylist (`Bash(rm *)`, `Bash(git push --force)`, `Bash(curl *)` etc.), the permission handler **shall** deny the call with error `AGT_AUTO_DENIED` and **shall not** prompt the user.

**REQ-AGT-002-008 (Event-Driven) — auto mode ambiguous classification**
**When** an agent with `permissionMode: auto` attempts a tool call matching neither safelist nor denylist, the permission handler **shall** fall back to `default` mode behavior (user prompt via `AskUserQuestion`).

### 5.3 State-Driven (상태 기반)

**REQ-AGT-002-009 (State-Driven) — auto mode disabled**
**While** `permission.auto_mode.enabled: false` in quality.yaml, the permission handler **shall** reject any agent spawn that declares `permissionMode: auto` with error `AGT_AUTO_MODE_DISABLED` and a link to the enablement documentation.

**REQ-AGT-002-010 (State-Driven) — bubble mode without parent**
**While** an agent with `permissionMode: bubble` is spawned without a parent session (e.g., root-level via CLI), the permission handler **shall** treat `bubble` as `default` and log a one-time warning.

### 5.4 Optional (선택)

**REQ-AGT-002-011 (Optional) — allowlist/denylist override**
**Where** the user provides a custom safelist or denylist via `.moai/config/sections/quality.yaml` under `permission.auto_mode.safelist` or `permission.auto_mode.denylist`, the classifier **shall** merge the user-defined lists with the built-in lists (user-denylist overrides user-safelist on conflict).

**REQ-AGT-002-012 (Optional) — auto mode telemetry**
**Where** `permission.auto_mode.telemetry_enabled: true`, the auto-approve and auto-deny events **shall** be appended to `.moai/reports/auto-mode-stats-{ISO-date}.jsonl` (one JSON line per event) for offline analysis.

### 5.5 Unwanted Behavior

**REQ-AGT-002-013 (Unwanted Behavior) — auto mode with bypassPermissions coexistence**
**If** an agent frontmatter declares both `permissionMode: auto` and nested `bypassPermissions` configuration (via hooks or session rules), **then** the parser **shall** reject the agent with error `AGT_PERMISSION_MODE_CONFLICT` and **shall not** allow silent privilege escalation.

**REQ-AGT-002-014 (Unwanted Behavior) — classifier failure**
**If** the `auto` mode classifier encounters an internal error (e.g., malformed safelist entry), **then** the permission handler **shall** fall back to `default` mode for that specific tool call and log the classifier error with stack trace.

---

## 6. Acceptance Criteria (수용 기준 요약)

상세 Given-When-Then 시나리오는 `acceptance.md`에서 정의한다.

핵심 기준:

- **AC-AGT-002-01**: `PermissionMode` enum이 6개 값을 모두 수용하고, unknown 값은 `AGT_INVALID_FRONTMATTER` 반환
- **AC-AGT-002-02**: `bubble` 모드 에이전트가 tool permission 요청 시 parent의 `AskUserQuestion`이 호출됨 (mock 기반 unit test)
- **AC-AGT-002-03**: `auto` 모드 + safelist tool → 자동 승인 및 `auto-approve.log` 1줄 기록
- **AC-AGT-002-04**: `auto` 모드 + denylist tool → `AGT_AUTO_DENIED` 반환, user prompt 없음
- **AC-AGT-002-05**: `auto` 모드 + ambiguous tool → `default` fallback, user prompt 발생
- **AC-AGT-002-06**: `permission.auto_mode.enabled: false` 환경에서 `permissionMode: auto` agent spawn 시도 → `AGT_AUTO_MODE_DISABLED`
- **AC-AGT-002-07**: parent 없이 `bubble` 모드 agent 실행 → warning 1회 + `default` 동작
- **AC-AGT-002-08**: 사용자 safelist와 builtin denylist가 충돌 시 denylist 우선 (security-first)
- **AC-AGT-002-09**: classifier 내부 에러 발생 시 default fallback 동작, panic 없음

---

## 7. Constraints (제약)

- [HARD] `auto` 모드는 v3.0에서 default disabled. 명시적 opt-in 필수.
- [HARD] `auto` 모드에서 LLM 호출 금지(v3.0 scope — rule-based only). LLM classifier는 v3.1+ SPEC에서 별도.
- [HARD] `bubble` 모드는 기존 `AskUserQuestion` 채널만 사용. 별도 통신 채널 구축 금지.
- [HARD] 하드코딩 금지: safelist/denylist는 `internal/core/agent/permission/classifier_lists.go`의 const 또는 config file에서 로드.
- [HARD] security-first 원칙: 충돌 시 항상 denylist 우선, bypassPermissions와 auto 충돌 시 reject.
- [HARD] feature flag(quality.yaml)는 SPEC-V3-SCH-001의 validator/v10으로 검증되어야 한다.
- [HARD] 16개 언어 중립성: classifier의 safelist/denylist는 Go toolchain 편중 없이 언어별 동등하게 확장 가능하도록 pluggable 설계.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크                                                           | 영향     | 완화                                                                                              |
|------------------------------------------------------------------|----------|---------------------------------------------------------------------------------------------------|
| `auto` 모드 denylist 누락으로 위험한 command 자동 실행            | Critical | 기본 disabled; 충돌 시 denylist 우선; 독립적 security review required before enabling by default  |
| `bubble` 모드가 parent 세션 종료 후 orphan으로 남음                | Medium   | parent session abort 시 bubble agent 자동 종료 (signal propagation)                               |
| safelist/denylist의 regex 표기가 복잡해져 오탐지                   | Medium   | v3.0은 permission-rule syntax (CC 동일) 전용; regex는 v3.1+                                       |
| `auto` 모드 enabled 상태로 실수 커밋                               | Medium   | `quality.yaml`의 `auto_mode.enabled: true`는 `moai doctor` warning 발생; CI에서 prod check        |
| classifier의 ambiguous 분류가 너무 많아 UX 저하                    | Low      | safelist 기본 목록을 보수적으로 구성; telemetry 수집 후 v3.2에서 expansion 결정                   |
| 사용자 custom safelist로 denylist 우회 시도                        | Low      | REQ-AGT-002-011에 따라 denylist가 항상 우선; security-first                                       |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3-AGT-001** (Agent frontmatter v2 bundle): `PermissionMode` enum의 기본 4개 값이 먼저 정의되어야 한다.
- **SPEC-V3-SCH-001** (Formal config schemas): `quality.yaml` 스키마의 validator/v10 기반 검증.

### 9.2 Blocks

- **SPEC-V3-AGT-003** (Fork subagent primitive): fork subagent는 `permissionMode: bubble`이 전제이므로 본 SPEC 완료 후 진행 가능.

### 9.3 Related

- **SPEC-HOOKS-006** (Hook permission decision protocol — sibling writer HOOKS domain): hook의 permission decision과 agent permissionMode는 독립적이지만, `auto` 모드에서 hook 레벨 decision이 있을 경우 hook이 우선한다.
- **SPEC-V3-TEAM-001** (Team mailbox): 팀 모드에서 `permission_request`/`permission_response` 메시지가 `bubble` 모드와 상호작용.

---

## 10. Traceability (추적성)

- 총 REQ 개수: 14 (Ubiquitous 4, Event-Driven 4, State-Driven 2, Optional 2, Unwanted Behavior 2)
- 예상 AC 개수: 9
- 관련 Wave 1 근거:
  - findings-wave1-agent-team.md §4.3 (PERMISSION_MODES enum table, PermissionMode.ts:PERMISSION_MODES)
  - findings-wave1-agent-team.md §8.1 (Gap: `permissionMode: bubble` forkSubagent.ts:67; `permissionMode: auto` PERMISSION_MODES)
  - findings-wave1-moai-current.md §6.3 (Current moai permissionMode enum in agent frontmatter)
- 구현 시 각 소스 파일에 `@SPEC:SPEC-V3-AGT-002:REQ-AGT-002-NNN` 주석 부착
- 코드 구현 예상 경로:
  - `internal/config/schema/agent.go` (REQ-AGT-002-001)
  - `internal/core/agent/permission/bubble.go` (REQ-AGT-002-005, 010)
  - `internal/core/agent/permission/auto.go` (REQ-AGT-002-002, 003, 006, 007, 008, 014)
  - `internal/core/agent/permission/classifier_lists.go` (REQ-AGT-002-003, 011)
  - `internal/config/schema/quality.go` (REQ-AGT-002-004, 009, 011, 012)
  - `internal/core/agent/permission/auto_test.go` (AC-AGT-002-03..05)
  - `internal/core/agent/permission/bubble_test.go` (AC-AGT-002-02, 07)

---

End of SPEC.
