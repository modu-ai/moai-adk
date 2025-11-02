---
name: moai-alfred-context-budget
description: "Claude Code context window optimization strategies, JIT retrieval, progressive loading, memory file patterns, and cleanup practices. Use when optimizing context usage, managing large projects, or implementing efficient workflows."
allowed-tools: "Read, Grep"
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

Learn more in `reference.md` for detailed JIT strategy, memory patterns, context budget calculations, and management practices.

**Related Skills**: moai-alfred-practices, moai-cc-memory
