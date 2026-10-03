# Acceptance — SPEC-TEST-ENV-HERMETIC-001

All commands run from the worktree root
(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1356`). Whole-package runs: Bash timeout of at
least 600 s, serial, under a `moai slot` lease (plan.md §C step 5). Document-level tree pin:
**`2de0a2cb6`** — it binds every RED-now cell below that carries no pin of its own. A RED-now
command is re-executable on any descendant tree that does not yet contain plan.md milestones
M1-M4 (the only commit between `2de0a2cb6` and the SPEC's plan-phase commit changes SPEC
artifacts, not Go sources).

## §D.0 Evidence ledger (RED-now carrier)

Each entry carries the four elements of `verification-completeness.md` §2.1: the command (a
single read-only invocation, no pipe / `&&` / `;`), its verbatim stdout, its exit code as its own
field, and the tree SHA. Durations in `go test` output vary run to run; the judged content is the
`--- FAIL` / `--- PASS` lines and the assertion messages. `env NAME=value go test …` is one
invocation; it sets the named axes on top of whatever the session already exports, so each ledger
command names **every** axis it depends on explicitly (a partial env measures nothing, plan.md B4).

### E-1 — internal/cli: three tests red when `MOAI_FACTORY_ROLE=lane` reaches the binary

- tree: `2de0a2cb6`
- command: `env MOAI_FACTORY_ROLE=lane go test ./internal/cli -count=1 -run '^(TestTodoClaim_LaneGovernance|TestTodoClaimMCP_Mirror|TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit)$'`
- exit code: `1`
- stdout (verbatim):

```
--- FAIL: TestTodoClaim_LaneGovernance (0.87s)
    todo_claim_test.go:291: arm 2 claim --lane: moai claim: refused — lane boundary: a lane session cannot mutate the queue (read-only here: bare todo, list, history, show, why, pr, triage); a lane takes its next card through moai factory next
--- FAIL: TestTodoClaimMCP_Mirror (0.21s)
    todo_claim_test.go:416: mcp todo_claim: todo_claim: moai claim: refused — lane boundary: a lane session cannot mutate the queue (read-only here: bare todo, list, history, show, why, pr, triage); a lane takes its next card through moai factory next
--- FAIL: TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit (0.07s)
    todo_test.go:190: moai add: refused — lane boundary: a lane session cannot mutate the queue (read-only here: bare todo, list, history, show, why, pr, triage); a lane takes its next card through moai factory next
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.514s
```

- **why red**: `TestMain` clears the 12 keys of `factoryAmbientEnvKeys` and that set lacks
  `MOAI_FACTORY_ROLE`; the lane-refusal predicate (`internal/cli/factory_card.go:57-71`) reads it,
  so the role marker reaches the tests and the queue mutation is refused. `TestTodoClaim_LaneGovernance`
  pins only the lane-label axis, so the role axis stays live.

### E-2 — internal/hook: two tests red under the lane env

- tree: `2de0a2cb6`
- command: `env MOAI_FACTORY_ROLE=lane MOAI_FACTORY_WORKER=lane-6 MOAI_FACTORY_WORKERS=0 MOAI_KANBAN_BACKEND=claude MOAI_KANBAN_ID=tm9i7y go test ./internal/hook -count=1 -run '^(TestStaleRunNoticeLegacyLeaderSpelling|TestStaleRunNoticeLegacySessionRecord)$'`
- exit code: `1`
- stdout (verbatim):

```
--- FAIL: TestStaleRunNoticeLegacyLeaderSpelling (0.07s)
    stale_run_m1_test.go:42: staleRunNoticeFor = "", want a notice
--- FAIL: TestStaleRunNoticeLegacySessionRecord (0.07s)
    stale_run_m1_test.go:85: staleRunNoticeFor = "", want a notice naming the legacy record role
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.955s
```

- **why red**: the hook `TestMain` scrubs none of the factory axes. With `MOAI_KANBAN_ID` and
  `MOAI_FACTORY_WORKERS` both non-empty, `staleRunNoticeFor` routes through the run-state gate
  (`internal/hook/stale_run_gate.go:259-260`), which answers "" for a run that is not recorded
  active; the tests expect the ungated relaunch notice.

### E-3 — internal/hook one-axis arms (diagnostic, not a RED-now cell)

These arms are compound (`unset … && go test …`, the form the lane protocol prescribes for a
scrubbed arm) and therefore sit outside the single-invocation RED-now form; they explain E-2 and
are re-observed at M3. Tree `2de0a2cb6`; each arm starts from the session's ambient lane env
(`MOAI_FACTORY_ROLE`, `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_WORKERS`, `MOAI_KANBAN_BACKEND`,
`MOAI_KANBAN_ID` all set) and removes only the named axis. Selector:
`-count=1 -v -run '^(TestStaleRunNoticeLegacyLeaderSpelling|TestStaleRunNoticeLegacySessionRecord)$'`.

| arm | command prefix | exit | verdict lines (verbatim) |
|-----|---------------|------|--------------------------|
| unset `MOAI_KANBAN_ID` | `unset MOAI_KANBAN_ID && go test ./internal/hook …` | 0 | `--- PASS: TestStaleRunNoticeLegacyLeaderSpelling (0.00s)` · `--- PASS: TestStaleRunNoticeLegacySessionRecord (0.00s)` · `ok  	github.com/modu-ai/moai-adk/internal/hook	0.751s` |
| unset `MOAI_FACTORY_WORKERS` | `unset MOAI_FACTORY_WORKERS && go test ./internal/hook …` | 0 | same two `--- PASS` lines · `ok  	github.com/modu-ai/moai-adk/internal/hook	0.672s` |
| unset `MOAI_FACTORY_ROLE` + `MOAI_FACTORY_WORKER` + `MOAI_KANBAN_BACKEND` | `unset MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER MOAI_KANBAN_BACKEND && go test ./internal/hook …` | 1 | `stale_run_m1_test.go:42: staleRunNoticeFor = "", want a notice` · `--- FAIL: TestStaleRunNoticeLegacyLeaderSpelling (0.08s)` · `stale_run_m1_test.go:85: staleRunNoticeFor = "", want a notice naming the legacy record role` · `--- FAIL: TestStaleRunNoticeLegacySessionRecord (0.08s)` · `FAIL	github.com/modu-ai/moai-adk/internal/hook	0.884s` |

Reading: each of `MOAI_KANBAN_ID` and `MOAI_FACTORY_WORKERS` alone is sufficient to flip the two
tests to PASS when removed; the other three axes are not. The responsible condition is the
conjunction of the two (matching the source gate). Not isolated: `MOAI_KANBAN` /
`MOAI_KANBAN_LEAD_NAME`, which the tests set themselves.

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

### E-5 — the guard files and their ancestry do not exist yet (basis of AC-THE-005)

- tree `2de0a2cb6`, exit `0`:
  `git log --format=%h --reverse 2de0a2cb6..HEAD -- internal/cli/factory_env_axes_test.go internal/hook/lane_env_axes_test.go`
- stdout (verbatim): *(empty)*

An empty stdout with exit 0 is a complete observation: no commit adds either guard file.

### E-6 — refused-tool disclosure (not a measurement)

Two plan-phase attempts were refused by the worktree-isolation guard, and one class of partial arm
was discarded; none supplies a cited figure (`verification-claim-integrity.md` §3.1):
(1) `env -u <AXIS> go test …` was refused ("cannot be shown not to be git"); the `unset <AXIS> &&
go test …` form of E-3 replaced it. (2) A first group of arms of the form `env <subset> go test …`
inherited the remaining ambient axes (they set, never unset) and measured nothing; discarded.
(3) A multi-SPEC loop with a runtime-computed path in `sed` was refused; plain separate commands
replaced it.

## §D AC Matrix

Classification (`verification-completeness.md` §2.1): **release-blocking** = the RED-now cell is
re-executable on `2de0a2cb6` and carries command, stdout, exit code, and tree SHA. **regression-guard**
= its RED cannot be re-executed on `2de0a2cb6` (the artifact it tests does not exist yet, or it is
green today); it is not recorded as a pass on that basis. 7 ACs (Tier M ceiling 16).

| AC | REQ | Class | Given / When / Then (judgment) | RED-now cell | Green path |
|----|-----|-------|--------------------------------|--------------|-----------|
| AC-THE-001 | REQ-THE-001 | **release-blocking** | Given the cli test binary; When `env MOAI_FACTORY_ROLE=lane go test ./internal/cli -count=1 -v -run '^(TestTodoClaim_LaneGovernance\|TestTodoClaimMCP_Mirror\|TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit)$'` runs; Then exit 0, exactly three `--- PASS:` lines for these names, zero `--- SKIP`, no `[no tests to run]` | **E-1** (tree `2de0a2cb6`, exit 1) — red because `MOAI_FACTORY_ROLE` is outside the cli `TestMain` scrub set | M2 flips it: same command, exit 0, three `--- PASS:` lines |
| AC-THE-002 | REQ-THE-002 | **release-blocking** | Given the hook test binary; When the E-2 command runs with `-v` added; Then exit 0, exactly two `--- PASS:` lines for the two StaleRunNotice tests, zero `--- SKIP`; AND `git diff 2de0a2cb6 -- internal/hook/stale_run_m1_test.go` leaves the assertion lines (`Fatalf` / `Errorf` / `Contains`) of both tests untouched | **E-2** (tree `2de0a2cb6`, exit 1) — red because the hook `TestMain` scrubs no factory axis and `MOAI_KANBAN_ID` ∧ `MOAI_FACTORY_WORKERS` route the notice through the run-state gate (E-3) | M3 flips it: E-2 command with `-v`, exit 0, two `--- PASS:` lines; assertion diff empty |
| AC-THE-003 | REQ-THE-001, REQ-THE-002, REQ-THE-003, REQ-THE-007 | **release-blocking** | Given the post-M3 tree; When the four whole-package runs of §D.3 execute (cli lane arm, cli scrubbed arm, hook lane arm, hook scrubbed arm, no `-run` selector, `-count=1`, under the lease); Then per package the lane-arm and scrubbed-arm failing-test sets are **equal**, the swept test counts are equal, both sets are empty modulo failures named as env-unrelated, and neither set names a guard test or any of the five observed reds | Witness by existence: **E-1 + E-2** (tree `2de0a2cb6`) show the lane-arm set contains at least those 5 tests; the scrubbed-arm greens are E-3 (hook, unset arms) and, for cli, Arm B of the local-only `.moai/reports/t1356/baseline.md` (`unset MOAI_FACTORY_ROLE`, not committed, not re-measured here), so the two sets differ now — red because the five observed reds are lane-only. The full pairs are recorded at M1 c1 in progress.md §E.2 (pre-guard tree) | M4 flips it: the final-tree pairs recorded in progress.md §E.2 with commands, exit codes, lists, counts, lease lines |
| AC-THE-004 | REQ-THE-005 | **regression-guard** | Given the guard committed at M1; When it runs on its own commit's tree, then after M2/M3, then with one axis removed from a scrub set, then on the synthetic inputs of plan.md D3; Then it is RED on the M1 tree naming the uncovered axes of both packages, GREEN on the final tree, RED naming the removed axis under the removal probe, and RED on each synthetic input (uncovered axis, empty-reason exemption, empty reference set) | **Not re-executable on `2de0a2cb6`**: the guard does not exist (E-5 shows no guard file); running a guard selector here would print `[no tests to run]` and exit 0, a vacuous green. RED is observed at M1 on the c2 tree and recorded with its own SHA | M1 (red) → M2 (cli green) → M3 (hook green); removal probe recorded at M4 |
| AC-THE-005 | REQ-THE-006 | **release-blocking** | Given the card branch; When `git log --format='%h %s' --reverse 2de0a2cb6..HEAD -- <guard files + scrub-set files>` is read and ancestry is tested with `git merge-base --is-ancestor`; Then the baseline-record commit (c1) and the guard commit (c2) are distinct commits, c1 is a strict ancestor of c2, and c2 is a strict ancestor of every commit that touches `internal/cli/factory_test.go` or `internal/hook/main_test.go` or any M4 pin; the SHAs are recorded in progress.md §E.2 | **E-5** (tree `2de0a2cb6`, exit 0, empty stdout) — red because neither guard file exists, so no ordering can hold | M1 creates c1 then c2; M2/M3/M4 commits follow |
| AC-THE-006 | REQ-THE-004, REQ-THE-008 | **regression-guard** | Given the branch diff vs `2de0a2cb6`; When `git diff --name-only 2de0a2cb6..HEAD` and `git diff -U0 2de0a2cb6..HEAD -- internal/cli internal/hook` are read; Then every path is a `*_test.go` file or under `.moai/specs/SPEC-TEST-ENV-HERMETIC-001/`; no `func Test…` line is removed; no `t.Skip` / `t.Skipf` / `t.SkipNow` is added | **N/A (preservation)**: green today — at `2de0a2cb6` the diff is empty, so the criterion holds vacuously; it guards the change, not the starting tree, and is not recorded as a pass on that basis | M4 |
| AC-THE-007 | REQ-THE-007 | **regression-guard** | Given the post-M2/M3 tree; When the E-4a and E-4b commands run (with `-v`); Then exit 0 and the `--- PASS:` lines of E-4a (two tests) and E-4b (the test and its four subtests) are present | **N/A (positive control)**: green today by E-4a / E-4b (tree `2de0a2cb6`, exit 0); it must stay green after the scrub | M2 / M3 |

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
  equal vacuously; the criterion requires the lane arm's explicit env to be echoed in the
  evidence (names of the axes set), no `-run` selector in either arm, and the swept test counts of
  the two arms to be equal and non-zero. A mutant that hides the difference by skipping the
  differing test is excluded by AC-THE-006.
- **AC-THE-005** — a mutant that squashes the guard and the fix into one commit, or authors the
  guard after the fix, satisfies a "guard exists" reading and violates REQ-THE-006; the criterion
  requires c1 ≠ c2, c1 a strict ancestor of c2, and c2 a strict ancestor of every fix commit, tested
  with `git merge-base --is-ancestor` (a commit is not its own strict ancestor).

(Regression-guard mutants, for completeness: AC-THE-004 — a guard that scans the wrong directory
finds zero references and fails its positive control; AC-THE-006 — a mutant adding an
unlisted env-gated skip is caught by the added-`t.Skip` clause.)

### §D.2 Preconditions (adoption discipline)

- **Two-cell adoption.** A release-blocking AC is unadopted until its RED-now cell is observed on the
  stated tree and pinned. AC-THE-001 / AC-THE-002 / AC-THE-005 carry observed cells (E-1, E-2, E-5);
  AC-THE-003 carries a witness-by-existence cell plus the M1 c1 pairs the run phase records before any
  fix commit. GREEN without the recorded RED counterpart is reported as a Gap, never a PASS.
- **RED for the right reason.** E-1/E-2 are red because of the scrub-set gap (stated per cell), not
  because of unrelated files; no green path runs through "someone fixes the unrelated files".
- **Ordering clause rewritten for git** (spec.md §F; `verification-claim-integrity.md` §2.3): the
  baseline artifact is gitignored, so ordering is witnessed by commit ancestry (AC-THE-005), not by a
  report commit.
- **Evidence pinning.** Every claim names the tree SHA it was measured on; a rebase re-measures.
- **Env-isolated form.** Scrubbed arms run as one compound `unset <VARS> && go test …`; `env -u`
  is refused by the worktree guard (E-6) and is not used.

### §D.3 Whole-package measurement obligation (AC-THE-003 command shapes)

Per package P ∈ {`./internal/cli`, `./internal/hook`}, four runs total, each inside a lease
(`moai slot acquire --resource heavy-test --max-duration 20m` … `moai slot release --resource
heavy-test`):

- **lane arm**: `env MOAI_FACTORY_ROLE=lane MOAI_FACTORY_WORKER=lane-6 MOAI_FACTORY_WORKERS=0 MOAI_KANBAN_BACKEND=claude MOAI_KANBAN_ID=tm9i7y MOAI_AUTONOMY_TIER=fully-autonomous MOAI_FACTORY_AUTO_DISPATCH=auto go test P -count=1 -timeout 20m -json`
- **scrubbed arm**: `unset <every family variable listed from internal/config/envkeys.go at pre-flight> && go test P -count=1 -timeout 20m -json`

Output goes to a file under `.moai/state/verify/t1356/` (machine-local scratch; the deciding
lines — the failing-test lists, the swept counts, the exit codes, the lease lines — are copied into
progress.md §E.2, which is the committed carrier). The failing set is the set of `"Test"` names on
`"Action":"fail"` rows; the swept count is the number of distinct `"Test"` names on `"Action":"pass"`
or `"fail"` rows. Expected: both sets empty modulo failures that appear identically in both arms,
which the run names as env-unrelated.

## §D.4 Given-When-Then scenarios

**AC-THE-001 / 002 — narrow flip.** Given the observed-red tests; When each narrow command is run
under its explicit lane env before and after the fix; Then the pre-fix run exits 1 naming the
refusal / the empty notice and the post-fix run exits 0 with the stated `--- PASS:` lines.

**AC-THE-003 — package-wide equality.** Given the post-M3 tree and the lease; When the lane arm and
the scrubbed arm of a package run unselected; Then the two failing sets and the two swept counts
are equal, and any shared failure is named env-unrelated.

**AC-THE-004 — guard fires on a new axis.** Given the guard and a scrub set missing one axis that
production references; When the package tests run; Then the run fails and the failure names that axis
and the production file referencing it; restoring the axis turns it green.

**AC-THE-005 — ordering from the commit graph.** Given the branch history; When ancestry is tested
for c1 → c2 → each fix commit; Then every test exits 0 and no two of the named commits coincide.

## §D.5 Closure Gate (Definition of Done)

- [ ] AC-THE-001, 002, 003, 005 PASS with recorded evidence (command, verbatim output, exit code, tree
      SHA); AC-THE-004, 006, 007 evidence recorded and not counted as release gates.
- [ ] progress.md §E.2 carries: the M1 c1 baseline pairs, the guard's red on the c2 tree, the M3
      one-axis arms, the M4 final pairs, every lease acquire/release line, the commit SHAs for AC-THE-005.
- [ ] `GOOS=windows GOARCH=amd64 go vet ./internal/cli ./internal/hook` exit 0;
      `golangci-lint run` delta vs the pre-flight baseline = 0 new findings.
- [ ] Branch integrated into local `develop` per the git-flow lane protocol; CI green on `origin/develop`
      per OS (the full-suite verdict surface).
- [ ] No non-test Go file in the diff; no skip, deletion, or blacklist.
- [ ] Card id `t1356` in every commit message.

## §D.6 Forward-Looking Checks (post-merge)

- **One CI observation cycle.** Watch the two guards and the two packages across one `origin/develop`
  cycle on every OS of the matrix; a red from a guard names a real uncovered axis.
- **Continued firing.** At the next change that adds a `MOAI_FACTORY_*` / `MOAI_KANBAN*` constant to
  `internal/config/envkeys.go`, the guard of any package that references it is the signal; if a new
  axis is added and no guard turns red, treat that as the guard having stopped firing (spec.md §E).

## §D.7 Quality Gate Criteria (TRUST 5)

- **Tested**: RED/GREEN pair per release-blocking AC; a durable guard per package; swept counts reported.
- **Readable**: English comments; the guard's comment states the hazard (ambient lane env, two prior
  recurrences) and cites this SPEC.
- **Unified**: `gofmt`; lint delta 0; env names via `internal/config/envkeys.go` constants only.
- **Secured**: no new input surface; no production change; no secret material in evidence.
- **Trackable**: Conventional commits + card id `t1356`; SPEC ID in scope; `related_specs` wired.
