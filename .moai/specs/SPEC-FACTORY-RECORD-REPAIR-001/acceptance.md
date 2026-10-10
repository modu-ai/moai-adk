---
id: SPEC-FACTORY-RECORD-REPAIR-001
version: "0.1.0"
updated: 2026-10-10
---

# Acceptance criteria: SPEC-FACTORY-RECORD-REPAIR-001

These criteria verify the requirements in `spec.md` (REQ-FRR-001 to REQ-FRR-015). Each is a Given-When-Then scenario. The requirement text itself lives in `spec.md`.

## Adoption status

The plan phase did not run `go test`, as the lane instructed, so no RED cell has been observed yet. Under verification-completeness §2, a criterion is adopted only when its RED-now cell is observed on the unchanged tree at `2aab5f797`, with four elements recorded: the command, its verbatim stdout, its exit code, and the tree SHA. Until then every criterion is pending. A criterion that is already green on the unchanged tree is marked as a regression guard. The "expected RED reason" lines state what the code read in `plan.md` § 검증된 전제 predicts. They are predictions, not observations.

## Acceptance criteria

### AC-FRR-001: Renewal extends a live lease (REQ-FRR-001)

Given card X is leased to lane L and L's lease has about 5 minutes left,
When L runs the factory renew verb for X,
Then the lease expiry becomes the request time plus the configured duration, the verb prints that expiry, and the card's version is unchanged.

- Verification: `go test ./internal/cli/ -run TestFactoryRenewExtendsLiveLease -count=1`
- Expected RED reason (predicted): the verb does not exist.
- Green path: M2.

### AC-FRR-002: The MCP tool matches the verb (REQ-FRR-002)

Given card X is leased to lane L,
When L renews X through the MCP tool, and then, on the same state, through the factory renew verb,
Then both return the same expiry fields, and for an expired lease both return the same refusal class.

- Verification: `go test ./internal/cli/ -run TestFactoryRenewMatchesMCP -count=1`
- Expected RED reason (predicted): no MCP tool is declared.
- Green path: M2.

### AC-FRR-003: The lease duration is configurable, with a 15-minute default (REQ-FRR-003)

Given no lease-duration key is configured,
When a lease is acquired on card X,
Then its expiry is the acquisition time plus 15 minutes.
And given the key is set to 20m, when a lease is acquired on card X, then its expiry is the acquisition time plus 20 minutes.

- Verification: `go test ./internal/config/ -run TestFactoryLeaseDurationDefault -count=1`; `go test ./internal/homestate/ -run TestCardLeaseUsesConfiguredDuration -count=1`
- Expected RED reason (predicted): the key and the field do not exist.
- Green path: M1.

### AC-FRR-004: An expired renewal is refused with no write (REQ-FRR-004)

Given card X's lease is held by L and expired two minutes ago,
When L runs the factory renew verb for X,
Then the request is refused, and the card's state, version, and event count are unchanged.

- Verification: `go test ./internal/homestate/ -run TestRenewRefusesExpiredWithoutWrite -count=1`
- Expected RED reason (predicted): the renewal reaps the lease first (`card_transition.go:452-457`), which bumps the version and appends `lease.expired`.
- Green path: M2.

### AC-FRR-005: The watchdog renews before it measures (REQ-FRR-005)

Given the lane stall watchdog skill is read for an awaken pass,
When its procedure and allowed tools are read,
Then the first procedural step renews the lane's own card lease, the renew command is among the allowed tools, and the measurement step follows the renewal.

- Verification: `go test ./internal/template/ -run TestLaneWatchdogRenewsFirst -count=1`
- Expected RED reason (predicted): the current procedure opens with measurement (§1) and lists no renew command (`allowed-tools`, line 17).
- Green path: M6.

### AC-FRR-006: One-shot recovery through the stage verb and the parent-checkout next verb (REQ-FRR-006)

Given card X is at run, leased to L, and L's lease expired three minutes ago,
When L runs the factory stage verb to run for X,
Then one invocation succeeds, the card's events read the expiry followed by the transition, and the card ends leased to L at stage run.
And given the same starting state, when L runs the factory next verb with the card id from the parent checkout, then one invocation succeeds with the same end state.
And given the requested edge is illegal from the card's stage, when L makes the request, then the request is refused and the expired row is unchanged.

- Verification: `go test ./internal/cli/ -run TestStageRecoversOwnExpiredLease -count=1`; `go test ./internal/cli/ -run TestNextRecoversOwnExpiredLeaseFromParentCheckout -count=1`; `go test ./internal/homestate/ -run TestRecoveryAllOrNothing -count=1`
- Expected RED reason (predicted): the transition returns `ErrLeaseExpired` (`card_transition.go:315`), and next refuses as in flight (`factory_card.go:1125-1128`).
- Green path: M3.

### AC-FRR-007: The expired refusal names the recovery verb (REQ-FRR-007)

Given card X's lease is held by L and has expired,
When a caller other than L requests a transition on X,
Then the refusal text names the recovery verb that L runs to re-lease X.

- Verification: `go test ./internal/homestate/ -run TestExpiredRefusalNamesRecoveryVerb -count=1`
- Expected RED reason (predicted): the refusal text names no verb (`card_transition.go:315`).
- Green path: M3.

### AC-FRR-008: A sync-audit FAIL on the card's own code returns the card to run (REQ-FRR-008)

Given card X is at sync-audit, leased to L, and its current sync verdict is FAIL and names blocking defects in the card's own code,
When L requests the sync-audit to run transition,
Then the card moves to run with stage run, the lease stays with L, and the evidence record is kept.

- Verification: `go test ./internal/homestate/ -run TestSyncAuditFailOwnScopeReturnsToRun -count=1`
- Expected RED reason (predicted): no sync-audit → run row exists (`card_transition.go:133-181`), so the request is an illegal transition.
- Green path: M4. The discriminator is Decisions owed, D3.

### AC-FRR-009: Other sync verdicts do not return the card to run (REQ-FRR-009)

Given card X is at sync-audit, leased to L, and its verdict is PASS, or FAIL without own-scope blocking defects,
When L requests the sync-audit to run transition,
Then the request is refused with the own-scope reason, and the card's state, version, and events are unchanged.

- Verification: `go test ./internal/homestate/ -run TestSyncAuditRefusesRunWithoutOwnScopeFail -count=1`
- Expected RED reason (predicted): on the unchanged tree the request is refused only because no edge exists. The RED cell must pin the own-scope reason string, so this criterion is red until the guard exists. A bare refusal check is a wrong-reason green and does not qualify.
- Green path: M4.

### AC-FRR-010: A holder's re-stamp keeps the stage and the lease (REQ-FRR-010)

Given card X is at plan-audit, leased to L, with a stale evidence hash, and L holds a verdict whose audited SHA is the new evidence commit,
When L requests the re-audit entry with the new hash,
Then the card stays at plan-audit under L's lease, the evidence hash equals the new SHA, and the event history keeps every prior event and adds one.
And the same holds at sync-audit.
And given the verdict's audited SHA differs from the new hash, when L requests the re-stamp, then the request is refused and the card is unchanged.

- Verification: `go test ./internal/homestate/ -run TestHolderRestampsEvidenceAtPlanAudit -count=1`; `go test ./internal/homestate/ -run TestHolderRestampsEvidenceAtSyncAudit -count=1`
- Expected RED reason (predicted): no same-stage row exists (`card_transition.go:133-181`).
- Green path: M4.

### AC-FRR-011: The lease records the worktree path (REQ-FRR-011)

Given card X is assigned to lane L and no worktree path is recorded,
When L acquires the lease from its own worktree,
Then the card records L's worktree path.
And given a different path is already recorded, when L acquires the lease, then the acquisition is refused and the card is unchanged.

- Verification: `go test ./internal/homestate/ -run TestLeaseRecordsWorktreePath -count=1`
- Expected RED reason (predicted): the lease write sets the holder, heartbeat, and expiry only (`card_transition.go:540-545`).
- Green path: M1.

### AC-FRR-012: The interim guard refuses an empty-stage run for a SPEC-bearing card (REQ-FRR-012)

Given card X has a valid SPEC id, an empty stage, no admitted plan verdict, and no recorded operator decision,
When X is moved to run,
Then the transition is refused and the row is unchanged.
And given an admitted plan verdict is recorded for X, when X is moved to run, then the transition is admitted.

- Verification: `go test ./internal/homestate/ -run TestEmptyStageToRunRefusedWithSpecID -count=1`; `go test ./internal/homestate/ -run TestEmptyStageToRunAdmittedWithPlanVerdict -count=1`
- Expected RED reason (predicted): `stageAdmits` admits the empty-stage run unconditionally (`card_transition.go:752-753`). The admitted-side test is green on the unchanged tree and is adopted only together with the refused side.
- Green path: M5.

### AC-FRR-013: Unchanged routes stay unchanged (REQ-FRR-013)

Given card Y has an empty SPEC id and an empty stage,
When Y is moved to run,
Then the transition is admitted as before, and the interim guard does not refuse the sync-audit to run transition.

- Verification: `go test ./internal/homestate/ -run TestEmptyStageToRunUnchangedWithoutSpecID -count=1`
- Status: regression guard. Green before and after by design, so it is not a release-blocking RED criterion.
- Green path: M5 keeps it green.

### AC-FRR-014: A non-holder can neither renew nor reap an expired lease (REQ-FRR-014)

Given card X's lease is held by L and has expired,
When a different label M runs the factory renew verb for X, and separately runs the factory stage verb for X,
Then both requests are refused, and the card's state, version, and events are unchanged.

- Verification: `go test ./internal/homestate/ -run TestNonHolderCannotReapOrRenew -count=1`
- Expected RED reason (predicted): the expiry branch reaps before the holder check (`card_transition.go:311-316` and `:452-457`), so the version and events change.
- Green path: M2 for the renewal and M3 for the transition.

### AC-FRR-015: The queue claim lease is unchanged (REQ-FRR-015)

Given the card lease duration is configured to 20m,
When a queue claim is made,
Then its lease expiry stays 15 minutes after the claim, as the existing claim-lease tests assert.

- Verification: the existing claim-lease test in `internal/factory/backlog_claim_test.go` (near lines 148-155 and 275-293), run as `go test ./internal/factory/ -run <name confirmed in pre-flight> -count=1`
- Status: regression guard. Green before and after by design.
- Green path: M1 keeps it green.

### AC-FRR-016: The pinned edge count equals the table (supports REQ-FRR-008 and REQ-FRR-010)

Given the transition table, after the sync-audit → run row and the same-stage rows are added,
When the accepted-pair test runs,
Then the pinned count equals the number of accepted pairs, and the `@MX:ANCHOR` reason states the same count.

- Verification: the AC-005 test in `internal/homestate/fr_transition_test.go` (near line 124), run as `go test ./internal/homestate/ -run <name confirmed in pre-flight> -count=1`
- Expected RED reason (predicted): once the rows exist and before the pin moves, the count test fails. The baseline count is measured in pre-flight item 2, not assumed to be 71.
- Green path: M4.

## Edge cases

- Expiry exactly at the request instant follows the existing rule in `LeaseExpired` (`card_record.go:146`). The RED test pins that rule; the criteria do not restate it.
- Expiry of a lease on a merging card returns the card to blocked, not to assigned, so recovery does not apply (plan M3, item 2). This is covered by the M3 test set and is not a separate criterion.
- A non-positive or unparseable lease duration is expected to be refused at config load, naming the key. This expectation is not yet a criterion and is to be confirmed with the lane.
- A holder whose worker row is missing follows the existing registration rule. AC-FRR-014 covers the non-holder side.
- A sync-audit FAIL with blocking defects outside the card's scope is covered by AC-FRR-009.
- A card whose SPEC id is empty while leased is covered by AC-FRR-013 (unchanged), and the residual is stated in spec.md.

## Quality gates (not acceptance criteria)

- QG-1 Format and lint: `gofmt` and the project linter run clean on every touched Go package. The lane records the verbatim output.
- QG-2 Race: `go test -race ./internal/homestate/` after the transaction changes, with the verbatim output recorded.
- QG-3 Coverage: `go test -cover` on each touched package, with output recorded. The TRUST 5 target is 85%. A shortfall is recorded as a Gap, not as a pass.
- QG-4 Template-First: the rendered `.claude/` and `.moai/` copies match the template bytes after `make build`. The check compares the template with each rendered file.
- QG-5 Verdict parser untouched: `git diff --name-only 2aab5f797..HEAD` lists no file under `internal/auditverdict/`.
- QG-6 Traceability: every commit on the branch names card t1621 in its message.

## Definition of Done

- Each of AC-FRR-001 to AC-FRR-016 has an observed RED cell with its four elements, and an observed green output recorded in the card's progress record. The regression guards (AC-FRR-013 and AC-FRR-015) are marked as such.
- QG-1 to QG-6 pass, or each shortfall is recorded as a Gap.
- Decisions owed D1 to D13 are answered in the lane's progress record before the plan audit runs.
- The plan audit runs on the session model, with codex as the required backend.
- Local landing only: no push, no pull request, and no remote done edge.
