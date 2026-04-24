---
id: SPEC-V3R2-MIG-002
title: Hook Registration Cleanup (orphan events, stub handlers, divergence)
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 2 SPEC writer (Layer 6/7/Cleanup)
priority: P1 High
phase: "v3.0.0 — Phase 8 — Migration Tool + Docs"
module: "internal/hook/, internal/cli/deps.go, .claude/hooks/moai/, internal/template/templates/.claude/settings.json.tmpl"
dependencies:
  - SPEC-V3R2-EXT-004
related_gap:
  - r6-hooks-audit
  - r6-orphan-handlers
related_theme: "Theme 2 — Runtime Hardening"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "hooks, cleanup, orphan, stub, divergence, handler, v3"
---

# SPEC-V3R2-MIG-002: Hook Registration Cleanup

## HISTORY

| Version | Date       | Author | Description                                                          |
|---------|------------|--------|----------------------------------------------------------------------|
| 0.1.0   | 2026-04-23 | Wave 2 | Initial SPEC — orphan handler cleanup + stub removal + settings sync  |

---

## 1. Goal (목적)

R6 audit §2는 MoAI hook 생태계의 3가지 균열을 식별했다: (a) `setupHandler` orphan (Go 핸들러만 존재, shell wrapper 없고 settings.json 미등록), (b) **14/27 handlers (~52%)가 thin logger** (business logic 없음), (c) `subagentStopHandler` 미구현 버그(tmux pane kill 누락). 본 SPEC은 이 세 issue를 codify 방식으로 해결한다: orphan handler 등록 또는 제거, stub logger의 upgrade/retire 판정, settings.json과 Go handler 간 divergence 제거.

### 1.1 배경

R6 §2.2 per-handler verdict 표:
- **Orphan**: `setupHandler` (no shell wrapper, no settings.json entry).
- **Stub loggers (14)**: config_change, file_changed, elicitation, elicitation_result, instructions_loaded, notification, subagent_stop, task_created, post_tool_failure, worktree_create/remove, setup, cwd_changed, permission_request (일부는 smart).
- **Missing logic**: subagent_stop은 tmux pane kill 필요 (known bug per agent-memory `feedback_team_tmux_cleanup.md`).

R6 §2 cross-cutting: "Top 5 handler gaps (registered but no-op): SubagentStop, ConfigChange, Setup, InstructionsLoaded, FileChanged".

### 1.2 비목표 (Non-Goals)

- 새 hook event 추가
- 기존 rich handler(SessionStart, PreToolUse 등)의 재작성
- Hook execution framework 변경 (기존 shell wrapper + Go handler 구조 유지)
- Hook timeout 설정 변경
- Hook event schema (JSON input/output) 변경
- Hook event ordering 변경

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns**: orphan handler 처리 (setup), stub logger 5개 upgrade decision, subagent_stop 버그 수정(tmux pane kill), settings.json ↔ handler ↔ wrapper 3-way sync 검증.
- Orphan handler 판정: `setupHandler` retire (R6 §2.2 권장) — Go 코드에서 제거하고 관련 참조 정리.
- Stub → real logic upgrade: `subagentStopHandler`(tmux pane kill), `configChangeHandler`(config reload trigger), `instructionsLoadedHandler`(CLAUDE.md character budget validation), `fileChangedHandler`(MX rescan trigger), `postToolUseFailureHandler`(error classification).
- Stub → retire: `notificationHandler`, `elicitationHandler`, `elicitationResultHandler`, `taskCreatedHandler` (R6 recommended retire or retain-only-in-Go + remove from settings.json).
- settings.json 3-way sync: 각 등록된 event는 (Go handler) + (shell wrapper) + (settings.json entry) 모두 존재.
- Hook protocol 준수: JSON-OR-ExitCode (pattern §T-5) 확인, 신규 로직도 동일 프로토콜.

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- 새 hook event 추가
- Hook JSON schema 변경
- Rich handler(SessionStart 등) 재작성
- Hook execution framework 변경
- Hook timing / timeout 튜닝
- Agent-scoped hook 추가

---

## 3. Environment (환경)

- 런타임: moai-adk-go (Go 1.26+), Claude Code hook runtime, `.claude/hooks/moai/*.sh`, `.claude/settings.json`
- 영향 디렉터리:
  - 수정: `internal/hook/subagent_stop.go`, `config_change.go`, `instructions_loaded.go`, `file_changed.go`, `post_tool_failure.go`
  - 수정: `internal/cli/deps.go` (handler registration 삭제/추가)
  - 제거: `internal/hook/setup.go`, `notification.go`, `elicitation.go`, `task_created.go` (settings.json entry도 제거)
  - 수정: `internal/template/templates/.claude/settings.json.tmpl`
  - 참조: `.claude/rules/moai/core/hooks-system.md` (event 표 업데이트)
- 외부 레퍼런스: R6 §2 hooks audit, `/Users/goos/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_team_tmux_cleanup.md`

---

## 4. Assumptions (가정)

- SPEC-V3R2-EXT-004 migration framework이 cleanup step 실행 경로를 제공한다.
- subagentStopHandler의 tmux pane kill 요구사항은 memory `feedback_team_tmux_cleanup.md`에 이미 근거를 두고 있다 (team config의 `tmuxPaneId` 필드 소비).
- settings.json 3-way sync는 `internal/template/settings_audit_test.go` 혹은 신규 테스트로 CI에 보장 가능하다.
- 삭제되는 stub handler들은 Claude Code native event 감지에 영향을 주지 않는다 (settings.json entry 없어도 이벤트는 fire 가능; MoAI만 subscribe 안 할 뿐).
- `setupHandler`는 현재 아무 business logic이 없으므로 안전하게 retire 가능.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-MIG002-001**
The v3 hook registry **shall** have strict 3-way sync: every registered Go handler has a matching shell wrapper AND settings.json entry (no orphans).

**REQ-MIG002-002**
Post-cleanup, the registered handler count **shall** be 22 or fewer (current 27 minus retirements).

**REQ-MIG002-003**
The v3 hook registry **shall** retire: `setupHandler`, `notificationHandler`, `elicitationHandler`, `elicitationResultHandler`, `taskCreatedHandler`.

**REQ-MIG002-004**
The `subagentStopHandler` **shall** implement tmux pane kill logic when the stopping subagent is a teammate with `tmuxPaneId` in team config.

**REQ-MIG002-005**
The `configChangeHandler` **shall** trigger config reload when `.moai/config/sections/*.yaml` changes.

**REQ-MIG002-006**
The `instructionsLoadedHandler` **shall** validate CLAUDE.md character budget per `.claude/rules/moai/development/coding-standards.md` (≤ 40,000 chars).

**REQ-MIG002-007**
The `fileChangedHandler` **shall** trigger MX annotation re-scan for externally-edited files.

**REQ-MIG002-008**
The `postToolUseFailureHandler` **shall** classify errors into at least 3 categories: `permission`, `timeout`, `other`, and surface user-friendly systemMessage.

### 5.2 Event-Driven Requirements

**REQ-MIG002-009**
**When** settings.json and Go handler registry diverge, CI **shall** fail with `HOOK_SYNC_DRIFT`.

**REQ-MIG002-010**
**When** a hook wrapper is present but no matching Go handler exists, CI **shall** fail with `HOOK_WRAPPER_ORPHAN`.

**REQ-MIG002-011**
**When** migration step runs, it **shall** remove shell wrappers corresponding to retired handlers.

**REQ-MIG002-012**
**When** a user's project has a retired handler registered in their local `settings.json`, the migration **shall** remove the entry.

### 5.3 State-Driven Requirements

**REQ-MIG002-013**
**While** `subagentStopHandler` receives a stop event for a teammate subagent, the handler **shall** read `~/.claude/teams/<team-name>/config.json` for `tmuxPaneId` and invoke `tmux kill-pane -t <id>` before exit.

**REQ-MIG002-014**
**While** CLAUDE.md exceeds 40,000 chars at session start, `instructionsLoadedHandler` **shall** emit warning `CLAUDE_MD_OVER_BUDGET` with current char count.

### 5.4 Optional Requirements

**REQ-MIG002-015**
**Where** the user has `noop` handlers enabled (opt-in via `.moai/config/sections/state.yaml: hooks.retain_noop: true`), the 5 retired handlers **may** be re-registered for observability tap.

**REQ-MIG002-016**
**Where** `fileChangedHandler` detects a non-source file (e.g., .md docs), it **may** skip MX re-scan.

### 5.5 Complex Requirements (Unwanted Behavior / Composite)

**REQ-MIG002-017 (Unwanted Behavior)**
**If** `subagentStopHandler` cannot find `tmuxPaneId` in team config, **then** the handler **shall** log the gap and continue without error (graceful degradation).

**REQ-MIG002-018 (Unwanted Behavior)**
**If** `configChangeHandler` triggers a reload that fails (malformed YAML), **then** the handler **shall** log `CONFIG_RELOAD_FAILED` with the error but not abort the Claude Code session.

**REQ-MIG002-019 (Complex: State + Event)**
**While** the migration is in progress, **when** a user's local `.claude/hooks/moai/handle-notification.sh` is detected, the migration **shall** move the file to `.moai/archive/hooks/v3.0/` before settings.json cleanup.

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-MIG002-01**: Given the v3 hook registry When inspected Then exactly 22 or fewer handlers are registered (maps REQ-MIG002-002, REQ-MIG002-003).
- **AC-MIG002-02**: Given `internal/hook/setup.go` When inspected post-migration Then the file no longer exists (maps REQ-MIG002-003).
- **AC-MIG002-03**: Given `subagentStopHandler` receives teammate stop event with `tmuxPaneId` When handler runs Then `tmux kill-pane -t <id>` is invoked (maps REQ-MIG002-004, REQ-MIG002-013).
- **AC-MIG002-04**: Given `.moai/config/sections/quality.yaml` is edited When `configChangeHandler` fires Then config reload is triggered (maps REQ-MIG002-005).
- **AC-MIG002-05**: Given CLAUDE.md with 42,000 chars When session starts Then `CLAUDE_MD_OVER_BUDGET` warning fires (maps REQ-MIG002-014).
- **AC-MIG002-06**: Given a file is edited externally When `fileChangedHandler` fires Then MX rescan is triggered (maps REQ-MIG002-007).
- **AC-MIG002-07**: Given a tool invocation fails with permission error When `postToolUseFailureHandler` runs Then category `permission` is classified and user-friendly message is surfaced (maps REQ-MIG002-008).
- **AC-MIG002-08**: Given settings.json has an entry but no Go handler exists When CI runs Then `HOOK_SYNC_DRIFT` failure (maps REQ-MIG002-009).
- **AC-MIG002-09**: Given `.claude/hooks/moai/handle-foo.sh` exists but no Go handler When CI runs Then `HOOK_WRAPPER_ORPHAN` failure (maps REQ-MIG002-010).
- **AC-MIG002-10**: Given user's local settings.json has notification entry When migration runs Then the entry is removed (maps REQ-MIG002-012).
- **AC-MIG002-11**: Given user's project has `handle-notification.sh` When migration runs Then the file is moved to `.moai/archive/hooks/v3.0/` (maps REQ-MIG002-019).
- **AC-MIG002-12**: Given `subagentStopHandler` with no `tmuxPaneId` When fired Then gap is logged but no error (maps REQ-MIG002-017).
- **AC-MIG002-13**: Given `configChangeHandler` reload fails (malformed yaml) When fired Then `CONFIG_RELOAD_FAILED` log + session continues (maps REQ-MIG002-018).
- **AC-MIG002-14**: Given `state.yaml: hooks.retain_noop: true` When migration runs Then retired handlers remain registered (maps REQ-MIG002-015).

---

## 7. Constraints (제약)

- 9-direct-dep 정책 준수.
- `tmux kill-pane` invocation은 tmux 설치 여부 확인 후 conditional (부재 시 graceful skip).
- Hook protocol(JSON-OR-ExitCode) 준수.
- Protected directories 규정 준수: `.moai/archive/`는 본 SPEC에서 생성 가능한 영역.
- settings.json 수정은 `gopkg.in/yaml.v3` + 기존 JSON 유틸 재사용.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 완화 |
|---|---|---|
| subagent_stop tmux kill이 race condition 유발 | 테스트 환경 불안정 | REQ-MIG002-017의 graceful degradation + tmux 존재 확인 |
| 사용자가 retired handler에 의존 | 기능 손실 | REQ-MIG002-015의 `retain_noop` opt-in 제공 |
| CLAUDE.md char budget validation이 false positive | 세션 시작 경고 noise | warning 수준; error 아님 |
| configChangeHandler reload가 Claude Code freeze 유발 | UX 저하 | REQ-MIG002-018의 fail-safe + async reload |
| settings.json 편집 실수로 JSON parse error | 세션 시작 실패 | 본 SPEC의 template/local sync CI로 사전 방지 |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3R2-EXT-004: migration framework이 cleanup step을 감쌈.

### 9.2 Blocks

- SPEC-V3R2-MIG-001: v2→v3 migrator가 본 SPEC의 cleanup step을 호출.

### 9.3 Related

- agent-memory `feedback_team_tmux_cleanup.md` (subagent_stop 버그 근거).
- R6 §2 hooks audit.

---

## 10. Traceability (추적성)

- REQ 총 19개: Ubiquitous 8, Event-Driven 4, State-Driven 2, Optional 2, Complex 3.
- AC 총 14개, 모든 REQ에 최소 1개 AC 매핑 (100% 커버리지).
- Wave 2 소스 앵커: R6 §2 hooks audit; agent-memory feedback_team_tmux_cleanup.md; `.claude/rules/moai/development/coding-standards.md` (40K cap).
- BC 영향: 없음 (retired handlers는 no-op이었으므로 behavior 손실 없음; `retain_noop` opt-in으로 복구 가능).
- 구현 경로 예상:
  - `internal/hook/subagent_stop.go` (tmux pane kill 로직)
  - `internal/hook/config_change.go` (reload trigger)
  - `internal/hook/instructions_loaded.go` (char budget check)
  - `internal/hook/file_changed.go` (MX rescan)
  - `internal/hook/post_tool_failure.go` (error classification)
  - `internal/hook/setup.go` 삭제, `notification.go`, `elicitation.go`, `task_created.go` 삭제
  - `internal/cli/deps.go` registration 업데이트
  - `internal/template/templates/.claude/settings.json.tmpl`
  - `internal/template/settings_audit_test.go` (확장)

---

End of SPEC.
