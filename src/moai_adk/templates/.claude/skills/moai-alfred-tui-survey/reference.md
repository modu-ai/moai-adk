# TUI Survey Reference - API & Integration Guide

Complete technical reference for the moai-alfred-tui-survey Skill.

_Last updated: 2025-10-22_

---

## Table of Contents

1. [AskUserQuestion API](#askuserquestion-api)
2. [Question Design Patterns](#question-design-patterns)
3. [Response Handling](#response-handling)
4. [Integration with Alfred Sub-agents](#integration-with-alfred-sub-agents)
5. [Decision Trees](#decision-trees)
6. [TUI Frameworks Comparison](#tui-frameworks-comparison)
7. [Validation & Error Handling](#validation--error-handling)

---

## AskUserQuestion API

### Complete API Specification

```typescript
interface AskUserQuestionParams {
  questions: Question[];
}

interface Question {
  question: string;          // Required: The question text
  header: string;            // Required: Short label (max 12 chars)
  multiSelect: boolean;      // Required: Allow multiple selections
  options: Option[];         // Required: 2-4 options
}

interface Option {
  label: string;             // Required: Option display text
  description: string;       // Required: Explanation of option
}

interface AskUserQuestionResponse {
  answers: {
    [questionIndex: string]: string | string[];
  };
}
```

### Field Constraints

| Field | Type | Required | Constraints | Notes |
|-------|------|----------|-------------|-------|
| `question` | string | Yes | 10-200 chars | Must end with `?` |
| `header` | string | Yes | Max 12 chars | Short label for UI |
| `multiSelect` | boolean | Yes | true/false | Checkbox vs radio |
| `options` | Option[] | Yes | 2-4 items | Too few = no choice, too many = fatigue |
| `label` | string | Yes | 3-50 chars | Concise option name |
| `description` | string | Yes | 20-200 chars | Explain implications |

### Example: Minimal Survey

```typescript
AskUserQuestion({
  questions: [
    {
      question: "Which approach should be used?",
      header: "Approach",
      multiSelect: false,
      options: [
        {
          label: "Option A",
          description: "Description of option A with trade-offs."
        },
        {
          label: "Option B",
          description: "Description of option B with trade-offs."
        }
      ]
    }
  ]
})
```

### Example: Multi-Question Survey

```typescript
AskUserQuestion({
  questions: [
    {
      question: "What is the primary goal?",
      header: "Goal",
      multiSelect: false,
      options: [
        {
          label: "Performance",
          description: "Optimize for speed and efficiency."
        },
        {
          label: "Simplicity",
          description: "Optimize for ease of maintenance."
        },
        {
          label: "Flexibility",
          description: "Optimize for extensibility and customization."
        }
      ]
    },
    {
      question: "Which features should be included?",
      header: "Features",
      multiSelect: true,
      options: [
        {
          label: "Feature A",
          description: "Description of feature A."
        },
        {
          label: "Feature B",
          description: "Description of feature B."
        },
        {
          label: "Feature C",
          description: "Description of feature C."
        }
      ]
    }
  ]
})
```

---

## Question Design Patterns

### Pattern 1: Binary Decision

**Use when**: Two mutually exclusive choices

```typescript
{
  question: "Should this feature be enabled?",
  header: "Enable",
  multiSelect: false,
  options: [
    {
      label: "Yes, enable",
      description: "Activate the feature immediately. PRO: Instant availability. CON: May need tuning."
    },
    {
      label: "No, skip",
      description: "Disable for now. PRO: No risk. CON: Feature unavailable."
    }
  ]
}
```

### Pattern 2: Multiple Choice (Exclusive)

**Use when**: 3-4 distinct options, pick one

```typescript
{
  question: "Which deployment strategy?",
  header: "Deploy",
  multiSelect: false,
  options: [
    {
      label: "Rolling deployment",
      description: "Gradual instance replacement. Zero downtime, 15 min duration."
    },
    {
      label: "Blue-green deployment",
      description: "Instant cutover to new environment. 5 min duration, easy rollback."
    },
    {
      label: "Canary deployment",
      description: "Gradual traffic shift (5% → 100%). Safest, 2 hour duration."
    }
  ]
}
```

### Pattern 3: Multiple Selection (Combinable)

**Use when**: Options are independent and can be combined

```typescript
{
  question: "Which security features should be enabled?",
  header: "Security",
  multiSelect: true,
  options: [
    {
      label: "Rate limiting",
      description: "Prevent abuse with request rate limits per user/IP."
    },
    {
      label: "CSRF protection",
      description: "Protect against cross-site request forgery attacks."
    },
    {
      label: "XSS prevention",
      description: "Set Content-Security-Policy and X-Frame-Options headers."
    }
  ]
}
```

### Pattern 4: Scope Definition

**Use when**: Need to define extent or depth of implementation

```typescript
{
  question: "What scope should this feature cover?",
  header: "Scope",
  multiSelect: false,
  options: [
    {
      label: "Minimal (core only)",
      description: "Essential functionality. Effort: 2-3 days."
    },
    {
      label: "Standard (common use cases)",
      description: "Core + frequent scenarios. Effort: 5-7 days."
    },
    {
      label: "Comprehensive (full feature set)",
      description: "All capabilities included. Effort: 10-14 days."
    }
  ]
}
```

### Pattern 5: Technical Choice

**Use when**: Selecting libraries, tools, or frameworks

```typescript
{
  question: "Which state management library?",
  header: "Library",
  multiSelect: false,
  options: [
    {
      label: "Redux Toolkit (v2.x)",
      description: "Industry standard. PRO: Battle-tested, DevTools. CON: More boilerplate."
    },
    {
      label: "Zustand (v5.x)",
      description: "Modern minimalist. PRO: Simple API. CON: Fewer resources."
    },
    {
      label: "Jotai (v2.x)",
      description: "Atomic state. PRO: Fine-grained reactivity. CON: Paradigm shift."
    }
  ]
}
```

### Pattern 6: Risk Assessment

**Use when**: Decision involves trade-offs between risk and speed

```typescript
{
  question: "What migration strategy?",
  header: "Migration",
  multiSelect: false,
  options: [
    {
      label: "Big bang (high risk, fast)",
      description: "Complete replacement. Timeline: 2-3 weeks. Risk: High downtime."
    },
    {
      label: "Incremental (medium risk, medium speed)",
      description: "Gradual rollout. Timeline: 4-6 weeks. Risk: Medium complexity."
    },
    {
      label: "Parallel run (low risk, slow)",
      description: "Support both systems. Timeline: 6-8 weeks. Risk: Low, safe rollback."
    }
  ]
}
```

---

## Response Handling

### Single-Select Response

```typescript
const response = await AskUserQuestion({
  questions: [
    {
      question: "Which approach?",
      header: "Approach",
      multiSelect: false,
      options: [
        { label: "Option A", description: "..." },
        { label: "Option B", description: "..." }
      ]
    }
  ]
});

// Response format
{
  answers: {
    "0": "Option A"  // String value
  }
}

// Access the selected option
const selectedOption = response.answers["0"];
if (selectedOption === "Option A") {
  // Execute Option A logic
}
```

### Multi-Select Response

```typescript
const response = await AskUserQuestion({
  questions: [
    {
      question: "Which features?",
      header: "Features",
      multiSelect: true,
      options: [
        { label: "Feature A", description: "..." },
        { label: "Feature B", description: "..." },
        { label: "Feature C", description: "..." }
      ]
    }
  ]
});

// Response format
{
  answers: {
    "0": ["Feature A", "Feature C"]  // Array of strings
  }
}

// Access the selected options
const selectedFeatures = response.answers["0"];
if (selectedFeatures.includes("Feature A")) {
  // Execute Feature A logic
}
if (selectedFeatures.includes("Feature C")) {
  // Execute Feature C logic
}
```

### Multi-Question Response

```typescript
const response = await AskUserQuestion({
  questions: [
    {
      question: "Primary goal?",
      header: "Goal",
      multiSelect: false,
      options: [/* ... */]
    },
    {
      question: "Which features?",
      header: "Features",
      multiSelect: true,
      options: [/* ... */]
    }
  ]
});

// Response format
{
  answers: {
    "0": "Performance",              // Question 1 (single-select)
    "1": ["Feature A", "Feature B"]  // Question 2 (multi-select)
  }
}

// Access responses
const goal = response.answers["0"];
const features = response.answers["1"];
```

### Handling "Other" Option

When user selects "Other" and provides custom input:

```typescript
const selectedOption = response.answers["0"];
if (selectedOption === "Other") {
  // Follow-up survey or prompt for details
  const detailsResponse = await AskUserQuestion({
    questions: [
      {
        question: "Please specify the custom approach:",
        header: "Custom",
        multiSelect: false,
        options: [
          { label: "Approach X", description: "..." },
          { label: "Approach Y", description: "..." }
        ]
      }
    ]
  });
}
```

---

## Integration with Alfred Sub-agents

### spec-builder Integration

```typescript
// In spec-builder sub-agent
class SpecBuilder {
  async clarifyScope(featureName: string) {
    const response = await AskUserQuestion({
      questions: [
        {
          question: `What should the ${featureName} feature include?`,
          header: "Scope",
          multiSelect: true,
          options: this.generateScopeOptions(featureName)
        }
      ]
    });

    return this.buildSpecFromAnswers(response.answers);
  }

  private generateScopeOptions(featureName: string): Option[] {
    // Context-aware option generation based on codebase analysis
    const analysis = this.analyzeExistingFeatures();
    return [
      {
        label: "Core functionality",
        description: `Essential ${featureName} capabilities. Effort: ${analysis.minimalEffort}`
      },
      {
        label: "Extended features",
        description: `Core + common use cases. Effort: ${analysis.standardEffort}`
      },
      {
        label: "Full implementation",
        description: `All capabilities. Effort: ${analysis.comprehensiveEffort}`
      }
    ];
  }
}
```

### code-builder pipeline Integration

```typescript
// In implementation-planner (Phase 1 of code-builder pipeline)
class ImplementationPlanner {
  async selectLibrary(purpose: string, candidates: Library[]) {
    const options = candidates.map(lib => ({
      label: `${lib.name} (${lib.version})`,
      description: `${lib.description} PRO: ${lib.pros}. CON: ${lib.cons}.`
    }));

    const response = await AskUserQuestion({
      questions: [
        {
          question: `Which library should be used for ${purpose}?`,
          header: "Library",
          multiSelect: false,
          options
        }
      ]
    });

    return candidates.find(lib =>
      response.answers["0"].startsWith(lib.name)
    );
  }
}
```

### doc-syncer Integration

```typescript
// In doc-syncer sub-agent
class DocSyncer {
  async selectDocumentationFormat() {
    const response = await AskUserQuestion({
      questions: [
        {
          question: "Which documentation format should be used?",
          header: "Format",
          multiSelect: false,
          options: [
            {
              label: "Markdown (standard)",
              description: "GitHub-flavored Markdown. Best for most projects."
            },
            {
              label: "MDX (interactive)",
              description: "Markdown + JSX components. For interactive docs."
            },
            {
              label: "AsciiDoc (technical)",
              description: "Rich semantic markup. For complex technical docs."
            }
          ]
        }
      ]
    });

    return response.answers["0"];
  }
}
```

---

## Decision Trees

### Tree 1: Feature Implementation Flow

```
Start: User requests "Add feature X"
    ↓
Context Analysis
    ↓
Q1: "How should feature X be implemented?"
    ├─ Option A → Q2a: "Which data store?"
    ├─ Option B → Q2b: "Which UI framework?"
    └─ Option C → Q2c: "Which API pattern?"
          ↓
Review Summary
    ↓
Confirmation
    ↓
Execute
```

### Tree 2: Refactoring Decision Flow

```
Start: User requests "Refactor component Y"
    ↓
Risk Assessment
    ↓
Q1: "What migration strategy?"
    ├─ Big bang → Q2a: "Maintenance window timing?"
    ├─ Incremental → Q2b: "Which parts first?"
    └─ Parallel → Q2c: "Rollout percentage?"
          ↓
Q3: "Rollback plan?"
    ↓
Review Summary
    ↓
Confirmation
    ↓
Execute
```

### Tree 3: Library Selection Flow

```
Start: User needs "State management"
    ↓
Codebase Analysis (existing patterns, complexity)
    ↓
Q1: "Which library?"
    ├─ Redux → Q2a: "Store structure?"
    ├─ Zustand → Q2b: "Domain stores?"
    └─ Jotai → Q2c: "Atom organization?"
          ↓
Q3: "Persistence strategy?"
    ↓
Review Summary
    ↓
Confirmation
    ↓
Execute
```

---

## TUI Frameworks Comparison

### Popular TUI Frameworks (2025)

| Framework | Language | Strengths | Use Cases |
|-----------|----------|-----------|-----------|
| **Rich** | Python | Colors, tables, progress bars | Logs, reports, dashboards |
| **Textual** | Python | Reactive UI, CSS-like styling | Interactive applications |
| **BubbleTea** | Go | Model-view-update architecture | CLI tools, forms |
| **Ratatui** | Rust | Performance, complex layouts | System monitors, editors |
| **Ink** | JavaScript/Node | React-like components | CLI apps with React patterns |

### Best Practices from TUI Frameworks

**Clarity (Rich, Textual)**:
- Use clear, concise text
- Provide obvious visual cues
- Avoid over-complication

**Simplicity (BubbleTea)**:
- Model-view-update separation
- Single source of truth for state
- Predictable state transitions

**Feedback (All frameworks)**:
- Immediate visual response to user input
- Loading indicators for async operations
- Clear error messages with suggestions

**Accessibility (Textual, Ink)**:
- Keyboard navigation support
- Screen reader compatibility
- High contrast color schemes

### Applying TUI Best Practices to AskUserQuestion

```typescript
// Clarity: Clear question + concise options
{
  question: "Which deployment strategy should be used?",  // Clear
  header: "Deploy",  // Concise label
  options: [
    {
      label: "Rolling",  // Short, clear
      description: "Gradual instance replacement. Zero downtime."  // Concise explanation
    }
  ]
}

// Simplicity: Limit to 2-4 options, avoid nested complexity
{
  multiSelect: false,  // Single choice simplifies decision
  options: [/* 3 options */]  // Not overwhelming
}

// Feedback: Review step provides confirmation
{
  question: "Review your selections. Ready to proceed?",
  options: [
    { label: "Approve and execute", description: "..." },
    { label: "Go back and modify", description: "..." }
  ]
}
```

---

## Validation & Error Handling

### Input Validation

```typescript
function validateSurvey(survey: AskUserQuestionParams): ValidationResult {
  const errors: string[] = [];

  // Validate questions array
  if (!survey.questions || survey.questions.length === 0) {
    errors.push("At least one question is required");
  }

  if (survey.questions.length > 4) {
    errors.push("Maximum 4 questions allowed (avoid decision fatigue)");
  }

  // Validate each question
  survey.questions.forEach((q, index) => {
    // Question text
    if (!q.question || q.question.length < 10) {
      errors.push(`Question ${index + 1}: Text too short (min 10 chars)`);
    }
    if (!q.question.endsWith('?')) {
      errors.push(`Question ${index + 1}: Must end with '?'`);
    }

    // Header
    if (!q.header || q.header.length > 12) {
      errors.push(`Question ${index + 1}: Header must be ≤12 chars`);
    }

    // Options
    if (!q.options || q.options.length < 2) {
      errors.push(`Question ${index + 1}: Minimum 2 options required`);
    }
    if (q.options.length > 4) {
      errors.push(`Question ${index + 1}: Maximum 4 options allowed`);
    }

    // Validate each option
    q.options.forEach((opt, optIndex) => {
      if (!opt.label || opt.label.length < 3) {
        errors.push(`Question ${index + 1}, Option ${optIndex + 1}: Label too short`);
      }
      if (!opt.description || opt.description.length < 20) {
        errors.push(`Question ${index + 1}, Option ${optIndex + 1}: Description too short`);
      }
    });
  });

  return {
    isValid: errors.length === 0,
    errors
  };
}
```

### Error Handling Patterns

```typescript
// Pattern 1: Validation before survey
try {
  const validation = validateSurvey(surveyParams);
  if (!validation.isValid) {
    console.error("Survey validation failed:", validation.errors);
    // Fix issues and retry
    return;
  }

  const response = await AskUserQuestion(surveyParams);
  processResponse(response);
} catch (error) {
  console.error("Survey execution failed:", error);
  // Fallback to default behavior or manual input
}

// Pattern 2: Retry with clarification
async function surveyWithRetry(surveyParams, maxRetries = 2) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await AskUserQuestion(surveyParams);

      if (response.answers["0"] === "Other") {
        // User selected "Other", follow up for details
        surveyParams = generateFollowUpSurvey(response);
        continue;
      }

      return response;
    } catch (error) {
      if (i === maxRetries - 1) throw error;
      console.warn(`Survey attempt ${i + 1} failed, retrying...`);
    }
  }
}

// Pattern 3: Graceful degradation
async function surveyOrFallback(surveyParams, fallback) {
  try {
    return await AskUserQuestion(surveyParams);
  } catch (error) {
    console.warn("Survey failed, using fallback:", fallback);
    return { answers: { "0": fallback } };
  }
}
```

---

## Best Practices Checklist

### Before Creating Survey

- [ ] Analyze codebase context (existing patterns, similar features)
- [ ] Identify specific ambiguity or decision point
- [ ] Verify decision requires user input (not deterministic)
- [ ] Check if similar decisions made before (consistency)
- [ ] Validate survey parameters (2-4 options, clear descriptions)

### During Survey Design

- [ ] Question text is clear and specific
- [ ] Header is ≤12 characters
- [ ] 2-4 options per question (not more, not less)
- [ ] Each option has label + description
- [ ] Trade-offs (PRO/CON) visible when relevant
- [ ] "Other" option available for flexibility
- [ ] Multi-select only for independent choices

### After Capturing Responses

- [ ] Show review summary of all selections
- [ ] Provide "Go back" option to revise
- [ ] Confirm understanding with echo: "You selected [X], which means [Y]"
- [ ] Log selections for future reference (SPEC, docs)
- [ ] Execute based on confirmed choices only

---

**Version**: 2.0.0
**Last Updated**: 2025-10-22
**Total Lines**: 550+
**Coverage**: API spec, patterns, integration, decision trees, validation
**Status**: Production-ready
