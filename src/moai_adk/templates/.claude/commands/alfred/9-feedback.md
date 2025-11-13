---
name: alfred:9-feedback
description: "Quickly create GitHub issues (automatic information collection + templates)"
allowed-tools:
- Bash(gh:*)
- Bash(uv:*)
- AskUserQuestion
- Skill
skills:
- moai-alfred-issue-labels
- moai-alfred-feedback-templates
---

# 🎯 MoAI-ADK Alfred 9-Feedback: GitHub Issue Quick Creation Tool

> **Purpose**: Record bugs, feature requests, improvement suggestions, and questions quickly and accurately on GitHub.

## 📋 Command Purpose

Enables developers to immediately record bugs or ideas as GitHub issues when discovered.

- ✅ **Fast**: Complete issue creation in 2-3 steps
- ✅ **Accurate**: Automatically collect version and environment information
- ✅ **Organized**: Structured templates by label
- ✅ **Simple**: Just run the command (`/alfred:9-feedback`)

**How to use**:
```bash
/alfred:9-feedback
```

Done!

---

## 🚀 Execution Process (2 Steps)

### Step 1: Execute Command
```bash
/alfred:9-feedback
```

Just enter this, and Alfred handles the rest.

---

### Step 2: Collect Required Information at Once (AskUserQuestion - multiSelect)

**With a single question**, select all of the following:

```
┌─ Issue Type (required, single selection)
│  ├─ 🐛 Bug Report - Problem occurred
│  ├─ ✨ Feature Request - Propose new feature
│  ├─ ⚡ Improvement - Improve existing feature
│  ├─ 📚 Documentation - Improve documentation
│  ├─ 🔄 Refactoring - Improve code structure
│  └─ ❓ Question - Ask the team
│
├─ Priority (default: medium)
│  ├─ 🔴 Critical - System down, data loss
│  ├─ 🟠 High - Major feature failure
│  ├─ 🟡 Medium - General priority
│  └─ 🟢 Low - Can be done later
│
└─ Template Selection (optional)
   ├─ ✅ Auto-generate Template (recommended)
   └─ 📝 Write Manually
```

---

### Step 3: Review & Fill Auto-Generated Template

Alfred automatically generates a template matching the selected issue type.

For example, when **Bug Report** is selected:

```markdown
## Bug Description

[Space for user input]

## Steps to Reproduce

1. [User input]
2. [User input]
3. [User input]

## Expected Behavior

[Space for user input]

## Actual Behavior

[Space for user input]

## Environment Information

🔍 Automatically collected information:
- MoAI-ADK version: 0.22.5
- Python version: 3.11.5
- OS: macOS 14.2
- Current branch: feature/SPEC-001
- Uncommitted changes: 3 files
```

Users only need to fill in the `[Space for user input]` sections.

---

Alfred automatically handles:

1. **Environment Information Collection** (`python3 .moai/scripts/feedback-collect-info.py`):
   - MoAI-ADK version
   - Python version, OS
   - Git status (current branch, uncommitted changes)
   - Current SPEC being worked on

2. **Label Mapping** (`Skill("moai-alfred-issue-labels")`):
   - Issue type → labels (e.g., bug → "bug", "reported")
   - Priority → labels (e.g., high → "priority-high")

3. **Auto-generate Title**: "🐛 [BUG] Bug description..."

4. **GitHub Issue Creation**:
   ```bash
   gh issue create \
     --title "🐛 [BUG] Bug description" \
     --body "## Bug Description\n...[template + environment info]..." \
     --label "bug" \
     --label "reported" \
     --label "priority-high"
   ```

5. **Display Result**:
   ```
   ✅ GitHub Issue #234 created successfully!

   📋 Title: 🐛 [BUG] Bug description
   🔴 Priority: High
   🏷️ Labels: bug, reported, priority-high
   🔗 URL: https://github.com/owner/repo/issues/234

   💡 Next: Reference this issue in commit messages or link to SPEC
   ```

---

## 📊 Label Mapping (via `Skill("moai-alfred-issue-labels")`)

| Type | Main Labels | Priority | Final Labels |
|------|-------------|----------|--------------|
| 🐛 Bug | bug, reported | High | bug, reported, priority-high |
| ✨ Feature | feature-request, enhancement | Medium | feature-request, enhancement, priority-medium |
| ⚡ Improvement | improvement, enhancement | Medium | improvement, enhancement, priority-medium |
| 📚 Documentation | documentation | Medium | documentation, priority-medium |
| 🔄 Refactoring | refactor | Medium | refactor, priority-medium |
| ❓ Question | question, help-wanted | Medium | question, help-wanted, priority-medium |

---

## ⚠️ Rules

### ✅ Must Do

- ✅ Collect required information at once with multiSelect (issue type, priority)
- ✅ Accurately preserve user input
- ✅ Execute auto-information collection script (`python3 .moai/scripts/feedback-collect-info.py`)
- ✅ Map labels with `Skill("moai-alfred-issue-labels")`
- ✅ Provide templates with `Skill("moai-alfred-feedback-templates")`
- ✅ Display Issue URL after creation

### ❌ Must Not Do

- ❌ Use command arguments (`/alfred:9-feedback --bug` is wrong → just use `/alfred:9-feedback`)
- ❌ Ask more than 4 questions
- ❌ Modify user input
- ❌ Create issues without labels
- ❌ Hard-code labels (use skill-based mapping)

---

## 💡 Key Advantages

1. **⚡ Fast**: Complete in 2-3 steps within 30 seconds
2. **🤖 Automated**: Automatically collect version and environment information
3. **📋 Accurate**: Structured templates by label
4. **🏷️ Meaningful**: Classification based on `moai-alfred-issue-labels` skill
5. **🔄 Reusable**: Share labels with `/alfred:1-plan`, `/alfred:3-sync`
6. **Multi-language**: All text written in user's conversation language

---

## 📝 Usage Example

**Step 1**: Execute command
```bash
/alfred:9-feedback
```

**Step 2**: Select required information
```
Issue Type: [🐛 Bug Report] selected
Priority: [🟠 High] selected
Template: [✅ Auto-generate] selected
```

**Step 3**: Fill template
```markdown
## Bug Description
Login button does not respond when clicked.

## Steps to Reproduce
1. Access homepage
2. Click login button in top right corner
3. No response

## Expected Behavior
Login modal should appear

## Actual Behavior
Nothing happens

## Environment Information
🔍 Automatically collected information:
- MoAI-ADK version: 0.22.5
- Python version: 3.11.5
- OS: macOS 14.2
```

**Result**: Issue #234 automatically created + URL displayed ✅

---

**Supported Version**: MoAI-ADK v0.22.5+
