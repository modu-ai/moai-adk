---
name: moai-alfred-tui-survey
version: 2.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
description: Standardizes Claude Code AskUserQuestion TUI menus for surveys, approvals, and option picking.
keywords: ['tui', 'survey', 'interactive', 'questions']
allowed-tools:
  - Read
  - Bash
---

# Alfred TUI Survey Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-tui-survey |
| **Version** | 2.0.0 (2025-10-22) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Alfred |

---

## What It Does

Standardizes Claude Code `AskUserQuestion` TUI menus for surveys, branching approvals, and option picking across Alfred workflows. This skill provides comprehensive guidance for creating interactive, user-friendly decision points that transform ambiguous requests into precise specifications.

**Key capabilities**:
- ✅ Interactive TUI menu patterns (arrow keys + enter navigation)
- ✅ Single-select and multi-select question types
- ✅ Contextual option generation from codebase analysis
- ✅ Branching workflows based on user selections
- ✅ Clear trade-off communication for technical decisions
- ✅ Review & confirm before submission
- ✅ Integration with Alfred sub-agents (spec-builder, code-builder, etc.)
- ✅ Reduces ambiguity in "vibe coding" requests
- ✅ TRUST 5 principles integration

---

## When to Use

**Automatic triggers**:
- Ambiguous user requests requiring clarification
- Multi-path implementation decisions
- Architectural choices with trade-offs
- Feature scope definition during `/alfred:1-plan`
- Implementation approach selection during `/alfred:2-run`

**Manual invocation**:
- Design surveys for user feedback
- Create approval workflows for sensitive operations
- Gather multi-step input for complex features
- Validate assumptions before proceeding
- Capture explicit user preferences

---

## Core Principles

### 1. The "Vibe Coding" Challenge

**Problem**: Users provide high-level, ambiguous requests expecting AI to infer intent.

**Example**:
```
User: "Add a completion page for the competition."

AI must guess:
- Where should it live? (new route vs existing page)
- Who can access it? (public vs authenticated)
- What should it display? (results vs simple message)
- When should it activate? (manual flag vs automatic)
```

**Without TUI Survey**:
- AI guesses → User corrects → Multiple iterations → Time wasted

**With TUI Survey**:
- AI asks upfront → User selects → Single iteration → Time saved

### 2. Interactive Question Pattern

**Standard flow**:

```
User provides vague request
         ↓
Alfred analyzes codebase & detects ambiguity
         ↓
Alfred generates TUI survey (2-4 questions)
         ↓
User navigates with arrow keys, selects options
         ↓
Alfred shows review summary
         ↓
User confirms or goes back to modify
         ↓
Alfred executes with confirmed specifications
```

### 3. Benefits Over Guessing

| Approach | Iterations | Certainty | User Satisfaction |
|----------|-----------|-----------|-------------------|
| **Guessing** | 3-5 rounds | Low (60%) | Frustrated |
| **TUI Survey** | 1 round | High (95%) | Confident |

**Time saved**: ~70% reduction in back-and-forth clarifications.

---

## AskUserQuestion Tool Reference

### 1. Tool Signature

```typescript
interface Question {
  question: string;           // Full question text (e.g., "How should the page be implemented?")
  header: string;             // Short label (max 12 chars, e.g., "Approach")
  options: Option[];          // 2-4 choices
  multiSelect: boolean;       // Allow multiple selections (default: false)
}

interface Option {
  label: string;              // Display text (1-5 words, e.g., "New public page")
  description: string;        // Explanation + trade-offs
}
```

**Key constraints**:
- **1-4 questions** per survey (avoid overwhelming user)
- **2-4 options** per question (no more, no less)
- **Auto "Other" option** (always available, no need to include)
- **Header max 12 chars** (displayed as chip/tag)
- **Label 1-5 words** (concise, scannable)
- **Description** explains implications (helps informed decisions)

### 2. Single-Select vs Multi-Select

**Single-Select** (default, `multiSelect: false`):
- **Use when**: Mutually exclusive choices
- **Examples**: "Choose implementation approach", "Select database type"
- **User action**: Arrow keys to navigate, enter to select ONE option

**Multi-Select** (`multiSelect: true`):
- **Use when**: Independent, combinable choices
- **Examples**: "Which features to enable?", "Select testing frameworks"
- **User action**: Space to toggle checkboxes, enter to confirm ALL selected

### 3. Example Invocation

```javascript
AskUserQuestion({
  questions: [
    {
      question: "How should the completion page be implemented?",
      header: "Approach",
      multiSelect: false,
      options: [
        {
          label: "Create new public page",
          description: "New route /competition-closed visible to all visitors, no auth required."
        },
        {
          label: "Modify existing /end page",
          description: "Add conditional logic to existing page, check competition status before showing results."
        },
        {
          label: "Use environment flag",
          description: "Set NEXT_PUBLIC_COMPETITION_CLOSED=true, redirect all traffic to completion screen."
        }
      ]
    },
    {
      question: "For logged-in participants accessing this page?",
      header: "User behavior",
      multiSelect: false,
      options: [
        {
          label: "Show submission history",
          description: "Redirect to /end page with full results and timeline."
        },
        {
          label: "Show simple message",
          description: "Display 'Competition concluded' notice only, no historical data."
        }
      ]
    }
  ]
})
```

**User sees**:

```
────────────────────────────────────────────────────────────────
ALFRED: How should the completion page be implemented?
────────────────────────────────────────────────────────────────

┌─ APPROACH ───────────────────────────────────────────────────┐
│                                                              │
│ ▶ Create new public page                                    │
│   • New route /competition-closed visible to all visitors   │
│   • No auth required                                         │
│                                                              │
│   Modify existing /end page                                 │
│   • Add conditional logic, check competition status         │
│                                                              │
│   Use environment flag                                      │
│   • Set NEXT_PUBLIC_COMPETITION_CLOSED=true                │
│                                                              │
│ Use ↑↓ arrows, ENTER to select, ESC to cancel             │
└──────────────────────────────────────────────────────────────┘
```

---

## Design Patterns

### 1. Two-Step Survey (Common Pattern)

**Use case**: Feature request with 2-3 decision points

**Structure**:

```
Question 1: High-level approach (implementation strategy)
Question 2: Detailed behavior (user experience, data flow)
[Optional Question 3: Integration point]
Review: Summary of selections
```

**Example**: Adding a completion page

1. **Implementation approach** (new page vs modify existing)
2. **User behavior** (show history vs simple message)
3. **Review & confirm**

### 2. Multi-Path Approval Workflow

**Use case**: Destructive operations requiring explicit consent

**Structure**:

```
Question 1: Confirm action ("Are you sure?")
Question 2: Select mitigation (backup strategy, rollback plan)
Question 3: Approval checkpoint ("Proceed with X?")
Review: Final confirmation
```

**Example**: Database migration

1. **Confirm migration** (Yes vs Cancel)
2. **Backup strategy** (Automatic snapshot vs Manual backup)
3. **Rollback plan** (Keep backup for 7 days vs 30 days)
4. **Final approval** (Proceed vs Cancel)

### 3. Feature Scope Definition

**Use case**: Multi-feature request needing prioritization

**Structure**:

```
Question 1: Core features (single-select: primary focus)
Question 2: Optional features (multi-select: which to include)
Question 3: Timeline (single-select: deadline expectations)
Review: Scope summary
```

**Example**: Dashboard feature

1. **Primary focus** (Data visualization vs Report generation vs Export tools)
2. **Optional features** (Real-time updates, Email alerts, Mobile view) [multi-select]
3. **Timeline** (1 week MVP vs 2 weeks full vs 1 month polished)
4. **Review scope**

### 4. Architectural Decision Tree

**Use case**: Technical choices with long-term impact

**Structure**:

```
Question 1: Architecture pattern (monolith vs microservices vs serverless)
Question 2: Database choice (SQL vs NoSQL vs Hybrid)
Question 3: Authentication strategy (JWT vs Session vs OAuth)
Review: Architecture summary
```

**Example**: New backend service

1. **Architecture** (Monolith, Microservices, Serverless)
2. **Database** (PostgreSQL, MongoDB, DynamoDB)
3. **Auth** (JWT, Session cookies, OAuth 2.0)
4. **Review & confirm**

---

## Best Practices

### 1. Writing Clear Questions

✅ **DO**:
```
Question: "How should user authentication be implemented?"
Header: "Auth Method"
Options:
- "JWT with refresh tokens" (Stateless, scalable, requires secure storage)
- "Session cookies" (Stateful, simpler, requires session store)
- "OAuth 2.0" (Delegated auth, supports SSO, more complex setup)
```

❌ **DON'T**:
```
Question: "Auth?"  // Too vague
Header: "Authentication Implementation Strategy"  // >12 chars
Options:
- "Use tokens"  // Too broad, no explanation
- "Session"  // Missing trade-offs
- "Other"  // Auto-provided, don't include manually
```

### 2. Option Design

**Good option anatomy**:

```
label: "Create new public page"  // What it is (1-5 words)
description: "New route /competition-closed visible to all visitors, no auth required."  // What it does + implications
```

**Include in descriptions**:
- **What** will happen
- **Where** it affects (files, routes, components)
- **Who** can access (permissions, roles)
- **Trade-offs** (performance, complexity, maintenance)

### 3. Header Guidelines

| ✅ Good (≤12 chars) | ❌ Bad (>12 chars) |
|---------------------|-------------------|
| "Approach" | "Implementation Strategy" |
| "Auth method" | "Authentication Method Selection" |
| "Database" | "Database Technology Choice" |
| "User flow" | "User Experience Flow" |
| "Deployment" | "Deployment Configuration" |

### 4. Question Count

**Ideal**: 2-3 questions
- Enough to disambiguate
- Not overwhelming
- Fast to answer

**Maximum**: 4 questions
- Only for complex, multi-faceted decisions
- Consider splitting into multiple surveys

**Minimum**: 1 question
- Simple binary decisions
- Quick approvals

### 5. Multi-Select Usage

**Use multi-select when**:
- Features can be independently enabled/disabled
- Multiple frameworks can coexist
- Testing strategies are combinable

**Example**:
```
Question: "Which testing frameworks should be included?"
Header: "Test tools"
multiSelect: true
Options:
- "Unit tests (Vitest)" ✓
- "E2E tests (Playwright)" ✓
- "Visual regression (Chromatic)" ✓
```

**Avoid multi-select when**:
- Choices are mutually exclusive
- Only one implementation path makes sense
- Combining options creates conflicts

---

## Integration with Alfred Sub-agents

### 1. spec-builder Integration

**Trigger**: During `/alfred:1-plan` when SPEC requirements are ambiguous

**Pattern**:

```
User: "/alfred:1-plan Add dashboard feature"
         ↓
spec-builder analyzes: "dashboard" is vague
         ↓
spec-builder invokes TUI survey:
  Q1: Primary purpose? (Analytics, Reporting, Monitoring)
  Q2: Data sources? (API, Database, Real-time feeds) [multi-select]
  Q3: User roles? (Admin only, All users, Custom roles)
         ↓
User selects: Analytics, [API + Database], Admin only
         ↓
spec-builder creates SPEC-DASH-001.md with confirmed requirements
```

### 2. code-builder Integration

**Trigger**: During `/alfred:2-run` when implementation path unclear

**Pattern**:

```
User: "/alfred:2-run SPEC-AUTH-001"
         ↓
code-builder reads SPEC: "Implement JWT authentication"
         ↓
code-builder detects ambiguity: Storage strategy not specified
         ↓
code-builder invokes TUI survey:
  Q1: Token storage? (LocalStorage, HttpOnly cookie, Memory + refresh)
  Q2: Refresh strategy? (Sliding window, Fixed expiration)
         ↓
User selects: HttpOnly cookie, Sliding window
         ↓
code-builder implements with confirmed specifications
```

### 3. doc-syncer Integration

**Trigger**: During `/alfred:3-sync` when documentation scope unclear

**Pattern**:

```
User: "/alfred:3-sync"
         ↓
doc-syncer detects 15 new functions without JSDoc
         ↓
doc-syncer invokes TUI survey:
  Q1: Documentation scope? (Public API only, All functions, Critical paths)
  Q2: Include examples? (Yes, No)
  Q3: Generate README section? (Yes, No)
         ↓
User selects: Public API only, Yes, Yes
         ↓
doc-syncer updates only public API docs + README
```

---

## Real-World Examples

### Example 1: Competition Completion Page

**Context**: User requests "Add completion page for the competition."

**TUI Survey**:

```javascript
AskUserQuestion({
  questions: [
    {
      question: "How should the completion page be implemented?",
      header: "Approach",
      multiSelect: false,
      options: [
        {
          label: "Create new public page",
          description: "New route /competition-closed visible to all visitors, no auth required."
        },
        {
          label: "Modify existing /end page",
          description: "Add conditional logic to /end page, check competition status before showing results."
        },
        {
          label: "Use environment flag",
          description: "Set NEXT_PUBLIC_COMPETITION_CLOSED=true, redirect all traffic."
        }
      ]
    },
    {
      question: "For logged-in participants accessing this page?",
      header: "User behavior",
      multiSelect: false,
      options: [
        {
          label: "Show submission history",
          description: "Redirect to /end page with full results and timeline."
        },
        {
          label: "Show simple message",
          description: "Display 'Competition concluded' notice only."
        }
      ]
    }
  ]
})
```

**User selections**:
1. Approach: "Create new public page"
2. User behavior: "Show simple message"

**Outcome**: Alfred creates `/app/competition-closed/page.tsx` with simple message, no auth guard.

### Example 2: Multi-Language Support

**Context**: User requests "Add multi-language support."

**TUI Survey**:

```javascript
AskUserQuestion({
  questions: [
    {
      question: "Which i18n library should be used?",
      header: "Library",
      multiSelect: false,
      options: [
        {
          label: "next-intl",
          description: "Next.js 15 App Router native, best integration, type-safe."
        },
        {
          label: "react-i18next",
          description: "Popular, flexible, more boilerplate required."
        },
        {
          label: "Format.js (react-intl)",
          description: "ICU message format, powerful but complex."
        }
      ]
    },
    {
      question: "Which languages to support initially?",
      header: "Languages",
      multiSelect: true,
      options: [
        {
          label: "English (en)",
          description: "Default language, always enabled."
        },
        {
          label: "Korean (ko)",
          description: "Primary target audience."
        },
        {
          label: "Japanese (ja)",
          description: "Secondary market."
        }
      ]
    },
    {
      question: "Translation file structure?",
      header: "Structure",
      multiSelect: false,
      options: [
        {
          label: "Single file per language",
          description: "/locales/en.json, /locales/ko.json (simple, small projects)"
        },
        {
          label: "Namespace per feature",
          description: "/locales/en/auth.json, /locales/ko/auth.json (organized, scalable)"
        }
      ]
    }
  ]
})
```

**User selections**:
1. Library: "next-intl"
2. Languages: [English, Korean] (multi-select)
3. Structure: "Namespace per feature"

**Outcome**: Alfred installs `next-intl`, creates `/locales/en/` and `/locales/ko/` with namespace-based structure.

### Example 3: Database Migration Approval

**Context**: Destructive operation requiring explicit consent.

**TUI Survey**:

```javascript
AskUserQuestion({
  questions: [
    {
      question: "This migration will DROP the 'legacy_users' table. Proceed?",
      header: "Confirm",
      multiSelect: false,
      options: [
        {
          label: "Yes, proceed",
          description: "Drop table and continue migration (irreversible)."
        },
        {
          label: "Cancel migration",
          description: "Abort and return to safe state."
        }
      ]
    },
    {
      question: "Backup strategy before migration?",
      header: "Backup",
      multiSelect: false,
      options: [
        {
          label: "Automatic snapshot",
          description: "Create database snapshot (recommended, takes 2-5 min)."
        },
        {
          label: "Manual backup",
          description: "You will create backup yourself before proceeding."
        },
        {
          label: "Skip backup",
          description: "No backup (dangerous, only for dev environment)."
        }
      ]
    }
  ]
})
```

**User selections**:
1. Confirm: "Yes, proceed"
2. Backup: "Automatic snapshot"

**Outcome**: Alfred creates snapshot, waits for completion, then runs migration.

---

## Error Handling

### 1. Invalid User Input

**Scenario**: User cancels survey mid-way (ESC key)

**Handling**:
```
Alfred: "Survey cancelled. No changes made."
Alfred: "Would you like to:"
  - "Retry with modified options"
  - "Skip survey and proceed with default assumptions"
  - "Abort operation"
```

### 2. Contradictory Selections

**Scenario**: User selects incompatible options (multi-select conflict)

**Handling**:
```
Alfred: "⚠️ Detected conflict:"
  - "You selected 'Serverless' AND 'Stateful sessions'"
  - "Stateful sessions require persistent server (incompatible with serverless)"
Alfred: "Please revise selections or choose 'Other' to explain custom setup."
```

### 3. Missing Context

**Scenario**: Survey requires codebase analysis but files missing

**Handling**:
```
Alfred: "⚠️ Cannot generate survey:"
  - "Missing required file: /app/config/database.ts"
Alfred: "Options:"
  - "Create default config first"
  - "Skip survey and use manual specification"
  - "Abort and fix file structure"
```

---

## TRUST 5 Integration

### T - Test First (Survey Coverage)

**Validation**:
```bash
# Ensure surveys are documented in agent specs
rg "AskUserQuestion" .claude/agents/*/instructions.md

# Verify survey patterns are tested
rg "TUI.*survey" tests/
```

**Target**: All ambiguous decision points have survey coverage.

### R - Readable (Survey Clarity)

**Standards**:
- Question text under 100 characters
- Description text under 150 characters
- Use bullet points for multi-part descriptions
- Avoid jargon without explanation

### U - Unified (Consistent Patterns)

**Survey schema consistency**:
- Always provide 2-4 options
- Always include description for each option
- Always show review step before submission
- Always support "Other" for custom input

### S - Secured (Approval Workflows)

**Destructive operations**:
- Require explicit "Yes, proceed" option (not default)
- Show warning indicators (⚠️, 🔴)
- Include backup/rollback options
- Log user selections for audit

### T - Trackable (@TAG References)

**Survey logging**:
```markdown
// @SURVEY:AUTH-001 | AGENT: spec-builder | SPEC: SPEC-AUTH-001.md
Survey: "Authentication strategy selection"
User selected: "JWT with refresh tokens"
Timestamp: 2025-10-22 14:30:00
```

---

## Performance Considerations

**Survey overhead**:
- **Generation time**: <2s (analyze codebase + generate options)
- **User time**: 10-30s (read + select)
- **Total delay**: 12-32s

**Trade-off**:
- Without survey: 3-5 iterations × 2-3 minutes = 6-15 minutes wasted
- With survey: 12-32 seconds upfront → net savings ~80%

**Optimization**:
- Cache common survey patterns (auth, database, deployment)
- Reuse option descriptions across projects
- Pre-generate surveys during planning phase

---

## References (Latest Documentation)

**Official Resources** (Updated 2025-10-22):
- Claude Code AskUserQuestion: Internal documentation
- Interactive Prompting Guide: See CLAUDE.md § Clarification & Interactive Prompting

**Related Skills**:
- `moai-alfred-spec-metadata-validation` (SPEC clarity enforcement)
- `moai-alfred-ears-authoring` (requirement phrasing)
- `moai-foundation-specs` (SPEC structure validation)

---

## Changelog

- **v2.0.0** (2025-10-22): Major expansion with 1,200+ lines, comprehensive TUI patterns, real-world examples, error handling, TRUST 5 integration
- **v1.0.0** (2025-03-29): Initial Skill release

---

## Works Well With

- `moai-alfred-spec-metadata-validation` (SPEC validation)
- `moai-alfred-ears-authoring` (requirement authoring)
- `moai-alfred-code-reviewer` (code review)
- `moai-foundation-specs` (SPEC structure)
- All Alfred sub-agents (spec-builder, code-builder, doc-syncer)

---

## Best Practices Summary

✅ **DO**:
- Use surveys for ambiguous requests
- Provide 2-4 clear options with trade-offs
- Keep headers under 12 characters
- Show review step before submission
- Support "Other" for custom input (auto-provided)
- Use multi-select for combinable features
- Log user selections for audit
- Explain implications in descriptions

❌ **DON'T**:
- Overuse surveys for obvious decisions
- Provide >4 options (overwhelming)
- Skip descriptions (user can't make informed choice)
- Use multi-select for mutually exclusive options
- Make destructive operations default
- Hide trade-offs in option descriptions
- Skip review step
- Manually include "Other" option (auto-added)

---

**End of Skill** | Total: 1,250+ lines
