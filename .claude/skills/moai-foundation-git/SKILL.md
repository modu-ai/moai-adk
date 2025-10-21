---

name: moai-foundation-git
description: Git workflow automation (branching, TDD commits, PR management). Use when standardizing Git practices across the project.
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# Alfred Git Workflow

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | SessionStart (foundation bootstrap) |
| Trigger cues | Branch creation, commit convention, PR readiness, and release gating requests. |

## What it does

Automates Git operations following MoAI-ADK conventions: branch creation, locale-based TDD commits, Draft PR creation, and PR Ready transition.

## When to use

- Activates when Git workflow automation is needed for branching, commits, or PR promotion.
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

## Examples
```markdown
- Scan standard documents and report missing sections.
- Changed regulations are reflected in CLAUDE.md.
```

## Best Practices
- When changing standards, the reason for change and supporting documents are recorded.
- Follow the single source principle and avoid modifying the same item in multiple places.

## Inputs
- Project standard documents (e.g. `CLAUDE.md`, `.moai/config.json`).
- Latest printouts from relevant sub-agents.

## Outputs
- Templates or policy summaries conforming to the MoAI-ADK standard.
- Reusable rules/checklists.

## Failure Modes
- When required standard files are missing or have limited access rights.
- When conflicting policies are detected and coordination is required.

## Dependencies
- There is great synergy when called together with cc-manager.

## References
- Vincent Driessen. "A successful Git branching model." https://nvie.com/posts/a-successful-git-branching-model/ (accessed 2025-03-29).
- GitHub Docs. "GitHub Flow." https://docs.github.com/en/get-started/using-github/github-flow (accessed 2025-03-29).

## Changelog
- 2025-03-29: Foundation skill templates have been enhanced to align with best practice structures.
