---
name: moai-reference-global-flags
description: >
  Global flags reference for all MoAI subcommands. This document defines
  standardized flags that apply across multiple workflows.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "reference"
  status: "active"
  updated: "2026-02-26"
  tags: "flags, reference, standardization"

# Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 3000

# Triggers
triggers:
  keywords: ["flags", "global", "options", "arguments"]
  agents: []
  phases: []
---

# Global Flags Reference

Standardized flags that apply across multiple MoAI subcommands.

## Execution Mode Flags (Mutually Exclusive)

These flags control how MoAI executes workflows and are mutually exclusive.

### `--team`
Force Agent Teams mode for parallel execution.

**Prerequisites:**
- `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` in settings.json env
- `workflow.team.enabled: true` in `.moai/config/sections/workflow.yaml`

**Fallback Behavior:**
If prerequisites are not met:
- Warn user about team mode unavailability
- Continue in sub-agent mode (solo)
- No data loss or state corruption

**Available in:**
- plan, run, sync, fix, loop, mx, review, moai (default)

### `--solo`
Force sub-agent mode (single agent per phase).

Disables complexity-based auto-selection and ensures single-agent execution.

**Available in:**
- moai (default)

### No flag (default)
System auto-selects based on complexity thresholds:
- Domains >= 3 OR Files >= 10 OR Complexity score >= 7 → Team mode
- Otherwise → Sub-agent mode

---

## Common Utility Flags

### `--resume <ID>`
Resume from a previous checkpoint or state.

**Parameter Format:** `--resume <ID>`

The `<ID>` parameter varies by command:
- `fix`, `loop`: Snapshot ID (e.g., `latest`, `iteration-001`, `memory-pressure`)
- `plan`, `run`, `moai`: SPEC ID (e.g., `SPEC-AUTH-001`)
- `context`: No additional parameter (uses context search)

**Examples:**
```bash
/moai:fix --resume latest
/moai:loop --resume iteration-001
/moai:run --resume SPEC-AUTH-001
```

**Available in:**
- fix, loop, context, moai, plan, run

### `--sequential` / `--seq`
Execute tasks sequentially instead of in parallel.

Useful for debugging or when parallel execution causes issues.

**Alias:** `--seq` (short form)

**Available in:**
- fix, loop

### `--dry-run` / `--dry`
Preview mode - show what would be done without making changes.

**Alias:** `--dry` (short form)

**Available in:**
- fix, clean, mx

---

## Development Mode Flags

### `--deepthink`
Activate Sequential Thinking MCP for deep analysis.

Used for complex problem analysis, architecture decisions, and technology trade-offs.

**Available in:**
- All commands (when supported)

---

## Flag Alias Patterns

MoAI uses consistent alias patterns across commands:

| Long Form | Short Form | Purpose |
|-----------|------------|---------|
| `--dry-run` | `--dry` | Preview mode |
| `--sequential` | `--seq` | Sequential execution |
| `--errors-only` | `--errors` | Fix errors only |
| `--max-iterations` | `--max` | Maximum iteration count |
| `--include-coverage` | `--coverage` | Include coverage checks |
| `--include-security` | `--security` | Include security checks |
| `--no-format` | `--no-fmt` | Skip formatting |
| `--resume-from` | `--resume` | Resume from checkpoint |

---

## Flag Standardization Rules

When adding new flags to MoAI workflows:

1. **Use established patterns**: Check if an existing flag pattern matches your use case
2. **Provide short aliases**: For flags longer than 15 characters, provide a short alias
3. **Document prerequisites**: For flags with prerequisites (like `--team`), document them clearly
4. **Specify fallback behavior**: Document what happens when prerequisites are not met
5. **Use angle brackets for required parameters**: `--resume <ID>`
6. **Use square brackets for optional parameters**: `--level <N>` (where N has a default)

---

## Version History

- v1.0.0 (2026-02-26): Initial standardization document
