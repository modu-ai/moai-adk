# moai-alfred-interactive-questions - Working Examples

_Last updated: 2025-10-23_

## Example 1: Simple Confirmation

**Scenario**: Destructive operation (delete 50 user records)

```typescript
const answer = await AskUserQuestion({
  questions: [{
    question: "This will permanently delete 50 user records. Continue?",
    header: "Confirm",
    multiSelect: false,
    options: [
      { label: "Yes, proceed", description: "Irreversible operation. Delete now." },
      { label: "No, cancel", description: "Abort and return." }
    ]
  }]
});

if (answer["Confirm"] === "Yes, proceed") {
  // Execute deletion
}
```

---

## Example 2: Implementation Approach

**Scenario**: "Add a completion page" (vague)

```typescript
const answer = await AskUserQuestion({
  questions: [
    {
      question: "How should the completion page be implemented?",
      header: "Approach",
      multiSelect: false,
      options: [
        { label: "New public route", description: "Separate page at /competition-closed" },
        { label: "Modify /end page", description: "Add conditional logic to existing page" },
        { label: "Environment flag", description: "Use NEXT_PUBLIC_COMPETITION_CLOSED" }
      ]
    },
    {
      question: "For logged-in participants?",
      header: "Behavior",
      multiSelect: false,
      options: [
        { label: "Show results", description: "Redirect to /end with full history" },
        { label: "Show message", description: "'Competition concluded' only" }
      ]
    }
  ]
});

const approach = answer["Approach"];
const behavior = answer["Behavior"];
// Proceed with confirmed specifications
```

---

## Example 3: Multi-Select (Independent Features)

**Scenario**: "Which testing frameworks?"

```typescript
const answer = await AskUserQuestion({
  questions: [{
    question: "Which testing frameworks to include?",
    header: "Test Tools",
    multiSelect: true,  // User can select multiple
    options: [
      { label: "Unit tests (Vitest)", description: "Fast unit testing." },
      { label: "E2E tests (Playwright)", description: "Browser automation." },
      { label: "Visual regression", description: "Screenshot comparison." }
    ]
  }]
});

const selectedTools = answer["Test Tools"];  // ["Unit tests (Vitest)", "E2E tests (Playwright)"]
// Install selected frameworks
```

---

## Example 4: Sequential Questions (Conditional)

**Scenario**: Q2 depends on Q1 answer

```typescript
// Question 1
const q1 = await AskUserQuestion({
  questions: [{
    question: "Enable authentication?",
    header: "Auth",
    options: [
      { label: "Yes", description: "User login required." },
      { label: "No", description: "Public access." }
    ]
  }]
});

// Question 2 (only if Q1 = "Yes")
if (q1["Auth"] === "Yes") {
  const q2 = await AskUserQuestion({
    questions: [{
      question: "Which auth provider?",
      header: "Provider",
      options: [
        { label: "JWT + email", description: "Traditional email/password." },
        { label: "OAuth", description: "Third-party login (Google, GitHub)." }
      ]
    }]
  });

  const provider = q2["Provider"];
  // Setup authentication
}
```

---

## Example 5: Feature Selection

**Scenario**: "Which i18n library?"

```typescript
const answer = await AskUserQuestion({
  questions: [{
    question: "Which i18n library?",
    header: "Library",
    multiSelect: false,
    options: [
      { label: "next-intl", description: "Next.js 15+ native, best DX." },
      { label: "react-i18next", description: "Popular, flexible, more setup." },
      { label: "Format.js", description: "ICU format, complex but powerful." }
    ]
  }]
});

const library = answer["Library"];  // "next-intl"
// Install and configure library
```

---

## Example 6: Error Recovery

**Scenario**: 3 tests failed, ask how to proceed

```typescript
const answer = await AskUserQuestion({
  questions: [{
    question: "3 tests failed in auth.test.ts. How should we proceed?",
    header: "Test Failure",
    multiSelect: false,
    options: [
      { label: "Debug immediately", description: "Stop and fix now." },
      { label: "Skip tests", description: "NOT RECOMMENDED - continues with failing tests." },
      { label: "Rollback changes", description: "Revert to last passing commit." }
    ]
  }]
});

if (answer["Test Failure"] === "Debug immediately") {
  // Trigger debug workflow
} else if (answer["Test Failure"] === "Rollback changes") {
  // Rollback git changes
}
```

---

## Example 7: /alfred:1-plan Integration

**Scenario**: spec-builder detects ambiguous SPEC title

```typescript
// Inside spec-builder
function shouldAskForClarification(title: string): boolean {
  return title.length < 20 || title.includes("feature/add/implement");
}

if (shouldAskForClarification(userTitle)) {
  const answer = await AskUserQuestion({
    questions: [{
      question: "Can you clarify what this SPEC should cover?",
      header: "Scope",
      options: [
        { label: "Authentication system", description: "User login, JWT tokens, refresh logic." },
        { label: "UI components", description: "Reusable components library." },
        { label: "API endpoints", description: "REST API implementation." }
      ]
    }]
  });

  specTitle = answer["Scope"];
}
```

---

## Example 8: /alfred:2-run Integration

**Scenario**: code-builder detects multiple implementation paths

```typescript
// Inside code-builder
if (multipleValidImplementationPaths) {
  const answer = await AskUserQuestion({
    questions: [{
      question: "How should we implement JWT storage?",
      header: "Storage",
      options: [
        { label: "HttpOnly cookie", description: "Most secure, no JS access." },
        { label: "LocalStorage", description: "Accessible but vulnerable to XSS." },
        { label: "Memory + refresh", description: "Balance between security and UX." }
      ]
    }]
  });

  const storageStrategy = answer["Storage"];
  implementWithStrategy(storageStrategy);
}
```

---

## Example 9: /alfred:3-sync Integration

**Scenario**: doc-syncer asks about sync mode

```typescript
// Inside doc-syncer
const answer = await AskUserQuestion({
  questions: [{
    question: "Which sync mode?",
    header: "Mode",
    multiSelect: false,
    options: [
      { label: "auto", description: "Smart detection, sync affected docs only." },
      { label: "force", description: "Regenerate all docs from scratch." },
      { label: "status", description: "Show status without making changes." }
    ]
  }]
});

const syncMode = answer["Mode"];
runSync(syncMode);
```

---

For detailed guidance, see **SKILL.md** and **reference.md**
