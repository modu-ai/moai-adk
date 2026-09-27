---
description: "Contract-mode gate rewiring — which gates a signed SPEC contract replaces, the autonomous Kickoff check, revocation, and the one-pass lifecycle"
paths: "**/.moai/specs/**/contract.yaml,**/.moai/specs/**/kickoff-receipt.json,**/contract-autonomy.md"
---

# Contract Autonomy — Gate Rewiring in Contract Mode

A signed SPEC contract (`.moai/specs/<SPEC-ID>/contract.yaml`) records, once, the decisions a
run would otherwise stop to ask for: the acceptance binding, the invariants, the ownership
globs, the approach, the permitted actions, and the escalation triggers. In contract mode the
orchestrator reads those decisions from the signed contract instead of asking for them turn by
turn. This file is the single source for that rewiring; every contract-mode block elsewhere is a
one-sentence pointer here.

## Scope and activation

This file applies only when `workflow.autonomy.mode: contract`. When the mode is `guided`, the
key is absent, or the value is invalid, the effective mode is `guided` and every gate works
exactly as written outside the contract-mode blocks.

- Read the mode and the configured Kickoff decider with `moai contract kickoff-check <SPEC-ID>
  --card <card> --json` (its `mode` and `decider` fields). That command resolves the
  configuration through the same reader `moai contract sign` uses, so an invalid value falls
  back to `guided` and `human` in one place.
- `workflow.autonomy.mode` is a different axis from the permission-bundle autonomy tier
  (`MOAI_AUTONOMY_TIER`). The tier widens what a session may do without a prompt; the mode
  decides whether a signed contract stands in for the run-time gates. Neither implies the
  other.

## The signing gate

In contract mode the plan→run gate is `moai contract kickoff-check <SPEC-ID> --card <card>`
exiting 0. The orchestrator does not emit the Implementation Kickoff Approval
`AskUserQuestion`; the human decision moves to the moment of signing.

When the configured decider is `human`, or autonomous Kickoff is inactive:

1. Report the signing command (`moai contract sign <SPEC-ID>`) and a short contract summary,
   then close the turn.
2. The operator signs at an interactive terminal (a typed confirmation; agent sessions are
   refused).
3. On the next turn, run `kickoff-check` again; exit 0 enters the run phase.

A kickoff decision whose outcome is `reject` or `human` is treated the same way: the signature is
refused and a human decision required — the orchestrator does not call the receipt signing path
and follows the human signing procedure above. The two values differ only in the record they
leave.

When `kickoff-check` exits non-zero, do not enter the run phase and do not fall back to guided
mode silently: report the reason codes it printed as an escalation.

## Equivalence clause (human signature only)

Every MoAI rule, skill, agent, or output style that presupposes "Implementation Kickoff Approval
was obtained" is satisfied in contract mode by a human signature of the SPEC's contract — a
signature recorded with `signer_kind: human` and `method: interactive-tty`, for which
`kickoff-check` exits 0. The operator's approval is not removed; it is given once, at signing,
with the contract in front of them.

This equivalence covers the human signature only. A signature produced from a kickoff receipt
is not equivalent by this clause; it can stand in for the gate only through the activation
conditions of § Autonomous Kickoff.

## Gate disposition

| Gate | guided | contract |
|---|---|---|
| Implementation Kickoff Approval | `AskUserQuestion` at plan→run | signature + `kickoff-check` exit 0 |
| Socratic interview (after signing) | ask when intent is unclear | satisfied by the signature; the plan-phase interview is unchanged |
| Approach approval | ask before non-trivial code | the contract's `approach` field, approved once at signing |
| Wait for assumption confirmation | list and wait | record the assumption in `progress.md` and proceed; escalate when it contradicts the contract |
| plan-audit FAIL retry and escalation | ask after the retry cap | repair and re-audit automatically up to the cap, then escalate |
| SPEC quality gate (WARNING/FAIL) | ask | repair automatically up to the same cap, then escalate |
| Documentation scope approval (`gate-sync-2`) | ask | not asked |
| Sync next-step and branch questions | ask | not asked |

The retry cap is the contract's `budget.audit_retries`, or
`workflow.autonomy.escalation.budget_default.audit_retries` when the contract has no budget.

An assumption or a new ambiguity that contradicts the contract's `acceptance`, `invariants`,
`ownership`, or `approach` stops the work and escalates under the matching `escalate_on` class.

## Gates a contract never touches

- The question-channel monopoly: every question to the user goes through `AskUserQuestion`,
  asked by the orchestrator only.
- The question after three failed CI auto-fix iterations.
- The context-limit `/clear` handoff.
- The constitution and its amendment procedure.
- The sync-auditor must-pass criteria.
- Any action on `main` or a release branch — a contract cannot permit it.
- Card selection, which stays the operator's act.
- The goal turn ceiling.
- Confirmation before a destructive command.
- The Report-Before-Ask gate for every question that is still asked.

Autonomous approval and revocation record decisions; neither asks the user anything.

## Escalation routing

An escalation is a report, not a question. The orchestrator writes the Report-Before-Ask
findings report and stops; it does not open an `AskUserQuestion` round in place of the gate the
contract removed. A card is needs-decision when it has an open escalation record under
`.moai/reports/<card>/escalation/`; the record format belongs to the escalation detector.

Route to an escalation, instead of a question, when:

- `kickoff-check` exits non-zero;
- the plan-audit or quality-gate retry cap is reached;
- an assumption or ambiguity contradicts the contract;
- a sync failure decision point is reached (quality gate, test failure, compatibility break,
  critical security finding, local CI mirror failure);
- the contract is revoked.

## One-pass lifecycle

A signed contract runs as one pass through seven stages. A stage advances only when its evidence
is recorded and the card has no open escalation record — including the record a revocation
writes, which `kickoff-check` reports as `revoked`, so the orchestrator runs `kickoff-check` at
every stage boundary.

| Stage | Input | Evidence to record | Advance condition |
|---|---|---|---|
| Discovery | the contract's `reobserve` list | reobserve each entry; the reproduction command and its output | evidence recorded and no open escalation record |
| RED | the acceptance criteria | the commit SHA of the failing test committed before the implementation, and its verbatim failing output | evidence recorded and no open escalation record |
| GREEN | the RED tests | the same tests' passing output | evidence recorded and no open escalation record |
| Qualification | the changed scope | commands and outputs of the scoped tests, lint, coverage, the mutation check, and the second-model review (`audit_multi`, model from `review.second_model`) | evidence recorded and no open escalation record |
| Closure | the qualification evidence | the status transition and `.moai/reports/<card>/verdict.md` | evidence recorded and no open escalation record |
| Integration | the integration branch | the re-measurement output on the merged tree after absorbing it | evidence recorded and no open escalation record |
| Push | the integrated branch | the push under the `push-develop` lease | evidence recorded and no open escalation record |

The Push stage is inactive until the stop-before-push on a missing second review is in place;
until then the lifecycle ends at Integration and the push stays a human act. When active, a push
requires `push-develop` in the contract's `actions`, `workflow.autonomy.contract.push_develop`
true, and the `push-develop` lease held.

## Guided mode

In `guided` mode this entire file is inactive. The contract-mode blocks elsewhere open with the
condition `workflow.autonomy.mode: contract` and do not apply; every gate stays as written.

## Autonomous Kickoff

This section is filled in when the autonomous Kickoff activation conditions are met. Until then
only a human signature passes `kickoff-check`.

## Revocation

`moai contract revoke <card> --spec <SPEC-ID>` withdraws a signed contract, whatever kind of
signature it carries. It appends a revoke event to the contract store and writes one escalation
record of kind `revoke`; after that `kickoff-check` reports `revoked` for that signature until a
new signature is given.

A run in progress stops at the next stage boundary, when the boundary's `kickoff-check` reports
the revocation. Revocation never deletes a worktree, deletes or renames a branch, pushes, changes
the queue, edits the contract, the signature, or any SPEC document, kills a process, or runs a git
write command.
