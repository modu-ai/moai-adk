---
description: "(dev-only) release maintainer harness — MoAI-ADK production release (Enhanced GitHub Flow, scripts/release.sh + GoReleaser). NOT distributed to user projects."
argument-hint: "[VERSION] [--hotfix]"
allowed-tools: Agent
---

> **[DEV-ONLY]** maintainer harness. NOT distributed to user projects. user-owned namespace (`moai update` preserves it).

Run the `release` harness. Pure human-gated specialist — no Runner, no manifest (no non-interactive fan-out).

Use the `harness-release-specialist` subagent with arguments: $ARGUMENTS

The specialist is human-gated: the orchestrator holds every AskUserQuestion gate (version selection, production-release final approval). The specialist returns blocker reports for any user decision.
