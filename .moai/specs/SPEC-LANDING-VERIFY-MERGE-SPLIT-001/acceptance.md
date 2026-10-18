# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — acceptance criteria (v0.1.0)

This file is the verification layer: every entry is an `AC-LVM-XXX` Given-When-Then, binary-testable. The GEARS requirements are `spec.md` §3. Counts: 22 requirements, 22 acceptance criteria (Tier L ceiling 25/25); the evidence ledger below holds 28 command observations because six criteria span two or three Go packages.

## Common setup

- **Test names are the interface.** The names below are proposed identifiers; the run-phase may rename one only with a matching edit here (a rename without it leaves a `--- PASS` line missing, which reads red).
- **Empty-sweep guard (verification-completeness §1.1).** Every Go AC runs `go test ./internal/<pkg> -run '^(A|B|…)$' -count=1 -v` and requires, together: (i) exactly one `--- PASS: <Name> ` line per listed name (the AC states N), (ii) no `[no tests to run]` in the output, (iii) exit 0. A selector that matches zero tests exits 0, so the name count is the verdict.
- **RED-now cells.** Four elements per cell, carried in the evidence ledger at the end of this file (the recommended carrier): the command, its verbatim stdout, its exit code, and the tree SHA — document-level pin `ec164d45d2e3c5a7902aa950ab212e8f385f5d08` (the current main at authoring, which is the base of this card's branch; binds every cell without its own pin). Every Go cell is the command `go test ./internal/<pkg> -list '^(names)$' -count=1`, observed with exit 0 and **zero names listed**. Each is red for one stated reason: the named tests do not exist because the behaviour is not implemented (count 0 ≠ N). This is a count-based red: the command itself exits 0, which is why the criterion is the number of names and never the exit code alone.
- **Judging build.** CLI observations used `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/c5bd52fe-ffbe-40e1-94ce-4491a79c0784/scratchpad/moai-t1872`, built from the same clean tree (`go build ./cmd/moai`, exit 0) — build commit equals the tree HEAD.
- **Green path.** Each AC names the milestone (`plan.md`) that flips it. M0 is the baseline commit; M1 the completeness gate; M2 auto-release; M3 capacity; M4 instrumentation; M5 ready view; M6 reuse; M7 rules and mirrors.
- **Gate-on fixtures.** Gate tests construct temporary git repositories with `t.TempDir()`; none reads this repository's live refs, `.moai/state`, or the user's home.

## Part 6 — completeness gate (M1)

### AC-LVM-001 — manifest derivation (REQ-LVM-001)

- **Given** a temporary repository whose card branch changes (a) a non-test file of package `p/a`, (b) only a test file of package `p/b`, (c) a file of `internal/cli`; and `p/c` imports `p/a`, `p/d` imports `p/b` only in a test, **When** the manifest generator runs, **Then** the manifest lists a whole-package call (no `-run`) for each of `p/a`, `p/b`, `internal/cli`, and for the importers `p/c` and `p/d`; only the `internal/cli` call may carry `-run`, and it names `35e9`; every call records identifier, command and resolved package set.
- Verify (N=3): `go test ./internal/factory -run '^(TestManifestDerivesChangedAndImporters|TestManifestNarrowingOnlyInternalCLI|TestManifestIncludesTestOnlyChange)$' -count=1 -v`.
- RED-now: E-001. Green: M1.

### AC-LVM-002 — attempt record (REQ-LVM-002)

- **Given** a manifest call run through the re-measure verb over a fixture `go test -json` stream with two packages, **When** the attempt finishes, **Then** the attempt record carries a unique attempt ID, call identifier, command, exit code, one terminal event per stream package, tree SHA, HEAD SHA, base, environment fingerprint, start/end times and termination class `completed`.
- Verify (N=2): `go test ./internal/factory -run '^(TestAttemptRecordCarriesAllFields|TestAttemptPackageTerminalEvents)$' -count=1 -v`.
- RED-now: E-002. Green: M1.

### AC-LVM-003 — abort versus restart (REQ-LVM-003)

- **Given** attempts ended by (a) SIGTERM, (b) the verb's timeout, (c) a stream cut mid-line or lacking one expected package's terminal event, **When** each is judged, **Then** each is `aborted`, is retained on disk, and counts as no pass; **and When** the call is run again, **Then** it is a new attempt with a new ID, the aborted record is not overwritten, and only the most recent completed attempt on the same tree counts.
- Verify (N=3): `go test ./internal/factory -run '^(TestAbortedAttemptNeverCounts|TestRestartMakesNewAttemptID|TestAbortClassesSignalTimeoutPartial)$' -count=1 -v`.
- RED-now: E-003. Green: M1.

### AC-LVM-004 — whole manifest, positive control (REQ-LVM-004)

- **Given** a manifest of three calls, **When** all three have a completed exit-0 attempt on the candidate tree with full package coverage, **Then** the verifier reports gate-complete (positive control); **When** only the last-recorded call is green and the other two have no attempt, **Then** it does not (last record alone is insufficient); **When** a call has an aborted attempt followed by a completed one on the same tree, **Then** the completed one counts.
- Verify (N=3): `go test ./internal/factory -run '^(TestManifestWholePassPositiveControl|TestLastRecordAloneInsufficient|TestMostRecentCompletedAttemptCounts)$' -count=1 -v`.
- RED-now: E-004. Green: M1.

### AC-LVM-005 — the nine reject classes (REQ-LVM-005)

- **Given** the AC-LVM-004 positive fixture, **When** each one mutation is applied alone — (1) manifest file removed, (2) one call's attempts removed, (3) one stream truncated, (4) one attempt aborted, (5) one attempt recorded on another tree or HEAD, (6) one attempt's GOFLAGS/tags differ from the manifest fingerprint, (7) a changed non-Go input with no verified reader call, (8) an overlap decision that cannot be established, (9) a call whose package swept zero tests where the manifest does not allow it — **Then** the verifier refuses readiness and refuses reuse in each case, naming the class; and the unmutated fixture still passes (control).
- Verify (N=9): `go test ./internal/factory -run '^(TestGateRejectsMissingManifest|TestGateRejectsMissingCall|TestGateRejectsPartialJSON|TestGateRejectsAbortedAttempt|TestGateRejectsOtherTreeOrHead|TestGateRejectsEnvFingerprintMismatch|TestGateRejectsUnverifiedNonGoInput|TestGateRejectsUndecidableOverlap|TestGateRejectsEmptySweep)$' -count=1 -v`.
- RED-now: E-005. Green: M1.

### AC-LVM-006 — one verifier, every consumer (REQ-LVM-006)

- **Given** gate on and a tree whose only record is a script-wrapped call (`structured_required=false`, exit 0), **When** the merge step, `moai factory complete`, merge-readiness condition (d), and the ready view consult the tree, **Then** all four refuse through the same verifier (same reason text), and a gate-complete tree passes all four.
- Verify: factory (N=2) `go test ./internal/factory -run '^(TestMergeStepConsultsCompletenessVerifier|TestScriptWrappedRecordRefusedWhenGateOn)$' -count=1 -v`; cli (N=1) `go test ./internal/cli -run '^TestFactoryCompleteConsultsCompletenessVerifier$' -count=1 -v`; factorylane (N=1) `go test ./internal/factorylane -run '^TestMergeTripleConditionDUsesVerifier$' -count=1 -v`. The ready-view consumer is asserted by AC-LVM-014's tests.
- RED-now: E-006a, E-006b, E-006c. Green: M1 (the ready-view leg flips at M5).

### AC-LVM-007 — default off, narrowing list (REQ-LVM-007)

- **Given** a fresh config and a project with no `workflow.landing_gate` block, **When** the landing path runs a card whose record is a script-wrapped call, **Then** the outcome equals the pre-change outcome (accepted exactly as today); and the shipped narrowing allowance list is exactly `[internal/cli]`.
- Verify: config (N=2) `go test ./internal/config -run '^(TestLandingGateDefaultsOff|TestRunNarrowingAllowanceInternalCLIOnly)$' -count=1 -v`; factory (N=1) `go test ./internal/factory -run '^TestGateOffMatchesTodaysBehaviour$' -count=1 -v`.
- RED-now: E-007a, E-007b. Green: M1.

## Ordering (REQ-LVM-008)

### AC-LVM-008 — the gate lands first, reuse last, and reuse is inert without it

- **Given** the run-phase branch, **When** the milestone commits are located by subject and compared in the commit graph, **Then** the baseline commit M0 precedes M1, M1 precedes each of M2 to M6, and M5 precedes M6; and the reverse comparison fails (negative control); **and When** reuse is requested with the gate disabled or the source record not gate-complete, **Then** it is refused.
- Verify, one invocation each: `git log --format=%H -1 --grep='^feat(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M<k> '` for k = 1..6 (each must print one 40-hex SHA, exit 0) and `git log --format=%H -1 --grep='^docs(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M0 '`; then `git merge-base --is-ancestor <S1> <Sk>` for k = 2..6 and `<S5> <S6>` and `<S0> <S1>` (each exit 0); negative control `git merge-base --is-ancestor <S6> <S1>` (exit 1); `git ls-files --error-unmatch .moai/specs/SPEC-LANDING-VERIFY-MERGE-SPLIT-001/evidence/m0-baseline.md` (exit 0: the baseline is a tracked file outside `.moai/reports`); functional `go test ./internal/factory -run '^TestReuseInertWithoutGate$' -count=1 -v` (N=1).
- RED-now: `git log --format=%H -1 --grep='^feat(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M1 '` → exit 0, empty stdout (no M1 commit exists); `git merge-base --is-ancestor HEAD HEAD` → exit 0 and `git merge-base --is-ancestor HEAD ec164d45d2` → exit 1 (the two exit-code polarities the check relies on, both observed at the pin); and E-008. Red reason: no milestone commit exists, so there is no SHA to compare. Green: M1..M6 landing in order.

## Part 5 — instrumentation (M4)

### AC-LVM-009 — events and metrics, fail-open (REQ-LVM-009)

- **Given** a landing sequence of three cards in a fixture repository, **When** each becomes ready, has a re-measure executed or a reuse considered, and lands, **Then** one event row is appended per occurrence and the metrics verb prints the landing-interval distribution, ready-wait per card, re-measure execution count, reuse count and shadow reuse-eligibility count equal to the fixture's known values; **and When** the log directory is unwritable, **Then** the landing outcome and exit code are unchanged.
- Verify: factory (N=3) `go test ./internal/factory -run '^(TestLandingEventRowsAppended|TestMetricsReportFields|TestEventLogFailureNeverBlocksLanding)$' -count=1 -v`; cli (N=1) `go test ./internal/cli -run '^TestIntegrationMetricsVerb$' -count=1 -v`.
- RED-now: E-009a, E-009b. Green: M4.

## Part 4 — dead-owner release (M2)

### AC-LVM-010 — auto-release (REQ-LVM-010)

- **Given** a lease record whose pid is a reaped process, **When** `slot status`, `slot acquire` (other session) or the guard reads it, **Then** the record file is removed and an `auto-release` audit row names the holder; **Given** a record with pid 0, **Then** it reads live and is not removed.
- Verify (N=3): `go test ./internal/factory -run '^(TestSlotDeadOwnerAutoReleased|TestSlotUnresolvedOwnerReadsLive|TestSlotAutoReleaseAuditRow)$' -count=1 -v`.
- RED-now: E-010, plus the premise probe of `research.md` §4 (today `slot status` leaves `probe-res.json` in place). Green: M2.

## Part 3 — capacity (M3)

### AC-LVM-011 — capacity semantics (REQ-LVM-011)

- **Given** capacity unset or 1, **Then** the second acquire is refused with exit 3 exactly as today; **Given** capacity 2, **Then** two distinct live holders are admitted and the third refused with exit 3.
- Verify (N=3): `go test ./internal/factory -run '^(TestSlotCapacityOneIsToday|TestSlotCapacityTwoAdmitsTwo|TestSlotCapacityTwoRefusesThird)$' -count=1 -v`.
- RED-now: E-011. Green: M3.

### AC-LVM-012 — headroom admission (REQ-LVM-012)

- **Given** capacity 2 and a first holder, **When** the second acquire arrives with load and available memory readings (injected) (a) inside thresholds, (b) outside, (c) unreadable, or (d) thresholds unset, **Then** only (a) is admitted.
- Verify (N=3): `go test ./internal/factory -run '^(TestSlotExtraHolderNeedsLoadAndMemoryHeadroom|TestSlotExtraHolderRefusedOnUnreadableSignal|TestSlotThresholdsUnsetMeansCapacityOne)$' -count=1 -v`.
- RED-now: E-012. Green: M3.

### AC-LVM-013 — wall-clock-bound exclusion (REQ-LVM-013)

- **Given** capacity 2 and a configured wall-clock-bound package list, **When** two holders declare intersecting package sets, **Then** the second is refused with exit 3 despite free capacity; **When** the sets are disjoint, **Then** it is admitted.
- Verify (N=2): `go test ./internal/factory -run '^(TestSlotRefusesIntersectingWallClockPackages|TestSlotAllowsDisjointPackages)$' -count=1 -v`.
- RED-now: E-013. Green: M3.

## Part 2 — ready view (M5)

### AC-LVM-014 — derived ready set (REQ-LVM-014)

- **Given** four merge-ready fixture cards — one fully prepared, one not containing the tip, one gate-incomplete, one with an audit naming an older HEAD — **When** the ready verb runs, **Then** exactly the first is listed and each other carries its first failing predicate.
- Verify: factory (N=2) `go test ./internal/factory -run '^(TestReadySetListsOnlyFullyPrepared|TestNotReadyCarriesFirstFailingPredicate)$' -count=1 -v`; cli (N=1) `go test ./internal/cli -run '^TestIntegrationReadyVerb$' -count=1 -v`.
- RED-now: E-014a, E-014b. Green: M5.

### AC-LVM-015 — view grants nothing; audit SHA binds (REQ-LVM-015)

- **Given** a listed-ready card, **When** the ready verb is deleted from the flow (or its output tampered), **Then** `complete` reaches the same verdict; no readiness file exists on disk; **Given** an audit verdict whose audited SHA is an ancestor of HEAD with an empty or patch-identical diff, **Then** the card is not ready.
- Verify: factory (N=2) `go test ./internal/factory -run '^(TestReadyViewPersistsNothing|TestAuditSHAMismatchNotReadyEvenIfPatchIdentical)$' -count=1 -v`; cli (N=1) `go test ./internal/cli -run '^TestCompleteIgnoresReadyView$' -count=1 -v`.
- RED-now: E-015a, E-015b. Green: M5.

## Part 1 — reuse (M6, last)

### AC-LVM-016 — reuse record (REQ-LVM-016)

- **Given** a gate-complete source record at base B, a moved set that touches nothing in scope, and a conflict-free re-absorb, **When** the reuse verb runs, **Then** a reuse record keyed by the new tree is written citing the source tree, B, the new base, the moved-commit list, the scope digest, the decision `no-overlap` and reasons.
- Verify (N=2): `go test ./internal/factory -run '^(TestReuseRecordWrittenWhenNoOverlap|TestReuseRecordCitesSourceAndMovedCommits)$' -count=1 -v`.
- RED-now: E-016. Green: M6.

### AC-LVM-017 — scope inventory (REQ-LVM-017)

- **Given** fixtures where a moved file is (a) inside a closure package directory but in a subfolder, (b) go.mod, (c) a `go:embed` file, a `testdata` file, and a file excluded by a build tag in a closure directory, **When** the scope is computed, **Then** each is inside the scope and the decision is overlap.
- Verify (N=3): `go test ./internal/factory -run '^(TestScopeClosureDirectoriesRecursive|TestScopeIncludesModuleFiles|TestScopeIncludesEmbedTestdataAndTagExcludedFiles)$' -count=1 -v`.
- RED-now: E-017. Green: M6.

### AC-LVM-018 — undecidable is overlap; declared limits (REQ-LVM-018)

- **Given** each of the seven undecidable classes (no observed package list; non-`go test -json` command; fingerprint or tag mismatch; `go list` failure; undeclared non-Go reads of a measured package; live-git dependence; reuse-record source), **Then** the decision is overlap with the class named; **Given** a moved non-Go file outside the closure, **Then** overlap, unless a declared never-read glob matches it.
- Verify (N=3): `go test ./internal/factory -run '^(TestReuseUndecidableIsOverlap|TestMovedNonGoFileOutsideClosureIsOverlap|TestNeverReadGlobNarrowsOnlyWhenDeclared)$' -count=1 -v` (`TestReuseUndecidableIsOverlap` carries seven named subtests, each a `--- PASS` line).
- RED-now: E-018. Green: M6.

### AC-LVM-019 — consumers recompute (REQ-LVM-019)

- **Given** a reuse record whose moved-commit list or decision was edited by hand, **When** the merge step, `complete`, or condition (d) reads it, **Then** each refuses with the existing re-measure-and-re-acquire code (exit 2 on the merge step); **Given** an overlapping moved set, **Then** the refusal text is byte-identical to today's `moved past the record's base` refusal.
- Verify: factory (N=2) `go test ./internal/factory -run '^(TestMergeStepRecomputesOverlapAndRefusesMismatch|TestOverlapKeepsTodaysRefusal)$' -count=1 -v`; cli (N=1) `go test ./internal/cli -run '^TestCompleteRefusesTamperedReuseRecord$' -count=1 -v`.
- RED-now: E-019a, E-019b. Green: M6.

### AC-LVM-020 — no chaining, cumulative moved set (REQ-LVM-020)

- **Given** a reuse record offered as a source, **Then** refused; **Given** a source at base B and two landings since, one of which touches scope, **Then** the cumulative set from B reads overlap even if the last hop alone does not.
- Verify (N=3): `go test ./internal/factory -run '^(TestReuseSourceMustBeMeasuredGateComplete|TestReuseNeverChains|TestMovedSetCumulativeFromSourceBase)$' -count=1 -v`.
- RED-now: E-020. Green: M6.

## Safety and documentation

### AC-LVM-021 — safety invariants (REQ-LVM-021)

- **Given** a reused landing, **Then** the audit evidence check still runs and a mismatching audited SHA refuses it; **Given** two ready cards, **Then** one window hold lands exactly one; **Given** the new verbs, **Then** none writes the integration branch ref (asserted by a fake ref-writer that fails on any write outside the merge step); **Given** each undecidable reading, **Then** the outcome is refusal.
- Verify (N=4): `go test ./internal/factory -run '^(TestReuseNeverSkipsAudit|TestNoTwoCardsPerWindowHold|TestNoIntegrationBranchWriteOutsideMergeStep|TestUndecidableRefuses)$' -count=1 -v`.
- RED-now: E-021. Green: M6 (final).

### AC-LVM-022 — rules and mirrors (REQ-LVM-022)

- **Given** the run-phase tree after M7, **Then** `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` and its template mirror each contain the literal phrases `re-measure completeness gate`, `ready set` and `reuse record`; `.claude/rules/moai/workflow/resource-slot-lease.md` and its mirror each contain `capacity` and `auto-release`; and each rule file is byte-identical to its mirror.
- Verify, one invocation each: `grep -c -F -e "re-measure completeness gate" <file>` (and the other phrases) for the four files, each exit 0 and count ≥ 1; `cmp .claude/rules/moai/workflow/factory-dispatch-mechanics.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md` and the same for `resource-slot-lease.md` (each exit 0).
- RED-now (pin as above): `grep -n -F -e "re-measure completeness gate"` on the rule and its mirror → exit 1, empty stdout (both files); on `resource-slot-lease.md` and its mirror → exit 1; `grep -n -F -e "capacity" .claude/rules/moai/workflow/resource-slot-lease.md` → exit 1, empty. Regression guard (green now, must stay green): the two `cmp` calls print nothing and exit 0 at the pin. Green: M7.

## Mutant probes (verification-completeness §2)

| Criterion | Mutant that must fail | Caught by |
|---|---|---|
| AC-LVM-004 | verifier reads only the newest record | `TestLastRecordAloneInsufficient` |
| AC-LVM-005 | verifier ignores the tree key | `TestGateRejectsOtherTreeOrHead` (and passes the positive control — which is why both directions exist) |
| AC-LVM-005 | verifier counts aborted as pass | `TestGateRejectsAbortedAttempt`, `TestGateRejectsPartialJSON` |
| AC-LVM-012 | verifier treats an unreadable load as zero | `TestSlotExtraHolderRefusedOnUnreadableSignal` |
| AC-LVM-015 | readiness cached to disk and trusted | `TestReadyViewPersistsNothing`, `TestCompleteIgnoresReadyView` |
| AC-LVM-018 | undecidable returns no-overlap | `TestReuseUndecidableIsOverlap` |
| AC-LVM-019 | consumer trusts the record's decision | `TestMergeStepRecomputesOverlapAndRefusesMismatch` |

## Quality gates and Definition of Done

- Every AC above green with the empty-sweep guard satisfied; the 28 RED-now observations re-run at M0 and committed as `evidence/m0-baseline.md`.
- `go vet` and `gofmt -l` clean on touched packages; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; package-scoped `golangci-lint` on touched packages; scoped `-run` coverage for added code with the remainder recorded as a CI-deferred Gap (the box's measured full-package budget makes full `internal/cli` coverage unrealistic locally — `research.md` references the scoped-coverage practice).
- `moai spec lint --strict` on this SPEC exit 0; no `\u` escapes in the catalogue; `make build` regenerated template assets (Template-First) with `git status --short` clean afterwards.
- No retried `complete`, no skipped test, no widened exception: the narrowing allowance list diff shows `internal/cli` only.

## Edge cases

- A card that changes no Go package: manifest holds zero Go calls and records `no_go_changes: true`; the gate then requires the card's declared non-Go verification call (if none declared, readiness is refused — fail-closed).
- A package with no test files: allowed to sweep zero tests only where the manifest recorded `no_test_files` for it at derivation.
- The manifest is regenerated after a re-absorb that adds importers: the new manifest replaces the old; attempts bound to the old manifest identity stop counting.

---

## Evidence ledger (RED-now observations)

All entries: tree `ec164d45d2e3c5a7902aa950ab212e8f385f5d08`, run 2026-10-18, unpiped (output redirected to a file and printed verbatim). `names_listed` counts the lines of the form `Test…` in the stdout.

#### E-001
```
command: go test ./internal/factory -list '^(TestManifestDerivesChangedAndImporters|TestManifestNarrowingOnlyInternalCLI|TestManifestIncludesTestOnlyChange)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.372s
names_listed: 0 (required: 3)
```
#### E-002
```
command: go test ./internal/factory -list '^(TestAttemptRecordCarriesAllFields|TestAttemptPackageTerminalEvents)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.333s
names_listed: 0 (required: 2)
```
#### E-003
```
command: go test ./internal/factory -list '^(TestAbortedAttemptNeverCounts|TestRestartMakesNewAttemptID|TestAbortClassesSignalTimeoutPartial)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.346s
names_listed: 0 (required: 3)
```
#### E-004
```
command: go test ./internal/factory -list '^(TestManifestWholePassPositiveControl|TestLastRecordAloneInsufficient|TestMostRecentCompletedAttemptCounts)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.318s
names_listed: 0 (required: 3)
```
#### E-005
```
command: go test ./internal/factory -list '^(TestGateRejectsMissingManifest|TestGateRejectsMissingCall|TestGateRejectsPartialJSON|TestGateRejectsAbortedAttempt|TestGateRejectsOtherTreeOrHead|TestGateRejectsEnvFingerprintMismatch|TestGateRejectsUnverifiedNonGoInput|TestGateRejectsUndecidableOverlap|TestGateRejectsEmptySweep)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.352s
names_listed: 0 (required: 9)
```
#### E-006a
```
command: go test ./internal/factory -list '^(TestMergeStepConsultsCompletenessVerifier|TestScriptWrappedRecordRefusedWhenGateOn)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.385s
names_listed: 0 (required: 2)
```
#### E-006b
```
command: go test ./internal/cli -list '^(TestFactoryCompleteConsultsCompletenessVerifier)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/cli	1.629s
names_listed: 0 (required: 1)
```
#### E-006c
```
command: go test ./internal/factorylane -list '^(TestMergeTripleConditionDUsesVerifier)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factorylane	0.342s
names_listed: 0 (required: 1)
```
#### E-007a
```
command: go test ./internal/config -list '^(TestLandingGateDefaultsOff|TestRunNarrowingAllowanceInternalCLIOnly)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/config	0.412s
names_listed: 0 (required: 2)
```
#### E-007b
```
command: go test ./internal/factory -list '^(TestGateOffMatchesTodaysBehaviour)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.429s
names_listed: 0 (required: 1)
```
#### E-008
```
command: go test ./internal/factory -list '^(TestReuseInertWithoutGate)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.264s
names_listed: 0 (required: 1)
```
#### E-009a
```
command: go test ./internal/factory -list '^(TestLandingEventRowsAppended|TestMetricsReportFields|TestEventLogFailureNeverBlocksLanding)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.310s
names_listed: 0 (required: 3)
```
#### E-009b
```
command: go test ./internal/cli -list '^(TestIntegrationMetricsVerb)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/cli	1.554s
names_listed: 0 (required: 1)
```
#### E-010
```
command: go test ./internal/factory -list '^(TestSlotDeadOwnerAutoReleased|TestSlotUnresolvedOwnerReadsLive|TestSlotAutoReleaseAuditRow)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.312s
names_listed: 0 (required: 3)
```
#### E-011
```
command: go test ./internal/factory -list '^(TestSlotCapacityOneIsToday|TestSlotCapacityTwoAdmitsTwo|TestSlotCapacityTwoRefusesThird)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.313s
names_listed: 0 (required: 3)
```
#### E-012
```
command: go test ./internal/factory -list '^(TestSlotExtraHolderNeedsLoadAndMemoryHeadroom|TestSlotExtraHolderRefusedOnUnreadableSignal|TestSlotThresholdsUnsetMeansCapacityOne)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.269s
names_listed: 0 (required: 3)
```
#### E-013
```
command: go test ./internal/factory -list '^(TestSlotRefusesIntersectingWallClockPackages|TestSlotAllowsDisjointPackages)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.323s
names_listed: 0 (required: 2)
```
#### E-014a
```
command: go test ./internal/factory -list '^(TestReadySetListsOnlyFullyPrepared|TestNotReadyCarriesFirstFailingPredicate)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.288s
names_listed: 0 (required: 2)
```
#### E-014b
```
command: go test ./internal/cli -list '^(TestIntegrationReadyVerb)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/cli	1.609s
names_listed: 0 (required: 1)
```
#### E-015a
```
command: go test ./internal/factory -list '^(TestReadyViewPersistsNothing|TestAuditSHAMismatchNotReadyEvenIfPatchIdentical)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.393s
names_listed: 0 (required: 2)
```
#### E-015b
```
command: go test ./internal/cli -list '^(TestCompleteIgnoresReadyView)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/cli	1.501s
names_listed: 0 (required: 1)
```
#### E-016
```
command: go test ./internal/factory -list '^(TestReuseRecordWrittenWhenNoOverlap|TestReuseRecordCitesSourceAndMovedCommits)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.329s
names_listed: 0 (required: 2)
```
#### E-017
```
command: go test ./internal/factory -list '^(TestScopeClosureDirectoriesRecursive|TestScopeIncludesModuleFiles|TestScopeIncludesEmbedTestdataAndTagExcludedFiles)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.318s
names_listed: 0 (required: 3)
```
#### E-018
```
command: go test ./internal/factory -list '^(TestReuseUndecidableIsOverlap|TestMovedNonGoFileOutsideClosureIsOverlap|TestNeverReadGlobNarrowsOnlyWhenDeclared)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.306s
names_listed: 0 (required: 3)
```
#### E-019a
```
command: go test ./internal/factory -list '^(TestMergeStepRecomputesOverlapAndRefusesMismatch|TestOverlapKeepsTodaysRefusal)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.319s
names_listed: 0 (required: 2)
```
#### E-019b
```
command: go test ./internal/cli -list '^(TestCompleteRefusesTamperedReuseRecord)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/cli	1.560s
names_listed: 0 (required: 1)
```
#### E-020
```
command: go test ./internal/factory -list '^(TestReuseSourceMustBeMeasuredGateComplete|TestReuseNeverChains|TestMovedSetCumulativeFromSourceBase)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.334s
names_listed: 0 (required: 3)
```
#### E-021
```
command: go test ./internal/factory -list '^(TestReuseNeverSkipsAudit|TestNoTwoCardsPerWindowHold|TestNoIntegrationBranchWriteOutsideMergeStep|TestUndecidableRefuses)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.300s
names_listed: 0 (required: 4)
```
