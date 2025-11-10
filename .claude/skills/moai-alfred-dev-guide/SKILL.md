---
name: "Orchestrating SPEC-First TDD Development"
version: 2.0.0
created: 2025-11-02
updated: 2025-11-11
status: active
description: "Guides agents through SPEC-First TDD workflow with context engineering, TRUST principles, and @TAG traceability. Essential for /alfred:1-plan, /alfred:2-run, /alfred:3-sync commands. Covers EARS requirements, JIT context loading, TDD RED-GREEN-REFACTOR cycle, and TAG chain validation."
keywords: ['tdd-workflow', 'spec-first', 'development-guide', 'context-engineering', 'trust-principles', 'tag-traceability', 'ears-format', 'research', 'workflow-optimization', 'test-driven-development']
allowed-tools: "Read, Bash(rg:*), Bash(grep:*), AskUserQuestion, TodoWrite"
---

# Alfred Development Guide Skill

## Core Workflow: SPEC → TEST → CODE → DOC

**No spec, no code. No tests, no implementation.**

### 1. SPEC Phase (`/alfred:1-plan`)
- Author detailed specifications first with `@SPEC:ID` tags
- Use EARS format (5 patterns: Ubiquitous, Event-driven, State-driven, Optional, Constraints)
- Store in `.moai/specs/SPEC-{ID}/spec.md`

### 2. TDD Phase (`/alfred:2-run`)
- **RED**: Write failing tests with `@TEST:ID` tags
- **GREEN**: Implement minimal code with `@CODE:ID` tags
- **REFACTOR**: Improve code while maintaining SPEC compliance
- Document with `@DOC:ID` tags

### 3. SYNC Phase (`/alfred:3-sync`)
- Verify @TAG chain integrity (SPEC→TEST→CODE→DOC)
- Synchronize documentation with implementation
- Generate sync report

## Key Principles

**Context Engineering**: Load only necessary documents at each phase
- `/alfred:1-plan` → product.md, structure.md, tech.md
- `/alfred:2-run` → SPEC-{ID}/spec.md, development-guide.md
- `/alfred:3-sync` → sync-report.md, TAG validation

**TRUST 5 Pillars**:
1. **T** – Test-driven (RED→GREEN→REFACTOR)
2. **R** – Readable (clear naming, documentation)
3. **U** – Unified (consistent patterns, language)
4. **S** – Secured (OWASP compliance, security reviews)
5. **E** – Evaluated (metrics, coverage ≥85%)

## Common Patterns

| Scenario | Action |
|----------|--------|
| Write tests first | RED phase: failing tests with @TEST:ID |
| Implement feature | GREEN phase: minimal code with @CODE:ID |
| Refactor safely | REFACTOR phase: improve code structure |
| Track changes | Always use @TAG system in code + docs |
| Validate completion | `/alfred:3-sync` verifies all links |

---

## Research Integration

### TDD Workflow Research Capabilities

**TDD Pattern Analysis**:
- **RED phase optimization**: Failing test generation strategies and coverage targets
- **GREEN phase efficiency**: Minimal implementation patterns and code quality metrics
- **REFACTOR phase research**: Code improvement techniques and safety net validation
- **Cycle time measurement**: Time tracking for each phase across different project types

**Workflow Research Areas**:
- **Context engineering research**: Optimal JIT loading strategies for different development phases
- **TRUST principle validation**: Quality metrics and compliance scoring methods
- **@TAG traceability studies**: ID assignment patterns and chain integrity optimization
- **SPEC-TDD correlation**: SPEC quality metrics and TDD success rate relationships

**Research Methodology**:
- **Development pattern analysis**: Track success metrics across different project types
- **Context load optimization**: Measure effectiveness of progressive disclosure in TDD
- **TAG system performance**: Evaluate different ID assignment strategies and chain integrity
- **Workflow efficiency scoring**: Compare different approaches to SPEC→TEST→CODE→DOC workflows

### SPEC Integration Research Framework

#### 1. EARS Format Research
- **EARS pattern effectiveness**: Study on requirement clarity vs. implementation accuracy
- **EARS optimization**: Best practices for each pattern type (Ubiquitous, Event-driven, etc.)
- **EARS validation metrics**: Quality scoring for SPEC completeness and testability
- **EARS evolution**: Pattern adaptation for different project sizes and complexity

#### 2. TAG System Research
- **TAG assignment strategies**: Optimal ID patterns for different project domains
- **TAG chain integrity**: Validation methods and impact tracking on traceability
- **TAG performance metrics**: Impact on code quality, test coverage, and maintainability
- **TAG documentation standards**: Best practices for HISTORY section maintenance

#### 3. TDD Cycle Optimization
```
TDD Research Framework:
├── Phase Analysis
│   ├── RED phase failure patterns
│   ├── GREEN phase efficiency metrics
│   ├── REFACTOR phase improvement strategies
│   └── Cycle time optimization techniques
├── Quality Research
│   ├── TRUST 5 validation methods
│   ├── Code quality measurement
│   ├── Test coverage correlation
│   └── Maintainability metrics
└── Context Engineering
        ├── JIT loading strategies
        ├── Progressive disclosure optimization
        ├── Context compression techniques
        └── Memory file organization
```

**Current Research Focus Areas**:
- TDD workflow efficiency across different project sizes
- SPEC quality impact on development success rates
- @TAG system optimization for traceability
- Context engineering for multi-session workflows
- TRUST principle validation and improvement

---

## Integration with Research System

The development guide integrates with MoAI-ADK's research framework by:

1. **Collecting usage patterns**: Track how teams use TDD workflows and identify optimization opportunities
2. **Validating research hypotheses**: Provide real-world testing ground for research initiatives
3. **Documenting best practices**: Capture lessons learned and improvement opportunities
4. **Benchmarking workflows**: Measure effectiveness of different approaches across project types

**Research Collaboration**:
- **Context budget team**: Share insights on memory optimization during TDD
- **TAG research team**: Provide data on traceability patterns and chain integrity
- **Workflow optimization team**: Contribute to automation and efficiency research
