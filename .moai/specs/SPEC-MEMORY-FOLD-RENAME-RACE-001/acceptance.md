# acceptance.md — SPEC-MEMORY-FOLD-RENAME-RACE-001

Acceptance criteria for the fold rename-window lost-update fix. Every criterion is a mechanically checkable command with an expected outcome; race-window criteria are judged by repeated runs with the race detector (premise P5 — never a single green run). Verification commands are single invocations (no pipes, no chaining).

## §1 The criteria

### AC-MRR-001 — The loss is reproduced before the fix (RED cell of the two-cell pair)

- **Given** the pre-fix tree (baseline `2aab5f797` + the M1 test-only seam commit), **When** `TestFoldRenameWindowConcurrentWriter` runs — the seam injects a concurrent writer's publish into the window between the last byte check and `os.Rename` — **Then** the fold returns success while the writer's line is gone from MEMORY.md, and the test FAILS on every iteration.
- **Command**: `go test -race -count=10 ./internal/cli -run '^TestFoldRenameWindowConcurrentWriter$'`
- **Expected now (pre-fix)**: exit non-zero; failure text names the absent writer line; 10/10 iterations red.
- **Status at plan phase**: PENDING EXECUTION — this cell is executed and its verbatim output + exit code + tree SHA recorded in progress.md §E.2 during run M1. The plan forbids M2 from starting before this observation is on record.
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
- **Command**: `go test -race -count=1 ./internal/cli -run '^TestFoldOnDoneContentionAbandonsWithoutWrite$'`
- **Expected**: exit 0.

### AC-MRR-007 — The fold family stays green (regression surface: t1502 fold 계열 + MEMORY.md index)

- **Given** the fixed tree, **When** the affected family runs (fold + review family — the callers of `applyFold`/`atomicWriteFoldFile` and their regressions), **Then** every test passes under the race detector.
- **Command**: `go test -race -count=1 -timeout 900s ./internal/cli -run '^Test(Fold|Review).*$'`
- **Expected**: exit 0; the baseline for this exact command was measured at plan phase on `2aab5f797` and recorded in progress.md §E.1 — green-after must equal green-before.
- **Full-suite note**: the whole `./internal/cli` suite is judged by CI (plan.md B-1); `go build ./...` exit 0 accompanies this criterion.

### AC-MRR-008 — Windows parity compiles

- **Given** the two build-tagged lock files, **When** the windows target builds, **Then** the build exits 0 with no build-tag or API errors.
- **Command**: `GOOS=windows go build ./internal/cli/...`
- **Expected**: exit 0.

## §2 Traceability

| REQ (spec.md §2) | AC |
|---|---|
| REQ-MRR-001 (whole-apply span) | AC-MRR-002, AC-MRR-004 |
| REQ-MRR-002 (one per store, both surfaces) | AC-MRR-004, AC-MRR-005 |
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
- The close-path step abandoned WHILE holding the lock: `defer release()` runs on the step's return (including the abandonment return paths) — covered by AC-MRR-006's recorder assertions.

## §4 Quality gates and Definition of Done

- TRUST 5: Tested (every REQ mapped, race criteria at `-count=10`), Readable/Unified (gofmt, package style), Secured (the lock file is inside the store dir, mode 0644, no secret surface), Trackable (conventional commits carrying the SPEC id and card id).
- DoD: AC-MRR-001's RED observed and recorded → AC-MRR-002..008 green with outputs on record → AC-MRR-007's post-fix green equals the pre-recorded baseline → progress.md §E.2/§E.3 populated.
- Indirect verification: none required beyond the family command — CI is the full-suite judge (plan.md B-1).
