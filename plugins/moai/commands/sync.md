---
description: Synchronize documentation, codemaps, and create pull request
argument-hint: "[SPEC-XXX] [--auto-merge] [--skip-mx]"
allowed-tools: Skill
---

Dispatch the moai workflow `sync` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `sync` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `sync` subcommand with the same arguments
