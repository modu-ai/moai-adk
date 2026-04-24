---
id: SPEC-V3R2-RT-001
title: "Hook JSON-OR-ExitCode Dual Protocol"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: GOOS
priority: P0 Critical
phase: "v3.0.0 — Phase 2 — Runtime Hardening"
module: "internal/hook/"
dependencies:
  - SPEC-V3R2-CON-001
  - SPEC-V3R2-RT-005
bc_id: [BC-V3R2-001]
related_principle: [P8 Hook Output = JSON Protocol, P2 ACI, P6 Permission Bubble]
related_pattern: [T-5, T-1, S-1]
related_problem: [P-H05, P-H19, P-C01]
related_theme: "Layer 3: Runtime"
breaking: true
lifecycle: spec-anchored
tags: "hook, protocol, json, v3r2, breaking, runtime"
---

# SPEC-V3R2-RT-001: Hook JSON-OR-ExitCode Dual Protocol

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-23 | GOOS | Initial v3 Round-2 draft from Wave 3 synthesis. Supersedes SPEC-V3-HOOKS-001 with scope narrowed to protocol semantics (separates handler coverage into SPEC-V3R2-RT-006 and settings provenance into SPEC-V3R2-RT-005). |

---

## 1. Goal (목적)

Adopt Claude Code's JSON-OR-ExitCode hook output contract as moai's native protocol for all 27 registered hook events. Hook handlers in `internal/hook/*.go` emit a typed `HookResponse` JSON payload on stdout; when stdout is empty or not valid JSON, the runtime falls back to legacy exit-code semantics. This unlocks five programmable capabilities that moai cannot express through exit codes alone: inject model-turn context (`additionalContext`), rewrite tool input mid-turn (`updatedInput`), declare permission decisions (`permissionDecision`), surface user-visible notifications (`systemMessage`), and block teammate idle progression (`continue: false`). Retry hints (`retry`) and file watchlists (`watchPaths`) round out the schema so every future Runtime SPEC (permission stack, sandbox, session state) has a single protocol to plug into. Master §5.4 names this as the structural dependency that blocks programmable Sprint Contract injection, MX tag emission, and config-reload triggers.

The dual protocol is deliberate: master §8 BC-V3R2-001 commits to backward-compatible rollout via wrappers-unchanged + handlers-rewritten so the 26 shell wrappers under `.claude/hooks/moai/` continue to function during the alpha.2 → rc.2 deprecation window. Removal of exit-code-only parsing is deferred to v4.0.

## 2. Scope (범위)

In-scope:

- Typed `HookResponse` Go struct in `internal/hook/response.go` exposing the 7 canonical fields from master §4.3 Layer 3 type block: `AdditionalContext string`, `PermissionDecision PermissionDecision`, `UpdatedInput map[string]any`, `SystemMessage string`, `Continue *bool`, `WatchPaths []string`, `Retry *RetryHint`.
- `PermissionDecision` string enum with values `"allow" | "ask" | "deny"` (wired via SPEC-V3R2-RT-002 resolver but declared here as the protocol field type).
- `HookSpecificOutput` discriminator pattern: per-event Go variant types (PreToolUse, PostToolUse, SessionStart, SubagentStop, ConfigChange, InstructionsLoaded, FileChanged, PreCompact, TeammateIdle, TaskCompleted, PostToolUseFailure, UserPromptSubmit, Stop, StopFailure, Notification, WorktreeCreate, WorktreeRemove, CwdChanged, PermissionRequest, PermissionDenied, SessionEnd, SubagentStart, PostCompact, Setup, Elicitation, ElicitationResult, TaskCreated) implementing a `HookEventName() string` contract.
- Dual-parse shim in `internal/hook/protocol.go`: attempt `encoding/json` unmarshal into `HookResponse` first; on parse failure or empty stdout, synthesize a legacy response from process exit code (0 → allow, 2 → deny with stderr as reason, other non-zero → deny + stderr as user message).
- Strict schema validation via `go-playground/validator/v10` tags (delivered in SPEC-V3R2-SCH-001 Constitution phase; this SPEC consumes the validator only).
- Context-injection wiring: `AdditionalContext` text is appended to the next model turn's system-context block when emitted by SessionStart, UserPromptSubmit, PreToolUse, or PostToolUse events.
- Input-mutation wiring: `UpdatedInput` replaces the pending tool input when PreToolUse returns it (relied on by SPEC-V3R2-RT-002 for permission-layer input rewriting).
- `continue: false` escalates to teammate-idle blocker (relied on by SPEC-V3R2-HRN-002 Sprint Contract quality-gate).
- One-shot deprecation warning: on first legacy exit-code-only output per session, emit a `systemMessage` pointing to the v3 migration guide. Gated by `MOAI_HOOK_LEGACY=1` opt-out for CI and air-gapped installs.
- Opt-in strict mode: `.moai/config/sections/system.yaml` key `hook.strict_mode: true` rejects exit-code-only output entirely with `HookProtocolLegacyRejected` error.

Out-of-scope (addressed by other SPECs):

- Handler completeness — which of the 27 events actually have business logic — SPEC-V3R2-RT-006.
- Permission stack 8-source resolution — SPEC-V3R2-RT-002.
- Sandbox execution — SPEC-V3R2-RT-003.
- Typed session state (`Checkpoint`, `BlockerReport`) — SPEC-V3R2-RT-004.
- Multi-layer settings with provenance (reader for config sources) — SPEC-V3R2-RT-005.
- Hardcoded path fix in shell wrappers — SPEC-V3R2-RT-007.
- Hook type extension (`prompt`/`agent`/`http` beyond `command`) — deferred to v3.1+ per master §13.
- @MX tag injection via PostToolUse `additionalContext` — protocol declared here, semantics in SPEC-V3R2-SPC-002.

## 3. Environment (환경)

Current moai-adk state (from research Wave 1-2):

- `internal/hook/types.go:167-311` defines a partial `HookOutput` struct missing the `AdditionalContext`, `WatchPaths`, `PermissionDecision`, and typed `HookSpecificOutput` union fields per r6-commands-hooks-style-rules.md §2.2.
- All 26 shell wrappers in `.claude/hooks/moai/` emit exit codes only; r6 §2.1 confirms no wrapper writes structured JSON to stdout today.
- 10 of 27 handlers are logging-only no-ops per r6 §A Hook Coverage Matrix (partial coverage 37%). Handler completeness is tracked in SPEC-V3R2-RT-006 but many of those stubs exist precisely because exit codes leave them nothing useful to emit — upgrading the protocol unblocks their upgrade path.
- `internal/hook/registry.go:18-325` dispatches by `EventType` with no consumption of richer fields.

Claude Code reference (master §8 BC-V3R2-001 + r3 §2 Decision 5):

- CC's `HookJSONOutput` is a discriminated union keyed on `hookEventName` (r3 §2 Decision 5, r3 §4 Adopt 4).
- CC injects `additionalContext` into model turn context assembly and applies `updatedInput` to pending tool calls (r3 §2 Decision 5).
- CC's exit-code fallback remains active in 2026.x for scripts without JSON producers (r3 §2 Decision 5).

Affected modules:

- `internal/hook/types.go` — struct expansion for typed HookResponse.
- `internal/hook/response.go` — new file, typed response + per-event variants.
- `internal/hook/protocol.go` — new file, dual-parse reader + exit-code synthesizer.
- `internal/hook/registry.go` — downstream consumption of rich fields.
- `.claude/hooks/moai/*.sh` — wrappers remain unchanged (they forward stdin/stdout transparently).
- `.moai/config/sections/system.yaml` — add `hook.strict_mode` key.

## 4. Assumptions (가정)

- Claude Code's HookJSONOutput schema remains stable for 2.1.111+ per master §10 R11 risk-register.
- `go-playground/validator/v10` is available via SPEC-V3R2-SCH-001 (Phase 1 Constitution).
- Existing 26 shell hook wrappers emit valid legacy exit-code-only output today; dual-parse fallback is the migration mechanism (r6 §2.1).
- External plugin-hook authors exist in small numbers; grace period via legacy shim is sufficient per master §8 BC-V3R2-001.
- Go's `encoding/json` with `json.RawMessage` handles `hookSpecificOutput` discrimination via two-pass decode (first pass reads `hookEventName`, second pass decodes into the concrete variant).
- Stdout buffering: handlers must print the complete JSON object in one write before process exit; partial writes are treated as parse failure and fall back to exit code.
- The 26 shell wrappers pipe child-process stdout 1:1 to Claude Code; no intermediary mutates the bytes.
- 1500ms SessionEnd timeout noted in r3 §3 Technical Debt item 5 still applies; hook authors targeting SessionEnd must not produce >64KB payloads.

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

- REQ-V3R2-RT-001-001: The `HookResponse` Go struct SHALL expose every top-level field named in master §4.3 Layer 3 type block (`AdditionalContext`, `PermissionDecision`, `UpdatedInput`, `SystemMessage`, `Continue`, `WatchPaths`, `Retry`) with `json:",omitempty"` tags matching Claude Code's HookJSONOutput schema byte-for-byte.
- REQ-V3R2-RT-001-002: The `PermissionDecision` type SHALL be a string enum with values `"allow"`, `"ask"`, `"deny"`, and the zero value `""` representing "no opinion" (delegate to stack resolver).
- REQ-V3R2-RT-001-003: The system SHALL provide concrete per-event variant types for all 27 Claude Code hook events enumerated in master §7.3 Table.
- REQ-V3R2-RT-001-004: Every variant type SHALL implement the `HookEventName() string` method returning the canonical Claude Code event name.
- REQ-V3R2-RT-001-005: The dual-parse reader SHALL attempt `encoding/json` unmarshal of stdout into `HookResponse` before consulting exit code.
- REQ-V3R2-RT-001-006: Every received `HookResponse` SHALL pass `validator/v10` schema validation before effects are enqueued downstream.
- REQ-V3R2-RT-001-007: The system SHALL expose a single deprecation warning banner per session per project when any legacy exit-code-only hook fires.

### 5.2 Event-Driven Requirements

- REQ-V3R2-RT-001-010: WHEN a hook wrapper writes a JSON object to stdout and the JSON parses successfully, the protocol reader SHALL populate `HookResponse` directly and bypass exit-code inspection.
- REQ-V3R2-RT-001-011: WHEN JSON parse fails or stdout contains only whitespace, the protocol reader SHALL fall back to exit-code synthesis (0 → allow, 2 → deny with stderr as reason, other non-zero → user-visible systemMessage).
- REQ-V3R2-RT-001-012: WHEN a hook returns `AdditionalContext` on SessionStart, UserPromptSubmit, PreToolUse, or PostToolUse, the system SHALL append the text to the next model turn's system-context block in the order the hooks fired.
- REQ-V3R2-RT-001-013: WHEN a PreToolUse hook returns `UpdatedInput` with a non-nil map, the system SHALL replace the pending tool-call input with the rewritten map before dispatching the tool.
- REQ-V3R2-RT-001-014: WHEN any hook returns `Continue: false`, the system SHALL halt the current turn and, for SubagentStop, block the teammate from idling until the orchestrator resolves the blocker.
- REQ-V3R2-RT-001-015: WHEN a hook returns `SystemMessage`, the system SHALL emit it to the user-visible status stream exactly once per hook invocation.
- REQ-V3R2-RT-001-016: WHEN a PostToolUse hook returns `AdditionalContext` containing `@MX:NOTE`, `@MX:WARN`, `@MX:ANCHOR`, `@MX:TODO`, or `@MX:LEGACY` markers, the system SHALL route the text into the @MX tag ingestion path defined in SPEC-V3R2-SPC-002 (integration point only; semantics in that SPEC).

### 5.3 State-Driven Requirements

- REQ-V3R2-RT-001-020: WHILE the environment variable `MOAI_HOOK_LEGACY=1` is set, the deprecation-warning banner SHALL be suppressed but dual-parse SHALL continue to accept both output forms.
- REQ-V3R2-RT-001-021: WHILE `.moai/config/sections/system.yaml` key `hook.strict_mode` is `true`, the system SHALL reject any hook whose stdout fails JSON parse with error `HookProtocolLegacyRejected` (halts the turn with user-visible message).
- REQ-V3R2-RT-001-022: WHILE the hook payload size exceeds 64 KiB, the system SHALL truncate `AdditionalContext` to 64 KiB and emit `SystemMessage: "AdditionalContext truncated to 64 KiB budget"`.

### 5.4 Optional Features

- REQ-V3R2-RT-001-030: WHERE a hook wrapper declares `api_version: 2` in its frontmatter (shell comment `# moai-hook-api-version: 2`), the system SHALL skip exit-code fallback for that wrapper even in non-strict mode.
- REQ-V3R2-RT-001-031: WHERE a hook returns the `Retry` field with a non-nil `RetryHint{Attempts int, Backoff string}`, the orchestrator MAY re-dispatch the hook up to the declared attempt count with exponential backoff bounded by `Backoff`.
- REQ-V3R2-RT-001-032: WHERE `WatchPaths` is returned by SessionStart, the system SHALL register file-system watches on the paths and fire `FileChanged` hook events when they change.

### 5.5 Unwanted Behavior

- REQ-V3R2-RT-001-040: IF a hook returns `HookSpecificOutput.HookEventName` that does not match the dispatched event, THEN the system SHALL reject the response with error `HookSpecificOutputMismatch`, log the mismatch to `.moai/logs/hook.log`, and treat the hook as failed.
- REQ-V3R2-RT-001-041: IF a hook returns both a `PermissionDecision` and a non-empty `UpdatedInput` for a PreToolUse event, THEN the system SHALL apply `UpdatedInput` first, then apply `PermissionDecision` against the updated input.
- REQ-V3R2-RT-001-042: IF a hook writes malformed JSON (parse error) AND stderr is also empty AND exit code is 0, THEN the system SHALL treat the hook as allow-continue with no side effects and log a warning to `.moai/logs/hook.log`.

### 5.6 Complex Requirements

- REQ-V3R2-RT-001-050: WHILE `hook.strict_mode: false` AND the hook emits legacy exit-code output, WHEN the session has already emitted the deprecation banner, THEN the system SHALL suppress further banner emissions until the next session but SHALL continue to honor the exit code (allow/deny/user-stderr).
- REQ-V3R2-RT-001-051: WHILE `BC-V3R2-001` is in its deprecation window (v3.0.0 through v3.x), WHEN a plugin-contributed hook emits legacy exit-code output, THEN the system SHALL apply the dual-parse fallback without raising an error regardless of strict-mode.

## 6. Acceptance Criteria (수용 기준)

- AC-V3R2-RT-001-01: Given a PreToolUse hook wrapper writes `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow","updatedInput":{"file_path":"/tmp/x"}}}` to stdout, When the hook fires, Then the tool input is replaced with `{"file_path":"/tmp/x"}` before execution and the resolver receives `permissionDecision: allow`. (maps REQ-V3R2-RT-001-010, -013)
- AC-V3R2-RT-001-02: Given a SessionStart hook returns `{"additionalContext":"ctx","watchPaths":["/abs/.env"]}`, When parsed, Then `HookResponse.AdditionalContext == "ctx"` and `HookResponse.WatchPaths == ["/abs/.env"]` with validator passing. (maps REQ-V3R2-RT-001-001, -012, -032)
- AC-V3R2-RT-001-03: Given a legacy hook wrapper exits with code 2 and stderr `"blocked"`, When dual-parse falls back, Then `HookResponse.PermissionDecision == "deny"` and the reason field contains `"blocked"`. (maps REQ-V3R2-RT-001-011)
- AC-V3R2-RT-001-04: Given `MOAI_HOOK_LEGACY=1` is set, When a legacy wrapper fires, Then no deprecation banner is emitted during that session. (maps REQ-V3R2-RT-001-020)
- AC-V3R2-RT-001-05: Given a PreToolUse hook returns `hookSpecificOutput.hookEventName == "PostToolUse"` (mismatch), When parsed, Then the protocol layer returns `HookSpecificOutputMismatch` error and the hook is treated as failed. (maps REQ-V3R2-RT-001-040)
- AC-V3R2-RT-001-06: Given `.moai/config/sections/system.yaml` has `hook.strict_mode: true`, When a legacy wrapper emits only an exit code (no stdout JSON), Then the system returns `HookProtocolLegacyRejected` and the turn halts. (maps REQ-V3R2-RT-001-021)
- AC-V3R2-RT-001-07: Given a PostToolUse hook returns `AdditionalContext: "@MX:WARN at line 42 — unbounded goroutine"`, When the turn concludes, Then the @MX ingestion path from SPEC-V3R2-SPC-002 receives the marker text. (maps REQ-V3R2-RT-001-016)
- AC-V3R2-RT-001-08: Given a SubagentStop hook returns `Continue: false` with `SystemMessage: "coverage below 85%"`, When evaluated, Then the teammate is prevented from idling and the orchestrator surfaces the blocker via AskUserQuestion. (maps REQ-V3R2-RT-001-014)
- AC-V3R2-RT-001-09: Given a hook returns `AdditionalContext` of 128 KiB, When the response is consumed, Then `AdditionalContext` is truncated to 64 KiB and `SystemMessage` contains the truncation notice. (maps REQ-V3R2-RT-001-022)
- AC-V3R2-RT-001-10: Given a PreToolUse hook returns both `PermissionDecision: "deny"` and `UpdatedInput: {...}`, When applied, Then `UpdatedInput` is merged into the pending input first and the `deny` decision blocks the tool call with the post-update input shown in the denial message. (maps REQ-V3R2-RT-001-041)
- AC-V3R2-RT-001-11: Given a hook wrapper declares `# moai-hook-api-version: 2` in its shell header, When it exits 0 with no JSON on stdout, Then the protocol reader does NOT fall back to exit-code synthesis and instead treats the empty response as the explicit no-op `HookResponse{}`. (maps REQ-V3R2-RT-001-030)
- AC-V3R2-RT-001-12: Given validator/v10 schema tags are applied to `HookResponse`, When `PermissionDecision` receives value `"yes"` (invalid), Then `Validate()` returns a non-nil error naming the offending field. (maps REQ-V3R2-RT-001-006)
- AC-V3R2-RT-001-13: Given `make build` regenerates embedded templates, When `go test ./internal/hook/... -run TestDualParse` runs, Then all 27 event variant round-trip serialization tests pass (marshal → unmarshal identity). (maps REQ-V3R2-RT-001-003, -004)
- AC-V3R2-RT-001-14: Given a plugin-contributed hook emits legacy exit-code output during a session where `hook.strict_mode: true`, When processed, Then dual-parse fallback is applied with no `HookProtocolLegacyRejected` error and the plugin origin is logged via `source: plugin` provenance from SPEC-V3R2-RT-005. (maps REQ-V3R2-RT-001-051)
- AC-V3R2-RT-001-15: Given the deprecation banner has already fired once in a session, When a second legacy-only hook emits exit-code output, Then no banner is printed but the exit code still produces the correct `PermissionDecision` synthesis. (maps REQ-V3R2-RT-001-050)

## 7. Constraints (제약)

- Technical: Go 1.22+; no new top-level dependencies beyond validator/v10 (SPEC-V3R2-SCH-001). `encoding/json` from stdlib covers the discriminator decode via two-pass pattern.
- Backward compat: exit-code fallback MUST function for the full v3.0.0 → v3.x deprecation window per master §8 BC-V3R2-001; removal deferred to v4.0.
- Platform: macOS / Linux / Windows. Windows hook wrappers use `CLAUDE_ENV_FILE` for env propagation (already present; no new logic).
- Performance: JSON parse of hook stdout MUST NOT add more than 5 ms p99 overhead on top of subprocess wall time for payloads up to 64 KiB.
- Binary size: struct + per-event variants MUST NOT grow `bin/moai` by more than 250 KiB.
- Wrappers unchanged: the 26 shell scripts under `.claude/hooks/moai/` MUST NOT require modification; they continue to forward stdin/stdout.
- SessionEnd timeout ceiling remains at 1500 ms per Claude Code runtime; hook authors targeting SessionEnd must bound payload emission time.

## 8. Risks & Mitigations (리스크 및 완화)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| External plugin-hook authors emit invalid JSON and rely on exit-code fallback indefinitely | M | M | Dual-parse spans the full v3.x minor cycle; `MOAI_HOOK_LEGACY=1` opt-out for CI/air-gapped; `hook.strict_mode: true` opt-in for teams wanting early rejection. |
| Discriminated-union mismatches produce cryptic errors | M | M | REQ-V3R2-RT-001-040 mandates specific `HookSpecificOutputMismatch` error with log trace; `moai doctor hook --validate` surfaces these pre-runtime. |
| Large `AdditionalContext` payloads balloon model token usage | L | M | REQ-V3R2-RT-001-022 caps at 64 KiB with user-visible truncation notice; migration doc publishes best-practice limits. |
| PreToolUse `UpdatedInput` mutation races with permission decision | L | M | REQ-V3R2-RT-001-041 defines deterministic order (input first, then decision). |
| 10 logging-only handlers continue returning no response after protocol upgrade | M | L | SPEC-V3R2-RT-006 explicitly enumerates 27-event business-logic coverage with per-event decisions; this SPEC only owns the wire format. |
| Shell-wrapper JSON forwarding breaks on non-UTF-8 bytes from language-specific test output | L | L | Wrappers already use `cat` semantics; validator/v10 rejects non-UTF-8 strings with clear error naming. |

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3R2-SCH-001 (provides validator/v10 integration).
- SPEC-V3R2-CON-001 (FROZEN-zone codification enables the protocol-is-structural declaration).
- SPEC-V3R2-RT-005 (provides Source tag for plugin-contributed hook provenance referenced in REQ-V3R2-RT-001-051).

### 9.2 Blocks

- SPEC-V3R2-RT-002 (permission stack consumes `PermissionDecision` field from PreToolUse hooks).
- SPEC-V3R2-RT-003 (sandbox enforcement uses `UpdatedInput` to mutate command arguments mid-turn).
- SPEC-V3R2-RT-006 (handler completeness relies on protocol being expressive enough for business logic).
- SPEC-V3R2-SPC-002 (@MX tag autonomous add/update/remove via PostToolUse `additionalContext`).
- SPEC-V3R2-HRN-002 (evaluator fresh-memory injection uses PreToolUse `additionalContext`).

### 9.3 Related

- SPEC-V3R2-MIG-001 (v2→v3 migrator adds validator/v10 import to hook types during migration).
- SPEC-V3R2-CON-003 (constitution consolidation moves hook-protocol text from CLAUDE.md Section 10 into hooks-system.md).

## 10. Traceability (추적성)

- Theme: master §4.3 Layer 3 Runtime; §5.4 Cross-Layer Hook JSON-OR-ExitCode; §8 BC-V3R2-001.
- Principle: P8 (Hook JSON Protocol); secondary P2 (ACI), P6 (Permission Bubble).
- Pattern: T-5 (Hook Dual Protocol), T-1 (ACI structured responses), S-1 (Permission stack).
- Problem: P-H05 (exit-code-only protocol, HIGH), P-H19 (59% partial event coverage, HIGH), P-C01 (no permission bubble, CRITICAL — protocol prerequisite).
- Master Appendix A: Principle P8 → primary SPEC-V3R2-RT-001.
- Master Appendix C: Pattern T-5 → primary SPEC-V3R2-RT-001 (priority 2).
- Wave 1 sources: r3-cc-architecture-reread.md §2 Decision 5 (JSON-OR-exitcode), §4 Adoption Candidate 4 (programmable hook protocol); r6-commands-hooks-style-rules.md §2.2 (handler audit), §A Hook Coverage Matrix.
- Wave 2 sources: design-principles.md P8 (Hook JSON Protocol); pattern-library.md T-5 (priority 2); problem-catalog.md P-H05.
- BC-ID: BC-V3R2-001 (hook handlers migrate to JSON-OR-ExitCode, AUTO migration).
- Priority: P0 Critical — blocks every higher-value Runtime SPEC in v3.0 Phase 2.
