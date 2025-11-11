---
name: moai-alfred-practices
version: 2.0.0
created: 2025-11-02
updated: 2025-11-11
status: active
description: "Practical workflows, context engineering strategies, and real-world execution examples for MoAI-ADK. Use when learning workflow patterns, optimizing context management, debugging issues, or implementing features end-to-end."
keywords: ['workflow-examples', 'context-engineering', 'jit-retrieval', 'agent-usage', 'debugging-patterns', 'feature-implementation', 'research', 'workflow-optimization', 'execution-patterns', 'best-practices']
allowed-tools: "Read, Glob, Grep, Bash, AskUserQuestion, TodoWrite"
---

## Skill Metadata

| Field | Value |
| ----- | ----- |
| Version | 1.0.0 |
| Tier | Alfred |
| Auto-load | When practical guidance is needed |
| Keywords | workflow-examples, context-engineering, jit-retrieval, agent-usage, debugging-patterns, feature-implementation |

## What It Does

JIT (Just-in-Time) context 관리 전략, Explore agent 효율적 사용법, SPEC → TDD → Sync 실행 순서, 자주 발생하는 문제 해결책을 제공합니다.

## When to Use

- ✅ 실제 task 실행 시 구체적 단계 필요
- ✅ Context 관리 최적화 필요 (큰 프로젝트)
- ✅ Explore agent 효율적 활용 방법 학습
- ✅ SPEC → TDD → Sync 실행 패턴 학습
- ✅ 자주 발생하는 문제 해결 방법 찾기

## Core Practices at a Glance

### 1. Context Engineering Strategy

#### JIT (Just-in-Time) Retrieval
- 필요한 context만 즉시 pull
- Explore로 manual file hunting 대체
- Task thread에서 결과 cache하여 재사용

#### Efficient Use of Explore
- Call graphs/dependency maps for core module changes
- Similar features 검색으로 구현 참고
- SPEC references나 TAG metadata로 변경사항 anchor

### 2. Context Layering

```
High-level brief      → Purpose, stakeholders, success criteria
        ↓
Technical core        → Entry points, domain models, utilities
        ↓
Edge cases           → Known bugs, constraints, SLAs
```

### 3. Practical Workflow Commands

```bash
/alfred:1-plan "Feature name"
  → Skill("moai-alfred-spec-metadata-extended") validation
  → SPEC 생성

/alfred:2-run SPEC-ID
  → TDD RED → GREEN → REFACTOR
  → Tests + Implementation

/alfred:3-sync
  → Documentation auto-update
  → TAG chain validation
  → PR ready
```

## 5 Practical Scenarios

1. **Feature Implementation**: New feature from SPEC to production
2. **Debugging & Triage**: Error analysis with fix-forward recommendations
3. **TAG System Management**: ID assignment, HISTORY updates
4. **Backup Management**: Automatic safety snapshots before risky actions
5. **Multi-Agent Collaboration**: Coordinate between debug-helper, spec-builder, tdd-implementer

## Key Principles

- ✅ **Context minimization**: Load only what's needed now
- ✅ **Explore-first**: Use Explore agent for large searches
- ✅ **Living documentation**: Sync after significant changes
- ✅ **Problem diagnosis**: Use debug-helper for error triage
- ✅ **Reproducibility**: Record rationale for SPEC deviations

---

## Research Integration

### Workflow Optimization Research Capabilities

**Execution Pattern Analysis**:
- **Workflow efficiency studies**: Measure time and resource consumption across different execution patterns
- **Context engineering research**: Optimize JIT retrieval strategies for different project types and sizes
- **Agent usage optimization**: Research on optimal task delegation and collaboration patterns
- **Debug pattern analysis**: Identify common error patterns and effective resolution strategies

**Optimization Research Areas**:
- **Parallel execution opportunities**: Research on identifying and implementing task parallelization
- **Shortcut pattern development**: Discover efficient alternatives for common operations
- **Automation potential analysis**: Identify manual processes that can be automated
- **Resource allocation optimization**: Study optimal distribution of computational and memory resources

**Research Methodology**:
- **Performance benchmarking**: Compare different workflow approaches for efficiency and quality
- **User behavior analysis**: Study how teams adopt different patterns and identify improvement opportunities
- **Success rate tracking**: Measure the effectiveness of different execution patterns across project types
- **Resource consumption analysis**: Monitor memory, time, and computational usage patterns

### Practical Workflow Research Framework

#### 1. Context Engineering Research
- **JIT retrieval optimization**: Research on optimal timing and content for context loading
- **Progressive disclosure studies**: Effectiveness analysis for different project sizes and complexity
- **Memory file organization**: Research on optimal structures for different team sizes
- **Context compression techniques**: Analysis of summarization vs. caching strategies

#### 2. Agent Collaboration Research
- **Agent selection patterns**: Research on optimal agent usage for different task types
- **Collaboration efficiency**: Study on task delegation and coordination strategies
- **Communication overhead**: Analysis of inter-agent communication and optimization
- **Specialization research**: Research on optimal agent specialization and workload distribution

#### 3. Workflow Pattern Optimization
```
Workflow Research Framework:
├── Performance Analysis
│   ├── Execution time measurement
│   ├── Resource consumption tracking
│   ├── Success rate analysis
│   └── Quality impact correlation
├── Pattern Recognition
│   ├── Common workflow identification
│   ├── Bottleneck detection
│   ├── Opportunity discovery
│   └── Best practice extraction
└── Optimization Strategies
        ├── Parallelization research
        ├── Automation potential
        ├── Resource allocation
        ├── Performance tuning
```

**Current Research Focus Areas**:
- Workflow efficiency optimization for different project sizes
- Context engineering improvements for large-scale projects
- Agent collaboration pattern standardization
- Automation opportunity identification
- Execution performance benchmarking

---

## Integration with Research System

The workflow practices system integrates with MoAI-ADK's research framework by:

1. **Collecting execution data**: Track how teams use different workflows and identify optimization opportunities
2. **Validating workflow hypotheses**: Provide real-world testing ground for new workflow patterns and optimizations
3. **Documenting best practices**: Capture successful patterns and share them across the organization
4. **Benchmarking performance**: Establish baselines and measure improvement over time

**Research Collaboration**:
- **Context budget team**: Share insights on memory optimization during workflow execution
- **Agent optimization team**: Provide data on agent usage patterns and collaboration efficiency
- **Quality assurance team**: Collaborate on quality metrics and validation methods
- **Behavioral research team**: Study workflow adoption patterns and user behavior changes

---

**Learn More**: See `reference.md` for step-by-step examples, full workflow sequences, and advanced patterns.

**Related Skills**: moai-alfred-rules, moai-alfred-agent-guide, moai-essentials-debug
