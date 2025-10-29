# 📌 GitHub Issue Label Mapping Configuration

> **MoAI-ADK Label Management** - Centralized configuration for issue type labels and priority indicators

**Last Updated**: 2025-10-29
**Version**: 1.0.0
**Related**: `/alfred:9-feedback`, `src/moai_adk/core/issue_creator.py`

---

## 🏷️ Issue Type Label Mapping

### Bug Issues (`--bug`)

**Primary Labels**:
- `bug` - Indicates a defect or unexpected behavior
- `reported` - Indicates user-reported bug

**Optional Labels** (added based on priority):
- `priority-critical` - System down, data loss risk
- `priority-high` - Major feature broken
- `priority-medium` - Normal bug
- `priority-low` - Minor issue

**Label Colors**:
```json
{
  "bug": { "color": "d73a49", "description": "Something isn't working" },
  "reported": { "color": "fc2929", "description": "User-reported issue" },
  "priority-critical": { "color": "ff0000", "description": "Critical priority - URGENT" },
  "priority-high": { "color": "ff6600", "description": "High priority" },
  "priority-medium": { "color": "ffcc00", "description": "Medium priority" },
  "priority-low": { "color": "00cc00", "description": "Low priority" }
}
```

**Example Issue**:
```
Title: 🐛 [BUG] Login button not responding on homepage
Labels: bug, reported, priority-high
Color: Red/Orange
```

---

### Feature Request Issues (`--feature`)

**Primary Labels**:
- `feature-request` - New feature proposal
- `enhancement` - Feature addition or improvement

**Optional Labels** (added based on priority):
- `priority-critical` - Blocking, must implement immediately
- `priority-high` - Important feature
- `priority-medium` - Normal priority feature (default)
- `priority-low` - Nice to have

**Label Colors**:
```json
{
  "feature-request": { "color": "a2eeef", "description": "New feature or request" },
  "enhancement": { "color": "0075ca", "description": "Improvement or enhancement" },
  "priority-critical": { "color": "ff0000", "description": "Critical priority" },
  "priority-high": { "color": "ff6600", "description": "High priority" },
  "priority-medium": { "color": "ffcc00", "description": "Medium priority" },
  "priority-low": { "color": "00cc00", "description": "Low priority" }
}
```

**Example Issue**:
```
Title: ✨ [FEATURE] Add dark mode theme support
Labels: feature-request, enhancement, priority-high
Color: Cyan/Light Blue
```

---

### Improvement Issues (`--improvement`)

**Primary Labels**:
- `improvement` - Code quality, performance, or design improvement
- `enhancement` - General enhancement

**Optional Labels** (added based on priority):
- `priority-critical` - Critical refactoring needed
- `priority-high` - Important improvement
- `priority-medium` - Normal priority (default)
- `priority-low` - Technical debt, can wait

**Label Colors**:
```json
{
  "improvement": { "color": "5ebcf6", "description": "Performance or code quality improvement" },
  "enhancement": { "color": "0075ca", "description": "Improvement or enhancement" },
  "priority-critical": { "color": "ff0000", "description": "Critical priority" },
  "priority-high": { "color": "ff6600", "description": "High priority" },
  "priority-medium": { "color": "ffcc00", "description": "Medium priority" },
  "priority-low": { "color": "00cc00", "description": "Low priority" }
}
```

**Example Issue**:
```
Title: ⚡ [IMPROVEMENT] Optimize database queries in checkout
Labels: improvement, enhancement, priority-high
Color: Light Blue
```

---

### Question/Discussion Issues (`--question`)

**Primary Labels**:
- `question` - Question or discussion needed
- `help-wanted` - Help needed to resolve

**Optional Labels** (added based on priority):
- `priority-critical` - Urgent decision needed
- `priority-high` - Important decision
- `priority-medium` - Normal discussion (default)
- `priority-low` - Optional discussion

**Label Colors**:
```json
{
  "question": { "color": "fbca04", "description": "Question for discussion" },
  "help-wanted": { "color": "fcfc03", "description": "We need help with this" },
  "priority-critical": { "color": "ff0000", "description": "Critical priority" },
  "priority-high": { "color": "ff6600", "description": "High priority" },
  "priority-medium": { "color": "ffcc00", "description": "Medium priority" },
  "priority-low": { "color": "00cc00", "description": "Low priority" }
}
```

**Example Issue**:
```
Title: ❓ [QUESTION] Should we migrate from Sequelize to Prisma ORM?
Labels: question, help-wanted, priority-high
Color: Yellow
```

---

## 🎯 Priority Emoji Mapping

| Priority | Emoji | Hex Color | Use Case |
|----------|-------|-----------|----------|
| Critical | 🔴 | #ff0000 | System outage, data loss, security breach |
| High | 🟠 | #ff6600 | Major feature broken, significant impact |
| Medium | 🟡 | #ffcc00 | Normal bugs/features (default) |
| Low | 🟢 | #00cc00 | Minor issues, nice-to-have features |

---

## 🚀 Issue Type Emoji Mapping

| Type | Emoji | Description |
|------|-------|-------------|
| Bug | 🐛 | Defect or unexpected behavior |
| Feature | ✨ | New functionality or capability |
| Improvement | ⚡ | Code quality, performance, or design improvement |
| Question | ❓ | Question, discussion, or decision needed |

---

## 📋 Default Label Sets by Type

### Configuration Object

```javascript
const ISSUE_TYPE_LABELS = {
  BUG: ["bug", "reported"],
  FEATURE: ["feature-request", "enhancement"],
  IMPROVEMENT: ["improvement", "enhancement"],
  QUESTION: ["question", "help-wanted"]
}

const PRIORITY_LABELS = {
  CRITICAL: "priority-critical",
  HIGH: "priority-high",
  MEDIUM: "priority-medium",
  LOW: "priority-low"
}
```

### Label Assignment Algorithm

1. **Start with type labels** (always added)
   - Example: Bug → ["bug", "reported"]

2. **Add priority label** (based on user selection)
   - Example: High → "priority-high"

3. **Result**: ["bug", "reported", "priority-high"]

---

## 🔄 Label Workflow

### Creating an Issue

```
User Input:
  /alfred:9-feedback

↓

Parse Type:
  type = "BUG"
  labels = ["bug", "reported"]

↓

Ask Priority:
  Priority = "HIGH"
  labels += ["priority-high"]

↓

Final Labels:
  ["bug", "reported", "priority-high"]

↓

GitHub Issue Created:
  Title: 🐛 [BUG] Login form crashes
  Labels: bug, reported, priority-high
  Color: Red
```

---

## 📊 Label Statistics Configuration

### Recommended Label Counts per Issue

| Type | Optimal | Min | Max |
|------|---------|-----|-----|
| Bug | 3-4 | 2 | 5 |
| Feature | 3-4 | 2 | 5 |
| Improvement | 3-4 | 2 | 5 |
| Question | 3-4 | 2 | 5 |

**Format**: `[type1, type2, priority]`

---

## 🔗 Related Label Categories (Optional)

When using with category options, additional labels can be added:

```json
{
  "categories": {
    "Frontend": "category-frontend",
    "Backend": "category-backend",
    "API": "category-api",
    "Database": "category-database",
    "DevOps": "category-devops",
    "Security": "category-security"
  }
}
```

**Example with Category**:
```
Title: 🐛 [BUG] Login button not responding
Labels: bug, reported, priority-high, category-frontend
```

---

## 🛠️ GitHub Labels Setup Script

### Creating Labels in GitHub Repository

```bash
# Install GitHub CLI first
gh label create "bug" \
  --description "Something isn't working" \
  --color "d73a49"

gh label create "reported" \
  --description "User-reported issue" \
  --color "fc2929"

gh label create "feature-request" \
  --description "New feature or request" \
  --color "a2eeef"

gh label create "enhancement" \
  --description "Improvement or enhancement" \
  --color "0075ca"

gh label create "improvement" \
  --description "Performance or code quality improvement" \
  --color "5ebcf6"

gh label create "question" \
  --description "Question for discussion" \
  --color "fbca04"

gh label create "help-wanted" \
  --description "We need help with this" \
  --color "fcfc03"

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

### Verify Labels

```bash
# List all labels
gh label list

# Output should show:
# bug       Something isn't working           d73a49
# reported  User-reported issue               fc2929
# feature-request  New feature or request     a2eeef
# enhancement      Improvement or enhancement 0075ca
# ... (and others)
```

---

## 📖 Label Usage Examples

### Example 1: Critical Bug

```
Issue: #456
Title: 🐛 [BUG] Payment processing fails for all transactions
Labels: bug, reported, priority-critical
Priority: CRITICAL (🔴)
```

**Why these labels?**
- `bug` - Indicates defect
- `reported` - User-reported issue
- `priority-critical` - Requires immediate attention

---

### Example 2: Nice-to-Have Feature

```
Issue: #457
Title: ✨ [FEATURE] Add tooltips to form fields
Labels: feature-request, enhancement, priority-low
Priority: LOW (🟢)
```

**Why these labels?**
- `feature-request` - New functionality
- `enhancement` - Feature improvement
- `priority-low` - Can be deferred

---

### Example 3: Performance Improvement

```
Issue: #458
Title: ⚡ [IMPROVEMENT] Reduce API response time by 50%
Labels: improvement, enhancement, priority-high
Priority: HIGH (🟠)
```

**Why these labels?**
- `improvement` - Code quality/performance
- `enhancement` - System improvement
- `priority-high` - Important for user experience

---

## 🔄 Maintenance & Updates

### When to Update Labels

1. **New issue type needed**: Add to `ISSUE_TYPE_LABELS`
2. **New priority level**: Add to `PRIORITY_LABELS`
3. **Category expansion**: Add to categories if using

### Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2025-10-29 | Initial label mapping configuration |

---

## 📞 Support

For issues with label configuration:

1. Check GitHub repository settings: Settings → Labels
2. Verify label colors and descriptions match this document
3. Run label setup script if labels are missing
4. Contact team for label updates

---

**Related Files**:
- `src/moai_adk/core/issue_creator.py` - Python implementation
- `.claude/commands/alfred/9-feedback.md` - Command definition
- `.moai/docs/quick-issue-creation-guide.md` - User guide

**Next**: Review `.moai/docs/quick-issue-creation-guide.md` for complete usage instructions.
