auditor-model: claude-opus-5-5

# Sync audit — card t2070 (worktree migration speed and announcements)

verdict: FAIL
audited_sha: 30392ed445f084485a55d277fbf5fd76becd4aff

Overall: FAIL. The FAIL is caused by the unmet required codex gate, not by the code. The independent review found no BLOCKING code defect, and on its own it scores PASS-WITH-DEBT, 0.89 (Functionality 92, Security 95, Craft 80, Consistency 86).

Why FAIL: `.moai/config/sections/workflow.yaml` sets `audit.gates.codex: required` (lines 36-39), but the leader's brief for this audit forbids calling any external review backend (codex_audit, audit_multi, glm, claude_audit). Per the auditor contract, an unmet required gate cannot produce a PASS. The receipt guard enforced this at subagent stop: "AUDIT_RECEIPT_VIOLATION: sync-auditor reported PASS-WITH-DEBT but the audit could not be corroborated — no receipt cited." I did not override the leader's prohibition.

gate_unmet: codex (required; not called by instruction).

To clear it, the leader authorizes a `mcp__moai__codex_audit` run (target baseBranch, project_root = this worktree), and the audit is repeated citing the returned receipt. If the codex leg returns no blocking finding, the substantive verdict below stands as PASS-WITH-DEBT with debts F1-F10.

Commit under audit is 30392ed445 (HEAD of WT-moai-update-mo); its code equals 2ac39de91c, and the card's code change is `git show 2ac39de91c`.
I found no BLOCKING defect inside the diff. There is one SHOULD-FIX that is likely to turn the CI lint job red (F1). It must be fixed, or disproved with golangci-lint, before the card lands.

Audit plan / cross-model: not run. The leader's brief forbids external review backends (codex, glm, audit_multi, claude_audit), and `moai verify audit-plan` was not invoked. No receipts were issued. See Gaps.

## 1. Claim

| # | Claim | Result |
|---|---|---|
| C1 | `processCWDIndex.usesWorktree` answers exactly like `processUsesWorktree` for every input | Holds. I read the code and tried to build a counter-example; none survived (see Evidence E3). The only difference is when the stat calls run (once at index build instead of per pair). |
| C2 | The eight concurrent manifest reads settle in the old order with the same verdicts (unborn-HEAD exception, first failure wins, timeout) and the same digest | Holds (E4, tests pass incl. `-race`) |
| C3 | The read-ahead cannot deadlock, cannot reorder journal or move-list writes, and keeps its bound; a stale read-ahead digest still ends in `reversed` | Holds under the invariant that `before()` is called exactly once per eligible tree, in order. It is the first statement of `migrateOneTreeWith` (migrate.go:1121) (E5) |
| C4 | The destination root is resolved once, with the same destinations and the same error behaviour; the dry run writes nothing | Destinations are the same (`Join(Join(R,""),n) == Join(R,n)`, and the resolver's error does not depend on the name). One behaviour change: the first error is memoized (F4). The dry-run byte-identity tests pass (E6) |
| C5 | The progress lines: dry run and `--json` stay silent, non-TTY output is sane, the throttle works | Holds (E7) |
| C6 | No safety check was removed | Holds. The manifest comparison before and after the move, the journal `pending` entry before the move, the reverse order, and the skip on a probe error are all unchanged (E8) |
| C7 | progress.md numbers are honestly attributed and its Gaps section is complete | Mostly holds. There are small arithmetic and wording drifts and one mis-guessed cause (F2, F9, F10) |

## 2. Evidence (commands run in this audit, verbatim output)

E1 — tree identity and scope
```
$ git rev-parse HEAD
30392ed445f084485a55d277fbf5fd76becd4aff
$ git diff 2ac39de91c 30392ed445 --stat
 .moai/reports/t2070/probe-end-to-end.go.txt   |  62 ++++++++++
 .moai/reports/t2070/probe-stage-timing.go.txt | 156 ++++++++++++++++++++++++++
 .moai/reports/t2070/progress.md               | 112 ++++++++++++++++++
 3 files changed, 330 insertions(+)
```

E2 — format / vet
```
$ gofmt -l internal/cli/worktree internal/config
(no output)
$ go vet ./internal/cli/worktree ./internal/config
(no output)
```

E3 — process-cwd index equivalence (code read: migrate.go:330-396 against migrate.go:400-407, classify.go:234-276, migrate.go:2476-2478)
- `canonicalTreePath` and `evalOrClean` are the same function (EvalSymlinks, else Clean). So `e.canon == treeCanon` is the old second disjunct, and it also subsumes the old `rp == rr → false` early exit of `underPath`.
- The empty tree or empty dir case: old `isUnderWorktreePrefix` returns false when the prefix is ""; `underPath` returns false when the path is "". New: `tree == "" || e.resolved == ""` → continue. `evalOrClean` never returns "" for a non-empty input, so the two match.
- A tree that does not exist, or whose stat fails: the old code tests HasPrefix before the stat, and so does the new. On a stat failure the old code returns false for that dir; the new code continues to the next dir. The results match.
- The ancestor walk: the recorded loop is byte-for-byte the old loop. It stops at the filesystem root or at a not-exist error, and skips any other stat error. Its termination never depends on the tree, so the recorded set is exactly what the old walk would visit before a match. Respelled paths, symlinked dirs, missing dirs and the `""`/`"/"` entries are covered by `TestProcessCWDIndex_AnswersLikeTheDirectComparison`, which uses the old function as its oracle.
- Counter-example search: I found none. The only residual is timing: ancestors are stat'ed once at index build instead of at each check (Residual-risk R3).

E4 — manifest capture (migrate.go:2262-2274, 2310-2370; reference oracle in the test at migrate_speed_t2070_test.go:80-106)
- `settle` is the old `run` second half. Once `g.err` is set, every later read settles to "", which is exactly what the old early `return ""` produced. The unborn-HEAD probe still runs lazily inside `settle`, in call order.
- One difference: after a failure, the old code did not execute the remaining reads; the new code has already executed them. They are read-only. The opportunistic index refresh of `git status`/`git diff` is non-dying on lock contention, so this has no correctness effect.

E5 — read-ahead (migrate.go:946-956, 998-1003, 1062-1115)
- The dispatcher acquires a slot before it launches capture i. `get(i)` waits on `done[i]` and then frees one slot. By induction, when `get(i)` is called, captures up to `i+ahead-1` have been launched, so there is no deadlock. This holds as long as `get` is called once per index, in order. `index = treeIndex-1` increases monotonically. The prefetch source filter (`!interrupted && MoaiMade && skip==""`) is the same predicate that reaches `treeIndex++`. `before()` is the first statement of `migrateOneTreeWith`, with no early return before it.
- `stop()` closes `quit` and the dispatcher exits. Captures already started finish under their own git bound (2 min each, `DefaultWorktreeGitTimeout`). Nothing waits on them, and none outlives a normal run, because every future is taken.
- All journal and move-list writes stay inside `migrateOneTreeWith`, sequential and in plan order. The read-ahead only reads.

E6 — targeted tests (one plain command each)
```
$ go test ./internal/cli/worktree -run "TestProcessCWDIndex_|TestTreeProgress_|TestCaptureTreeManifest_|TestMigrationEngine_|TestPlanMigration_|TestClassify_AgentNamedTreeFollowsItsRoot" -count=1 -v
--- PASS: TestPlanMigration_SkipsEveryTreeWhenLsofTimesOut (3.21s)
--- PASS: TestPlanMigration_SkipsATreeAnActiveProcessIsInside (6.39s)
--- PASS: TestCaptureTreeManifest_UnbornHEAD_IsNotAnError (0.34s)
--- PASS: TestMigrationEngine_WorktreeListTimeoutIsAnError (6.73s)
--- PASS: TestPlanMigration_ResolvesTheDestinationRootOnce (3.77s)
--- PASS: TestPlanMigration_DryRunResolvesTheRootOnceAndWritesNothing (4.27s)
--- PASS: TestCaptureTreeManifest_EqualsTheSequentialCapture (5.34s)
--- PASS: TestCaptureTreeManifest_FailureReadsTheSameAsTheSequentialCapture (0.38s)
--- PASS: TestCaptureTreeManifest_TimeoutStillFailsTheCapture (4.02s)
--- PASS: TestCaptureTreeManifest_ReadsRunTogether (3.68s)
--- PASS: TestClassify_AgentNamedTreeFollowsItsRoot (0.01s)
--- PASS: TestProcessCWDIndex_AnswersLikeTheDirectComparison (0.14s)
--- PASS: TestProcessCWDIndex_ManyTreesAgainstManyProcessesIsQuick (3.56s)
--- PASS: TestTreeProgress_PlanAnnouncesTheCountAndThatItCanTakeLong (0.00s)
--- PASS: TestTreeProgress_PlanAnnouncesNothingForNoTrees (0.00s)
--- PASS: TestTreeProgress_ScanCountsTreesAndIsThrottled (0.00s)
--- PASS: TestTreeProgress_MoveShowsElapsedAndRemainingTime (0.00s)
--- PASS: TestMigrationEngine_AnnouncesBeforeThePlanRuns (7.00s)
--- PASS: TestMigrationEngine_ReadsManifestsAheadButBoundedAndMovesInPlanOrder (10.25s)
--- PASS: TestMigrationEngine_AManifestReadAheadStillMustMatchAfterTheMove (8.73s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	68.857s

$ go test -race ./internal/cli/worktree -run "TestCaptureTreeManifest_|TestMigrationEngine_ReadsManifestsAhead|TestMigrationEngine_AManifestReadAhead|TestProcessCWDIndex_AnswersLike" -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	27.246s

$ go test ./internal/cli/worktree -run "TestMigrateDryRun_IsByteIdentical|TestMigrate_DryRunWritesNothing|TestApply_|TestRollback" -count=1
--- FAIL: TestApply_PendingIntentButNeverMoved_IsSettledThenMoved (3.18s)
    migrate_hardening_e_t2053_test.go:118: move-list results = pending,reversed,moved, want pending,reversed,pending,moved
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	112.328s
```
(Only that one test failed. The dry-run byte-identity and writes-nothing tests passed in the same run. Cause: F2.)

E7 — progress lines (migrate.go:145-151, 928-932, 1329-1335; migrate_progress.go; tui/progress_line.go:85-167)
- The migrate verb passes `progress` only for `apply && !asJSON`, and update consolidation only for `!dryRun`. The engine forwards it to the plan only when `apply`. So the dry run and `--json` stay silent.
- On non-TTY, each `Update` is one plain `"  ○ msg\n"` line with no escapes. `planEnd` sets `scanLine = nil` after `Done`, so the `Done`-twice panic in `tui` is unreachable.
- Throttle: the first and the last tree are always shown, and the rest at most once per second (pinned by `TestTreeProgress_ScanCountsTreesAndIsThrottled`).

E8 — safety checks intact (migrate.go:1120-1198)
- The manifest is checked first (fail-safe skip on error). The order is then unchanged: `EnsureInstanceToken`, then `journal.pending`, then the intent line, then the move, the re-key, the rebind, and the after-move `captureManifestChecked` compare → `reverseAfterFailure`.
- The plan's probe-error skip and the active-process skip are still in place (migrate.go:591-603).

E9 — unused function (F1)
```
$ git grep -nw "migrateOneTree" 98ac41c7ca -- internal
98ac41c7ca:internal/cli/worktree/migrate.go:847:		out := migrateOneTree(root, p, mlPath, trigger, journal)
... (definition and comments)
$ grep -rnw "migrateOneTree" internal
internal/cli/worktree/migrate.go:1035:// treeOutcome is migrateOneTree's verdict for one tree.
internal/cli/worktree/migrate.go:1044:// migrateOneTree moves one tree through the journaled step order —
internal/cli/worktree/migrate.go:1051:func migrateOneTree(root string, p migratePlan, mlPath, trigger string, journal *migrationJournal) treeOutcome {
internal/cli/worktree/migrate.go:1117:// migrateOneTreeWith is migrateOneTree with the before-the-move manifest handed
```
`.golangci.yml` enables `unused` (line 30). The diff removed the only call site (the engine now calls `migrateOneTreeWith`).

## 3. Findings (structured defect list)

- **F1 [SHOULD-FIX] [blocking for landing if confirmed] [in-diff]** — internal/cli/worktree/migrate.go:1051. `migrateOneTree` now has zero references in production or tests (E9). With `unused` enabled in `.golangci.yml`, CI's lint job is expected to report it. This is not observed: golangci-lint was outside this audit's allowed commands, so it stays a high-confidence hypothesis. Required fix: delete `migrateOneTree`, since `manifestCapture` stays in use through the prefetcher; or run `golangci-lint run ./internal/cli/worktree` and record that it is clean. Confidence high.
- **F2 [SHOULD-FIX] [optional for this card] [outside — movelist.go, pre-existing since t1640 0e35ff49f5, with reproduction]** — internal/cli/worktree/movelist.go:494-523. `revLines` is documented "last first" but returns lines in file order. So the dedupe in `appendMoveListLine` compares a new line with the FIRST line of the move list, not the latest. The t2053 fixture writes a `pending` line that is byte-identical to the new run's `pending` line when both fall in the same second: the ts is second-resolution, and trigger, paths, ident and reason are equal. The new intent line is then silently dropped, which gives `pending,reversed,moved` (reproduced in E6). This explains why t2070's speedup raises the failure rate that progress.md measured (4/10 → 7/10): the faster apply lands more often in the same second as the fixture write. In production the effect is that a pre-move intent line can be dropped when it duplicates the file's first line, which is rare. Not mutation-confirmed: the diagnosis comes from reading the code. progress.md's guessed cause ("identical to the latest one") is close but wrong on which line. Required fix (separate card, ideally landing before or with t2070 so the package's CI is not flaky-red): make `revLines` actually reverse, or compare against the last line; then re-run the t2053 test `-count=20`. Security relevance: the intent line is a crash-recovery record. Severity medium, so I am not escalating it as High.
- **F3 [MINOR] [optional] [in-diff]** — internal/config/defaults.go:279-287. The comment says the bound "caps the git processes at this many captures of eight reads each". The after-move `captureManifestChecked` runs alongside the read-ahead, so the real peak is 4 captures (about 32 git processes, plus the unborn-HEAD probes). This I/O fan-out on large real trees was not measured. Required fix: correct the comment.
- **F4 [MINOR] [optional] [in-diff]** — internal/cli/worktree/migrate.go:496-507. A destination-root error is now memoized: one transient resolver failure skips every remaining tree, where it previously skipped one. This is the fail-safe direction (skip, never move), but it is a behaviour change that progress.md does not name. Required fix: a one-line note in the code comment, or in progress.md.
- **F5 [MINOR] [optional] [in-diff test]** — migrate_speed_t2070_test.go:61. The name `TestPlanMigration_DryRunResolvesTheRootOnceAndWritesNothing` claims "WritesNothing" but asserts only the call count. The property rests on the existing `TestMigrate_DryRunWritesNothing` and `TestMigrateDryRun_IsByteIdentical`, which passed (E6). Required fix: rename it, or add the check.
- **F6 [MINOR] [optional] [in-diff test]** — migrate_speed_t2070_test.go:191 (400 ms) and :322 (3 s). Wall-clock thresholds on a loaded machine are a flake risk. The 400 ms test also asserts peak ≥ 2, which is the load-independent signal. Required fix: none needed now; widen the thresholds if CI flakes.
- **F7 [MINOR] [optional] [in-diff]** — internal/cli/worktree/migrate.go:400. `processUsesWorktree` is now production-dead and kept only as the test oracle, with no comment saying so. (A duplicate orphan doc comment at migrate.go:2480 is pre-existing and outside the diff.) Required fix: add an "oracle for TestProcessCWDIndex_*" note.
- **F8 [MINOR] [optional] [in-diff]** — migrate_progress.go:74. The message reads "checking N registered tree(s)", but N counts only the trees under the two legacy folders, not all registered trees (the operator's 648 registered vs. the in-scope count). Required fix: "N tree(s) under the legacy worktree folders", or similar.
- **F9 [MINOR] [optional] [evidence doc]** — progress.md:38. "Manifest (before + after) is 55-70% of one tree's move": the table's own numbers give 56% / 72% / 72%. Required fix: change it to 56-72%.
- **F10 [MINOR] [optional] [evidence doc]** — progress.md §Verification and Gaps. golangci-lint, which CI runs and which F1 depends on, is not named as unrun. The `/verify` gap covers it only implicitly. Required fix: name it.

Item (3) of the card (agent-* in the move count): the answer is correct per the classifier and pinned by `TestClassify_AgentNamedTreeFollowsItsRoot`. The real-repo placement is honestly listed as unobserved.

## 4. Baseline-attribution

- Tree: this worktree at HEAD 30392ed445f084485a55d277fbf5fd76becd4aff (code = 2ac39de91c); base for the scope comparison is 98ac41c7ca (the parent merge).
- Toolchain: the installed `go` and `gofmt` invoked by shell path. All test runs compile from this tree, so the judging build is made from the tree under measurement. No moai binary measurement is cited.
- progress.md's timing tables were NOT re-measured here. They stand on the lane's probe runs (sources committed as `.moai/reports/t2070/probe-*.go.txt`). I checked their arithmetic only: 13.88 s / 69 = 201 ms; 7.29 s / 129 = 56 ms; 9.82 → 0.56 s ≈ 17.5x; 51.85 → 15.26 / 13.25 s ≈ 3.4-3.9x; 42.03 s / 104 = 404 ms. All are consistent, except the share in F9. The "648 trees → two minutes" is an extrapolation (648 × 201 ms ≈ 130 s) from the fixture's per-tree cost; it is labelled as a cause, not a measurement on the real project.

## 5. Gaps (not observed in this audit)

- golangci-lint was not run (outside the allowed commands), so F1 is unconfirmed.
- No cross-model or second-opinion backend was called, and `moai verify audit-plan` was not run. The leader's brief forbade both. No `audit_receipt` exists.
- The mutants described in progress.md (index ancestor check disabled; read-ahead bound 1000) were not re-executed, because editing sources was not permitted. I judged them plausible and correctly targeted: both would be caught by the named tests as written.
- Coverage of the new code was not measured. A package-wide `-cover` was not permitted, and a subset cover figure would be meaningless. By reading the code, every new function is exercised by at least one passing test.
- The probe timings and the before/after end-to-end numbers were not re-run.
- Windows was not executed (vet, compile and darwin tests only). TTY rendering was not observed.
- The F2 diagnosis was reached by reading the code and is not mutation-confirmed.

## 6. Residual-risk

- R1: On a real project with large trees, up to about 32 concurrent git reads (3 read-ahead captures plus the after-move capture, 8 reads each) could raise I/O contention. A read past the 2-minute bound fails safe as "skipped, manifest unavailable", not as a wrong move. Not measured.
- R2: A tree edited between its read-ahead capture and its move now ends `reversed` (a visible failure). The old order would have moved it. This is safe but noisier, and it is disclosed in progress.md.
- R3: The process index stats ancestors once at build time. A directory created or removed during the plan is judged on the earlier snapshot. The lsof list itself was already a single snapshot, so the exposure is unchanged in kind.
- R4: If CI runs the full package, the pre-existing F2 flake now fails more often (lane-measured 7/10 on this machine). Without a fix to F2, the card's PR may go red for reasons unrelated to its diff.

## Iteration history

- Iteration 1a: the code review alone came out PASS-WITH-DEBT at 30392ed445. The receipt guard refused it, because the required codex gate had no receipt.
- Iteration 1b (this file): FAIL at 30392ed445, caused only by the unmet required codex gate, which the leader's brief forbade calling. The code findings are unchanged: F1 (confirm or fix before landing), F2 (follow-up card), F3-F10 minor.
