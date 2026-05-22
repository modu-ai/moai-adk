---
name: moai-workflow-spec
description: >
  SPEC workflow orchestration with EARS format requirements, acceptance criteria,
  and Plan-Run-Sync integration for MoAI-ADK development. Use when creating SPEC
  documents or defining acceptance criteria.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Bash(git:*), Bash(ls:*), Bash(wc:*), Bash(mkdir:*), Grep, Glob, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.2.0"
  category: "workflow"
  status: "active"
  updated: "2026-01-08"
  modularized: "true"
  tags: "workflow, spec, ears, requirements, moai-adk, planning"
  author: "MoAI-ADK Team"
  context: "fork"
  agent: "Plan"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["SPEC", "requirement", "EARS", "acceptance criteria", "user story", "planning", "specification", "requirements gathering"]
  phases: ["plan"]
  agents: ["manager-spec", "manager-strategy", "Plan"]
---

# SPEC Workflow Management

## Quick Reference

SPEC Workflow Orchestration using EARS format for systematic requirement definition and Plan-Run-Sync workflow integration.

Core Capabilities:

- EARS Format Specifications: Five requirement patterns for unambiguous requirements
- Requirement Clarification: Four-step systematic process with assumption analysis
- SPEC Document Templates: Standardized 3-file structure (spec.md / plan.md / acceptance.md)
- Plan-Run-Sync Integration: Seamless workflow connection
- Parallel Development: Git Worktree-based SPEC isolation
- Quality Gates: TRUST 5 framework validation

EARS Five Patterns:

| Pattern | Format | Use |
|---------|--------|-----|
| Ubiquitous | "The system shall always X" | Always active |
| Event-Driven | "WHEN event THEN action" | Trigger-response |
| State-Driven | "IF condition THEN action" | Conditional behavior |
| Unwanted | "The system shall not X" | Prohibition |
| Optional | "Where possible, provide X" | Nice-to-have |

When to Use:

- Feature planning and requirement definition
- SPEC document creation and maintenance
- Parallel feature development coordination
- Quality assurance and validation planning
- Requirements gathering from user story narratives

Quick Commands:

```bash
/moai:1-plan "user authentication system"                   # Create new SPEC
/moai:1-plan "login" "signup" --worktree                    # Parallel SPECs
/moai:1-plan "payment processing" --branch                  # New branch
/moai:1-plan SPEC-001 "add OAuth support"                   # Update existing
```

---

## Implementation Guide

### Core Concepts

SPEC-First Development Philosophy:

- EARS format ensures unambiguous requirements
- Requirement clarification prevents scope creep
- Systematic validation through test scenarios
- Integration with DDD workflow for implementation
- Quality gates enforce completion criteria
- Constitution reference ensures project-wide consistency

### Constitution Reference (SDD 2025 Standard)

Constitution defines the project DNA that all SPECs must respect. Before creating any SPEC, verify alignment with `.moai/project/tech.md`.

Constitution Components: Technology Stack, Naming Conventions, Forbidden Libraries, Architectural Patterns, Security Standards, Logging Standards.

Constitution Verification: All SPEC technology choices align with Constitution stack versions, no forbidden libraries, naming conventions respected, architectural boundaries preserved.

WHY: Constitution prevents architectural drift and ensures maintainability.

### SPEC Workflow Stages

| Stage | Activity |
|-------|----------|
| 1 | User Input Analysis — parse natural-language feature description |
| 2 | Requirement Clarification — 4-step systematic process |
| 3 | EARS Pattern Application — structure requirements using five patterns |
| 4 | Success Criteria Definition — establish completion metrics |
| 5 | Test Scenario Generation — create verification test cases |
| 6 | SPEC Document Generation — produce standardized markdown |

### EARS Format

Five patterns cover all requirement types. Each pattern has a specific use case and test strategy.

See [EARS deep dive with examples per pattern](references/ears-deep-dive.md) for use cases, examples, and test strategies for Ubiquitous, Event-Driven, State-Driven, Unwanted, and Optional requirements.

### Requirement Clarification Process

5-step systematic process:

- Step 0: Assumption Analysis (Philosopher Framework) — surface technical, business, team, integration assumptions
- Step 0.5: Root Cause Analysis (Five Whys) — surface problem to root cause for problem-driven SPECs
- Step 1: Scope Definition — supported methods, validation rules, failure handling, session management
- Step 2: Constraint Extraction — performance, security, compatibility, scalability
- Step 3: Success Criteria — coverage targets, response time percentiles, functional completion, quality gates
- Step 4: Test Scenario Creation — normal, error, edge, security cases

See [requirement clarification detailed workflow](references/requirement-clarification.md) for assumption documentation templates and Five Whys application.

### Plan-Run-Sync Workflow Integration

PLAN (/moai:1-plan): manager-spec analyzes input → EARS requirements → clarification → SPEC creation in `.moai/specs/` → optional `--branch` or `--worktree`.

RUN (/moai:2-run): manager-develop loads SPEC → ANALYZE-PRESERVE-IMPROVE (DDD) or RED-GREEN-REFACTOR (TDD) per `quality.yaml development_mode` → moai-workflow-testing reference → expert agent delegation → manager-quality validation.

SYNC (/moai:3-sync): manager-docs synchronizes documentation → API docs from SPEC → README and architecture updates → CHANGELOG → version control commit.

### Parallel Development with Git Worktree

Worktree provides isolated working directories per SPEC for parallel development without branch switching. Benefits: parallel development, clear ownership boundaries, dependency isolation, risk reduction.

See [worktree workflow patterns](references/worktree-workflow.md) for creation commands and team collaboration examples.

---

## Resources

### SPEC File Organization

Standard 3-File Format:

- `.moai/specs/SPEC-{ID}/spec.md` — EARS format specification
- `.moai/specs/SPEC-{ID}/plan.md` — implementation plan, milestones, technical approach
- `.moai/specs/SPEC-{ID}/acceptance.md` — acceptance criteria, Given-When-Then scenarios

[HARD] Every SPEC directory MUST contain all 3 files. Missing files create incomplete requirements.

State files: `.moai/state/last-session-state.json`. Generated docs: `.moai/docs/api-documentation.md`.

### SPEC Metadata Schema

Canonical 12 required fields (enforced by `internal/spec/lint.go` `FrontmatterSchemaRule`): id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags.

Status enum (8 values): draft → planned → in-progress → implemented → completed | superseded | archived | rejected.

Optional fields: issue_number, depends_on, lint.skip, bc_id, tier (S/M/L LEAN tier).

Full schema at `.claude/rules/moai/development/spec-frontmatter-schema.md` (SSOT).

### SPEC Lifecycle Management

Three lifecycle levels:

| Level | Description | Maintenance |
|-------|-------------|-------------|
| spec-first | SPEC discarded after implementation | None |
| spec-anchored | SPEC maintained alongside implementation | Quarterly review |
| spec-as-source | SPEC is single source of truth, only SPEC edited by humans | Changes regenerate impl |

Transitions: spec-first → spec-anchored when production-critical, spec-anchored → spec-as-source when compliance or regeneration workflow required. Downgrade requires explicit justification.

### Quality Metrics

SPEC Quality Indicators: requirement clarity (all EARS patterns used), test coverage (all requirements have scenarios), constraint completeness, success criteria measurability.

Validation Checklist: All EARS requirements testable, no ambiguous language ("should", "might", "usually"), all error cases documented, performance targets quantified, security requirements OWASP-compliant.

### Token Management

| Phase | Token Budget |
|-------|--------------|
| PLAN | ~30% |
| RUN | ~60% |
| SYNC | ~10% |

Context Optimization: SPEC document persists in `.moai/specs/`. Session state in `.moai/state/`. Minimal context transfer through SPEC ID reference. Agent delegation reduces token overhead.

---

## SPEC Scope and Classification

### What Belongs in .moai/specs/

The `.moai/specs/` directory is EXCLUSIVELY for SPEC documents that define features to be implemented.

Valid SPEC Content: feature requirements in EARS format, implementation plans with milestones, acceptance criteria with Given/When/Then scenarios, technical specifications for new functionality, user stories with clear deliverables.

SPEC Characteristics: forward-looking (what WILL be built), actionable, testable, structured (EARS).

### What Does NOT Belong in .moai/specs/

| Document Type | Why Not SPEC | Correct Location |
|---------------|--------------|------------------|
| Security Audit | Analyzes existing code | `.moai/reports/security-audit-{DATE}/` |
| Performance Report | Documents current metrics | `.moai/reports/performance-{DATE}/` |
| Dependency Analysis | Reviews existing dependencies | `.moai/reports/dependency-review-{DATE}/` |
| Architecture Overview | Documents current state | `.moai/docs/architecture.md` |
| API Reference | Documents existing APIs | `.moai/docs/api-reference.md` |
| Meeting Notes | Records decisions made | `.moai/reports/meeting-{DATE}/` |
| Retrospective | Analyzes past work | `.moai/reports/retro-{DATE}/` |

### Exclusion Rules

[HARD] Reports analyze what EXISTS → `.moai/reports/`. SPECs define what will be BUILT → `.moai/specs/`.

[HARD] Documentation explains HOW TO USE → `.moai/docs/`. SPECs define WHAT TO BUILD → `.moai/specs/`.

---

## Works Well With

- moai-foundation-core: SPEC-First DDD methodology and TRUST 5 framework
- moai-workflow-testing: DDD implementation and test automation
- moai-workflow-project: Project initialization and configuration
- moai-workflow-worktree: Git Worktree management for parallel development
- manager-spec: SPEC creation and requirement analysis agent
- manager-develop: DDD/TDD implementation based on SPEC requirements
- manager-quality: TRUST 5 quality validation and gate enforcement

For migration scenarios and validation scripts: [reference/migration-guide.md](reference/migration-guide.md).

---

Version: 1.3.1 (SPEC-V3R6-SKILL-COMPRESS-001 body compression)
Last Updated: 2026-05-23
Integration Status: Complete - Plan-Run-Sync workflow with SDD 2025 features

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "The SPEC is obvious, I can skip EARS format" | EARS exists because obvious requirements are the first to be misinterpreted. The format forces disambiguation. |
| "Acceptance criteria are redundant with the requirements" | Requirements describe intent. Acceptance criteria describe observable evidence. Both are needed. |
| "I will refine the SPEC during implementation" | Late refinement means wasted implementation. SPEC is the cheap place to change your mind. |
| "Research is a nice-to-have, not a blocker" | Skipping research produces SPECs that conflict with existing code. research.md prevents rework. |
| "Annotation cycle is just user friction" | Annotation catches misunderstandings before code is written. It is the cheapest feedback loop in the pipeline. |
| "This SPEC is small, I do not need a separate file" | Every SPEC is a persistent contract. In-message SPECs cannot be referenced by /moai run SPEC-XXX. |

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- Requirements written in imperative prose instead of EARS (WHEN X, SHALL Y)
- Acceptance criteria phrased as subjective judgments ("feels fast", "looks clean")
- SPEC document missing research.md sibling when modifying existing code
- Annotation cycle skipped or reduced to a single-turn "looks good"
- Requirements use "should" where they mean "shall" (optional vs mandatory ambiguity)
- SPEC-ID not registered in `.moai/specs/` directory

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] SPEC file exists at `.moai/specs/SPEC-XXX/spec.md` with unique ID
- [ ] Every requirement uses EARS keywords (WHEN, WHILE, WHERE, IF, SHALL)
- [ ] Every acceptance criterion is observable (test output, file existence, metric threshold)
- [ ] research.md exists when the SPEC touches existing code
- [ ] Annotation cycle completed with explicit user approval marker
- [ ] SPEC references existing SPEC-IDs it depends on or supersedes
- [ ] Non-goals section present to prevent scope creep

<!-- moai:evolvable-end -->
