---
id: SPEC-TEST-ENV-HERMETIC-001
title: "Test env hermeticity sweep — tests that read the factory/kanban lane gate axes must not change verdict with the ambient env of the session that runs them"
version: "0.1.0"
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

All figures below were measured on tree `2de0a2cb6` (branch `WT-test-env-hermetic-sweep`, no local
change at measurement time) in a lane session whose ambient env carries
`MOAI_FACTORY_ROLE=lane`, `MOAI_FACTORY_WORKER=lane-6`, `MOAI_FACTORY_WORKERS=0`,
`MOAI_KANBAN_BACKEND=claude`, `MOAI_KANBAN_ID=tm9i7y`, `MOAI_AUTONOMY_TIER=fully-autonomous`,
`MOAI_FACTORY_AUTO_DISPATCH=auto`. The deciding command and verbatim output of each item is
carried in `acceptance.md` §D.0 (evidence ledger, ids E-1..E-5). The earlier local-only baseline
`.moai/reports/t1356/baseline.md` (gitignored by operator directive 2026-09-14 — `.moai/reports/*`;
it is never force-added) is cited as an on-disk measurement, not as a committed artifact; the
ledger re-measures its two decisive facts so the committed record does not depend on it.

1. **internal/cli — three observed reds, single responsible axis.** `TestTodoClaim_LaneGovernance`,
   `TestTodoClaimMCP_Mirror`, and `TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit` fail
   with `refused — lane boundary` when `MOAI_FACTORY_ROLE=lane` reaches the test binary (E-1).
   Mechanism: the lane-refusal predicate in `internal/cli/factory_card.go:57-71` reads the role
   marker; the package `TestMain` clears 12 factory/kanban keys (`factoryAmbientEnvKeys`,
   `internal/cli/factory_test.go:27-40`) and that set does **not** contain `MOAI_FACTORY_ROLE`
   (nor `MOAI_AUTONOMY_TIER`, `MOAI_FACTORY_CLEAR_POLICY`, `MOAI_FACTORY_AUTO_DISPATCH`,
   `MOAI_FACTORY_MANAGED`, all of which production code in the package references).
   `TestTodoClaim_LaneGovernance` pins only the lane-label axis (`t.Setenv(EnvMoaiFactoryWorker, "")`),
   which is exactly the partial pin the card names.
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
   `qasLaneEnv`, `nmLaneEnv`, `clearKanbanLauncherEnv`). A precedent for a drift guard exists:
   `internal/cli/ptycaptest/selfcheck_test.go` keeps its `kanbanVars` in step with
   `internal/config/envkeys.go`, and `internal/config/envkeys_factory_role_test.go` already
   enumerates the hook guard's `os.Getenv` call sites against a closed set.
6. **Reach.** Production sources referencing the lane/kanban gate axes (measured with a
   reference scan over the ten family constants): 22 files in `internal/cli`, 8 in
   `internal/hook`, 3 in `internal/discovery`, 1 in `internal/cli/ptycaptest`. Test files
   referencing the five gate-axis constants: 47 in `internal/cli`, 28 in `internal/hook`.

## §B Scope

**In Scope**:
- Making the observed-red tests hermetic with respect to the lane gate axes, in `internal/cli` and
  `internal/hook` (REQ-THE-001, REQ-THE-002).
- Measuring — not assuming — which further tests flip under the lane env, by whole-package
  failing-set comparison, and fixing only those that measure red (REQ-THE-003, REQ-THE-004).
- A recurrence guard in each of the two packages (REQ-THE-005) landed before the fix (REQ-THE-006).
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
- `internal/discovery` and `internal/cli/ptycaptest` also reference the axes in production code;
  they are surveyed by one narrow lane-vs-scrubbed measurement (plan.md M4) and fixed only if a
  test there measures red. `ptycaptest` already has its own drift guard.

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
  list, and the difference of the two lists, naming every failure that is identical in both arms
  as env-unrelated.
- **REQ-THE-004** — **When** the measurement of REQ-THE-003 shows a test that flips between the
  two arms and is not one of the five observed reds, the run shall fix that test only with its
  own RED/GREEN pair, and shall not skip, delete, or list any test as a substitute for a fix.
- **REQ-THE-005** — **When** production code in a guarded package references a lane/kanban gate
  axis that the package's test-binary scrub set does not cover and that carries no reasoned
  exemption, the package's test run shall fail and shall name the uncovered axis.
- **REQ-THE-006** — The recurrence guard of REQ-THE-005 shall land in a commit that is a strict
  ancestor of every commit that changes either package's scrub set, so that the guard is red on
  the tree that precedes the fix.
- **REQ-THE-007** — **Where** a test composes lane env deliberately (through its own helper, or
  through a re-executed child whose parent composed the env as payload), the hermeticity change
  shall leave that test's verdict unchanged.
- **REQ-THE-008** — The change shall touch `*_test.go` files and this SPEC's artifacts only and
  shall introduce no new production env axis.

REQ count 8 (Tier M ceiling 16). The brevity of the list is deliberate: the card is a follow-up
sweep, and the verification weight sits in `acceptance.md`.

## §D Design Options Considered

The card offers two mechanisms: pin all axes with `t.Setenv` per test, **or** standardize a
gate-seeding helper. The package already uses a third idiom (a binary-wide scrub in `TestMain`).
Three options are compared; **the lane lead decides from audit evidence** — nothing here is an
operator decision.

Measured sizes (this run, tree `2de0a2cb6`): cli has 10 scrub/seed helper definitions and 45
`clearFactoryTestEnv(` call sites in 11 test files; hook has 6 helper definitions; 47 cli and 28
hook test files reference the five gate-axis constants; the cli `TestMain` scrub set has 12 keys.
Line counts below are estimates, not measurements.

| | **A — extend the package-level scrub** | **B — per-test `t.Setenv` pinning** | **C — shared seeding/scrubbing helper** |
|---|---|---|---|
| Content | Add the missing axes (at minimum `MOAI_FACTORY_ROLE`) to cli's existing scrub set; add an equivalent `TestMain` scrub to hook | Pin every axis a test's code path reads, inside each flip-prone test | One helper per package (or one shared package) that the existing helpers delegate to |
| Sites touched (observed set only) | cli: 1 file (the key slice, a few entries); hook: 2 files (`TestMain` + a key set), plus the 2 guard files = **about 5 files** | 5 tests in 3 files (`todo_claim_test.go`, `todo_test.go`, `stale_run_m1_test.go`) | 16 helper bodies (10 cli + 6 hook) + 2 `TestMain` + helper definition(s) = **about 20 sites** |
| Sites touched (survey hypothesis) | unchanged — one set covers the whole binary | up to ~357 test functions (257 + ~100) across the cli files the survey counted, each re-measured | unchanged for the set, plus re-pointing of any helper whose key set differs on purpose |
| Failure mode | A binary-wide scrub **hides** ambient dependence instead of pinning it: a test that reads an axis without pinning still passes, only because the binary stripped the value. Mitigated by the guard (REQ-THE-005) and by tests keeping their own pins where the axis is the subject | Per-test pins leave every **unmeasured** test exposed; a new test or a new axis re-opens the hazard. `t.Setenv` panics under `t.Parallel`, so parallel tests (e.g. `TestParseKanbanFlagUnifiedEntry`) cannot take it without dropping `t.Parallel` | Over-unification: the 16 existing helpers have non-identical key sets on purpose (some stamp a lane, some clear it, some forward to a child). A shared Option C package is a non-test addition to the production tree, in tension with REQ-THE-008 |
| Precedent in tree | `clearFactoryAmbientEnv` in `TestMain` (card t1252) and the `CLAUDE_PROJECT_DIR` scrub in hook `TestMain` (card t1165) — the same bug class, same idiom | `b12538f0a` (card t1354) on one hook test | none |

**Assessment (requested by the card; not an operator decision) — Option A, plus Option B only
for tests that the whole-package measurement still shows red after A**, because the
simplicity ladder puts reuse of an existing in-package pattern ahead of new code (ladder rung 2),
the class of defect is binary-wide rather than per-test (the survey nominates hundreds of
functions), and Option A is the only option whose cost does not grow with the nominated set.

**Precondition under which the recommendation holds**: (i) the M1 pre-fix whole-package pairs and
the M4 post-fix pairs show that adding the missing axes to the scrub set flips **no** test red in
the scrubbed arm (no test relies on the ambient value of an added axis); (ii) the re-executed
child processes in `internal/hook` that compose their own env (`os.Args[0]` sites) do not carry a
factory axis as payload — if one does, it takes the pin-marker pattern cli already has
(`factoryEnvPinnedEnv`). If (i) or (ii) fails for a test, that test is handled with Option B and
the exemption is recorded with its reason. Option C is not recommended while REQ-THE-008 stands;
it is the right answer only if the lane lead lifts the test-files-only constraint (§H O6).

## §E Recurrence Guard — three-part specification and continued firing

(Per `.claude/rules/moai/development/verification-completeness.md` §1.2 and §1.3.)

The guard is one test per package (cli, hook). It compares **the lane/kanban gate axes that
production code of the package references** against **the package's test-binary scrub set plus a
small table of reasoned exemptions**, and fails when an axis is in neither.

- **(a) WHEN it runs to be meaningful.** On every run of the package's tests that includes it:
  locally under any selector that matches it, and in CI on the full-package `go test` per OS
  matrix (the verdict surface on `origin/develop`). The guard is a static comparison of two sets
  read from the same tree, so it is meaningful at every commit that can change either side — it is
  never scheduled at a moment where the two sides cannot yet differ. It is independent of the
  ambient env, so unlike the five observed reds it is red on a plain shell too.
- **(b) the INPUT that turns it red.** Any one of: (1) a production file in the package newly
  references a lane/kanban gate axis constant — a constant in `internal/config/envkeys.go` whose
  value starts with `MOAI_FACTORY_` or `MOAI_KANBAN`, or is `MOAI_AUTONOMY_TIER` — that is neither
  in the scrub set nor in the exemption table; (2) an exemption row with an empty reason; (3) the
  reference scan finds **zero** referenced axes, or the scrub set is empty (an empty sweep asserts
  nothing, `verification-completeness.md` §1.1). Input (1) is the case that reintroduces the
  card's hazard: a future axis added to production code with no matching scrub entry. The family
  is read from `envkeys.go` at test time, so a new constant joins it without editing the guard.
- **(c) who sees the red.** The author running the package's tests (exit code 1, a `--- FAIL` line
  naming the uncovered axis constant and the production file that references it), the lane's
  scoped verification run, and CI on `origin/develop`, which the leader reads.
- **Continued firing (§1.3).** *If this guard stopped running tomorrow, what would differ in what
  a reader sees?* Two answers are built in and one residual is named. Built in: (1) the positive
  control of input (3) makes a scan that has silently stopped matching (moved files, renamed
  constants) go red instead of green; (2) CI runs the full package, so a narrow `-run` selector
  cannot exclude it on the verdict surface. Residual, not eliminated: deleting or renaming the
  guard test, or running only a selector that omits it locally, is not detected by the guard
  itself — deletion shows in the diff and in review, and the AC-THE-005 ancestry check ties the
  guard to the fix commits only at the time the SPEC closes. §H O4 records the alternative of
  placing the check in `TestMain`, which rides every selector but fails every run of the package.

## §F Ordering evidence

The baseline artifact `.moai/reports/t1356/baseline.md` is gitignored, so the commit graph cannot
witness measure-before-change from a report commit (`verification-claim-integrity.md` §2.3). The
ordering clause is therefore rewritten into what git can witness: the **guard test commit and the
whole-package baseline record commit are strict ancestors of every commit that changes a scrub
set or a test's env handling**, and the guard is red on the tree of its own commit. The baseline
whole-package pairs are recorded in this SPEC's committed `progress.md` §E.2 (a committed carrier,
unlike `.moai/reports/`) in a commit of their own. AC-THE-005 makes this checkable from `git log`
and `git merge-base --is-ancestor`.

## §G Constraints and Residual Risks

### Constraints
- **Heavy runs are serialized and leased.** Whole-package runs of `internal/cli` and
  `internal/hook` are minute-scale. Each takes a lease first (`moai slot acquire --resource
  heavy-test --max-duration <cap>` … `moai slot release --resource heavy-test`, per
  `.claude/rules/local/gitflow-lane-protocol.md` §8), runs as **one compound invocation**
  `unset <VARS> && go test …` for the scrubbed arm, and never as `go test ./...`. The cap is the
  holder's own declared bound and is set above the whole-package runtime recorded at M1.
- **Local verification is scoped to the change.** Only the two packages, the narrow selectors of
  `acceptance.md`, and the compile checks named in plan.md; full-suite judgment is CI on
  `origin/develop`.
- **`t.TempDir()` for temp dirs; env axis names via `internal/config/envkeys.go` constants.**
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
  flip-prone test reading an env var outside the family is not found by this sweep.
- **R5 — Option A hides rather than pins.** Disclosed in §D; the guard converts "hidden" into
  "the set of hidden axes is explicit and reviewed".
- **R6 — develop tip may carry env-unrelated reds.** A failure identical in both arms is named and
  classified env-unrelated (REQ-THE-003); it is neither fixed nor hidden here.
- **R7 — machine load flips verdicts.** The lease serializes heavy runs; a verdict that changes
  between a repeated identical arm is reported as load noise, not attributed to env.

## §H Open Questions (non-blocking)

- **O1** — Family definition for the guard: derive from `envkeys.go` constant **values** by prefix
  (`MOAI_FACTORY_`, `MOAI_KANBAN`, `MOAI_AUTONOMY_TIER`; recommended in §E because a new constant
  joins automatically) or from an explicit list in the guard (simpler, but a new axis outside the
  list is invisible)? A future axis with a different prefix escapes the value-prefix rule.
- **O2** — Which of `MOAI_AUTONOMY_TIER`, `MOAI_FACTORY_CLEAR_POLICY`, `MOAI_FACTORY_AUTO_DISPATCH`,
  `MOAI_FACTORY_MANAGED` belong in cli's scrub set versus the exemption table? Production
  references them; whether any test depends on their ambient value is unmeasured. M2 decides by
  measurement; an axis that a test legitimately reads from ambient takes an exemption with a reason.
- **O3** — Should the guard be generalized into one cross-package registry check covering
  `internal/discovery` (3 production files reference `MOAI_KANBAN_ID`) and any future package, or
  stay one-per-package? Not decided; discovery gets a narrow measurement at M4.
- **O4** — Guard placement: an ordinary test function (recommended; CI covers selector filtering)
  or `TestMain` (rides every selector, as the cwd-residue guard of SPEC-CLI-TEST-CWD-ISOLATION-001
  does, but fails unrelated narrow runs)?
- **O5** — Should the six hook and ten cli per-test helpers later delegate to the package scrub
  set (a follow-up consolidation)? Left for a separate card; this SPEC adds no delegation.
- **O6** — Is the card's "standardize a gate-seeding helper" option wanted enough to lift
  REQ-THE-008 for a shared test-support package (Option C)?

## §I Cross-References

- `internal/cli/factory_card.go:57-71`, `internal/hook/contract_sign_guard.go:144-152`,
  `internal/hook/stale_run_gate.go:259-260`, `internal/hook/session_stale_run.go:27-35,172-179` —
  the production gates the evidence names.
- `internal/cli/factory_test.go:27-77` (scrub set, `clearFactoryAmbientEnv`, `clearFactoryTestEnv`),
  `internal/cli/main_test.go:349-361` (`TestMain` call), `internal/hook/main_test.go:67-96`
  (`TestMain`).
- `internal/config/envkeys.go` (`EnvFactoryRole` etc.), `internal/config/envkeys_factory_role_test.go`,
  `internal/cli/ptycaptest/selfcheck_test.go:265-290` — constant SSOT and the two drift-guard precedents.
- SPEC-CLI-TEST-CWD-ISOLATION-001 — sibling test-isolation SPEC (cwd residue), structure precedent.
- `.claude/rules/moai/development/verification-completeness.md` §1.2, §1.3, §2, §2.1, §4;
  `.claude/rules/moai/core/verification-claim-integrity.md` §2.3;
  `.claude/rules/local/gitflow-lane-protocol.md` §8.
- `.moai/reports/t1356/baseline.md` — local-only on-disk measurement (gitignored; not committed).

## §J HISTORY

| Date | Author | Change |
|------|--------|--------|
| 2026-10-03 | manager-spec | v0.1.0 plan-phase authoring (card t1356, Tier M, Class C). Evidence re-measured on tree `2de0a2cb6`: cli RED under `MOAI_FACTORY_ROLE=lane` (3 tests), hook RED under the lane env (2 tests), hook one-axis arms isolating the `MOAI_KANBAN_ID` ∧ `MOAI_FACTORY_WORKERS` conjunction, positive controls. Design Options A/B/C compared; recommendation stated with its precondition. Recurrence guard and ordering clause specified. |
