---
id: SPEC-MOAI-HYGIENE-001
title: ".moai Hygiene — Size-Based Audit-Log Rotation and Finished-Session State GC"
version: "0.3.0"
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

2. **Per-session state files of finished sessions.** `PruneObservationLogs` covers `logs/trace-*.jsonl` only. Nothing covers session-keyed residue elsewhere under `.moai/state/`: measured dead-session counts — `state/context-usage` 1,017 of 1,116 files; per-session traces under `logs/` 889 (owned by REQ-OBH-002, out of scope here); `state/todo` per-session files 201; `state/codex-stop-chain` 145; `state/codex-stop-cap` 19; `state/agent-stops` 31; `routing-pending-*.json` 80; `state/verify` scratch 911 files older than 30 days; stale `spec-close-*.lock` 15 (excluded from GC scope — REQ-HYG-011); `state/goal` 26.

The session registry is **not** a safe liveness source on its own: all 61 registry entries reported a live pid at measurement time (one pid, 40177, shared by sessions registered on 09-26 — pid reuse), while only 34 had a heartbeat under 24 h — and the registry holds 61 entries against 1,116 `context-usage` files, so registry-absent and entry-missing candidates are the majority case and must be classified explicitly (REQ-HYG-007), never left to a single-signal inference.

Measured caveat that shapes the design: file mtime can be bulk-touched (hundreds of card report dirs shared one identical mtime), and preserved mtimes (`cp -p`, rsync) survive onto unrelated files — so **mtime is never a deletion datum**: a candidate whose age cannot be dated from a timestamp recorded in its own content is spared, never deleted (REQ-HYG-009).

Two independent units are specified: a **rotator** (size-bound for registered audit sinks) and a **session-state GC** (liveness-evaluated residue removal). They share no failure domain: the rotator failing closed never blocks the GC and vice versa.

## §B Requirements (GEARS)

### §B.1 Rotation unit

- **REQ-HYG-001 (Ubiquitous):** The rotator shall hold each registered audit sink to a file-count bound of exactly `1 + audit_log_kept_rotations` files (defaults: 10 MiB = 10×1024×1024 rotation threshold, the established trace-writer scale; kept rotations: 1 — the primary plus one displaced chunk): whenever the rotator runs, it shall rotate every registered sink it observes at or above `audit_log_max_bytes`. Each displaced chunk is a rotation-time snapshot and may itself exceed the threshold when a sink grew between runs; no file is ever truncated in place, and a legacy oversized file (the measured 14.1 MB first chunk) rotates on the first pass.

- **REQ-HYG-002 (When):** When the rotator rotates a sink, it shall do so under a cross-process exclusive lockfile, shall re-check the sink's size and existence **inside the lock** — skipping without action when the sink is absent or has fallen below the threshold (a decision made from a pre-lock stat is never acted on) — and shall replace the previous chunk through a staged, crash-recoverable sequence over two rotator-owned artifact names: `<name>.1` (the displaced chunk) and `<name>.1.staging` (a chunk mid-placement). The sequence: the primary is renamed to `<name>.1.staging` first; the existing `<name>.1` is removed only while a staged replacement exists; `<name>.1.staging` is then promoted to `<name>.1`. At pass start, under the lock, the rotator shall recover an orphan `<name>.1.staging` left by a crashed pass by completing the interrupted placement (remove the stale `<name>.1`, promote the staging chunk) — **a staged chunk is never deleted and never overwritten**. Where cross-process exclusion cannot be established — Windows, where the in-process mutex is not cross-process, unless a cross-process sidecar lock (LockFileEx-based) is taken and verified — the rotation is skipped (`skipped-locked` when the sidecar lock is held, `skipped-platform` when the sidecar cannot be established) and retried on the next run; **a chunk is never destroyed on a path whose exclusivity is unverified**.

- **REQ-HYG-003 (Unwanted, When detected):** The rotator shall not read, write, rotate, or delete any path that is not a registered sink or a rotator-owned artifact of one. The sink registry is a closed, named list: `rule-load-audit.jsonl`, `agent-model-audit.jsonl`, `preedit-session-guard.log`, `codex-adapter.jsonl`, `status-transition-audit.log`, `subagent-write-guard.log`, `agent-stop-audit.jsonl`, `hook-runtime.log`, and the hygiene audit sink itself (REQ-HYG-004) — together with the rotator-owned chunk artifacts `<name>.1` and `<name>.1.staging` of those sinks, which no pass may treat as foreign and no pass may overwrite with content it did not stage. When a source file opens an append-only sink under `.moai/logs/` without a corresponding registry entry, the registry-completeness guard test shall fail and name the unregistered writer.

- **REQ-HYG-004 (Ubiquitous):** The rotator and the GC shall append rows to their own audit sink `.moai/logs/hygiene-audit.jsonl`: in report mode exactly **one summary row per unit per run** (counts, outcomes, kept-reason tally — never per-candidate rows), in apply mode one row per action taken or skipped with its reason and signal evidence. **Rotation of the hygiene sink happens only in apply mode** — the same opt-in gate that governs every other rotation governs this one (REQ-HYG-013 holds for the hygiene sink without exemption); in report mode an over-threshold hygiene sink is left untouched and the summary row records the skip. Between apply runs the report-mode growth rate is bounded by the summary granularity (on the order of two small rows per session start); this is a disclosed known limitation, read together with decision-index Q7/Q8. In apply mode the sink is bounded by REQ-HYG-001 like every registered sink.

### §B.2 Session-state GC unit

- **REQ-HYG-005 (Unwanted):** The GC shall not evaluate, report on, or delete any path outside the closed target registry. A **session key** is a session UUID in the shape the writers stamp (36-character hyphenated UUID); a candidate whose name does not resolve to a session key in that shape has no key and is out of scope (spared) — except the lock class, which is excluded on its own terms (REQ-HYG-011), not spared by the session-key rule. Targets, with their per-class shape and dating datum (a class whose dating datum is absent from a candidate's body is content-undatable ⇒ spared):

  | Target class | Shape | Dating datum (content-recorded) |
  |---|---|---|
  | `state/context-usage/<key>.json` | file | `captured_at` (writer: `internal/statusline/context_usage.go`) |
  | `state/goal/<key>.json` + `.verdict.json` + `.html` | grouped triple | `recorded_at` in the `.json` member; the triple is deleted together or not at all — a partial deletion that strands an undatable sibling is prohibited |
  | `state/codex-stop-chain/<key>.json` | file | the record's verdict timestamp (field pinned by the M3 per-class dating RED test) |
  | `state/codex-stop-cap/<key>.json` | file | the counter's recorded timestamp (field pinned at M3) |
  | `state/agent-stops/<key>.json` | file | newest entry `stopped_at` (writer: `internal/hook/agent_stop_guard.go`) |
  | `state/routing-pending-<key>.json` | file | `timestamp` (writer: `internal/harness/routing/types.go`) |
  | `state/verify/<key>/` (excluding `snapshots/`) | directory | dated **per entry** from the entry's own recorded timestamp (field pinned at M3); entries are removed entry-by-entry as they individually qualify, and the directory itself is removed only when empty |
  | session-keyed files directly under `state/todo/` | file | the file's recorded timestamp (field pinned at M3); the shared `backlog.json`, `backlog.db`, and every directory entry are excluded regardless |

- **REQ-HYG-006 (Unwanted):** Neither unit shall act through a symlinked path component **strictly below the project's fully-resolved `.moai` root**. The root itself is resolved through any symlinks above it (macOS `/var` → `/private/var` included) and never counts as a violation; a candidate is refused when its fully-resolved real path escapes the resolved root or when any path component below the resolved root is a symlink — outcome `symlink-refused`, path untouched. The resolution check is re-evaluated immediately before each action; the residual between that re-check and the action (a parent swapped inside the final window) is a **named residual** of this SPEC, detected and refused by the re-check on the next pass, and asserted as such by the parent-swap test.

- **REQ-HYG-007 (Ubiquitous):** The GC shall classify every candidate session's liveness as exactly one of LIVE, DEAD, or INDETERMINATE. It shall evaluate three independent signal families: (a) the session registry's recorded pid, probed alive **and** consistent with a process-start-time fingerprint (a live pid with no recorded or unreadable fingerprint yields no affirmative pid signal); (b) transcript activity — a transcript file for the session key found under the scanned profile roots whose modification time falls inside the transcript activity window (default 48 h); **a resolvable root with no transcript file for the key yields an unmeasured transcript signal, never a negative** (the session may live under a root this install does not scan — worktree and secondary-profile sessions are the measured shape); (c) the registry heartbeat recency against the staleness window (`HygieneHeartbeatStaleWindow`, default 24 h). The verdict is LIVE when any affirmative signal holds; DEAD only when **every** signal is affirmatively negative; INDETERMINATE when no signal is affirmative and at least one is unmeasured. Where the registry file is absent or carries no entry for the candidate's session key — the majority case (61 entries against 1,116 files measured) — the pid and heartbeat signals are **unmeasured**, and the verdict can only be LIVE (transcript affirmative) or INDETERMINATE (transcript negative or unmeasured), never DEAD. A registry pid probe alone shall never decide liveness.

- **REQ-HYG-008 (Unwanted):** The GC shall not delete the state of a LIVE or INDETERMINATE session, and the run report shall state the keeping reason for every kept candidate.

- **REQ-HYG-009 (While apply mode, When a candidate is DEAD):** While the hygiene mode is `apply`, when a candidate is DEAD and its age is datable **from the timestamp recorded in its own content** per the REQ-HYG-005 class table, and that age meets the minimum-age floor (default 7 days), the GC shall delete the candidate and record an audit row naming the path, the reason, and the signal evidence. A candidate whose body carries no recorded timestamp — or whose class datum is missing from it — is **undatable and spared regardless of its mtime**; file mtime is never a deletion datum (the measured bulk-touch caveat and preserved-mtime hazard make mtime-only dating unsafe in both directions). Deletion is idempotent: an already-vanished path is recorded as `already-gone`, never as an error.

### §B.3 Execution and safety

- **REQ-HYG-010 (While report mode):** While the hygiene mode is `report` (the shipped default), the rotator and the GC shall produce decisions and summary rows without modifying any candidate file, without rotating anything (including the hygiene sink — REQ-HYG-004), and without creating any lockfile; the lock-class acquire attempt (REQ-HYG-014) is likewise apply-mode-only — report mode records `would-probe` for lock candidates. The project tree is byte-identical after a report run except for at most one summary row per unit appended to `.moai/logs/hygiene-audit.jsonl`.

- **REQ-HYG-011 (Unwanted):** The GC shall not evaluate, probe, acquire, or delete any `spec-close-*.lock` file: **the lock class is excluded from the target registry in this SPEC's scope** and any lock-named file encountered in a scan is reported `lock-class-excluded` and left untouched. Two verified reasons fix the exclusion: lock files are created empty (O_EXCL, `internal/spec/lock.go`), so content-only dating can never establish their age; and removing a lock whose inode another holder still has open re-admits a second live lock at the recreated path (a two-live-locks window, demonstrated in plan-audit iteration 2) unless the lock writer re-verifies the inode after acquire — a writer-side change outside this SPEC's module (`internal/hygiene`). Stale-lock reclamation is an upstream concern (decision-index Q7); the probe-then-remove and acquire-then-unlink shapes are prohibited on this class by the same exclusion.

- **REQ-HYG-012 (Ubiquitous):** The rotator and the GC shall fail independently: a rotator error shall not prevent the GC run, and a GC error shall not prevent the rotator run; each reports its own outcome.

- **REQ-HYG-013 (Where the operator has not opted in on the invoking surface):** Where `workflow.hygiene.mode` is absent or `report`, the SessionStart auto path shall not rotate or delete anything. The CLI mutates only when `--apply` is passed on that invocation — `--apply` overrides the config mode **for that CLI invocation only**; the config mode alone never makes a CLI invocation mutate. Absent the opt-in on its own surface, no unit rotates or deletes on any surface — the hygiene audit sink included (REQ-HYG-004).

- **REQ-HYG-014 (When a SessionStart hook fires):** When a SessionStart hook fires, the hygiene engine shall run the rotator and then the GC in the configured mode, best-effort: bounded scans (stat-based enumeration, with per-candidate probes — fingerprint reads, content-timestamp parsing, and the lock-named scan for the excluded class — each cheap and per-path), per-unit errors logged and never propagated, and session launch never blocked or delayed beyond the hook timeout budget.

- **REQ-HYG-015 (Ubiquitous) [HARD]:** Every rotator and GC test shall construct its fixture tree exclusively under `t.TempDir()`; no test shall read, write, rotate, or delete the repository's own `.moai` tree; the real repository shall ever see only report-mode output from its own sessions' hooks unless the operator has set `apply`; and — enforced at runtime, not only by convention — when the process is a Go test (`testing.Testing()`), the units' entry points shall refuse any root that is not inside the test's temporary directory.

- **REQ-HYG-016 (Ubiquitous):** Every size, count, age, window, and mode threshold shall live as a named constant in `internal/config/defaults.go` with a `workflow.yaml` config-key override; no call site shall carry a magic number. The named set: `HygieneAuditLogMaxBytes` (10 MiB), `HygieneAuditLogKeptRotations` (1), `HygieneTranscriptActivityWindow` (48 h), `HygieneHeartbeatStaleWindow` (24 h), `HygieneMinAgeDays` (7), mode default `report`.

## §C Constraints

- **Hook budget:** SessionStart runs under the 5 s MoAI policy budget. Enumeration is stat-based; per-candidate probes are cheap reads; nothing retries or waits.
- **Concurrency:** sink appenders take no lock (plain `O_APPEND` writers across hook processes). The rotator's lockfile serializes rotators per sink; the rotation decision is made under the lock (REQ-HYG-002). Named residuals, each documented in code, not silently assumed: (1) rows appended by a process holding the **displaced previous chunk's** inode open at the moment of its removal are lost with the inode — the window opens only while a staged replacement exists; (2) the symlink re-check's final window (REQ-HYG-006) — a parent swapped inside it is caught on the next pass; (3) report mode grows the hygiene sink at summary-row rate until the first apply run (REQ-HYG-004).
- **Platform:** Windows rotation requires a cross-process sidecar lock (LockFileEx-based); where it cannot be taken or verified, rotation skips (`skipped-locked`/`skipped-platform`) and retries — no destruction on unverified exclusivity. Rename-over-open-file failures (sharing violations) skip and retry likewise. The spec-close lock class is excluded everywhere (REQ-HYG-011).
- **Symlinks:** refusal scoped strictly below the resolved `.moai` root (REQ-HYG-006); the GC suite carries a parent-swap test including the swap-after-check residual arm.
- **Compatibility:** no sink writer changes and no `internal/spec` writer changes (the lock-class exclusion is what keeps that true). Readers of the primary sink files are unaffected (rotation preserves the primary path). `agent-model-audit.jsonl` keeps its existing whole-file age-out (REQ-OBH-002); rotation and age-out are orthogonal.
- **Config:** keys under `workflow.yaml` (`hygiene:` block); named defaults in `internal/config/defaults.go` per the hardcoding-prevention convention.
- **Safety:** closed registries on both units; fail-closed liveness (registry-absent, entry-missing, transcript-absent, missing-fingerprint all explicitly unmeasured classes); mtime never a deletion datum; per-class dating table (REQ-HYG-005) with undatable⇒spared; report-mode default; idempotent deletion; per-file failures skip the file (the SessionStart path must never gate a launch on cleanup — same contract as `internal/factory/record_prune.go`).

## §D Scope and Exclusions

In scope: the two units, their registries, the liveness evaluator, config keys and defaults, SessionStart wiring, the `moai clean` flag extension, the guard tests, and documentation.

### Out of Scope — separate cards (operator-assigned)

- Memory stores and the MEMORY.md lifecycle (hygiene audit Area 2 — separate card).
- Worktree accumulation and reaping (hygiene audit Area 3 — H1–H7 cards).
- SPEC archive lifecycle and `spec.Locate` resolution (hygiene audit Area 4 — separate card).
- Reports-tree lifecycle (card-close-date archiving, `reports/worktrees` folding, loose `session-*.md` — M5 separate card).

### Out of Scope — already-owned or writer-owned surfaces (verified in-tree)

- `lessons-inbox.jsonl` rotation — owned by `internal/hook/inbox_lifecycle.go` and the `moai inbox` verbs.
- `harness/usage-log.jsonl` retention — owned by `internal/harness/retention.go` (30-day prune → monthly gz), working as measured.
- Per-session trace files (`logs/trace-*.jsonl`) — owned by `PruneObservationLogs` (REQ-OBH-002) plus the trace writer's own 10 MB rotation (REQ-OBS-006).
- `backups/`, `evolution/telemetry`, `.moai/cache/` retention, and local-branch reaping — no session-keyed shape; separate concern.
- `state/handoff/consumed/` records — the resume-message audit trail; kept by design.
- `spec-close-*.lock` reclamation — excluded here (REQ-HYG-011); requires writer-side inode re-verification in `internal/spec` (decision-index Q7).
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

- **Rotator pass:** acquire the cross-process lock → recover any orphan `<name>.1.staging` (complete the interrupted placement; a staged chunk is never deleted or overwritten) → **inside the lock** re-stat size and existence → skip when absent or under threshold → else rename primary to `<name>.1.staging` → remove the existing `<name>.1` (a staged replacement now exists) → promote `<name>.1.staging` to `<name>.1` → release → append the audit row. Windows: the same sequence only under a taken-and-verified LockFileEx sidecar lock; otherwise `skipped-locked`/`skipped-platform`. A stale pre-lock observation leads to a no-op skip, never to a removal. File-count bound per sink: primary + keptRotations chunks (REQ-HYG-001).
- **GC pass:** enumerate target-registry candidates → resolve the session key (36-char hyphenated UUID; lock-named files ⇒ `lock-class-excluded`) → evaluate the three-signal liveness verdict (registry absent/entry missing ⇒ pid+heartbeat unmeasured; transcript absent under a resolvable root ⇒ transcript unmeasured; heartbeat against the 24 h staleness window) → decide (report both modes) → in apply mode delete only DEAD + content-datable (per-class table) + age-floor-met, entry-by-entry for directory classes, grouped for the goal triple → append rows (report mode: one summary row per unit) → emit the run report (kept candidates carry their keeping reason).
- **Liveness signals:** (a) registry pid probe + process start-time fingerprint (reuses the `internal/session` probe seam and the `internal/homestate` fingerprint machinery; a missing/unreadable fingerprint leaves the pid signal unmeasured); (b) transcript mtime inside the 48 h window (absent-under-resolvable-root ⇒ unmeasured); (c) heartbeat recency against `HygieneHeartbeatStaleWindow`. Any affirmative → LIVE; all measurable and negative → DEAD; any unmeasured with none affirmative → INDETERMINATE (kept, reason reported).
- **Modes:** `report` (default) / `apply`; the auto path is gated by `workflow.hygiene.mode`, the CLI by `--apply` for that invocation (REQ-HYG-013). All rotation, including the hygiene sink's own, is apply-mode-only.
- **Known limitation (D25, recorded):** the fail-closed verdict rules mean most of the measured residue (entry-missing majority; transcript-absent classes) is reclaimed only when an affirmative DEAD case is provable — the expected reclaim fraction is well below the motivating 1,928-file figure, and the evidence-heavier second DEAD route (authoritative transcript-root proof) is parked at decision-index Q8.

## §G Provenance / Evidence Basis

- Hygiene audit (read-only, 2026-10-04): `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/8cc9d4cb-10a5-401b-b43b-b8829dd40c92/scratchpad/hygiene/hygiene.md` + `hygiene.json` (full tables). Installed-build caveat recorded there applies to its CLI-sourced figures; the sink sizes and state counts above are its independently measured git/python figures.
- Code owners read in this plan phase: `internal/hook/prune_logs.go` (+ `session_end.go:116` wiring), `internal/hook/trace/writer.go` (rotation precedent, 10 MB), `internal/harness/retention.go` (locked-prune precedent + documented residual window), `internal/factory/record_prune.go` (conservative GC contract), `internal/session/session_pid.go` (three-valued liveness probe seam), `internal/session/registry.go`, `internal/homestate/process_fingerprint_*.go` (pid-reuse defense), `internal/cli/worktree/sweep_cwd_posix.go` (lsof fail-closed precedent), `internal/hook/inbox_lifecycle.go` (lessons-inbox already owned), sink writer constants (§B.1 list), state-shape writers (`internal/statusline/context_usage.go` — `captured_at`, `internal/goal/state.go` — `recorded_at`, `internal/cli/codex_stop_chain.go`, `internal/codexwiring/stop_budget.go`, `internal/hook/agent_stop_guard.go` — `stopped_at`, `internal/harness/routing/types.go` — `timestamp`, `internal/verify/store.go` + `schema.go`, `internal/spec/lock.go` — flock(2) Unix / O_EXCL Windows, empty-file creation, no post-acquire inode re-verification).
- Plan-audit iteration 1 (FAIL 0.73, audited `8039ea714`): D1–D14 folded in v0.2.0. Iteration 2 (FAIL 0.81, audited `58c01b4e1`): D15–D22 blocking + D23–D25 optional folded in this v0.3.0 — report-mode rotation moved to apply-only resolving the four-REQ collision (D15, one decision in REQ-HYG-004); staging/displaced artifact names made rotator-owned registry entries with under-lock crash recovery (D16); Windows rotation gated on a verified cross-process sidecar lock, else skip+retry (D17); lock class excluded from GC scope — empty-file dating impossible + two-live-locks window, writer-side fix out of module (D18); symlink refusal scoped strictly below the resolved root with the swap-after-check residual named (D19); transcript-absent ⇒ unmeasured row + `HygieneHeartbeatStaleWindow` constant (D20); per-class shape/dating/grouping table with M3 field-pinning RED tests (D21); runtime `testing.Testing()` root guard + escape-pattern greps (D22); rotation-count bound phrasing (D23); report-mode `would-probe` + hash-exclusion statement (D24); reclaim-expectation limitation + Q8 (D25).
- Card t1508 lineage (SPEC-AGENTS-CONTRACT-001) is on a held branch and is NOT a dependency of this SPEC (D14: drop or resolve at close).

## §H Cross-References

- SPEC-OBSERVE-HYGIENE-001 — the age-based prune this SPEC complements (rotation bounds size; REQ-OBH-002 bounds age). Not superseded.
- SPEC-WORKTREE-SWEEP-001 (REQ-WS-006) — the fail-closed unanswerable-predicate disposition adopted for unmeasurable signals.
- `internal/factory/record_prune.go` — the datable-or-spare GC contract reused for REQ-HYG-009.
- `internal/spec/lock.go` — the lock-class platform semantics whose gaps (empty-file dating, no post-acquire inode re-verification) ground the REQ-HYG-011 exclusion.
- decision-index.md — Q1–Q6 design decisions plus Q7 (lock-class upstream reclamation) and Q8 (evidence-heavier DEAD route).

## HISTORY

- 2026-10-05 — v0.3.0 by manager-spec (card t1518, plan-audit iteration-2 remediation, final iteration). D15–D22 blocking + D23–D25 optional folded per the verdict's numbered fixes; REQ count held at 16 (ceiling), AC count 16; the D15 resolution is one decision recorded in REQ-HYG-004 (all rotation apply-mode-only, no exemption clauses), and the D18 resolution excludes the lock class from GC scope (REQ-HYG-011) with the upstream path parked at decision-index Q7.
- 2026-10-05 — v0.2.0 by manager-spec (iteration-1 remediation). FAIL 0.73 verdict folded: MP-8 four-element evidence ledger from direct execution on the pre-implementation tree (8039ea714); REQ set consolidated 19 → 16 (Tier M ceiling); D2–D14 applied.
- 2026-10-05 — v0.1.0 created by manager-spec (card t1518, plan phase, worktree `WT-audit-log-gc` @ develop `6643c7bba`). Evidence basis: 2026-10-04 hygiene audit. Status: draft.
