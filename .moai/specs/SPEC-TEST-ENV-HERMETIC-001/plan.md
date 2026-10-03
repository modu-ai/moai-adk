# Plan — SPEC-TEST-ENV-HERMETIC-001

Milestones are ordered by decision-reversibility: the guards and the scrub-set decisions (the
choices most likely to change after the measurements) come first; mechanical steps come last.
Priorities are High / Medium / Low; no time estimates.

## §A Context

- **Worktree**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1356`, branch
  `WT-test-env-hermetic-sweep`, plan-phase measurement tree `2de0a2cb6` (re-measured at HEAD
  `a5a63a0bc`, Go sources identical; re-read `git rev-parse --short HEAD` and `git branch
  --show-current` before any commit).
- **Card**: t1356, Class C (plan → plan-audit → run → sync). Git-flow lane protocol: no card PR;
  integration into local `develop` via the integration window; CI on `origin/develop` is the
  full-suite verdict surface. The card id `t1356` appears in every commit message.
- **Artifacts**: spec.md + plan.md + acceptance.md (Tier M) + progress.md + decision-index.md
  (the decision gate is on in `.moai/config/sections/interview.yaml`).
- **Development mode**: TDD — the RED guard commit precedes the fix commits (REQ-THE-006).
- **Evidence status**: the five observed reds, the cli scrubbed arm, the hook one-axis arms and
  the discovery narrow pair, and the plan-time listed counts are carried in acceptance.md §D.0
  (ledger E-1..E-8).
  `.moai/reports/t1356/baseline.md` and `.moai/reports/t1356/plan-audit.md` are local-only and
  gitignored; never `git add -f` them and never re-add a gitignore negation.
- **PRESERVE**: every production file; `internal/cli/ptycaptest/` (own guard); the six hook and ten
  cli helpers' bodies (no delegation in this SPEC); `factoryEnvPinnedEnv` semantics; all
  currently-green assertions.
- **Existing infrastructure to EXTEND**: cli `factoryAmbientEnvKeys` + `clearFactoryAmbientEnv`
  (called from `main_test.go:356`); hook `TestMain` (`main_test.go:67`, same file that already
  scrubs `CLAUDE_PROJECT_DIR`, card t1165).

## §B Known Issues

- **B1 — `t.Setenv` panics under `t.Parallel`.** Any M4 pin must drop `t.Parallel()` from that
  test or the test is handled differently; audit the touched test for parallel markers first.
- **B2 — a scrubbed arm needs `unset … && go test …` as ONE compound invocation.** A separate
  `unset` does not reach the next command (each Bash call is a fresh process).
- **B3 — `env -u … go test` is refused by the worktree guard** (measured at plan time: "cannot be
  shown not to be git"), and so is `go test -list '.*' … | tail` (iteration 2: a pipe after a `go`
  command with a regex argument); `unset <VARS> && go test …`, `env NAME=value go test …`, a plain
  `go test -list '.*' <pkg>`, and `unset <VARS> && go test … -json > <file>` are accepted. Use
  those forms. A refusal is recorded as a Gap, never silently substituted
  (`verification-claim-integrity.md` §3.1).
- **B4 — `env NAME=value go test` does not remove other ambient axes.** A "lane arm" must set the
  whole modelled lane env explicitly **and** unset every other family axis in the same compound
  invocation, and a "scrubbed arm" must unset the whole family; a partial arm measures neither.
  Plan-phase partial arms were discarded for this reason (progress.md §E.1).
- **B5 — go test caching.** Every evidence-bearing run uses `-count=1`.
- **B6 — empty sweep.** A `-run` selector that matches nothing prints `[no tests to run]` and exits
  0; every narrow-run AC counts `--- PASS` lines, not exit codes alone (`verification-completeness.md` §1.1).
- **B7 — cross-platform.** Guard path logic uses `filepath`; `GOOS=windows GOARCH=amd64 go vet` on
  the two packages type-checks the test files for the Windows build.
- **B8 — scope discipline on a shared machine.** Stage by explicit pathspec only; never `git add
  -A`; re-read `git status --short` immediately before staging.
- **B9 — hook child re-exec sites.** `internal/hook` re-executes its own test binary in four test
  files, one `exec.Command(os.Args[0], …)` site each: `slot_lease_guard_test.go:165`,
  `session_start_drift_fill_burst_test.go:204`, `factory_handoff_race_test.go:38`,
  `factory_handoff_bind_test.go:403` (`session_start_drift_fill_test.go` mentions
  `os.Executable()` in comments and an assertion message and spawns no child). A `TestMain` scrub
  also reaches those children.
- **B10 — the range base moves when develop is absorbed.** Ordering and diff checks name
  `<BASE>` = `2de0a2cb6`. If the branch absorbs local `develop`, a literal pin would pull other
  cards' commits into `<BASE>..HEAD`; re-derive `<BASE>` once with `git merge-base develop HEAD`,
  record it in progress.md §E.2, and require `git rev-list --count <BASE>..HEAD` ≥ 1
  (`.claude/rules/local/gitflow-lane-protocol.md` §8, pre-merge evaluation only). At plan time
  `git merge-base develop HEAD` = `2de0a2cb6`.

## §C Pre-Flight (run-phase session, before any commit; record outputs in progress.md §E.2)

```bash
git rev-parse --short HEAD
git branch --show-current
# expect: 2de0a2cb6 or a descendant containing only this SPEC's plan-phase commits / WT-test-env-hermetic-sweep
```

Then, one command per call:

1. Re-establish the five observed reds with the RED-now commands of acceptance.md §D.0 (E-1, E-2)
   and the green arm E-1b — same commands, same tree lineage; record verbatim output and the tree SHA.
2. List the family from `internal/config/envkeys.go` (constants whose value starts `MOAI_FACTORY_`
   or `MOAI_KANBAN`, plus `MOAI_AUTONOMY_TIER`; 17 at `a5a63a0bc`) — this list builds both arms'
   env; do not copy it from this plan. Read the session's own family env, once, before any arm,
   with `env | grep -E '^(MOAI_FACTORY_|MOAI_KANBAN|MOAI_AUTONOMY_TIER)'` (a form the worktree guard
   accepts; it printed the nine axes at plan time, exit `0`, acceptance.md E-9): the lane arm
   reproduces it verbatim (nine axes at plan time: eight non-empty plus `MOAI_FACTORY_CLEAR_POLICY`
   set empty) and unsets the other eight. Record the output as the arm env lines of progress.md
   §E.2. The read must show at least `MOAI_FACTORY_ROLE=lane` and non-empty `MOAI_FACTORY_WORKER`,
   `MOAI_FACTORY_WORKERS` and `MOAI_KANBAN_ID`; a session without them cannot measure the lane arm,
   so the run returns a blocker report and AC-THE-003 stays failed (acceptance.md §D.3 clause (f)).
3. Hook child re-exec census: for each `os.Args[0]` site in `internal/hook/*_test.go` (B9), record
   whether its `cmd.Env` carries a family axis as payload. Expected: none do; if one does, M3 uses
   the pin-marker pattern.
4. `GOOS=windows GOARCH=amd64 go vet ./internal/cli ./internal/hook` and `golangci-lint run
   --timeout=2m` baselines (distinguish NEW from pre-existing findings later).
5. Acquire the lease before any whole-package run: `moai slot status --resource heavy-test`, then
   `moai slot acquire --resource heavy-test --max-duration <cap>` with `<cap>` = `max(20m, 1.5 x
   the longest whole-package runtime recorded so far in progress.md §E.2)` (20m for the first c1
   arm, recomputed for each later lease), and run `go test` with `-timeout` = the cap minus 2m, so
   the timeout is strictly below the lease; each arm runs as a background Bash call, because the
   cap minus 2m is at least 18m and a foreground call ends at 600 s; release with `moai slot release
   --resource heavy-test` immediately after each arm. Exit 3 = held by another session: report and wait, never `--force`.
6. Independent test counts: `go test -list '.*' ./internal/cli > <file>` and `go test -list '.*'
   ./internal/hook > <file>` as plain commands (B3), then `grep -cE '^(Test|Example|Fuzz)' <file>`;
   L = number of output lines beginning `Test`, `Example` or `Fuzz` (benchmarks excluded).
   Recorded as L0 for the pre-guard tree. Plan-time figures at `669cf18c9`: cli 4884, hook 1322
   (acceptance.md E-8); `go test -list` runs no test and takes seconds on a warm build cache.

## §D Constraints (Hard)

1. **Test-first ordering (REQ-THE-006).** The commit chain is fixed, with exact subjects and file
   sets (the single checkable statement is AC-THE-005):
   - **c1** `test(SPEC-TEST-ENV-HERMETIC-001): baseline record (card t1356)` — touches
     `.moai/specs/SPEC-TEST-ENV-HERMETIC-001/progress.md` only;
   - **c2** the guard commit — adds the two guard files and nothing else
     (`internal/cli/factory_env_axes_test.go`, `internal/hook/lane_env_axes_test.go`);
   - **c2r** `test(SPEC-TEST-ENV-HERMETIC-001): guard red record (card t1356)` — touches progress.md
     only;
   - then **c3** (cli fix), **c4** (hook fix), **c5+** (contingent pins), evidence, sync.
   No fix commit before c2r exists, and no fix commit is also c2.
2. **Test files only (REQ-THE-008).** `git diff --name-only <BASE>..HEAD` contains only `*_test.go`
   and `.moai/specs/SPEC-TEST-ENV-HERMETIC-001/*`, judged at the M4 tip and unchanged at the sync
   tip (M5 adds SPEC-directory paths only).
3. **No skip / delete / blacklist (REQ-THE-004).** The applied-behaviour test is a helper-process
   test: in the parent it re-executes the binary, in the child it asserts — it ends with a
   `return`, never a `t.Skip`.
4. **Heavy runs leased and serial** (§C step 5); never `go test ./...`.
5. **English code comments; env names via `internal/config/envkeys.go` constants.** The only
   literal forbidden in `internal/hook` by `internal/config/envkeys_factory_role_test.go:60` is
   the exact quoted string `"MOAI_FACTORY_ROLE"` (it scans every `.go` file there, test files
   included); a guard that matches the unquoted prefixes `MOAI_FACTORY_` / `MOAI_KANBAN` is legal,
   and the one full name needed, `MOAI_AUTONOMY_TIER`, is `config.EnvAutonomyTier`.
6. **Conventional commits** with card id `t1356`; `🗿 MoAI` trailer per repo convention.
7. **MX tags**: new guard helpers are test code; no `@MX:ANCHOR` is expected; none added unless the
   scan reveals fan_in ≥ 3.

## §E Self-Verification (design decisions)

- **D1 — Option A + guard pair (assessment, not an operator decision).** See spec.md §D; the lane
  lead decides from audit evidence. The plan below is written for A; M4 is the Option-B contingency.
- **D2 — guards anchor on the judgment target.** The coverage test compares sets read from the
  same tree (production references vs scrub set ∪ cited exemptions); it does not mock the env and
  does not depend on the ambient env.
- **D3 — coverage logic is a pure comparison over inputs**, so it is exercised on synthetic inputs
  (uncovered axis, exempted axis with reason and citation, exemption with empty reason, exemption
  citing an absent file or a file that does not mention the axis, empty reference set, family below
  the floor), and the same comparison runs once against the real tree. The synthetic cases are the
  re-executable RED once the guards are committed. A reference is counted by either form of
  spec.md §A.6 — the identifier `config.<Name>` or the axis's quoted literal value — so a literal
  such as `os.Getenv("MOAI_AUTONOMY_TIER")` (`internal/cli/codex_sync_gate.go:262`) is not missed;
  the guard reads the axis values from `envkeys.go` and builds the quoted needle at run time, so no
  quoted `"MOAI_FACTORY_ROLE"` literal appears in `internal/hook` source (the one literal that
  `internal/config/envkeys_factory_role_test.go:60` forbids there, spec.md §G).
- **D4 — guards are ordinary test functions, not `TestMain`** (spec.md §H O4). Liveness: the
  coverage test asserts its own swept count (non-zero references found, non-empty scrub set, family
  size at least the floor 17 recorded at c2) and prints it under `-v`; every liveness assertion
  reports with `t.Errorf`, so the comparison always runs to completion; it also asserts that the
  sibling package's guard file exists and declares its guard tests (the unasked stale-guard
  signal, spec.md §E). The floor is a constant in the guard file, lowered in the same change that
  deliberately removes a family constant; the reference scan sees Go source only (spec.md §G R4).
- **D5 — hook scrub-set variable is declared empty in c2** (so the tree compiles and the guards are
  red for the right reason: the set is empty against referenced axes) and filled in c4; the scrub
  function and its `TestMain` call are added in c4, not c2, so c2 carries no unused symbol. The
  hook coverage test is therefore red at c2 for **two** reasons, both reported because liveness
  uses `t.Errorf` (D4): the empty-scrub-set liveness message and the thirteen uncovered-axis
  names; the cli coverage test at c2 (non-empty scrub set) carries the six axis names only; the
  applied tests carry the surviving axes. The c2r record states which message(s) the observed red
  carries, per package and per test.
- **D6 — the whole-package arms are real arms.** The lane arm reproduces the measuring session's
  family env and unsets the other family axes; the scrubbed arm unsets the full family; both run
  the whole package with no `-run` selector, and each arm's swept count is checked against the
  independently listed count L of the same tree (acceptance.md §D.3), so an arm cannot pass by
  dying early.
- **D7 — the applied-behaviour test is a self re-execution.** One test function per package, two
  modes keyed on a witness variable: in the parent it builds the child env explicitly (parent env
  minus every family axis and minus the cli pin marker, plus every family axis set to a sentinel,
  plus the witness) and runs `os.Args[0] -test.run=^<ThisTest>$ -test.v`; in the child it asserts
  absence of every referenced non-exempt axis after the child's `TestMain`. The probe never calls
  the scrub function itself — only `TestMain` may, or the check proves nothing. Chosen over a
  source scan of `TestMain` because it reads behaviour, not text (spec.md §E). Three mechanics
  the probe carries: (a) the child runs under a bounded `context.WithTimeout`
  (`exec.CommandContext`), as the existing precedent does at
  `internal/cli/codex_launcher_exec_posix_test.go:121` (20 s); the cli child's `TestMain` also runs
  `warmUpCommandTree` (`internal/cli/main_test.go:380`), which is slow at a machine load near 50,
  so the bound is sized from a child runtime measured and recorded under that load, never below
  the precedent's 20 s; (b) the parent logs the child's combined output **on success as well as
  on failure** (`t.Log`), because AC-THE-008 requires a child `--- PASS:` line visible in the log;
  (c) the family axes are removed from the child env by **case-insensitive** key comparison,
  because Windows env names are case-insensitive and the CI matrix includes Windows (spec.md R3).
- **D8 — exemption rows are `{axis, reason, citation}`.** The coverage test requires a non-empty
  reason and a citation naming a `*_test.go` file in the package directory that exists and
  references the axis; the table starts empty and a row is added only when a measurement shows a
  test needing the ambient value — the M4 whole-package scrubbed-arm test that went red when the
  axis was stripped, named in progress.md §E.2 beside the row (acceptance.md §D.5). A citation
  alone does not prove that need, and the variant of an unscrubbed axis plus a padded row with a
  real citation is left to review (acceptance.md §D.1).
- **D9 — guard test names (fixed so AC commands are exact).** cli: `TestFactoryEnvAxesCovered`
  (coverage) and `TestFactoryEnvAxesScrubApplied` (applied); hook: `TestLaneEnvAxesCovered` and
  `TestLaneEnvAxesScrubApplied`.

## §F Milestones

### M1 — Baseline record, then the RED recurrence guards — Priority High
Files (exact):
- c1 (record only, subject and file set per §D.1): progress.md §E.2 baseline — L0 for both packages
  (§C step 6); the whole-package pairs of `internal/cli` and `internal/hook` on the pre-guard tree
  (4 leased runs, acceptance.md §D.3), each with command, exit code, failing-test list, terminal
  top-level count T against L, lease acquire/release lines, and the sorted failing-name file of
  each arm (acceptance.md §D.3 commands 6-7, copied verbatim — the c1 side of clause (e)); the
  lane-arm env line of each package's c1 lane arm and the clause (f) counts of §D.3 command 10
  (3 / 2 on the lane names files, 0 on the scrubbed ones); the hook child census; the discovery
  narrow pair re-recorded (E-7 is the plan-time measurement); the session family env read (§C step 2).
- c2 (guards): `internal/cli/factory_env_axes_test.go` (new — cli coverage test and applied test;
  reads `factoryAmbientEnvKeys` from `factory_test.go`); `internal/hook/lane_env_axes_test.go`
  (new — hook coverage test and applied test, the **empty** scrub-set declaration, and the
  exemption table skeleton). No other file.
- c2r (record only): progress.md §E.2 — for each package the coverage test and the applied test run
  on the c2 tree, on a plain shell and under the lane env (it must be red in both, being
  ambient-independent): command, verbatim output naming the uncovered axes (expected from spec.md
  §A.6: six cli axes, thirteen hook axes), exit code, and the c2 SHA. For each coverage test the
  cell states which message(s) the observed red carries — expected: the hook coverage test, the
  empty-scrub-set liveness message **and** the thirteen axis names (D5); the cli coverage test,
  the six axis names only. The record is written so that the content witness of AC-THE-005
  (acceptance.md §D.8 step 7) holds: each of the four guard-test `--- FAIL:` lines is present and
  the c2 SHA is named.
Steps: pre-flight (§C); record c1; write c2; run each guard pair and record the **red**; record c2r.
Exit: AC-THE-005

### M2 — internal/cli scrub-set fix + measured diff — Priority High
Files: `internal/cli/factory_test.go` (extend `factoryAmbientEnvKeys` with `config.EnvFactoryRole`
— the observed-red cause, E-1 / E-1b — and the other five uncovered axes `EnvAutonomyTier`,
`EnvFactoryClearPolicy`, `EnvFactoryAutoDispatch`, `EnvMoaiFactoryManaged`,
`EnvMoaiFactorySlowLaunchMS`; any axis left out is added to the guard's exemption table with a
reason and a cited test file in `internal/cli/factory_env_axes_test.go`). Default is to scrub all
six: static read at plan time shows every test site that touches the five non-Role axes sets or
clears it itself (SlowLaunchMS: five `t.Setenv` sites, no ambient read), so an exemption needs a
measurement that shows a test going red because the axis was stripped — which only the M4
whole-package scrubbed arm can show; M2 does not decide it alone.
Steps: apply; run the narrow AC command of AC-THE-001 under the explicit lane env → green with 3
`--- PASS`; the cli coverage test and applied test green; the existing pin/clear tests
(`TestFactoryAmbientEnvClearedInTestMain`, `TestFactoryEnvPinnedSkipsTestMainClear`) green.
Whole-package pairs are measured once at M4 on the final tree; M2 does not run them.
Exit: AC-THE-001

### M3 — internal/hook fix — Priority High
Files: `internal/hook/lane_env_axes_test.go` (fill the scrub-set declaration with the thirteen
referenced axes; add the scrub function; finalize the exemption table), `internal/hook/main_test.go`
(`TestMain` calls the scrub before the first test, beside the existing `CLAUDE_PROJECT_DIR` scrub;
a re-executed child composing env is honored through the pin-marker pattern if the §C census found
one).
Steps (first): isolate the responsible axis for the two StaleRunNotice tests by a one-axis-at-a-time
measurement **on the then-current tree** — five single-axis arms of the form `unset <AXIS> && go
test ./internal/hook -count=1 -v -run '^(TestStaleRunNoticeLegacyLeaderSpelling|TestStaleRunNoticeLegacySessionRecord)$'`
(plan-phase result, to be re-observed: `MOAI_KANBAN_ID` and `MOAI_FACTORY_WORKERS` each alone flip
the verdict; ROLE/WORKER/BACKEND do not) — record all five before choosing the set. These arms are
scoped to the two observed reds, whose gate reads only the ID ∧ WORKERS conjunction; the other
eight referenced axes (`MOAI_KANBAN`, `MOAI_KANBAN_SPEC`, `MOAI_KANBAN_LABEL`,
`MOAI_KANBAN_SETTINGS_INJECTED`, `MOAI_KANBAN_LEAD_ADDR`, `MOAI_KANBAN_CARD`,
`MOAI_KANBAN_LEAD_NAME`, `MOAI_FACTORY_AUTO_DISPATCH`) are **not** isolated by single-axis arms —
an arm that removes one of them cannot move those two tests — and are decided by the guard pair
and AC-THE-003's whole-package scrubbed arm. Then apply and run the narrow AC command of
AC-THE-002 under the explicit lane env → green with 2 `--- PASS`; the positive control, the hook
coverage test and the hook applied test green (the green path of AC-THE-004 and AC-THE-008, which
are bound to the M4 exit because their mutation probes run there).
Exit: AC-THE-002, AC-THE-007

### M4 — Contingent per-test pins, remaining survey, whole-package pairs, mutation probes — Priority Medium
Files: only those measured red by the M1 baseline pairs that M2/M3 did not already fix (named in
progress.md §E.2 before editing); plus the final measurement record.
Steps: (1) `go test -list` L1 for both packages (L1 ≥ L0 + 2 for the two guard tests of each) and
the final whole-package pairs on the post-M3 tree under the lease — cli lane arm, cli scrubbed arm,
hook lane arm, hook scrubbed arm — with the validity checks of acceptance.md §D.3 (T = L, no timeout
panic / goroutine-leak / non-test failure line), recording failing sets and their difference
(AC-THE-003), and the c1-containment check of clause (e) for each arm type of each package
(acceptance.md §D.3 commands 6-8: `final − c1` by `comm -13`, required empty; a name outside c1
is a change-induced regression, not env-unrelated), plus command 9 (`comm -3` of the two arms'
names, empty) and clause (f): each final lane arm replays the c1 lane-arm env line byte for byte
(never the then-current session env) and the two lines per package are compared; (2) for any test
still differing: RED/GREEN pair, Option-B pin with `t.Setenv` of
every axis the code path reads, `t.Parallel()` dropped where B1 applies, no skip. A test that
went red in the all-axes-unset scrubbed arm because an axis was stripped is fixed the same way, by
a pin that makes its verdict arm-independent; only when a test truly needs the ambient value does the
axis take an exemption row (D8) instead, and the scrubbed arm is then re-run leaving that axis at its
lane value (acceptance.md §D.3), so the axis is identical-valued in both arms and the equality
requirement holds with the row in place; (3) mutation probes, each
recorded and then reverted (confirmed by an empty `git diff --stat`): P1 remove one axis from a
declared scrub set → coverage test red naming it; P2 remove the `TestMain` scrub call (cli
`clearFactoryAmbientEnv()`, hook scrub call) → applied test red naming every surviving axis;
P3 pad an exemption row (reason "n/a", citation to a file that does not mention the axis) →
coverage test red; P4 point the reference scan at an empty directory → liveness red; P5 move one
guard file aside → the sibling package's coverage test red; (4) `go vet` + Windows `GOOS=windows
GOARCH=amd64 go vet ./internal/cli ./internal/hook` + `golangci-lint run` delta vs the §C baselines;
(5) the AC-THE-005 / AC-THE-006 checks from `git log` — for AC-THE-005 the step 6 enumeration is
re-evaluated here (M1 evaluated the shape, the chain and the step 7 content witness); (6) list the
final exemption tables of both guard files in progress.md §E.2 with, for each surviving row, the
name of the whole-package scrubbed-arm test that went red when that axis was stripped (an empty
table needs no row); (7) populate progress.md §E.2/§E.3.
Exit: AC-THE-003, AC-THE-004, AC-THE-006, AC-THE-008

### M5 — Sync — Priority Low
manager-docs owns `progress.md` §E.4 and the `implemented → completed` transition on the single
sync commit. Test-only change; the sync commit touches the SPEC directory only and no CHANGELOG
entry is added (a test-only change; REQ-THE-008's diff therefore holds at the sync tip as at M4).
The leader integrates per the git-flow lane protocol and pushes in batch.

## §G Anti-Patterns

- **AP-1** — skipping, deleting, or listing a flip-prone test (REQ-THE-004).
- **AP-2** — a "lane arm" that is only a partial env, or a "scrubbed arm" that unsets a subset
  (B4); both measure nothing.
- **AP-3** — declaring the sweep done from the static survey's list.
- **AP-4** — writing the fix first and "confirming" the guard red afterwards (REQ-THE-006).
- **AP-5** — citing a `-run` selector result as the whole-package verdict (AC-THE-003 requires the
  unselected package).
- **AP-6** — local `go test ./...`.
- **AP-7** — a guard whose reference scan matches nothing and passes (D4).
- **AP-8** — an exemption without a reason, or with a citation that proves nothing (D8).
- **AP-9** — adding an axis to the scrub set because it is convenient, without checking that a
  test does not rely on its ambient value (spec.md §D precondition (i)).
- **AP-10** — an arm accepted because it reported a non-zero swept count: both arms can die early
  for the same env-independent reason (timeout panic, goroutine-leak report); only the independent
  listed count L catches it.
- **AP-11** — an applied-behaviour probe that calls the scrub function itself, or runs with the
  parent's ambient env: it then proves the function works, not that `TestMain` applies it.
- **AP-12** — naming a failure env-unrelated because it is identical in both final arms, without
  checking that it is in the c1 failing set of the same arm type (clause (e)): a change-induced
  failure can be identical in both arms. The check is by full test path, subtests included: a
  top-level-only comparison misses a new failing subtest under an already-red parent.
- **AP-13** — a "lane arm" built from a session that carries no lane axes, or accepted without the
  five reds in its c1 failing set (clause (f)): both arms are then equal and empty, and the
  criterion passes without exercising the lane env.

## §H Cross-References

- spec.md §A evidence and family table, §C REQ-THE-001..009, §D options, §E guard spec, §F ordering,
  §H open questions.
- acceptance.md §D — AC matrix and evidence ledger.
- `.claude/rules/moai/development/verification-completeness.md` §1.2, §1.3, §2, §2.1;
  `.claude/rules/moai/core/verification-claim-integrity.md` §2.3, §3.1;
  `.claude/rules/local/gitflow-lane-protocol.md` §8 (slot lease, verification scope, range base).
- SPEC-CLI-TEST-CWD-ISOLATION-001 plan.md — structure and RED-first precedent.
