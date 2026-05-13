---
name: moai-workflow-harness
description: >
  Harness management workflow: surfaces the harness learning subsystem
  (observer + 4-tier evolution + 5-layer safety) to the user. Routes to
  CLI `moai harness {status,apply,rollback,disable}` and bridges Tier 4
  proposals via AskUserQuestion (delegated to moai-harness-learner skill).
  Use when inspecting harness state, approving auto-update proposals,
  rolling back snapshots, or disabling the learning subsystem.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-05-14"
  tags: "harness, learning, observer, tier-4, safety, evolution, apply, rollback"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["harness", "harness status", "harness apply", "harness rollback", "harness disable", "tier 4", "harness proposal", "harness evolve"]
  agents: ["moai-harness-learner"]
  phases: ["harness"]
---

<!-- @MX:NOTE: [AUTO] Thin orchestration wrapper for `moai harness` CLI verbs (SPEC-V3R3-HARNESS-LEARNING-001 REQ-HL-009). Heavy lifting lives in CLI (internal/cli/harness.go) + moai-harness-learner skill. -->
<!-- @MX:REASON: [AUTO] This workflow exists so users can invoke harness lifecycle from inside Claude Code without dropping to terminal. CLI is authoritative; this workflow is a discoverability + AskUserQuestion bridge. -->

# Workflow: harness — Harness Learning Subsystem Management

Purpose: Surface the harness learning subsystem (observer, 4-tier proposals, 5-layer safety pipeline) to the user. This workflow is a thin orchestration layer over the `moai harness` CLI; the CLI itself does not interact with the user, and Tier 4 approval flows are delegated to the `moai-harness-learner` skill which owns the `AskUserQuestion` bridge.

Authoritative source:
- SPEC: `SPEC-V3R3-HARNESS-LEARNING-001` (status: completed)
- CLI: `internal/cli/harness.go` (REQ-HL-009: 4 verbs)
- Skill: `.claude/skills/moai-harness-learner/SKILL.md` (AskUserQuestion bridge)
- Config: `.moai/config/sections/harness.yaml` (learning subsystem on/off)
- Artifacts: `.moai/harness/` (usage-log.jsonl, proposals/, learning-history/snapshots/)

---

## Input

`$ARGUMENTS` — the first word is the verb (`status`, `apply`, `rollback`, `disable`, or empty for help). Remaining arguments are passed to the CLI.

## Verb Routing

[HARD] Extract the FIRST WORD from `$ARGUMENTS`. Match against the verb table below. If empty or unrecognized, show the help block (Phase 0).

| Verb | CLI invocation | Skill bridge | Purpose |
|------|----------------|--------------|---------|
| `status` | `moai harness status` | none | Inspect tier distribution, rate-limit window, pending proposal count |
| `apply` | `moai harness apply` | moai-harness-learner | Surface next Tier 4 proposal via AskUserQuestion, approve/reject, run 5-layer safety pipeline |
| `rollback <YYYY-MM-DD>` | `moai harness rollback <date>` | none | Restore snapshot from date |
| `disable` | `moai harness disable` | none | Set `learning.enabled: false` in harness.yaml |

---

## Phase 0: Help (default when no verb)

If `$ARGUMENTS` is empty or matches `help` / `--help` / `-h`, render the verb table above plus a one-line summary in the user's `conversation_language`. Then stop. Do not invoke the CLI.

---

## Phase 1: Pre-Execution Sanity Check

Before invoking any verb, verify:

1. Project root is detected (`.moai/config/config.yaml` exists). If absent, abort with guidance to run `moai init` first.
2. Harness learning subsystem state: read `.moai/config/sections/harness.yaml` `learning.enabled` field.
   - If `enabled: false` and verb is anything other than `status`, surface a warning that the subsystem is disabled (via AskUserQuestion: continue / abort).
3. For `rollback`: the date argument is mandatory. If missing, abort with usage hint `moai harness rollback <YYYY-MM-DD>`.

---

## Phase 2: Verb Dispatch

### 2.1 status

Run:

```
moai harness status --project-root <project_root>
```

Capture stdout and render to the user as a Markdown block. Expected payload structure (per `internal/cli/harness.go`):

```
Harness Learning Status
  Enabled:        <bool>
  Log entries:    <int>
  Patterns:       <int>

Tier Distribution:
  observation:    <int> patterns
  heuristic:      <int> patterns
  rule:           <int> patterns
  auto_update:    <int> patterns (Tier 4 — pending user approval)

Pending Proposals: <int>
  [PENDING] <proposal_id>: <skill_path> (<field>)
  ...
```

No user prompt. Stop after rendering.

### 2.2 apply

[HARD] Delegate the approval flow to the `moai-harness-learner` skill. Do NOT call `AskUserQuestion` from this workflow body directly — the skill owns the bridge so the orchestrator's prompt construction stays consistent across CLI and direct skill invocations.

Steps:

1. Invoke: `Skill("moai-harness-learner")` with argument `apply`.
2. The skill internally runs `moai harness apply --project-root <project_root>`, parses the JSON payload, surfaces the proposal via `AskUserQuestion` (approve / reject), and writes the user's decision to `.moai/harness/proposals/<id>/decision.json`.
3. On approve: the skill runs the L1→L2→L3→L4→L5 safety pipeline (FrozenGuard → Canary → Contradiction → RateLimit → HumanApproval-already-collected) via `internal/harness/safety/` package. Applier writes the frontmatter change and creates a snapshot at `.moai/harness/learning-history/snapshots/<date>/`.
4. On reject: the proposal file is moved to `.moai/harness/proposals/rejected/`.
5. Render the final outcome (Applied / Rejected / Safety-Blocked) to the user.

Edge case: if `moai harness apply` returns no pending proposals, render "No Tier 4 proposals awaiting approval" and stop. Do not surface AskUserQuestion.

### 2.3 rollback

Validate the date argument matches `YYYY-MM-DD`. If invalid format, abort with usage hint.

Run:

```
moai harness rollback <YYYY-MM-DD> --project-root <project_root>
```

The CLI restores frontmatter from the snapshot directory. Render the list of restored files to the user.

Safety: rollback does NOT delete the current state — it overlays the snapshot. The current state is preserved at `.moai/harness/learning-history/snapshots/<rollback-timestamp>/pre-rollback/` (per `internal/harness/retention.go`).

### 2.4 disable

Before invoking the CLI, confirm intent via AskUserQuestion (the only place in this workflow that prompts the user directly, because the consequence is global subsystem deactivation):

- Question: "Disable harness learning subsystem? Observer will stop collecting events and no new proposals will be generated."
- Options: `Disable` (recommended for production), `Keep enabled`, `Disable temporarily (1h)`.

On `Disable` selection:

```
moai harness disable --project-root <project_root>
```

The CLI sets `learning.enabled: false` in `.moai/config/sections/harness.yaml`. Snapshots and proposals are preserved.

On `Keep enabled`: stop without changes.

On `Disable temporarily`: not currently supported by the CLI. Render guidance to re-enable manually via config edit, and stop without changes.

---

## Phase 3: Post-Execution Summary

After any successful verb execution, render a one-paragraph summary in the user's `conversation_language` covering:

1. What was done (verb + key result).
2. Where the artifact lives (`.moai/harness/...` path).
3. Suggested next step (e.g., after `apply`: "Run `/moai harness status` to verify the new tier distribution.").

---

## Error Handling

| Symptom | Likely cause | Recovery |
|---------|-------------|----------|
| `moai: command not found` | Binary not in PATH | Run `which moai`; reinstall via `make install` (dev) or `go install` (user) |
| `Error: project root not detected` | Not in MoAI project | `cd` to project root or run `moai init` |
| `Error: learning.enabled: false` on `apply` | Subsystem disabled | Re-enable via `moai harness enable` (manual config edit, no CLI verb yet) or accept the disabled state |
| `Error: no proposals matching <date>` on `rollback` | Snapshot doesn't exist for date | Run `moai harness status` to list available snapshot dates |
| AskUserQuestion schema not loaded | Deferred tool preload missed | The skill auto-preloads via `ToolSearch(query: "select:AskUserQuestion")` |

---

## Cross-references

- SPEC: `.moai/specs/SPEC-V3R3-HARNESS-LEARNING-001/spec.md` (REQ-HL-001 ~ REQ-HL-012)
- SPEC: `.moai/specs/SPEC-V3R3-PROJECT-HARNESS-001/spec.md` (16Q interview integration)
- SPEC: `.moai/specs/SPEC-V3R3-HARNESS-001/spec.md` (Meta-Harness Skill core + generated artifacts)
- Skill: `.claude/skills/moai-harness-learner/SKILL.md` (AskUserQuestion bridge)
- Skill: `.claude/skills/moai-meta-harness/SKILL.md` (project-specific harness generation)
- README: `.moai/harness/README.md` (subsystem overview + CLI reference)
- Rules: `.claude/rules/moai/NOTICE.md` (Apache-2.0 attribution to revfactory/harness)

<!-- Verifies REQ-HL-009: 4 verbs (status/apply/rollback/disable) exposed to orchestrator -->
<!-- Verifies REQ-HL-010: AskUserQuestion delegated to moai-harness-learner skill (orchestrator-only invariant) -->
<!-- Verifies REQ-HL-011: 5-layer safety pipeline invoked on apply (FrozenGuard → Canary → Contradiction → Rate → Human) -->
