---
name: moai-workflow-spec
description: >
  SPEC workflow orchestration with EARS format requirements, acceptance criteria,
  and Plan-Run-Sync integration for MoAI-ADK development. Use when creating SPEC
  documents or defining acceptance criteria.

when_to_use: >
  Use for SPEC workflow orchestration: EARS-format requirements,
  acceptance criteria, user stories, requirements gathering, planning, and
  Plan-Run-Sync integration for MoAI-ADK development.

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

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000
---

# SPEC Workflow Management

## Quick Reference

SPEC Workflow Orchestration using GEARS notation (current) — backed by the EARS legacy backward-compatibility window — for systematic requirement definition and Plan-Run-Sync workflow integration.

Lint behavior canonicalized per the GEARS migration policy.

Core Capabilities:

- GEARS-Format Specifications (current): Five requirement patterns with the unified compound clause `[Where ...][While ...][When ...] The <subject> shall <behavior>` and a generalized `<subject>` (any noun, not only "the system")
- EARS Legacy Reference: All EARS patterns preserved per the lint engine's backward-compatibility policy to keep pre-v3 SPECs (those authored before GEARS became canonical) readable
- Requirement Clarification: Four-step systematic process with assumption analysis
- SPEC Document Templates: Standardized 3-file structure (spec.md / plan.md / acceptance.md)
- Plan-Run-Sync Integration: Seamless workflow connection
- Parallel Development: Git Worktree-based SPEC isolation
- Quality Gates: TRUST 5 framework validation

GEARS Five Patterns (current notation):

| Pattern | GEARS form (current) | EARS form (legacy) | Notes |
|---------|----------------------|--------------------|-------|
| Ubiquitous | "The <subject> shall <behavior>" | "The system shall <behavior>" | `<subject>` may be any noun: system, component, service, agent, function, artifact |
| Event-driven | "**When** <event-detected>, the <subject> shall <behavior>" | "WHEN <event>, the system shall <action>" | Unchanged trigger semantics |
| State-driven | "**While** <state>, the <subject> shall <behavior>" | "WHILE <state>, the system shall <action>" | Unchanged — promoted as a first-class pattern |
| Capability gate | "**Where** <capability / feature flag / static config>, the <subject> shall <behavior>" | "WHERE <feature exists>, the system shall <action>" | Reframed — represents capability gate / feature flag / static config (no longer "Optional") |
| Event-detected (replaces IF/THEN) | "**When** <undesired-condition-detected>, the <subject> shall <response>" | `IF <condition> THEN <action>` **[DEPRECATED — use WHEN <event-detected>]** | The `IF/THEN` modality was removed; describe the same intent as a detected event |

Unified compound clause: `**Where** <precondition> **While** <state> **When** <event> the <subject> shall <behavior>` — any subset of the three modifiers may chain.

See [GEARS notation reference](https://adk.mo.ai.kr/en/workflow-commands/moai-plan/#gears-notation).

> **IF/THEN deprecated callout**: Authoring guidance previously used `IF <condition> THEN <action>` to describe state-conditioned behavior. In GEARS that intent is expressed as `When <condition-detected>` (event-detected form). The lint engine emits a `LegacyEARSKeyword` warning (non-strict) or error (`moai spec lint --strict`) on residual `IF/THEN` in new SPECs. The 6-month backward-compatibility window remains active for legacy SPECs.

Generalized subject substitution: GEARS replaces the hardcoded "the system" subject with `<subject>`, which may be any noun. Authors writing NEW SPECs MAY use the generalized form. Examples of valid non-"the system" subjects:

- "The skill shall present GEARS as the primary notation." (Ubiquitous, `<subject>` = skill)
- "The agent shall return a blocker report instead of prompting the user." (Ubiquitous, `<subject>` = agent)
- "**When** a SPEC author opens the file, the component shall display the deprecation banner." (Event-driven, `<subject>` = component)

Pre-v3 SPECs (those authored before GEARS became canonical) keep "The system" as the default subject for readability; existing readers do not need to relearn the canonical phrase.

EARS Five Patterns (legacy — 6-month backward-compatibility window):

| Pattern | Format | Use |
|---------|--------|-----|
| Ubiquitous | "The system shall always X" | Always active |
| Event-Driven | "WHEN event THEN action" | Trigger-response |
| State-Driven | "WHILE state, the system shall ..." | Conditional behavior (use `WHILE`, not legacy `IF/THEN`) |
| Unwanted | "The system shall not X" | Prohibition |
| Optional | "Where possible, provide X" | Nice-to-have |

The legacy `IF/THEN` modality is replaced by GEARS `When <event-detected>` — see callout above.

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

### GEARS Format (current)

GEARS (Generalized EARS) is the canonical SPEC notation as of v3.0.0. It preserves Ubiquitous / `When` (event-driven) / `While` (state-driven) and reframes `Where` as a capability gate. The legacy `IF/THEN` modality is replaced by `When <event-detected>`.

GEARS notation is exhaustively described in [docs-site GEARS notation reference](https://adk.mo.ai.kr/en/workflow-commands/moai-plan/#gears-notation) and the canonical GEARS migration policy record.

Compound clause example (with non-"the system" subject):

> **Where** the project is initialized **While** strict mode is active **When** a SPEC author runs `moai spec lint`, the lint engine shall emit a `LegacyEARSKeyword` finding for every residual `IF/THEN` modality.

This example chains all three GEARS modifiers (`Where`, `While`, `When`) and uses `<subject>` = "lint engine" rather than "the system".

### EARS Format (legacy — 6-month backward-compatibility window)

Five patterns cover all requirement types. Each pattern has a specific use case and test strategy. Pre-v3 SPECs (those authored before GEARS became canonical) continue to use EARS notation and remain valid per the lint engine's backward-compatibility policy.

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

### [NEEDS CLARIFICATION] Marker Convention

**[NEEDS CLARIFICATION: <topic>]** markers identify unresolved questions in plan.md and research.md that MUST be settled before Implementation Kickoff Approval (plan→run HUMAN GATE).

**Placement**: ONLY in plan.md and research.md (NEVER in spec.md or acceptance.md).

**Format**: 
- `[NEEDS CLARIFICATION: <specific topic>]` — inline marker for open questions
- Each marker MUST be addressable via orchestrator AskUserQuestion before run-phase entry
- plan-auditor detects unclarified markers and flags as "clarification gate" finding

**3-Layer Distinction**:
- `[NEEDS CLARIFICATION: <topic>]` — plan/research artifact blocker (user Q required)
- `TODO` — code-level implementation debt (no user Q needed)
- `@MX:TODO` — code-level annotation for untested/incomplete code

**Processing**:
- plan-auditor scans for `[NEEDS CLARIFICATION]` markers during audit
- If any remain, plan-auditor recommends resolution before Implementation Kickoff Approval
- Orchestrator runs AskUserQuestion rounds to resolve each marked topic
- Implementation Kickoff Approval (mandatory human gate) proceeds only after all clarifications are resolved

### Plan-Run-Sync Workflow Integration

PLAN (/moai:1-plan): manager-spec analyzes input → EARS requirements → clarification → SPEC creation in `.moai/specs/` → optional `--branch` or `--worktree`.

RUN (/moai:2-run): manager-develop loads SPEC → ANALYZE-PRESERVE-IMPROVE (DDD) or RED-GREEN-REFACTOR (TDD) per `quality.yaml` `constitution.development_mode` → moai-workflow-testing reference → per-spawn Agent(general-purpose) domain delegation → quality-gate validation (Stop hook / /moai gate).

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

Canonical 12 required fields (enforced by the SPEC frontmatter lint rule): id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags.

Status enum (8 values): draft → in-progress → implemented → completed | superseded | archived | rejected. (`planned` is retained in the enum as legacy-optional — NOT in the active flow; no agent authors a `draft → planned` transition. See `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum.)

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

### Out of Scope Classification Rules

These routing rules decide what is out of scope for a SPEC document (and where it belongs instead). When authoring a SPEC's own exclusions section, express each excluded item as a `### Out of Scope — <topic>` H3 sub-heading with `-` bullets so the section satisfies the `OutOfScopeRule` lint.

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
- /moai gate skill (or sync-phase-quality-gate.sh Stop hook): TRUST 5 quality validation and gate enforcement (former manager-quality role)

For migration scenarios and validation scripts: [reference/migration-guide.md](reference/migration-guide.md).

---

Version: 1.3.1 (skill body compression pass)
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
- [ ] Out of Scope section present to prevent scope creep — at least one `### Out of Scope — <topic>` H3 sub-heading with a `-` bullet entry (satisfies the `OutOfScopeRule` lint)

<!-- moai:evolvable-end -->
