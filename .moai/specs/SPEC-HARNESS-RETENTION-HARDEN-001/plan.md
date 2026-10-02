# SPEC-HARNESS-RETENTION-HARDEN-001 — Implementation Plan (card t1432)

Derived from `spec.md` (the SSOT). Version 0.2.0, revised for plan-audit iteration 2. Milestones are ordered by decision-reversibility: the ones most likely to change come first (log-rewrite semantics, state-file path semantics), the mechanical and documentation steps last; M0 is a process step that must precede every change. Priority labels and phase ordering only; no time estimates.

## §A Context

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
| B5 | The draft tests and mutant files behind the RED-now ledger live in a session scratch directory. | Leader or orchestrator decision: hoist them into a card evidence path before the plan-audit re-run if the auditor must re-execute the ledger (`acceptance.md` Gap G-3). |
| B6 | The M0 commit leaves the branch with failing tests by design. | Branch-local; the leader's merge puts a green tip on develop. A green-history variant (baseline artifact first, tests with the fix) would lose the tracked-test witness and is not recommended. |
| B7 | Per-PR CI runs the `test` job on `ubuntu-latest` only (`.github/workflows/ci.yml:125`); macOS and Windows test legs run at release time (`release-pr-multi-os.yml:98`, READ). | Every probe in this plan was observed on darwin; Linux behaviour of the read-only-handle seam, the FIFO tests and the stderr capture is first exercised by CI. The local Windows vet (AC-HRH-011) is the only per-card Windows guard. |
| B8 | Force-adding `.moai/reports/t1432/*` conflicts in spirit with the `.gitignore:235` comment ("local-only artifacts … never on the remote"). | Precedent: 1589 tracked files under `.moai/reports`, 8 of them t1425's. The leader decides whether the plan branch's evidence files are pushed. |

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

Exit: `git log --diff-filter=A --format=%h -- internal/harness/retention_statepath_test.go` names T, and for every commit F listed by `git log --reverse --format=%h -- internal/harness/retention.go`, `git merge-base --is-ancestor T F` exits 0.

### M1 — Event loss: tail-carry and terminator (Priority High; REQ-HRH-006, -007, -008; AC-HRH-007, -008, -009)

The decision most likely to change, because it alters what the log rewrite writes. Approach (WHAT, not the final code): the prune opens the log, takes its size from the open handle, and reads exactly that prefix; only whole newline-terminated lines inside the prefix are classified, and the byte offset after the last of them is recorded. Immediately before the rename it reads whatever the log holds past that offset, once, and appends those bytes verbatim to the temporary replacement file, adding one newline when the tail is non-empty and does not end with one, then renames. A final line without a terminator, and any line whose write was in flight at the measurement, is wholly in the tail, never split between the classified part and the tail.

- Touches `partitionEvents` (returns the classified offset) and `overwriteWithEvents` (takes the offset and the log path, appends the tail and the terminator). The function body of `PruneStaleEntries` is not edited (AC-HRH-009).
- Keep the no-late-event output byte-identical to today's for a log whose last line is terminated (AC-HRH-008 case a pins this with the existing unparsed-lines and nothing-stale tests); the one stated difference is a final unterminated stale line, archived one interval later.
- Do not touch `observer.go` or take any new lock (REQ-HRH-008, AC-HRH-009).
- Mutation runs, recorded in `verdict.md`: a verbatim tail copy with no terminator (fails AC-HRH-008 case b), classifying the final unterminated line (fails case c), dropping the tail copy (fails AC-HRH-007).

### M2 — State-file path hygiene: F4, F7, ownership, heal safety, warning (Priority High; REQ-HRH-001..005; AC-HRH-001..006)

A single helper returns the opened, lockable state file. Approach (WHAT):

1. Inspect the path without following links (`os.Lstat`). Absent: create exclusively (never follows a link, fails if the path exists); an "already exists" result means another process created it first, so inspect again. Regular file: open read-write without create, then compare the opened file's identity with the inspection and refuse to proceed (inspect again) on a mismatch, before any stamp write. Symbolic link: never opened. Any other kind: open as today (a directory fails, a FIFO fails at the lock), keeping today's errors.
2. For a symbolic link, or a regular file whose open fails with a permission error, ask the owner check whether the current user owns the entry (the link's own record, not its target's). Not owned, or owner undeterminable: leave the entry, return an error naming the path, and write the single warning line (`[WARN] harness/retention: …`, the form at `internal/harness/safety/frozen_guard.go:124`) to standard error. Owned: heal.
3. Heal: re-inspect the entry and remove it only if it is still the inspected one (same file identity, same type, same permission mode, same modification time); otherwise inspect again without removing. Create the replacement exclusively. At most three inspections in total, then skip with an error. The lock, the stamp read and the stamp write run on the opened file as today.
4. The owner check is one function-valued unexported field on `Retention`, installed by `NewRetention`. The real check lives in the `//go:build !windows` file (the entry's owner id from its platform stat record compared with the effective user id) with a `//go:build windows` twin that reports "not owned" for every entry. This field is the single test-only field the operator allowed (§B D5); tests replace it to reach the foreign-owned branch.
5. Tests, all in `retention_owner_test.go` carrying `//go:build !windows` unless noted: AC-HRH-004 (serial, swaps `os.Stderr`, injects "not owned" for a symbolic link and for a regular file), AC-HRH-005 (the real check on a file the test made and on `/`, skipped for the `/` step as uid 0), AC-HRH-006 (the heal helper with a stand-in healer that renames a fresh file over the inspected link). AC-HRH-001, -002 and -003 (b) already landed in M0 and flip here. The existing directory-at-state-path and read-only-directory tests stay green unmodified.
6. Mutation runs, recorded in `verdict.md`: AC-HRH-004 owner result ignored, removal before the ownership check, warning omitted; AC-HRH-005 check always "owned"; AC-HRH-006 unconditional removal; AC-HRH-003 a heal that removes a directory, and a heal that ignores the removal failure.
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
- List every Gap: skipped tests with platform and uid, unobserved Linux and Windows behaviour, and the unmeasured heal burst.

## §G Anti-patterns to avoid

- Fixing F5 or F6 in code without a reproduction (the leader's narrowing): the disclosure is the repair.
- Taking a lock on the append path to close the event-loss window (D3 A is operator-held).
- Opening the state file with the same call and "handling" a symbolic link after the fact: the link must be dealt with before any open that could truncate.
- Deleting a state-path entry the current user does not own, or deleting one without re-checking that it is still the inspected entry.
- Treating a runtime skip as a build constraint: a FIFO test without `//go:build !windows` does not compile on Windows.
- A package-level mutable seam variable in a package whose tests run in parallel.
- Reading a `go test -run` exit 0 without its swept count.
- Claiming a Windows result from a build or vet run.

## §H Cross-references

`spec.md` §A-§F; `acceptance.md`; `decision-index.md`; `.moai/reports/t1425/sync-audit.md` §3, `.moai/reports/t1425/sync-audit-delta.md` §3 and §1 Residual-risk, `.moai/reports/t1425/decision-records.md`; `.moai/reports/t1432/plan-audit.md` (iteration 1); `.moai/specs/SPEC-AGENT-TEAM-RETIRE-001/spec.md` REQ-ATR-001; `.claude/rules/moai/core/verification-claim-integrity.md` §2.2, §2.3; `.claude/rules/moai/development/verification-completeness.md` §1.1, §2, §2.1.
