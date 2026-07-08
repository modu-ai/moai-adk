---
description: >
  Creates comprehensive SPEC documents using GEARS notation (EARS retained
  as legacy reference, 6-month backward-compat window) as the first step
  of the Plan-Run-Sync workflow. Handles project exploration, SPEC file
  generation, validation, and optional Git environment setup with worktree
  or branch creation. Use when planning features or creating specifications.
user-invocable: false
metadata:
  version: "2.8.0"
  category: "workflow"
  status: "active"
  updated: "2026-05-25"
  tags: "plan, spec, gears, ears, requirements, specification, design"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["plan", "spec", "design", "architect", "requirements", "feature request"]
  agents: ["manager-spec", "Explore", "manager-git"]
  phases: ["plan"]
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->
<!-- Emits one line per Phase entry/exit to stderr in format: [trace] /moai plan Phase <N> <enter|exit> -->

# Plan Workflow Orchestration

## Phase Owners (per the canonical agent catalog policy)

Phase Owners: `manager-spec` (SPEC artifact authoring — spec.md/plan.md/acceptance.md/design.md/research.md/progress.md §F.1) + `Explore` (read-only codebase investigation; Anthropic built-in agent) + `manager-git` (worktree/branch creation when Phase 3 worktree env opt-in).

Cross-reference: per-SPEC Phase 0.5 SKIP rationale recorded at `.moai/specs/SPEC-{ID}/progress.md` § Phase 0.5 SKIP Rationale; Phase 0.95 Mode Selection autopilot logging at `progress.md` § Phase 0.95 Mode Selection.

## Purpose

Create comprehensive SPEC documents using **GEARS notation** (Generalized EARS — the canonical SPEC authoring form as of v3.0.0) as the first step of the Plan-Run-Sync workflow. EARS notation is retained as the explicit 6-month backward-compatibility legacy reference for the 88 pre-v3 SPECs (legacy window expires 2026-11-22 per the canonical GEARS migration policy). Handles project exploration, SPEC file generation, validation, and optional Git environment setup with worktree or branch creation.

Canonical GEARS authoring guide: `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format.

For phase overview and token budgets, see: `.claude/rules/moai/workflow/spec-workflow.md`

---

## Phase Routing Table

| Phase / Section | Sub-skill | Description |
|---|---|---|
| Step 0: Brain Proposal Detection | `plan/context-discovery.md` | Brain IDEA scan and SPEC candidate surfacing |
| Phase 1A: Project Exploration | `plan/context-discovery.md` | Explore subagent codebase analysis |
| Phase 0.3: Clarity Evaluation | `plan/context-discovery.md` | Clarity scoring (1-10) and skip conditions |
| Phase 0.3.1: Deep Interview Loop | `plan/clarity-interview.md` | 1-5 round topic-focused interview |
| Phase 0.4: UltraThink Auto-Activation | `plan/clarity-interview.md` | Complexity-based extended reasoning activation |
| Phase 0.5: Deep Research | `plan/clarity-interview.md` | Explore subagent research.md artifact |
| Phase 1.25: Design Direction | `plan/clarity-interview.md` | UI/UX intent-first design direction |
| Phase 1B: SPEC Planning | `plan/clarity-interview.md` | manager-spec GEARS structure (EARS legacy retained) + candidate proposal |
| Decision Point 1 + Annotation Cycle | `plan/clarity-interview.md` | Plan review HUMAN GATE + 1-6 iteration cycle |
| Phase 1.5: Pre-Creation Validation | `plan/spec-assembly.md` | Document type classification + SPEC ID validation |
| Phase 2: SPEC Document Creation | `plan/spec-assembly.md` | spec.md + plan.md + acceptance.md + spec-compact.md |
| Phase 2.3: Independent SPEC Review | `plan/spec-assembly.md` | plan-auditor adversarial audit + retry loop |
| Phase 2.5: GitHub Issue Creation | `plan/spec-assembly.md` | gh issue create + bidirectional reference |
| Phase 3: Git Environment Setup | `plan/spec-assembly.md` | BODP Gate + Worktree/Branch/Current path |
| Phase 3.5: MX Tag Planning | `plan/spec-assembly.md` | ANCHOR/WARN/NOTE target identification |
| Phase 3.6: SPEC Quality Gate | `plan/spec-assembly.md` | GEARS ↔ AC coverage (EARS legacy form accepted for pre-v3 SPECs) + security scope check |
| Decision Point 2/3/3.5 | `plan/spec-assembly.md` | Dev environment + next action + execution mode |
| Completion Criteria | `plan/spec-assembly.md` | All checklist items + audit-ready signal |
| Test Scenarios | `plan/spec-assembly.md` | Normal/Existing Assets/Error flow examples |

---

## Invocation Flow

```
/moai plan [description] [--worktree|--branch] [--team] [--no-issue]
  └─ context-discovery.md
       ├─ Step 0: Brain proposal scan
       ├─ Phase 1A: Explore (optional)
       └─ Phase 0.3: Clarity evaluation (1-10 score)
            └─ clarity-interview.md
                 ├─ Phase 0.3.1: Deep interview (1-5 rounds, if score 4-10)
                 ├─ Phase 0.4: UltraThink auto-activation (if complexity >= 7)
                 ├─ Phase 0.5: Deep research → research.md (recommended)
                 ├─ Phase 1.25: Design direction (if UI/UX keywords)
                 ├─ Phase 1B: manager-spec SPEC planning
                 └─ Decision Point 1: HUMAN GATE (Proceed / Annotate / Draft / Cancel)
                      └─ spec-assembly.md
                           ├─ Phase 1.5: Pre-creation validation gate
                           ├─ Phase 2: SPEC document creation
                           │    └─ [HARD] Pre-write frontmatter checklist (12 fields)
                           ├─ Phase 2.3: plan-auditor review (≤3 iterations)
                           ├─ Phase 2.5: GitHub Issue creation (conditional)
                           ├─ Phase 3: Git environment (BODP Gate → worktree/branch/current)
                           ├─ Phase 3.5: MX tag planning
                           ├─ Phase 3.6: SPEC quality gate
                           └─ Decision Point 2/3/3.5: Execution mode selection
```

---

## Cross-References

- SPEC workflow overview: `.claude/rules/moai/workflow/spec-workflow.md`
- AskUserQuestion protocol: `.claude/rules/moai/core/askuser-protocol.md`
- BODP gate algorithm: `.claude/rules/moai/development/branch-origin-protocol.md`
- Worktree isolation: `.claude/rules/moai/workflow/worktree-integration.md`
- MX tag protocol: `.claude/rules/moai/workflow/mx-tag-protocol.md`
- Session handoff (Block 0): `.claude/rules/moai/workflow/session-handoff.md`

## Audit-Ready Signal

On successful plan completion (all SPEC files created, user approved), append to `.moai/specs/SPEC-{ID}/progress.md`:

```
- plan_complete_at: {ISO-8601 timestamp}
- plan_status: audit-ready
```

This signal marks the plan artifacts as finalized and enables the Plan Audit Gate at `/moai run` Phase 0.5.

---

## Team Mode Routing

When --team flag is provided or auto-selected, the plan phase MUST switch to team orchestration:

1. Verify prerequisites: workflow.team.enabled == true AND CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 env var is set
2. If prerequisites met: Read ${CLAUDE_SKILL_DIR}/team/plan.md and execute the team workflow (spawn researcher + analyst + architect via Agent(name=...) — the team forms implicitly on first spawn)
3. If prerequisites NOT met: Warn user then fallback to standard sub-agent mode (manager-spec)

Team composition: researcher (haiku) + analyst (inherit) + architect (inherit)

For detailed team orchestration steps, see ${CLAUDE_SKILL_DIR}/team/plan.md.

---

Version: 2.8.0
Updated: 2026-05-25
Changes: Added test scenarios, Phase 0.9 JIT Language Detection.

---

## Custom Harness Extension (Optional)

@.moai/harness/plan-extension.md

*(이 파일은 `/moai project --harness`로 생성됩니다. 파일이 없으면 자동으로 skip됩니다.)*

## Sentinel Error Keys

A CI audit verifies the literal `MODE_PIPELINE_ONLY_UTILITY` sentinel remains present in this skill body (shared with `design.md`). Passing `--mode pipeline` to `/moai plan` is rejected because plan is a Multi-Agent subcommand; pipeline mode is reserved for utility subcommands.
