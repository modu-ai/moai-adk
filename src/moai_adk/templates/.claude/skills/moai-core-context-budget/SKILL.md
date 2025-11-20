---
name: moai-core-context-budget
version: 4.0.0
updated: 2025-11-20
status: stable
tier: core
description: Context window management and token budget optimization
allowed-tools: [Read, Bash]
---

# Context Budget Manager

**Efficient Context Window Usage**

> **Focus**: Token Optimization, Memory Management  
> **Limits**: Claude 3.5 Sonnet (200K tokens)

---

## Overview

Strategies for managing limited context window in LLM conversations.

### Key Concepts

1.  **Token Budget**: Total available tokens (200K for Claude 3.5 Sonnet).
2.  **Context Allocation**: System prompt, conversation history, code files.
3.  **Optimization**: Minimize unnecessary tokens, prioritize critical info.

---

## Token Budget Breakdown

### Typical Allocation

```
Total: 200,000 tokens

System Prompt: ~5,000 tokens (2.5%)
- Agent instructions
- Skill descriptions
- Core rules

Conversation History: ~50,000 tokens (25%)
- Recent messages
- Summarized older context

Code & Files: ~140,000 tokens (70%)
- Viewed files
- Generated code
- Documentation

Reserved: ~5,000 tokens (2.5%)
- Response generation
```

---

## Optimization Strategies

### 1. Progressive File Viewing

View only necessary portions of large files.

```python
# Bad: View entire 10,000-line file
view_file("large_file.py")  # Uses ~50K tokens

# Good: View specific functions
view_file("large_file.py", start_line=100, end_line=200)  # ~5K tokens
```

### 2. Summarization

Summarize old conversation context.

**Technique**:

- Keep last 10 messages verbatim
- Summarize older messages into bullet points
- Discard very old, low-value exchanges

### 3. File Selection Priority

View files in order of importance:

1.  **Critical**: Core logic, main entry points
2.  **Important**: Related modules, tests
3.  **Optional**: Config files, documentation (defer if budget low)

### 4. Code Minimization

When sharing code for review, remove:

- Comments (if code is self-explanatory)
- Import statements (unless relevant)
- Blank lines

---

## Budget Monitoring

### Check Remaining Tokens

```python
# Estimate used tokens
used_tokens = (
    len(system_prompt) / 4 +  # ~4 chars per token
    len(conversation_history) / 4 +
    len(viewed_files) / 4
)

remaining_tokens = 200_000 - used_tokens
```

### Low Budget Actions

When <20% budget remains:

1.  **Summarize** conversation history
2.  **Close** unnecessary files
3.  **Defer** non-critical tasks
4.  **Split** task into smaller chunks

---

## Best Practices

1.  **View Incrementally**: Start with outlines (`view_file_outline`), then details.
2.  **Cache Key Info**: Reference discussed patterns instead of re-viewing files.
3.  **Batch Operations**: Group related file views to minimize overhead.
4.  **Prune History**: Periodically summarize old context.

---

## Validation Checklist

- [ ] **File Views**: Only necessary portions viewed?
- [ ] **History**: Old messages summarized?
- [ ] **Priority**: Critical files viewed first?
- [ ] **Monitoring**: Token usage tracked?

---

## Related Skills

- `moai-core-session-state`: Session management
- `moai-core-rules`: Core development rules

---

**Last Updated**: 2025-11-20
