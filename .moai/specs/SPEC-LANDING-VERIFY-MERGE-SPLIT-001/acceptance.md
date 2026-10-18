# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — acceptance criteria (v0.5.0, piece A)

Verification layer: each entry is an `AC-LVM-XXX` Given-When-Then, binary-testable. The GEARS requirements are `spec.md` §3. Counts: 18 requirements, 18 acceptance criteria (Tier L ceiling 25/25), one-to-one. The evidence ledger at the end holds 24 observations because several criteria span two or three packages.

## Common setup

- **Test names are the interface.** The names are proposals; the run phase may rename one only with a matching edit here (a rename leaves a `--- PASS` line missing, which reads red).
- **Empty-sweep guard (verification-completeness §1.1).** Every Go AC runs `go test ./internal/<pkg> -run '^(A|B|…)$' -count=1 -v` and requires together: (i) exactly one `--- PASS: <Name> ` line per listed name (N stated per AC), (ii) no `[no tests to run]`, (iii) exit 0. A selector matching zero tests exits 0, so the name count is the verdict.
- **RED-now cells.** Four elements per cell, carried in the evidence ledger at the end: command, verbatim stdout, exit code, tree SHA. Document-level pin: `8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2` (no `_test.go` file differs from main at `ec164d45d2`, and none has changed since), which binds every cell without its own pin. A Go cell is `grep -rhoE "^func (A|B)\(" internal/<pkg> --include='*_test.go'`, observed with exit 1 and empty stdout; red reason: the named test functions do not exist because the behaviour is not implemented (found 0 ≠ N). This form needs no Go build (resource rule 8c1e); the green command is the `go test` line of each AC.
- **Judging build.** CLI observations used the prebuilt `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/c5bd52fe-ffbe-40e1-94ce-4491a79c0784/scratchpad/moai-t1872`; its `version` output prints no commit, so "built from `ec164d45d2`" is the lead's statement, not an observation.
- **Green path.** Milestones (`plan.md`): M0 baseline; M1 gate core; M2 receipt commit field; M3 audit binding; M4 lane guidance. The enabling switch is not a milestone of this branch (REQ-LVM-012).
- **Fixtures.** Gate tests build temporary git repositories with `t.TempDir()` and canned `go test -json` streams; none reads this repository's refs, `.moai/state`, or the user's home, and none touches the real primary checkout. Receipt fixtures use the real writer function, not hand-built JSON.

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

### AC-LVM-005 — the twelve attempt-level reject classes (REQ-LVM-005)

- **Given** the AC-LVM-004 positive fixture, **When** exactly one mutation is applied — (1) manifest removed, (2) a call's attempts removed, (3) a call's only record is a script-wrapped command, (4) a stream lacks an expected package's terminal event, (5) a stream cut mid-line, (6) an attempt aborted, (7) an attempt on another tree or HEAD, (8) fingerprint differs, (9) a changed non-Go input with neither reader call nor zero-readers record, (10) the scope cannot be derived, (11) a package passes with zero executed tests (selector matched nothing, or a `TestMain` that exits 0) where the manifest does not allow it, (12) the attempt's command differs from the call's (`-run '^TestX$'`, `-short`, `-skip`, a different package list, a changed `-count` or `-p`, a pipe or an extra segment) — **Then** the verifier refuses in each case naming the class; the unmutated fixture still passes (control); and a `-run` selector is accepted only on an `internal/cli` call whose manifest entry names 35e9.
- Verify (N=12): `go test ./internal/factory -run '^(TestGateRejectsMissingManifest|TestGateRejectsMissingCall|TestGateRejectsScriptWrappedRecord|TestGateRejectsMissingPackageEvent|TestGateRejectsPartialJSON|TestGateRejectsAbortedAttempt|TestGateRejectsOtherTreeOrHead|TestGateRejectsEnvFingerprintMismatch|TestGateRejectsUnverifiedNonGoInput|TestGateRejectsUnderivableScope|TestGateRejectsEmptySweep|TestGateRejectsCommandDiffersFromManifest)$' -count=1 -v`.
- RED-now: E-005. Green: M1.

### AC-LVM-006 — one verifier, every consumer (REQ-LVM-006)

- **Given** the gate on and a tree whose only record is a script-wrapped call (exit 0), **When** the merge step, `moai factory complete`, merge-readiness condition (d) and `moai integration manifest verify` consult it, **Then** all four refuse with the same reason text; a gate-complete tree passes all four.
- Verify: factory (N=1) `go test ./internal/factory -run '^TestMergeStepConsultsCompletenessVerifier$' -count=1 -v`; cli (N=2) `go test ./internal/cli -run '^(TestFactoryCompleteConsultsCompletenessVerifier|TestManifestVerifyVerbSharesVerifier)$' -count=1 -v`; factorylane (N=1) `go test ./internal/factorylane -run '^TestMergeTripleConditionDUsesVerifier$' -count=1 -v`.
- RED-now: E-006a, E-006b, E-006c. Green: M1.

### AC-LVM-007 — default off (REQ-LVM-007)

- **Given** a fresh config and no `workflow.landing_gate` block, **Then** the landing path's outcome for a script-wrapped record equals the pre-change outcome (oracle: the characterization test recorded at M1 against the unchanged `ValidateRemeasureRecord` verdicts of the same fixtures); the shipped template default is false; the narrowing list is exactly `[internal/cli]`.
- Verify: config (N=2) `go test ./internal/config -run '^(TestLandingGateDefaultsOff|TestRunNarrowingAllowanceInternalCLIOnly)$' -count=1 -v`; factory (N=1) `go test ./internal/factory -run '^TestGateOffMatchesTodaysBehaviour$' -count=1 -v`.
- RED-now: E-007a, E-007b. Green: M1.

### AC-LVM-008 — verdict locator and equals-HEAD (REQ-LVM-008)

- **Given** the gate on and a SPEC card whose `.moai/reports/<card-id>/` holds `sync-audit.md`, `sync-audit-iter2.md`, `sync-audit-iter10.md`, **Then** the located verdict is `-iter10` (numeric, not lexical); a SPEC card with only a `verdict.md` or with no file is refused; a SPEC-less card is located at `.moai/reports/<card-id>/verdict.md` and refused when that file is absent; `complete` passes only when the verdict's audited commit equals HEAD, and refuses when it is an earlier commit (a) with a non-empty change since, (b) with an empty or patch-identical change since — including an `--allow-empty` commit or a no-change merge after the audited commit, which is the empty-diff-ancestor negative — (c) followed by a `progress.md` edit under the SPEC directory, or (d) followed by a commit that touches only `.moai/reports/<card-id>/`; there is no tolerated difference.
- Verify: factorylane (N=8) `go test ./internal/factorylane -run '^(TestSyncAuditVerdictLocatorPicksNewestIteration|TestSPECCardWithoutSyncAuditVerdictRefused|TestSPECLessCardVerdictLocatedAtVerdictMd|TestSyncAuditVerdictMustNameCurrentHead|TestEarlierPassNeverSubstitutes|TestEmptyDiffAncestorRefused|TestProgressBackfillAfterAuditRefuses|TestEvidenceDirDiffAfterAuditRefused)$' -count=1 -v`; cli (N=1) `go test ./internal/cli -run '^TestCompleteRefusesStaleAuditHead$' -count=1 -v`.
- RED-now: E-008a, E-008b. Green: M3.

### AC-LVM-009 — receipt commit and target fields, writer and entry points (REQ-LVM-009)

- **Given** the real writer called for a tree at HEAD `H` for review target `T`, **Then** the receipt carries `audited_commit` `H` (full 40-hex) and `audit_target` `T`; **Given** the tree's HEAD moved while the call ran, **Then** `audited_commit` is empty; **Given** a real entry point (`codex_audit` with a mocked backend that moves HEAD during the call), **Then** the receipt carries the HEAD captured at entry or an empty field, never the later HEAD; **Given** a legacy receipt JSON without the fields, **Then** it decodes as unbound and the file is byte-identical after a read-and-verify pass.
- Verify: auditreceipt (N=3) `go test ./internal/auditreceipt -run '^(TestReceiptRoundTripsAuditedCommit|TestLegacyReceiptDecodesAsUnbound|TestExistingReceiptNeverRewritten)$' -count=1 -v`; cli (N=4) `go test ./internal/cli -run '^(TestWriterStampsAuditedCommit|TestWriterLeavesEmptyWhenHeadMoved|TestEntryPointCapturesStartHead|TestWriterRecordsAuditTarget)$' -count=1 -v`.
- RED-now: E-009a, E-009b. Green: M2.

### AC-LVM-010 — receipt binding in `complete` (REQ-LVM-010)

- **Given** the gate on and a verdict citing receipts, **Then** `complete` passes when one cited receipt, looked up at the store root of the audited tree, is readable, a codex or multi-backend audit receipt, has `tree_root` equal to the normalized path of the audited tree itself, an empty unmet gate, verdict pass, and `audited_commit` equal to HEAD; and each of (b) an older-HEAD receipt, (c) an absent receipt, (d) an unreadable (corrupt) receipt, (e) a legacy receipt without the field, (f) a receipt with the field empty, (g) a receipt with a non-empty unmet gate, (h) a receipt whose `tree_root` is another tree's path for the same commit, (i) a receipt whose commit is an ancestor of HEAD with an empty diff, fails to qualify and refuses when it is the only citation; **Given** a config-orphaned card worktree `W` (no configuration of its own) whose primary checkout is `P`, **Then** a receipt filed at `P`'s store with `tree_root` = `W` qualifies for `W` (positive control; the store root is not the compared quantity) and a receipt filed there with `tree_root` = `P` for the same commit is refused for `W`; a verdict citing one qualifying receipt plus one failed or inconclusive extra call passes (the extra is recorded), and one citing only failed calls refuses. The passing fixtures' receipts are produced by the real writer.
- Verify: factorylane (N=11) `go test ./internal/factorylane -run '^(TestReceiptAtCurrentHeadPasses|TestReceiptFromOlderHeadRefused|TestAbsentReceiptRefused|TestUnreadableReceiptRefused|TestLegacyReceiptWithoutCommitRefused|TestReceiptWithEmptyCommitRefused|TestUnmetGateReceiptRefused|TestReceiptFromOtherTreeRootRefused|TestReceiptForConfigOrphanedWorktreeQualifies|TestReceiptAtEmptyDiffAncestorRefused|TestFailedExtraReceiptDoesNotBlockButNeverReplacesPass)$' -count=1 -v`; cli (N=2) `go test ./internal/cli -run '^(TestRealWriterReceiptSatisfiesBinding|TestRealWriterReceiptForOrphanedWorktreeSatisfiesBinding)$' -count=1 -v`.
- RED-now: E-010a, E-010b. Green: M3.

### AC-LVM-011 — lane guidance, one check per procedure step (REQ-LVM-011)

- **Given** the tree after M4, **Then** `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` and its template mirror each contain every literal phrase below, and the two files are byte-identical.
- Verify, one invocation per phrase and file (18 calls, each exit 0): `grep -c -F -e "<phrase>" <file>` for the phrases `re-measure completeness gate`, `manifest derive`, `remeasure --call`, `manifest verify`, `restarted as a new attempt`, `committed content of the integration branch`, `gate_state_unreadable`, `built from the landing HEAD`, `before the final audit`; `cmp .claude/rules/moai/workflow/factory-dispatch-mechanics.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md` (exit 0).
- RED-now (pin as above): the 18 greps each exit 1 (observed with `grep -q`; recorded in the ledger note). Regression guard (green now, must stay green): the `cmp` exits 0 at the pin. Green: M4.

### AC-LVM-012 — the branch never enables; the switch is separate; recovery is ordinary (REQ-LVM-012)

- **Given** the run-phase branch, **Then** the milestone commits are in graph order M0 < M1 < M2 < M3 < M4; M0 changes only the baseline file; no commit of the branch touches the repository's own `workflow.yaml`; the reverse ordering fails (negative control); **Given** a later one-path card that sets the key false, **Then** its derived manifest holds the reader calls of `.moai/config/sections/workflow.yaml`, it lands through the gated path, and it is refused when the audit binding fails.
- Verify, one invocation each: `git log --format=%H -1 --grep='^docs(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M0 '` and `^feat(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M<k> ` for k = 1..4 (each prints one 40-hex SHA, exit 0) → S0..S4; `git merge-base --is-ancestor <S0> <S1>`, `<S1> <S2>`, `<S2> <S3>`, `<S3> <S4>` (each exit 0); negative control `<S4> <S0>` (exit 1); `git diff --name-only <S0>^ <S0>` (exactly `.moai/specs/SPEC-LANDING-VERIFY-MERGE-SPLIT-001/evidence/m0-baseline.md`); `git ls-files --error-unmatch` on that path (exit 0); `git log --format=%H <S0>^..<S4> -- .moai/config/sections/workflow.yaml` (exit 0, **empty** stdout: the branch never enables, which also rejects an enable-then-revert history); cli (N=2) `go test ./internal/cli -run '^(TestKeyOffCardLandsThroughGatedPath|TestKeyOffCardStillNeedsAuditBinding)$' -count=1 -v`.
- Post-landing witness (regression-guard class, not release-blocking: the switch commit does not exist at this SPEC's landing, `verification-completeness` §2.1 undecidable disposition), run by the leader when the switch commit `<SW>` exists: `git diff --name-only <SW>^ <SW>` prints exactly `.moai/config/sections/workflow.yaml`; `git merge-base --is-ancestor <S4> <SW>` exit 0; with the by-path binary, `moai integration manifest verify --card <any-card>` prints `gate: on (source: <integration-branch>:.moai/config/sections/workflow.yaml @<SW>…)` — the observed effect, independent of the primary checkout's commit (AC-LVM-017 is the fixture form of the same observation).
- RED-now (pin as above): the M0 and M1 `git log` commands exit 0 with empty stdout; `git ls-files --error-unmatch …/evidence/m0-baseline.md` exit 1; `git grep -n -F -e landing_gate HEAD -- .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml` exit 1; E-012. The polarities of `merge-base --is-ancestor` were observed at the pin (self → 0; non-ancestor → 1). Green: M0..M4 landing in order.

### AC-LVM-013 — zero-Go card (REQ-LVM-013)

- **Given** a card that changes only a rule file no test reads, **Then** the manifest is empty with its derivation evidence (changed-path list, `go list` proof, reader-search result) and is admitted by the merge step, `complete` and condition (d) alike (positive); it is refused when the evidence is missing, and when re-derivation yields a different changed-path list (a later commit added a path); the audit binding applies unchanged.
- Verify (N=4): `go test ./internal/factory -run '^(TestZeroGoCardAdmittedWithEvidence|TestZeroGoCardRefusedWithoutEvidence|TestZeroGoCardRefusedOnStaleChangedPaths|TestZeroGoCardAdmittedThroughAllThreeConsumers)$' -count=1 -v`.
- RED-now: E-013. Green: M1.

### AC-LVM-014 — manifest producer, end to end (REQ-LVM-014)

- **Given** a fixture repository with a two-package change, **When** a lane runs `manifest derive --card <id>`, then each call through `remeasure --call <call-id> -- <command>`, then `manifest verify --card <id>`, **Then** the manifest exists at `.moai/state/remeasure/manifests/<card-id>.json` under the shared root and `verify` exits 0 printing `gate: on|off (source: …)`; after a re-absorb that adds an importer, `verify` exits non-zero until a fresh `derive` and the new call's attempt exist; `show` prints the manifest; a refusal names the class.
- Verify (N=5): `go test ./internal/cli -run '^(TestManifestDeriveWritesStoredManifest|TestManifestRegeneratedAfterReabsorb|TestManifestShowAndVerifyVerbs|TestManifestVerifyPrintsGateStateAndSource|TestDeriveRunVerifyEndToEnd)$' -count=1 -v`.
- RED-now: E-014. Green: M1.

### AC-LVM-015 — safety invariants (REQ-LVM-015)

- **Given** a gate-on landing, **Then** the existing audit check still runs and a mismatching audit refuses it; the new verbs write no integration-branch ref (a fake ref-writer fails on any write outside the merge step); each undecidable reading refuses; and a call whose command narrows below the whole package is refused unless the manifest entry itself names the `internal/cli` 35e9 narrowing (the class-12 mutant of AC-LVM-005 is the named witness).
- Verify (N=4): `go test ./internal/factory -run '^(TestGateNeverSkipsAuditCheck|TestNoIntegrationBranchWriteOutsideMergeStep|TestUndecidableRefuses|TestNarrowingAllowedOnlyWhenManifestNamesIt)$' -count=1 -v`.
- RED-now: E-015. Green: M1 (final check at M4).

### AC-LVM-016 — manifest currency by re-derivation (REQ-LVM-016)

- **Given** a manifest derived at an earlier HEAD that lists only package `p`, **When** a later commit changes package `q` and the lane runs only the existing `p` call at the current HEAD, **Then** every consumer refuses as the thirteenth class (stale manifest, `q` missing); **Given** a later commit that adds a reader of a changed non-Go file, **Then** the refusal names the reader; **Given** a later harmless commit that leaves the derived set unchanged, **Then** the stored manifest still equals the derivation (positive control for the manifest comparison only: the attempts stay bound to tree and HEAD, class 7, so the calls are re-run); **Given** a verifier process whose GOOS, GOARCH, CGO_ENABLED, GOFLAGS or tags differ from the manifest's recorded environment, **Then** the re-derivation runs under the recorded environment and no spurious class-13 refusal occurs; the merge step, `complete`, condition (d) and `manifest verify` each re-derive.
- Verify (N=5): `go test ./internal/factory -run '^(TestGateRejectsStaleManifestMissingCall|TestGateRejectsManifestWithNewReaderSince|TestAllConsumersReDeriveManifest|TestCurrentManifestSurvivesHarmlessLaterCommit|TestReDerivationUsesManifestEnvironmentNotVerifierEnv)$' -count=1 -v`.
- RED-now: E-016. Green: M1.

### AC-LVM-017 — gate state read from the integration branch's committed blob (REQ-LVM-017)

- **Given** a temporary repository with an integration branch `main`, a card worktree, and a stand-in primary checkout deliberately left detached on an older commit that never receives branch commits (the negative control; the real primary is never touched), with `CLAUDE_PROJECT_DIR` set to the stand-in and the reader run from the card worktree (the real resolution order of `integrationLockRoot()`), **When** a commit on `main` sets `workflow.landing_gate.enabled: true`, **Then** the reader reports on and the landing verbs apply the gate, even though the stand-in's file says false or is absent; **When** a following commit on `main` sets it false, **Then** the reader reports off; **Given** the card branch carries its own `enabled: true`, **Then** it is ignored; **Given** no integration branch configured (no window record, empty target), **Then** off — the project has not opted in; **Given** a branch configured whose ref does not exist, **Then** the landing is refused with the named error `gate_state_integration_branch_unresolvable`; **Given** the branch resolves and the path is absent, **Then** off, and **Given** the path is present but the read fails, the YAML is unparseable, or the value is non-boolean (`"yes"`, `1`), **Then** on with the notice `gate_state_unreadable` and an audit row — and a landing that satisfies the gate is still admitted (no deadlock for an unreadable file, including the card that repairs it). No-deadlock reasoning for the unresolvable branch: the merge step cannot run without that branch, and repairing the branch configuration or ref needs no landing.
- Verify: factory (N=11) `go test ./internal/factory -run '^(TestGateStateReadFromIntegrationBranchBlob|TestEnableCommitOnIntegrationBranchTakesEffect|TestRevertCommitOnIntegrationBranchTakesEffect|TestStalePrimaryCheckoutDoesNotAffectGateState|TestCardBranchCopyOfKeyIgnored|TestAbsentKeyReadsOff|TestNoIntegrationBranchConfiguredReadsOff|TestConfiguredButUnresolvableBranchRefusesNamed|TestUnreadableFileTreatedOnWithNotice|TestUnreadableFileStillAdmitsCompliantLanding|TestAbsentPathVersusFailedReadDiscriminated)$' -count=1 -v`; cli (N=1) `go test ./internal/cli -run '^TestCompleteReadsGateFromIntegrationBranch$' -count=1 -v`.
- RED-now: E-017a, E-017b. Green: M1.

### AC-LVM-018 — citation line grammar and its documentation (REQ-LVM-018)

- **Given** a located verdict (SPEC card or SPEC-less card), **Then** the line `audit_receipt: rcpt-<a>,rcpt-<b>` yields two ids; and the gate-on `complete` refuses when the line is missing, appears twice, is empty, reads `none`, carries a malformed id, or cites an id absent from the store, for a SPEC-less card as for a SPEC card; and the sync-auditor protocol, its template mirror and the audit artifact convention each contain the phrase `audit_receipt machine line`.
- Verify: factorylane (N=8) `go test ./internal/factorylane -run '^(TestAuditReceiptLineParsesCommaList|TestAuditReceiptLineMissingRefused|TestAuditReceiptLineNoneRefused|TestAuditReceiptLineEmptyRefused|TestAuditReceiptLineMalformedIdRefused|TestAuditReceiptLineDuplicateRefused|TestCitedIdAbsentFromStoreRefused|TestSPECLessCardCitationLineRequired)$' -count=1 -v`; one invocation per file (each exit 0): `grep -c -F -e "audit_receipt machine line" .claude/agents/moai/sync-auditor.md`, the same on `internal/template/templates/.claude/agents/moai/sync-auditor.md` and `.moai/docs/audit-artifact-convention.md`; `cmp .claude/agents/moai/sync-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md` (exit 0 at the pin and must stay 0: regression guard).
- RED-now: E-018 and the three phrase greps (each exit 1, recorded in the ledger note). Green: M3 (parser, tests) and M4 (documents).

## Mutant probes (verification-completeness §2)

| Criterion | Mutant that must fail | Caught by |
|---|---|---|
| AC-LVM-004 | verifier reads only the newest record | `TestLastRecordAloneInsufficient` |
| AC-LVM-005 | verifier ignores the tree key | `TestGateRejectsOtherTreeOrHead` (the positive control passes it, which is why both directions exist) |
| AC-LVM-005 | verifier counts an aborted attempt as pass | `TestGateRejectsAbortedAttempt`, `TestGateRejectsPartialJSON` |
| AC-LVM-005 | verifier accepts a script-wrapped exit-0 record | `TestGateRejectsScriptWrappedRecord` |
| AC-LVM-005 | verifier accepts `go test -json -run '^TestX$' ./pkg` | `TestGateRejectsCommandDiffersFromManifest` |
| AC-LVM-005 | verifier accepts a package that passed with zero tests | `TestGateRejectsEmptySweep` |
| AC-LVM-016 | verifier compares attempts only, never re-derives the manifest | `TestGateRejectsStaleManifestMissingCall` |
| AC-LVM-017 | state read from the primary checkout's working tree | `TestStalePrimaryCheckoutDoesNotAffectGateState` |
| AC-LVM-017 | unreadable file silently reads off | `TestUnreadableFileTreatedOnWithNotice` |
| AC-LVM-017 | an unconfigured project reads on, or a missing configured branch reads off | `TestNoIntegrationBranchConfiguredReadsOff`, `TestConfiguredButUnresolvableBranchRefusesNamed` |
| AC-LVM-016 | re-derivation under the verifier's own environment | `TestReDerivationUsesManifestEnvironmentNotVerifierEnv` |
| AC-LVM-008 | binding accepts an ancestor whose diff is empty or limited to the evidence directory | `TestEarlierPassNeverSubstitutes`, `TestEmptyDiffAncestorRefused`, `TestEvidenceDirDiffAfterAuditRefused` |
| AC-LVM-010 | receipt `tree_root` compared with the store root instead of the audited tree's path (refuses every orphaned-worktree receipt) | `TestReceiptForConfigOrphanedWorktreeQualifies`, `TestReceiptFromOtherTreeRootRefused` |
| AC-LVM-010 | binding matches a receipt by card id or time instead of the audited commit | `TestReceiptFromOlderHeadRefused`, `TestLegacyReceiptWithoutCommitRefused` |
| AC-LVM-010 | binding applies "every cited receipt" so a verdict with no citation passes vacuously | `TestAuditReceiptLineMissingRefused`, `TestAuditReceiptLineNoneRefused` |
| AC-LVM-012 | a branch commit enables the key and a later one removes it | the empty `git log <S0>^..<S4> -- workflow.yaml` |

## Quality gates and Definition of Done

- Every AC green with the empty-sweep guard satisfied; the RED-now cells re-run at M0 and committed as `evidence/m0-baseline.md`.
- `go vet` and `gofmt -l` clean on touched packages; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; package-scoped `golangci-lint`; scoped `-run` coverage for the added code with the remainder recorded as a CI-deferred Gap. Go calls are made only while the host admits them (resource rule 8c1e).
- `moai spec lint --strict` on this SPEC exit 0; no `\u` escapes; `make build` regenerated template assets (Template-First) with `git status --short` clean after.
- No skipped test, no excluded test in any manifest call, no widened allowance: the narrowing list diff shows `internal/cli` only.
- **Not claimed:** that this card's own landing runs under the gate. Piece A lands under the current rules (board `0577` P2); the first gated landing is the next one after a later, separate enabling commit made with a by-path binary built from main (spec C5).

## Edge cases

- A package with no test files: may sweep zero tests only where the manifest recorded `no_test_files` for it at derivation, naming the reason.
- A re-absorb that adds importers makes the stored manifest stale (REQ-LVM-016); a fresh `derive` replaces it and attempts bound to the superseded identity stop counting.
- A card whose only change is outside Go and has readers: the readers become calls; the empty-manifest path applies only when there are none.
- Recovery if the gate refuses its own reversal: the key-off one-path card lands through the gated path (its reader calls run, its audit is bound; an unreadable state reads on but admits a compliant landing). If the gated path itself is the defect, the operator-held manual edit of the integration branch's configuration is the accepted residual, with the operator as owner and not performed by an agent.
- Measured transition fact: of the 8 sampled merge-ready SPEC cards, 2 carry a top-level `audit_receipt:` line in their newest sync-audit verdict; the others cite receipts in prose and are re-audited at landing anyway (board `2c51`), producing the machine line.

---

## Evidence ledger (RED-now observations)

All entries: tree `8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2`, run 2026-10-19, unpiped (output redirected to a file and printed verbatim). `names_found` counts stdout lines of the form `func Test…(`.

#### E-001
```
command: grep -rhoE "^func (TestManifestDerivesChangedAndImporters|TestManifestNarrowingOnlyInternalCLI|TestManifestIncludesTestOnlyChange|TestManifestRecordsNonGoReaderSearch|TestManifestDerivesEmbedOwnerPackages|TestManifestHandlesDeletedPackage)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 6)
```
#### E-002
```
command: grep -rhoE "^func (TestAttemptRecordCarriesAllFields|TestAttemptPackageTerminalEvents|TestAttemptRecordsPerPackageTestCounts)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 3)
```
#### E-003
```
command: grep -rhoE "^func (TestAbortedAttemptNeverCounts|TestRestartMakesNewAttemptID|TestAbortClassesSignalTimeoutPartial)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 3)
```
#### E-004
```
command: grep -rhoE "^func (TestManifestWholePassPositiveControl|TestLastRecordAloneInsufficient|TestMostRecentCompletedAttemptCounts)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 3)
```
#### E-005
```
command: grep -rhoE "^func (TestGateRejectsMissingManifest|TestGateRejectsMissingCall|TestGateRejectsScriptWrappedRecord|TestGateRejectsMissingPackageEvent|TestGateRejectsPartialJSON|TestGateRejectsAbortedAttempt|TestGateRejectsOtherTreeOrHead|TestGateRejectsEnvFingerprintMismatch|TestGateRejectsUnverifiedNonGoInput|TestGateRejectsUnderivableScope|TestGateRejectsEmptySweep|TestGateRejectsCommandDiffersFromManifest)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 12)
```
#### E-006a
```
command: grep -rhoE "^func (TestMergeStepConsultsCompletenessVerifier)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 1)
```
#### E-006b
```
command: grep -rhoE "^func (TestFactoryCompleteConsultsCompletenessVerifier|TestManifestVerifyVerbSharesVerifier)\(" internal/cli --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 2)
```
#### E-006c
```
command: grep -rhoE "^func (TestMergeTripleConditionDUsesVerifier)\(" internal/factorylane --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 1)
```
#### E-007a
```
command: grep -rhoE "^func (TestLandingGateDefaultsOff|TestRunNarrowingAllowanceInternalCLIOnly)\(" internal/config --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 2)
```
#### E-007b
```
command: grep -rhoE "^func (TestGateOffMatchesTodaysBehaviour)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 1)
```
#### E-008a
```
command: grep -rhoE "^func (TestSyncAuditVerdictLocatorPicksNewestIteration|TestSPECCardWithoutSyncAuditVerdictRefused|TestSPECLessCardVerdictLocatedAtVerdictMd|TestSyncAuditVerdictMustNameCurrentHead|TestEarlierPassNeverSubstitutes|TestEmptyDiffAncestorRefused|TestProgressBackfillAfterAuditRefuses|TestEvidenceDirDiffAfterAuditRefused)\(" internal/factorylane --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 8)
```
#### E-008b
```
command: grep -rhoE "^func (TestCompleteRefusesStaleAuditHead)\(" internal/cli --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 1)
```
#### E-009a
```
command: grep -rhoE "^func (TestReceiptRoundTripsAuditedCommit|TestLegacyReceiptDecodesAsUnbound|TestExistingReceiptNeverRewritten)\(" internal/auditreceipt --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 3)
```
#### E-009b
```
command: grep -rhoE "^func (TestWriterStampsAuditedCommit|TestWriterLeavesEmptyWhenHeadMoved|TestEntryPointCapturesStartHead|TestWriterRecordsAuditTarget)\(" internal/cli --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 4)
```
#### E-010a
```
command: grep -rhoE "^func (TestReceiptAtCurrentHeadPasses|TestReceiptFromOlderHeadRefused|TestAbsentReceiptRefused|TestUnreadableReceiptRefused|TestLegacyReceiptWithoutCommitRefused|TestReceiptWithEmptyCommitRefused|TestUnmetGateReceiptRefused|TestReceiptFromOtherTreeRootRefused|TestReceiptForConfigOrphanedWorktreeQualifies|TestReceiptAtEmptyDiffAncestorRefused|TestFailedExtraReceiptDoesNotBlockButNeverReplacesPass)\(" internal/factorylane --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 11)
```
#### E-010b
```
command: grep -rhoE "^func (TestRealWriterReceiptSatisfiesBinding|TestRealWriterReceiptForOrphanedWorktreeSatisfiesBinding)\(" internal/cli --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 2)
```
#### E-012
```
command: grep -rhoE "^func (TestKeyOffCardLandsThroughGatedPath|TestKeyOffCardStillNeedsAuditBinding)\(" internal/cli --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 2)
```
#### E-013
```
command: grep -rhoE "^func (TestZeroGoCardAdmittedWithEvidence|TestZeroGoCardRefusedWithoutEvidence|TestZeroGoCardRefusedOnStaleChangedPaths|TestZeroGoCardAdmittedThroughAllThreeConsumers)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 4)
```
#### E-014
```
command: grep -rhoE "^func (TestManifestDeriveWritesStoredManifest|TestManifestRegeneratedAfterReabsorb|TestManifestShowAndVerifyVerbs|TestManifestVerifyPrintsGateStateAndSource|TestDeriveRunVerifyEndToEnd)\(" internal/cli --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 5)
```
#### E-015
```
command: grep -rhoE "^func (TestGateNeverSkipsAuditCheck|TestNoIntegrationBranchWriteOutsideMergeStep|TestUndecidableRefuses|TestNarrowingAllowedOnlyWhenManifestNamesIt)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 4)
```
#### E-016
```
command: grep -rhoE "^func (TestGateRejectsStaleManifestMissingCall|TestGateRejectsManifestWithNewReaderSince|TestAllConsumersReDeriveManifest|TestCurrentManifestSurvivesHarmlessLaterCommit|TestReDerivationUsesManifestEnvironmentNotVerifierEnv)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 5)
```
#### E-017a
```
command: grep -rhoE "^func (TestGateStateReadFromIntegrationBranchBlob|TestEnableCommitOnIntegrationBranchTakesEffect|TestRevertCommitOnIntegrationBranchTakesEffect|TestStalePrimaryCheckoutDoesNotAffectGateState|TestCardBranchCopyOfKeyIgnored|TestAbsentKeyReadsOff|TestNoIntegrationBranchConfiguredReadsOff|TestConfiguredButUnresolvableBranchRefusesNamed|TestUnreadableFileTreatedOnWithNotice|TestUnreadableFileStillAdmitsCompliantLanding|TestAbsentPathVersusFailedReadDiscriminated)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 11)
```
#### E-017b
```
command: grep -rhoE "^func (TestCompleteReadsGateFromIntegrationBranch)\(" internal/cli --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 1)
```
#### E-018
```
command: grep -rhoE "^func (TestAuditReceiptLineParsesCommaList|TestAuditReceiptLineMissingRefused|TestAuditReceiptLineNoneRefused|TestAuditReceiptLineEmptyRefused|TestAuditReceiptLineMalformedIdRefused|TestAuditReceiptLineDuplicateRefused|TestCitedIdAbsentFromStoreRefused|TestSPECLessCardCitationLineRequired)\(" internal/factorylane --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 8)
```

Non-Go cells, exit codes observed at the pin: the 18 `grep -q -F -e "<phrase>" <rule or mirror>` calls of AC-LVM-011 → 1 each; `grep -q -F -e "audit_receipt machine line"` on `.claude/agents/moai/sync-auditor.md`, its template mirror and `.moai/docs/audit-artifact-convention.md` → 1 each; `git grep -n -F -e landing_gate HEAD -- .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml` → 1; `git log --format=%H -1 --grep='^feat(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M1 '` → 0 with empty stdout; the M0 subject → 0, empty; `git ls-files --error-unmatch .moai/specs/SPEC-LANDING-VERIFY-MERGE-SPLIT-001/evidence/m0-baseline.md` → 1. Regression guards (green now): `cmp` of the factory-dispatch-mechanics rule and mirror → 0; `cmp` of the sync-auditor protocol and mirror → 0.
