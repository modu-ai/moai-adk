---
description: Scan codebase and generate architecture documentation in codemaps/
argument-hint: "[--force] [--area AREA]"
allowed-tools: Skill
---

Dispatch the moai workflow `codemaps` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `codemaps` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `codemaps` subcommand with the same arguments
