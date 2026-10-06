---
description: Create SPEC document with EARS format requirements and acceptance criteria
argument-hint: "\"description\" [--branch] [--resume SPEC-XXX]"
allowed-tools: Skill
---

Dispatch the moai workflow `plan` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `plan` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `plan` subcommand with the same arguments
