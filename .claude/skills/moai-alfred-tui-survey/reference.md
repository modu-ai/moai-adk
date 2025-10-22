# moai-alfred-tui-survey - Technical Reference

_Last updated: 2025-10-22_

## AskUserQuestion API Specification

### Function Signature

```typescript
AskUserQuestion(params: {
  questions: Question[]
}): Promise<Record<string, string | string[]>>
```

### Question Schema

```typescript
interface Question {
  question: string;          // The question text shown to user
  header: string;            // Short label (max 12 chars) for the TUI header
  multiSelect: boolean;      // Allow multiple selections (default: false)
  options: Option[];         // Array of 2-4 options
}

interface Option {
  label: string;             // Display text for the option (1-5 words recommended)
  description: string;       // Explanation of what this option means or does
}
```

### Return Value

```typescript
{
  [header: string]: string | string[]
}
```

- **Single-select**: Returns `{ "HeaderText": "Selected label" }`
- **Multi-select**: Returns `{ "HeaderText": ["Label1", "Label2", ...] }`

### Constraints

| Parameter | Constraint | Rationale |
|-----------|------------|-----------|
| `header` | Max 12 characters | TUI layout constraints |
| `question` | Clear, specific text | User understanding |
| `options` | 2-4 options per question | Avoid choice overload |
| `label` | 1-5 words | TUI readability |
| `description` | Concise (1-2 sentences) | Quick scanning |

---

## TUI Rendering Patterns

### Single-Select UI

```
┌─ HEADER TEXT ────────────────────────────────────────────────┐
│                                                              │
│ ▶ Option 1 Label                                             │
│   • Description for option 1.                               │
│                                                              │
│   Option 2 Label                                            │
│   • Description for option 2.                               │
│                                                              │
│   Option 3 Label                                            │
│   • Description for option 3.                               │
│                                                              │
│ Use ↑↓ arrows to navigate, ENTER to select                 │
└──────────────────────────────────────────────────────────────┘
```

**Indicators**:
- `▶` = Currently selected option
- `•` = Bullet point for description

### Multi-Select UI

```
┌─ HEADER TEXT (Multi-Select) ─────────────────────────────────┐
│                                                              │
│ [x] Option 1 Label                                          │
│     • Description for option 1.                             │
│                                                              │
│ [ ] Option 2 Label                                          │
│     • Description for option 2.                             │
│                                                              │
│ [x] Option 3 Label                                          │
│     • Description for option 3.                             │
│                                                              │
│ Use ↑↓ to navigate, SPACE to toggle, ENTER to submit       │
└──────────────────────────────────────────────────────────────┘
```

**Indicators**:
- `[x]` = Selected
- `[ ]` = Unselected
- SPACE bar toggles selection
- ENTER submits all selections

---

## Best Practices

### 1. Question Design

#### Do's ✅
- **Be specific**: "Which database migration strategy?" not "What should we do?"
- **Provide context**: Include relevant numbers, file names, or scope
- **Order by safety**: Safest/recommended option first
- **Flag risks**: Use "NOT RECOMMENDED" or "CAUTION" labels where appropriate
- **Explain trade-offs**: Mention time, resource, or complexity implications

#### Don'ts ❌
- **Vague questions**: "Proceed?" without context
- **Too many options**: 8+ options overwhelm users
- **Ambiguous labels**: "Option A" vs "Option B" without meaning
- **Hidden consequences**: Failing to mention downtime, data loss, etc.
- **Biased phrasing**: Leading questions that push toward one answer

### 2. Header Text Guidelines

| ✅ Good | ❌ Bad | Why |
|---------|--------|-----|
| `Framework` | `Framework Selection Wizard` | Too long (max 12 chars) |
| `DB Type` | `Database Type for Your Project` | Too wordy |
| `Sync Mode` | `Synchronization Configuration Mode` | Exceeds limit |
| `Auth` | `A` | Too cryptic |

### 3. Option Label Guidelines

| ✅ Good | ❌ Bad | Why |
|---------|--------|-----|
| `Full TDD (RED → GREEN → REFACTOR)` | `TDD` | Too cryptic for new users |
| `Blue-Green Deployment` | `Deployment Strategy #1` | Not self-documenting |
| `85% (recommended)` | `85%` | Missing recommendation cue |

### 4. Description Guidelines

**Format**: `<What it does> <Key implication>`

Examples:
- ✅ "Zero downtime, requires 2x resources temporarily. Safest option."
- ✅ "Gradual rollout, minimal resource overhead, 5-10 min downtime window."
- ❌ "This is a deployment strategy." (Too vague)
- ❌ "Uses blue-green pattern with environment swapping and load balancer configuration changes." (Too technical)

---

## Common Patterns

### Pattern 1: Yes/No Confirmation

```typescript
AskUserQuestion({
  questions: [{
    question: "This will [ACTION]. [SCOPE]. Continue?",
    header: "Confirm",
    multiSelect: false,
    options: [
      {
        label: "Yes, proceed",
        description: "[WHAT HAPPENS IF YES]"
      },
      {
        label: "No, cancel",
        description: "[WHAT HAPPENS IF NO]"
      }
    ]
  }]
})
```

**When to use**: Risky operations, destructive actions, irreversible changes.

### Pattern 2: Strategy Selection

```typescript
AskUserQuestion({
  questions: [{
    question: "How should we [ACTION] the [THING]?",
    header: "Strategy",
    multiSelect: false,
    options: [
      {
        label: "[APPROACH 1 NAME]",
        description: "[TRADE-OFFS]. [RECOMMENDATION]."
      },
      {
        label: "[APPROACH 2 NAME]",
        description: "[TRADE-OFFS]."
      },
      {
        label: "[APPROACH 3 NAME]",
        description: "[TRADE-OFFS]."
      }
    ]
  }]
})
```

**When to use**: Multiple valid implementation approaches, architectural decisions.

### Pattern 3: Feature Selection (Multi-Select)

```typescript
AskUserQuestion({
  questions: [{
    question: "Which [FEATURES] do you want to enable?",
    header: "Features",
    multiSelect: true,
    options: [
      {
        label: "[FEATURE 1]",
        description: "[WHAT IT DOES]. [DEPENDENCIES]."
      },
      {
        label: "[FEATURE 2]",
        description: "[WHAT IT DOES]. [DEPENDENCIES]."
      }
    ]
  }]
})
```

**When to use**: Optional add-ons, plugin selection, capability toggles.

### Pattern 4: Error Recovery

```typescript
AskUserQuestion({
  questions: [{
    question: "[N] [ERRORS] in [FILE]. How should we proceed?",
    header: "Error",
    multiSelect: false,
    options: [
      {
        label: "Debug immediately",
        description: "Stop and fix errors now."
      },
      {
        label: "Skip and continue",
        description: "Mark as @skip. NOT RECOMMENDED."
      },
      {
        label: "Rollback changes",
        description: "Revert to last passing commit."
      }
    ]
  }]
})
```

**When to use**: Test failures, build errors, deployment issues.

---

## Sequential Question Flow

### Simple Flow (2-3 Questions)

```typescript
// Question 1
const q1 = await AskUserQuestion({
  questions: [{
    question: "General question?",
    header: "Q1",
    options: [...]
  }]
});

// Question 2 (conditional based on Q1)
if (q1["Q1"] === "Yes") {
  const q2 = await AskUserQuestion({
    questions: [{
      question: "Follow-up question?",
      header: "Q2",
      options: [...]
    }]
  });
}

// Question 3 (always asked)
const q3 = await AskUserQuestion({
  questions: [{
    question: "Final question?",
    header: "Q3",
    options: [...]
  }]
});
```

### Batch Questions (1 Call, Multiple Questions)

```typescript
// Ask multiple related questions at once
const answers = await AskUserQuestion({
  questions: [
    {
      question: "Question 1?",
      header: "Q1",
      multiSelect: false,
      options: [...]
    },
    {
      question: "Question 2?",
      header: "Q2",
      multiSelect: false,
      options: [...]
    }
  ]
});

// Access answers
const answer1 = answers["Q1"];
const answer2 = answers["Q2"];
```

---

## Integration Points

### With /alfred:0-project

**Typical Questions**:
1. Documentation language (en/ko/ja)
2. Project mode (personal/team)
3. Optional features (auth, notifications, etc.)

**Example**:
```typescript
const langAnswer = await AskUserQuestion({
  questions: [{
    question: "Which language for documentation?",
    header: "Language",
    options: [
      { label: "English (en)", description: "All docs in English." },
      { label: "한국어 (ko)", description: "All docs in Korean." }
    ]
  }]
});

// Store in .moai/config.json
config.language = langAnswer["Language"];
```

### With /alfred:1-plan

**Typical Questions**:
1. SPEC scope confirmation
2. Domain prefix selection (AUTH, UI, DATA, etc.)
3. Git branch strategy

**Example**:
```typescript
const branchAnswer = await AskUserQuestion({
  questions: [{
    question: "Should we create a feature branch for SPEC-AUTH-001?",
    header: "Git Branch",
    options: [
      {
        label: "Yes, create feature/AUTH-001",
        description: "Isolate changes in dedicated branch."
      },
      {
        label: "No, work on current branch",
        description: "Continue on develop."
      }
    ]
  }]
});

if (branchAnswer["Git Branch"].startsWith("Yes")) {
  // Create branch
  await Bash("git checkout -b feature/AUTH-001");
}
```

### With /alfred:2-run

**Typical Questions**:
1. Implementation mode (TDD, Prototype, Test-After)
2. Which SPEC(s) to implement (if multiple pending)
3. Error recovery strategy (when tests fail)

**Example**:
```typescript
const modeAnswer = await AskUserQuestion({
  questions: [{
    question: "How should we implement SPEC-AUTH-001?",
    header: "Mode",
    options: [
      {
        label: "Full TDD (RED → GREEN → REFACTOR)",
        description: "Write failing tests first. Recommended for critical features."
      },
      {
        label: "Prototype First",
        description: "Quick POC, add tests after."
      }
    ]
  }]
});

const implementationStrategy = modeAnswer["Mode"];
// Proceed with selected strategy
```

### With /alfred:3-sync

**Typical Questions**:
1. Sync mode (auto/force/status/project)
2. PR Ready confirmation
3. Documentation scope (full/partial)

**Example**:
```typescript
const syncAnswer = await AskUserQuestion({
  questions: [{
    question: "Which synchronization mode?",
    header: "Sync Mode",
    options: [
      {
        label: "auto (recommended)",
        description: "Detect changes, sync affected docs only."
      },
      {
        label: "force",
        description: "Regenerate all docs from scratch."
      }
    ]
  }]
});

const mode = syncAnswer["Sync Mode"].split(" ")[0]; // Extract "auto" or "force"
// Run sync with selected mode
```

---

## Error Handling

### Invalid Input

```typescript
try {
  const answer = await AskUserQuestion({
    questions: [{
      question: "Select an option:",
      header: "Choice",
      options: [
        { label: "Option 1", description: "..." },
        { label: "Option 2", description: "..." }
      ]
    }]
  });
} catch (error) {
  console.error("User cancelled survey or invalid input:", error);
  // Fallback to default behavior
}
```

### User Cancellation (ESC Key)

When user presses ESC, the survey is cancelled and an error is thrown. Handle gracefully:

```typescript
const answer = await AskUserQuestion(...).catch(() => {
  console.log("User cancelled. Using default option.");
  return { "Header": "Default Option" };
});
```

### Custom Input Validation (Other Option)

```typescript
// User selects "Other" and provides custom input
const customValue = answers["Header"]; // May be a custom string

// Validate
if (!isValidValue(customValue)) {
  console.error(`Invalid custom value: ${customValue}`);
  // Ask again or use fallback
}
```

---

## Performance Considerations

### Minimize Question Rounds

❌ **Bad** (5 separate calls):
```typescript
const q1 = await AskUserQuestion(...);
const q2 = await AskUserQuestion(...);
const q3 = await AskUserQuestion(...);
const q4 = await AskUserQuestion(...);
const q5 = await AskUserQuestion(...);
```

✅ **Good** (1 batch call):
```typescript
const answers = await AskUserQuestion({
  questions: [
    { question: "Q1?", header: "Q1", options: [...] },
    { question: "Q2?", header: "Q2", options: [...] },
    { question: "Q3?", header: "Q3", options: [...] }
  ]
});
```

### Use Conditional Logic

Only ask follow-up questions when relevant:

```typescript
const q1 = await AskUserQuestion({
  questions: [{
    question: "Do you want authentication?",
    header: "Auth",
    options: [
      { label: "Yes", description: "Enable auth." },
      { label: "No", description: "Skip auth." }
    ]
  }]
});

// Only ask about auth provider if user said "Yes"
if (q1["Auth"] === "Yes") {
  const q2 = await AskUserQuestion({
    questions: [{
      question: "Which auth provider?",
      header: "Provider",
      options: [...]
    }]
  });
}
```

---

## Accessibility

### Clear Navigation Instructions

Always include keyboard instructions in TUI:
- "Use ↑↓ arrows to navigate, ENTER to select"
- "Use ↑↓ to navigate, SPACE to toggle, ENTER to submit" (multi-select)

### Option Descriptions

- **Required**: Every option MUST have a description
- **Purpose**: Helps users understand trade-offs without external documentation
- **Length**: 1-2 sentences max

### Visual Indicators

- **Single-select**: Use `▶` to indicate currently focused option
- **Multi-select**: Use `[x]` for selected, `[ ]` for unselected
- **Warnings**: Use "NOT RECOMMENDED" or "CAUTION:" prefix for risky options

---

## Testing Guidelines

### Unit Testing AskUserQuestion Calls

Mock the response:

```typescript
// Mock
jest.mock('AskUserQuestion', () => ({
  AskUserQuestion: jest.fn().mockResolvedValue({
    "Header": "Option 1"
  })
}));

// Test
const result = await someFunction();
expect(AskUserQuestion).toHaveBeenCalledWith({
  questions: [{
    question: expect.stringContaining("..."),
    header: "Header",
    options: expect.arrayContaining([...])
  }]
});
```

### Integration Testing

Test the full question flow:

```typescript
test('sequential question flow works correctly', async () => {
  // Simulate user answers
  const mockAnswers = [
    { "Q1": "Yes" },
    { "Q2": "Option A" }
  ];

  let callCount = 0;
  AskUserQuestion.mockImplementation(() => {
    return Promise.resolve(mockAnswers[callCount++]);
  });

  const result = await runWorkflow();

  expect(AskUserQuestion).toHaveBeenCalledTimes(2);
  expect(result.q1Answer).toBe("Yes");
  expect(result.q2Answer).toBe("Option A");
});
```

---

## Changelog

- **v2.0.0** (2025-10-22): Comprehensive reference update with API specs, patterns, integration points
- **v1.0.0** (2025-03-29): Initial reference documentation

---

_For working examples, see examples.md. For usage guidance, see SKILL.md._
