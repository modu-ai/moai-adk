# Alfred Interactive Questions - API Reference

> **Main Skill**: [SKILL.md](SKILL.md)  
> **Examples**: [examples.md](examples.md)

---

## AskUserQuestion API Specification

### Function Signature

```typescript
async function AskUserQuestion(params: {
  questions: Question[]
}): Promise<Record<string, string | string[]>>

interface Question {
  question: string;         // The question text
  header: string;          // Column header (max 12 chars)
  multiSelect: boolean;    // true = multiple selections, false = single
  options: Option[];       // 2-4 options recommended
}

interface Option {
  label: string;          // 1-5 words, displayed in TUI
  description: string;    // Rationale and context
}
```

### Return Value

**Single-select** (`multiSelect: false`):
```typescript
{ "Header": "Selected Label" }
```

**Multi-select** (`multiSelect: true`):
```typescript
{ "Header": ["Label1", "Label2", "Label3"] }
```

### Auto-Generated "Other" Option

Claude Code automatically adds an "Other" option to every question, allowing users to provide custom input. You don't need to manually include it.

---

## Question Types

### Single-Select Questions

**Use when**: Mutually exclusive choices (choose ONE).

**Example domains**:
- Database selection (PostgreSQL OR MySQL OR SQLite)
- Authentication method (JWT OR OAuth OR SAML)
- Deployment target (AWS OR GCP OR Azure)

**Configuration**:
```typescript
{
  multiSelect: false,
  options: [
    { label: "PostgreSQL", description: "Relational, ACID-compliant." },
    { label: "MongoDB", description: "Document store, flexible schema." },
    { label: "Redis", description: "Key-value, in-memory cache." }
  ]
}
```

### Multi-Select Questions

**Use when**: Independent options that can be combined (check any that apply).

**Example domains**:
- Testing frameworks (Unit tests AND E2E tests AND Visual regression)
- Feature flags (Enable analytics AND Enable notifications)
- Deployment environments (Staging AND Production)

**Configuration**:
```typescript
{
  multiSelect: true,
  options: [
    { label: "Unit tests (Vitest)", description: "Fast, modern." },
    { label: "E2E tests (Playwright)", description: "Browser automation." },
    { label: "Visual regression", description: "Screenshot comparison." }
  ]
}
```

---

## Parameter Constraints

### Question Limits

| Parameter | Min | Max | Reason |
|-----------|-----|-----|--------|
| **Questions per call** | 1 | 4 | Avoid user fatigue |
| **Options per question** | 2 | 4 | Prevent choice overload |
| **Header length** | 1 | 12 chars | TUI layout constraints |
| **Label length** | 1 word | 5 words | Quick scanning |
| **Description length** | 10 chars | 200 chars | Provide context without overwhelming |

### Header Guidelines

**Good headers** (≤12 chars):
- ✅ "Approach"
- ✅ "Database"
- ✅ "Auth Method"
- ✅ "Deploy To"

**Bad headers** (too long):
- ❌ "Implementation Strategy" (23 chars)
- ❌ "Which database should we use?" (31 chars)

### Label Guidelines

**Good labels** (1-5 words):
- ✅ "New component"
- ✅ "Extend existing"
- ✅ "PostgreSQL with pgvector"

**Bad labels**:
- ❌ "Option 1" (vague)
- ❌ "Use the new standalone component approach with React hooks" (too long)

### Description Guidelines

**Good descriptions** (concise, informative):
- ✅ "Isolated, reusable, easier to test."
- ✅ "Relational database with ACID guarantees."
- ✅ "Fast in-memory cache, no persistence."

**Bad descriptions**:
- ❌ "Good option." (too vague)
- ❌ "This is a database that stores data in tables with rows and columns and supports SQL queries..." (too long)

---

## Integration Patterns by Sub-agent

### spec-builder Integration

**When**: SPEC title or scope is ambiguous.

**Example triggers**:
- "Add feature" without details
- "Refactor auth" without scope
- "Improve performance" without metrics

**Sample invocation**:
```typescript
if (titleIsAmbiguous(specTitle)) {
  const answer = await AskUserQuestion({
    questions: [{
      question: "Can you clarify what this SPEC should cover?",
      header: "Scope",
      multiSelect: false,
      options: [
        { label: "New feature", description: "Add new functionality." },
        { label: "Bug fix", description: "Resolve existing issue." },
        { label: "Refactor", description: "Improve code quality." }
      ]
    }]
  });
  specTitle = answer["Scope"];
}
```

### code-builder Integration

**When**: Implementation approach is unclear or multiple valid paths exist.

**Example triggers**:
- Multiple design patterns apply
- Trade-off between performance and simplicity
- Error recovery after failing tests

**Sample invocation**:
```typescript
if (multipleValidImplementationPaths) {
  const answer = await AskUserQuestion({
    questions: [{
      question: "How should we implement this?",
      header: "Approach",
      multiSelect: false,
      options: [
        { label: "Singleton pattern", description: "Single global instance." },
        { label: "Factory pattern", description: "Create instances dynamically." },
        { label: "Dependency injection", description: "Pass dependencies externally." }
      ]
    }]
  });
  implementationStrategy = answer["Approach"];
}
```

### doc-syncer Integration

**When**: Sync scope, mode, or PR Ready status is unclear.

**Example triggers**:
- Full documentation regeneration vs partial update
- PR Ready confirmation before merging
- Documentation coverage decision

**Sample invocation**:
```typescript
const syncMode = await AskUserQuestion({
  questions: [{
    question: "Which sync mode?",
    header: "Mode",
    multiSelect: false,
    options: [
      { label: "auto", description: "Smart detection of changes." },
      { label: "force", description: "Regenerate all documentation." },
      { label: "partial", description: "Update only changed sections." }
    ]
  }]
});
```

---

## Advanced Patterns

### Conditional Branching (Sequential Questions)

**When to use**: Question 2 depends on Question 1 answer.

**Example**: Authentication setup
```typescript
// Question 1: Enable authentication?
const q1 = await AskUserQuestion({
  questions: [{
    question: "Enable authentication?",
    header: "Auth",
    multiSelect: false,
    options: [
      { label: "Yes", description: "User login required." },
      { label: "No", description: "Public access only." }
    ]
  }]
});

// Question 2: Only if Q1 = "Yes"
if (q1["Auth"] === "Yes") {
  const q2 = await AskUserQuestion({
    questions: [{
      question: "Which auth provider?",
      header: "Provider",
      multiSelect: false,
      options: [
        { label: "JWT + email", description: "Traditional email/password." },
        { label: "OAuth (Google)", description: "Third-party Google login." },
        { label: "SAML", description: "Enterprise SSO." }
      ]
    }]
  });
}
```

### Batching Related Questions

**When to use**: Multiple independent questions that naturally flow together.

**Example**: Project setup
```typescript
const answers = await AskUserQuestion({
  questions: [
    {
      question: "Which database?",
      header: "Database",
      multiSelect: false,
      options: [
        { label: "PostgreSQL", description: "Relational, ACID-compliant." },
        { label: "MongoDB", description: "Document store, flexible schema." }
      ]
    },
    {
      question: "Which testing frameworks?",
      header: "Testing",
      multiSelect: true,  // Multiple selections allowed
      options: [
        { label: "Unit tests", description: "Fast, isolated tests." },
        { label: "E2E tests", description: "Full user flow testing." },
        { label: "Visual tests", description: "Screenshot comparison." }
      ]
    }
  ]
});

// Access answers:
const db = answers["Database"];         // "PostgreSQL"
const testing = answers["Testing"];     // ["Unit tests", "E2E tests"]
```

### Option Grouping (Hierarchical Selection)

**When to use**: Too many options → first narrow category, then specific choice.

**Example**: Database selection with 8+ options
```typescript
// Step 1: Narrow by category
const category = await AskUserQuestion({
  questions: [{
    question: "Database type?",
    header: "DB Type",
    multiSelect: false,
    options: [
      { label: "Relational (SQL)", description: "PostgreSQL, MySQL, etc." },
      { label: "Document (NoSQL)", description: "MongoDB, CouchDB, etc." },
      { label: "Key-Value", description: "Redis, Memcached, etc." }
    ]
  }]
});

// Step 2: Specific choice within category
if (category["DB Type"] === "Relational (SQL)") {
  const specific = await AskUserQuestion({
    questions: [{
      question: "Which SQL database?",
      header: "SQL DB",
      multiSelect: false,
      options: [
        { label: "PostgreSQL", description: "Advanced features, extensions." },
        { label: "MySQL", description: "Popular, wide support." },
        { label: "SQLite", description: "Embedded, serverless." }
      ]
    }]
  });
}
```

---

## Best Practices (Complete Guide)

### DO

1. **Be specific in questions**
   - ✅ "Which i18n library for Next.js 15?"
   - ❌ "What should we use?"

2. **Provide contextual information**
   - Include affected files, scope, or impact
   - Example: "This will modify 3 files in src/auth/"

3. **Order options logically**
   - Safest/most common option first
   - Risky/experimental options last
   - Mark dangerous options clearly

4. **Flag risks explicitly**
   - Use prefixes: "⚠️ CAUTION:", "🚨 NOT RECOMMENDED:"
   - Example: "⚠️ Force push (data loss risk)"

5. **Explain trade-offs**
   - Mention time, complexity, resources
   - Example: "Fast but less accurate" vs "Slower but comprehensive"

6. **Use single-select for exclusive choices**
   - Database: PostgreSQL OR MySQL (not both)
   - Auth: JWT OR OAuth (not both)

7. **Use multi-select for combinable options**
   - Testing: Unit AND E2E AND Visual (all applicable)
   - Features: Analytics AND Logging AND Monitoring (all applicable)

8. **Include descriptions for every option**
   - User needs rationale to make informed choice
   - Mention key pros/cons or use cases

9. **Keep headers short (≤12 chars)**
   - TUI layout constraint
   - "Approach" not "Implementation Strategy"

10. **Batch related questions**
    - Ask 2-3 at once if they naturally flow
    - Reduces back-and-forth

### DON'T

1. **Don't overuse questions**
   - Only ask when genuinely ambiguous
   - Don't ask for decisions that have obvious defaults

2. **Don't provide too many options (>4)**
   - Choice paralysis sets in
   - Use hierarchical selection instead

3. **Don't use vague labels**
   - ❌ "Option A", "Approach 2"
   - ✅ "PostgreSQL", "Factory pattern"

4. **Don't skip descriptions**
   - User needs context to decide
   - Every option must have a description

5. **Don't hide trade-offs**
   - Always mention implications (time, complexity, risk)
   - Don't present false equivalence

6. **Don't make destructive actions default**
   - Risky option should be clearly marked
   - Never pre-select dangerous operations

7. **Don't mix concerns in one question**
   - One decision per question
   - Don't ask "Which database and auth method?" as one question

8. **Don't manually add "Other" option**
   - It's auto-provided by Claude Code
   - You'll duplicate it unnecessarily

9. **Don't nest more than 2 levels deep**
   - Keep conditional flow linear
   - Avoid Q1 → Q2 → Q3 → Q4 chains

10. **Don't ask trivial questions**
    - If answer is obvious from context, don't ask
    - Example: Don't ask "Create file?" when user said "Create file X"

---

## Anti-Patterns to Avoid

### ❌ Too Many Options (Choice Paralysis)

**Bad**:
```typescript
options: [
  { label: "PostgreSQL" }, { label: "MySQL" }, { label: "MariaDB" },
  { label: "SQLite" }, { label: "MongoDB" }, { label: "CouchDB" },
  { label: "Cassandra" }, { label: "Redis" }
]
```

**Good** (Group first):
```typescript
// Question 1: Category
options: [
  { label: "Relational (SQL)", description: "PostgreSQL, MySQL, etc." },
  { label: "Document (NoSQL)", description: "MongoDB, CouchDB, etc." },
  { label: "Key-Value", description: "Redis, Memcached, etc." }
]

// Question 2: Specific within category
options: [
  { label: "PostgreSQL", description: "Advanced features." },
  { label: "MySQL", description: "Popular, wide support." },
  { label: "SQLite", description: "Embedded, serverless." }
]
```

### ❌ Vague Descriptions

**Bad**:
```typescript
{ label: "Option 1", description: "Good choice." }
{ label: "Option 2", description: "Also works." }
```

**Good**:
```typescript
{ label: "PostgreSQL", description: "Relational database with ACID guarantees and advanced features (JSON, full-text search)." }
{ label: "MongoDB", description: "Document store with flexible schema, good for rapid prototyping." }
```

### ❌ Missing Context

**Bad**:
```typescript
question: "Which database?"
```

**Good**:
```typescript
question: "Which database for the user authentication system? (Will store ~10K users, requires ACID guarantees)"
```

### ❌ Hidden Trade-offs

**Bad**:
```typescript
{ label: "Approach A", description: "Uses pattern X." }
{ label: "Approach B", description: "Uses pattern Y." }
```

**Good**:
```typescript
{ label: "Approach A", description: "Fast implementation (2 hours) but harder to test." }
{ label: "Approach B", description: "Slower implementation (5 hours) but more maintainable and testable." }
```

---

## Performance Optimization

### Batch Questions When Possible

**Avoid** (3 separate calls):
```typescript
const q1 = await AskUserQuestion({...});
const q2 = await AskUserQuestion({...});
const q3 = await AskUserQuestion({...});
```

**Prefer** (1 call with 3 questions):
```typescript
const answers = await AskUserQuestion({
  questions: [
    { question: "Q1?", header: "H1", ... },
    { question: "Q2?", header: "H2", ... },
    { question: "Q3?", header: "H3", ... }
  ]
});
```

**Exception**: When Q2 depends on Q1 answer (conditional branching).

### Pre-generate Options from Codebase Analysis

**Avoid**: Analyzing codebase multiple times to generate options.

**Prefer**: Analyze once, generate all options upfront.

```typescript
// Analyze codebase once
const availableDatabases = analyzeCodebase();

// Use results in question
const answer = await AskUserQuestion({
  questions: [{
    question: "Which database?",
    header: "Database",
    multiSelect: false,
    options: availableDatabases.map(db => ({
      label: db.name,
      description: db.description
    }))
  }]
});
```

### Minimize Sequential Calls

**Avoid**: Long chains of dependent questions.

**Prefer**: Smart defaults or inferred decisions when possible.

```typescript
// Only ask critical questions
// Infer or use sensible defaults for non-critical decisions
```

---

## Error Handling Patterns

### User Cancels Survey (ESC Key)

```typescript
try {
  const answer = await AskUserQuestion({
    questions: [...]
  });
  
  // Proceed with selected options
  implementFeature(answer);
  
} catch (error) {
  console.log("User cancelled survey");
  
  // Option 1: Use sensible default
  implementFeature({ "Approach": "default" });
  
  // Option 2: Abort operation
  return { status: "cancelled", message: "Operation aborted by user" };
}
```

### Validate Custom Input ("Other" Option)

```typescript
const answer = await AskUserQuestion({
  questions: [{
    question: "Which database?",
    header: "Database",
    multiSelect: false,
    options: [
      { label: "PostgreSQL", description: "..." },
      { label: "MySQL", description: "..." }
    ]
  }]
});

const VALID_OPTIONS = ["PostgreSQL", "MySQL"];

if (!VALID_OPTIONS.includes(answer["Database"])) {
  // User selected "Other" and provided custom input
  const customDb = answer["Database"];
  
  // Validate custom input
  if (!isSupportedDatabase(customDb)) {
    throw new Error(`Unsupported database: ${customDb}`);
  }
  
  console.log(`Using custom database: ${customDb}`);
}
```

### Handle Multi-Select Empty Selection

```typescript
const answer = await AskUserQuestion({
  questions: [{
    question: "Which testing frameworks?",
    header: "Testing",
    multiSelect: true,
    options: [...]
  }]
});

const selected = answer["Testing"];

// Check if user selected at least one option
if (!selected || selected.length === 0) {
  console.warn("No testing frameworks selected. Using default: Unit tests");
  implementTests(["Unit tests"]);
} else {
  implementTests(selected);
}
```

---

## Official Documentation References

**Claude Code Built-in Tools**:
- AskUserQuestion (native Claude Code tool)
- Interactive Prompting (CLAUDE.md § Clarification & Interactive Prompting)

**Related MoAI-ADK Documentation**:
- CLAUDE-RULES.md § Interactive Question Rules
- CLAUDE-PRACTICES.md § Ambiguity Handling
- CLAUDE-AGENTS-GUIDE.md § Sub-agent Collaboration

---

**End of Reference** | 2025-10-27
