# Feedback Templates

Structured GitHub issue templates for consistent feedback submission.

## 6 Template Types

```
Feedback Template Types:
├── 🐛 Bug Report
│   ├── Description
│   ├── Reproduction steps
│   ├── Expected vs Actual behavior
│   └── Environment info
├── ✨ Feature Request
│   ├── Feature description
│   ├── Usage scenarios
│   ├── Expected effects
│   └── Priority
├── ⚡ Improvement
│   ├── Current state
│   ├── Improved state
│   ├── Performance/Quality impact
│   └── Implementation complexity
├── 🔄 Refactor
│   ├── Refactoring scope
│   ├── Current vs Improved structure
│   ├── Improvement reasons
│   └── Impact analysis
├── 📚 Documentation
│   ├── Document content
│   ├── Target audience
│   ├── Document structure
│   └── Related docs
└── ❓ Question/Discussion
    ├── Background
    ├── Question or proposal
    ├── Options
    └── Decision criteria
```

## Bug Report Template

```markdown
## Bug Description
[Brief description of the bug]

## Reproduction Steps
1. [First step]
2. [Second step]
3. [Step where bug occurs]

## Expected Behavior
[What should happen normally]

## Actual Behavior
[What actually happens]

## Environment
- MoAI-ADK Version: [version]
- Python Version: [version]
- OS: [Windows/macOS/Linux]

## Additional Information
[Screenshots, error messages, logs]
```

## Feature Request Template

```markdown
## Feature Description
[Brief description of the feature]

## Usage Scenarios
1. [Scenario 1]
2. [Scenario 2]

## Expected Effects
[Expected outcomes and benefits]

## Priority
- [ ] High
- [ ] Medium
- [ ] Low

## Additional Context
[Any additional information]
```

## Usage Integration

**Auto-triggered by `/moai:9-feedback` command**:
1. User executes `/moai:9-feedback "description"`
2. Skill selects appropriate template type
3. Template is populated with user input
4. GitHub issue is created automatically

## Success Metrics

- **Feedback Completeness**: 95% GitHub issues with complete information
- **Response Time**: Issues resolved 40% faster with complete templates
