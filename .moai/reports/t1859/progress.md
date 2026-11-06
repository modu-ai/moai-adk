# Card t1859 progress (Class B, no SPEC)

Defect: a `kickoff` card whose owner lane is dead could be neither reassigned (`ReassignCard` refused non-holding, non-assigned states) nor given the audit kickoff decision (T8a: "no registered owner to lease to").

Fix (shape A): `ReassignCard` also admits `kickoff` and `needs-decision`. State, Decision* fields stay; owner swaps; registered-lane, run-binding, and cross-run live-lease checks unchanged.

## RED (before the guard change; tests added, production unchanged except an unused helper + doc comment)

Command: `go test -count=1 -run 'TestReassignCard' ./internal/homestate/`

```
--- FAIL: TestReassignCardAdmitsDecisionWaitingCardsWithADeadOwner (0.78s)
    --- FAIL: .../kickoff
        card_reassign_test.go:265: reassign of a kickoff card: illegal card transition: card rd1 is kickoff; reassignment moves an assigned card, or a card whose lease expired
    --- FAIL: .../needs-decision
        card_reassign_test.go:265: reassign of a needs-decision card: illegal card transition: card rd1 is needs-decision; reassignment moves an assigned card, or a card whose lease expired
--- FAIL: TestReassignCardDecisionWaitingKeepsTheTargetRefusals (0.26s)
    card_reassign_test.go:306: unregistered target: err = illegal card transition: card rd2 is kickoff; ...
--- FAIL: TestReassignCardUnwedgesTheAuditKickoffDecision (0.97s)
    card_reassign_test.go:327: reassign: illegal card transition: card wedge is kickoff; ...
FAIL	github.com/modu-ai/moai-adk/internal/homestate
```

`TestReassignCardStillRefusesOtherNonHoldingStates` passed in RED and GREEN (it pins picked/blocked/done/failed still refuse).

## GREEN

- `go build ./...` -> ok
- `go vet ./internal/homestate/ ./internal/cli/...` -> ok
- `go test -count=1 ./internal/homestate/` -> `ok  github.com/modu-ai/moai-adk/internal/homestate  237.695s`
- `go test -count=1 -run '^(TestFactoryAssignReassignHappyPath|TestFactoryAssignReassignRefusals)$' ./internal/cli/` -> `ok  github.com/modu-ai/moai-adk/internal/cli  10.719s`

## Files changed

- internal/homestate/card_reassign.go (guard, `decisionWaitingState` helper, doc comment, refusal message)
- internal/homestate/card_reassign_test.go (4 new tests)
- internal/cli/factory_card.go (two flag help strings that became false)

## Gaps

- Full `internal/cli` package, `go test ./...`, and golangci-lint were not run.
- No live-DB check: the real t1840 card was not reassigned or decided; verification is by in-memory fixture only.
- No test of the CLI path `assign --reassign` against a kickoff card (only the two existing name-exact CLI reassign tests were run, unchanged).
- Docs/templates mentioning reassignment were grepped under internal/ and cmd/ only (`.claude`, docs not found containing the old wording via targeted grep).
- An earlier backgrounded test run printed FAIL/ok output from before the guard edit; it was discarded and not cited.
