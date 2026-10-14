# plan.md — SPEC-MEMORY-FOLD-RENAME-RACE-001

Implementation plan for card t1568. Tier M. Baseline tree: `2aab5f797` (worktree branch `WT-fold-rename-race`, after the lane's base absorb from `81786284e` — lane decision D-1). Development mode: TDD (RED-first is the card's explicit mandate and M1 below).

## §A Context

### A.1 The defect and the approach

The fold apply path ends `atomicWriteFoldFile` with a check→rename tail (`internal/cli/memory_fold.go:635-649`) that nothing serializes: a concurrent writer publishing between the last byte comparison and `os.Rename` is destroyed while the fold reports success. The approach is a per-store advisory lock (flock on POSIX, LockFileEx on Windows) held for the whole `applyFold` span — acquired before the first pre-write check, released after the last rename — so two fold processes on one store never overlap their windows. Requirements: spec.md §2 (REQ-MRR-001..008). Mechanism rationale and rejected alternatives: design.md. Decisions OD-1..OD-5 map one-to-one onto decision-index.md Q1..Q5; every row is implementation-level with the published default applied.

### A.2 Files the run phase will touch (certain set — 7)

| File | New/Edit | What |
|---|---|---|
| `internal/cli/memory_fold.go` | edit | M1: add two test-only seam fields — `mutateBeforeRename` (invoked immediately before `os.Rename`) and `mutateBetweenWrites` (invoked in `applyFold` between the two `atomicWriteFoldFile` calls, at the re-read/effective-state position) — with their invocations (test-only scaffolding, B-3 set/restore discipline). M2: acquire/release the store lock around the `applyFold` body. No existing check moves. |
| `internal/cli/memory_fold_test.go` | edit | M1: the RED test. M2: the test's writer becomes the lock-taking (cooperating) form. M5 (not run — §F): the card-close subtest of `TestFoldStoreLockSpanHeldThroughApply` (AC-MRR-009) was planned here and is not written. |
| `internal/cli/fold_store_lock_unix.go` | new | Unexported `foldStoreLock` — `unix.Flock` `LOCK_EX\|LOCK_NB` with a bounded retry loop; pattern copied from `internal/sessionmsg/lock_unix.go` (including the `unacquiredFD = -1` sentinel lesson). |
| `internal/cli/fold_store_lock_windows.go` | new | Windows parity — `LockFileEx` `LOCKFILE_EXCLUSIVE_LOCK\|LOCKFILE_FAIL_IMMEDIATELY`; pattern from `internal/sessionmsg/lock_windows.go`. |
| `internal/cli/fold_store_lock_test.go` | new | Lock unit tests: acquire/refuse/release, retry-bound. Cross-process: the fold-level child-process criterion AC-MRR-010 belongs to M5, which is not run (§F), so the file holds no cross-process test; the planned form is built on the helper-process pattern of `internal/execerr/execerr_test.go`. Two-lock-object contention is an in-process check, so it is not an alternative for the cross-process property. |
| `internal/cli/fold_store_lock_windows_test.go` | new | Windows-tagged (`//go:build windows`) lock-semantics test — acquire → contending acquire refused → release → acquire succeeds on the real `LockFileEx` path; executed by the Windows CI job (AC-MRR-008's post-close judge); its card-close leg is `GOOS=windows go vet ./internal/cli/`. |
| `internal/cli/memory_fold_wiring_test.go` | edit | The AC-MRR-006 cell (`TestFoldOnDoneContentionAbandonsWithoutWrite`), written in M3 on the wiring test's existing conventions (`wireFixture`, `runWireClose`, `waitWorkerExit`). M5 (not run — §F): the cell-2 HOLDING precondition was planned here and is not written. Added at the iteration-4 repair (D5): the M3 run edited this file, and the earlier table omitted it. |

Named contingencies (each +1 file, none expected): (i) if the auditor requires the lock to live outside `internal/cli` for t1595 adoption, it promotes to `internal/filelock/` with the same body; (ii) if `go vet` on windows flags a build-tag detail, a `fold_store_lock_stub.go` may be needed — it would be a defect, not a design change.

### A.3 The window test's two forms (why the test changes in M2)

The RED test (M1) injects a direct write at the new seam — on the pre-fix tree no lock exists, so the write lands in the window and the rename destroys it: the loss is observed deterministically. After the fix, a direct seam write would STILL be destroyed (the lock cannot protect a writer that ignores it — spec.md P6), so in M2 the same test's writer becomes the cooperating form: it attempts a non-blocking acquire of the very lock the fold now holds, records "serialized-out" on refusal, and the test performs the writer's publish after the fold returns. The final assertion is the contract form (REQ-MRR-004; the refused-then-re-published branch is REQ-MRR-008): **the writer's line is present in the final store, whatever the fold's outcome**. Pre-fix: line written in-window, then destroyed → absent → RED. Post-fix: in-window publish refused by the held lock, post-apply publish lands → present → GREEN. Same test body, no goroutine timing, no deadlock (non-blocking try only), deterministic under `-race -count=10`.

## §B Known Issues

- B-1 — The whole `./internal/cli` suite is structurally red inside a card worktree (shared-tree fixture contention; recorded repo lesson from t1542). Local verification is scoped to the affected family (`-run` pattern below); CI on `origin/develop` judges the full suite.
- B-2 — Lane-local environment can falsely redden env-reading guard tests; verification runs in the env-scrubbed compound form (`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test …`).
- B-3 — `foldTestSeam` is documented as "never set while another fold apply runs" (`memory_fold.go:62-64`); the new seam field inherits that contract, and the new tests set/restore it the way the existing `mutateDuringWrite` tests do (single apply per seam, `t.Cleanup` restore).

## §C Pre-flight

1. Baseline family measurement on `2aab5f797` with the FINAL selector `^(TestMemoryFold|TestReviewArchiveUpdate|TestReviewSequentialAbandonedTempOwnership).*$` — green-before on record: `ok  github.com/modu-ai/moai-adk/internal/cli  42.974s`, exit 0 (attempt 4, progress.md §E.1). Selector history: the first selector `^Test(Fold|Review).*$` was wrong (53 Review-prefix tests, ZERO fold tests; its 414.385s baseline is a recorded mis-measurement — audit iter-1 D1); the second selector `^(TestMemoryFold|TestReviewArchiveUpdate).*$` (38.008s) missed `TestReviewSequentialAbandonedTempOwnership` (audit iter-2 A4) and is superseded by the final 29-test form. The selector's swept set is verified before every use: `go test -list '^(TestMemoryFold|TestReviewArchiveUpdate|TestReviewSequentialAbandonedTempOwnership).*$' ./internal/cli` → 29 tests, including `TestMemoryFold_ArchiveRecheckedBeforeMemoryRename`, the 13 `TestMemoryFoldOnDone_*`, and `TestReviewSequentialAbandonedTempOwnership` (verified at plan phase, exit 0). The very first attempt ran on the pre-absorb base `81786284e` and hit its own 240s timeout — a tool artifact, not a red. If red before any change, STOP and report: a red baseline is a gap, not a starting point.
2. Confirm the window lines still match spec.md P1 (`grep -n "os.Rename" internal/cli/memory_fold.go` → `:649`) before M1 edits; the plan's line citations are pinned to `2aab5f797`.
3. `git rev-parse --short HEAD` immediately before the M1 commit — never a value read earlier in the turn.

## §D Constraints

- D-1 — The REQ-MFB-004 check ordering is frozen (spec.md C-3): the lock acquisition sits at the TOP of `applyFold` (before the first `checkFoldUnchanged`) and the release is deferred; nothing between the existing checks moves. The `orderProbe` sequence (`effective-start` → `bytes-done`) must be observable unchanged by the existing regression test `TestReviewArchiveUpdateDuringEffectiveScan`.
- D-2 — Bounded wait: the retry loop's deadline is a named constant in the new lock file (value: 2 seconds — same magnitude as the close-path bound's production ceiling; exact value settled in review). On deadline expiry: the clean refusal error of REQ-MRR-003, naming the store path.
- D-3 — The lock file is `<store-dir>/.moai-fold.lock`, mode 0644, created `O_CREAT|O_RDWR|O_CLOEXEC` (POSIX) / `O_CREAT` (Windows). It is never removed on release (removal races a concurrent opener; flock's kernel lifetime makes removal unnecessary). Premise (spec.md §4, operator decision `d-20261010T042423Z-722b`): the lock path names a regular file that the fold never replaces. The open follows symbolic links and performs no identity check between the descriptor and the path, and this plan does not add either; a link planted at the lock path is out of scope under the threat model stated in spec.md §4. The residual risk is spec.md §5 R-1.
- D-4 — The card-close path's `writesForbidden` checks and the `run.temp` recovery keep their exact positions; the lock changes nothing about abandonment (REQ-MRR-006).
- D-5 — No `go.mod` change (REQ-MRR-007; AC-MRR-007 checks `go.mod` and `go.sum` against the pinned base); no edit outside the seven files of §A.2.

## §E Self-Verification

Each milestone carries its own verifying commands (§F); the milestone is done when those commands' outputs are observed and recorded in progress.md §E.2 (run phase). No milestone is closed on an unrun command, except as the exception below states.

**AC-MRR-008 post-close exception.** AC-MRR-008's observation is the Windows leg of `release-pr-multi-os.yml`, which the release process runs at the release window (`release/*` pull requests), after card close; card close is judged on the local gates, and AC-MRR-008's result is carried to the release window and is not recorded as a card-close pass.

## §F Milestones

Order: decision-reversibility first — the RED proof (most likely to reshape the design if the loss fails to reproduce) leads; mechanical close-out sits at the bottom.

### M1 — RED: the loss reproduced (Priority High)

1. Add two test-only seam fields to `foldTestSeam` and invoke them (both B-3 set/restore discipline): `mutateBeforeRename func(dir, name string)` between the `bytes-done` order probe (`:643-645`) and the rename (`:649`) — after every check, before `os.Rename`; and `mutateBetweenWrites func(dir string)` in `applyFold`, after the first rename (the archive append) completes and before the second `atomicWriteFoldFile` (the re-read/effective-state position) — the observation point that separates a whole-apply lock from a per-write one (AC-MRR-009).
2. Add `TestFoldRenameWindowConcurrentWriter` to `internal/cli/memory_fold_test.go`: t.TempDir store with MEMORY.md carrying a t1568-shaped STRONG line and a linked archive index; apply the fold (`--yes` path, seam set/restored per B-3); the seam hook appends a distinct writer line to MEMORY.md (direct write — the pre-fix world has no lock).
3. Observe RED: the fold returns success; MEMORY.md no longer carries the writer's line; the assertion "the writer's line is present" fails.
   - RED-now command: `go test -race -count=10 ./internal/cli -run '^TestFoldRenameWindowConcurrentWriter$'`
   - Expected (pre-fix): FAIL on every iteration, failure text naming the absent writer line. Record verbatim output + exit code + tree SHA (progress.md §E.2, AC-MRR-001's RED cell).

### M2 — GREEN: the store lock (Priority High)

1. `fold_store_lock_unix.go` / `fold_store_lock_windows.go`: `foldStoreLock` with `acquire(dir) error` (NB try + bounded retry per D-2), `release()`, both build-tagged; `fold_store_lock_test.go` unit cells (acquire→second acquire refused within the bound→release→acquire succeeds).
2. `applyFold`: acquire at entry (D-1 position), `defer release()`; on bounded-wait expiry return the REQ-MRR-003 refusal.
3. Flip the window test's writer to the cooperating form (§A.3) and observe GREEN: `go test -race -count=10 ./internal/cli -run '^TestFoldRenameWindowConcurrentWriter$'` → PASS, every iteration.
4. Parameterize the same test over both write surfaces (archive append call site and MEMORY.md rewrite call site — premise P3): the contract holds on both (AC-MRR-004).
5. Add the span-observation test `TestFoldStoreLockSpanHeldThroughApply` (AC-MRR-009): a second independent lock object samples a non-blocking acquire at FOUR points — `mutateDisk` (success — the lock is not yet held), `mutateDuringWrite` (refused), `mutateBetweenWrites` (refused), and `mutateBeforeRename` (refused). The (success, refusal, refusal, refusal) tuple pins the whole-apply span of REQ-MRR-001 and fails on BOTH mutant classes: a per-rename lock yields (success, refusal, success, …) at the third sample; a per-write lock (acquire/release per `atomicWriteFoldFile` call — the codex-demonstrated iter-2 mutant) yields (success, refusal, success, refusal) at the between-writes sample.

### M3 — Surface cells (Priority Medium)

1. Contention refusal (AC-MRR-003): hold the lock in-test, run the fold verb on that store → non-zero exit, the error names the store, both store files byte-identical to pre-state (asserted inside the test).
2. Close-path abandonment under contention (AC-MRR-006): held lock + the wiring's bounded step → the step reports its one stderr line and begins no write (existing wiring-test conventions).
3. Lock-file invisibility (AC-MRR-005): `moai memory doctor --dir <fixture>` and a fold preview on a store carrying `.moai-fold.lock` produce byte-identical output to the same store without it.
4. Windows parity (AC-MRR-008, regression guard): `fold_store_lock_windows_test.go` executes the lock semantics on the real `LockFileEx` path (acquire → contending acquire refused → release → acquire succeeds). Its post-close judge is the Windows leg of `release-pr-multi-os.yml` (`windows-latest`, `go test -json -race -timeout 35m ./...`; the matrix job carries no `continue-on-error`): the only Windows surface that executes `internal/cli` root packages, because the PR gate `pr-multi-os-gate.yml` filters on `internal/hook/**` and `internal/cli/worktree/**`, not on the `internal/cli` root (measured at plan-audit iter-2 D6; re-read at iter-4). Its card-close legs are `GOOS=windows go build ./internal/cli/...` and `GOOS=windows go vet ./internal/cli/` (the vet leg type-checks the windows-tagged test file, which the build does not), both exit 0. **AC-MRR-008 post-close exception.** AC-MRR-008's observation is the Windows leg of `release-pr-multi-os.yml`, which the release process runs at the release window (`release/*` pull requests), after card close; card close is judged on the local gates, and AC-MRR-008's result is carried to the release window and is not recorded as a card-close pass.

### M4 — Close-out (Priority Medium)

1. Full affected family: `go test -race -count=1 -timeout 900s ./internal/cli -run '^(TestMemoryFold|TestReviewArchiveUpdate|TestReviewSequentialAbandonedTempOwnership).*$'` → PASS (B-1 scope; CI judges the rest; the selector's swept set re-verified by the §C.1 `-list` gate).
2. `go build ./...` exit 0.
3. Update progress.md §E.2/§E.3 evidence; verify every card-close AC command's output is on record. AC-MRR-008 is excluded from this obligation by the exception in §E, in the same words: **AC-MRR-008 post-close exception.** AC-MRR-008's observation is the Windows leg of `release-pr-multi-os.yml`, which the release process runs at the release window (`release/*` pull requests), after card close; card close is judged on the local gates, and AC-MRR-008's result is carried to the release window and is not recorded as a card-close pass.

### M5 — Post-close cells: cross-process, card-close span, cell-2 HOLDING (NOT RUN — operator decision `d-20261010T042423Z-722b`, option (b))

Status at plan-audit iteration 4: M1–M4 are recorded complete, and spec.md reads `status: completed`. Three repaired criteria need test code that HEAD does not contain. Checked at iteration 4: no process-spawning construct in the lock tests; the span test samples only the verb path; the cell-2 park point samples nothing. M5 ran only if the operator chose it. The two options were: **(a)** run M5 and reopen the SPEC through the amendment procedure (`completed → in-progress`, with a HISTORY `## Amendments` row recording `prior_completed_sha` equal to the prior close's `sync_commit_sha`, per `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix); or **(b)** accept the three cells as an explicit post-close exception, which leaves REQ-MRR-001's cross-process clause and the card-close span unverified and recorded as residual risk.

**Decision (2026-10-10).** The operator chose option (b) (board record `d-20261010T042423Z-722b`, operator first-hand, selected option "D2-B + M5-(b)"). M5 is not run. The SPEC stays `completed`, and no amendment procedure is opened. The three cells are accepted as a post-close exception: AC-MRR-010 stays unadopted, AC-MRR-009 is judged on the verb path only, and AC-MRR-006 is judged without the cell-2 precondition. The unverified clauses are recorded in spec.md §5 (R-2 cross-process; R-3 card-close entry point). Steps 1–4 below are kept as the record of what option (a) would have run. None of them is executed, and no test file is edited for them.

1. **AC-MRR-010 (cross-process, fold level), in `fold_store_lock_test.go`.** Step 1 is the RED-now: on the pinned pre-fix tree `84bdf072c` (POSIX), the parent holds `<store>/.moai-fold.lock` through a raw `flock` (no M2 API exists there), and the child fold applies while the parent holds it — the defect. Record the verbatim output, exit code, and tree SHA in progress.md §E.2 before any GREEN work. The GREEN is the criterion on HEAD.
2. **AC-MRR-009 (card-close subtest), in `memory_fold_test.go`.** Add a second subtest to `TestFoldStoreLockSpanHeldThroughApply` that drives `foldOnDoneStep` through the wiring fixture (`wireFixture`, `seedWireCard`, `runWireClose` in `memory_fold_wiring_test.go`) with a bound that cannot expire during the sample. Samples 2–4 are applyFold-internal seams and fire on both entry points (premise P2). Sample 1 (`mutateDisk`) fires only inside `newMemoryFoldCmd` (`memory_fold.go:183–187`), so this subtest takes sample 1 from the test immediately before it invokes the step; the expected outcome (acquired) is unchanged.
3. **AC-MRR-006 cell 2 (HOLDING precondition), in `memory_fold_wiring_test.go`.** Inside the parked `mutateBetweenWrites` callback, before the `lockHeld` signal, a non-blocking acquire by a second lock object must be refused. An acquisition there fails the cell.
4. Re-run the commands of AC-MRR-006, AC-MRR-009, and AC-MRR-010, and record their verbatim output in progress.md §E.2 (an M5 block). AC-MRR-010 is adopted once its RED-now (step 1) and its GREEN are both on record.

M5 would have changed test files only and no production code; every file it names is listed in §A.2.

## §G Anti-Patterns

- Do NOT "fix" by re-running `checkFoldUnchanged` closer to the rename — the window shrinks; it never closes (spec.md P6). The lock is the fix.
- Do NOT remove the lock file on release (D-3) — the removal/open race re-opens exactly the class being closed.
- Do NOT widen the seam contract: `mutateBeforeRename` is test-only, set by one test at a time, restored via `t.Cleanup` (B-3).
- Do NOT let the fix touch line classification, dedupe, budget, or preview rendering (Out of Scope).
- Do NOT run the whole `./internal/cli` suite locally as the judge (B-1).

## §H Cross-References

- spec.md §2 (REQ-MRR-001..008), §4 (the lock-path premise), and §5 (residual risk accepted at close); acceptance.md (AC-MRR-001..010), design.md (mechanism + rejected alternatives), decision-index.md (Q1..Q5, defaults applied).
- `SPEC-MEMORY-FOLD-BUDGET-001` (the write path's origin; REQ-MFB-004 ordering frozen by D-1), `SPEC-MEMORY-INDEX-FOLD-001` (no-code overlap).
- Card t1595 item 8 (same-store two-fold; Out of Scope here; the lock is adoptable there).
- `internal/sessionmsg/lock_unix.go` / `lock_windows.go` (the copied pattern), `internal/cli/memory_fold.go:635-649` (the window).
