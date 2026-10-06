# verdict — t1481 (SPEC-FACTORY-DECISION-AUTO-001)

verdict: PASS-WITH-DEBT

## Claim

Factory decision automation landed on `WT-decision-automation`: decision board, shared verdict
admission predicate, audit kickoff decider (T8a), ceiling policy, FOUNDER defaults, wait recheck
doctrine, bind cache, degraded-inbox surfacing. 21 of 25 ACs PASS, 5 PARTIAL (recorded debts).
Final sync audit: PASS-WITH-DEBT 88.

## Evidence

- Plan audit: iter4 PASS 0.885 (`.moai/reports/t1481/plan-audit-iter4.md`).
- Sync audits: FAIL 72 → FAIL 78 → FAIL 80 → PASS-WITH-DEBT 88
  (`.moai/reports/t1481/sync-audit.md`, `-2.md`, `-3.md`, `-4.md`).
- Last test runs (-race): `ok github.com/modu-ai/moai-adk/internal/auditverdict 1.129s`;
  `ok github.com/modu-ai/moai-adk/internal/homestate 89.062s` (whole package);
  `ok github.com/modu-ai/moai-adk/internal/cli 14.316s` (selectors
  `TestFDA_|TestDecisionCmd|TestFR_AC015|TestSD_AC016`); golangci-lint v2.1.6 `0 issues.`
- AC matrix with commands and verbatim output: `.moai/specs/SPEC-FACTORY-DECISION-AUTO-001/progress.md` §E.2.

## Baseline-attribution

Measured in this card tree at `b8cd707d0` (code unchanged by the close commit). Pre-flight base:
`cb8b7e03a`.

## Gaps

- No codex audit receipt (this tree has no local codex gate config); codex card review inconclusive
  (blank output).
- M0 load measurements not run; sync-audit-4dim script change not executed; ceiling procedure has no
  executable fixture.
- The repository-wide test verdict belongs to CI on develop (pending).

## Residual-risk

- Land LAST among v3.2.0 cards: older plan verdicts lacking `must_pass_failed` / `blocking_count` /
  `plan_artifact_hash` are refused at Kickoff/T7; in-flight cards need re-audit.
- D1: the audit decider is inert until a writer emits `audit_ready: true`.
- D3: the queue hold read is not atomic across stores (→ t1458).
- D4–D6, F1 case-variant ordering, convention "exactly once" wording; D6 delta plan re-audit owed.
- Expected textual conflict with t1480 in `internal/homestate/card_transition.go`.
