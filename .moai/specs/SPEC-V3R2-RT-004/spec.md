---
id: SPEC-V3R2-RT-004
title: "Typed Session State + Phase Checkpoint"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: GOOS
priority: P1 High
phase: "v3.0.0 — Phase 2 — Runtime Hardening"
module: "internal/session/"
dependencies:
  - SPEC-V3R2-CON-001
  - SPEC-V3R2-RT-005
bc_id: []
related_principle: [P5 Typed State + Durable Checkpoint, P3 Fresh-Context Iteration, P11 File-First Primitives]
related_pattern: [X-3, M-1, R-6]
related_problem: [P-C02, P-C05]
related_theme: "Layer 3: Runtime"
breaking: false
lifecycle: spec-anchored
tags: "session, state, checkpoint, typed, v3r2, runtime, file-first"
---

# SPEC-V3R2-RT-004: Typed Session State + Phase Checkpoint

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-23 | GOOS | Initial v3 Round-2 draft. New SPEC — no v3-legacy predecessor. Addresses P-C02 (no sub-agent context isolation) and P-C05 (no cache-prefix discipline). Non-breaking: adds typed schema on top of existing `.moai/state/` files. |

---

## 1. Goal (목적)

Formalize moai's cross-phase and cross-iteration state as a typed schema with immutable updates checkpointed at every phase boundary. Master §5.6 commits to "file-first state + fresh-context iteration": state lives on disk in `.moai/state/` as human-readable YAML/Markdown/JSON, and the LLM context that consumes the state is ephemeral per Principle 3. This SPEC converts that commitment into concrete Go types (`PhaseState`, `Checkpoint`, `BlockerReport`), a canonical directory layout, provenance tags per source (user / project / local / session / hook), and an `interrupt()`-equivalent blocker pathway that surfaces to AskUserQuestion at the orchestrator.

The dual rationale is structural and safety-oriented:

- **Structural (Principle 5 + 6)**: Every phase boundary (plan → run → sync) writes a checkpoint file that fully captures the inputs the next phase needs. The next phase reads the file, not the prior conversation. This closes problem-catalog.md P-C02 ("No sub-agent context isolation primitive in moai") by making context a file on disk instead of a transcript.
- **Safety (Principle 4 + 11)**: Sprint Contract state (SPEC-V3R2-HRN-002) and evaluator fresh-judgment memory boundaries require a durable substrate that survives agent-level memory resets. Checkpoint files are that substrate.

Master §4.3 Layer 3 type block names `PhaseState{Phase, SPECID, Checkpoint, BlockerRpt, UpdatedAt}` as the canonical type. Master §5.6 enumerates the `.moai/state/` layout: `task-ledger.md` (append-only, Magentic O-3), `progress.md` (Ralph shape), `runs/{iter-id}/prompt.md|response.md|artifacts/`, `sprint-contract.yaml`, `checkpoint-{phase}.json`.

## 2. Scope (범위)

In-scope:

- `Phase` typed enum in `internal/session/state.go` with values `"plan" | "run" | "sync" | "design" | "review" | "fix" | "loop" | "db" | "mx"`.
- `PhaseState` Go struct exposing `Phase Phase`, `SPECID string`, `Checkpoint Checkpoint`, `BlockerRpt *BlockerReport`, `UpdatedAt time.Time`.
- `Checkpoint` typed interface with per-phase concrete variants: `PlanCheckpoint`, `RunCheckpoint`, `SyncCheckpoint` (extensible).
- `BlockerReport` struct for `interrupt()`-equivalent surfacing: `Kind`, `Message`, `Context map[string]any`, `RequestedAction string`, `ProvenanceSource Source`.
- `SessionStore` Go interface with `Checkpoint(phase PhaseState) error`, `Hydrate(phase Phase, specID string) (PhaseState, error)`, `AppendTaskLedger(entry TaskLedgerEntry) error`, `WriteRunArtifact(iterID, name string, body []byte) error`.
- Canonical `.moai/state/` file layout per master §5.6:
  - `task-ledger.md` — append-only Magentic ledger (O-3).
  - `progress.md` — Ralph-shape per iteration summary.
  - `activity.log` — structured event stream.
  - `errors.log` — structured error stream.
  - `runs/{iter-id}/prompt.md`, `runs/{iter-id}/response.md`, `runs/{iter-id}/artifacts/` — Ralph iteration record.
  - `sprint-contract.yaml` — SPEC-V3R2-HRN-002 state (referenced but owned there).
  - `checkpoint-{phase}.json` — typed phase checkpoint.
- Provenance tag on every state mutation: which `Source` tier introduced the update — `user` / `project` / `local` / `session` / `hook` (subset of the 8-tier stack from SPEC-V3R2-RT-005, scoped to state-layer sources).
- Validator/v10 schema tags on all typed checkpoints; `SessionStore.Checkpoint()` validates before writing.
- Crash-resume semantics: `STALE_SECONDS` primitive adopted from Ralph (default 3600s); on hydration, if `UpdatedAt` is older than `STALE_SECONDS`, hydration fails and the user is prompted via AskUserQuestion whether to resume or restart fresh.
- `moai state dump` subcommand prints the current `PhaseState` with provenance per field.
- `moai state show-blocker` prints any unresolved `BlockerReport` and routes acknowledgment back into checkpoint.

Out-of-scope (addressed by other SPECs):

- Settings file multi-layer merge (user/project/local settings resolver) — SPEC-V3R2-RT-005.
- Hook JSON protocol — SPEC-V3R2-RT-001.
- Sprint Contract durable state schema — SPEC-V3R2-HRN-002 (owns `sprint-contract.yaml`).
- Ralph `/moai loop` outer-loop flow — SPEC-V3R2-WF-003.
- Memdir typed taxonomy (user / feedback / project / reference) — SPEC-V3R2-EXT-001 (different layer: memdir is cross-project persistent, session state is per-SPEC).
- Agent `memory:` field in frontmatter — SPEC-V3R2-ORC-001.

## 3. Environment (환경)

Current moai-adk state:

- `.moai/state/` directory exists per CLAUDE.local.md §2 protected directories list; contents are currently ad-hoc markdown without typed schema.
- No Go struct in `internal/config/types.go` describes phase state or checkpoints; r6-commands-hooks-style-rules.md §5.2 confirms 5 yaml sections lack loaders including the ones that would interact with state (constitution, context, interview, design, harness).
- Subagent stateless context is the v2 paradigm per agent-common-protocol.md; no primitive exists to pass typed state across agent boundaries (problem-catalog.md P-C02).
- Prompt assembly order is unspecified per problem-catalog.md P-C05; the system prompt may re-order input between turns, losing Anthropic prompt-cache benefits.
- Ralph `/moai loop` is documented in CLAUDE.md §15 and moai-workflow-loop skill; canonical state files (`progress.md`, `activity.log`, `errors.log`) exist informally but are not Go-typed.

Claude Code reference:

- r3 §1.1 type table: Session has typed `sessionId`, `cwd`, `projectRoot`, `parentSessionId`.
- r3 §2 Decision 2 "Cache-prefix discipline on system prompts": CC freezes `(systemPrompt, userContext, systemContext)` ordering and uses `memoize()` on each piece. moai needs this discipline.
- r3 §4 Adopt 3 "Sub-agent context isolation as a primitive": adopt `{systemPrompt, toolPool, permissionMode, cwd, mcpServers, fileStateCache, transcript}` as typed objects.

Wave 1-2 sources:

- design-principles.md P5: "Typed state + durable checkpoint at phase boundaries" — DSPy signatures, LangGraph checkpointer, MS Agent Framework 1.0 GA April 2026.
- pattern-library.md X-3 (Output Style) provides frontmatter-as-typed-schema precedent, applied here to state files.
- pattern-library.md R-6 (Ralph fresh-context) — canonical file-first layout.

Affected modules:

- `internal/session/state.go` — new file, typed state.
- `internal/session/checkpoint.go` — new file, per-phase concrete variants.
- `internal/session/blocker.go` — new file, BlockerReport + routing.
- `internal/session/store.go` — new file, file-first SessionStore implementation.
- `internal/cli/state.go` — new file, `moai state` subcommand.
- `.moai/state/` — directory layout standardized.
- `internal/config/types.go` — add `SessionConfig` if needed for STALE_SECONDS configuration.

## 4. Assumptions (가정)

- `.moai/state/` is already in the protected-directories list per CLAUDE.local.md §2 and is excluded from template drift checks.
- STALE_SECONDS default of 3600s (1 hour) covers typical developer iteration without over-prompting; configurable via `.moai/config/sections/ralph.yaml` key `stale_seconds`.
- File-first state files are ASCII-safe (YAML, Markdown, JSON); binary artifacts go under `runs/{iter-id}/artifacts/` with no assumption about encoding.
- Provenance tags reuse the `Source` enum subset from SPEC-V3R2-RT-005: `SrcUser`, `SrcProject`, `SrcLocal`, `SrcSession`, `SrcHook`. `SrcPolicy`, `SrcPlugin`, `SrcSkill`, `SrcBuiltin` do not contribute session state directly.
- Crash-resume assumes the filesystem survived the crash (no disk corruption); power-loss scenarios may leave partial writes, handled via atomic rename pattern (write `*.tmp`, `os.Rename` to final).
- Concurrent writes to the same state file from two agents in parallel team mode are prevented by per-agent checkpoint paths (`checkpoint-{phase}-{agent-name}.json`); the orchestrator merges before user-facing display.
- `sprint-contract.yaml` lifecycle is owned by SPEC-V3R2-HRN-002; this SPEC only commits to the file path and the reader/writer interface.
- `BlockerReport` surfacing to AskUserQuestion is routed through the orchestrator only (agent-common-protocol.md HARD rule); subagents write the blocker to disk and stop, never prompt directly.

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

- REQ-V3R2-RT-004-001: The `Phase` type SHALL be a typed string enum with exactly 9 values covering all moai subcommands that write session state: `"plan"`, `"run"`, `"sync"`, `"design"`, `"review"`, `"fix"`, `"loop"`, `"db"`, `"mx"`.
- REQ-V3R2-RT-004-002: Every `PhaseState` instance SHALL include non-empty `Phase`, `SPECID`, `UpdatedAt` fields and optional `Checkpoint`/`BlockerRpt` fields.
- REQ-V3R2-RT-004-003: `Checkpoint` SHALL be a Go interface with per-phase concrete variants (`PlanCheckpoint`, `RunCheckpoint`, `SyncCheckpoint` ship in v3.0; others added per phase as SPECs land).
- REQ-V3R2-RT-004-004: Every state file in `.moai/state/` SHALL be validated against a validator/v10 schema before write; invalid state SHALL cause `Checkpoint()` to return error without mutating disk.
- REQ-V3R2-RT-004-005: Every `PhaseState` mutation SHALL carry a provenance tag naming the `Source` that introduced the update (user / project / local / session / hook).
- REQ-V3R2-RT-004-006: The canonical `.moai/state/` layout SHALL match master §5.6 verbatim (task-ledger.md, progress.md, activity.log, errors.log, runs/{iter-id}/, sprint-contract.yaml, checkpoint-{phase}.json).
- REQ-V3R2-RT-004-007: `moai state dump` SHALL print the current `PhaseState` with per-field provenance.
- REQ-V3R2-RT-004-008: `SessionStore` interface SHALL expose `Checkpoint`, `Hydrate`, `AppendTaskLedger`, `WriteRunArtifact`, `RecordBlocker`, `ResolveBlocker` methods.

### 5.2 Event-Driven Requirements

- REQ-V3R2-RT-004-010: WHEN a phase completes successfully, the orchestrator SHALL call `SessionStore.Checkpoint(PhaseState{Phase, SPECID, Checkpoint, UpdatedAt: now})` before advancing to the next phase.
- REQ-V3R2-RT-004-011: WHEN a phase begins, the orchestrator SHALL call `SessionStore.Hydrate(Phase, SPECID)` to load the prior-phase checkpoint; the checkpoint is the sole input to the new phase's LM context.
- REQ-V3R2-RT-004-012: WHEN a subagent cannot proceed due to missing context or ambiguous user intent, it SHALL call `SessionStore.RecordBlocker(BlockerReport{...})` and return to the orchestrator; the blocker file path SHALL be named `blocker-{phase}-{SPECID}-{timestamp}.json`.
- REQ-V3R2-RT-004-013: WHEN the orchestrator receives a blocker, it SHALL surface the `BlockerReport.Message` and `BlockerReport.RequestedAction` via AskUserQuestion and route the user response to `SessionStore.ResolveBlocker`.
- REQ-V3R2-RT-004-014: WHEN a hydration operation reads a checkpoint file whose `UpdatedAt` is older than `STALE_SECONDS`, the system SHALL return error `CheckpointStale` and the orchestrator SHALL prompt the user via AskUserQuestion whether to resume or restart fresh.
- REQ-V3R2-RT-004-015: WHEN a Ralph `/moai loop` iteration completes, `WriteRunArtifact` SHALL persist `prompt.md`, `response.md`, and any artifact bytes under `runs/{iter-id}/`.

### 5.3 State-Driven Requirements

- REQ-V3R2-RT-004-020: WHILE a `BlockerReport` is outstanding for a given `(Phase, SPECID)`, `SessionStore.Checkpoint` for the same key SHALL refuse to advance the phase with error `BlockerOutstanding`.
- REQ-V3R2-RT-004-021: WHILE two parallel team-mode agents write per-agent checkpoint files, the orchestrator-side merge SHALL produce a single `PhaseState` with `ProvenanceSource: SrcSession` and list per-agent origin paths in an audit subfield.
- REQ-V3R2-RT-004-022: WHILE `.moai/config/sections/ralph.yaml` key `stale_seconds: <value>` is set, the hydration staleness check SHALL use `<value>` instead of the 3600s default.

### 5.4 Optional Features

- REQ-V3R2-RT-004-030: WHERE `moai state dump --phase run` is invoked, the system SHALL emit only the run-phase checkpoint with its provenance chain.
- REQ-V3R2-RT-004-031: WHERE a project declares `.moai/config/sections/state.yaml` key `retention_days: N`, older `runs/{iter-id}/` directories SHALL be eligible for `moai clean` pruning.
- REQ-V3R2-RT-004-032: WHERE `moai state show-blocker` is invoked AND a blocker is outstanding, the system SHALL print the BlockerReport in human-readable form and exit non-zero.
- REQ-V3R2-RT-004-033: WHERE `--resume` flag is passed to a moai subcommand, the phase hydrator SHALL skip the stale-check AskUserQuestion prompt.

### 5.5 Unwanted Behavior

- REQ-V3R2-RT-004-040: IF a concurrent `Checkpoint()` races with another on the same file, THEN the later write SHALL detect the mid-write state via `flock` (Linux/macOS) or advisory lock file (Windows), retry up to 3 times with 10 ms backoff, and fail with error `CheckpointConcurrent` on repeated loss.
- REQ-V3R2-RT-004-041: IF a checkpoint file fails validator/v10 validation on read (corrupted on disk), THEN hydration SHALL return `CheckpointInvalid` error with the offending field named and the system SHALL prompt via AskUserQuestion to restore from `*.bak` or restart fresh.
- REQ-V3R2-RT-004-042: IF a subagent attempts to call `AskUserQuestion` directly instead of `RecordBlocker`, THEN the agent-common-protocol.md CI lint from SPEC-V3R2-ORC-002 SHALL reject the agent file.
- REQ-V3R2-RT-004-043: IF `runs/{iter-id}/` contains non-UTF-8 artifact bytes AND the artifact type is declared as text, THEN `WriteRunArtifact` SHALL return `ArtifactEncodingMismatch` error.

### 5.6 Complex Requirements

- REQ-V3R2-RT-004-050: WHILE a phase transition is in flight (Checkpoint written but Hydrate of next phase not yet complete), WHEN the user aborts the session (Ctrl-C, terminal close), THEN on next session start the orchestrator SHALL detect the in-flight transition, read the most recent `checkpoint-{phase}.json`, and offer via AskUserQuestion to resume at the next phase or roll back.
- REQ-V3R2-RT-004-051: WHILE team mode is active AND multiple teammates write their own checkpoint fragments, WHEN any teammate records a blocker, THEN the orchestrator SHALL pause the phase transition for all teammates and surface the blocker at the parent terminal (per bubble-mode semantics from SPEC-V3R2-RT-002).

## 6. Acceptance Criteria (수용 기준)

- AC-V3R2-RT-004-01: Given a plan phase completes with `PlanCheckpoint{TaskDAG, RiskSummary, SelectedHarness}`, When `Checkpoint()` is called, Then `.moai/state/checkpoint-plan.json` is written atomically and validator-valid. (maps REQ-V3R2-RT-004-002, -004, -010)
- AC-V3R2-RT-004-02: Given run phase begins for `SPEC-V3R2-RT-004`, When `Hydrate("run", "SPEC-V3R2-RT-004")` is called, Then the returned PhaseState contains the prior PlanCheckpoint. (maps REQ-V3R2-RT-004-011)
- AC-V3R2-RT-004-03: Given a subagent encounters missing acceptance-criteria context, When it records a BlockerReport, Then `.moai/state/blocker-run-SPEC-V3R2-RT-004-*.json` exists and `RecordBlocker()` returned nil. (maps REQ-V3R2-RT-004-012)
- AC-V3R2-RT-004-04: Given an outstanding blocker exists for `(run, SPEC-V3R2-RT-004)`, When Checkpoint() is retried for the same phase, Then the write fails with `BlockerOutstanding` error. (maps REQ-V3R2-RT-004-020)
- AC-V3R2-RT-004-05: Given a checkpoint file's `UpdatedAt` is 2 hours old and `stale_seconds: 3600` is default, When Hydrate() reads it, Then the system returns `CheckpointStale` and prompts via AskUserQuestion. (maps REQ-V3R2-RT-004-014)
- AC-V3R2-RT-004-06: Given `--resume` flag is set on `moai run`, When Hydrate() encounters a stale checkpoint, Then the stale-check prompt is suppressed and hydration proceeds. (maps REQ-V3R2-RT-004-033)
- AC-V3R2-RT-004-07: Given `moai state dump --phase plan` runs, When stdout is captured, Then it contains the PlanCheckpoint fields and per-field provenance tags. (maps REQ-V3R2-RT-004-007, -030)
- AC-V3R2-RT-004-08: Given team mode with 3 teammates each writing checkpoint fragments, When the orchestrator merges, Then the merged PhaseState has `ProvenanceSource: SrcSession` and an audit subfield listing 3 per-agent paths. (maps REQ-V3R2-RT-004-021)
- AC-V3R2-RT-004-09: Given a corrupted `checkpoint-run.json` with invalid JSON, When Hydrate() is called, Then `CheckpointInvalid` error is returned naming the offending field. (maps REQ-V3R2-RT-004-041)
- AC-V3R2-RT-004-10: Given two concurrent Checkpoint() calls on the same file, When both race, Then one succeeds and the other fails with `CheckpointConcurrent` after 3 retries. (maps REQ-V3R2-RT-004-040)
- AC-V3R2-RT-004-11: Given a subagent body contains literal `AskUserQuestion(`, When CI lint runs (from SPEC-V3R2-ORC-002), Then the build fails naming the agent file. (maps REQ-V3R2-RT-004-042)
- AC-V3R2-RT-004-12: Given a Ralph iteration completes, When `WriteRunArtifact("iter-007", "prompt.md", bytes)` is called, Then `.moai/state/runs/iter-007/prompt.md` exists with the bytes. (maps REQ-V3R2-RT-004-015)
- AC-V3R2-RT-004-13: Given `.moai/config/sections/state.yaml` has `retention_days: 14`, When `moai clean` runs, Then `runs/` directories older than 14 days are eligible for pruning. (maps REQ-V3R2-RT-004-031)
- AC-V3R2-RT-004-14: Given session aborted mid-transition from plan to run, When next session starts, Then orchestrator detects in-flight state and offers resume-or-rollback via AskUserQuestion. (maps REQ-V3R2-RT-004-050)
- AC-V3R2-RT-004-15: Given validator/v10 tagged `RunCheckpoint.Harness string` with `validate:"oneof=minimal standard thorough"`, When a value `"ultra"` is written, Then Checkpoint() returns validation error naming the field. (maps REQ-V3R2-RT-004-004)

## 7. Constraints (제약)

- Technical: Go 1.22+; validator/v10 from SPEC-V3R2-SCH-001. Atomic rename pattern for crash-safety (write `*.tmp`, `os.Rename` to final). Advisory file locks via `golang.org/x/sys/unix` on Linux/macOS, native Windows file locks via `golang.org/x/sys/windows` on Windows.
- Backward compat: Non-breaking — existing `.moai/state/` ad-hoc markdown files are not disturbed; typed checkpoints are added alongside. No migration required for v2.x users (they simply gain the typed layer incrementally as phases re-run under v3).
- Platform: macOS / Linux / Windows. All filesystem operations use `filepath.Join` / `filepath.Abs` per CLAUDE.local.md §6 test isolation rules.
- Performance: Checkpoint write MUST complete in under 50 ms p99 for states up to 1 MiB (typical PlanCheckpoint is under 64 KiB). Hydrate MUST complete in under 30 ms p99 for the same range.
- Size: `.moai/state/runs/{iter-id}/artifacts/` directory MAY grow unbounded; `retention_days` opt-in pruning via REQ-V3R2-RT-004-031 is the mitigation.
- File encoding: text artifacts are UTF-8; binary artifacts permitted under `artifacts/` but not under top-level state files.
- Cache-prefix discipline (P-C05): this SPEC freezes the `(systemPrompt, userContext, systemContext)` assembly order at the hydrate step; the order is documented in `internal/session/hydrate.go` with a `// cache-prefix: DO NOT REORDER` comment.

## 8. Risks & Mitigations (리스크 및 완화)

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Typed checkpoints bloat state files beyond plain-markdown readability | L | L | JSON with indent=2; `moai state dump` provides pretty-print; YAML variants supported for human-edited state (sprint-contract.yaml). |
| Race conditions in team mode produce partial checkpoint writes | M | M | Atomic rename + advisory file lock per REQ-040; per-agent checkpoint paths avoid the common case. |
| STALE_SECONDS default of 1 hour annoys users with long debugging sessions | M | L | Configurable via ralph.yaml; `--resume` flag bypasses the prompt. |
| Provenance metadata adds complexity to state files without user-visible benefit | L | L | `moai state dump` is the one consumer; provenance is a single string per field. |
| File-first state over-stresses local disk for long-running Ralph loops | L | M | `retention_days` opt-in pruning (REQ-031); Ralph `runs/` directories are structured for selective cleanup. |
| BlockerReport routing through orchestrator duplicates work if multiple teammates block | M | M | REQ-V3R2-RT-004-051 consolidates blockers at the parent terminal via bubble-mode. |
| Validator/v10 schema drift between v3.0 and v3.1 breaks checkpoint reads | L | H | Schema version field in every checkpoint struct; migration framework (SPEC-V3R2-EXT-004) handles v1→v2 schema lifts. |

## 9. Dependencies (의존성)

### 9.1 Blocked by

- SPEC-V3R2-SCH-001 (validator/v10 integration).
- SPEC-V3R2-RT-005 (provides `Source` enum subset used for provenance tagging).
- SPEC-V3R2-CON-001 (file-first state declaration lives in the FROZEN zone per P11).

### 9.2 Blocks

- SPEC-V3R2-HRN-002 (Sprint Contract state durability relies on SessionStore.Checkpoint for `sprint-contract.yaml`).
- SPEC-V3R2-WF-003 (multi-mode router's `--mode loop` Ralph fresh-context iteration uses `runs/{iter-id}/` layout).
- SPEC-V3R2-WF-004 (Agentless fixed pipelines still record minimal `PhaseState` for audit trail).

### 9.3 Related

- SPEC-V3R2-RT-001 (hook JSON `continue: false` populates `BlockerReport.RequestedAction`).
- SPEC-V3R2-RT-002 (bubble-mode routing of BlockerReport to parent terminal).
- SPEC-V3R2-EXT-001 (typed memory taxonomy is distinct — memdir is per-user/cross-project; session state is per-SPEC).
- SPEC-V3R2-EXT-004 (migration runner applies schema migrations to checkpoint files).
- SPEC-V3R2-ORC-001 (agent roster reduction — builder-platform and manager-cycle read/write PhaseState).
- SPEC-V3R2-CON-003 (consolidation pass moves file-reading-optimization rule references into session-context comments).

## 10. Traceability (추적성)

- Theme: master §4.3 Layer 3 Runtime; §5.6 File-First State + Fresh-Context Iteration.
- Principle: P5 (Typed State + Durable Checkpoint at Phase Boundaries); P3 (Fresh-Context Iteration — compatible via state-durable-context-ephemeral resolution); P11 (File-First Primitives).
- Pattern: X-3 (typed frontmatter-as-schema), M-1 partial (state distinct from memdir but reuses Source provenance), R-6 (Ralph fresh-context file layout).
- Problem: P-C02 (no sub-agent context isolation primitive, HIGH); P-C05 (no cache-prefix discipline, MEDIUM).
- Master Appendix A: Principle P5 → primary SPEC-V3R2-RT-004; P3 → secondary.
- Master Appendix C: Pattern M-1 → partial SPEC-V3R2-RT-004 (memdir is SPEC-V3R2-EXT-001).
- Wave 1 sources: r3-cc-architecture-reread.md §1.1 (typed Session/QueryEngine), §2 Decision 2 (cache-prefix discipline), §4 Adopt 3 (sub-agent context primitive).
- Wave 2 sources: design-principles.md P5 (Typed State + Durable Checkpoint), P3 (Fresh-Context), P11 (File-First); pattern-library.md X-3, M-1, R-6; problem-catalog.md P-C02, P-C05.
- BC-ID: none (non-breaking — additive typed schema on top of existing `.moai/state/`).
- Priority: P1 High — enables HRN-002 Sprint Contract durability and WF-003 Ralph loop mode; not on the CRITICAL path but heavily cited downstream.
