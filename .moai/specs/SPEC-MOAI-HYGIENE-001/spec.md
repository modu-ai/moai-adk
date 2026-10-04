---
id: SPEC-MOAI-HYGIENE-001
title: ".moai Hygiene — Size-Based Audit-Log Rotation and Finished-Session State GC"
version: "0.1.0"
status: draft
created: 2026-10-05
updated: 2026-10-05
author: manager-spec
priority: P2
phase: "v3.2.0"
module: "internal/hygiene"
lifecycle: spec-anchored
era: V3R6
tier: M
related_specs: [SPEC-OBSERVE-HYGIENE-001, SPEC-WORKTREE-SWEEP-001]
tags: "audit-log-rotation, log-hygiene, session-state-gc, liveness, fail-closed, dry-run-default, sessionstart-hook, worktree-card-t1518"
---

# SPEC-MOAI-HYGIENE-001 — .moai Hygiene: Audit-Log Rotation + Finished-Session State GC

## §A Problem and Context

A read-only hygiene audit of a long-lived checkout (2026-10-04, evidence basis §G) measured two unbounded growth classes that no existing mechanism owns:

1. **Append-only audit sinks with no size bound.** Seven sinks under `.moai/logs/` grow without limit because `PruneObservationLogs` (SPEC-OBSERVE-HYGIENE-001, REQ-OBH-002) removes whole files by *age*, and an append-only sink's mtime is always fresh — the age-out never fires while the sink is written daily. Measured: `rule-load-audit.jsonl` 14.1 MB (first row 2026-09-22 — ~1 MB/day peak growth), `agent-model-audit.jsonl` 7.0 MB, `preedit-session-guard.log` 7.0 MB, `codex-adapter.jsonl` 5.5 MB, `status-transition-audit.log` 4.3 MB, `subagent-write-guard.log` 3.0 MB, `agent-stop-audit.jsonl` 2.9 MB. The sweep adds `hook-runtime.log` 2.3 MB (age-out only, same never-fires shape). Only size-based rotation bounds these; the trace writer (`internal/hook/trace/writer.go`, REQ-OBS-006, 10 MB threshold, keep-1) already establishes the in-codebase rotation scale for exactly this reason.

2. **Per-session state files of finished sessions.** `PruneObservationLogs` covers `logs/trace-*.jsonl` only. Nothing covers session-keyed residue elsewhere under `.moai/state/`: measured dead-session counts — `state/context-usage` 1,017 of 1,116 files; per-session traces under `logs/` 889 (owned by REQ-OBH-002, out of scope here); `state/todo` per-session files 201; `state/codex-stop-chain` 145; `state/codex-stop-cap` 19; `state/agent-stops` 31; `routing-pending-*.json` 80; `state/verify` scratch 911 files older than 30 days; stale `spec-close-*.lock` 15; `state/goal` 26.

The session registry is **not** a safe liveness source on its own: all 61 registry entries reported a live pid at measurement time (one pid, 40177, shared by sessions registered on 09-26 — pid reuse), while only 34 had a heartbeat under 24 h. Liveness for GC must therefore combine independent signals and fail closed.

Measured caveat that shapes the design: file mtime can be bulk-touched (hundreds of card report dirs shared one identical mtime), so mtime is never a sufficient age signal — a candidate whose age cannot be dated from its own content is spared, never deleted.

Two independent units are specified: a **rotator** (size-bound for registered audit sinks) and a **session-state GC** (liveness-evaluated residue removal). They share no failure domain: the rotator failing closed never blocks the GC and vice versa.

## §B Requirements (GEARS)

### §B.1 Rotation unit

- **REQ-HYG-001 (Ubiquitous):** The rotator shall bound each registered audit sink to at most `audit_log_max_bytes + audit_log_kept_rotations × audit_log_max_bytes` bytes on disk. Defaults: `audit_log_max_bytes = 10 MiB` (10×1024×1024 — the established trace-writer threshold), `audit_log_kept_rotations = 1`.

- **REQ-HYG-002 (When):** When the rotator runs and a registered sink's size is at or above `audit_log_max_bytes`, the rotator shall rotate the sink to its `<name>.1` successor — removing the previous `<name>.1` first, then renaming — while holding an exclusive lockfile, so two concurrent rotators never interleave removes and renames.

- **REQ-HYG-003 (Unwanted):** The rotator shall not read, write, rotate, or delete any path that is not a registered sink. The sink registry is a closed, named list: `rule-load-audit.jsonl`, `agent-model-audit.jsonl`, `preedit-session-guard.log`, `codex-adapter.jsonl`, `status-transition-audit.log`, `subagent-write-guard.log`, `agent-stop-audit.jsonl`, `hook-runtime.log`, and the hygiene audit sink itself (REQ-HYG-005).

- **REQ-HYG-004 (When detected):** When a source file opens an append-only sink under `.moai/logs/` without a corresponding sink-registry entry, the registry-completeness guard test shall fail and name the unregistered writer.

- **REQ-HYG-005 (Ubiquitous):** The rotator and the GC shall append one JSON row per evaluated action — action taken or skipped, with the reason — to their own audit sink `.moai/logs/hygiene-audit.jsonl`, and that sink shall itself appear in the sink registry so the rotator bounds it too.

### §B.2 Session-state GC unit

- **REQ-HYG-006 (Unwanted):** The GC shall not evaluate, report on, or delete any path outside the closed target registry. Targets (session-keyed, shapes verified against their writers): `state/context-usage/<sessionID>.json`, `state/goal/<session>.json` + `<session>.verdict.json` + `<session>.html`, `state/codex-stop-chain/<session>.json`, `state/codex-stop-cap/<session>.json`, `state/agent-stops/<session-id>.json`, `state/routing-pending-<session>.json`, `state/verify/<session>/` scratch subdirectories (the `snapshots/` subtree excluded), session-keyed files directly under `state/todo/`, and the lock class `state/spec-close-<SPEC-ID>.lock` (REQ-HYG-013).

- **REQ-HYG-007 (Ubiquitous):** The GC shall classify every candidate session's liveness as exactly one of LIVE, DEAD, or INDETERMINATE: LIVE when any affirmative liveness signal holds; DEAD only when every signal is affirmatively negative; INDETERMINATE when no signal is affirmative and at least one signal could not be measured.

- **REQ-HYG-008 (Ubiquitous):** The GC's liveness verdict shall combine at least two independent signal families, all three of which it evaluates: (a) the session registry's recorded pid, probed alive **and** consistent with a process-start-time fingerprint so a reused pid is not read as its original session; (b) transcript activity — a transcript file for the session id whose modification time falls inside the transcript activity window (default 48 h); (c) the registry heartbeat recency. A registry pid probe alone shall never decide liveness.

- **REQ-HYG-009 (Unwanted):** The GC shall not delete the state of a LIVE or INDETERMINATE session, and the run report shall state the keeping reason for every kept candidate.

- **REQ-HYG-010 (While apply mode, When a candidate is DEAD):** While the hygiene mode is `apply`, when a candidate is DEAD, its age is datable from its own content (a recorded timestamp field, with file mtime as fallback only), and that age meets the minimum-age floor (default 7 days), the GC shall delete the candidate and append an audit row naming the path, the reason, and the signal evidence. Deletion is idempotent: an already-vanished path is recorded as `already-gone`, never as an error.

- **REQ-HYG-011 (Ubiquitous):** The GC shall spare any candidate whose age cannot be dated from content or mtime, and any candidate whose session key cannot be resolved from its name — an undatable candidate is kept, never deleted.

- **REQ-HYG-012 (While report mode):** While the hygiene mode is `report` (the shipped default), the rotator and the GC shall produce decisions, reasons, and audit rows without modifying any candidate file — the project tree is byte-identical after a report run except for rows appended to `.moai/logs/hygiene-audit.jsonl`.

- **REQ-HYG-013 (When detected):** When a `spec-close-*.lock` candidate is older than the minimum-age floor and no process holds an open file descriptor on it, the GC shall treat it as removable (deletion still gated by REQ-HYG-010's mode rule); where the fd probe is absent or fails, the lock is INDETERMINATE and kept — the fail-closed disposition SPEC-WORKTREE-SWEEP-001 (REQ-WS-006) establishes for unanswerable predicates.

- **REQ-HYG-014 (Ubiquitous):** The rotator and the GC shall fail independently: a rotator error shall not prevent the GC run, and a GC error shall not prevent the rotator run; each reports its own outcome.

### §B.3 Execution and safety

- **REQ-HYG-015 (Where the operator has not opted in):** Where `workflow.hygiene.mode` is absent or `report`, neither unit shall rotate or delete anything regardless of the invoking surface; mutation requires the explicit operator setting `workflow.hygiene.mode: apply` (auto path) or the `--apply` flag (CLI path).

- **REQ-HYG-016 (When a SessionStart hook fires):** When a SessionStart hook fires, the hygiene engine shall run the rotator and then the GC in the configured mode, best-effort: stat-only scans, at most one fd probe per run, per-unit errors logged and never propagated, and session launch never blocked or delayed beyond the hook timeout budget.

- **REQ-HYG-017 (Where the operator runs the CLI):** Where the operator runs `moai clean --audit-logs` or `moai clean --session-state`, the CLI shall print the dry-run report by default and mutate only when `--apply` is passed.

- **REQ-HYG-018 (Ubiquitous) [HARD]:** Every rotator and GC test shall construct its fixture tree exclusively under `t.TempDir()`; no test shall read, write, rotate, or delete the repository's own `.moai` tree, and the real repository shall ever see only report-mode output from its own sessions' hooks unless the operator has set `apply`.

- **REQ-HYG-019 (Ubiquitous):** Every size, count, age, window, and mode threshold shall live as a named constant in `internal/config/defaults.go` with a `workflow.yaml` config-key override; no call site shall carry a magic number.

## §C Constraints

- **Hook budget:** SessionStart runs under the 5 s MoAI policy budget. The hygiene pass is stat-only for sinks and session-keyed candidates (tens of stats, sub-millisecond); the fd probe (one `lsof` invocation) runs only when lock-class candidates exist and is skipped, not retried, on failure.
- **Concurrency:** sink appenders take no lock (they are plain `O_APPEND` writers across hook processes). The rotator's lockfile serializes rotators only. Known residual, documented and accepted for audit sinks: rows appended by a process holding the removed `<name>.1` inode open at the moment of its removal are lost with the inode — a one-rename-wide window, the same residual class the harness retention pruner documents. On Windows, rename-over-open-file fails (sharing violation); the rotation is skipped and retried on the next run.
- **Compatibility:** no sink writer changes. Readers of the primary sink files are unaffected (rotation preserves the primary path). `agent-model-audit.jsonl` keeps its existing whole-file age-out (REQ-OBH-002); rotation and age-out are orthogonal (a post-age-out absent file is a rotator no-op).
- **Platform:** lockfile semantics reuse `internal/lockfile` (in-process mutex only on Windows — documented residual, same as the harness retention pruner).
- **Config:** keys under `workflow.yaml` (`hygiene:` block); named defaults in `internal/config/defaults.go` per the hardcoding-prevention convention.
- **Safety:** closed registries on both units; fail-closed liveness; report-mode default; idempotent deletion; per-file failures skip the file (the SessionStart path must never gate a launch on cleanup — same contract as `internal/factory/record_prune.go`).

## §D Scope and Exclusions

In scope: the two units, their registries, the liveness evaluator, config keys and defaults, SessionStart wiring, the `moai clean` flag extension, the guard tests, and documentation.

### Out of Scope — separate cards (operator-assigned)

- Memory stores and the MEMORY.md lifecycle (hygiene audit Area 2 — separate card).
- Worktree accumulation and reaping (hygiene audit Area 3 — H1–H7 cards).
- SPEC archive lifecycle and `spec.Locate` resolution (hygiene audit Area 4 — separate card).
- Reports-tree lifecycle (card-close-date archiving, `reports/worktrees` folding, loose `session-*.md` — M5 separate card).

### Out of Scope — already-owned surfaces (verified in-tree)

- `lessons-inbox.jsonl` rotation — owned by `internal/hook/inbox_lifecycle.go` and the `moai inbox` verbs.
- `harness/usage-log.jsonl` retention — owned by `internal/harness/retention.go` (30-day prune → monthly gz), working as measured.
- Per-session trace files (`logs/trace-*.jsonl`) — owned by `PruneObservationLogs` (REQ-OBH-002) plus the trace writer's own 10 MB rotation (REQ-OBS-006).
- `backups/`, `evolution/telemetry`, `.moai/cache/` retention, and local-branch reaping — no session-keyed shape; separate concern.
- `state/handoff/consumed/` records — the resume-message audit trail; kept by design.
- Refactoring sink writers to rotate in-process — deliberately not taken; one standalone rotator is specified instead (see decision-index Q1).

## §E Acceptance Criteria Overview

Acceptance criteria live in `acceptance.md` as Given-When-Then scenarios with runnable commands (AC-HYG-001 through AC-HYG-015), each carrying its canonical `(maps REQ-HYG-...)` coverage line. Requirement-to-AC traceability (HYG-N = REQ-HYG-0NN):

| HYG-N | AC coverage |
|---|---|
| HYG-001 | AC-HYG-001, AC-HYG-002 |
| HYG-002 | AC-HYG-001, AC-HYG-004 |
| HYG-003 | AC-HYG-003 |
| HYG-004 | AC-HYG-005 |
| HYG-005 | AC-HYG-009 |
| HYG-006 | AC-HYG-007, AC-HYG-008 |
| HYG-007 | AC-HYG-006 |
| HYG-008 | AC-HYG-006 |
| HYG-009 | AC-HYG-008 |
| HYG-010 | AC-HYG-008, AC-HYG-009 |
| HYG-011 | AC-HYG-008 |
| HYG-012 | AC-HYG-007 |
| HYG-013 | AC-HYG-015 |
| HYG-014 | AC-HYG-010 |
| HYG-015 | AC-HYG-007, AC-HYG-013 |
| HYG-016 | AC-HYG-012 |
| HYG-017 | AC-HYG-013 |
| HYG-018 | AC-HYG-014 |
| HYG-019 | AC-HYG-011 |

## §F Design Sketch (behavior level)

- **Rotator pass:** stat every registered sink → for each over-threshold sink: acquire lockfile → remove stale `<name>.1` → rename sink to `<name>.1` → release → append audit row. Under-threshold and absent sinks are no-ops. Steady-state bound per sink ≈ 2 × 10 MiB; against the measured ~1 MB/day peak growth, a 10 MiB chunk represents roughly ten days of tail, and keep-1 preserves the most recent full chunk for forensics.
- **GC pass:** enumerate target-registry candidates → resolve session key per candidate → evaluate the three-signal liveness verdict → decide (report both modes) → delete only in apply mode and only DEAD + datable + age-floor-met → append audit rows → emit the run report (kept candidates carry their keeping reason).
- **Liveness signals:** (a) registry pid probe + process start-time fingerprint (reuses the `internal/session` probe seam and the `internal/homestate` fingerprint machinery so a reused pid is not mistaken for its original session); (b) transcript mtime inside the 48 h window; (c) registry heartbeat recency. Any affirmative → LIVE; all measurable and negative → DEAD; any unmeasurable with none affirmative → INDETERMINATE (kept, reason reported). The bulk-mtime caveat cannot cause over-deletion: a bulk touch marks sessions LIVE (over-keeping), and the content-datable rule spares undatable records outright.
- **Modes:** `report` (default) / `apply`; auto path gated by `workflow.hygiene.mode`, CLI path by `--apply`.

## §G Provenance / Evidence Basis

- Hygiene audit (read-only, 2026-10-04): `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/8cc9d4cb-10a5-401b-b43b-b8829dd40c92/scratchpad/hygiene/hygiene.md` + `hygiene.json` (full tables). Installed-build caveat recorded there applies to its CLI-sourced figures; the sink sizes and state counts above are its independently measured git/python figures.
- Code owners read in this plan phase: `internal/hook/prune_logs.go` (+ `session_end.go:116` wiring), `internal/hook/trace/writer.go` (rotation precedent, 10 MB), `internal/harness/retention.go` (locked-prune precedent + documented residual window), `internal/factory/record_prune.go` (conservative GC contract: datable-or-spare, name-shape invariants, retentionDays≤0 off, absent-dir-is-state), `internal/session/session_pid.go` (three-valued liveness probe seam), `internal/session/registry.go`, `internal/homestate/process_fingerprint_*.go` (pid-reuse defense), `internal/cli/worktree/sweep_cwd_posix.go` (lsof fail-closed precedent), `internal/hook/inbox_lifecycle.go` (lessons-inbox already owned), sink writer constants (§B.1 list), state-shape writers (`internal/statusline/context_usage.go`, `internal/goal/state.go`, `internal/cli/codex_stop_chain.go`, `internal/codexwiring/stop_budget.go`, `internal/hook/agent_stop_guard.go`, `internal/harness/routing/types.go`, `internal/verify/store.go`, `internal/spec/lock.go`).
- Card t1508 lineage (SPEC-AGENTS-CONTRACT-001) is on a held branch and is NOT a dependency of this SPEC.

## §H Cross-References

- SPEC-OBSERVE-HYGIENE-001 — the age-based prune this SPEC complements (rotation bounds size; REQ-OBH-002 bounds age). Not superseded.
- SPEC-WORKTREE-SWEEP-001 (REQ-WS-006) — the fail-closed unanswerable-predicate disposition adopted for the fd probe.
- `internal/factory/record_prune.go` — the datable-or-spare GC contract reused for REQ-HYG-010/011.
- decision-index.md — the six design decisions surfaced during assembly, four delegated by card t1518 to this SPEC.

## HISTORY

- 2026-10-05 — v0.1.0 created by manager-spec (card t1518, plan phase, worktree `WT-audit-log-gc` @ develop `6643c7bba`). Evidence basis: 2026-10-04 hygiene audit. Status: draft. Requirements REQ-HYG-001..019 in GEARS notation; Tier M artifact set authored.
