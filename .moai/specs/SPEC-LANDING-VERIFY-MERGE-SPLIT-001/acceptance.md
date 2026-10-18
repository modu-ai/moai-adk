# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — acceptance criteria (v0.9.0, piece A2)

Verification layer: each entry is an `AC-LVM-XXX` Given-When-Then, binary-testable. The GEARS requirements are `spec.md` §3. Counts: 7 requirements, 7 acceptance criteria (Tier L ceiling 25/25), one-to-one. The criteria of the requirements that moved to SPEC-LANDING-COMPLETENESS-GATE-001 are listed in `spec.md` §7 and are not carried here; their pre-split text is preserved at commit `9713b6757d`. The evidence ledger at the end holds 12 Go-name observations (several criteria span two packages) and 23 non-Go observations, each with its four elements.

## Common setup

- **Test names are the interface.** The names are proposals; the run phase may rename one only with a matching edit here (a rename leaves a `--- PASS` line missing, which reads red).
- **Empty-sweep guard (verification-completeness §1.1).** Every Go AC runs `go test ./internal/<pkg> -run '^(A|B|…)$' -count=1 -v` and requires together: (i) exactly one `--- PASS: <Name> ` line per listed name (N stated per AC), (ii) no `[no tests to run]`, (iii) exit 0. A selector matching zero tests exits 0, so the name count is the verdict.
- **RED-now cells.** Four elements per cell, carried in the evidence ledger at the end: command, verbatim stdout, exit code, tree SHA. Document-level pin: `8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2`, which binds every cell without its own pin. No file under `internal/`, `.claude/` or `cmd/` differs between that pin and the authoring HEAD `49df956ac8` (`git diff --name-only 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2 HEAD -- internal .claude cmd` printed nothing, exit 0), so a cell observed at the pin and one observed at that HEAD read the same tree for the surfaces checked. A Go cell is `grep -rhoE "^func (A|B)\(" internal/<pkg> --include='*_test.go'`, observed with exit 1 and empty stdout; red reason: the named test functions do not exist because the behaviour is not implemented (found 0 ≠ N). This form needs no Go build (resource rule 8c1e); the green command is the `go test` line of each AC.
- **Judging build.** CLI observations used the prebuilt `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/c5bd52fe-ffbe-40e1-94ce-4491a79c0784/scratchpad/moai-t1872`; its `version` output prints no commit, so "built from `ec164d45d2`" is the lead's statement, not an observation.
- **Green path.** Milestones (`plan.md`): M0 baseline and the characterization test of today's `complete`; M1 receipt commit field; M2 audit binding (the gate state enters as a seam that reads off); M3 gate-state reader, key and status display; M4 lane guidance. The enabling switch is not a milestone of this branch (REQ-LVM-012).
- **Fixtures.** Gate and binding tests build temporary git repositories with `t.TempDir()`; none reads this repository's refs, `.moai/state`, or the user's home, and none touches the real primary checkout. Receipt fixtures use the real writer function, not hand-built JSON.

### AC-LVM-007 — default off (REQ-LVM-007)

- **Given** a fresh config and no `workflow.landing_gate` block, **Then** the zero-value configuration and the template `workflow.yaml` both read the key false; and `moai factory complete` for a card whose verdict names an earlier commit and cites no bound receipt reaches the same outcome as it does today (oracle: the characterization test `TestCompleteGateOffMatchesTodaysBehaviour`, committed in the M0 commit on the unchanged code; AC-LVM-012 witnesses by ancestry that it is in the graph ahead of the edit commit, and by path that the edit commit does not contain it).
- Verify: config (N=1) `go test ./internal/config -run '^TestLandingGateDefaultsOff$' -count=1 -v`; cli (N=1) `go test ./internal/cli -run '^TestCompleteGateOffMatchesTodaysBehaviour$' -count=1 -v`.
- RED-now: E-007a, E-007b. Green: M0 (characterization test, green on the unchanged code) and M3 (key, defaults).

### AC-LVM-008 — verdict locator and equals-HEAD (REQ-LVM-008)

- **Given** the gate on and a SPEC card whose `.moai/reports/<card-id>/` holds `sync-audit.md`, `sync-audit-iter2.md`, `sync-audit-iter10.md`, **Then** the located verdict is `-iter10` (numeric, not lexical); a SPEC card with only a `verdict.md` or with no file is refused; a SPEC-less card is located at `.moai/reports/<card-id>/verdict.md` and refused when that file is absent; `complete` passes only when the verdict's audited commit equals HEAD, and refuses when it is an earlier commit (a) with a non-empty change since, (b) with an empty or patch-identical change since — including an `--allow-empty` commit or a no-change merge after the audited commit, which is the empty-diff-ancestor negative — (c) followed by a `progress.md` edit under the SPEC directory, or (d) followed by a commit that touches only `.moai/reports/<card-id>/`; there is no tolerated difference; an audited commit written as an abbreviation of 12 to 40 hex digits that is a prefix of HEAD passes, while one of 11 digits or fewer, or naming another commit, refuses; and every refusal neither acquires nor releases the integration window and changes no record, the card staying at `merge-ready` (a window the lane already holds stays held).
- Verify: factorylane (N=8) `go test ./internal/factorylane -run '^(TestSyncAuditVerdictLocatorPicksNewestIteration|TestSPECCardWithoutSyncAuditVerdictRefused|TestSPECLessCardVerdictLocatedAtVerdictMd|TestSyncAuditVerdictMustNameCurrentHead|TestEarlierPassNeverSubstitutes|TestEmptyDiffAncestorRefused|TestProgressBackfillAfterAuditRefuses|TestEvidenceDirDiffAfterAuditRefused)$' -count=1 -v`; cli (N=1) `go test ./internal/cli -run '^TestCompleteRefusesStaleAuditHead$' -count=1 -v`.
- RED-now: E-008a, E-008b. Green: M2.

### AC-LVM-009 — receipt commit and target fields, writer and entry points (REQ-LVM-009)

- **Given** the real writer called for a tree at HEAD `H` for review target `T`, **Then** the receipt carries `audited_commit` `H` (full 40-hex) and `audit_target` `T`; **Given** the tree's HEAD moved while the call ran, **Then** `audited_commit` is empty; **Given** a real entry point (`codex_audit` with a mocked backend that moves HEAD during the call), **Then** the receipt carries the HEAD captured at entry or an empty field, never the later HEAD; **Given** a legacy receipt JSON without the fields, **Then** it decodes as unbound and the file is byte-identical after a read-and-verify pass.
- Verify: auditreceipt (N=3) `go test ./internal/auditreceipt -run '^(TestReceiptRoundTripsAuditedCommit|TestLegacyReceiptDecodesAsUnbound|TestExistingReceiptNeverRewritten)$' -count=1 -v`; cli (N=4) `go test ./internal/cli -run '^(TestWriterStampsAuditedCommit|TestWriterLeavesEmptyWhenHeadMoved|TestEntryPointCapturesStartHead|TestWriterRecordsAuditTarget)$' -count=1 -v`.
- RED-now: E-009a, E-009b. Green: M1.

### AC-LVM-010 — receipt binding in `complete` (REQ-LVM-010)

- **Given** the gate on and a verdict citing receipts, **Then** `complete` passes when one cited receipt, looked up at the store root of the audited tree, is readable, a codex or multi-backend audit receipt, has `tree_root` equal to `auditreceipt.Canonical` of the audited tree's path (a symlinked spelling of the same tree qualifies), an empty unmet gate, verdict pass, and `audited_commit` equal to HEAD; and each of (b) an older-HEAD receipt, (c) an absent receipt, (d) an unreadable (corrupt) receipt, (e) a legacy receipt without the field, (f) a receipt with the field empty, (g) a receipt with a non-empty unmet gate, (h) a receipt whose `tree_root` is another tree's path for the same commit, (i) a receipt whose commit is an ancestor of HEAD with an empty diff, fails to qualify and refuses when it is the only citation; **Given** a config-orphaned card worktree `W` (no configuration of its own) whose primary checkout is `P`, **Then** a receipt filed at `P`'s store with `tree_root` = `W` qualifies for `W` (positive control; the store root is not the compared quantity) and a receipt filed there with `tree_root` = `P` for the same commit is refused for `W`; a verdict citing one qualifying receipt plus one failed or inconclusive extra call passes (the extra is recorded), and one citing only failed calls refuses. The passing fixtures' receipts are produced by the real writer.
- Verify: factorylane (N=11) `go test ./internal/factorylane -run '^(TestReceiptAtCurrentHeadPasses|TestReceiptFromOlderHeadRefused|TestAbsentReceiptRefused|TestUnreadableReceiptRefused|TestLegacyReceiptWithoutCommitRefused|TestReceiptWithEmptyCommitRefused|TestUnmetGateReceiptRefused|TestReceiptFromOtherTreeRootRefused|TestReceiptForConfigOrphanedWorktreeQualifies|TestReceiptAtEmptyDiffAncestorRefused|TestFailedExtraReceiptDoesNotBlockButNeverReplacesPass)$' -count=1 -v`; cli (N=2) `go test ./internal/cli -run '^(TestRealWriterReceiptSatisfiesBinding|TestRealWriterReceiptForOrphanedWorktreeSatisfiesBinding)$' -count=1 -v`.
- RED-now: E-010a, E-010b. Green: M2.

### AC-LVM-012 — the branch never enables; the switch is separate; recovery is ordinary; the lane guidance (REQ-LVM-012)

- **Given** the run-phase branch, **Then** the milestone commits are in graph order M0 < M1 < M2 < M3 < M4; M0 changes exactly two files, the baseline and the characterization test; the M0 commit does not touch `internal/cli/factory_card.go`, and the commit that edits it (M2) does not contain the test file; no commit of the branch touches the repository's own `.moai/config/sections/workflow.yaml`; the reverse ordering fails (negative control); **Given** a later one-path card that sets the key false, **Then** `complete` applies the audit binding to it (it proceeds past the binding only with a verdict and a bound receipt at its HEAD, and is refused when the binding fails); and `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` and its template mirror each contain every literal phrase below and are byte-identical.
- Verify, one invocation each: `git log --format=%H -1 --grep='^test(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M0 '` and `^feat(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M<k> ` for k = 1..4 (each prints one 40-hex SHA, exit 0) → S0..S4; `git merge-base --is-ancestor <S0> <S1>`, `<S1> <S2>`, `<S2> <S3>`, `<S3> <S4>` (each exit 0); negative control `<S4> <S0>` (exit 1); `git diff --name-only <S0>^ <S0>` (exactly the two paths `.moai/specs/SPEC-LANDING-VERIFY-MERGE-SPLIT-001/evidence/m0-baseline.md` and `internal/cli/factory_complete_gate_off_test.go`); `git ls-files --error-unmatch` on each (exit 0); `git diff --name-only <S0>^ <S0> -- internal/cli/factory_card.go` (exit 0, empty stdout); `git diff --name-only <S2>^ <S2> -- internal/cli/factory_complete_gate_off_test.go` (exit 0, empty stdout: the edit commit does not contain the test); `git grep -c -F -e TestCompleteGateOffMatchesTodaysBehaviour <S0> -- internal/cli/factory_complete_gate_off_test.go` (exit 0); `git log --format=%H --reverse <S0>..<S4> -- internal/cli/factory_card.go` (the first printed SHA is S2); `git log --format=%H <S0>^..<S4> -- .moai/config/sections/workflow.yaml` (exit 0, **empty** stdout: the branch never enables, which also rejects an enable-then-revert history); cli (N=2) `go test ./internal/cli -run '^(TestKeyOffCardLandsThroughGatedPath|TestKeyOffCardStillNeedsAuditBinding)$' -count=1 -v`; the guidance, 8 calls (each exit 0): `grep -c -F -e "<phrase>" <file>` for the phrases `committed content of the integration branch`, `gate_state_unreadable`, `built from the landing HEAD`, `before the final audit` on the rule and on its template mirror; and `cmp .claude/rules/moai/workflow/factory-dispatch-mechanics.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md` (exit 0).
- Post-landing witness (regression-guard class, not release-blocking: the switch commit does not exist at this SPEC's landing, `verification-completeness` §2.1 undecidable disposition), run by the leader when the switch commit `<SW>` exists (made only after this SPEC and SPEC-LANDING-COMPLETENESS-GATE-001 have both landed, board 6404): `git diff --name-only <SW>^ <SW>` prints exactly `.moai/config/sections/workflow.yaml`; `git merge-base --is-ancestor <S4> <SW>` exit 0; with the by-path binary, `moai integration status` prints `gate: on (source: <integration-branch>:.moai/config/sections/workflow.yaml @<SW>…)` — the observed effect, independent of the primary checkout's commit (AC-LVM-017 is the fixture form of the same observation).
- RED-now (pin as above): ledger entries N-001 to N-008 (the guidance phrases, each exit 1), N-014 (`landing_gate` absent, exit 1), N-015 and N-016 (the M0 and M1 subject searches, exit 0 with empty stdout), N-017 and N-018 (the two paths untracked, exit 1), N-019 and N-020 (the polarities of `merge-base --is-ancestor`), E-012. Regression guard (green now, must stay green): N-021 (the rule and its mirror, `cmp` exit 0). Green: M0..M4 landing in order (graph legs), M2/M3 (cli tests), M4 (guidance).

### AC-LVM-017 — gate state read from the integration branch's committed blob (REQ-LVM-017)

- **Given** a temporary repository with an integration branch `main`, a card worktree, and a stand-in primary checkout deliberately left detached on an older commit that never receives branch commits (the negative control; the real primary is never touched), with `CLAUDE_PROJECT_DIR` set to the stand-in and the reader run from the card worktree (the real resolution order of `integrationLockRoot()`), **When** a commit on `main` sets `workflow.landing_gate.enabled: true`, **Then** the reader reports on and `complete` applies the audit binding, even though the stand-in's file says false or is absent; **When** a following commit on `main` sets it false, **Then** the reader reports off; **Given** the card branch carries its own `enabled: true`, **Then** it is ignored; **Given** no integration branch configured (no window record, empty target), **Then** off — the project has not opted in; **Given** a tag with the same name as the integration branch holding `enabled: false` while the branch holds `true`, **Then** the reader reports on from the branch and the displayed `@<sha>` is the branch commit (the tag never shadows the branch); **Given** a branch configured whose `refs/heads/<branch>` does not resolve (a tag of that name alone does not count), **Then** the landing is refused with the named error `gate_state_integration_branch_unresolvable`; **Given** the branch resolves and the path is absent, **Then** off, and **Given** the path is present but the read fails, the YAML is unparseable, or the value is non-boolean (`"yes"`, `1`), **Then** on with the notice `gate_state_unreadable` and one JSON line (event, branch, SHA, path, cause, UTC time) appended to `.moai/logs/landing-gate-audit.jsonl` under the integration lock root — and a landing that satisfies the gate is still admitted (no deadlock for an unreadable file, including the card that repairs it). `complete` exits non-zero before any state change when it refuses; `moai integration status` prints `gate: on|off (source: …)` with the notice or the named error when one applies and keeps its own exit status. No-deadlock reasoning for the unresolvable branch: the merge step cannot run without that branch, and repairing the branch configuration or ref needs no landing.
- Verify: factory (N=12) `go test ./internal/factory -run '^(TestSameNamedTagDoesNotShadowBranch|TestGateStateReadFromIntegrationBranchBlob|TestEnableCommitOnIntegrationBranchTakesEffect|TestRevertCommitOnIntegrationBranchTakesEffect|TestStalePrimaryCheckoutDoesNotAffectGateState|TestCardBranchCopyOfKeyIgnored|TestAbsentKeyReadsOff|TestNoIntegrationBranchConfiguredReadsOff|TestConfiguredButUnresolvableBranchRefusesNamed|TestUnreadableFileTreatedOnWithNotice|TestUnreadableFileStillAdmitsCompliantLanding|TestAbsentPathVersusFailedReadDiscriminated)$' -count=1 -v`; cli (N=2) `go test ./internal/cli -run '^(TestCompleteReadsGateFromIntegrationBranch|TestIntegrationStatusPrintsGateStateAndSource)$' -count=1 -v`.
- RED-now: E-017a, E-017b. Green: M3.

### AC-LVM-018 — citation line grammar and its documentation (REQ-LVM-018)

- **Given** a located verdict (SPEC card or SPEC-less card), **Then** the line `audit_receipt: rcpt-<a>,rcpt-<b>` yields two ids; and the gate-on `complete` refuses when the line is missing, appears twice, is empty, reads `none`, carries a malformed id, or cites an id absent from the store, for a SPEC-less card as for a SPEC card; and the sync-auditor protocol, its template mirror, the audit artifact convention and the two Codex agent definitions emitted from the protocol (the template `.toml` regenerated by `make agents-emit`, the local `.toml` copied from it) each contain the phrase `audit_receipt machine line`.
- Verify: factorylane (N=8) `go test ./internal/factorylane -run '^(TestAuditReceiptLineParsesCommaList|TestAuditReceiptLineMissingRefused|TestAuditReceiptLineNoneRefused|TestAuditReceiptLineEmptyRefused|TestAuditReceiptLineMalformedIdRefused|TestAuditReceiptLineDuplicateRefused|TestCitedIdAbsentFromStoreRefused|TestSPECLessCardCitationLineRequired)$' -count=1 -v`; one invocation per file (each exit 0): `grep -c -F -e "audit_receipt machine line" .claude/agents/moai/sync-auditor.md`, the same on `internal/template/templates/.claude/agents/moai/sync-auditor.md` and `.moai/docs/audit-artifact-convention.md`; the same on `.codex/agents/moai/sync-auditor.toml` and `internal/template/templates/.codex/agents/moai/sync-auditor.toml`; `cmp .claude/agents/moai/sync-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md` and `cmp .codex/agents/moai/sync-auditor.toml internal/template/templates/.codex/agents/moai/sync-auditor.toml` (each exit 0 at the pin and must stay 0: regression guards); green command for the emitted copy: `make agents-emit-check` (exit 0; a Go call, made only while the host admits it).
- RED-now: E-018 and N-009 to N-013 (the five phrase greps, each exit 1); guards N-022 and N-023. Green: M2 (parser, tests) and M4 (documents and regeneration).

## Mutant probes (verification-completeness §2)

| Criterion | Mutant that must fail | Caught by |
|---|---|---|
| AC-LVM-007 | the binding applied while the key is false or absent | `TestCompleteGateOffMatchesTodaysBehaviour` |
| AC-LVM-008 | binding accepts an ancestor whose diff is empty or limited to the evidence directory | `TestEarlierPassNeverSubstitutes`, `TestEmptyDiffAncestorRefused`, `TestEvidenceDirDiffAfterAuditRefused` |
| AC-LVM-008 | locator ranks iterations lexically (`-iter10` below `-iter2`) | `TestSyncAuditVerdictLocatorPicksNewestIteration` |
| AC-LVM-008 | a refusal taken after the window was acquired or a record changed | `TestCompleteRefusesStaleAuditHead` |
| AC-LVM-010 | receipt-side ancestor whose diff to HEAD is empty accepted | `TestReceiptAtEmptyDiffAncestorRefused` |
| AC-LVM-010 | receipt `tree_root` compared with the store root instead of the audited tree's path (refuses every orphaned-worktree receipt) | `TestReceiptForConfigOrphanedWorktreeQualifies`, `TestReceiptFromOtherTreeRootRefused` |
| AC-LVM-010 | binding matches a receipt by card id or time instead of the audited commit | `TestReceiptFromOlderHeadRefused`, `TestLegacyReceiptWithoutCommitRefused` |
| AC-LVM-010 | binding applies "every cited receipt" so a verdict with no citation passes vacuously | `TestAuditReceiptLineMissingRefused`, `TestAuditReceiptLineNoneRefused` |
| AC-LVM-009 | the writer stamps the HEAD read at write time instead of the HEAD captured at entry | `TestEntryPointCapturesStartHead` |
| AC-LVM-017 | state read from the primary checkout's working tree | `TestStalePrimaryCheckoutDoesNotAffectGateState` |
| AC-LVM-017 | state read from the card branch's own copy of the key | `TestCardBranchCopyOfKeyIgnored` |
| AC-LVM-017 | the branch read as a bare name (`<branch>:<path>`), so a same-named tag with another value outranks it and turns the gate off | `TestSameNamedTagDoesNotShadowBranch` |
| AC-LVM-012 | the characterization test added in the same commit as the `factoryCompleteCard` edit | the `git diff --name-only <S2>^ <S2> -- internal/cli/factory_complete_gate_off_test.go` leg (non-empty stdout) |
| AC-LVM-017 | unreadable file silently reads off | `TestUnreadableFileTreatedOnWithNotice` |
| AC-LVM-017 | an unconfigured project reads on, or a missing configured branch reads off | `TestNoIntegrationBranchConfiguredReadsOff`, `TestConfiguredButUnresolvableBranchRefusesNamed` |
| AC-LVM-012 | a branch commit enables the key and a later one removes it | the empty `git log <S0>^..<S4> -- workflow.yaml` |

## Quality gates and Definition of Done

- Every AC green with the empty-sweep guard satisfied; the RED-now cells re-run at M0 and committed as `evidence/m0-baseline.md`.
- `go vet` and `gofmt -l` clean on touched packages; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; package-scoped `golangci-lint`; scoped `-run` coverage for the added code with the remainder recorded as a CI-deferred Gap. Go calls are made only while the host admits them (resource rule 8c1e).
- `moai spec lint --strict` on this SPEC exit 0; no `\u` escapes; `make agents-emit` regenerated the Codex agent copy of the sync-auditor protocol and `make build` (which runs the read-only `agents-emit-check`) is green, with `git status --short` clean after (Template-First).
- No skipped test, no widened default permission. The diff of `factory_card.go` leaves steps 3 and 4 (`factory_card.go:2260-2300`) and the merge step untouched: the re-measure record checks are exactly as before (spec C9).
- **Not claimed:** that this card's own landing runs under the gate. Piece A2 lands under the current rules (board `0577` P2); the first gated landing is the next one after a later, separate enabling commit, made only after piece A1 has also landed, with a by-path binary built from a main containing both pieces (spec C5, C10). Nor is it claimed that the gate, once on, checks re-measure completeness: that is SPEC-LANDING-COMPLETENESS-GATE-001.

## Edge cases

- The gate on while no card carries a bound receipt yet (the receipt-writing server not yet running the REQ-LVM-009 build): every `complete` refuses with the unbound-receipt cause. This is fail-closed and is the cost recorded in spec C5/C6, not a wedge: the refusal precedes the window, and re-auditing with a conforming server produces the receipt.
- A SPEC card whose only audit file is `verdict.md`: refused (the SPEC locator reads `sync-audit*.md`).
- A re-absorb or any commit after the audit: the verdict no longer equals HEAD; the card takes the HEAD-confirmation audit again (board `2c51`).
- Recovery if the gate refuses its own reversal: the key-off one-path card lands through the gated path (its audit is bound; an unreadable state reads on but admits a compliant landing). If the gated path itself is the defect, the operator-held manual edit of the integration branch's configuration is the accepted residual, with the operator as owner and not performed by an agent.
- Measured transition fact: of the 8 sampled merge-ready SPEC cards, 2 carry a top-level `audit_receipt:` line in their newest sync-audit verdict; the others cite receipts in prose and are re-audited at landing anyway (board `2c51`), producing the machine line.

---

## Evidence ledger (RED-now observations)

Entries E-007a, E-007b and E-017b were observed 2026-10-19 at HEAD `49df956ac8`, and E-017a (now 12 names) at HEAD `edee2983cc`; the other Go-name entries were observed at the pin and are carried unchanged because no file under `internal/`, `.claude/`, `cmd/` or `.codex/` differs between the pin and `edee2983cc` (`git diff --name-only` printed nothing). The non-Go entries N-001 to N-023 follow the Go entries; they were observed 2026-10-19 at HEAD `edee2983cc`, and the paths they read are unchanged since the pin. All entries: tree `8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2`, unpiped (output redirected to a file and printed verbatim). `names_found` counts stdout lines of the form `func Test…(`.

#### E-007a
```
command: grep -rhoE "^func (TestLandingGateDefaultsOff)\(" internal/config --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 1)
```
#### E-007b
```
command: grep -rhoE "^func (TestCompleteGateOffMatchesTodaysBehaviour)\(" internal/cli --include='*_test.go'
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
#### E-017a
```
command: grep -rhoE "^func (TestSameNamedTagDoesNotShadowBranch|TestGateStateReadFromIntegrationBranchBlob|TestEnableCommitOnIntegrationBranchTakesEffect|TestRevertCommitOnIntegrationBranchTakesEffect|TestStalePrimaryCheckoutDoesNotAffectGateState|TestCardBranchCopyOfKeyIgnored|TestAbsentKeyReadsOff|TestNoIntegrationBranchConfiguredReadsOff|TestConfiguredButUnresolvableBranchRefusesNamed|TestUnreadableFileTreatedOnWithNotice|TestUnreadableFileStillAdmitsCompliantLanding|TestAbsentPathVersusFailedReadDiscriminated)\(" internal/factory --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 12)
```
#### E-017b
```
command: grep -rhoE "^func (TestCompleteReadsGateFromIntegrationBranch|TestIntegrationStatusPrintsGateStateAndSource)\(" internal/cli --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 2)
```
#### E-018
```
command: grep -rhoE "^func (TestAuditReceiptLineParsesCommaList|TestAuditReceiptLineMissingRefused|TestAuditReceiptLineNoneRefused|TestAuditReceiptLineEmptyRefused|TestAuditReceiptLineMalformedIdRefused|TestAuditReceiptLineDuplicateRefused|TestCitedIdAbsentFromStoreRefused|TestSPECLessCardCitationLineRequired)\(" internal/factorylane --include='*_test.go'
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout:
names_found: 0 (required: 8)
```

#### N-001
```
command: grep -c -F -e "committed content of the integration branch" .claude/rules/moai/workflow/factory-dispatch-mechanics.md
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout: 0
```
#### N-002
```
command: grep -c -F -e "gate_state_unreadable" .claude/rules/moai/workflow/factory-dispatch-mechanics.md
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout: 0
```
#### N-003
```
command: grep -c -F -e "built from the landing HEAD" .claude/rules/moai/workflow/factory-dispatch-mechanics.md
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout: 0
```
#### N-004
```
command: grep -c -F -e "before the final audit" .claude/rules/moai/workflow/factory-dispatch-mechanics.md
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout: 0
```
#### N-005
```
command: grep -c -F -e "committed content of the integration branch" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout: 0
```
#### N-006
```
command: grep -c -F -e "gate_state_unreadable" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout: 0
```
#### N-007
```
command: grep -c -F -e "built from the landing HEAD" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout: 0
```
#### N-008
```
command: grep -c -F -e "before the final audit" internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout: 0
```
#### N-009
```
command: grep -c -F -e "audit_receipt machine line" .claude/agents/moai/sync-auditor.md
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout: 0
```
#### N-010
```
command: grep -c -F -e "audit_receipt machine line" internal/template/templates/.claude/agents/moai/sync-auditor.md
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout: 0
```
#### N-011
```
command: grep -c -F -e "audit_receipt machine line" .moai/docs/audit-artifact-convention.md
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout: 0
```
#### N-012
```
command: grep -c -F -e "audit_receipt machine line" .codex/agents/moai/sync-auditor.toml
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout: 0
```
#### N-013
```
command: grep -c -F -e "audit_receipt machine line" internal/template/templates/.codex/agents/moai/sync-auditor.toml
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout: 0
```
#### N-014
```
command: git grep -n -F -e landing_gate HEAD -- .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout: 
```
#### N-015
```
command: git log --format=%H -1 --grep='^test(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M0 '
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 0
stdout: 
```
#### N-016
```
command: git log --format=%H -1 --grep='^feat(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M1 '
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 0
stdout: 
```
#### N-017
```
command: git ls-files --error-unmatch .moai/specs/SPEC-LANDING-VERIFY-MERGE-SPLIT-001/evidence/m0-baseline.md
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout: 
```
#### N-018
```
command: git ls-files --error-unmatch internal/cli/factory_complete_gate_off_test.go
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout: 
```
#### N-019
```
command: git merge-base --is-ancestor HEAD HEAD
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 0
stdout: 
```
#### N-020
```
command: git merge-base --is-ancestor HEAD 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 1
stdout: 
```
#### N-021
```
command: cmp .claude/rules/moai/workflow/factory-dispatch-mechanics.md internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 0
stdout: 
```
#### N-022
```
command: cmp .claude/agents/moai/sync-auditor.md internal/template/templates/.claude/agents/moai/sync-auditor.md
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 0
stdout: 
```
#### N-023
```
command: cmp .codex/agents/moai/sync-auditor.toml internal/template/templates/.codex/agents/moai/sync-auditor.toml
tree: 8a3014bee99ed50cbd6c2cc833ee40e1e1e9f9b2
exit: 0
stdout: 
```

N-019 reads `git merge-base --is-ancestor HEAD HEAD` (self → 0) and N-020 the reverse polarity (HEAD is a descendant of the pin, so asking whether HEAD is an ancestor of the pin → 1); N-014 to N-018 read the committed history of `edee2983cc` and are history-dependent, so they were taken at that HEAD and stay valid at the pin only for the paths they name.
