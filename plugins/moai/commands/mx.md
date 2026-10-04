---
description: Scan codebase and add @MX code-level annotations for AI context
argument-hint: "[--all] [--dry] [--priority P1-P4] [--force] [--team]"
allowed-tools: Skill
---

Dispatch the moai workflow `mx` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `mx` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `mx` subcommand with the same arguments
