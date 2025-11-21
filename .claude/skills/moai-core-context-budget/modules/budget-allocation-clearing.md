---
name: budget-allocation-clearing
parent: moai-core-context-budget
description: Context budget breakdown and clearing strategies
---

# Module 1: Budget Allocation & Clearing

## Context Budget Breakdown

```yaml
# Claude Code Context Budget (200K tokens)
Total Context Window: 200,000 tokens

Allocation:
  System Prompt: 2,000 tokens (1%)
  Tool Definitions: 5,000 tokens (2.5%)
  Session History: 30,000 tokens (15%)
  Project Context: 40,000 tokens (20%)
  Available for Response: 123,000 tokens (61.5%)
```

## Monitoring Context Usage

```bash
# Check current context usage
/context

# Example output:
# Context Usage: 156,234 / 200,000 tokens (78%)
# WARNING: Approaching 80% threshold
```

## When to Clear Context

**Decision Tree**:
- Task completed → /clear immediately
- Context > 80% → /clear + document decisions
- Debugging (3+ attempts) → /clear stale logs
- Switching tasks → /clear + update memory
- Poor quality → /clear + re-state requirements

## Clearing Workflow

```bash
# Step 1: Complete task
implement_feature

# Step 2: Document BEFORE clearing
echo "Auth implementation complete" >> .moai/memory/decisions.md

# Step 3: Execute /clear

# Step 4: Start fresh
start_next_task
```

## Anti-Patterns

**Bad Context**:
- Session History: 80K tokens (40% - too much)
- Project Context: 90K tokens (45% - overloaded)
- Available: 23K tokens (11.5% - insufficient)

**Good Context**:
- Session History: 15K tokens (7.5%)
- Project Context: 25K tokens (12.5%)
- Available: 155K tokens (77.5% - optimal)

---

**Reference**: [Claude Code Context](https://docs.claude.com/context-windows)
