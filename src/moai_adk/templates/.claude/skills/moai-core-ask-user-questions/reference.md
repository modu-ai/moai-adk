# Alfred Ask User Questions - Reference Documentation

## Official Documentation

### Core References
- **AskUserQuestion API**: [Claude Code Documentation - Built-in Tools]
- **Alfred 4-Step Workflow**: `CLAUDE.md` - 4-Step Agent-Based Workflow Logic
- **Language Configuration**: `.moai/config/config.json` - conversation_language

### AskUserQuestion API Specification

#### Function Signature
```typescript
AskUserQuestion({
  questions: [
    {
      question: string,           // The question text (user language)
      header?: string,            // Short category label (≤12 chars, no emojis)
      multiSelect?: boolean,      // Allow multiple selections (default: false)
      options: [
        {
          label: string,          // Option label (1-5 words, user language)
          description: string     // Option description (user language)
        }
      ]
    }
  ]
})
```

#### Language Enforcement
- **CRITICAL**: All text MUST use user's configured conversation_language
- **Source of Truth**: `.moai/config/config.json` → `language.conversation_language`
- **Runtime Check**: `cat .moai/config.json | jq '.language.conversation_language'`
- **Zero Tolerance**: No exceptions, no fallbacks to English

#### Constraints and Best Practices
- **1-4 questions maximum**: Prevent user fatigue
- **2-4 options per question**: Avoid choice overload
- **Header ≤12 characters**: TUI layout compatibility
- **Label 1-5 words**: Quick scanning
- **Description required**: Enable informed decisions
- **Auto "Other" option**: Custom input always available

#### Error Handling
```typescript
try {
  const answer = await AskUserQuestion({...});
} catch (error) {
  // User cancelled (ESC key) or survey error
  console.log("User cancelled or survey failed");
  // Fall back to default or abort operation
}
```

### Integration Patterns

#### Alfred Sub-agent Integration
| Sub-agent | Trigger Pattern | Example Use Case |
|-----------|-----------------|------------------|
| spec-builder | Vague requirements | "Add dashboard feature" |
| tdd-implementer | Multiple implementation paths | "Which approach for authentication?" |
| doc-syncer | Scope clarification | "Full sync or partial update?" |
| git-manager | Destructive operations | "Delete this branch?" |

#### Question Design Patterns

##### Pattern 1: Implementation Approach
```typescript
{
  question: "How should we implement the user authentication system?",
  header: "Authentication",
  options: [
    {
      label: "JWT Tokens",
      description: "Stateless authentication with access/refresh tokens"
    },
    {
      label: "Session-Based",
      description: "Server-side sessions with database storage"
    },
    {
      label: "OAuth Provider",
      description: "Third-party authentication (Google, GitHub, etc.)"
    }
  ]
}
```

##### Pattern 2: Scope Confirmation
```typescript
{
  question: "Which components should be included in this refactoring?",
  header: "Scope",
  multiSelect: true,
  options: [
    {
      label: "Database Layer",
      description: "Update database schema and migration scripts"
    },
    {
      label: "API Endpoints",
      description: "Modify REST API controllers and routes"
    },
    {
      label: "Frontend Components",
      description: "Update React components and state management"
    }
  ]
}
```

## External References

### User Interface Design
- **Terminal UI Best Practices**: [Ink - React for CLIs](https://github.com/vadimdemedes/ink)
- **Command Line Interface Guidelines**: [CLI Guidelines](https://clig.dev/)
- **Survey Design Principles**: [SurveyMonkey Best Practices](https://www.surveymonkey.com/mp/survey-best-practices/)

### Decision Support
- **Choice Architecture**: [Nudge Theory](https://en.wikipedia.org/wiki/Nudge_theory)
- **Cognitive Load Theory**: [Cognitive Load in UX](https://www.nngroup.com/articles/minimize-cognitive-load/)

---

**Last Updated**: 2025-11-11
**Related Skills**: moai-core-agent-guide, moai-core-workflow
