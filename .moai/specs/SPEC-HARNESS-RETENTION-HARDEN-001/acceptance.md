# SPEC-HARNESS-RETENTION-HARDEN-001 — Acceptance Criteria (card t1432)

Verification layer for `spec.md` §C. Every criterion is `Given … When … Then …` and binary-testable. Test names are handles for the run phase, not requirements on identifiers. Evidence rules follow `.claude/rules/moai/development/verification-completeness.md` §2 and §2.1.

**Document-level pin.** Every RED-now cell below was measured on tree `db6d88a2a` (committed HEAD of `WT-harness-retention-debt` when the plan was first written; its Go code is identical to base `1e2151a38`, `git diff --quiet 1e2151a38 db6d88a2a -- internal cmd` exit 0). A ledger entry that carries another SHA says so: the entries re-run for revision 0.3.0 carry `2ebc10f8f`, whose Go code is also identical to base (ledger E-031).

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
| AC-HRH-006 | REQ-HRH-005 | heal safety | RG | not obtainable: helper absent (Gap G-1) | `--- PASS` at M2 plus one mutant killed at M2 |
| AC-HRH-007 | REQ-HRH-006 | event loss | RB | E-003 (exit 1, `late-event count = 0`) | flips at M1: `--- PASS`, exit 0 |
| AC-HRH-008 | REQ-HRH-007 | event loss, terminator | RB (cases b, c); RG (case a) | E-004, E-005 (exit 1); case a green at base | flips at M1: `--- PASS`, exit 0 |
| AC-HRH-009 | REQ-HRH-008 | append and pre-lock path | RG | E-013 green at base, E-028 (the `-L` form); controls E-014, E-029, E-015, E-030 | stays exit 0 |
| AC-HRH-010 | REQ-HRH-009, 010, 011 | disclosure | RB | E-010 (`0`, exit 1), E-011 (`1`, exit 0) | flips at M5: each sentinel count ≥ 1, old sentence 0 |
| AC-HRH-011 | REQ-HRH-012 | cross-platform | RG | E-016 green at base; controls E-017, E-018 | stays exit 0 |
| AC-HRH-012 | REQ-HRH-013 | N1 | RB | E-009 (exit 1, `panic: test timed out after 40s`) | flips at M4: fails in ≤ 15 s under mD |
| AC-HRH-013 | REQ-HRH-014 | N2 / mB | RB (survivor RED) | E-006 survives (exit 0), E-007 killed by the draft (exit 1) | lands at M0: `--- PASS` unmutated (E-020), fails under mB |
| AC-HRH-014 | REQ-HRH-015 | N2 / mC | RB (survivor RED) | E-008 survives (exit 0); kill not obtainable (Gap G-2) | flips at M3: `--- PASS` unmutated, fails under mC |

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

### AC-HRH-006 — A concurrent healer's fresh state file is never removed (REQ-HRH-005)

- **Given** a symbolic link owned by the user at the state path that the heal helper has inspected, and then a stand-in for a concurrent healer that creates a healthy regular state file elsewhere in the directory and renames it over the link (so the new file's identity differs from the inspected link's),
- **When** the helper is asked to remove the inspected entry,
- **Then** it does not remove the new file (same identity and bytes afterwards) and reports that the entry changed; with the entry unchanged the same call removes it.
- Verify: `go test -count=1 -v -run '^TestHealDoesNotRemoveAFreshStateFile$' ./internal/harness/` prints `--- PASS` and exits 0. The rename keeps the old link alive while the new file is created, so the two identities cannot coincide through inode reuse.
- Class RG: the helper is absent at base (G-1). Adoption proof at M2: a mutant with an unconditional removal fails. This deterministic interleaving is the proof for two concurrent healers; the burst of N processes is not exercised and not measured (`spec.md` §F).

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
- Sentinels: `residual window` (REQ-HRH-009); `no cross-process exclusion`, `burst of hook processes`, `F5: not reproduced, not measured` (REQ-HRH-010); `lock waiters block with no timeout`, `5 s hook timeout`, `appended before the wait`, `F6: not reproduced, not measured` (REQ-HRH-011).
- Verify, one command per sentinel: `grep -c -F "<sentinel>" internal/harness/retention.go` prints a number ≥ 1 for each; `grep -c -F "events other hooks append in that window are lost" internal/harness/retention.go` prints 0; `grep -c -E 't1[0-9]{3}' internal/harness/retention.go` prints 0 (base: `0`, exit 1, so the guard keeps card ids out of source; card history belongs to the decision record).
- RED-now: E-010 (the eight sentinels together match 0 lines at base, exit 1) and E-011 (the old sentence matches 1 line, exit 0). The `async` obligation of REQ-HRH-011 and the sentence-level reading are reviewer-read; this is stated as the pass condition, not hidden.

### AC-HRH-011 — The shared lock package is untouched and Windows builds, vets and compiles the tests (REQ-HRH-012)

- **Given** the final tree,
- **Then** every file under `internal/lockfile` is byte-identical to the base, and `internal/harness` and `internal/lockfile` build and vet for Windows, vet compiling their test files, so a FIFO test without a `//go:build !windows` constraint, or a POSIX-only production file without a Windows twin, fails here.
- Verify: `git diff --quiet 1e2151a38 -- internal/lockfile` exits 0; `GOOS=windows go build ./internal/harness/ ./internal/lockfile/` exits 0; `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` exits 0. Verification level: build and vet only (no Windows runtime).
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

## Edge cases

- The effective uid is 0, or the platform is Windows: permission, ownership and symbolic-link tests skip or are excluded by their build constraint; the build/vet criterion (AC-HRH-011) still runs.
- The state path is a symbolic link to a directory or to a path that does not exist: the link (owned by the current user) is removed, never followed, and the prune continues; the target is untouched. A user-owned link to a root-owned directory such as `/` is healed the same way, because the link's own owner decides (AC-HRH-005 (d)).
- A FIFO at the state path, or a link to a FIFO: outside this SPEC. `PruneStaleEntries` blocks in the lock-free pre-check (OBSERVED, E-025; `spec.md` §E, §F); no criterion pins or repairs it and none uses a FIFO at the state path.
- Two processes meet the same fault at the same instant: the pruner removes an entry only if it is still the one it inspected; the residual gap and the unmeasured burst are disclosed (`spec.md` §F, AC-HRH-006).
- A late line is malformed JSON: it is carried verbatim like any other tail bytes, with the terminator the rule adds only when missing.
- Late events older than the retention cutoff: carried, kept in the log, and archived by a later interval's prune.
- The final unterminated line is the only stale event: the prune classifies nothing as stale, returns without a rewrite and adds no terminator, so the line stays unclassified until an append completes it or a later prune rewrites the log (`spec.md` §B D3); stated, not pinned by a test.

## Quality gate criteria

- `go vet ./internal/harness/ ./internal/lockfile/` exit 0; the project linter at the CI-pinned version reports no new finding in `retention.go` or the touched tests (state the judging build next to the tree HEAD, `verification-claim-integrity.md` §2.2).
- `go test -race -count=1 -v ./internal/harness/` exits 0 and every pre-existing `TestPrune*` function passes; `git diff --name-only 1e2151a38 -- 'internal/harness/*_test.go'` lists only `retention_killed_test.go` among test files that existed at the base.
- `internal/harness` package coverage stays at or above the 85 percent TRUST 5 floor (the t1425 delta audit measured 87.5 percent, `sync-audit-delta.md`); the new branches (heal, ownership refusal, tail-carry, locked-phase seam) are covered by AC-HRH-001..008 and -014.
- `go test -race -count=20` on the killed-pruner and concurrent-process tests exits 0.

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

## Definition of Done

1. The commit that adds the new failing tests is an ancestor of every commit that touches `internal/harness/retention.go`, witnessed on the commit graph and evaluated on the card branch before it is merged into develop: `git log --diff-filter=A --format=%h -- internal/harness/retention_statepath_test.go` names the tests commit T; the commits touching production code are listed by `git log --reverse --format=%h 1e2151a38..HEAD -- internal/harness/retention.go` (bounded to the card's own range: the unbounded form lists seven commits older than the base, E-021, for which no check of T can pass), and `git merge-base --is-ancestor T <each>` exits 0 for every one. An empty listing means no production commit exists yet and is a Gap, not a pass (it is expected to be empty at M0 itself, E-022). The check can fail: `git merge-base --is-ancestor` exits 1 on a mis-ordered pair, and the bounded form lists a hash on a range that holds a change (E-024, E-023). `.moai/reports/t1432/red-baseline.md` and every other `.moai/reports/t1432/*` file are committed with `git add -f` (the path is ignored by `.gitignore:235`; `git ls-files .moai/reports/t1432` must list each) in commit T; the ignored report is evidence content, the tracked test files are the witness.
2. AC-HRH-001..014 each PASS with the command and verbatim output recorded in `verdict.md`; any criterion whose RED was not observed, and any skipped test, is listed as a Gap with its platform and uid.
3. `internal/lockfile` and `observer.go` are byte-identical to `1e2151a38`, and the bodies of `PruneStaleEntries`, `readStamp`, `readStampFile` and `stampIsFresh` are untouched (AC-HRH-009).
4. The residual window, F5 and F6 are disclosed in the source documentation with the labels in `spec.md` §A.
5. No criterion claims a Windows runtime observation; Windows evidence is stated as build and vet only.
6. The three mutation runs of AC-HRH-004, the three of AC-HRH-005 (a check that always answers "owned"; an owner read that follows the link; every symbolic link treated as owned), the one of AC-HRH-006 and the two of AC-HRH-003 (a heal that removes a directory; a heal that ignores the removal failure) are recorded as killed; mB, mC and mD are re-run and shown killed or failing fast.
7. The leader has been told which `spec.md` §B options remain operator-held and unselected, and that the single allowed test-only field was spent on the owner lookup.
8. The completion report states `Windows runtime not observed` in its Gaps section and names the intermediate red commit T by SHA in one line (`spec.md` §D).
