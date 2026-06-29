---
description: "Plan Phase 0.3.1/0.4/0.5/1.25/1B/DP1 — Deep interview loop, UltraThink activation, deep research, design direction, SPEC planning, and plan review annotation cycle"
user-invocable: false
metadata:
  parent: moai-workflow-plan
  phase: "Phase 0.3.1 through Decision Point 1: Clarity Interview, Research, and Plan Review"
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->

### Phase 0.3.1: Deep Interview Loop (Conditional)

Purpose: Gather missing context through a structured, topic-focused interview before research begins. Each round presents curated options so the user can answer quickly.

**Entry condition:** Clarity score 4-10 AND skip conditions not met (from Phase 0.3).

**Guard:** [HARD] During the interview loop, the agent MUST NOT write implementation code or start codebase exploration. The sole output is `.moai/specs/SPEC-{ID}/interview.md`.

**Round topics:**

| Round | Focus Topic | Example Questions |
|---|---|---|
| 1 | Scope | What is included and explicitly excluded? |
| 2 | Constraints | Performance, security, compatibility, technology limits |
| 3 | Success criteria | How do we know when this is done and working correctly? |
| 4 | Edge cases | What unusual or failure scenarios must be handled? |
| 5 | Priority | What is the minimum viable slice if scope must be cut? |

**Per-round execution:**

For each round:

1. Formulate 3 recommended options relevant to the current topic and the user's request context.
2. Present via AskUserQuestion with exactly 4 options:
   - Option 1: [Recommended based on context] (Recommended): [Detailed description of this answer]
   - Option 2: [Alternative]: [Description]
   - Option 3: [Alternative]: [Description]
   - Option 4: Type your own answer: Enter a custom response if none of the above match
3. Record the user's answer.
4. Re-evaluate clarity score after each round.
5. If updated clarity score drops to 3 or below: end the loop early (user's answers added no useful information).
6. If updated clarity score reaches 8 or above: end the loop early (sufficient clarity achieved).
7. Display round counter: "Interview round {N}/{max_rounds}"

**Output:** Write all interview answers to `.moai/specs/SPEC-{ID}/interview.md` with this structure:

```
# Interview: {SPEC Title}

## Interview Phase 1: Scope
Question: {question asked}
Answer: {user's answer}

## Interview Phase 2: Constraints
...

## Clarity Score
Initial: {N}/10
Final: {N}/10
Rounds completed: {N}
```

**Context passing:** Pass `interview.md` to Phase 0.5 (Deep Research) and Phase 1B (SPEC Planning) as additional context. Both agents MUST read interview.md before proceeding.

### Phase 0.4: UltraThink Auto-Activation (Conditional)

Purpose: Automatically activate deep analysis mode for complex SPECs that benefit from structured reasoning.

**Activation condition**: Evaluate task complexity from Phase 1A exploration results or user request:
- Complexity score >= 7 (multi-domain, cross-cutting concerns)
- Request involves architectural decisions (new module, system redesign, migration)
- Request touches security-critical areas (auth, payment, data isolation)
- User explicitly includes `ultrathink` keyword in request

**UltraThink (primary deep reasoning trigger)**:
- `ultrathink`: Claude Code native deep analysis mode — activates Adaptive Thinking (Opus 4.7+) within the current agent context. Triggered by keyword detection in user input.

When UltraThink auto-activates:
- Log: "UltraThink mode activated: [reason]"
- Apply extended reasoning to Phase 0.5 research and Phase 1B SPEC creation
- Produce deeper analysis in research.md with trade-off comparisons and risk assessments
- Consider alternative approaches and document rejection rationale
- Optionally apply extended reasoning (ultrathink / Adaptive Thinking) for structured step-by-step decomposition; document each step in research.md when used

**Skip condition**: Simple, well-scoped features (complexity < 5, single domain, clear requirements). Log: "UltraThink skipped: low complexity task."

### Phase 0.5: Deep Research (Recommended)

Agent: Explore subagent (deep codebase analysis)

Purpose: Produce a persistent research.md artifact documenting deep codebase understanding. This document serves as a verification surface — MoAI and the user can review it and correct misunderstandings before planning begins.

When to run:
- Feature involves modifying existing code
- Feature has cross-module dependencies
- User explicitly requests research phase

When to skip:
- Simple, isolated additions (new file with no dependencies)
- User provides explicit "skip research" instruction

Tasks for the Explore subagent:
- Read target code areas in depth — understand how they work deeply, their intricacies and specificities
- Study related systems in great detail — trace data flow, identify implicit contracts and side effects
- Discover reference implementations in the existing codebase — find similar patterns that can guide the new implementation
- Search for relevant open-source examples or documented patterns that align with the project's conventions
- Document all findings in a structured research.md file

Research directives (Deep Reading patterns):
- Use language that demands thoroughness: "read deeply", "study in great detail", "understand the intricacies"
- Avoid surface-level scanning — agent must trace through actual execution paths
- Every finding must include specific file paths and line references

Output: `.moai/specs/SPEC-{ID}/research.md` containing:
- Architecture analysis with file paths and dependency maps
- Existing patterns and conventions discovered
- Reference implementations found (internal codebase or documented patterns)
- Risks, constraints, and implicit contracts identified
- Recommendations for the implementation approach

### Phase 1.25: Design Direction (Conditional)

Purpose: Establish design intent and direction for UI/UX-related SPECs before SPEC planning begins. Based on the Intent-First design philosophy from the interface-design methodology.

When to run:
- SPEC description contains 2+ UI/UX keywords: ui, frontend, interface, design, component, page, screen, layout, form, dashboard, button, modal, view, sidebar, navigation, widget, chart, table
- User explicitly requests design direction

When to skip:
- No UI/UX keywords detected in SPEC description
- User explicitly requests "skip design" or uses --prototype flag
- Backend-only, infrastructure, or documentation SPECs

Agent: per-spawn `Agent(general-purpose)` frontend specialist (with moai-design-craft skill; frontend whitelist per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 8)

Tasks:
1. Check if `.moai/design/system.md` exists and has content
2. If system.md exists: Load as design context, skip Intent-First process
3. If system.md is empty or missing: Execute Intent-First process:
   - Answer: Who is this human? What must they accomplish? What should this feel like?
   - Produce domain exploration: 5+ domain concepts, 5+ color world entries, 1 signature element
   - Identify 3+ defaults to avoid (generic patterns to reject)
4. Generate design direction artifact

Output: `.moai/specs/SPEC-{ID}/design-direction.md` containing:
- Intent statement (who, what, feel)
- Domain concepts and vocabulary
- Color world exploration
- Signature element definition
- Defaults to avoid
- Reference to `.moai/design/system.md` if exists

Design direction guard: [HARD] During Phase 1.25, the agent MUST NOT write implementation code. Focus exclusively on design exploration and direction definition.

After Phase 1.25: Offer to persist design decisions to `.moai/design/system.md` if it was newly created or updated. Use AskUserQuestion: "Save design direction to project-level design memory (.moai/design/system.md)?"

### Phase 1B: SPEC Planning (Required)

Agent: manager-spec subagent

Input: User request plus Phase 1A results (if executed), plus design-direction.md (if Phase 1.25 executed)

Tasks for manager-spec:
- Analyze project documents (product.md, structure.md, tech.md)
- Propose 1-3 SPEC candidates with proper naming
- Check for duplicate SPECs in .moai/specs/
- Design GEARS structure for each candidate using the 5 GEARS patterns (Ubiquitous, Event-driven `When`, State-driven `While`, Capability-gate `Where`, Event-detected unwanted). EARS legacy form is accepted for pre-v3 SPECs until 2026-11-22; new SPECs MUST use GEARS. Canonical authoring reference: `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format.
- Create implementation plan with technical constraints
- Identify library versions (production stable only, no beta/alpha)
- Search for reference implementations: Identify similar patterns in the existing codebase or well-documented approaches that can guide implementation
- When reference implementations are found, include them in the plan as "Reference: {file_path}:{line_range}" to improve implementation quality

Output: Implementation plan with SPEC candidates, GEARS-notation requirements (EARS legacy form accepted for pre-v3 SPECs until 2026-11-22), and technical constraints.

Implementation guard: [HARD] During Phases 0.5, 1A, and 1B, all agent prompts MUST include the instruction: "DO NOT write implementation code. Focus exclusively on research, analysis, and planning." This separation of thinking and typing is the foundation of effective AI-assisted development.

### Decision Point 1: Plan Review and Annotation Cycle

<!-- moai:evolvable-start id="gate-plan-1" -->
### HUMAN GATE: Plan Review

**Previous phase output:** SPEC draft with GEARS-notation requirements (EARS retained as legacy reference for pre-v3 SPECs) and acceptance criteria
**Approval question:** Does this SPEC capture the correct requirements and scope?
**Cannot proceed until:**
- [ ] User has reviewed the SPEC document
- [ ] User has confirmed acceptance criteria are testable
- [ ] User has approved the proposed file changes
- [ ] No open questions remain in the SPEC
<!-- moai:evolvable-end -->

Tool: AskUserQuestion (at orchestrator level only)

Options:
- Proceed with SPEC Creation (Recommended): Plan is approved, continue to Phase 1.5 then Phase 2
- Annotate Plan: Add inline notes to plan.md for revision (starts annotation cycle)
- Save as Draft: Save plan.md with status draft, create commit, print resume command, exit
- Cancel: Discard plan, exit with no files created

If "Proceed": Continue to Phase 1.5 then Phase 2.
If "Annotate": Enter Annotation Cycle (see below).
If "Draft": Save plan.md with status draft, create commit, print resume command, exit.
If "Cancel": Discard plan, exit with no files created.

#### Annotation Cycle (1-6 iterations)

Purpose: Allow users to iteratively refine the plan through inline notes before any code is written. This prevents expensive failures by catching architectural misunderstandings, missed conventions, and scope issues early.

Process:
1. User reviews plan.md (and research.md if available) in their editor
2. User adds inline notes directly into the document (e.g., "NOTE: use drizzle:generate for migrations, not raw SQL")
3. User signals completion via AskUserQuestion
4. MoAI delegates to manager-spec subagent: "Address all inline notes in the plan document and update it accordingly. DO NOT implement any code."
5. manager-spec updates plan.md, removing addressed notes and incorporating feedback
6. MoAI presents updated plan to user for another review cycle

Iteration limits:
- Maximum 6 annotation cycles per plan
- After each cycle, present options: Proceed / Annotate Again / Save Draft / Cancel
- Track iteration count and display: "Annotation cycle {N}/6"

Guard rule: [HARD] During annotation cycles, the explicit instruction "DO NOT implement any code — only update the plan document" MUST be included in every agent prompt. This prevents premature code generation.

---

**Next phase:** Read `workflows/plan/spec-assembly.md` to continue with Phase 1.5 Pre-Creation Validation Gate.
