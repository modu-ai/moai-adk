---
description: Code review with security and @MX tag compliance check
argument-hint: "[--staged] [--branch] [--security] [--deep] [--patch] [--commit <SHA>]"
allowed-tools: Skill
---

Dispatch the moai workflow `review` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `review` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `review` subcommand with the same arguments
