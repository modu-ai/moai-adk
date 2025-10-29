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

## 🚀 Interactive Execution Flow (2-Phase Batched Design)

**UX Improvement**: Reduced from 4 sequential questions to 2 batched phases (50% turn reduction)

### Step 1: Start Command
```bash
/alfred:9-feedback
```

Alfred responds and proceeds to Phase 1.

---

### Step 2: Phase 1 - Metadata Collection (AskUserQuestion)

**Batch 2 Questions in 1 Turn**:

**Question 1**: "What type of issue do you want to create?"
**Options**:
```
[ ] 🐛 Bug Report - Something isn't working
[ ] ✨ Feature Request - Suggest new functionality
[ ] ⚡ Improvement - Enhance existing features
[ ] ❓ Question/Discussion - Ask the team
```

**Question 2**: "What's the priority level?"
**Options**:
```
[ ] 🔴 Critical - System down, data loss, security breach
[ ] 🟠 High - Major feature broken, significant impact
[✓] 🟡 Medium - Normal priority (default)
[ ] 🟢 Low - Minor issues, nice-to-have
```

**Example AskUserQuestion Call**:
```python
AskUserQuestion(
    questions=[
        {
            "question": "What type of issue do you want to create?",
            "options": [
                "🐛 Bug Report - Something isn't working",
                "✨ Feature Request - Suggest new functionality",
                "⚡ Improvement - Enhance existing features",
                "❓ Question/Discussion - Ask the team"
            ]
        },
        {
            "question": "What's the priority level?",
            "options": [
                "🔴 Critical - System down, data loss, security breach",
                "🟠 High - Major feature broken, significant impact",
                "🟡 Medium - Normal priority (default)",
                "🟢 Low - Minor issues, nice-to-have"
            ]
        }
    ]
)
```

**User Response Example**:
```
Selected: 🐛 Bug Report
Selected: 🟠 High
```

---

### Step 3: Phase 2 - Content Collection (AskUserQuestion)

**Batch 2 Questions in 1 Turn**:

**Question 1**: "What is the issue title?"
**Options**:
```
[ ] Other (type your custom title)
```

**Question 2**: "Provide a detailed description (optional)"
**Options**:
```
[ ] Skip description
[ ] Other (type your description)
```

**Example AskUserQuestion Call**:
```python
AskUserQuestion(
    questions=[
        {
            "question": "What is the issue title? (Be concise)",
            "options": [
                "Other (type your custom title)"
            ]
        },
        {
            "question": "Provide a detailed description (optional)",
            "options": [
                "Skip description",
                "Other (type your description)"
            ]
        }
    ]
)
```

**User Response Example**:
```
Other: Login button on homepage not responding to clicks
Other: When I click the login button on the homepage, nothing happens.
Tested on Chrome 120.0 on macOS 14.2.
Expected: Login modal should appear
Actual: No response
```

---

### Step 4: Parse Responses & Validate

**Extract from Phase 1**:
```python
# Parse issue type from first response
issue_type = parse_issue_type_from_label(phase1_response[0])
# Examples: "🐛 Bug Report" → "bug"

# Parse priority from second response
priority = parse_priority_from_label(phase1_response[1])
# Examples: "🟠 High" → "high"
```

**Extract from Phase 2**:
```python
# Parse title from first response (handle "Other" case)
title = validate_title(phase2_response[0])
# Examples: "Other: Login button..." → "Login button..."

# Parse description from second response (handle "Skip" case)
description = validate_description(phase2_response[1])
# Examples: "Skip description" → ""
#           "Other: When I click..." → "When I click..."
```

**Fallback Logic**:
```python
# Default to "Medium" if priority parsing fails
if not priority or priority not in ["critical", "high", "medium", "low"]:
    priority = "medium"

# Require title to be non-empty
if not title or title.strip() == "":
    raise ValueError("Issue title cannot be empty")

# Description is optional (empty string is valid)
description = description or ""
```

---

### Step 5: Create Issue (Automatic)

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

## 🔧 Helper Functions

### parse_issue_type_from_label(label: str) → str

Converts user-selected label to issue type identifier.

**Examples**:
```python
"🐛 Bug Report - Something isn't working" → "bug"
"✨ Feature Request - Suggest new functionality" → "feature"
"⚡ Improvement - Enhance existing features" → "improvement"
"❓ Question/Discussion - Ask the team" → "question"
```

**Implementation**:
```python
def parse_issue_type_from_label(label: str) -> str:
    if "🐛" in label or "Bug" in label:
        return "bug"
    elif "✨" in label or "Feature" in label:
        return "feature"
    elif "⚡" in label or "Improvement" in label:
        return "improvement"
    elif "❓" in label or "Question" in label:
        return "question"
    return "bug"  # Default fallback
```

---

### parse_priority_from_label(label: str) → str

Converts user-selected label to priority level.

**Examples**:
```python
"🔴 Critical - System down, data loss, security breach" → "critical"
"🟠 High - Major feature broken, significant impact" → "high"
"🟡 Medium - Normal priority (default)" → "medium"
"🟢 Low - Minor issues, nice-to-have" → "low"
```

**Implementation**:
```python
def parse_priority_from_label(label: str) -> str:
    if "🔴" in label or "Critical" in label:
        return "critical"
    elif "🟠" in label or "High" in label:
        return "high"
    elif "🟡" in label or "Medium" in label:
        return "medium"
    elif "🟢" in label or "Low" in label:
        return "low"
    return "medium"  # Default fallback
```

---

### validate_title(response: str) → str

Extracts title from user response, handling "Other" case.

**Examples**:
```python
"Other: Login button not responding" → "Login button not responding"
"Login button not responding" → "Login button not responding"
"" → raises ValueError
```

**Implementation**:
```python
def validate_title(response: str) -> str:
    # Remove "Other:" prefix if present
    title = response.replace("Other:", "").strip()

    if not title:
        raise ValueError("Issue title cannot be empty")

    return title
```

---

### validate_description(response: str) → str

Extracts description from user response, handling "Skip" and "Other" cases.

**Examples**:
```python
"Skip description" → ""
"Other: When I click the button..." → "When I click the button..."
"When I click the button..." → "When I click the button..."
"" → ""
```

**Implementation**:
```python
def validate_description(response: str) -> str:
    # Handle "Skip description" case
    if "Skip" in response or response.strip() == "":
        return ""

    # Remove "Other:" prefix if present
    description = response.replace("Other:", "").strip()

    return description
```

---

## ⚠️ Important Rules

### ✅ What to Do

- ✅ Use **batched AskUserQuestion** calls (2 phases instead of 4 sequential)
- ✅ Phase 1: Collect Type + Priority in 1 turn
- ✅ Phase 2: Collect Title + Description in 1 turn
- ✅ Use "Other" option for text input (Title, Description)
- ✅ Handle "Skip description" response for optional field
- ✅ Validate parsed values before issue creation
- ✅ Preserve exact user wording in title and description
- ✅ Show issue URL after creation

### ❌ What NOT to Do

- ❌ Accept command arguments (`/alfred:9-feedback --bug` is wrong—just use `/alfred:9-feedback`)
- ❌ Ask 4 sequential questions (old UX pattern—use batched 2-phase instead)
- ❌ Skip validation of parsed responses
- ❌ Rephrase user's input
- ❌ Create issues without labels
- ❌ Proceed with empty title (must validate)

---

## 💡 Key Benefits

1. **🚀 No Arguments Needed**: Just `/alfred:9-feedback`
2. **⚡ 50% Faster UX**: 2 turns instead of 4 (50% improvement)
3. **💬 Conversational**: Intuitive batched dialog
4. **🏷️ Auto-labeled**: Labels applied automatically
5. **🔗 Team Visible**: Issues immediately visible
6. **⏱️ Fast**: Create issues in 15 seconds (down from 30)

---

**Supported since**: MoAI-ADK v0.7.0+
