# SPEC-HARNESS-RETENTION-HARDEN-001 — Implementation Plan (card t1432)

Derived from `spec.md` (the SSOT). Version 0.3.0, revised for plan-audit iteration 3 (the last). Milestones are ordered by decision-reversibility: the ones most likely to change come first (log-rewrite semantics, state-file path semantics), the mechanical and documentation steps last; M0 is a process step that must precede every change. Priority labels and phase ordering only; no time estimates.

**Amendment 0.4.0 [0.4.0].** The spec is amended in place (`spec.md` `### Amendments`): the heal check-then-remove window is closed with a heal lock (D4.c option C, REQ-HRH-005 rewritten, REQ-HRH-016 new). This plan gains milestones M7 to M10 below M6 and the notes marked `[0.4.0]`; M0 to M6 and every unmarked line are the 0.3.0 text, kept as history. M7 is a process step like M0 (the observed RED tests in their own commit, before any implementation commit), so it precedes M8 although M8 holds the decisions most likely to change; within the amendment the order is otherwise by reversibility (M8 the heal-lock decisions, M9 test debt, M10 verification).

## §A Context

- [0.4.0] Amendment base: tree `7639c04c1` on `WT-harness-retention-debt` (the card's run and sync-complete tree). Further production edits: two new build-tagged files in `internal/harness` (`retention_heal_unix.go`, `retention_heal_windows.go`) and a few lines in `retention.go` (`healStateEntry` and the constants); one new test file (`retention_heallock_test.go`) and a path guard in one existing test of this card (`retention_owner_test.go`); one `.gitignore` line. Still untouched: `internal/lockfile`, `observer.go`, the four pre-lock functions.
- Base: develop `1e2151a38` (code-identical to the plan commit `db6d88a2a`), branch `WT-harness-retention-debt`, worktree `.moai/worktrees/t1432`.
- Production edits: `internal/harness/retention.go`, plus two build-tagged owner-check files in the same package (a `//go:build !windows` file and a `//go:build windows` twin defining the same symbol). Test edits: `internal/harness/retention_killed_test.go` (N1) plus five new test files in the same package: `retention_statepath_test.go`, `retention_tail_test.go`, `retention_stampbytes_test.go`, `retention_stampwrite_test.go`, `retention_owner_test.go`.
- Evidence: `.moai/reports/t1432/red-baseline.md` and `.moai/reports/t1432/verdict.md`. The path is ignored by `.gitignore:235`; this repository tracks card evidence there by force-add (`git ls-files .moai/reports/t1425` lists tracked files), so every `.moai/reports/t1432/*` file is committed with `git add -f`. The tracked test files, not the ignored report, are the ordering witness (M0).
- Defaults chosen in `spec.md` §B: D1 disclosure (F5), D2 record only (F6), D3 tail-carry (event loss), D4 replace-on-fault restricted to entries the current user owns, with a stderr warning and conditional removal, D5 a parameter seam for the stamp-write pin and the one allowed field for the owner lookup. No milestone below touches `internal/lockfile` or `observer.go`.
- Windows: no runtime. `GOOS=windows go build` and `go vet` (which compiles test files) on the two packages are the only Windows verification.

## §B Known issues and open items carried into the run phase

| # | Item | State |
|---|---|---|
| B1 | F5 and F6 are "not reproduced, not measured"; the plan makes no code repair for them. | Settled by the leader's relayed verdict; the non-default options stay in `spec.md` §B as operator-held. |
| B2 | N1 fix mechanism (non-blocking read open vs a timer that unblocks by opening the write side). | Open: run phase chooses and records which, after observing the pre-fix hang (M0). The non-blocking form can race a pruner that has not reached its write-open. |
| B3 | Whether a concurrent `O_APPEND` write can be seen half-complete at the size reading. | Not observed. M1 tests the boundary rule deterministically; no kernel claim. |
| B4 | Operator-held options still unselected: D1-C, D2-B and C, D3-A, D4.a-B. | If any is selected, this SPEC is amended (`spec.md` §B and the affected REQs) before M0; it is not a run-phase decision. The Q4 restriction, the event-loss default and the one allowed test-only field are answered by the relayed verdicts. |
| B5 | The draft tests and mutant files behind the RED-now ledger are hoisted to `.moai/reports/t1432/red-now-drafts/` (tracked); the overlay JSON files there still name the plan phase's session scratch copies, and the FIFO-hang probe files (ledger E-025) are untracked. | Leader decision: regenerate the overlay JSON files against the committed sources and `git add -f` the probe files (`acceptance.md` Gap G-3); until then the ledger commands run only where the scratch copies exist. |
| B6 | The M0 commit leaves the branch with failing tests by design. | Branch-local; the leader's merge puts a green tip on develop. A green-history variant (baseline artifact first, tests with the fix) would lose the tracked-test witness and is not recommended. |
| B7 | Per-PR CI runs the `test` job on `ubuntu-latest` only (`.github/workflows/ci.yml:125`); macOS and Windows test legs run at release time (`release-pr-multi-os.yml:98`, READ). | Every probe in this plan was observed on darwin; Linux behaviour of the read-only-handle seam, the FIFO tests and the stderr capture is first exercised by CI. The local Windows vet (AC-HRH-011) is the only per-card Windows guard. |
| B8 | Force-adding `.moai/reports/t1432/*` conflicts in spirit with the `.gitignore:235` comment ("local-only artifacts … never on the remote"). | Precedent: 1589 tracked files under `.moai/reports`, 8 of them t1425's. The leader has accepted the force-added evidence (relayed 2026-10-03) and decides whether the plan branch's evidence files are pushed. |
| B9 | A FIFO at the state path hangs the lock-free pre-check (OBSERVED, ledger E-025); REQ-HRH-008 forbids a call added to that path. | Recorded, not repaired (`spec.md` §E, §F). If the leader wants it repaired, `spec.md` is amended first (REQ-HRH-008 and a new criterion) before M0; it is not a run-phase decision. |
| B10 [0.4.0] | The heal-lock wait bound (2 s) and poll interval (10 ms) are engineering choices: derived from the 5 s hook timeout and the t1425 lane's 1.79 s prune, not measured under load. | Open (decision-index Q9). The run phase uses the `spec.md` §B D4.c default and records it; a different figure goes back through a spec amendment, not a run-phase edit. |
| B11 [0.4.0] | The existing swap test `TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal` has an owner-check stand-in that renames a fresh file over whichever path it is asked about; once the heal lock asks the same owner check about the heal-lock path, the stand-in would act on that path too. | Settled in the design: M7 adds a one-line path guard (act only on the state path) to that stand-in; it should stay green at base and that is not observed until M7; it is the only edit to a test that exists at this tree. |
| B14 [0.4.0] | The amendment edits `spec.md`, so the plan-artifact hash changes and any cached plan-auditor PASS verdict is invalid (`spec-workflow.md` § Amendment as cache-invalidating event). | One cold delta plan-audit, scoped to the amended passages, runs before M7; if it fails, the run does not start. `/moai run` Phase 1 re-executes on the changed hash. |
| B12 [0.4.0] | The drafts, the probes and the mutant copies behind ledger E-032 to E-046 live in the session scratch directory (Gap G-9). | The leader hoists them to a tracked evidence path or M7 does, regenerating the overlay JSON files, before the scratch directory is purged. |
| B13 [0.4.0] | The CHANGELOG sentence "narrows but does not close" and the record `residual-risk-removal-window.md` describe the 0.3.0 close. | Not this SPEC's run phase: the amended sync (manager-docs) updates the sentence and the leader decides the record's status (Definition of Done 13). |

## §C Pre-flight (run phase, before M0)

Re-read branch and commit state (`git rev-parse --short HEAD`, `git branch --show-current`) and run the parallel-session divergence check before the first commit. Take the package verification slot lease the t1425 lane used for the `internal/harness` test runs, and release it before reporting. Scope every test run to `./internal/harness/` (and `./internal/lockfile/` for the unchanged-contract check); push and let CI run the full suite. Scrub the kanban environment in the same invocation as each test run (`unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ...`), because the harness tests read the environment. Build the project tool from the tree before citing any `moai spec lint` result and state its commit next to the tree HEAD (`verification-claim-integrity.md` §2.2).

## §D Constraints

Restated from `spec.md` §D: baseline-first commit ordering witnessed by tracked tests; Windows by build/vet only, with the build-tag clause of REQ-HRH-012; no work added to the append path or the lock-free pre-check; no new key, exported symbol or dependency (one unexported test-only field, spent on the owner lookup); `internal/lockfile` and `observer.go` byte-identical.

## §E Self-verification (what the run agent must be able to show)

- Every RB criterion in `acceptance.md` has its RED-now ledger cell (plan phase) and a second RED observed in M0 with the tree-resident test. Per criterion, the M0 record carries four fields — the single-invocation command, its verbatim stdout, its exit code as its own field, and the tree SHA from `git rev-parse --short HEAD` — and a fifth, the milestone that flips it. A criterion whose M0 RED differs from its ledger cell, or is missing, is a Gap, not a PASS.
- Every RG criterion is adopted by the mutation run its green-path cell names, each mutant observed failing the criterion's test and then removed; the verdict pastes the failing output.
- Every `go test -run` result is read after its swept count (`acceptance.md` swept-count requirement): the `--- PASS` line for each named test, never `[no tests to run]` or `--- SKIP`.

## §F Milestones

### M0 — Observed RED baseline commit (Priority High; own commit T, precedes every fix commit)

Goal: make the ordering witnessable from the commit graph (`verification-claim-integrity.md` §2.3) with tracked artifacts, and keep the baseline commit compilable.

1. Author only tests that compile against the unmodified package, using the public API and the existing helpers: `retention_statepath_test.go` (AC-HRH-001, -002, and AC-HRH-003 case b), `retention_tail_test.go` carrying `//go:build !windows` (AC-HRH-007 and AC-HRH-008 cases b and c), and `retention_stampbytes_test.go` (AC-HRH-013). Do **not** add `retention_owner_test.go` (AC-HRH-004, -005, -006: they need the owner field, the owner check and the heal helper) or `retention_stampwrite_test.go` (AC-HRH-014: it needs the extracted locked phase); a test that names an absent symbol breaks the build of the whole package and turns every other RED into a build failure.
2. Confirm the package builds with the new files: `go vet ./internal/harness/` exits 0 and `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` exits 0 on the commit's tree.
3. Run each new test against the unmodified `retention.go`, without an overlay, and record per criterion the command, verbatim stdout, exit code and tree SHA in `.moai/reports/t1432/red-baseline.md`. Expected: AC-HRH-001, -002, -007 and -008 (b, c) fail; AC-HRH-003 (b) and AC-HRH-013 pass at base (a pin and a test-gap criterion).
4. Re-observe the three mutants with overlays outside the tree: mD (replace `err = r.prune(retentionDays, now)` with `err = nil`) under `go test -count=1 -timeout 40s -run '^TestPruneStamp_StampExistsBeforeTheWork$'` (AC-HRH-012, the `panic: test timed out` output); mB (remove the `Truncate(0)` block in the stamp writer) with the whole package, then with the new mB test (AC-HRH-013, survivor then kill); mC (replace the stamp-write error check with `_ = writeStamp(sf, now)`) with the whole package (AC-HRH-014, survivor). The plan-phase overlay files in `acceptance.md` E-006 to E-009 are the model.
5. Record the AC-HRH-010 baseline greps (E-010, E-011) and the card-id guard count.
6. Commit the three test files and `red-baseline.md` (`git add -f .moai/reports/t1432/red-baseline.md`, tests staged by explicit pathspec) in one commit T whose subject names the card, e.g. `test(SPEC-HARNESS-RETENTION-HARDEN-001): observed RED baseline for retention hardening (card t1432)`.

Exit (evaluated on the card branch, before the merge into develop): `git log --diff-filter=A --format=%h -- internal/harness/retention_statepath_test.go` names T, and for every commit F listed by `git log --reverse --format=%h 1e2151a38..HEAD -- internal/harness/retention.go`, `git merge-base --is-ancestor T F` exits 0.

- The listing is bounded to the card's own range. The unbounded form lists seven commits that all predate the base (ledger E-021), and T, a descendant of the base, can never be their ancestor, so an unbounded check cannot pass.
- An empty bounded listing means no production commit exists yet. At M0 itself it is expected to be empty (ledger E-022); it is never read as a pass, and it is a Gap if it is still empty at M6.
- The check can fail, which is what makes it a check: the bounded form lists a hash on a range that holds a change (ledger E-023), and `git merge-base --is-ancestor` exits 1 on a deliberately mis-ordered pair, a later commit as the first argument (ledger E-024).

### M1 — Event loss: tail-carry and terminator (Priority High; REQ-HRH-006, -007, -008; AC-HRH-007, -008, -009)

The decision most likely to change, because it alters what the log rewrite writes. Approach (WHAT, not the final code): the prune opens the log, takes its size from the open handle, and reads exactly that prefix; only whole newline-terminated lines inside the prefix are classified, and the byte offset after the last of them is recorded. Immediately before the rename it reads whatever the log holds past that offset, once, and appends those bytes verbatim to the temporary replacement file, adding one newline when the tail is non-empty and does not end with one, then renames. A final line without a terminator, and any line whose write was in flight at the measurement, is wholly in the tail, never split between the classified part and the tail.

- Touches `partitionEvents` (returns the classified offset) and `overwriteWithEvents` (takes the offset and the log path, appends the tail and the terminator). The pre-lock functions (`PruneStaleEntries`, `readStamp`, `readStampFile`, `stampIsFresh`) are not edited (AC-HRH-009).
- Keep the no-late-event output byte-identical to today's for a log whose last line is terminated (AC-HRH-008 case a pins this with the existing unparsed-lines and nothing-stale tests); the one stated difference is a final unterminated stale line, archived one interval later.
- Do not touch `observer.go` or take any new lock (REQ-HRH-008, AC-HRH-009).
- Mutation runs, recorded in `verdict.md`: a verbatim tail copy with no terminator (fails AC-HRH-008 case b), classifying the final unterminated line (fails case c), dropping the tail copy (fails AC-HRH-007).

### M2 — State-file path hygiene: F4, F7, ownership, heal safety, warning (Priority High; REQ-HRH-001..005; AC-HRH-001..006)

A single helper returns the opened, lockable state file. Approach (WHAT):

1. Inspect the path without following links (`os.Lstat`). Absent: create exclusively (never follows a link, fails if the path exists); an "already exists" result means another process created it first, so inspect again. Regular file: open read-write without create, then compare the opened file's identity with the inspection and refuse to proceed (inspect again) on a mismatch, before any stamp write. Symbolic link: never opened. Any other kind: open as today (a directory fails the open), keeping today's errors; a FIFO never reaches this step, because the lock-free pre-check blocks on it first (pre-existing, not repaired: `spec.md` §E).
2. For a symbolic link, or a regular file whose open fails with a permission error, ask the owner check whether the current user owns the entry (the link's own record, not its target's). Not owned, or owner undeterminable: leave the entry, return an error naming the path, and write the single warning line (`[WARN] harness/retention: …`, the form at `internal/harness/safety/frozen_guard.go:124`) to standard error. Owned: heal.
3. Heal: re-inspect the entry and remove it only if it is still the inspected one (same file identity, same type, same permission mode, same modification time); otherwise inspect again without removing. Create the replacement exclusively. At most three inspections in total, then skip with an error. The lock, the stamp read and the stamp write run on the opened file as today. Split the heal helper into an inspection step and a removal step, so that AC-HRH-006 can interpose a stand-in healer between them.
4. The owner check is one function-valued unexported field on `Retention`, installed by `NewRetention`. The real check takes the entry's path and reads the owner itself with a no-follow stat (the link's own record, REQ-HRH-001), so a lookup that follows links cannot hide in the caller; it compares that owner id with the effective user id and lives in the `//go:build !windows` file, with a `//go:build windows` twin that reports "not owned" for every entry. This field is the single test-only field the operator allowed (§B D5); tests replace it to reach the foreign-owned branch.
5. Tests, all in `retention_owner_test.go` carrying `//go:build !windows` unless noted: AC-HRH-004 (serial, swaps `os.Stderr`, injects "not owned" for a symbolic link and for a regular file), AC-HRH-005 (the real check on a file the test made, on `/`, on a link the test made pointing at `/`, and on a root-owned symbolic link from a fixed candidate list; skipped as uid 0, and the last step skipped where no candidate qualifies), AC-HRH-006 (the heal helper with a stand-in healer that renames a fresh file over the inspected link). AC-HRH-001, -002 and -003 (b) already landed in M0 and flip here. The existing directory-at-state-path and read-only-directory tests stay green unmodified.
6. Mutation runs, recorded in `verdict.md`: AC-HRH-004 owner result ignored, removal before the ownership check, warning omitted; AC-HRH-005 check always "owned", check reading the owner of the entry a link points at (a Stat-based read), check treating every symbolic link as owned; AC-HRH-006 unconditional removal; AC-HRH-003 a heal that removes a directory, and a heal that ignores the removal failure.
7. Run `GOOS=windows go build` and `go vet` on both packages after this milestone. The time-of-check gap between the inspection and the open is closed for writes by the post-open identity check and is not pinned by a test (`spec.md` §F); say so in the verdict.

### M3 — N2: stamp-write pin and the locked-phase seam (Priority Medium; REQ-HRH-014, -015; AC-HRH-013, -014)

- mB: the test landed in M0 with no production change; M3 re-runs the mutant kill and records it.
- mC: extract the locked phase of `pruneExclusive` (fresh-clock stamp re-check, stamp write, prune, orphan sweep) into a function that takes the already-opened state file and the retention days; `pruneExclusive` opens the file through the M2 helper, locks it and calls the function. The new test in `retention_stampwrite_test.go` (`//go:build !windows`) opens a state file read-only, calls the function, and asserts a non-nil error, a byte-identical log and no archive directory. Existing tests cover the move. Mutation run: mC (the stamp-write error ignored) fails the test, closing Gap G-2.
- If CI shows on Linux that a read-only handle's truncate does not fail, return a blocker report: `spec.md` §B D5 option B (the field) is then the fallback and the owner lookup needs another pin.

### M4 — N1: fail fast (Priority Medium; REQ-HRH-013; AC-HRH-012)

Edit `drain` in `retention_killed_test.go` so it can never block on a pruner that did not reach its archive step: either open the read side non-blocking, or run the blocking open in a goroutine and, on a deadline, unblock it by opening the FIFO write side non-blocking (which succeeds only if a reader is blocked), then fail with a message naming the missing archive step. Decide after re-observing the hang in M0; record the choice and the reason in `verdict.md`. The file already carries `//go:build !windows` and keeps it.

### M5 — Disclosure comments: residual window, F5, F6 (Priority Low; REQ-HRH-009, -010, -011; AC-HRH-010)

Comment-only edits in `retention.go`: the `PruneStaleEntries` doc comment (above the function, so AC-HRH-009's function-scoped history check stays empty) and the `pruneExclusive` doc comment, carrying the sentinel phrases of AC-HRH-010 (`residual window`, `no cross-process exclusion`, `burst of hook processes`, `F5: not reproduced, not measured`, `lock waiters block with no timeout`, `5 s hook timeout`, `appended before the wait`, `F6: not reproduced, not measured`) and removing the sentence the baseline carries (`events other hooks append in that window are lost`). State technical facts only; no card id and no design-question history (that point lives in `decision-index.md` Q2 and `progress.md`). Respect the `code_comments` setting (English). No behaviour change.

### M6 — Verification and close-out (Priority Medium)

- `go test -race -count=1 -v ./internal/harness/` and `go vet ./internal/harness/ ./internal/lockfile/` exit 0; `go test -count=20 -run '^(TestPruneStamp_StampExistsBeforeTheWork|TestPruneConcurrentProcessesSingleRewrite)$' ./internal/harness/` exit 0.
- `GOOS=windows go build ./internal/harness/ ./internal/lockfile/` and `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` exit 0.
- `git diff --quiet 1e2151a38 -- internal/lockfile internal/harness/observer.go` exit 0, and the function-scoped `git log -L` check of AC-HRH-009 prints nothing.
- Re-run mB, mC and mD and the M1 and M2 mutants, and paste the failing outputs into `verdict.md`.
- State the judging build for every project-tool measurement next to the tree HEAD (`verification-claim-integrity.md` §2.2).
- List every Gap: skipped tests with platform and uid, unobserved Linux and Windows behaviour, and the unmeasured heal burst. The completion report's Gaps section states `Windows runtime not observed`, and one line names the intermediate red commit T by SHA (`spec.md` §D).
- Evaluate the M0 Exit ordering check on the card branch before the merge into develop; an empty bounded listing at this point is a Gap, not a pass.

## §F.1 Amendment milestones M7 to M10 [0.4.0]

Run after the delta plan-audit of the amended artifacts (B14). Same discipline as M0 to M6: one milestone per delegation, commits by explicit pathspec, tests scoped to `./internal/harness/` under the slot lease, the kanban environment scrubbed in the same invocation, every `go test -run` read after its swept count.

### M7 — Observed RED baseline for the heal lock (Priority High; own commit T2; precedes every heal-lock production commit)

Goal: make the amendment's ordering witnessable from the commit graph (`verification-claim-integrity.md` §2.3), as M0 did, and keep the baseline commit compilable.

1. Author `internal/harness/retention_heallock_test.go` with `//go:build !windows`: the tests of AC-HRH-006 (b), (c), (d), AC-HRH-015 and AC-HRH-016, compiling against the unmodified package. They name the heal-lock file by the literal `.prune-heal`, take the lock themselves with `syscall.Flock`, and use only existing symbols and helpers (`captureStderr`, `blockedPruner`, `writeStaleLog`, `stampSuffix`); a test that names an absent symbol (`pruneHealSuffix`, a new helper) breaks the build of the whole package and turns every other RED into a build failure. The drafts behind E-036 to E-039 are the model; the AC-HRH-015 test carries a hard cap (8 s) and the AC-HRH-016 test a 4.5 s cap per case, so an unbounded wait or a hang fails the test instead of hanging the run.
2. Add the path guard (act only on the state path) to the owner-check stand-in of `TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal` in `retention_owner_test.go` (B11).
3. Confirm the package builds: `go vet ./internal/harness/` and `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` exit 0 on the commit's tree.
4. Run each new test against the unmodified `retention.go`, without an overlay, and record per criterion the command, verbatim stdout, exit code and tree SHA in `.moai/reports/t1432/red-baseline-amend.md`. Expected: AC-HRH-006 (b), AC-HRH-015 and AC-HRH-016 fail; (c) and (d) pass at base (regression-guards; (d) vacuous there); the existing swap test and AC-HRH-006 (a) still pass. A criterion whose M7 RED differs from its ledger cell (E-036 to E-038) is a Gap.
5. Record the baselines E-041 (the ninth sentinel) and E-042 (`.gitignore`), and hoist the drafts and the overlay-regenerated probe to a tracked evidence path (B12).
6. Commit the test files and `red-baseline-amend.md` (`git add -f .moai/reports/t1432/red-baseline-amend.md`, tests staged by explicit pathspec) in one commit T2, e.g. `test(SPEC-HARNESS-RETENTION-HARDEN-001): observed RED baseline for the heal lock (card t1432)`.

Exit (evaluated on the card branch, before the merge into develop): Definition of Done 9, with its controls (E-046). The listing is bounded to `7639c04c1..HEAD`; at M7 it is expected to be empty and is never read as a pass.

### M8 — The heal lock (Priority High; REQ-HRH-005, -010, -016; AC-HRH-006 b to d, -010 ninth sentinel, -011, -015, -016)

The decisions most likely to change (the wait bound, the inspection and open form of the heal-lock file, what the owner check is asked). Approach (WHAT; names are suggestions, the criteria bind behaviour):

1. In `retention.go`: constants `pruneHealSuffix` (`.prune-heal`, beside `pruneStateSuffix`), `pruneHealWait` (2 s) and `pruneHealPoll` (10 ms), each with a doc comment stating the technical fact (the 5 s hook timeout, why 2 s) and no card id or design history.
2. `retention_heal_unix.go` (`//go:build !windows`): one helper that, given the heal-lock path and the owner-check function, returns a release function or an error naming the path. It inspects the path with `os.Lstat` (at most three inspections); creates it exclusively with mode 0600 when absent (`O_CREATE|O_EXCL|O_RDWR|O_NOFOLLOW|O_NONBLOCK`, so it never follows a link and never truncates); opens an existing regular file read-write with `O_NOFOLLOW|O_NONBLOCK` and no `O_CREATE` or `O_TRUNC`; requires the opened file to be regular and `os.SameFile` with the inspection; asks the owner check about the path; then polls `syscall.Flock(LOCK_EX|LOCK_NB)` every `pruneHealPoll` until `pruneHealWait` on the real clock (`EWOULDBLOCK` and `EINTR` mean poll again, any other error means unusable). It never removes, replaces, truncates or chmods the file. A symbolic link, a directory, a FIFO or any other non-regular entry is never opened (the decisive FIFO rule, `spec.md` §B D4.c item 6).
3. `retention_heal_windows.go` (`//go:build windows`): the same unexported symbol, returning a no-op release and no error, creating no file; its comment states that the heal lock gives no exclusion on Windows, that the window of option A remains there, and that the Windows runtime is not observed.
4. In `healStateEntry`, after the ownership check passes: acquire the heal lock; on an error write the one warning line (`[WARN] harness/retention: …` naming the heal-lock path, the form of the state-entry warning) and return the error naming that path, so `openStateFile` returns it and the prune is skipped before any stamp is written; otherwise run `removeStateEntryIfUnchanged` while holding the lock and release the lock before returning, so it is never held when `openStateFile` creates the replacement or when `pruneExclusive` locks the state file. `openStateFile`'s absent and healthy-regular branches are not edited (AC-HRH-006 c).
5. Disclosure comments (English, no card id, no design history): `healStateEntry` and the `@MX:REASON` of `openStateFile` say the removal runs under the heal lock; the `PruneStaleEntries` or `pruneExclusive` doc comment carries the sentence `heal lock gives no exclusion on Windows` (REQ-HRH-010; the comment goes above the function so AC-HRH-009's function-scoped check stays empty). Update the `@MX:REASON` that today states the concurrent-healer guarantee without the heal lock.
6. `.gitignore`: add `.moai/harness/usage-log.jsonl.prune-heal` after the `.prune-state` line (`.gitignore:352`); the template `.gitignore` is not touched (E-042).
7. Verify: M7's tests flip (AC-HRH-006 b, c, d, AC-HRH-015, AC-HRH-016 each `--- PASS`, swept count read); the full package under `-race`; `GOOS=windows go build` and `go vet` of both packages (AC-HRH-011); AC-HRH-009's two commands; the ninth sentinel of AC-HRH-010 (count at least 1).
8. Mutation runs, recorded in `verdict.md`, each regenerated from the then-current `retention.go` (copies in the scratch directory are bound to `7639c04c1`, the N6 hazard): heal lock never taken; taken on the common path; held across the prune; open follows links; owner check skipped; hostile entry removed and recreated; unbounded wait; heal proceeding after the bound; no error or no warning. State that a lock released before the removal is not killed by a deterministic test (`acceptance.md` AC-HRH-006).
9. The race probe of Definition of Done 10, with its controls. Do not touch `observer.go`, `internal/lockfile` or the four pre-lock functions.

### M9 — Test debt N1, N4, N2 (Priority Low; optional, only if cheap; AC-HRH-006 e, AC-HRH-008 d)

Land the three draft tests of E-045 (the model) in files this card created: N1 and N4 in `retention_owner_test.go`, N2 in `retention_tail_test.go`. Each is adopted by its mutant killed, the mutants regenerated from the then-current files (E-044 shows the three survive the suite today). "Cheap" means the draft's shape suffices with no new seam or field; where one is needed, stop and carry the debt to the verdict as an accepted item (`sync-audit-delta.md` D4 to D6). The test commit needs no production change and may precede or follow M8.

### M10 — Verification and close-out of the amendment (Priority Medium)

- `go test -race -count=1 -v ./internal/harness/`, `go vet ./internal/harness/ ./internal/lockfile/` and the quality-gate runs of `acceptance.md` exit 0; the heal-lock tests under `-race -count=5`, AC-HRH-015 once.
- `GOOS=windows go build` and `go vet` of both packages exit 0; `git diff --quiet 7639c04c1 -- internal/lockfile internal/harness/observer.go` exit 0; the AC-HRH-009 `git log -L` form over `7639c04c1..HEAD` prints nothing.
- The mutation outputs of M8 and M9 pasted failing into `verdict.md`; the race probe and its two controls with their trial counts (Definition of Done 10); `git check-ignore` exit 0 for the heal-lock path.
- State the judging build for every project-tool measurement next to the tree HEAD (`verification-claim-integrity.md` §2.2); list every Gap (skipped tests with platform and uid, Linux and Windows unobserved, the unmeasured production frequency and the 2 s bound under load); `Windows runtime not observed` and the SHA of T2 in the completion report.
- Evaluate the Definition of Done 9 ordering check on the card branch before the merge into develop; an empty bounded listing at this point is a Gap.
- Hand off Definition of Done 13 (the CHANGELOG sentence, the status of the residual-risk record) to its owners.

## §G Anti-patterns to avoid

- Fixing F5 or F6 in code without a reproduction (the leader's narrowing): the disclosure is the repair.
- Taking a lock on the append path to close the event-loss window (D3 A is operator-held).
- Opening the state file with the same call and "handling" a symbolic link after the fact: the link must be dealt with before any open that could truncate.
- Deleting a state-path entry the current user does not own, or deleting one without re-checking that it is still the inspected entry.
- Treating a runtime skip as a build constraint: a FIFO test without `//go:build !windows` does not compile on Windows.
- A package-level mutable seam variable in a package whose tests run in parallel.
- Reading a `go test -run` exit 0 without its swept count.
- Claiming a Windows result from a build or vet run.
- Repairing the FIFO hang inside the lock-free pre-check on this card: REQ-HRH-008 forbids a call added to that path, and the condition is recorded, not repaired (`spec.md` §E).
- [0.4.0] Building the heal lock in `internal/lockfile` (REQ-HRH-012), or adding a try-lock there.
- [0.4.0] Taking the heal lock on the common path (an absent or healthy state entry), or holding it across the prune: it covers the re-inspection and the removal only.
- [0.4.0] Opening the heal-lock path through a link, with a blocking open, with truncation, or after skipping the owner check; removing, replacing or "healing" a damaged heal-lock entry.
- [0.4.0] Driving the wait from the injected `nowFn` clock instead of the real clock; a wait with no bound.
- [0.4.0] Presenting the audit's bare-helper probe as the post-fix evidence: the lock is taken by the helper's caller, so that probe still counts removals after the change by design; the evidence is the probe through the heal entry point with a lock-honouring swapper, with its two controls.
- [0.4.0] Claiming a Windows result from a build or vet run, or claiming the Windows window is closed.

## §H Cross-references

[0.4.0] `.moai/reports/t1432/sync-audit.md` (F1), `.moai/reports/t1432/sync-audit-delta.md` (D1 to D6), `.moai/reports/t1432/residual-risk-removal-window.md`; `.claude/rules/moai/development/spec-frontmatter-schema.md` (the `completed → in-progress (amendment)` transition); `.claude/rules/moai/workflow/spec-workflow.md` § Amendment as cache-invalidating event.

`spec.md` §A-§F; `acceptance.md`; `decision-index.md`; `.moai/reports/t1425/sync-audit.md` §3, `.moai/reports/t1425/sync-audit-delta.md` §3 and §1 Residual-risk, `.moai/reports/t1425/decision-records.md`; `.moai/reports/t1432/plan-audit.md` (iteration 1); `.moai/reports/t1432/plan-audit-iter2.md` (iteration 2); `.moai/specs/SPEC-AGENT-TEAM-RETIRE-001/spec.md` REQ-ATR-001; `.claude/rules/moai/core/verification-claim-integrity.md` §2.2, §2.3; `.claude/rules/moai/development/verification-completeness.md` §1.1, §2, §2.1.
