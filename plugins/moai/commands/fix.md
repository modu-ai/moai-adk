---
description: Auto-detect and fix LSP errors, linting issues, and type errors
argument-hint: "[--dry] [--seq] [--level N] [--resume] [--team]"
allowed-tools: Skill
---

Dispatch the moai workflow `fix` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `fix` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `fix` subcommand with the same arguments
