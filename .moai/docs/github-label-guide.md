# GitHub Label Standards Guide

## Overview

This guide defines the labeling system for MoAI-ADK GitHub Issues and Pull Requests. Labels are essential for issue tracking, prioritization, and automation.

---

## Label Hierarchy

### Tier 1: SPEC Labels (Mandatory for SPEC Issues)

Used exclusively for SPEC documents and planning phases.

| Label | Color | Description | Usage |
|-------|-------|-------------|-------|
| `spec` | 🟢 #0E8A16 | SPEC document related | **REQUIRED** for all SPEC issues |
| `planning` | 🔵 #1D76DB | Planning phase (Tier 1 work) | **REQUIRED** for all SPEC issues during planning |

**Rule**: All SPEC issues MUST have both `spec` and `planning` labels, plus one priority label.

### Tier 2: Priority Labels (Mandatory for all Issues)

Indicates issue urgency and importance.

| Label | Color | Description | When to Use | Examples |
|-------|-------|-------------|------------|----------|
| `critical` | 🔴 #B60205 | Blocks users, security risk, data loss | System-breaking issues | Windows compatibility failure, auth bypass, data corruption |
| `high` | 🟠 #D93F0B | Core features, major impact | Important features/bugs affecting many users | SPEC implementation, major bugs, critical features |
| `medium` | 🟡 #85E89D | Important but not blocking | Improvements, minor bugs | Refactoring, logging enhancements, small features |
| `low` | 🟢 #85E89D | Nice-to-have improvements | Cosmetic changes, optimization | Performance tweaks, UI polish, code style |

**Rule**: Each issue MUST have exactly ONE priority label.

### Tier 3: Issue Type Labels

Categorizes the nature of the issue.

| Label | Color | Description | When to Use |
|-------|-------|-------------|------------|
| `bug` | 🔴 #d73a4a | Something isn't working | For bug reports and defect issues |
| `enhancement` | 🔵 #a2eeef | Enhancement to existing feature | For improvements to existing functionality |
| `feature-request` | 🟢 #A8E6CF | New feature request | For new capability requests from users |
| `improvement` | 🟡 #FFD93D | Code quality, refactoring, tech debt | For code improvements, refactoring, tech debt |
| `documentation` | 🔵 #0075ca | Documentation improvements | For documentation-only changes |
| `question` | 🟣 #d876e3 | Discussion/help requested | For questions and discussions |

### Tier 4: Meta Labels (Optional)

Special labels for issue source and handling.

| Label | Color | Description | When to Use |
|-------|-------|-------------|------------|
| `reported` | 🔴 #FF6B6B | User-reported via /alfred:9-feedback | Applied automatically by feedback system |
| `help wanted` | 🔵 #008672 | Seeking community help | For issues open to contributor help |

---

## Labeling Rules by Issue Type

### SPEC Issues (Planning & Implementation)

**Format**: `[SPEC-{ID}] {Title} (v{version})`

**Phase 1: Planning Phase**

Required labels:
- ✅ `spec` (identity)
- ✅ `planning` (phase indicator)
- ✅ One priority label

Example:
```
Title: [SPEC-AUTH-001] JWT Authentication System (v0.0.1)
Labels: spec, planning, high
```

**Phase 2: Implementation Phase**

When `/alfred:2-run` begins:
- ✅ `spec` (identity) - keep
- ❌ `planning` (remove when implementation starts)
- ✅ One priority label - keep

Labels are updated by automation.

### Bug Reports

**Format**: `🐛 [BUG] {Brief description}`

Required labels:
- ✅ `bug`
- ✅ One priority label (default: `high`)
- ✅ `reported` (if created via `/alfred:9-feedback`)

Example:
```
Title: 🐛 [BUG] SessionStart hook fails on Windows
Labels: bug, critical, reported
```

### Feature Requests

**Format**: `✨ [FEATURE] {Description}`

Required labels:
- ✅ `feature-request`
- ✅ One priority label (default: `medium`)
- ✅ `reported` (if created via `/alfred:9-feedback`)

Example:
```
Title: ✨ [FEATURE] Multi-language support
Labels: feature-request, high, reported
```

### Improvements & Refactoring

**Format**: `⚡ [IMPROVEMENT] {Description}`

Required labels:
- ✅ `improvement`
- ✅ One priority label (default: `medium`)

Example:
```
Title: ⚡ [IMPROVEMENT] Refactor template processor for clarity
Labels: improvement, medium
```

### Questions & Discussions

**Format**: `❓ [QUESTION] {Description}`

Required labels:
- ✅ `question`
- Optional: `help wanted`

Example:
```
Title: ❓ [QUESTION] How to extend Alfred with custom agents?
Labels: question, help wanted
```

---

## Automated Labeling

### Workflow Automation

#### `/alfred:9-feedback` (Issue Creator)

Automatically adds labels based on issue type:
- **Bug**: `bug`, `reported`, priority
- **Feature**: `feature-request`, `reported`, priority
- **Improvement**: `improvement`, `reported`, priority
- **Question**: `question`, `reported`

#### `spec-issue-sync.yml` (GitHub Actions)

Automatically applied to SPEC issues from PRs:

**Initial Creation**:
- Always adds: `spec`, `planning`, priority label

**Updates**:
- Validates mandatory labels
- Auto-corrects missing labels
- Ensures only one priority label exists

#### `git-manager` (Alfred Agent)

When creating SPEC PRs:
- Ensures mandatory labels are present
- Updates labels if Issue already exists
- Uses fallback search for unlabeled Issues

### Manual Labeling

#### Command-line

```bash
# Add labels to an issue
gh issue edit {number} --add-label "spec,planning,high"

# Remove a label
gh issue edit {number} --remove-label "backlog"

# List all labels
gh label list

# Find issues with specific label
gh issue list --label "bug" --state open
```

---

## Best Practices

### 1. Always Use Priority Labels

Every issue must have exactly ONE priority label for proper triaging.

```
✅ GOOD:  Labels: bug, critical
❌ WRONG: Labels: bug (no priority)
❌ WRONG: Labels: bug, high, medium (multiple priorities)
```

### 2. Keep SPEC Issues Properly Tagged

SPEC issues are the source of truth for features.

```
✅ GOOD:  [SPEC-AUTH-001] JWT Authentication (v0.1.0)
          Labels: spec, planning, high

❌ WRONG: [SPEC] Authentication
          Labels: spec (missing planning and priority)
```

### 3. Use Consistent Naming

Follow emoji + category format for clarity.

```
✅ GOOD:
- 🐛 [BUG] Session timeout error
- ✨ [FEATURE] Export reports
- ⚡ [IMPROVEMENT] Reduce bundle size
- ❓ [QUESTION] How to extend Alfred?

❌ WRONG:
- Bug in session handling
- Add export feature
- Code quality improvement
- Question about extensibility
```

### 4. Report via /alfred:9-feedback

The feedback system ensures consistent labeling.

```bash
/alfred:9-feedback
# Creates properly labeled issues with "reported" label
```

### 5. Don't Create Invalid Labels

Workflow validation prevents label typos. Invalid labels are rejected.

```
✅ VALID:   high, critical, medium, low
❌ INVALID: priority-high, high-priority, HIGH
```

---

## Label Deprecations & Removals

### Removed Labels (No Longer Used)

These labels were removed to reduce clutter:

- ~~`backlog`~~ - Use GitHub's native "Open" status instead
- ~~`in-progress`~~ - Use GitHub's PR/Issue assignee instead
- ~~`done`~~ - Use GitHub's native "Closed" status instead
- ~~`duplicate`~~ - Use GitHub's merge/close functionality
- ~~`invalid`~~ - Use GitHub's close with reason instead
- ~~`wontfix`~~ - Use GitHub's close with reason instead
- ~~`good first issue`~~ - Use `help wanted` instead
- ~~`help-wanted`~~ - Corrected to `help wanted` (GitHub standard)

---

## Troubleshooting

### Issue Created Without Labels

**Cause**: Manual creation or API bypass

**Solution**: Add labels manually
```bash
gh issue edit {number} --add-label "spec,planning,high"
```

### Label Appears with Typo (e.g., "priority-high")

**Cause**: Code using old label format

**Solution**: Update to direct label names (see issue_creator.py fixes)

### SPEC Issue Missing Planning Label

**Cause**: Phase advancement (moving from planning to implementation)

**Solution**: Automated workflows update this - verify completion

### Too Many Priority Labels

**Cause**: User manually added multiple priorities

**Solution**: Remove extras, keep only one
```bash
gh issue edit {number} --remove-label "low,medium"
```

---

## Compliance Checklist

Use this checklist before closing an issue:

- [ ] Issue has exactly ONE priority label?
- [ ] All required labels for issue type present?
- [ ] SPEC issues have `spec` + `planning` + priority?
- [ ] Title follows format with emoji + category?
- [ ] No typos or invalid label names?
- [ ] Automated workflows ran (check for label updates)?

---

## References

- [GitHub About labels](https://docs.github.com/en/issues/using-labels-and-milestones-to-track-work/managing-labels)
- [MoAI-ADK Issue Creation](/alfred:9-feedback)
- [SPEC Document Guide](.moai/specs/)

---

**Last Updated**: 2025-10-31
**Maintained By**: MoAI-ADK Team
**Questions?**: Create an issue with `question` label
