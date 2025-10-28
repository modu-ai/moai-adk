# 🎯 Quick Issue Creation Guide

> **MoAI-ADK v0.7.0+** - Create GitHub Issues instantly with `/alfred:9-feedback` interactive dialog

## Overview

The Quick Issue Creation system allows developers to report bugs, request features, suggest improvements, and ask questions directly to GitHub Issues without leaving their development context.

**Key Benefit**: Convert problems into tracked GitHub Issues in seconds, maintaining development flow.

---

## 🚀 Quick Start

### Basic Usage

```bash
# Report a bug
/alfred:9-help --bug 'Login button not responding on homepage'

# Request a feature
/alfred:9-help --feature 'Add dark mode theme support'

# Suggest an improvement
/alfred:9-help --improvement 'Optimize database query in UserService'

# Ask a question
/alfred:9-help --question 'What is the recommended approach for API versioning?'
```

### What Happens Next

1. **Parsing**: Alfred extracts issue type and content
2. **Priority Selection**: You select issue priority (critical/high/medium/low)
3. **Issue Creation**: GitHub Issue is created with:
   - Formatted title (with emoji and type indicator)
   - Your description
   - Automatic labels based on type and priority
   - Metadata footer
4. **Confirmation**: You receive the issue URL for immediate sharing

**Example Output**:
```
✅ GitHub Issue #456 created successfully
📋 Title: 🐛 [BUG] Login button not responding on homepage
🔴 Priority: High
🏷️  Labels: bug, reported, priority-high
🔗 URL: https://github.com/owner/repo/issues/456
```

---

## 📋 Issue Types & Labels

### 🐛 Bug Reports (`--bug`)

Use this when you discover a problem or unexpected behavior.

**Automatic Labels**: `bug`, `reported`

**Example**:
```bash
/alfred:9-help --bug 'Payment form crashes when credit card has < 4 digits'
```

---

### ✨ Feature Requests (`--feature`)

Use this to propose new functionality.

**Automatic Labels**: `feature-request`, `enhancement`

**Example**:
```bash
/alfred:9-help --feature 'Add webhook support for payment notifications'
```

---

### ⚡ Improvement Suggestions (`--improvement`)

Use this to suggest enhancements to existing features.

**Automatic Labels**: `improvement`, `enhancement`

**Example**:
```bash
/alfred:9-help --improvement 'Reduce database queries in checkout process by 50%'
```

---

### ❓ Questions & Discussions (`--question`)

Use this to ask questions or start discussions.

**Automatic Labels**: `question`, `help-wanted`

**Example**:
```bash
/alfred:9-help --question 'Should we migrate from Sequelize to Prisma ORM?'
```

---

## 🎯 Priority Levels

When you create an issue, you'll be prompted to select a priority:

| Level | Emoji | Label | When to Use |
|-------|-------|-------|------------|
| 🔴 Critical | 🔴 | `priority-critical` | System down, data loss risk, security breach |
| 🟠 High | 🟠 | `priority-high` | Major feature broken, significant impact |
| 🟡 Medium | 🟡 | `priority-medium` | Normal bugs, typical features, default |
| 🟢 Low | 🟢 | `priority-low` | Minor issues, nice-to-have features |

---

## 💡 Real-World Examples

### Example 1: Emergency Bug Report

**Scenario**: During production support, you find that users cannot reset their passwords.

```bash
/alfred:9-help --bug 'Password reset email not being sent after clicking "Forgot Password"'
```

**Result**: Issue #234 created and visible to team immediately with critical priority.

### Example 2: Feature Request from Code Review

**Scenario**: During code review, you think of a feature that would improve the codebase.

```bash
/alfred:9-help --feature 'Add request rate limiting middleware to prevent abuse'
```

**Result**: Issue #235 created for backlog planning.

### Example 3: Performance Improvement Suggestion

**Scenario**: You notice slow database queries in the user service.

```bash
/alfred:9-help --improvement 'Add database index on users.email for faster lookups'
```

**Result**: Issue #236 created for technical debt backlog.

### Example 4: Architecture Question

**Scenario**: You're uncertain about the best approach for API design.

```bash
/alfred:9-help --question 'Should we use REST or GraphQL for the new mobile API?'
```

**Result**: Issue #237 created for team discussion.

---

## 🔧 Prerequisites

### Required

1. **GitHub CLI installed**
   ```bash
   # macOS
   brew install gh

   # Ubuntu/Debian
   sudo apt install gh

   # Or visit: https://cli.github.com
   ```

2. **Authenticated with GitHub**
   ```bash
   gh auth login
   ```

3. **Git repository initialized**
   ```bash
   git init
   git remote add origin https://github.com/owner/repo.git
   ```

---

## 🔄 Related Commands

| Command | Purpose |
|---------|---------|
| `/alfred:0-project` | Initialize project |
| `/alfred:1-plan` | Create SPEC documents |
| `/alfred:2-run` | Implement features |
| `/alfred:3-sync` | Sync documentation |
| `/alfred:9-help` | **Create issues (this command)** |

---

## ✨ Summary

The `/alfred:9-help` command enables:

- ✅ **Fast issue creation** - Seconds, not minutes
- ✅ **Standardized format** - Consistent labels and metadata
- ✅ **Priority management** - Clear issue prioritization
- ✅ **Team visibility** - Issues immediately visible and discussable
- ✅ **Workflow integration** - Works with MoAI-ADK specs and planning

**Start using it now**:
```bash
/alfred:9-help --bug 'Describe the issue you just found'
```

Happy issue reporting! 🎉
