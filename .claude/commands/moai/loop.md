---
description: 모든 문제가 해결되거나 최대 반복 횟수에 도달할 때까지 반복 수정
argument-hint: "[--max N] [--auto-fix] [--seq]"
allowed-tools: Skill
---

Dispatch the moai workflow `loop` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `loop` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `loop` subcommand with the same arguments
