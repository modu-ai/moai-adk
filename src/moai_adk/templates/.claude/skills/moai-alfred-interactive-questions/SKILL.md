---
name: moai-alfred-interactive-questions
version: 3.0.0
created: 2025-10-22
updated: 2025-10-23
status: active
description: Guide Alfred sub-agents to actively invoke AskUserQuestion for ambiguous decisions.
keywords: ['interactive', 'clarification', 'decision-making', 'AskUserQuestion']
allowed-tools:
  - AskUserQuestion
---

# Alfred Interactive Questions - Skill Guide

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-interactive-questions |
| **Version** | 3.0.0 (2025-10-23) |
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

## AskUserQuestion API Reference

### Minimal Invocation

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

### Key Constraints

| Constraint | Reason |
|-----------|--------|
| **1-4 questions max** | Avoid user fatigue |
| **2-4 options per Q** | Prevent choice overload |
| **Header ≤12 chars** | TUI layout fit |
| **Label 1-5 words** | Quick scanning |
| **Description required** | Enables informed choice |
| **Auto "Other" option** | Always available for custom input |

### Question Types

**Single-Select** (`multiSelect: false`):
- Mutually exclusive: "Choose ONE database"
- Returns: `{ "Header": "Selected Label" }`

**Multi-Select** (`multiSelect: true`):
- Independent options: "Which features to enable?" (check multiple)
- Returns: `{ "Header": ["Label1", "Label2"] }`

---

## Usage Patterns (Top 5)

### Pattern 1: Implementation Approach

**Trigger**: "Add feature X" or "Build Y" without specifics

```typescript
await AskUserQuestion({
  questions: [{
    question: "How should we implement this feature?",
    header: "Approach",
    multiSelect: false,
    options: [
      { label: "New standalone component", description: "Isolated, reusable." },
      { label: "Extend existing component", description: "Shared state, simpler." },
      { label: "Use environment flag", description: "Conditional visibility." }
    ]
  }]
});
```

### Pattern 2: Confirmation (Risky Operations)

**Trigger**: Destructive action (delete, migrate, reset)

```typescript
await AskUserQuestion({
  questions: [{
    question: "This will permanently delete 50 records. Proceed?",
    header: "Confirm",
    multiSelect: false,
    options: [
      { label: "Yes, proceed", description: "Irreversible. Continue." },
      { label: "No, cancel", description: "Abort operation." }
    ]
  }]
});
```

### Pattern 3: Multi-Option Feature Selection

**Trigger**: "Which framework/library/approach?"

```typescript
await AskUserQuestion({
  questions: [{
    question: "Which i18n library?",
    header: "Library",
    multiSelect: false,
    options: [
      { label: "next-intl", description: "Next.js 15+ native, best DX." },
      { label: "react-i18next", description: "Popular, flexible, more setup." },
      { label: "Format.js", description: "ICU format, powerful, complex." }
    ]
  }]
});
```

### Pattern 4: Multi-Select (Independent Features)

**Trigger**: "Which features to enable/include?"

```typescript
await AskUserQuestion({
  questions: [{
    question: "Which testing frameworks?",
    header: "Test Tools",
    multiSelect: true,  // Multiple selections allowed
    options: [
      { label: "Unit tests (Vitest)", description: "Fast, modern." },
      { label: "E2E tests (Playwright)", description: "Browser automation." },
      { label: "Visual regression", description: "Screenshot comparison." }
    ]
  }]
});
```

### Pattern 5: Sequential Questions (Conditional Flow)

**Trigger**: Dependent decisions (Q2 depends on Q1 answer)

```typescript
// Question 1
const q1 = await AskUserQuestion({
  questions: [{
    question: "Enable authentication?",
    header: "Auth",
    options: [
      { label: "Yes", description: "User login required." },
      { label: "No", description: "Public access." }
    ]
  }]
});

// Question 2 (only if Q1 = "Yes")
if (q1["Auth"] === "Yes") {
  const q2 = await AskUserQuestion({
    questions: [{
      question: "Which auth provider?",
      header: "Provider",
      options: [
        { label: "JWT + email", description: "Traditional." },
        { label: "OAuth", description: "Third-party." }
      ]
    }]
  });
}
```

---

## Best Practices

### ✅ DO

- **Be specific**: "Which database type?" not "What should we use?"
- **Provide context**: Include file names, scope, or impact
- **Order logically**: General → Specific; safest option first
- **Flag risks**: Use "NOT RECOMMENDED" or "CAUTION:" prefixes
- **Explain trade-offs**: Mention time, resources, complexity
- **Single-select for exclusive**: "Choose ONE"
- **Multi-select for combinable**: "Check any that apply"
- **Include descriptions**: Every option needs rationale
- **Keep headers short**: "Approach" not "Implementation Strategy"
- **Batch related questions**: Ask 2-3 at once if they flow naturally

### ❌ DON'T

- **Overuse questions**: Only ask when ambiguous
- **Too many options**: 2-4 per question max
- **Vague labels**: "Option A", "Use tokens", "Option 2"
- **Skip descriptions**: User needs rationale
- **Hide trade-offs**: Always mention implications
- **Make destructive default**: Risky option = NOT pre-selected
- **Mix concerns**: One decision per question
- **Manually add "Other"**: It's auto-provided
- **Nest more than 2 levels deep**: Keep flow linear

---

## Integration with Alfred Sub-agents

### spec-builder (`/alfred:1-plan`)

**When to ask**:
- SPEC title is vague ("Add feature")
- Scope undefined (what exactly?)
- Domain prefix unclear (AUTH vs UI vs DATA?)

**Example**:
```typescript
if (titleIsAmbiguous(specTitle)) {
  const answer = await AskUserQuestion({
    questions: [{
      question: "Can you clarify what this SPEC should cover?",
      header: "Scope",
      options: [...]
    }]
  });
  specTitle = answer["Scope"];
}
```

### code-builder (`/alfred:2-run`)

**When to ask**:
- Implementation approach unclear
- Multiple valid paths exist
- Error recovery needed (failing tests)

**Example**:
```typescript
if (multipleValidImplementationPaths) {
  const answer = await AskUserQuestion({
    questions: [{
      question: "How should we implement this?",
      header: "Approach",
      options: [...]
    }]
  });
  implementationStrategy = answer["Approach"];
}
```

### doc-syncer (`/alfred:3-sync`)

**When to ask**:
- Sync scope unclear (full vs partial)
- PR Ready status uncertain
- Documentation coverage decision needed

**Example**:
```typescript
const syncMode = await AskUserQuestion({
  questions: [{
    question: "Which sync mode?",
    header: "Mode",
    options: [
      { label: "auto", description: "Smart detection." },
      { label: "force", description: "Regenerate all." }
    ]
  }]
});
```

---

## Real-World Examples

### Example: Vague Feature Request

**User**: "Add a completion page for the competition."

**Alfred detects ambiguity**:
- Where should it live? (new route vs modify existing)
- Who can access it? (public vs authenticated)
- What should it display? (results vs simple message)

**AskUserQuestion invocation**:
```typescript
const answer = await AskUserQuestion({
  questions: [
    {
      question: "How should the completion page be implemented?",
      header: "Approach",
      multiSelect: false,
      options: [
        { label: "New public route", description: "New page visible to all visitors." },
        { label: "Modify existing page", description: "Add conditional logic to /end page." },
        { label: "Environment flag", description: "Set NEXT_PUBLIC_COMPETITION_CLOSED=true." }
      ]
    },
    {
      question: "For logged-in participants?",
      header: "Behavior",
      multiSelect: false,
      options: [
        { label: "Show full results", description: "Redirect to /end page with history." },
        { label: "Show simple message", description: "Display 'Competition concluded' only." }
      ]
    }
  ]
});
```

**Result**: Alfred knows exactly what to build.

---

## Anti-Patterns to Avoid

### ❌ Too Many Options
```typescript
// 8+ options = choice paralysis
options: [
  { label: "PostgreSQL" }, { label: "MySQL" }, { label: "MariaDB" },
  { label: "SQLite" }, { label: "MongoDB" }, { label: "CouchDB" },
  { label: "Cassandra" }, { label: "Redis" }
]
```

### ✅ Group Options
```typescript
// First ask: which TYPE? Then: specific within type
questions: [{
  question: "Database type?",
  header: "DB Type",
  options: [
    { label: "Relational (SQL)", description: "PostgreSQL, MySQL, etc." },
    { label: "Document (NoSQL)", description: "MongoDB, CouchDB, etc." },
    { label: "Key-Value", description: "Redis, Memcached, etc." }
  ]
}]
```

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
  // Handle custom input validation
  validateCustomInput(answer["Header"]);
}
```

---

## Performance Tips

- **Batch questions**: Ask 2-3 related questions in one call
- **Minimize calls**: Don't ask sequentially unless truly dependent
- **Pre-generate options**: Analyze codebase once to generate all options

---

## References

**Official Claude Code Documentation**:
- AskUserQuestion tool (built-in Claude Code)
- Interactive Prompting (CLAUDE.md § Clarification & Interactive Prompting)

**Related Skills**:
- `moai-alfred-spec-metadata-validation` (SPEC clarity)
- `moai-alfred-ears-authoring` (requirement phrasing)
- `moai-foundation-specs` (SPEC structure)

---

## Summary

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

**End of Skill** | Refactored 2025-10-23
