---
name: moai-core-context-budget
description: Enterprise Claude Code context window optimization with 2025 best practices
modularized: true
modules:
  - budget-allocation-clearing
  - memory-mcp-optimization
  - quality-chunking-patterns
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: context, moai, core, budget  


## Quick Reference (30 seconds)

# Token Optimization & Context Window Budget Management

**200K token context optimization** with aggressive clearing strategies, memory management, and quality-over-quantity principles for Claude Code.

**Key Principles (2025)**:
1. **Avoid Last 20%** - Performance degrades in final fifth
2. **Aggressive Clearing** - `/clear` every 1-3 messages
3. **Lean Memory Files** - Keep each file < 500 lines
4. **Disable Unused MCPs** - Minimize tool definitions
5. **Quality > Quantity** - 10% relevant beats 90% noise

---

## Modules

### 1. [Budget Allocation & Clearing](modules/budget-allocation-clearing.md)
Context budget breakdown, allocation strategies, and aggressive clearing patterns.

**Topics**:
- 200K token allocation breakdown
- When to execute `/clear`
- Context monitoring and optimization
- Anti-patterns and pitfalls

### 2. [Memory & MCP Optimization](modules/memory-mcp-optimization.md)
Memory file management, rotation strategies, and MCP server configuration.

**Topics**:
- Memory file structure (<500 lines)
- File rotation automation
- MCP server context impact
- Phase-based MCP strategy

### 3. [Quality & Chunking Patterns](modules/quality-chunking-patterns.md)
Quality-focused context loading and strategic task chunking for large projects.

**Topics**:
- Quality checklist for context
- Task size estimation
- Chunk dependency management
- Common pitfalls to avoid

---

## Best Practices Checklist

- [ ] Context maintained below 80%
- [ ] `/clear` executed every 1-3 messages
- [ ] Memory files < 500 lines each
- [ ] Unused MCP servers disabled
- [ ] Tasks chunked < 120K tokens
- [ ] Only relevant files loaded

---

## Version History

**v4.0.0** (2025-11-21)
- Modularized structure (3 focused modules)
- Enhanced 2025 best practices
- Aggressive clearing patterns

---

**Last Updated**: 2025-11-21  
**Optimization**: Modularized for progressive loading
