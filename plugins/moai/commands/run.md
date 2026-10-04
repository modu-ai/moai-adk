---
description: Implement SPEC requirements using DDD/TDD methodology
argument-hint: "SPEC-XXX [--team] [--resume SPEC-XXX]"
allowed-tools: Skill
---

Dispatch the moai workflow `run` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `run` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `run` subcommand with the same arguments
