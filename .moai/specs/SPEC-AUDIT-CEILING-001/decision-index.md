# decision-index.md — SPEC-AUDIT-CEILING-001

Authored at plan phase per `interview.decision_gate: on`
(`.moai/config/sections/interview.yaml:6`). One row per decision surfaced
during assembly that the operator has not settled. Labels:
`DECIDED` / `POLICY-COVERED` / `EVIDENCE-NEEDED` / `FOUNDER`. No open row
carries a recommendation. Q4 gates plan.md M1, and Q2/Q3/Q5 gate plan.md M3
the same way. Every FOUNDER row's SPEC-embedded default stands while the row
is open and is kickoff-amendable: an operator answer at kickoff supersedes
the default, and the affected REQ/AC set is re-audited before run entry.
v0.4.0 refresh (2026-10-06): Q0 records the D9 operator decision; the iter3
defect dispositions D31-D36 are tabulated below; Q6 fell away with the
scope cut.

### Q0: After the iter3 STOP (cap exhausted, score regression), how does card t1500 resume?

Label: DECIDED
Authority anchor: operator decision D9 on card t1500 (2026-10-06, relayed
via the team-lead dispatch; recorded in `progress.md` §E.1 v0.4.0 entry)
Why recorded: The card's own ceiling accounting was exhausted (iter1 0.69 →
iter2 0.83 → iter3 0.81 + STOP). The operator chose resume-with-narrowed-
scope over accept-with-debt and over a cap extension.
Operator verdict: OPERATOR-DECIDED — resume with reduced scope: (1) a
CLI-side per-SPEC audit-iteration counter, (2) a single path that records
the policy outcome when the cap is reached (record debt then proceed /
split scope / hold — no questions asked), (3) codify that a required-backend
fail blocks run entry. Extra rules surfaced by iter3 (e.g. the cross-card
re-audit dedupe) are SPLIT OFF to follow-up work. The narrowed re-plan
starts a fresh audit series under the card's own ceiling accounting.

### iter3 defect dispositions (D31-D36) under the D9 scope cut

| Defect | Disposition | Where it landed in v0.4.0 |
|---|---|---|
| D31 ceiling ladder has no arm for a fully-passing verdict at/over the ceiling | KEPT — fixed in v0.4.0 | The ladder is core scope 2; posture (a) selected (admit) per D9's no-question mandate. New REQ-ACE-013 pass-through arm + design.md §2 rung 0 + §G risk 5 + AC-ACE-013 (`TestCeilingPolicyPassThrough`) |
| D32 receipt producer fed only from multi-model ConvergenceResult — single-model audits on required-resolving trees have no writer | KEPT — fixed in v0.4.0 | REQ-ACE-008 extended with the single-model producer arm (own verdict + the backend it ran; uncovered backends stay absent and refuse); design.md §3; plan.md M1; AC-ACE-008 extended |
| D33 same-SHA dedupe vs cross-card never-collapsed clause cannot both hold — ceiling blind to no-repair churn | DEFERRED — follow-up card material | Explicitly named in D9's split-off list. The contradiction is dissolved by simplification: REQ-ACE-001 reverts to the plain (SPEC id, iteration number) identity; the D22 audited-state machinery and the never-collapsed clause are removed by this scope cut (cross-layer sweep: design.md §1, AC-ACE-001, §C edge 9). The ceiling-blindness is recorded as an accepted limitation (spec.md §E + §G risk 2 + design.md §1) |
| D34 §G risk 2 describes the superseded v0.2.0 identity, contradicting REQ-ACE-001 | KEPT — fixed in v0.4.0 | §G risk 2 rewritten to the simplified identity v0.4.0 selects |
| D35 LEDGER-ACE-013-A omits the split-outcome member grep | DROPPED with its surface | Former REQ-ACE-013 (doc-text vocabulary) deleted by the scope cut; the ledger and AC-ACE-013's old text are gone; the follow-up card owns the replacement vocabulary check |
| D36 LEDGER-ACE-014-A verifies the sum, not per-name | DROPPED with its surface | Former REQ-ACE-014 (§9 inventory) deleted by the scope cut; LEDGER-ACE-014-A/B and AC-ACE-014's old text are gone |

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
Why unresolved: The card mandates blocking when a required backend FAILS (t1469/t1482). The absence case is this SPEC's extension: refusing implements fail-closed strictly but blocks trees whose auditors do not yet emit receipts; admitting preserves compatibility but leaves the requirement decorative until exporters catch up. The SPEC defaults to refuse (REQ-ACE-010) and gates plan.md M1 on this verdict. The D32 single-model producer arm narrows the exposure (a single-model audit now writes its own receipt), but a required backend the audit did not run still refuses.
Operator verdict:

### Q5: Which non-admitted verdicts qualify for the debt-admission outcome at the ceiling?

Label: FOUNDER
Authority anchor: (none — the card names the outcome ("부채 기록 후 진행") without fixing the eligibility predicate, and no committed authority defines it)
Why unresolved: The SPEC maps debt-admission to the label-only predicate (spec.md REQ-ACE-004, design.md §2): `overall_score` at or above the tier threshold, `must_pass_failed` = 0, `blocking_count` = 0, a `plan_artifact_hash` that binds the current plan artifacts, no duplicate keys, no REQ-ACE-009/010 receipt refusal, and at least one finding to enumerate — score-threshold and hash-binding shortfalls hold, and receipt refusals never convert. A looser reading (any non-must-pass failure) or a stricter one (score within a band of the threshold) are both defensible; the choice changes how often a ceiling hit ends a card versus parks it.
Operator verdict:

### Q6: Which missing gates join the auto-semantics §9 inventory in this SPEC? [FELL AWAY — D9 scope cut, 2026-10-06]

Label: (row retired — the surface it governed was deleted)
Why it fell away: Former REQ-ACE-014 (the §9 inventory expansion) and the
phase-execution.md rewrite (former REQ-ACE-013) were removed from this SPEC
by operator decision D9; the run-gate doc reconciliation is follow-up card
material. The former SPEC-embedded selection (the 11 rows of v0.3.0 §D.2)
is preserved in the v0.3.0 tree (`git show 9dd4d5c74:.moai/specs/SPEC-AUDIT-CEILING-001/spec.md`)
for the follow-up author. No M-gate depends on this row.
Operator verdict: N/A — surface deleted by the scope cut.
