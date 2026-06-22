---
description: "(dev-only) release-update maintainer harness — Claude Code upstream change tracker (release-notes version-delta sweep + docs sync). NOT distributed to user projects."
argument-hint: "[--since vX.Y.Z | --dry]"
allowed-tools: Agent
---

> **[DEV-ONLY]** maintainer harness. NOT distributed to user projects. user-owned namespace (`moai update` preserves it).

Run the `release-update` harness. Manifest (SSOT): `.claude/commands/harness/release-update/manifest.json`. Runner: `.claude/workflows/harness-release-update-run.js` (non-interactive CC-release-notes research fan-out only).

Use the `harness-release-update-specialist` subagent with arguments: $ARGUMENTS

The specialist is human-gated: the orchestrator holds every AskUserQuestion gate (user approval, PR creation); the Runner models only the non-interactive research sweep. The specialist returns blocker reports for any user decision.
