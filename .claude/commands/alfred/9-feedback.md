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

## ⚠️ Important Rules

### ✅ What to Do

- ✅ Ask all 4 questions in sequence (type → title → description → priority)
- ✅ Preserve exact user wording in title and description
- ✅ Use AskUserQuestion for all user inputs
- ✅ Allow skipping description (optional field)
- ✅ Show issue URL after creation

### ❌ What NOT to Do

- ❌ Accept command arguments (`/alfred:9-feedback --bug` is wrong—just use `/alfred:9-feedback`)
- ❌ Skip questions or change order
- ❌ Rephrase user's input
- ❌ Create issues without labels

---

## 💡 Key Benefits

1. **🚀 No Arguments Needed**: Just `/alfred:9-feedback`
2. **💬 Conversational**: Intuitive step-by-step dialog
3. **🏷️ Auto-labeled**: Labels applied automatically
4. **🔗 Team Visible**: Issues immediately visible
5. **⏱️ Fast**: Create issues in 30 seconds

---

**Supported since**: MoAI-ADK v0.7.0+
