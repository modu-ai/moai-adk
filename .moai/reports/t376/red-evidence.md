# RED evidence — SPEC-STATUS-TRANSITION-VALIDITY-001 (card t376)

The verbatim pre-GREEN failing output. `.log` files are gitignored in this repository
(`.gitignore:106` — `*.log`), so the RED runs are transcribed here to survive the worktree.
Tree at capture time: `a5d963db6` (`WT-status-transition-gap`).

Note what the M1 run establishes beyond "the rule is absent": every one of the 20
must-stay-silent cases passed **vacuously** on this tree, because a rule that does not exist
is silent about everything. That is precisely why AC-STV-010's live control exists, and it is
the last line of the M1 block below — an all-silent run is a harness failure, not a clean pass.

## M1 — `go test ./internal/spec/ -run 'TestStatusTransition' -count=1` (exit 1)

```text
--- FAIL: TestStatusTransitionValidityRule (20.49s)
    --- FAIL: TestStatusTransitionValidityRule/draft_to_completed_is_caught (0.86s)
        lint_transition_test.go:354: AC-STV-001: want exactly 1 StatusTransitionInvalid, got 0: []
    --- FAIL: TestStatusTransitionValidityRule/completed_to_draft_is_caught (1.11s)
        lint_transition_test.go:354: AC-STV-002: want exactly 1 StatusTransitionInvalid, got 0: []
    --- FAIL: TestStatusTransitionValidityRule/completed_to_implemented_reversal_is_caught (0.84s)
        lint_transition_test.go:354: AC-STV-014: want exactly 1 StatusTransitionInvalid, got 0: []
    --- FAIL: TestStatusTransitionValidityRule/draft_to_completed_without_trailer_is_caught (0.70s)
        lint_transition_test.go:354: AC-STV-003: want exactly 1 StatusTransitionInvalid, got 0: []
    --- FAIL: TestStatusTransitionValidityRule/synced_token_fires_its_own_code (0.75s)
        lint_transition_test.go:372: AC-STV-015: want exactly 1 StatusTokenUnrecognized, got 0: []
    --- FAIL: TestStatusTransitionValidityRule/approved_token_fires_its_own_code (0.88s)
        lint_transition_test.go:372: AC-STV-015: want exactly 1 StatusTokenUnrecognized, got 0: []
    --- FAIL: TestStatusTransitionValidityRule/cancelled_token_fires_its_own_code (0.91s)
        lint_transition_test.go:372: AC-STV-015: want exactly 1 StatusTokenUnrecognized, got 0: []
    --- FAIL: TestStatusTransitionValidityRule/case_variant_token_fires_its_own_code (0.95s)
        lint_transition_test.go:372: AC-STV-015: want exactly 1 StatusTokenUnrecognized, got 0: []
    lint_transition_test.go:405: AC-STV-010 live control: no case in this run produced a finding — the harness, not the corpus, is what this failure is about
--- FAIL: TestStatusTransitionTrailerIndependence (0.85s)
    lint_transition_test.go:426: trailer="manager-develop": want exactly 1 StatusTransitionInvalid, got 0: []
--- FAIL: TestStatusTransitionFindingGates (0.79s)
    lint_transition_test.go:470: AC-STV-018: want exactly 1 StatusTransitionInvalid, got 0: []
--- FAIL: TestStatusTransitionCheckOrder (1.25s)
    --- FAIL: TestStatusTransitionCheckOrder/token_check_precedes_pair_check (0.74s)
        lint_transition_test.go:509: want 1 StatusTokenUnrecognized, got 0: []
--- FAIL: TestStatusTransitionRuleIsObservationOnly (1.04s)
    lint_transition_test.go:536: AC-STV-012: control case did not fire — the run under test may not have exercised the rule
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/spec	24.910s
FAIL
```

## M4 — `go test ./internal/spec/ -run 'TestApplyEraDemotion|TestDemotionCause' -count=1` (exit 1)

```text
# github.com/modu-ai/moai-adk/internal/spec [github.com/modu-ai/moai-adk/internal/spec.test]
internal/spec/lint_demotion_cause_test.go:20:16: undefined: demotionCause
internal/spec/lint_demotion_cause_test.go:26:18: undefined: demotionCause
internal/spec/lint_demotion_cause_test.go:32:18: undefined: demotionCause
internal/spec/lint_demotion_cause_test.go:38:18: undefined: demotionCause
internal/spec/lint_demotion_cause_test.go:81:30: undefined: demotionCause
internal/spec/lint_demotion_cause_test.go:86:6: undefined: demotionCause
FAIL	github.com/modu-ai/moai-adk/internal/spec [build failed]
FAIL
```
