---
description: 코드베이스 스캔 및 codemaps/에 아키텍처 문서 생성
argument-hint: "[--force] [--area 영역]"
allowed-tools: Skill
---

Dispatch the moai workflow `codemaps` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `codemaps` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `~/.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `codemaps` subcommand with the same arguments
