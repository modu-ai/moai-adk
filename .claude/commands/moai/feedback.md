---
description: 피드백 수집 및 GitHub 이슈 생성 (버그 리포트, 기능 요청)
argument-hint: "[\"설명\"]"
allowed-tools: Skill
---

Dispatch the moai workflow `feedback` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `feedback` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `feedback` subcommand with the same arguments
