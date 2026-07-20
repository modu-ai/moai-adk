---
description: "(user-owned) oss-docs harness — README 4-locale + docs-site publishing for the moai-adk-go OSS project (adk.mo.ai.kr, Vercel auto-deploy)."
argument-hint: "[task description | --dry]"
allowed-tools: Agent
---

> **[USER-OWNED]** `hns-` namespace harness. `moai update` preserves it. NOT distributed to user projects (never touches `internal/template/templates/`).

Run the `oss-docs` harness. Manifest (SSOT): `.claude/commands/harness/oss-docs/manifest.json`. Runner: `.claude/workflows/hns-oss-docs-run.js` (non-interactive scope → author → translate → verify pipeline only).

Dispatch the specialists declared in the manifest (content-author, locale-translator, structure-curator) with arguments: $ARGUMENTS

The harness is human-gated at publish points: the orchestrator holds every AskUserQuestion gate (publish approval, commit, push, PR); specialists return blocker reports for any user decision. The Runner never commits or pushes.
