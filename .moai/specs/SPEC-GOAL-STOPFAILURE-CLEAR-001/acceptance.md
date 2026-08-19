---
id: SPEC-GOAL-STOPFAILURE-CLEAR-001
artifact: acceptance
version: "0.1.0"
status: implemented
created: 2026-08-19
updated: 2026-08-19
---

| ID | Criterion | Verification |
|---|---|---|
| AC-GSF-001 | An armed goal is gone after a `StopFailure` carrying `authentication_failed` | `TestStopFailureHandler_DisarmsGoalOnUnrecoverable` |
| AC-GSF-002 | The same holds for `oauth_org_not_allowed` and `billing_error` | same test, table-driven |
| AC-GSF-003 | An armed goal survives `rate_limit`, `overloaded`, `server_error`, `max_output_tokens`, `unknown`, and an empty type | `TestStopFailureHandler_KeepsGoalOnRecoverable` |
| AC-GSF-004 | The returned `systemMessage` names the disarm when one happened, and is unchanged when none did | assertion inside both tests |
| AC-GSF-005 | The handler returns no error on a missing goal, an unresolvable root, or an unwritable state dir | `TestStopFailureHandler_FailOpen` |
| AC-GSF-006 | `oauth_org_not_allowed` appears in the `StopFailure` matcher in both `settings.json` and the template | `grep -c oauth_org_not_allowed` on both files |
| AC-GSF-007 | The classifier decides every documented `error_type` explicitly | `TestIsUnrecoverableStopFailure` covers all ten |
| AC-GSF-008 | Build, vet, and the two affected packages pass | `go build ./... && go vet ./... && go test ./internal/goal/... ./internal/hook/...` |
