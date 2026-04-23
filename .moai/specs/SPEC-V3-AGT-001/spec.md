---
id: SPEC-V3-AGT-001
title: "Agent Frontmatter Expansion (v2 bundle)"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P0 Critical
phase: "v3.0.0 — Phase 3 Agent Runtime v2"
module: "internal/config/schema/, internal/cli/deps.go"
dependencies:
  - SPEC-V3-SCH-001
related_gap:
  - gm#56
  - gm#57
  - gm#58
  - gm#59
  - gm#60
  - gm#61
  - gm#62
  - gm#63
related_theme: "Theme 4 — Agent Frontmatter Expansion"
breaking: false
lifecycle: spec-anchored
tags: "agent, frontmatter, memory, omitClaudeMd, initialPrompt, requiredMcpServers, v3"
---

# SPEC-V3-AGT-001: Agent Frontmatter Expansion (v2 bundle)

## HISTORY

| Version | Date       | Author | Description                          |
|---------|------------|--------|--------------------------------------|
| 0.1.0   | 2026-04-22 | GOOS   | Initial draft (Wave 4 SPEC writer)   |

---

## 1. Goal (목적)

moai 에이전트 프론트매터 스키마를 Claude Code v2 스키마와 동등 수준으로 확장하여, 22개 기존 에이전트가 추가 필드(`memory`, `initialPrompt`, `requiredMcpServers`, `omitClaudeMd`, `maxTurns`, `criticalSystemReminder_EXPERIMENTAL`, `hooks` array in frontmatter)를 opt-in으로 사용할 수 있도록 한다. 특히 `omitClaudeMd: true`는 읽기 전용 research 에이전트(`researcher`, `plan-auditor`, `evaluator-active`)에 적용되어 1 spawn당 ~15 Ktok을 절감한다.

### 1.1 배경

현재 moai 에이전트 프론트매터는 `name`, `description`, `tools`, `model`, `permissionMode`, `memory`, `skills`, `hooks` 8개 필드를 지원한다(findings-wave1-moai-current.md §6.3). Claude Code의 `AgentJsonSchema`(findings-wave1-agent-team.md §4.1, loadAgentsDir.ts:73-99)는 동일한 8개 필드 외에 다음 8개 필드를 추가로 정의한다:

- `memory: user|project|local` — 에이전트 스코프 영속 메모리(findings-wave1-agent-team.md §4.2, agentMemory.ts)
- `initialPrompt` — 첫 user turn에 prepend(findings-wave1-agent-team.md §4.2, loadAgentsDir.ts:91)
- `requiredMcpServers` — 지정 MCP 서버 미연결 시 에이전트 비활성(loadAgentsDir.ts:122)
- `omitClaudeMd` — CLAUDE.md 계층 스킵(loadAgentsDir.ts:128-132; 5-15 Gtok/week 절감 per CC BQ analysis)
- `maxTurns` — agentic turn 하드캡(loadAgentsDir.ts:649)
- `criticalSystemReminder_EXPERIMENTAL` — 모든 user turn에 re-inject(loadAgentsDir.ts:121)
- `background` — 프론트매터 수준 기본 async 실행(loadAgentsDir.ts:93; moai는 spawn 시점 `run_in_background`만 지원)
- `hooks: { SessionStart: [...], ... }` — 에이전트 레벨 세션 훅(findings-wave1-agent-team.md §4.5; moai는 단일 배열 형태만 partial 지원)

BigQuery 텔레메트리 분석(findings-wave1-agent-team.md §8.1)에 따르면 `omitClaudeMd: true`를 Explore/Plan에 적용한 CC는 주당 5-15 Gtok을 절감한다. moai의 `researcher`, `plan-auditor`, `evaluator-active`는 유사한 특성(읽기 전용 또는 심판 역할, CLAUDE.md context 의존도 낮음)을 가지므로 동일한 최적화 이득을 기대할 수 있다.

### 1.2 Non-Goals

- `permissionMode: bubble` 및 `permissionMode: auto` 추가(SPEC-V3-AGT-002에서 처리)
- fork subagent 프리미티브(SPEC-V3-AGT-003에서 처리)
- `isolation: remote`(Anthropic 내부 CCR 의존; §10 Non-Goals에서 reject)
- `color` 필드(UI 전용; SPEC-V3-OUT-001 범위)

---

## 2. Scope (범위)

### 2.1 In Scope

- `internal/config/schema/agent.go` 신설: `AgentFrontmatter` 구조체 + validator/v10 태그
- `internal/template/deployer.go` 프론트매터 파서 확장: 8개 신규 필드 파싱
- `internal/cli/deps.go` 에이전트 spawn dispatcher: 신규 필드 해석 로직
- `internal/core/agent/memory_scope.go`: `memory: user|project|local` 스코프 해석 및 스냅샷 초기화
- `internal/core/agent/claude_md_loader.go`: `omitClaudeMd: true` 시 CLAUDE.md 계층 로드 스킵
- `internal/core/agent/mcp_gate.go`: `requiredMcpServers` 검증(30초 타임아웃)
- `.claude/agents/*.md` 22개 기존 에이전트 문서 v3 호환성 확인(validation warning만 발생, non-blocking)
- `.claude/agents/researcher.md`, `plan-auditor.md`, `evaluator-active.md`에 `omitClaudeMd: true` 추가(opt-in 적용)

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Claude Code `permissionMode: auto|bubble` 구현(SPEC-V3-AGT-002 범위)
- Fork subagent primitive(SPEC-V3-AGT-003 범위)
- Plan/Explore/Verification 등 CC 내장 에이전트의 moai 포트(SPEC-V3-AGT-002 범위)
- 기존 22개 에이전트의 내용 변경(옵트인 필드 추가 외)
- `color`, `isolation: remote`, `disallowedTools`(별도 SPEC 또는 Non-Goals)
- CLAUDE.md 자체 로더 재작성(기존 로더에 gate만 추가)
- 에이전트 메모리 파일 포맷 변경(SPEC-V3-MEM-001 범위)
- MCP 서버 연결 재시도 정책(단순 30초 타임아웃 후 fail-fast만 구현)

---

## 3. Environment (환경)

- 런타임: moai-adk-go v3.0.0+ (Go 1.23+)
- Claude Code v2.1.111+ (findings-wave1-agent-team.md §4)
- Opus 4.7 Adaptive Thinking 호환 (CLAUDE.md §12)
- 대상 디렉터리: `internal/config/schema/`, `internal/core/agent/`, `.claude/agents/`
- 의존 라이브러리: `github.com/go-playground/validator/v10`(SPEC-V3-SCH-001에서 도입)
- OS 동등성: macOS / Linux / Windows
- 영향 받는 에이전트 수: 22 (manager 8 + expert 8 + builder 3 + evaluator 2 + researcher 1; findings-wave1-moai-current.md §6.2)

---

## 4. Assumptions (가정)

- SPEC-V3-SCH-001이 먼저 머지되어 `validator/v10`이 직접 의존성으로 추가된 상태다.
- 22개 기존 에이전트의 frontmatter는 v2 스키마에서 error 없이 파싱된다(추가 필드는 optional).
- `omitClaudeMd: true` 대상 3개 에이전트(`researcher`, `plan-auditor`, `evaluator-active`)는 CLAUDE.md context 없이도 업무 수행 가능하다(Wave 1 분석 근거).
- MCP 서버 연결은 30초 이내 완료되거나 명확히 실패한다(CC 동작 동일, findings-wave1-agent-team.md §5.2 dispatch tree step 6).
- 에이전트 메모리 경로 충돌은 SPEC-V3-MEM-001의 path validation으로 방지된다.
- `criticalSystemReminder_EXPERIMENTAL`은 v3.0에서 실험 기능이며 default disabled다.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous (항시)

**REQ-AGT-001-001 (Ubiquitous) — 스키마 확장**
The `internal/config/schema/agent.go` **shall** define `AgentFrontmatter` struct with at least the following fields: `Name`, `Description`, `Tools`, `Model`, `PermissionMode`, `Memory`, `InitialPrompt`, `RequiredMcpServers`, `OmitClaudeMd`, `MaxTurns`, `CriticalSystemReminderExperimental`, `Background`, `Skills`, `Hooks`, `Effort`, `Isolation`.

**REQ-AGT-001-002 (Ubiquitous) — validator tag 강제**
The `AgentFrontmatter` struct **shall** include validator/v10 tags enforcing:
- `Memory` ∈ `{"", "user", "project", "local"}` (empty allowed)
- `PermissionMode` ∈ `{"default", "plan", "acceptEdits", "bypassPermissions"}` (bubble/auto는 SPEC-V3-AGT-002에서 추가)
- `MaxTurns` ≥ 0
- `Isolation` ∈ `{"", "none", "worktree"}` (remote는 Non-Goal)
- `Effort` ∈ `{"", "low", "medium", "high", "xhigh", "max"}`

**REQ-AGT-001-003 (Ubiquitous) — 기본값**
The `AgentFrontmatter` struct **shall** default the following fields when absent: `Memory = ""` (no agent-scoped memory), `OmitClaudeMd = false`, `MaxTurns = 0` (no cap), `Background = false`, `Isolation = ""`.

**REQ-AGT-001-004 (Ubiquitous) — 역호환성**
The frontmatter parser **shall** accept all existing 22 agent definitions (pre-v3 frontmatter) without validation errors.

### 5.2 Event-Driven (이벤트 기반)

**REQ-AGT-001-005 (Event-Driven) — omitClaudeMd 동작**
**When** an agent is spawned with `omitClaudeMd: true`, the spawn dispatcher **shall** skip the CLAUDE.md hierarchy loading for that agent's system prompt assembly and log the skip event to `.moai/logs/agent-spawn.log`.

**REQ-AGT-001-006 (Event-Driven) — initialPrompt injection**
**When** an agent with non-empty `initialPrompt` is spawned, the spawn dispatcher **shall** prepend the `initialPrompt` content to the first user-role message of the agent's conversation history.

**REQ-AGT-001-007 (Event-Driven) — requiredMcpServers 검증**
**When** an agent with non-empty `requiredMcpServers` is spawned, the spawn dispatcher **shall** wait up to 30 seconds for each listed MCP server to report `connected` status, then:
- if all connected → proceed with spawn
- if any server still disconnected → abort spawn with structured error `AGT_MCP_UNAVAILABLE` listing the offending server names

**REQ-AGT-001-008 (Event-Driven) — maxTurns 적용**
**When** an agent's turn count reaches `maxTurns`, the agent runtime **shall** terminate the agent loop and return a result with `stop_reason: "max_turns"`.

**REQ-AGT-001-009 (Event-Driven) — memory 스코프 초기화**
**When** an agent with non-empty `memory` field is spawned for the first time in a project, the spawn dispatcher **shall** create the agent-scoped memory directory at one of:
- `memory: user` → `~/.claude/agents/{agent-name}/memory/`
- `memory: project` → `.moai/agents/{agent-name}/memory/`
- `memory: local` → `.moai/agents/{agent-name}/memory.local/` (git-ignored)

### 5.3 State-Driven (상태 기반)

**REQ-AGT-001-010 (State-Driven) — Background 프론트매터 + spawn 우선순위**
**While** `background: true` is declared in the agent frontmatter, the spawn dispatcher **shall** treat the agent as default-async. The spawn-time `run_in_background` parameter, if present, overrides the frontmatter value.

**REQ-AGT-001-011 (State-Driven) — criticalSystemReminder 재주입**
**While** `criticalSystemReminder_EXPERIMENTAL` is non-empty AND the experimental feature flag `MOAI_AGT_CRITICAL_REMINDER` is truthy, the agent runtime **shall** append the reminder text (wrapped in `<system-reminder>...</system-reminder>`) to every user-role message of the agent's conversation before sending to the model.

### 5.4 Optional (선택)

**REQ-AGT-001-012 (Optional) — opt-in omitClaudeMd for 3 agents**
**Where** the SPEC is shipped, the template frontmatter of `.claude/agents/researcher.md`, `.claude/agents/plan-auditor.md`, and `.claude/agents/evaluator-active.md` **shall** include `omitClaudeMd: true` as an opt-in optimization (documented in each agent's HISTORY section).

**REQ-AGT-001-013 (Optional) — skills as YAML array**
**Where** an agent frontmatter declares `skills:`, the parser **shall** accept both CSV string form (legacy) and YAML array form (CC-compat). Array form is the preferred representation.

### 5.5 Unwanted Behavior

**REQ-AGT-001-014 (Unwanted Behavior) — unknown field strictness**
**If** an agent frontmatter contains a field name not defined in the v3 `AgentFrontmatter` schema, **then** the parser **shall** emit a warning (non-blocking) to stderr listing the unknown field and continue parsing. Unless `MOAI_CONFIG_STRICT=1` is set, in which case the parser **shall** reject the agent with error `AGT_UNKNOWN_FIELD`.

**REQ-AGT-001-015 (Unwanted Behavior) — validation failure**
**If** the agent frontmatter fails validator/v10 validation (out-of-range `maxTurns`, invalid `memory` enum, etc.), **then** the parser **shall** reject the agent with error `AGT_INVALID_FRONTMATTER` and include the specific field name and expected values in the error message.

---

## 6. Acceptance Criteria (수용 기준 요약)

상세 Given-When-Then 시나리오는 `acceptance.md`에서 정의한다.

핵심 기준:

- **AC-AGT-001-01**: 22개 기존 에이전트가 v3 스키마에서 0 error로 파싱됨
- **AC-AGT-001-02**: `researcher` 에이전트 spawn 시 CLAUDE.md 로드 스킵 확인 (`.moai/logs/agent-spawn.log`에 skip 이벤트 기록)
- **AC-AGT-001-03**: `researcher` 에이전트의 per-spawn token usage가 v2.12 대비 약 15 Ktok 감소 (측정 방법: `moai doctor agent --measure-tokens`)
- **AC-AGT-001-04**: `requiredMcpServers: [context7]`를 가진 테스트 에이전트가 context7 미연결 환경에서 `AGT_MCP_UNAVAILABLE` 에러를 30±2초 이내에 반환
- **AC-AGT-001-05**: `maxTurns: 5`를 가진 에이전트가 5턴 도달 시 `stop_reason: max_turns`로 종료
- **AC-AGT-001-06**: `memory: project`를 가진 에이전트 첫 spawn 시 `.moai/agents/{name}/memory/` 디렉터리 자동 생성
- **AC-AGT-001-07**: `initialPrompt`가 선언된 에이전트의 첫 turn에 initialPrompt 내용이 prepend됨 (agent transcript 검증)
- **AC-AGT-001-08**: `MOAI_CONFIG_STRICT=1` 환경에서 unknown field가 있으면 parse 실패, 기본 환경에서는 warning만 출력
- **AC-AGT-001-09**: `go test ./internal/config/schema/... ./internal/core/agent/...` 전체 통과 (coverage ≥ 85%)

---

## 7. Constraints (제약)

- [HARD] 22개 기존 에이전트의 행동은 새 필드 opt-in 없이는 v2.12와 byte-identical 동등해야 한다.
- [HARD] `omitClaudeMd: true`는 v3.0에서 default false로 고정한다 (Master §9 Open Question #9 Recommended default).
- [HARD] 하드코딩 금지(CLAUDE.local.md §14): `MAX_MCP_WAIT_SECONDS = 30`, `DEFAULT_MAX_TURNS = 0` 등은 `internal/config/defaults.go`에 const로 정의.
- [HARD] 템플릿 우선 원칙(CLAUDE.local.md §2): `.claude/agents/*.md` 변경은 `internal/template/templates/.claude/agents/`에 먼저 적용 후 `make build`.
- [HARD] 16개 언어 중립성(CLAUDE.local.md §15): 스키마와 파서는 특정 사용자 프로젝트 언어에 의존하지 않는다.
- [HARD] validator/v10 외 추가 의존성 금지: moai의 9-direct-dep 철학 유지 (Master §1.2).
- [HARD] `criticalSystemReminder_EXPERIMENTAL`은 feature flag 뒤에 gate. default disabled.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크                                                            | 영향   | 완화                                                                                              |
|-------------------------------------------------------------------|--------|---------------------------------------------------------------------------------------------------|
| `omitClaudeMd: true` 적용 에이전트가 CLAUDE.md 규칙을 위반        | High   | v3.0에서 3개 에이전트만 opt-in; 각 에이전트 기능 회귀 테스트 수행; R-005 in Master §7             |
| `requiredMcpServers` 30초 타임아웃이 실서비스 환경에서 너무 짧음  | Medium | `MOAI_MCP_WAIT_SECONDS` 환경변수로 override 허용; default 30s는 CC와 동등                         |
| `memory: local` 디렉터리가 `.gitignore` 누락으로 커밋됨           | Medium | 템플릿 `.gitignore`에 `.moai/agents/*/memory.local/` 패턴 포함(SPEC-V3-SCH-001 배포 시 동시 적용) |
| 기존 skills CSV 표기와 YAML array 혼재로 파싱 충돌                | Low    | 두 표기 모두 accept; REQ-AGT-001-013 명시적 규정; unit test로 검증                                |
| `initialPrompt`가 slash command와 충돌                            | Low    | CC 동작 그대로 따름(findings §4.2: loadAgentsDir.ts:91); 회귀 케이스 문서화                       |
| `criticalSystemReminder_EXPERIMENTAL` 남용으로 token budget 초과  | Low    | feature flag gate; 미사용이 default; 사용 시 docs에 warning                                       |
| unknown field warning 너무 많아 noise                              | Low    | non-blocking warning; 한 agent당 최대 1회 출력 (dedup)                                            |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3-SCH-001** (Formal config schemas): `validator/v10` 직접 의존성 추가가 먼저 필요하다.

### 9.2 Blocks

- **SPEC-V3-AGT-002** (Built-in moai agents): 본 SPEC의 frontmatter 확장을 전제로 `Explore`/`Plan` 변형 정의.
- **SPEC-V3-AGT-003** (Fork subagent primitive): 본 SPEC의 `permissionMode` enum 확장을 전제로 `bubble` 추가(간접 의존).

### 9.3 Related

- **SPEC-V3-MEM-001** (Memory 2.0): `memory: user|project|local` 경로 validation은 MEM-001의 `validateMemoryPath`를 재사용한다.
- **SPEC-V3-SKL-001** (Skill frontmatter v2): skills YAML array 파싱 규칙 공유.
- **SPEC-HOOKS-004** (Hook handler richness — sibling writer): `hooks` array 파싱 시 hook v2 output contract 준수.

---

## 10. Traceability (추적성)

- 본 SPEC의 모든 REQ ID는 `plan.md` 마일스톤 및 `acceptance.md` 시나리오로 역참조된다.
- 구현 시 각 소스 파일에 `@SPEC:SPEC-V3-AGT-001:REQ-AGT-001-NNN` 주석을 부착한다.
- 총 REQ 개수: 15 (Ubiquitous 4, Event-Driven 5, State-Driven 2, Optional 2, Unwanted Behavior 2)
- 예상 AC 개수: 9
- 관련 Wave 1 근거:
  - findings-wave1-agent-team.md §4.1 (AgentJsonSchema, loadAgentsDir.ts:73-99)
  - findings-wave1-agent-team.md §4.2 (Extra fields table; loadAgentsDir.ts:91, :122, :128-132, :121, :649, :93)
  - findings-wave1-agent-team.md §5.2 (Dispatch decision tree — MCP servers wait 30s at step 6)
  - findings-wave1-agent-team.md §8.1 (Gap table: `memory`, `initialPrompt`, `requiredMcpServers`, `omitClaudeMd`, `maxTurns`, `criticalSystemReminder`, `skills`, `background`)
  - findings-wave1-moai-current.md §6.3 (Current moai agent frontmatter fields in use)
- 코드 구현 예상 경로:
  - `internal/config/schema/agent.go` (REQ-AGT-001-001, 002, 003)
  - `internal/config/schema/agent_test.go` (REQ-AGT-001-014, 015)
  - `internal/core/agent/claude_md_loader.go` (REQ-AGT-001-005)
  - `internal/core/agent/initial_prompt.go` (REQ-AGT-001-006)
  - `internal/core/agent/mcp_gate.go` (REQ-AGT-001-007)
  - `internal/core/agent/turn_counter.go` (REQ-AGT-001-008)
  - `internal/core/agent/memory_scope.go` (REQ-AGT-001-009)
  - `internal/core/agent/background.go` (REQ-AGT-001-010)
  - `internal/core/agent/critical_reminder.go` (REQ-AGT-001-011)
  - `.claude/agents/researcher.md`, `plan-auditor.md`, `evaluator-active.md` (REQ-AGT-001-012)

---

End of SPEC.
