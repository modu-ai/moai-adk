---
id: SPEC-V3-HOOKS-004
title: "Hook Matcher & Filter System — if condition, tool matchers, event matchers"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P1 High
phase: "v3.0.0 — Phase 2 Hook Protocol v2 Core"
module: "internal/hook/condition.go, internal/hook/matcher.go"
dependencies:
  - SPEC-V3-HOOKS-001
related_gap:
  - gm#8
  - gm#13
  - gm#14
related_theme: "Theme 1: Hook Protocol v2 — Matcher Upgrade"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "hook, v3, matcher, condition, permission-rule"
---

# SPEC-V3-HOOKS-004: Hook Matcher & Filter System

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-22 | GOOS | Initial v3 draft from Wave 4 bundle (Hooks/Commands) |

---

## 1. Goal (목적)

Upgrade moai-adk's hook matching from simple string matchers to Claude Code's full matcher system. Two mechanisms:

1. The `matcher` field supports exact names, pipe-separated lists (`Write|Edit|Bash`), regex patterns, and the wildcard `*`. Per-event `matchQuery` resolution determines which payload field the matcher is applied to.
2. The `if` field filters hooks BEFORE spawn using permission-rule syntax (e.g., `Bash(git *)`, `Read(*.ts)`). This avoids the cost of spawning a subprocess only to have it no-op.

Dedup across sources uses the composite key `{shell}\0{command}\0{if}` so the same hook appearing in multiple tiers is only run once. This SPEC's matcher and filter primitives are a prerequisite for HOOKS-006's source precedence semantics.

## 2. Scope (범위)

In-scope:
- `matcher` field upgrade: exact / pipe-separated / regex / `*`.
- Per-event `matchQuery` resolution: `tool_name` (tool events), `source` (SessionStart/ConfigChange), `trigger` (Setup/PreCompact/PostCompact), `notification_type` (Notification), `reason` (SessionEnd), `error` (StopFailure), `agent_type` (Subagent*), `mcp_server_name` (Elicitation*), `load_reason` (InstructionsLoaded), `basename(file_path)` (FileChanged), none (TeammateIdle, TaskCreated, TaskCompleted, Stop, UserPromptSubmit, WorktreeCreate, WorktreeRemove, CwdChanged).
- `if` field: permission-rule syntax evaluator (e.g., `Bash(git *)`, `Read(*.ts)`) implemented in `internal/hook/condition.go`.
- Integration with existing permission rule parser in moai-adk (if any). If no parser exists, port minimum subset from CC's permission-rule grammar.
- Dedup key computation: `{shell}\0{command}\0{if}` per SPEC-V3-HOOKS-003 once-hook convention.
- Matcher + if filter applied to all 4 hook types defined in SPEC-V3-HOOKS-002 (command / prompt / agent / http).

Out-of-scope:
- Source precedence merge logic (which source wins when same hook appears in multiple tiers) → SPEC-V3-HOOKS-006.
- Per-hook `shell: powershell` selection → deferred (current moai is bash-only on Unix, `cmd`/`pwsh` via template on Windows).
- Workspace trust gating (`shouldSkipHookDueToTrust`) → deferred.

## 3. Environment (환경)

Current moai-adk state:
- Hook matching is limited to event-type dispatch in `internal/hook/registry.go:18-325` (findings-wave1-moai-current.md §5.1). No per-hook matcher pattern exists.
- No `if` condition filter; every registered hook for an event spawns unconditionally.
- No dedup between multiple sources (single-tier loader).

Claude Code reference:
- `utils/hooks.ts:1346-1381` — matcher semantics (exact/pipe/regex/`*`) (findings-wave1-hooks-commands.md §3.8).
- `utils/hooks.ts:1616-1669` — per-event matchQuery resolution table (findings-wave1-hooks-commands.md §3.8).
- `utils/hooks.ts:1390-1421` — `if` condition evaluator using permission-rule syntax (findings-wave1-hooks-commands.md §14.3).
- `utils/hooks.ts:1723-1801` — dedup key computation (findings-wave1-hooks-commands.md §3.9).
- `schemas/hooks.ts:19-27` — `if` field schema (findings-wave1-hooks-commands.md §2.1).

Affected modules:
- `internal/hook/matcher.go` — new file: matcher compilation (string → exact / set / regex).
- `internal/hook/condition.go` — new file: permission-rule expression evaluator.
- `internal/hook/registry.go` — integrate matcher + if filter into dispatch.
- `internal/hook/types.go` — add `Matcher string` and `If string` fields to HookDeclaration.

## 4. Assumptions (가정)

- Permission-rule syntax is stable (`Tool(pattern)` with glob support). Reference grammar exists in CC at `utils/hooks.ts:1390-1421` and companion permission rule parser.
- Regex matchers are expressible via Go's `regexp` package (RE2 syntax); we reject patterns that `regexp.Compile` cannot accept.
- Matchers are case-sensitive by default (matching CC behavior).
- `{shell}\0{command}\0{if}` uses literal null-byte separator because that sequence cannot occur in any of the three components on any supported platform.

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

- REQ-HOOKS-004-001: The `HookDeclaration` schema SHALL expose optional string fields `matcher` and `if`.
- REQ-HOOKS-004-002: An empty `matcher` or value `*` SHALL match every event instance (no filter).
- REQ-HOOKS-004-003: A `matcher` value matching the regex `^[a-zA-Z0-9_|]+$` SHALL be treated as an exact or pipe-separated list match (no regex compile).
- REQ-HOOKS-004-004: A `matcher` value not matching the simple-form regex SHALL be compiled as a Go `regexp` pattern; invalid regex SHALL produce `HookMatcherCompileError` at load time.
- REQ-HOOKS-004-005: The system SHALL resolve each event's `matchQuery` payload field according to the table in Section 3.8 of findings-wave1-hooks-commands.md (e.g., PreToolUse → `tool_name`, SessionStart → `source`, FileChanged → `basename(file_path)`).
- REQ-HOOKS-004-006: The `if` field SHALL be evaluated as a permission-rule expression against the hook input payload before subprocess spawn.
- REQ-HOOKS-004-007: The system SHALL compute a dedup key `{shell}\0{command}\0{if}` for each hook declaration and refuse to enqueue duplicates within the same source tier.

### 5.2 Event-driven Requirements

- REQ-HOOKS-004-010: WHEN the resolved `matchQuery` for an event is empty (e.g., TeammateIdle), the matcher SHALL be ignored and the hook SHALL always match.
- REQ-HOOKS-004-011: WHEN an `if` expression evaluates to false, the hook SHALL NOT spawn and the dispatcher SHALL log a trace-level skip event.
- REQ-HOOKS-004-012: WHEN an `if` expression fails to parse at load time, the system SHALL reject the hook declaration with `HookIfConditionParseError` and continue loading other declarations.
- REQ-HOOKS-004-013: WHEN a dedup collision is detected within one source tier, the second declaration SHALL be discarded and a warn-level log emitted naming both sources.

### 5.3 State-driven Requirements

- REQ-HOOKS-004-020: WHILE a regex-compiled matcher has been cached during session start, subsequent dispatches SHALL reuse the compiled `regexp.Regexp` pointer without recompilation.
- REQ-HOOKS-004-021: WHILE the configuration key `hook.matcher.case_insensitive: true` is set, matcher comparisons SHALL use `strings.EqualFold` for exact/list forms and the regex SHALL be compiled with `(?i)` prefix.

### 5.4 Optional Features

- REQ-HOOKS-004-030: WHERE a hook declaration specifies `if` but the event has no payload fields required for the expression (e.g., `Stop` has no `tool_name` so `Bash(*)` cannot match), the system SHALL evaluate the expression against an empty payload and SHALL treat pattern-mismatch as false.

### 5.5 Complex Requirements

- REQ-HOOKS-004-040: IF an `if` expression contains a reference to a payload field absent from the event schema (e.g., `tool_input.file_path` on a Notification event), THEN the evaluator SHALL return false and log a warning; ELSE the expression is evaluated normally.
- REQ-HOOKS-004-041: IF both `matcher: "*"` and `if: "Bash(git *)"` are set on a PreToolUse hook, THEN the matcher gate passes for all events but the `if` gate filters to git-related Bash calls only; ELSE no filtering is applied beyond the matcher.

## 6. Acceptance Criteria (수용 기준 요약)

- AC-HOOKS-004-01: Given a PreToolUse hook with `matcher: "Write|Edit"`, When a Read tool fires, Then the hook does not spawn. (maps REQ-HOOKS-004-003)
- AC-HOOKS-004-02: Given a PreToolUse hook with `matcher: "^(Write|Edit|Bash)$"` (regex), When a Bash tool fires, Then the hook spawns. (maps REQ-HOOKS-004-004)
- AC-HOOKS-004-03: Given a PreToolUse hook with `matcher: "Bash"`, `if: "Bash(git *)"`, When tool input is `{"command": "ls"}`, Then the hook does not spawn because the if-condition evaluates false. (maps REQ-HOOKS-004-006, REQ-HOOKS-004-011)
- AC-HOOKS-004-04: Given a PreToolUse hook with `matcher: "Bash"`, `if: "Bash(git *)"`, When tool input is `{"command": "git commit"}`, Then the hook spawns. (maps REQ-HOOKS-004-011)
- AC-HOOKS-004-05: Given an `if` field with syntax `Bash(git *` (missing closing paren), When the declaration loads, Then the system emits `HookIfConditionParseError` and skips the declaration. (maps REQ-HOOKS-004-012)
- AC-HOOKS-004-06: Given two identical hook declarations in the same source tier with matching `{shell, command, if}`, When loaded, Then only one is kept and a warn log names both file paths. (maps REQ-HOOKS-004-007, REQ-HOOKS-004-013)
- AC-HOOKS-004-07: Given a TeammateIdle hook (no matchQuery), When the event fires, Then matcher evaluation is skipped and the hook runs. (maps REQ-HOOKS-004-010)
- AC-HOOKS-004-08: Given `hook.matcher.case_insensitive: true`, When a hook declares `matcher: "write"` and a Write tool fires, Then the hook matches. (maps REQ-HOOKS-004-021)
- AC-HOOKS-004-09: Given a FileChanged event where file_path is `/abs/path/.envrc` and matcher is `.envrc|.env`, When resolved, Then the basename `.envrc` matches and the hook spawns. (maps REQ-HOOKS-004-005)

## 7. Constraints (제약)

- Technical: Go 1.22+. `regexp` stdlib is sufficient; no third-party regex engines.
- Backward compat: Empty `matcher` and empty `if` preserve v2 unconditional-match behavior.
- Platform: Matcher and condition logic is platform-neutral; path-based matchers use `filepath.Base` for FileChanged.
- Performance: Matcher evaluation MUST add ≤ 0.5 ms p99 per hook per event dispatch.
- Security: Regex DoS protection via Go's RE2 engine (linear time guarantee); no backreferences, no lookahead.

## 8. Risks & Mitigations (리스크 및 완화)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Regex matcher misuse (overly broad patterns match unintended events) | M | M | `moai doctor hook --validate` flags matchers that would compile to a catch-all; documented best practices in migration guide. |
| `if` expression grammar drifts from permission-rule parser | L | M | Shared grammar module between permission rules and hook conditions; golden tests ensure parity. |
| Dedup collision silently discards user's intended duplicate | L | L | Warn-level log emits file paths of both; `moai doctor hook --list` shows dedup decisions. |
| Payload field references in `if` break when event schema evolves | M | M | Unknown field evaluates to false (REQ-HOOKS-004-040); tests cover every event's schema. |

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3-HOOKS-001 (HookDeclaration struct shape must exist first).

### 9.2 Blocks

- SPEC-V3-HOOKS-006 (source precedence hierarchy dedup key derived from matcher+if).

### 9.3 Related

- SPEC-V3-HOOKS-002 (matcher/if applies to all four hook types).
- SPEC-V3-HOOKS-003 (once-hook dedup key shares `{shell}\0{command}\0{if}` format).

## 10. Traceability (추적성)

- Theme: master-v3 Section 3.1 (Theme 1 — Hook Protocol v2) matcher + if condition sub-feature.
- Gap rows: gm#8 (High — if condition with permission-rule syntax), gm#13 (Medium — matcher patterns), gm#14 (Medium — hook dedup).
- BC-ID: None (purely additive; existing hooks without matcher/if still work).
- Wave 1 sources: findings-wave1-hooks-commands.md §3.8 (matcher semantics + matchQuery table), §3.9 (dedup keys), §14.3 (feature table), §12 (source references including `utils/hooks.ts:1346-1381, 1390-1421, 1723-1801`).
- Priority: P1 High (unlocks cost-efficient hook dispatch; ships Phase 2).
