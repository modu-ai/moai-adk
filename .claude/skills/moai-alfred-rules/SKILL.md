---
name: moai-alfred-rules
version: 2.0.0
created: 2025-11-02
updated: 2025-11-11
status: active
description: "Mandatory rules for Skill invocation, AskUserQuestion usage, TRUST principles, TAG validation, and TDD workflow. Use when validating workflow compliance, checking quality gates, enforcing MoAI-ADK standards, or verifying rule adherence."
keywords: ['skill-invocation', 'ask-user-question', 'trust', 'tag', 'tdd', 'quality-gates', 'workflow-compliance', 'research', 'enforcement-patterns', 'rule-validation']
allowed-tools: "Read, Glob, Grep, Bash, AskUserQuestion, TodoWrite"
---

## Skill Metadata

| Field | Value |
| ----- | ----- |
| Version | 1.0.0 |
| Tier | Alfred |
| Auto-load | When validating rules or quality gates |
| Keywords | skill-invocation, ask-user-question, trust, tag, tdd, quality-gates, workflow-compliance |

## What It Does

MoAI-ADK의 10가지 필수 Skill 호출 규칙, 5가지 AskUserQuestion 시나리오, TRUST 5 품질 게이트, TAG 체인 규칙, TDD 워크플로우를 정의합니다.

## When to Use

- ✅ Skill() 호출이 mandatory인지 optional인지 판단 필요
- ✅ 사용자 질문이 ambiguous할 때 AskUserQuestion 사용 여부 결정
- ✅ 코드/커밋이 TRUST 5를 준수하는지 확인
- ✅ TAG 체인 무결성 검증
- ✅ 커밋 메시지 형식 확인
- ✅ 품질 게이트(quality gate) 검증

## Core Rules at a Glance

### AGENT-FIRST PRINCIPLE (v5.0.0)

**RULE #1: ALWAYS delegate to agents - NEVER execute directly**

Commands MUST orchestrate, agents MUST execute:
```bash
# ❌ WRONG: Command doing domain work
"Design REST API for user management"

# ✅ CORRECT: Delegate to domain expert
Task(
  subagent_type="backend-expert",
  description="Design REST API for user management",
  prompt="You are the backend-expert agent. Design comprehensive user management API."
)
```

### Architecture Enforcement Rules

1. **Commands**: Orchestration ONLY - Never implement features directly
2. **Agents**: Domain expertise ownership - Handle complex reasoning
3. **Skills**: Knowledge capsules - Called by agents with context

### 10 Mandatory Skill Invocations

| User Request | Skill | Pattern |
|---|---|---|
| TRUST validation, quality check | `moai-foundation-trust` | `Skill("moai-foundation-trust")` |
| TAG validation, orphan detection | `moai-foundation-tags` | `Skill("moai-foundation-tags")` |
| SPEC authoring, spec validation | `moai-foundation-specs` | `Skill("moai-foundation-specs")` |
| EARS syntax, requirement formatting | `moai-foundation-ears` | `Skill("moai-foundation-ears")` |
| Git workflow, branch management | `moai-foundation-git` | `Skill("moai-foundation-git")` |
| Language detection, stack detection | `moai-foundation-langs` | `Skill("moai-foundation-langs")` |
| Debugging, error analysis | `moai-essentials-debug` | `Skill("moai-essentials-debug")` |
| Refactoring, code improvement | `moai-essentials-refactor` | `Skill("moai-essentials-refactor")` |
| Performance optimization | `moai-essentials-perf` | `Skill("moai-essentials-perf")` |
| Code review, quality review | `moai-essentials-review` | `Skill("moai-essentials-review")` |

### 5 AskUserQuestion Scenarios

Use `AskUserQuestion` when:
1. Tech stack choice unclear (multiple frameworks/languages)
2. Architecture decision needed (monolith vs microservices)
3. User intent ambiguous (multiple valid interpretations)
4. Existing component impacts unknown (breaking changes)
5. Resource constraints unclear (budget, timeline)

### TRUST 5 Quality Gates

- **Test**: 85%+ coverage required
- **Readable**: No code smells, SOLID principles
- **Unified**: Consistent patterns, no duplicate logic
- **Secured**: OWASP Top 10 checks, no secrets
- **Trackable**: @TAG chain intact (SPEC→TEST→CODE→DOC)

### TAG Chain Integrity Rules

- Assign as `<DOMAIN>-<###>` (e.g., `AUTH-003`)
- Create `@TEST` before `@CODE`
- Document in HISTORY section
- Never have orphan TAGs (TAG without corresponding code)

## Progressive Disclosure

---

## Research Integration

### Enforcement Pattern Research Capabilities

**Rule Effectiveness Analysis**:
- **Skill invocation studies**: Analyze which Skills are most/least effective in different scenarios
- **AskUserQuestion optimization**: Research optimal question formats, timing, and user experience
- **Trust principle validation**: Measure impact of TRUST 5 compliance on code quality and maintainability
- **TAG system effectiveness**: Study the correlation between TAG usage and project success metrics

**Enforcement Research Areas**:
- **Quality gate optimization**: Research on optimal thresholds and validation methods
- **Workflow compliance patterns**: Identify common violations and preventive measures
- **Rule adaptation strategies**: How to tailor enforcement rules for different team sizes and expertise levels
- **Performance impact studies**: Measure the overhead of rule enforcement vs. quality benefits

**Research Methodology**:
- **Compliance rate tracking**: Monitor success/failure rates of different rules across projects
- **Quality impact measurement**: Correlate rule enforcement with bug reduction, code quality, and maintainability
- **User experience research**: Study the friction points of rule enforcement and optimization opportunities
- **Rule effectiveness scoring**: Evaluate which rules provide the best quality/effort ratio

### Rule Validation Research Framework

#### 1. Trust Principle Research
- **Trust pillar effectiveness**: Study on individual TRUST principles vs. combined impact
- **Trust validation methods**: Research on best practices for quality measurement and scoring
- **Trust correlation studies**: Analyze relationship between TRUST compliance and project success
- **Trust threshold optimization**: Research optimal minimum requirements for different project types

#### 2. TAG System Research
- **TAG usage patterns**: Research on optimal TAG assignment and naming conventions
- **TAG chain integrity**: Study the effectiveness of TAG traceability on code maintainability
- **TAG performance metrics**: Measure impact on development speed, quality, and collaboration
- **TAG evolution strategies**: Research how TAG systems adapt to growing projects

#### 3. Quality Gate Optimization
```
Quality Gate Research Framework:
├── Threshold Analysis
│   ├── Optimal coverage levels for different project types
│   ├── Code smell detection improvement
│   ├── Security validation effectiveness
│   └── Performance impact measurement
├── Compliance Patterns
│   ├── Common violation identification
│   ├── Prevention strategy research
│   ├── Team adaptation studies
│   └── Skill integration analysis
└── Quality Metrics
        ├── Bug reduction correlation
        ├── Code quality improvement
        ├── Maintenance cost reduction
        └── Development efficiency impact
```

**Current Research Focus Areas**:
- Rule effectiveness optimization for different team sizes
- Quality gate balance between enforcement and productivity
- AskUserQuestion improvement strategies
- TAG system evolution and scalability
- Trust principle validation across domains

---

## Integration with Research System

The rules enforcement system integrates with MoAI-ADK's research framework by:

1. **Collecting violation data**: Track common rule violations and identify systemic issues
2. **Validating rule effectiveness**: Provide real-world testing ground for new rules and modifications
3. **Benchmarking quality standards**: Establish baselines and measure improvement over time
4. **Identifying optimization opportunities**: Find rules that are too strict or too permissive

**Research Collaboration**:
- **TAG research team**: Provide data on TAG system effectiveness and optimization opportunities
- **Quality assurance team**: Collaborate on quality metrics and validation methods
- **Context engineering team**: Share insights on rule impact on context usage
- **Behavioral research team**: Study rule adoption patterns and user behavior changes

---

**Version**: 2.0.0
**Related Skills**: moai-foundation-trust, moai-foundation-tags, moai-alfred-practices