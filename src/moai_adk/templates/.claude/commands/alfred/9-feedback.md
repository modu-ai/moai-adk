---
name: alfred:9-feedback
description: "Create GitHub issues interactively"
allowed-tools:
- Bash(gh:*)
- AskUserQuestion
- Skill
skills:
- moai-alfred-issue-labels
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

Use AskUserQuestion with:

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
1. **Load label schema** via `Skill("moai-alfred-issue-labels")`
   - Resolves semantic label taxonomy
   - Maps type → labels (e.g., bug → "bug", "reported")
   - Maps priority → labels (e.g., high → "priority-high")
2. **Formats title with emoji**: "🐛 [BUG] Login button not responding..."
3. **Prepares body**: User description + creation timestamp + referenced from /alfred:9-feedback
4. **Executes gh CLI**:
   ```bash
   gh issue create \
     --title "🐛 [BUG] Login button not responding to clicks" \
     --body "When I click the login button on the homepage, nothing happens..." \
     --label "bug" \
     --label "reported" \
     --label "priority-high"
   ```
5. **Parses issue number** from response

**Label Mapping** (via `moai-alfred-issue-labels` skill):

| Type | Primary Labels | Priority | Final Labels |
|------|---|---|---|
| 🐛 Bug | bug, reported | High | bug, reported, priority-high |
| ✨ Feature | feature-request, enhancement | Medium | feature-request, enhancement, priority-medium |
| ⚡ Improvement | improvement, enhancement | Medium | improvement, enhancement, priority-medium |
| ❓ Question | question, help-wanted | Medium | question, help-wanted, priority-medium |

**Success Output**:
```
✅ GitHub Issue #234 created successfully!

📋 Title: 🐛 [BUG] Login button not responding to clicks
🔴 Priority: High
🏷️  Labels: bug, reported, priority-high (via moai-alfred-issue-labels)
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
- ✅ Load `Skill("moai-alfred-issue-labels")` to resolve semantic labels
- ✅ Apply labels from skill mapping (type + priority → labels)
- ✅ Show issue URL after creation with applied labels

### ❌ What NOT to Do

- ❌ Accept command arguments (`/alfred:9-feedback --bug` is wrong—just use `/alfred:9-feedback`)
- ❌ Skip questions or change order
- ❌ Rephrase user's input
- ❌ Create issues without labels (always use skill-based mapping)
- ❌ Hardcode label values (use skill mapping instead)

---

## 💡 Key Benefits

1. **🚀 No Arguments Needed**: Just `/alfred:9-feedback`
2. **💬 Conversational**: Intuitive step-by-step dialog
3. **🏷️ Semantic Labels**: Auto-labeled via `moai-alfred-issue-labels` skill
4. **🔗 Team Visible**: Issues immediately visible on GitHub
5. **⏱️ Fast**: Create issues in 30 seconds
6. **🔄 Reusable**: Label mapping shared with other commands (`/alfred:1-plan`, `/alfred:3-sync`)

---

**Supported since**: MoAI-ADK v0.7.0+
