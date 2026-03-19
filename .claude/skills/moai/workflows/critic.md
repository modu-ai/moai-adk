---
name: moai-workflow-critic
description: >
  Multi-perspective SPEC critique workflow. Coordinates 4 critic agents
  to critique SPEC documents from tech, business, user, and ops perspectives.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-03-19"
  tags: "critic, critique, review, spec, devil's advocate"

progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

triggers:
  keywords: ["critic", "critique", "review spec", "devil's advocate", "크리틱", "비판"]
  agents: ["manager-critic", "critic-tech", "critic-business", "critic-user", "critic-ops"]
  phases: ["critic"]
---

# Critic Workflow Orchestration

## Purpose

Generate multi-perspective critical questions about SPEC documents before implementation.
Inserts between Plan and Run: PLAN → CRITIC → RUN → SYNC.

## Scope

- Validates SPEC worth before investing implementation effort
- Generates critic.md as a companion artifact to spec.md
- Optional: can be skipped at user's choice

## Input

- $ARGUMENTS: SPEC-ID to critique (e.g., SPEC-AUTH-001)
- --skip-persona tech|business|user|ops: Skip specific critic
- --auto: Auto-dismiss LOW severity questions

## Context Loading

Before execution, load these files:
- .moai/specs/SPEC-{ID}/spec.md
- .moai/specs/SPEC-{ID}/plan.md
- .moai/specs/SPEC-{ID}/acceptance.md
- .moai/config/sections/lessons.yaml

---

## Phase Sequence

### Phase 1: SPEC Validity Check

Verify required files exist in .moai/specs/SPEC-{ID}/:
- spec.md (required)
- plan.md (required)
- acceptance.md (required)

If any missing: Report error, suggest running `/moai plan` first, exit.

Create SPEC summary (~1K tokens) for critic agents:
- Key requirements from spec.md EARS structure
- Implementation approach from plan.md
- Success criteria from acceptance.md

### Phase 2: Parallel Critic Invocation

Agent: manager-critic subagent

The manager-critic orchestrates 4 critic agents in parallel:

1. **critic-tech** (haiku, permissionMode: plan):
   - Prompt: "Analyze this SPEC from a technical architecture perspective. Focus on architecture risks, failure scenarios, and scalability concerns."
   - Input: SPEC summary
   - Output: 3 questions with severity (HIGH/MEDIUM/LOW) and category

2. **critic-business** (haiku, permissionMode: plan):
   - Prompt: "Analyze this SPEC from a business value perspective. Focus on revenue impact, cost-benefit, and market fit."
   - Input: SPEC summary
   - Output: 3 questions with severity and category

3. **critic-user** (haiku, permissionMode: plan):
   - Prompt: "Analyze this SPEC from a user experience perspective. Focus on user value, UX impact, and adoption barriers."
   - Input: SPEC summary
   - Output: 3 questions with severity and category

4. **critic-ops** (haiku, permissionMode: plan):
   - Prompt: "Analyze this SPEC from an operations perspective. Focus on operational costs, monitoring needs, and deployment risks."
   - Input: SPEC summary
   - Output: 3 questions with severity and category

Skip any persona specified in --skip-persona flag.

Total: Up to 12 questions generated in parallel (~27K token budget).

### Phase 3: Question Consolidation

Manager-critic consolidates results:
- Merge all questions (up to 12)
- Remove semantic duplicates
- Sort by severity: HIGH → MEDIUM → LOW
- Assign sequential IDs: Q1 through Q12

### Phase 4: User Q&A Session

Tool: AskUserQuestion (via manager-critic)

For each question, present:
```
[SEVERITY] Category — Critic Perspective
Question: {question text}
Context: {why this matters}
```

User options per question:
- **Answer** (Recommended): Provide response → recorded in critic report
- **Dismiss**: Skip this question → recorded as dismissed
- **Modify SPEC**: This reveals a gap → flag for SPEC update

If --auto flag is set: All LOW severity questions auto-dismissed.

#### Consensus Mode Integration

When invoked from plan.md `--consensus` flow:
- Phase 4 runs with auto-dismiss for LOW questions
- "Modify SPEC" selections are collected but NOT immediately applied — they are returned to the plan.md consensus loop (Decision Point 2.6) for batch application
- The critic report includes a `consensus_modifications` field listing all requested SPEC changes

### Phase 5: Critic Report Generation

Generate `.moai/specs/SPEC-{ID}/critic.md`:

```markdown
---
spec_id: SPEC-{ID}
critic_date: {ISO 8601}
total_questions: 12
answered: 8
dismissed: 4
spec_modifications: 2
---

# CRITIC Report

## Summary
- Total Questions: 12
- Answered: 8
- Dismissed: 4
- SPEC Modifications Suggested: 2

## Tech Critic

### Q1: [HIGH] Architecture Risk
**Question**: {question text}
**Answer**: {user answer}
**Impact on SPEC**: {none / Added requirement R-XXX}

### Q2: [MEDIUM] Scalability
**Question**: {question text}
**Status**: Dismissed

## Business Critic
...

## User Critic
...

## Ops Critic
...
```

### Phase 6: Learning System Integration

If .moai/config/sections/lessons.yaml has `enabled: true`:
- Record high-value patterns (questions → SPEC modifications)
- Record low-value patterns (consistently dismissed questions)
- Append to .moai/lessons/lessons.jsonl

---

## Decision Points

### Decision Point: Post-Critic Action
Tool: AskUserQuestion

Options:
- **Proceed to Implementation** (Recommended): Continue to /moai run SPEC-{ID}
- **Modify SPEC**: Return to /moai plan to update the SPEC based on critic findings
- **Re-Critique**: Run critic again with different perspectives
- **Cancel**: Archive critic report and exit

---

## Completion Criteria

- All selected critic agents completed successfully
- All questions presented to user with responses collected
- critic.md generated in .moai/specs/SPEC-{ID}/
- Learning patterns recorded (if lessons system enabled)

---

Version: 1.1.0
Updated: 2026-03-19
Changes: Added Consensus Mode Integration for --consensus flow from plan.md. Added consensus_modifications field to critic report.
