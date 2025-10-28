---
name: alfred:9-feedback
description: "Interactive GitHub Issue creation - Step-by-step dialog to create issues without command arguments"
allowed-tools:
- Bash(gh:*)
- Task
- AskUserQuestion
---

# 🎯 MoAI-ADK Alfred 9-Feedback: Interactive GitHub Issue Creation

> **Purpose**: Create GitHub Issues through an interactive multi-step dialog. Simple command → guided questions → automatic issue creation.

## 📋 Command Purpose

Enable developers to instantly report bugs, request features, suggest improvements, and ask questions through conversational dialogs. No command arguments needed—just run `/alfred:9-feedback` and answer questions.

**Command Format**:
```bash
/alfred:9-feedback
```

That's it! Alfred guides you through the rest.

---

## 🚀 Interactive Execution Flow

### Step 1: Start Command
```bash
/alfred:9-feedback
```

Alfred responds and proceeds to Step 2.

---

### Step 2: Select Issue Type (AskUserQuestion)

**Question**: "What type of issue do you want to create?"

**Options**:
```
[ ] 🐛 Bug Report - Something isn't working
[ ] ✨ Feature Request - Suggest new functionality
[ ] ⚡ Improvement - Enhance existing features
[ ] ❓ Question/Discussion - Ask the team
```

**User Selection**: Selects one (e.g., 🐛 Bug Report)

---

### Step 3: Enter Issue Title (AskUserQuestion)

**Question**: "What is the issue title? (Be concise)"

**Example Input**:
```
Login button on homepage not responding to clicks
```

---

### Step 4: Enter Description (AskUserQuestion)

**Question**: "Provide a detailed description (optional—press Enter to skip)"

**Example Input**:
```
When I click the login button on the homepage, nothing happens.
Tested on Chrome 120.0 on macOS 14.2.
Expected: Login modal should appear
Actual: No response
```

Or just press Enter to skip.

---

### Step 5: Select Priority (AskUserQuestion)

**Question**: "What's the priority level?"

**Options**:
```
[ ] 🔴 Critical - System down, data loss, security breach
[ ] 🟠 High - Major feature broken, significant impact
[✓] 🟡 Medium - Normal priority (default)
[ ] 🟢 Low - Minor issues, nice-to-have
```

**User Selection**: Selects priority (e.g., 🟠 High)

---

### Step 6: Create Issue (Automatic)

Alfred automatically:
1. Formats title with emoji: "🐛 [BUG] Login button not responding..."
2. Prepares body with user description + metadata
3. Assigns labels: bug, reported, priority-high
4. Executes: `gh issue create --title ... --body ... --label ...`
5. Parses issue number from response

**Success Output**:
```
✅ GitHub Issue #234 created successfully!

📋 Title: 🐛 [BUG] Login button not responding to clicks
🔴 Priority: High
🏷️  Labels: bug, reported, priority-high
🔗 URL: https://github.com/owner/repo/issues/234

💡 Next: Reference this issue in your commits or link to a SPEC document
```

---

## 📋 Issue Type → Label Mapping

| User Selection | Issue Type | Automatic Labels |
|---|---|---|
| 🐛 Bug Report | bug | `bug`, `reported` |
| ✨ Feature Request | feature | `feature-request`, `enhancement` |
| ⚡ Improvement | improvement | `improvement`, `enhancement` |
| ❓ Question/Discussion | question | `question`, `help-wanted` |

**Priority Label** (added based on Step 5):
- `priority-critical` for 🔴 Critical
- `priority-high` for 🟠 High
- `priority-medium` for 🟡 Medium
- `priority-low` for 🟢 Low

---

## ⚠️ Important Rules

### ✅ What to Do

- ✅ Ask all 5 questions in sequence (type → title → description → priority → confirm)
- ✅ Preserve exact user wording in title and description
- ✅ Use AskUserQuestion for all user inputs
- ✅ Allow skipping description (optional field)
- ✅ Default priority to Medium if user doesn't select
- ✅ Show issue URL after creation

### ❌ What NOT to Do

- ❌ Accept command arguments (`/alfred:9-help --bug` is wrong—just use `/alfred:9-help`)
- ❌ Skip questions or change order
- ❌ Rephrase user's input
- ❌ Create issues without labels
- ❌ Modify existing issues—only create new ones

---

## 💡 Real-World Examples

### Example 1: Bug Report Flow

```
User: /alfred:9-feedback

Alfred: 🎯 Quick GitHub Issue Creation - What type of issue?
User:   🐛 Bug Report

Alfred: What's the title?
User:   Login button crash on mobile

Alfred: Description? (Optional)
User:   Tapping login button on iPhone 15 freezes app for 5 seconds then crashes.
        iOS 17.2

Alfred: Priority?
User:   🟠 High

Alfred: ✅ Creating issue...
        ✅ GitHub Issue #234 created!
        📋 🐛 [BUG] Login button crash on mobile
        🟠 Priority: High
        🏷️  bug, reported, priority-high
        🔗 https://github.com/owner/repo/issues/234
```

### Example 2: Feature Request Flow

```
User: /alfred:9-feedback

Alfred: What type of issue?
User:   ✨ Feature Request

Alfred: What's the title?
User:   Add dark mode theme support

Alfred: Description? (Optional)
User:   Users are requesting a dark mode option.
        Should affect all pages and persist in user preferences.

Alfred: Priority?
User:   🟡 Medium

Alfred: ✅ Creating issue...
        ✅ GitHub Issue #235 created!
        📋 ✨ [FEATURE] Add dark mode theme support
        🟡 Priority: Medium
        🏷️  feature-request, enhancement, priority-medium
        🔗 https://github.com/owner/repo/issues/235
```

### Example 3: Quick Improvement (Skip Description)

```
User: /alfred:9-feedback

Alfred: What type of issue?
User:   ⚡ Improvement

Alfred: What's the title?
User:   Optimize N+1 query in user listing

Alfred: Description? (Optional)
User:   (Press Enter to skip)

Alfred: Priority?
User:   🟠 High

Alfred: ✅ Creating issue...
        ✅ GitHub Issue #236 created!
        📋 ⚡ [IMPROVEMENT] Optimize N+1 query in user listing
        🟠 Priority: High
        🏷️  improvement, enhancement, priority-high
        🔗 https://github.com/owner/repo/issues/236
```

---

## 🔧 Technical Implementation

### Dialog Questions (AskUserQuestion)

**Question 1** (multiSelect: false):
```
header: "Issue Type"
question: "What type of issue do you want to create?"
options: [
  { label: "🐛 Bug Report", description: "..." },
  { label: "✨ Feature Request", description: "..." },
  { label: "⚡ Improvement", description: "..." },
  { label: "❓ Question/Discussion", description: "..." }
]
```

**Question 2** (free text):
```
header: "Issue Title"
question: "What's the issue title?"
multiSelect: false
(Accept any text input)
```

**Question 3** (free text, optional):
```
header: "Description"
question: "Provide detailed description (press Enter to skip)"
multiSelect: false
(Accept any text input or empty)
```

**Question 4** (multiSelect: false):
```
header: "Priority"
question: "What's the priority level?"
options: [
  { label: "🔴 Critical", description: "..." },
  { label: "🟠 High", description: "..." },
  { label: "🟡 Medium", description: "..." },
  { label: "🟢 Low", description: "..." }
]
```

### Issue Creation Process

```python
# After all 4 questions:
1. Determine labels from issue_type
2. Determine priority_label from priority
3. Combine all labels: [type_label1, type_label2, priority_label]
4. Format title: "{emoji} [{TYPE}] {user_title}"
5. Format body: "{user_description}\n\n---\n\nType: {type}\nPriority: {priority}\nCreated via: /alfred:9-help"
6. Execute: gh issue create --title "{title}" --body "{body}" --label "{labels}"
7. Parse issue number from output URL
8. Display success with URL
```

---

## 🔄 Related Commands

- `/alfred:0-project` - Initialize project
- `/alfred:1-plan` - Create SPEC documents
- `/alfred:2-run` - Implement features (RED-GREEN-REFACTOR)
- `/alfred:3-sync` - Synchronize documentation
- `/alfred:9-feedback` - **Create issues interactively (this command)**

---

## 📚 Learn More

For detailed information:
- `.moai/docs/quick-issue-creation-guide.md` - Complete usage guide with examples
- `.moai/memory/issue-label-mapping.md` - Label configuration and setup
- `.moai/memory/interactive-dialogs.md` - AskUserQuestion best practices

---

## ✨ Key Benefits

1. **🚀 No Arguments Needed**: Just `/alfred:9-help`—Alfred asks what you need
2. **💬 Conversational**: Step-by-step dialog is intuitive and friendly
3. **🏷️ Auto-labeled**: Labels applied automatically based on selections
4. **🔗 Team Visible**: Issues appear immediately in GitHub for team coordination
5. **⏱️ Fast**: Create issues in 30 seconds without leaving your IDE

---

**Supported since**: MoAI-ADK v0.7.0+
**Updated**: 2025-10-29 (Interactive dialog version)
