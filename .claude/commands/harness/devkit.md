---
description: "(dev-only) devkit dev-maintainer harness — consolidates the three numeric maintainer commands (CC upstream tracker / GitHub issue-PR workflow / production release) into one /harness:devkit entry. NOT distributed to user projects."
argument-hint: "release-update [--since vX.Y.Z | --dry] | github issues|pr [...] | release [VERSION | --hotfix]"
allowed-tools: Skill
---

> **[DEV-ONLY]** maintainer harness. NOT distributed to user projects. user-owned namespace (`moai update` preserves it).

Run the `devkit` harness. Manifest (SSOT): `.claude/commands/harness/manifest.json`. Runner: `.claude/workflows/harness-devkit-run.js`.

Dispatch by first argument to the matching specialist (per manifest `specialists`):
- `release-update` → `harness-devkit-release-update-specialist`
- `github` → `harness-devkit-github-specialist`
- `release` → `harness-devkit-release-specialist`

All three capabilities are human-gated: the orchestrator holds every AskUserQuestion gate; the Runner models only the non-interactive research fan-out. Specialists return blocker reports for any user decision.

Use Skill("moai") with arguments: harness devkit $ARGUMENTS
