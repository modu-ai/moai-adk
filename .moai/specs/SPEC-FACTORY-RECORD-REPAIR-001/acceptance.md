---
id: SPEC-FACTORY-RECORD-REPAIR-001
version: "0.2.0"
updated: 2026-10-10
---

# Acceptance criteria: SPEC-FACTORY-RECORD-REPAIR-001

These criteria verify the active requirements in `spec.md` (REQ-FRR-001 to REQ-FRR-007, REQ-FRR-011, REQ-FRR-014, REQ-FRR-015, and REQ-FRR-016 to REQ-FRR-019). Each is a Given-When-Then scenario. The requirement text itself lives in `spec.md`. The active criteria are AC-FRR-001 to AC-FRR-007, AC-FRR-011, AC-FRR-014, AC-FRR-015, and AC-FRR-017 to AC-FRR-021 (AC-FRR-017 to -021 are item 3, HISTORY 0.3.0). The identifiers AC-FRR-008 to -010, -012, -013, and -016 are retired and are not reused (`spec.md` HISTORY 0.2.0).

## Adoption status

This revision runs no tests, and no criterion is recorded as a pass. Every active criterion is a regression guard (decision-index Q1, operator verdict `d-20261010T053034Z-1a19`). None gates the release. The run phase records the output of each Verification line in `progress.md` §E.2. Where a named test does not yet exist, the run phase writes it. A compile or discovery failure is TOOL_FAILURE and is never recorded as RED (`.claude/rules/moai/workflow/tdd-result-contract.md`).

**Verification convention.** Each Verification line runs one named test with `-v` and `-count=1`. A check passes only when the output contains `=== RUN <Name>` and `--- PASS: <Name>` for that test, with an executed count above zero. `[no tests to run]`, a zero executed count, or a missing line is a FAIL.

**Surface rule for predicted failures.** A test for a surface that does not exist yet reaches it through the public command or tool surface (the command tree or the MCP tool list), so the test compiles on the unchanged tree and fails on an assertion. A test that names a symbol which does not exist is TOOL_FAILURE, not a failure to predict.

## Acceptance criteria

### AC-FRR-001: Renewal extends a live lease (REQ-FRR-001)

Given card X is leased to lane L and L's lease has exactly 5 minutes left,
When L runs the factory renew verb for X,
Then the lease expiry becomes the request time plus the configured duration, the verb prints that expiry, and the card's version is unchanged.

- Classification: regression guard (decision-index Q1, operator verdict `d-20261010T053034Z-1a19`). Not a release gate.
- Verification: `go test ./internal/cli/ -run '^TestFactoryRenewExtendsLiveLease$' -v -count=1`; passes only with `=== RUN TestFactoryRenewExtendsLiveLease` and `--- PASS: TestFactoryRenewExtendsLiveLease`.
- Predicted failure (not observed): the renew verb is not registered, so the test fails on its expiry assertion once it drives the command tree. A compile failure would be TOOL_FAILURE.
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M2.

### AC-FRR-002: The MCP tool matches the verb (REQ-FRR-002)

Given card X is leased to lane L,
When L renews X through the MCP tool, and then, on the same state, through the factory renew verb,
Then both return the same expiry fields, and for an expired lease both return the same refusal class.

- Classification: regression guard (decision-index Q1, operator verdict `d-20261010T053034Z-1a19`). Not a release gate.
- Verification: `go test ./internal/cli/ -run '^TestFactoryRenewMatchesMCP$' -v -count=1`; passes only with `=== RUN TestFactoryRenewMatchesMCP` and `--- PASS: TestFactoryRenewMatchesMCP`.
- Predicted failure (not observed): no MCP renew tool is declared, so the MCP leg fails on its result assertion.
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M2.

### AC-FRR-003: The lease duration is configurable, with a 15-minute default (REQ-FRR-003)

Given no lease-duration key is configured,
When a lease is acquired on card X,
Then its expiry is the acquisition time plus 15 minutes.
And given the key is set to 20m, when a lease is acquired on card X, then its expiry is the acquisition time plus 20 minutes.

- Classification: regression guard (decision-index Q1, operator verdict `d-20261010T053034Z-1a19`; key name per Q18). Not a release gate.
- Verification: `go test ./internal/config/ -run '^TestFactoryLeaseDurationDefault$' -v -count=1`; `go test ./internal/homestate/ -run '^TestCardLeaseUsesConfiguredDuration$' -v -count=1`. Each passes only with its own `=== RUN` and `--- PASS` lines.
- Predicted failure (not observed): the key and the configured field do not exist, so the default assertion, which reads the key, and the override assertion both fail.
- Stops firing: if either test is renamed, moved, or excluded, its ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M1.

### AC-FRR-004: An expired renewal is refused with no write (REQ-FRR-004)

Given card X's lease is held by L and expired two minutes ago,
When L runs the factory renew verb for X,
Then the request is refused, and the card's state, version, and event count are unchanged.

- Classification: regression guard (decision-index Q1, operator verdict `d-20261010T053034Z-1a19`; refusal per Q9 (a)). Not a release gate.
- Verification: `go test ./internal/homestate/ -run '^TestRenewRefusesExpiredWithoutWrite$' -v -count=1`; passes only with `=== RUN TestRenewRefusesExpiredWithoutWrite` and `--- PASS: TestRenewRefusesExpiredWithoutWrite`.
- Predicted failure (not observed): `RenewLease` reaps the expired lease before it checks the holder (`internal/homestate/card_transition.go:452-457`), which bumps the version and appends `lease.expired`, so the unchanged-version assertion fails. At HEAD 35d242f34 the command the audit ran for this name printed `[no tests to run]`, because no test of this name exists yet (plan-audit-iter1 E4; re-confirmed at HEAD 53ea27f4f, plan-audit-iter2 E7).
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M2.

### AC-FRR-005: The watchdog renews before it measures (REQ-FRR-005)

Given the lane stall watchdog skill is read for an awaken pass,
When its procedure and allowed tools are read,
Then the first procedural step renews the lane's own card lease, the renew command is among the allowed tools, and the measurement step follows the renewal.

- Classification: regression guard (decision-index Q1, operator verdict `d-20261010T053034Z-1a19`). Not a release gate.
- Verification: `go test ./internal/template/ -run '^TestLaneWatchdogRenewsFirst$' -v -count=1`; passes only with `=== RUN TestLaneWatchdogRenewsFirst` and `--- PASS: TestLaneWatchdogRenewsFirst`.
- Predicted failure (not observed): the current procedure opens with measurement (§1) and lists no renew command (`allowed-tools`, SKILL.md line 17).
- Limit: this checks procedure text, not runtime behaviour, so a text-only change passes it (spec.md Residual risk).
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M6.

### AC-FRR-006: One-shot recovery through the stage verb and the parent-checkout next verb (REQ-FRR-006)

Given card X is at run, leased to L, and L's lease expired three minutes ago,
When L runs the factory stage verb to run for X,
Then one invocation succeeds, the card's events read the expiry followed by the transition, and the card ends leased to L at stage run.
And given the same starting state, when L runs the factory next verb with the card id from the parent checkout, then one invocation succeeds with the same end state.
And given the requested edge is illegal from the card's stage, when L makes the request, then the request is refused and the expired row is unchanged.

- Classification: regression guard (decision-index Q1, operator verdict `d-20261010T053034Z-1a19`). Not a release gate.
- Verification: `go test ./internal/cli/ -run '^TestStageRecoversOwnExpiredLease$' -v -count=1`; `go test ./internal/cli/ -run '^TestNextRecoversOwnExpiredLeaseFromParentCheckout$' -v -count=1`; `go test ./internal/homestate/ -run '^TestRecoveryAllOrNothing$' -v -count=1`. Each passes only with its own `=== RUN` and `--- PASS` lines.
- Predicted failure (not observed): the transition returns `ErrLeaseExpired` after reaping (`internal/homestate/card_transition.go:315`), and the next verb refuses as in flight (`internal/cli/factory_card.go:1125-1128`).
- Stops firing: if any of the three tests is renamed, moved, or excluded, its ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M3.

### AC-FRR-007: The expired refusal names the recovery command (REQ-FRR-007)

Given card X's lease is held by L and has expired,
When a caller other than L requests a transition on X, and separately L's own renewal of X is refused as expired,
Then each refusal text names the factory stage verb that L runs to re-lease X.

- Classification: regression guard (decision-index Q1 and Q8, operator verdict `d-20261010T053034Z-1a19`, option (a)). Not a release gate.
- Verification: `go test ./internal/homestate/ -run '^TestExpiredRefusalNamesRecoveryVerb$' -v -count=1`; passes only with `=== RUN TestExpiredRefusalNamesRecoveryVerb` and `--- PASS: TestExpiredRefusalNamesRecoveryVerb`.
- Predicted failure (not observed): the refusal texts name no verb (`internal/homestate/card_transition.go:315` and `:456`).
- Sequence on this tree before the repair: the first transition request reaps the expired lease, then `moai factory next --card <card-id>` (or the MCP factory_next with project_root set to the parent checkout) re-leases it, then a same-stage return follows. REQ-FRR-006 replaces that sequence with one step. The named command is the stage verb because a holder inside its card worktree cannot run the next verb (`internal/cli/factory_card.go:189-201`).
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M3.

### AC-FRR-011: The lease records the worktree path at acquisition (REQ-FRR-011)

Given card X is assigned to lane L, no worktree path is recorded for X, and the factory materializes the card worktree for X,
When L acquires the lease on X,
Then the card records that worktree path in the same transaction that writes the holder, heartbeat, and expiry, and no lease is held on X without it.
And given the materialization of the card worktree for X fails, when L attempts the acquisition, then the acquisition is refused, no lease is held on X, and no path is recorded.
And given a different worktree path is already recorded for X, when L acquires the lease, then the acquisition is refused and the card is unchanged.

- Classification: regression guard (decision-index Q10, operator verdict `d-20261010T053034Z-1a19`, option (ii) at T3; Q1). Not a release gate.
- Verification: `go test ./internal/homestate/ -run '^TestLeaseRecordsWorktreePath$' -v -count=1`; `go test ./internal/cli/ -run '^TestLeaseRefusedOnMaterializationFailure$' -v -count=1`. Each passes only with its own `=== RUN` and `--- PASS` lines.
- Predicted failure (not observed): the lease write sets the holder, heartbeat, and expiry only (`internal/homestate/card_transition.go:540-545`), so the recorded path stays empty and the first assertion fails. On the unchanged tree the lease is acquired before the card worktree is materialized (`internal/cli/factory_card.go:1590-1600`), so the materialization-failure case acquires the lease where it must refuse.
- The re-lease in AC-FRR-006 is an acquisition and must satisfy this criterion as well.
- Stops firing: if either test is renamed, moved, or excluded, its ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M1.

### AC-FRR-014: A non-holder can neither renew nor reap an expired lease, and cannot renew a live one (REQ-FRR-014)

Given card X's lease is held by L,
When a different label M runs the factory renew verb for X, or calls the MCP renew tool for X, while the lease is live, and separately runs the factory stage verb for X after the lease has expired,
Then each request is refused, and the card's state, version, and events are unchanged.

- Classification: regression guard (decision-index Q1, operator verdict `d-20261010T053034Z-1a19`). Not a release gate.
- Verification: `go test ./internal/homestate/ -run '^TestNonHolderCannotReapOrRenew$' -v -count=1`; `go test ./internal/cli/ -run '^TestNonHolderRenewRefusedCLIAndMCP$' -v -count=1`. Each passes only with its own `=== RUN` and `--- PASS` lines.
- Predicted failure (not observed): on the expired-lease case, the expiry branch reaps before the holder check (`internal/homestate/card_transition.go:311-316` and `:452-457`), so the version and events change. On the live-lease case, a missing renew verb fails the request with a different error; the test pins the holder-refusal class (`ErrLeaseHolder`), so a missing verb cannot satisfy it.
- Stops firing: if either test is renamed, moved, or excluded, its ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M2 for the renewal; M3 for the transition.

### AC-FRR-015: The queue claim lease is unchanged (REQ-FRR-015)

Given the card lease duration is configured to 20 minutes, set on the card lease duration variable `FactoryLeaseDuration` in `internal/homestate` that the card lease reads,
When a queue claim is made,
Then its lease expiry stays 15 minutes after the claim, as the existing claim-lease tests assert.

- Classification: regression guard (green before and after by design; not a failure to predict).
- Verification: `go test ./internal/factory/ -run '^TestBacklogClaim_ReclaimsExpired$' -v -count=1`; `go test ./internal/factory/ -run '^TestBacklogClaim_RenewAfterExpiry$' -v -count=1`; `go test ./internal/factory/ -run '^TestBacklogClaim_QueueLeaseIgnoresCardLeaseDuration$' -v -count=1`. Each passes only with its own `=== RUN` and `--- PASS` lines.
- The first two tests exist, and the first asserts a duration of about 15 minutes (`internal/factory/backlog_claim_test.go:154-155`). The third is new: it sets the card lease duration to 20 minutes, claims, and asserts the 15-minute queue expiry. It follows the `TestBacklogClaim_<Descriptor>` naming of that file. It is the mutant check: a queue claim that followed the card duration would fail it.
- Stops firing: if the new test or either existing test is renamed, moved, or excluded, its ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M1 keeps it green.

### AC-FRR-017: A pushed card reaches done through the receipt gate (REQ-FRR-016)

Given card X is in state `pushed` in a repository with a remote, its merge commit is contained in the remote-tracking ref of the integration branch, and a leader approval receipt for X verifies against X's current version, evidence SHA, and owner label,
When the human decider requests the `done` transition for X with X's backlog identity,
Then the request is accepted, X is in state `done` at the next version, and X's event history gains one move from `pushed` to `done`.

- Classification: regression guard, following the classification of Q1 (operator verdict `d-20261010T053034Z-1a19`). Not a release gate.
- Verification: `go test ./internal/homestate/ -run '^TestPushedToDoneOpensWithReceipt$' -v -count=1`; passes only with `=== RUN TestPushedToDoneOpensWithReceipt` and `--- PASS: TestPushedToDoneOpensWithReceipt`.
- Predicted failure (not observed): `buildTransitionTable` has no edge from `pushed` to `done` (`internal/homestate/card_transition.go:133-181`), so the request returns `ErrIllegalTransition`. A compile failure would be TOOL_FAILURE.
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M8.

### AC-FRR-018: A refused done request leaves a pushed card unchanged, and the other edges keep their refusals (REQ-FRR-017)

Given card X is in state `pushed` in a repository with a remote, and every condition of AC-FRR-017 holds except one,
When the human decider requests the `done` transition for X,
Then the request is refused with the refusal that names the failing condition, not the illegal-transition refusal, and X's state, version, and event history are unchanged.
And given a card in state merged-local in the same repository, when `done` is requested for it, then the request is refused while the remote is configured, and the card is unchanged.
And given a card in state `pushed`, when `ci-green` is requested for it, then the request is refused as the reserved edge, and the card is unchanged.

- Classification: regression guard, following the classification of Q1 (operator verdict `d-20261010T053034Z-1a19`). Not a release gate.
- Verification: `go test ./internal/homestate/ -run '^TestPushedToDoneRefusalsLeaveRowUnchanged$' -v -count=1`; `go test ./internal/homestate/ -run '^TestMergedLocalDoneAndCiGreenEdgesUnchanged$' -v -count=1`. Each passes only with its own `=== RUN` and `--- PASS` lines.
- Predicted failure (not observed): on the unchanged tree every `pushed` → `done` request returns `ErrIllegalTransition`, so `TestPushedToDoneRefusalsLeaveRowUnchanged` fails on its named-refusal assertion. An unchanged-row check alone would pass vacuously, which is why the named refusal is asserted. `TestMergedLocalDoneAndCiGreenEdgesUnchanged` is green before and after by design (a regression guard, not a failure to predict).
- Stops firing: if either test is renamed, moved, or excluded, its ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M8.

### AC-FRR-019: The transition-count pin accepts exactly one new pair (REQ-FRR-016, REQ-FRR-017)

Given the record's transition table, and a fixture that places each pair of card states under the conditions that pair needs,
When every (from, to) pair of the card states is requested,
Then exactly 72 pairs are accepted and 369 are refused, the one pair newly accepted is `pushed` → `done`, every other pair keeps the outcome it had before this change (71 accepted), and every refused pair leaves the record unchanged.

- Classification: regression guard, following the classification of Q1 (operator verdict `d-20261010T053034Z-1a19`). Not a release gate. The count pin is re-authored in place, as decision-index Q2 anticipates when an edge is added.
- Verification: `go test ./internal/homestate/ -run '^TestFR_AC005_TransitionTableEdgeCount$' -v -count=1`; passes only with `=== RUN TestFR_AC005_TransitionTableEdgeCount` and `--- PASS: TestFR_AC005_TransitionTableEdgeCount`.
- Predicted failure (not observed): on the unchanged tree the test counts 71 accepted and 370 refused, against the re-authored expectation of 72 and 369, so its count assertion fails.
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M8.

### AC-FRR-020: The doctor reports out-of-band writes and exits non-zero (REQ-FRR-018)

Given a run whose record holds three cards: card A, written outside the transition API with its state set to `done` and no event that moves it into `done`; card B, whose version was raised with no event that carries the new version; and card C, which agrees with its history,
When the operator runs `moai factory doctor` on that record,
Then the output names card A and card B with the run, the recorded value, and the history value, does not name card C as a disagreement, and the exit status is non-zero.
And given a record in which every card agrees with its history, when the operator runs `moai factory doctor` on it, then the output reports no disagreement and the exit status is zero.

- Classification: regression guard, following the classification of Q1 (operator verdict `d-20261010T053034Z-1a19`). Not a release gate.
- Verification: `go test ./internal/cli/ -run '^TestDoctorReportsOutOfBandWrite$' -v -count=1`; passes only with `=== RUN TestDoctorReportsOutOfBandWrite` and `--- PASS: TestDoctorReportsOutOfBandWrite`.
- Predicted failure (not observed): the `factory` command group has no `doctor` subcommand (`internal/cli/factory_handoff_recover.go:20`), so the command returns an unknown-command error and the report assertions fail. The test goes through the command tree, so it compiles on the unchanged tree.
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M8.

### AC-FRR-021: The doctor is read-only and never reports an empty sweep as clean (REQ-FRR-019)

Given a record that holds cards for the run,
When the operator runs `moai factory doctor` on it,
Then the card rows and the event count are unchanged by the run, and the output reports the number of cards compared.
And given a run that holds no cards, when the operator runs `moai factory doctor` on it, then the output reports an empty sweep with zero cards compared, the exit status is non-zero, and the output carries no clean result (verification-completeness §1.1, the empty swept set).

- Classification: regression guard, following the classification of Q1 (operator verdict `d-20261010T053034Z-1a19`). Not a release gate.
- Verification: `go test ./internal/cli/ -run '^TestDoctorReadOnlyAndEmptySweep$' -v -count=1`; passes only with `=== RUN TestDoctorReadOnlyAndEmptySweep` and `--- PASS: TestDoctorReadOnlyAndEmptySweep`.
- Predicted failure (not observed): as for AC-FRR-020, the verb is not registered, so the command returns an unknown-command error and the assertions fail.
- Stops firing: if the test is renamed, moved, or excluded, the ledger row loses its `=== RUN` line and reads FAIL with `[no tests to run]`, which the sync audit reads from `progress.md` §E.2.
- Green path: M8.

## Edge cases

- Expiry exactly at the request instant follows the existing rule in `LeaseExpired` (`card_record.go:146`). The predicted-failure test pins that rule; the criteria do not restate it.
- Expiry of a lease on a merging card returns the card to blocked, not to assigned, so recovery does not apply (plan M3, item 2). This is covered by the M3 test set and is not a separate criterion.
- A non-positive or unparseable lease duration is expected to be refused at config load, naming the key. This expectation is not yet a criterion and is to be confirmed with the lane (decision-index, Optional, deferred).
- A holder whose worker row is missing follows the existing registration rule. AC-FRR-014 covers the non-holder side.

## Quality gates (not acceptance criteria)

- QG-1 Format and lint: `gofmt` and the project linter run clean on every touched Go package. The lane records the verbatim output.
- QG-2 Race: `go test -race ./internal/homestate/` after the transaction changes, with the verbatim output recorded.
- QG-3 Coverage: `go test -cover` on each touched package, with output recorded. The TRUST 5 target is 85%. A shortfall is recorded as a Gap, not as a pass.
- QG-4 Template-First, scoped to what this SPEC changes: after `make build`, each file this SPEC changes matches its template. The files are the lease-duration key in `.moai/config/sections/workflow.yaml` and its template, the watchdog `SKILL.md` in `.claude/` and in the template tree, and the catalogue and `factory-dispatch-mechanics.md` mirrors. The comparison is key-scoped for `workflow.yaml`, because the rendered file differs from the template on purpose. The repository-local overrides that must survive a template sync are, in `.moai/config/sections/workflow.yaml`: `branch_guard.enabled: true` (template `false`), `branch_guard.deny_commits_on: [main]` (template `[]`), and `commit_identity_guard.enabled: true` (template `false`). A sync that overwrote them would remove the protection on `main`. The lane records the verbatim comparison output.
- QG-5 Verdict parser untouched: `git diff --name-only 2aab5f797..HEAD` lists no file under `internal/auditverdict/`. The lane records the resolved HEAD SHA beside the output, because HEAD moves.
- QG-6 Traceability: every commit on the branch names card t1621 in its message.

## Definition of Done

- Each Verification line of AC-FRR-001 to AC-FRR-007, AC-FRR-011, AC-FRR-014, AC-FRR-015, and AC-FRR-017 to AC-FRR-021 shows `=== RUN` and `--- PASS` for its named test, with an executed count above zero, recorded in `progress.md` §E.2 after the run-phase fix commit.
- Every active criterion is a regression guard. None gates the release (decision-index Q1, operator verdict `d-20261010T053034Z-1a19`).
- Item (3) of the card re-scope is carried by AC-FRR-017 to AC-FRR-021 under leader ruling `d-20261010T060606Z-c5c4`. Its sub-choices are open in decision-index Q21, and the operator may confirm or overrule them.
- QG-1 to QG-6 pass, or each shortfall is recorded as a Gap.
- Decisions are recorded in `decision-index.md` before the plan re-audit runs. Q21 is open, and the autonomous Kickoff stays blocked until it is answered.
- Local landing only: no push, no pull request, and no remote operation from this lane.
