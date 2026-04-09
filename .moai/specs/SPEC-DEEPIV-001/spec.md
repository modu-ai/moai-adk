---
id: SPEC-DEEPIV-001
version: "1.0.0"
status: draft
created: "2026-04-07"
updated: "2026-04-07"
author: GOOS
priority: P1
issue_number: 0
---

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-04-07 | 1.0.0 | Initial draft |

---

# SPEC-DEEPIV-001: Deep Interview (심층 인터뷰)

## Overview

`/moai plan`과 `/moai project` 워크플로우에 소크라테스식 요구사항 명확화 단계를 자동 통합. 모호한 요구사항을 구조화된 질문-답변 루프로 구체화하여 SPEC 품질을 향상.

## Motivation

현재 `/moai plan`은 사용자 입력을 그대로 SPEC 생성에 사용. 모호한 요구사항("기능 개선")은 잘못된 SPEC으로 이어지고, 이는 구현 단계에서 재작업을 유발. 사전에 요구사항을 명확화하면 전체 파이프라인 효율이 향상됨.

## Requirements (EARS Format)

### REQ-DEEPIV-001 (Event-Driven)
When `/moai plan` is invoked, the system shall evaluate the clarity score (1-10) of the user's input before proceeding to Phase 0.5 (Deep Research).

### REQ-DEEPIV-002 (State-Driven)
When the clarity score is 4 or higher (ambiguous), the system shall automatically start Deep Interview before Phase 0.5.

### REQ-DEEPIV-003 (State-Driven)
When the clarity score is 3 or lower (clear), the system shall skip Deep Interview and proceed directly to Phase 0.5.

### REQ-DEEPIV-004 (Ubiquitous)
Each Deep Interview round shall present 2-3 questions via AskUserQuestion with recommended multiple-choice options AND a free-text input option.

### REQ-DEEPIV-005 (Ubiquitous)
The Deep Interview shall complete within a maximum of 5 rounds for `/moai plan` and 3 rounds for `/moai project`.

### REQ-DEEPIV-006 (Event-Driven)
When `/moai project` is invoked for a new project, the system shall use Deep Interview to replace the current 4-question sequence (Phase 0.5).

### REQ-DEEPIV-007 (Event-Driven)
When `/moai project` is invoked for an existing project, the system shall use Deep Interview between Phase 1 (codebase analysis) and Phase 2 (user confirmation).

### REQ-DEEPIV-008 (Ubiquitous)
The Deep Interview shall produce an `interview.md` artifact in the SPEC directory (plan) or `.moai/project/interview.md` (project).

### REQ-DEEPIV-009 (Optional)
When the user includes `--skip-interview` flag, the Deep Interview shall be bypassed.

### REQ-DEEPIV-010 (Ubiquitous)
Each AskUserQuestion call shall follow MoAI constraints: max 4 options, no emoji, first option is recommended, conversation_language.

## Affected Files

### New Files
- `internal/template/templates/.moai/config/sections/interview.yaml` - Configuration

### Modified Files
- `internal/template/templates/.claude/skills/moai/workflows/plan.md` - Add Phase 0.3
- `internal/template/templates/.claude/skills/moai/workflows/project.md` - Add Deep Interview phase

## Technical Design

### Clarity Score Evaluation (Phase 0.3)

MoAI evaluates the user's input text against these criteria:

| Criterion | Score Impact |
|-----------|-------------|
| Technical keywords >= 3 | -3 (clearer) |
| Specific action verb (add, remove, migrate, refactor) | -2 (clearer) |
| File/module name mentioned | -2 (clearer) |
| SPEC-ID present (resume) | Skip entirely |
| Generic nouns only (feature, improvement, update) | +3 (more ambiguous) |
| No scope boundary (no module/file mentioned) | +2 (more ambiguous) |
| No success criteria mentioned | +1 (more ambiguous) |

Base score: 5. Adjust by criteria. Clamp to 1-10.

### AskUserQuestion Format (Deep Interview)

Each question follows this pattern:

```
Question: [Socratic question in conversation_language]
Options:
- [Recommended answer based on codebase research] (Recommended): [Description of what this means and its implications]
- [Alternative answer 2]: [Description]
- [Alternative answer 3]: [Description]
- Type your answer: Enter a custom response for questions not covered by the options above
```

The 4th option "Type your answer" enables free-text input. When selected, MoAI follows up with a text prompt.

### Interview Rounds by Workflow

**`/moai plan` (max 5 rounds)**:

| Round | Focus | Question Examples |
|-------|-------|-------------------|
| 1 | Scope | "What specifically do you want to change? (module, feature, behavior)" |
| 2 | Constraints | "What systems/modules are affected? What must NOT change?" |
| 3 | Success | "How will you verify this is working correctly?" |
| 4 | Edge cases | "What happens when [boundary condition]?" |
| 5 | Priority | "What is the most critical requirement if scope must be reduced?" |

**`/moai project` (max 3 rounds)**:

| Round | Focus | Question Examples |
|-------|-------|-------------------|
| 1 | Vision | "What is the core purpose? Who are the primary users?" |
| 2 | Technology | "What tech stack? What external integrations?" |
| 3 | Scope | "What is the MVP? What are the current priorities?" |

### Re-evaluation After Each Round

After each round, re-calculate clarity score with new information. If score drops to 3 or below, interview ends early.

### Integration in plan.md

Insert **Phase 0.3** between current Phase 1A and Phase 0.5:

```
Phase 1A: Project Exploration (existing)
  |
Phase 0.3: Clarity Evaluation (NEW)
  |-- Score <= 3: Skip to Phase 0.5
  |-- Score >= 4: Enter Deep Interview
  |
Phase 0.3.1: Deep Interview Loop (NEW, max 5 rounds)
  |-- Each round: AskUserQuestion with 3 choices + free-text
  |-- Re-evaluate clarity after each round
  |-- Output: .moai/specs/SPEC-{ID}/interview.md
  |
Phase 0.5: Deep Research (existing, now receives interview.md as context)
```

### Integration in project.md

**New Project**: Replace Phase 0.5 Questions 1-4 with Deep Interview (3 rounds):
```
Phase 0: Project Type Detection (existing)
  |
Phase 0.3: Deep Interview (NEW, replaces Phase 0.5)
  |-- Round 1: Vision + project type
  |-- Round 2: Technology + language
  |-- Round 3: Features + scope
  |-- Output: .moai/project/interview.md
  |
Phase 1: Codebase Analysis (skipped for new projects)
```

**Existing Project**: Insert between Phase 1 and Phase 2:
```
Phase 1: Codebase Analysis (existing)
  |
Phase 1.5: Deep Interview (NEW, max 3 rounds)
  |-- Questions informed by codebase analysis results
  |-- Output: .moai/project/interview.md
  |
Phase 2: User Confirmation (existing, now includes interview findings)
```

### Configuration (interview.yaml)

```yaml
interview:
  enabled: true
  clarity_threshold: 4
  plan:
    max_rounds: 5
    questions_per_round: 3
  project:
    max_rounds: 3
    questions_per_round: 3
  skip_conditions:
    - resume_spec_id_present
    - skip_interview_flag
    - technical_keywords_gte_5
```

## Dependencies

- None (independent skill-level change)

## Non-Goals

- Replacing existing AskUserQuestion decision points
- Adding new Claude Code hook events
- Modifying Go binary code (skill-only change)
