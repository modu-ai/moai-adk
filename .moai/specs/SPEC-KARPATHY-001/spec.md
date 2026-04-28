---
id: SPEC-KARPATHY-001
version: "0.1.0"
status: completed
created_at: 2026-04-28
updated_at: 2026-04-28
author: manager-spec
priority: High
labels: [karpathy, anti-patterns, constitution, skills, rules, coding-principles, templates]
issue_number: null
title: "Karpathy Coding Principles Integration into MoAI-ADK Orchestration"
breaking: false
bc_id: []
lifecycle: spec-anchored
related_specs: [SPEC-V3R3-PATTERNS-001, SPEC-V3R2-CON-001]
complexity: 6
harness_level: standard
development_mode: ddd
domains: [rules, skills, templates]
---

# SPEC-KARPATHY-001: Karpathy Coding Principles Integration into MoAI-ADK Orchestration

## HISTORY

| Version | Date       | Author       | Description                                                                                          |
|---------|------------|--------------|------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-04-28 | manager-spec | Initial draft. Integrate Andrej Karpathy's 4 coding principles + anti-pattern catalog into MoAI-ADK. |
| 1.0.0   | 2026-04-28 | moai         | Implementation complete. §1.2 Non-Goals updated (NOTICE.md attribution added per user request). |

---

## 1. Goal (Purpose)

Integrate Andrej Karpathy's 4 coding principles (Think Before Coding, Simplicity First, Surgical Changes, Goal-Driven Execution) plus a concrete anti-pattern catalog into MoAI-ADK's orchestration layer. The integration closes identified coverage gaps between Karpathy's principles and MoAI's existing Agent Core Behaviors through targeted constitution amendments, a new anti-pattern reference skill, and a quick-reference rule file.

### 1.1 Background

The `forrestchang/andrej-karpathy-skills` repository (94K stars) packages Karpathy's coding philosophy into 4 principles with an EXAMPLES.md (~300 lines) containing 8 concrete wrong-vs-right anti-pattern categories.

GAP analysis between Karpathy principles and MoAI-ADK Agent Core Behaviors:

| Karpathy Principle        | MoAI-ADK Coverage                                         | GAP                                                                 |
|---------------------------|-----------------------------------------------------------|---------------------------------------------------------------------|
| Think Before Coding       | 95% -- Surface Assumptions + Manage Confusion + Discovery | Fully covered                                                       |
| Simplicity First          | 90% -- Enforce Simplicity + Scope Discipline              | **Missing: quantitative LOC trigger ("200->50 rewrite")**           |
| Surgical Changes          | 85% -- Scope Discipline + Multi-File Decomposition        | **Missing: "match existing style" positive directive**              |
| Goal-Driven Execution     | 80% -- Reproduction-First Bug Fix + TRUST 5               | **Missing: goal->test pattern for non-SPEC ad-hoc tasks**           |

The BIGGEST unique value is EXAMPLES.md anti-pattern catalog -- MoAI has abstract principles but NO concrete code-level wrong/right examples.

### 1.2 Non-Goals

- **NOTICE.md changes**: ~~Excluded~~ → User later requested attribution for Karpathy open-source material (session mid-flight requirement change)
- **New agent creation**: Integration uses existing agents via chaining
- **Behavioral rule removal**: Additive only, no breaking changes
- **Language-specific tooling**: 16-language neutrality maintained throughout
- **Upstream sync automation**: One-time integration, no automated sync with forrestchang repo

---

## 2. Scope

### 2.1 Deliverables

| # | Target Path                                   | Type   | REQ        | Description                                                    |
|---|-----------------------------------------------|--------|------------|----------------------------------------------------------------|
| 1 | `.claude/rules/moai/core/moai-constitution.md`| Amend  | KP-001/002/003 | Constitution amendments: 3 additions to Agent Core Behaviors  |
| 2 | `.claude/skills/moai/references/anti-patterns.md` | New  | KP-004     | Progressive-disclosure anti-pattern reference skill (8 categories) |
| 3 | `.claude/rules/moai/development/karpathy-quickref.md` | New | KP-005     | Quick reference rule mapping 4 principles to 6 behaviors      |
| 4 | `internal/template/templates/` (mirrors of 2+3) | Sync | KP-006     | Template-First propagation of all new files                   |

### 2.2 Constitution Amendment Details

Amendments target the `<!-- moai:evolvable-start id="agent-core-behaviors" -->` section:

**Amendment A (REQ-KP-001)**: Extend Behavior 4 (Enforce Simplicity) with quantitative trigger:
- Add LOC ratio check: if implementation exceeds 3x estimated minimum viable LOC, flag for rewrite
- Max 6 lines added

**Amendment B (REQ-KP-002)**: Extend Behavior 5 (Maintain Scope Discipline) with style-matching directive:
- Add "match existing code style, naming, conventions" as positive directive
- Max 4 lines added

**Amendment C (REQ-KP-003)**: Extend Behavior 6 (Verify, Don't Assume) with goal-to-test pattern:
- Add goal->test transformation pattern for non-SPEC ad-hoc tasks
- Max 6 lines added

Total constitution addition: max 16 lines (within +20 line constraint).

### 2.3 Anti-Pattern Reference Skill Details

8 anti-pattern categories with code examples:

| # | Category                   | Mapped Behavior(s)          | Trigger Keywords                         |
|---|----------------------------|-----------------------------|------------------------------------------|
| 1 | Hidden Assumptions         | Behavior 1 (Surface)        | assume, implicit, silently               |
| 2 | Multiple Interpretations   | Behavior 2 (Confusion)      | ambiguous, unclear, either/or            |
| 3 | Over-Abstraction           | Behavior 4 (Simplicity)     | factory, interface, abstract, generic    |
| 4 | Speculative Features       | Behavior 5 (Scope)          | might need, future-proof, just in case   |
| 5 | Drive-by Refactoring       | Behavior 5 (Scope)          | while I was here, clean up, improve      |
| 6 | Style Drift                | Behavior 5 (Scope)          | naming, convention, style, format        |
| 7 | Vague Goals                | Behavior 6 (Verify)         | should work, basically, roughly          |
| 8 | Test-First Execution       | Behavior 6 + Rule 4         | skip test, test later, obvious           |

Skill configuration:
- Primary examples in Go, secondary in Python/TypeScript (16-language neutrality)
- Progressive disclosure: Level 1 (~100 tokens metadata), Level 2 (~5000 tokens full examples)
- Triggers: keywords + agent types + phases
- user-invocable: false (background knowledge, loaded by trigger matching)

### 2.4 Quick Reference Rule Details

Rule file mapping Karpathy's 4 principles to MoAI's 6 Agent Core Behaviors with concrete checkpoint questions per principle.

- Paths frontmatter: `**/*.go,**/*.py,**/*.ts,**/*.js,**/*.java,**/*.rs,**/*.c,**/*.cpp,**/*.rb,**/*.php,**/*.kt,**/*.swift,**/*.dart,**/*.ex,**/*.scala,**/*.hs,**/*.zig` (all 16 supported languages)
- Content: principle name, mapped behavior(s), 3-5 checkpoint questions, cross-reference to anti-pattern skill

### 2.5 Template-First Synchronization

All new files must exist in both locations:
- Local: `.claude/rules/moai/` and `.claude/skills/moai/`
- Template mirror: `internal/template/templates/.claude/rules/moai/` and `internal/template/templates/.claude/skills/moai/`

---

## 3. EARS Requirements

### REQ-KP-001: Constitution Amendment -- Quantitative Simplicity Trigger (Event-Driven)

**When** an agent produces implementation code, **the system SHALL** enforce a quantitative simplicity check: if implementation exceeds 3x the estimated minimum viable LOC, the agent MUST flag it for rewrite consideration and document justification.

**Target**: `.claude/rules/moai/core/moai-constitution.md`, Agent Core Behavior 4 (Enforce Simplicity)

### REQ-KP-002: Constitution Amendment -- Match Existing Style Directive (State-Driven)

**While** modifying existing code, **the agent SHALL** match the existing code style, naming patterns, and conventions, even if personal preference differs. Consistency takes precedence over personal style within the scope of the task.

**Target**: `.claude/rules/moai/core/moai-constitution.md`, Agent Core Behavior 5 (Maintain Scope Discipline)

### REQ-KP-003: Constitution Amendment -- Goal-to-Test Pattern for Ad-Hoc Tasks (Optional)

**Where** non-SPEC ad-hoc tasks are being executed (bug fixes, small features, one-off changes), **the agent SHALL** transform the stated goal into a verifiable test case before implementing.

**Target**: `.claude/rules/moai/core/moai-constitution.md`, Agent Core Behavior 6 (Verify, Don't Assume)

### REQ-KP-004: Anti-Pattern Reference Skill Creation (Ubiquitous)

**The system SHALL** provide a progressive-disclosure skill `moai-reference-anti-patterns` containing 8 anti-pattern categories with concrete wrong/right code examples, each mapped to the corresponding MoAI Agent Core Behavior.

Requirements:
- Primary examples in Go, secondary in Python/TypeScript (16-language neutrality)
- Progressive disclosure: Level 1 (~100 tokens), Level 2 (~5000 tokens)
- Trigger agents: expert-backend, expert-frontend, manager-quality, evaluator-active
- Trigger phases: run, review
- user-invocable: false

**Target**: `.claude/skills/moai/references/anti-patterns.md`

### REQ-KP-005: Quick Reference Card (Ubiquitous)

**The system SHALL** provide a quick reference rule file mapping Karpathy's 4 principles to MoAI's 6 Agent Core Behaviors with concrete checkpoint questions for each principle.

**Target**: `.claude/rules/moai/development/karpathy-quickref.md`

### REQ-KP-006: Template Propagation (Ubiquitous)

**All** new files **SHALL** be created in `internal/template/templates/` first, then propagated to local copies. `make build` SHALL succeed and `go test ./internal/template/...` SHALL pass after propagation.

### REQ-KP-007: Workflow Chaining Integration (Ubiquitous)

**The system SHALL** enforce Karpathy principles through existing agent/skill chaining:
- manager-spec: Anti-pattern awareness during SPEC creation (via karpathy-quickref.md paths trigger)
- expert-backend/frontend: Constitution amendments apply automatically (always loaded)
- manager-quality: Anti-pattern catalog validation during review (via skill trigger)
- evaluator-active: Score implementations against anti-pattern examples (via skill trigger)

This is achieved through constitution amendments (always loaded) + skill trigger configuration (on-demand loading). No new agent invocation patterns are required.

---

## 4. Exclusions (What NOT to Build)

- **NOTICE.md or attribution files**: User explicitly excluded copyright notice changes
- **New agent definitions**: Integration uses existing agent chaining architecture
- **Breaking changes to existing rules**: Additive only, no modification of existing Behavior text
- **Language-specific bias**: Templates must maintain 16-language neutrality
- **Upstream sync with forrestchang repo**: One-time integration, no automated sync
- **IDE plugin or external tooling**: MoAI-ADK internal orchestration only
- **New SPEC workflow phases**: Uses existing plan->run->sync pipeline
- **Custom MX tag types**: Uses existing @MX tag vocabulary

---

## 5. Constraints

- [HARD] NO NOTICE.md changes (user explicit request)
- [HARD] Template-First: all file changes to `internal/template/templates/` before local copies
- [HARD] 16-language neutrality: no language favored in templates
- [HARD] Additive only: no breaking changes to existing rules
- [HARD] Token budget: Constitution additions max +20 lines total
- [HARD] Skill progressive disclosure: Level 1 under 150 tokens, Level 2 under 6000 tokens
- [HARD] Constitution amendments target ONLY the evolvable zone (`moai:evolvable-start id="agent-core-behaviors"`)
- Frozen zone (Zone Registry CONST-V3R2-025..046) remains untouched

---

## 6. Acceptance Hooks

- AC-001 ~ AC-006 -> see `acceptance.md`
- Implementation order: see `plan.md`

---

## 7. References

- forrestchang/andrej-karpathy-skills -- https://github.com/forrestchang/andrej-karpathy-skills (94K stars)
- Karpathy 4 principles: Think Before Coding, Simplicity First, Surgical Changes, Goal-Driven Execution
- `.claude/rules/moai/core/moai-constitution.md` -- Constitution target (Evolvable zone: Agent Core Behaviors)
- `.claude/rules/moai/core/zone-registry.md` -- Zone registry (CONST-V3R2-025..046 Frozen zone confirmation)
- SPEC-V3R3-PATTERNS-001 -- Pattern Cookbook (precedent for reference file integration)
- SPEC-V3R2-CON-001 -- Zone Registry (frozen zone definitions, ensures no frozen zone violation)
