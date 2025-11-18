# 📌 GitHub Issue Label Mapping Configuration

> **MoAI-ADK Label Management** - Centralized configuration for issue type labels and priority indicators

**Version**: 1.0.0
**Related**: `/alfred:9-feedback`

---

## 🏷️ Issue Type Label Mapping

### Bug Issues (`--bug`)

**Primary Labels**: `bug`, `reported`

**Optional Labels** (based on priority):
- `priority-critical` - System down, data loss risk
- `priority-high` - Major feature broken
- `priority-medium` - Normal bug
- `priority-low` - Minor issue

---

### Feature Request Issues (`--feature`)

**Primary Labels**: `feature-request`, `enhancement`

**Optional Labels** (based on priority):
- `priority-critical` - Blocking, must implement immediately
- `priority-high` - Important feature
- `priority-medium` - Normal priority feature (default)
- `priority-low` - Nice to have

---

### Improvement Issues (`--improvement`)

**Primary Labels**: `improvement`, `enhancement`

**Optional Labels** (based on priority):
- `priority-critical` - Critical refactoring needed
- `priority-high` - Important improvement
- `priority-medium` - Normal priority (default)
- `priority-low` - Technical debt, can wait

---

### Question/Discussion Issues (`--question`)

**Primary Labels**: `question`, `help-wanted`

**Optional Labels** (based on priority):
- `priority-critical` - Urgent decision needed
- `priority-high` - Important decision
- `priority-medium` - Normal discussion (default)
- `priority-low` - Optional discussion

---

## 🎯 Priority Emoji Mapping

| Priority | Emoji | Use Case |
|----------|-------|----------|
| Critical | 🔴 | System outage, data loss, security breach |
| High | 🟠 | Major feature broken, significant impact |
| Medium | 🟡 | Normal bugs/features (default) |
| Low | 🟢 | Minor issues, nice-to-have features |

---

## 🚀 Issue Type Emoji Mapping

| Type | Emoji | Description |
|------|-------|-------------|
| Bug | 🐛 | Defect or unexpected behavior |
| Feature | ✨ | New functionality or capability |
| Improvement | ⚡ | Code quality, performance, or design improvement |
| Question | ❓ | Question, discussion, or decision needed |

---

## 🛠️ GitHub Labels Setup

To create labels in your GitHub repository, use GitHub CLI:

```bash
# Authenticate first
gh auth login

# Create bug labels
gh label create "bug" \
  --description "Something isn't working" \
  --color "d73a49"

gh label create "reported" \
  --description "User-reported issue" \
  --color "fc2929"

# Create feature labels
gh label create "feature-request" \
  --description "New feature or request" \
  --color "a2eeef"

gh label create "enhancement" \
  --description "Improvement or enhancement" \
  --color "0075ca"

# Create improvement label
gh label create "improvement" \
  --description "Performance or code quality improvement" \
  --color "5ebcf6"

# Create question labels
gh label create "question" \
  --description "Question for discussion" \
  --color "fbca04"

gh label create "help-wanted" \
  --description "We need help with this" \
  --color "fcfc03"

# Create priority labels
gh label create "priority-critical" \
  --description "Critical priority - URGENT" \
  --color "ff0000"

gh label create "priority-high" \
  --description "High priority" \
  --color "ff6600"

gh label create "priority-medium" \
  --description "Medium priority" \
  --color "ffcc00"

gh label create "priority-low" \
  --description "Low priority" \
  --color "00cc00"
```

---

## ✨ Summary

This configuration enables:
- ✅ **Standardized issue labels** across all issue types
- ✅ **Priority indicators** with emoji for visual identification
- ✅ **Type emoji** for quick issue categorization
- ✅ **Automated label assignment** via `/alfred:9-feedback`

For more information, see `.moai/docs/quick-issue-creation-guide.md`
