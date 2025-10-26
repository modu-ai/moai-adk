# Alfred Interactive Questions - Real-World Examples

> **Main Skill**: [SKILL.md](SKILL.md)  
> **API Reference**: [reference.md](reference.md)

---

## Example 1: Vague Feature Request

### Scenario

**User**: "Add a completion page for the competition."

**Alfred detects ambiguity**:
- Where should it live? (new route vs modify existing)
- Who can access it? (public vs authenticated)
- What should it display? (results vs simple message)

### AskUserQuestion Invocation

```typescript
const answer = await AskUserQuestion({
  questions: [
    {
      question: "How should the completion page be implemented?",
      header: "Approach",
      multiSelect: false,
      options: [
        { 
          label: "New public route", 
          description: "New page at /competition-complete, visible to all visitors." 
        },
        { 
          label: "Modify existing page", 
          description: "Add conditional logic to existing /end page based on competition status." 
        },
        { 
          label: "Environment flag", 
          description: "Set NEXT_PUBLIC_COMPETITION_CLOSED=true to show message on homepage." 
        }
      ]
    },
    {
      question: "For logged-in participants, what should they see?",
      header: "Behavior",
      multiSelect: false,
      options: [
        { 
          label: "Show full results", 
          description: "Redirect to /end page with their submission history and rankings." 
        },
        { 
          label: "Show simple message", 
          description: "Display 'Competition concluded. Thank you for participating!' only." 
        }
      ]
    }
  ]
});

// Result:
// {
//   "Approach": "New public route",
//   "Behavior": "Show full results"
// }
```

### Outcome

Alfred now knows **exactly** what to build:
- Create `/app/competition-complete/page.tsx`
- New public route
- Logged-in users redirected to `/end` page with full results
- Non-logged-in users see simple completion message

---

## Example 2: Implementation Approach (Pattern 1)

### Scenario

**User**: "Add i18n support to the app."

**Alfred detects**:
- Multiple valid i18n libraries exist
- Each has different trade-offs (DX, complexity, performance)
- User hasn't specified preference

### AskUserQuestion Invocation

```typescript
const answer = await AskUserQuestion({
  questions: [{
    question: "Which i18n library should we use for Next.js 15?",
    header: "Library",
    multiSelect: false,
    options: [
      { 
        label: "next-intl", 
        description: "Native Next.js 15 support, best DX, automatic type safety." 
      },
      { 
        label: "react-i18next", 
        description: "Most popular, flexible, more setup required, larger bundle." 
      },
      { 
        label: "Format.js", 
        description: "ICU message format, powerful pluralization, steeper learning curve." 
      }
    ]
  }]
});

// Result: { "Library": "next-intl" }
```

### Outcome

Alfred proceeds with `next-intl`, avoiding:
- ❌ Guessing wrong library
- ❌ Implementing then refactoring later
- ❌ 3-5 rounds of "actually, can we use X instead?"

---

## Example 3: Risky Operation Confirmation (Pattern 2)

### Scenario

**User**: "Delete all test users from production database."

**Alfred detects**:
- Destructive operation (data loss)
- Production environment (high risk)
- Explicit confirmation required

### AskUserQuestion Invocation

```typescript
const answer = await AskUserQuestion({
  questions: [{
    question: "⚠️ This will PERMANENTLY delete 50 test users from PRODUCTION database. Are you sure?",
    header: "Confirm",
    multiSelect: false,
    options: [
      { 
        label: "No, cancel", 
        description: "Abort operation. No changes will be made." 
      },
      { 
        label: "Yes, proceed", 
        description: "🚨 IRREVERSIBLE. Delete 50 users now." 
      }
    ]
  }]
});

// Result: { "Confirm": "No, cancel" } (user cancelled)
```

### Outcome

Alfred **aborts** operation without executing destructive command. User explicitly chose to cancel after seeing full risk disclosure.

---

## Example 4: Multi-Select Feature Selection (Pattern 4)

### Scenario

**User**: "Set up testing for the new project."

**Alfred detects**:
- Multiple testing frameworks can coexist
- User might want unit + E2E + visual testing
- No explicit preference stated

### AskUserQuestion Invocation

```typescript
const answer = await AskUserQuestion({
  questions: [{
    question: "Which testing frameworks should we set up? (Select all that apply)",
    header: "Test Tools",
    multiSelect: true,  // Multiple selections allowed
    options: [
      { 
        label: "Unit tests (Vitest)", 
        description: "Fast, modern, Vite-native. ~2min setup." 
      },
      { 
        label: "E2E tests (Playwright)", 
        description: "Browser automation, cross-browser. ~10min setup." 
      },
      { 
        label: "Visual regression", 
        description: "Screenshot comparison with Playwright. ~5min setup." 
      }
    ]
  }]
});

// Result: { "Test Tools": ["Unit tests (Vitest)", "E2E tests (Playwright)"] }
```

### Outcome

Alfred installs and configures:
- ✅ Vitest for unit tests
- ✅ Playwright for E2E tests
- ❌ Skips visual regression (not selected)

---

## Example 5: Sequential Questions (Conditional Flow, Pattern 5)

### Scenario

**User**: "Add authentication to the app."

**Alfred detects**:
- First decision: Enable auth? (Yes/No)
- Second decision: If yes, which provider? (depends on Q1)

### AskUserQuestion Invocation (Part 1)

```typescript
// Question 1: Enable authentication?
const q1 = await AskUserQuestion({
  questions: [{
    question: "Should we enable user authentication?",
    header: "Auth",
    multiSelect: false,
    options: [
      { 
        label: "Yes", 
        description: "User login required for protected routes." 
      },
      { 
        label: "No", 
        description: "Public access only, no authentication." 
      }
    ]
  }]
});

// Result: { "Auth": "Yes" }
```

### AskUserQuestion Invocation (Part 2, Conditional)

```typescript
// Question 2: Only ask if Q1 = "Yes"
if (q1["Auth"] === "Yes") {
  const q2 = await AskUserQuestion({
    questions: [{
      question: "Which authentication provider?",
      header: "Provider",
      multiSelect: false,
      options: [
        { 
          label: "JWT + email", 
          description: "Traditional email/password with JWT tokens. ~30min setup." 
        },
        { 
          label: "OAuth (Google)", 
          description: "Third-party login via Google. Faster for users. ~20min setup." 
        },
        { 
          label: "SAML", 
          description: "Enterprise SSO (for corporate environments). ~2hr setup." 
        }
      ]
    }]
  });
  
  // Result: { "Provider": "OAuth (Google)" }
}
```

### Outcome

Alfred implements:
- ✅ Authentication enabled
- ✅ OAuth with Google provider
- ✅ Skips JWT setup (not selected)
- ❌ If Q1 was "No", Q2 never asked (saves time)

---

## Example 6: Option Grouping (Hierarchical Selection)

### Scenario

**User**: "Set up a database for the project."

**Alfred detects**:
- 8+ database options exist
- Too many for one question → choice paralysis
- Need hierarchical selection: Category → Specific

### AskUserQuestion Invocation (Step 1: Category)

```typescript
// Step 1: Narrow by database type
const category = await AskUserQuestion({
  questions: [{
    question: "What type of database does your project need?",
    header: "DB Type",
    multiSelect: false,
    options: [
      { 
        label: "Relational (SQL)", 
        description: "PostgreSQL, MySQL, SQLite. ACID guarantees, structured data." 
      },
      { 
        label: "Document (NoSQL)", 
        description: "MongoDB, CouchDB. Flexible schema, JSON documents." 
      },
      { 
        label: "Key-Value", 
        description: "Redis, Memcached. In-memory cache, fast reads." 
      }
    ]
  }]
});

// Result: { "DB Type": "Relational (SQL)" }
```

### AskUserQuestion Invocation (Step 2: Specific Choice)

```typescript
// Step 2: Specific database within category
if (category["DB Type"] === "Relational (SQL)") {
  const specific = await AskUserQuestion({
    questions: [{
      question: "Which SQL database?",
      header: "SQL DB",
      multiSelect: false,
      options: [
        { 
          label: "PostgreSQL", 
          description: "Advanced features (JSON, full-text search, extensions). Industry standard." 
        },
        { 
          label: "MySQL", 
          description: "Most popular, wide hosting support, good performance." 
        },
        { 
          label: "SQLite", 
          description: "Embedded, serverless, single-file. Good for prototypes and small apps." 
        }
      ]
    }]
  });
  
  // Result: { "SQL DB": "PostgreSQL" }
}
```

### Outcome

Alfred avoids:
- ❌ Presenting 8 options at once (choice paralysis)
- ❌ User confusion ("What's the difference between MariaDB and MySQL?")

Instead:
- ✅ First narrows to category (3 options)
- ✅ Then presents specific choices within category (3 options)
- ✅ Clear, manageable decision flow

---

## Example 7: Batch Related Questions

### Scenario

**User**: "Initialize a new Next.js project with best practices."

**Alfred detects**:
- Multiple independent decisions needed
- All decisions are related to project setup
- Can ask 3 questions at once (no dependencies)

### AskUserQuestion Invocation

```typescript
const answers = await AskUserQuestion({
  questions: [
    {
      question: "Which database?",
      header: "Database",
      multiSelect: false,
      options: [
        { label: "PostgreSQL", description: "Relational, ACID-compliant, advanced features." },
        { label: "MongoDB", description: "Document store, flexible schema, rapid prototyping." }
      ]
    },
    {
      question: "Which testing frameworks? (Select all that apply)",
      header: "Testing",
      multiSelect: true,  // Multiple selections allowed
      options: [
        { label: "Unit tests", description: "Vitest for fast unit testing." },
        { label: "E2E tests", description: "Playwright for browser automation." },
        { label: "Visual tests", description: "Screenshot comparison testing." }
      ]
    },
    {
      question: "Enable TypeScript strict mode?",
      header: "TypeScript",
      multiSelect: false,
      options: [
        { label: "Yes (recommended)", description: "Strict type checking, catch errors early." },
        { label: "No", description: "Loose type checking, faster initial development." }
      ]
    }
  ]
});

// Result:
// {
//   "Database": "PostgreSQL",
//   "Testing": ["Unit tests", "E2E tests"],
//   "TypeScript": "Yes (recommended)"
// }
```

### Outcome

Alfred configures project with:
- ✅ PostgreSQL database
- ✅ Vitest unit tests
- ✅ Playwright E2E tests
- ❌ No visual regression tests
- ✅ TypeScript strict mode enabled

All decided in **one interaction** instead of 3 separate back-and-forth exchanges.

---

## Example 8: Error Recovery After Test Failure

### Scenario

**Context**: Alfred is in `/alfred:2-run` phase, tests are failing.

**Alfred detects**:
- Tests fail after implementation
- Multiple possible fixes exist
- User input needed to decide repair strategy

### AskUserQuestion Invocation

```typescript
const answer = await AskUserQuestion({
  questions: [{
    question: "Tests are failing. How should we proceed?",
    header: "Fix Strategy",
    multiSelect: false,
    options: [
      { 
        label: "Fix implementation", 
        description: "Adjust code to pass existing tests (tests are correct)." 
      },
      { 
        label: "Update tests", 
        description: "Modify tests to match new implementation (implementation is correct)." 
      },
      { 
        label: "Debug together", 
        description: "Review both implementation and tests to identify root cause." 
      }
    ]
  }]
});

// Result: { "Fix Strategy": "Fix implementation" }
```

### Outcome

Alfred:
- ✅ Keeps tests unchanged
- ✅ Fixes implementation to satisfy test expectations
- ✅ Avoids "test modification to make them pass" anti-pattern
- ✅ Follows RED → GREEN → REFACTOR TDD cycle correctly

---

## Example 9: SPEC Scope Clarification (spec-builder)

### Scenario

**User**: "Create SPEC for performance improvements."

**Alfred detects** (spec-builder sub-agent):
- SPEC title is vague ("performance improvements")
- Scope undefined (backend? frontend? database?)
- Multiple domains affected

### AskUserQuestion Invocation

```typescript
const answer = await AskUserQuestion({
  questions: [
    {
      question: "Which area needs performance improvements?",
      header: "Domain",
      multiSelect: false,
      options: [
        { label: "Backend API", description: "Optimize API response times, database queries." },
        { label: "Frontend", description: "Improve page load times, bundle size, rendering." },
        { label: "Database", description: "Add indexes, optimize queries, caching." },
        { label: "All of the above", description: "Comprehensive performance audit." }
      ]
    },
    {
      question: "What's the performance target?",
      header: "Target",
      multiSelect: false,
      options: [
        { label: "< 200ms API", description: "API responses under 200ms (P95)." },
        { label: "< 2s page load", description: "Pages load under 2 seconds (P95)." },
        { label: "< 100ms queries", description: "Database queries under 100ms (P95)." }
      ]
    }
  ]
});

// Result:
// {
//   "Domain": "Backend API",
//   "Target": "< 200ms API"
// }
```

### Outcome

spec-builder generates precise SPEC:
- ✅ Title: `[PERF-001] Optimize Backend API Response Times to <200ms`
- ✅ Scope: Backend API performance only
- ✅ Clear success criteria: P95 latency < 200ms
- ❌ Avoids vague SPEC like "Improve performance"

---

## Example 10: Doc Sync Mode Selection (doc-syncer)

### Scenario

**User**: "Sync documentation."

**Alfred detects** (doc-syncer sub-agent):
- Sync mode unclear (auto-detect vs force-regenerate vs partial)
- User hasn't specified preference
- Different modes have different time/accuracy trade-offs

### AskUserQuestion Invocation

```typescript
const answer = await AskUserQuestion({
  questions: [{
    question: "Which documentation sync mode?",
    header: "Sync Mode",
    multiSelect: false,
    options: [
      { 
        label: "auto", 
        description: "Smart detection of changes. Fast, minimal updates. (Recommended)" 
      },
      { 
        label: "force", 
        description: "Regenerate ALL documentation. Slow but guaranteed accurate." 
      },
      { 
        label: "partial", 
        description: "Update only specified sections. Fast, targeted." 
      }
    ]
  }]
});

// Result: { "Sync Mode": "auto" }
```

### Outcome

doc-syncer:
- ✅ Uses auto-detection mode
- ✅ Only updates changed sections
- ✅ Faster execution (1-2 min vs 10 min for force mode)
- ✅ User sees clear time/accuracy trade-off before deciding

---

## Summary: Pattern Catalog

| Pattern | Use Case | Example Scenario |
|---------|----------|------------------|
| **Pattern 1: Implementation Approach** | Multiple valid implementation paths | "Add i18n support" (which library?) |
| **Pattern 2: Confirmation** | Risky/destructive operations | "Delete production data" (are you sure?) |
| **Pattern 3: Multi-Option Selection** | Choose ONE from multiple frameworks/tools | "Which testing framework?" |
| **Pattern 4: Multi-Select** | Enable/disable independent features | "Which features to enable?" (check multiple) |
| **Pattern 5: Sequential Questions** | Q2 depends on Q1 answer | "Enable auth?" → If yes, "Which provider?" |
| **Pattern 6: Option Grouping** | Too many options → hierarchical selection | Database type → Specific database |
| **Pattern 7: Batch Questions** | Multiple independent decisions | Project setup (DB + Testing + TypeScript) |
| **Pattern 8: Error Recovery** | Handle failures with user guidance | Test failure → Fix implementation vs update tests |
| **Pattern 9: SPEC Clarification** | Vague SPEC scope/title | "Performance improvements" → Which domain? |
| **Pattern 10: Mode Selection** | Choose workflow mode/strategy | Doc sync mode (auto vs force vs partial) |

---

**End of Examples** | 2025-10-27
