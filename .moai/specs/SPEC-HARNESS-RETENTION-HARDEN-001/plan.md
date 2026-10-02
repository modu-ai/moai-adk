# SPEC-HARNESS-RETENTION-HARDEN-001 — Implementation Plan (card t1432)

Derived from `spec.md` (the SSOT). Milestones are ordered by decision-reversibility: the ones most likely to change come first (log-rewrite semantics, state-file path semantics), the mechanical and documentation steps last. Priority labels and phase ordering only; no time estimates.

## §A Context

- Base: develop `1e2151a38`, branch `WT-harness-retention-debt`, worktree `.moai/worktrees/t1432`. Production edits: `internal/harness/retention.go` only. Test edits: `internal/harness/retention_killed_test.go` (N1) plus up to three new test files in the same package. Evidence: `.moai/reports/t1432/red-baseline.md` and `.moai/reports/t1432/verdict.md` (card evidence path convention, `AGENTS.md` §3).
- Defaults chosen in `spec.md` §B: D1 disclosure (F5), D2 record only (F6), D3 tail-carry (event loss), D4 replace-on-fault (F4+F7), D5 unexported seam field (N2 mC). No milestone below touches `internal/lockfile` or `observer.go`.
- Windows: no runtime. `GOOS=windows go build` and `go vet` on the two packages are the only Windows verification.

## §B Known issues and open items carried into the run phase

| # | Item | State |
|---|---|---|
| B1 | F5 and F6 are "not reproduced, not measured"; the plan makes no code repair for them. | Settled by the leader's scope narrowing; the non-default options stay in `spec.md` §B as operator-held. |
| B2 | N1 fix mechanism (non-blocking read open vs a timer that unblocks by opening the write side). | Open: run phase chooses and records which, after observing the pre-fix hang (M0). The non-blocking form can race a pruner that has not reached its write-open. |
| B3 | Whether a concurrent `O_APPEND` write can be seen half-complete at the size reading. | Not observed. M1 tests the boundary rule (last newline-terminated line) deterministically; no kernel claim. |
| B4 | The leader may lift D1-C, D2-B/C, D3-A to the operator before Kickoff. | If the operator selects one, this SPEC is amended (`spec.md` §B and the affected REQs) before M0; it is not a run-phase decision. |

## §C Pre-flight (run phase, before M0)

Re-read branch and commit state (`git rev-parse --short HEAD`, `git branch --show-current`) and run the parallel-session divergence check before the first commit. Take the package verification slot lease the t1425 lane used for the `internal/harness` test runs, and release it before reporting. Scope every test run to `./internal/harness/` (and `./internal/lockfile/` for the unchanged-contract check); push and let CI run the full suite. Scrub the kanban environment in the same invocation as each test run (`unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ...`), because the harness tests read the environment.

## §D Constraints

Restated from `spec.md` §D: baseline-first commit ordering; Windows by build/vet only; no work added to the append path; no new key, symbol or dependency; `internal/lockfile` and `observer.go` byte-identical.

## §E Self-verification (what the run agent must be able to show)

Each behaviour-changing AC in `acceptance.md` has its RED observed in M0 against the unmodified retention code, the verbatim output pasted into `red-baseline.md`, and the same commands re-run GREEN after the fix commit. A criterion whose RED is not observed in M0 is a Gap, not a PASS.

## §F Milestones

### M0 — Observed RED baseline (Priority High; own commit, precedes every fix commit)

Goal: make the ordering witnessable from the commit graph (`verification-claim-integrity.md` §2.3).

1. Add the failing tests for AC-HRH-001, -002, -003, -005 and the mutant-killing tests for AC-HRH-013, -014 (test files only; production code untouched). Run each against the unmodified `retention.go` and record verbatim command and output in `.moai/reports/t1432/red-baseline.md`.
2. Re-observe the N1 hang (AC-HRH-012): apply the audit's mutant mD (replace the `r.prune(...)` call with `return nil` after the stamp) as a temporary overlay or patch, run `go test -run '^TestPruneStamp_StampExistsBeforeTheWork$' -timeout 40s ./internal/harness/`, record the `panic: test timed out` output, and remove the mutation.
3. Re-observe the two surviving mutants (mB: remove `Truncate(0)` in `writeStamp`; mC: ignore the `writeStamp` error in `pruneExclusive`) against the existing suite, record `ok` results, remove the mutations. These are the REDs for AC-HRH-013 and -014 (a test gap's RED is the survivor).
4. Record the baseline grep counts for the comment ACs (AC-HRH-008, -009, -010): at base, `not reproduced, not measured` 0, `lock waiters block` 0, `5 s hook timeout` 0, `residual window` 0, and the sentence `events other hooks append in that window are lost` 1 (all observed during plan authoring on `retention.go`; re-observe in M0).
5. Commit `.moai/reports/t1432/red-baseline.md` together with the new test files in one commit whose subject names the card, e.g. `test(SPEC-HARNESS-RETENTION-HARDEN-001): observed RED baseline for retention hardening (card t1432)`. These tests fail on the branch at this commit by design; the intermediate commit is branch-local and is not what CI gates. If the leader prefers a green history, split: baseline artifact alone first, tests with the fix commit; the artifact then names the test source it ran.

Exit: `git log --reverse --format=%h -- .moai/reports/t1432/red-baseline.md` lists a commit that precedes every commit touching `internal/harness/retention.go`.

### M1 — Event loss: tail-carry (Priority High; REQ-HRH-004, -005, -006; AC-HRH-005, -006, -007, -008)

The decision most likely to change, because it alters what the log rewrite writes. Approach (WHAT, not the final code): the prune reads a measured prefix of the log, classifies only that prefix, and records the byte offset of the end of the last newline-terminated line it classified. Immediately before the rename, it reads whatever the log gained past that offset and appends those bytes verbatim to the temporary replacement file, then renames. A line whose write was in flight at the size reading is wholly in the tail, never split between the classified part and the tail. After the tail copy the pruner may re-check for further growth a bounded number of times; the final window is documented, not eliminated.

- Touches `partitionEvents` (returns the classified offset) and `overwriteWithEvents` (takes the offset and the log path, appends the tail). Update the `@MX:WARN` on `PruneStaleEntries` to the residual-window wording (REQ-HRH-006); the `@MX:REASON` keeps the Windows sentence (M5 rewrites it).
- Keep the no-late-event output byte-identical to today's (AC-HRH-006 pins this with the existing unparsed-lines and nothing-stale tests).
- Do not touch `observer.go` or take any new lock (REQ-HRH-005, AC-HRH-007).
- Test boundary rule: a log whose last line has no trailing newline at the moment of the prefix measurement ends up whole in the replacement exactly once (no duplicate, no split).

### M2 — State-file path hygiene: F4 + F7 (Priority High; REQ-HRH-001, -002, -003; AC-HRH-001..004)

A single helper opens the state file: inspect the path without following links (`os.Lstat`); a symbolic link is removed (`os.Remove` removes the link, not the target) and the path recreated with exclusive create; a regular file whose read-write open fails with a permission error is removed and recreated the same way; one retry bound if the exclusive create reports the path already exists (another process healed first); any other failure (a directory, a FIFO, a file in a non-writable directory) returns today's `retention: prune state open failed` error unchanged. The lock, the stamp read and the stamp write are unchanged and run on the opened file.

- `Lstat`, `Remove` and exclusive create are portable; no build tags.
- The existing directory-at-state-path and read-only-directory tests must stay green unmodified (AC-HRH-015): neither is a symbolic link nor a removable regular file.
- Permission-based tests skip when the effective uid is 0 or on Windows, following `TestPruneStamp_UnwritableStateSkipsPruneAndRecordSucceeds`.

### M3 — N2: stamp-write pins and the seam (Priority Medium; REQ-HRH-011, -012; AC-HRH-013, -014)

- mB test: pre-seed the state file with a nine-fractional-digit stamp older than the interval, run a prune whose clock value formats with fewer bytes (a whole second), and assert the state file's exact content equals the new stamp with no trailing bytes and parses fresh. No production change.
- mC test: add an unexported function-valued field on `Retention` (default: the real stamp writer, set in `NewRetention`; a nil value falls back to the real writer so a literal-constructed `Retention` still works). The test replaces it with a failing writer and asserts the prune returns an error and the log is byte-identical, and that the archive directory is not created. A package variable is rejected because it would race the package's parallel tests.
- If the leader selects D5 option B (no seam), AC-HRH-014 is withdrawn by amendment and mutant mC is recorded as a known survivor.

### M4 — N1: fail fast (Priority Medium; REQ-HRH-010; AC-HRH-012)

Edit `drain` in `retention_killed_test.go` so it can never block on a pruner that did not reach its archive step: either open the read side non-blocking, or run the blocking open in a goroutine and, on a deadline, unblock it by opening the FIFO write side non-blocking (which succeeds only if a reader is blocked), then fail with a message naming the missing archive step. Decide in M0 after observing the hang; record the choice and the reason in `red-baseline.md`.

### M5 — Disclosure comments: F5, F6, residual window (Priority Low; REQ-HRH-006, -007, -008; AC-HRH-008, -009, -010)

Comment-only edits in `retention.go` (the `PruneStaleEntries` `@MX:WARN`/`@MX:REASON` and the `pruneExclusive` `@MX:NOTE`), carrying the sentinel phrases `residual window`, `not reproduced, not measured` (F5 and F6 each), `lock waiters block` and `5 s hook timeout`, and removing the sentence the baseline carries (`events other hooks append in that window are lost`). Respect the `code_comments` setting (English). No behaviour change.

### M6 — Verification and close-out (Priority Medium)

- `go test -race -count=1 ./internal/harness/` and `go vet ./internal/harness/ ./internal/lockfile/` exit 0; `go test -count=20 -run '^(TestPruneStamp_StampExistsBeforeTheWork|TestPruneConcurrentProcessesSingleRewrite)$' ./internal/harness/` exit 0.
- `GOOS=windows go build ./internal/harness/ ./internal/lockfile/` and `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` exit 0.
- `git diff --quiet 1e2151a38 -- internal/lockfile internal/harness/observer.go` exit 0.
- Re-run the three mutants (mB, mC, mD) and show them killed or failing fast, outputs pasted into `verdict.md`.
- State the judging build for every project-tool measurement next to the tree HEAD (`verification-claim-integrity.md` §2.2).

## §G Anti-patterns to avoid

- Fixing F5 or F6 in code without a reproduction (the leader's narrowing): the disclosure is the repair.
- Taking a lock on the append path to close the event-loss window (D3 A is operator-held).
- Opening the state file with the same call and "handling" a symbolic link after the fact: the link must be dealt with before the open that truncates.
- A package-level mutable seam variable in a package whose tests run in parallel.
- Claiming a Windows result from a build or vet run.

## §H Cross-references

`spec.md` §A-§F; `acceptance.md`; `decision-index.md`; `.moai/reports/t1425/sync-audit.md` §3, `.moai/reports/t1425/sync-audit-delta.md` §3 and §1 Residual-risk, `.moai/reports/t1425/decision-records.md`; `.moai/specs/SPEC-AGENT-TEAM-RETIRE-001/spec.md` REQ-ATR-001; `.claude/rules/moai/core/verification-claim-integrity.md` §2.3.
