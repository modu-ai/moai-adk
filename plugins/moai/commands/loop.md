---
description: Iteratively fix issues until all resolved or max iterations reached
argument-hint: "[--max N] [--auto-fix] [--seq]"
allowed-tools: Skill
---

Dispatch the moai workflow `loop` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `loop` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `loop` subcommand with the same arguments
