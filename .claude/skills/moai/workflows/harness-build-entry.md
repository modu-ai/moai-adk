---
description: >
  Harness build entry workflow. Turns a natural-language harness-creation
  request into a concrete harness via Context-First Discovery (domain / goal /
  constraints / scope extraction), harness name derivation (derived from the
  request, NOT statically supplied), explicit orchestrator-issued approval,
  then delegation to the Builder Workflow. Conducts AskUserQuestion Socratic
  rounds when intent clarity is below 100%.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-06-19"
  tags: "harness, build, natural-language, context-first-discovery, name-derivation, approval-gate"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["harness build", "build a harness", "create a harness", "harness for"]
  agents: ["moai-meta-harness"]
  phases: ["harness"]
---

# Workflow: harness-build-entry — Natural-Language Harness Build Entry

Purpose: This workflow is the natural-language branch of the `harness` subcommand. It takes a free-form harness-creation request (e.g. "build a harness for CLI template development") and turns it into a concrete harness through a four-step pipeline: Context-First Discovery, harness `<name>` derivation, explicit orchestrator-issued approval, then delegation to the Builder Workflow.

The orchestrator executes this workflow body directly (it is NOT a subagent). Subagents reachable from this workflow MUST NOT invoke `AskUserQuestion` — they return structured blocker reports and the orchestrator re-runs the round (asymmetric boundary per `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary).

## Reach

This workflow is reached when the `harness` subcommand dispatcher (in `SKILL.md` § harness) classifies `$ARGUMENTS` as a natural-language request — i.e. the FIRST token is NOT one of the reserved verbs (`status` / `apply` / `rollback` / `disable`). Reserved verbs route to the sibling `harness.md` learning-lifecycle workflow instead.

## Authoritative Sources

- AskUserQuestion contract: `.claude/rules/moai/core/askuser-protocol.md` (canonical reference — preload procedure, Socratic interview structure, option-description standards, bias prevention)
- Orchestrator-subagent boundary: `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary
- Context-First Discovery: CLAUDE.md §7 Rule 5 (trigger conditions + Socratic interview)
- Skill namespace policy: `.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy (`harness-*` user-owned vs `moai-harness-*` template-builder)
- Companion learning-lifecycle workflow: `${CLAUDE_SKILL_DIR}/workflows/harness.md` (Branch A — reserved verbs)
- Builder module (orchestrator-direct 4 phases): `${CLAUDE_SKILL_DIR}/workflows/harness-builder.md` (ANALYZE / PLAN / GENERATE / ACTIVATE — the orchestrator-side logic Phase 4 below transitions into)

## Input

`$ARGUMENTS` — a natural-language harness-creation request (the full text after the `harness` subcommand keyword, MINUS any reserved verb). Example: `build a harness for CLI template development`.

## Phase 0: Reserved-Verb Guard

[HARD] If `$ARGUMENTS` (trimmed, first token) matches a reserved verb (`status` / `apply` / `rollback` / `disable`), STOP — this is a misroute. The dispatcher should have sent it to the learning-lifecycle workflow (`${CLAUDE_SKILL_DIR}/workflows/harness.md`). Re-emit the routing guidance and halt. This guard is defense-in-depth; the dispatcher in `SKILL.md` already filters, but this workflow body re-verifies to catch direct-invocation edge cases.

## Phase 1: Context-First Discovery (extract domain / goal / constraints / scope)

Apply CLAUDE.md §7 Rule 5 (Context-First Discovery). The orchestrator extracts a preliminary profile from the raw request:

1. **Domain** — the primary subject area the harness will serve (e.g., "CLI template development", "research", "code review"). Extracted from the noun phrase following "for" / "to" / "that" in the request.
2. **Goal** — the outcome the user wants from the harness (e.g., "automate template generation", "parallelize research fan-out"). Extracted from the action verb + object.
3. **Constraints** — boundaries the harness must respect (e.g., 16-language neutrality, template-content neutrality, namespace isolation). May be implicit in the domain; surface them explicitly.
4. **Scope** — which files / surfaces the harness will touch (e.g., CLI template files, docs-site, hook scripts). Extracted from the domain + project structure.

Emit the preliminary profile as a structured block BEFORE the Socratic round so the user can see what was extracted and correct it.

### Phase 1.5: AskUserQuestion Socratic Rounds (when clarity < 100%)

If the extracted profile has ANY ambiguous field (domain too vague, goal unstated, constraints unenumerated, scope unclear), conduct AskUserQuestion Socratic rounds per `.claude/rules/moai/core/askuser-protocol.md`:

1. `ToolSearch(query: "select:AskUserQuestion")` — preload the deferred tool schema.
2. Compose a round of ≤4 questions, ≤4 options each, first option marked `(권장)` / `(Recommended)`, all text in the user's `conversation_language` (read from `.moai/config/sections/language.yaml`).
3. Each subsequent round MUST narrow ambiguity by building on previous answers (no repeated questions).
4. Continue until intent clarity reaches 100%.
5. Consolidate the confirmed profile into a structured block and proceed to Phase 2.

[HARD] Do NOT skip the Socratic rounds when clarity is below 100%. The name derivation (Phase 2) and approval gate (Phase 3) both depend on a fully-resolved profile.

## Phase 2: Harness `<name>` Derivation

Derive the harness `<name>` from the confirmed profile (Phase 1 + 1.5). The name is NOT statically supplied by the user — the orchestrator derives it. Naming rules:

- Lowercase, hyphen-separated, ≤32 characters.
- Reflects the domain (e.g., domain "CLI template development" → name `cli-template-dev`; domain "research" → name `research`).
- MUST use the `harness-` prefix ONLY if it will live under the user-owned `.claude/skills/harness-*/` namespace. If it is a project-level harness without the `harness-` skill prefix, omit the prefix (the `/harness:<name>` command namespace is separate from the skill namespace).
- MUST NOT collide with an existing harness name. Check `.claude/commands/harness/<name>.md` existence before confirming.

Surface the derived name to the user as part of the Phase 3 approval gate. If the user rejects the derived name via the "Modify" option, re-derive from the refined profile (do NOT ask the user to type a name statically — re-derivation keeps the name semantically tied to the request).

## Phase 3: Approval Gate (orchestrator-issued AskUserQuestion)

[HARD] Before delegating to the Builder Workflow, the orchestrator MUST obtain explicit approval via `AskUserQuestion`. This gate is mandatory and score-independent (a strong Context-First Discovery profile never authorizes skipping it — parallel to the Implementation Kickoff Approval human gate).

`ToolSearch(query: "select:AskUserQuestion")` → `AskUserQuestion` with the canonical four-option pattern (first option `(권장)` / `(Recommended)`):

- **Build (권장)** — Proceed to Phase 4 (delegate to the Builder Workflow with the confirmed profile + derived name).
- **Modify profile** — Return to Phase 1.5 with the user's refinement (e.g., narrow the domain, add a constraint). Re-derive the name in Phase 2.
- **Rename** — Re-derive the harness `<name>` from the same profile with a different naming heuristic (the user hints at a preferred stem). Do NOT ask the user to type the name statically.
- **Abort** — Stop. No files are modified. The request is recorded in `.moai/harness/build-requests/` for retrospective analysis (best-effort; the directory is created on first use).

## Phase 4: Transition to the Orchestrator-Direct Builder

On `Build` approval, the orchestrator transitions directly into the Builder — it does NOT delegate to a dynamic-workflow script and does NOT spawn a separate Builder agent. The Builder is **orchestrator-side logic**: the orchestrator continues executing in the same session, running the 4 signal-driven phases (ANALYZE / PLAN / GENERATE / ACTIVATE) using ordinary `Agent()` spawn. Intermediate results are held in the orchestrator's session context.

**Read the Builder module for the full phase logic**: `${CLAUDE_SKILL_DIR}/workflows/harness-builder.md`. That module documents:

- **ANALYZE** — orchestrator parallel `Agent(agentType: "Explore", effort: "low")` fan-out across the codebase + docs + existing harness surfaces + SPEC history (read-only, main tree). Produces a domain profile + task-pattern inventory.
- **PLAN** — orchestrator spawns a single `Agent(model: "opus", effort: "xhigh")` that selects/combines patterns from the 6-pattern catalog, defines specialist roles, maps each to an execution primitive, and drafts the manifest. The orchestrator then runs an **AskUserQuestion approval gate** at the PLAN→GENERATE boundary (first-class, because the orchestrator holds the boundary — this is the self-contradiction resolution that made the Builder orchestrator-direct).
- **GENERATE** — orchestrator fan-out emits the 5 artifact types (thin-wrapper command, Runner Workflow, specialist sub-agents, companion skills, manifest.json). Conditional `Agent(isolation: "worktree")` per specialist whose manifest declares `isolation: worktree`.
- **ACTIVATE** — orchestrator-direct dry-run + `/goal` autonomous convergence + optional with/without A/B. The A/B is **skipped** for tasks within the model's solo reliable range (load-bearing minimum), with the skip recorded + rationale.

**There is no `harness-build.js` script.** The Builder is orchestrator-side logic, not a dynamic-workflow script. The Runner (`harness-<name>-run.js`) generated by GENERATE stays a dynamic-workflow script — only creation is orchestrator-direct; execution runs inside the generated `/harness:<name>` command.

**Carry-over invariant.** The confirmed profile + derived name from Phases 1-3 above is the single source of truth for the manifest's `source_request` field. The Builder MUST carry the original natural-language request verbatim into `manifest.json.source_request`.

## Phase 5: Post-Build Summary

After the Builder's ACTIVATE phase completes, render a one-paragraph summary in the user's `conversation_language` covering:

1. What was built (harness name + domain + entry command `/harness:<name>`).
2. Where the artifacts live (`.claude/commands/harness/<name>.md`, `.claude/workflows/harness-<name>-run.js`, `.claude/agents/harness/<name>-*-specialist.md`, `.claude/skills/harness-<name>-*/`, `manifest.json`).
3. Whether ACTIVATE's A/B was run or skipped (load-bearing-minimum rationale).
4. Suggested next step (e.g., "Run `/harness:<name>` to execute the harness on a sample task").

## Error Handling

| Symptom | Likely cause | Recovery |
|---------|-------------|----------|
| Reserved-verb misroute (Phase 0 fires) | Dispatcher misrouted a `status`/`apply`/`rollback`/`disable` call | Re-emit routing guidance; halt this workflow; user re-invokes via the correct path |
| AskUserQuestion schema not loaded | Deferred-tool preload missed | The workflow body explicitly preloads via `ToolSearch(query: "select:AskUserQuestion")` before each Socratic round and the Phase 3 gate |
| Derived name collides with existing harness | `.claude/commands/harness/<name>.md` already exists | Re-derive with a domain-specific suffix (e.g., `<name>-v2`) OR ask the user via the "Rename" Phase 3 option |
| PLAN→GENERATE gate returns "Revise manifest" | User rejected the draft manifest at the Builder's approval gate | Return to PLAN with the user's refinement; re-derive specialists/patterns; re-present the gate |

## Cross-references

- Skill namespace policy: `.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy (`harness-*` user-owned)
- Companion learning-lifecycle workflow: `${CLAUDE_SKILL_DIR}/workflows/harness.md` (Branch A)
- moai SKILL.md § harness (dispatcher — argument-branching routing rule)
- AskUserQuestion canonical: `.claude/rules/moai/core/askuser-protocol.md`
- Orchestrator-subagent boundary: `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary
- Context-First Discovery: CLAUDE.md §7 Rule 5
