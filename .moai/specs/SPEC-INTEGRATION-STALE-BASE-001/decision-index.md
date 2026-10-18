# Decision index: SPEC-INTEGRATION-STALE-BASE-001

Revision 0.1.0, authored at plan phase for card t1834 (Tier M). The decision gate is on (`.moai/config/sections/interview.yaml`, `decision_gate: on`), so this index is owed alongside the Tier artifacts. It carries no status field: the lifecycle lives in `spec.md` alone.

## How to read a row

- **Label** is one of `DECIDED`, `POLICY-COVERED`, `EVIDENCE-NEEDED`, or `FOUNDER`.
- **Class** is `product-level` or `implementation-level`. A row with no Class line is product-level.
- **Authority anchor** names a committed file and section, or states that none was found.
- **Default** and **Alternate** appear only where the published Default rule ranks the options (preserves current behavior; else single-revert undo; else smaller user-visible surface). A Default is a policy application, not a recommendation.
- An `Operator verdict:` line is left empty at authoring except for an implementation-level row that carries a Default, which takes `DEFAULT-APPLIED` at plan close.

## Summary

| Row | Subject | Label | Class | Operator verdict: line |
|---|---|---|---|---|
| Q1 | Shipped default of `max_behind_landings` (0 or a positive number) | FOUNDER | implementation-level | DEFAULT-APPLIED 2026-10-17T14:50Z lane / manager-spec |
| Q2 | Whether this repository's live `workflow.yaml` opts the gate in (`enabled: true`) | FOUNDER | implementation-level | DEFAULT-APPLIED 2026-10-17T14:50Z lane / manager-spec |
| Q3 | Whether queue promotion also runs the gate (deferred scope question; non-blocking) | FOUNDER | product-level | not required for Kickoff: promotion gating is Out of Scope (spec.md §C); a widening answer would be a new SPEC |
| Q4 | Order against card t1847 on the shared verbs `acquire` and `factory complete` (scheduling; non-blocking) | FOUNDER | implementation-level | not required for Kickoff: the SPEC is self-consistent in either order |
| Q5 | Widening the run-phase allow-list for the sync-audit iter8/iter9 repairs and bringing the SPEC wording in line (added 2026-10-19; plan.md §C, spec.md HISTORY 0.1.1); partly reverted and reduced to a single-target gate by ruling `d-20261018T172707Z-8372` (spec.md HISTORY 0.1.2) | FOUNDER | implementation-level | leader rulings `d-20261018T163916Z-ffc2` and `d-20261018T172707Z-8372` cited by id (decision board; digests not pinned here) |

## Rows

### Q1: What is the shipped default of `workflow.base_freshness_gate.max_behind_landings`?

Label: FOUNDER
Class: implementation-level
Authority anchor: none found. The card text asks for "a configuration key (default in config yaml)" and names no number; no committed setting or constitution clause decides it.
Why unresolved: a positive default admits lanes that `RunMergeStep` cause 2 refuses by construction (`internal/factory/integration_merge_step.go:332-335`), while 0 matches that check exactly; which trade the operator wants for queue-wait slack is a judgment, and the landing rate that would ground a number is not measured here.
Default: 0 (rule: the option that preserves current behavior; with 0 the gate refuses at entry exactly what the merge step already refuses inside the window)
Alternate: a small positive number, to be chosen from a measured landing rate
Operator verdict: DEFAULT-APPLIED 2026-10-17T14:50Z lane / manager-spec

### Q2: Does the live `.moai/config/sections/workflow.yaml` ship `enabled: true`?

Label: FOUNDER
Class: implementation-level
Authority anchor: none found for this key. The sibling opt-ins in the live file are operator-local choices (`branch_guard.enabled: true`, `local_main_integration.enabled: true`, `.moai/config/sections/workflow.yaml:187-195`); the template ships them off.
Why unresolved: turning the gate on in the live file changes how this repository's lanes land cards; the card asks for the mechanism and does not say whether to arm it here.
Default: false (rule: the option that preserves current behavior)
Alternate: true in the live file only, the template staying false
Operator verdict: DEFAULT-APPLIED 2026-10-17T14:50Z lane / manager-spec

### Q3: Should the freshness gate also run when the queue promotes a waiting ticket?

Label: FOUNDER
Class: product-level
Authority anchor: none found. The existing promotion gate (`CandidateGrantGate`, `internal/factory/integration_window_ops.go:451-465`) takes only a card id, and the staleness that arises while a lane waits is today governed by merge cause 2.
Why unresolved: gating promotion would need the gate to resolve a card's worktree inside the window-queue package and would drop a waiting lane from the queue when main moves, which changes the queue's observable behavior for every waiter; that is a judgment about product behavior, not an implementation detail.
Operator verdict: not required for Kickoff (deferred scope question; the SPEC is self-consistent either way)

### Q4: In what order do t1834 and t1847 land, given both edit `acquire` and `factory complete`?

Label: FOUNDER
Class: implementation-level
Authority anchor: none found. t1846's SPEC records the same kind of request for t1831 (`SPEC-LANDING-SURFACE-RULES-001/spec.md` §A.5, in tree `.moai/worktrees/t1846`, not committed to this tree); t1847's SPEC does not exist in any tree read.
Why unresolved: ordering is a leader scheduling decision; this plan only keeps its edits to one added call at each verb so that either order merges cleanly.
Operator verdict: not required for Kickoff (scheduling matter for the leader)

### Q5: Do the repairs forced by the sync audits (iter8 F2, iter9 G1/G2) stay in card t1834, with the allow-list widened and the SPEC wording amended?

Label: FOUNDER
Class: implementation-level
Authority anchor: none verifiable in the committed tree. The decision is a decision-board ruling, `d-20261018T163916Z-ffc2`, relayed to this author by the orchestrator; no digest prefix and no verbatim board line was available, so the citation cannot be pinned and the row does not claim `DECIDED` (a board citation that cannot be pinned routes to `FOUNDER`).
Why unresolved: a second route existed — move the three repairs (`internal/factory/integration_remeasure.go` with its gitenv test, `internal/cli/integration_remeasure.go`) into a follow-up card and leave this card's allow-list at the plan-time 15 paths — and this author cannot show from a committed source that the ruling excludes it. What the amendment records is the ruling as relayed: the repairs stay in t1834, the allow-list gains those three paths plus the sync-phase `CHANGELOG.md`, and requirements REQ-ISB-004, -005, -006 and -015 are amended in place without a new REQ or AC identifier (Tier M budget of 16 held; the Tier M file-count guidance of 5 to 15 is exceeded at 19 paths in amendment 0.1.1 and at 18 plus 3 `.moai` entries after amendment 0.1.2, and the tier is left unchanged, plan.md A.1).
Operator verdict: leader ruling `d-20261018T163916Z-ffc2` (decision board), recorded by id on 2026-10-19 as relayed by the orchestrator; not a `DEFAULT-APPLIED` entry and not verified against the board by this author.

Reduction ruling (2026-10-19, amendment 0.1.2): leader ruling `d-20261018T172707Z-8372`, relayed by the orchestrator and cited by id only (no digest prefix and no verbatim board line available, so it is not pinned and does not claim `DECIDED`), reduced the gate to ONE target-branch resolution, the configured integration target. The widening recorded above was therefore partly reverted: `f932d66dee` (`moai integration remeasure --branch`) and `9fc118b44e` (the refusal remedy naming `remeasure --branch <branch>`) were reverted by `e8a273aed9` and `87e8c01b72`, so `internal/cli/integration_remeasure.go` leaves the allow-list; `e81bf9429e` (common-root target read) and `087a270e48` (git-environment scrub) stay. The fix `1c05af5837` (audit iter10 H1/H2) makes the gate measure the configured target only, refuses an explicit `acquire --branch` naming another branch (same sentinel and exit code 16, with the gate off one advisory line), and gives the gate and the re-measure the same `refs/heads/<branch>` spelling; it adds `internal/factory/base_freshness_target_test.go`. The allow-list is now 18 paths plus 3 `.moai` entries against the Tier M guide of 5 to 15; whether that changes the tier is the plan-auditor's and the leader's question, and this row records none. Operator verdict for this part: leader ruling `d-20261018T172707Z-8372` (decision board), recorded by id on 2026-10-19; not verified against the board by this author.

## Routing

1. **No row blocks the Kickoff.** Q3 is a deferred scope question and Q4 a scheduling matter; the SPEC is self-consistent whichever way either is answered, and a widening answer to Q3 would be a new SPEC.
2. **DEFAULT-APPLIED at plan close:** Q1 and Q2, both implementation-level with a ranked Default, so the audited hash covers them.
3. **Q5 was added after plan close** (2026-10-19, amendment 0.1.1, then reduced by amendment 0.1.2 under ruling `d-20261018T172707Z-8372`) and is outside the audited plan hash; it records a run-phase scope change, not a plan-time decision, and blocks nothing.
