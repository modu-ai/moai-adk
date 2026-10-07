---
description: 하네스 학습 서브시스템 관리(status/apply/rollback/disable) 또는 자연어 분석으로 하네스 생성
argument-hint: "{status|apply|rollback <YYYY-MM-DD>|disable|list|edit <name>|remove <name>|doctor|<natural-language request>}"
allowed-tools: Skill
---

Dispatch the moai workflow `harness` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `harness` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `~/.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `harness` subcommand with the same arguments
