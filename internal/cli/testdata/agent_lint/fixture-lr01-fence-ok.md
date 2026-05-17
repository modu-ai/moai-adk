---
name: test-lr01-fence-ok-agent
description: Agent with AskUserQuestion only inside fenced code blocks
tools: Read, Write, Edit
effort: high
---

# LR-01 Fence-OK Agent

This agent demonstrates that AskUserQuestion inside fenced code blocks is exempt.

## Usage Example

```
// This is how the orchestrator calls AskUserQuestion:
AskUserQuestion({
  questions: [{
    question: "What would you like to do?",
    options: [
      { label: "Option A (권장)", description: "First choice" },
      { label: "Option B", description: "Second choice" }
    ]
  }]
})
```

The above is documentation only — this agent does not call AskUserQuestion.
