# Acceptance — SPEC-TEST-ENV-HERMETIC-001

All commands run from the worktree root
(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1356`). Whole-package runs: serial, under a
`moai slot` lease, each as a background Bash call, because the foreground Bash maximum is 600 s
while `-timeout` is the lease cap minus 2m (at least 18m; plan.md §C step 5). Document-level tree pin:
**`a5a63a0bc`** — HEAD at the iteration-2 measurements; its Go sources are identical to the
original measurement tree `2de0a2cb6` (`git rev-list --count 2de0a2cb6..a5a63a0bc` prints `1`, and
that commit's `git diff-tree` lists five `.moai/specs/SPEC-TEST-ENV-HERMETIC-001/` paths and no Go
file). The pin binds every RED-now cell below that carries no pin of its own, and a RED-now command
is re-executable on any descendant tree that does not yet contain plan.md milestones M1-M4.
`<BASE>` in the git-range checks is `2de0a2cb6` until the branch absorbs local `develop`; after an
absorption it is re-derived once with `git merge-base develop HEAD` and recorded (plan.md B10).

## §D.0 Evidence ledger (RED-now carrier)

Each entry carries the four elements of `verification-completeness.md` §2.1: the command, its
stdout, its exit code as its own field, and the tree SHA. A **single-invocation** command has no
pipe, `&&`, or `;`; entries E-1b, E-3 and the E-7 scrubbed arm are compound (`unset … && go test
…`, the form the lane protocol prescribes for a scrubbed arm), so they are labelled **diagnostic**
and are not RED-now cells. Stdout is quoted as the run returned it (the tool merges stdout and
stderr); durations and temp-directory suffixes vary run to run, and the judged content is the
`--- FAIL` / `--- PASS` lines and the assertion messages. `env NAME=value go test …` sets the
named axes on top of whatever the session already exports, so each ledger command names **every**
axis it depends on explicitly (a partial env measures nothing, plan.md B4).

### E-1 — internal/cli: three tests red when `MOAI_FACTORY_ROLE=lane` reaches the binary

- tree: `a5a63a0bc` (Go sources = `2de0a2cb6`)
- command: `env MOAI_FACTORY_ROLE=lane go test ./internal/cli -count=1 -v -run '^(TestTodoClaim_LaneGovernance|TestTodoClaimMCP_Mirror|TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit)$'`
- exit code: `1`
- stdout (as returned, including the trailing `FAIL` lines):

```
=== RUN   TestTodoClaim_LaneGovernance
    todo_claim_test.go:291: arm 2 claim --lane: moai claim: refused — lane boundary: a lane session cannot mutate the queue (read-only here: bare todo, list, history, show, why, pr, triage); a lane takes its next card through moai factory next
--- FAIL: TestTodoClaim_LaneGovernance (1.27s)
=== RUN   TestTodoClaimMCP_Mirror
    todo_claim_test.go:416: mcp todo_claim: todo_claim: moai claim: refused — lane boundary: a lane session cannot mutate the queue (read-only here: bare todo, list, history, show, why, pr, triage); a lane takes its next card through moai factory next
--- FAIL: TestTodoClaimMCP_Mirror (0.42s)
=== RUN   TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit
    todo_test.go:190: moai add: refused — lane boundary: a lane session cannot mutate the queue (read-only here: bare todo, list, history, show, why, pr, triage); a lane takes its next card through moai factory next
--- FAIL: TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit (0.17s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.379s
FAIL
```

- **why red**: `TestMain` clears the 12 keys of `factoryAmbientEnvKeys` and that set lacks
  `MOAI_FACTORY_ROLE`; the lane-refusal predicate (`internal/cli/factory_card.go:57-71`) reads it,
  so the role marker reaches the tests and the queue mutation is refused. `TestTodoClaim_LaneGovernance`
  pins only the lane-label axis, so the role axis stays live.

### E-1b — internal/cli scrubbed arm: the same three tests pass when only `MOAI_FACTORY_ROLE` is unset (diagnostic)

Compound form, so outside the single-invocation RED-now form: it is the green counterpart that
attributes E-1's red to that one axis. The session's other lane axes stay ambient.

- tree: `a5a63a0bc`
- command: `unset MOAI_FACTORY_ROLE && go test ./internal/cli -count=1 -v -run '^(TestTodoClaim_LaneGovernance|TestTodoClaimMCP_Mirror|TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit)$'`
- exit code: `0`
- stdout (as returned; the two lines inside the third test are test-binary stderr):

```
=== RUN   TestTodoClaim_LaneGovernance
--- PASS: TestTodoClaim_LaneGovernance (3.09s)
=== RUN   TestTodoClaimMCP_Mirror
--- PASS: TestTodoClaimMCP_Mirror (1.52s)
=== RUN   TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit
moai-cli-test: userHomeDirFn redirected real home to sandbox /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/moai-cli-home-3663858279
2026/10/03 12:19:12 WARN config sections directory not found, using defaults path=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit4110972749/001/.moai/config/sections
--- PASS: TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit (3.47s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	9.397s
```

### E-2 — internal/hook: two tests red under the lane env

- tree: `a5a63a0bc` (Go sources = `2de0a2cb6`)
- command: `env MOAI_FACTORY_ROLE=lane MOAI_FACTORY_WORKER=lane-6 MOAI_FACTORY_WORKERS=0 MOAI_KANBAN_BACKEND=claude MOAI_KANBAN_ID=tm9i7y go test ./internal/hook -count=1 -v -run '^(TestStaleRunNoticeLegacyLeaderSpelling|TestStaleRunNoticeLegacySessionRecord)$'`
- exit code: `1`
- stdout (as returned, including the trailing `FAIL` lines):

```
=== RUN   TestStaleRunNoticeLegacyLeaderSpelling
    stale_run_m1_test.go:42: staleRunNoticeFor = "", want a notice
--- FAIL: TestStaleRunNoticeLegacyLeaderSpelling (0.13s)
=== RUN   TestStaleRunNoticeLegacySessionRecord
    stale_run_m1_test.go:85: staleRunNoticeFor = "", want a notice naming the legacy record role
--- FAIL: TestStaleRunNoticeLegacySessionRecord (0.10s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	1.191s
FAIL
```

- **why red**: the hook `TestMain` scrubs none of the factory axes. With `MOAI_KANBAN_ID` and
  `MOAI_FACTORY_WORKERS` both non-empty, `staleRunNoticeFor` routes through the run-state gate
  (`internal/hook/stale_run_gate.go:259-260`), which answers "" for a run that is not recorded
  active; the tests expect the ungated relaunch notice.

### E-3 — internal/hook one-axis arms (diagnostic, not a RED-now cell)

These arms are compound and therefore sit outside the single-invocation RED-now form; they
explain E-2 and are re-observed at M3. Measured at plan time on tree `2de0a2cb6` and not re-run in
iteration 2 (the Go sources are identical at `a5a63a0bc`); each arm starts from the session's
ambient lane env (`MOAI_FACTORY_ROLE`, `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_WORKERS`,
`MOAI_KANBAN_BACKEND`, `MOAI_KANBAN_ID` all set) and removes only the named axis. Selector:
`-count=1 -v -run '^(TestStaleRunNoticeLegacyLeaderSpelling|TestStaleRunNoticeLegacySessionRecord)$'`.

| arm | command prefix | exit | verdict lines (verbatim) |
|-----|---------------|------|--------------------------|
| unset `MOAI_KANBAN_ID` | `unset MOAI_KANBAN_ID && go test ./internal/hook …` | 0 | `--- PASS: TestStaleRunNoticeLegacyLeaderSpelling (0.00s)` · `--- PASS: TestStaleRunNoticeLegacySessionRecord (0.00s)` · `ok  	github.com/modu-ai/moai-adk/internal/hook	0.751s` |
| unset `MOAI_FACTORY_WORKERS` | `unset MOAI_FACTORY_WORKERS && go test ./internal/hook …` | 0 | same two `--- PASS` lines · `ok  	github.com/modu-ai/moai-adk/internal/hook	0.672s` |
| unset `MOAI_FACTORY_ROLE` + `MOAI_FACTORY_WORKER` + `MOAI_KANBAN_BACKEND` | `unset MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_KANBAN_BACKEND && go test ./internal/hook …` | 1 | `stale_run_m1_test.go:42: staleRunNoticeFor = "", want a notice` · `--- FAIL: TestStaleRunNoticeLegacyLeaderSpelling (0.08s)` · `stale_run_m1_test.go:85: staleRunNoticeFor = "", want a notice naming the legacy record role` · `--- FAIL: TestStaleRunNoticeLegacySessionRecord (0.08s)` · `FAIL	github.com/modu-ai/moai-adk/internal/hook	0.884s` |

Reading: each of `MOAI_KANBAN_ID` and `MOAI_FACTORY_WORKERS` alone is sufficient to flip the two
tests to PASS when removed; the other three axes are not. The responsible condition is the
conjunction of the two (matching the source gate). Not isolated: `MOAI_KANBAN` /
`MOAI_KANBAN_LEAD_NAME`, which the tests set themselves. These arms say nothing about the eight
other axes the hook package references (plan.md M3).

### E-4 — positive controls, green under the lane env (basis of AC-THE-007)

- E-4a — tree `2de0a2cb6`, exit `0`:
  `env MOAI_FACTORY_ROLE=lane go test ./internal/cli -count=1 -v -run '^(TestFactoryEnvPinnedSkipsTestMainClear|TestFactoryAmbientEnvClearedInTestMain)$'`

```
=== RUN   TestFactoryAmbientEnvClearedInTestMain
--- PASS: TestFactoryAmbientEnvClearedInTestMain (0.00s)
=== RUN   TestFactoryEnvPinnedSkipsTestMainClear
--- PASS: TestFactoryEnvPinnedSkipsTestMainClear (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.230s
```

- E-4b — tree `2de0a2cb6`, exit `0`:
  `env MOAI_FACTORY_ROLE=lane MOAI_FACTORY_WORKER=lane-6 MOAI_FACTORY_WORKERS=0 MOAI_KANBAN_BACKEND=claude MOAI_KANBAN_ID=tm9i7y go test ./internal/hook -count=1 -v -run '^TestContractRoleScopedAllowWithoutLaneMarker$'`

```
=== RUN   TestContractRoleScopedAllowWithoutLaneMarker
=== RUN   TestContractRoleScopedAllowWithoutLaneMarker/marker_unset
=== RUN   TestContractRoleScopedAllowWithoutLaneMarker/marker_other_value
=== RUN   TestContractRoleScopedAllowWithoutLaneMarker/marker_legacy_worker
=== RUN   TestContractRoleScopedAllowWithoutLaneMarker/marker_legacy_agent
--- PASS: TestContractRoleScopedAllowWithoutLaneMarker (0.00s)
    --- PASS: TestContractRoleScopedAllowWithoutLaneMarker/marker_unset (0.00s)
    --- PASS: TestContractRoleScopedAllowWithoutLaneMarker/marker_other_value (0.00s)
    --- PASS: TestContractRoleScopedAllowWithoutLaneMarker/marker_legacy_worker (0.00s)
    --- PASS: TestContractRoleScopedAllowWithoutLaneMarker/marker_legacy_agent (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.763s
```

Disclosure (iteration 2): `TestFactoryAmbientEnvClearedInTestMain` asserts the launch gate and
three keys (`MOAI_KANBAN_ID`, `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_WORKERS`,
`internal/cli/factory_test.go:96-111`) and never reads `MOAI_FACTORY_ROLE`, so E-4a is green before
and after adding that axis to the scrub set and cannot separate the two states; it is kept as the
pin-marker preservation control only. The control that does read the added axes is the
applied-behaviour test of AC-THE-008.

### E-5 — the guard commits do not exist yet (basis of AC-THE-005)

Tree `a5a63a0bc`, each exit `0` with empty stdout (an empty stdout with exit 0 is a complete
observation: no such commit exists). Positive controls prove each form lists a commit when one
exists.

- E-5a (c1 absent): `git log --format='%h %s' --reverse 2de0a2cb6..HEAD --grep='^test(SPEC-TEST-ENV-HERMETIC-001): baseline record (card t1356)$'` → stdout empty.
- E-5a2 (c2r absent): `git log --format='%h %s' --reverse 2de0a2cb6..HEAD --grep='^test(SPEC-TEST-ENV-HERMETIC-001): guard red record (card t1356)$'` → stdout empty.
- E-5b (c2 absent): `git log --format=%h --reverse --diff-filter=A 2de0a2cb6..HEAD -- internal/cli/factory_env_axes_test.go internal/hook/lane_env_axes_test.go` → stdout empty.
- E-5c (no internal commit yet): `git rev-list --no-merges --reverse 2de0a2cb6..HEAD -- internal` → stdout empty.
- Control for the `--grep` form: `git log --format='%h %s' --reverse 2de0a2cb6..HEAD --grep='^feat(SPEC-TEST-ENV-HERMETIC-001): plan-phase artifacts'` → `a5a63a0bc feat(SPEC-TEST-ENV-HERMETIC-001): plan-phase artifacts (Tier M, card t1356)`.
- Control for the `--diff-filter=A` form: `git log --format=%h --reverse --diff-filter=A 2de0a2cb6..HEAD -- .moai/specs/SPEC-TEST-ENV-HERMETIC-001/spec.md` → `a5a63a0bc`.
- Control for the `rev-list … -- <path>` form, with an **explicit upper bound**: `git rev-list --no-merges --reverse 2de0a2cb6..a5a63a0bc -- .moai` → `a5a63a0bc0a8bd0125bdf881b98dfee20c9746ae`. (Written with `..HEAD` the same form prints two lines once the revision commit `669cf18c9` exists — re-measured at HEAD `669cf18c9`: `a5a63a0bc0a8…` and `669cf18c98233c5a5086d45de7851be745e14775` — so an open upper bound is not a stable control; E-5a, E-5a2, E-5b and E-5c re-run at `669cf18c9` still print empty stdout.)
- E-5d (c2r content witness absent; basis of the AC-THE-005 step 7), tree `669cf18c9` named as the tree-ish in every command (never `HEAD`, whose content moves): `git grep -c -e '--- FAIL: TestFactoryEnvAxesCovered ' 669cf18c9 -- .moai/specs/SPEC-TEST-ENV-HERMETIC-001/progress.md` → stdout empty, exit `1` (git grep's no-match status; observed with a trailing `; echo "exit=$?"`). Controls for the form, same tree: `git grep -c -e '--- FAIL: TestTodoClaim_LaneGovernance ' 669cf18c9 -- .moai/specs/SPEC-TEST-ENV-HERMETIC-001/acceptance.md` → `669cf18c9:.moai/specs/SPEC-TEST-ENV-HERMETIC-001/acceptance.md:1`, exit `0`; `git grep -c --all-match -e '--- FAIL: TestTodoClaim_LaneGovernance ' -e '--- FAIL: TestStaleRunNoticeLegacyLeaderSpelling ' 669cf18c9 -- .moai/specs/SPEC-TEST-ENV-HERMETIC-001/acceptance.md` → `669cf18c9:.moai/specs/SPEC-TEST-ENV-HERMETIC-001/acceptance.md:3`, exit `0`; the same `--all-match` with the second name replaced by the absent `--- FAIL: TestFactoryEnvAxesCovered ` (with its trailing space) → stdout empty, exit `1`; `git grep -c -e 'REQ-THE-001' 669cf18c9 -- .moai/specs/SPEC-TEST-ENV-HERMETIC-001/spec.md` → `669cf18c9:.moai/specs/SPEC-TEST-ENV-HERMETIC-001/spec.md:2`, exit `0`.
- Control for the strict-ancestor predicate (why it is two tests, not one):
  `git merge-base --is-ancestor a5a63a0bc a5a63a0bc` exits `0` and `git rev-list --count a5a63a0bc..a5a63a0bc` prints `0`
  (a commit is an ancestor of itself, but the count of commits between is zero);
  `git merge-base --is-ancestor 2de0a2cb6 a5a63a0bc` exits `0` and `git rev-list --count 2de0a2cb6..a5a63a0bc` prints `1`;
  `git merge-base --is-ancestor a5a63a0bc 2de0a2cb6` exits `1`.
- Control for the file-set check: `git diff-tree --no-commit-id --name-only -r a5a63a0bc` lists the
  five `.moai/specs/SPEC-TEST-ENV-HERMETIC-001/{acceptance,decision-index,plan,progress,spec}.md` paths.

### E-6 — refused-tool disclosure (not a measurement)

Plan-phase and iteration-2 attempts refused by the worktree-isolation guard, and one class of
partial arm that was discarded; none supplies a cited figure (`verification-claim-integrity.md` §3.1):
(1) `env -u <AXIS> go test …` was refused ("cannot be shown not to be git"); the `unset <AXIS> &&
go test …` form of E-3 replaced it. (2) A first group of arms of the form `env <subset> go test …`
inherited the remaining ambient axes (they set, never unset) and measured nothing; discarded.
(3) A multi-SPEC loop with a runtime-computed path in `sed` was refused; plain separate commands
replaced it. (4) Iteration 2: `go test -list '.*' ./internal/discovery | tail -40; echo
"exit=${PIPESTATUS[0]}"` was refused (pipe plus `PIPESTATUS` after a `go` command); the plain `go
test -list '.*' ./internal/discovery` replaced it. (The iteration-2 note that `go test -list`
for `./internal/cli` and `./internal/hook` could not be taken is withdrawn: with a warm build cache
it takes seconds and runs no test; the plan-time figures are E-8.)

### E-7 — internal/discovery narrow pair (diagnostic; basis of the survey claim)

Tree `a5a63a0bc`, run from inside the measuring lane session (its own family env is nine axes).
The package has no `TestMain` scrub and references one family axis in production (`MOAI_KANBAN_ID`).
`go test -list '.*' ./internal/discovery` printed 15 test names (exit 0).

- E-7a lane arm — `env MOAI_FACTORY_ROLE=lane MOAI_FACTORY_WORKER=lane-6 MOAI_FACTORY_WORKERS=0 MOAI_KANBAN_BACKEND=claude MOAI_KANBAN_ID=tm9i7y MOAI_AUTONOMY_TIER=fully-autonomous MOAI_FACTORY_AUTO_DISPATCH=auto go test ./internal/discovery -count=1 -v` — exit `0`; verdict lines: 14 `--- PASS`, `--- SKIP: TestHelperLeaderChild (0.00s)` (`factory_discovery_live_test.go:91: helper only`), `PASS`, `ok  	github.com/modu-ai/moai-adk/internal/discovery	9.589s`.
- E-7b scrubbed arm — `unset MOAI_AUTONOMY_TIER MOAI_KANBAN MOAI_KANBAN_SPEC MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_BACKEND MOAI_KANBAN_CARD MOAI_KANBAN_LEAD_NAME MOAI_FACTORY_WORKERS MOAI_FACTORY_SLOW_LAUNCH_MS MOAI_FACTORY_WORKER MOAI_FACTORY_MANAGED MOAI_FACTORY_ROLE MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_AUTO_DISPATCH && go test ./internal/discovery -count=1 -v` — exit `0`; the same 14 `--- PASS` and the same one `--- SKIP`, `ok  	github.com/modu-ai/moai-adk/internal/discovery	13.855s`.
- E-7c floor-form control — `unset MOAI_FACTORY_ROLE && go test ./internal/discovery -count=1 -json > <file>` exit `0`; on the file, `grep -cE '"Action":"(pass|fail|skip)".*"Test":"[^/"]+"' <file>` printed `15` (14 `"Action":"pass"`, 1 `"Action":"skip"`, 0 `"Action":"fail"`), equal to the 15 listed names. It shows the redirect form passes the guard and that skips are terminal rows, so the floor counts `pass`, `fail` and `skip`.
- E-7d lane-arm form control — `unset MOAI_KANBAN MOAI_KANBAN_SPEC MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_CARD MOAI_KANBAN_LEAD_NAME MOAI_FACTORY_SLOW_LAUNCH_MS MOAI_FACTORY_MANAGED && env MOAI_FACTORY_ROLE=lane MOAI_FACTORY_WORKER=lane-6 MOAI_FACTORY_WORKERS=0 MOAI_KANBAN_BACKEND=claude MOAI_KANBAN_ID=tm9i7y MOAI_AUTONOMY_TIER=fully-autonomous MOAI_FACTORY_AUTO_DISPATCH=auto MOAI_KANBAN_SETTINGS_INJECTED=1 MOAI_FACTORY_CLEAR_POLICY= go test ./internal/discovery -count=1 -json > <file>` — exit `0` (the compound lane-arm form of §D.3 passes the guard); on the file the terminal-count grep printed `15` and the invalid-marker grep of §D.3 command 4 printed `0`.

Reading: the two arms are equal (same pass set, same pre-existing helper skip), so no `internal/discovery` test flips under the lane env on this tree.

### E-8 — plan-time listed test counts L for cli and hook (diagnostic; basis of the L figures of §D.3)

Tree `669cf18c9` — HEAD at iteration 3; its Go sources equal `2de0a2cb6` (`git diff --name-only
2de0a2cb6..669cf18c9` lists the five SPEC-directory paths and no Go file). Warm build cache, machine
load average above 50. `go test -list` compiles the test binary and prints test names; it runs no
test. Each count takes two plain commands (the listing redirected to a file, then a count on it):

- E-8a — `go test -list '.*' ./internal/cli > <file>` exit `0`; `grep -cE '^(Test|Example|Fuzz)' <file>` exit `0`, stdout `4884`. The lines of the file not matching are `BenchmarkIsTrivialCommand`, `BenchmarkTodoAuditAnalyze` and `ok  	github.com/modu-ai/moai-adk/internal/cli	1.833s`.
- E-8b — `go test -list '.*' ./internal/hook > <file>` exit `0`; the same `grep -cE` count, stdout `1322`. The lines not matching are `BenchmarkConfigChange_AsyncReturn`, `BenchmarkFileChanged_AsyncReturn`, `BenchmarkNotification_AsyncReturn`, `BenchmarkSubagentWriteGuardDenyPath`, `BenchmarkTaskCreated_AsyncReturn` and `ok  	github.com/modu-ai/moai-adk/internal/hook	1.035s`.

Plan-time L: cli **4884**, hook **1322**. The run phase records L0 again on the pre-guard tree (it
differs if the branch absorbs `develop`, plan.md B10).

### E-9 — form controls for §D.3 commands 6-11 and the AC-THE-006 diff greps (diagnostic)

Iteration 4. Scratch Go module outside the repository (tree-independent: it demonstrates the
extraction forms, not repository behaviour; arms are `go test -C <module> ./lane -count=1 -json`,
each exit `1` unless stated, files sorted with `LC_ALL=C sort -o`). Selector
`^(TestLegacyLeader|TestLegacyRecord)$`, gate red only when `LANE_ID` and `LANE_WORKERS` are both set.

- Command 6 as now written (`"Test":"[^"]+"`, subtests included): arm A (`LANE_ID=x LANE_WORKERS=0`)
  2 names, arm B (same env, one test selected) 1 name, arm C (no env, exit `0`) empty file with grep
  exit `1`. `comm -13 B A` → `"Action":"fail","Package":"mf4/lane","Test":"TestLegacyRecord"`;
  `comm -13 C A` → both names; `comm -13 A B` and `comm -13 A A` → nothing (all exit `0`).
- Subtest case, `TestParent` with subtests `sa` and `sb`; c1 arm `ARM=c1` (sa fails), final arm
  `ARM=final` (sa and sb fail). New form: c1 names `TestParent/sa`, `TestParent`; final names add
  `TestParent/sb`; `comm -13 <c1> <final>` printed `"Action":"fail","Package":"mf4/lane","Test":"TestParent/sb"`,
  exit `0`. The former top-level-only form (`"Package":"[^"]+","Test":"[^/"]+"`) named only
  `"Action":"fail","Package":"mf4/lane","Test":"TestParent"` in both arms (exit `0`) and its
  `comm -13 <c1> <final>` printed **nothing**, exit `0`: the clause (e) hole, closed by the new form.
  `comm -3 <c1> <final>` printed the one `TestParent/sb` line (clause (b) sees the divergence);
  `comm -3 A A` printed nothing; `comm -3 A C` (lane vs scrubbed on the A/C arms above) printed both
  names, the red state.
- Package-level failures with no test-level row: `TestMain` failing after every test passed — T `1`,
  command 4 `0` (grep exit `1`), command 5 `0` (exit `1`), names file empty (exit `1`), arm exit `1`, so
  the arm is invalid by command 5; a build failure (`./bf`) — T `0`, command 4 `1`, command 5 `0`, arm
  exit `1`, invalid by commands 1 and 4. Neither can pass as an empty failing set.
- Clause (f) control on the real packages (tree `47ae8ecb1`, Go sources = `2de0a2cb6`), the compound
  lane form `unset <the eight family axes the session lacks> && env <the nine session axes verbatim> go
  test <pkg> -count=1 -json -run <the observed red names> > <file>`: cli lane arm exit `1`, hook lane
  arm exit `1`; the 17-axis scrubbed arms exit `0`. On the sorted names files command 10 printed `3`
  (cli lane) / `0` (cli scrubbed, grep exit `1`) / `2` (hook lane) / `0` (hook scrubbed, grep exit `1`);
  on the raw cli lane arm file the same pattern printed `15`, because the names also sit on `run`,
  `output` and `pass` rows — hence the names file. `comm -3 <hook lane names> <hook scrubbed names>`
  printed the two StaleRunNotice names (clause (b) is red on the c1 tree, by design).
- The family env read: `env | grep -E '^(MOAI_FACTORY_|MOAI_KANBAN|MOAI_AUTONOMY_TIER)'` printed the nine
  axes of §D.3, exit `0` (a session without family axes would print nothing and exit `1`; not
  observed here, and that is the INVALID outcome of clause (f)).
- AC-THE-006 diff greps: `git diff -U0 --diff-filter=M 2de0a2cb6~60 2de0a2cb6 -- internal/cli
  internal/hook > <file>` is 4176 lines; the removed-assertion grep printed `45` and the bare-`return`
  grep `5`; on `2de0a2cb6..47ae8ecb1` the saved diff has 0 lines and the first grep printed `0`
  (exit `1`).
- MF-3 escape-aware control (iteration 5; scratch module `t1356e9` under `/tmp/t1356-e9`,
  tree-independent — it tests the extraction form, not repository behaviour; go1.26.8
  darwin/arm64; measured at tree `960ea3012`). Runs: `env ARM=<arm> go test -C /tmp/t1356-e9 .
  -count=1 -json`, c1 exit `1`, final exit `1` (all four greps exit `0`). `TestParent` carries
  subtests `q"uote` and `q"uoted` (red arm-switched: `q"uote` on c1, `q"uoted` on final) and
  `back\\slash` (red in both arms). The v0.4.0 regex `"Test":"[^"]+"` truncates both quote-bearing
  names to `"Test":"TestParent/q\"`, so the sorted c1 and final old-form names files are identical
  and `LC_ALL=C comm -13 <c1 old> <final old>` prints **nothing**, exit `0` — the new failing name
  is missed. The escape-aware regex `"Test":"([^"\\]|\\.)+"` extracts
  `"Action":"fail","Package":"t1356e9","Test":"TestParent/q\"uote"`,
  `"Action":"fail","Package":"t1356e9","Test":"TestParent/q\"uoted"` and
  `"Action":"fail","Package":"t1356e9","Test":"TestParent/back\\slash"` whole, and its
  `LC_ALL=C comm -13 <c1 new> <final new>` prints
  `"Action":"fail","Package":"t1356e9","Test":"TestParent/q\"uoted"` — the new failing name —
  exit `0`. The same escape-aware form over the skip rows is §D.3 command 11.

## §D AC Matrix

Classification (`verification-completeness.md` §2.1): **release-blocking** = the RED-now cell is
re-executable on the stated tree and carries command, stdout, exit code, and tree SHA.
**regression-guard** = its RED cannot be re-executed on `a5a63a0bc` (the artifact it tests does not
exist yet, or it is green today); it is not recorded as a pass on that basis. AC-THE-004 and
AC-THE-008 carry the class **regression-guard at plan time; becomes release-blocking when its
four-element cell is recorded at c2r**: nothing at plan time is certified by them, because the
guard does not exist on `a5a63a0bc`, and the label changes only when the four elements (command,
verbatim stdout, exit code, c2 SHA) are recorded at the guard-red record commit c2r on the c2 tree —
a tree on which the command is re-executable — and the content witness of §D.8 step 7 holds. A cell
that cannot exist at plan time is **completed by the record** at the commit it names (c1 for
AC-THE-003's whole-package pairs, c2r for AC-THE-004 and AC-THE-008); that phrase is used for it
throughout. A closure with no such record leaves the AC unadopted and blocks (§D.5), and the sync
auditor re-executes one of the four c2r cells at the c2 tree and records its own verbatim stdout and
exit code (§D.8 step 7). 8 ACs (Tier M ceiling 16).

Exit binding: the `Exit:` line of an AC in plan.md names the milestone at whose end its whole Then
clause can be evaluated; the Green path column names the milestone that flips the green path.
AC-THE-004 and AC-THE-008 turn green at M2 (cli) and M3 (hook) but record their mutation probes at
M4, so they are bound to the M4 exit. AC-THE-005 is bound to M1, which evaluates the shape and the
chain of c1, c2 and c2r (§D.8 steps 0-5 and 7); M4 re-evaluates the full enumeration of step 6.

| AC | REQ | Class | Given / When / Then (judgment) | RED-now cell | Green path |
|----|-----|-------|--------------------------------|--------------|-----------|
| AC-THE-001 | REQ-THE-001 | **release-blocking** | Given the cli test binary; When `env MOAI_FACTORY_ROLE=lane go test ./internal/cli -count=1 -v -run '^(TestTodoClaim_LaneGovernance\|TestTodoClaimMCP_Mirror\|TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit)$'` runs; Then exit 0, exactly three `--- PASS:` lines for these names, zero `--- SKIP`, no `[no tests to run]` | **E-1** (tree `a5a63a0bc`, exit 1; the AC's own command) — red because `MOAI_FACTORY_ROLE` is outside the cli `TestMain` scrub set; diagnostic green arm **E-1b** (compound) shows the same three pass when that one axis is unset | M2 flips it: same command, exit 0, three `--- PASS:` lines |
| AC-THE-002 | REQ-THE-002 | **release-blocking** | Given the hook test binary; When the E-2 command runs; Then exit 0, exactly two `--- PASS:` lines for the two StaleRunNotice tests, zero `--- SKIP`; AND `git diff <BASE> -- internal/hook/stale_run_m1_test.go` leaves the assertion lines (`Fatalf` / `Errorf` / `Contains`) of both tests untouched | **E-2** (tree `a5a63a0bc`, exit 1; the AC's own command) — red because the hook `TestMain` scrubs no factory axis and `MOAI_KANBAN_ID` ∧ `MOAI_FACTORY_WORKERS` route the notice through the run-state gate (E-3) | M3 flips it: E-2 command, exit 0, two `--- PASS:` lines; assertion diff empty |
| AC-THE-003 | REQ-THE-001, REQ-THE-002, REQ-THE-003, REQ-THE-007 | **release-blocking** | Given the post-M3 tree; When the four whole-package runs of §D.3 execute (cli and hook, each a lane arm and a scrubbed arm, no `-run` selector, `-count=1`, under the lease); Then per package (a) each arm is **valid** — exit code recorded, its terminal top-level count T equals the independently listed count L of the same tree, no `panic: test timed out` / goroutine-leak / build-or-setup-failure line, and a non-zero exit shows at least one test-level failure row; (b) the two arms' failing-test sets are equal — by full test path, subtests included, so §D.3 command 9 prints nothing, an axis of a surviving exemption row being left at its lane value in the scrubbed arm and so identical in both — and their skipped sets are equal; (c) both failing sets are empty modulo failures named as env-unrelated, each of which is identical in both arms and present in the c1 failing set; (d) neither set names a guard test or any of the five observed reds; (e) every name in a final-tree arm's failing set (full test path) is also in the c1 failing set of the same arm type — the difference `final − c1` of §D.3 commands 6-8 prints nothing for the lane arm and for the scrubbed arm of each package (a name absent from c1 is a regression caused by the change, never env-unrelated); (f) the lane arm is **a lane**: its env (every set `NAME=value`, every unset name) is recorded as one line per arm in progress.md §E.2 for the c1 and the final lane arm of each package, the four lines are byte-identical (the final arm replays the c1 line, not the then-current session, and the final arm's child-visible env line of §D.3 — the `env`-in-place-of-the-run form — is recorded beside it and identical to the c1 line), each carries every axis of the pre-flight family read (plan.md §C step 2) and at least `MOAI_FACTORY_ROLE=lane` with non-empty `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_WORKERS` and `MOAI_KANBAN_ID`, and in the c1 arms the lane names file holds the five observed reds (§D.3 command 10 prints 3 for cli, 2 for hook) while the c1 scrubbed names file holds none (prints 0). A session whose pre-flight read lacks a required axis, or a c1 lane arm without the five reds, makes the lane arm INVALID and AC-THE-003 FAIL — never skipped, never passed; the run returns a blocker report | Witness by existence: **E-1 + E-2** (tree `a5a63a0bc`) show the lane-arm set contains at least those 5 tests; the scrubbed-arm greens are **E-1b** (cli) and E-3 (hook, unset arms), so the two sets differ now — red because the five observed reds are lane-only. Cell completed by the c1 record: the full pairs on the pre-guard tree, with their sorted failing-name files and the lane-arm env lines, are recorded in the baseline record commit (progress.md §E.2) | M4 flips it: the final-tree pairs recorded in progress.md §E.2 with commands, exit codes, lists, T and L, the `comm -13` and `comm -3` outputs, the env lines, lease lines |
| AC-THE-004 | REQ-THE-005 | **regression-guard at plan time; becomes release-blocking when its four-element cell is recorded at c2r** | Given the coverage tests committed in the guard commit (cli `TestFactoryEnvAxesCovered`, hook `TestLaneEnvAxesCovered`); When each is run with `-count=1 -v -run '^<name>$'` on the c2 tree, on the final tree, under mutations P1, P3, P4 and P5 (plan.md M4), and on the synthetic inputs of plan.md D3; Then it is RED on the c2 tree naming each uncovered axis and the production file that references it (expected: the six cli axes; for hook the thirteen axes **and** the empty-scrub-set liveness message, because plan.md D5 declares the hook scrub set empty at c2 and liveness assertions use `t.Errorf` — spec.md §E input (3); the cell records which message(s) the observed red carries), GREEN on the final tree with `--- PASS` and a liveness line showing referenced ≥ 1, scrub-set size ≥ 1 and family size ≥ the floor 17, RED naming the removed axis (P1), RED on a padded exemption row (P3) and on each synthetic input (uncovered axis, empty-reason exemption, exemption citing an absent file or a file that does not mention the axis, empty reference set, family below the floor), RED on an empty reference scan (P4), and RED naming the missing file when the sibling package's guard file is absent (P5) | **Not re-executable on `a5a63a0bc`**: the guard does not exist (E-5b shows no guard file); a selector here would print `[no tests to run]` and exit 0, a vacuous green. Cell completed by the c2r record: the exact command, its verbatim stdout, its exit code as its own field, and the c2 SHA together with the `git rev-parse --short HEAD` read taken as the immediately preceding step of the run (equal to it) and the `git status --short` read (expected empty) beside it, per package, recorded in progress.md §E.2 and witnessed by §D.8 step 7 | Green path M2 (cli green) → M3 (hook green); mutation probes recorded at M4, the AC's Exit milestone |
| AC-THE-005 | REQ-THE-006 | **release-blocking** | Given the card branch; When the procedure of §D.8 is run (it names c1, c2 and c2r by exact subject or file set, checks each commit's file set, tests the chain with `git merge-base --is-ancestor` plus `git rev-list --count`, and enumerates **every** commit that touches `internal/`); Then c1, c2 and c2r are each found exactly once with the prescribed shape, each is a strict ancestor of the next in that order, and every `internal/` commit other than c2 is a strict descendant of c2r; the c2r record carries each of the four guard-test red names and the c2 SHA (§D.8 step 7); the SHAs are recorded in progress.md §E.2 | **E-5** (tree `a5a63a0bc`, exit 0, empty stdout for E-5a, E-5a2, E-5b and E-5c) and **E-5d** (exit 1, empty stdout) — red because none of the three commits exists, so the exactly-once checks of §D.8 and the content witness of step 7 cannot hold | M1 creates c1, c2, c2r in that order and evaluates steps 0-5 and 7; later commits descend from them and M4 re-evaluates the step 6 enumeration |
| AC-THE-006 | REQ-THE-004, REQ-THE-008 | **regression-guard** | Given the branch diff vs `<BASE>`, evaluated at the M4 tip and again at the sync tip; When `git diff --name-only <BASE>..HEAD` and `git diff -U0 <BASE>..HEAD -- internal/cli internal/hook` are read; Then every path is a `*_test.go` file or under `.moai/specs/SPEC-TEST-ENV-HERMETIC-001/` (the sync commit adds SPEC-directory paths only); no `func Test…` line present at `<BASE>` is removed; no `t.Skip` / `t.Skipf` / `t.SkipNow` is added; no assertion line is removed and no early bare `return` / `testing.Short()` guard is added in a pre-existing test file (the greps of §D.1, last paragraph) | **N/A (preservation)**: green today — on `a5a63a0bc` the diff names only SPEC-directory paths, so the criterion holds vacuously; it guards the change, not the starting tree, and is not recorded as a pass on that basis | M4 (re-read at the sync tip) |
| AC-THE-007 | REQ-THE-007 | **regression-guard** | Given the post-M2/M3 tree; When the E-4a and E-4b commands run (with `-v`); Then exit 0 and the `--- PASS:` lines of E-4a (two tests) and E-4b (the test and its four subtests) are present; and the applied-behaviour test of AC-THE-008, which reads every added axis, is green | **N/A (positive control)**: green today by E-4a / E-4b (exit 0); E-4a does not read `MOAI_FACTORY_ROLE` (disclosure under E-4), so the separation control for the added axes is AC-THE-008; both must stay green once the scrub is applied | M2 / M3 |
| AC-THE-008 | REQ-THE-009, REQ-THE-005 | **regression-guard at plan time; becomes release-blocking when its four-element cell is recorded at c2r** | Given the applied-behaviour tests committed in the guard commit (cli `TestFactoryEnvAxesScrubApplied`, hook `TestLaneEnvAxesScrubApplied`); When each is run with `-count=1 -v -run '^<name>$'` on the c2 tree, on the final tree, and under mutation P2 (the `TestMain` scrub call removed); Then it is RED on the c2 tree naming every referenced axis still present after the child's `TestMain` (expected: six cli axes, thirteen hook axes), GREEN on the final tree with its own `--- PASS:` line and a child `--- PASS:` line visible in the log, RED naming every surviving axis under P2 even though the declared scrub set is full, and RED if the child was pinned (cli) or the witness list was empty | **Not re-executable on `a5a63a0bc`**: the test does not exist (E-5b). Cell completed by the c2r record exactly as AC-THE-004 — exact command, verbatim stdout, exit code as its own field, c2 SHA with the pre-run `HEAD` read and the `git status --short` read (expected empty) beside it, per package, witnessed by §D.8 step 7. The two observed reds E-1 and E-2 are the production-shaped instances of this failure (an axis survives `TestMain`) | Green path M2 (cli green) → M3 (hook green); P2 recorded at M4, the AC's Exit milestone |

### §D.1 Mutant probes (release-blocking ACs)

- **AC-THE-001** — a mutant that deletes the three tests, adds `t.Skip`, or leaves them failing
  in a way the exit code hides would satisfy an exit-0-only reading; the criterion is judged on
  three named `--- PASS:` lines, zero skips, and AC-THE-006's no-skip/no-removal diff. A mutant
  that pins `MOAI_FACTORY_ROLE` inside these three tests only satisfies AC-THE-001 while
  violating REQ-THE-001 for the rest of the package; AC-THE-003 (whole-package, no selector) is the
  paired criterion that catches it.
- **AC-THE-002** — a mutant that rewrites the two tests to expect the empty notice (making the
  red green by changing the assertion) satisfies "exit 0" and violates REQ-THE-002; the assertion-diff
  clause catches it. A mutant that pins only `MOAI_KANBAN_ID` passes these two tests and leaves
  `MOAI_FACTORY_WORKERS` live for the next test that gates on the conjunction; AC-THE-003 catches it.
- **AC-THE-003** — a mutant that runs both arms under the same scrubbed env, runs the lane arm
  with a `-run` selector, or computes the failing set from a truncated stream would make the sets
  equal vacuously; the criterion requires (clause (f)) the lane arm's env to be recorded
  byte-identically at c1 and at the final tree with every modelled axis present, the c1 lane arm to
  hold the five observed reds (a session carrying no lane axis yields an arm with none: that arm is
  INVALID and the AC fails, instead of passing on two empty sets), no `-run` selector in either arm,
  and each arm's terminal top-level count T to equal the independently listed count L. A mutant in which both arms
  die early for the same env-independent reason — a `-timeout` panic, a `goleak.VerifyTestMain`
  failure in the hook binary, or a crash after N tests — reports equal non-zero counts and equal
  truncated failing sets; T ≠ L and the timeout / goroutine-leak marker lines make such an arm
  invalid, and an invalid arm is no measurement. A mutant that hides the difference by skipping the
  differing test is excluded by AC-THE-006 and by the equal-skipped-sets clause. The mutant that
  clauses (a)-(d) alone would accept is a fix that breaks a test X which sets a family axis
  itself and then calls `clearFactoryTestEnv(t)` (the slice now carries six more axes through
  44 call sites in 11 test files): X fails in **both** final arms, so the sets are equal and a
  shared failure could be named env-unrelated. Clause (e) closes it: X is absent from the c1
  failing set of its arm type, so `final − c1` is non-empty and the AC fails; a failure that was
  already red in c1 in the same arm type is the only kind that can be named env-unrelated. The
  same holds under a parent that is already red at c1: the names files carry the full test path, so
  a new failing subtest (`TestParent/sb`) is absent from c1 even though `TestParent` is present.
- **AC-THE-004** — caught: a padded exemption row whose citation is empty, or names a file that
  is absent or does not mention the axis (P3); a mutant that scans the wrong directory and finds
  zero references (liveness input, P4); a scan that silently drops most family constants (the
  family-size floor); a deleted guard file in one package (the sibling package's coverage test,
  P5). **Not closed by this AC**: the harmful variant is an axis X with **no** scrub plus an
  exemption row `{X, "n/a", <an existing *_test.go that references X>}`. It passes the coverage
  test (the row has a reason and a citation that exists and mentions X) and the applied-behaviour
  test (both skip exempt axes), and nearly every axis is referenced by some cli test file, so the
  citation condition is close to vacuous there. That variant is **left to review**, and the
  closure DoD (§D.5) gives review its evidence: the final exemption tables are listed in
  progress.md §E.2 and each surviving row cites the M4 whole-package scrubbed-arm test that went
  red when that axis was stripped.
- **AC-THE-005** — the mutants this procedure must catch: a fix commit authored before the guard
  commit, a `todo_test.go` pin committed before the guard commit (no file list names it; the
  enumeration of every `internal/` commit does), the baseline record squashed into the guard
  commit (the file-set check on c2 sees `progress.md`), the guard squashed with a fix (c2 also
  touches `factory_test.go` or hook `main_test.go`, so its file set is not exactly the two guard
  files), c1 equal to c2 (the strict-ancestor pair is exit 0 with a count of `0`), and the red
  never recorded before the first fix (no c2r).
- **AC-THE-008** — a mutant that fills the declared scrub list but never calls the scrub from
  `TestMain` passes the coverage test and is caught here (P2); a mutant that makes the probe call
  the scrub function itself is caught by reviewing the child path (plan.md D7, AP-11) and by P2
  still being red; a child that ran nothing is caught by the required child `--- PASS:` line and
  the non-empty witness list.

(Regression-guard mutants, for completeness: AC-THE-006 — a mutant adding an unlisted env-gated
skip is caught by the added-`t.Skip` clause. A bare `return` or `if testing.Short() { return }` at
the top of a pre-existing test is neither a skip nor a removed `func Test…` line; it is read from
the modified-file diff saved to a file, `git diff -U0 --diff-filter=M <BASE>..HEAD -- internal/cli internal/hook > <file>`:
`grep -cE '^-[^-].*(\.(Fatalf?|Errorf?|FailNow|Fail)\(|strings\.Contains|assert\.|require\.)' <file>`
must print `0` (a rewrite that re-indents or removes an assertion removes the line), and
`grep -cE '^\+[[:space:]]*(return$|if testing\.Short\(\))' <file>` must print `0` or every hit is
named and justified in progress.md §E.2. Controls (E-9): on `2de0a2cb6~60..2de0a2cb6` the two
greps print `45` and `5`; on `2de0a2cb6..47ae8ecb1` the saved diff is empty and the first prints `0`
with grep exit `1`. Residual, left to review: a hollowing built from added lines alone — a `defer`
that swallows the failure, an early `panic`-and-recover — is not visible to a line grep, and
neither is a hollowing by a weakened condition: `- if got == "" {` / `+ if false {` removes a line
carrying no assertion token, so the removed-assertion grep stays `0` (SF-6).)

### §D.2 Preconditions (adoption discipline)

- **Two-cell adoption.** A release-blocking AC is unadopted until its RED-now cell is observed on the
  stated tree and pinned. AC-THE-001 / AC-THE-002 / AC-THE-005 carry observed cells (E-1, E-2, E-5);
  AC-THE-003 carries a witness-by-existence cell, completed by the c1 record of the pairs the run
  phase writes before any fix commit; AC-THE-004 / AC-THE-008 have their cell completed by the c2r
  record (their cells cannot exist before the guard does). GREEN without the recorded RED
  counterpart is reported as a Gap, never a PASS.
- **RED for the right reason.** E-1/E-2 are red because of the scrub-set gap (stated per cell), not
  because of unrelated files; the guard's red is the uncovered or surviving axis it names; no
  green path runs through "someone fixes the unrelated files".
- **Ordering clause rewritten for git** (spec.md §F; `verification-claim-integrity.md` §2.3): the
  baseline artifact is gitignored, so ordering is witnessed by commit ancestry (AC-THE-005), not by a
  report commit.
- **Evidence pinning.** Every claim names the tree SHA it was measured on; a rebase or an
  absorption of `develop` re-measures.
- **Env-isolated form.** Scrubbed arms run as one compound `unset <VARS> && go test …`; `env -u`
  is refused by the worktree guard (E-6) and is not used.

### §D.3 Whole-package measurement obligation (AC-THE-003 command shapes)

Per package P ∈ {`./internal/cli`, `./internal/hook`}, on the pre-guard tree (baseline record) and
again on the final tree — four arm runs each time — inside a lease (`moai slot acquire --resource
heavy-test --max-duration <cap>` … `moai slot release --resource heavy-test`). The lease cap is
`max(20m, 1.5 x the longest whole-package runtime recorded so far in progress.md §E.2)` — 20m for
the first c1 arm, recomputed for each later lease — and `-timeout` below is the cap minus 2m, so
it is strictly below the cap and the lease cannot lapse at the instant the test timeout fires
(spec.md §G):

- **listed count L** (plain command, run before the arms on the same tree): `go test -list '.*' P`
  redirected to a file, then `grep -cE '^(Test|Example|Fuzz)' <file>`; L = the number of output
  lines beginning `Test`, `Example` or `Fuzz` (`Benchmark` lines excluded; plan-time L: cli 4884,
  hook 1322, E-8). L on the final tree is at least L on the pre-guard tree plus the two guard
  tests of that package.
- **lane arm**: `unset <every family variable absent from the measuring session's env> && env <the measuring session's family env, verbatim from the pre-flight read> go test P -count=1 -timeout <cap - 2m> -json > <arm file>`.
  The lane arm intentionally models the measuring session (nine axes at plan time: `MOAI_FACTORY_ROLE=lane`, `MOAI_FACTORY_WORKER=lane-6`, `MOAI_FACTORY_WORKERS=0`, `MOAI_KANBAN_BACKEND=claude`, `MOAI_KANBAN_ID=tm9i7y`, `MOAI_AUTONOMY_TIER=fully-autonomous`, `MOAI_FACTORY_AUTO_DISPATCH=auto`, `MOAI_KANBAN_SETTINGS_INJECTED=1`, `MOAI_FACTORY_CLEAR_POLICY=` set empty), not every possible lane (spec.md R8).
  The family env is read once, before any arm, by `env | grep -E '^(MOAI_FACTORY_|MOAI_KANBAN|MOAI_AUTONOMY_TIER)'`
  (plan.md §C step 2; it prints the nine lines above in this session, exit `0`, and does not print
  the unrelated `MOAI_PROFILE_LEASE_TOKEN`). The arm's env is recorded as one line,
  `lane-arm env: set NAME=value … ; unset NAME …` (names sorted, a set-empty axis written `NAME=`),
  in progress.md §E.2 beside the arm; the final lane arm's command carries the c1 line verbatim.
  An arm is a lane only when clause (f) of AC-THE-003 holds: a family read that lacks a required
  axis, or a c1 lane arm without the five reds, makes the arm INVALID. The final-tree lane arm also
  records the env its child actually sees — `unset <the family axes> && env <the nine lane axes>
  env > <file>`, the arm command's prefix with `env` in place of the test run — beside the arm,
  formatted the same way, and that line must be identical to the c1 recorded env line (SF-4: the
  replayed c1 line alone is self-recorded; this witness reads the child-visible env).
- **scrubbed arm**: `unset <every family variable listed from internal/config/envkeys.go at pre-flight> && go test P -count=1 -timeout <cap - 2m> -json > <arm file>`.
  An axis of a surviving exemption row (spec.md §E) is not stripped by the binary, so it is left at
  its lane-arm value in this arm — `unset <the other family variables> && env <exempt NAME=value> go
  test …` — and is identical-valued in both arms by construction, so clause (b) holds for it; the
  arm's env is recorded as one line the same way.

Arm files go to a path under `.moai/state/verify/t1356/` (machine-local scratch, never a citation
target); the deciding lines are copied into progress.md §E.2, the committed carrier. Per arm, five
read-only commands on the arm file, each recorded with its output:

1. terminal top-level count T: `grep -cE '"Action":"(pass|fail|skip)".*"Test":"[^/"]+"' <arm file>` (top-level only, so it compares with the listed count L);
2. failing test rows, subtests included: `grep -E '"Action":"fail".*"Test":"([^"\\]|\\.)+"' <arm file>` (a failing subtest is its own row, named by its full path, `TestParent/sb`; the name pattern is escape-aware — `go test -json` writes a quote or backslash inside a name escaped, and the naive `[^"]+` stops at the first `\"`, truncating the name, MF-3);
3. skipped test rows, subtests included: the same escape-aware pattern as command 2, with `"Action":"skip"`;
4. invalid-arm markers: `grep -cE 'panic: test timed out|goleak: |\[(build|setup) failed\]' <arm file>` must print `0` (grep exits 1 when it prints `0`; the printed count is the reading);
5. when the arm's exit code is non-zero: `grep -cE '"Action":"fail".*"Test":"' <arm file>` must print at least `1` (otherwise the failure is a non-test failure and the arm is invalid).

An arm is **valid** only when T equals L, command 4 prints `0`, and command 5 holds; a lane arm is
valid only when clause (f) of AC-THE-003 also holds — command 10 witnesses clause (f) on the **c1
lane arms only** (it runs on the recorded c1 names files), and a final-tree lane arm's lane-ness
rests on its recorded env line replaying the c1 line plus its child-visible env line, both
identical to the c1 line (SF-4). Both arms of a package must be
valid before their sets are compared; the failing sets and the skipped sets must be equal (the
failing sets by full test path, command 9; the skipped sets by command 11, the same form over the
skip rows), and both failing sets empty modulo failures that appear
identically in both arms **and** are present in the c1 failing set of the same arm type, which the
run names as env-unrelated. The unit of every comparison below is the Go test row, a subtest
included, named by its full path. (E-7c shows command 1 on a real arm file; the plan-time L figures are E-8.)

**c1 containment (clause (e)), equality (clause (b)) and the lane control (clause (f)).** For each
package and each arm type (lane, scrubbed), the failing names of the final-tree arm must be a subset
of the failing names of the c1 arm of the same type. Five more plain commands, no pipe, each
recorded with its output:

6. names file: `grep -oE '"Action":"fail","Package":"[^"]+","Test":"([^"\\]|\\.)+"' <arm file> > <names file>` —
   each match is one failing test row by its full path, subtests included (`go test -json` writes
   the keys in the order Time, Action, Package, Test, Elapsed, so the matched text is stable; the
   package-level fail row carries no `Test` key and does not match). The name pattern is
   escape-aware (`([^"\\]|\\.)+`): a quote or backslash inside a subtest name is written escaped
   in the JSON, so the naive `[^"]+` stops at the first `\"` — `TestParent/q"uote` extracts as
   `TestParent/q\` — and two quote-bearing names sharing the text up to the quote collapse into
   one line, hiding a new failure from `comm -13` (MF-3, E-9). `grep` exits 1 and leaves an
   empty file when the arm has no failing test row, which is a valid empty set once the arm is
   valid (commands 1, 4 and 5); a package-level failure with no test-level row (a build failure, a
   `TestMain` failure after the tests) also yields an empty file, and commands 4 and 5 then make the
   arm invalid, never a pass;
7. `LC_ALL=C sort -o <names file> <names file>`;
8. `LC_ALL=C comm -13 <c1 names file of the same arm type> <final names file>` prints the names present only
   in the final file and **must print nothing** (exit 0);
9. clause (b): `LC_ALL=C comm -3 <lane names file> <scrubbed names file>` of the same tree **must print
   nothing** on the final tree (on the c1 tree it prints the five reds, the red state);
10. clause (f), c1 arms only: `grep -cE '"Test":"(<name1>|<name2>|…)"' <sorted names file>`
   with the package's observed red names — cli names `TestTodoClaim_LaneGovernance`, `TestTodoClaimMCP_Mirror`,
   `TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit`; hook names
   `TestStaleRunNoticeLegacyLeaderSpelling`, `TestStaleRunNoticeLegacySessionRecord` — **must print
   `3` / `2` on the c1 lane names file** and `0` on the c1 scrubbed names file (grep exits 1 then). It
   runs on the names file, not the arm file, which carries those names on `run`, `output` and `pass`
   rows too (E-9);
11. skipped-set equality (the clause "the skipped sets are equal"): per arm,
   `grep -oE '"Action":"skip","Package":"[^"]+","Test":"([^"\\]|\\.)+"' <arm file> > <skip names file>` —
   the escape-aware pattern of command 6 over the skip rows — then
   `LC_ALL=C sort -o <skip names file> <skip names file>`, and
   `LC_ALL=C comm -3 <c1 skip names file of the same arm type> <final skip names file>`
   **must print nothing** (exit 0; grep exits 1 and leaves an empty file when the arm has no skipped
   test row, which is a valid empty set once the arm is valid — command 3 reads the same rows on the
   arm file, command 11 makes them comparable, N1).

A name printed by command 8 or 9 is not final on one run: the affected arm is run once more on the
same tree under the same lease (the identical command, itself a valid arm, its output recorded
beside the initial run); every name printed in **either** run is recorded in progress.md §E.2 with
both outputs; a name printed in **both** runs fails (b) or (e); a name printed only in the repeat —
absent from the initial run — also counts as a hit against (b) or (e) and is never load noise; load
noise is the reverse, a name printed in the initial run and absent from the repeat, recorded with
both outputs (spec.md R7). Command 9 compares two arms, so **both** arms of the pair are repeated,
and the repeat comparison for command 9 runs between the two repeat files. The comparison stays by
the full path of command 6, and the repeat is a whole-package arm, not an isolated re-run, so a
repeat never excuses a subtest and never excuses a test that fails only in package context.

The sorted c1 names files and the lane-arm env lines are copied verbatim into progress.md §E.2 in the
baseline record commit, so the comparison does not depend on scratch surviving. Form controls (the
three arms A/B/C, a subtest case, a package-level failure, the lane control, and the clause (e)
mutant under an already-red parent): ledger E-9.

### §D.4 Given-When-Then scenarios

**AC-THE-001 / 002 — narrow flip.** Given the observed-red tests; When each narrow command is run
under its explicit lane env before and after the fix; Then the pre-fix run exits 1 naming the
refusal / the empty notice and the post-fix run exits 0 with the stated `--- PASS:` lines.

**AC-THE-003 — package-wide equality with a floor.** Given the post-M3 tree and the lease; When the
lane arm and the scrubbed arm of a package run unselected; Then each arm is valid (T equals the
listed count L, no timeout / leak marker), the lane arm is a lane (clause (f): its env recorded and
identical at c1 and final, the five reds in the c1 lane arm), the two failing sets (full test paths,
`comm -3` prints nothing) and the two skipped sets are equal, and any shared failure is named
env-unrelated only when it is also in the c1 failing set of the same arm type (`comm -13` of §D.3
prints nothing).

**AC-THE-004 — the coverage guard fires on a new axis.** Given the guard and a scrub set missing one
axis that production references; When the package tests run; Then the run fails and the failure
names that axis and the production file referencing it; restoring the axis turns it green.

**AC-THE-008 — the applied-behaviour guard fires when the scrub is declared but not applied.** Given
a full declared scrub set and a `TestMain` that no longer applies it; When the package's applied
test re-executes the test binary with every family axis set; Then the run fails naming every axis
the child still sees; restoring the `TestMain` call turns it green.

**AC-THE-005 — ordering from the commit graph.** Given the branch history; When the §D.8 procedure
runs; Then every check exits as stated and no two of the named commits coincide.

### §D.5 Closure Gate (Definition of Done)

- [ ] AC-THE-001, 002, 003, 004, 005, 008 PASS with recorded evidence (command, verbatim output, exit
      code, tree SHA); AC-THE-006, 007 evidence recorded and not counted as release gates.
- [ ] AC-THE-004 and AC-THE-008 have their cell **completed by the c2r record**: the c2-tree RED
      records (the exact command, verbatim stdout, the exit code as its own field, the c2 SHA, the
      pre-run `git rev-parse --short HEAD` read equal to it and the `git status --short` read
      (expected empty) taken beside it, per package; for the hook coverage
      test the message(s) the observed red carries) are in progress.md §E.2 from the guard-red
      record commit, and the §D.8 step 7 content witness holds on c2r; an absent record leaves the
      AC unadopted and blocks closure. The sync auditor re-executes one of the four cells at the c2
      tree — materialized read-only with `git archive -o <scratch>/c2.tar <c2>` and `tar -xf
      <scratch>/c2.tar -C <scratch dir>`, never `git worktree add` or a checkout — and records its
      own verbatim stdout and exit code beside the cell (string presence in progress.md is not
      execution evidence; the re-execution closes that gap for the one cell it re-executes, and the
      other three cells stay witnessed by their recorded strings, SF-2).
- [ ] progress.md §E.2 carries: the baseline pairs with L, T and the arm-validity checks, the sorted
      c1 failing-name files, the four lane-arm env lines (c1 and final, per package, byte-identical),
      the final-tree lane arms' child-visible env lines (the §D.3 `env`-in-place-of-the-run form,
      identical to the c1 lines)
      and the clause (f) outputs of §D.3 command 10 (`3` / `2` for the c1 lane arms, `0` for the c1
      scrubbed arms), the discovery pair re-recorded, the guard reds on the c2 tree, the M3
      one-axis arms, the final pairs with the `comm -13` outputs of §D.3 command 8, the `comm -3`
      outputs of command 9 and the `comm -3` outputs of command 11 (empty for each arm type of each
      package), the mutation probes P1-P5
      (each reverted: empty `git diff --stat` afterwards), every lease acquire/release line, and
      the SHAs of c1, c2 and c2r (plus `<BASE>` where re-derived).
- [ ] The **final exemption tables** of both guard files are listed in progress.md §E.2, and each
      surviving row cites the name of the M4 whole-package scrubbed-arm test that went red when that
      axis was stripped (the first scrubbed arm unsets every axis; the arm is then re-run with the
      exempt axis left at its lane-arm value, §D.3); an empty table needs no row. (This is the
      evidence review reads for the one variant AC-THE-004 does not close, §D.1.)
- [ ] The guard tests are listed in both packages: `go test -list '^(TestFactoryEnvAxesCovered|TestFactoryEnvAxesScrubApplied)$' ./internal/cli` prints both names and `go test -list '^(TestLaneEnvAxesCovered|TestLaneEnvAxesScrubApplied)$' ./internal/hook` prints both names (the closure-time stale-guard signal, spec.md §E).
- [ ] `GOOS=windows GOARCH=amd64 go vet ./internal/cli ./internal/hook` exit 0;
      `golangci-lint run` delta vs the pre-flight baseline = 0 new findings.
- [ ] Branch integrated into local `develop` per the git-flow lane protocol; CI green on `origin/develop`
      per OS (the full-suite verdict surface).
- [ ] No non-test Go file in the diff at the M4 tip or the sync tip; no skip, deletion, or blacklist.
- [ ] Card id `t1356` in every commit message.

### §D.6 Forward-Looking Checks (post-merge)

- **One CI observation cycle.** Watch the two guard pairs and the two packages across one `origin/develop`
  cycle on every OS of the matrix; a red from a guard names a real uncovered or surviving axis.
- **Continued firing.** At the next change that adds a `MOAI_FACTORY_*` / `MOAI_KANBAN*` constant to
  `internal/config/envkeys.go`, the guard of any package that references it is the signal; if a new
  axis is added and no guard turns red, treat that as the guard having stopped firing (spec.md §E).
  Deleting both guard files together is the one change no guard detects (spec.md §E residual).

### §D.7 Quality Gate Criteria (TRUST 5)

- **Tested**: RED/GREEN pair per release-blocking AC; a durable guard pair per package; arm validity
  and swept counts reported against an independent listing.
- **Readable**: English comments; the guard's comment states the hazard (ambient lane env, two prior
  recurrences) and cites this SPEC.
- **Unified**: `gofmt`; lint delta 0; env names via `internal/config/envkeys.go` constants only.
- **Secured**: no new input surface; no production change; no secret material in evidence.
- **Trackable**: Conventional commits + card id `t1356`; SPEC ID in scope; `related_specs` wired.

### §D.8 AC-THE-005 procedure (the single statement of the ordering check)

`<BASE>` is `2de0a2cb6` (or the value re-derived and recorded per plan.md B10). Every command is a
single read-only invocation; the SHAs are copied into progress.md §E.2.

0. Positive control: `git rev-list --count <BASE>..HEAD` prints at least `1`.
1. **c1** — `git log --format=%h --reverse <BASE>..HEAD --grep='^test(SPEC-TEST-ENV-HERMETIC-001): baseline record (card t1356)$'` prints exactly one line.
2. **c2r** — the same with `--grep='^test(SPEC-TEST-ENV-HERMETIC-001): guard red record (card t1356)$'` prints exactly one line.
3. **c2** — `git log --format=%h --reverse --diff-filter=A <BASE>..HEAD -- internal/cli/factory_env_axes_test.go internal/hook/lane_env_axes_test.go` prints exactly one line (both guard files are added by one commit).
4. **Shapes** — `git diff-tree --no-commit-id --name-only -r <commit>` for c1 and for c2r prints exactly `.moai/specs/SPEC-TEST-ENV-HERMETIC-001/progress.md`; for c2 it prints exactly the two guard paths and nothing else (so c2 is not also a fix commit and not also the record).
5. **Chain** — for each pair (c1, c2) and (c2, c2r): `git merge-base --is-ancestor <earlier> <later>` exits 0 **and** `git rev-list --count <earlier>..<later>` prints at least `1`. The two tests together are the strict inequality: `--is-ancestor A A` exits 0, but `git rev-list --count A..A` prints `0` (E-5 controls).
6. **Enumeration** — `git rev-list --no-merges --reverse <BASE>..HEAD -- internal` lists every commit under `internal/`; for every listed commit x other than c2: `git merge-base --is-ancestor <c2r> x` exits 0 **and** `git rev-list --count <c2r>..x` prints at least `1`. Because c2r descends from c2, this also places c2 before every other internal commit.
7. **Content witness on c2r** — the c2r record is author-written text, so this step witnesses that
   it carries the four guard-test reds and names c2 (`<c2>` and `<c2r>` from steps 3 and 2;
   `<progress>` = `.moai/specs/SPEC-TEST-ENV-HERMETIC-001/progress.md`):
   - `git grep -c --all-match -e '--- FAIL: TestFactoryEnvAxesCovered ' -e '--- FAIL: TestFactoryEnvAxesScrubApplied ' -e '--- FAIL: TestLaneEnvAxesCovered ' -e '--- FAIL: TestLaneEnvAxesScrubApplied ' <c2r> -- <progress>` prints one line `<c2r>:<progress>:<N>`. `--all-match` prints the file only when **each** of the four patterns matches at least one line, so one printed line is the witness for every name, and an absent name makes the output empty (N counts the matching lines);
   Each name pattern ends in a space (`go test -v` prints `--- FAIL: <name> (<seconds>s)`), so a longer name sharing the prefix does not satisfy it; the c2r record therefore quotes the verbatim `-v` verdict lines.
   - `git grep -c -e '<c2>' <c2r> -- <progress>` prints `<c2r>:<progress>:<N>` with N ≥ 1 (the cell names the c2 SHA, in the abbreviated form step 3 printed).
   Positive and negative controls for the forms are ledger entry E-5d (a present dashed pattern, `--all-match` with two present names and with one absent name, a REQ id in spec.md at a commit); the step is red now because progress.md at `669cf18c9` carries no guard-test name (E-5d). The sync auditor re-executes one of the four recorded cells at the c2 tree and records its own stdout and exit code (§D.5), because the cell is re-executable there by construction; the witness proves that the strings exist, and only that re-execution shows the guards were run.

Evaluation point: M1 evaluates steps 0-5 and 7 (c1, c2 and c2r exist, in shape and chain, and c2r
carries the record); M4 re-evaluates step 6 — the full enumeration — once the fix commits exist, and
the same procedure is run again at the sync tip.

The gitignored baseline is never committed: nothing here uses `git add -f` or a re-added gitignore negation.
