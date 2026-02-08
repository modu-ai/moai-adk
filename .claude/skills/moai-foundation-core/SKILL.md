---
name: moai-foundation-core
description: >
  Provides MoAI-ADK foundational principles including TRUST 5 quality framework,
  SPEC-First DDD methodology, delegation patterns, progressive disclosure,
  and agent catalog reference.
  Use when referencing TRUST 5 gates, SPEC workflow, EARS format, DDD methodology,
  agent delegation patterns, or MoAI orchestration rules.
  Do NOT use for context and token management (use moai-foundation-context instead)
  or strategic analysis (use moai-foundation-philosopher instead).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.5.0"
  category: "foundation"
  status: "active"
  updated: "2026-01-21"
  modularized: "true"
  tags: "foundation, core, orchestration, agents, commands, trust-5, spec-first-ddd"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords:
    - "trust-5"
    - "spec-first"
    - "ddd"
    - "delegation"
    - "agent"
    - "token"
    - "progressive disclosure"
    - "modular"
    - "workflow"
    - "orchestration"
    - "quality gate"
    - "spec"
    - "ears format"
  agents:
    - "manager-spec"
    - "manager-ddd"
    - "manager-strategy"
    - "manager-quality"
    - "builder-agent"
    - "builder-skill"
  phases:
    - "plan"
    - "run"
    - "sync"
---

# MoAI Foundation Core

Foundational principles and architectural patterns that power MoAI-ADK's AI-driven development workflow.

Core Philosophy: Quality-first, domain-driven, modular, and efficient AI development through proven patterns and automated workflows.

## Quick Reference

What is MoAI Foundation Core?

Six essential principles that ensure quality, efficiency, and scalability in AI-powered development:

1. TRUST 5 Framework - Quality gate system (Tested, Readable, Unified, Secured, Trackable)
2. SPEC-First DDD - Specification-driven domain-driven development workflow
3. Delegation Patterns - Task orchestration via specialized agents (never direct execution)
4. Token Optimization - 200K budget management and context efficiency
5. Progressive Disclosure - Three-tier knowledge delivery (Quick, Implementation, Advanced)
6. Modular System - File splitting and reference architecture for scalability

Quick Access:

- Quality standards in modules/trust-5-framework.md
- Development workflow in modules/spec-first-ddd.md
- Agent coordination in modules/delegation-patterns.md
- Budget management in modules/token-optimization.md
- Content structure in modules/progressive-disclosure.md
- File organization in modules/modular-system.md
- Agent catalog in modules/agents-reference.md
- Command reference in modules/commands-reference.md
- Security and constraints in modules/execution-rules.md

Use Cases:

- New agent creation with quality standards
- New skill development with structural guidelines
- Complex workflow orchestration
- Token budget planning and optimization
- Documentation architecture design
- Quality gate configuration

---

## Implementation Guide

### 1. TRUST 5 Framework - Quality Assurance System

Purpose: Automated quality gates ensuring code quality, security, and maintainability.

Five Pillars:

Tested Pillar: Maintain comprehensive test coverage with characterization tests ensuring behavior preservation. Execute pytest with coverage reporting. Block merge and generate missing tests on failure. Characterization tests capture current behavior for legacy code, while specification tests validate domain requirements for new code. High coverage ensures code reliability and reduces production defects. Preserves behavior during refactoring and reduces debugging time by 60-70 percent.

Readable Pillar: Use clear and descriptive naming conventions. Execute ruff linter checks. Issue warning and suggest refactoring improvements on failure. Clear naming improves code comprehension and team collaboration. Reduces onboarding time by 40 percent and improves maintenance velocity.

Unified Pillar: Apply consistent formatting and import patterns. Execute black formatter and isort checks. Auto-format code or issue warning on failure. Consistency eliminates style debates and merge conflicts. Reduces code review time by 30 percent and improves readability.

Secured Pillar: Comply with OWASP security standards. Execute security-expert agent analysis. Block merge and require security review on failure. Security vulnerabilities create critical business and legal risks. Prevents 95+ percent of common security vulnerabilities.

Trackable Pillar: Write clear and structured commit messages. Match Git commit message regex patterns. Suggest proper commit message format on failure. Clear history enables debugging, auditing, and collaboration. Reduces issue investigation time by 50 percent.

Integration Points: Pre-commit hooks for automated validation, CI/CD pipelines for quality gate enforcement, Agent workflows for core-quality validation, Documentation for quality metrics.

Detailed Reference: modules/trust-5-framework.md

---

### 2. SPEC-First DDD - Development Workflow

Purpose: Specification-driven development ensuring clear requirements before implementation.

Three-Phase Workflow:

Phase 1 SPEC (/moai:1-plan): workflow-spec generates EARS format. Output is .moai/specs/SPEC-XXX/spec.md. Execute /clear to save 45-50K tokens.

Phase 2 DDD (/moai:2-run): ANALYZE for requirements, PRESERVE for existing behavior, IMPROVE for enhancement. Validate with at least 85% coverage.

Phase 3 Docs (/moai:3-sync): API documentation, architecture diagrams, project reports.

EARS Format: Ubiquitous for system-wide always active requirements. Event-driven for trigger-based when X do Y requirements. State-driven for conditional while X do Y requirements. Unwanted for prohibited shall not do X requirements. Optional for nice-to-have where possible do X requirements.

Token Budget: SPEC takes 30K, DDD takes 180K, Docs takes 40K, Total is 250K.

Key Practice: Execute /clear after Phase 1 to initialize context.

Detailed Reference: modules/spec-first-ddd.md

---

### 3. Delegation Patterns - Agent Orchestration

Purpose: Task delegation to specialized agents, avoiding direct execution.

Core Principle: MoAI must delegate all work through Task() to specialized agents. Direct execution bypasses specialization, quality gates, and token optimization. Proper delegation improves task success rate by 40 percent and enables parallel execution.

Delegation Syntax: Call Task with subagent_type parameter for specialized agent, prompt parameter for clear specific task, and context parameter with relevant data dictionary.

Three Patterns:

Sequential for dependencies: Call Task to api-designer for design, then Task to backend-expert for implementation with design context.

Parallel for independent work: Call Promise.all with Task to backend-expert and Task to frontend-expert simultaneously.

Conditional for analysis-based: Call Task to debug-helper for analysis, then based on analysis.type, call Task to security-expert or other appropriate agent.

Agent Selection: Simple tasks with 1 file use 1-2 agents sequential. Medium tasks with 3-5 files use 2-3 agents sequential. Complex tasks with 10+ files use 5+ agents mixed.

Detailed Reference: modules/delegation-patterns.md

---

### 4. Token Optimization - Budget Management

Purpose: Efficient 200K token budget through strategic context management.

Budget Allocation:

SPEC Phase takes 30K tokens. Strategy is to load requirements only and execute /clear after completion. Specification phase requires minimal context for requirement analysis. Saves 45-50K tokens for implementation phase.

DDD Phase takes 180K tokens. Strategy is selective file loading, load only implementation-relevant files. Implementation requires deep context but not full codebase. Enables 70 percent larger implementations within budget.

Docs Phase takes 40K tokens. Strategy is result caching and template reuse. Documentation builds on completed work artifacts. Reduces redundant file reads by 60 percent.

Total Budget is 250K tokens across all phases. Phase separation with context reset between phases provides clean context boundaries and prevents token bloat. Enables 2-3x larger projects within same budget.

Token Saving Strategies:

Phase Separation: Execute /clear between phases, after /moai:1-plan to save 45-50K, when context exceeds 150K, after 50+ messages.

Selective Loading: Load only necessary files.

Context Optimization: Target 20-30K tokens.

Model Selection: Sonnet for quality, Haiku for speed and cost with 70% cheaper rates for 60-70% total savings.

Detailed Reference: modules/token-optimization.md

---

### 5. Progressive Disclosure - Content Architecture

Purpose: Three-tier knowledge delivery balancing value with depth.

Three Levels:

Quick Reference Level: 30 seconds time investment, core principles and essential concepts, approximately 1,000 tokens. Rapid value delivery for time-constrained users. Users gain 80 percent understanding in 5 percent of time.

Implementation Level: 5 minutes time investment, workflows, practical examples, integration patterns, approximately 3,000 tokens. Bridges concept to execution with actionable guidance. Enables immediate productive work without deep expertise.

Advanced Level: 10+ minutes time investment, deep technical dives, edge cases, optimization techniques, approximately 5,000 tokens. Provides mastery-level knowledge for complex scenarios. Reduces escalations by 70 percent through comprehensive coverage.

SKILL.md Structure (maximum 500 lines): Quick Reference section, Implementation Guide section, Advanced Patterns section, Works Well With section.

Module Architecture: SKILL.md as entry point with cross-references, modules directory for deep dives with unlimited size, examples.md for working samples, reference.md for external links.

File Splitting when exceeding 500 lines: SKILL.md contains Quick at 80-120 lines, Implementation at 180-250 lines, Advanced at 80-140 lines, References at 10-20 lines. Overflow content goes to modules/topic.md.

Detailed Reference: modules/progressive-disclosure.md

---

### 6. Modular System - File Organization

Purpose: Scalable file structure enabling unlimited content.

Standard Structure: Create .claude/skills/skill-name/ directory containing SKILL.md as core file under 500 lines, modules directory for extended content with unlimited size including patterns.md, examples.md for working samples, reference.md for external links, scripts directory for utilities (optional), templates directory (optional).

File Principles: SKILL.md stays under 500 lines with progressive disclosure and cross-references. modules directory is topic-focused with no limits and self-contained content. examples.md is copy-paste ready with comments. reference.md contains API docs and resources.

Cross-Reference Syntax: Reference modules as Details in modules/patterns.md, reference examples as Examples in examples.md#auth, reference external docs as External in reference.md#api.

Discovery Flow: SKILL.md to Topic to modules/topic.md to Deep dive.

Detailed Reference: modules/modular-system.md

---

## Advanced Implementation

Advanced patterns including cross-module integration, quality validation, and error handling are available in the detailed module references.

Key Advanced Topics:

- Cross-Module Integration: Combining TRUST 5 + SPEC-First DDD
- Token-Optimized Delegation: Parallel execution with context reset
- Progressive Agent Workflows: Escalation patterns
- Quality Validation: Pre/Post execution validation
- Error Handling: Delegation failure recovery

Detailed Reference: examples.md for working code samples

---

## Works Well With

Agents: agent-factory for creating agents with foundation principles, skill-factory for generating skills with modular architecture, core-quality for automated TRUST 5 validation, workflow-spec for EARS format specification, workflow-ddd for ANALYZE-PRESERVE-IMPROVE execution, workflow-docs for documentation with progressive disclosure.

Skills: moai-cc-claude-md for CLAUDE.md with foundation patterns, moai-cc-configuration for config with TRUST 5, moai-cc-memory for token optimization, moai-context7-integration for MCP integration.

Tools: AskUserQuestion for direct user interaction and clarification needs.

Commands: /moai:1-plan for SPEC-First Phase 1, /moai:2-run for DDD Phase 2, /moai:3-sync for Documentation Phase 3, /moai:9-feedback for continuous improvement, /clear for token management.

Foundation Modules (Extended Documentation): modules/agents-reference.md for 26-agent catalog with 7-tier hierarchy, modules/commands-reference.md for 6 core commands workflow, modules/execution-rules.md for security, Git strategy, and compliance.

---

## Quick Decision Guide

New Agent: Primary principle is TRUST 5 and Delegation. Supporting principles are Token Optimization and Modular.

New Skill: Primary principle is Progressive and Modular. Supporting principles are TRUST 5 and Token Optimization.

Workflow: Primary principle is Delegation Patterns. Supporting principles are SPEC-First and Token Optimization.

Quality: Primary principle is TRUST 5 Framework. Supporting principle is SPEC-First DDD.

Budget: Primary principle is Token Optimization. Supporting principles are Progressive and Modular.

Docs: Primary principle is Progressive and Modular. Supporting principle is Token Optimization.

Module Deep Dives: modules/trust-5-framework.md, modules/spec-first-ddd.md, modules/delegation-patterns.md, modules/token-optimization.md, modules/progressive-disclosure.md, modules/modular-system.md, modules/agents-reference.md, modules/commands-reference.md, modules/execution-rules.md.

Full Examples: examples.md
External Resources: reference.md
