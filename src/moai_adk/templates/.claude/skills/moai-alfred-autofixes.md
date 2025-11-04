---
name: moai-alfred-autofixes
description: Safety protocol for automatic code fixes, merge conflicts, and user approval workflow
tier: alfred
freedom: low
tags: [auto-fix, merge-conflicts, safety, approval, protocol]
---

# Auto-Fix & Merge Conflict Protocol

When Alfred detects issues that could automatically fix code (merge conflicts, overwritten changes, deprecated code, etc.), follow this protocol BEFORE making any changes:

## Step 1: Analysis & Reporting

- Analyze the problem thoroughly using git history, file content, and logic
- Write a clear report (plain text, NO markdown) explaining:
  - Root cause of the issue
  - Files affected
  - Proposed changes
  - Impact analysis

**Example Report Format**:
```
Detected Merge Conflict:

Root Cause:
- Commit c054777b removed language detection from develop
- Merge commit e18c7f98 (main → develop) re-introduced the line

Impact:
- .claude/hooks/alfred/shared/handlers/session.py
- src/moai_adk/templates/.claude/hooks/alfred/shared/handlers/session.py

Proposed Fix:
- Remove detect_language() import and call
- Delete "🐍 Language: {language}" display line
- Synchronize both local + package templates
```

## Step 2: User Confirmation (AskUserQuestion)

- Present the analysis to the user
- Use AskUserQuestion to get explicit approval
- Options should be clear: "Should I proceed with this fix?" with YES/NO choices
- Wait for user response before proceeding

## Step 3: Execute Only After Approval

- Only modify files after user confirms
- Apply changes to both local project AND package templates
- Maintain consistency between `/` and `src/moai_adk/templates/`

## Step 4: Commit with Full Context

- Create commit with detailed message explaining:
  - What problem was fixed
  - Why it happened
  - How it was resolved
- Reference the conflict commit if applicable

## Critical Rules

- ❌ NEVER auto-modify without user approval
- ❌ NEVER skip the report step
- ✅ ALWAYS report findings first
- ✅ ALWAYS ask for user confirmation (AskUserQuestion)
- ✅ ALWAYS update both local + package templates together

## Template Synchronization Rules

**Package templates are the source of truth**:
- Changes to `src/moai_adk/templates/` must be synchronized to local project paths
- Local modifications should not override package template changes
- When conflicts arise between local and package templates, prefer package template changes

**Synchronization checklist**:
- [ ] Change applied to `src/moai_adk/templates/` path
- [ ] Change applied to corresponding local project path
- [ ] File contents verified identical
- [ ] Git commit confirms both paths updated
