---
id: SPEC-TEST-ENV-HERMETIC-001
title: "Test env hermeticity sweep — tests that read the factory/kanban lane gate axes must not change verdict with the ambient env of the session that runs them"
version: "0.4.0"
status: draft
created: 2026-10-03
updated: 2026-10-03
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/cli, internal/hook"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "test-hermeticity, env-isolation, factory-lane, ambient-env, recurrence-guard, test-only"
related_specs: [SPEC-CLI-TEST-CWD-ISOLATION-001, SPEC-FACTORY-SELF-DISPATCH-001, SPEC-AUTONOMY-PRECONDITION-001, SPEC-ROLE-NAMING-CODE-001, SPEC-STALE-RUN-LABEL-001]
---

# SPEC-TEST-ENV-HERMETIC-001

## §A Problem / Motivation

A test that reads a **multi-axis env gate** — a production predicate that decides on several
environment variables at once — and pins only some of the axes passes or fails depending on the
env of the session that runs it. A factory-lane session exports `MOAI_FACTORY_ROLE`,
`MOAI_FACTORY_WORKER`, `MOAI_FACTORY_WORKERS`, `MOAI_KANBAN_BACKEND`, `MOAI_KANBAN_ID` (and more),
so a test that is green on a plain developer shell turns red inside a lane, and the red looks like
a regression of whatever the lane just changed. It is a **fake red**: the failure carries no
information about the code under test. The hazard is the same shape as the cwd residue of
SPEC-CLI-TEST-CWD-ISOLATION-001 — a different answer, not an error — and it recurred twice on one
gate before this card (t1350 §6 and t1354, both `TestContractRoleScopedAllowWithoutLaneMarker`,
fixed in place by commit `b12538f0a`, which pinned all three axes of `contractLaneGate`).

This SPEC generalizes that one-test fix into a measured sweep plus a guard against recurrence.
Per the card, the sweep is **centered on observed red cases** — it does not blacklist or skip
the suite.

### Observed evidence

All figures below were first measured on tree `2de0a2cb6` (branch `WT-test-env-hermetic-sweep`,
no local change at measurement time) in a lane session whose ambient env carries
`MOAI_FACTORY_ROLE=lane`, `MOAI_FACTORY_WORKER=lane-6`, `MOAI_FACTORY_WORKERS=0`,
`MOAI_KANBAN_BACKEND=claude`, `MOAI_KANBAN_ID=tm9i7y`, `MOAI_AUTONOMY_TIER=fully-autonomous`,
`MOAI_FACTORY_AUTO_DISPATCH=auto` (plus `MOAI_KANBAN_SETTINGS_INJECTED=1` and an empty
`MOAI_FACTORY_CLEAR_POLICY`). Iteration 2 re-measured the decisive commands on HEAD `a5a63a0bc`,
whose Go sources are identical to `2de0a2cb6` (`git rev-list --count 2de0a2cb6..a5a63a0bc` = 1 and
that commit's `git diff-tree` lists five SPEC-directory paths, no Go file). The deciding command
and verbatim output of each item is carried in `acceptance.md` §D.0 (evidence ledger, ids
E-1..E-8). The earlier local-only baseline `.moai/reports/t1356/baseline.md` (gitignored by
operator directive 2026-09-14 — `.moai/reports/*`; it is never force-added) is context only: the
ledger carries every fact the committed record relies on, including the cli scrubbed-arm green
(E-1b), so the committed record does not depend on it.

1. **internal/cli — three observed reds, single responsible axis.** `TestTodoClaim_LaneGovernance`,
   `TestTodoClaimMCP_Mirror`, and `TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit` fail
   with `refused — lane boundary` when `MOAI_FACTORY_ROLE=lane` reaches the test binary (E-1), and
   all three pass when that one axis is unset (E-1b). Mechanism: the lane-refusal predicate in
   `internal/cli/factory_card.go:57-71` reads the role marker; the package `TestMain` clears 12 keys
   (`factoryAmbientEnvKeys`, `internal/cli/factory_test.go:27-40`: 11 family axes plus
   `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS`) and that set does **not** contain `MOAI_FACTORY_ROLE`.
   The family axes the set lacks are **six**: `MOAI_FACTORY_ROLE`, `MOAI_AUTONOMY_TIER`,
   `MOAI_FACTORY_CLEAR_POLICY`, `MOAI_FACTORY_AUTO_DISPATCH`, `MOAI_FACTORY_MANAGED`, and
   `MOAI_FACTORY_SLOW_LAUNCH_MS` (read at `internal/cli/factory_launch_timing.go:172`); production
   code in the package references all six (§A.6). `TestTodoClaim_LaneGovernance` pins only the
   lane-label axis (`t.Setenv(EnvMoaiFactoryWorker, "")`), which is exactly the partial pin the
   card names.
2. **internal/hook — two observed reds, responsible axes isolated by measurement.**
   `TestStaleRunNoticeLegacyLeaderSpelling` and `TestStaleRunNoticeLegacySessionRecord` fail
   (`staleRunNoticeFor = ""`) under a lane env (E-2). The package `TestMain`
   (`internal/hook/main_test.go:67`) clears `CLAUDE_PROJECT_DIR`, the profile-lease variables and
   the git env — and **none** of the factory axes. Plan-phase one-axis arms (E-3) show the
   verdict flips to PASS when **either** `MOAI_KANBAN_ID` **or** `MOAI_FACTORY_WORKERS` alone is
   unset, and stays FAIL when `MOAI_FACTORY_ROLE`, `MOAI_FACTORY_WORKER` and `MOAI_KANBAN_BACKEND`
   are unset together. The gate is therefore a **conjunction** of those two axes
   (`internal/hook/stale_run_gate.go:259-260`: run id non-empty AND fan-out env non-empty →
   run-state gate, which answers "" for a run that is not recorded active). The earlier baseline
   attributed this to six variables unset together; E-3 narrows it.
3. **Positive control.** `TestContractRoleScopedAllowWithoutLaneMarker` (all three axes of
   `contractLaneGate` pinned, `internal/hook/contract_sign_guard.go:144-152`) is green under the
   lane env (E-4): the ambient env alone does not redden a hermetic test, so the five reds above
   are the observed failures.
4. **Survey nominations are hypothesis.** A static read-only survey nominated roughly 257
   `moai todo`-family test functions and about 100 factory-surface functions in `internal/cli` as
   flip-prone through `MOAI_FACTORY_ROLE` (mechanism CERTAIN, per-function verdict PROBABLE) and
   two more hook tests; both of the two hook nominations measured **green** under the ambient env
   (the trace overstated them). Only the five reds above are measured. The remaining set is
   measured by the run phase (REQ-THE-003), not trusted from the survey.
5. **No shared seeding helper exists.** Each package carries overlapping, non-identical scrub
   helpers — hook: 6 (`clearKanbanEnv`, `scrubKanbanEnv`, `sdScrubRuleEnv`, `m3ScrubEnv`,
   `srlGateEnv`, `rebindEnv`); cli: 10 (`clearFactoryAmbientEnv`, `clearFactoryTestEnv`,
   `sdLaneEnv`, `sdClearLaneEnv`, `sdScrubLauncherEnv`, `laneEnv`, `factoryLaneEnv`,
   `qasLaneEnv`, `nmLaneEnv`, `clearKanbanLauncherEnv`) — one definition each, counted at HEAD
   `a5a63a0bc`. A precedent for a drift guard exists:
   `internal/cli/ptycaptest/selfcheck_test.go` keeps its `kanbanVars` in step with
   `internal/config/envkeys.go`, and `internal/config/envkeys_factory_role_test.go` already
   enumerates the hook guard's `os.Getenv` call sites against a closed set.
6. **The axis family and its reach.** The guard's family is read by one rule: the constants in
   `internal/config/envkeys.go` whose value starts with `MOAI_FACTORY_` or `MOAI_KANBAN`, plus
   `MOAI_AUTONOMY_TIER`. At HEAD `a5a63a0bc` that rule yields **17** constants (the same 17 as the
   name rule `Env(AutonomyTier|MoaiKanban*|MoaiFactory*|Factory*)`): `MOAI_AUTONOMY_TIER`,
   `MOAI_KANBAN`, `MOAI_KANBAN_SPEC`, `MOAI_KANBAN_ID`, `MOAI_KANBAN_LABEL`,
   `MOAI_KANBAN_SETTINGS_INJECTED`, `MOAI_KANBAN_LEAD_ADDR`, `MOAI_KANBAN_BACKEND`,
   `MOAI_KANBAN_CARD`, `MOAI_KANBAN_LEAD_NAME`, `MOAI_FACTORY_WORKERS`,
   `MOAI_FACTORY_SLOW_LAUNCH_MS`, `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_MANAGED`,
   `MOAI_FACTORY_ROLE`, `MOAI_FACTORY_CLEAR_POLICY`, `MOAI_FACTORY_AUTO_DISPATCH`. Per package
   (production = non-`_test.go` files in the package directory; a **reference** is either the
   identifier `config.<Name>` **or** the axis's quoted literal value — either form counts, so
   `os.Getenv("MOAI_AUTONOMY_TIER")` at `internal/cli/codex_sync_gate.go:262` is not missed; the
   counts below hold under either form, because the one quoted literal in `internal/cli` is also
   referenced by identifier and `internal/hook` production carries no quoted family literal at
   HEAD `669cf18c9`):

   | package | family axes referenced by production | covered by the test binary's start-up scrub today | not covered — decided by this SPEC |
   |---------|---------------------------------------|---------------------------------------------------|-----------------------------------|
   | `internal/cli` | **17** of 17 | **11** (Workers, Worker, Kanban, KanbanID, KanbanSpec, KanbanLabel, KanbanSettingsInjected, KanbanLeadAddr, KanbanBackend, KanbanCard, KanbanLeadName) | **6**: Role, AutonomyTier, ClearPolicy, AutoDispatch, Managed, SlowLaunchMS |
   | `internal/hook` | **13** of 17 (Kanban, KanbanSpec, KanbanID, KanbanLabel, KanbanSettingsInjected, KanbanLeadAddr, KanbanBackend, KanbanCard, KanbanLeadName, Workers, Worker, Role, AutoDispatch) | **0** | **13** (the four unreferenced axes — AutonomyTier, ClearPolicy, Managed, SlowLaunchMS — are outside the guard's judgment) |
   | `internal/discovery` | 1 (KanbanID) | no `TestMain` scrub | narrow lane-vs-scrubbed pair (E-7) |
   | `internal/cli/ptycaptest` | 9 (the Kanban axes) | own drift guard | not measured here (R4) |

   Reach, measured by one command per directory,
   `grep -lE 'config\.Env(AutonomyTier|MoaiKanban[A-Za-z]*|MoaiFactory[A-Za-z]*|Factory[A-Za-z]*)\b' <dir>/*.go`
   split into non-test and `_test.go` files at HEAD `a5a63a0bc`: `internal/cli` 24 production /
   55 test files; `internal/hook` 9 / 33; `internal/discovery` 3 / 2; `internal/cli/ptycaptest`
   1 / 1 (`internal/cli` production reaches 25 files when the one quoted literal
   `"MOAI_AUTONOMY_TIER"` at `codex_sync_gate.go:262` is counted). Test files referencing the five
   lane axes (`EnvFactoryRole`, `EnvMoaiFactoryWorker`, `EnvMoaiFactoryWorkers`,
   `EnvMoaiKanbanBackend`, `EnvMoaiKanbanID`): 46 in `internal/cli`, 28 in `internal/hook`.

## §B Scope

**In Scope**:
- Making the observed-red tests hermetic with respect to the lane gate axes, in `internal/cli` and
  `internal/hook` (REQ-THE-001, REQ-THE-002).
- Measuring — not assuming — which further tests flip under the lane env, by whole-package
  failing-set comparison, and fixing only those that measure red (REQ-THE-003, REQ-THE-004).
- A recurrence guard pair in each of the two packages — a coverage comparison and an
  applied-behaviour probe (REQ-THE-005, REQ-THE-009) — landed before the fix (REQ-THE-006).
- Preserving every test that sets a lane env itself (REQ-THE-007).

**Out of Scope**: the items below.

### Out of Scope — whole-suite skip, blacklist, or `t.Skip` lists
- No test is skipped, deleted, or listed as "known env-sensitive". The card forbids a whole-suite
  blacklist; a skipped test measures nothing and hides the next real regression.

### Out of Scope — production-code change
- No change to any non-`*_test.go` file: the lane gates (`contractLaneGate`, the lane
  admission/refusal predicates, the stale-run gate) keep their behavior. No new production env
  axis is introduced.

### Out of Scope — template mirrors
- Test-only change under `internal/`; no `internal/template/templates/` counterpart exists for the
  touched files, so the Template-First mirror obligation does not apply.

### Out of Scope — other axes and other packages beyond the measured survey
- Env axes unrelated to the lane/kanban gates (`CLAUDE_*`, `MOAI_HOME`, git env) are owned by
  their existing TestMain scrubs and are not revisited.
- `internal/discovery` references one axis in production code. Its narrow lane-vs-scrubbed pair
  was measured at plan time (E-7: equal, 14 pass and 1 pre-existing helper skip in both arms) and
  is re-recorded once in the baseline record commit (plan.md M1 c1); a contingent fix, if a test
  there ever measures red, is an M4 pin. `internal/cli/ptycaptest` keeps its own drift guard and
  is not measured here (R4).

### Out of Scope — a shared non-test helper package
- A new importable test-support package under `internal/` (design Option C) is not part of the
  recommended scope: it is a non-`_test.go` addition to the production tree. It is compared in §D
  and left as an open question (§H).

## §C Requirements (GEARS)

Acceptance criteria are enumerated canonically in `acceptance.md` §D (Tier M). This section
carries the requirement layer only.

- **REQ-THE-001** — **While** the `internal/cli` test binary runs inside a session whose env
  carries the factory/kanban lane gate axes, the package's set of failing tests shall equal its
  set of failing tests under a fully scrubbed env.
- **REQ-THE-002** — **While** the `internal/hook` test binary runs inside a session whose env
  carries the factory/kanban lane gate axes, the package's set of failing tests shall equal its
  set of failing tests under a fully scrubbed env.
- **REQ-THE-003** — **When** the run phase begins, the run shall measure the whole-package
  failing-test set of `internal/cli` and of `internal/hook` under a lane env and under a scrubbed
  env, both before and after the change, and shall record each command, exit code, failing-test
  list, the swept test count against the package's independently listed test count, and the
  difference of the two failing lists, treating an arm that ends in a timeout panic, a
  goroutine-leak report, or a non-test failure as no measurement; the run shall name a failure
  env-unrelated only when it is identical in both arms **and** present in the c1 (pre-guard tree)
  failing set of the same arm type, and shall record the difference `final − c1` of each final-tree
  arm's failing names against that c1 set and require it empty — a name absent from c1 is a
  regression caused by the change, not an env-unrelated failure. The unit of every failing-set
  comparison is the Go test row, a subtest included and named by its full path, and the run shall
  record the lane arm's env (names set with values, names unset) and treat a lane arm that lacks a
  modelled lane axis, or whose c1 failing set lacks the five observed reds, as no measurement and
  the criterion as failed.
- **REQ-THE-004** — **When** the measurement of REQ-THE-003 shows a test that flips between the
  two arms and is not one of the five observed reds, the run shall fix that test only with its
  own RED/GREEN pair, and shall not skip, delete, or list any test as a substitute for a fix.
- **REQ-THE-005** — **When** production code in a guarded package references a lane/kanban gate
  axis that the package's test binary does not strip at start-up and that carries no reasoned,
  cited exemption, the package's test run shall fail and shall name the uncovered axis.
- **REQ-THE-006** — The recurrence guards of REQ-THE-005 and REQ-THE-009 shall land in a commit
  that is a strict ancestor of every other commit on the branch that changes a file under
  `internal/`, and the whole-package baseline record shall land in a commit that is a strict
  ancestor of that guard commit, so that the guards are red on the tree that precedes the fix.
- **REQ-THE-007** — **Where** a test composes lane env deliberately (through its own helper, or
  through a re-executed child whose parent composed the env as payload), the hermeticity change
  shall leave that test's verdict unchanged — the test being the Go test row, so a subtest's
  verdict counts apart from its parent's.
- **REQ-THE-008** — The change shall touch `*_test.go` files and this SPEC's artifacts only and
  shall introduce no new production env axis.
- **REQ-THE-009** — **When** the guarded package's applied-behaviour test starts the package
  test binary with every lane/kanban gate axis present in its environment and the binary's
  start-up scrub leaves a production-referenced axis that carries no exemption still present, the
  applied-behaviour test shall fail and shall name each such axis.

REQ count 9 (Tier M ceiling 16). The brevity of the list is deliberate: the card is a follow-up
sweep, and the verification weight sits in `acceptance.md`.

## §D Design Options Considered

The card offers two mechanisms: pin all axes with `t.Setenv` per test, **or** standardize a
gate-seeding helper. The package already uses a third idiom (a binary-wide scrub in `TestMain`).
Three options are compared; **the lane lead decides from audit evidence** — nothing here is an
operator decision.

Measured sizes (HEAD `a5a63a0bc`): cli has 10 scrub/seed helper definitions and 44
`clearFactoryTestEnv(` call sites (45 occurrences with its definition) in 11 test files; hook has 6
helper definitions; 46 cli and 28 hook test files reference the five lane axes (§A.6); the cli
`TestMain` scrub set has 12 keys. Line counts below are estimates, not measurements.

| | **A — extend the package-level scrub** | **B — per-test `t.Setenv` pinning** | **C — shared seeding/scrubbing helper** |
|---|---|---|---|
| Content | Add the missing axes (at minimum `MOAI_FACTORY_ROLE`) to cli's existing scrub set; add an equivalent `TestMain` scrub to hook | Pin every axis a test's code path reads, inside each flip-prone test | One helper per package (or one shared package) that the existing helpers delegate to |
| Sites touched (observed set only) | cli: 2 files (`factory_test.go` key slice, a few entries, plus the new guard file); hook: 2 files (`main_test.go` `TestMain` plus the new guard file) = **4 files** | 5 tests in 3 files (`todo_claim_test.go`, `todo_test.go`, `stale_run_m1_test.go`) | 16 helper bodies (10 cli + 6 hook) + 2 `TestMain` + helper definition(s) = **about 20 sites** |
| Sites touched (survey hypothesis) | unchanged — one set covers the whole binary | up to ~357 test functions (257 + ~100) across the cli files the survey counted, each re-measured | unchanged for the set, plus re-pointing of any helper whose key set differs on purpose |
| Failure mode | A binary-wide scrub **hides** ambient dependence instead of pinning it: a test that reads an axis without pinning still passes, only because the binary stripped the value. Mitigated by the guard pair (REQ-THE-005, REQ-THE-009) and by tests keeping their own pins where the axis is the subject | Per-test pins leave every **unmeasured** test exposed; a new test or a new axis re-opens the hazard. `t.Setenv` panics under `t.Parallel`, so parallel tests (e.g. `TestParseKanbanFlagUnifiedEntry`) cannot take it without dropping `t.Parallel` | Over-unification: the 16 existing helpers have non-identical key sets on purpose (some stamp a lane, some clear it, some forward to a child). A shared Option C package is a non-test addition to the production tree, in tension with REQ-THE-008 |
| Precedent in tree | `clearFactoryAmbientEnv` in `TestMain` (card t1252) and the `CLAUDE_PROJECT_DIR` scrub in hook `TestMain` (card t1165) — the same bug class, same idiom | `b12538f0a` (card t1354) on one hook test | none |

**Assessment (requested by the card; not an operator decision) — Option A, plus Option B only
for tests that the whole-package measurement still shows red after A**, because the
simplicity ladder puts reuse of an existing in-package pattern ahead of new code (ladder rung 2),
the class of defect is binary-wide rather than per-test (the survey nominates hundreds of
functions), and Option A is the only option whose cost does not grow with the nominated set.

**Precondition under which the recommendation holds**: (i) the M1 pre-fix whole-package pairs and
the M4 post-fix pairs show that adding the missing axes to the scrub set flips **no** test red in
the scrubbed arm (no test relies on the ambient value of an added axis); (ii) the re-executed
child processes in `internal/hook` that compose their own env (four `os.Args[0]` sites, plan.md
B9) do not carry a factory axis as payload — if one does, it takes the pin-marker pattern cli
already has (`factoryEnvPinnedEnv`). If (i) or (ii) fails for a test, that test is handled with
Option B and the exemption is recorded with its reason. Option C is not recommended while
REQ-THE-008 stands; it is the right answer only if the lane lead lifts the test-files-only
constraint (§H O6).

## §E Recurrence Guard — specification, applied behaviour, and continued firing

(Per `.claude/rules/moai/development/verification-completeness.md` §1.2 and §1.3.)

The guard is a pair of tests per package (cli, hook). Both read the same three inputs from the
tree at test time: the **family** (§A.6, read from `internal/config/envkeys.go`), the
**references** (family axes that production code of the package references), and the package's
**scrub set** — the axes its test binary strips at start-up — plus a small **exemption table**.

- **Coverage test (declared behaviour).** Fails when a referenced axis is in neither the scrub set
  nor the exemption table, naming the axis and the production file that references it. An
  exemption row is accepted only when its reason is non-empty **and** it cites a `*_test.go` file
  in the package directory that exists and itself references the axis (a test that needs the
  ambient value); an empty reason, an empty citation, or a citation to a file that is absent or
  does not mention the axis is a failure, so a padded row does not pass. A citation shows that a
  test reads the axis, not that it needs the ambient value — that last step is review's, and the
  evidence review is given is a closure obligation (acceptance.md §D.5): each surviving
  exemption row cites the M4 whole-package scrubbed-arm test that went red when that axis was
  stripped. The one variant these checks cannot see is therefore named, not claimed closed: an
  axis with no scrub plus an exemption row citing an existing test file that references it
  passes both this test and the applied-behaviour test (both skip exempt axes). The reference
  scan counts a reference by either form of §A.6 (identifier or quoted literal) and sees Go
  source only (§G R4). The comparison always runs to completion: every liveness assertion of
  input (3) reports with `t.Errorf` (never `t.Fatalf` or `FailNow` ahead of the comparison), so
  one red carries every message that applies.
- **Applied-behaviour test.** The declared scrub set proves a list, not that the binary applies
  it: a package could fill the list and never call it. This test re-executes the package's own
  test binary (`os.Args[0]` with a `-test.run` selector naming this test only) with an
  environment it builds explicitly — the parent's env minus every family axis (and minus the cli
  pin marker `factoryEnvPinnedEnv`, so the child is not pinned), plus every family axis set to a
  sentinel value and a witness variable naming the axes it set. In the child the test asserts that
  every referenced axis without an exemption is absent after the child's `TestMain` ran; the
  parent requires the child to exit 0 with a `--- PASS` line for that test, and the witness list
  to be non-empty, so a child that ran nothing is not a pass. The result does not depend on the
  ambient env of the parent session, and the mechanism is test-file-only (a child process of the
  test binary), so REQ-THE-008 holds; existing tests already re-execute the package binary the
  same way (`internal/hook/slot_lease_guard_test.go:165`, `factory_handoff_race_test.go:38`,
  `internal/cli/codex_launcher_exec_posix_test.go:100-123`). A source scan that `TestMain` calls
  the scrub function was considered and not chosen: it passes on a call inside a comment or a
  dead branch, which the child cannot.
- **(a) WHEN it runs to be meaningful.** On every run of the package's tests that includes it:
  locally under any selector that matches it, and in CI on the full-package `go test` per OS
  matrix (the verdict surface on `origin/develop`). Both tests compare or apply state of the tree
  they run on, so they are meaningful at every commit that can change either side — never
  scheduled at a moment where the two sides cannot yet differ. Both are independent of the ambient
  env, so unlike the five observed reds they are red on a plain shell too.
- **(b) the INPUT that turns it red.** Any one of: (1) a production file newly references a family
  axis that is in neither the scrub set nor the exemption table; (2) a padded exemption row
  (empty reason, empty citation, or a citation to a file that does not reference the axis);
  (3) a liveness input of the coverage test — the reference scan finds **zero** referenced axes,
  or the scrub set is empty, or the family read from `envkeys.go` has fewer members than the
  floor recorded at c2 (**17**, §A.6) (an empty sweep asserts nothing,
  `verification-completeness.md` §1.1; a scan that silently drops most constants would otherwise
  stay green on `referenced >= 1`). Each of the three is its own `t.Errorf`, so the hook coverage
  test at c2 — whose scrub set is deliberately empty (plan.md D5) — carries **both** the
  empty-scrub-set message and the thirteen uncovered-axis names; the cli coverage test at c2
  (non-empty scrub set) carries the six uncovered-axis names only. A deliberate edit that
  removes a family constant lowers the floor in the same change; (4) the binary's start-up scrub is not
  applied — the child still sees a referenced axis; (5) the sibling package's guard file is absent
  or no longer declares its guard tests. Inputs (1)-(2) reintroduce the card's hazard for a future
  axis; input (4) is the hazard in the form a declared-only guard would miss. The family is read
  from `envkeys.go` at test time, so a new constant joins it without editing the guard.
- **(c) who sees the red.** The author running the package's tests (exit code 1, a `--- FAIL`
  line naming the axis and, for the coverage test, the production file that references it), the
  lane's scoped verification run, and CI on `origin/develop`, which the leader reads.
- **Continued firing (§1.3).** *If this guard stopped running tomorrow, what would differ in what
  a reader sees?* Built in: (1) the liveness input (3) makes a scan that has silently stopped
  matching (moved files, renamed constants) go red instead of green; (2) CI runs the full package,
  so a narrow `-run` selector cannot exclude it on the verdict surface; (3) an **unasked
  stale-guard signal** — each package's coverage test asserts that the sibling package's guard
  file exists and declares its guard tests (a read of `../hook/lane_env_axes_test.go` from cli and
  `../cli/factory_env_axes_test.go` from hook, the precedent being
  `internal/config/envkeys_factory_role_test.go`), so deleting or renaming one guard turns the
  other package's run red without anyone asking; (4) the closure gate lists both guard tests in
  both packages with `go test -list` (acceptance.md §D.5). Residual, not eliminated: deleting both
  guard files in one change, or running only a selector that omits them locally, is not detected
  by the guards themselves; the removal-of-a-test clause of AC-THE-006 covers tests present at
  `2de0a2cb6`, not the guards. Deletion shows in the diff and in review. §H O4 records the
  alternative of placing the check in `TestMain`, which rides every selector but fails every run
  of the package.

## §F Ordering evidence

The baseline artifact `.moai/reports/t1356/baseline.md` is gitignored, so the commit graph cannot
witness measure-before-change from a report commit (`verification-claim-integrity.md` §2.3). The
ordering clause is therefore rewritten into what git can witness: three commits fixed in shape
and order — the **baseline record commit** (progress.md only), the **guard commit** (the guard
files only), and the **guard-red record commit** (progress.md only) — and the guard-red record
commit is a strict ancestor of every commit other than c2 that touches `internal/`. The whole-package
baseline pairs are recorded in this SPEC's committed `progress.md` §E.2 (a committed carrier,
unlike `.moai/reports/`) in the baseline record commit; the guard's red on its own tree is
recorded in the guard-red record commit. Nothing commits the gitignored baseline: no
`git add -f`, no re-added negation. The exact subjects, the file-set checks, and the
`git merge-base --is-ancestor` / `git rev-list --count` enumeration that make this checkable live
in one place, AC-THE-005.

## §G Constraints and Residual Risks

### Constraints
- **Heavy runs are serialized and leased.** Whole-package runs of `internal/cli` and
  `internal/hook` are minute-scale. Each takes a lease first (`moai slot acquire --resource
  heavy-test --max-duration <cap>` … `moai slot release --resource heavy-test`, per
  `.claude/rules/local/gitflow-lane-protocol.md` §8), runs as **one compound invocation**
  `unset <VARS> && go test …` for the scrubbed arm, and never as `go test ./...`. The cap is the
  holder's own declared bound: `max(20m, 1.5 x the longest whole-package runtime recorded so far
  in progress.md §E.2)` (20m before any runtime exists — the first c1 arm — and recomputed for
  each later lease), and `go test -timeout` is set strictly below it (the cap minus 2m), so the
  lease cannot lapse at the instant the test timeout fires. The cap minus 2m is at least 18m while
  a foreground Bash call ends at 600 s, so each whole-package arm runs as a background Bash call.
- **Local verification is scoped to the change.** Only the two packages, the narrow selectors of
  `acceptance.md`, and the compile checks named in plan.md; full-suite judgment is CI on
  `origin/develop`.
- **`t.TempDir()` for temp dirs; env axis names via `internal/config/envkeys.go` constants.** The
  only literal forbidden in `internal/hook` by `internal/config/envkeys_factory_role_test.go:60`
  is the exact quoted string `"MOAI_FACTORY_ROLE"` (the scan reads every `.go` file there, test
  files included); a guard that matches the unquoted prefixes `MOAI_FACTORY_` / `MOAI_KANBAN` is
  legal, and the one full name a guard needs, `MOAI_AUTONOMY_TIER`, is reached as
  `config.EnvAutonomyTier`.
- **`t.Setenv`/`t.Chdir` tests must not call `t.Parallel()`** (Go panics on the combination).
- **No OTEL env is touched** (`internal/hook/CLAUDE.md`: never `t.Setenv("OTEL_EXPORTER_*")` in
  parallel tests); the change sets and clears only the lane/kanban axes.

### Residual Risks
- **R1 — tests that rely on a lane env they set themselves.** Helpers such as `sdLaneEnv` stamp
  the lane env after `TestMain`, so a binary-wide scrub does not reach them; a parent that composes
  env for a re-executed child needs the pin marker. Verified by AC-THE-003 (same failing set in
  both arms) and AC-THE-007 (the pin-marker and ambient-clear tests stay green).
- **R2 — parallel tests and the OTEL caveat.** Option A scrubs once in `TestMain`, before any
  parallelism, so it cannot trip the `t.Setenv`-under-`t.Parallel` panic; Option B pins can.
- **R3 — Windows / other-OS key sets.** The axis names are constants shared across OSes and
  `os.Unsetenv` is portable; Windows env names are case-insensitive but the constants are upper
  case. The guard reads files through `filepath`. Verified by a `GOOS=windows` compile of the two
  packages' tests (plan.md M4). No per-OS key-set difference is claimed beyond that.
- **R4 — the unmeasured set stays a hypothesis.** The ~357 nominated functions are not individually
  measured; their status is established only by the M1 baseline pairs and the M4 final pairs. A
  flip-prone test reading an env var outside the family is not found by this sweep, and
  `internal/cli/ptycaptest` is covered only by its own drift guard. The guard pair sees **Go
  references only**: a hook test that executes a shell gate and hands it a family value as
  payload (`tierEnv(...)` in `internal/hook/sync_gate_failstate_test.go:303`) is outside both
  tests, and a family axis read only by a non-Go program is invisible to the reference scan.
- **R5 — Option A hides rather than pins.** Disclosed in §D; the guard pair converts "hidden" into
  "the set of hidden axes is explicit, applied, and reviewed".
- **R6 — develop tip may carry env-unrelated reds.** A failure identical in both arms **and**
  present in the c1 baseline of the same arm type is named and classified env-unrelated
  (REQ-THE-003); it is neither fixed nor hidden here. A failure identical in both final arms but
  absent from c1 is a change-induced regression, never env-unrelated (AC-THE-003 clause (e)).
- **R7 — machine load flips verdicts.** The lease serializes heavy runs; a verdict that changes
  between a repeated identical arm is reported as load noise, not attributed to env. A name that
  appears in a final arm and not in c1 is repeated once as a whole-package arm (acceptance.md §D.3),
  and only a name failing in both runs counts; the comparison stays by full test path.
- **R8 — the lane arm models the measuring session, not every possible lane.** The lane arm sets
  the nine family axes the measuring session exported and unsets the other eight; a real lane with
  a different subset is not reproduced by it. The coverage test, the applied-behaviour test, and
  the scrubbed arm bound that gap: every referenced axis is declared stripped, shown stripped, and
  no test goes red when stripped. The arm is a lane only when its recorded env carries the modelled
  axes and the c1 lane arm reproduces the five reds (AC-THE-003 clause (f)); a session without lane
  axes yields an invalid arm and a failed criterion, never a vacuous pass.
- **R9 — in-process production stamping.** `enterFactoryLaneMode` (`internal/cli/factory.go:797-801`)
  sets the lane worker, role, lane-count, clear-policy and dispatch markers in the process that
  calls it and returns a restore function (the role stamp site is pinned by source in
  `factory_m4_test.go:462-469`); the start-up scrub does not cover a marker stamped later in the
  same binary, and no measurement here proves that none escapes a test.

## §H Open Questions (non-blocking)

- **O1** — Family definition for the guard: derive from `envkeys.go` constant **values** by prefix
  (`MOAI_FACTORY_`, `MOAI_KANBAN`, `MOAI_AUTONOMY_TIER`; recommended in §E because a new constant
  joins automatically) or from an explicit list in the guard (simpler, but a new axis outside the
  list is invisible)? A future axis with a different prefix escapes the value-prefix rule.
- **O2** — Which of the five cli axes other than the observed-red cause — `MOAI_AUTONOMY_TIER`,
  `MOAI_FACTORY_CLEAR_POLICY`, `MOAI_FACTORY_AUTO_DISPATCH`, `MOAI_FACTORY_MANAGED`,
  `MOAI_FACTORY_SLOW_LAUNCH_MS` — belong in cli's scrub set versus the exemption table?
  Production references them. Static read at plan time: every test site that touches
  `MOAI_FACTORY_SLOW_LAUNCH_MS` sets it itself (five `t.Setenv` sites in
  `factory_launch_timing_test.go:20,71` and `codex_debug_composition_test.go:85,112,122`) and none
  reads it from ambient; the other four are set or cleared by their tests the same way. Whether
  any test depends on an ambient value is decided by the M4 whole-package scrubbed arm: a test that
  goes red because the axis was stripped is the evidence for an exemption row (reason plus cited
  test file); otherwise the axis stays in the scrub set. The final exemption tables, and for each
  surviving row the name of that red test, are listed in progress.md §E.2 at closure
  (acceptance.md §D.5); an empty table needs no row.
- **O3** — Should the guard be generalized into one cross-package registry check covering
  `internal/discovery` (3 production files reference `MOAI_KANBAN_ID`; its narrow pair measured
  equal at plan time, E-7) and any future package, or stay one-per-package? Not decided.
- **O4** — Guard placement: an ordinary test function (recommended; CI covers selector filtering,
  and the sibling-file check of §E adds an unasked stale-guard signal) or `TestMain` (rides every
  selector, as the cwd-residue guard of SPEC-CLI-TEST-CWD-ISOLATION-001 does, but fails unrelated
  narrow runs)?
- **O5** — Should the six hook and ten cli per-test helpers later delegate to the package scrub
  set (a follow-up consolidation)? Left for a separate card; this SPEC adds no delegation.
- **O6** — Is the card's "standardize a gate-seeding helper" option wanted enough to lift
  REQ-THE-008 for a shared test-support package (Option C)?

## §I Cross-References

- `internal/cli/factory_card.go:57-71`, `internal/hook/contract_sign_guard.go:144-152`,
  `internal/hook/stale_run_gate.go:259-260`, `internal/hook/session_stale_run.go:27-35,172-179`,
  `internal/cli/factory_launch_timing.go:172` — the production gates the evidence names.
- `internal/cli/factory_test.go:27-77` (scrub set, pin marker, `clearFactoryAmbientEnv`,
  `clearFactoryTestEnv`), `internal/cli/main_test.go:349-361` (`TestMain` call),
  `internal/hook/main_test.go:67-96` (`TestMain`).
- `internal/config/envkeys.go` (`EnvFactoryRole` etc.), `internal/config/envkeys_factory_role_test.go`,
  `internal/cli/ptycaptest/selfcheck_test.go:265-290` — constant SSOT and the two drift-guard precedents.
- SPEC-CLI-TEST-CWD-ISOLATION-001 — sibling test-isolation SPEC (cwd residue), structure precedent.
- `.claude/rules/moai/development/verification-completeness.md` §1.2, §1.3, §2, §2.1, §4;
  `.claude/rules/moai/core/verification-claim-integrity.md` §2.3;
  `.claude/rules/local/gitflow-lane-protocol.md` §8.
- `.moai/reports/t1356/baseline.md` — local-only on-disk measurement (gitignored; not committed);
  `.moai/reports/t1356/plan-audit.md` and `.moai/reports/t1356/plan-audit-iter2.md` — the
  iteration-1 and iteration-2 plan-audit reports, and `.moai/reports/t1356/plan-audit-iter3.md`
  (iteration 3) (local, gitignored, cited by path only).

## §J HISTORY

| Date | Author | Change |
|------|--------|--------|
| 2026-10-03 | manager-spec | v0.1.0 plan-phase authoring (card t1356, Tier M, Class C). Evidence re-measured on tree `2de0a2cb6`: cli RED under `MOAI_FACTORY_ROLE=lane` (3 tests), hook RED under the lane env (2 tests), hook one-axis arms isolating the `MOAI_KANBAN_ID` ∧ `MOAI_FACTORY_WORKERS` conjunction, positive controls. Design Options A/B/C compared; recommendation stated with its precondition. Recurrence guard and ordering clause specified. |
| 2026-10-03 | manager-spec | v0.2.0 plan-audit iteration 1 revision (FAIL 0.79 vs 0.80; findings F1-F13). Guard spec gains an applied-behaviour test and a tightened exemption rule (REQ-THE-005 reworded, REQ-THE-009 added, AC-THE-004 and AC-THE-008 added as guard criteria whose cell is completed at the c2r record); ordering check rewritten to enumerate every `internal/` commit with prescribed c1/c2/c2r shapes (REQ-THE-006 widened to match); the 17-axis family and per-package referenced/covered/undecided sets re-measured, the sixth cli axis `MOAI_FACTORY_SLOW_LAUNCH_MS` carried through §A, §H O2, decision-index Q4 and plan M2; reach figures replaced by a reproducible command; AC-THE-003 gains an independent swept-count floor; evidence ledger re-recorded with `-v` and full stdout (E-1, E-2), cli scrubbed arm added (E-1b), discovery narrow pair added (E-7); stale-guard signal, REQ-THE-008 judging point, and minor count corrections. |
| 2026-10-03 | manager-spec | v0.3.0 plan-audit iteration 2 revision (FAIL 0.86 vs 0.80, driven by one must-fix mutant hole; findings D1-D11; this feeds the final permitted audit). REQ-THE-003 and AC-THE-003 gain the c1-containment clause (e): a failure is env-unrelated only when identical in both arms and present in the c1 failing set of the same arm type, with the `final − c1` difference recorded and required empty (R6 reworded to match); AC-THE-004's mutant-probe text no longer overstates closure (the unscrubbed-axis-plus-padded-citation variant is named and left to review, with a closure DoD item listing the final exemption tables and each surviving row's red test); the hook coverage test's c2 red carries both the empty-scrub-set liveness message and the thirteen axis names (liveness uses `t.Errorf`); the class label of AC-THE-004 and AC-THE-008 is relabelled (the v0.2.0 \"on adoption\" label is retired) and one phrase is used across AC-THE-003, 004 and 008 for a cell completed by a later record; an AC-THE-005 step 7 content witness on the c2r commit; AC-THE-004 and AC-THE-008 bound to the M4 exit; lease cap and `-timeout` relation stated; the guard's reference rule pinned (identifier or quoted literal) with a family-size liveness floor of 17; REQ-THE-009's trigger reworded; the E-5 control gets an explicit upper bound and plan-time `go test -list` counts are recorded (E-8). |
| 2026-10-03 | manager-spec | v0.4.0 delta revision for the leader-approved fourth plan-audit (iteration 3: FAIL 0.87, two must-fix holes in AC-THE-003; findings MF-1, MF-2, SF-1..SF-4). AC-THE-003 gains clause (f), a lane-arm positive control (env recorded and identical at c1 and final, every modelled axis present, five reds in the c1 lane arm, otherwise INVALID and failed) and a pre-flight env-read step; the failing-name comparison is by full test path, subtests included (REQ-THE-003 and REQ-THE-007 state the unit); exemption axes are left at their lane value in the scrubbed arm; the c2r cell carries exact command, exit code field and the c2 SHA with the pre-run HEAD read, and the sync re-execution records its own stdout; a one-repeat rule for a name absent from c1; AC-THE-006 gains assertion-removal and bare-return greps; `LC_ALL=C` on `sort` and `comm`; the §F wording slip is fixed. No REQ or AC added. |
