---
id: SPEC-FACTORY-RECORD-REPAIR-001
version: "0.1.1"
updated: 2026-10-10
---

# Acceptance criteria: SPEC-FACTORY-RECORD-REPAIR-001

These criteria verify the requirements in `spec.md` (REQ-FRR-001 to REQ-FRR-015). Each is a Given-When-Then scenario. The requirement text itself lives in `spec.md`.

## Adoption status

This repair pass runs no tests. No RED cell is observed in this document, and no criterion is recorded as a pass. The one command result listed (AC-FRR-016) is carried over from the plan audit at tree 35d242f344e1bd68a34c47591f5be4f3c017a699; it is not a measurement of this pass, and pre-flight item 2 re-measures it.

**Classification rule (lane ruling).** A criterion whose observed RED needs a test that does not exist yet cannot be observed on the unchanged tree. Under `verification-completeness.md` §2.1, the undecidable disposition, such a criterion loses release-blocking eligibility, is classified regression-guard at plan time, and is not recorded as a pass. The release gate moves to the run phase: a separate RED commit, made before the fix commit, must carry an observed EXPECTED_RED whose failure is an assertion failure. A compile or discovery failure is TOOL_FAILURE and is never written as RED (`.claude/rules/moai/workflow/tdd-result-contract.md`). The commit graph witnesses the order (`verification-claim-integrity.md` §2.3). The observed output goes in the evidence ledger, `progress.md` §E.2. The disposition is the DEFAULT-APPLIED route recorded in `decision-index.md` as Q1; the leader's ruling d-20261010T010044Z-57f8 is cited as an evidence note only.

**Plan-time classification.** Every criterion is a regression-guard at plan time. Thirteen of them (AC-FRR-001 to AC-FRR-012 and AC-FRR-014) gate the release only through their run-phase RED commit. AC-FRR-013 and AC-FRR-015 are green before and after by design. AC-FRR-016 is green on the unchanged tree, so the RED direction of `verification-completeness.md` §2 is vacuous for it. Each criterion states its classification in its own entry.

**Verification convention.** Each Verification line runs one named test with `-v` and `-count=1`. A check passes only when the output contains `=== RUN <Name>` and `--- PASS: <Name>` for that test, with an executed count above zero. `[no tests to run]`, a zero executed count, or a missing line is a FAIL.

**Surface rule for run-phase REDs.** A test for a surface that does not exist yet reaches it through the public command or tool surface (the command tree or the MCP tool list), so the test compiles on the unchanged tree and fails on an assertion. A test that names a symbol which does not exist is TOOL_FAILURE, not RED.

## Acceptance criteria

### AC-FRR-001: Renewal extends a live lease (REQ-FRR-001)

Given card X is leased to lane L and L's lease has exactly 5 minutes left,
When L runs the factory renew verb for X,
Then the lease expiry becomes the request time plus the configured duration, the verb prints that expiry, and the card's version is unchanged.

- Classification: regression-guard at plan time; the run-phase RED commit gates release (see Adoption status).
- Verification: `go test ./internal/cli/ -run '^TestFactoryRenewExtendsLiveLease$' -v -count=1`; passes only with `=== RUN TestFactoryRenewExtendsLiveLease` and `--- PASS: TestFactoryRenewExtendsLiveLease`.
- Predicted RED (not observed): the renew verb is not registered, so the test fails on its expiry assertion once it drives the command tree. A compile failure would be TOOL_FAILURE.
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M2.

### AC-FRR-002: The MCP tool matches the verb (REQ-FRR-002)

Given card X is leased to lane L,
When L renews X through the MCP tool, and then, on the same state, through the factory renew verb,
Then both return the same expiry fields, and for an expired lease both return the same refusal class.

- Classification: regression-guard at plan time; the run-phase RED commit gates release (see Adoption status).
- Verification: `go test ./internal/cli/ -run '^TestFactoryRenewMatchesMCP$' -v -count=1`; passes only with `=== RUN TestFactoryRenewMatchesMCP` and `--- PASS: TestFactoryRenewMatchesMCP`.
- Predicted RED (not observed): no MCP renew tool is declared, so the MCP leg fails on its result assertion.
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M2.

### AC-FRR-003: The lease duration is configurable, with a 15-minute default (REQ-FRR-003)

Given no lease-duration key is configured,
When a lease is acquired on card X,
Then its expiry is the acquisition time plus 15 minutes.
And given the key is set to 20m, when a lease is acquired on card X, then its expiry is the acquisition time plus 20 minutes.

- Classification: regression-guard at plan time; the run-phase RED commit gates release (see Adoption status).
- Verification: `go test ./internal/config/ -run '^TestFactoryLeaseDurationDefault$' -v -count=1`; `go test ./internal/homestate/ -run '^TestCardLeaseUsesConfiguredDuration$' -v -count=1`. Each passes only with its own `=== RUN` and `--- PASS` lines.
- Predicted RED (not observed): the key and the configured field do not exist, so the default assertion, which reads the key, and the override assertion both fail.
- Stops firing: if either test is renamed, moved, or excluded, its ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M1.

### AC-FRR-004: An expired renewal is refused with no write (REQ-FRR-004)

Given card X's lease is held by L and expired two minutes ago,
When L runs the factory renew verb for X,
Then the request is refused, and the card's state, version, and event count are unchanged.

- Classification: regression-guard at plan time; the run-phase RED commit gates release (see Adoption status).
- Verification: `go test ./internal/homestate/ -run '^TestRenewRefusesExpiredWithoutWrite$' -v -count=1`; passes only with `=== RUN TestRenewRefusesExpiredWithoutWrite` and `--- PASS: TestRenewRefusesExpiredWithoutWrite`.
- Predicted RED (not observed): `RenewLease` reaps the expired lease before it checks the holder (`internal/homestate/card_transition.go:452-457`), which bumps the version and appends `lease.expired`, so the unchanged-version assertion fails. At HEAD 35d242f34 the command the audit ran for this name printed `[no tests to run]`, because no test of this name exists yet (plan-audit-iter1 E4).
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M2.

### AC-FRR-005: The watchdog renews before it measures (REQ-FRR-005)

Given the lane stall watchdog skill is read for an awaken pass,
When its procedure and allowed tools are read,
Then the first procedural step renews the lane's own card lease, the renew command is among the allowed tools, and the measurement step follows the renewal.

- Classification: regression-guard at plan time; the run-phase RED commit gates release (see Adoption status).
- Verification: `go test ./internal/template/ -run '^TestLaneWatchdogRenewsFirst$' -v -count=1`; passes only with `=== RUN TestLaneWatchdogRenewsFirst` and `--- PASS: TestLaneWatchdogRenewsFirst`.
- Predicted RED (not observed): the current procedure opens with measurement (§1) and lists no renew command (`allowed-tools`, SKILL.md line 17).
- Limit: this checks procedure text, not runtime behaviour, so a text-only change passes it (spec.md Residual risk).
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M6.

### AC-FRR-006: One-shot recovery through the stage verb and the parent-checkout next verb (REQ-FRR-006)

Given card X is at run, leased to L, and L's lease expired three minutes ago,
When L runs the factory stage verb to run for X,
Then one invocation succeeds, the card's events read the expiry followed by the transition, and the card ends leased to L at stage run.
And given the same starting state, when L runs the factory next verb with the card id from the parent checkout, then one invocation succeeds with the same end state.
And given the requested edge is illegal from the card's stage, when L makes the request, then the request is refused and the expired row is unchanged.

- Classification: regression-guard at plan time; the run-phase RED commit gates release (see Adoption status).
- Verification: `go test ./internal/cli/ -run '^TestStageRecoversOwnExpiredLease$' -v -count=1`; `go test ./internal/cli/ -run '^TestNextRecoversOwnExpiredLeaseFromParentCheckout$' -v -count=1`; `go test ./internal/homestate/ -run '^TestRecoveryAllOrNothing$' -v -count=1`. Each passes only with its own `=== RUN` and `--- PASS` lines.
- Predicted RED (not observed): the transition returns `ErrLeaseExpired` after reaping (`internal/homestate/card_transition.go:315`), and the next verb refuses as in flight (`internal/cli/factory_card.go:1125-1128`).
- Stops firing: if any of the three tests is renamed, moved, or excluded, its ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M3.

### AC-FRR-007: The expired refusal names the recovery command (REQ-FRR-007)

Given card X's lease is held by L and has expired,
When a caller other than L requests a transition on X, and separately L's own renewal of X is refused as expired,
Then each refusal text names the factory stage verb that L runs to re-lease X.

- Classification: regression-guard at plan time; the run-phase RED commit gates release (see Adoption status).
- Verification: `go test ./internal/homestate/ -run '^TestExpiredRefusalNamesRecoveryVerb$' -v -count=1`; passes only with `=== RUN TestExpiredRefusalNamesRecoveryVerb` and `--- PASS: TestExpiredRefusalNamesRecoveryVerb`.
- Predicted RED (not observed): the refusal texts name no verb (`internal/homestate/card_transition.go:315` and `:456`).
- Sequence on this tree before the repair: the first transition request reaps the expired lease, then `moai factory next --card <card-id>` (or the MCP factory_next with project_root set to the parent checkout) re-leases it, then a same-stage return follows. REQ-FRR-006 replaces that sequence with one step. The named command is the stage verb because a holder inside its card worktree cannot run the next verb (`internal/cli/factory_card.go:189-201`).
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M3.

### AC-FRR-008: A sync-audit FAIL on the card's own code returns the card to run (REQ-FRR-008)

Given card X is at sync-audit, leased to L with sync evidence hash H, and the leader has recorded a decision of kind ruling with scope card:X that names the sync-audit verdict file path and an audited SHA equal to H,
When L requests the sync-audit to run transition and cites that ruling,
Then the card moves to run with stage run, the lease stays with L, and the evidence record is kept.

- Classification: regression-guard at plan time; the run-phase RED commit gates release (see Adoption status).
- Verification: `go test ./internal/homestate/ -run '^TestSyncAuditFailOwnScopeReturnsToRun$' -v -count=1`; passes only when the output contains `=== RUN TestSyncAuditFailOwnScopeReturnsToRun` and `--- PASS: TestSyncAuditFailOwnScopeReturnsToRun` with an executed count above zero. `[no tests to run]` is FAIL.
- Predicted RED (not observed): no sync-audit to run row exists (`internal/homestate/card_transition.go:133-181`), so the request fails as an illegal transition at its admitted-transition assertion; that is an assertion failure, not a compile failure.
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M4.

### AC-FRR-009: A sync-audit to run request without a matching ruling is refused (REQ-FRR-009)

Given card X is at sync-audit, leased to L, with sync evidence hash H,
When L requests the sync-audit to run transition citing (a) no ruling, (b) a ruling scoped to another card Y, or (c) a ruling whose audited SHA differs from H,
Then each request is refused with the reason token `ruling-missing`, `ruling-scope-mismatch`, or `ruling-sha-mismatch` respectively, and the card's state, version, and event history are unchanged.

- Classification: regression-guard at plan time; the run-phase RED commit gates release (see Adoption status).
- Verification: `go test ./internal/homestate/ -run '^TestSyncAuditRefusesRunWithoutMatchingRuling$' -v -count=1`; passes only when the output contains `=== RUN TestSyncAuditRefusesRunWithoutMatchingRuling` and `--- PASS: TestSyncAuditRefusesRunWithoutMatchingRuling`, with a `--- PASS` line for each subtest (`ruling-missing`, `ruling-scope-mismatch`, `ruling-sha-mismatch`). `[no tests to run]` or a zero executed count is FAIL.
- Predicted RED (not observed): on the unchanged tree each request is refused as an illegal transition, because no edge exists, and carries none of the three reason tokens, so the reason-token assertion fails. A bare refusal check would be green for the wrong reason and does not qualify.
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M4.

### AC-FRR-010: A holder's re-stamp keeps the stage and the lease (REQ-FRR-010)

Given card X is at plan-audit, leased to L, with a stale evidence hash, and L holds a verdict whose audited SHA is the new evidence commit,
When L requests the re-audit entry with the new hash,
Then the card stays at plan-audit under L's lease, the evidence hash equals the new SHA, and the event history keeps every prior event and adds one.
And the same holds at sync-audit.
And given the verdict's audited SHA differs from the new hash, when L requests the re-stamp, then the request is refused and the card is unchanged.

- Classification: regression-guard at plan time; the run-phase RED commit gates release (see Adoption status).
- Verification: `go test ./internal/homestate/ -run '^TestHolderRestampsEvidenceAtPlanAudit$' -v -count=1`; `go test ./internal/homestate/ -run '^TestHolderRestampsEvidenceAtSyncAudit$' -v -count=1`; `go test ./internal/homestate/ -run '^TestHolderRestampRefusesMismatchedAuditedSHA$' -v -count=1`. Each passes only with its own `=== RUN` and `--- PASS` lines.
- Predicted RED (not observed): no same-stage audit row exists (`internal/homestate/card_transition.go:133-181`), so the two accepting tests fail as illegal transitions. The mismatch test pins the evidence-mismatch refusal class, so the missing-row refusal cannot satisfy it.
- Stops firing: if any of the three tests is renamed, moved, or excluded, its ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M4.

### AC-FRR-011: The lease records the worktree path (REQ-FRR-011)

Given card X is assigned to lane L, no worktree path is recorded for X, and the factory materializes the card worktree for X,
When L acquires the lease on X,
Then the card records that worktree path in the same acquisition, and no lease is held on X without it.
And given a different worktree path is already recorded for X, when L acquires the lease, then the acquisition is refused and the card is unchanged.

- Classification: regression-guard at plan time; the run-phase RED commit gates release (see Adoption status).
- Verification: `go test ./internal/homestate/ -run '^TestLeaseRecordsWorktreePath$' -v -count=1`; passes only with `=== RUN TestLeaseRecordsWorktreePath` and `--- PASS: TestLeaseRecordsWorktreePath`.
- Predicted RED (not observed): the lease write sets the holder, heartbeat, and expiry only (`internal/homestate/card_transition.go:540-545`), so the recorded path stays empty and the first assertion fails.
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M1.

### AC-FRR-012: The interim guard refuses an empty-stage run for a SPEC-bearing card (REQ-FRR-012)

Given card X has an empty stage, no admitted plan verdict, and a non-empty SPEC id, valid or invalid, where valid means the id matches `specIDPattern` (`internal/homestate/card_picked.go:15`),
When X is moved to run,
Then the transition is refused and the row is unchanged.
And given card X also has a recorded operator decision, when X is moved to run without an admitted plan verdict, then the transition is refused, because a recorded operator decision alone does not admit it.
And given an admitted plan verdict is recorded for a card with a valid SPEC id and an empty stage, when that card is moved to run, then the transition is admitted.

- Classification: regression-guard at plan time; the run-phase RED commit gates release (see Adoption status).
- Verification: `go test ./internal/homestate/ -run '^TestEmptyStageToRunRefusedWithSpecID$' -v -count=1`; `go test ./internal/homestate/ -run '^TestEmptyStageToRunAdmittedWithPlanVerdict$' -v -count=1`. Each passes only with its own `=== RUN` and `--- PASS` lines.
- Predicted RED (not observed): `stageAdmits` admits an empty-stage run unconditionally (`internal/homestate/card_transition.go:752-753`), so the refused-side assertions fail. The admitted-side test is predicted green on the unchanged tree and is adopted only together with the refused side.
- Stops firing: if either test is renamed, moved, or excluded, its ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M5.

### AC-FRR-013: Unchanged routes stay unchanged (REQ-FRR-013)

Given card Y has an empty SPEC id and an empty stage,
When Y is moved to run,
Then the transition is admitted as before.
The sync-audit route named in REQ-FRR-013 is verified by AC-FRR-008 and AC-FRR-009, whose refusal reasons are the REQ-FRR-009 tokens and not the interim guard's.

- Classification: regression-guard (green before and after by design; not a RED gate). The guard test is written and observed green before M5 starts.
- Verification: `go test ./internal/homestate/ -run '^TestEmptyStageToRunUnchangedWithoutSpecID$' -v -count=1`; passes only with `=== RUN TestEmptyStageToRunUnchangedWithoutSpecID` and `--- PASS: TestEmptyStageToRunUnchangedWithoutSpecID`.
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M5 keeps it green.

### AC-FRR-014: A non-holder can neither renew nor reap an expired lease, and cannot renew a live one (REQ-FRR-014)

Given card X's lease is held by L,
When a different label M runs the factory renew verb for X, or calls the MCP renew tool for X, while the lease is live, and separately runs the factory stage verb for X after the lease has expired,
Then each request is refused, and the card's state, version, and events are unchanged.

- Classification: regression-guard at plan time; the run-phase RED commit gates release (see Adoption status).
- Verification: `go test ./internal/homestate/ -run '^TestNonHolderCannotReapOrRenew$' -v -count=1`; `go test ./internal/cli/ -run '^TestNonHolderRenewRefusedCLIAndMCP$' -v -count=1`. Each passes only with its own `=== RUN` and `--- PASS` lines.
- Predicted RED (not observed): on the expired-lease case, the expiry branch reaps before the holder check (`internal/homestate/card_transition.go:311-316` and `:452-457`), so the version and events change. On the live-lease case, a missing renew verb fails the request with a different error; the test pins the holder-refusal class (`ErrLeaseHolder`), so a missing verb cannot satisfy it.
- Stops firing: if either test is renamed, moved, or excluded, its ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M2 for the renewal; M3 for the transition.

### AC-FRR-015: The queue claim lease is unchanged (REQ-FRR-015)

Given the card lease duration is configured to 20 minutes, set on the card lease duration variable `FactoryLeaseDuration` in `internal/homestate` that the card lease reads,
When a queue claim is made,
Then its lease expiry stays 15 minutes after the claim, as the existing claim-lease tests assert.

- Classification: regression-guard (green before and after by design; not a RED gate).
- Verification: `go test ./internal/factory/ -run '^TestBacklogClaim_ReclaimsExpired$' -v -count=1`; `go test ./internal/factory/ -run '^TestBacklogClaim_RenewAfterExpiry$' -v -count=1`; `go test ./internal/factory/ -run '^TestBacklogClaim_QueueLeaseIgnoresCardLeaseDuration$' -v -count=1`. Each passes only with its own `=== RUN` and `--- PASS` lines.
- The first two tests exist, and the first asserts a duration of about 15 minutes (`internal/factory/backlog_claim_test.go:154-155`). The third is new: it sets the card lease duration to 20 minutes, claims, and asserts the 15-minute queue expiry. It follows the `TestBacklogClaim_<Descriptor>` naming of that file. It is the mutant check: a queue claim that followed the card duration would fail it.
- Stops firing: if the new test or either existing test is renamed, moved, or excluded, its ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M1 keeps it green.

### AC-FRR-016: The transition-table edge count is pinned (count-pin check; no requirement link)

Given the transition table as it stands after this SPEC's rows are added,
When the accepted-pair test runs,
Then the pinned count equals the number of distinct accepted (from, to) pairs, and the `@MX:ANCHOR` reason at `internal/homestate/card_transition.go:130` states the same count.

- Classification: regression-guard. Its RED-now cell is green on the unchanged tree, the vacuous direction of `verification-completeness.md` §2, so no RED is available for it at plan time.
- Requirement link: none. This is a count-pin check on the transition table. It is not evidence for REQ-FRR-008 or REQ-FRR-010.
- RED-now cell, recorded from plan-audit-iter1 and not re-run in this pass:
  - Command: `go test ./internal/homestate/ -run TestFR_AC005_TransitionTableEdgeCount -count=1 -v`
  - Verbatim stdout, as quoted in plan-audit-iter1.md E5 (the raw capture is taken at pre-flight item 2):

    ```text
    === RUN   TestFR_AC005_TransitionTableEdgeCount
    fr_transition_test.go:165: requested pairs: 71 accepted, 370 refused, 441 total; production table rows: 73
    --- PASS: TestFR_AC005_TransitionTableEdgeCount (1.70s)
    PASS
    ok  	github.com/modu-ai/moai-adk/internal/homestate	1.913s
    ```

  - Exit code: 0
  - Tree SHA: 35d242f344e1bd68a34c47591f5be4f3c017a699, the audited HEAD. Its code tree is identical to 2aab5f797 (plan-audit-iter1 E2).
  - Gap: this pass ran no test. The observation is carried over from the audit, and the run phase re-observes it.
- Verification: `go test ./internal/homestate/ -run '^TestFR_AC005_TransitionTableEdgeCount$' -v -count=1`; passes only with `=== RUN TestFR_AC005_TransitionTableEdgeCount` and `--- PASS: TestFR_AC005_TransitionTableEdgeCount`.
- Two numbers, two quantities: 71 is the number of distinct accepted (from, to) pairs, which the test pins (`internal/homestate/fr_transition_test.go:146-166`). 73 is `len(TransitionEdges())`, the number of table entries. That count includes the entries the builder generates per state (`internal/homestate/card_transition.go:160-178`), so it is not a count of pairs. Only the 71 is pinned.
- Predicted after the rows are added: the pin moves from 71 to 74, which is 71 plus the three new pairs named in plan M4 (sync audit to run, plan audit to plan audit, and sync audit to sync audit). The lane re-measures the count.
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M4 moves the pin to the re-measured count.

## Edge cases

- Expiry exactly at the request instant follows the existing rule in `LeaseExpired` (`card_record.go:146`). The RED test pins that rule; the criteria do not restate it.
- Expiry of a lease on a merging card returns the card to blocked, not to assigned, so recovery does not apply (plan M3, item 2). This is covered by the M3 test set and is not a separate criterion.
- A non-positive or unparseable lease duration is expected to be refused at config load, naming the key. This expectation is not yet a criterion and is to be confirmed with the lane (decision-index, Optional, deferred).
- A holder whose worker row is missing follows the existing registration rule. AC-FRR-014 covers the non-holder side.
- A sync-audit FAIL whose defects lie outside the card's scope has no ruling scoped to that card, so AC-FRR-009 refuses the transition.
- A card whose SPEC id is empty while leased is covered by AC-FRR-013 (unchanged), and the residual is stated in spec.md.

## Quality gates (not acceptance criteria)

- QG-1 Format and lint: `gofmt` and the project linter run clean on every touched Go package. The lane records the verbatim output.
- QG-2 Race: `go test -race ./internal/homestate/` after the transaction changes, with the verbatim output recorded.
- QG-3 Coverage: `go test -cover` on each touched package, with output recorded. The TRUST 5 target is 85%. A shortfall is recorded as a Gap, not as a pass.
- QG-4 Template-First, scoped to what this SPEC changes: after `make build`, each file this SPEC changes matches its template. The files are the lease-duration key in `.moai/config/sections/workflow.yaml` and its template, the watchdog `SKILL.md` in `.claude/` and in the template tree, and the catalogue and `factory-dispatch-mechanics.md` mirrors. The comparison is key-scoped for `workflow.yaml`, because the rendered file differs from the template on purpose. The repository-local overrides that must survive a template sync are, in `.moai/config/sections/workflow.yaml`: `branch_guard.enabled: true` (template `false`), `branch_guard.deny_commits_on: [main]` (template `[]`), and `commit_identity_guard.enabled: true` (template `false`). A sync that overwrote them would remove the protection on `main`. The lane records the verbatim comparison output.
- QG-5 Verdict parser untouched: `git diff --name-only 2aab5f797..HEAD` lists no file under `internal/auditverdict/`. The lane records the resolved HEAD SHA beside the output, because HEAD moves.
- QG-6 Traceability: every commit on the branch names card t1621 in its message.

## Definition of Done

- Run-phase release gate (AC-FRR-001 to AC-FRR-012 and AC-FRR-014): each criterion has a RED commit, made before its fix commit, whose recorded output shows EXPECTED_RED as an assertion failure. Its green output is recorded in `progress.md` §E.2 after the fix commit.
- Regression guards (AC-FRR-013, AC-FRR-015, and AC-FRR-016): green output recorded in `progress.md` §E.2. The AC-FRR-013 guard test is observed green before M5 starts.
- Each Verification line of AC-FRR-001 to AC-FRR-016 shows `=== RUN` and `--- PASS` for its named test, with an executed count above zero.
- QG-1 to QG-6 pass, or each shortfall is recorded as a Gap.
- Decisions are recorded in `decision-index.md` before the plan re-audit runs. A decision surfaced later is recorded before the Phase 1 Plan Audit Gate runs at the start of run, because that gate reads the index then.
- The plan re-audit runs on the session model, with codex as the required backend.
- Local landing only: no push, no pull request, and no remote done edge.
