---

name: moai-alfred-git-workflow
description: Automates Git operations with MoAI-ADK conventions (feature branch, locale-based TDD commits, Draft PR, PR Ready transition). Use when orchestrating GitFlow checkpoints, commits, or PR transitions.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred Git Workflow

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | /alfred:2-run Git automation |
| Trigger cues | Branch provisioning, commit batching, draft PR preparation within Alfred flows. |

## What it does

Automates Git operations following MoAI-ADK conventions: branch creation, locale-based TDD commits, Draft PR creation, and PR Ready transition.

## When to use

- Activates when Alfred must manage branches, commits, or PR transitions.
- “Create branch”, “Create PR”, “Create commit”
- Automatically invoked by `/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync`
- Git workflow automation needed

## How it works

**1. Branch Creation**:
```bash
git checkout develop
git checkout -b feature/SPEC-AUTH-001
```

**2. Locale-based TDD Commits**:
- **Korean (ko)**: 🔴 RED: [Test Description]
- **English (en)**: 🔴 RED: [Test description]
- **Japanese (ja)**: 🔴 RED: [テスト説明]
- **Chinese (zh)**: 🔴 RED: [测试说明]

Configured via `.moai/config.json`:
```json
{"project": {"locale": "ko"}}
```

**3. Draft PR Creation**:
Creates Draft PR with SPEC reference and test checklist.

**4. PR Ready Transition** (via `/alfred:3-sync`):
- Updates PR from Draft → Ready
- Adds quality gate checklist
- Verifies TRUST 5-principles

## Best Practices
- The text shown to the user is written using TUI/report expressions.
- When running the tool, a summary of commands and results are recorded.

## Examples
```markdown
- Call this skill inside the /alfred command to generate a report.
- Add summary to Completion Report.
```

## Inputs
- MoAI-ADK project context (`.moai/project/`, `.claude/` templates, etc.).
- Parameters passed from user commands or higher commands.

## Outputs
- Reports, checklists or recommendations for your Alfred workflow.
- Structured data for subsequent subagent calls.

## Failure Modes
- When required input documents are missing or permissions are limited.
- When disruptive changes are required without user approval.

## Dependencies
- Cooperation with higher-level agents such as cc-manager and project-manager is required.

## References
- Vincent Driessen. "A successful Git branching model." https://nvie.com/posts/a-successful-git-branching-model/ (accessed 2025-03-29).
- GitHub Docs. "GitHub Flow." https://docs.github.com/en/get-started/using-github/github-flow (accessed 2025-03-29).

## Changelog
- 2025-03-29: Added input/output/failure response to Alfred-specific skills.

## Works well with

- alfred-ears-authoring (SPEC ID-based branch naming)
- alfred-trust-validation (PR Ready quality check)
