---
description: lint+format+type-check+test 병렬 실행 (pre-commit 품질 게이트)
argument-hint: "[--fix]"
allowed-tools: Skill
---

Dispatch the moai workflow `gate` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `gate` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `~/.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `gate` subcommand with the same arguments
