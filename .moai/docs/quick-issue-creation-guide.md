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

**Creates**:
- 🐛 [BUG] Payment form crashes when credit card has < 4 digits
- Labels: bug, reported, priority-{your-selection}

---

### ✨ Feature Requests (`--feature`)

Use this to propose new functionality.

**Automatic Labels**: `feature-request`, `enhancement`

**Example**:
```bash
/alfred:9-help --feature 'Add webhook support for payment notifications'
```

**Creates**:
- ✨ [FEATURE] Add webhook support for payment notifications
- Labels: feature-request, enhancement, priority-{your-selection}

---

### ⚡ Improvement Suggestions (`--improvement`)

Use this to suggest enhancements to existing features.

**Automatic Labels**: `improvement`, `enhancement`

**Example**:
```bash
/alfred:9-help --improvement 'Reduce database queries in checkout process by 50%'
```

**Creates**:
- ⚡ [IMPROVEMENT] Reduce database queries in checkout process by 50%
- Labels: improvement, enhancement, priority-{your-selection}

---

### ❓ Questions & Discussions (`--question`)

Use this to ask questions or start discussions.

**Automatic Labels**: `question`, `help-wanted`

**Example**:
```bash
/alfred:9-help --question 'Should we migrate from Sequelize to Prisma ORM?'
```

**Creates**:
- ❓ [QUESTION] Should we migrate from Sequelize to Prisma ORM?
- Labels: question, help-wanted, priority-{your-selection}

---

## 🎯 Priority Levels

When you create an issue, you'll be prompted to select a priority:

| Level | Emoji | Label | When to Use |
|-------|-------|-------|------------|
| 🔴 Critical | 🔴 | `priority-critical` | System down, data loss risk, security breach |
| 🟠 High | 🟠 | `priority-high` | Major feature broken, significant impact |
| 🟡 Medium | 🟡 | `priority-medium` | Normal bugs, typical features, default |
| 🟢 Low | 🟢 | `priority-low` | Minor issues, nice-to-have features |

**Example Priority Selection**:
```
Priority Level?
[ ] 🔴 Critical - Blocking, urgent
[ ] 🟠 High - Important, should fix soon
[✓] 🟡 Medium - Normal priority (default)
[ ] 🟢 Low - Nice to have, can wait
```

---

## 💡 Real-World Examples

### Example 1: Emergency Bug Report

**Scenario**: During production support, you find that users cannot reset their passwords.

```bash
/alfred:9-help --bug 'Password reset email not being sent after clicking "Forgot Password"'
```

1. Alfred parses the bug report
2. You select: **Critical** priority
3. Result:
   - 🐛 [BUG] Password reset email not being sent after clicking "Forgot Password"
   - Labels: bug, reported, priority-critical
   - Issue #234 created and visible to team immediately

**Benefits**:
- Team is instantly aware of the issue
- Issue is prioritized as critical
- Link can be shared in Slack/Discord
- No context switching from development

---

### Example 2: Feature Request from Code Review

**Scenario**: During code review, you think of a feature that would improve the codebase.

```bash
/alfred:9-help --feature 'Add request rate limiting middleware to prevent abuse'
```

1. Alfred parses the feature request
2. You select: **High** priority
3. Result:
   - ✨ [FEATURE] Add request rate limiting middleware to prevent abuse
   - Labels: feature-request, enhancement, priority-high
   - Issue #235 created for backlog planning

**Benefits**:
- Idea is captured immediately
- Can be discussed in issue thread
- Team can estimate and prioritize
- Linked to code review discussion

---

### Example 3: Performance Improvement Suggestion

**Scenario**: You notice slow database queries in the user service.

```bash
/alfred:9-help --improvement 'Add database index on users.email for faster lookups'
```

1. Alfred parses the improvement
2. You select: **Medium** priority
3. Result:
   - ⚡ [IMPROVEMENT] Add database index on users.email for faster lookups
   - Labels: improvement, enhancement, priority-medium
   - Issue #236 created for technical debt backlog

---

### Example 4: Architecture Question

**Scenario**: You're uncertain about the best approach for API design.

```bash
/alfred:9-help --question 'Should we use REST or GraphQL for the new mobile API?'
```

1. Alfred parses the question
2. You select: **High** priority (for decision needed)
3. Result:
   - ❓ [QUESTION] Should we use REST or GraphQL for the new mobile API?
   - Labels: question, help-wanted, priority-high
   - Issue #237 created for team discussion

**Benefits**:
- Question is visible to entire team
- Architectural decision is documented
- Can link to related specs and ADRs

---

## 🔗 Integration with MoAI-ADK Workflow

### Linking to SPECs

When an issue is related to a SPEC document, reference it:

**In the SPEC document** (`.moai/specs/SPEC-API-001/spec.md`):
```markdown
# @SPEC:API-001 API Rate Limiting

## Related GitHub Issues
- See [Issue #235 - Request rate limiting middleware](https://github.com/owner/repo/issues/235) for discussion
```

**In the GitHub Issue**:
```markdown
## Related SPEC
- See [SPEC-API-001](https://github.com/owner/repo/blob/main/.moai/specs/SPEC-API-001/spec.md) for requirements
```

### Linking to ADRs

For architectural decisions:

**Create SPEC document**:
```bash
/alfred:1-plan "API Design Decision - REST vs GraphQL"
```

**Then create issue for discussion**:
```bash
/alfred:9-help --question 'Should we use REST or GraphQL for the new mobile API?'
```

**Link them in the SPEC**:
```markdown
## Discussion
- GitHub Issue #237 contains team discussion and decision rationale
```

---

## ⚡ Pro Tips & Best Practices

### ✅ Do This

1. **Be specific and detailed**
   ```bash
   # ✅ Good: Specific problem with context
   /alfred:9-help --bug 'Login API returns 500 error when email has + symbol'

   # ❌ Bad: Too vague
   /alfred:9-help --bug 'Login broken'
   ```

2. **Use realistic priority levels**
   ```bash
   # ✅ Good: Accurate priority
   /alfred:9-help --bug 'Typo in README file'
   # User selects: Low priority

   # ❌ Bad: Inflated priority
   /alfred:9-help --bug 'Typo in README file'
   # User selects: Critical (not appropriate)
   ```

3. **Include reproduction steps for bugs**
   ```bash
   /alfred:9-help --bug '1. Open Chrome 2. Go to example.com 3. Click signup button 4. Error occurs'
   ```

4. **Reference related work**
   ```bash
   /alfred:9-help --feature 'Add 2FA support (related to SPEC-SECURITY-001)'
   ```

### ❌ Don't Do This

1. **Multiple issues in one command**
   ```bash
   # ❌ Bad: Combine unrelated issues
   /alfred:9-help --bug 'Login broken and homepage is slow'

   # ✅ Good: Create separate issues
   /alfred:9-help --bug 'Login form returning 500 error'
   /alfred:9-help --improvement 'Homepage performance optimization needed'
   ```

2. **Emotional language**
   ```bash
   # ❌ Bad: Emotional
   /alfred:9-help --bug 'This code is awful and broken everywhere'

   # ✅ Good: Professional
   /alfred:9-help --bug 'Error handling in payment module needs refactoring'
   ```

3. **Duplicate existing issues**
   - Check GitHub Issues before creating
   - Reference existing issue if similar:
     ```bash
     # Check first: https://github.com/owner/repo/issues
     # If #456 exists for similar issue, comment there instead
     ```

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
   # Follow prompts to authenticate
   ```

3. **Git repository initialized**
   ```bash
   git init
   git remote add origin https://github.com/owner/repo.git
   ```

### Verify Setup

```bash
# Check GitHub CLI is ready
gh auth status

# Output should show:
# ✓ Logged in to github.com as your-username
# ✓ Git operations for github.com configured to use https protocol
# ✓ Token: ghu_... (with limited access)
```

---

## 🚨 Troubleshooting

### Problem: "GitHub CLI is not installed"

```bash
# Solution: Install GitHub CLI
brew install gh  # macOS
# or visit https://cli.github.com
```

### Problem: "GitHub CLI is not authenticated"

```bash
# Solution: Authenticate
gh auth login
# Choose: GitHub.com
# Choose: HTTPS
# Authorize browser login
```

### Problem: "Repository not found"

```bash
# Solution: Ensure you're in the correct directory
pwd  # Verify you're in the repo root

# Or check remote
git remote -v
# Should show: origin https://github.com/owner/repo.git
```

### Problem: Issue created but labels missing

```bash
# Solution: Check GitHub token permissions
gh auth status --show-token

# Ensure token has 'repo' and 'read:org' scopes
```

---

## 📊 Monitoring Created Issues

### View Your Recent Issues

```bash
# List 10 most recent issues
gh issue list --limit 10

# List only your issues
gh issue list --assignee @me

# List high-priority issues
gh issue list --label priority-high
```

### Close an Issue

If created by mistake:

```bash
# Close issue #456
gh issue close 456

# Close with comment
gh issue close 456 -c "Duplicate of #123"
```

---

## 🔄 Related Commands

| Command | Purpose | Related to |
|---------|---------|-----------|
| `/alfred:0-project` | Initialize project | Creates foundation for issues |
| `/alfred:1-plan` | Create SPEC documents | Can be linked to issues |
| `/alfred:2-run` | Implement features | Issues can reference specs |
| `/alfred:3-sync` | Sync documentation | Update issue links in docs |
| `/alfred:9-help` | **Create issues** | Quick feedback loop |

---

## 📚 Additional Resources

- **CLAUDE.md**: MoAI-ADK project guidance
- **`.moai/memory/issue-label-mapping.md`**: Complete label configuration
- **GitHub Docs**: [Creating issues](https://docs.github.com/en/issues/tracking-your-work-with-issues/creating-an-issue)
- **GitHub CLI Docs**: [Issue commands](https://cli.github.com/manual/gh_issue)

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
