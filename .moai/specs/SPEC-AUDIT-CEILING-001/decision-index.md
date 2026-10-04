# decision-index.md — SPEC-AUDIT-CEILING-001

Authored at plan phase per `interview.decision_gate: on`
(`.moai/config/sections/interview.yaml:6`). One row per decision surfaced
during assembly that the operator has not settled. Labels:
`DECIDED` / `POLICY-COVERED` / `EVIDENCE-NEEDED` / `FOUNDER`. No row carries
a recommendation. All rows currently carry empty operator verdicts; Q4 gates
plan.md M1.

### Q1: How many delta audit rounds run automatically once the tier ceiling is reached?

Label: POLICY-COVERED
Authority anchor: `.moai/config/sections/harness.yaml` `plan_audit_ceiling_policy.auto_delta_rounds` (:82-84, value 1)
Why unresolved: Nothing is unresolved about the COUNT — the operator setting fixes it at 1. The row exists because the CLI will now enforce this number mechanically, so any future change to the key changes run behavior, and the operator should know the value is load-bearing.
Operator verdict:

### Q2: Where exactly does the refusal fire — at the tier ceiling, or only after the auto delta round is also exhausted?

Label: FOUNDER
Authority anchor: (none — the only prose that composes the two, plan-auditor.md § Retry Loop Contract :714, is outside the admissible authority register)
Why unresolved: The card and harness.yaml fix the two numbers but not their composition. The SPEC's current design (design.md §2) follows the prose: effective ceiling = tier ceiling + auto_delta_rounds. An operator preferring the harder reading (refuse immediately at the tier ceiling) changes REQ-ACE-003's trigger and the M3 tests.
Operator verdict:

### Q3: What evidence defines the round count — the exported iteration verdict files, or a dedicated CLI-maintained counter state file?

Label: FOUNDER
Authority anchor: (none — `.moai/docs/audit-artifact-convention.md` mandates per-iteration files but is outside the admissible authority register, and no config setting addresses the counter's source)
Why unresolved: Iteration files make the count auditable from the same artifacts auditors already export (design.md §1) but depend on export discipline; a state file is write-authoritative but creates a second truth that can disagree with the files. The SPEC designs for iteration files. Switching to a state file changes REQ-ACE-001 and AC-ACE-001.
Operator verdict:

### Q4: When a required backend is configured and the verdict file carries NO convergence receipt at all, should admission refuse or admit?

Label: FOUNDER
Authority anchor: (none — fail-closed is the doctrine default (auto-semantics.md §7), but that file is outside the admissible authority register, and the card text covers only the fail-present case)
Why unresolved: The card mandates blocking when a required backend FAILS (t1469/t1482). The absence case is this SPEC's extension: refusing implements fail-closed strictly but blocks trees whose auditors do not yet emit receipts; admitting preserves compatibility but leaves the requirement decorative until exporters catch up. The SPEC defaults to refuse (REQ-ACE-010) and gates plan.md M1 on this verdict.
Operator verdict:

### Q5: Which non-admitted verdicts qualify for the debt-admission outcome at the ceiling?

Label: FOUNDER
Authority anchor: (none — the card names the outcome ("부채 기록 후 진행") without fixing the eligibility predicate, and no committed authority defines it)
Why unresolved: The SPEC maps debt-admission to `must_pass_failed == 0 && blocking_count == 0` (design.md §2) — i.e., only a score-threshold or hash-binding shortfall converts to recorded debt. A looser reading (any non-must-pass failure) or a stricter one (score within a band of the threshold) are both defensible; the choice changes how often a ceiling hit ends a card versus parks it.
Operator verdict:

### Q6: Which missing gates join the auto-semantics §9 inventory in this SPEC?

Label: FOUNDER
Authority anchor: (none — the card directs "add the missing rows" without enumerating them; the research note's 30-row table is a research artifact, not authority)
Why unresolved: The SPEC selects the 11 rows of spec.md §D.3 — the plan/run/sync gate rows whose sources were verified in this session. The note's remaining rows (intake/plan interview gates, card mechanics, harness-learning applies, LSEL, goal ceilings) are unverified or belong to other subsystems. Adding more rows widens M4; trimming narrows the reconciliation.
Operator verdict:
