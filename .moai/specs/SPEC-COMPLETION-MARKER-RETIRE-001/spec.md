---
id: SPEC-COMPLETION-MARKER-RETIRE-001
title: "Completion-Marker Feature + Dormant Persistent-Mode Subsystem Retirement"
version: "0.1.0"
status: in-progress
created: 2026-06-11
updated: 2026-06-15
author: manager-spec
priority: P2
phase: "v3.6.0"
module: "internal/hook,internal/hook/lifecycle,internal/config,pkg/models,internal/template/templates,.claude/output-styles,.claude/skills,docs-site"
lifecycle: spec-anchored
tags: "retirement, completion-marker, persistent-mode, dead-config, output-style, hooks, docs-site, dormant-subsystem, v3r6"
era: V3R6
tier: M
---

# SPEC-COMPLETION-MARKER-RETIRE-001 — Completion-Marker Feature + Dormant Persistent-Mode Subsystem Retirement

## §A. Purpose

Retire, as a single coupled unit, (1) the MoAI completion markers `<moai>DONE</moai>` / `<moai>COMPLETE</moai>` and (2) the entire dormant persistent-mode subsystem. The two are one inseparable dead/dormant unit: the completion markers are the only signal that deactivates persistent-mode, and persistent-mode is never activated in production. Removing the markers while leaving persistent-mode scaffolding in place would be strictly worse — the scaffolding would retain no deactivation path. The user has confirmed the premise that **persistent-mode will NOT be revived**; therefore both are retired together.

## §B. Problem Statement

The completion-marker / persistent-mode pair exhibits three distinct decay symptoms, all independently verified by direct grep plus a 5-lens read-only workflow audit.

### B.1 — Three-layer dead config (markers are configurable but unwired)

The completion markers are defined in three layers, but only the Go hardcode is on a live path:

1. **LIVE hardcode** — `internal/hook/stop.go:16-19` defines `defaultCompletionMarkers = []string{"<moai>DONE</moai>", "<moai>COMPLETE</moai>"}`. This is the only definition wired into a running handler.
2. **DEAD config struct** — `internal/config/types.go:344-354` defines `CompletionConfig` / `MarkersConfig`; defaults at `internal/config/defaults.go:355-361`. No production code path injects these struct values into the Stop handler.
3. **DEAD YAML** — `.moai/config/sections/workflow.yaml:7-11` and the template `internal/template/templates/.moai/config/sections/workflow.yaml:12-16` expose `completion.markers.{done,complete}`. Editing this YAML has zero runtime effect.

Additionally `pkg/models/config.go:187` carries a `LogCompletionMarkers bool` field (`LSPStateLogging`) that no production code reads.

The "configurable-but-unwired" YAML is misleading debt: a user editing `workflow.completion.markers` reasonably expects the change to take effect, but the production Stop handler always uses the hardcoded defaults via `internal/cli/deps.go:159 NewStopHandler()` (which never calls the config-aware `NewStopHandlerWithMarkers`).

### B.2 — Dormant persistent-mode subsystem (never activated in production)

`internal/hook/lifecycle/persistent_mode.go` implements an activate/deactivate/check state machine over `.moai/state/persistent-mode.json`. The activation entry-point `ActivatePersistentMode()` (`persistent_mode.go:28`) has **zero production callers** — every grep match for it lives in `_test.go`. Because the file is never written in production:

- `mode` is always `nil` in the Stop handler.
- The deactivation branch (`stop.go:78-82`) and the stop-blocking branch (`stop.go:89-96`) are unreachable dead paths.
- The PreCompact handler's read of the same file (`internal/hook/compact.go:97-107` via `readPersistentMode` at `compact.go:142-159`) always returns an empty section.

### B.3 — Markers' only live effect is an observation-only log line

Today, emitting a completion marker produces exactly one observable runtime effect: the observation-only log line at `stop.go:109-121` (`slog.Info("completion marker detected", ...)`). The Stop hook always returns an empty `HookOutput{}` (allow) regardless. The marker does not change control flow, does not block stop, and does not deactivate anything (because nothing is active). The log line is the entire surviving runtime behavior.

### B.4 — Prompt-layer behavioral contract (the user-visible signal)

Separate from the runtime, the completion markers also exist as a **prompt-layer behavioral contract** with two distinct roles:

1. **Completion / Session-Handoff terminal token** — the orchestrator emits `<moai>DONE</moai>` / `<moai>COMPLETE</moai>` as the terminal token of the Completion Report / Session Handoff. The marker emission is instructed by 8 consumer files (live `.claude/` tree + template mirror; re-derived by live grep, see §C / acceptance.md AC-CMR-008):
   - `.claude/output-styles/moai/moai.md` (the XML-exception rule line + §6/§7/§8 handoff usages)
   - `.claude/output-styles/moai/einstein.md` (literal list + handoff usage)
   - `.claude/skills/moai/SKILL.md` (line 323 — "add the appropriate completion marker")
   - `.claude/skills/moai/workflows/loop.md` (the `/moai loop` exit-signal — see role 2 below)
   - `.claude/skills/moai/workflows/release.md` (line 731 terminal token)
   - `.claude/rules/moai/workflow/spec-workflow.md` (lines 281-282 marker glossary)
   - `.claude/agents/local/release-update-specialist.md` (line 334 `Emit: <moai>DONE</moai>`)
   - Template mirrors under `internal/template/templates/.claude/...` for the template-managed subset (`output-styles/moai/{moai,einstein}.md`, `skills/moai/SKILL.md`, `skills/moai/workflows/loop.md`, `rules/moai/workflow/spec-workflow.md`). Note: `release.md` and `release-update-specialist.md` are dev-only (CLAUDE.local.md §21 / §24 `.claude/agents/local/`) and have NO template mirror — they are edited in the live tree only. The `agent-memory/plan-auditor/` hit is a memory artifact, not a consumer (excluded).

   The actual `moai.md` rule wording (line 48) is: `- [HARD] **No XML tags in user-facing output** — except completion markers <moai>DONE</moai> / <moai>COMPLETE</moai>`. The marker is an **EXCEPTION carved out of** an existing "no XML" rule (NOT, as an earlier draft mis-stated, the "sole permitted XML"). A second rule at line 729 repeats the exception: `User-facing output: Markdown only, never raw XML (except <moai> markers)`. Both exception clauses must be removed (see REQ-CMR-011).

2. **`/moai loop` (Ralph Engine) live loop-termination signal** — `loop.md:64-66` Step 1 ("Completion Check") reads the marker from the previous iteration's response and exits the loop on detection; `loop.md:100` ("Prompt user to add completion marker or continue") and `loop.md:157` (Completion Conditions: "Completion marker detected in response") reinforce this. This is the crux of REQ-CMR-015: unlike the dormant runtime path, this is a **LIVE prompt-layer loop-exit token**. Dropping the marker without a replacement would leave `/moai loop` with no explicit success-exit signal (it would rely only on the other 4 exit conditions: zero-errors+tests+coverage, max-iterations, memory-pressure, user-interruption). The replacement signal is the central design decision (plan.md §C).

This prompt-layer surface is the user-visible "completion signal" UX layer and the loop-control contract; it is resolved as the central design decision (see plan.md §C, defaulting to Option (a) pending GATE-2).

## §C. Requirements (GEARS)

### Runtime removal (Go)

- **REQ-CMR-001** (Ubiquitous): The Stop handler shall no longer reference completion markers — the `defaultCompletionMarkers` variable, the `hasCompletionMarker` method, the marker-detection loop, and the persistent-mode branches shall be removed from `internal/hook/stop.go`.
- **REQ-CMR-002** (Ubiquitous): The persistent-mode subsystem (`internal/hook/lifecycle/persistent_mode.go`) shall be retired in its entirety, including `PersistentMode`, `ActivatePersistentMode`, `DeactivatePersistentMode`, `CheckPersistentMode`, and `IsExpired`.
- **REQ-CMR-003** (Ubiquitous): The PreCompact handler shall no longer read `.moai/state/persistent-mode.json` — the `readPersistentMode` function and its "Execution Mode" section assembly in `internal/hook/compact.go` shall be removed.
- **REQ-CMR-004** (State-driven): While the Stop hook processes a session-end event, the handler shall preserve all non-marker behavior — `stop_hook_active` short-circuit, telemetry pruning, reflective-learning analysis, and the empty-`HookOutput{}` allow return — unchanged. (Verified by AC-CMR-001 second command: surviving `TestStopHandler_Handle*` tests pass after removal.)
- **REQ-CMR-005** (Ubiquitous): The Stop-handler constructors shall be reduced to a single marker-free `NewStopHandler()`; the `NewStopHandlerWithMarkers(markers []string)` constructor shall be removed.

### Config struct + YAML removal

- **REQ-CMR-006** (Ubiquitous): The `CompletionConfig` and `MarkersConfig` types shall be removed from `internal/config/types.go`, the `Completion` field shall be removed from `WorkflowConfig`, and the corresponding defaults shall be removed from `internal/config/defaults.go` `NewDefaultWorkflowConfig`.
- **REQ-CMR-007** (Ubiquitous): The `completion:` block shall be removed from both `.moai/config/sections/workflow.yaml` and `internal/template/templates/.moai/config/sections/workflow.yaml`, keeping struct↔YAML symmetry so the config loader-completeness and struct-YAML-symmetry CI guards remain green.
- **REQ-CMR-008** (Ubiquitous): The `LogCompletionMarkers` field shall be removed from `LSPStateLogging` in `pkg/models/config.go`.

### Test retirement

- **REQ-CMR-009** (Event-driven): When the test suite runs after removal, every test genuinely coupled to completion markers or the persistent-MODE subsystem shall have been retired or rewritten so the full suite passes. The complete coupled set was re-derived by live grep (`grep -rln 'persistent\|Persistent\|<moai>\|completionMarker\|defaultCompletionMarkers\|DetectInOutput\|NewStopHandlerWithMarkers\|CompletionConfig\|MarkersConfig\|LogCompletionMarkers' --include='*_test.go' internal/ pkg/`) and classified as follows:

  **DELETE — whole file (asserts only marker/persistent-mode behavior):**
  - `internal/hook/stop_completion_test.go` (156 LOC, 4 tests — all marker-detection)
  - `internal/hook/lifecycle/persistent_mode_test.go` (227 LOC — the retired subsystem's own tests)

  **DELETE — specific functions (rewrite-not-applicable; the asserted behavior is removed):**
  - `internal/hook/stop_test.go`: 6 persistent-mode functions — `TestStopHandler_PersistentMode` (128), `_CompletionMarker` (161), `_Expired` (205), `_Inactive` (253), `_NoFile` (288), `_StopHookActive_OverridesPersistentMode` (316), plus all `defaultCompletionMarkers` / `NewStopHandlerWithMarkers` usages. PRESERVE the 4 non-marker tests (`TestStopHandler_EventType` 18, `_Handle` 28, `_Handle_StopHookActive` 76, `_Handle_StopHookNotActive` 101).
  - `internal/hook/compact_test.go`: `TestCompactHandler_Handle_ReadsPersistentMode` (179). PRESERVE the worktrees-section tests.
  - `internal/hook/compact_coverage_test.go`: 3 functions — `TestReadPersistentMode_MissingFile` (63), `_ValidJSON` (73), `_MalformedJSON` (90). PRESERVE `TestReadWorktrees_*` (13/24/46). **(This file was the auditor's D2 omission — it WILL compile-break after `persistent_mode.go` removal because it calls `readPersistentMode`.)**

  **EDIT — remove specific assertions (file survives, adjacent assertions preserved):**
  - `internal/config/defaults_test.go`: remove **3** lines (not 2) — `Completion.DetectInOutput` (394), `Completion.Markers.Done` (437), `Completion.Markers.Complete` (438) — and reduce the "AC-WSE-007 36-assertion oracle" comment count by 3.
  - `internal/config/workflow_nested_test.go`: remove the `Completion.Markers.Done` / `Complete` assertion block (62-66) from the AC-WSE-003 oracle.
  - `internal/config/types_test.go`: remove the 3 field-reachability rows (321 `Completion/DetectInOutput`, 322 `Completion/Markers/Complete`, 323 `Completion/Markers/Done`) that break once the struct fields are dropped.

  **EXCLUDED as false positives** (substring collisions on `PersistentPreRunE` / `persistent flag` / `oauth-token-persistent` / `persistent contention` — none reference completion markers or the persistent-MODE subsystem; MUST NOT be touched): `internal/cli/coverage_improvement_test.go`, `internal/cli/github_test.go`, `internal/cli/oauth_token_preservation_test.go`, `internal/hook/handoff/persist_test.go`.

### Output-style + template + prompt-layer

- **REQ-CMR-010** (Ubiquitous): The marker-emission instruction shall be removed from the complete consumer set (re-derived by live grep `grep -rln '<moai>\|moai>DONE\|moai>COMPLETE' .claude/ internal/template/templates/.claude/`), and the Completion Report / Session Handoff shall signal completion via the banner / prose alone. The complete set is the 8 live-tree consumers enumerated in §B.4 role 1 (`output-styles/moai/{moai,einstein}.md`, `skills/moai/SKILL.md`, `skills/moai/workflows/{loop,release}.md`, `rules/moai/workflow/spec-workflow.md`, `agents/local/release-update-specialist.md`) PLUS their template mirrors under `internal/template/templates/.claude/` for the template-managed subset. **Template scope (D6) explicitly INCLUDES `.claude/agents/` namespace coverage** — but note the only agent consumer (`release-update-specialist.md`) is dev-only and has no template mirror (CLAUDE.local.md §21/§24); the template-mirrored consumers are the 5 in `output-styles/skills/rules`. Verified by AC-CMR-008 across BOTH trees.
- **REQ-CMR-011** (Ubiquitous): The two marker-exception clauses in `moai.md` shall be removed so the "No XML tags in user-facing output" rule has NO marker carve-out: (a) line 48 `- [HARD] **No XML tags in user-facing output** — except completion markers <moai>DONE</moai> / <moai>COMPLETE</moai>` shall drop the `— except completion markers ...` clause; (b) line 729 `User-facing output: Markdown only, never raw XML (except <moai> markers)` shall drop the `(except <moai> markers)` clause. (The markers are an EXCEPTION carved out of an existing rule, not the "sole permitted XML" — D9 correction.)
- **REQ-CMR-012** (Event-driven): When a template source file under `internal/template/templates/` is edited as part of this retirement, `make build` shall be run so the embedded template (`go:embed`) reflects the change, and template↔mirror byte-parity shall be preserved for any SSOT-mirrored rule files.

### Documentation (docs-site 4-locale)

- **REQ-CMR-013** (Ubiquitous): The public docs-site completion-marker references shall be removed or rewritten in all four locales (en / ko / ja / zh) in parallel, so no locale lags another — covering `utility-commands/moai.md`, `advanced/hooks-guide.md`, and the auxiliary EN files that mention the marker (`utility-commands/moai-loop.md`, `getting-started/introduction.md`, `core-concepts/what-is-moai-adk.md`) with their locale siblings where present.

### `/moai loop` termination-signal replacement (the live prompt-layer consumer — D8)

- **REQ-CMR-015** (Ubiquitous): Because `/moai loop` (Ralph Engine) uses the completion marker as a LIVE prompt-layer loop-exit signal (`loop.md:64-66` Step 1, `:100`, `:157`; see §B.4 role 2), the marker-emission removal shall be accompanied by a concrete replacement loop-termination signal so `/moai loop` retains an explicit success-exit path. Per the §C design decision (default Option (a), pending GATE-2): the replacement is an **explicit natural-language completion sentence** that the Ralph Engine loop evaluator keys on (e.g., the Completion Report banner line / a sentence such as "All loop completion conditions satisfied; exiting loop."). The `loop.md` Step 1 Completion Check and the §Completion Conditions list shall be rewritten to reference this natural-language signal instead of `<moai>DONE</moai>` / `<moai>COMPLETE</moai>`, with the other 4 exit conditions (zero-errors+tests+coverage, max-iterations, memory-pressure, user-interruption) preserved unchanged. This requirement is user-confirmable at GATE-2; if the user selects Option (b) (non-marker terminal token) instead, the replacement token doubles as the loop-exit signal and this REQ is satisfied by that token.

### SPEC traceability

- **REQ-CMR-014** (Ubiquitous): The SPEC shall record that `SPEC-PERSIST-001` (origin of the marker→deactivate contract, REQ-PERSIST-002) and `SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001` (origin of the AC-WSE-003 / AC-WSE-007 marker oracle assertions) are superseded with respect to the completion-marker / persistent-mode surface by this retirement, and that their acceptance oracles for the marker values are being inverted (asserting absence rather than presence).

## §D. Acceptance Criteria

See `acceptance.md` for the full AC-CMR-* enumeration with verifiable commands.

## §E. Dependencies & Traceability

- **Supersedes (partial, marker/persistent-mode surface only)**: `SPEC-PERSIST-001`, `SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001`.
- **Related**: `SPEC-V3R6-SEQ-THINKING-RETIRE-001` and `SPEC-LSPMCP-RETIRE-001` (precedent retirement SPECs — convention reference for `-RETIRE-` naming, dead-config removal, and 4-locale docs sync).

## §F. Exclusions (What NOT to Build)

This section enumerates work explicitly out of scope. Heading uses h3 + dash bullets per `moai spec lint --strict` `OutOfScopeRule`.

### Out of Scope

- Reviving, re-wiring, or replacing the persistent-mode SUBSYSTEM — the user has confirmed persistent-mode will NOT be revived; no replacement `.moai/state/persistent-mode.json` state machine or stop-blocking runtime path is built. (Distinct from REQ-CMR-015, which replaces only the `/moai loop` prompt-layer EXIT SENTENCE, not the persistent-mode runtime.)
- Introducing a new XML / machine-greppable terminal token in transcripts — under the recommended Option (a) the marker is dropped and replaced by a natural-language completion sentence (REQ-CMR-015), NOT a new greppable token. Building a structured/greppable replacement token is excluded unless the user selects Option (b) at GATE-2 (in which case it is in scope per that decision).
- Changing the Stop hook's allow/block decision semantics — the handler already always returns allow; this SPEC removes dead branches without altering the surviving decision.
- Refactoring unrelated hook handlers, the telemetry pruning logic, or the reflective-learning analysis in `stop.go` — only marker/persistent-mode code is touched (Surgical Changes).
- Modifying the `.moai/state/worktrees.json` reader in `compact.go` (`readWorktrees`) — only `readPersistentMode` is removed; the worktrees section is preserved.
- Removing or renaming other `workflow.yaml` sections (`auto_clear`, `loop_prevention`, `team`, `token_budget`, `worktree`) — only the `completion:` block is removed.
- Touching the stale transient agent worktree at `.claude/worktrees/agent-*/` — those are runtime worktree artifacts, not source of truth.
- Adding new lint rules or CI guards to prevent marker re-introduction — out of scope for this retirement; a follow-up guard SPEC may be filed separately if desired.
