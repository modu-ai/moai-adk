# plan.md — SPEC-FACTORY-RUN-RETIRE-001 (card t1107)

## §A Context

Implementation plan for owner-liveness reconciliation of factory runs. The requirements layer is
`spec.md` §B; the verification layer is `acceptance.md`. This file carries the approach, the
milestone order, and the constraints the run phase must not cross.

Baseline: branch `WT-factory-run-retire`, base local develop `176d8b658`, worktree
`.claude/worktrees/t1107`.

Milestones are ordered by **decision reversibility** — the data-model change first, then the new
interface, then the user-facing surfaces, with verification last. A reviewer reading top-down meets
the decisions most expensive to reverse first.

## §B Known issues in the current code

- `runs` carries no owner identity (`internal/homestate/factory.go` DDL).
- `RecordRun` (`internal/homestate/runtime.go:23`) is the sole `runs.status` writer and only ever
  writes `'active'`.
- `ResolveActiveRun` (`internal/factorymsg/store.go:241`) counts active rows and fails closed.
- `recordFactoryRunStart` (`internal/cli/factory.go:258`) runs **before** `execOrSpawnClaude`, so a
  launch that fails after the record leaves a stale `active` row with no lead peer at all.

## §C Pre-flight

1. Confirm the sibling-card file sets are still disjoint from this SPEC's set (see §D) before the
   first edit, and again before integration.
2. Confirm `.moai/reports/t1107/verdict.md` is present and readable — it is the SPEC's evidence base
   and the run phase cites it rather than re-deriving the mechanism.
3. Re-measure the two instrument facts from §A of spec.md in the run tree before M1 (the zero-hit
   `UPDATE runs` search plus its `status='active'` control), so the "no retirement path exists"
   premise is attributed to the run-phase tree rather than carried over.

## §D Constraints

- **[HARD] Never create a path that retires a live lead's run.** Liveness is PID **plus**
  process-start identity. `indeterminate` is treated as live.
- **[HARD] Do not weaken `ResolveActiveRun`'s fail-closed behaviour.** Reconciliation may only
  remove provably-dead owners from the active set.
- No new runner and no new state store. The `runs` table and `factory.db` already exist; this SPEC
  adds columns to `runs` and a status value, nothing more.
- Reuse before writing: `homestate.ProbeProcessIdentity`, `homestate.CurrentProcessFingerprint`,
  `platformPIDState`, and the existing `migrateFactoryV1ToV2` migration shape are all in service and
  are the intended building blocks.
- File set owned by this SPEC: `internal/homestate/factory.go`, `internal/homestate/runtime.go`, a
  new `internal/homestate/factory_run_retire.go`, `internal/factorymsg/store.go`,
  `internal/cli/launch_exec_windows.go` (the REQ-002b restamp call site),
  `internal/cli/factory_handoff_recover.go` (the existing `moai factory` command group), a new
  `test/integration/harness/it08_factory_run_retire_test.go`, plus package tests. Files owned by
  t1082 and t1109 are listed in spec.md §E; touching one halts with a blocker report.
- **Overlap status, stated honestly**: this tree cannot establish non-overlap with t1082, because
  t1082's files do not exist here. The zero-overlap figure is the team lead's measurement against
  t1082's own worktree (2026-09-23), cited as that. Re-confirm before integration, and treat the
  same-package `internal/factorymsg/store.go` clash as a live residual (spec.md §E).
- Template neutrality (C1-C8) applies to anything under `internal/template/templates/`; `make build`
  after any template edit; `make agents-emit` after any agent `.md` edit. Neither is expected here.

## §E Self-verification

Each milestone closes only when its ACs are recorded in `progress.md` §E.2 with the command run and
its verbatim output, attributed to that run and that tree.

## §F Milestones

### M1 — Owner identity on the run record (data-model change; least reversible)

- Add `lead_pid INTEGER NOT NULL DEFAULT 0` and `lead_process_start TEXT NOT NULL DEFAULT ''` to the
  `runs` DDL; bump `factorySchemaVersion` 2 → 3 with a `migrateFactoryV2ToV3` following the shape of
  the existing `migrateFactoryV1ToV2` (`ALTER TABLE ... ADD COLUMN` inside one transaction, then the
  `meta.schema_version` update).
- Extend `homestate.FactoryRun` with the two fields and stamp them in `RecordRun`'s existing
  transaction. The caller supplies `os.Getpid()` and `homestate.CurrentProcessFingerprint()`.
- Add the `retired` status value and a `run.retired` event kind. Retirement is an `UPDATE` of
  `status`; the row is never deleted (REQ-010).
- **Decision recorded here for review — the owner is the SESSION process, and the platform branch is
  explicit.** On POSIX `syscall.Exec` preserves PID and start time, so the launcher stamp already
  IS the session identity and no restamp is needed. On Windows `child.Start()` creates a distinct
  session process, so the launcher restamps the row with `child.Process.Pid` + `childFingerprint` —
  the same two values it already hands `registerFactoryLaunchPending`
  (`launch_exec_windows.go:50,55`) — immediately after the child's identity is probed live and
  before the peer registration.

  An earlier draft claimed one unconditional stamp worked on both shapes. That was measured false
  (spec.md §A.1): a Windows launcher killed while its child survives would probe dead and its live
  session's run would be retired, and the column and the peer fallback would name different
  processes. The branch costs about ten lines at one call site and removes both.

  The pre-restamp window is correct rather than tolerated: between `RecordRun` and the restamp the
  launcher IS the only process, so a launch that fails in that window leaves a row whose stamped
  owner dies with it and is correctly reaped — which is exactly the rejected-alternative (e)
  rationale the stamped column exists for.

### M2 — The liveness predicate and the reconciler (new interface)

- `internal/homestate/factory_run_retire.go`: a classifier over a run row returning
  `live | dead | indeterminate`, and `ReconcileActiveRuns` which retires the `dead` ones and reports
  what it retired and what it left.
- **Interface decision for review**: `homestate` cannot import `factorymsg` (that would be an import
  cycle — `factorymsg` imports `homestate`), so the legacy-row fallback of REQ-006 cannot live in
  `homestate`. `ReconcileActiveRuns` therefore takes a fallback lookup as a function parameter,
  `func(runID string) (pid int, processStart string, ok bool)`, and `factorymsg` supplies the
  closure that opens the run's broker and reads its `role='lead'` peer. Rejected alternative: moving
  the whole reconciler into `factorymsg`, which would put a `runs`-table writer outside the package
  that owns the schema.
- A run with `lead_pid = 0` consults the fallback; a run with neither source is `indeterminate`.

### M3 — Resolver wiring and the failure message (user-facing)

- `ResolveActiveRun`: on `len(runs) > 1`, reconcile, re-query, then apply the unchanged 0/1/many
  switch. Zero stays `NO_ACTIVE_FACTORY`; many stays `AMBIGUOUS_FACTORY`.
- The `AMBIGUOUS_FACTORY` error text gains the remaining run ids with their classifications, so the
  operator reads *why* each survived rather than an opaque refusal. The sentinel substring
  `AMBIGUOUS_FACTORY` is preserved — existing tests match on it.
- The explicit-`--factory-run` branch is unchanged: it validates one named run and returns.

### M4 — The operator maintenance surface (user-facing)

- Extend the existing `moai factory` command group (`internal/cli/factory_handoff_recover.go`
  `newFactoryCommand`) with `runs`: list every run with status, owner classification, and
  timestamps; `--retire <run-id>` retires exactly the named run and refuses when its owner is
  classified `live` (REQ-009).
- No interactive prompt: the subagent boundary forbids it, and the command is operator-invoked.

### M5 — Three-door execution evidence

- Exercise `moai cc -f`, `moai glm -f`, and `moai codex -f` in the isolated sandbox and capture the
  `runs` table after each. This milestone exists because the three doors were **read, not run**
  during reproduction; re-reading the source does not close it.

### M6 — Migration, mutation, and cross-platform placement

- Legacy-row migration path: a v2 database with two unstamped `active` rows, reconciled via the peer
  fallback.
- Both mutation directions (AC-015a / AC-015b).
- **Cross-platform placement, not a cross-platform verdict.** The liveness/reconciler exercise goes
  at `test/integration/harness/it08_factory_run_retire_test.go` behind `//go:build integration` —
  measured as the only path the three-OS `test-integration` job runs
  (`go test -tags=integration ./test/integration/harness/...`, `ci.yml:381`). The unit `test` job is
  ubuntu-only (`ci.yml:124`) and `ci.yml` has no `pull_request` trigger for `develop`, so the
  three-OS result arrives on the **develop push after integration**, not before. The run phase
  records darwin locally and records the rest as deferred — it does not claim a verdict it cannot
  obtain (AC-013).
- Affected-package tests: `go test ./internal/homestate/... ./internal/factorymsg/... ./internal/cli/...`.

## §G Anti-patterns

- Making `AMBIGUOUS_FACTORY` rarer by choosing a run. That is not a fix; it is the defect wearing the
  fix's clothes.
- Treating `indeterminate` as dead to make a test pass. The asymmetry is deliberate: a stale run has
  a working escape (`--factory-run`); a wrongly-retired live run does not.
- A test that opens factory state without the REQ-012 isolation. It will silently act on the
  developer's real project via `CanonicalProjectRoot` and its green result will mean nothing.
- Running the full suite locally to judge this change. Run the affected packages; CI judges the rest
  (CLAUDE.local.md §4).

## §H Cross-references

- `.moai/reports/t1107/verdict.md` — the reproduction, and the evidence base for §A of spec.md.
- `internal/homestate/profile_lease.go:202` — `ProbeProcessIdentity`, the liveness primitive.
- `internal/cli/factory_launch_pending.go:34` — the lead peer registration the REQ-006 fallback reads.
- `internal/homestate/factory.go` `migrateFactoryV1ToV2` — the migration shape M1 follows.
