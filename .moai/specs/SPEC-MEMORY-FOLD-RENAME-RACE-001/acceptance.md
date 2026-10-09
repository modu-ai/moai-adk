# acceptance.md — SPEC-MEMORY-FOLD-RENAME-RACE-001

Acceptance criteria for the fold rename-window lost-update fix. Every criterion is a mechanically checkable command with an expected outcome; race-window criteria are judged by repeated runs with the race detector (premise P5 — never a single green run). Verification commands are single invocations (no pipes, no chaining).

## §1 The criteria

### AC-MRR-001 — The loss is reproduced before the fix (RED cell of the two-cell pair)

- **Given** the pre-fix tree (baseline `2aab5f797` + the M1 test-only seam commit), **When** `TestFoldRenameWindowConcurrentWriter` runs — the seam injects a concurrent writer's publish into the window between the last byte check and `os.Rename` — **Then** the fold returns success while the writer's line is gone from MEMORY.md, and the test FAILS on every iteration.
- **Command**: `go test -race -count=10 ./internal/cli -run '^TestFoldRenameWindowConcurrentWriter$'`
- **Expected now (pre-fix)**: exit non-zero; failure text names the absent writer line; 10/10 iterations red.
- **Status at plan phase**: PENDING EXECUTION by choice, not by impossibility — the existing `orderProbe("bytes-done")` seam (`memory_fold.go:643-645`) already fires inside the window, so a plan-phase RED is executable today by setting it per plan.md B-3's set/restore discipline. The deferral to M1 is the clean-dedicated-seam choice: `orderProbe`'s contract is stage recording (owned by the ordering regression), not mutation, and the dedicated `mutateBeforeRename` seam keeps the mutation contract separate while firing after the abandonment check — the true last observable moment. This cell is executed and its verbatim output + exit code + tree SHA recorded in progress.md §E.2 during run M1; the plan forbids M2 from starting before the observation is on record.
- **Why red for the right reason**: the only change on the tree is the test-only seam + test (no production change), so the red is the defect, not collateral.

### AC-MRR-002 — No silent loss under the lock (green path of the two-cell pair)

- **Given** the fixed tree (M2), **When** the same test runs with its writer in the cooperating (lock-taking) form, **Then** the writer's line is present in the final store on every iteration and the fold's outcome is success — the window can no longer destroy a cooperating writer's publish.
- **Command**: `go test -race -count=10 ./internal/cli -run '^TestFoldRenameWindowConcurrentWriter$'`
- **Expected (post-fix)**: exit 0; `PASS`, 10/10 iterations green under the race detector.
- **Flips**: M2 step 3. The test body's contract assertion ("writer's line present at end") is identical in both cells — only the writer's lock-taking changes (plan.md §A.3).

### AC-MRR-003 — Contended fold refuses cleanly

- **Given** a store whose lock is held by another holder in the test, **When** `moai memory fold --card t1568 --yes --dir <store>` runs, **Then** the process exits non-zero, the error names the contended store, and both MEMORY.md and the archive index are byte-identical to their pre-invocation state (byte equality asserted inside the test).
- **Command**: `go test -count=1 ./internal/cli -run '^TestFoldStoreLockContentionRefusesCleanly$'`
- **Expected**: exit 0 (the test passes); the test's internal assertions carry the byte-identity proof.

### AC-MRR-004 — Both write surfaces are covered

- **Given** the window test parameterized over the two `atomicWriteFoldFile` call sites (archive append; MEMORY.md rewrite — premise P3), **When** it runs, **Then** the no-silent-loss contract (AC-MRR-002's assertion) holds on both surfaces.
- **Command**: `go test -race -count=10 ./internal/cli -run '^TestFoldRenameWindowConcurrentWriter$'`
- **Expected**: exit 0; the test's subtests cover both surfaces, both green.

### AC-MRR-005 — The lock file is invisible to the store's own tooling

- **Given** a store directory carrying `.moai-fold.lock`, **When** `moai memory doctor --dir <store>` and a fold preview run on it, **Then** their outputs are byte-identical to the same store without the lock file (asserted by comparison inside the test), and no unlinked-archive listing entry names the lock file.
- **Command**: `go test -count=1 ./internal/cli -run '^TestFoldLockFileInvisibleToTooling$'`
- **Expected**: exit 0.

### AC-MRR-006 — The card-close step's bound is intact under contention

- **Given** a contended store and the card-close fold step (MFB wiring) with a shortened test bound, **When** the bound expires while the step waits, **Then** the step reports at most one stderr line and begins no write (the wiring test's existing recorder/seam conventions assert this).
- **Given** the step HOLDING the lock with its apply in flight when the bound expires, **When** the caller's timeout branch reports abandonment and the worker's exit is observed (the wiring's existing worker-exit synchronization), **Then** a subsequent non-blocking acquire on the same store succeeds — the worker's deferred release ran after its apply observed the abandonment.
- **Command**: `go test -race -count=1 ./internal/cli -run '^TestFoldOnDoneContentionAbandonsWithoutWrite$'` (both cells live in the same test)
- **Expected**: exit 0.

### AC-MRR-007 — The fold family stays green (regression surface: t1502 fold 계열 + MEMORY.md index)

- **Given** the fixed tree, **When** the affected family runs (the fold regression family: all 27 `TestMemoryFold*` tests — including the write-ordering regression `TestMemoryFold_ArchiveRecheckedBeforeMemoryRename`, the abandonment and timeout-cleanup cells, and the 13 `TestMemoryFoldOnDone_*` wiring tests — plus `TestReviewArchiveUpdateDuringEffectiveScan` and `TestReviewSequentialAbandonedTempOwnership`, the codex-review round-2 regression over the fold wiring's per-worker temp ownership), **Then** every test passes under the race detector. The selector's swept set is verified non-empty and family-complete by `go test -list '^(TestMemoryFold|TestReviewArchiveUpdate|TestReviewSequentialAbandonedTempOwnership).*$' ./internal/cli` → 29 tests (plan.md §C.1).
- **Command**: `go test -race -count=1 -timeout 900s ./internal/cli -run '^(TestMemoryFold|TestReviewArchiveUpdate|TestReviewSequentialAbandonedTempOwnership).*$'`
- **Expected**: exit 0; the baseline for this exact command was measured at plan phase on `2aab5f797` and recorded in progress.md §E.1 — green-after must equal green-before. (An earlier recorded baseline, `ok … 414.385s`, was measured with the wrong selector `^Test(Fold|Review).*$` — 53 Review-prefix tests, zero fold tests — and is kept on record only as a mis-measurement, never cited as the green-before.)
- **Full-suite note**: the whole `./internal/cli` suite is judged by CI (plan.md B-1); `go build ./...` exit 0 accompanies this criterion.

### AC-MRR-008 — Windows parity: semantics executed on the release multi-OS gate's Windows leg

- **Given** a windows-tagged lock-semantics unit test (`fold_store_lock_windows_test.go`, `//go:build windows`: acquire → contending acquire refused → release → acquire succeeds, on the real `LockFileEx` path), **When** the Windows leg of `release-pr-multi-os.yml` (`windows-latest`, `go test -json -race -timeout 35m ./...` — the full suite, every leg blocking) runs it, **Then** it passes. That leg is the judge because it is the only Windows surface that executes `internal/cli` root packages: the PR gate (`pr-multi-os-gate.yml`) never fires its matrix on `internal/cli` root changes (its paths-filter covers `internal/hook/**`, `internal/cli/worktree/**`, `go.mod`, `go.sum`, and the workflow itself) and its test command covers `./internal/hook/... ./internal/cli/worktree/...` only — measured at plan-audit iter-2 (D6). No local surface executes the Windows path; `GOOS=windows go build ./internal/cli/...` exit 0 remains the local compile gate.
- **Command**: the `release-pr-multi-os.yml` Windows leg result — its `go test -json -race -timeout 35m ./...` step, verdict read from the job result. Local compile gate: `GOOS=windows go build ./internal/cli/...`
- **Expected**: the Windows leg green on the lock-semantics test; local build exit 0. **Timing, recorded honestly**: this leg runs at the release window — after card close — so AC-MRR-008's green is the one criterion whose observation may land after this card's own close (§4's DoD carries the same statement). A red Windows leg blocks the release, which is the enforcement point.

### AC-MRR-009 — The whole-apply span is held (span observation; kills the per-rename AND per-write mutants)

- **Given** a store and a second independent lock object (its own descriptor), **When** the fold apply runs on the verb path and the test samples a non-blocking acquire at FOUR seam points — before `applyFold` (the `mutateDisk` seam: the lock is not yet held → acquire SUCCEEDS), inside the first write (the `mutateDuringWrite` seam: REFUSED), between the two writes (the test-only `mutateBetweenWrites` seam in `applyFold`, after the first rename completes and before the second `atomicWriteFoldFile`: REFUSED), and inside the second write (the `mutateBeforeRename` seam: REFUSED) — **Then** the recorded outcomes are exactly (success, refusal, refusal, refusal): the lock is held continuously from `applyFold` entry through the final rename.
- **Why it kills both mutant classes**: a per-rename lock (acquired around each rename, released between) yields (success, refusal, success, …) — the third sample acquires; a per-write lock (acquired/released per `atomicWriteFoldFile` call) holds at both interior write samples but yields (success, refusal, SUCCESS, refusal) — the between-writes sample acquires. Only the whole-apply span produces all three refusals. (Audit iter-2 D7: codex demonstrated the per-write mutant passing the 3-tuple with samples (true, false, false); the fourth sample is its kill.)
- **Command**: `go test -race -count=10 ./internal/cli -run '^TestFoldStoreLockSpanHeldThroughApply$'`
- **Expected**: exit 0, 10/10 iterations green under the race detector.

## §2 Traceability

| REQ (spec.md §2) | AC |
|---|---|
| REQ-MRR-001 (whole-apply span) | AC-MRR-002, AC-MRR-004, AC-MRR-009 |
| REQ-MRR-002 (one per store, both surfaces) | AC-MRR-004, AC-MRR-005, AC-MRR-009 |
| REQ-MRR-003 (bounded clean refusal) | AC-MRR-003 |
| REQ-MRR-004 (no silent loss) | AC-MRR-001 (RED) + AC-MRR-002 (green) |
| REQ-MRR-005 (tooling-invisible lock file) | AC-MRR-005 |
| REQ-MRR-006 (abandonment bound intact) | AC-MRR-006 |
| REQ-MRR-007 (POSIX+Windows, no new dep) | AC-MRR-008, AC-MRR-007 (`go build ./...` + unchanged `go.mod`) |

Every requirement has at least one criterion; no orphan criterion.

## §3 Edge cases

- The lock file pre-exists but is unlocked (stale from a killed process): flock's kernel lifetime means the file's presence alone never blocks — the fold acquires normally. Covered by the lock unit test's "acquire on an existing unlocked file" cell.
- The lock file cannot be created (read-only store): the acquire error surfaces as the fold's clean refusal — a store that cannot be locked is a store the fold must not write; the refusal path of AC-MRR-003 covers the shape.
- Two folds on DIFFERENT stores: distinct lock files, no interference — the unit test holds two locks simultaneously.
- The close-path step abandoned WHILE holding the lock: the caller's timeout branch returns before the worker does, so the release is the WORKER's deferred `release()` — it lands when the worker's apply observes the abandonment and returns (synchronized in tests via the wiring's worker-exit signal). Covered by AC-MRR-006's second cell; the release is not synchronous with the caller's return, and in a one-shot CLI process the kernel releases the flock at exit regardless (recorded boundary).

## §4 Quality gates and Definition of Done

- TRUST 5: Tested (every REQ mapped, race criteria at `-count=10`), Readable/Unified (gofmt, package style), Secured (the lock file is inside the store dir, mode 0644, no secret surface), Trackable (conventional commits carrying the SPEC id and card id).
- DoD: AC-MRR-001's RED observed and recorded → AC-MRR-002..009 green with outputs on record → AC-MRR-007's post-fix green equals the pre-recorded baseline → progress.md §E.2/§E.3 populated. AC-MRR-008 timing exception, stated where the chain is read: its judge is `release-pr-multi-os.yml`'s Windows leg, which runs at the release window — after card close — so the chain is complete at card close except this one criterion, whose observation is carried to the release gate; a red Windows leg blocks the release.
- Indirect verification: none required beyond the family command — CI is the full-suite judge (plan.md B-1).
