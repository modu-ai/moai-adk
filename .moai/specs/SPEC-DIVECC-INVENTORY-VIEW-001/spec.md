---
id: SPEC-DIVECC-INVENTORY-VIEW-001
title: "Unified Agent/Harness Inventory View (moai inventory)"
version: "0.1.1"
status: draft
created: 2026-06-23
updated: 2026-06-23
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "cli, inventory, runtime, observability, dogfooding, divecc"
tier: M
era: V3R6
---

# SPEC-DIVECC-INVENTORY-VIEW-001 — Unified Agent/Harness Inventory View

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-23 | manager-spec | Initial plan-phase draft. Final candidate (N6) of Epic Dive-into-CC. Premise VERIFIED by orchestrator read-only reconnaissance (three inventory surfaces inspected: `internal/cli/session.go`, `internal/cli/worktree/list.go`, `internal/cli/harness/v4lifecycle.go`). Closing N6 completes the Epic 7/7. |
| 0.1.1 | 2026-06-23 | manager-spec | plan-auditor PASS-WITH-DEBT (0.86) remediation. D1: REQ-INV-010 reworded — out-of-project worktree behavior attributed to REQ-INV-008 error-degradation (worktrees error outside a git repo; sessions+harnesses degrade-to-empty), not a uniform all-zero report. D2: acceptance.md gains AC-INV-005c (non-nil provider whose `List()` errors). D3: plan §A.5/M4 names `renderCard` (drops `renderSummaryLine`). D4: §A.2/§E add `v4lifecycle_cmd.go` (`NewHarnessV4ListCmd`) command-factory attribution. |

---

## §A. Background

### A.1 Epic provenance (Dive-into-CC dogfooding)

This is the **final candidate (N6)** of **Epic Dive-into-CC** (see `../SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001/ROADMAP.md`). N1·N2·N3·N4·N5·N7 are all completed; closing N6 completes the Epic 7/7. The Epic applies findings from an academic reverse-engineering analysis of Claude Code (arXiv:2604.14228 "Dive into Claude Code", companion repo github.com/VILA-Lab/Dive-into-Claude-Code) to moai-adk's own harness — a self-improvement (dogfooding) exercise.

The paper open-direction that motivates N6: **"runtime-is-first-class"** (open-direction 1). Durable execution, checkpoints, and agent/harness inventory should be **user-visible and inspectable**, not scattered across separate commands. moai-adk already exposes three independent inventory surfaces; N6 composes them into one unified read-only view so an operator can see runtime state at a glance.

### A.2 Verified ground-truth (NOT hypothesis)

The premise that three independent inventory surfaces exist today — and that NO unified inventory command exists — was grounded by read-only inspection of the moai-adk tree this plan-phase. It is recorded here as **observed fact**, not hypothesis, per `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 (a defect/state claim is valid only when the domain's tooling — here, `Read`/`grep` over `internal/cli/` — was actually run and its output observed). The three surfaces and their backing functions:

**Surface 1 — SESSION inventory.** The `moai session list [--json] [--filter-spec=<ID>]` command (factory `newSessionListCmd` in `internal/cli/session.go`) lists active Claude Code sessions. Backing data: `session.QueryActiveWork(optSpecID string) ([]session.Entry, error)` reads the registry file `.moai/state/active-sessions.json`. The `session.Entry` struct carries json-tagged fields `{SessionID, SpecID, Phase, StartedAt, LastHeartbeat, PID, Host, CWD}`. The command ALREADY has a `--json` path (`json.MarshalIndent` of `[]Entry`, `internal/cli/session.go` lines ~140-146). When the registry file is absent, `QueryActiveWork` returns an empty slice (0 sessions), NOT an error.

**Surface 2 — WORKTREE inventory.** The `moai worktree list [--verbose]` command (`internal/cli/worktree/list.go`) lists active Git worktrees. Backing: `WorktreeProvider.List() ([]git.Worktree, error)` (the provider is `git.WorktreeManager`, var declared in `internal/cli/worktree/root.go`) queries Git directly via `git worktree list --porcelain`; there is NO JSON registry file. The `git.Worktree` struct (`internal/core/git/types.go`) carries `{Path, Branch, HEAD}`. This command does **NOT** have a `--json` flag today — it is text-only (`wtCard` rendering). The unified view serializes worktree entries INTERNALLY; it does NOT add `--json` to `moai worktree list`.

**Surface 3 — HARNESS inventory.** The `moai harness list [--json]` command (v4 lifecycle) lists user-owned harnesses. The `--json`-bearing command factory is `NewHarnessV4ListCmd` in `internal/cli/harness/v4lifecycle_cmd.go:47-97`; its backing function `harness.ListHarnesses(projectRoot string) ([]harness.HarnessEntry, error)` lives in the sibling `internal/cli/harness/v4lifecycle.go` and scans `.claude/commands/harness/*.md`, joining each with its `manifest.json`. (Command-factory vs backing-function split: the `--json` flag is owned by `NewHarnessV4ListCmd` in `v4lifecycle_cmd.go`; the enumeration logic the unified view calls is `ListHarnesses` in `v4lifecycle.go`.) The `harness.HarnessEntry` struct carries `{Name, Domain, EntryCommand, RunnerWorkflow, ManifestMissing, CommandPath, ManifestPath}`. The command ALREADY has a `--json` path. When the harness directory is absent, `ListHarnesses` returns `(nil, nil)` (0 harnesses), NOT an error. The run-phase unified view calls ONLY `ListHarnesses()` (in `v4lifecycle.go`); it does NOT call `NewHarnessV4ListCmd`.

**Negative signal — no unified inventory command exists today (verified).** `moai status` (project SPEC/config summary) and `moai doctor` (environment health — `internal/cli/doctor.go`, a composite grouped-output across diagnostic checks) are NOT inventory composition: neither enumerates sessions/worktrees/harnesses. The unified inventory view is therefore **greenfield** — a new command, not an extension of an existing one.

### A.3 ROADMAP drift correction (must be reflected)

The Epic ROADMAP §N6 text references **`/harness:devkit list`** as the harness inventory surface. That reference is **STALE**: `/harness:devkit` was split into three independent dev-only harnesses (per the dev-only-commands isolation policy), and the live harness inventory surface today is the Go CLI **`moai harness list --json`** (v4 lifecycle, `harness.ListHarnesses()`). This SPEC references `moai harness list` / `harness.ListHarnesses()`, NOT `/harness:devkit list`. The drift is recorded here so a future reader does not re-introduce the stale surface name.

### A.4 What this SPEC is (and is not)

This SPEC delivers a **new read-only CLI command** that COMPOSES the three existing inventory surfaces into one view. It is **NOT** a modification of the three backing packages. Concretely the deliverable is:

1. A new top-level command `moai inventory [--json]` (factory `newInventoryCmd()` registered in `internal/cli/root.go` `init()`).
2. A human-readable default output: a compact summary (Sessions / Worktrees / Harnesses — count + key fields per row).
3. A `--json` output: a structured `UnifiedInventoryReport` (`{"sessions": {...}, "worktrees": {...}, "harnesses": {...}}`).
4. Graceful degradation: when any surface is empty or its backing directory is absent, the unified view reports a 0-count for that surface, never an error.

The three backing packages (`internal/session`, `internal/cli/worktree`, `internal/cli/harness`) are CALLED via their existing exported functions; they are NOT modified.

---

## §B. Requirements (GEARS)

> GEARS notation per `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format. Subjects are generalized (the command, the report, the view) rather than the hardcoded "the system".

### B.1 Command existence & registration

- **REQ-INV-001 (Ubiquitous)**: The unified inventory command `moai inventory` **shall** exist as a new top-level cobra command, constructed by a factory `newInventoryCmd()` and registered on the root command via `rootCmd.AddCommand(newInventoryCmd())` in `internal/cli/root.go` `init()`.

- **REQ-INV-002 (Ubiquitous)**: The `moai inventory` command **shall** be a single top-level command (NOT a subcommand group, NOT an extension of `moai status` or `moai doctor`).

### B.2 Read-only composition of the three surfaces

- **REQ-INV-003 (Ubiquitous)**: The unified view **shall** compose all three inventory surfaces — sessions, worktrees, harnesses — by calling their existing exported functions: `session.QueryActiveWork("")` for sessions, `WorktreeProvider.List()` (the `git.WorktreeManager.List()` provider) for worktrees, and `harness.ListHarnesses(projectRoot)` for harnesses.

- **REQ-INV-004 (Ubiquitous)**: The unified view **shall** present each surface as a count plus a key-field summary per entry (the "3-surface count summary MVP" breadth): for sessions the key fields are `{SessionID (short), SpecID, Phase}`; for worktrees `{Branch, Path, HEAD (short)}`; for harnesses `{Name, Domain, ManifestMissing}`.

### B.3 Output modes

- **REQ-INV-005 (When)**: **When** the `--json` flag is supplied, the command **shall** emit a single structured `UnifiedInventoryReport` JSON object with three top-level keys `sessions`, `worktrees`, `harnesses`, each carrying a count and an entries array, via `json.MarshalIndent` written to `cmd.OutOrStdout()` (mirroring the `internal/cli/session.go` JSON convention).

- **REQ-INV-006 (When)**: **When** the `--json` flag is absent, the command **shall** emit a compact human-readable summary of the three surfaces (count + key fields per row) to `cmd.OutOrStdout()`, using the existing render utilities (`internal/cli/render.go`) where applicable.

### B.4 Graceful degradation (empty / absent surfaces)

- **REQ-INV-007 (When)**: **When** a surface's backing data is empty or its backing directory/file is absent (e.g. `.moai/state/active-sessions.json` does not exist → 0 sessions; `.claude/commands/harness/` does not exist → 0 harnesses; no worktrees beyond the main checkout), the command **shall** report a 0-count (or main-only) result for that surface and **shall not** return a non-zero exit code or an error for that condition alone.

- **REQ-INV-008 (When)**: **When** one surface's backing call returns a genuine error (NOT an empty result — e.g. `WorktreeProvider == nil` because the git module is unavailable, or a registry read I/O error), the command **shall** surface that error for the affected surface to stderr while still rendering the surfaces that succeeded, and **shall** exit non-zero only when no surface could be rendered.

### B.5 Project-root resolution & out-of-project invocation

- **REQ-INV-009 (When)**: **When** the command needs to locate the project root (for `harness.ListHarnesses(projectRoot)` and for anchoring the session registry path), it **shall** resolve it via the established `resolveProjectRoot(cmd)` helper pattern (`--project-root` flag → `os.Getwd()` fallback), matching `internal/cli/harness.go`.

- **REQ-INV-010 (Where)**: **Where** `moai inventory` is invoked outside a git repository, the command **shall** report sessions and harnesses as 0-count (per REQ-INV-007) and **shall** degrade the worktree surface per REQ-INV-008 (`worktrees.error` populated from the git error), still exiting 0 since ≥1 surface rendered successfully. The asymmetry is intentional: `session.QueryActiveWork` and `harness.ListHarnesses` degrade-to-empty when their backing file/directory is absent (`.moai/state/active-sessions.json` absent → 0 sessions; `.claude/commands/harness/` absent → 0 harnesses), whereas `WorktreeProvider.List()` runs `git worktree list --porcelain` and returns a non-nil error (`fatal: not a git repository`) when invoked outside a git repo — it does NOT return an empty slice. The out-of-project worktree behavior is therefore the REQ-INV-008 error-degradation path, not the REQ-INV-007 zero-count path.

### B.6 Zero backing-package modification (invariant)

- **REQ-INV-011 (Ubiquitous)**: The command **shall not** modify any file in `internal/session/`, `internal/cli/worktree/`, or `internal/cli/harness/`. The unified view only CALLS the existing exported functions (`session.QueryActiveWork`, `WorktreeProvider.List`, `harness.ListHarnesses`) and serializes their output; it adds no new exported function to, and changes no behavior of, those packages.

- **REQ-INV-012 (Ubiquitous)**: The command **shall not** add a `--json` flag to `moai worktree list` — the worktree surface lacks JSON output today, and the unified view serializes worktree entries INTERNALLY (within `internal/cli/inventory.go`) rather than by extending the worktree command.

### B.7 Subagent boundary (CLI discipline)

- **REQ-INV-013 (Ubiquitous)**: The new command **shall not** invoke `AskUserQuestion` or `mcp__askuser__*` — it runs in CLI/subagent context where the orchestrator owns user interaction (C-HRA-008 per `internal/cli/CLAUDE.md` § Subagent boundary). It emits exit codes + JSON/text only.

---

## §C. Out of Scope

The exclusions below are out of scope for this SPEC. Each is expressed as an `### Out of Scope — <topic>` H3 sub-heading with bullet items, satisfying the `OutOfScopeRule` (`MissingExclusions`) lint.

### Out of Scope — Observability statistics

- Composing N4-style failure clusters, tier-promotions, or any `moai-harness-learner` observability statistics into the inventory view. The MVP breadth is a count + key-field summary of the three inventory surfaces only. Surfacing observability state (failure signatures, evolution decisions, usage-log analytics) is deferred to a potential follow-up SPEC.

### Out of Scope — Cross-correlation across surfaces

- Computing or rendering a cross-correlation mapping between the surfaces — e.g. `session.spec_id` ↔ worktree branch name ↔ harness domain. The MVP presents the three surfaces side-by-side as independent count summaries; joining them into a correlated runtime graph (which session is working in which worktree on which harness) is deferred to a potential follow-up SPEC.

### Out of Scope — Modifying the backing inventory commands

- Modifying, refactoring, or extending `moai session list`, `moai worktree list`, or `moai harness list` (or their backing packages `internal/session`, `internal/cli/worktree`, `internal/cli/harness`). This SPEC adds a NEW command that composes the existing exported functions read-only; it changes none of the three sources.
- Adding a `--json` flag to `moai worktree list`. The worktree surface remains text-only at its own command; the unified view serializes worktree entries internally.

### Out of Scope — New inventory data / registry schema

- Introducing a new persistent inventory registry, a new on-disk state file, or a new schema. The unified view is a read-only projection of data the three existing surfaces already own; it persists nothing.
- Adding live/watch/streaming modes (e.g. `--watch`), filtering flags beyond what the surfaces already expose, or pagination. The MVP is a single point-in-time read.

### Out of Scope — Other Epic Dive-into-CC candidates

- N6 is the LAST candidate of Epic Dive-into-CC; N1·N2·N3·N4·N5·N7 are already completed. This SPEC covers only the unified inventory view. Any further runtime-first-class surfaces (durable-execution timeline, checkpoint browser) are new Epics, not part of N6.

---

## §D. Acceptance Criteria (summary)

Full Given-When-Then acceptance criteria are in `acceptance.md`. Summary of the binding gates:

- AC-INV-001: `moai inventory` command exists and is registered on the root command (built via `newInventoryCmd()`).
- AC-INV-002: the command composes all three surfaces via the named existing exported functions (read-only).
- AC-INV-003: `--json` emits a structured `UnifiedInventoryReport` with the three top-level keys.
- AC-INV-004: human-readable default emits a compact 3-surface summary.
- AC-INV-005: empty / absent surfaces yield a 0-count result, never an error (graceful degradation).
- AC-INV-006: zero backing-package modification — `internal/session`, `internal/cli/worktree`, `internal/cli/harness` are unchanged; `moai worktree list` gains no `--json` flag.
- AC-INV-007: the new command does not invoke `AskUserQuestion` (subagent boundary grep is 0).

---

## §E. Cross-References

- `../SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001/ROADMAP.md` § N6 — Epic Dive-into-CC, final candidate (with the `/harness:devkit list` → `moai harness list` drift correction recorded in §A.3 of this SPEC).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 — the state-claim grounding invariant that binds §A.2 (three surfaces verified, no unified command verified).
- `internal/cli/session.go` — SESSION surface; `newSessionListCmd` factory + `session.QueryActiveWork` + the `--json` `json.MarshalIndent` convention to mirror.
- `internal/cli/worktree/list.go` + `internal/core/git/types.go` (`Worktree` struct) + `internal/core/git/worktree.go` (`WorktreeManager.List`) — WORKTREE surface.
- `internal/cli/harness/v4lifecycle.go` (`harness.ListHarnesses` + `HarnessEntry` struct — the backing function the unified view calls) + `internal/cli/harness/v4lifecycle_cmd.go:47-97` (`NewHarnessV4ListCmd` — the `--json`-bearing command factory; documentation precision for the command-factory vs backing-function split, NOT called by run-phase) — HARNESS surface.
- `internal/cli/doctor.go` — composite grouped-output reference pattern for the human-readable default rendering.
- `internal/cli/harness.go` `resolveProjectRoot(cmd)` — project-root resolution helper to reuse (REQ-INV-009).
- `internal/cli/CLAUDE.md` § Subagent boundary (C-HRA-008), § Cobra subcommand registration, § Output streams — CLI module conventions the run-phase implementation must follow.
