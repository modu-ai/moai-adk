# moai-alfred-tui-survey - Working Examples

_Last updated: 2025-10-22_

## Example 1: Simple Yes/No Approval

**Scenario**: User needs to approve a risky operation (like deleting data).

**AskUserQuestion Call**:
```typescript
AskUserQuestion({
  questions: [{
    question: "This will permanently delete 150 user records. Continue?",
    header: "Confirm Delete",
    multiSelect: false,
    options: [
      {
        label: "Yes, proceed with deletion",
        description: "Irreversible operation. All 150 records will be removed."
      },
      {
        label: "No, cancel operation",
        description: "Abort deletion and return to previous screen."
      }
    ]
  }]
})
```

**TUI Rendering**:
```
────────────────────────────────────────────────────────────────
ALFRED: This will permanently delete 150 user records. Continue?
────────────────────────────────────────────────────────────────

┌─ CONFIRM DELETE ─────────────────────────────────────────────┐
│                                                              │
│ ▶ Yes, proceed with deletion                                │
│   • Irreversible operation. All 150 records will be removed.│
│                                                              │
│   No, cancel operation                                      │
│   • Abort deletion and return to previous screen.           │
│                                                              │
│ Use ↑↓ arrows to navigate, ENTER to select                 │
│ Type custom answer or press ESC to cancel                   │
└──────────────────────────────────────────────────────────────┘
```

**Best Practices**:
- ✅ Clear consequence description
- ✅ Default selection on safest option
- ✅ Explicit escape route

---

## Example 2: Architecture Decision (Multi-Option)

**Scenario**: Choosing database migration strategy with trade-offs.

**AskUserQuestion Call**:
```typescript
AskUserQuestion({
  questions: [{
    question: "How should we handle the schema migration for the auth system?",
    header: "Migration Strategy",
    multiSelect: false,
    options: [
      {
        label: "Blue-Green Deployment",
        description: "Zero downtime, requires 2x resources temporarily. Safest option."
      },
      {
        label: "Rolling Update",
        description: "Gradual rollout, minimal resource overhead, 5-10 min downtime window."
      },
      {
        label: "In-Place Update",
        description: "Fastest approach, 15-30 min downtime, no additional resources."
      },
      {
        label: "Canary Release",
        description: "Test on 10% users first, longest timeline (2-3 days), lowest risk."
      }
    ]
  }]
})
```

**TUI Rendering**:
```
────────────────────────────────────────────────────────────────
ALFRED: How should we handle the schema migration for the auth system?
────────────────────────────────────────────────────────────────

┌─ MIGRATION STRATEGY ─────────────────────────────────────────┐
│                                                              │
│ ▶ Blue-Green Deployment                                      │
│   • Zero downtime, requires 2x resources temporarily.       │
│   • Safest option.                                          │
│                                                              │
│   Rolling Update                                            │
│   • Gradual rollout, minimal resource overhead.             │
│   • 5-10 min downtime window.                               │
│                                                              │
│   In-Place Update                                           │
│   • Fastest approach, 15-30 min downtime.                   │
│   • No additional resources.                                │
│                                                              │
│   Canary Release                                            │
│   • Test on 10% users first, longest timeline (2-3 days).   │
│   • Lowest risk.                                            │
│                                                              │
│ Use ↑↓ arrows to navigate, ENTER to select                 │
└──────────────────────────────────────────────────────────────┘
```

**Best Practices**:
- ✅ Trade-offs explicitly stated
- ✅ Options ordered by recommendation (safest first)
- ✅ Concise but informative descriptions

---

## Example 3: Multi-Select Feature Selection

**Scenario**: User selecting multiple features to enable during project setup.

**AskUserQuestion Call**:
```typescript
AskUserQuestion({
  questions: [{
    question: "Which optional features do you want to enable for your project?",
    header: "Feature Selection",
    multiSelect: true,  // Allow multiple selections
    options: [
      {
        label: "Authentication (OAuth + JWT)",
        description: "User login with Google/GitHub, JWT tokens. +2 dependencies."
      },
      {
        label: "Real-time Notifications",
        description: "WebSocket-based push notifications. +1 dependency (Socket.IO)."
      },
      {
        label: "Email Service",
        description: "Transactional emails via SendGrid. Requires API key setup."
      },
      {
        label: "Analytics Dashboard",
        description: "Built-in analytics with charts. +3 dependencies (Chart.js, etc.)."
      }
    ]
  }]
})
```

**TUI Rendering**:
```
────────────────────────────────────────────────────────────────
ALFRED: Which optional features do you want to enable for your project?
────────────────────────────────────────────────────────────────

┌─ FEATURE SELECTION (Multi-Select) ───────────────────────────┐
│                                                              │
│ [x] Authentication (OAuth + JWT)                            │
│     • User login with Google/GitHub, JWT tokens.            │
│     • +2 dependencies.                                      │
│                                                              │
│ [ ] Real-time Notifications                                 │
│     • WebSocket-based push notifications.                   │
│     • +1 dependency (Socket.IO).                            │
│                                                              │
│ [x] Email Service                                           │
│     • Transactional emails via SendGrid.                    │
│     • Requires API key setup.                               │
│                                                              │
│ [ ] Analytics Dashboard                                     │
│     • Built-in analytics with charts.                       │
│     • +3 dependencies (Chart.js, etc.).                     │
│                                                              │
│ Use ↑↓ to navigate, SPACE to toggle, ENTER to submit       │
└──────────────────────────────────────────────────────────────┘

Selected: Authentication, Email Service
```

**Best Practices**:
- ✅ Clear indication of multi-select mode
- ✅ Checkbox UI ([x] / [ ])
- ✅ Dependencies and requirements disclosed upfront

---

## Example 4: Sequential Question Flow

**Scenario**: Project initialization requiring multiple decisions.

**AskUserQuestion Call (Question 1)**:
```typescript
AskUserQuestion({
  questions: [{
    question: "Which framework would you like to use?",
    header: "Framework",
    multiSelect: false,
    options: [
      { label: "Next.js", description: "React with SSR, recommended for production." },
      { label: "Vite + React", description: "Fast dev server, SPA architecture." },
      { label: "SvelteKit", description: "Lightweight, great DX, smaller bundle." }
    ]
  }]
})
```

**AskUserQuestion Call (Question 2 - conditional on Q1 answer)**:
```typescript
// If user selected "Next.js"
AskUserQuestion({
  questions: [{
    question: "Which Next.js routing approach?",
    header: "Routing",
    multiSelect: false,
    options: [
      { label: "App Router (recommended)", description: "Latest Next.js 13+ pattern with Server Components." },
      { label: "Pages Router", description: "Traditional Next.js routing, stable and well-documented." }
    ]
  }]
})
```

**TUI Rendering (Sequential)**:
```
[QUESTION 1] Which framework would you like to use?
→ User selects: Next.js

[QUESTION 2] Which Next.js routing approach?
→ User selects: App Router (recommended)

[REVIEW] Summary
────────────────────────────────────────────
✓ Framework: Next.js
✓ Routing: App Router (recommended)

Ready to submit? [Submit] [← Go back]
```

**Best Practices**:
- ✅ Questions flow logically (general → specific)
- ✅ Conditional questions based on previous answers
- ✅ Summary review before final submission

---

## Example 5: SPEC Implementation Mode Selection

**Scenario**: `/alfred:2-run` asking user how to implement a SPEC.

**AskUserQuestion Call**:
```typescript
AskUserQuestion({
  questions: [
    {
      question: "How should we implement SPEC-AUTH-001 (JWT Authentication)?",
      header: "Implementation Mode",
      multiSelect: false,
      options: [
        {
          label: "Full TDD (RED → GREEN → REFACTOR)",
          description: "Write failing tests first, then implement. Recommended for critical features."
        },
        {
          label: "Prototype First",
          description: "Quick POC implementation, add tests after. Good for exploring solutions."
        },
        {
          label: "Test-After (TAD)",
          description: "Implement feature, then write comprehensive tests. Faster initial delivery."
        }
      ]
    },
    {
      question: "Should we create a new feature branch?",
      header: "Git Workflow",
      multiSelect: false,
      options: [
        {
          label: "Yes, create feature/AUTH-001",
          description: "Isolate changes in a dedicated branch. Recommended for team workflows."
        },
        {
          label: "No, work on current branch",
          description: "Continue on current branch (e.g., during prototyping)."
        }
      ]
    }
  ]
})
```

**TUI Rendering (Multi-Question Survey)**:
```
────────────────────────────────────────────────────────────────
ALFRED: Implementation Strategy for SPEC-AUTH-001
────────────────────────────────────────────────────────────────

[QUESTION 1/2] How should we implement SPEC-AUTH-001 (JWT Authentication)?

┌─ IMPLEMENTATION MODE ────────────────────────────────────────┐
│                                                              │
│ ▶ Full TDD (RED → GREEN → REFACTOR)                         │
│   • Write failing tests first, then implement.              │
│   • Recommended for critical features.                      │
│                                                              │
│   Prototype First                                           │
│   • Quick POC implementation, add tests after.              │
│   • Good for exploring solutions.                           │
│                                                              │
│   Test-After (TAD)                                          │
│   • Implement feature, then write comprehensive tests.      │
│   • Faster initial delivery.                                │
└──────────────────────────────────────────────────────────────┘

→ Selection: Full TDD (RED → GREEN → REFACTOR)

[QUESTION 2/2] Should we create a new feature branch?

┌─ GIT WORKFLOW ───────────────────────────────────────────────┐
│                                                              │
│ ▶ Yes, create feature/AUTH-001                              │
│   • Isolate changes in a dedicated branch.                  │
│   • Recommended for team workflows.                         │
│                                                              │
│   No, work on current branch                                │
│   • Continue on current branch (e.g., during prototyping).  │
└──────────────────────────────────────────────────────────────┘

→ Selection: Yes, create feature/AUTH-001

[REVIEW] Implementation Strategy Summary
────────────────────────────────────────────────────────────────
✓ Implementation Mode: Full TDD (RED → GREEN → REFACTOR)
✓ Git Workflow: Yes, create feature/AUTH-001

Ready to submit these answers?

 [✓ Submit answers]  [← Go back and modify]
```

**Best Practices**:
- ✅ Multi-question surveys grouped logically
- ✅ Progress indicator (1/2, 2/2)
- ✅ Comprehensive review step before execution

---

## Example 6: Error Recovery Decision

**Scenario**: Test failures detected, asking user how to proceed.

**AskUserQuestion Call**:
```typescript
AskUserQuestion({
  questions: [{
    question: "3 tests failed in auth.test.ts. How should we proceed?",
    header: "Test Failure",
    multiSelect: false,
    options: [
      {
        label: "Debug failures immediately",
        description: "Stop current workflow, analyze and fix test failures."
      },
      {
        label: "Skip tests and continue",
        description: "Mark tests as @skip, continue implementation. NOT RECOMMENDED."
      },
      {
        label: "Rollback recent changes",
        description: "Revert to last passing commit, investigate separately."
      },
      {
        label: "Update test expectations",
        description: "Tests may be outdated, adjust assertions to match new behavior."
      }
    ]
  }]
})
```

**TUI Rendering**:
```
────────────────────────────────────────────────────────────────
ALFRED: 3 tests failed in auth.test.ts. How should we proceed?
────────────────────────────────────────────────────────────────

┌─ TEST FAILURE ───────────────────────────────────────────────┐
│                                                              │
│ ▶ Debug failures immediately                                │
│   • Stop current workflow, analyze and fix test failures.   │
│                                                              │
│   Skip tests and continue                                   │
│   • Mark tests as @skip, continue implementation.           │
│   • NOT RECOMMENDED.                                        │
│                                                              │
│   Rollback recent changes                                   │
│   • Revert to last passing commit.                          │
│   • Investigate separately.                                 │
│                                                              │
│   Update test expectations                                  │
│   • Tests may be outdated.                                  │
│   • Adjust assertions to match new behavior.                │
│                                                              │
│ Use ↑↓ arrows to navigate, ENTER to select                 │
└──────────────────────────────────────────────────────────────┘
```

**Best Practices**:
- ✅ Context provided (3 tests, auth.test.ts)
- ✅ Warnings on risky options (Skip tests = NOT RECOMMENDED)
- ✅ Safest option listed first

---

## Example 7: Documentation Language Selection

**Scenario**: `/alfred:0-project` asking for documentation language preference.

**AskUserQuestion Call**:
```typescript
AskUserQuestion({
  questions: [{
    question: "Which language would you like to use for project initialization and documentation?",
    header: "Language",
    multiSelect: false,
    options: [
      {
        label: "English (en)",
        description: "All dialogs and documentation in English."
      },
      {
        label: "한국어 (ko)",
        description: "All dialogs and documentation in Korean."
      },
      {
        label: "日本語 (ja)",
        description: "All dialogs and documentation in Japanese."
      }
    ]
  }]
})
```

**TUI Rendering**:
```
────────────────────────────────────────────────────────────────
ALFRED: Which language would you like to use for project initialization and documentation?
────────────────────────────────────────────────────────────────

┌─ LANGUAGE ───────────────────────────────────────────────────┐
│                                                              │
│ ▶ English (en)                                               │
│   • All dialogs and documentation in English.               │
│                                                              │
│   한국어 (ko)                                                 │
│   • All dialogs and documentation in Korean.                │
│                                                              │
│   日本語 (ja)                                                 │
│   • All dialogs and documentation in Japanese.              │
│                                                              │
│ Use ↑↓ arrows to navigate, ENTER to select                 │
└──────────────────────────────────────────────────────────────┘
```

**Best Practices**:
- ✅ Native language labels (한국어, 日本語)
- ✅ Clear scope (dialogs AND documentation)
- ✅ Simple, non-technical language

---

## Example 8: Custom Input with "Other" Option

**Scenario**: User needs to specify a custom value not in the preset options.

**AskUserQuestion Call**:
```typescript
AskUserQuestion({
  questions: [{
    question: "Which test coverage target should we enforce?",
    header: "Coverage Target",
    multiSelect: false,
    options: [
      {
        label: "85% (recommended)",
        description: "TRUST 5 principle standard. Balances quality and speed."
      },
      {
        label: "90% (strict)",
        description: "Higher quality gate, slower development cycle."
      },
      {
        label: "75% (relaxed)",
        description: "Faster iteration, lower quality assurance."
      }
    ]
  }]
})
```

**TUI Rendering**:
```
────────────────────────────────────────────────────────────────
ALFRED: Which test coverage target should we enforce?
────────────────────────────────────────────────────────────────

┌─ COVERAGE TARGET ────────────────────────────────────────────┐
│                                                              │
│ ▶ 85% (recommended)                                          │
│   • TRUST 5 principle standard.                             │
│   • Balances quality and speed.                             │
│                                                              │
│   90% (strict)                                              │
│   • Higher quality gate, slower development cycle.          │
│                                                              │
│   75% (relaxed)                                             │
│   • Faster iteration, lower quality assurance.              │
│                                                              │
│   Other (custom value)                                      │
│   • Specify a custom coverage percentage.                   │
│                                                              │
│ Use ↑↓ arrows to navigate, ENTER to select                 │
│ Type custom answer when 'Other' is selected                 │
└──────────────────────────────────────────────────────────────┘

→ User selects "Other" → Prompt: "Enter custom coverage target (0-100):"
→ User types: "80"
→ Confirmation: "Coverage target set to 80%"
```

**Best Practices**:
- ✅ "Other" option always available (automatic)
- ✅ Clear prompt for custom input
- ✅ Validation of custom values (0-100 range)

---

## Example 9: Branching Logic Based on Answers

**Scenario**: Follow-up questions depend on previous answers.

**Question Flow**:
```
Q1: "Do you want to enable authentication?"
    → Yes: Ask Q2 ("Which auth provider?")
    → No: Skip to Q3 ("Database selection")

Q2: "Which authentication provider?"
    → OAuth: Ask Q2a ("Google, GitHub, or both?")
    → JWT: Ask Q2b ("Token expiry duration?")
    → Passkeys: Ask Q2c ("FIDO2 or WebAuthn?")

Q3: "Which database?"
    → PostgreSQL / MySQL / MongoDB / SQLite
```

**Implementation Pattern**:
```typescript
// Question 1
const authAnswer = await AskUserQuestion({
  questions: [{
    question: "Do you want to enable authentication?",
    header: "Authentication",
    options: [
      { label: "Yes", description: "Enable user authentication." },
      { label: "No", description: "Skip authentication setup." }
    ]
  }]
});

if (authAnswer["Authentication"] === "Yes") {
  // Question 2 (conditional)
  const providerAnswer = await AskUserQuestion({
    questions: [{
      question: "Which authentication provider?",
      header: "Auth Provider",
      options: [
        { label: "OAuth (Google/GitHub)", description: "Third-party login." },
        { label: "JWT (email/password)", description: "Traditional credentials." },
        { label: "Passkeys (FIDO2)", description: "Passwordless authentication." }
      ]
    }]
  });

  // Further branching based on providerAnswer...
}

// Question 3 (always asked)
const dbAnswer = await AskUserQuestion({
  questions: [{
    question: "Which database?",
    header: "Database",
    options: [...]
  }]
});
```

**Best Practices**:
- ✅ Conditional logic based on previous answers
- ✅ Skip irrelevant questions
- ✅ Maintain context across question flow

---

## Example 10: /alfred:3-sync Mode Selection

**Scenario**: Asking user which sync mode to use.

**AskUserQuestion Call**:
```typescript
AskUserQuestion({
  questions: [{
    question: "Which synchronization mode should we use?",
    header: "Sync Mode",
    multiSelect: false,
    options: [
      {
        label: "auto (recommended)",
        description: "Intelligently detect changes and sync affected documents only."
      },
      {
        label: "force",
        description: "Regenerate all documents from scratch. Use after major refactors."
      },
      {
        label: "status",
        description: "Show sync status without making changes. Read-only mode."
      },
      {
        label: "project",
        description: "Sync project-level docs only (product.md, structure.md, tech.md)."
      }
    ]
  }]
})
```

**TUI Rendering**:
```
────────────────────────────────────────────────────────────────
ALFRED: Which synchronization mode should we use?
────────────────────────────────────────────────────────────────

┌─ SYNC MODE ──────────────────────────────────────────────────┐
│                                                              │
│ ▶ auto (recommended)                                         │
│   • Intelligently detect changes.                           │
│   • Sync affected documents only.                           │
│                                                              │
│   force                                                     │
│   • Regenerate all documents from scratch.                  │
│   • Use after major refactors.                              │
│                                                              │
│   status                                                    │
│   • Show sync status without making changes.                │
│   • Read-only mode.                                         │
│                                                              │
│   project                                                   │
│   • Sync project-level docs only.                           │
│   • (product.md, structure.md, tech.md)                     │
│                                                              │
│ Use ↑↓ arrows to navigate, ENTER to select                 │
└──────────────────────────────────────────────────────────────┘
```

**Best Practices**:
- ✅ Mode names match command arguments (`/alfred:3-sync auto`)
- ✅ Recommended option clearly marked
- ✅ Use cases explained (when to use each mode)

---

## Integration with MoAI-ADK Commands

### /alfred:0-project
- **Language selection**: User picks documentation language (en/ko/ja)
- **Mode selection**: personal vs team project setup
- **Feature selection**: Multi-select optional features (auth, notifications, analytics)

### /alfred:1-plan
- **SPEC scope confirmation**: Review planned SPEC titles before creation
- **Domain assignment**: Choose domain prefix (AUTH, UI, DATA, API, etc.)
- **Branching strategy**: Create feature branch vs work on current branch

### /alfred:2-run
- **Implementation mode**: TDD vs Prototype vs Test-After
- **SPEC selection**: Which SPEC(s) to implement (if multiple pending)
- **Error handling**: How to proceed when tests fail

### /alfred:3-sync
- **Sync mode**: auto/force/status/project
- **PR Ready confirmation**: Move Draft PR to Ready for review?
- **Documentation scope**: Full sync vs selective updates

---

## Anti-Patterns to Avoid

### ❌ Bad: Too Many Options
```typescript
// 8+ options overwhelm the user
AskUserQuestion({
  questions: [{
    question: "Choose a database:",
    header: "Database",
    options: [
      { label: "PostgreSQL" },
      { label: "MySQL" },
      { label: "MariaDB" },
      { label: "SQLite" },
      { label: "MongoDB" },
      { label: "CouchDB" },
      { label: "Cassandra" },
      { label: "Redis" }
    ]
  }]
})
```

### ✅ Good: Grouped Options
```typescript
// Group by category, ask in sequence
AskUserQuestion({
  questions: [{
    question: "Which database type?",
    header: "DB Type",
    options: [
      { label: "Relational (SQL)", description: "PostgreSQL, MySQL, etc." },
      { label: "Document (NoSQL)", description: "MongoDB, CouchDB, etc." },
      { label: "Key-Value", description: "Redis, etc." }
    ]
  }]
})

// Then follow up with specific choice within category
```

### ❌ Bad: Vague Option Labels
```typescript
options: [
  { label: "Option A", description: "This is option A." },
  { label: "Option B", description: "This is option B." }
]
```

### ✅ Good: Descriptive Labels
```typescript
options: [
  {
    label: "Full backup (30 min downtime)",
    description: "Safest option, requires maintenance window."
  },
  {
    label: "Incremental backup (5 min downtime)",
    description: "Faster, but depends on previous backups."
  }
]
```

### ❌ Bad: No Context
```typescript
AskUserQuestion({
  questions: [{
    question: "Proceed?",
    header: "Confirm",
    options: [
      { label: "Yes" },
      { label: "No" }
    ]
  }]
})
```

### ✅ Good: Clear Context
```typescript
AskUserQuestion({
  questions: [{
    question: "This will modify 47 files and add 3 new dependencies. Proceed with implementation?",
    header: "Confirm Changes",
    options: [
      {
        label: "Yes, proceed",
        description: "Apply changes and install dependencies."
      },
      {
        label: "No, review first",
        description: "Show file list and dependency details before proceeding."
      }
    ]
  }]
})
```

---

## Keyboard Shortcuts Reference

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate options |
| `SPACE` | Toggle selection (multiSelect mode) |
| `ENTER` | Confirm selection |
| `ESC` | Cancel survey |
| `←` / `→` | Navigate back/forward in multi-question flow |

---

## Summary

The `moai-alfred-tui-survey` Skill enables:
- ✅ **Structured decision-making** via interactive TUI menus
- ✅ **Reduced ambiguity** by presenting explicit choices
- ✅ **Better UX** with visual navigation (arrow keys, checkboxes)
- ✅ **Branching logic** for conditional question flows
- ✅ **Context preservation** across multi-step surveys

Use this Skill whenever:
- User intent is ambiguous
- Multiple valid approaches exist
- Architectural decisions require trade-off analysis
- Approvals are needed before risky operations

---

_For more details, see SKILL.md and reference.md_
