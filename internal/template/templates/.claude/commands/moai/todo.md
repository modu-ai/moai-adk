---
description: Compatibility alias for the canonical GTD task-management entry point
argument-hint: "[\"<description>\"|list|next|done <n>]"
allowed-tools: Skill
---

Dispatch the moai workflow `gtd` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `gtd` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `gtd` subcommand with the same arguments
