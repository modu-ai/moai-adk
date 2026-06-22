---
description: "(dev-only) github maintainer harness — GitHub issue-fix and PR-review via gh CLI. NOT distributed to user projects."
argument-hint: "issues|pr [--all | --label LABEL | NUMBER]"
allowed-tools: Agent
---

> **[DEV-ONLY]** maintainer harness. NOT distributed to user projects. user-owned namespace (`moai update` preserves it).

Run the `github` harness. Pure human-gated specialist — no Runner, no manifest (no non-interactive fan-out).

Use the `harness-github-specialist` subagent with arguments: $ARGUMENTS

The specialist is human-gated: the orchestrator holds every AskUserQuestion gate (PR creation approval, review submission). The specialist returns blocker reports for any user decision.
