---
description: Collect feedback and create GitHub issue (bug report, feature request)
argument-hint: "[\"description\"]"
allowed-tools: Skill
---

Dispatch the moai workflow `feedback` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `feedback` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `feedback` subcommand with the same arguments
