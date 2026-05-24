---
id: SPEC-V3R6-HARNESS-PROPOSAL-GEN-001
title: "V3R4 Self-Evolving Harness Loop Closure — Automatic SPEC Proposal Generator"
version: "0.1.0"
status: draft
created: 2026-05-24
updated: 2026-05-24
author: manager-spec
priority: P1
phase: "v3.0.0 R6 — Harness Self-Evolution Closure"
module: "internal/harness/proposalgen, internal/cli/harness, .moai/proposals/"
lifecycle: spec-anchored
tags: "harness, proposal, learning-loop, v3r4, self-evolving, tier-m"
depends_on: [SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001]
breaking: false
bc_id: []
related_theme: "Self-Evolving Harness v2 — Learning Loop Closure"
target_release: v3.0.0
tier: M
---

# SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 — V3R4 Self-Evolving Harness Loop Closure: Automatic SPEC Proposal Generator

## §1. Purpose & Background

### §1.1 Vision and Predecessor Context

The V3R4 self-evolving harness was designed to observe its own usage patterns (hooks: `PostToolUse`, `SubagentStop`, `Stop`, `UserPromptSubmit`), cluster recurring patterns (V3R4-003 SimHash 64-bit embedding-cluster classifier), and surface tier promotions to a learning log. SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001 (HCW-001, merged at commit `577f10308` on 2026-05-24) **closed the wiring gap** — runtime hook events now feed the classifier through `moai hook harness-classify`, producing `.moai/harness/learning-history/tier-promotions.jsonl`.

The current learning loop has two halves:

```
[runtime events] → [classifier] → tier-promotions.jsonl   ← HCW-001 (CLOSED)
                                          ↓
                                  [proposal generator]    ← THIS SPEC (OPEN)
                                          ↓
                                  draft SPEC proposal     ← user gate via orchestrator
                                          ↓
                                  manager-spec subagent   ← auto-delegation post-approval
```

This SPEC delivers the second half: a generator that consumes `tier-promotions.jsonl`, maps learned patterns into draft SPEC proposals, and (via the orchestrator's AskUserQuestion gate) optionally auto-delegates to manager-spec for full SPEC authoring.

### §1.2 Current Data Limitation (CRITICAL SCOPE CLARIFICATION)

[HARD] As of plan-phase creation (2026-05-24, source data verified), the live `tier-promotions.jsonl` contains **8 records spanning 4 unique pattern_key values**, all of which are **system hook event patterns**:

- `agent_invocation:Bash:` (1 observation, tier=`observation`)
- `subagent_stop:unknown:` (41 observations, tier=`auto_update`)
- `session_stop::` (23 observations, tier=`auto_update`)
- `user_prompt::` (33 observations, tier=`auto_update`)

These patterns describe Claude Code session lifecycle events, not actionable code-change or error patterns that would map onto a new SPEC. Mapping any of these to a draft SPEC would produce false-positive proposals (e.g., "create SPEC to handle UserPromptSubmit events" — the events ARE the substrate, not a feature gap).

[HARD] The correct behavior with current data is a **graceful no-op**: the generator parses `tier-promotions.jsonl`, applies the mapping rules, finds zero actionable patterns, writes nothing to `.moai/proposals/`, and emits a diagnostic log (`reason: no-actionable-patterns`). The generator's value materializes when **richer learning sources** (code-change patterns, error patterns, repeated tool failure patterns) feed `tier-promotions.jsonl` in future V3R6+ iterations.

This SPEC is "future-data ready" by design — the loop closure must exist before the loop has data worth acting on.

### §1.3 Goals

- Close the V3R4 self-evolving harness learning loop by emitting actionable draft SPEC proposals when (and only when) `tier-promotions.jsonl` contains patterns that map onto identifiable feature gaps.
- Provide a CLI subcommand (`moai harness propose`) that orchestrators can invoke after run/sync phases or on demand.
- Define the orchestrator AskUserQuestion gate contract (Approve/Modify/Reject) for proposal review without violating the subagent boundary.
- Preserve the V3R4 hard subagent boundary: the CLI MUST NOT invoke `AskUserQuestion`; the orchestrator owns user interaction.
- Establish the `.moai/proposals/<draft-id>/` directory schema as a stable contract for downstream manager-spec consumption.

### §1.4 Non-Goals

- [HARD] MUST NOT produce false-positive proposals from system-event-only data (current dataset). When mapping rules find no actionable pattern, emit a graceful no-op + diagnostic log.
- [HARD] MUST NOT auto-delegate to manager-spec without the user gate. The CLI's auto-delegate output is a structured payload; only the orchestrator-driven `Approve` decision triggers re-spawn.
- [HARD] MUST NOT modify SPEC bodies after generation. The generator emits a draft once; subsequent edits flow through manager-spec as a regular SPEC authoring round.
- [HARD] MUST NOT consume sources beyond `.moai/harness/learning-history/tier-promotions.jsonl` in this SPEC scope. Additional sources (`usage-log.jsonl`, `observations.yaml`, git history) are deferred to follow-up SPECs.
- [HARD] MUST NOT invoke `AskUserQuestion` from the CLI binary (subagent boundary HARD per `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary).
- MUST NOT introduce new dependencies on external services (network calls, third-party APIs) — generator runs fully offline against local JSONL.
- MUST NOT assume any specific programming language in the generated proposal body (16-language neutrality per CLAUDE.local.md §15).
- MUST NOT delete or rewrite `tier-promotions.jsonl` — the generator is a consumer, not a manager.

## §2. EARS Requirements

### §2.1 Reader Contract (REQ-PGN-001..003)

- **REQ-PGN-001** (Ubiquitous): The proposal generator SHALL parse `.moai/harness/learning-history/tier-promotions.jsonl` line-by-line using the canonical `internal/harness.Promotion` schema (fields: `Ts time.Time`, `PatternKey string`, `FromTier string`, `ToTier string`, `ObservationCount int`, `Confidence float64`; JSON tags `ts`, `pattern_key`, `from_tier`, `to_tier`, `observation_count`, `confidence` in snake_case).
- **REQ-PGN-002** (Event-Driven): WHEN a line in `tier-promotions.jsonl` fails JSON unmarshal or schema validation, the reader SHALL skip the malformed line, accumulate a diagnostic count, and continue processing remaining lines. The diagnostic count SHALL appear in the CLI's JSON output as `malformed_lines: <count>`.
- **REQ-PGN-003** (Unwanted): IF `tier-promotions.jsonl` does not exist or is empty, THEN the reader SHALL return an empty promotion set without error and the generator SHALL emit a structured no-op response with `reason: "tier-promotions.jsonl absent or empty"`.

### §2.2 Mapper Contract (REQ-PGN-004..005)

- **REQ-PGN-004** (State-Driven): WHILE iterating promotions, the mapper SHALL produce a draft proposal candidate ONLY when ALL of the following hold for a given `pattern_key`:
  - `Confidence >= 0.70` (confidence threshold)
  - `ToTier ∈ {"recommendation", "approval_required"}` (excludes `observation` and `auto_update` — those tiers are pre-actionable)
  - `pattern_key` matches the actionable-pattern regex `^(code_change|error_pattern|tool_failure|repeated_edit):[a-z_]+:[^:]+$` (excludes system-event prefixes such as `session_stop`, `user_prompt`, `subagent_stop`, `agent_invocation`)
- **REQ-PGN-005** (Optional): WHERE the mapper produces zero candidates after evaluating all promotions, the generator SHALL emit a no-op response with `reason: "no-actionable-patterns"` and `evaluated_patterns: <total_unique_pattern_keys>`. No filesystem writes SHALL occur in this case (REQ-PGN-007 scaffolder is a no-op).

### §2.3 Scaffolding Contract (REQ-PGN-006..007)

- **REQ-PGN-006** (Ubiquitous): The scaffolder SHALL emit, for each actionable proposal candidate, a directory `.moai/proposals/PROPOSAL-<YYYYMMDD>-<short-hash>/` containing exactly two files:
  - `spec.md` — language-neutral draft SPEC with placeholder frontmatter (`status: draft`, `priority: P3`, `lifecycle: exploratory`) and a body section `## Origin` linking back to the source `pattern_key` + observation count + confidence.
  - `proposal.json` — structured metadata (source `pattern_key`, `observation_count`, `confidence`, `tier`, `generated_at` RFC3339 timestamp, `generator_version`).
- **REQ-PGN-007** (Unwanted): IF the mapper returns zero candidates (REQ-PGN-005 no-op path), THEN the scaffolder SHALL NOT create the `.moai/proposals/` directory tree if it does not already exist, and SHALL NOT write any files.

### §2.4 CLI Contract (REQ-PGN-008..009)

- **REQ-PGN-008** (Ubiquitous): The CLI SHALL expose `moai harness propose` as a new subcommand with the following flags:
  - `--auto` (boolean): When set, the CLI exits with `auto_delegate: true` in its JSON output, signaling the orchestrator to launch the AskUserQuestion gate. Default: false (CLI emits proposals and exits without gate signal).
  - `--dry-run` (boolean): When set, the CLI evaluates and reports candidates without writing to `.moai/proposals/`. Default: false.
  - `--limit N` (integer, default 5): Maximum number of proposals to emit per invocation. When candidates exceed N, sort by `Confidence` descending and truncate.
  - `--input PATH` (string, default `.moai/harness/learning-history/tier-promotions.jsonl`): Override the input JSONL path. Used by tests against frozen baseline fixtures (per AC-PGN-004) and by operators inspecting alternate learning-history snapshots. When omitted, the CLI reads the canonical live path.
- **REQ-PGN-009** (Event-Driven): WHEN `moai harness propose` completes, it SHALL emit a JSON object to stdout with the shape `{"proposals": [...], "reason": "<string>", "malformed_lines": <int>, "evaluated_patterns": <int>, "auto_delegate": <bool>}` and SHALL exit with code 0 on success (including no-op), code 1 on unrecoverable error (e.g., permission denied on `.moai/proposals/`), code 2 on malformed CLI flags.

### §2.5 Orchestrator AskUserQuestion Gate Contract (REQ-PGN-010..011)

- **REQ-PGN-010** (Ubiquitous): WHEN the orchestrator invokes `moai harness propose --auto` and the JSON output contains `auto_delegate: true` AND `len(proposals) >= 1`, the orchestrator SHALL present an AskUserQuestion gate with exactly three options: `Approve` (권장 / Recommended — auto-delegate to manager-spec), `Modify` (user provides edits via the auto-appended Other option, orchestrator re-emits the gate), `Reject` (proposal discarded, logged to `.moai/proposals/<draft-id>/REJECTED.md`).
- **REQ-PGN-011** (State-Driven): WHILE the gate awaits user response, the orchestrator SHALL NOT spawn manager-spec or modify any `.moai/proposals/` artifact. ON `Approve`, the orchestrator SHALL spawn manager-spec with the proposal `spec.md` content injected into the prompt. ON `Reject`, the orchestrator SHALL write `.moai/proposals/<draft-id>/REJECTED.md` containing the reason from the Other-option free-form input (or "no reason provided" if empty).

### §2.6 Subagent Boundary HARD (REQ-PGN-012..013)

- **REQ-PGN-012** (Unwanted): IF any source file under `internal/cli/harness/` references the symbol `AskUserQuestion` (any form: literal string, function call, type reference, comment containing the invocation pattern `AskUserQuestion(`), THEN the package SHALL fail the static-check test `TestPropose_NoAskUserQuestion` and CI SHALL block the merge. (Comments describing the orchestrator's behavior in prose — e.g., "the orchestrator calls AskUserQuestion after this CLI exits" — are permitted; only invocation patterns are forbidden, mirroring the precedent in `internal/cli/worktree/new_test.go`.)
- **REQ-PGN-013** (Ubiquitous): The test `TestPropose_NoAskUserQuestion` in `internal/cli/harness/propose_test.go` SHALL scan all `*.go` source files under `internal/cli/harness/` (excluding `*_test.go`) for the literal substring `AskUserQuestion(` and report a test failure if any match is found. This test mirrors the structural pattern of `TestNew_NoAskUserQuestion` in `internal/cli/worktree/new_test.go`.

### §2.7 Current-Data Graceful No-Op (REQ-PGN-014)

- **REQ-PGN-014** (Ubiquitous): With the live `tier-promotions.jsonl` as of 2026-05-24 (8 records, 4 unique `pattern_key` values: `agent_invocation:Bash:`, `subagent_stop:unknown:`, `session_stop::`, `user_prompt::` — all system events), `moai harness propose --dry-run` SHALL exit 0 with stdout JSON `{"proposals": [], "reason": "no-actionable-patterns", "malformed_lines": 0, "evaluated_patterns": 4, "auto_delegate": false}` and SHALL NOT create the `.moai/proposals/` directory.

## §3. Acceptance Reference

See `.moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/acceptance.md` for the AC-PGN-001..008 table with verification commands and expected outcomes.

## §4. Test Strategy

This SPEC follows TDD methodology per `.moai/config/sections/quality.yaml` `development_mode: tdd`. Each REQ has a corresponding test in either `internal/harness/proposalgen/*_test.go` or `internal/cli/harness/*_test.go`.

| Phase | Action |
|-------|--------|
| RED | Author failing tests for REQ-PGN-001 (reader), REQ-PGN-004 (mapper), REQ-PGN-006 (scaffolder), REQ-PGN-008 (CLI flags), REQ-PGN-013 (subagent boundary) before any implementation |
| GREEN | Implement minimal reader → mapper → scaffolder → CLI subcommand to pass tests |
| REFACTOR | Extract shared helpers (e.g., `pattern_key` parsing), tighten error wrapping, achieve ≥85% coverage |

Coverage target: `go test -cover ./internal/harness/proposalgen/... ./internal/cli/harness/...` ≥ 85%.

## §5. Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| R1 — Generator emits false positives if mapper regex is too permissive | REQ-PGN-004 explicitly excludes system-event prefixes via regex; AC-PGN-002 verifies no-op on current data |
| R2 — Subagent boundary regression (CLI calls AskUserQuestion) | REQ-PGN-013 CI guard test mirrors proven `TestNew_NoAskUserQuestion` pattern; static check fails fast |
| R3 — `.moai/proposals/` directory pollution across runs | REQ-PGN-006 uses date+hash directory naming for uniqueness; REQ-PGN-007 no-op skips creation; orchestrator owns retention policy (out of scope) |
| R4 — Mapper schema drifts from `internal/harness.Promotion` struct | REQ-PGN-001 binds reader to canonical struct via import; compile-time check |
| R5 — Future data sources change pattern_key format | This SPEC scopes to current `<event_type>:<subject>:` format; new formats will trigger a follow-up SPEC with mapper extension |
| R6 — Auto-delegate gate misuse (orchestrator skips Approve and spawns manager-spec) | REQ-PGN-010/011 are documentary contracts validated in plan.md §3; not enforceable at CLI layer (orchestrator-side discipline) |

## §6. Dependencies

- **Required**: SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001 (HCW-001, merged at `577f10308`) — produces `tier-promotions.jsonl`. Without HCW-001, the input file does not exist and REQ-PGN-003 returns the empty-input path.
- **Schema source**: `internal/harness.Promotion` struct at `internal/harness/types.go:191` — canonical JSON schema for promotion records.
- **CLI extension point**: `internal/cli/harness_route.go:58` (`newHarnessRouterCmd()` — the existing V3R4 `moai harness` parent registered at `internal/cli/root.go:104`). This SPEC's `moai harness propose` is registered as a NEW subcommand under this existing parent (alongside existing `status` / `apply` / `rollback` / `disable` siblings); NO new parent cobra.Command is created. Trade-off evaluated in plan.md §3.2.
- **Subagent boundary precedent**: `internal/cli/worktree/new_test.go` `TestNew_NoAskUserQuestion` — REQ-PGN-013 mirrors this pattern verbatim for `internal/cli/harness/`.

## §7. Out of Scope (Sub-section for spec-lint MissingExclusions compliance)

### §7.1 Out of Scope — Additional Learning Sources

This SPEC consumes ONLY `tier-promotions.jsonl`. The following sources are explicitly out of scope and deferred to follow-up SPECs:

- `.moai/harness/usage-log.jsonl` — raw event stream
- `.moai/harness/observations.yaml` — aggregated observations
- Git history pattern mining
- LSP/diagnostic feedback loops

### §7.2 Out of Scope — Proposal Retention and Garbage Collection

The lifecycle of `.moai/proposals/<draft-id>/` directories after generation (cleanup, archival, deduplication across runs) is out of scope. The orchestrator and/or a future `moai harness clean` subcommand will own retention policy.

### §7.3 Out of Scope — Orchestrator-Side Workflow Wiring

This SPEC defines the AskUserQuestion gate contract (REQ-PGN-010/011) as a documentary requirement that the orchestrator MUST follow. Actual workflow body modifications (e.g., adding `moai harness propose` invocation to `/moai sync` or `/moai run` workflows) are out of scope and deferred to a follow-up SPEC that focuses on workflow integration.

### §7.4 Out of Scope — Proposal Edit-After-Generation

Once a proposal is emitted to `.moai/proposals/<draft-id>/spec.md`, the generator MUST NOT re-emit or modify it. Subsequent refinement flows through manager-spec as a regular SPEC authoring round. Idempotency across multiple `moai harness propose` invocations on the same input is achieved through the date+hash directory naming (REQ-PGN-006) — re-running on the same `tier-promotions.jsonl` produces directories with the same name (idempotent overwrite-safe), but explicit overwrite policy is out of scope.

### §7.5 Out of Scope — Cross-Language Proposal Generation

Generated proposals use language-neutral SPEC format (EARS-style requirements in prose, no executable code in `spec.md` body). Mapping a learned pattern into a Go-specific, Python-specific, or TypeScript-specific implementation skeleton is explicitly out of scope. Implementation language selection is a manager-spec / manager-develop concern downstream of the proposal gate.
