---
title: /moai harness
weight: 55
draft: false
---

Operates the V3R4 Self-Evolving Harness learning subsystem. Documents the 4-tier evolution ladder (observer → heuristic → rule → frozen-zone) and the 5-layer safety pipeline (frozen-guard → canary → contradiction → rate-limit → human oversight).

{{< callout type="info" >}}
**Slash command**: Type `/moai harness` in Claude Code to invoke this command directly.
{{< /callout >}}

## Overview

`/moai harness` provides four verbs (`status`, `apply`, `rollback`, `disable`) to safely operate the MoAI-ADK Self-Evolving learning subsystem. The PostToolUse hook records every tool invocation in `.moai/harness/usage-log.jsonl` (append-only). Proposals are classified along a 4-tier evolution ladder, and any Tier-4 change to a frozen zone requires explicit user approval through `AskUserQuestion`.

Key concepts:

- **Observer**: The PostToolUse hook appends every tool use to `.moai/harness/usage-log.jsonl`.
- **4-Tier Evolution Ladder**: observation → heuristic → rule → frozen-zone proposal.
- **5-Layer Safety Pipeline**: Every evolution proposal must pass five safety checks before being applied.
- **CLI Retirement**: Since V3R4, every verb is executed by file-system operations inside the workflow body. The Go binary does not expose a `moai harness` subcommand.

## Command Syntax

```bash
/moai harness {status | apply | rollback <YYYY-MM-DD> | disable}
```

- An empty argument prints help text.
- All verbs run in the orchestrator main context with privileged file-system access.

## Verbs

### status

Prints the current harness learning state, pending Tier-4 proposals, and the 7-day rate-limit window usage.

- **Read-only**: Performs no file modification.
- **Output includes**:
  - `learning.enabled` setting from `.moai/config/sections/harness.yaml`
  - Number of pending Tier-4 proposals in `.moai/harness/proposals/`
  - Count of applied entries within the 7-day window from `.moai/harness/learning-history/applied/`
  - Recent tier promotion events (`tier-promotions.jsonl`)
  - Frozen Guard violation log (`frozen-guard-violations.jsonl`)

### apply

Submits the oldest pending Tier-4 proposal to the 5-Layer Safety pipeline. Application is preceded by an `AskUserQuestion` round driven by the orchestrator; the user must approve explicitly.

- **Preconditions**:
  - Fewer than one application within the 7-day window (REQ-HRN-FND-012 rate-limit floor).
  - Proposal payload integrity verified.
- **User options (Recommended / Modify / Defer / Reject)**: The first option is marked `(Recommended)`. On Apply, a pre-modification snapshot is stored under `.moai/harness/learning-history/snapshots/<ISO-DATE>/`.

### rollback `<YYYY-MM-DD>`

Reverts the most recent application by restoring the snapshot of the given date. If subsequent evolutions have accumulated, a conflict report is printed and user re-approval is requested.

- **Argument**: ISO-8601 date (YYYY-MM-DD). Invalid formats are rejected.
- **Effect**: `.moai/harness/learning-history/applied/<DATE>.json` is moved to `rolled-back/`, and affected files are restored to the snapshot.

### disable

Pauses harness learning by setting `learning.enabled: false`. PostToolUse observation continues, but the 4-tier classifier and proposal generator are inactive.

- **When to use**: When evolution proposals look suspicious or external audits are in progress.
- **Re-enable**: Set `learning.enabled: true` in `.moai/config/sections/harness.yaml`.

## 4-Tier Evolution Ladder

| Tier | Classification | Auto-apply | Notes |
|------|----------------|------------|-------|
| Tier-1 | Observation | n/a (manual review) | Passive log accumulation only |
| Tier-2 | Heuristic | Suggestion only | Orchestrator recommends to user |
| Tier-3 | Rule | Non-frozen areas may auto-apply | Canary must pass |
| Tier-4 | Frozen-zone | **User approval required** | Must clear 5-Layer Safety |

Frozen zones are defined in `.claude/rules/moai/design/constitution.md` §2 and `.claude/rules/moai/core/zone-registry.md`.

## 5-Layer Safety Pipeline

1. **L1 Frozen Guard**: Blocks modification attempts on frozen-zone targets.
2. **L2 Canary**: Simulates impact in an isolated sandbox.
3. **L3 Contradiction**: Detects conflicts with other active rules.
4. **L4 Rate Limit**: At most one application per 7-day window (REQ-HRN-FND-012).
5. **L5 Human Oversight**: Orchestrator-led `AskUserQuestion` approval round.

If any layer rejects, `apply` aborts and the proposal stays `pending`.

## Examples

```bash
# 1) Inspect current state
/moai harness status

# 2) Review and apply the oldest pending Tier-4 proposal
/moai harness apply

# 3) Roll back the most recent application using yesterday's snapshot
/moai harness rollback 2026-05-21

# 4) Pause learning
/moai harness disable
```

## References

- [`.claude/skills/moai/workflows/harness.md`](https://github.com/modu-ai/moai-adk) — workflow body SSOT
- [`SPEC-V3R4-HARNESS-001`](https://github.com/modu-ai/moai-adk) — V3R4 foundation SPEC (supersedes three V3R3 harness SPECs)
- [`/moai plan`](/en/workflow-commands/moai-plan) — SPEC creation
- [`/moai run`](/en/workflow-commands/moai-run) — DDD/TDD implementation
- [`/moai sync`](/en/workflow-commands/moai-sync) — documentation sync + PR
