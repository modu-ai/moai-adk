---
description: Identify and safely remove dead code with test verification
argument-hint: "[--dry] [--safe-only] [--file PATH]"
allowed-tools: Skill
---

Dispatch the moai workflow `clean` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `clean` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `clean` subcommand with the same arguments
