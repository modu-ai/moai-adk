---
name: moai-alfred-feedback-templates
description: Korean language feedback templates for issue types and Alfred commands with localized messaging patterns
version: 1.0.0
modularized: false
tags:
  - enterprise
  - framework
  - feedback
  - architecture
  - templates
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0
**modularized**: false
**last_updated**: 2025-11-22
**compliance_score**: 75%
**auto_trigger_keywords**: feedback, moai, templates, alfred

## Quick Reference (30 seconds)

# GitHub Issue Writing Template Collection v1.0.0

## Implementation Guide

## Overview

Structured templates by label for writing GitHub issues consistently and clearly.

Using the appropriate template for each label provides:

- ✅ **Clarity**: Issue content is clear and structured
- ✅ **Efficiency**: Reduces missing information
- ✅ **Consistency**: Unifies issue format across team
- ✅ **Traceability**: All important information recorded

## 🐛 Bug Report Template

### When to use?

- Unexpected behavior or error occurs
- Feature doesn't work properly
- Performance degradation or abnormal behavior

### Template

```
## Bug Description

[Briefly describe what the bug is]

## Steps to Reproduce

1. [First step]
2. [Second step]
3. [Step where bug occurs]

## Expected Behavior

[Describe what should happen normally]

## Actual Behavior

[Describe what actually happens]

## Environment Information

- MoAI-ADK Version: [version]
- Python Version: [version]
- OS: [Windows/macOS/Linux]
- Browser: [optional]

## Additional Information

[Screenshots, error messages, logs, etc.]
```

### Example

```
## Bug Description

SPEC not created when executing /moai:1-plan command.

## Steps to Reproduce

1. Execute command `/moai:1-plan "Add new feature"`
2. Complete all steps 1-4
3. Click create button after final confirmation

## Expected Behavior

New SPEC document should be created in .moai/specs/SPEC-XXX/ folder

## Actual Behavior

Only "Spec creation failed" error message output, folder not created

## Environment Information

- MoAI-ADK Version: 0.22.5
- Python Version: 3.11.5
- OS: macOS 14.2
```

## ✨ Feature Request Template

### When to use?

- New feature needed
- Want to suggest improvement for existing feature
- Request to add new command or tool

### Template

```
## Feature Description

[Describe what feature is needed]

## Usage Scenario

[Explain when/how this feature will be used]

Example: When user tries to do X, if Y feature exists, they can easily do Z.

## Expected Benefits

- [Benefit 1]
- [Benefit 2]
- [Benefit 3]

## Implementation Ideas (Optional)

[Suggest ideas on how this feature could be implemented]

## Priority

- 🔴 Urgent: Immediately needed
- 🟠 High: Soon needed
- 🟡 Medium: Moderately needed (default)
- 🟢 Low: Later is fine
```

### Example

```
## Feature Description

/moai:9-feedback command needs automatic information collection feature.
Would be good to automatically collect environment info (version, OS, Git status) when bug reporting.

## Usage Scenario

When user who found bug executes /moai:9-feedback,
automatically collect MoAI-ADK version, Python version, current Git status, recent error logs
and include in issue body.

## Expected Benefits

- Reduced bug reporting time (eliminate manual input)
- Prevent missing environment information
- Have all necessary info from start for bug analysis
```

## ⚡ Improvement Template

### When to use?

- Code quality improvement suggestion
- Performance optimization idea
- User experience (UX) improvement
- Documentation improvement

### Template

```
## Improvement Content

[Describe what you want to improve]

## Current State

[Describe how it currently is]

## Improved State

[Describe how it will change after improvement]

## Performance/Quality Impact

- Performance: [improvement, e.g., 50% response time reduction]
- Usability: [improvement, e.g., 2 steps reduced]
- Maintainability: [effect, e.g., 30% code complexity reduction]

## Implementation Complexity

- ⚪ Low: 1-2 hours
- 🟡 Medium: Half day
- 🔴 High: 1+ days
```

### Example

```
## Improvement Content

Improve usability by reducing AskUserQuestion steps

## Current State

4-step questions when executing /moai:9-feedback (issue type → title → description → priority)

## Improved State

Consolidated to 1 step: Required info (type, priority) at once + auto template generation

## Performance/Quality Impact

- Usability: 4 steps → 1-2 steps (50% reduction)
- Time: ~90 seconds → ~30 seconds (67% reduction)
```

## 🔄 Refactoring Template

### When to use?

- Restructuring existing code
- Improving design patterns
- Resolving technical debt
- Separating or integrating modules

### Template

```
## Refactoring Scope

[Clearly specify what to refactor]

Example: Template class in src/moai_adk/core/template_engine.py

## Current Structure

[Current code structure or issues]

## Improved Structure

[How it will change after refactoring]

## Improvement Reasons

- [Reason 1]
- [Reason 2]
- [Reason 3]

## Impact Analysis

- Modules changed: [module list]
- Tests affected: [test list]
- Compatibility: [compatibility maintenance status]
```

### Example

```
## Refactoring Scope

Unify frontmatter of command files in .claude/commands/alfred/

## Current Structure

Different frontmatter format for each command:
- Files with/without skills section mixed

## Improved Structure

All commands with same frontmatter standard:
```

name: alfred:X
skills: [...]

```

## Improvement Reasons

- Consistency: Unify frontmatter format
- Maintainability: Clarify skill addition/removal
- Automation: Simplify parsing script
```

## 📚 Documentation Template

### When to use?

- Improving documents like README, CLAUDE.md
- Writing new guide/tutorial
- Improving clarity of existing documentation
- Improving code comments/docstrings

### Template

```
## Documentation Content

[Describe what the documentation is about]

## Target Audience

[Who are the primary readers of this documentation?]

Example: New developers, team leaders, API users

## Documentation Structure

[Sketch the structure of the document]

Example:
1. Overview
2. Installation/Setup
3. Usage
4. FAQ
5. References

## Content to Include

- [Content 1]
- [Content 2]
- [Content 3]

## Related Documents

[Existing documents to reference]
```

### Example

```
## Documentation Content

Usage guide for /moai:9-feedback

## Target Audience

MoAI-ADK developers, team members new to bug/feature reporting

## Documentation Structure

1. Introduction (what it is)
2. Step-by-step usage (how to use)
3. Label descriptions (which labels to use)
4. Tips and precautions
5. FAQ

## Content to Include

- How to execute command
- Screenshots for each step
- Label selection guide
- Explanation of automatic environment info collection
```

## ❓ Question/Discussion Template

### When to use?

- Questions for the team
- Matters requiring decision
- Technical concerns or suggestions

### Template

```
## Background

[Explain the background of this question/discussion]

## Question or Suggestion

[Clearly present the core question]

## Options

- [ ] Option 1
- [ ] Option 2
- [ ] Option 3
- [ ] Other

## Decision Criteria

[Explain criteria for decision]

Example: Development time, performance impact, team learning curve, etc.

## Additional Information

[Related additional information or references]
```

### Example

```
## Background

Reviewing /moai:9-feedback AskUserQuestion design improvement.
Received feedback that current 4-step questions are too many.

## Question or Suggestion

Which approach would be best to reduce steps?

## Options

- [ ] Collect required info at once with multiSelect then auto-generate template
- [ ] Maximize defaults (priority default medium, description optional)
- [ ] Summary info only in Step 1, detailed input in Step 2
- [ ] Change to cli format with script (e.g., `alfred:9-feedback bug "title" -d "description"`)

## Decision Criteria

- Usability (minimize steps)
- Information collection accuracy (prevent missing required info)
- Korean language support consistency
```

## 📊 Template Comparison

| Label            | Key Elements               | Minimum Fields       | Additional Info      |
| --------------- | ----------------------- | --------------- | -------------- |
| **bug**         | Reproduction steps, expected vs actual | Description, environment      | Screenshots, logs |
| **feature**     | Scenario, effects          | Description, use case | Implementation ideas  |
| **improvement** | Before/after improvement, reasons         | Description, expected effects | Complexity, impact |
| **refactor**    | Scope, reasons              | Current vs improved    | Impact analysis      |
| **docs**        | Target audience, structure         | Content overview       | Item list to include |
| **question**    | Background, options            | Criteria            | Related info      |

## 🎯 Template Usage Tips

### DO ✅

- ✅ Fill out all sections of the template
- ✅ Write specifically and in detail
- ✅ Enter environment/version info accurately
- ✅ Describe reproduction steps clearly step-by-step
- ✅ Attach screenshots or error messages

### DON'T ❌

- ❌ Don't skip template sections
- ❌ Don't make vague descriptions like "doesn't work"
- ❌ Don't omit environment information
- ❌ Don't mix multiple issues in one issue

## 🔗 References

- **Command**: `/moai:9-feedback`
- **Label Classification**: `Skill("moai-core-issue-labels")`
- **Previous Version**: Supported from v0.22.5+

**Last Updated**: 2025-11-12
**Status**: Production Ready (v1.0.0)

## Advanced Patterns
