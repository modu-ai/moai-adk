---
name: moai-alfred-tui-survey
description: Standardizes Claude Code AskUserQuestion TUI menus for surveys, approvals, and option picking. Use when gathering user decisions, clarifying ambiguous requirements, or presenting implementation choices through interactive terminal menus.
allowed-tools:
  - AskUserQuestion
  - Read
  - Bash
version: 2.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
keywords: ['tui', 'survey', 'interactive', 'questions', 'menus', 'user-input', 'clarification', 'approval']
---

# Alfred TUI Survey Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-tui-survey |
| **Version** | 2.0.0 (2025-10-22) |
| **Allowed tools** | AskUserQuestion (interactive surveys), Read (context), Bash (validation) |
| **Auto-load** | On demand when ambiguity detected or approval needed |
| **Tier** | Alfred |

---

## What It Does

The **moai-alfred-tui-survey** Skill standardizes interactive user engagement through Claude Code's AskUserQuestion tool, transforming vague requests into precise specifications through structured TUI (Terminal User Interface) surveys.

**Core Capabilities**:
- Interactive clarification of ambiguous requirements
- Structured decision-making for architectural choices
- User approval workflows for critical operations
- Multi-option selection menus with clear trade-offs
- Context-aware question generation based on codebase analysis
- Progressive disclosure of implementation options

**Key Innovation**: Instead of AI guessing user intent from incomplete instructions ("Vibe Coding"), this Skill enables AI to present concrete options with clear trade-offs via terminal UI menus, capturing explicit user choices before execution.

---

## When to Use

### Automatic Triggers

This Skill activates automatically when:
- **Ambiguity detected**: Multiple valid implementation approaches exist
- **Architectural decisions**: Choice affects system design or data flow
- **Breaking changes**: Operation may impact existing functionality
- **User experience decisions**: Multiple UX patterns are viable
- **Security/compliance**: Explicit approval required before proceeding

### Manual Invocation

Invoke explicitly when:
- Planning complex features with trade-offs
- Refactoring with multiple migration strategies
- Designing APIs with competing patterns
- Selecting libraries or frameworks
- Clarifying high-level requirements like "Add a dashboard" or "Optimize performance"

### Example Scenarios

| User Request | Why TUI Survey Needed | Clarification Required |
|--------------|----------------------|------------------------|
| "Add a dashboard" | Multiple layout patterns exist | Layout type, data sources, authentication |
| "Optimize performance" | Many bottlenecks possible | Target metric, acceptable trade-offs, profiling approach |
| "Refactor auth system" | Multiple migration paths | Scope, backwards compatibility, deployment strategy |
| "Add multi-language support" | Implementation varies widely | Scope (UI only vs full i18n), default language, library choice |

---

## Core Principles

### 1. Reduce Ambiguity, Don't Guess

**Problem**: Traditional "Vibe Coding" forces AI to infer intent from incomplete instructions, leading to:
- Ambiguous implementations
- Unnecessary code modifications
- Multiple refinement rounds
- Wasted developer time

**Solution**: Present concrete options upfront, capture explicit choices, execute with certainty.

### 2. Context-Aware Question Generation

Survey questions are derived from:
- Existing codebase patterns (`Read` tool analysis)
- Current project structure (`Glob` for similar features)
- Domain best practices (skill knowledge)
- SPEC and TAG references (traceability context)

### 3. Progressive Clarification

Start with high-level choices, drill down based on selections:
```
Question 1: Implementation approach (architectural decision)
    ↓
Question 2: User experience details (UX decision)
    ↓
Question 3: Data handling strategy (technical decision)
    ↓
Review: Summary of all selections with "Go back" option
```

### 4. Always Provide Review Step

Before executing, show summary:
- All selected options
- Implications of choices
- Option to revise any answer
- Clear "Submit" vs "Go back" actions

---

## TUI Survey Patterns

### Pattern 1: Architectural Decision Survey

**Use when**: Multiple system design approaches are viable

**Structure**:
```
Question: "How should the [feature] be implemented?"
Options:
  - Approach A (description + trade-offs)
  - Approach B (description + trade-offs)
  - Approach C (description + trade-offs)
  - [Other] (custom input)
```

**Example**:
```
Question: "How should the completion page be implemented?"
Options:
  1. Create new public page (/competition-closed)
     • Unguarded route, visible to all visitors
     • No authentication required
     • PRO: Simple, clear separation | CON: Additional route

  2. Modify existing /end page with conditional logic
     • Check if competition is active before showing results
     • PRO: Reuses existing code | CON: Adds complexity

  3. Use environment-based gating
     • Set NEXT_PUBLIC_COMPETITION_CLOSED=true
     • PRO: Runtime configuration | CON: Requires env management
```

### Pattern 2: Feature Scope Survey

**Use when**: Feature scope is unclear or has multiple valid extents

**Structure**:
```
Question: "What scope should [feature] cover?"
Options:
  - Minimal (core functionality only)
  - Standard (common use cases)
  - Comprehensive (full feature set)
  - [Other] (custom scope definition)
```

**Example**:
```
Question: "What scope should multi-language support cover?"
Options:
  1. UI labels only (buttons, menus, navigation)
     • Scope: Frontend i18n | Effort: 2-3 days

  2. UI + static content (about pages, help docs)
     • Scope: Frontend + CMS integration | Effort: 5-7 days

  3. Full i18n (UI, content, backend messages, emails)
     • Scope: Full-stack internationalization | Effort: 10-14 days
```

### Pattern 3: Technical Choice Survey

**Use when**: Multiple libraries, frameworks, or tools can solve the problem

**Structure**:
```
Question: "Which [tool/library] should be used for [purpose]?"
Options:
  - Option A (maturity, ecosystem, trade-offs)
  - Option B (maturity, ecosystem, trade-offs)
  - Option C (maturity, ecosystem, trade-offs)
  - [Other] (specify alternative)
```

**Example**:
```
Question: "Which state management library should be used?"
Options:
  1. Redux Toolkit (v2.x)
     • Maturity: Industry standard | Ecosystem: Extensive
     • PRO: DevTools, middleware | CON: Boilerplate

  2. Zustand (v5.x)
     • Maturity: Stable, modern | Ecosystem: Growing
     • PRO: Simple API, minimal boilerplate | CON: Fewer resources

  3. Jotai (v2.x)
     • Maturity: Modern atomic approach | Ecosystem: Niche
     • PRO: Fine-grained reactivity | CON: Paradigm shift
```

### Pattern 4: Migration Strategy Survey

**Use when**: Refactoring or upgrading with multiple migration paths

**Structure**:
```
Question: "What migration strategy should be used?"
Options:
  - Big bang (full rewrite, single deployment)
  - Incremental (gradual migration, feature flags)
  - Parallel run (new + old systems coexist)
  - [Other] (custom strategy)
```

**Example**:
```
Question: "How should the auth system be refactored?"
Options:
  1. Big bang migration (full rewrite)
     • Timeline: 2-3 weeks | Risk: High
     • PRO: Clean slate | CON: Risky deployment

  2. Incremental with feature flags
     • Timeline: 4-6 weeks | Risk: Medium
     • PRO: Gradual rollout | CON: Code duplication

  3. Parallel run (adapter pattern)
     • Timeline: 6-8 weeks | Risk: Low
     • PRO: Safe rollback | CON: Complex maintenance
```

### Pattern 5: Approval Workflow Survey

**Use when**: Explicit approval needed before executing

**Structure**:
```
Question: "Review the proposed changes. Proceed?"
Options:
  - Approve and execute
  - Approve with modifications (specify)
  - Reject and reconsider
```

**Example**:
```
Question: "This will delete 50 unused files. Proceed?"
Details:
  - 50 files identified as unused (last modified >6 months ago)
  - Backup created at .moai-backups/pre-cleanup-2025-10-22
  - Reversible via Git checkout

Options:
  1. Approve and execute cleanup
  2. Review file list first (show details)
  3. Cancel operation
```

---

## Question Design Best Practices

### Structure Each Question

```yaml
question: "Clear, specific question ending with ?"
header: "Short label (max 12 chars)"  # e.g., "Auth method", "Library"
multiSelect: false  # or true for checkboxes
options:
  - label: "Option 1 name"
    description: "What this option means and its implications"
  - label: "Option 2 name"
    description: "What this option means and its implications"
```

### Writing Effective Questions

**DO**:
- ✅ Be specific: "How should the completion page be implemented?"
- ✅ Provide context: Include current state, constraints, implications
- ✅ Show trade-offs: PRO/CON, effort estimates, risk levels
- ✅ Limit options: 2-4 main choices (avoid analysis paralysis)
- ✅ Include "Other" option: Allow custom input
- ✅ Use present tense: "How should...", "Which...", "What..."

**DON'T**:
- ❌ Be vague: "What do you want?" (too open-ended)
- ❌ Use jargon without explanation: Assume user knowledge level
- ❌ Omit implications: Show what each choice affects
- ❌ Overload options: >4 options cause decision fatigue
- ❌ Force choices: Always provide escape hatch
- ❌ Use past tense: "How did you want...?" (sounds passive)

### Option Design Patterns

**Good Option Format**:
```
Option Label (concise, 3-7 words)
  • Description of what this option does
  • PRO: Key advantage | CON: Key limitation
  • Effort: Time estimate | Risk: High/Medium/Low
```

**Example**:
```
Create new microservice
  • Separate service with own database and API
  • PRO: Full isolation, independent scaling
  • CON: Operational overhead, network latency
  • Effort: 2-3 weeks | Risk: Medium
```

### Multi-Select vs Single-Select

**Single-Select** (default):
- Use for mutually exclusive choices
- Architecture patterns (can't be both monolith AND microservices)
- Migration strategies (pick one path)
- Primary library selection

**Multi-Select** (`multiSelect: true`):
- Use for independent, combinable options
- Feature toggles (enable multiple features)
- Validation rules (apply multiple checks)
- Testing strategies (unit + integration + e2e)

---

## Integration with MoAI Workflow

### Phase 1: Plan (`/alfred:1-plan`)

**Use TUI Survey for**:
- SPEC scope clarification ("What should this feature include?")
- EARS pattern selection ("Which requirement type fits?")
- TAG domain assignment ("Which domain does this belong to?")

**Example**:
```
User: "/alfred:1-plan Add notification system"
  ↓
TUI Survey: "What type of notifications?"
  - In-app only
  - Email notifications
  - Push notifications (mobile)
  - All of the above
  ↓
TUI Survey: "Real-time or batch?"
  - Real-time (WebSocket/SSE)
  - Batch (scheduled jobs)
  - Hybrid approach
```

### Phase 2: Run (`/alfred:2-run`)

**Use TUI Survey for**:
- Library/framework selection
- Implementation approach (TDD RED → GREEN → REFACTOR)
- Error handling strategy
- Data validation approach

**Example**:
```
User: "/alfred:2-run SPEC-001"
  ↓
TUI Survey: "Which validation library?"
  - Zod (TypeScript-first)
  - Yup (JSON schema)
  - Joi (Node.js standard)
  ↓
Execute TDD cycle with selected library
```

### Phase 3: Sync (`/alfred:3-sync`)

**Use TUI Survey for**:
- Documentation language/format
- Changelog structure
- PR title/description style
- Merge strategy

**Example**:
```
User: "/alfred:3-sync"
  ↓
TUI Survey: "PR is ready. Merge strategy?"
  - Squash and merge (clean history)
  - Merge commit (preserve commits)
  - Rebase and merge (linear history)
  ↓
Execute Git operations with selected strategy
```

---

## AskUserQuestion Tool Specification

### Tool Parameters

```typescript
{
  questions: [
    {
      question: string,          // The question to ask
      header: string,            // Short label (max 12 chars)
      multiSelect: boolean,      // Allow multiple selections?
      options: [
        {
          label: string,         // Option display text
          description: string    // Explanation of option
        }
      ]
    }
  ]
}
```

### Response Format

```typescript
{
  answers: {
    [questionIndex: string]: string | string[]
  }
}
```

### Example Invocation

```typescript
AskUserQuestion({
  questions: [
    {
      question: "How should the API authentication be implemented?",
      header: "Auth method",
      multiSelect: false,
      options: [
        {
          label: "JWT with refresh tokens",
          description: "Stateless auth with short-lived access tokens and long-lived refresh tokens"
        },
        {
          label: "Session-based with cookies",
          description: "Server-side sessions stored in Redis, client receives httpOnly cookie"
        },
        {
          label: "OAuth2 with third-party provider",
          description: "Delegate authentication to Google/GitHub/etc., receive tokens"
        }
      ]
    },
    {
      question: "Which security features should be enabled?",
      header: "Security",
      multiSelect: true,
      options: [
        {
          label: "Rate limiting",
          description: "Prevent abuse with request rate limits per user/IP"
        },
        {
          label: "CSRF protection",
          description: "Protect against cross-site request forgery attacks"
        },
        {
          label: "XSS prevention headers",
          description: "Set Content-Security-Policy and X-Frame-Options headers"
        }
      ]
    }
  ]
})
```

---

## Failure Modes & Mitigation

### Failure Mode 1: User Selects "Other" Without Details

**Symptom**: User chooses "Other" but provides vague custom input

**Mitigation**:
- Follow up with clarifying questions
- Provide examples of what details are needed
- Show similar patterns from codebase for reference

**Example**:
```
User selects: "Other: Something with webhooks"
  ↓
Follow-up TUI Survey:
  "Could you clarify the webhook approach?"
  - Outgoing webhooks (send events to external services)
  - Incoming webhooks (receive events from external services)
  - Bidirectional webhook system
```

### Failure Mode 2: Too Many Nested Questions

**Symptom**: Survey becomes overwhelming with 5+ sequential questions

**Mitigation**:
- Limit to 3-4 questions per survey
- Group related choices into single multi-select
- Split complex decisions into multiple phases

**Example**:
```
❌ BAD: 7 sequential questions about feature details

✅ GOOD:
  Phase 1: High-level approach (1-2 questions)
  → Execute initial setup
  Phase 2: Detailed configuration (1-2 questions)
  → Finalize implementation
```

### Failure Mode 3: Options Not Mutually Exclusive

**Symptom**: User wants to combine options marked as single-select

**Mitigation**:
- Review if multi-select is more appropriate
- Add "Hybrid" or "Combination" option
- Explain why options are mutually exclusive in descriptions

**Example**:
```
Question: "Authentication strategy?"
  - JWT only
  - Session only
  - Hybrid (JWT for API, sessions for web) ← Add this option
```

### Failure Mode 4: Insufficient Context in Questions

**Symptom**: User can't make informed decision due to missing information

**Mitigation**:
- Always analyze codebase context before presenting survey (`Read` tool)
- Include current state in question description
- Show examples from existing code patterns

**Example**:
```
Question: "How should error handling be implemented?"
Context: "Current codebase uses try/catch blocks in 15 files,
          Result<T, E> pattern in 5 files (services layer)."
Options:
  - Expand Result pattern (align with existing services)
  - Standardize on try/catch (align with majority)
  - Introduce error boundary pattern (new approach)
```

---

## Best Practices Checklist

### Before Creating Survey

- [ ] Analyze codebase context (existing patterns, similar features)
- [ ] Identify specific ambiguity or decision point
- [ ] Verify that decision requires user input (not deterministic)
- [ ] Check if similar decisions have been made before (consistency)

### During Survey Design

- [ ] Limit to 3-4 questions maximum
- [ ] Each option has clear label + description
- [ ] Trade-offs (PRO/CON) are visible when relevant
- [ ] "Other" option is included for flexibility
- [ ] Multi-select is used only for independent choices
- [ ] Headers are concise (≤12 characters)

### After Capturing Responses

- [ ] Show review summary of all selections
- [ ] Provide "Go back" option to revise
- [ ] Confirm understanding with brief echo: "You selected [X], which means [Y]"
- [ ] Log selections for future reference (SPEC, docs)

---

## Examples

See [examples.md](examples.md) for complete real-world scenarios:
1. **Competition Completion Page** - Architectural decision with scope clarification
2. **State Management Library Selection** - Technical choice with trade-off analysis
3. **Auth System Refactor** - Migration strategy with risk assessment
4. **Feature Approval Workflow** - Explicit approval before destructive operation

---

## References

See [reference.md](reference.md) for:
- Complete AskUserQuestion API specification
- TUI interaction patterns from popular frameworks
- Decision tree templates for common scenarios
- Integration points with Alfred sub-agents

---

## Works Well With

- **moai-alfred-spec-metadata-validation**: Clarify SPEC scope during `/alfred:1-plan`
- **moai-alfred-ears-authoring**: Select EARS patterns interactively
- **moai-alfred-refactoring-coach**: Choose refactoring strategies
- **moai-foundation-specs**: Confirm SPEC metadata choices

---

## Changelog

- **v2.0.0** (2025-10-22): Complete rewrite with TUI survey patterns, question design best practices, codebase context integration, failure mode analysis, and extensive examples
- **v1.0.0** (2025-03-29): Initial Skill release with basic AskUserQuestion integration

---

**Version**: 2.0.0
**Last Updated**: 2025-10-22
**Status**: Production-ready
**Framework**: MoAI-ADK + Alfred workflow orchestration
