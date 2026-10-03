# SPEC-HARNESS-RETENTION-HARDEN-001 — Acceptance Criteria (card t1432)

Verification layer for `spec.md` §C. Every criterion is `Given … When … Then …` and binary-testable. Test names are handles for the run phase, not requirements on identifiers. Evidence rules follow `.claude/rules/moai/development/verification-completeness.md` §2 and §2.1.

**Document-level pin.** Every RED-now cell below was measured on tree `db6d88a2a` (committed HEAD of `WT-harness-retention-debt` when the plan was first written; its Go code is identical to base `1e2151a38`, `git diff --quiet 1e2151a38 db6d88a2a -- internal cmd` exit 0). A ledger entry that carries another SHA says so: the entries re-run for revision 0.3.0 carry `2ebc10f8f`, whose Go code is also identical to base (ledger E-031).

**Amendment 0.4.0 pin [0.4.0].** Every RED-now, green-at-base and survivor cell the amendment adds (ledger E-032 to E-046, below the original ledger) was measured on tree `7639c04c1` (committed HEAD of `WT-harness-retention-debt` when the amendment was written, clean status before and after; it carries the card's run and sync commits and no heal-lock code). A cell that carries another SHA says so. The amended or new criteria are AC-HRH-006 (rewritten), AC-HRH-008 (optional case d), AC-HRH-010 (one sentinel), AC-HRH-011 (one clause), AC-HRH-015 and AC-HRH-016 (new); every other criterion and ledger entry is the 0.3.0 text, kept as history. The amendment's drafts live in the session scratch directory `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/783a9ff3-1235-40b6-8b28-8f6b61be9e2c/scratchpad/t1432-amend/` and are instruments, not part of the change (Gap G-9).

**Classes.** Each criterion carries one class:

- **RB (release-blocking)** — a behaviour-changing criterion whose RED was observed on the pre-implementation tree (ledger id in the RED-now column) and which a named milestone flips.
- **RG (regression-guard)** — a criterion green at base (a pin), or whose RED cannot be obtained before the code exists. It is not recorded as a pass on the strength of a starting observation; its adoption proof is the mutation run named in its green-path cell (§1.1: a check is complete only after its failure was observed on a known failing input).

**How the RED-now cells were obtained.** The tests do not exist on the tree yet, so a plain `go test -run` on the tree prints `ok … [no tests to run]` and exits 0 (ledger E-019, a vacuous green). RED-now therefore runs a draft of the test, kept outside the tree's own package files, injected with `go test -overlay`. The drafts (`zz_*_test.go`), the overlay JSON files and the mutant copies of `retention.go` (`retention_m*.go`) are committed in `.moai/reports/t1432/red-now-drafts/`; they are measurement instruments, not part of the change. The command form and its stated precondition are in the ledger preamble. M0 authors the tree-resident tests under the same test names and records a second RED-now per criterion without an overlay in `.moai/reports/t1432/red-baseline.md`; a criterion whose second RED differs from its ledger cell is a Gap.

**Swept-count requirement.** A `go test -run` verification is read only after its swept count: its verify command carries `-v` and the output must contain a `--- PASS: <Name>` line for every named test (or `--- FAIL`). `[no tests to run]`, `[no test files]` and `--- SKIP` are Gaps, never passes (§1.1). Permission-based and symbolic-link tests skip when the effective uid is 0 or on Windows; the verdict names the platform and uid of every run.

**Environment.** Test commands run with the kanban variables unset in the same shell invocation (`plan.md` §C); that prefix is not part of the cited command.

## Traceability

| AC | REQ | Item | Class | RED-now (ledger) | Green path |
|---|---|---|---|---|---|
| AC-HRH-001 | REQ-HRH-001 | F4 | RB | E-001 (exit 1, victim changed, path still a link) | flips at M2: `--- PASS`, exit 0 |
| AC-HRH-002 | REQ-HRH-002 | F7 | RB | E-002 (exit 1, `permission denied` twice) | flips at M2: `--- PASS`, exit 0 |
| AC-HRH-003 | REQ-HRH-003 | F7 boundary | RG | E-012 (case b green at base); case a is the existing test | stays green; mutation proof at M2 |
| AC-HRH-004 | REQ-HRH-003, 004 | Q4 refusal and warning | RG | not obtainable: needs the owner seam (Gap G-1) | `--- PASS` at M2 plus three mutants killed at M2 |
| AC-HRH-005 | REQ-HRH-001, 003, 004 | Q4 default owner check, link-owner rule | RG | not obtainable: symbol absent (Gap G-1) | `--- PASS` at M2 plus three mutants killed at M2 |
| AC-HRH-006 [0.4.0] | REQ-HRH-005 | heal safety, heal-lock serialization | RB (case b); RG (cases a, c, d); optional M9 (case e) | case b: E-036 (exit 1, `did not wait`, entry replaced); case a green at base (E-040); cases c, d green at base (E-039); case e survivors E-044 | case b flips at M8: `--- PASS`, exit 0; cases a, c, d stay green; mutants killed at M8 |
| AC-HRH-007 | REQ-HRH-006 | event loss | RB | E-003 (exit 1, `late-event count = 0`) | flips at M1: `--- PASS`, exit 0 |
| AC-HRH-008 | REQ-HRH-007 | event loss, terminator | RB (cases b, c); RG (case a); optional M9 (case d) [0.4.0] | E-004, E-005 (exit 1); case a green at base; case d survivor E-044 (N2) | flips at M1: `--- PASS`, exit 0; case d adopted at M9 |
| AC-HRH-009 | REQ-HRH-008 | append and pre-lock path | RG | E-013 green at base, E-028 (the `-L` form); controls E-014, E-029, E-015, E-030 | stays exit 0 |
| AC-HRH-010 | REQ-HRH-009, 010, 011 | disclosure | RB | E-010 (`0`, exit 1), E-011 (`1`, exit 0); the ninth sentinel (0.4.0): E-041 (`0`, exit 1) | flips at M5 (eight sentinels) and M8 (the ninth, `heal lock gives no exclusion on Windows`): each sentinel count ≥ 1, old sentence 0 |
| AC-HRH-011 | REQ-HRH-012 | cross-platform | RG | E-016 green at base; controls E-017, E-018; heal-lock controls E-043 (0.4.0) | stays exit 0 |
| AC-HRH-012 | REQ-HRH-013 | N1 | RB | E-009 (exit 1, `panic: test timed out after 40s`) | flips at M4: fails in ≤ 15 s under mD |
| AC-HRH-013 | REQ-HRH-014 | N2 / mB | RB (survivor RED) | E-006 survives (exit 0), E-007 killed by the draft (exit 1) | lands at M0: `--- PASS` unmutated (E-020), fails under mB |
| AC-HRH-014 | REQ-HRH-015 | N2 / mC | RB (survivor RED) | E-008 survives (exit 0); kill not obtainable (Gap G-2) | flips at M3: `--- PASS` unmutated, fails under mC |
| AC-HRH-015 [0.4.0] | REQ-HRH-016 | heal lock, contention timeout | RB | E-037 (exit 1, error `<nil>`, entry replaced, no warning) | flips at M8: `--- PASS`, exit 0, three mutants killed |
| AC-HRH-016 [0.4.0] | REQ-HRH-016 | heal lock, hostile heal-lock entry | RB | E-038 (exit 1, four subtests red, no hang) | flips at M8: `--- PASS`, exit 0, three mutants killed |

## Acceptance scenarios

### AC-HRH-001 — A symlinked state path is replaced and its target is untouched (REQ-HRH-001)

- **Given** a log with one stale event, a 64-byte victim file, and a symbolic link owned by the current user at `<log>.prune-state` pointing at the victim,
- **When** `PruneStaleEntries(30)` runs,
- **Then** the call returns nil; the victim's bytes are identical to before; the state path is a regular file (not a link) whose content parses as a fresh stamp; the stale event is gone from the log and its month archive exists.
- Verify: `go test -count=1 -v -run '^TestPruneStateSymlinkReplacedTargetUntouched$' ./internal/harness/` prints `--- PASS: TestPruneStateSymlinkReplacedTargetUntouched ` (the name followed by a space) and exits 0.
- RED-now: E-001. Red for the stated reason: the base code opens the link read-write (`retention.go:121`) and writes the stamp into the victim, so the victim changes and the path stays a link; the prune itself completes, so a "skip the prune" mutant would pass the first Then and fail the last (mutant probe: also fails a mutant that removes the target instead of the link). The victim here is owned by the current user, so this criterion does not tell a link-owner read from a target-owner read; AC-HRH-005 (d) does.

### AC-HRH-002 — An unwritable state file owned by the user is replaced and retention continues (REQ-HRH-002)

- **Given** a log with one stale event and a state file of mode 0400 owned by the current user in a writable directory,
- **When** the first `PruneStaleEntries(30)` runs, and then a second instance runs one day later against a log that has gained a new stale event,
- **Then** both calls return nil, both stale events are archived and gone from the log, and the state file is a regular file the current user can open read-write whose file identity differs from the 0400 file's (replacement, not a permission change).
- Verify: `go test -count=1 -v -run '^TestPruneStateUnwritableFileReplaced$' ./internal/harness/` prints `--- PASS: TestPruneStateUnwritableFileReplaced ` (the name followed by a space) and exits 0 (skips as uid 0 or on Windows; a skip is a Gap).
- RED-now: E-002. Red for the stated reason: both calls return `retention: prune state open failed: … permission denied` and the stale events stay in the log. The identity assertion kills a mutant that only runs `chmod`.
- Limit, stated: the scope of the heal to a permission error (REQ-HRH-002), and the skip for any other open failure (REQ-HRH-003), are reviewer-read; no portable way exists to make the open of a regular file fail for another reason, so a mutant that heals on every open error passes this test.

### AC-HRH-003 — An unreplaceable state path still skips safely (REQ-HRH-003)

- **Given** (a) the state path is a directory, and (b) the state file is mode 0400 and owned by the user inside a directory made non-writable (mode 0555),
- **When** `PruneStaleEntries(30)` runs in each case,
- **Then** each call returns a non-nil error, the log is byte-identical to before, no archive directory exists, the state-path entry is unchanged, and `RecordEvent` through an observer carrying that retention still succeeds.
- Verify: the existing `TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds` (case a) stays green unmodified, plus `go test -count=1 -v -run '^TestPruneStateUnreplaceableInReadOnlyDirSkips$' ./internal/harness/` (case b) prints `--- PASS` and exits 0.
- Both cases are non-FIFO by construction (a directory; a regular file in a read-only directory), so the criterion stays decidable: a FIFO at the state path hangs the lock-free pre-check, which makes a test of it undecidable (`spec.md` §A, §F; ledger E-025). That hang is pre-existing and outside this criterion.
- Class RG: case b is green at base (E-012). Adoption proof at M2: a mutant whose heal removes a directory fails case a; a mutant that ignores the removal failure fails case b.

### AC-HRH-004 — A foreign-owned entry is left byte-identical, and the pruner warns (REQ-HRH-003, REQ-HRH-004)

- **Given** the pruner's owner check replaced (through the one unexported test-only field) by one that reports "not owned", and (a) a symbolic link at the state path pointing at a victim file, (b) a regular state file of mode 0400,
- **When** `PruneStaleEntries(30)` runs in each case with standard error captured,
- **Then** each call returns a non-nil error naming the state path; the log, the archive directory and the state-path entry are byte-identical to before (the link: same file identity and same target text; the regular file: same identity, mode and bytes), the victim is unchanged, and standard error holds exactly one line that starts with `[WARN] harness/retention:` and contains the state path.
- Verify: `go test -count=1 -v -run '^TestPruneStateForeignOwnedLeftUntouchedAndWarns$' ./internal/harness/` prints `--- PASS` and exits 0. The test is serial (no `t.Parallel()`) because it swaps `os.Stderr`, and its file carries `//go:build !windows`.
- Class RG: RED-now is not obtainable on the unmodified tree because the owner seam does not exist (G-1); a non-root user cannot build a foreign-owned file (`chown 0` fails with `Operation not permitted`, observed). Adoption proof at M2: three mutants must each fail this test — the owner result ignored (always "owned"), the removal done before the ownership check, and the warning omitted.

### AC-HRH-005 — The real owner check answers correctly, reads a link's own owner, and Windows answers conservatively (REQ-HRH-001, REQ-HRH-003, REQ-HRH-004)

- **Given** the production owner check (it takes the entry's path and reads the owner itself with a no-follow stat), (a) a file the test created in its temporary directory, (b) the root directory `/` (owned by uid 0), (c) a retention built by the constructor, (d) a symbolic link the test created in its temporary directory pointing at `/` (the link's own owner is the current user, its target's owner is uid 0: OBSERVED, E-026), and (e) a symbolic link owned by another user, the first of the fixed candidates `/var`, `/tmp`, `/etc`, `/bin` whose `Lstat` reports a symbolic link with an owner other than the effective user (root-owned on darwin: OBSERVED, E-027; Linux not observed),
- **When** the check is evaluated on (a), (b), (d) and (e), and the constructor-installed check is evaluated on (b),
- **Then** (a) is owned; (b) is not owned for a non-root user; (d) is owned, because a link's own owner decides and not its target's; (e) is not owned, because a foreign-owned link is never healable whatever its target's owner; the constructor installs the real check (the same answer on (b)); steps (b), (d) and (e) skip when the effective uid is 0, and (e) is also a recorded skip when no candidate qualifies; for Windows the twin file defines the same symbol and answers "not owned" for every entry, which is verified by build and vet only.
- Verify: `go test -count=1 -v -run '^TestOwnerCheckDefault$' ./internal/harness/` prints `--- PASS` and exits 0 (file carries `//go:build !windows`); `GOOS=windows go build ./internal/harness/` exits 0 (AC-HRH-011).
- Class RG: the symbol is absent at base (G-1). Adoption proof at M2: three mutants must each fail this test. (1) A check that always answers "owned" fails (b) and (e). (2) A check that reads the owner of the entry a link points at (a follow-the-link stat) fails (d): the link owned by the current user and pointing at `/` reads as not owned; this mutant passes AC-HRH-001..004 and -006, which use a user-owned target or an injected answer (plan-audit iteration 2, finding B2). (3) A check that treats every symbolic link as owned fails (e); it passes (d), which is why (e) exists. Limitations stated, not hidden: the wiring and steps (b), (d) are pinned only where the test user is not root, and (e) only where a foreign-owned link is among the candidates (darwin: `/var`, `/tmp`, `/etc`); the reverse case, a foreign-owned link pointing at a user-owned target, cannot be built by a non-root user and is covered only through (e) and the injected-check test AC-HRH-004.

### AC-HRH-006 [0.4.0] — Healers serialize on the heal lock; a concurrent healer's fresh state file is never removed (REQ-HRH-005)

Rewritten in 0.4.0. Case (a) is the 0.3.0 criterion kept as case (a); cases (b) to (d) are new; case (e) is optional (plan M9).

- **(a) Given** a symbolic link owned by the user at the state path that the heal helper has inspected, and then a stand-in for a concurrent healer that creates a healthy regular state file elsewhere in the directory and renames it over the link (so the new file's identity differs from the inspected link's), **when** the helper is asked to remove the inspected entry, **then** it does not remove the new file (same identity and bytes afterwards) and reports that the entry changed; with the entry unchanged the same call removes it.
  - Verify: `go test -count=1 -v -run '^(TestHealDoesNotRemoveAFreshStateFile|TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal)$' ./internal/harness/` prints two `--- PASS` lines and exits 0 (E-040). The rename keeps the old link alive while the new file is created, so the two identities cannot coincide through inode reuse.
- **(b) Given** a faulty symbolic link owned by the current user at the state path, a log with one stale event, and another descriptor holding an exclusive `flock` on `<log>.prune-heal`, **when** `PruneStaleEntries(30)` starts, **then** (i) after 300 ms the pruner has not returned and the state-path entry is still the inspected link (same file identity, same type); and **when** the holder then renames a fresh, healthy regular state file carrying a still-fresh stamp over the link (as a winning healer would) and releases the lock, **then** (ii) the pruner returns nil, the fresh file has the same identity and bytes, and the log is byte-identical (the pruner re-inspected under the lock, did not remove the winner's file and adopted its stamp).
  - Verify: `go test -count=1 -v -run '^TestPruneHealSerializesOnTheHealLock$' ./internal/harness/` prints `--- PASS` and exits 0. The file carries `//go:build !windows`, takes the lock itself with `syscall.Flock`, and names the heal-lock file by the literal `.prune-heal`, so it compiles against the unmodified package (M7).
  - Why (i) cannot be falsely red: a slow runner only delays the pruner, which is blocked on the lock either way; it can make (i) pass vacuously only if the pruner has not yet reached the lock, which (ii) then covers (the pruner must still end with the winner's file intact). Step (ii) can be falsely red only if the test itself is stalled for longer than the 2 s bound between (i) and the release, so that the pruner legitimately times out; not measured under load (Gap G-12).
- **(c) Given** an absent state path, and a healthy regular state file whose stamp is expired, **when** `PruneStaleEntries(30)` runs on each, **then** the prune runs (the stale event is gone from the log) and no heal-lock entry exists afterwards.
  - Verify: `go test -count=1 -v -run '^TestPruneCommonPathCreatesNoHealLock$' ./internal/harness/` prints `--- PASS` and exits 0 (E-039).
- **(d) Given** a faulty symbolic link owned by the current user and a month archive path made a FIFO so the pruner blocks in its archive step after it has healed the entry and stamped, **when** the test, with the pruner blocked there, takes a non-blocking exclusive `flock` on `<log>.prune-heal`, **then** the `flock` succeeds (the heal lock is held only across the re-inspection and the removal, never across the prune).
  - Verify: `go test -count=1 -v -run '^TestPruneHealLockIsFreeWhilePrunerIsInItsArchiveStep$' ./internal/harness/` prints `--- PASS` and exits 0 (E-039; the file carries `//go:build !windows`).
- **(e) Optional, plan M9, Priority Low (N1 and N4 of `sync-audit-delta.md`).** The removal helper does not remove a swapped-in file whose modification time equals the inspected entry's (a `Chtimes` on the fresh file before the removal call), and the file arm of the heal (an unwritable regular file) removes the entry only through the conditional removal (a stand-in owner check renames a fresh stamped file over it). Verify: `go test -count=1 -v -run '^(TestHealRemovalIsNotDecidedByModTimeAlone|TestPruneHealFileArmKeepsAFreshStateFileSwappedInDuringTheHeal)$' ./internal/harness/`. Survivor RED: E-044 (mutants N1 and N4 pass the whole package); kills and unmutated passes observed with drafts: E-045. If M9 is not done, the two items are listed as carried debt in the verdict; they are never silently dropped.
- RED-now: E-036, case (b), exit 1. Red for the stated reason: no heal lock exists at this tree, so the pruner heals at once (`the pruner returned (<nil>) while another descriptor held the heal lock: it did not wait`; `the state-path entry was removed or replaced while the heal lock was held by another descriptor`). Cases (a), (c), (d) are green at base (E-040, E-039): they pin behaviour the change must keep, and case (d) is vacuous at base (no lock is taken at all) and earns its adoption from its mutant.
- Class: RB for case (b); RG for (a), (c), (d); optional for (e). Mutant probe, to be run at M8 and pasted failing into the verdict: heal lock never taken (fails (b): the RED-now itself); heal lock taken on the common path (fails (c)); heal lock held across the prune (fails (d)); for (a) a removal that is unconditional (fails (a)). **Stated limit:** a mutant that releases the lock before the removal, or takes it after the re-inspection, satisfies (b) to (d) and is not killed by any deterministic test, because no seam sits between the re-inspection and the removal; it is covered by the post-fix race probe (Definition of Done 10, an expectation until measured) and by review.
- The raw race cannot be reproduced deterministically without a seam inside the helper, so (b) is the contract test: at this tree it fails because the lock is not honoured, after the change it passes because the lock serializes. The statistical evidence is the probe, DoD 10: pre-fix, through the production entry point and with a lock-honouring swapper, 5 removals in 1270 trials (E-035); post-fix expectation, 0 removals in at least 20000 trials with no early stop. The audit's own probe (E-034) calls the bare removal helper and will still count removals after the change, by design (the lock is taken by its caller); it is a control, not the post-fix evidence.

### AC-HRH-007 — A late event survives the prune (REQ-HRH-006)

- **Given** a log with one stale and one fresh event and the archive path made a FIFO so the pruner blocks in its archive step after it has read the log,
- **When** a second writer appends `late-event` through the append helper while the pruner is blocked, and the pruner is then released,
- **Then** the pruner returns nil, the log contains `fresh` and then `late-event` (in that order, `late-event` verbatim), does not contain the stale event, and the archive contains the stale event.
- Verify: `go test -count=1 -v -run '^TestPruneCarriesLateEvents$' ./internal/harness/` prints `--- PASS` and exits 0 (file carries `//go:build !windows`).
- RED-now: E-003. Red for the stated reason: the late event is absent after the prune (`late-event count = 0`) because the base pruner rewrites the log from what it read.

### AC-HRH-008 — The replacement log is terminated, a partial line is carried once, and quiescent output is unchanged (REQ-HRH-007)

- **Given** (a) a log nothing is appended to during the prune, with unparsed lines around a stale event; (b) a log to which a fragment without a terminating newline is appended while the pruner is blocked in its archive step; (c) a log that holds one other stale event and whose final line is a stale event with no terminating newline,
- **When** the prune runs in each case, and in (b) a normal event is then appended through the append helper,
- **Then** in (a) the replacement equals what the baseline writes (the existing `TestPruneKeepsUnparsedLinesVerbatim` and `TestPruneNothingStaleLeavesLogUntouched` pass unmodified); in (b) the fragment occurs exactly once, the replacement ends with a newline, and the appended event is on its own line and parses; in (c) the final stale event is still in the log after this prune (the prune rewrites the log because the other stale event is archived, and the rewrite adds the terminator) and is gone after a prune one interval later.
- Verify: `go test -count=1 -v -run '^(TestPruneKeepsUnparsedLinesVerbatim|TestPruneNothingStaleLeavesLogUntouched|TestPruneTailPartialLineCarriedAndTerminated|TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval)$' ./internal/harness/` prints four top-level `--- PASS` lines (`TestPruneKeepsUnparsedLinesVerbatim` also prints one indented `--- PASS` line per subtest) and exits 0 (the new tests carry `//go:build !windows`).
- RED-now: E-004 (fragment occurs 0 times) and E-005 (the stale final line is archived at once, so it is absent), both exit 1; case a is green at base and is the pin. Mutants (probed against the drafts): a verbatim tail copy with no terminator fails (b) at its last assertion; classifying the final unterminated line fails (c). When the final unterminated stale line is the only stale event, see Edge cases.
- **(d) Optional, plan M9, Priority Low (N2 of `sync-audit-delta.md`) [0.4.0].** **Given** a log with one stale and one fresh event and nothing appended during the prune, **when** the prune runs, **then** the replacement log is exactly the kept line: one newline in total and no trailing blank line (a raw-byte check; the log readers drop blank lines, so only a raw count sees one). Verify: `go test -count=1 -v -run '^TestPruneWithoutLateEventAddsNoBlankLine$' ./internal/harness/`. Survivor RED: E-044 (mutant N2, a newline written when the tail is empty, passes the whole package); kill and unmutated pass observed with a draft: E-045. If M9 is not done the item is carried debt in the verdict.

### AC-HRH-009 — The append path and the pre-lock path gain nothing (REQ-HRH-008)

- **Given** the final tree,
- **When** the observer source and the pre-lock functions — `PruneStaleEntries` (the in-memory interval check and the log existence check) and the stamp-read helpers `readStamp`, `readStampFile` and `stampIsFresh` — are compared to base `1e2151a38`,
- **Then** `internal/harness/observer.go` is byte-identical, and no commit after the base changes a line of those four functions (a stray call would land there); the existing `TestPruneStamp_FreshStampNeedsNoLock` still passes unmodified.
- Verify: `git diff --quiet 1e2151a38 -- internal/harness/observer.go` exits 0, and `git log --format=%h -s -L '/^func (r \*Retention) PruneStaleEntries/,/^}/:internal/harness/retention.go' -L '/^func readStamp(/,/^}/:internal/harness/retention.go' -L '/^func readStampFile(/,/^}/:internal/harness/retention.go' -L '/^func stampIsFresh(/,/^}/:internal/harness/retention.go' 1e2151a38..HEAD` prints nothing (exit 0). The four `-L` options must follow the order of the functions in the file (a later option's search starts after the end of the previous range); a function that is missing, renamed or moved above its predecessor makes git exit 128 instead of printing nothing, so the sweep cannot shrink silently (ledger E-030). The heal helper, the tail copy and the owner check live outside these functions (called only from `pruneExclusive` and below).
- Class RG (green at base: E-013 for `observer.go`, E-028 for the `-L` form). Positive controls: the same `git diff --quiet` form exits 1 on a commit that changed `observer.go` (E-014); the four-function `-L` form prints `fe211e9c9` on the range that added the three stamp-read helpers and edited `PruneStaleEntries` (E-029); the `PruneStaleEntries`-only form prints a hash on a commit that changed that function and nothing on a range that did not (E-015). An empty `-L` output is read as a pass only where `1e2151a38..HEAD` holds commits (it does once M0 has landed) and the controls above held.

### AC-HRH-010 — The residual window, F5 and F6 are disclosed with their labels, and history stays out of source (REQ-HRH-009, REQ-HRH-010, REQ-HRH-011)

- **Given** the final `internal/harness/retention.go`,
- **When** it is searched for the sentinel phrases below, one per obligation,
- **Then** each sentinel occurs at least once, the old sentence occurs 0 times, no card id appears in the file, and a reviewer reads the F5 and F6 sentences against REQ-HRH-010 and REQ-HRH-011 (the sentinels alone do not prove that the sentences state the obligations).
- Sentinels: `residual window` (REQ-HRH-009); `no cross-process exclusion`, `burst of hook processes`, `F5: not reproduced, not measured` (REQ-HRH-010); `lock waiters block with no timeout`, `5 s hook timeout`, `appended before the wait`, `F6: not reproduced, not measured` (REQ-HRH-011); `heal lock gives no exclusion on Windows` (REQ-HRH-010, the clause added in 0.4.0) [0.4.0].
- Verify, one command per sentinel: `grep -c -F "<sentinel>" internal/harness/retention.go` prints a number ≥ 1 for each; `grep -c -F "events other hooks append in that window are lost" internal/harness/retention.go` prints 0; `grep -c -E 't1[0-9]{3}' internal/harness/retention.go` prints 0 (base: `0`, exit 1, so the guard keeps card ids out of source; card history belongs to the decision record).
- RED-now: E-010 (the eight sentinels together match 0 lines at base, exit 1) and E-011 (the old sentence matches 1 line, exit 0). For the ninth sentinel (0.4.0): E-041, `grep -c -F "heal lock gives no exclusion on Windows" internal/harness/retention.go` prints `0` and exits 1 on tree `7639c04c1`, where the other eight are already present (landed at M5); it flips at M8, with the heal-lock code, to a count of at least 1. The `async` obligation of REQ-HRH-011 and the sentence-level reading are reviewer-read; this is stated as the pass condition, not hidden.

### AC-HRH-011 — The shared lock package is untouched and Windows builds, vets and compiles the tests (REQ-HRH-012)

- **Given** the final tree,
- **Then** every file under `internal/lockfile` is byte-identical to the base, and `internal/harness` and `internal/lockfile` build and vet for Windows, vet compiling their test files, so a FIFO test without a `//go:build !windows` constraint, or a POSIX-only production file without a Windows twin, fails here.
- Verify: `git diff --quiet 1e2151a38 -- internal/lockfile` exits 0; `GOOS=windows go build ./internal/harness/ ./internal/lockfile/` exits 0; `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` exits 0. Verification level: build and vet only (no Windows runtime).
- **[0.4.0] The heal-lock files are covered by the same two commands.** The new POSIX-only production code (the heal-lock helper: `syscall.Flock`, `O_NOFOLLOW`, `O_NONBLOCK`) lives in `retention_heal_unix.go` (`//go:build !windows`) with a `retention_heal_windows.go` twin defining the same symbol, and `retention_heallock_test.go` carries `//go:build !windows`; a missing twin or a missing tag makes the Windows build or vet fail. Level: build and vet only; Windows runtime not observed, and the twin's "acquires no exclusion" behaviour is reviewer-read. This clause is a regression-guard with no RED-now of its own (the twin cannot differ before it exists); its positive controls are E-043: with an untagged test file calling `syscall.Flock` the Windows vet exits 1 (`undefined: syscall.Flock`), and with the draft heal-lock tests (tagged) it exits 0.
- Class RG (green at base, E-016). Positive controls: with an untagged FIFO test file the Windows vet exits 1 (E-017) and with the draft files tagged it exits 0 (E-018). Why this is the only per-card guard: the `test` job in `.github/workflows/ci.yml` runs on `ubuntu-latest` only (`:125`), and its Windows leg is a cross-compile of the binary (`go build`, test files not compiled, `:509`, `:545`); the three-OS `go vet ./...` and `go test` run in `.github/workflows/release-pr-multi-os.yml` (matrix `:98`, steps `:152`, `:210`) at release time (READ, not run here), so a Windows test-file compile error would otherwise surface at release.

### AC-HRH-012 — A pruner that never archives fails the killed-pruner test quickly (REQ-HRH-013)

- **Given** mutant mD (the pruner stamps and then returns without pruning) applied as a temporary overlay,
- **When** `TestPruneStamp_StampExistsBeforeTheWork` runs with `-timeout 40s`,
- **Then** it exits non-zero in at most 15 seconds wall time, the output names the missing archive step and does not contain `panic: test timed out`; and with no mutation the same test passes.
- Verify: `go test -count=1 -v -timeout 40s -overlay <mD overlay> -run '^TestPruneStamp_StampExistsBeforeTheWork$' ./internal/harness/` under mD, timed from outside (`timeout 30`), plus the unmutated run printing `--- PASS` and exiting 0.
- RED-now: E-009. Red for the stated reason: with the unmodified test, the test goroutine blocks in the FIFO's blocking open and the go test alarm fires (`panic: test timed out after 40s`, `FAIL … 40.722s`, exit 1).

### AC-HRH-013 — Mutant mB (no truncation of a longer stamp) is killed (REQ-HRH-014)

- **Given** a state file holding an older stamp with nine fractional digits and a prune whose clock formats with fewer bytes,
- **When** the prune records its stamp,
- **Then** the state file content equals exactly the new stamp (no trailing bytes) and parses as fresh; with `Truncate(0)` removed from the stamp writer the test fails.
- Verify: `go test -count=1 -v -run '^TestPruneStampShorterOverLongerIsExact$' ./internal/harness/` prints `--- PASS` and exits 0 unmutated, and prints `--- FAIL` and exits 1 under mB.
- RED-now: a survivor RED. E-006: the whole package with mB applied prints `ok` and exits 0 (the existing suite is blind to mB); E-007: the draft test under mB prints `state file = "2026-10-02T00:00:00Z123456789Z", want exactly "2026-10-02T00:00:00Z"` and exits 1 (the criterion detects mB); E-020: the draft unmutated passes.

### AC-HRH-014 — Mutant mC (stamp-write error ignored) is killed (REQ-HRH-015)

- **Given** the locked phase of the prune called with a state file opened read-only, so that its stamp write fails (OBSERVED on darwin: `Truncate` returns `invalid argument`),
- **When** the phase runs on a log with a stale event,
- **Then** it returns a non-nil error, the log is byte-identical, and no archive directory is created; with the stamp-write error ignored the test fails.
- Verify: `go test -count=1 -v -run '^TestPruneStampWriteFailureSkipsPrune$' ./internal/harness/` prints `--- PASS` and exits 0 unmutated, and exits 1 under mC. The test file carries `//go:build !windows`; the outcome, not the mechanism, is the criterion, so a different seam that produces the same failure satisfies it.
- RED-now: a survivor RED. E-008: the whole package with mC applied prints `ok` and exits 0. The kill observation needs the extracted locked phase, which exists only after M3 (Gap G-2). Linux behaviour of a read-only handle is not observed; if it differs, the seam choice in `spec.md` §B D5 is revisited before M3.

### AC-HRH-015 [0.4.0] — A heal lock held past the bound fails closed within the bound (REQ-HRH-016)

- **Given** a faulty symbolic link owned by the current user at the state path (pointing at a 64-byte victim), a log with one stale event, and another descriptor that holds an exclusive `flock` on `<log>.prune-heal` for the whole call,
- **When** `PruneStaleEntries(30)` runs with standard error captured, bounded by a hard cap of 8 s so that an unbounded wait fails the test instead of hanging it,
- **Then** the call returns in under 4.5 s (the bound is 2 s, the hook timeout 5 s) with a non-nil error naming the heal-lock path; the state-path entry is unchanged (same file identity, same mode, still a link); the victim is unchanged; the log is byte-identical and no archive directory exists; standard error holds exactly one line that starts with `[WARN] harness/retention:` and contains the heal-lock path; and the heal-lock entry is the same regular file, still empty (not removed, replaced or truncated).
- Verify: `go test -count=1 -v -run '^TestPruneHealLockHeldPastTheBoundFailsClosed$' ./internal/harness/` prints `--- PASS` and exits 0. The test is serial (no `t.Parallel()`) because it swaps `os.Stderr`, and its file carries `//go:build !windows`.
- RED-now: E-037, exit 1. Red for the stated reason: no heal lock exists at this tree, so the pruner heals at once (`want an error naming …prune-heal, got <nil>`; `the state-path entry was removed or replaced although the heal lock was not acquired`; `log changed`; no warning), after 0.03 s, far inside the ceiling, so the failure is in the outcome and not in the timing.
- Mutant probe, at M8, each pasted failing into the verdict: an unbounded blocking wait (fails the 4.5 s ceiling after the 8 s cap); the heal proceeding after the bound (fails: the entry is replaced); no error or no warning (fails). The ceiling assertion could be falsely red on a runner so loaded that the 2 s wait stretches past 4.5 s; not measured under load (Gap G-12). The criterion pins an upper bound; that a waiter does wait for a lock that is released in time is pinned by AC-HRH-006 (b).

### AC-HRH-016 [0.4.0] — A hostile heal-lock entry fails closed, is never followed, truncated or removed, and adds no hang path (REQ-HRH-016)

- **Given** a faulty symbolic link owned by the current user at the state path and, at the heal-lock path `<log>.prune-heal`, (a) a symbolic link to a 64-byte victim file, (b) a directory, (c) a FIFO, (d) a regular file for which the pruner's owner check answers "not owned" (the check replaced through the existing test-only field by one that says "owned" for every path except the heal-lock path),
- **When** `PruneStaleEntries(30)` runs in each case with standard error captured, the call bounded at 4.5 s,
- **Then** each call returns (no hang) with a non-nil error naming the heal-lock path; the state-path entry is unchanged; the state-path victim is unchanged; the log is byte-identical; standard error holds exactly one line that starts with `[WARN] harness/retention:` and contains the heal-lock path; and the heal-lock entry is untouched: in (a) still a link, with the victim behind it byte-identical (never followed, never truncated); in (b) still a directory; in (c) still a FIFO; in (d) the same regular file, still empty.
- Verify: `go test -count=1 -v -run '^TestPruneHealLockHostileEntryFailsClosed$' ./internal/harness/` prints `--- PASS` for the test and for its four subtests, and exits 0 (serial, `//go:build !windows`).
- RED-now: E-038, exit 1, four subtests red. Red for the stated reason: the unmodified pruner never looks at the heal-lock path, so it heals the state-path link at once and every assertion about the outcome fails (`want an error naming …prune-heal, got <nil>`; `the state-path entry was removed or replaced although the heal lock was unusable`; `log changed`; no warning).
- Limits, stated and not hidden: the FIFO subtest pins the outcome (fail closed, return within the cap), not the mechanism; on darwin an `O_RDWR` open of a FIFO does not block (E-032 P3), so the cap would catch a hang only on a platform where it does (Linux not observed). The interleavings between the inspection and the open (a link or FIFO swapped in) cannot be forced without a seam and are not pinned (`spec.md` §F). The real owner check on the heal-lock file is the one pinned by AC-HRH-005; (d) pins the wiring from the heal lock to the owner check, not the check.
- Mutant probe, at M8, each pasted failing into the verdict: the heal-lock open follows links (fails (a): the victim is opened and the heal proceeds); the heal-lock owner check skipped (fails (d)); a hostile heal-lock entry removed and recreated, as the state entry is (fails (a), (b), (c)).

## Edge cases

- The effective uid is 0, or the platform is Windows: permission, ownership and symbolic-link tests skip or are excluded by their build constraint; the build/vet criterion (AC-HRH-011) still runs.
- The state path is a symbolic link to a directory or to a path that does not exist: the link (owned by the current user) is removed, never followed, and the prune continues; the target is untouched. A user-owned link to a root-owned directory such as `/` is healed the same way, because the link's own owner decides (AC-HRH-005 (d)).
- A FIFO at the state path, or a link to a FIFO: outside this SPEC. `PruneStaleEntries` blocks in the lock-free pre-check (OBSERVED, E-025; `spec.md` §E, §F); no criterion pins or repairs it and none uses a FIFO at the state path.
- Two processes meet the same fault at the same instant [0.4.0]: healers serialize on the heal lock and the pruner removes an entry only if, under the lock, it is still the one it inspected; a waiter that finds the winner's fresh file leaves it and adopts it (AC-HRH-006 b). The heal window remains on Windows and against a binary that does not honour the lock, and the production frequency of the window and of lock timeouts is unmeasured (`spec.md` §F). A lock that cannot be used fails closed (AC-HRH-015, AC-HRH-016).
- The heal-lock entry is a leftover, empty regular file from an earlier heal [0.4.0]: it is reused; the common path never opens it, and a heal that finds it opens it without truncating it.
- A late line is malformed JSON: it is carried verbatim like any other tail bytes, with the terminator the rule adds only when missing.
- Late events older than the retention cutoff: carried, kept in the log, and archived by a later interval's prune.
- The final unterminated line is the only stale event: the prune classifies nothing as stale, returns without a rewrite and adds no terminator, so the line stays unclassified until an append completes it or a later prune rewrites the log (`spec.md` §B D3); stated, not pinned by a test.

## Quality gate criteria

- `go vet ./internal/harness/ ./internal/lockfile/` exit 0; the project linter at the CI-pinned version reports no new finding in `retention.go` or the touched tests (state the judging build next to the tree HEAD, `verification-claim-integrity.md` §2.2).
- `go test -race -count=1 -v ./internal/harness/` exits 0 and every pre-existing `TestPrune*` function passes; `git diff --name-only 1e2151a38 -- 'internal/harness/*_test.go'` lists only `retention_killed_test.go` among test files that existed at the base.
- `internal/harness` package coverage stays at or above the 85 percent TRUST 5 floor (the t1425 delta audit measured 87.5 percent, `sync-audit-delta.md`); the new branches (heal, ownership refusal, tail-carry, locked-phase seam) are covered by AC-HRH-001..008 and -014.
- `go test -race -count=20` on the killed-pruner and concurrent-process tests exits 0.
- [0.4.0] `go test -race -count=5` on the heal-lock tests of AC-HRH-006 (b to d) and AC-HRH-016 exits 0, and AC-HRH-015 once under `-race` (its run spends the 2 s bound by design); `go vet ./internal/harness/ ./internal/lockfile/`, `GOOS=windows go build` and `GOOS=windows go vet` of both packages exit 0; `git diff --quiet 7639c04c1 -- internal/lockfile internal/harness/observer.go` exits 0 and the AC-HRH-009 `git log -L` form over `7639c04c1..HEAD` prints nothing.

## Evidence ledger (plan-phase measurements)

Entries cited by the tables above. Each carries the command, its verbatim stdout (long outputs are trimmed to the deciding lines and say so), its exit code as its own field, and the tree SHA. Commands run from the worktree root. This ledger quotes stack frames and a compiler message that name the `syscall` package; the build-tag obligation they bear on is REQ-HRH-012 (`//go:build !windows` for every FIFO test file, `//go:build windows` twins for POSIX-only production files).

**Command form.** Every ledger command is a single invocation with no placeholder. Commands that use `-overlay` read the overlay JSON files committed in `.moai/reports/t1432/red-now-drafts/` (relative to the worktree root); each maps a path inside `internal/harness/` to a source file, namely a `zz_*_test.go` draft or a `retention_m*.go` mutant copy. Precondition, stated rather than hidden: an overlay JSON holds absolute paths, and as committed these name the plan phase's session scratch directory (`/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/783a9ff3-1235-40b6-8b28-8f6b61be9e2c/scratchpad/t1432-iter2`), which holds byte-identical copies of the committed sources (`diff -rq` reports no differing file). Where that directory is absent the JSON files must first be regenerated with their values pointing at the committed sources; that is setup, not a cited command, and the leader or M0 owns it (Gap G-3). Entries marked `tree: 2ebc10f8f` were re-run in revision 0.3.0 with this form; their timings and temporary-directory names differ from the earlier runs, the failing lines and exit codes do not.

```
id: E-001   criterion: AC-HRH-001   tree: 2ebc10f8f
command: go test -count=1 -overlay .moai/reports/t1432/red-now-drafts/overlay-drafts.json -run '^TestPruneStateSymlinkReplacedTargetUntouched$' ./internal/harness/
stdout:
--- FAIL: TestPruneStateSymlinkReplacedTargetUntouched (0.04s)
    zz_statepath_test.go:40: victim changed: err=<nil> content="2026-10-02T00:00:00Z"
    zz_statepath_test.go:44: state path is not a regular file: err=<nil>
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	1.839s
FAIL
exit: 1
```

```
id: E-002   criterion: AC-HRH-002   tree: 2ebc10f8f
command: go test -count=1 -overlay .moai/reports/t1432/red-now-drafts/overlay-drafts.json -run '^TestPruneStateUnwritableFileReplaced$' ./internal/harness/
stdout:
--- FAIL: TestPruneStateUnwritableFileReplaced (0.00s)
    zz_statepath_test.go:77: first prune returned retention: prune state open failed: open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestPruneStateUnwritableFileReplaced361675888/001/usage-log.jsonl.prune-state: permission denied, want nil
    zz_statepath_test.go:80: first stale event still in the log
    zz_statepath_test.go:88: second prune returned retention: prune state open failed: open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestPruneStateUnwritableFileReplaced361675888/001/usage-log.jsonl.prune-state: permission denied, want nil
    zz_statepath_test.go:91: second stale event still in the log
    zz_statepath_test.go:95: state file not openable read-write: open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestPruneStateUnwritableFileReplaced361675888/001/usage-log.jsonl.prune-state: permission denied
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	0.712s
FAIL
exit: 1
```

```
id: E-003   criterion: AC-HRH-007   tree: 2ebc10f8f
command: go test -count=1 -overlay .moai/reports/t1432/red-now-drafts/overlay-drafts.json -run '^TestPruneCarriesLateEvents$' ./internal/harness/
stdout:
--- FAIL: TestPruneCarriesLateEvents (0.34s)
    zz_tail_test.go:69: subjects after prune: map[fresh:1]
    zz_tail_test.go:71: late-event count = 0, want 1
    zz_tail_test.go:81: order wrong: [{"timestamp":"2026-10-01T00:00:00Z","event_type":"feedback","subject":"fresh","context_hash":"h","tier_increment":0,"schema_version":"v2.1"}]
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	0.986s
FAIL
exit: 1
```

```
id: E-004   criterion: AC-HRH-008 (case b)   tree: 2ebc10f8f
command: go test -count=1 -overlay .moai/reports/t1432/red-now-drafts/overlay-drafts.json -run '^TestPruneTailPartialLineCarriedAndTerminated$' ./internal/harness/
stdout:
--- FAIL: TestPruneTailPartialLineCarriedAndTerminated (0.33s)
    zz_tail_test.go:106: partial fragment occurs 0 times, want 1
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	1.002s
FAIL
exit: 1
```

```
id: E-005   criterion: AC-HRH-008 (case c)   tree: 2ebc10f8f
command: go test -count=1 -overlay .moai/reports/t1432/red-now-drafts/overlay-drafts.json -run '^TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval$' ./internal/harness/
stdout:
--- FAIL: TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval (0.01s)
    zz_tail_test.go:138: subjects after first prune: map[fresh:1]
    zz_tail_test.go:140: stale-final count after first prune = 0, want 1 (kept in the log this interval)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	0.664s
FAIL
exit: 1
```

```
id: E-006   criterion: AC-HRH-013 (survivor)   tree: 2ebc10f8f
command: go test -count=1 -overlay .moai/reports/t1432/red-now-drafts/overlay-mB.json ./internal/harness/
note: overlay-mB.json replaces retention.go with retention_mB.go (same directory), which differs from the tree file only by the removed Truncate(0) block (`diff` observed in 0.3.0: lines 210-212 deleted)
stdout:
ok  	github.com/modu-ai/moai-adk/internal/harness	2.096s
exit: 0
```

```
id: E-007   criterion: AC-HRH-013 (kill)   tree: 2ebc10f8f
command: go test -count=1 -overlay .moai/reports/t1432/red-now-drafts/overlay-mB-newtest.json -run '^TestPruneStampShorterOverLongerIsExact$' ./internal/harness/
stdout:
--- FAIL: TestPruneStampShorterOverLongerIsExact (0.00s)
    zz_stampbytes_test.go:27: state file = "2026-10-02T00:00:00Z123456789Z", want exactly "2026-10-02T00:00:00Z"
    zz_stampbytes_test.go:30: state file does not parse as a fresh stamp: "2026-10-02T00:00:00Z123456789Z"
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	0.764s
FAIL
exit: 1
```

```
id: E-008   criterion: AC-HRH-014 (survivor)   tree: 2ebc10f8f
command: go test -count=1 -overlay .moai/reports/t1432/red-now-drafts/overlay-mC.json ./internal/harness/
note: retention_mC.go (same directory) differs from the tree file only at lines 142-144 (the stamp-write error check replaced by `_ = writeStamp(sf, now)`; `diff` observed in 0.3.0)
stdout:
ok  	github.com/modu-ai/moai-adk/internal/harness	1.578s
exit: 0
```

```
id: E-009   criterion: AC-HRH-012   tree: 2ebc10f8f
command: go test -count=1 -timeout 40s -overlay .moai/reports/t1432/red-now-drafts/overlay-mD.json -run '^TestPruneStamp_StampExistsBeforeTheWork$' ./internal/harness/
note: retention_mD.go (same directory) differs from the tree file only at line 146 (`err = r.prune(retentionDays, now)` replaced by `err = nil`; `diff` observed in 0.3.0). The goroutine dump is trimmed to the deciding lines (the first three lines and the blocking-open frames); an earlier run's full output is `.moai/reports/t1432/red-now-drafts/mD-out.txt`
stdout:
panic: test timed out after 40s
	running tests:
		TestPruneStamp_StampExistsBeforeTheWork (40s)
...
syscall.Open({0x43cb4c46b600?, 0x1001ef7c0?}, 0x1000000, 0x0)
...
github.com/modu-ai/moai-adk/internal/harness.TestPruneStamp_StampExistsBeforeTheWork.TestPruneStamp_StampExistsBeforeTheWork.func1.func7()
	/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1432/internal/harness/retention_killed_test.go:43 +0x28
...
FAIL	github.com/modu-ai/moai-adk/internal/harness	40.722s
FAIL
exit: 1
```

```
id: E-010   criterion: AC-HRH-010 (presence)   tree: db6d88a2a
command: grep -c -F -e "residual window" -e "no cross-process exclusion" -e "burst of hook processes" -e "F5: not reproduced, not measured" -e "lock waiters block with no timeout" -e "5 s hook timeout" -e "appended before the wait" -e "F6: not reproduced, not measured" internal/harness/retention.go
stdout:
0
exit: 1
```

```
id: E-011   criterion: AC-HRH-010 (old sentence)   tree: db6d88a2a
command: grep -c -F "events other hooks append in that window are lost" internal/harness/retention.go
stdout:
1
exit: 0
```

```
id: E-012   criterion: AC-HRH-003 (case b, pin)   tree: 2ebc10f8f
command: go test -count=1 -overlay .moai/reports/t1432/red-now-drafts/overlay-unreplaceable.json -run '^TestPruneStateUnreplaceableInReadOnlyDirSkips$' -v ./internal/harness/
stdout:
=== RUN   TestPruneStateUnreplaceableInReadOnlyDirSkips
--- PASS: TestPruneStateUnreplaceableInReadOnlyDirSkips (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	1.083s
exit: 0
```

```
id: E-013   criterion: AC-HRH-009 (green at base)   tree: db6d88a2a
command: git diff --quiet db6d88a2a -- internal/lockfile internal/harness/observer.go
stdout: (empty)
exit: 0
```

```
id: E-014   criterion: AC-HRH-009 (positive control: the form goes non-zero on a changed path)   tree: db6d88a2a
command: git diff --quiet 6dc31a727~1 6dc31a727 -- internal/harness/observer.go
stdout: (empty)
exit: 1
```

```
id: E-015   criterion: AC-HRH-009 (controls for the PruneStaleEntries-only history form)   tree: db6d88a2a
command: git log --format=%h -s -L '/^func (r \*Retention) PruneStaleEntries/,/^}/:internal/harness/retention.go' fe211e9c9~1..fe211e9c9
stdout:
fe211e9c9
exit: 0
command: git log --format=%h -s -L '/^func (r \*Retention) PruneStaleEntries/,/^}/:internal/harness/retention.go' fe211e9c9..fe4e8f44e
stdout: (empty; this range changed `pruneExclusive`, not `PruneStaleEntries`)
exit: 0
```

```
id: E-016   criterion: AC-HRH-011 (green at base)   tree: db6d88a2a
command: GOOS=windows go vet ./internal/harness/ ./internal/lockfile/
stdout: (empty)
exit: 0
```

```
id: E-017   criterion: AC-HRH-011 (positive control: a FIFO test with no //go:build !windows constraint)   tree: 2ebc10f8f
command: GOOS=windows go vet -overlay .moai/reports/t1432/red-now-drafts/overlay-untagged.json ./internal/harness/
stdout:
# github.com/modu-ai/moai-adk/internal/harness [github.com/modu-ai/moai-adk/internal/harness.test]
/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/783a9ff3-1235-40b6-8b28-8f6b61be9e2c/scratchpad/t1432-iter2/zz_untagged_fifo_test.go:11:20: undefined: syscall.Mkfifo
exit: 1
note: the control file carries no //go:build !windows constraint, which is the defect this control exposes; E-018 is the tagged counterpart. The path in the message is the overlay's target, a copy of the committed `zz_untagged_fifo_test.go`.
```

```
id: E-018   criterion: AC-HRH-011 (the draft tests, FIFO file tagged //go:build !windows)   tree: 2ebc10f8f
command: GOOS=windows go vet -overlay .moai/reports/t1432/red-now-drafts/overlay-drafts.json ./internal/harness/
stdout: (empty)
exit: 0
```

```
id: E-019   criterion: (the vacuous form this document avoids)   tree: db6d88a2a
command: go test -count=1 -run '^TestPruneStateSymlinkReplacedTargetUntouched$' ./internal/harness/
stdout:
ok  	github.com/modu-ai/moai-adk/internal/harness	0.711s [no tests to run]
exit: 0
```

```
id: E-020   criterion: AC-HRH-013 (draft green when unmutated)   tree: 2ebc10f8f
command: go test -count=1 -overlay .moai/reports/t1432/red-now-drafts/overlay-drafts.json -run '^TestPruneStampShorterOverLongerIsExact$' ./internal/harness/
stdout:
ok  	github.com/modu-ai/moai-adk/internal/harness	0.707s
exit: 0
```

```
id: E-021   criterion: plan.md M0 Exit and Definition of Done 1 (why the listing is bounded)   tree: 2ebc10f8f
command: git log --reverse --format=%h 1e2151a38 -- internal/harness/retention.go
stdout:
68f023289
fe0901cdc
fe211e9c9
987dcc61b
8172295ca
035a7bc31
fe4e8f44e
exit: 0
note: the unbounded history of the file holds seven commits, all reachable from the base. T, a descendant of the base, is never an ancestor of any of them, so `git merge-base --is-ancestor T F` is false for each (E-024 shows the first one) and an unbounded ordering check cannot pass. This is plan-audit iteration 2, finding B1.
```

```
id: E-022   criterion: plan.md M0 Exit (the bounded listing at the plan tip)   tree: 2ebc10f8f
command: git log --reverse --format=%h 1e2151a38..2ebc10f8f -- internal/harness/retention.go
stdout: (empty)
exit: 0
note: no commit of this card touches the file yet, as expected before any fix. An empty listing means "no production commit yet", never a pass.
```

```
id: E-023   criterion: plan.md M0 Exit (positive control: the bounded form lists a hash on a range that holds a change)   tree: 2ebc10f8f
command: git log --reverse --format=%h fe211e9c9~1..fe211e9c9 -- internal/harness/retention.go
stdout:
fe211e9c9
exit: 0
```

```
id: E-024   criterion: plan.md M0 Exit (positive control: the ordering test can fail)   tree: 2ebc10f8f
command: git merge-base --is-ancestor 2ebc10f8f 68f023289
stdout: (empty)
exit: 1
note: a deliberately mis-ordered pair (a later commit as the first argument); 68f023289 is the first commit of E-021.
command: git merge-base --is-ancestor 68f023289 2ebc10f8f
stdout: (empty)
exit: 0
note: the correctly ordered pair.
```

```
id: E-025   criterion: spec.md §A, §E, §F (the FIFO hang, plan-audit iteration 2 finding B3; the probe is not a criterion RED)   tree: 2ebc10f8f
command: go test -count=1 -v -overlay .moai/reports/t1432/red-now-drafts/overlay-b3.json -run '^TestZZProbeB3FifoAtStatePath$' ./internal/harness/
stdout:
=== RUN   TestZZProbeB3FifoAtStatePath
    zz_fifo_state_test.go:29: PROBE-B3 BLOCKED: PruneStaleEntries did not return within 3s on a FIFO state path
--- PASS: TestZZProbeB3FifoAtStatePath (3.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	3.797s
exit: 0
note: the probe passes by design; its verdict is the logged line. It makes a FIFO at `<log>.prune-state` and calls `PruneStaleEntries(30)` on the unmodified package. Probe, overlay and an earlier output are in `.moai/reports/t1432/red-now-drafts/` (`zz_b3_fifo_state_test.go`, `overlay-b3.json`, `probe-b3.out`; Gap G-7). Cause READ at `retention.go:94` and `:187-189`.
```

```
id: E-026   criterion: AC-HRH-005 (d) premise: a user-owned link and its target's owner   tree: 2ebc10f8f
setup: ln -sfn / /private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/783a9ff3-1235-40b6-8b28-8f6b61be9e2c/scratchpad/t1432-iter3/link-to-root (run by the current user)
command: stat -f 'link-own: %Sp uid=%u' /private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/783a9ff3-1235-40b6-8b28-8f6b61be9e2c/scratchpad/t1432-iter3/link-to-root
stdout:
link-own: lrwxr-xr-x uid=501
exit: 0
command: stat -L -f 'target-own: %Sp uid=%u' /private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/783a9ff3-1235-40b6-8b28-8f6b61be9e2c/scratchpad/t1432-iter3/link-to-root
stdout:
target-own: drwxr-xr-x uid=0
exit: 0
command: id -u
stdout:
501
exit: 0
note: the link's own owner (501) is the current user and its target's owner (0) is not, so an owner read that follows the link answers differently from the no-follow read. The iteration-2 audit measured the same (its probe P-OWN). darwin only.
```

```
id: E-027   criterion: AC-HRH-005 (e) premise: root-owned symbolic links among the fixed candidates   tree: 2ebc10f8f
command: stat -f '%Sp %u %Su %N' /var /tmp /etc
stdout:
lrwxr-xr-x 0 root /var
lrwxr-xr-x 0 root /tmp
lrwxr-xr-x 0 root /etc
exit: 0
note: darwin only; `stat` without `-L` reports the link itself. Linux is not observed (`/bin` is a root-owned link on a merged-usr layout; the criterion skips where no candidate qualifies).
```

```
id: E-028   criterion: AC-HRH-009 (the four-function history form, green at the plan tip)   tree: 2ebc10f8f
command: git log --format=%h -s -L '/^func (r \*Retention) PruneStaleEntries/,/^}/:internal/harness/retention.go' -L '/^func readStamp(/,/^}/:internal/harness/retention.go' -L '/^func readStampFile(/,/^}/:internal/harness/retention.go' -L '/^func stampIsFresh(/,/^}/:internal/harness/retention.go' 1e2151a38..2ebc10f8f
stdout: (empty)
exit: 0
note: the range 1e2151a38..2ebc10f8f is empty of commits touching the file (E-022), so the empty output is the expected green at base, not yet a pass; it becomes the pass condition on the final tree.
```

```
id: E-029   criterion: AC-HRH-009 (positive control: the four-function form prints a hash on a range that holds a change)   tree: 2ebc10f8f
command: git log --format=%h -s -L '/^func (r \*Retention) PruneStaleEntries/,/^}/:internal/harness/retention.go' -L '/^func readStamp(/,/^}/:internal/harness/retention.go' -L '/^func readStampFile(/,/^}/:internal/harness/retention.go' -L '/^func stampIsFresh(/,/^}/:internal/harness/retention.go' fe211e9c9~1..fe211e9c9
stdout:
fe211e9c9
exit: 0
note: fe211e9c9 added `readStamp`, `readStampFile` and `stampIsFresh` and edited `PruneStaleEntries`; a single-function form on each of the three helpers over the history up to the base printed exactly `fe211e9c9` (observed in 0.3.0), so each of the three regular expressions matches.
```

```
id: E-030   criterion: AC-HRH-009 (positive control: out-of-order -L options fail loudly)   tree: 2ebc10f8f
command: git log --format=%h -s -L '/^func readStampFile(/,/^}/:internal/harness/retention.go' -L '/^func readStamp(/,/^}/:internal/harness/retention.go' 1e2151a38..2ebc10f8f
stdout:
fatal: -L parameter '^func readStamp(' starting at line 195: regexec() failed to match
exit: 128
note: a later `-L` option's search starts after the end of the previous range, so the options must follow the order of the functions in the file; a missing or renamed function fails the same way, so the sweep cannot shrink silently.
```

```
id: E-031   criterion: document-level pin (the plan tip's Go code equals the base)   tree: 2ebc10f8f
command: git diff --quiet 1e2151a38 2ebc10f8f -- internal cmd
stdout: (empty)
exit: 0
```

Gaps recorded by the ledger:

- **G-1** AC-HRH-004, -005, -006 have no RED-now: the owner seam, the real owner check and the heal helper exist only after the run phase. They are regression-guards whose adoption proof is the mutation runs named in their green-path cells, observed at M2.
- **G-2** AC-HRH-014's kill observation needs the extracted locked phase (M3). Only the survivor (E-008) is observed.
- **G-3** The drafts and the mutant files are hoisted to `.moai/reports/t1432/red-now-drafts/` (tracked), but the overlay JSON files there still name the plan phase's session scratch copies (see the Command form). The leader or M0 regenerates them against the committed sources; until then the overlay commands run only where that scratch directory exists.
- **G-4** E-003 to E-005 depend on a bounded delay (300 ms after the stamp appears); a slow runner can only make a draft pass vacuously, never fail it falsely.
- **G-5** No Linux or Windows observation exists for any entry; every run was on darwin, uid 501, `go1.26.8`.
- **G-6** Refused command, recorded per `verification-claim-integrity.md` §3.1: the first re-run of E-009 in 0.3.0, written with `| head` to bound its output, was refused by the worktree guard ("construct too complex to verify"); it was re-run as the plain single command shown, whose full output was read. The kanban environment variables were not scrubbed in the 0.3.0 re-runs (a separate or compound `unset` form is refused by the guard); the runs read no kanban state, and their failing lines equal the earlier scrubbed runs'.
- **G-7** The B3 probe files (`zz_b3_fifo_state_test.go`, `overlay-b3.json`, `probe-b3.out`) are in `.moai/reports/t1432/red-now-drafts/` but untracked at this revision; the leader force-adds them. The overlay names a scratch copy of the probe (`zz_fifo_state_test.go`), identical in content.
- **G-8** E-026 and E-027 are darwin only: the Linux owner of root-owned symbolic links, and whether a Linux CI runner finds a candidate for AC-HRH-005 (e), are not observed.

## Evidence ledger — amendment 0.4.0 [0.4.0]

Entries E-032 to E-046, all measured in this session on tree `7639c04c1` (`git rev-parse --short HEAD` printed `7639c04c1`, `git branch --show-current` printed `WT-harness-retention-debt`, `git status --short` empty before the runs and empty after them). `S` below abbreviates the session scratch directory `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/783a9ff3-1235-40b6-8b28-8f6b61be9e2c/scratchpad`; every command spells it out literally in the run, and the cited commands here replace it by `S` only to keep lines short (the draft files are under `S/t1432-amend/`; the audit's probe under `S/audit/`). Package test runs held the slot lease `go-test-internal-harness` (acquired and released; `moai slot status` read `free` before and after). Each package test command was run as one compound invocation `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test …; echo "exit: $?"`; the `unset` prefix and the `echo` suffix are not part of the cited command, and the `exit:` field is that suffix's value. Where no `exit:` line was printed the tool reports a non-zero status explicitly and printed none, which is exit 0. Overlay commands read overlay JSON files that map a path inside `internal/harness/` to a draft or mutant source in `S/t1432-amend/` (Gap G-9); the committed tree was never modified.

```
id: E-032   criterion: spec.md §A heal-lock primitives (platform premises of AC-HRH-015, AC-HRH-016)   tree: 7639c04c1
command: go run S/t1432-amend/prim/main.go
note: a standalone stdlib program (draft instrument, not in the tree); darwin arm64, uid 501, go1.26.8; it opens two descriptors of one file, a symbolic link, a FIFO, a directory and a dangling link, as the heal-lock helper would
stdout:
P1 first_lock_err=<nil> second_nb_err=resource temporarily unavailable second_is_EWOULDBLOCK=true
P1 after_first_closed_second_nb_err=<nil>
P2 open_symlink_nofollow_err=open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/prim5270296/link: too many levels of symbolic links opened=false
P2 victim_bytes="VICTIM-BYTES"
P3 open_fifo_returned=true stat_err=<nil> is_regular=false mode=prw-r--r--
P4 open_dir_rdwr_err=open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/prim5270296/adir: is a directory opened=false
P5 create_excl_over_dangling_err=open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/prim5270296/dangling: file exists opened=false target_created=false
exit: 0
```

```
id: E-033   criterion: spec.md §A heal-lock primitives (compile reach)   tree: 7639c04c1
command: env GOOS=linux go vet S/t1432-amend/prim/main.go
stdout: (empty)
exit: 0
note: compile and vet only, for the symbols syscall.Flock, LOCK_NB, O_NOFOLLOW, O_NONBLOCK, EWOULDBLOCK; Linux runtime not observed.
command: env GOOS=windows go doc syscall.Flock
stdout:
doc: no symbol Flock in package syscall
exit: 1
command: env GOOS=windows go doc syscall.O_NOFOLLOW
stdout:
doc: no symbol O_NOFOLLOW in package syscall
exit: 1
```

```
id: E-034   criterion: spec.md §A "Heal window" (the first audit probe re-run unchanged)   tree: 7639c04c1
command: go test -count=1 -v -timeout 150s -overlay S/audit/ov_race.json -run '^TestZZProbeRemoveWindow$' ./internal/harness/
note: the audit's probe (`S/audit/zz_audit_race_test.go`, one swap per trial against the bare helper removeStateEntryIfUnchanged, stops after 5 violations); not mine, re-run unchanged. The figure is the trial count at which the 5th removal occurred, not a rate.
stdout:
=== RUN   TestZZProbeRemoveWindow
    zz_audit_race_test.go:67: PROBE remove-window: trials=2135 helper_removed_the_swapped_in_fresh_entry=5 helper_first=2119 swap_first=11
--- PASS: TestZZProbeRemoveWindow (7.28s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	8.318s
exit: 0
```

```
id: E-035   criterion: spec.md §A "Heal window"; AC-HRH-006 (b), definition of done 10 (the pre-fix control of the post-fix probe)   tree: 7639c04c1
command: go test -count=1 -v -timeout 120s -overlay S/t1432-amend/overlay-healprobe.json -run '^TestZZProbeHealWindow$' ./internal/harness/
note: draft `zz_heal_race_probe_test.go`. Through the production entry point healStateEntry; the swapper takes an exclusive flock on <log>.prune-heal around its rename of a fresh file over the inspected link, i.e. it behaves as a healer that honours the heal lock. At this tree the production code ignores that lock, so the result is the pre-fix control. Stops after 5 violations (a violation is a trial after which the state path is absent: the heal removed the swapped-in file).
stdout:
=== RUN   TestZZProbeHealWindow
    zz_heal_race_probe_test.go:93: PROBE heal-window: trials=1270 heal_removed_the_swapped_in_fresh_entry=5 path_holds_F=1265 path_holds_other=0 heal_errors=0
--- PASS: TestZZProbeHealWindow (4.12s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	5.243s
exit: 0
```

```
id: E-036   criterion: AC-HRH-006 (b)   tree: 7639c04c1
command: go test -count=1 -overlay S/t1432-amend/overlay-heallock.json -run '^TestPruneHealSerializesOnTheHealLock$' ./internal/harness/
note: draft `zz_heallock_test.go`; the test takes the heal lock itself (syscall.Flock on the literal <log>.prune-heal), so it compiles against the unmodified package
stdout:
--- FAIL: TestPruneHealSerializesOnTheHealLock (0.01s)
    zz_heallock_test.go:102: the pruner returned (<nil>) while another descriptor held the heal lock: it did not wait
    zz_heallock_test.go:106: the state-path entry was removed or replaced while the heal lock was held by another descriptor: err=<nil>
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/harness	0.789s
FAIL
exit: 1
```

```
id: E-037   criterion: AC-HRH-015   tree: 7639c04c1
command: go test -count=1 -overlay S/t1432-amend/overlay-heallock.json -run '^TestPruneHealLockHeldPastTheBoundFailsClosed$' ./internal/harness/
note: the temporary-directory path in the messages is trimmed to `<tmp>`; everything else is verbatim. This is the second run, with the 8 s hang cap added to the draft; the first run (without the cap) printed the same failing lines.
stdout:
--- FAIL: TestPruneHealLockHeldPastTheBoundFailsClosed (0.03s)
    zz_heallock_test.go:187: want an error naming <tmp>/001/usage-log.jsonl.prune-heal, got <nil>
    zz_heallock_test.go:190: the state-path entry was removed or replaced although the heal lock was not acquired: err=<nil>
    zz_heallock_test.go:196: log changed
    zz_heallock_test.go:199: archive directory exists: <nil>
    zz_heallock_test.go:202: want exactly one warning line naming <tmp>/001/usage-log.jsonl.prune-heal, got ""
FAIL
FAIL	github.com/modu-ui/moai-adk/internal/harness	0.892s
FAIL
exit: 1
```

```
id: E-038   criterion: AC-HRH-016   tree: 7639c04c1
command: go test -count=1 -v -overlay S/t1432-amend/overlay-heallock.json -run '^TestPruneHealLockHostileEntryFailsClosed$' ./internal/harness/
note: trimmed to the deciding lines: the symlink subtest's four failure lines (the other three subtests print the same four lines with their own paths) and the summary; `<tmp>` stands for the temporary directory
stdout:
=== RUN   TestPruneHealLockHostileEntryFailsClosed/symlink
    zz_heallock_test.go:277: want an error naming <tmp>/001/usage-log.jsonl.prune-heal, got <nil>
    zz_heallock_test.go:280: the state-path entry was removed or replaced although the heal lock was unusable: err=<nil>
    zz_heallock_test.go:286: log changed
    zz_heallock_test.go:289: want exactly one warning line naming <tmp>/001/usage-log.jsonl.prune-heal, got ""
--- FAIL: TestPruneHealLockHostileEntryFailsClosed (0.10s)
    --- FAIL: TestPruneHealLockHostileEntryFailsClosed/symlink (0.02s)
    --- FAIL: TestPruneHealLockHostileEntryFailsClosed/directory (0.00s)
    --- FAIL: TestPruneHealLockHostileEntryFailsClosed/fifo (0.03s)
    --- FAIL: TestPruneHealLockHostileEntryFailsClosed/not-owned (0.05s)
FAIL
FAIL	github.com/modu-ui/moai-adk/internal/harness	0.947s
FAIL
exit: 1
```

```
id: E-039   criterion: AC-HRH-006 (c) and (d), green at base (regression-guards)   tree: 7639c04c1
command: go test -count=1 -v -overlay S/t1432-amend/overlay-heallock.json -run '^TestPruneCommonPathCreatesNoHealLock$' ./internal/harness/
stdout:
=== RUN   TestPruneCommonPathCreatesNoHealLock
=== RUN   TestPruneCommonPathCreatesNoHealLock/absent
=== RUN   TestPruneCommonPathCreatesNoHealLock/healthy-expired-stamp
--- PASS: TestPruneCommonPathCreatesNoHealLock (0.07s)
    --- PASS: TestPruneCommonPathCreatesNoHealLock/absent (0.07s)
    --- PASS: TestPruneCommonPathCreatesNoHealLock/healthy-expired-stamp (0.01s)
PASS
ok  	github.com/modu-ui/moai-adk/internal/harness	1.044s
exit: 0
command: go test -count=1 -v -overlay S/t1432-amend/overlay-heallock.json -run '^TestPruneHealLockIsFreeWhilePrunerIsInItsArchiveStep$' ./internal/harness/
stdout:
=== RUN   TestPruneHealLockIsFreeWhilePrunerIsInItsArchiveStep
=== PAUSE TestPruneHealLockIsFreeWhilePrunerIsInItsArchiveStep
=== CONT  TestPruneHealLockIsFreeWhilePrunerIsInItsArchiveStep
--- PASS: TestPruneHealLockIsFreeWhilePrunerIsInItsArchiveStep (0.31s)
PASS
ok  	github.com/modu-ui/moai-adk/internal/harness	0.926s
exit: 0
note: both pass at base because no heal lock exists; (d) is vacuous at base and earns adoption from its mutant (heal lock held across the prune), to be run at M8. A swept-count run of all five draft test names printed `--- FAIL` for AC-HRH-006 (b), AC-HRH-015, AC-HRH-016 (and its four subtests) and `--- PASS` for (c) and (d), so no name selected zero tests.
```

```
id: E-040   criterion: AC-HRH-006 (a), green at base   tree: 7639c04c1
command: go test -count=1 -v -run '^(TestHealDoesNotRemoveAFreshStateFile|TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal)$' ./internal/harness/
stdout:
=== RUN   TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal
=== PAUSE TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal
=== RUN   TestHealDoesNotRemoveAFreshStateFile
=== PAUSE TestHealDoesNotRemoveAFreshStateFile
=== CONT  TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal
=== CONT  TestHealDoesNotRemoveAFreshStateFile
--- PASS: TestHealDoesNotRemoveAFreshStateFile (0.01s)
--- PASS: TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal (0.02s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	1.198s
exit: 0
```

```
id: E-041   criterion: AC-HRH-010, the ninth sentinel   tree: 7639c04c1
command: grep -c -F "heal lock gives no exclusion on Windows" internal/harness/retention.go
stdout:
0
exit: 1
```

```
id: E-042   criterion: definition of done 11 (the heal-lock path is gitignored)   tree: 7639c04c1
command: git check-ignore -v .moai/harness/usage-log.jsonl.prune-heal
stdout: (empty)
exit: 1
command: git check-ignore -v .moai/harness/usage-log.jsonl.prune-state
stdout:
.gitignore:352:.moai/harness/usage-log.jsonl.prune-state	.moai/harness/usage-log.jsonl.prune-state
exit: 0
note: the second command is the positive control: the state-file path is ignored, by the exact line `.gitignore:352`; the heal-lock path is not ignored at this tree.
command: grep -n "harness\|prune" internal/template/templates/.gitignore
stdout:
219:!.agents/skills/moai-harness/
exit: 0
note: the distributed template carries no `.prune-state` entry (its one `harness` line is an unrelated allow-list line), so the template is not changed.
```

```
id: E-043   criterion: AC-HRH-011 clause (heal-lock files), controls   tree: 7639c04c1
command: env GOOS=windows go vet ./internal/harness/ ./internal/lockfile/
stdout: (empty)
exit: 0
command: env GOOS=windows go vet -overlay S/t1432-amend/overlay-untagged-flock.json ./internal/harness/
stdout:
# github.com/modu-ai/moai-adk/internal/harness [github.com/modu-ai/moai-adk/internal/harness.test]
S/t1432-amend/zz_untagged_flock_test.go:19:14: undefined: syscall.Flock
S/t1432-amend/zz_untagged_flock_test.go:19:41: undefined: syscall.LOCK_EX
S/t1432-amend/zz_untagged_flock_test.go:19:57: undefined: syscall.LOCK_NB
exit: 1
command: env GOOS=windows go vet -overlay S/t1432-amend/overlay-heallock.json ./internal/harness/
stdout: (empty)
exit: 0
note: the second command is the control (a test file calling syscall.Flock with no //go:build !windows constraint breaks the Windows vet); the third injects the tagged draft heal-lock tests and passes. Build and vet only; Windows runtime not observed.
```

```
id: E-044   criterion: AC-HRH-006 (e), AC-HRH-008 (d) (survivor RED: the three test-debt mutants pass the whole package)   tree: 7639c04c1
note: each mutant is a copy of the tree's `retention.go` edited in the scratch directory (`S/t1432-amend/mut/`), applied with an overlay; the tree never held a mutant. N1: the identity comparison in removeStateEntryIfUnchanged reduced to the modification time only (violates D4.c). N2: appendLogTail writes a newline when the tail is empty. N4: the "file" arm of healStateEntry calls os.Remove directly (the symbolic-link arm keeps the conditional removal). The same three survive in `sync-audit-delta.md` (D4, D6, D5; its figures not relied on).
command: go test -count=1 -overlay S/t1432-amend/mut/overlay-N1.json ./internal/harness/
stdout:
ok  	github.com/modu-ai/moai-adk/internal/harness	6.144s
exit: 0
command: go test -count=1 -overlay S/t1432-amend/mut/overlay-N2.json ./internal/harness/
stdout:
ok  	github.com/modu-ai/moai-adk/internal/harness	6.691s
exit: 0
command: go test -count=1 -overlay S/t1432-amend/mut/overlay-N4.json ./internal/harness/
stdout:
ok  	github.com/modu-ai/moai-adk/internal/harness	8.247s
exit: 0
```

```
id: E-045   criterion: AC-HRH-006 (e), AC-HRH-008 (d) (kill observation with drafts)   tree: 7639c04c1
command: go test -count=1 -v -overlay S/t1432-amend/overlay-m9.json -run '^(TestHealRemovalIsNotDecidedByModTimeAlone|TestPruneHealFileArmKeepsAFreshStateFileSwappedInDuringTheHeal|TestPruneWithoutLateEventAddsNoBlankLine)$' ./internal/harness/
note: draft `zz_m9_test.go`, unmutated; the three `=== RUN`, `=== PAUSE` and `=== CONT` lines are trimmed
stdout:
--- PASS: TestHealRemovalIsNotDecidedByModTimeAlone (0.03s)
--- PASS: TestPruneWithoutLateEventAddsNoBlankLine (0.03s)
--- PASS: TestPruneHealFileArmKeepsAFreshStateFileSwappedInDuringTheHeal (0.03s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	1.060s
exit: 0
command: go test -count=1 -v -overlay S/t1432-amend/mut/overlay-m9-N1.json -run '^TestHealRemovalIsNotDecidedByModTimeAlone$' ./internal/harness/
stdout (trimmed to the failing lines):
    zz_m9_test.go:56: removal of a changed entry with an equal modification time: removed=true err=<nil>, want false and nil
    zz_m9_test.go:59: fresh state file was removed or changed: ""
--- FAIL: TestHealRemovalIsNotDecidedByModTimeAlone (0.02s)
exit: 1
command: go test -count=1 -v -overlay S/t1432-amend/mut/overlay-m9-N2.json -run '^TestPruneWithoutLateEventAddsNoBlankLine$' ./internal/harness/
stdout (trimmed to the failing lines):
    zz_m9_test.go:135: replacement log = "{\"timestamp\":\"2026-10-01T00:00:00Z\",\"event_type\":\"feedback\",\"subject\":\"fresh\",\"context_hash\":\"h\",\"tier_increment\":0,\"schema_version\":\"v2.1\"}\n\n": want exactly one newline-terminated kept line and no blank line
--- FAIL: TestPruneWithoutLateEventAddsNoBlankLine (0.02s)
exit: 1
command: go test -count=1 -v -overlay S/t1432-amend/mut/overlay-m9-N4.json -run '^TestPruneHealFileArmKeepsAFreshStateFileSwappedInDuringTheHeal$' ./internal/harness/
stdout (trimmed to the failing lines):
    zz_m9_test.go:109: the fresh state file was removed or replaced by the heal's file arm: err=<nil>
--- FAIL: TestPruneHealFileArmKeepsAFreshStateFileSwappedInDuringTheHeal (0.03s)
exit: 1
```

```
id: E-046   criterion: definition of done 9 (the amendment's ordering check, and its controls)   tree: 7639c04c1
command: git log --reverse --format=%h 7639c04c1..HEAD -- internal/harness/retention.go internal/harness/retention_heal_unix.go internal/harness/retention_heal_windows.go
stdout: (empty)
exit: 0
note: the bounded listing at the amendment base is empty: no heal-lock production commit exists yet. An empty listing means "no production commit yet", never a pass (it is expected at M7 itself, and is a Gap if still empty at M10).
command: git log --reverse --format=%h 1e2151a38..7639c04c1 -- internal/harness/retention.go
stdout:
c1cc3fe67
f52dd1b1c
b3a469eab
9289a92b6
exit: 0
note: positive control: the bounded form lists the card's four earlier production commits on a range that holds changes.
command: git merge-base --is-ancestor 9289a92b6 ac40cf3bf
stdout: (empty)
exit: 1
command: git merge-base --is-ancestor ac40cf3bf 9289a92b6
stdout: (empty)
exit: 0
note: positive controls: a mis-ordered pair (the later commit first) exits 1, the correctly ordered pair exits 0.
```

Gaps recorded by the amendment ledger:

- **G-9** The drafts, probes and mutant copies of E-032 to E-046 live in the session scratch directory, which is session-scoped and untracked; the overlay JSON files name it by absolute path. The leader or M7 hoists them to a tracked evidence path (as `red-now-drafts/` did for 0.3.0) and regenerates the overlay files; until then the commands run only where that directory exists.
- **G-10** Every entry is darwin arm64, uid 501, go1.26.8 on APFS; Linux and Windows runtime are not observed (E-033 is compile-only for Linux). The kernel's release of a lock when its holder is killed is reasoned from `flock` semantics, not observed.
- **G-11** AC-HRH-006 (b), AC-HRH-015 and AC-HRH-016 have a RED-now (E-036 to E-038) but no kill observation: the heal-lock code does not exist, so no mutant of it can be built. Their adoption proof is the mutation runs named in their cells, to be observed at M8. AC-HRH-006 (c) and (d) are green at base and vacuous there (E-039).
- **G-12** The timing assertions (the 300 ms of AC-HRH-006 (b), the 4.5 s ceiling of AC-HRH-015, the 4.5 s cap of AC-HRH-016) were not run under machine load; background load is prohibited here. The 2 s bound is an engineering choice, not measured under load (decision-index Q9).
- **G-13** Refused command, recorded per `verification-claim-integrity.md` §3.1: the swept-count run of the five heal-lock draft tests, first written as `go test … | grep …`, was refused by the worktree guard ("construct too complex to verify"); it was re-run with the output redirected to a scratch file and the top-level result lines read with a plain `grep` (the E-039 note records what it showed). The post-fix race probe (definition of done 10) and the mutant kills of the heal lock are expectations until M8 and M10.

## Definition of Done

1. The commit that adds the new failing tests is an ancestor of every commit that touches `internal/harness/retention.go`, witnessed on the commit graph and evaluated on the card branch before it is merged into develop: `git log --diff-filter=A --format=%h -- internal/harness/retention_statepath_test.go` names the tests commit T; the commits touching production code are listed by `git log --reverse --format=%h 1e2151a38..HEAD -- internal/harness/retention.go` (bounded to the card's own range: the unbounded form lists seven commits older than the base, E-021, for which no check of T can pass), and `git merge-base --is-ancestor T <each>` exits 0 for every one. An empty listing means no production commit exists yet and is a Gap, not a pass (it is expected to be empty at M0 itself, E-022). The check can fail: `git merge-base --is-ancestor` exits 1 on a mis-ordered pair, and the bounded form lists a hash on a range that holds a change (E-024, E-023). `.moai/reports/t1432/red-baseline.md` and every other `.moai/reports/t1432/*` file are committed with `git add -f` (the path is ignored by `.gitignore:235`; `git ls-files .moai/reports/t1432` must list each) in commit T; the ignored report is evidence content, the tracked test files are the witness.
2. AC-HRH-001..014 each PASS with the command and verbatim output recorded in `verdict.md`; any criterion whose RED was not observed, and any skipped test, is listed as a Gap with its platform and uid.
3. `internal/lockfile` and `observer.go` are byte-identical to `1e2151a38`, and the bodies of `PruneStaleEntries`, `readStamp`, `readStampFile` and `stampIsFresh` are untouched (AC-HRH-009).
4. The residual window, F5 and F6 are disclosed in the source documentation with the labels in `spec.md` §A.
5. No criterion claims a Windows runtime observation; Windows evidence is stated as build and vet only.
6. The three mutation runs of AC-HRH-004, the three of AC-HRH-005 (a check that always answers "owned"; an owner read that follows the link; every symbolic link treated as owned), the one of AC-HRH-006 and the two of AC-HRH-003 (a heal that removes a directory; a heal that ignores the removal failure) are recorded as killed; mB, mC and mD are re-run and shown killed or failing fast.
7. The leader has been told which `spec.md` §B options remain operator-held and unselected, and that the single allowed test-only field was spent on the owner lookup.
8. The completion report states `Windows runtime not observed` in its Gaps section and names the intermediate red commit T by SHA in one line (`spec.md` §D).

Amendment 0.4.0 items (9 to 14) [0.4.0]; items 1 to 8 are the 0.3.0 text and still hold (item 1's bounded range `1e2151a38..HEAD` now also contains the amendment's commits, all descendants of T, so it still passes):

9. The amendment's ordering, witnessed on the commit graph and evaluated on the card branch before it is merged into develop: the commit T2 that adds `internal/harness/retention_heallock_test.go` (`git log --diff-filter=A --format=%h -- internal/harness/retention_heallock_test.go` names it) is an ancestor of every commit listed by `git log --reverse --format=%h 7639c04c1..HEAD -- internal/harness/retention.go internal/harness/retention_heal_unix.go internal/harness/retention_heal_windows.go`, checked with `git merge-base --is-ancestor T2 <each>`. An empty listing is a Gap, not a pass (it is empty at M7 itself, E-046). The controls are E-046: a bounded range that holds changes lists their hashes, and a mis-ordered pair exits 1. `.moai/reports/t1432/red-baseline-amend.md` and every other new `.moai/reports/t1432/*` file are committed with `git add -f` in T2 (the path is ignored by `.gitignore:235`).
10. The post-fix race evidence is a measurement, not the expectation written here: the draft `zz_heal_race_probe_test.go` (adapted if the signature of the heal entry point changed), run against the final tree, through the production heal entry point and with a swapper that honours the lock, over at least 20000 trials with no early stop, counts 0 removals of the swapped-in file. Two controls, run in the same session against the same final tree, show that the probe can fail and that the lock is what makes it pass: the same heal probe with a swapper that does NOT take the lock still counts removals (a process that does not honour the lock defeats it), and the audit's unchanged bare-helper probe (E-034) still counts removals (the removal helper is unchanged; the lock is taken by its caller). Pre-fix reference: 5 removals in 1270 trials (E-035). The result is pasted verbatim into the verdict with its trial count; a probe that stopped early or ran fewer trials is a Gap.
11. The repository `.gitignore` ignores the heal-lock path: `git check-ignore -v .moai/harness/usage-log.jsonl.prune-heal` exits 0 (RED-now: exit 1, E-042). The distributed template `.gitignore` is not changed (E-042: it carries no `.prune-state` entry either).
12. AC-HRH-006 (b), (c), (d), AC-HRH-015 and AC-HRH-016 each PASS with the command and verbatim output in `verdict.md`; the mutants named in their cells are recorded killed, and the limit of AC-HRH-006 (a lock released before the removal is not killed by any deterministic test) is stated in the verdict. The optional M9 cases, AC-HRH-006 (e) and AC-HRH-008 (d), are either PASS with their kills recorded or listed as carried debt.
13. The sync phase of the amendment hands the following to the owners named in the Status Transition Ownership Matrix, not to this SPEC's run phase: the `CHANGELOG.md` sentence that says the heal race "narrows but does not close" (manager-docs, at the amended sync), and the status of `.moai/reports/t1432/residual-risk-removal-window.md` (superseded by the amendment; the leader session decides the note).
14. The completion report states `Windows runtime not observed` in its Gaps section, names the intermediate red commit T2 by SHA in one line, and lists every skipped test with its platform and uid.
