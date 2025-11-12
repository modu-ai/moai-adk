---
name: "moai-alfred-ask-user-questions"
version: "4.0.0"
created: 2025-10-22
updated: 2025-11-12
status: stable
description: Enterprise interactive survey orchestrator with AskUserQuestion tool integration, multi-select support, conditional branching, error recovery, and production-grade decision automation across all Alfred workflows; activates for requirement clarification, architectural decisions, risky operations, feature selection, and complex multi-step user interactions
keywords: ['interactive-surveys', 'user-clarification', 'decision-making', 'AskUserQuestion', 'multi-select', 'conditional-flow', 'error-recovery', 'workflow-automation', 'enterprise-ux', 'production-surveys']
allowed-tools: 
  - AskUserQuestion
---

# Enterprise Interactive Survey Orchestrator v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-ask-user-questions |
| **Version** | 4.0.0 Enterprise (2025-11-12) |
| **Core Tool** | `AskUserQuestion` (Claude Code built-in) |
| **Auto-load** | When Alfred detects ambiguity in requests |
| **Tier** | Alfred (Workflow Orchestration) |
| **Allowed tools** | AskUserQuestion |
| **Lines of Content** | 850+, with 10+ production examples |
| **Progressive Disclosure** | 3-level (quick-reference, patterns, advanced) |

---

## 🚀 What It Does (Enterprise Context)

**Purpose**: Empower Alfred sub-agents to **actively conduct enterprise-grade surveys** for requirement clarification, architectural decisions, and complex decision automation.

Leverages Claude Code's native `AskUserQuestion` tool to collect explicit, structured user input that transforms vague requests into precise specifications with guaranteed UX quality across all models.

**Enterprise Capabilities**:
- ✅ Single-select & multi-select option types (independent or dependent)
- ✅ 1-4 questions per survey (cognitive load optimization)
- ✅ 2-4 options per question with trade-off analysis
- ✅ Automatic "Other" option for custom input validation
- ✅ Conditional branching based on user answers
- ✅ Error recovery and retry logic
- ✅ Integration across all Alfred commands (Plan/Run/Sync)
- ✅ Multi-language support (user's configured language)
- ✅ Accessibility-first TUI design
- ✅ Reduces ambiguity → single interaction vs 3-5 iterations

---

## 🎯 When to Use (Decision Framework)

### ✅ MANDATORY ASK when user intent is ambiguous:

1. **Vague noun phrases**: "Add dashboard", "Refactor auth", "Improve performance"
   - Missing concrete specification or scope
   
2. **Missing scope definition**: No specification of WHERE, WHO, WHAT, HOW, WHEN
   - Could affect 5+ files or multiple modules
   
3. **Multiple valid paths**: ≥2 reasonable implementation approaches
   - Different trade-offs (speed vs quality, simple vs comprehensive)
   
4. **Trade-off decisions**: Performance vs reliability, cost vs features
   - No single objectively best answer
   
5. **Risky operations**: Destructive actions (delete, migrate, reset)
   - Explicit informed consent required

6. **Architectural decisions**: Technology selection, API design, database choice
   - Long-term impact requires clarification

### ❌ DON'T ask when:
- User explicitly specified exact requirements
- Decision is automatic (no choices, pure routing)
- Single obvious path exists (no alternatives)
- Quick yes/no confirmation only (keep it brief)
- Information already provided in conversation

---

## 🎨 Design Principles (Enterprise Standards)

### Core Principle: **Certainty Over Guessing**

**Golden Rule**: When in doubt, **ask the user** instead of assuming.

**Why**:
- ✅ User sees exactly what you'll do → no surprises
- ✅ Single interaction vs 3-5 rounds of back-and-forth
- ✅ Fast → execute with certainty
- ✅ Reduces "vibe coding" frustration
- ✅ Builds trust through transparency

**Pattern**:
```
Ambiguous request detected
         ↓
Call AskUserQuestion({questions: [...]})
         ↓
User selects from clear options
         ↓
Proceed with confirmed specifications
```

---

## 🏗️ Architecture: 3-Level Progressive Disclosure

### Level 1: Quick Start (Minimal Invocation)

**Single-Select Pattern**:
```typescript
const answer = await AskUserQuestion({
  questions: [
    {
      question: "How should we implement this?",
      header: "Approach",          // max 12 chars
      multiSelect: false,
      options: [
        {
          label: "Option 1",       // 1-5 words
          description: "What it does and why you'd pick it."
        },
        {
          label: "Option 2",
          description: "Alternative with different trade-offs."
        }
      ]
    }
  ]
});

// Returns: { "Approach": "Option 1" }
```

**Multi-Select Pattern** (independent features):
```typescript
const answer = await AskUserQuestion({
  questions: [
    {
      question: "Which features should we enable?",
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

// Returns: { "Features": ["Feature A", "Feature C"] }
```

### Level 2: Enterprise Patterns

**Batch Questions** (related decisions):
```typescript
const answer = await AskUserQuestion({
  questions: [
    {
      question: "Which database technology?",
      header: "Database",
      options: [
        { label: "PostgreSQL", description: "Relational, ACID, mature" },
        { label: "MongoDB", description: "Document, flexible schema" }
      ]
    },
    {
      question: "Cache strategy?",
      header: "Caching",
      options: [
        { label: "Redis", description: "In-memory, fast" },
        { label: "Memcached", description: "Distributed cache" }
      ]
    }
  ]
});
```

**Conditional Flow** (dependent decisions):
```typescript
let questions = [
  {
    question: "What type of deployment?",
    header: "Deployment Type",
    options: [
      { label: "Docker", description: "Containerized" },
      { label: "Kubernetes", description: "Orchestrated" },
      { label: "Serverless", description: "Functions-as-a-Service" }
    ]
  }
];

const initialAnswer = await AskUserQuestion({ questions });

// Follow-up based on initial choice
if (initialAnswer["Deployment Type"] === "Kubernetes") {
  const kubeAnswer = await AskUserQuestion({
    questions: [
      {
        question: "Which cluster provider?",
        header: "K8s Provider",
        options: [
          { label: "AWS EKS", description: "Amazon Elastic Kubernetes" },
          { label: "GCP GKE", description: "Google Kubernetes Engine" },
          { label: "Self-Managed", description: "On-premises cluster" }
        ]
      }
    ]
  });
}
```

### Level 3: Advanced (Error Handling & Validation)

**Custom Input Validation** ("Other" option):
```typescript
const answer = await AskUserQuestion({
  questions: [
    {
      question: "Which framework?",
      header: "Framework",
      options: [
        { label: "React", description: "UI library" },
        { label: "Vue", description: "Progressive framework" },
        { label: "Other", description: "Custom framework or library" }
      ]
    }
  ]
});

// Validate custom input
if (answer["Framework"] === "Other" || 
    !VALID_FRAMEWORKS.includes(answer["Framework"])) {
  const customAnswer = validateCustomInput(answer["Framework"]);
  if (!customAnswer) {
    // Re-ask with guidance
    return retryWithGuidance();
  }
}
```

**Error Recovery**:
```typescript
try {
  const answer = await AskUserQuestion({...});
} catch (error) {
  if (error.type === "UserCancelled") {
    // User pressed ESC - use default or abort
    return fallbackToDefault();
  } else if (error.type === "InvalidInput") {
    // Validate and retry
    return retryWithValidation();
  }
}
```

**Risky Operation Confirmation**:
```typescript
const answer = await AskUserQuestion({
  questions: [
    {
      question: "This will DELETE 15 branches and merge to main. Continue?",
      header: "Destructive Op",
      options: [
        {
          label: "Proceed",
          description: "Delete branches (CANNOT BE UNDONE)"
        },
        {
          label: "Dry Run",
          description: "Show what would be deleted"
        },
        {
          label: "Cancel",
          description: "Abort entire process"
        }
      ]
    }
  ]
});

if (answer["Destructive Op"] === "Proceed") {
  // Require additional final confirmation
  const final = await AskUserQuestion({
    questions: [
      {
        question: "Type DELETE to confirm irreversible action:",
        header: "Final Confirmation"
      }
    ]
  });
  
  if (final["Final Confirmation"] === "DELETE") {
    executeDeletion();
  }
}
```

---

## 📋 Key Constraints (TUI Optimization)

| Constraint | Reason | Example |
|-----------|--------|---------|
| **1-4 questions max** | Avoid user fatigue | Use follow-up surveys instead |
| **2-4 options per Q** | Prevent choice overload | Avoid 5+ options (decision paralysis) |
| **Header ≤12 chars** | TUI layout fit | "DB Choice" not "Which Database Technology" |
| **Label 1-5 words** | Quick scanning | "PostgreSQL" not "SQL Database by PostgreSQL" |
| **Description required** | Enables informed choice | Always explain trade-offs |
| **Auto "Other" option** | Always available | System adds automatically for custom input |
| **No HTML/markdown** | Plain text TUI | Use formatting sparingly |
| **Language matching** | User experience | Always match configured conversation_language |

---

## 🔄 Integration with Alfred Sub-agents

| Sub-agent | When to Ask | Example Trigger | Questions |
|-----------|-------------|-----------------|-----------|
| **spec-builder** (`/alfred:1-plan`) | SPEC title vague, scope undefined | "Add feature" without specifics | "Feature type?", "Scope?", "Users affected?" |
| **tdd-implementer** (`/alfred:2-run`) | Implementation approach unclear | Multiple valid implementation paths | "Architecture?", "Libraries?", "Constraints?" |
| **doc-syncer** (`/alfred:3-sync`) | Sync scope unclear | Full vs partial sync decision | "Full sync?", "Which files?", "Auto-commit?" |
| **qa-validator** | Review depth unclear | Quick vs comprehensive check | "Review level?", "Security focus?", "Performance?" |

---

## 🎓 Top 10 Usage Patterns

### Pattern 1: Feature Type Clarification
**Trigger**: "Add dashboard feature" without specifics  
**Question**: "Which dashboard type: Analytics, Admin, or Profile?"  
**Outcome**: Narrowed scope → faster implementation

### Pattern 2: Implementation Approach Selection
**Trigger**: Multiple valid tech choices  
**Question**: "JWT tokens, session-based, or OAuth?"  
**Outcome**: Locked architecture → deterministic development

### Pattern 3: Risky Operation Confirmation
**Trigger**: Destructive action (delete, migrate, reset)  
**Question**: "This will delete X. Proceed?" with retry confirmation  
**Outcome**: Explicit consent + audit trail

### Pattern 4: Multi-Feature Selection
**Trigger**: "Which features to enable?"  
**Question**: Multi-select of independent features  
**Outcome**: Precise feature set → no scope creep

### Pattern 5: Sequential Conditional Decisions
**Trigger**: Dependent choices (Q2 depends on Q1)  
**Question**: First survey → follow-up based on answer  
**Outcome**: Progressive narrowing → precise specification

### Pattern 6: Technology Stack Selection
**Trigger**: "Build with what stack?"  
**Question**: Database, cache, queue, API type  
**Outcome**: Full stack locked → team alignment

### Pattern 7: Performance vs Reliability
**Trigger**: Trade-off between conflicting goals  
**Question**: "Optimize for speed, reliability, or cost?"  
**Outcome**: Explicit requirements → informed trade-offs

### Pattern 8: Custom Input Handling
**Trigger**: "Other" option selected  
**Question**: "Please describe..." with validation  
**Outcome**: Unexpected inputs handled gracefully

### Pattern 9: Experience Level Calibration
**Trigger**: Unclear target audience  
**Question**: "Beginner, intermediate, or advanced?"  
**Outcome**: Content adapted to expertise level

### Pattern 10: Approval Workflow
**Trigger**: Major decision needs consensus  
**Question**: Multi-team approval with options  
**Outcome**: Documented decision with stakeholder buy-in

---

## ✅ Best Practices Summary

### DO's
- **Be specific**: "Which database type?" not "What should we use?"
- **Provide context**: Include file names, scope, or impact
- **Order logically**: General → Specific; safest option first
- **Flag risks**: Use "NOT RECOMMENDED" or "CAUTION:" prefixes
- **Explain trade-offs**: Mention time, resources, complexity, performance
- **Limit options**: 2-4 per question (not 5+ for decision paralysis)
- **Validate custom input**: Check "Other" responses for validity
- **Batch related Q's**: Keep related decisions together (max 4 questions)

### DON'Ts
- **Overuse questions**: Only ask when genuinely ambiguous
- **Too many options**: 5+ options cause decision paralysis
- **Vague labels**: "Option A", "Use tokens", "Option 2"
- **Skip descriptions**: User needs rationale for informed choice
- **Hide trade-offs**: Always mention implications and costs
- **Ask for obvious**: Single clear path = no question needed
- **Recursive surveys**: Avoid asking the same question twice
- **Ignore language**: Always match user's configured conversation_language

---

## 🔗 Related Skills

- `moai-alfred-personas` (Communication styles by user level)
- `moai-alfred-spec-authoring` (SPEC clarity & structure)
- `moai-foundation-specs` (SPEC format & requirements)
- `moai-alfred-language-detection` (Conversation language handling)

---

## 📚 Quick Reference Card

| Scenario | Action | Key Points |
|----------|--------|-----------|
| **Vague request** | Ask for clarification | 1-4 questions max |
| **Multiple approaches** | Let user choose | Show trade-offs clearly |
| **Risky operation** | Get explicit consent | Require final confirmation |
| **Feature selection** | Use multi-select | Independent options only |
| **Dependent decisions** | Use sequential surveys | Ask follow-ups based on answers |
| **Custom input** | Validate carefully | Re-ask if invalid |
| **Accessibility** | Plain text UI | No complex formatting |

---

## Token Budget Optimization

- **Average per survey**: 500-800 tokens
- **Typical workflow**: 1-2 surveys per task (1,000-1,600 tokens total)
- **Benefit**: Eliminates 3-5 clarification rounds (3,000-5,000 tokens saved)
- **ROI**: Net savings of 1,400-3,500 tokens per interaction

---

**For detailed API specifications**: [reference.md](reference.md)  
**For real-world examples**: [examples.md](examples.md)  
**Last Updated**: 2025-11-12  
**Status**: Production Ready (Enterprise v4.0.0)
