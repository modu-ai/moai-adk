---
name: moai-workflow-challenge
description: >
  Multi-perspective SPEC critique system. Launches 4 critic agents (tech, business,
  user, ops) to generate critical questions about SPEC documents before implementation.
  Validates that what you're building is worth building.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-03-19"
  tags: "challenge, critique, review, devil's advocate, pre-mortem, spec review"

progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

triggers:
  keywords: ["challenge", "critique", "review spec", "devil's advocate", "챌린지", "비판"]
  agents: ["manager-challenge", "critic-tech", "critic-business", "critic-user", "critic-ops"]
  phases: ["challenge"]
---

# Challenge Workflow Orchestration

## Purpose

Generate multi-perspective critical questions about SPEC documents before implementation.
Validates that what you're building is worth building from 4 angles:
tech architecture, business value, user experience, and operations.

## Scope

- Inserts between Plan and Run phases: PLAN → CHALLENGE → RUN
- Reads SPEC documents created by /moai plan
- Generates challenge.md report in the SPEC directory
- Optional: can be skipped via Decision Point

## Input

- $ARGUMENTS: SPEC-ID to challenge (e.g., SPEC-AUTH-001)
- --skip-persona tech|business|user|ops: Skip specific critic perspective
- --auto: Automatically dismiss LOW severity questions

## Context Loading

Before execution, load these essential files:
- .moai/specs/SPEC-{ID}/spec.md (SPEC document)
- .moai/specs/SPEC-{ID}/plan.md (implementation plan)
- .moai/specs/SPEC-{ID}/acceptance.md (acceptance criteria)
- .moai/config/sections/lessons.yaml (learning system config)

---

## Phase Sequence

### Phase 1: SPEC Validity Check

Verify required files exist:
- .moai/specs/SPEC-{ID}/spec.md
- .moai/specs/SPEC-{ID}/plan.md
- .moai/specs/SPEC-{ID}/acceptance.md

If any missing: Report error and exit.

Create SPEC summary (~1K tokens) for critic agents:
- Extract key requirements from spec.md
- Extract implementation approach from plan.md
- Extract success criteria from acceptance.md

### Phase 2: Parallel Critic Invocation

Launch 4 critic agents in parallel via Agent() tool:

1. **critic-tech** (haiku, ~3K budget):
   Input: SPEC summary + "Analyze from technical architecture perspective"
   Output: 3 questions with severity and category

2. **critic-business** (haiku, ~3K budget):
   Input: SPEC summary + "Analyze from business value perspective"
   Output: 3 questions with severity and category

3. **critic-user** (haiku, ~3K budget):
   Input: SPEC summary + "Analyze from user experience perspective"
   Output: 3 questions with severity and category

4. **critic-ops** (haiku, ~3K budget):
   Input: SPEC summary + "Analyze from operations perspective"
   Output: 3 questions with severity and category

Skip any critic specified in --skip-persona flag.

### Phase 3: Question Consolidation

- Collect all questions (up to 12)
- Remove semantic duplicates (similar questions from different critics)
- Sort by severity: HIGH → MEDIUM → LOW
- Assign unique IDs: Q1, Q2, ... Q12

### Phase 4: User Q&A Session

Present questions to user via AskUserQuestion:

For each question (grouped by severity):
- Display: [SEVERITY] Category - Critic Perspective
- Display: Question text and context
- Options:
  - Answer: User provides response → recorded in report
  - Dismiss: Question noted but not answered → recorded as dismissed
  - Modify SPEC: This question reveals a gap → flag for SPEC modification

If --auto flag: Automatically dismiss all LOW severity questions.

### Phase 5: Challenge Report Generation

Generate `.moai/specs/SPEC-{ID}/challenge.md`:

```yaml
---
spec_id: SPEC-{ID}
challenge_date: {ISO 8601 timestamp}
total_questions: {N}
answered: {N}
dismissed: {N}
spec_modifications: {N}
---
```

Followed by questions grouped by critic perspective with answers.

### Phase 6: Learning System Integration

Record patterns for the lessons system:
- Questions answered with SPEC modifications → high-value critique patterns
- Questions consistently dismissed → low-value patterns
- Save to .moai/lessons/lessons.jsonl if lessons system is enabled

---

## Completion Criteria

- All critic agents completed successfully
- Questions presented to user and responses collected
- challenge.md generated in SPEC directory
- Learning patterns recorded (if lessons enabled)

---

Version: 1.0.0
Updated: 2026-03-19
