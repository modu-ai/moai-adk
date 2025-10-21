---

name: moai-alfred-tui-survey
description: Standardizes Claude Code Tools AskUserQuestion TUI menus for surveys, branching approvals, and option picking across Alfred workflows. Use when gathering approvals or decisions via Alfred’s TUI menus.
allowed-tools:
  - Read
  - Write
  - Edit
  - TodoWrite
---

# Alfred TUI Survey Skill

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), TodoWrite (todo_write) |
| Auto-load | On demand when AskUserQuestion menus are built |
| Trigger cues | Branch approvals, survey menus, decision gating via AskUserQuestion. |

## What it does

Provides ready-to-use patterns for Claude Code's AskUserQuestion TUI selector so Alfred agents can gather user choices, approvals, or survey answers with structured menus instead of ad-hoc text prompts.

## When to use

- Activates when Alfred needs to gather structured choices through the TUI selector.
- Need confirmation before advancing to a risky/destructive step.
- Choosing between alternative implementation paths or automation levels.
- Collecting survey-like answers (persona, tech stack, priority, risk level).
- Any time a branched workflow depends on user-selected options rather than free-form text.

## How it works

1. **Detect gate** – Pause at steps that require explicit user choice.
2. **Shape options** – Offer 2–5 focused choices with concise labels.
3. **Render menu** – Emit the `AskUserQuestion({...})` block for the selector.
4. **Map follow-ups** – Note how each option alters the next action/agent.
5. **Fallback** – Provide plain-text instructions when the UI cannot render.

### Templates & examples

- [Single-select deployment decision](examples.md#single-select-template)
- [Multi-select diagnostics checklist](examples.md#multi-select-variation)
- [Follow-up drill-down prompt](examples.md#follow-up-prompt-for-deeper-detail)

## Best Practices
- Reduce context switching by putting questions, options, and follow-up actions on one screen.
- Options are sorted according to comparison criteria such as risk and priority.
- If safety measures are required, such as approval/suspension/cancellation, attach a warning message.
- Submission results are recorded for reuse in the Sync step or reports.

## Inputs
- Decision scenarios and candidate options collected by Alfred workflows.
- Definition of follow-up actions for each option (next command, subagent, TODO, etc.).

## Outputs
- AskUserQuestion block and choice → follow-up action mapping.
- Summary of selection results and TODO items that require further confirmation.

## Failure Modes
- Ambiguous or duplicate options cause selection delays.
- It is difficult to judge priorities when all options look the same.
- AskUserQuestion does not work in a TUI-inactive environment, so plan B is needed.

## Dependencies
- Works with the main `/alfred:*` command and schedules follow-up actions with TodoWrite when necessary.
- When combined with moai-foundation-ears · moai-foundation-tags, requirements → options → TAG records are connected.

## References
- Jakubovic, J. "Designing Effective CLI Dialogs." ACM Queue, 2021.
- NCurses. "Programming Guide." https://invisible-island.net/ncurses/man/ncurses.3x.html (accessed 2025-03-29).

## Changelog
- 2025-03-29: Added input/output/failure response to Alfred-specific skills.

## Works well with

- `moai-foundation-ears` – Combine structured requirement patterns with menu-driven confirmations.
- `moai-alfred-git-workflow` – Use menus to choose branch/worktree strategies.
- `moai-alfred-code-reviewer` – Capture reviewer focus areas through guided selection.

## Examples
```markdown
- In the `/alfred:1-plan` step, user priorities are collected and the results are written to the PLAN board.
- Check user approval before high-risk operations while running `/alfred:2-run`.
```
