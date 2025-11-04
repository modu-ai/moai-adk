---
name: moai-alfred-ask-user-questions
version: 3.1.0
created: 2025-10-22
updated: 2025-11-04
status: active
description: Guide Alfred sub-agents to actively invoke AskUserQuestion for ambiguous decisions.
keywords: ['interactive', 'clarification', 'decision-making', 'AskUserQuestion']
allowed-tools:
  - AskUserQuestion
---

# Alfred Ask User Questions - Skill Guide

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-ask-user-questions |
| **Version** | 3.1.0 (2025-10-27) |
| **Core Tool** | `AskUserQuestion` (Claude Code built-in) |
| **Auto-load** | When Alfred detects ambiguity in requests |
| **Tier** | Alfred (Workflow Orchestration) |

---

## What It Does

**Purpose**: Empower Alfred sub-agents to **actively ask clarifying questions** whenever user intent is ambiguous, rather than guessing.

Leverages Claude Code's native `AskUserQuestion` tool to collect explicit, structured user input that transforms vague requests into precise specifications.

**Key capabilities**:
- ✅ Single-select & multi-select option types
- ✅ 1-4 questions per survey (avoid fatigue)
- ✅ 2-4 options per question (prevent choice overload)
- ✅ Automatic "Other" option for custom input
- ✅ Conditional branching based on answers
- ✅ Integration across all Alfred commands (Plan/Run/Sync)
- ✅ Reduces ambiguity → fewer iterations → faster execution

---

## When to Ask (Trigger Patterns)

### ✅ ASK when user intent is ambiguous:

1. **Vague noun phrases**: "Add dashboard", "Refactor auth", "Improve performance"
2. **Missing scope**: No specification of WHERE, WHO, WHAT, HOW, WHEN
3. **Multiple valid paths**: ≥2 reasonable implementation approaches
4. **Trade-off decisions**: Speed vs quality, simple vs comprehensive, etc.
5. **Risky operations**: Destructive actions needing explicit consent

### ❌ DON'T ask when:
- User explicitly specified exact requirements
- Decision is automatic (no choices)
- Single obvious path exists
- Quick yes/no confirmation only (maybe, but keep it brief)

---

## Core Principle: Just Ask!

**Golden Rule**: When in doubt, **ask the user** instead of guessing.

**Why**:
- ✅ User sees exactly what you'll do → no surprises
- ✅ Single interaction vs 3-5 rounds of back-and-forth
- ✅ Fast → execute with certainty
- ✅ Reduces "vibe coding" frustration

**Pattern**:
```
Ambiguous request detected
         ↓
Call AskUserQuestion({questions: [...]})
         ↓
User selects from clear options
         ↓
Proceed with confirmed specifications
```

---

## Quick Start: Minimal Invocation

```typescript
const answer = await AskUserQuestion({
  questions: [
    {
      question: "How should we implement this?",
      header: "Approach",          // max 12 chars
      multiSelect: false,
      options: [
        {
          label: "Option 1",       // 1-5 words
          description: "What it does and why you'd pick it."
        },
        {
          label: "Option 2",
          description: "Alternative with different trade-offs."
        }
      ]
    }
  ]
});

// Returns: { "Approach": "Option 1" }
```

---

## Key Constraints

| Constraint | Reason |
|-----------|--------|
| **1-4 questions max** | Avoid user fatigue |
| **2-4 options per Q** | Prevent choice overload |
| **Header ≤12 chars** | TUI layout fit |
| **Label 1-5 words** | Quick scanning |
| **Description required** | Enables informed choice |
| **Auto "Other" option** | Always available for custom input |

---

## Top 5 Usage Patterns

### Pattern 1: Implementation Approach
**Trigger**: "Add feature X" or "Build Y" without specifics

### Pattern 2: Confirmation (Risky Operations)
**Trigger**: Destructive action (delete, migrate, reset)

### Pattern 3: Multi-Option Feature Selection
**Trigger**: "Which framework/library/approach?"

### Pattern 4: Multi-Select (Independent Features)
**Trigger**: "Which features to enable/include?"

### Pattern 5: Sequential Questions (Conditional Flow)
**Trigger**: Dependent decisions (Q2 depends on Q1 answer)

> **Detailed examples and code**: See [examples.md](examples.md)

---

## Integration with Alfred Sub-agents

| Sub-agent | When to Ask | Example Trigger |
|-----------|-------------|-----------------|
| **spec-builder** (`/alfred:1-plan`) | SPEC title vague, scope undefined | "Add feature" without specifics |
| **code-builder** (`/alfred:2-run`) | Implementation approach unclear | Multiple valid implementation paths |
| **doc-syncer** (`/alfred:3-sync`) | Sync scope unclear | Full vs partial sync decision |

> **Detailed integration patterns**: See [reference.md](reference.md)

---

## Best Practices Summary

### ✅ DO
- **Be specific**: "Which database type?" not "What should we use?"
- **Provide context**: Include file names, scope, or impact
- **Order logically**: General → Specific; safest option first
- **Flag risks**: Use "NOT RECOMMENDED" or "CAUTION:" prefixes
- **Explain trade-offs**: Mention time, resources, complexity

### ❌ DON'T
- **Overuse questions**: Only ask when ambiguous
- **Too many options**: 2-4 per question max
- **Vague labels**: "Option A", "Use tokens", "Option 2"
- **Skip descriptions**: User needs rationale
- **Hide trade-offs**: Always mention implications

> **Complete best practices guide**: See [reference.md](reference.md)

---

## Error Handling

### User Cancels (ESC key)
```typescript
try {
  const answer = await AskUserQuestion({...});
} catch (error) {
  console.log("User cancelled survey");
  // Fall back to default or abort
}
```

### Validate Custom Input ("Other" option)
```typescript
const answer = await AskUserQuestion({...});

if (answer["Header"] === "Other" || !VALID_OPTIONS.includes(answer["Header"])) {
  validateCustomInput(answer["Header"]);
}
```

---

## Related Skills

- `moai-alfred-spec-metadata-validation` (SPEC clarity)
- `moai-alfred-ears-authoring` (requirement phrasing)
- `moai-foundation-specs` (SPEC structure)

---

## Quick Reference

**When to use**:
- User intent is ambiguous
- Multiple valid approaches exist
- Architectural decisions with trade-offs
- Approvals needed before risky ops

**How**:
1. Detect ambiguity (vague request, multiple paths, etc.)
2. Call `AskUserQuestion({ questions: [...] })`
3. User selects from clear options
4. Proceed with confirmed specifications

**Benefits**:
- ✅ Certainty instead of guessing
- ✅ Single interaction vs 3-5 iterations
- ✅ Faster, happier users
- ✅ Less "vibe coding" frustration

---

**For detailed API specifications**: [reference.md](reference.md)  
**For real-world examples**: [examples.md](examples.md)

---

**End of Skill** | Refactored 2025-10-27
