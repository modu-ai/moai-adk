# Output styles

Adapt Claude Code for uses beyond software engineering

## Built-in output styles

- **Default**: Designed for efficient software engineering tasks
- **Explanatory**: Provides educational insights about the codebase along the way

Use `/output-style` to switch between them or create your own custom output style.

## Using output styles

### Switching output styles

To change your output style:

```bash
# Interactive selection
/output-style

# Direct style selection
/output-style explanatory
/output-style default
```

### How they work

Output styles modify the system prompt that guides Claude's behavior. Unlike other customization methods:

- **Output styles**: Replace the entire system prompt
- **CLAUDE.md**: Adds additional context as a user message
- **`--append-system-prompt`**: Appends to the default system prompt

## Creating custom output styles

You can create your own output styles to customize Claude's behavior for specific use cases.

### Interactive creation

Use the `/output-style:new` command:

```bash
/output-style:new I want an output style that acts as a patient tutor for learning Python
```

### Manual creation

Create a markdown file in your output styles directory:

**File structure:**
```
.claude/output-styles/tutor.md
```

**Content:**
```markdown
---
name: Python Tutor
description: Patient tutor for learning Python programming
---

You are a patient and encouraging Python tutor. Always:

- Explain concepts clearly with examples
- Ask if the student understands before moving on
- Provide practice exercises
- Celebrate progress and learning milestones

When helping with code:
- Show multiple ways to solve problems
- Explain why certain approaches are better
- Point out common pitfalls to avoid
```

### Storage locations

Output styles can be stored at different levels:

- **User-level**: `~/.claude/output-styles/` (available across all projects)
- **Project-level**: `.claude/output-styles/` (specific to current project)

Project-level styles take precedence over user-level styles.

## Comparison with related features

| Feature | Purpose | Scope |
|---------|---------|-------|
| **Output styles** | Complete behavior change | Entire session |
| **CLAUDE.md** | Project context and guidelines | Single project |
| **Custom slash commands** | Specific repeatable tasks | Individual commands |
| **`--append-system-prompt`** | Add to default prompt | Single session |
