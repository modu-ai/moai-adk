---
description: >
  Factory Mode contract for the --factory / -f entry switch on the session
  launchers. Seeds a one-session plan -> run -> verify -> sync chain that
  extends the full-pipeline contract with a plan-phase chain head and a
  verify exit gate at run-phase exit.
user-invocable: false
metadata:
  version: "0.1.0"
  category: "workflow"
  status: "active"
  tags: "factory, launcher, pipeline, verify-gate, chain"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["factory", "chain", "verify gate"]
  agents: ["manager-spec", "manager-develop", "manager-docs"]
  phases: ["plan", "run", "sync"]
---

# Factory Mode

Factory Mode is an **entry switch**, not a subcommand and not a runtime. Passing `--factory` (or `-f`) to a session launcher seeds a session whose orchestrator drives a `plan -> run -> verify -> sync` chain end to end, powered by an armed goal preset evaluated by the existing `stop-goal` Stop-hook evaluator. No new hook, no new evaluator, no new daemon, and no new subcommand is introduced.

An optional SPEC identifier may follow the switch. When it is supplied, the chain targets that SPEC; when it is absent, the chain begins at plan-phase from the operator's first prompt.

## Contract

The `factory` pipeline contract **extends** the `full-pipeline` contract defined in `workflows/moai.md` § run→sync chaining policy. It inherits the run→sync auto-chain and the clause preserving the sync-internal gates verbatim, and adds exactly two deltas:

1. **A plan-phase chain head** — the chain starts at plan rather than at an explicitly-invoked phase.
2. **A verify exit gate** — a security review positioned at the exit of run-phase, specified in `workflows/run.md` § Verify Exit Gate.

There is no second chaining mechanism. Every other property of how phases chain is the inherited contract, unmodified.

## The four stages

| Stage | What runs | Notes |
|---|---|---|
| **plan** | SPEC authoring and the independent plan audit | the chain head; the audit gate is unchanged |
| **run** | the configured implementation cycle to acceptance-criterion convergence | unchanged |
| **verify** | `/moai review --security --deep --repo` | the exit gate of run-phase, not a stage of sync; outcome routing and the rung attribute are specified in `workflows/run.md` § Verify Exit Gate |
| **sync** | documentation, changelog, and the phase close | entered via the inherited auto-chain |

## Human gates

Four human gates fire across a factory chain. Exactly one is added by this contract; the other three are inherited unchanged. No fifth gate exists.

| # | Gate | Origin | Boundary |
|---|---|---|---|
| 1 | Implementation Kickoff Approval | inherited (plan→run) | the chain does not enter run-phase until it is cleared, and the goal preset is armed only afterwards, alongside the work it drives |
| 2 | the verify CRITICAL/HIGH decision | **added by this contract** | an orchestrator-issued `AskUserQuestion` round at the run exit gate — the sole HUMAN GATE Factory Mode introduces |
| 3 | `gate-sync-1` (pre-sync quality) | inherited via the extended contract | fires unchanged inside the chained sync phase |
| 4 | `gate-sync-2` (documentation scope) | inherited via the extended contract | fires unchanged inside the chained sync phase |

All four are orchestrator-issued question rounds, never Stop-hook blocks. That distinction matters for the block-cap note below: raising the block cap cannot skip any of them.

## Backend exclusion

Factory Mode is rejected on the mixed-backend launcher (`moai cg`) with the sentinel `FACTORY_MODE_UNSUPPORTED_BACKEND`, and no session is launched. That launcher runs a leader on one backend and teammates on another, which contradicts the one-session / one-backend / one-chain premise the chain rests on — the verify stage would run under an indeterminate backend. The rejection is deliberate, not a gap to be adapted around.

## State record

A factory session carries a session-keyed record under `.moai/state/factory/`:

| Field | Written by | Meaning |
|---|---|---|
| `session_id` | launcher | the session the record belongs to |
| `spec_id` | launcher | the targeted SPEC identifier, or empty when the chain heads at plan |
| `backend` | launcher | which backend the session runs on |
| `entered_at` | launcher | when Factory Mode was entered |
| `deepscan_dir` | orchestrator | the results directory the verify stage produced |
| `verify_rung` | orchestrator | the rigor rung of the verify result — `PRIMARY`, `FALLBACK`, or `DEGRADED` |
| `verify_reentries` | orchestrator | how many verify re-entries the chain has consumed |

The record is written **best-effort and fail-open**: a write failure never blocks a launch. The chain then degrades to a session with no record, which is a session with an unusable dedup input — and an unusable input resolves toward running the check, which is the safe direction.

The three orchestrator-written fields are filled in independently as the chain progresses, so a record carrying `deepscan_dir` but not `verify_rung` is reachable. This is why `verify_rung` is read as an **allow-list, never a deny-list**: sync-phase suppression of the security analysis requires `verify_rung` to be **recorded and equal to** `PRIMARY` or `FALLBACK`. Every other value — `DEGRADED`, an unrecognized string, an empty string, or a field never written — yields no suppression. A predicate whose behavior on a missing field is *suppress* is wrong by construction.

## Block-cap blast radius

Factory Mode raises the consecutive Stop-hook block cap for the session at launch, because the chain's goal is armed mid-session and the launch-time goal read cannot see it (see `.claude/rules/moai/workflow/goal-directive.md` § Raising the block cap for an infinite goal, which names both trigger conditions).

The raise is **session-wide**, not scoped to the factory chain. Two consequences an operator should be able to recognize:

- Declining at the first gate still leaves the session carrying the raised ceiling.
- Arming an unrelated goal later in the same session inherits that ceiling too, so an unrelated loop may run considerably longer than it would in a non-factory session.

This is a longer unattended leash, not a gate bypass — all four human gates above are question rounds rather than Stop-hook blocks, so a raised block cap cannot skip any of them. Scoping the cap to a single goal would require a per-goal runtime mechanism the runtime does not expose.
