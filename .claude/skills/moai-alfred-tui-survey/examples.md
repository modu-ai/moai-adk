# Alfred TUI Survey Examples

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

## Multi-select variation

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

## Follow-up prompt for deeper detail

```typescript
if (selection.includes("Integration tests")) {
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
