---
name: {skill-name}
description: {What this skill does and when to use it - 200 characters max}
---

# {Skill Title}

> {One-sentence value proposition}

---

## 🎯 Purpose

{Explain what this skill does and why it's useful. Be specific about the problem it solves.}

**Problem**: {What problem does this address?}
**Solution**: {How does this skill solve it?}
**Benefit**: {What value does it provide?}

---

## 📋 When to Use

Claude automatically invokes this skill when:

- {Trigger condition 1 with keywords}
- {Trigger condition 2 with context}
- {Trigger condition 3 with workflow state}

**Manual invocation**:
```bash
# Example command or query
"{Example user request that would trigger this skill}"
```

---

## 💡 How It Works

### Core Functionality

{Detailed explanation of what the skill does step by step}

### Example Workflow

```
User Request
    ↓
Skill Activation
    ↓
{Step 1: ...}
    ↓
{Step 2: ...}
    ↓
Result Delivered
```

---

## 🔗 Integration

### Works Well With

- **{related-skill-1}**: {How they complement each other}
- **{related-skill-2}**: {Why use them together}

### MoAI-ADK Workflow

```
/alfred:1-plan → /alfred:2-run → /alfred:3-sync
                        ↑
                  This skill activates here
                  (when {condition})
```

---

## 📚 Examples

### Example 1: {Use Case Name}

**User Request**: "{Example user input}"

**Skill Response**: {What the skill does}

**Result**: {What the user gets}

### Example 2: {Another Use Case}

**User Request**: "{Another example input}"

**Skill Response**: {What happens}

**Result**: {Output}

---

## ⚠️ Best Practices

### Do's ✅

- {Recommended practice 1}
- {Recommended practice 2}

### Don'ts ❌

- {What to avoid 1}
- {What to avoid 2}

---

## 🛠️ Supporting Files (Optional)

```
.claude/skills/{skill-name}/
├── SKILL.md (this file)
├── examples/
│   └── example-1.md
├── scripts/
│   └── helper.py
└── templates/
    └── template.txt
```

{Explain any additional files if present}

---

**Version**: 1.0.0
**Last Updated**: {YYYY-MM-DD}
**Author**: @{YourGitHubID}
