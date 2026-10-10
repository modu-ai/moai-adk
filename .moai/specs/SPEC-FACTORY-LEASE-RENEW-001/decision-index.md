# Decision index: SPEC-FACTORY-LEASE-RENEW-001

Revision 0.1.0, authored at plan phase for card t1663 (Tier M). The decision gate is on (`.moai/config/sections/interview.yaml`, `interview.decision_gate: on`), so this index is owed alongside the Tier artifacts. It carries no status field: the lifecycle lives in `spec.md` alone.

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
| Q1 | Values and names of the card lease duration and the watchdog renewal period | FOUNDER | product-level | empty |
| Q2 | Refusal or clamping when a lease key violates its bound | FOUNDER | product-level | empty |
| Q3 | Renewal of an expired own lease: refuse with no write, or reap then refuse | FOUNDER | product-level | empty |
| Q4 | Repository-local overrides survive a template sync | POLICY-COVERED | implementation-level | not required (POLICY-COVERED) |
| Q5 | Overlap with the draft SPEC-FACTORY-RECORD-REPAIR-001 | FOUNDER | implementation-level | DEFAULT-APPLIED 2026-10-10T09:15:17Z lane-17 / manager-spec |

## Routing

1. **Open FOUNDER rows with an empty `Operator verdict:` line:** Q1, Q2, and Q3. The autonomous Kickoff stays blocked until the operator writes these lines.
2. **DEFAULT-APPLIED at plan close:** Q5. It is implementation-level and carries a ranked Default.
3. **POLICY-COVERED, no verdict needed:** Q4.
4. **DECIDED:** none. No committed authority decides these questions.

## Rows

### Q1: What values and names do the two lease keys carry: the card lease duration, and the watchdog renewal period?

Label: FOUNDER
Class: product-level (the two keys are added to the template's `workflow.yaml`, which is a template default; the renewal period sets how often a lane's standing wake fires)
Authority anchor: none. At plan time the completed SPEC-FACTORY-ATOMIC-LEASE-001 was searched for renew and expired terms, and none of its matches decides a renewal period or a cadence. SPEC-FACTORY-RECORD-REPAIR-001 Q18 records a verdict on the duration key name, cited to a board record without a digest, so it is context and not an anchor. Decision d-20261010T064450Z-68b2 removes the 20-minute cadence; it is a board record without a digest and is context only.
Why unresolved: REQ-FLR-007 fixes the bound at one half of the card lease duration, and REQ-FLR-012 fixes the standing wake to follow the renewal period. The value inside the bound is a product choice, and the names are also open for confirmation. Values (the default duration is 15 minutes, so the whole-minute maximum is 7 minutes):
- (a) 5 minutes, the value that spec.md REQ-FLR-009 carries.
- (b) 7 minutes, the whole-minute maximum at the default duration.
- (c) another whole-minute value from 1 to 7 minutes at the default duration.
The consequences are stated without a ranking. A shorter period means more standing-wake turns for each awake lane, and a wake that falls after the prompt-cache window re-writes the session prefix (`.claude/rules/moai/workflow/cache-aware-execution.md` directive 9). The late-pass margin is the duration minus the period: 10 minutes at 5 minutes, and 8 minutes at 7 minutes. Names: spec.md carries `workflow.factory_lease.duration` (the draft's name) and `workflow.factory_lease.renew_interval` (new). The operator confirms both names or names other keys.
Default: none (the published Default rule ranks neither the values nor the names: no option preserves current behaviour, because no renewal period exists today).
Operator verdict:

### Q2: When a lease key violates its bound, does the factory refuse to open, or clamp the value to the bound and warn?

Label: FOUNDER
Class: product-level (a refusal to open the factory is visible to any operator whose workflow configuration sets either key outside its bound)
Authority anchor: none.
Why unresolved: REQ-FLR-006 and REQ-FLR-007 refuse to open the factory and name the key and the bound. The other option clamps the effective value to the bound and records a warning at open. The first makes a misconfiguration visible at once. The second keeps the factory open, and the effective value differs from the configured one except for the warning. The published Default rule does not rank them, because its third criterion (the smaller user-visible surface) turns on whether a refusal counts as user-visible surface.
Options:
- (a) refuse to open, naming the key and the bound (spec.md REQ-FLR-006 and REQ-FLR-007).
- (b) clamp the effective value to the bound, and record a warning at open.
Default: none (see Why unresolved).
Operator verdict:

### Q3: When the holder renews a lease that has already expired, does the renewal refuse with no write, or reap the lease and then refuse?

Label: FOUNDER
Class: product-level (the renew verb's response to an expired lease is user-visible, and the current behaviour writes)
Authority anchor: none. SPEC-FACTORY-RECORD-REPAIR-001 Q9 cites an operator direction in a board record without a digest, so it is context. The completed SPEC-FACTORY-ATOMIC-LEASE-001 matches its renew and expired terms only for the lock that the lease transactions take, and those matches do not decide this question.
Why unresolved: The current `RenewLease` reaps an expired lease, which bumps the card version and appends a lease event, and then refuses with the expired-lease class (`internal/homestate/card_transition.go:441-456`). spec.md REQ-FLR-004 states refusal with no write. If the second option is selected, REQ-FLR-004 and AC-FLR-004 are reworded in the same revision.
Options:
- (a) refuse with no write (spec.md REQ-FLR-004 as written).
- (b) reap the lease, then refuse (the current behaviour).
Default: (b) (rule: the option that preserves current behaviour)
Alternate: (a)
Note: The Default line records the published ranking. A product-level row takes no automatic application, so the `Operator verdict:` line decides. This row does not change the wording of spec.md.
Operator verdict:

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
Why unresolved: The draft carries REQ-FRR-001 to REQ-FRR-005, REQ-FRR-009, and REQ-FRR-010, which cover the same renew verb, duration key, watchdog renewal, and queue-claim independence as this SPEC. Two drafts that describe one behaviour can both be implemented, which duplicates the run-phase work and its verdicts.
Options:
- (a) leave the draft unchanged. Both drafts stand, and each card decides its own retirement.
- (b) retire or narrow the overlapping requirements of the draft under its own card, before this SPEC's run phase.
Default: (a) (rule: the option that preserves current behaviour; the draft is unchanged)
Alternate: (b)
Operator verdict: DEFAULT-APPLIED 2026-10-10T09:15:17Z lane-17 / manager-spec
