---
id: SPEC-FACTORY-LEASE-RENEW-001
version: "0.1.0"
updated: 2026-10-10
---

# Acceptance criteria: SPEC-FACTORY-LEASE-RENEW-001

These criteria verify the active requirements in `spec.md` (REQ-FLR-001 to REQ-FLR-012). Each criterion is a Given-When-Then scenario. The requirement text lives in `spec.md`, and the mapping from each criterion to its requirement is in `plan.md` §E. The active criteria are AC-FLR-001 to AC-FLR-013.

## Adoption status

At plan time no test has run, and no criterion is recorded as a pass. Every criterion carries a predicted RED (not observed) and names the test that the run phase writes or reads. The run phase records, in `progress.md` §E.2, the observed RED output of each named test before the GREEN output of that test. A criterion is recorded as PASS only from its GREEN output. No criterion is classified as a release gate in this revision.

A compile or discovery failure is TOOL_FAILURE and is never recorded as RED (`.claude/rules/moai/workflow/tdd-result-contract.md`).

**Verification convention.** Each Verification line runs one named test with `-v` and `-count=1`. A GREEN output counts only when it contains `=== RUN <Name>` and `--- PASS: <Name>` for that test, with an executed count above zero. `[no tests to run]`, a zero executed count, or a missing line is not GREEN. A test that is renamed, moved, or excluded loses its `=== RUN` line and so reads as not GREEN, which the sync audit reads from `progress.md` §E.2.

**Surface rule for predicted failures.** A test for a surface that does not exist yet reaches it through the public command or tool surface (the command tree, the MCP tool list, or a public handle such as `OpenFactory`). It names no Go symbol that does not exist yet, so it compiles on the unchanged tree and fails on an assertion.

**Mutant rule.** A case of a criterion that is green on the unchanged tree has no predicted RED of its own. Its RED input is a mutant that violates the requirement for that case. The run phase executes the mutant, records its failing output as the RED, and then restores the code.

## Acceptance criteria

### AC-FLR-001: Renewal extends a live lease (REQ-FLR-001)

Given card X is leased to lane L and the lease has 5 minutes left,
When L runs the factory renew verb for X,
Then the lease expiry becomes the request time plus the configured card lease duration, the verb prints that expiry, and the card's version is unchanged.

- Classification: behavioural check. Not a release gate.
- Verification: `go test ./internal/cli/ -run '^TestFactoryRenewExtendsLiveLease$' -v -count=1`.
- Predicted RED (not observed): the renew verb is not registered, so the test fails on its expiry assertion once it drives the command tree.
- Green path: M2.

### AC-FLR-002: The MCP tool matches the verb (REQ-FLR-002)

Given card X is leased to lane L,
When L renews X through the MCP renewal tool, and then, on the same state, through the factory renew verb,
Then both return the same expiry fields, and for an expired lease both return the same refusal class.

- Classification: behavioural check. Not a release gate.
- Verification: `go test ./internal/cli/ -run '^TestFactoryRenewMatchesMCP$' -v -count=1`.
- Predicted RED (not observed): no renewal tool is in the MCP tool list, so the MCP leg fails on its tool-presence assertion.
- Green path: M2.

### AC-FLR-003: A non-holder cannot renew a live lease (REQ-FLR-003)

Given card X's lease is held by lane L,
When a different lane M runs the factory renew verb for X, or calls the MCP renewal tool for X, while the lease is live,
Then each request is refused with the holder-refusal class, and the card's state, version, and events are unchanged.

- Classification: behavioural check. Not a release gate.
- Verification: `go test ./internal/cli/ -run '^TestNonHolderRenewRefusedCLIAndMCP$' -v -count=1`.
- Predicted RED (not observed): with no renew verb, each request fails with an unknown-command error, which does not match the holder-refusal class that the test asserts.
- Green path: M2.

### AC-FLR-004: An expired renewal is refused with no write (REQ-FLR-004)

Given card X's lease is held by lane L and expired two minutes ago,
When L runs the factory renew verb for X,
Then the request is refused with the expired-lease class, and the card's state, version, and event count are unchanged.

- Classification: behavioural check. Not a release gate.
- Verification: `go test ./internal/homestate/ -run '^TestRenewRefusesExpiredWithoutWrite$' -v -count=1`.
- Predicted RED (not observed): `RenewLease` reaps the expired lease before it checks the holder (`internal/homestate/card_transition.go:441-456`), which bumps the version and appends a lease event, so the unchanged-version assertion fails.
- Decision link: if decision-index Q3 selects reap-then-refuse, this criterion and REQ-FLR-004 are reworded in the same revision. Under that option the criterion is not GREEN as written.
- Green path: M2.

### AC-FLR-005: The card lease duration is read from the key, with a 15-minute default (REQ-FLR-005)

Given the project's workflow configuration does not set `workflow.factory_lease.duration`,
When the factory is opened with `OpenFactory` and a lease is acquired on card X,
Then the lease expiry is the acquisition time plus 15 minutes.
And given the same configuration sets `workflow.factory_lease.duration` to `20m`, when the factory is opened and a lease is acquired on card X, then the lease expiry is the acquisition time plus 20 minutes.

- Classification: behavioural check. Not a release gate.
- Verification: `go test ./internal/homestate/ -run '^TestCardLeaseUsesConfiguredDuration$' -v -count=1`.
- Predicted RED (not observed): the key is not read, so the 20-minute case fails with a 15-minute expiry. The default case is green on the unchanged tree, so its RED input is a mutant that uses a different default, executed in the run phase.
- Green path: M1.

### AC-FLR-006: A non-positive card lease duration refuses to open (REQ-FLR-006)

Given the project's workflow configuration sets `workflow.factory_lease.duration` to `0s`,
When the factory is opened with `OpenFactory`,
Then the open is refused, and the refusal names `workflow.factory_lease.duration`.

- Classification: behavioural check. Not a release gate.
- Verification: `go test ./internal/homestate/ -run '^TestFactoryOpenRefusesNonPositiveLeaseDuration$' -v -count=1`.
- Predicted RED (not observed): the open succeeds, so the refusal assertion fails.
- Green path: M1.

### AC-FLR-007: The renewal period is bounded at one half of the card lease duration (REQ-FLR-007)

Given the card lease duration is 10 minutes,
When the factory is opened with `workflow.factory_lease.renew_interval` set to 6 minutes, then the open is refused and the refusal names `workflow.factory_lease.renew_interval` and the bound of 5 minutes; when the key is set to 5 minutes, the open succeeds.
And given the card lease duration is 15 minutes, when the key is set to 7m30s, the open is refused for a value that is not a whole number of minutes; when the key is set to 8 minutes, the open is refused for exceeding the bound of 7 minutes 30 seconds; when the key is set to 7 minutes, the open succeeds.
And given the card lease duration is 4 hours, when the key is set to 60 minutes, the open is refused for a value above 59 minutes.

- Classification: behavioural check. Not a release gate. This criterion is the behavioural enforcement of the bound stated in REQ-FLR-007.
- Verification: `go test ./internal/homestate/ -run '^TestFactoryOpenRefusesRenewIntervalOutsideBound$' -v -count=1`; `go test ./internal/homestate/ -run '^TestFactoryOpenAcceptsRenewIntervalAtBound$' -v -count=1`.
- Predicted RED (not observed): the key is not read, so each refusal case opens the factory and its refusal assertion fails. The acceptance cases (5 minutes at 10 minutes, and 7 minutes at 15 minutes) are green on the unchanged tree. Their RED input is a mutant that refuses at the bound, which the run phase executes.
- Green path: M1.

### AC-FLR-008: The queue claim lease ignores the card lease duration key (REQ-FLR-008)

Given `workflow.factory_lease.duration` is set to `20m` in the project's workflow configuration and the factory has been opened,
When a queue claim is made,
Then the expiry that the claim writes is 15 minutes after the claim.

- Classification: behavioural check, regression guard. Not a release gate.
- Verification: `go test ./internal/factory/ -run '^TestBacklogClaim_QueueLeaseIgnoresCardLeaseDuration$' -v -count=1`, a new test; `go test ./internal/factory/ -run '^TestBacklogClaim_ReclaimsExpired$' -v -count=1` and `go test ./internal/factory/ -run '^TestBacklogClaim_RenewAfterExpiry$' -v -count=1`, which must stay GREEN.
- Predicted RED: none on the unchanged tree, because the claim already uses 15 minutes. The RED input is the mutant rule's mutant: a queue claim that reads the card lease duration key. The run phase executes it, records its failing output as the RED, and restores the code.
- Green path: M1 keeps it GREEN.

### AC-FLR-009: The renewal period default and the effective values are reported (REQ-FLR-009)

Given the project's workflow configuration sets neither lease key,
When `moai factory status --json` runs in that project,
Then the output reports `lease_duration` as `15m0s` and `renew_interval` as `5m0s`.

- Classification: behavioural check. Not a release gate.
- Verification: `go test ./internal/cli/ -run '^TestFactoryStatusReportsLeaseKeys$' -v -count=1`.
- Predicted RED (not observed): the status output carries neither field, so the assertion on both fields fails.
- Green path: M1.

### AC-FLR-010: The watchdog renews before it measures (REQ-FLR-010)

Given the lane stall watchdog skill is read for a pass,
When its procedure and its allowed tools are read,
Then the first procedural step renews the lane's own card lease through the factory renew verb, the renew command is among the allowed tools, and the measurement step follows the renewal.

- Classification: procedure-text check, regression guard. Limit: this checks procedure text, not a live pass, so a text-only change passes it.
- Verification: `go test ./internal/template/ -run '^TestLaneWatchdogRenewsFirst$' -v -count=1`.
- Predicted RED (not observed): the current procedure opens with measurement (§1), and its `allowed-tools` lists no renew command (SKILL.md line 17).
- Green path: M3.

### AC-FLR-011: A refused watchdog renewal is recorded and measurement still runs (REQ-FLR-011)

Given the watchdog procedure is read,
When its renewal step and its refusal branch are read,
Then the procedure states that a renewal that the factory record refuses is recorded in the watchdog snapshot, and that the lane progress measurement still runs in the same pass.

- Classification: procedure-text check, regression guard. Limit: as for AC-FLR-010, the behaviour of a watchdog turn in a live lane is not observed.
- Verification: `go test ./internal/template/ -run '^TestLaneWatchdogRecordsRefusalAndMeasures$' -v -count=1`.
- Predicted RED (not observed): the current procedure states no renewal refusal branch.
- Green path: M3.

### AC-FLR-012: One late pass is tolerated, and an expired lease is refused (REQ-FLR-001, REQ-FLR-004, REQ-FLR-007, REQ-FLR-009, REQ-FLR-012)

Given the card lease duration D and the renewal period I are configured, with I no longer than one half of D,
When a lane renews on schedule at times 0, I, 2I, and so on, through 3D,
Then every renewal succeeds.
And given the renewal due at time T plus I is delayed by D minus I minus one second, when it is requested, then it succeeds and extends the lease to its request time plus D.
And given the last renewal was at time T, when the next renewal is requested at exactly T plus D, and again at T plus D plus one second, then both requests are refused with the expired-lease class, and the card's state and version are unchanged.

- Classification: behavioural check, simulated with the injected clock of the factory tests. Not a release gate. It checks the margin that REQ-FLR-007 states. It does not measure wake timing.
- Verification: `go test ./internal/cli/ -run '^TestRenewCadenceKeepsLeaseLive$' -v -count=1`.
- Predicted RED (not observed): the renew verb is not registered, so the first on-schedule renewal fails its assertion.
- Green path: M4.

### AC-FLR-013: The standing wake is armed at the renewal period (REQ-FLR-012)

Given the standing-wake paragraph of `auto-semantics.md` is read,
When its arm expression is read,
Then the expression is a minute step of `workflow.factory_lease.renew_interval`, the fixed literal `7,27,47 * * * *` is absent, and the paragraph states that no two consecutive passes are further apart than one renewal period.

- Classification: procedure-text check, regression guard. Limit: the factory record does not observe the armed expression (spec.md Residual risk).
- Verification: `go test ./internal/template/ -run '^TestAutoSemanticsArmFollowsRenewInterval$' -v -count=1`.
- Predicted RED (not observed): the paragraph carries the literal `7,27,47 * * * *` (`auto-semantics.md:108`).
- Green path: M3.

## Edge cases

- Expiry at the request instant counts as expired (`internal/homestate/card_record.go:143-146`). AC-FLR-012 pins the instant.
- A renewal period that changes between two passes takes effect at the next arm. The standing wake is checked at card intake and after every `/clear`, as `auto-semantics.md` §5.1 already requires.
- A card whose lease has lapsed is outside the margin in AC-FLR-012. Its restoration is transferred to card t1639.

## Quality gates (not acceptance criteria)

- QG-1 Format and lint: `gofmt` and the project linter run clean on every touched Go package. The lane records the verbatim output.
- QG-2 Race: `go test -race ./internal/homestate/ ./internal/factory/` after the `RenewLease` change, with the verbatim output recorded.
- QG-3 Coverage: `go test -cover` on each touched package, with the output recorded. The TRUST 5 target is 85%. A shortfall is recorded as a Gap, not as a pass.
- QG-4 Template-First, scoped to what this SPEC changes: after `make build`, each changed file matches its template. The files are the lease keys in `workflow.yaml` (compared by key, because the rendered file differs from the template on purpose), the watchdog `SKILL.md` and the standing-wake rule in `.claude/` (compared byte for byte with their template copies), the MCP catalogue rule, and the `factory-dispatch-mechanics.md` mirror. The repository-local overrides that a template sync must keep are, in `.moai/config/sections/workflow.yaml`: `branch_guard.enabled: true`, `branch_guard.deny_commits_on: [main]`, and `commit_identity_guard.enabled: true`. The template ships `branch_guard.enabled: false`, `branch_guard.deny_commits_on: []`, and `commit_identity_guard.enabled: false`. A sync that overwrote the overrides would remove the protection on `main`. The lane records the verbatim comparison output.
- QG-5 Traceability: every commit on the branch names card t1663 in its message.

## Definition of Done

- Each Verification line of AC-FLR-001 to AC-FLR-013 has, in `progress.md` §E.2, an observed RED output followed by a GREEN output for its named test. The GREEN output carries `=== RUN` and `--- PASS` with an executed count above zero. Where a case is green on the unchanged tree, the observed RED is the output of its mutant, as the mutant rule states.
- No criterion is recorded as PASS before its GREEN output is recorded, and no criterion is recorded as PASS from a RED output.
- QG-1 to QG-5 pass, or each shortfall is recorded as a Gap.
- No criterion is a release gate in this revision.
- The `Operator verdict:` line is left empty at authoring; the operator writes it. This applies to each open FOUNDER row of decision-index.md.
- Local landing only: no push, no pull request, and no remote operation from this lane.
