---
description: Create and run E2E tests across web, mobile, and desktop with auto-detection and CLI-first toolchains
argument-hint: "[--tool TOOL] [--platform web|mobile|desktop|desktop-native] [--record] [--journey NAME] [--autofix]"
allowed-tools: Skill
---

Dispatch the moai workflow `e2e` with: $ARGUMENTS
- Harness with the Skill tool (Claude Code): invoke `Skill("moai")` with arguments: `e2e` $ARGUMENTS
- Harness without a skill loader (Codex CLI): read `.agents/skills/moai/SKILL.md` (the mirrored dispatcher body) and follow its routing for the `e2e` subcommand with the same arguments
