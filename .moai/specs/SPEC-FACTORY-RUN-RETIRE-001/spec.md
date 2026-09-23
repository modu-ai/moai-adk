---
id: SPEC-FACTORY-RUN-RETIRE-001
title: "Factory run retirement — owner-liveness reconciliation so a dead lead's run leaves 'active'"
version: "0.1.0"
status: draft
created: 2026-09-23
updated: 2026-09-23
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/homestate, internal/factorymsg, internal/cli"
lifecycle: spec-anchored
tags: "factory, run-lifecycle, liveness, ambiguous-factory, migration, card-t1107"
tier: M
card: t1107
related_specs: [SPEC-FACTORY-MIXED-HOOK-001, SPEC-FACTORY-WORKER-NAMING-001, SPEC-FACTORY-MODE-001]
---

# SPEC-FACTORY-RUN-RETIRE-001 — Factory run retirement

## §A Background and Motivation

A factory run is recorded `status='active'` and **no non-test code path ever takes it out of that
state**. Launching the factory lead twice in the same project in different seconds therefore leaves
two `active` rows, and every subsequent worker join fails closed with `AMBIGUOUS_FACTORY` until the
operator passes `--factory-run <id>` explicitly.

The defect is reproduced, not inferred. The evidence-bearing reproduction report is
`.moai/reports/t1107/verdict.md` (card t1107, measured 2026-09-23 on darwin against a
`make build` binary of local develop `176d8b658`); its §4 Gaps and §5 Residual-risk are inputs to
this SPEC and are addressed or explicitly carried forward in §D and §F below.

Measured mechanism (re-measured in this tree while authoring this SPEC):

- `internal/kanban/bootstrap.go:97` — `NewRunID()` is `base36(time.Now().Unix())`, **second**
  granularity. Two lead launches in the same second collapse to one row through the
  `ON CONFLICT(run_id) DO UPDATE` clause; the trigger is specifically a **different-second**
  relaunch.
- `internal/homestate/runtime.go:23` `RecordRun` — `INSERT ... 'active' ... ON CONFLICT DO UPDATE
  SET status='active'`. This is the only writer of the `runs.status` column.
- `internal/factorymsg/store.go:274-276` — two or more `active` rows produce `AMBIGUOUS_FACTORY`.
- `grep -rn "UPDATE runs\|DELETE FROM runs" --include='*.go' . | grep -v _test.go` → **0 hits**
  (exit 1). Control with the same search shape: `status='active'` → 2 non-test hits, so the
  instrument works and the absence is not a search failure.

Two further facts, measured in this tree, shape the prescription:

1. **The `runs` row carries no owner identity.** The `runs` DDL
   (`internal/homestate/factory.go`) has `run_id`, `lead_session_id`, `lead_backend`, `status`,
   `manifest_json`, `created_at`, `updated_at` — no PID, no process-start fingerprint. Nothing in
   the row can answer "is this run's lead still alive?".
2. **A lead process identity IS already recorded elsewhere, per run.** `registerFactoryLaunchPending`
   (`internal/cli/factory_launch_pending.go:34`) registers a `role='lead'` peer carrying `PID` and
   `ProcessStart` into the per-run broker database (`peers` table, `internal/factorymsg/store.go`).
   This is the t1074 PID-plus-start-identity convention, already in service.

The recording process is a valid liveness proxy on both platform shapes: on POSIX
`execOrSpawnClaude` uses `syscall.Exec`, which preserves the PID and the process start time across
the replacement, so the PID stamped before the exec IS the session's PID; on Windows the launcher
spawns a child and blocks in `child.Wait()` for the session's whole lifetime, so the launcher PID is
live for exactly as long as the session is.

## §B Requirements (GEARS)

### Owner identity

- **REQ-001** (Ubiquitous): The factory run record shall carry the launching lead's process owner
  identity — a process id and a process-start fingerprint — so that a run's liveness is answerable
  from the run record itself.
- **REQ-002** (Event-driven): **When** a factory run start is recorded, the factory state shall
  stamp the recording process's PID and its process-start fingerprint onto that run's row in the
  same transaction that writes the row.
- **REQ-003** (Ubiquitous): The owner-liveness predicate shall classify a run's owner as exactly one
  of `live`, `dead`, or `indeterminate`, using PID **together with** process-start identity — a
  bare PID probe is insufficient because process ids are reused.

### Reconciliation

- **REQ-004** (Event-driven): **When** factory run resolution observes more than one active run,
  the resolver shall first retire every active run whose owner is classified `dead`, then re-resolve
  over the remaining active runs.
- **REQ-005** (Unwanted): The reconciler shall not retire a run whose owner is classified `live` or
  `indeterminate`.
- **REQ-006** (Event-detected): **When** a run row carries no owner stamp — the shape of every row
  written before this SPEC lands — the reconciler shall take the run's registered `role='lead'` peer
  identity (PID plus process start) as the liveness source, and shall classify the owner
  `indeterminate` when neither source yields an identity.
- **REQ-010** (Ubiquitous): Retirement shall be a status transition on the existing `runs` row
  (`active` → `retired`) that preserves the row and appends a `run.retired` event; it shall not
  delete the row.

### Failure surface

- **REQ-007** (Event-detected): **When** reconciliation leaves two or more active runs, the resolver
  shall still fail closed with `AMBIGUOUS_FACTORY`, and the failure shall name each remaining active
  run id together with its owner classification.
- **REQ-014** (Unwanted): The fail-closed semantics of run resolution shall not be weakened — zero
  active runs shall remain `NO_ACTIVE_FACTORY`, and two or more active runs after reconciliation
  shall remain `AMBIGUOUS_FACTORY`. Selecting "the newest" or "the first" active run is prohibited.

### Operator surface

- **REQ-008** (Capability gate): **Where** the operator invokes the factory run maintenance command,
  the CLI shall report every active run with its owner classification, and shall retire a run only
  when the operator names that run id explicitly.
- **REQ-009** (Unwanted): The maintenance command shall not retire a run whose owner is classified
  `live`, regardless of the operator naming it.

### Doors, tests, platforms

- **REQ-011** (Ubiquitous): Run recording and run resolution shall behave identically at the three
  lead entry doors — `moai cc -f`, `moai glm -f`, and `moai codex -f` — each of which reaches the
  same recorder.
- **REQ-012** (Ubiquitous): Every test or reproduction that opens factory state shall isolate
  `HOME`, `MOAI_HOME`, and the launched-binary path (`MOAI_CLAUDE_BIN`), and shall use a project
  directory outside this repository's worktree set, because `homestate.CanonicalProjectRoot`
  converges a linked worktree onto the primary checkout — a test that omits this isolation mutates
  the developer's live factory state rather than a fixture.
- **REQ-013** (Ubiquitous): The owner-liveness predicate and the reconciler shall be exercised on
  darwin, linux, and windows.

## §C Design Decision — the retirement mechanism

**Chosen: owner-liveness reconciliation at run-resolution time** (REQ-004). When resolution would
otherwise raise `AMBIGUOUS_FACTORY`, every active run whose owner is provably dead is retired first,
and resolution re-runs over what remains.

Why this one:

- It does not depend on the lead's cooperation. On POSIX the lead process is *replaced* by
  `syscall.Exec`, so no Go code of ours runs after the launch — there is no lead-exit callback to
  hang retirement on, and a crashed or `SIGKILL`ed lead would defeat one anyway.
- It fires at exactly the moment the defect is felt. The user-visible symptom is the worker join
  being refused; reconciling on that path fixes the symptom at its own trigger.
- It covers prevention and migration with one mechanism. Rows already accumulated are reconciled by
  the same pass that prevents new accumulation, which is what REQ-006 makes possible.
- It preserves fail-closed behaviour (REQ-014). Reconciliation only ever *reduces* the active set by
  removing provably dead owners; ambiguity that survives it is real ambiguity.

An explicit operator command (REQ-008/009) ships alongside it, not as an alternative: a run whose
owner is `indeterminate` is never auto-retired by design (REQ-005), so without a deliberate operator
surface such a row would be permanently unremovable. The command shares the one liveness predicate;
it is a second entry to the same mechanism, not a second mechanism.

### Alternatives rejected

| Alternative | Why rejected |
|---|---|
| **(a) Retire on lead exit** (SessionEnd hook, or a deferred call in the launcher) | On POSIX `syscall.Exec` replaces the process, so nothing of ours runs after launch — there is no exit point to hook in the launcher. A Claude Code SessionEnd hook could cover clean exits only: a crash, a `SIGKILL`, or a disabled-hooks session leaves the row active, which is the defect unchanged. Best-effort cleanup that silently misses the crash case is worse than none, because it makes the residue look like a rare anomaly rather than a systematic leak. |
| **(b) Run lifetime / TTL** | Wrong predicate for this domain. A lead session legitimately idles for hours between dispatches, so any TTL short enough to unblock a user promptly is short enough to retire a live lead's run — a direct REQ-005 violation. A TTL long enough to be safe does not unblock the user at all. Time is not evidence of death here; process identity is. |
| **(c) Explicit operator command alone** | Necessary but insufficient. The operator meets this defect as an opaque `AMBIGUOUS_FACTORY` refusal; requiring them to already know a maintenance verb exists puts the cure behind the same knowledge gap as the disease. Retained as the `indeterminate`-residue surface (REQ-008), demoted from primary. |
| **(d) Widen `ResolveActiveRun` to pick one of several active runs** (newest, or the one matching `MOAI_KANBAN_ID`) | Makes the symptom disappear without touching the cause, and silently joins a worker to an arbitrary run. Explicitly forbidden by the card constraint and by REQ-014: two active runs IS ambiguous, and the error is correct. |
| **(e) Owner identity from the broker `peers` row alone, with no `runs` column** | Simpler (no schema change) and it is the REQ-006 fallback — but it cannot see a run whose record was written and whose launch then failed before the peer was registered (`recordFactoryRunStart` at `internal/cli/cc.go:191` precedes `execOrSpawnClaude`). That path leaves a permanently-`indeterminate` row, i.e. a second generator of the same defect. The stamped column closes it; the peer lookup is kept as the legacy-row fallback where the column is empty. |

## §D Gaps carried from the reproduction

Each Gap in `.moai/reports/t1107/verdict.md` §4 is dispositioned here:

| Verdict gap | Disposition in this SPEC |
|---|---|
| Not reproduced with a real `claude` session (a stub was used) | Unchanged. The defect and the fix both live in the state layer *before* `exec`, so the run phase also verifies with a stub binary. Recorded as residual risk (§F), not covered. |
| GLM and Codex lead doors read but never executed | **Covered** — REQ-011 plus AC-011/AC-012, which are satisfied only by a recorded invocation and its captured `runs` output. |
| Lead-process-death behaviour never measured (the stub lead exits immediately) | **Covered** — REQ-003/004/005 plus AC-004/AC-005, which measure both the dead-owner and live-owner branches with a controlled long-lived process. |
| Accumulated scale in real use never measured | **Partly covered.** REQ-008's report surfaces the count on demand; this SPEC does not measure the existing population of any real installation, and does not need to — reconciliation is per-resolution, not a one-shot sweep sized to a population. |
| Windows never measured (darwin only) | **Covered as scope** by REQ-013 and AC-013 (CI matrix). Stated plainly: the originating reproduction is darwin-only. |

## §E Exclusions

### Out of Scope — the kanban backlog runtime record

- `kanban.RecordFactoryRunStart` also writes a `TodoRuntimeRun` into the backlog store. Measured:
  that record carries no status field (`internal/kanban/todo_runtime.go:20`) — it is a provenance
  seed, not a lifecycle state — so it has no `active` to leave and is not reconciled here.

### Out of Scope — worker roster and peer lifecycle

- Retiring, expiring, or reaping `workers` rows and broker `peers` rows. This SPEC reads the lead
  peer identity (REQ-006) and writes nothing to it. Worker-slot lifecycle is a separate axis.

### Out of Scope — run-id granularity

- Widening `NewRunID()` beyond second granularity. Second granularity is what makes the same-second
  case collapse harmlessly; changing it would widen the defect's trigger, not narrow it, and the fix
  here is independent of id granularity.

### Out of Scope — files owned by sibling cards

- `internal/cli/factory_lane_handoff*.go`, `internal/factorymsg/handoff*.go`,
  `internal/cli/mcp_codex.go`, `internal/config/defaults.go` (card t1082), and
  `internal/hook/factory_messages*.go` (card t1109). Measured overlap with this SPEC's file set is
  zero; if the run phase finds it needs one of these, it halts with a blocker report naming the file
  rather than editing it.

### Out of Scope — release activities

- No release cut, push, or PR creation. Integration follows the git-flow lane protocol.

## §F Residual risk

- The stub-binary verification shape is inherited from the reproduction: a real session's timing
  around the launcher's deferred restore paths is not exercised.
- `internal/factorymsg/store.go` sits in the same package as card t1082's `handoff*.go` edits.
  Different files, no measured overlap, but a same-package semantic clash at integration is possible
  and is resolved by the integrating lane.
- `ProbeProcessIdentity` returns `indeterminate` on any probe error. A host where probing routinely
  fails would never auto-retire anything; this degrades to today's behaviour plus an explicit
  operator surface, and never to retiring a live run.

## §G History

- 2026-09-23 — v0.1.0 — manager-spec — initial plan-phase draft (card t1107, Tier M, class C),
  authored from the reproduced defect in `.moai/reports/t1107/verdict.md`.
