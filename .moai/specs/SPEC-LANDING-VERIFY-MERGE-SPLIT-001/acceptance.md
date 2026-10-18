# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — acceptance criteria (v0.2.0, piece A)

Verification layer: each entry is an `AC-LVM-XXX` Given-When-Then, binary-testable. The GEARS requirements are `spec.md` §3. Counts: 12 requirements, 12 acceptance criteria (Tier L ceiling 25/25), one-to-one. The evidence ledger at the end holds 14 command observations because three criteria span two packages.

## Common setup

- **Test names are the interface.** The names below are proposals; the run phase may rename one only with a matching edit here (a rename leaves a `--- PASS` line missing, which reads red).
- **Empty-sweep guard (verification-completeness §1.1).** Every Go AC runs `go test ./internal/<pkg> -run '^(A|B|…)$' -count=1 -v` and requires together: (i) exactly one `--- PASS: <Name> ` line per listed name (N stated per AC), (ii) no `[no tests to run]`, (iii) exit 0. A selector matching zero tests exits 0, so the name count is the verdict.
- **RED-now cells.** Four elements per cell, carried in the evidence ledger at the end (the recommended carrier): command, verbatim stdout, exit code, tree SHA. Document-level pin: `ec164d45d2e3c5a7902aa950ab212e8f385f5d08` (main at authoring; the base of this card's branch), which binds every cell without its own pin. Every Go cell is `go test ./internal/<pkg> -list '^(names)$' -count=1`, observed with exit 0 and zero names listed; red reason: the named tests do not exist because the behaviour is not implemented (count 0 ≠ N). The command itself exits 0, which is why the criterion is the name count and never the exit code alone.
- **Judging build.** CLI observations used `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/c5bd52fe-ffbe-40e1-94ce-4491a79c0784/scratchpad/moai-t1872`, built with `go build ./cmd/moai` from the same clean tree (build commit = tree HEAD).
- **Green path.** Milestones (`plan.md`): M0 baseline; M1 gate core; M2 audit binding (Q13); M3 lane guidance; M4 the enable switch.
- **Fixtures.** Gate tests build temporary git repositories with `t.TempDir()` and canned `go test -json` streams; none reads this repository's refs, `.moai/state`, or the user's home.

### AC-LVM-001 — manifest derivation (REQ-LVM-001)

- **Given** a temporary repository whose card branch changes (a) a non-test file of package `p/a`, (b) only a test file of package `p/b`, (c) a file of `internal/cli`, (d) a non-Go file outside any package directory; `p/c` imports `p/a`; `p/d` imports `p/b` only in a test, **When** the manifest generator runs, **Then** it lists a whole-package call (no `-run`) for `p/a`, `p/b`, `internal/cli`, `p/c`, `p/d`; only the `internal/cli` call may carry `-run` and it names `35e9`; every call records identifier, command and resolved package set; and for (d) the manifest holds either reader calls or a zero-readers record that names the search command and its empty result.
- Verify (N=4): `go test ./internal/factory -run '^(TestManifestDerivesChangedAndImporters|TestManifestNarrowingOnlyInternalCLI|TestManifestIncludesTestOnlyChange|TestManifestRecordsNonGoReaderSearch)$' -count=1 -v`.
- RED-now: E-001. Green: M1.

### AC-LVM-002 — attempt record (REQ-LVM-002)

- **Given** a manifest call run in the structured form over a fixture stream of two packages, **When** the attempt finishes, **Then** its record carries a unique attempt ID, call identifier, command, exit code, one terminal event per stream package, tree SHA, HEAD SHA, base, environment fingerprint, start and end times, and termination class `completed`.
- Verify (N=2): `go test ./internal/factory -run '^(TestAttemptRecordCarriesAllFields|TestAttemptPackageTerminalEvents)$' -count=1 -v`.
- RED-now: E-002. Green: M1.

### AC-LVM-003 — abort versus restart (REQ-LVM-003)

- **Given** attempts ended by (a) SIGTERM, (b) the verb's timeout, (c) a stream cut mid-line, **Then** each is `aborted`, retained on disk, and counts as no pass; **When** the call is run again, **Then** it is a new attempt with a new ID, the aborted record is not overwritten, and only the most recent completed attempt on the same tree counts.
- Verify (N=3): `go test ./internal/factory -run '^(TestAbortedAttemptNeverCounts|TestRestartMakesNewAttemptID|TestAbortClassesSignalTimeoutPartial)$' -count=1 -v`.
- RED-now: E-003. Green: M1.

### AC-LVM-004 — whole manifest, positive control (REQ-LVM-004)

- **Given** a manifest of three calls, **When** all three have a completed exit-0 attempt on the candidate tree with full package coverage, **Then** the verifier reports gate-complete (positive control); **When** only the last-recorded call is green and the other two have no attempt, **Then** it does not; **When** a call has an aborted attempt followed by a completed one on the same tree, **Then** the completed one counts.
- Verify (N=3): `go test ./internal/factory -run '^(TestManifestWholePassPositiveControl|TestLastRecordAloneInsufficient|TestMostRecentCompletedAttemptCounts)$' -count=1 -v`.
- RED-now: E-004. Green: M1.

### AC-LVM-005 — the eleven reject classes (REQ-LVM-005)

- **Given** the AC-LVM-004 positive fixture, **When** exactly one mutation is applied — (1) manifest removed, (2) one call's attempts removed, (3) one call's only record is a script-wrapped command with no package report, (4) one stream lacks an expected package's terminal event, (5) one stream cut mid-line, (6) one attempt aborted, (7) one attempt recorded on another tree or HEAD, (8) one attempt's GOFLAGS/tags differ from the manifest fingerprint, (9) a changed non-Go input with neither reader call nor zero-readers record, (10) `go list` fails so the scope cannot be derived, (11) a package swept zero tests where the manifest does not allow it — **Then** the verifier refuses in each case, naming the class; and the unmutated fixture still passes (control).
- Verify (N=11): `go test ./internal/factory -run '^(TestGateRejectsMissingManifest|TestGateRejectsMissingCall|TestGateRejectsScriptWrappedRecord|TestGateRejectsMissingPackageEvent|TestGateRejectsPartialJSON|TestGateRejectsAbortedAttempt|TestGateRejectsOtherTreeOrHead|TestGateRejectsEnvFingerprintMismatch|TestGateRejectsUnverifiedNonGoInput|TestGateRejectsUnderivableScope|TestGateRejectsEmptySweep)$' -count=1 -v`.
- RED-now: E-005. Green: M1.

### AC-LVM-006 — one verifier, three consumers (REQ-LVM-006)

- **Given** the gate enabled and a tree whose only record is a script-wrapped call (exit 0), **When** the merge step, `moai factory complete`, and merge-readiness condition (d) consult the tree, **Then** all three refuse through the same verifier (same reason text); a gate-complete tree passes all three.
- Verify: factory (N=1) `go test ./internal/factory -run '^TestMergeStepConsultsCompletenessVerifier$' -count=1 -v`; cli (N=1) `go test ./internal/cli -run '^TestFactoryCompleteConsultsCompletenessVerifier$' -count=1 -v`; factorylane (N=1) `go test ./internal/factorylane -run '^TestMergeTripleConditionDUsesVerifier$' -count=1 -v`.
- RED-now: E-006a, E-006b, E-006c. Green: M1.

### AC-LVM-007 — default off, narrowing list (REQ-LVM-007)

- **Given** a fresh config and a project with no `workflow.landing_gate` block, **When** the landing path runs a card whose record is a script-wrapped call, **Then** the outcome equals the pre-change outcome; and the shipped narrowing allowance list is exactly `[internal/cli]`.
- Verify: config (N=2) `go test ./internal/config -run '^(TestLandingGateDefaultsOff|TestRunNarrowingAllowanceInternalCLIOnly)$' -count=1 -v`; factory (N=1) `go test ./internal/factory -run '^TestGateOffMatchesTodaysBehaviour$' -count=1 -v`.
- RED-now: E-007a, E-007b. Green: M1.

### AC-LVM-008 — verdict bound to the current HEAD (REQ-LVM-008)

- **Given** the gate enabled and a SPEC card whose sync-audit verdict names an earlier commit, **When** `complete` evaluates it, **Then** it refuses (a) when the change since that commit is non-empty, and (b) also when it is empty or patch-identical (an earlier PASS never substitutes); **Given** a verdict naming HEAD, or naming a commit that differs from HEAD only under the card's evidence directory, **Then** the check passes (the allowance the SPEC-less path already applies).
- Verify: factorylane (N=3) `go test ./internal/factorylane -run '^(TestSyncAuditVerdictMustNameCurrentHead|TestEarlierPassNeverSubstitutes|TestEvidenceOnlyDiffToleratedAsToday)$' -count=1 -v`; cli (N=1) `go test ./internal/cli -run '^TestCompleteRefusesStaleAuditHead$' -count=1 -v`.
- RED-now: E-008a, E-008b. Green: M2.

### AC-LVM-009 — receipt at the current HEAD (REQ-LVM-009)

- **Given** the gate enabled, **When** the required backend's receipt (a) exists for HEAD and reads pass, **Then** the check passes; **When** it is (b) from an earlier HEAD, (c) absent, or (d) unreadable, **Then** `complete` refuses.
- Verify (N=2): `go test ./internal/factorylane -run '^(TestRequiredBackendReceiptAtHeadMustPass|TestReceiptFromOlderHeadRefused)$' -count=1 -v`.
- RED-now: E-009. Green: M2.

### AC-LVM-010 — lane guidance (REQ-LVM-010)

- **Given** the run-phase tree after M3, **Then** `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` and its template mirror each contain the literal phrases `re-measure completeness gate` and `structured form`, and the two files are byte-identical.
- Verify, one invocation each (exit 0, count ≥ 1): `grep -c -F -e "re-measure completeness gate" <file>` and `grep -c -F -e "structured form" <file>` for the rule and the mirror; `cmp .claude/rules/moai/workflow/factory-dispatch-mechanics.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md` (exit 0).
- RED-now (pin as above): `grep -n -F -e "re-measure completeness gate"` and `-e "structured form"` on the rule and the mirror → each exit 1, empty stdout (four observations, recorded in the ledger note below). Regression guard (green now, must stay green): the `cmp` prints nothing, exit 0. Green: M3.

### AC-LVM-011 — the ordered switch (REQ-LVM-011)

- **Given** the run-phase branch, **When** the milestone commits are located by subject and compared in the commit graph, **Then** M0 precedes M1, M1 precedes M2, M2 precedes M3, M3 precedes M4; M0 changes only the baseline file; M4 changes exactly one path; the key is absent or false at M3 and true at M4; the negative and positive gate tests pass on the M4 tree; and the reverse ordering fails (negative control).
- Verify, one invocation each: `git log --format=%H -1 --grep='^docs(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M0 '` and the same with `^feat(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M<k> ` for k = 1..4 (each prints one 40-hex SHA, exit 0) → S0..S4; `git merge-base --is-ancestor <S0> <S1>`, `<S1> <S2>`, `<S2> <S3>`, `<S3> <S4>` (each exit 0); negative control `git merge-base --is-ancestor <S4> <S0>` (exit 1); `git diff --name-only <S0>^ <S0>` (prints exactly `.moai/specs/SPEC-LANDING-VERIFY-MERGE-SPLIT-001/evidence/m0-baseline.md`); `git ls-files --error-unmatch .moai/specs/SPEC-LANDING-VERIFY-MERGE-SPLIT-001/evidence/m0-baseline.md` (exit 0: a tracked file outside `.moai/reports`); `git diff --name-only <S3> <S4>` (prints exactly `.moai/config/sections/workflow.yaml`); `git diff -U0 <S3> <S4> -- .moai/config/sections/workflow.yaml` (the added lines are exactly the `landing_gate:` block with `enabled: true`); `git grep -n -F -e "landing_gate" <S3> -- .moai/config/sections/workflow.yaml` (exit 1: absent before the switch); and the AC-LVM-001..009 test commands run on the tree at S4 (each green; the S3→S4 diff being configuration only, this witnesses the tests green before the switch).
- RED-now (pin as above): `git log --format=%H -1 --grep='^feat(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M1 '` → exit 0, empty stdout; same for the M0 subject → exit 0, empty; `git ls-files --error-unmatch …/evidence/m0-baseline.md` → exit 1 ("did not match any file(s) known to git"); `git grep -n -F -e "landing_gate"` on `.moai/config/sections/workflow.yaml` → exit 1. The two exit polarities of `merge-base --is-ancestor` were observed at the pin (HEAD of HEAD → 0; HEAD of main-at-authoring → 1). Red reason: no milestone commit exists, so there is no SHA to compare. Green: M0..M4 landing in order.

### AC-LVM-012 — safety invariants (REQ-LVM-012)

- **Given** a reused-nothing, gate-enabled landing, **Then** the existing audit check still runs and a mismatching audit refuses it; **Given** the new verbs, **Then** none writes the integration branch ref (a fake ref-writer fails on any write outside the merge step); **Given** each undecidable reading, **Then** the outcome is refusal; and no manifest call omits a test (a call's test selection equals the whole package, except the named `internal/cli` allowance).
- Verify (N=3): `go test ./internal/factory -run '^(TestGateNeverSkipsAuditCheck|TestNoIntegrationBranchWriteOutsideMergeStep|TestUndecidableRefuses)$' -count=1 -v`.
- RED-now: E-012. Green: M1 (final check at M4).

## Mutant probes (verification-completeness §2)

| Criterion | Mutant that must fail | Caught by |
|---|---|---|
| AC-LVM-004 | verifier reads only the newest record | `TestLastRecordAloneInsufficient` |
| AC-LVM-005 | verifier ignores the tree key | `TestGateRejectsOtherTreeOrHead` (the positive control passes it, which is why both directions exist) |
| AC-LVM-005 | verifier counts an aborted attempt as pass | `TestGateRejectsAbortedAttempt`, `TestGateRejectsPartialJSON` |
| AC-LVM-005 | verifier accepts a script-wrapped exit-0 record | `TestGateRejectsScriptWrappedRecord` |
| AC-LVM-008 | binding accepts an ancestor whose diff is empty | `TestEarlierPassNeverSubstitutes` |
| AC-LVM-009 | receipt matched by card id instead of HEAD | `TestReceiptFromOlderHeadRefused` |
| AC-LVM-011 | switch lands with a code change in the same commit | `git diff --name-only <S3> <S4>` |

## Quality gates and Definition of Done

- Every AC green with the empty-sweep guard satisfied; the RED-now cells re-run at M0 and committed as `evidence/m0-baseline.md`.
- `go vet` and `gofmt -l` clean on touched packages; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; package-scoped `golangci-lint`; scoped `-run` coverage for added code with the remainder recorded as a CI-deferred Gap.
- `moai spec lint --strict` on this SPEC exit 0; no `\u` escapes in the catalogue; `make build` regenerated template assets (Template-First) with `git status --short` clean after.
- No skipped test, no excluded test in any manifest call, no widened allowance: the diff of the narrowing list shows `internal/cli` only.
- The card's own landing runs under the enabled gate: its record is a manifest pass in the structured form (the first live use).

## Edge cases

- A card that changes no Go package: the manifest holds zero Go calls and `no_go_changes: true`; readiness then needs the card's declared non-Go verification call, and with none declared it is refused (fail-closed).
- A package with no test files: may sweep zero tests only where the manifest recorded `no_test_files` for it at derivation (the clarification of board `148c` applies: such a package is named in the record with its reason, not skipped silently).
- A re-absorb that adds importers regenerates the manifest; attempts bound to the superseded manifest identity stop counting.

---

## Evidence ledger (RED-now observations)

All entries: tree `ec164d45d2e3c5a7902aa950ab212e8f385f5d08`, run 2026-10-19, unpiped (output redirected to a file and printed verbatim). `names_listed` counts stdout lines of the form `Test…`.

Non-Go cells (AC-LVM-010, 011), each exit code observed at the pin: `grep -n -F -e "re-measure completeness gate" <rule>` → 1; same on the template mirror → 1; `grep -n -F -e "structured form" <rule>` → 1; mirror → 1; `grep -n -F -e "landing_gate" .moai/config/sections/workflow.yaml` → 1; `git log --format=%H -1 --grep='^feat(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M1 '` → 0 with empty stdout; the M0 subject → 0 with empty stdout; `git ls-files --error-unmatch .moai/specs/SPEC-LANDING-VERIFY-MERGE-SPLIT-001/evidence/m0-baseline.md` → 1.

#### E-001
```
command: go test ./internal/factory -list '^(TestManifestDerivesChangedAndImporters|TestManifestNarrowingOnlyInternalCLI|TestManifestIncludesTestOnlyChange|TestManifestRecordsNonGoReaderSearch)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.504s
names_listed: 0 (required: 4)
```
#### E-002
```
command: go test ./internal/factory -list '^(TestAttemptRecordCarriesAllFields|TestAttemptPackageTerminalEvents)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.395s
names_listed: 0 (required: 2)
```
#### E-003
```
command: go test ./internal/factory -list '^(TestAbortedAttemptNeverCounts|TestRestartMakesNewAttemptID|TestAbortClassesSignalTimeoutPartial)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.431s
names_listed: 0 (required: 3)
```
#### E-004
```
command: go test ./internal/factory -list '^(TestManifestWholePassPositiveControl|TestLastRecordAloneInsufficient|TestMostRecentCompletedAttemptCounts)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.367s
names_listed: 0 (required: 3)
```
#### E-005
```
command: go test ./internal/factory -list '^(TestGateRejectsMissingManifest|TestGateRejectsMissingCall|TestGateRejectsScriptWrappedRecord|TestGateRejectsMissingPackageEvent|TestGateRejectsPartialJSON|TestGateRejectsAbortedAttempt|TestGateRejectsOtherTreeOrHead|TestGateRejectsEnvFingerprintMismatch|TestGateRejectsUnverifiedNonGoInput|TestGateRejectsUnderivableScope|TestGateRejectsEmptySweep)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.408s
names_listed: 0 (required: 11)
```
#### E-006a
```
command: go test ./internal/factory -list '^(TestMergeStepConsultsCompletenessVerifier)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.362s
names_listed: 0 (required: 1)
```
#### E-006b
```
command: go test ./internal/cli -list '^(TestFactoryCompleteConsultsCompletenessVerifier)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/cli	1.777s
names_listed: 0 (required: 1)
```
#### E-006c
```
command: go test ./internal/factorylane -list '^(TestMergeTripleConditionDUsesVerifier)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factorylane	0.381s
names_listed: 0 (required: 1)
```
#### E-007a
```
command: go test ./internal/config -list '^(TestLandingGateDefaultsOff|TestRunNarrowingAllowanceInternalCLIOnly)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/config	0.441s
names_listed: 0 (required: 2)
```
#### E-007b
```
command: go test ./internal/factory -list '^(TestGateOffMatchesTodaysBehaviour)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.324s
names_listed: 0 (required: 1)
```
#### E-008a
```
command: go test ./internal/factorylane -list '^(TestSyncAuditVerdictMustNameCurrentHead|TestEarlierPassNeverSubstitutes|TestEvidenceOnlyDiffToleratedAsToday)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factorylane	0.218s
names_listed: 0 (required: 3)
```
#### E-008b
```
command: go test ./internal/cli -list '^(TestCompleteRefusesStaleAuditHead)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/cli	1.493s
names_listed: 0 (required: 1)
```
#### E-009
```
command: go test ./internal/factorylane -list '^(TestRequiredBackendReceiptAtHeadMustPass|TestReceiptFromOlderHeadRefused)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factorylane	0.211s
names_listed: 0 (required: 2)
```
#### E-012
```
command: go test ./internal/factory -list '^(TestGateNeverSkipsAuditCheck|TestNoIntegrationBranchWriteOutsideMergeStep|TestUndecidableRefuses)$' -count=1
tree: ec164d45d2e3c5a7902aa950ab212e8f385f5d08
exit: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/factory	0.327s
names_listed: 0 (required: 3)
```
