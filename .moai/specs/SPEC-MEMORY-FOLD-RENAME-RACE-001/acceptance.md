# acceptance.md — SPEC-MEMORY-FOLD-RENAME-RACE-001

Acceptance criteria for the fold rename-window lost-update fix. Every criterion is a mechanically checkable command with an expected outcome; race-window criteria are judged by repeated runs with the race detector (premise P5 — never a single green run). Verification commands are single invocations (no pipes, no chaining). Each criterion carries a class (§0), and the class states what its pass proves.

## §0 Criterion classes

- **Adopted (two-cell).** A RED-now cell and a green-path cell, authored as a pair (verification-completeness §2). The RED-now cell carries four elements together: the command, its verbatim stdout, its exit code, and the tree SHA it was measured on. Those elements sit in an evidence entry in §5 that the criterion cites by id. The green-path cell names the milestone that flips the criterion. An adopted criterion's pass proves the fix.
- **Regression guard.** Must-pass at card close, with a green-path cell and no RED-now cell. The criterion states why no RED-now cell exists. Its pass shows that an invariant still holds; it does not prove the fix by itself.
- **Unadopted.** The RED-now cell is not yet on record. An unadopted criterion is never cited as a pass.
- **Post-close.** The observation lands after card close. Its result is carried to the release window and is not a card-close pass. §4 states the exception.
- **Release-blocking.** No criterion in this SPEC holds this class. A criterion may hold it only with a four-element RED-now cell that can be re-executed on a tree the card controls (verification-completeness §2.1). v0.1.3 carried a release-gate claim on AC-MRR-008; v0.1.4 reclassified it (§1, AC-MRR-008).

## §1 The criteria

### AC-MRR-001 — The loss is reproduced before the fix (RED half of the adopted pair)

- **Class**: adopted (two-cell, RED half). Not release-blocking.
- **Given** the pre-fix tree — the M1 commit `84bdf072c`: the two test-only seams and the RED test over the base production code. `git diff --stat 2aab5f797 2f436db17 -- internal` is empty, and the M1 diff to `memory_fold.go` adds only the seam fields and their call sites. **When** `TestFoldRenameWindowConcurrentWriter` runs, and the seam injects a concurrent writer's publish between the last byte check and `os.Rename`, **Then** the fold returns success while the writer's line is gone from the store file, on both write surfaces, and the test fails on every iteration.
- **Command** (the working directory is a materialization of `84bdf072c` made with `git archive`, not the card worktree): `go test -race -count=10 ./internal/cli -run '^TestFoldRenameWindowConcurrentWriter$'`
- **Expected (pre-fix)**: exit non-zero; 10 of 10 iterations FAIL; 20 of 20 subtests FAIL (`archive-append`, `memory-rewrite`); the failure text names the absent writer line. **Observed in the iteration-4 repair pass**: exit 1; 10 top-level FAIL lines; 20 subtest FAIL lines; 20 defect-message lines (§5, E-RED-001).
- **Why red for the right reason**: the fold reports success (`runFoldOK` passes) while the writer's publish, made at the `mutateBeforeRename` point after every pre-rename check, is destroyed by `os.Rename`. A compile error or a fixture error is not this red. progress.md §E.2 keeps an earlier fixture red as a wrong-reason record and does not cite it as this criterion's RED.
- **Premise note**: the existing `orderProbe("bytes-done")` seam also fires inside the window (spec.md §1.2 P1). The dedicated `mutateBeforeRename` seam is used because it keeps the mutation contract separate from `orderProbe`'s stage-recording contract, and because it fires after the abandonment check.
- **Environment note**: the run recorded in progress.md §E.2 was prefixed with the B-2 env scrub. The three files this test touches reference no `MOAI_KANBAN_*` variable at `84bdf072c` (checked), so the scrub does not bear on the outcome. The repair-pass run was unprefixed.

### AC-MRR-002 — No silent loss under the lock (green half of the adopted pair)

- **Class**: adopted (green half).
- **Given** the fixed tree (M2 and later), **When** the same test runs with its writer in the cooperating (lock-taking) form, **Then** the writer's line is present in the final store on every iteration and the fold's outcome is success.
- **Command**: `go test -race -count=10 ./internal/cli -run '^TestFoldRenameWindowConcurrentWriter$'`
- **Expected**: exit 0; `ok`; 10 of 10 iterations green under the race detector. **Observed at HEAD `a7b89e294` in the repair pass**: `ok  github.com/modu-ai/moai-adk/internal/cli  2.192s`, exit 0 (§5, E-HEAD-001).
- **Flips**: M2 step 3. The writer's non-blocking acquire is refused inside the apply, and the post-release publish lands.

### AC-MRR-003 — Contended fold refuses cleanly

- **Class**: regression guard (must-pass at card close).
- **Given** a store whose lock is held by another holder in the test, **When** `moai memory fold --card t1568 --yes --dir <store>` runs, **Then** the process exits non-zero, the error names the contended store, and both store files are byte-identical to their pre-invocation state (asserted inside the test).
- **Command**: `go test -count=1 ./internal/cli -run '^TestFoldStoreLockContentionRefusesCleanly$'`
- **Expected**: exit 0. Recorded in run M3 (progress.md §E.2): `ok  github.com/modu-ai/moai-adk/internal/cli  2.864s`, exit 0.
- **Why no RED-now**: the test was authored in M3 (commit `1ab3503f1`) and holds the store lock through the M2 lock API. Neither the test nor that API exists at `84bdf072c`, so a pre-fix run could only fail to compile, and a compile failure is not the defect (verification-completeness §2). A raw-`flock` re-authoring could supply a RED. That is a scope option (plan §F, M5 note), not adopted here.

### AC-MRR-004 — Both write surfaces are covered

- **Class**: adopted (two-cell, per surface).
- **Given** the window test over its two write surfaces — the archive append (`project_card_archive_2026_10.md`) and the MEMORY.md rewrite (premise P3) — **When** it runs, **Then** the no-silent-loss contract of AC-MRR-002 holds on each surface.
- **Command**: `go test -race -count=10 ./internal/cli -run '^TestFoldRenameWindowConcurrentWriter$'`. The per-surface cells are the subtests `archive-append` and `memory-rewrite`; the top-level result is their conjunction.
- **RED-now (per surface)**: E-RED-001 (§5). On `84bdf072c` both subtests FAIL, and each failure message names its own surface.
- **Green-path**: the M2 flip. Both subtests go green, and the AC-MRR-002 observation covers both.
- **Expected**: exit 0; `ok`; 10 of 10 iterations.

### AC-MRR-005 — The lock file is invisible to the store's own tooling

- **Class**: regression guard (must-pass at card close).
- **Given** a store directory carrying `.moai-fold.lock`, **When** `moai memory doctor --dir <store> --json` and a fold preview run on it, **Then** their outputs are byte-identical to the same store without the lock file (compared inside the test), and no unlinked-archive listing entry names the lock file.
- **Command**: `go test -count=1 ./internal/cli -run '^TestFoldLockFileInvisibleToTooling$'`
- **Expected**: exit 0. Recorded in run M3: `ok  github.com/modu-ai/moai-adk/internal/cli  0.819s`, exit 0.
- **Why no RED-now**: the subject of the criterion is the lock file, which exists only after the fix. At `84bdf072c` there is no file for the tooling to react to, and the test seeds the file through the M2 API. The criterion guards REQ-MRR-005 against regression.

### AC-MRR-006 — The card-close step's bound is intact under contention

- **Class**: regression guard (must-pass at card close). Cell 2's precondition is added in M5 (plan §F).
- **Cell 1 — Given** a contended store and the card-close fold step under a shortened bound (200 ms), **When** the bound expires while the step waits, **Then** the step reports at most one stderr line and begins no write.
- **Cell 2 — Given** the step holding the lock with its apply in flight when the bound expires, **When** the caller's timeout branch reports abandonment and the worker's exit is observed, **Then** a subsequent non-blocking acquire on the same store succeeds, because the worker's deferred release ran after its apply observed the abandonment.
- **Cell 2 precondition (checked; M5)**: while the apply is parked at its `mutateBetweenWrites` point and before the test signals "lock held", a non-blocking acquire by a second lock object is refused. An acquisition at that point fails cell 2, because the step did not hold the lock when the bound expired.
- **Command**: `go test -race -count=1 ./internal/cli -run '^TestFoldOnDoneContentionAbandonsWithoutWrite$'` (both cells live in the same test).
- **Expected**: exit 0. Recorded in run M3: `ok  github.com/modu-ai/moai-adk/internal/cli  5.325s`, exit 0. The precondition becomes part of the expected outcome once M5 lands.
- **Why no RED-now**: both cells set the contention through the M2 lock API. The test was authored in M3 (`memory_fold_wiring_test.go`, commit `1ab3503f1`), so no pre-fix run exists that could be red for the defect.

### AC-MRR-007 — The fold family stays green (regression surface: t1502 fold family and MEMORY.md index)

- **Class**: regression guard (must-pass at card close).
- **Given** the fixed tree, **When** the affected family runs — the 27 `TestMemoryFold*` tests (including the write-ordering regression `TestMemoryFold_ArchiveRecheckedBeforeMemoryRename`, the abandonment and timeout-cleanup cells, and the 13 `TestMemoryFoldOnDone_*` wiring tests), plus `TestReviewArchiveUpdateDuringEffectiveScan` and `TestReviewSequentialAbandonedTempOwnership` (the codex-review round-2 regression over the wiring's per-worker temp ownership) — **Then** every test passes under the race detector, and the module graph is unchanged.
- **Command 1**: `go test -race -count=1 -timeout 900s ./internal/cli -run '^(TestMemoryFold|TestReviewArchiveUpdate|TestReviewSequentialAbandonedTempOwnership).*$'`
- **Expected 1**: exit 0; equal in health to the green-before measured at plan phase (progress.md §E.1, attempt 4: `ok  github.com/modu-ai/moai-adk/internal/cli  42.974s`, exit 0). The swept set is the 29 tests verified by the `go test -list` gate of plan §C.1 before each use.
- **Command 2**: `git diff --exit-code 2aab5f797 HEAD -- go.mod go.sum`
- **Expected 2**: exit 0, no output. This is the REQ-MRR-007 "no new module dependency" check. Both ends are pinned: `2aab5f797` is the pre-implementation base and `HEAD` is the tree under test, so the comparison is between two commits, not against a moving ref. **Observed at HEAD `a7b89e294` in the repair pass**: exit 0 (§5, E-HEAD-001).
- **Full-suite note**: the whole `./internal/cli` suite is judged by CI (plan §B-1). `go build ./...` exit 0 accompanies this criterion (recorded at M4).

### AC-MRR-008 — Windows parity (regression guard with a post-close observation)

- **Class**: regression guard, with a post-close observation (§4). **Reclassified in v0.1.4 from release-blocking.** v0.1.3 stated that "a red Windows leg blocks the release" on this criterion. The Windows-leg verdict cannot be re-executed on a tree the card controls, so under verification-completeness §2.1 the criterion loses release-blocking eligibility. No criterion in this SPEC gates a release.
- **Given** `fold_store_lock_windows_test.go` (`//go:build windows`; `TestFoldStoreLockWindowsSemantics`: acquire → contending acquire refused → release → acquire succeeds on the real `LockFileEx` path), **When** the Windows leg of `release-pr-multi-os.yml` runs it (`windows-latest`, `go test -json -race -timeout 35m ./...`; the `full-matrix-test` matrix job carries no `continue-on-error`, and it runs for `release/*` pull requests whose diff contains Go code, and on manual dispatch), **Then** the test passes on that leg. That leg is the only Windows surface that executes `internal/cli` root packages. The PR gate `pr-multi-os-gate.yml` filters on `internal/hook/**`, `internal/cli/worktree/**`, `go.mod`, `go.sum`, and its own workflow file, and its test command covers only `./internal/hook/...` and `./internal/cli/worktree/...`.
- **Card-close command 1**: `GOOS=windows go build ./internal/cli/...`. Expected exit 0 (recorded at M2 and M4; observed at HEAD `a7b89e294` in the repair pass).
- **Card-close command 2**: `GOOS=windows go vet ./internal/cli/`. Expected exit 0. It type-checks the windows-tagged test file, which the build does not. **Observed at HEAD `a7b89e294` in the repair pass**: exit 0.
- **Post-close observation**: the Windows leg's verdict on the release pull request, read from the job result. It is an external CI result, so it is not a card-close criterion (§4).

### AC-MRR-009 — The whole-apply span is held

- **Class**: regression guard (must-pass at card close on both entry points once M5 lands). The verb-path check is in place. The card-close check is absent at HEAD and is required by M5.
- **Given** a store and a second independent lock object (its own descriptor), **When** the fold apply runs on each entry point — the verb (`moai memory fold --yes`) and the card-close step (`foldOnDoneStep`, through the wiring fixture) — and the test samples a non-blocking acquire at four points, **Then** the recorded outcomes are exactly (acquired, refused, refused, refused) on each entry point. The four points are:
  1. Before the apply, with the lock not yet held: **acquired**. On the verb path this is the `mutateDisk` seam, which fires in `newMemoryFoldCmd` (`memory_fold.go:183–187`). `mutateDisk` is not on the card-close path, so the card-close subtest takes this sample immediately before it invokes the step. The expected outcome is unchanged.
  2. Inside the first write (the archive append; `mutateDuringWrite` for that surface): **refused**.
  3. Between the two writes (`mutateBetweenWrites`, after the archive append and before the MEMORY.md rewrite): **refused**.
  4. Inside the second write (the MEMORY.md rewrite; `mutateBeforeRename` for that surface): **refused**.
- **Command**: `go test -race -count=10 ./internal/cli -run '^TestFoldStoreLockSpanHeldThroughApply$'`
- **Expected**: exit 0; 10 of 10 iterations green, with both entry-point subtests run. A run that executes no subtest is not a pass. **Observed at HEAD `a7b89e294` in the repair pass** (verb path only, as implemented at HEAD): `ok  github.com/modu-ai/moai-adk/internal/cli  1.980s`, exit 0 (§5, E-HEAD-001).
- **Why it kills both mutant classes**: a per-rename lock yields (acquired, refused, acquired, …), because the third sample acquires. A per-write lock, acquired and released once per `atomicWriteFoldFile` call, yields (acquired, refused, acquired, refused), because the between-writes sample acquires. Only a whole-apply span yields (acquired, refused, refused, refused). The per-write mutant was demonstrated by the iteration-2 codex backend; it was not re-run in this pass.
- **Why no RED-now**: the test was authored in M2 (`5d60d2e8e`) and uses the M2 lock API, so it cannot run at `84bdf072c`.

### AC-MRR-010 — Cross-process exclusion, observed through a child process

- **Class**: unadopted until M5 records its RED-now cell; then adopted. If the operator chooses plan §F, M5 option (b), it stays unadopted, and REQ-MRR-001's cross-process clause is carried as residual risk.
- **Given** a store directory and a parent process that holds the store lock, **When** a child process runs the fold verb on the same store — the test binary re-executed with a `-test.run` filter and an environment marker, the helper-process pattern of `internal/execerr/execerr_test.go` — **Then** the child's fold is refused while the parent holds the lock (non-zero exit; the error names the store; the store bytes are unchanged), and the child's fold succeeds after the parent releases.
- **Sentinel rule**: the child prints one sentinel line per outcome (`FOLD-CHILD=refused`, then `FOLD-CHILD=applied`), and the parent records each through `t.Logf`. A child that runs zero tests exits 0 with no sentinel, so it fails this criterion. The empty-sweep rule of verification-completeness §1.1 applies.
- **Command**: `go test -count=1 -v ./internal/cli -run '^TestFoldStoreLockCrossProcess$'`
- **Expected (GREEN; on HEAD once M5 lands)**: exit 0; the verbatim output contains `--- PASS: TestFoldStoreLockCrossProcess`, both sentinel lines in the order refused → applied, and no `[no tests to run]` line.
- **Mutant it must kill**: an in-process lock (a process-local mutex in place of the file lock) and a no-op lock both let the child apply while the parent holds the lock, so the refused expectation fails. The in-process criteria (AC-MRR-002, -003, -005, -006, -009) pass under the in-process mutant. That is the iteration-2 codex overlay result, reported here and not re-run.
- **RED-now (required for adoption; M5 step 1)**: on `84bdf072c` (POSIX), the parent holds `<store>/.moai-fold.lock` through a raw `flock` (no M2 API exists there), and the child fold applies while the parent holds it. The expected red is the refused expectation failing for the defect's reason. Record the verbatim output, exit code, and tree SHA in progress.md §E.2 before the GREEN.
- **Windows**: the RED-now is POSIX-only. The GREEN also runs on the Windows leg after card close, which is a post-close observation (§4).

## §2 Traceability

| REQ (spec.md §2) | AC |
|---|---|
| REQ-MRR-001 (whole-apply span; cross-process clause) | AC-MRR-002, AC-MRR-004, AC-MRR-009 (both entry points), AC-MRR-010 (cross-process) |
| REQ-MRR-002 (one lock per store, both surfaces) | AC-MRR-004, AC-MRR-005, AC-MRR-009 |
| REQ-MRR-003 (bounded clean refusal) | AC-MRR-003 |
| REQ-MRR-004 (no silent loss) | AC-MRR-001 (RED half), AC-MRR-002 (green half), AC-MRR-004 (per surface) |
| REQ-MRR-005 (tooling-invisible lock file) | AC-MRR-005 |
| REQ-MRR-006 (abandonment bound intact) | AC-MRR-006 (cells 1–2; cell-2 precondition in M5) |
| REQ-MRR-007 (POSIX and Windows, no new dependency) | AC-MRR-007 (family; go.mod and go.sum), AC-MRR-008 (card-close build and vet; post-close Windows leg) |

Every requirement has at least one criterion. Every criterion traces to at least one requirement, so there is no orphan criterion. The counts are 7 requirements and 10 criteria, within the Tier M ceilings of 16 each.

## §3 Edge cases

- The lock file pre-exists but is unlocked (stale from a killed process): flock's kernel lifetime means the file's presence alone never blocks — the fold acquires normally. Covered by the lock unit test's "acquire on an existing unlocked file" cell.
- The lock file cannot be created (read-only store): the acquire error surfaces as the fold's clean refusal — a store that cannot be locked is a store the fold must not write; the refusal path of AC-MRR-003 covers the shape.
- Two folds on DIFFERENT stores: distinct lock files, no interference — the unit test holds two locks simultaneously.
- The close-path step abandoned WHILE holding the lock: the caller's timeout branch returns before the worker does, so the release is the WORKER's deferred `release()` — it lands when the worker's apply observes the abandonment and returns (synchronized in tests via the wiring's worker-exit signal). Covered by AC-MRR-006's second cell; the release is not synchronous with the caller's return, and in a one-shot CLI process the kernel releases the flock at exit regardless (recorded boundary).

## §4 Quality gates, Definition of Done, and open items

- TRUST 5: Tested (every REQ mapped; race criteria at `-count=10`), Readable/Unified (gofmt, package style), Secured (the lock file is inside the store directory, mode 0644, no secret surface), Trackable (conventional commits carrying the SPEC id and card id).
- **AC-MRR-008 post-close exception.** AC-MRR-008's observation is the Windows leg of `release-pr-multi-os.yml`, which the release process runs at the release window (`release/*` pull requests), after card close; card close is judged on the local gates, and AC-MRR-008's result is carried to the release window and is not recorded as a card-close pass.
- **Definition of Done (card close)**:
  1. AC-MRR-001's RED-now is recorded (§5, E-RED-001). AC-MRR-004's per-surface RED comes from the same observation.
  2. AC-MRR-002, -003, -004, -005, -006, -007, and -009 are green, with their command output on record in progress.md §E.2/§E.3. The card-close subtest of AC-MRR-009 and the cell-2 precondition of AC-MRR-006 are included once M5 lands.
  3. AC-MRR-007's post-fix green equals the green-before recorded at plan phase, and its `go.mod`/`go.sum` comparison exits 0.
  4. AC-MRR-010 is adopted only with its RED-now on record (M5). If the operator chooses option (b) (plan §F), its cross-process clause is recorded as residual risk; until the operator chooses, the clause stays open.
  5. The AC-MRR-008 post-close exception above is stated where the chain is read (plan §E, M3 item 4, M4 item 3).
  6. progress.md §E.2/§E.3 is populated by manager-develop; §E.4 by manager-docs.
- Indirect verification: none required beyond the family command. CI is the full-suite judge (plan §B-1).
- **Open items at iteration 4 (not closed by this document)**: (a) the AC-MRR-006 cell-2 precondition, the AC-MRR-009 card-close subtest, and AC-MRR-010 need test code that is absent at HEAD (plan §F, M5; the scope decision is open); (b) AC-MRR-003, -005, and -009 have no RED-now cell (§0, with the reasons in each criterion); (c) AC-MRR-010 is unadopted until its RED-now is recorded; (d) AC-MRR-008's Windows observation lands at the release window.

## §5 Evidence entries

### E-RED-001 — RED-now observation for AC-MRR-001 and AC-MRR-004 (repair pass, iteration 4)

- **Tree**: `84bdf072c` (the M1 commit), materialized with `git archive` into an empty scratch directory. The card worktree was not the measured tree.
- **Command**: `go test -race -count=10 ./internal/cli -run '^TestFoldRenameWindowConcurrentWriter$'`, with the working directory set to the materialized tree.
- **Exit code**: 1.
- **Verbatim stdout, iteration 1 of 10.** The other nine iterations are identical in structure. The counts below cover the full captured stdout.

```
--- FAIL: TestFoldRenameWindowConcurrentWriter (0.01s)
    --- FAIL: TestFoldRenameWindowConcurrentWriter/archive-append (0.00s)
        memory_fold_test.go:1466: REQ-MRR-004 violated: the fold returned success while the writer's line published inside the archive-append rename window was destroyed from project_card_archive_2026_10.md
    --- FAIL: TestFoldRenameWindowConcurrentWriter/memory-rewrite (0.00s)
        memory_fold_test.go:1466: REQ-MRR-004 violated: the fold returned success while the writer's line published inside the memory-rewrite rename window was destroyed from MEMORY.md
```

- **Counts over the full stdout (53 lines)**: 10 top-level `--- FAIL: TestFoldRenameWindowConcurrentWriter` lines; 20 subtest `--- FAIL` lines; 20 `REQ-MRR-004 violated` lines.
- **Same observation at the original run**: progress.md §E.2, M1 block (same tree, exit 1).

### E-HEAD-001 — observations at HEAD `a7b89e294` in the repair pass

| Command | Observed | Criteria |
|---|---|---|
| `go test -race -count=10 ./internal/cli -run '^TestFoldRenameWindowConcurrentWriter$'` | `ok  github.com/modu-ai/moai-adk/internal/cli  2.192s`, exit 0 | AC-MRR-002, AC-MRR-004 |
| `go test -race -count=10 ./internal/cli -run '^TestFoldStoreLockSpanHeldThroughApply$'` | `ok  github.com/modu-ai/moai-adk/internal/cli  1.980s`, exit 0 (verb path only, as implemented at HEAD) | AC-MRR-009 (verb path) |
| `GOOS=windows go build ./internal/cli/...` | exit 0 | AC-MRR-008 card-close command 1 |
| `GOOS=windows go vet ./internal/cli/` | exit 0 | AC-MRR-008 card-close command 2 |
| `git diff --exit-code 2aab5f797 HEAD -- go.mod go.sum` | exit 0 | AC-MRR-007 command 2 |

Code facts used by the repairs (read at HEAD and at `84bdf072c`):
- At `84bdf072c` no lock API exists (`git grep` for `foldStoreLock` and `newFoldStoreLock` returns nothing). The tests for AC-MRR-003, -005, -006, and -009 are absent there.
- At `84bdf072c` `TestFoldRenameWindowConcurrentWriter` already has the `archive-append` and `memory-rewrite` subtests.
- `mutateDisk` is invoked only inside `newMemoryFoldCmd` (`memory_fold.go:183–187`), not in `applyFold` and not in `foldOnDoneStep`.
- `mutateDuringWrite` and `mutateBeforeRename` are invoked inside `atomicWriteFoldFile`, and `mutateBetweenWrites` inside `applyFold`. All three fire on both entry points.
- The cell-2 `mutateBetweenWrites` callback in `memory_fold_wiring_test.go` signals and parks, with no probe.
