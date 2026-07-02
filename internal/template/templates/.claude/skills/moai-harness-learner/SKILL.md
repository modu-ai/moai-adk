---
name: moai-harness-learner
description: Harness learning subsystem coordinator. Produces Tier 4 auto-update proposal payloads consumed by the orchestrator (which surfaces them via AskUserQuestion) and orchestrates Apply/Rollback flows. Triggers when harness learning proposals are pending or learning lifecycle management is needed.
when_to_use: >
  Use for harness learning lifecycle management: producing Tier 4 auto-update
  proposal payloads for the orchestrator (surfaced via AskUserQuestion), and
  coordinating Apply/Rollback flows when learning proposals are pending.
allowed-tools: Bash,Read,Write,Edit
user-invocable: false
---

# moai-harness-learner

<!-- @MX:NOTE: [AUTO] V3R4 contract — this skill body is preserved unchanged per the harness foundation policy §10 exclusion #10 (text annotation only, no behavioral change). The 4-tier observation/heuristic/rule/auto_update ladder defined here is preserved verbatim under REQ-HRN-FND-011. The orchestrator-only AskUserQuestion contract is asserted by REQ-HRN-FND-015 (cross-reference: .claude/rules/moai/core/agent-common-protocol.md § User Interaction Boundary). The downstream replacement of the frequency-count classifier with an embedding-cluster algorithm is deferred to the harness classifier-upgrade policy. -->

Coordinator skill for the Harness Learning Subsystem (the harness-learning policy, superseded by the harness foundation policy as the active V3R4 foundation; this V3R3 SPEC's 4-tier ladder is preserved unchanged).
Produces Tier 4 auto-update proposal payloads consumed by the MoAI orchestrator; the orchestrator surfaces them to the user via AskUserQuestion and orchestrates Apply/Rollback flows. Canonical contract: `.claude/rules/moai/core/askuser-protocol.md § Orchestrator-Subagent Boundary` (the constitutional rule/002/003).

## Quick Reference

**Role**: Orchestrator-side bridge between CLI (`moai harness`) and AskUserQuestion.

**Key constraint** [HARD]: `moai harness apply` returns a JSON payload representing a Tier 4 auto-update proposal. This skill produces the payload; the orchestrator surfaces it via `AskUserQuestion`. The CLI itself does NOT prompt the user. Canonical contract: `.claude/rules/moai/core/askuser-protocol.md § Orchestrator-Subagent Boundary`.

**Common triggers**:
- `moai harness status` — check tier distribution and pending proposals
- `moai harness apply` — load next pending proposal (returns JSON payload)
- `moai harness rollback <date>` — restore snapshot
- `moai harness disable` — set learning.enabled: false

**Workflow**:
1. Run `moai harness status` to inspect state.
2. Run `moai harness apply` to get the proposal payload.
3. Hand payload to the orchestrator for `AskUserQuestion` surfacing (approve / reject).
4. On approve: write approval to proposals dir and signal CLI to proceed.
5. On reject: remove proposal file (no changes applied).

---

## Implementation Guide

### Step 1: Status Check

```bash
moai harness status --project-root <project_root>
```

Output includes:
- `enabled` state
- Tier distribution (observation / heuristic / rule / auto_update)
- Rate limit window status
- Number of pending proposals

### Step 2: Fetch Proposal Payload

```bash
moai harness apply --project-root <project_root>
```

The command outputs a JSON block with:
- `id` — proposal identifier
- `target_path` — file to be modified
- `field_key` — `description` or `triggers`
- `new_value` — proposed new content
- `pattern_key` — what triggered this proposal
- `observation_count` — how many times this pattern was observed

### Step 3: Produce structured payload for orchestrator consumption

[HARD] This skill produces a structured payload representing the Tier 4 auto-update proposal; the MoAI orchestrator surfaces it via `AskUserQuestion`. Canonical contract: `.claude/rules/moai/core/askuser-protocol.md § Orchestrator-Subagent Boundary`.

**Payload schema**:

- `proposal_id` — proposal identifier
- `target_path` — file to be modified
- `field_key` — `description` or `triggers`
- `current_value` — existing content (for diff context)
- `new_value` — proposed new content
- `observation_count` — pattern observation count
- `confidence` — auto-update confidence score (0.0–1.0)
- `recommended_action` — `approve` (default) | `reject` | `inspect` | `defer`

The skill emits this payload as its tool output. The orchestrator reads the payload, preloads `AskUserQuestion` via `ToolSearch(query: "select:AskUserQuestion")`, and surfaces the four-option decision (approve / reject / inspect / defer) to the user. On user approval, the orchestrator re-delegates to this skill with `action=apply`; on rejection, `action=skip`. The "(권장)" recommendation suffix and per-option descriptions are constructed by the orchestrator from the payload's `recommended_action` field per `askuser-protocol.md § Socratic Interview Structure`.

### Step 4: On Approve

The skill applies the change by invoking the safety pipeline directly. Since the CLI `apply` only surfaces the payload (not executes), the actual write happens via the harness package's `Apply()` function, gated by the 5-Layer Safety Pipeline.

For the coordinator skill, the simplest flow is:
1. User selects "approve"
2. Write `approved: true` to `.moai/harness/proposals/<id>.decision`
3. Run `moai harness apply --execute` (if the CLI supports it) or call the harness API directly.

### Step 5: On Reject

1. Delete `.moai/harness/proposals/<id>.json`
2. Confirm deletion to user.

### Rollback Flow

```bash
# List available snapshots
ls .moai/harness/learning-history/snapshots/

# Rollback to a specific snapshot
moai harness rollback 2026-04-27T00-00-00.000000000Z --project-root <project_root>
```

### Disable Learning

```bash
moai harness disable --project-root <project_root>
```

Sets `learning.enabled: false` in `.moai/config/sections/harness.yaml`.
Comments and key ordering are preserved (YAML round-trip).

---

## Works Well With

- `moai-meta-harness` — generates the `harness-*` skills that are targets of auto-updates
- `moai-workflow-tdd` — TDD cycle generates events that feed into the observer
- `moai-foundation-quality` — quality gates run after auto-updates to validate correctness

## Safety Architecture Reference

The 5-Layer Safety Pipeline (L1 Frozen Guard → L2 Canary Check → L3 Contradiction Detector → L4 Rate Limiter → L5 Human Oversight) protects every Tier 4 auto-update:

| Layer | Guard | Action on violation |
|-------|-------|---------------------|
| L1 | Frozen Guard | Block — FROZEN paths are never modified |
| L2 | Canary Check | Block — if effectiveness drops >0.10 |
| L3 | Contradiction Detector | Block — if trigger conflicts arise |
| L4 | Rate Limiter | Block — max 3 per week, 24h cooldown |
| L5 | Human Oversight | Orchestrator surfaces user-approval via AskUserQuestion (this skill emits payload) |

[HARD] L1 Frozen paths (never auto-modified at runtime):
- `.claude/agents/moai/**` (template-managed agents; `.claude/agents/harness/` is a user-owned allowed-write target, NOT frozen)
- `.claude/skills/moai-*/**`
- `.claude/rules/moai/**`
- `.moai/project/brand/**` (guard default-deny — not in the allowed-prefix list, so writes are blocked even though it is not a frozen prefix)

Only user-area skills (`.claude/skills/harness-*/`) and agents (`.claude/agents/harness/`) are valid auto-update targets.
