---
description: 완료 조건 목표를 관리하거나 승인 기반 자율 미션을 생성·운영
argument-hint: "{\"<condition>\"|--auto \"<mission>\"|approve|run|status [--all]|revoke|resume|clear|render}"
allowed-tools: Skill
---

Dispatch the moai workflow `goal` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `goal` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `goal` subcommand with the same arguments
