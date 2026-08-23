---
id: SPEC-V3R4-CC2X-MCP-001
title: "CC 2.1.238/239 MCP Behavior Notes — stdio Discover Ordering, .mcp.json Trust Gating, Elicitation/5xx Fixes"
version: 0.1.0
status: draft
created: 2026-08-23
updated: 2026-08-23
author: manager-spec
priority: medium
phase: "v3.1.2"
module: mcp
lifecycle: spec-anchored
era: V3R6
tier: S
tags: "release-update, mcp, stdio, trust-gating, mcp-server"
parent: SPEC-V3R4-CC2X-ADOPT-002
---

# SPEC-V3R4-CC2X-MCP-001 (STUB)

Empty stub created by the 2026-08-23 `/harness:release-update` sweep (Phase 5
Option C). Scope, evidence, and tier tables live in the umbrella research:
[`../SPEC-V3R4-CC2X-ADOPT-002/research.md`](../SPEC-V3R4-CC2X-ADOPT-002/research.md).

Scope (one line): note CC 2.1.238's stdio `server/discover`-after-`initialize`
ordering fix (affects `moai mcp-server` startup expectations — smoke-test under
CC 2.1.239 first, umbrella Open Question 3) and the `.mcp.json` folder-trust
gating now applying under `claude -p` (headless fresh-clone caveat) in
`.claude/rules/moai/core/moai-mcp-tools.md` + docs-site 4-locale if user-facing;
optionally record the elicitation-fullscreen and remote-MCP 5xx-reconnect fixes.
Plan/run/sync bodies not yet authored.

### Out of Scope

- `moai mcp-server` Go code changes — no code change indicated by the delta; the smoke-test informs the note only.
- `headersHelper` marketplace/plugin features — MoAI ships no plugin marketplaces.
- MCP elicitation UI and remote-MCP 5xx reconnect fixes — recorded in the umbrella research only; no MoAI-owned surface to edit.
