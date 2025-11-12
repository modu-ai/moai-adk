---
name: moai-alfred-context-budget
version: 2.0.0
created: 2025-11-02
updated: 2025-11-11
status: active
description: "Claude Code context window optimization strategies, JIT retrieval, progressive loading, memory file patterns, and cleanup practices. Use when optimizing context usage, managing large projects, or implementing efficient workflows."
keywords: ['context-budget', 'memory-optimization', 'jit-retrieval', 'progressive-disclosure', 'token-management', 'research', 'memory-patterns', 'optimization']
allowed-tools: "Read, Grep, AskUserQuestion, TodoWrite"
---

## What It Does

Claude Code context budget (100K-200K tokens)의 효율적 관리 전략, progressive disclosure, JIT retrieval, memory file 구조화 방법을 제공합니다.

## When to Use

- ✅ Context limit 도달 위험성 있을 때
- ✅ 대규모 프로젝트의 context 관리 필요
- ✅ Session handoff 준비
- ✅ Memory file 구조 설계/정리

## Context Budget Overview

```
Total: 100K-200K tokens
├── System Prompt: ~2K
├── Tools & Instructions: ~5K
├── Session History: ~30K
├── Project Context: ~40K
└── Available for Response: ~23K
```

## Progressive Disclosure Pattern

1. **Metadata** (always): Skill name, description, triggers
2. **Content** (on-demand): Full SKILL.md when invoked
3. **Supporting** (JIT): Examples, templates when referenced

## JIT Retrieval Principles

- Pull only what you need for immediate step
- Prefer `Explore` over manual file hunting
- Cache critical insights for reuse
- Remove unnecessary files regularly

## Best Practices

✅ DO:
- Use Explore for large searches
- Cache results in memory files
- Keep memory files < 500 lines each
- Update session-summary.md before task switches
- Archive completed work

❌ DON'T:
- Load entire src/ directory upfront
- Duplicate context between files
- Store memory files outside `.moai/memory/`
- Leave stale session notes (archive or delete)
- Cache raw code (summarize instead)

---

## Research Integration

### Memory Optimization Research Capabilities

**Context Budget Pattern Analysis**:
- **Token allocation studies**: Optimal distribution between system prompts, tools, and project context
- **Progressive disclosure efficiency**: Best practices for metadata → content → supporting data retrieval
- **Memory file structure optimization**: Ideal size limits, naming conventions, and retention policies
- **JIT retrieval algorithm research**: Pattern matching for context-aware file selection

**Memory Management Research Areas**:
- **Context compression techniques**: Summarization vs caching for different file types
- **Session handoff optimization**: Memory preservation strategies for multi-session workflows
- **Memory cleanup automation**: Trigger conditions and retention policies for different project sizes
- **Memory file organization**: Hierarchical structures for large-scale projects vs small teams

**Research Methodology**:
- **Token usage tracking**: Monitor consumption patterns across project types
- **Context efficiency scoring**: Measure success rates of progressive disclosure
- **Memory file impact analysis**: Evaluate the effect of different memory structures on task completion
- **JIT retrieval performance**: Compare manual file hunting vs automated context loading

### Memory Pattern Research Categories

#### 1. Project Size-Based Patterns
- **Small projects (<50 files)**: Minimal memory, focus on current session
- **Medium projects (50-200 files)**: Structured memory with hierarchical organization
- **Large projects (>200 files)**: Advanced memory with automated cleanup and compression

#### 2. Workflow-Specific Memory Strategies
- **SPEC-first workflows**: Memory prioritization for SPEC files and test cases
- **Team mode projects**: Shared memory patterns and collaboration memory
- **Personal mode projects**: Individual memory optimization and preference learning

#### 3. Memory File Research Framework
```
Memory File Analysis Framework:
├── Structure Analysis
│   ├── File size optimization
│   ├── Content categorization
│   ├── Access frequency tracking
│   └── Retention policy optimization
├── Performance Research
│   ├── Load time measurements
│   ├── Context impact scoring
│   ├── Token savings calculation
│   └── Task completion correlation
└── Best Practices Development
        ├── Memory architecture patterns
        ├── Content compression techniques
        ├── Cleanup automation strategies
        └── Memory migration guides
```

**Current Research Focus Areas**:
- Context budget allocation optimization
- Memory file size vs. utility relationship studies
- Progressive disclosure effectiveness metrics
- Memory pattern documentation standardization

---

Learn more in `reference.md` for detailed JIT strategy, memory patterns, context budget calculations, and management practices.

**Related Skills**: moai-alfred-practices, moai-cc-memory, moai-foundation-tags
