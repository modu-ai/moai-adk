# Decision index: SPEC-FACTORY-LEASE-RENEW-001

Revision 0.1.1, authored at plan phase for card t1663 (Tier M), and revised after plan audit iteration 1 (`.moai/reports/t1663/plan-audit.md`). The decision gate is on (`.moai/config/sections/interview.yaml`, `interview.decision_gate: on`), so this index is owed alongside the Tier artifacts. It carries no status field: the lifecycle lives in `spec.md` alone.

## How to read a row

- **Label** is one of `DECIDED`, `POLICY-COVERED`, `EVIDENCE-NEEDED`, or `FOUNDER`.
- **Class** is `product-level` or `implementation-level`. A row with no Class line is product-level.
- **Authority anchor** names a committed file and section, or states that none was found. A board record carries no digest, so a board record alone does not make a row DECIDED. Each board citation in this index is context only.
- **Why unresolved** states what the operator must settle.
- **Default** and **Alternate** appear only where the published Default rule ranks the options. A Default is a policy application, not a recommendation.
- The `Operator verdict:` line is left empty at authoring; the operator writes it. Only an implementation-level row that carries a Default takes `DEFAULT-APPLIED`, and only at plan close.

An open FOUNDER row blocks the autonomous Kickoff until its `Operator verdict:` line is written.

## Summary

| Row | Subject | Label | Class | Operator verdict: line |
|---|---|---|---|---|
| Q1 | Values, names, and armed form of the card lease duration and the watchdog renewal period | FOUNDER | product-level | 영실이 판단(운영자 위임) 10-10 — 본문 확정; decision d-20261010T095818Z-c501; armed form (sub-question d) covered by decision d-20261010T095818Z-c501 |
| Q2 | Refusal or clamping when a lease key violates its bound | FOUNDER | product-level | 영실이 판단(운영자 위임) 10-10 — 본문 확정; decision d-20261010T095818Z-c501 |
| Q3 | Renewal of an expired own lease: refuse with no write, or reap then refuse | FOUNDER | product-level | 영실이 판단(운영자 위임) 10-10 — 본문 확정; decision d-20261010T095731Z-5ddc |
| Q4 | Repository-local overrides survive a template sync | POLICY-COVERED | implementation-level | not required (POLICY-COVERED) |
| Q5 | Overlap with the draft SPEC-FACTORY-RECORD-REPAIR-001 | FOUNDER | implementation-level | DEFAULT-APPLIED 2026-10-10T09:40:46Z lane-17 / manager-spec |
| Q6 | Stall window reference: the previous snapshot, or the last observed channel change | FOUNDER | product-level | 영실이 판단(운영자 위임) 10-10 — 본문 확정; decision d-20261010T095731Z-5ddc |

## Routing

1. **FOUNDER rows with recorded operator verdicts:** Q1, Q2, Q3, and Q6 carry recorded operator verdicts (decision ids in the Rows section), so no FOUNDER row has an empty `Operator verdict:` line.
2. **DEFAULT-APPLIED at plan close:** Q5. It is implementation-level and carries a ranked Default.
3. **POLICY-COVERED, no verdict needed:** Q4.
4. **DECIDED:** none. No committed authority decides these questions.

## Rows

### Q1: What values, names, and armed form do the two lease keys carry: the card lease duration, the watchdog renewal period, and the start minute of the armed wake?

Label: FOUNDER
Class: product-level (the two keys are added to the template's `workflow.yaml`, which is a template default; the renewal period sets how often a lane's standing wake fires; the armed form sets the minutes at which it fires)
Authority anchor: none. At plan time the completed SPEC-FACTORY-ATOMIC-LEASE-001 was searched for renew and expired terms, and none of its matches decides a renewal period or a cadence. SPEC-FACTORY-RECORD-REPAIR-001 Q18 records a verdict on the duration key name, cited to a board record without a digest, so it is context and not an anchor. Decision d-20261010T064450Z-68b2 removes the 20-minute cadence; it is a board record without a digest and is context only.
Why unresolved: REQ-FLR-007 fixes the bound at one half of the card lease duration. REQ-FLR-012 fixes the standing wake to a minute step from minute 0. The value inside the bound is a product choice, and the names are also open for confirmation. Values (the default duration is 15 minutes, so the whole-minute maximum is 7 minutes):
- (a) 5 minutes, the value that spec.md REQ-FLR-009 carries.
- (b) 7 minutes, the whole-minute maximum at the default duration.
- (c) another whole-minute value from 1 to 7 minutes at the default duration.
The consequences are stated without a ranking. A shorter period means more standing-wake turns for each awake lane, and a wake that falls after the prompt-cache window re-writes the session prefix (`.claude/rules/moai/workflow/cache-aware-execution.md` directive 9). The late-pass margin is the duration minus the period: 10 minutes at 5 minutes, and 8 minutes at 7 minutes. Names: spec.md carries `workflow.factory_lease.duration` (the draft's name) and `workflow.factory_lease.renew_interval` (new). The operator confirms both names or names other keys.
Armed form (sub-question d): the wake fires at minute 0 with `*/N` (as REQ-FLR-012 is written). That form has the gap property for every N with no per-value check, and it fires together with other top-of-hour schedules. An off-minute start, such as the draft's `7,27,47` literal, is possible only with a gap check for the value in force (plan B6). The operator confirms `*/N` or names an off-minute start.
Default: none. The published Default rule ranks no option for (a) to (c), because no renewal period exists today. For (d), the rule's first criterion would select the off-minute start that the current literal uses, but that start fails the gap property for some values in force, so no ranking is recorded here; the operator verdict decides.
Operator verdict: 영실이 판단(운영자 위임) 10-10 — 본문 확정; decision d-20261010T095818Z-c501; armed form (sub-question d) covered by decision d-20261010T095818Z-c501

### Q2: When a lease key violates its bound, does the factory refuse to open, or clamp the value to the bound and warn?

Label: FOUNDER
Class: product-level (a refusal to open the factory is visible to any operator whose workflow configuration sets either key outside its bound)
Authority anchor: none.
Why unresolved: REQ-FLR-006 and REQ-FLR-007 refuse to open the factory and name the key and the bound. The other option clamps the effective value to the bound and records a warning at open. The first makes a misconfiguration visible at once. The second keeps the factory open, and the effective value differs from the configured one except for the warning. The published Default rule does not rank them, because its third criterion (the smaller user-visible surface) turns on whether a refusal counts as user-visible surface.
Options:
- (a) refuse to open, naming the key and the bound (spec.md REQ-FLR-006 and REQ-FLR-007).
- (b) clamp the effective value to the bound, and record a warning at open.
Default: none (see Why unresolved).
Operator verdict: 영실이 판단(운영자 위임) 10-10 — 본문 확정; decision d-20261010T095818Z-c501

### Q3: When the holder renews a lease that has already expired, does the renewal refuse with no write, or reap the lease and then refuse?

Label: FOUNDER
Class: product-level (the renew verb's response to an expired lease is user-visible, and the current behaviour writes)
Authority anchor: none. SPEC-FACTORY-RECORD-REPAIR-001 Q9 cites an operator direction in a board record without a digest, so it is context. The completed SPEC-FACTORY-ATOMIC-LEASE-001 matches its renew and expired terms only for the lock that the lease transactions take, and those matches do not decide this question.
Why unresolved: The current `RenewLease` reaps an expired lease, which bumps the card version and appends a lease event, and then refuses with the expired-lease class (`internal/homestate/card_transition.go:441-456`). spec.md REQ-FLR-004 states refusal with no write. If the second option is selected, REQ-FLR-004 and AC-FLR-004 are reworded in the same revision, and the M2 rewrite of the reaped-state assertion in `TestRenewLeaseRefusalsAndExpiry` is not made (plan M2 item 5).
Options:
- (a) refuse with no write (spec.md REQ-FLR-004 as written).
- (b) reap the lease, then refuse (the current behaviour).
Default: (b) (rule: the option that preserves current behaviour)
Alternate: (a)
Note: The Default line records the published ranking. A product-level row takes no automatic application, so the `Operator verdict:` line decides. This row does not change the wording of spec.md.
Operator verdict: 영실이 판단(운영자 위임) 10-10 — 본문 확정; decision d-20261010T095731Z-5ddc

### Q4: Do repository-local overrides survive a template sync?

Label: POLICY-COVERED
Class: implementation-level
Authority anchor: `.moai/config/sections/workflow.yaml`, the tracked rendered file, sections `branch_guard` (`enabled: true`, `deny_commits_on: [main]`) and `commit_identity_guard` (`enabled: true`). The template at `internal/template/templates/.moai/config/sections/workflow.yaml` ships `branch_guard.enabled: false`, `branch_guard.deny_commits_on: []`, and `commit_identity_guard.enabled: false`.
Why unresolved: nothing under the setting. The QG-4 comparison in acceptance.md compares the rendered and template files by key, so a sync that overwrote these values would fail it.
Operator verdict: not required (POLICY-COVERED)

### Q5: How is the overlap with the draft SPEC-FACTORY-RECORD-REPAIR-001 settled?

Label: FOUNDER
Class: implementation-level (the draft is SPEC bookkeeping; no shipped command changes with either option)
Authority anchor: none.
Why unresolved: The draft carries REQ-FRR-001 to REQ-FRR-005, REQ-FRR-009, and REQ-FRR-010, which cover the same renew verb, duration key, watchdog renewal, and queue-claim independence as this SPEC. Two drafts that describe one behaviour can both be implemented, which duplicates the run-phase work and its verdicts. The draft's REQ-FRR-005 also keeps the 20-minute cadence literal (`7,27,47 * * * *`) that REQ-FLR-012 replaces, so the two drafts contradict each other on the cadence (audit FLR-O2). A ruling that retires or narrows the draft's overlapping requirements would also retire that contradiction.
Options:
- (a) leave the draft unchanged. Both drafts stand, and each card decides its own retirement.
- (b) retire or narrow the overlapping requirements of the draft under its own card, before this SPEC's run phase.
Default: (a) (rule: the option that preserves current behaviour; the draft is unchanged)
Alternate: (b)
Operator verdict: DEFAULT-APPLIED 2026-10-10T09:40:46Z lane-17 / manager-spec

### Q6: Is the stall window measured from the previous snapshot, or from the last observed channel change?

Label: FOUNDER
Class: product-level (the stall verdict is the signal the leader reads from the watchdog, and its timing changes in the shipped watchdog procedure and in the snapshot format)
Authority anchor: none decides it. `.claude/rules/moai/workflow/auto-semantics.md` §3 states the rule as "no change across the N-minute window", and §4 implements it as a comparison with the previous snapshot. The two statements agree only while the wake period is N or more.
Why unresolved: At the renewal period of REQ-FLR-009 (5 minutes), and at any period below N = 15 minutes, a comparison with the previous snapshot never gives a stall verdict, because the previous observation is always younger than N. The options:
- (a) keep the comparison with the previous snapshot (the text as it stands). At a period below N the stall verdict cannot be given at all.
- (b) measure the N-minute window from the last observed channel change, with one additive snapshot field, `last_change_at` (spec.md REQ-FLR-013). At a period of N or more the verdict on every pass equals option (a)'s (a derivation from the rule text; AC-FLR-015 checks one instance, at 20 minutes). Below N, a stall is reported no later than N plus one period after the last change.
Default: (a) (rule: the option that preserves current behaviour)
Alternate: (b)
Note: The Default line records the published ranking. A product-level row takes no automatic application, so the `Operator verdict:` line decides. The SPEC body states option (b). Option (a) is the rule-ranked option, but at the adopted period it leaves the verdict unreachable, which is why the body does not follow the ranking.
Operator verdict: 영실이 판단(운영자 위임) 10-10 — 본문 확정; decision d-20261010T095731Z-5ddc
