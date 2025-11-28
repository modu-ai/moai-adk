#!/usr/bin/env python3
"""
Generate Nextra MDX pages for all MoAI-ADK commands.

This script:
1. Reads command files from .claude/commands/moai/
2. Converts them to Nextra-compatible MDX format
3. Adds usage examples and related resources
4. Generates pages/commands/*.mdx files
"""

import re
from pathlib import Path


COMMANDS_DIR = Path("/Users/goos/worktrees/MoAI-ADK/SPEC-NEXTRA-001/.claude/commands/moai")
OUTPUT_DIR = Path("/Users/goos/worktrees/MoAI-ADK/SPEC-NEXTRA-001/docs/pages/commands")


def convert_command_to_mdx(command_file: Path) -> str:
    """Convert command markdown to Nextra MDX format."""

    with open(command_file, 'r', encoding='utf-8') as f:
        content = f.read()

    command_name = command_file.stem
    command_display = f"/moai:{command_name}"

    # Extract title
    title_match = re.search(r'^#\s+(.+)$', content, re.MULTILINE)
    title = title_match.group(1) if title_match else command_display

    # Add MDX frontmatter and formatting
    mdx = f"""---
title: {title}
description: {command_display} - MoAI-ADK Core Command
---

{content}

---

## Related Commands

{"- [/moai:0-project](/commands/moai-0-project) - Initialize new project" if command_name != "0-project" else ""}
{"- [/moai:1-plan](/commands/moai-1-plan) - Generate SPEC document" if command_name != "1-plan" else ""}
{"- [/moai:2-run](/commands/moai-2-run) - Execute TDD implementation" if command_name != "2-run" else ""}
{"- [/moai:3-sync](/commands/moai-3-sync) - Sync documentation" if command_name != "3-sync" else ""}
{"- [/moai:9-feedback](/commands/moai-9-feedback) - Submit feedback" if command_name != "9-feedback" else ""}
- [/clear](/commands/clear) - Reset context window

## Related Skills

- [moai-foundation-core](/skills/moai-foundation-core) - Core execution framework
- [moai-workflow-docs](/skills/moai-workflow-docs) - Documentation workflow
- [moai-foundation-quality](/skills/moai-foundation-quality) - Quality assurance

## See Also

- [Agent Guide](/advanced/agents-guide) - Agent delegation patterns
- [SPEC Format](/core-concepts/spec-format) - EARS specification format
- [TDD Workflow](/core-concepts/workflow) - Development workflow

---

**Last Updated**: 2025-11-28
**Command Type**: Core Workflow
**Version**: 2.0
"""

    return mdx


def generate_all_commands():
    """Generate MDX pages for all commands (except 99-release)."""

    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    commands_to_process = [
        "0-project.md",
        "1-plan.md",
        "2-run.md",
        "3-sync.md",
        "9-feedback.md"
    ]

    generated_count = 0

    for command_file_name in commands_to_process:
        command_file = COMMANDS_DIR / command_file_name

        if not command_file.exists():
            print(f"⚠️  Command file not found: {command_file_name}")
            continue

        print(f"Processing {command_file_name}...")

        mdx_content = convert_command_to_mdx(command_file)

        output_name = f"moai-{command_file.stem}.mdx"
        output_file = OUTPUT_DIR / output_name

        with open(output_file, 'w', encoding='utf-8') as f:
            f.write(mdx_content)

        print(f"  ✅ Generated {output_name} ({len(mdx_content)} chars)")
        generated_count += 1

    # Also create /clear command page
    clear_mdx = """---
title: /clear - Context Window Reset
description: Reset the conversation context to free up tokens
---

# /clear - Context Window Reset

Reset the conversation context window to free up tokens and start fresh. This is a critical command for managing the 200K token budget in long sessions.

## Overview

The `/clear` command resets Alfred's conversation context, clearing all previous messages and loaded skills while preserving essential session state. This is necessary when:

- Context approaches 180K tokens (90% of 200K limit)
- Starting a new major task or SPEC
- Experiencing performance degradation
- Need to reset after complex multi-agent workflows

## Syntax

```bash
/clear
```

**Parameters**: None (this command takes no arguments)

## When to Use

### Automatic Triggers

Alfred will suggest `/clear` when:

- Context usage > 180K tokens (critical threshold)
- Multiple large skills are loaded
- Complex multi-agent coordination completes
- Session spans multiple hours

### Manual Triggers

Execute `/clear` when:

- Starting a new SPEC implementation
- Switching between unrelated tasks
- Before major architectural work
- After completing Phase completion (plan → run → sync)

## What Gets Preserved

The `/clear` command preserves:

✅ Session metadata (user preferences, config)
✅ Last session state (.moai/memory/last-session-state.json)
✅ Current working directory
✅ Git branch and status
✅ Project configuration

## What Gets Cleared

The `/clear` command removes:

❌ All conversation history
❌ Loaded skills and their content
❌ Agent delegation history
❌ Cached analysis results
❌ Intermediate work state

## Usage Examples

### Example 1: After SPEC Completion

```bash
# Complete SPEC-001 implementation
/moai:3-sync SPEC-001

# Clear context before starting new work
/clear

# Start fresh with SPEC-002
/moai:1-plan "new feature description"
```

### Example 2: Token Budget Management

```bash
# Alfred warns: "Context at 185K tokens"
/clear

# Continue previous work with fresh context
# Alfred loads from last-session-state.json
```

### Example 3: Workflow Transition

```bash
# Finish backend development work
Task(subagent_type="code-backend", prompt="Complete API implementation")

# Clear context before switching to frontend
/clear

# Start frontend work with fresh context
Task(subagent_type="code-frontend", prompt="Build UI components")
```

## Best Practices

### DO
✅ Execute `/clear` when context > 180K tokens
✅ Clear after completing major workflow phases
✅ Clear between unrelated tasks
✅ Clear before complex multi-agent orchestration
✅ Clear at natural workflow boundaries

### DON'T
❌ Clear in the middle of active agent coordination
❌ Clear during TDD RED-GREEN-REFACTOR cycle
❌ Clear when Alfred is mid-analysis
❌ Overuse (only when necessary)
❌ Clear without checking current context usage

## Advanced Usage

### Aggressive Clear Strategy

For maximum token efficiency:

```python
# Clear after every major phase
/moai:1-plan "feature"  # ~30K tokens
/clear                  # Reset to 0

/moai:2-run SPEC-001    # ~50K tokens
/clear                  # Reset to 0

/moai:3-sync SPEC-001   # ~20K tokens
/clear                  # Reset to 0
```

### Context Engineering

Strategic clearing for multi-task workflows:

```bash
# Task 1: Backend API
Task(subagent_type="code-backend", prompt="Build auth API")
/clear  # Independent 200K context

# Task 2: Frontend UI (new 200K context)
Task(subagent_type="code-frontend", prompt="Build login UI")
/clear

# Task 3: Integration (new 200K context)
Task(subagent_type="workflow-tdd", prompt="Test integration")
```

## Troubleshooting

### Issue: Work Lost After Clear

**Symptom**: Previous context lost after `/clear`

**Solution**:
- Alfred auto-saves to `.moai/memory/last-session-state.json`
- Check session state file for restoration
- Use Git commits to preserve code changes
- Use `/moai:3-sync` before clearing to save docs

### Issue: Unclear When to Clear

**Symptom**: Unsure if context is too large

**Solution**:
```bash
# Alfred will warn you automatically
# Look for messages like:
# "Context usage: 185K/200K tokens (92%)"
# "Recommend executing /clear"

# Or track manually during complex work
```

### Issue: Clearing Too Frequently

**Symptom**: Constant clearing disrupts workflow

**Solution**:
- Only clear at natural boundaries (between phases)
- Let Alfred manage token budget automatically
- Use Task() delegation for separate contexts
- Reserve `/clear` for critical thresholds

## Performance Impact

| Metric | Before /clear | After /clear |
|--------|---------------|--------------|
| Context Tokens | 180K-200K | 0K |
| Response Time | Slower | Faster |
| Token Budget | Limited | Full 200K |
| Skill Loading | Cached | Fresh |

## Related Commands

- [/moai:1-plan](/commands/moai-1-plan) - Often followed by `/clear`
- [/moai:2-run](/commands/moai-2-run) - Clear before starting implementation
- [/moai:3-sync](/commands/moai-3-sync) - Clear after documentation sync

## Related Skills

- [moai-foundation-context](/skills/moai-foundation-context) - Context optimization strategies
- [moai-foundation-core](/skills/moai-foundation-core) - Token budget management

---

**Last Updated**: 2025-11-28
**Command Type**: System Management
**Version**: 2.0
"""

    clear_file = OUTPUT_DIR / "clear.mdx"
    with open(clear_file, 'w', encoding='utf-8') as f:
        f.write(clear_mdx)

    print(f"  ✅ Generated clear.mdx ({len(clear_mdx)} chars)")
    generated_count += 1

    print(f"\n{'='*60}")
    print(f"Command pages generated!")
    print(f"  ✅ Total: {generated_count} command pages")
    print(f"{'='*60}")


if __name__ == "__main__":
    generate_all_commands()
