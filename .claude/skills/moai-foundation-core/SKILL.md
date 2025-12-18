---
name: moai-foundation-core
description: MoAI-ADK's foundational principles - TRUST 5, SPEC-First TDD, delegation patterns, token optimization, progressive disclosure, modular architecture, agent catalog, command reference, and execution rules for building AI-powered development workflows
version: 2.3.0
modularized: true
updated: 2025-12-03
status: active
tags:
 - foundation
 - core
 - orchestration
 - agents
 - commands
 - trust-5
 - spec-first-tdd
allowed-tools: Read, Grep, Glob
---

# MoAI Foundation Core

Foundational principles and architectural patterns that power MoAI-ADK's AI-driven development workflow.

Core Philosophy: Quality-first, test-driven, modular, and efficient AI development through proven patterns and automated workflows.

## Quick Reference (30 seconds)

What is MoAI Foundation Core?
Six essential principles that ensure quality, efficiency, and scalability in AI-powered development:

1. TRUST 5 Framework - Quality gate system (Test-first, Readable, Unified, Secured, Trackable)
2. SPEC-First TDD - Specification-driven test-driven development workflow
3. Delegation Patterns - Task orchestration via specialized agents (never direct execution)
4. Token Optimization - 200K budget management and context efficiency
5. Progressive Disclosure - Three-tier knowledge delivery (Quick → Implementation → Advanced)
6. Modular System - File splitting and reference architecture for scalability

Quick Access:
- Quality standards → [TRUST 5 Module](modules/trust-5-framework.md)
- Development workflow → [SPEC-First TDD Module](modules/spec-first-tdd.md)
- Agent coordination → [Delegation Patterns Module](modules/delegation-patterns.md)
- Budget management → [Token Optimization Module](modules/token-optimization.md)
- Content structure → [Progressive Disclosure Module](modules/progressive-disclosure.md)
- File organization → [Modular System Module](modules/modular-system.md)
- Agent catalog → [Agents Reference Module](modules/agents-reference.md) NEW
- Command reference → [Commands Reference Module](modules/commands-reference.md) NEW
- Security & constraints → [Execution Rules Module](modules/execution-rules.md) NEW

Use Cases:
- New agent creation with quality standards
- New skill development with structural guidelines
- Complex workflow orchestration
- Token budget planning and optimization
- Documentation architecture design
- Quality gate configuration

---

## Implementation Guide (5 minutes)

### 1. TRUST 5 Framework - Quality Assurance System

Purpose: Automated quality gates ensuring code quality, security, and maintainability.

Five Pillars:

Test-first Pillar:
- Requirement: Maintain test coverage at or above 85 percent
- Validation: Execute pytest with coverage reporting
- Failure Action: Block merge and generate missing tests
- WHY: High coverage ensures code reliability and reduces production defects
- IMPACT: Catches bugs early, reduces debugging time by 60-70 percent

Readable Pillar:
- Requirement: Use clear and descriptive naming conventions
- Validation: Execute ruff linter checks
- Failure Action: Issue warning and suggest refactoring improvements
- WHY: Clear naming improves code comprehension and team collaboration
- IMPACT: Reduces onboarding time by 40 percent, improves maintenance velocity

Unified Pillar:
- Requirement: Apply consistent formatting and import patterns
- Validation: Execute black formatter and isort checks
- Failure Action: Auto-format code or issue warning
- WHY: Consistency eliminates style debates and merge conflicts
- IMPACT: Reduces code review time by 30 percent, improves readability

Secured Pillar:
- Requirement: Comply with OWASP security standards
- Validation: Execute security-expert agent analysis
- Failure Action: Block merge and require security review
- WHY: Security vulnerabilities create critical business and legal risks
- IMPACT: Prevents 95+ percent of common security vulnerabilities

Trackable Pillar:
- Requirement: Write clear and structured commit messages
- Validation: Match Git commit message regex patterns
- Failure Action: Suggest proper commit message format
- WHY: Clear history enables debugging, auditing, and collaboration
- IMPACT: Reduces issue investigation time by 50 percent

Integration Points:
- Pre-commit hooks → Automated validation
- CI/CD pipelines → Quality gate enforcement
- Agent workflows → core-quality validation
- Documentation → Quality metrics

Detailed Reference: [TRUST 5 Framework Module](modules/trust-5-framework.md)

---

### 2. SPEC-First TDD - Development Workflow

Purpose: Specification-driven development ensuring clear requirements before implementation.

Three-Phase Workflow:

```
Phase 1: SPEC (/moai:1-plan)
 workflow-spec → EARS format
 Output: .moai/specs/SPEC-XXX/spec.md
 Execute /clear (saves 45-50K tokens)

Phase 2: TDD (/moai:2-run)
 RED: Failing tests
 GREEN: Passing code
 REFACTOR: Optimize
 Validate: ≥85% coverage

Phase 3: Docs (/moai:3-sync)
 API documentation
 Architecture diagrams
 Project reports
```

EARS Format:
- Ubiquitous: System-wide (always active)
- Event-driven: Trigger-based (when X, do Y)
- State-driven: Conditional (while X, do Y)
- Unwanted: Prohibited (shall not do X)
- Optional: Nice-to-have (where possible, do X)

Token Budget: SPEC 30K | TDD 180K | Docs 40K | Total 250K

Key Practice: Execute /clear after Phase 1 to initialize context.

Detailed Reference: [SPEC-First TDD Module](modules/spec-first-tdd.md)

---

### 3. Delegation Patterns - Agent Orchestration

Purpose: Task delegation to specialized agents, avoiding direct execution.

Core Principle [HARD]: Alfred must delegate all work through Task() to specialized agents.

WHY: Direct execution bypasses specialization, quality gates, and token optimization.
IMPACT: Proper delegation improves task success rate by 40 percent and enables parallel execution.

Delegation Syntax:
```python
result = await Task(
 subagent_type="specialized_agent",
 prompt="Clear, specific task",
 context={"relevant": "data"}
)
```

Three Patterns:

Sequential (dependencies):
```python
design = Task(subagent_type="api-designer", prompt="Design API")
code = Task(subagent_type="backend-expert", prompt="Implement", context={"design": design})
```

Parallel (independent):
```python
results = await Promise.all([
 Task(subagent_type="backend-expert", prompt="Backend"),
 Task(subagent_type="frontend-expert", prompt="Frontend")
])
```

Conditional (analysis-based):
```python
analysis = Task(subagent_type="debug-helper", prompt="Analyze")
if analysis.type == "security":
 Task(subagent_type="security-expert", prompt="Fix")
```

Agent Selection:
- Simple (1 file): 1-2 agents sequential
- Medium (3-5 files): 2-3 agents sequential
- Complex (10+ files): 5+ agents mixed

Detailed Reference: [Delegation Patterns Module](modules/delegation-patterns.md)

---

### 4. Token Optimization - Budget Management

Purpose: Efficient 200K token budget through strategic context management.

Budget Allocation:

SPEC Phase:
- Token Budget: 30K tokens
- Strategy: Load requirements only, execute /clear after completion
- WHY: Specification phase requires minimal context for requirement analysis
- IMPACT: Saves 45-50K tokens for implementation phase

TDD Phase:
- Token Budget: 180K tokens
- Strategy: Selective file loading, load only implementation-relevant files
- WHY: Implementation requires deep context but not full codebase
- IMPACT: Enables 70 percent larger implementations within budget

Docs Phase:
- Token Budget: 40K tokens
- Strategy: Result caching and template reuse
- WHY: Documentation builds on completed work artifacts
- IMPACT: Reduces redundant file reads by 60 percent

Total Budget:
- Combined Budget: 250K tokens across all phases
- Strategy: Phase separation with context reset between phases
- WHY: Clean context boundaries prevent token bloat
- IMPACT: Enables 2-3x larger projects within same budget

Token Saving Strategies:

1. Phase Separation: /clear between phases
 - After /moai:1-plan (saves 45-50K)
 - When context > 150K
 - After 50+ messages

2. Selective Loading: Load only necessary files

3. Context Optimization: 20-30K tokens target

4. Model Selection: Sonnet (quality) | Haiku (speed/cost)

Monitoring: /context command, track budget, suggest /clear

Cost Savings: Haiku 70% cheaper → 60-70% total savings

Detailed Reference: [Token Optimization Module](modules/token-optimization.md)

---

### 5. Progressive Disclosure - Content Architecture

Purpose: Three-tier knowledge delivery balancing value with depth.

Three Levels:

Quick Reference Level:
- Time Investment: 30 seconds
- Content: Core principles and essential concepts
- Token Usage: Approximately 1,000 tokens
- WHY: Rapid value delivery for time-constrained users
- IMPACT: Users gain 80 percent understanding in 5 percent of time

Implementation Level:
- Time Investment: 5 minutes
- Content: Workflows, practical examples, integration patterns
- Token Usage: Approximately 3,000 tokens
- WHY: Bridges concept to execution with actionable guidance
- IMPACT: Enables immediate productive work without deep expertise

Advanced Level:
- Time Investment: 10+ minutes
- Content: Deep technical dives, edge cases, optimization techniques
- Token Usage: Approximately 5,000 tokens
- WHY: Provides mastery-level knowledge for complex scenarios
- IMPACT: Reduces escalations by 70 percent through comprehensive coverage

SKILL.md Structure (≤500 lines):
```markdown
## Quick Reference (30s)
## Implementation Guide (5min)
## Advanced Patterns (10+min)
## Works Well With
```

Module Architecture:
- SKILL.md: Entry point, cross-references
- modules/: Deep dives, unlimited
- examples.md: Working samples
- reference.md: External links

File Splitting (when >500 lines):
```
SKILL.md (500 lines)
 Quick (80-120)
 Implementation (180-250)
 Advanced (80-140)
 References (10-20)

Overflow → modules/[topic].md
```

Detailed Reference: [Progressive Disclosure Module](modules/progressive-disclosure.md)

---

### 6. Modular System - File Organization

Purpose: Scalable file structure enabling unlimited content.

Standard Structure:
```
.claude/skills/skill-name/
 SKILL.md # Core (≤500 lines)
 modules/ # Extended (unlimited)
 patterns.md
 examples.md # Working samples
 reference.md # External links
 scripts/ # Utilities (optional)
 templates/ # Templates (optional)
```

File Principles:

1. SKILL.md: ≤500 lines, progressive disclosure, cross-references
2. modules/: Topic-focused, no limits, self-contained
3. examples.md: Copy-paste ready, commented
4. reference.md: API docs, resources

Cross-Reference Syntax:
```markdown
Details: [Module](modules/patterns.md)
Examples: [Examples](examples.md#auth)
External: [Reference](reference.md#api)
```

Discovery Flow: SKILL.md → Topic → modules/[topic].md → Deep dive

Detailed Reference: [Modular System Module](modules/modular-system.md)

---

## Advanced Implementation (10+ minutes)

Advanced patterns including cross-module integration, quality validation, and error handling are available in the detailed module references.

Key Advanced Topics:
- Cross-Module Integration: Combining TRUST 5 + SPEC-First TDD
- Token-Optimized Delegation: Parallel execution with context reset
- Progressive Agent Workflows: Escalation patterns
- Quality Validation: Pre/Post execution validation
- Error Handling: Delegation failure recovery

Detailed Reference: [examples.md](examples.md) for working code samples

---

## Works Well With

Agents:
- agent-factory - Create agents with foundation principles
- skill-factory - Generate skills with modular architecture
- core-quality - Automated TRUST 5 validation
- workflow-spec - EARS format specification
- workflow-tdd - RED-GREEN-REFACTOR execution
- workflow-docs - Documentation with progressive disclosure

Skills:
- moai-cc-claude-md - CLAUDE.md with foundation patterns
- moai-cc-configuration - Config with TRUST 5
- moai-cc-memory - Token optimization
- moai-context7-integration - MCP integration

Tools:
- AskUserQuestion - Use directly for all user interaction and clarification needs

Commands:
- /moai:1-plan - SPEC-First Phase 1
- /moai:2-run - TDD Phase 2
- /moai:3-sync - Documentation Phase 3
- /moai:9-feedback - Continuous improvement
- /clear - Token management

Foundation Modules (Extended Documentation):
- [Agents Reference](modules/agents-reference.md) - 26-agent catalog with 7-tier hierarchy
- [Commands Reference](modules/commands-reference.md) - 6 core commands workflow
- [Execution Rules](modules/execution-rules.md) - Security, Git strategy, compliance

---

## Quick Decision Matrix

| Scenario | Primary Principle | Supporting Principles |
|----------|------------------|----------------------|
| New Agent | TRUST 5, Delegation | Token Opt, Modular |
| New Skill | Progressive, Modular | TRUST 5, Token Opt |
| Workflow | Delegation Patterns | SPEC-First, Token Opt |
| Quality | TRUST 5 Framework | SPEC-First TDD |
| Budget | Token Optimization | Progressive, Modular |
| Docs | Progressive, Modular | Token Optimization |

Module Deep Dives:
- [TRUST 5 Framework](modules/trust-5-framework.md)
- [SPEC-First TDD](modules/spec-first-tdd.md)
- [Delegation Patterns](modules/delegation-patterns.md)
- [Token Optimization](modules/token-optimization.md)
- [Progressive Disclosure](modules/progressive-disclosure.md)
- [Modular System](modules/modular-system.md)
- [Agents Reference](modules/agents-reference.md) NEW
- [Commands Reference](modules/commands-reference.md) NEW
- [Execution Rules](modules/execution-rules.md) NEW

Full Examples: [examples.md](examples.md)
External Resources: [reference.md](reference.md)

---

Version: 2.4.0
Last Updated: 2025-12-06
Status: Active (optimized to <500 lines)
