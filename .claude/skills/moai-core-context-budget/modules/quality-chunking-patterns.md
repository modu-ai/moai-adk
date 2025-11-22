---
name: quality-chunking-patterns
parent: moai-core-context-budget
description: Quality-focused context and task chunking
---

# Module 3: Quality & Chunking Patterns

## Quality Over Quantity

**Principle**: 10% context with highly relevant info > 90% with noise

## Context Quality Checklist

Before adding to context:

1. **Relevance**: Directly supports current task?
2. **Freshness**: Current information (< 1 hour)?
3. **Actionability**: Will Claude use this?
4. **Uniqueness**: Not duplicated elsewhere?

## Task Chunking Strategy

```typescript
interface Task {
  name: string;
  estimatedTokens: number;
}

const MAX_CHUNK_TOKENS = 120_000; // 60% of 200K

// Example: Auth system (250K total)
const chunks: Task[] = [
  { name: "User model & hashing", estimatedTokens: 80_000 },
  { name: "JWT generation", estimatedTokens: 70_000 },
  { name: "Login/logout endpoints", estimatedTokens: 60_000 },
  { name: "Session middleware", estimatedTokens: 40_000 }
];

// Workflow:
// 1. Complete Chunk 1
// 2. /clear
// 3. Document results
// 4. Start Chunk 2
```

## Common Pitfalls

**Bad Chunking**:
- Chunk 1 (200K): Everything at once → overflow

**Good Chunking**:
- Chunk 1 (60K): User auth only
- Chunk 2 (70K): Payment processing
- Chunk 3 (50K): Notifications

## High-Quality Context Example

```yaml
Total: 30K tokens (15%)

Contents:
  - CLAUDE.md (2K)
  - src/auth/jwt.ts (5K) - current file
  - src/types/auth.ts (3K) - needed types
  - session-summary.md (4K) - current state
  - tests/auth.test.ts (8K) - reference
  - Last 5 messages (8K) - recent context

Quality: HIGH - every token is relevant
```

---

**Reference**: [Context Optimization Guide](https://sparkco.ai/blog/claude-context-window-2025)
