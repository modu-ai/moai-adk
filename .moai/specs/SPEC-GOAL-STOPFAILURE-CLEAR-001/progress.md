---
id: SPEC-GOAL-STOPFAILURE-CLEAR-001
artifact: progress
version: "0.1.0"
status: implemented
created: 2026-08-19
updated: 2026-08-19
---

## Run-phase Evidence

Worktree `.claude/worktrees/t139`, branch `worktree-t139`, base `e7aeec088` (`merge-base --is-ancestor release/v3.1.1 HEAD` → rc=0).

### Reproduction first (RED)

The failing test was written and observed failing before any production change:

```
$ go test ./internal/hook/ -run 'StopFailure|Unrecoverable' -count=1
--- FAIL: TestStopFailureHandler_DisarmsGoalOnUnrecoverable/authentication_failed
    goal still armed after an unrecoverable authentication_failed; it will spin idle turns to the ceiling
    systemMessage does not mention the goal that was disarmed: "Authentication failed. Check your API key or run 'moai glm setup'."
--- FAIL: TestStopFailureHandler_DisarmsGoalOnUnrecoverable/oauth_org_not_allowed
--- FAIL: TestStopFailureHandler_DisarmsGoalOnUnrecoverable/billing_error
FAIL
```

The recoverable-side and fail-open tests passed at RED, because they assert behavior that already held — they exist to keep it holding.

### A defect the existing suite caught mid-implementation

The first implementation guarded absence with `if _, err := goal.LoadGoal(...); err != nil`. That is wrong: `LoadGoal` reports an absent goal as `(nil, nil)`, not as an error (`internal/goal/state.go`), so the guard fell through and the handler announced a disarm that had disarmed nothing. Four pre-existing `TestStopFailureHandler_Handle` cases failed on the appended sentence and surfaced it.

Fixed by checking the value (`g == nil`), and `TestStopFailureHandler_FailOpen` now asserts the absence of the word "disarmed" rather than only the absence of an error — an error-only assertion is what let the claim through in the first place.

### GREEN

```
$ go test ./internal/hook/ -run 'StopFailure|Unrecoverable' -count=1
ok  	github.com/modu-ai/moai-adk/internal/hook	1.885s
```

### Full affected-package verification

```
$ go build ./...   → exit 0
$ go vet ./...     → exit 0
$ go test ./internal/goal/... ./internal/hook/... -count=1   → exit 0
ok  github.com/modu-ai/moai-adk/internal/goal              0.471s
ok  github.com/modu-ai/moai-adk/internal/hook             45.997s
ok  github.com/modu-ai/moai-adk/internal/hook/{handoff,memo,memo/taxonomy,mx,mx/complexity,perf,quality,security,testutil,trace}  (10/10)

$ go test ./internal/template/... -count=1
ok  github.com/modu-ai/moai-adk/internal/template  35.240s

$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/ -run 'Neutrality|Parity|Mirror|Leak|Internal' -count=1
ok  github.com/modu-ai/moai-adk/internal/template  1.201s

$ make build → catalog.yaml updated successfully (12899 bytes), build OK
```

The full suite was NOT run locally, per the lane-local verification rule.

## AC matrix

| AC | Result | Evidence |
|---|---|---|
| AC-GSF-001 | PASS | `TestStopFailureHandler_DisarmsGoalOnUnrecoverable/authentication_failed` |
| AC-GSF-002 | PASS | same test, `oauth_org_not_allowed` + `billing_error` subtests |
| AC-GSF-003 | PASS | `TestStopFailureHandler_KeepsGoalOnRecoverable`, 8 subtests incl. the empty type |
| AC-GSF-004 | PASS | message assertions in both tests above |
| AC-GSF-005 | PASS | `TestStopFailureHandler_FailOpen`, 3 subtests |
| AC-GSF-006 | PASS | `grep -c oauth_org_not_allowed .claude/settings.json internal/template/templates/.claude/settings.json.tmpl` → 1, 1 |
| AC-GSF-007 | PASS | `TestIsUnrecoverableStopFailure`, all 10 documented types + the empty string |
| AC-GSF-008 | PASS | the batch above |

## Gaps

- **Context-overflow disarm is not delivered**, because the documented `error_type` enum has no value for it. Named in the SPEC rather than silently rounded into "unrecoverable errors are handled".
- **The end-to-end path is not exercised.** The tests drive `Handle` directly; that Claude Code actually emits `StopFailure` with these `error_type` values on a real dead turn is taken from the documented protocol table, not observed here.
- **`invalid_request` and `model_not_found` are classified recoverable** — defensible (the goal is not their cause and a corrected request resumes the work) but it is a judgement, not a measurement. `TestIsUnrecoverableStopFailure` pins it so a change is deliberate.
- The 30-minute check-in half of the upstream change is out of scope by instruction.

## Phase note

plan → run → sync ran serially in this one lane, landing as two commits rather than three: the plan artifacts, then implementation plus the status transition. The three-phase close's separate `implemented` → `completed` sync commit was folded into the second commit because this card ships inside a shared release branch whose integration commit the lead owns.
