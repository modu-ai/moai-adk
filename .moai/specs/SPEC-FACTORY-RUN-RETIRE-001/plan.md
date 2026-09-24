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
  new `internal/homestate/factory_run_retire.go`, `internal/factorymsg/store.go`, a new
  **build-tag-free** `internal/cli/factory_run_owner.go` (the REQ-002b restamp seam),
  `internal/cli/launch_exec_windows.go`, `internal/cli/codex_direct_windows.go` and
  `internal/cli/codex_launcher.go` (the three call sites
  that invoke that seam), `internal/cli/factory_handoff_recover.go` (the existing `moai factory`
  command group), a new `test/integration/harness/it08_factory_run_retire_test.go`, plus package
  tests. Files owned by t1082 and t1109 are listed in spec.md §E; touching one halts with a blocker
  report.
- **`codex_launcher.go` joined the set at v0.4.0** (the pane door, D11). It is not in either sibling
  card's set: t1082 owns `factory_lane_handoff*.go`, `handoff*.go`, `mcp_codex.go`,
  `defaults.go`; t1109 owns `hook/factory_messages*.go`. Re-confirm before integration.
- **`codex_direct_windows.go` IS in the set — spawn shape, seam call required.** Measured shape at
  `internal/cli/codex_direct_windows.go:14-30`: `cmd.Start()` →
  `ProbeProcessIdentity(cmd.Process.Pid)` → `registerFactoryLaunchPending(..., cmd.Process.Pid,
  start)` → `cmd.Wait()`. The launching process survives the launch, so REQ-002b binds this door.
  It needs **no rule of its own** — the existing spawn rule (§F M1) already covers it — but it does
  need the **seam call**, because the restamp only happens where a call site makes it happen.
  Without that one edit, the Windows `moai codex -f` door leaves the launcher's identity on the run
  row, and a launcher that dies while its codex child lives gets that run retired. "Needs no new
  rule" and "needs no edit" are different claims; conflating them was defect D18. Neither sibling
  card owns this file (t1082: `factory_lane_handoff*.go`, `handoff*.go`, `mcp_codex.go`,
  `defaults.go`; t1109: `hook/factory_messages*.go`) — re-confirm before integration.
- **`codex_direct_posix.go` is deliberately NOT in the set — replace shape, no edit.**
  `syscall.Exec` at `internal/cli/codex_direct_posix.go:39` replaces the launching process, and the
  record-time stamp at `:34` has already written `os.Getpid()` +
  `homestate.CurrentProcessFingerprint()`, so that stamp already IS the session identity and there
  is no restamp to add. Its absence is a stated conclusion, not an oversight.
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
- **Decision recorded here for review — the owner is the SESSION process, and the branch is by
  DOOR SHAPE, not by platform.** Three shapes, one rule (spec.md §A.1):
  - **replace** (`launch_exec_posix.go:33`, `codex_direct_posix.go:34`) — `syscall.Exec` preserves
    PID and start time, so the record-time stamp already IS the session identity. No restamp.
  - **spawn** (`launch_exec_windows.go:54`, `codex_direct_windows.go:24`) — the launcher restamps
    with `child.Process.Pid` + `childFingerprint`, the same two values it already hands
    `registerFactoryLaunchPending`, right after the child's identity is probed live.
  - **pane** (`codex_launcher.go:230`) — the launcher restamps with the tmux pane identity that
    `defaultCodexSpawnPaneIdentity` has already resolved live, before it returns and exits.

- **The seam carries no build tag, and that is load-bearing, not tidiness.** Put the restamp in a
  new `internal/cli/factory_run_owner.go` taking `(root, runID, pid, fingerprint)`; the three call
  sites each pass the identity they resolved. Writing it inside `launch_exec_windows.go` would trap
  it behind `//go:build windows`, where a darwin host cannot compile a call to it — and the pane
  door that needs the identical restamp is on darwin. That is D15.

  An earlier draft claimed one unconditional stamp worked on every shape. That was measured false
  (spec.md §A.1) on spawn and again on pane: a launcher that dies while its session survives leaves
  a stamped identity that probes dead, its live session's run gets retired, and the column and the
  peer fallback name different processes. One seam plus three call sites removes all of it.

  The pre-restamp window is correct rather than tolerated: between `RecordRun` and the restamp the
  launcher IS the only process, so a launch that fails in that window leaves a row whose stamped
  owner dies with it and is correctly reaped — which is exactly the rejected-alternative (e)
  rationale the stamped column exists for. On the pane door that window is wider — from
  `recordFactoryRunStart` (`codex_launcher.go:515`) until the pane identity resolves inside
  `runCodexLaunch` — and REQ-002d governs its failure exit: refuse, and leave no run carrying the
  launcher's identity.

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
  classified `live` **or `indeterminate`** (REQ-005, which binds every retirement path — the command
  shares the reconciler's predicate rather than carrying a second, laxer copy of it. The laxer copy
  was defect D14).
- No interactive prompt: the subagent boundary forbids it, and the command is operator-invoked.

### M5 — Door execution evidence, pane door included

- Exercise `moai cc -f`, `moai glm -f`, `moai codex -f`, and **`moai codex -f --spawn`** in the
  isolated sandbox, capturing the `runs` table plus the stamped owner against the run's lead peer
  after each (AC-011). This milestone exists because every door was **read, not run** during
  reproduction; re-reading the source does not close it.
- The `--spawn` leg needs a live tmux server in the sandbox. Where the sandbox cannot provide one,
  that is a **blocker report naming the obstacle**, never a silent downgrade to a source read — the
  pane door is the one this SPEC's scope was extended to cover.
- Exercise the REQ-002d refusal too (AC-012): force the identity resolver to fail, then confirm the
  launch exits non-zero and no `runs` row carries the launcher's PID.
- `codex_direct_posix.go` is **replace**-shaped: it needs no edit and no execution evidence here.
  `syscall.Exec` preserves the identity the record-time stamp already named (spec.md §A.1), so
  there is nothing for this milestone to exercise.
- `codex_direct_windows.go` is **spawn**-shaped and therefore **is** an edit target — it carries the
  REQ-002b restamp call like every other non-replace door. Only its *execution* evidence is deferred:
  this is a darwin host and that door sits behind `//go:build windows`, so the run arrives post-merge
  on the three-OS `test-integration` job. Its restamp call is covered **pre-merge** by AC-016's
  source-level leg, which reads the call site rather than executing the door.

### M6 — Migration, mutation, and cross-platform placement

- Legacy-row migration path: a v2 database with two unstamped `active` rows, reconciled via the peer
  fallback.
- Both mutation directions (AC-015a / AC-015b).
- **Cross-platform placement now has a verified green path (spec.md §A.2).** The liveness/reconciler
  exercise goes at `test/integration/harness/it08_factory_run_retire_test.go` behind
  `//go:build integration` — the only path the three-OS `test-integration` job runs. That job has no
  `if:` of its own and inherits its gate through `needs: test`, so a develop push touching Go code
  runs it on ubuntu, macos, and windows; run `35802361895` is the confirming observation.
- **Verify selection, not just presence.** Close leg 1 of AC-013 with
  `go test -tags=integration -v ./test/integration/harness/ -run TestFactoryRunRetire` and confirm a
  `--- PASS: TestFactoryRunRetire` line. A selector matching zero tests exits 0 and prints `ok`, so
  a bare `ok` would let the three-OS job run nothing while every surface reported green.
- **Do not claim the three-OS verdict at close.** `ci.yml` has no `pull_request` trigger for
  `develop` and this project does not push `WT-` branches, so leg 2 lands only on the develop push
  after integration. Record darwin locally; record leg 2 as pending with its run id to follow.
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
