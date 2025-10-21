---
name: moai-alfred-tui-survey
description: "Standardizes Claude Code Tools AskUserQuestion TUI menus for surveys, branching approvals, and option picking across Alfred workflows."
allowed-tools:
  - Read
  - Write
  - Edit
  - TodoWrite
---

# Alfred TUI Survey Skill

## What it does

Provides ready-to-use patterns for Claude Code's AskUserQuestion TUI selector so Alfred agents can gather user choices, approvals, or survey answers with structured menus instead of ad-hoc text prompts.

## When to use

- Need confirmation before advancing to a risky/destructive step.
- Choosing between alternative implementation paths or automation levels.
- Collecting survey-like answers (persona, tech stack, priority, risk level).
- Any time a branched workflow depends on user-selected options rather than free-form text.

## How it works

1. **Detect decision points** – Identify moments where the agent must pause and let the user choose.
2. **Prepare option set** – Offer 2–5 clearly distinct options with concise labels and rich descriptions.
3. **Render AskUserQuestion** – Output a `AskUserQuestion({...})` block so Claude Code shows the interactive TUI selector.
4. **Map follow-up actions** – Document how each option changes the next action (branch, agent hand-off, command retry, etc.).
5. **Confirm / fallback** – If the user skips the menu or the tool is unavailable, gracefully fall back to text clarification.

## Single-select template

```typescript
AskUserQuestion({
  questions: [{
    header: "Decision point: Deployment Strategy",
    question: "How should we roll out the new release?",
    options: [
      {
        label: "Canary release",
        description: "Gradually roll out to a small user segment; monitor metrics first."
      },
      {
        label: "Blue/Green",
        description: "Keep the current version live while preparing the new stack in parallel."
      },
      {
        label: "Full deploy",
        description: "Immediate production rollout after smoke tests succeed."
      }
    ],
    multiSelect: false
  }]
})
```

### Multi-select variation

Use when the user may choose more than one option (checklist style):

```typescript
AskUserQuestion({
  questions: [{
    header: "Select diagnostics to run",
    question: "Which checks should run before proceeding?",
    options: [
      { label: "Unit tests", description: "Fast verification for core modules." },
      { label: "Integration tests", description: "Service-level interactions and DB calls." },
      { label: "Security scan", description: "Dependency vulnerability audit." }
    ],
    multiSelect: true
  }]
})
```

### Free-form follow-up

After the menu submission, summarize the choice(s) and request clarifying details if necessary:

```typescript
// Pseudocode follow-up
if (selection.includes("Integration tests")) {
  // AskUserQuestion again if more detail needed
  AskUserQuestion({
    questions: [{
      header: "Integration test scope",
      question: "Which environment should host integration tests?",
      options: [
        { label: "Staging", description: "Use the shared staging cluster." },
        { label: "Ephemeral env", description: "Provision a one-off test environment." }
      ],
      multiSelect: false
    }]
  })
}
```

## Operational checklist

- [ ] Always pair AskUserQuestion menus with this skill when a decision gate appears.
- [ ] Keep option labels short (≤40 chars) and descriptions actionable.
- [ ] Explicitly note downstream behavior for each option in the surrounding narrative.
- [ ] Prefer TUI menus over open-ended text for survey questions, priority rankings, or risk acknowledgements.
- [ ] Provide manual fallback instructions if the UI cannot render (e.g., "Reply with the option name").

## Works well with

- `moai-foundation-ears` – Combine structured requirement patterns with menu-driven confirmations.
- `moai-alfred-git-workflow` – Use menus to choose branch/worktree strategies.
- `moai-alfred-code-reviewer` – Capture reviewer focus areas through guided selection.
