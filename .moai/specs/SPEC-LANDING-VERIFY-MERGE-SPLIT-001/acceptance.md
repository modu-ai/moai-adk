# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — acceptance criteria (v0.3.0, piece A)

Verification layer: each entry is an `AC-LVM-XXX` Given-When-Then, binary-testable. The GEARS requirements are `spec.md` §3. Counts: 15 requirements, 15 acceptance criteria (Tier L ceiling 25/25), one-to-one. The evidence ledger at the end holds 20 observations because several criteria span two or three packages.

## Common setup

- **Test names are the interface.** The names are proposals; the run phase may rename one only with a matching edit here (a rename leaves a `--- PASS` line missing, which reads red).
- **Empty-sweep guard (verification-completeness §1.1).** Every Go AC runs `go test ./internal/<pkg> -run '^(A|B|…)$' -count=1 -v` and requires together: (i) exactly one `--- PASS: <Name> ` line per listed name (N stated per AC), (ii) no `[no tests to run]`, (iii) exit 0. A selector matching zero tests exits 0, so the name count is the verdict.
- **RED-now cells.** Four elements per cell, carried in the evidence ledger at the end: command, verbatim stdout, exit code, tree SHA. Document-level pin: `8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2` (the tree of the plan commit being revised; no `_test.go` file differs from main at `ec164d45d2`), which binds every cell without its own pin. A Go cell is `grep -rhoE "^func (A|B)\(" internal/<pkg> --include='*_test.go'`, observed with exit 1 and empty stdout; red reason: the named test functions do not exist because the behaviour is not implemented (found 0 ≠ N). This form needs no Go build (resource rule 8c1e); the green command is the `go test` line of each AC.
- **Judging build.** CLI observations used the prebuilt `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/c5bd52fe-ffbe-40e1-94ce-4491a79c0784/scratchpad/moai-t1872`, built from main `ec164d45d2` (its `version` output prints no commit, so the commit is the lead's statement, not an observation).
- **Green path.** Milestones (`plan.md`): M0 baseline; M1 gate core; M2 receipt commit field; M3 audit binding; M4 lane guidance. The enabling switch is not a milestone of this branch (REQ-LVM-012).
- **Fixtures.** Gate tests build temporary git repositories with `t.TempDir()` and canned `go test -json` streams; none reads this repository's refs, `.moai/state`, or the user's home. Receipt fixtures use the real writer function, not a hand-built JSON.

### AC-LVM-001 — manifest derivation (REQ-LVM-001)

- **Given** a temporary repository whose card branch changes (a) a non-test file of `p/a`, (b) only a test file of `p/b`, (c) a file of `internal/cli`, (d) a non-Go file outside any package directory, (e) a file embedded by package `p/e` that lives in `p/a`'s directory, (f) deletes package `p/x`; `p/c` imports `p/a`, `p/d` imports `p/b` only in a test, **When** the producer runs, **Then** the manifest lists a whole-package call (no `-run`) for `p/a`, `p/b`, `internal/cli`, `p/c`, `p/d`, `p/e`; only the `internal/cli` call may carry `-run`, naming `35e9`; every call records identifier, exact command and package set; (d) has reader calls or a zero-readers record naming the search command and its empty result; (f) is a recorded base-versus-candidate graph comparison, not a failing `go list`.
- Verify (N=6): `go test ./internal/factory -run '^(TestManifestDerivesChangedAndImporters|TestManifestNarrowingOnlyInternalCLI|TestManifestIncludesTestOnlyChange|TestManifestRecordsNonGoReaderSearch|TestManifestDerivesEmbedOwnerPackages|TestManifestHandlesDeletedPackage)$' -count=1 -v`.
- RED-now: E-001. Green: M1.

### AC-LVM-002 — attempt record (REQ-LVM-002)

- **Given** a manifest call run in the structured form over a fixture stream of two packages (one with 4 test events, one with 0), **Then** the record carries unique attempt ID, call identifier, exact argv, exit code, per-package terminal event and executed-test count (4 and 0), tree SHA, HEAD SHA, base, environment fingerprint, start/end times, judging build identity and termination `completed`.
- Verify (N=3): `go test ./internal/factory -run '^(TestAttemptRecordCarriesAllFields|TestAttemptPackageTerminalEvents|TestAttemptRecordsPerPackageTestCounts)$' -count=1 -v`.
- RED-now: E-002. Green: M1.

### AC-LVM-003 — abort versus restart (REQ-LVM-003)

- **Given** attempts ended by (a) SIGTERM, (b) the verb's timeout, (c) a stream cut mid-line, **Then** each is `aborted`, retained, and counts as no pass; **When** the call is run again, **Then** it is a new attempt with a new ID, the aborted record is intact, and only the most recent completed attempt on the same tree counts.
- Verify (N=3): `go test ./internal/factory -run '^(TestAbortedAttemptNeverCounts|TestRestartMakesNewAttemptID|TestAbortClassesSignalTimeoutPartial)$' -count=1 -v`.
- RED-now: E-003. Green: M1.

### AC-LVM-004 — whole manifest, positive control (REQ-LVM-004)

- **Given** a manifest of three calls, **When** all three have a completed exit-0 attempt on the candidate tree with full package coverage and non-zero counts, **Then** the verifier reports gate-complete (positive control); **When** only the last-recorded call is green and the others have no attempt, **Then** it does not; **When** a call has an aborted attempt followed by a completed one on the same tree, **Then** the completed one counts.
- Verify (N=3): `go test ./internal/factory -run '^(TestManifestWholePassPositiveControl|TestLastRecordAloneInsufficient|TestMostRecentCompletedAttemptCounts)$' -count=1 -v`.
- RED-now: E-004. Green: M1.

### AC-LVM-005 — the twelve reject classes (REQ-LVM-005)

- **Given** the AC-LVM-004 positive fixture, **When** exactly one mutation is applied — (1) manifest removed, (2) a call's attempts removed, (3) a call's only record is a script-wrapped command, (4) a stream lacks an expected package's terminal event, (5) a stream cut mid-line, (6) an attempt aborted, (7) an attempt on another tree or HEAD, (8) fingerprint differs, (9) a changed non-Go input with neither reader call nor zero-readers record, (10) the scope cannot be derived, (11) a package passes with zero executed tests (selector matched nothing, or a `TestMain` that exits 0) where the manifest does not allow it, (12) the attempt's command differs from the call's (`-run '^TestX$'`, `-short`, `-skip`, a different package list, a changed `-count` or `-p`, a pipe or an extra segment), — **Then** the verifier refuses in each case naming the class; the unmutated fixture still passes (control); and a `-run` selector is accepted only on an `internal/cli` call whose manifest entry names 35e9.
- Verify (N=12): `go test ./internal/factory -run '^(TestGateRejectsMissingManifest|TestGateRejectsMissingCall|TestGateRejectsScriptWrappedRecord|TestGateRejectsMissingPackageEvent|TestGateRejectsPartialJSON|TestGateRejectsAbortedAttempt|TestGateRejectsOtherTreeOrHead|TestGateRejectsEnvFingerprintMismatch|TestGateRejectsUnverifiedNonGoInput|TestGateRejectsUnderivableScope|TestGateRejectsEmptySweep|TestGateRejectsCommandDiffersFromManifest)$' -count=1 -v`.
- RED-now: E-005. Green: M1.

### AC-LVM-006 — one verifier, every consumer (REQ-LVM-006)

- **Given** the gate enabled and a tree whose only record is a script-wrapped call (exit 0), **When** the merge step, `moai factory complete`, merge-readiness condition (d) and `moai integration manifest verify` consult it, **Then** all four refuse with the same reason text; a gate-complete tree passes all four.
- Verify: factory (N=1) `go test ./internal/factory -run '^TestMergeStepConsultsCompletenessVerifier$' -count=1 -v`; cli (N=2) `go test ./internal/cli -run '^(TestFactoryCompleteConsultsCompletenessVerifier|TestManifestVerifyVerbSharesVerifier)$' -count=1 -v`; factorylane (N=1) `go test ./internal/factorylane -run '^TestMergeTripleConditionDUsesVerifier$' -count=1 -v`.
- RED-now: E-006a, E-006b, E-006c. Green: M1.

### AC-LVM-007 — gate state and source (REQ-LVM-007)

- **Given** a fresh config and no `workflow.landing_gate` block, **Then** the landing path's outcome for a script-wrapped record equals the pre-change outcome (oracle: the characterization test recorded at M1 against the unchanged `ValidateRemeasureRecord` verdicts of the same fixtures); the shipped template default is false; the narrowing list is exactly `[internal/cli]`; **Given** a card branch whose own `workflow.yaml` sets the key true while the integration root's does not, **Then** the gate stays off.
- Verify: config (N=2) `go test ./internal/config -run '^(TestLandingGateDefaultsOff|TestRunNarrowingAllowanceInternalCLIOnly)$' -count=1 -v`; factory (N=2) `go test ./internal/factory -run '^(TestGateOffMatchesTodaysBehaviour|TestGateStateReadFromIntegrationRootNotCardTree)$' -count=1 -v`.
- RED-now: E-007a, E-007b. Green: M1.

### AC-LVM-008 — verdict bound to the current HEAD (REQ-LVM-008)

- **Given** the gate enabled and a SPEC card, **Then** `complete` refuses when the verdict's audited commit is an earlier commit (a) with a non-empty change since, (b) with an empty or patch-identical change since; it refuses when HEAD differs from the audited commit by (c) a `progress.md` edit under the SPEC directory, or (d) a path under `.moai/reports/` but another card's directory; it passes when HEAD equals the audited commit or differs only under `.moai/reports/<card-id>/`.
- Verify: factorylane (N=5) `go test ./internal/factorylane -run '^(TestSyncAuditVerdictMustNameCurrentHead|TestEarlierPassNeverSubstitutes|TestOnlyEvidenceDirDiffTolerated|TestProgressBackfillAfterAuditRefuses|TestOtherReportsPathNotTolerated)$' -count=1 -v`; cli (N=1) `go test ./internal/cli -run '^TestCompleteRefusesStaleAuditHead$' -count=1 -v`.
- RED-now: E-008a, E-008b. Green: M3.

### AC-LVM-009 — receipt commit field and writer (REQ-LVM-009)

- **Given** the real writer called for a tree at HEAD `H`, **Then** the written receipt carries audited commit `H` (a full 40-hex SHA); **Given** the tree's HEAD moved while the call ran, **Then** the field is empty; **Given** a legacy receipt JSON without the field, **Then** it decodes as unbound and the file is byte-identical after a read-and-verify pass (never rewritten).
- Verify: auditreceipt (N=3) `go test ./internal/auditreceipt -run '^(TestReceiptRoundTripsAuditedCommit|TestLegacyReceiptDecodesAsUnbound|TestExistingReceiptNeverRewritten)$' -count=1 -v`; cli (N=2) `go test ./internal/cli -run '^(TestWriterStampsAuditedCommit|TestWriterLeavesEmptyWhenHeadMoved)$' -count=1 -v`.
- RED-now: E-009a, E-009b. Green: M2.

### AC-LVM-010 — receipt binding in `complete` (REQ-LVM-010)

- **Given** the gate enabled and a verdict citing a receipt, **Then** `complete` passes when the receipt is readable, a codex or multi-backend audit receipt, unmet-gate empty, verdict pass, audited commit equal to HEAD; and refuses each of (b) a receipt for an older HEAD, (c) an absent receipt, (d) an unreadable (corrupt) receipt, (e) a legacy receipt without the field, (f) a receipt with the field empty, (g) a receipt with a non-empty unmet gate; the passing fixture's receipt is produced by the real writer.
- Verify: factorylane (N=7) `go test ./internal/factorylane -run '^(TestReceiptAtCurrentHeadPasses|TestReceiptFromOlderHeadRefused|TestAbsentReceiptRefused|TestUnreadableReceiptRefused|TestLegacyReceiptWithoutCommitRefused|TestReceiptWithEmptyCommitRefused|TestUnmetGateReceiptRefused)$' -count=1 -v`; cli (N=1) `go test ./internal/cli -run '^TestRealWriterReceiptSatisfiesBinding$' -count=1 -v`.
- RED-now: E-010a, E-010b. Green: M3.

### AC-LVM-011 — lane guidance, one check per procedure step (REQ-LVM-011)

- **Given** the tree after M4, **Then** `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` and its template mirror each contain every literal phrase below, and the two files are byte-identical.
- Verify, one invocation per phrase and file (14 calls, each exit 0): `grep -c -F -e "<phrase>" <file>` for the phrases `re-measure completeness gate`, `manifest derive`, `remeasure --call`, `manifest verify`, `restarted as a new attempt`, `built from the landing HEAD`, `before the final audit`; `cmp .claude/rules/moai/workflow/factory-dispatch-mechanics.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md` (exit 0).
- RED-now (pin as above): the 14 greps each exit 1 (observed with `grep -q`; recorded in the ledger note). Regression guard (green now, must stay green): the `cmp` exits 0 at the pin. Green: M4.

### AC-LVM-012 — the branch never enables; the switch is separate; recovery is ordinary (REQ-LVM-012)

- **Given** the run-phase branch, **Then** the milestone commits are in graph order M0 < M1 < M2 < M3 < M4; M0 changes only the baseline file; no commit of the branch touches the repository's own `workflow.yaml`; and the reverse ordering fails (negative control); **Given** a later one-path card that sets the key false, **Then** it lands through the zero-Go path with the audit binding applied, and is refused when the audit binding fails.
- Verify, one invocation each: `git log --format=%H -1 --grep='^docs(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M0 '` and `^feat(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M<k> ` for k = 1..4 (each prints one 40-hex SHA, exit 0) → S0..S4; `git merge-base --is-ancestor <S0> <S1>`, `<S1> <S2>`, `<S2> <S3>`, `<S3> <S4>` (each exit 0); negative control `<S4> <S0>` (exit 1); `git diff --name-only <S0>^ <S0>` (exactly `.moai/specs/SPEC-LANDING-VERIFY-MERGE-SPLIT-001/evidence/m0-baseline.md`); `git ls-files --error-unmatch` on that path (exit 0, tracked, outside `.moai/reports`); `git log --format=%H <S0>^..<S4> -- .moai/config/sections/workflow.yaml` (exit 0, **empty** stdout: the branch never enables, which also rejects an enable-then-revert history); cli (N=2) `go test ./internal/cli -run '^(TestKeyOffCardLandsThroughZeroGoPath|TestKeyOffCardStillNeedsAuditBinding)$' -count=1 -v`.
- Post-landing witness (regression-guard class, not release-blocking: the switch commit does not exist at this SPEC's landing, `verification-completeness` §2.1 undecidable disposition), run by the leader when the switch commit `<SW>` exists: `git diff --name-only <SW>^ <SW>` prints exactly `.moai/config/sections/workflow.yaml`; `git merge-base --is-ancestor <S4> <SW>` exit 0; `git grep -n -F -e "manifest verify" <SW>^ -- internal/cli` exit 0 (the gate code is in the switch's parent).
- RED-now (pin as above): the M0 and M1 `git log` commands exit 0 with empty stdout; `git ls-files --error-unmatch …/evidence/m0-baseline.md` exit 1; `git grep -n -F -e landing_gate HEAD -- .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml` exit 1; E-012. The polarities of `merge-base --is-ancestor` were observed at the pin (self → 0; non-ancestor → 1). Green: M0..M4 landing in order.

### AC-LVM-013 — zero-Go card (REQ-LVM-013)

- **Given** a card that changes only a rule file no test reads, **Then** the manifest is empty with an evidence block (changed-path list, `go list` proof, reader-search result) and is admitted by the merge step, `complete` and condition (d) alike (positive); it is refused when the evidence block is missing, and when the recorded changed-path list differs from the repository's recomputed list (a later commit added a path); the audit binding applies unchanged.
- Verify (N=4): `go test ./internal/factory -run '^(TestZeroGoCardAdmittedWithEvidence|TestZeroGoCardRefusedWithoutEvidence|TestZeroGoCardRefusedOnStaleChangedPaths|TestZeroGoCardAdmittedThroughAllThreeConsumers)$' -count=1 -v`.
- RED-now: E-013. Green: M1.

### AC-LVM-014 — manifest producer, end to end (REQ-LVM-014)

- **Given** a fixture repository with a two-package change, **When** a lane runs `manifest derive --card <id>`, then each call through `remeasure --call <call-id> -- <command>`, then `manifest verify --card <id>`, **Then** the manifest file exists at `.moai/state/remeasure/manifests/<card-id>.json` under the integration root, `verify` exits 0; after a re-absorb that adds an importer and a fresh `derive`, `verify` exits non-zero until the new call has an attempt; `show` prints the manifest; a refusal names the class.
- Verify (N=4): `go test ./internal/cli -run '^(TestManifestDeriveWritesStoredManifest|TestManifestRegeneratedAfterReabsorb|TestManifestShowAndVerifyVerbs|TestDeriveRunVerifyEndToEnd)$' -count=1 -v`.
- RED-now: E-014. Green: M1.

### AC-LVM-015 — safety invariants (REQ-LVM-015)

- **Given** a gate-enabled landing, **Then** the existing audit check still runs and a mismatching audit refuses it; the new verbs write no integration-branch ref (a fake ref-writer fails on any write outside the merge step); each undecidable reading refuses; and a call whose command narrows below the whole package is refused unless the manifest entry itself names the `internal/cli` 35e9 narrowing (the class-12 mutant of AC-LVM-005 is the named witness).
- Verify (N=4): `go test ./internal/factory -run '^(TestGateNeverSkipsAuditCheck|TestNoIntegrationBranchWriteOutsideMergeStep|TestUndecidableRefuses|TestNarrowingAllowedOnlyWhenManifestNamesIt)$' -count=1 -v`.
- RED-now: E-015. Green: M1 (final check at M4).

## Mutant probes (verification-completeness §2)

| Criterion | Mutant that must fail | Caught by |
|---|---|---|
| AC-LVM-004 | verifier reads only the newest record | `TestLastRecordAloneInsufficient` |
| AC-LVM-005 | verifier ignores the tree key | `TestGateRejectsOtherTreeOrHead` (the positive control passes it, which is why both directions exist) |
| AC-LVM-005 | verifier counts an aborted attempt as pass | `TestGateRejectsAbortedAttempt`, `TestGateRejectsPartialJSON` |
| AC-LVM-005 | verifier accepts a script-wrapped exit-0 record | `TestGateRejectsScriptWrappedRecord` |
| AC-LVM-005 | verifier accepts `go test -json -run '^TestX$' ./pkg` | `TestGateRejectsCommandDiffersFromManifest` |
| AC-LVM-005 | verifier accepts a package that passed with zero tests | `TestGateRejectsEmptySweep` |
| AC-LVM-008 | binding accepts an ancestor whose diff is empty | `TestEarlierPassNeverSubstitutes` |
| AC-LVM-010 | binding matches a receipt by card id or time instead of the audited commit | `TestReceiptFromOlderHeadRefused`, `TestLegacyReceiptWithoutCommitRefused` |
| AC-LVM-012 | a branch commit enables the key and a later one removes it | the empty `git log <S0>^..<S4> -- workflow.yaml` |

## Quality gates and Definition of Done

- Every AC green with the empty-sweep guard satisfied; the RED-now cells re-run at M0 and committed as `evidence/m0-baseline.md`.
- `go vet` and `gofmt -l` clean on touched packages; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; package-scoped `golangci-lint`; scoped `-run` coverage for the added code with the remainder recorded as a CI-deferred Gap (the box's measured budget). Go calls are made only while the host admits them (resource rule 8c1e).
- `moai spec lint --strict` on this SPEC exit 0; no `\u` escapes; `make build` regenerated template assets (Template-First) with `git status --short` clean after.
- No skipped test, no excluded test in any manifest call, no widened allowance: the narrowing list diff shows `internal/cli` only.
- **Not claimed:** that this card's own landing runs under the gate. Piece A lands under the current rules (board `0577` P2); the first gated landing is the next one after a later, separate enabling commit made with a by-path binary built from main (spec C5).

## Edge cases

- A package with no test files: may sweep zero tests only where the manifest recorded `no_test_files` for it at derivation, naming the reason.
- A re-absorb that adds importers regenerates the manifest; attempts bound to the superseded manifest identity stop counting.
- A card whose only change is outside Go and has readers: the readers become calls; the zero-call path applies only when there are none.
- Recovery if a defective gate refuses its own revert: the one-path key-off card lands through the zero-Go path; if the zero-Go path itself is the defect, the operator-held manual edit of the integration root's configuration is the accepted residual, recorded with the operator as owner and not performed by an agent.

---

## Evidence ledger (RED-now observations)

All entries: tree `8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2`, run 2026-10-19, unpiped (output redirected to a file and printed verbatim). `names_found` counts stdout lines of the form `func Test…(`.

Non-Go cells (AC-LVM-011, 012), exit codes observed at the pin: the 14 `grep -q -F -e "<phrase>" <rule or mirror>` calls of AC-LVM-011 → 1 each; `git grep -n -F -e landing_gate HEAD -- .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml` → 1; `git log --format=%H -1 --grep='^feat(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M1 '` → 0 with empty stdout; the M0 subject → 0, empty; `git ls-files --error-unmatch .moai/specs/SPEC-LANDING-VERIFY-MERGE-SPLIT-001/evidence/m0-baseline.md` → 1.

#### E-001
```
command: grep -rhoE "^func (TestManifestDerivesChangedAndImporters|TestManifestNarrowingOnlyInternalCLI|TestManifestIncludesTestOnlyChange|TestManifestRecordsNonGoReaderSearch|TestManifestDerivesEmbedOwnerPackages|TestManifestHandlesDeletedPackage)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 6)
```
#### E-002
```
command: grep -rhoE "^func (TestAttemptRecordCarriesAllFields|TestAttemptPackageTerminalEvents|TestAttemptRecordsPerPackageTestCounts)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 3)
```
#### E-003
```
command: grep -rhoE "^func (TestAbortedAttemptNeverCounts|TestRestartMakesNewAttemptID|TestAbortClassesSignalTimeoutPartial)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 3)
```
#### E-004
```
command: grep -rhoE "^func (TestManifestWholePassPositiveControl|TestLastRecordAloneInsufficient|TestMostRecentCompletedAttemptCounts)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 3)
```
#### E-005
```
command: grep -rhoE "^func (TestGateRejectsMissingManifest|TestGateRejectsMissingCall|TestGateRejectsScriptWrappedRecord|TestGateRejectsMissingPackageEvent|TestGateRejectsPartialJSON|TestGateRejectsAbortedAttempt|TestGateRejectsOtherTreeOrHead|TestGateRejectsEnvFingerprintMismatch|TestGateRejectsUnverifiedNonGoInput|TestGateRejectsUnderivableScope|TestGateRejectsEmptySweep|TestGateRejectsCommandDiffersFromManifest)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 12)
```
#### E-006a
```
command: grep -rhoE "^func (TestMergeStepConsultsCompletenessVerifier)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 1)
```
#### E-006b
```
command: grep -rhoE "^func (TestFactoryCompleteConsultsCompletenessVerifier|TestManifestVerifyVerbSharesVerifier)\(" internal/cli --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 2)
```
#### E-006c
```
command: grep -rhoE "^func (TestMergeTripleConditionDUsesVerifier)\(" internal/factorylane --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 1)
```
#### E-007a
```
command: grep -rhoE "^func (TestLandingGateDefaultsOff|TestRunNarrowingAllowanceInternalCLIOnly)\(" internal/config --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 2)
```
#### E-007b
```
command: grep -rhoE "^func (TestGateOffMatchesTodaysBehaviour|TestGateStateReadFromIntegrationRootNotCardTree)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 2)
```
#### E-008a
```
command: grep -rhoE "^func (TestSyncAuditVerdictMustNameCurrentHead|TestEarlierPassNeverSubstitutes|TestOnlyEvidenceDirDiffTolerated|TestProgressBackfillAfterAuditRefuses|TestOtherReportsPathNotTolerated)\(" internal/factorylane --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 5)
```
#### E-008b
```
command: grep -rhoE "^func (TestCompleteRefusesStaleAuditHead)\(" internal/cli --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 1)
```
#### E-009a
```
command: grep -rhoE "^func (TestReceiptRoundTripsAuditedCommit|TestLegacyReceiptDecodesAsUnbound|TestExistingReceiptNeverRewritten)\(" internal/auditreceipt --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 3)
```
#### E-009b
```
command: grep -rhoE "^func (TestWriterStampsAuditedCommit|TestWriterLeavesEmptyWhenHeadMoved)\(" internal/cli --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 2)
```
#### E-010a
```
command: grep -rhoE "^func (TestReceiptAtCurrentHeadPasses|TestReceiptFromOlderHeadRefused|TestAbsentReceiptRefused|TestUnreadableReceiptRefused|TestLegacyReceiptWithoutCommitRefused|TestReceiptWithEmptyCommitRefused|TestUnmetGateReceiptRefused)\(" internal/factorylane --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 7)
```
#### E-010b
```
command: grep -rhoE "^func (TestRealWriterReceiptSatisfiesBinding)\(" internal/cli --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 1)
```
#### E-012
```
command: grep -rhoE "^func (TestKeyOffCardLandsThroughZeroGoPath|TestKeyOffCardStillNeedsAuditBinding)\(" internal/cli --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 2)
```
#### E-013
```
command: grep -rhoE "^func (TestZeroGoCardAdmittedWithEvidence|TestZeroGoCardRefusedWithoutEvidence|TestZeroGoCardRefusedOnStaleChangedPaths|TestZeroGoCardAdmittedThroughAllThreeConsumers)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 4)
```
#### E-014
```
command: grep -rhoE "^func (TestManifestDeriveWritesStoredManifest|TestManifestRegeneratedAfterReabsorb|TestManifestShowAndVerifyVerbs|TestDeriveRunVerifyEndToEnd)\(" internal/cli --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 4)
```
#### E-015
```
command: grep -rhoE "^func (TestGateNeverSkipsAuditCheck|TestNoIntegrationBranchWriteOutsideMergeStep|TestUndecidableRefuses|TestNarrowingAllowedOnlyWhenManifestNamesIt)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 (no test file changed since ec164d45d2)
exit: 1
stdout:
names_found: 0 (required: 4)
```
