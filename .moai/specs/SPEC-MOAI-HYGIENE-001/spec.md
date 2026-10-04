---
id: SPEC-MOAI-HYGIENE-001
title: ".moai Hygiene — Size-Based Audit-Log Rotation and Finished-Session State GC"
version: "0.2.0"
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

The session registry is **not** a safe liveness source on its own: all 61 registry entries reported a live pid at measurement time (one pid, 40177, shared by sessions registered on 09-26 — pid reuse), while only 34 had a heartbeat under 24 h — and the registry holds 61 entries against 1,116 `context-usage` files, so registry-absent and entry-missing candidates are the majority case and must be classified explicitly (REQ-HYG-007), never left to a single-signal inference.

Measured caveat that shapes the design: file mtime can be bulk-touched (hundreds of card report dirs shared one identical mtime), and preserved mtimes (`cp -p`, rsync) survive onto unrelated files — so **mtime is never a deletion datum**: a candidate whose age cannot be dated from a timestamp recorded in its own content is spared, never deleted (REQ-HYG-009).

Two independent units are specified: a **rotator** (size-bound for registered audit sinks) and a **session-state GC** (liveness-evaluated residue removal). They share no failure domain: the rotator failing closed never blocks the GC and vice versa.

## §B Requirements (GEARS)

### §B.1 Rotation unit

- **REQ-HYG-001 (Ubiquitous):** The rotator shall bound each registered audit sink in steady state to at most `audit_log_max_bytes + audit_log_kept_rotations × audit_log_max_bytes` bytes on disk: whenever the rotator runs, it shall rotate every registered sink it observes at or above `audit_log_max_bytes` (defaults: 10 MiB = 10×1024×1024, the established trace-writer threshold; kept rotations: 1). Between rotator runs a sink may transiently exceed the bound, and a legacy oversized file (the measured 14.1 MB first chunk) rotates on the first pass — it is never truncated in place.

- **REQ-HYG-002 (When):** When the rotator rotates a sink, it shall do so under an exclusive lockfile, shall re-check the sink's size and existence **inside the lock** — skipping without action when the sink is absent or has fallen below the threshold (a decision made from a pre-lock stat is never acted on) — and shall replace the previous chunk atomically: the primary is renamed to a staging name first, the displaced previous chunk is removed only after the fresh chunk is in its final `<name>.1` position. No interleaving of two rotators shall remove a freshly rotated chunk, and a failed rename shall never leave the tree with the previous chunk destroyed.

- **REQ-HYG-003 (Unwanted, When detected):** The rotator shall not read, write, rotate, or delete any path that is not a registered sink. The sink registry is a closed, named list: `rule-load-audit.jsonl`, `agent-model-audit.jsonl`, `preedit-session-guard.log`, `codex-adapter.jsonl`, `status-transition-audit.log`, `subagent-write-guard.log`, `agent-stop-audit.jsonl`, `hook-runtime.log`, and the hygiene audit sink itself (REQ-HYG-004). When a source file opens an append-only sink under `.moai/logs/` without a corresponding registry entry, the registry-completeness guard test shall fail and name the unregistered writer.

- **REQ-HYG-004 (Ubiquitous):** The rotator and the GC shall append rows to their own audit sink `.moai/logs/hygiene-audit.jsonl`: in report mode exactly **one summary row per unit per run** (counts, outcomes, kept-reason tally — never per-candidate rows), in apply mode one row per action taken or skipped with its reason and signal evidence. The sink itself appears in the sink registry so the rotator bounds it **in every mode** — the hygiene log's own rotation is hygiene of the hygiene log, not candidate mutation, and is exempt from the mode gate so the default report mode cannot grow an unbounded sink of the kind this SPEC exists to bound.

### §B.2 Session-state GC unit

- **REQ-HYG-005 (Unwanted):** The GC shall not evaluate, report on, or delete any path outside the closed target registry. Targets (session-keyed, shapes verified against their writers): `state/context-usage/<sessionKey>.json`, `state/goal/<sessionKey>.json` + `<sessionKey>.verdict.json` + `<sessionKey>.html`, `state/codex-stop-chain/<sessionKey>.json`, `state/codex-stop-cap/<sessionKey>.json`, `state/agent-stops/<sessionKey>.json`, `state/routing-pending-<sessionKey>.json`, `state/verify/<sessionKey>/` scratch subdirectories (the `snapshots/` subtree excluded), session-keyed files directly under `state/todo/` (the shared `backlog.json`, `backlog.db`, and every directory entry excluded), and the lock class `state/spec-close-<SPEC-ID>.lock` (REQ-HYG-011). A **session key** is a session UUID in the shape the writers stamp (36-character hyphenated UUID); a candidate whose name does not resolve to a session key in that shape has no key and is out of scope (spared).

- **REQ-HYG-006 (Unwanted):** Neither unit shall traverse a symlink: any registered or candidate path whose resolution encounters a symlinked component (parent or leaf) shall be refused — the unit records a `symlink-refused` outcome and leaves the path untouched. Path resolution shall be anchored (resolve the fully-resolved real path and require it to remain inside the project's `.moai` root, or traverse via directory file descriptors); a path is never deleted through a name that could have been re-pointed between enumeration and action.

- **REQ-HYG-007 (Ubiquitous):** The GC shall classify every candidate session's liveness as exactly one of LIVE, DEAD, or INDETERMINATE. It shall evaluate three independent signal families: (a) the session registry's recorded pid, probed alive **and** consistent with a process-start-time fingerprint (a live pid with no recorded or unreadable fingerprint yields no affirmative pid signal); (b) transcript activity — a transcript file for the session key whose modification time falls inside the transcript activity window (default 48 h); (c) the registry heartbeat recency. The verdict is LIVE when any affirmative signal holds; DEAD only when **every** signal is affirmatively negative; INDETERMINATE when no signal is affirmative and at least one is unmeasured. Where the registry file is absent or carries no entry for the candidate's session key — the majority case (61 entries against 1,116 files measured) — the pid and heartbeat signals are **unmeasured**, and the verdict can only be LIVE (transcript affirmative) or INDETERMINATE (transcript negative or unmeasured), never DEAD. A registry pid probe alone shall never decide liveness.

- **REQ-HYG-008 (Unwanted):** The GC shall not delete the state of a LIVE or INDETERMINATE session, and the run report shall state the keeping reason for every kept candidate.

- **REQ-HYG-009 (While apply mode, When a candidate is DEAD):** While the hygiene mode is `apply`, when a candidate is DEAD and its age is datable **from a timestamp recorded in its own content**, and that age meets the minimum-age floor (default 7 days), the GC shall delete the candidate and record an audit row naming the path, the reason, and the signal evidence. A candidate whose body carries no recorded timestamp is **undatable and spared regardless of its mtime** — file mtime is never a deletion datum (the measured bulk-touch caveat and preserved-mtime hazard make mtime-only dating unsafe in both directions). Deletion is idempotent: an already-vanished path is recorded as `already-gone`, never as an error.

### §B.3 Execution and safety

- **REQ-HYG-010 (While report mode):** While the hygiene mode is `report` (the shipped default), the rotator and the GC shall produce decisions and summary rows without modifying any candidate file and without creating the rotator lockfile — the project tree is byte-identical after a report run except for at most one summary row per unit appended to `.moai/logs/hygiene-audit.jsonl` (plus that sink's own rotation, REQ-HYG-004).

- **REQ-HYG-011 (When detected):** When the GC evaluates a `spec-close-*.lock` candidate, it shall not probe-then-remove (a fd probe followed by unlink races a close op acquiring the old inode and yields two live locks). Instead the GC shall attempt to acquire the lock itself, non-blockingly: on Unix (`flock(2)` semantics per `internal/spec/lock.go`) an acquired lock means the lock was free, and the file is removed while the GC holds the acquired handle, which is then released; `EWOULDBLOCK` means the lock is held and the candidate is kept with reason `lock-held`. On Windows the O_EXCL create-file protocol has no acquirable handle for an existing lock file, so the lock class is always kept on Windows and reported as `platform-unsupported`, never deleted.

- **REQ-HYG-012 (Ubiquitous):** The rotator and the GC shall fail independently: a rotator error shall not prevent the GC run, and a GC error shall not prevent the rotator run; each reports its own outcome.

- **REQ-HYG-013 (Where the operator has not opted in on the invoking surface):** Where `workflow.hygiene.mode` is absent or `report`, the SessionStart auto path shall not rotate or delete anything. The CLI mutates only when `--apply` is passed on that invocation — `--apply` overrides the config mode **for that CLI invocation only**; the config mode alone never makes a CLI invocation mutate. Absent the opt-in on its own surface, no unit rotates or deletes on any surface.

- **REQ-HYG-014 (When a SessionStart hook fires):** When a SessionStart hook fires, the hygiene engine shall run the rotator and then the GC in the configured mode, best-effort: bounded scans (stat-based enumeration, with per-candidate probes — fingerprint reads, content-timestamp parsing, the single non-blocking lock attempt for the lock class — each cheap and per-path), per-unit errors logged and never propagated, and session launch never blocked or delayed beyond the hook timeout budget.

- **REQ-HYG-015 (Ubiquitous) [HARD]:** Every rotator and GC test shall construct its fixture tree exclusively under `t.TempDir()`; no test shall read, write, rotate, or delete the repository's own `.moai` tree, and the real repository shall ever see only report-mode output from its own sessions' hooks unless the operator has set `apply`.

- **REQ-HYG-016 (Ubiquitous):** Every size, count, age, window, and mode threshold shall live as a named constant in `internal/config/defaults.go` with a `workflow.yaml` config-key override; no call site shall carry a magic number.

## §C Constraints

- **Hook budget:** SessionStart runs under the 5 s MoAI policy budget. Enumeration is stat-based; per-candidate probes are cheap reads; the lock-class acquire is one non-blocking attempt per candidate; nothing retries or waits.
- **Concurrency:** sink appenders take no lock (plain `O_APPEND` writers across hook processes). The rotator's lockfile serializes rotators only, and the rotation decision is made under the lock (REQ-HYG-002). Known residual, documented and accepted for audit tails: rows appended by a process holding the **displaced previous chunk's** inode open at the moment of its removal are lost with the inode — a window that opens only after the fresh chunk is already in place, the same residual class the harness retention pruner documents. On Windows, rename-over-open-file fails (sharing violation): the rotation is skipped and retried next run; the lockfile is an in-process mutex on Windows only, so cross-process rotation serialization is best-effort there and the skip-and-retry path is the backstop.
- **Symlinks:** both units refuse symlinked path components (REQ-HYG-006); the GC suite carries a parent-swap test.
- **Compatibility:** no sink writer changes. Readers of the primary sink files are unaffected (rotation preserves the primary path). `agent-model-audit.jsonl` keeps its existing whole-file age-out (REQ-OBH-002); rotation and age-out are orthogonal (a post-age-out absent file is a rotator no-op).
- **Config:** keys under `workflow.yaml` (`hygiene:` block); named defaults in `internal/config/defaults.go` per the hardcoding-prevention convention.
- **Safety:** closed registries on both units; fail-closed liveness (registry-absent and entry-missing explicitly classified); mtime never a deletion datum; report-mode default; idempotent deletion; per-file failures skip the file (the SessionStart path must never gate a launch on cleanup — same contract as `internal/factory/record_prune.go`).

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

Acceptance criteria live in `acceptance.md` as Given-When-Then scenarios with runnable commands (AC-HYG-001 through AC-HYG-016), each carrying its canonical `(maps REQ-HYG-...)` coverage line and a four-element RED-now evidence cell (command, verbatim stdout, exit code, tree SHA) in the acceptance evidence ledger. Requirement-to-AC traceability (HYG-N = REQ-HYG-0NN):

| HYG-N | AC coverage |
|---|---|
| HYG-001 | AC-HYG-001, AC-HYG-002, AC-HYG-003 |
| HYG-002 | AC-HYG-001, AC-HYG-004 |
| HYG-003 | AC-HYG-005 |
| HYG-004 | AC-HYG-007, AC-HYG-009 |
| HYG-005 | AC-HYG-008 |
| HYG-006 | AC-HYG-011 |
| HYG-007 | AC-HYG-006 |
| HYG-008 | AC-HYG-008 |
| HYG-009 | AC-HYG-008, AC-HYG-009 |
| HYG-010 | AC-HYG-007 |
| HYG-011 | AC-HYG-012 |
| HYG-012 | AC-HYG-010 |
| HYG-013 | AC-HYG-007, AC-HYG-014 |
| HYG-014 | AC-HYG-013 |
| HYG-015 | AC-HYG-016 |
| HYG-016 | AC-HYG-015 |

## §F Design Sketch (behavior level)

- **Rotator pass:** acquire per-sink lockfile → **inside the lock** re-stat size and existence → skip when absent or under threshold → else rename primary to staging → remove the displaced previous chunk only after the fresh chunk is staged at its final `<name>.1` position → release → append the audit row. The decision is never made from a pre-lock stat; a stale pre-lock observation leads to a no-op skip, never to a removal. Steady-state bound per sink ≈ 2 × 10 MiB; against the measured ~1 MB/day peak growth, a 10 MiB chunk represents roughly ten days of tail, and keep-1 preserves the most recent full chunk for forensics.
- **GC pass:** enumerate target-registry candidates → resolve the session key (36-char hyphenated UUID shape) → evaluate the three-signal liveness verdict (registry absent or entry missing ⇒ pid+heartbeat unmeasured ⇒ LIVE-or-INDETERMINATE only) → decide (report both modes) → delete only in apply mode and only DEAD + content-datable + age-floor-met → append rows (report mode: one summary row per unit) → emit the run report (kept candidates carry their keeping reason).
- **Liveness signals:** (a) registry pid probe + process start-time fingerprint (reuses the `internal/session` probe seam and the `internal/homestate` fingerprint machinery so a reused pid is not mistaken for its original session; a missing or unreadable fingerprint leaves the pid signal unmeasured); (b) transcript mtime inside the 48 h window; (c) registry heartbeat recency. Any affirmative → LIVE; all measurable and negative → DEAD; any unmeasured with none affirmative → INDETERMINATE (kept, reason reported). The bulk-mtime caveat cannot cause over-deletion: a bulk touch marks sessions LIVE (over-keeping), and mtime alone never deletes (REQ-HYG-009).
- **Lock class:** GC-side non-blocking acquire (acquired ⇒ free ⇒ remove while holding; `EWOULDBLOCK` ⇒ held ⇒ kept; Windows ⇒ always kept, `platform-unsupported`).
- **Modes:** `report` (default) / `apply`; the auto path is gated by `workflow.hygiene.mode`, the CLI by `--apply` for that invocation (REQ-HYG-013).

## §G Provenance / Evidence Basis

- Hygiene audit (read-only, 2026-10-04): `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/8cc9d4cb-10a5-401b-b43b-b8829dd40c92/scratchpad/hygiene/hygiene.md` + `hygiene.json` (full tables). Installed-build caveat recorded there applies to its CLI-sourced figures; the sink sizes and state counts above are its independently measured git/python figures.
- Code owners read in this plan phase: `internal/hook/prune_logs.go` (+ `session_end.go:116` wiring), `internal/hook/trace/writer.go` (rotation precedent, 10 MB), `internal/harness/retention.go` (locked-prune precedent + documented residual window), `internal/factory/record_prune.go` (conservative GC contract: datable-or-spare, name-shape invariants, retentionDays≤0 off, absent-dir-is-state), `internal/session/session_pid.go` (three-valued liveness probe seam), `internal/session/registry.go`, `internal/homestate/process_fingerprint_*.go` (pid-reuse defense), `internal/cli/worktree/sweep_cwd_posix.go` (lsof fail-closed precedent), `internal/hook/inbox_lifecycle.go` (lessons-inbox already owned), sink writer constants (§B.1 list), state-shape writers (`internal/statusline/context_usage.go`, `internal/goal/state.go`, `internal/cli/codex_stop_chain.go`, `internal/codexwiring/stop_budget.go`, `internal/hook/agent_stop_guard.go`, `internal/harness/routing/types.go`, `internal/verify/store.go`, `internal/spec/lock.go` — flock(2) Unix / O_EXCL Windows semantics recorded in REQ-HYG-011).
- Plan-audit iteration 1 verdict (FAIL 0.73, `.moai/reports/t1518/plan-audit.md`, audited SHA 8039ea714): D1–D14 remediation folded into this revision — 16-REQ consolidation (D9), registry-absent classification (D2), mtime prohibition convergence (D3), re-stat-under-lock + atomic chunk replacement (D4), symlink refusal (D5), report-mode row granularity + self-sink bound (D6), GC-side non-blocking lock acquire (D7), mode precedence (D8), bound restated (D10), widened greps + seeded controls (D11), typo (D12), lockfile/fingerprint/stat-only/session-key minors (D13). SPEC-AGENTS-CONTRACT-001 stays a non-dependency disclaimer (D14; dropped at close if still unlanded).
- Card t1508 lineage (SPEC-AGENTS-CONTRACT-001) is on a held branch and is NOT a dependency of this SPEC.

## §H Cross-References

- SPEC-OBSERVE-HYGIENE-001 — the age-based prune this SPEC complements (rotation bounds size; REQ-OBH-002 bounds age). Not superseded.
- SPEC-WORKTREE-SWEEP-001 (REQ-WS-006) — the fail-closed unanswerable-predicate disposition adopted for unmeasurable signals.
- `internal/factory/record_prune.go` — the datable-or-spare GC contract reused for REQ-HYG-009.
- `internal/spec/lock.go` — the lock-class platform semantics (flock(2) / O_EXCL) REQ-HYG-011 builds on.
- decision-index.md — the six design decisions surfaced during assembly, four delegated by card t1518 to this SPEC.

## HISTORY

- 2026-10-05 — v0.2.0 by manager-spec (card t1518, plan-audit iteration-1 remediation). FAIL 0.73 verdict folded: MP-8 four-element evidence ledger authored from direct execution on the pre-implementation tree (8039ea714); REQ set consolidated 19 → 16 (Tier M ceiling); D2–D14 applied per the verdict's numbered fixes.
- 2026-10-05 — v0.1.0 created by manager-spec (card t1518, plan phase, worktree `WT-audit-log-gc` @ develop `6643c7bba`). Evidence basis: 2026-10-04 hygiene audit. Status: draft.
