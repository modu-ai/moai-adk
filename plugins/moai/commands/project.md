---
description: Generate project documentation (product.md, structure.md, tech.md, codemaps/)
argument-hint: "[--force] [--area AREA]"
allowed-tools: Skill
---

Dispatch the moai workflow `project` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `project` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `project` subcommand with the same arguments
