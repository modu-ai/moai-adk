---
id: SPEC-GLM-WEBTOOL-ROUTING-001
title: "GLM-backend web search / web fetch / image read routing to z.ai MCP servers"
version: "0.1.0"
status: draft
created: 2026-06-03
updated: 2026-06-03
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli, .claude/rules/moai/core"
lifecycle: spec-anchored
tags: "glm, mcp, websearch, webfetch, vision, routing, doctrine"
tier: M
related_specs: [SPEC-GLM-MCP-001]
---

# SPEC-GLM-WEBTOOL-ROUTING-001 — GLM-backend web tooling routing to z.ai MCP servers

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-03 | manager-spec | Initial plan-phase authoring. Extends SPEC-GLM-MCP-001 (which introduced `moai glm tools enable`). Adds routing doctrine + corrects the registration machinery so the three official z.ai servers are registered separately. |
| 0.1.0 | 2026-06-03 | manager-spec | plan-audit D1-D4 debt remediation (no version bump — plan-phase clarifications): D1 added `tier: M` frontmatter (pins 0.80 Tier-M PASS threshold); D2 tightened AC-GWR-022 + Scenario 7 + plan.md §C/§E to require package-level `go test ./internal/template/` (both `TestTemplateNeutralityAudit` C1/C2 AND `TestTemplateNoInternalContentLeak` C3/C7) — the narrow `-run TestTemplateNeutralityAudit` defers internal date/SHA scanning to the sibling leak test; D4 committed REQ-GWR-A1 doctrine path to `.claude/rules/moai/core/glm-web-tooling.md` (removed "suggested" hedge to match AC-GWR-001). |

## A. Background / Problem

When MoAI runs under the GLM backend — either `moai glm` (whole session is GLM) or the GLM teammates of `moai cg` (hybrid: Claude leader + GLM teammate panes) — Claude Code's BUILT-IN `WebSearch` / `WebFetch` tools route through the z.ai Anthropic-compatible gateway (`https://api.z.ai/api/anthropic`, host `api.z.ai`). That gateway intermittently returns HTTP 529 (overload), which breaks research and fetch operations. Image reading via the built-in `Read` tool on image files likewise hits a known base64-encoding failure path (422) under GLM.

z.ai's official GLM Coding Plan ships dedicated MCP servers that run server-side and bypass this failure mode. The MoAI rule-base currently has NO doctrine instructing agents/orchestrator to prefer these z.ai MCP tools under GLM, and the registration machinery (`internal/cli/glm_tools.go`, the dev `.mcp.json`) is inaccurate: it registers a single npx server while claiming three capabilities (Vision, Web Search, Web Reader).

This SPEC closes both gaps: (1) author a HARD routing doctrine, and (2) align the registration machinery so the three official servers actually exist and each `enable` tool-name argument registers the correct server.

## B. Goal (testable, one sentence)

Under a GLM backend session, MoAI agents/orchestrator SHALL route web search → the z.ai `webSearchPrime` MCP tool, web fetch → the z.ai `webReader` MCP tool, and image reading → the z.ai vision MCP tools instead of the built-in `WebSearch` / `WebFetch` / `Read`, with the three supporting MCP servers correctly registered by `moai glm tools enable`.

## C. Predecessor relationship

This SPEC EXTENDS `SPEC-GLM-MCP-001`, which introduced the `moai glm tools enable|disable [vision|websearch|webreader|all]` subcommand and the auto-enable-on-GLM-launch behavior. SPEC-GLM-MCP-001 left a latent defect: the tool-name argument is cosmetic — every value builds the same single npx `zai-mcp-server` entry. This SPEC corrects that defect and layers the routing doctrine on top.

## D. Definitions

- **GLM backend session** — a Claude Code process whose `ANTHROPIC_BASE_URL` environment variable contains the substring `api.z.ai`. Equivalently, the runtime `LLMMode` is `LLMModeGLM` (`internal/runtime/cache_control.go`). Established by `moai glm` (whole session) or by the GLM teammate panes of `moai cg`.
- **cg-leader pane** — under `moai cg`, the leader pane has its GLM env STRIPPED from process env and `settings.local.json`; it runs on the Claude backend. Built-in `WebSearch` / `WebFetch` / `Read` work normally there.
- **z.ai web_search_prime server** — remote HTTP MCP at `https://api.z.ai/api/mcp/web_search_prime/mcp`, tool `webSearchPrime`. Bearer auth.
- **z.ai web_reader server** — remote HTTP MCP at `https://api.z.ai/api/mcp/web_reader/mcp`, tool `webReader`. Bearer auth.
- **z.ai vision server (`zai-mcp-server`)** — LOCAL stdio MCP via npx `@z_ai/mcp-server`, model GLM-4.6V. Eight tools (`image_analysis`, `video_analysis`, `extract_text_from_screenshot`, `diagnose_error_screenshot`, `understand_technical_diagram`, `analyze_data_visualization`, `ui_to_artifact`, `ui_diff_check`). Input is a LOCAL FILE PATH, not base64.
- **MCP tool namespacing** — Claude Code exposes an MCP server's tool as `mcp__<server-name>__<toolName>`. To produce the canonical references `mcp__web_search_prime__webSearchPrime`, `mcp__web_reader__webReader`, `mcp__zai-mcp-server__image_analysis`, the servers MUST be named `web_search_prime`, `web_reader`, `zai-mcp-server` respectively.

## E. Requirements (GEARS notation)

### REQ-GWR-A — Routing doctrine (NEW canonical rule file)

- **REQ-GWR-A1 (Ubiquitous)**: The doctrine rule file SHALL exist at the canonical core-rule path `.claude/rules/moai/core/glm-web-tooling.md` (decided — not a suggestion) and SHALL be the single source of truth for GLM-backend web-tooling routing.
- **REQ-GWR-A2 (Ubiquitous)**: The doctrine SHALL define the GLM-backend detection signal: a session is GLM-backed when `ANTHROPIC_BASE_URL` contains `api.z.ai`; and SHALL distinguish `moai glm` (whole session GLM) from `moai cg` (leader = Claude, teammate panes = GLM).
- **REQ-GWR-A3 (While, HARD)**: While a session is GLM-backed, MoAI agents and the orchestrator SHALL route web search to `mcp__web_search_prime__webSearchPrime`, web fetch to `mcp__web_reader__webReader`, and image reading to a `mcp__zai-mcp-server__*` vision tool (default `image_analysis`), instead of the built-in `WebSearch` / `WebFetch` / `Read`-on-image.
- **REQ-GWR-A4 (While, HARD — prohibition)**: While a session is GLM-backed, MoAI agents and the orchestrator SHALL NOT invoke the built-in `WebSearch` or `WebFetch`, nor `Read` on an image file, because those route through the 529-prone `api.z.ai/api/anthropic` gateway and the base64→422 image path.
- **REQ-GWR-A5 (Where — exception)**: Where the current pane is the `moai cg` leader pane (Claude backend), the HARD prohibition in REQ-GWR-A4 SHALL NOT apply; the built-in tools are permitted there. The HARD rule applies only to `moai glm` whole-session and `moai cg` GLM teammate panes.
- **REQ-GWR-A6 (While)**: While these z.ai MCP tools are deferred (not `alwaysLoad: true`), an agent SHALL preload the tool schema via `ToolSearch(query: "select:...")` before first use.
- **REQ-GWR-A7 (Ubiquitous)**: The doctrine SHALL state that the vision tool input is a LOCAL FILE PATH (not base64), and SHALL enumerate the eight vision tools so an agent can select the most specific one (e.g. `extract_text_from_screenshot` for OCR, `understand_technical_diagram` for diagrams).
- **REQ-GWR-A8 (Ubiquitous)**: The doctrine SHALL contain a routing table, an anti-pattern catalogue, and a cross-references section.

### REQ-GWR-B — Cross-links from existing reference points

- **REQ-GWR-B1 (Ubiquitous)**: The six highest-leverage WebSearch/WebFetch reference points SHALL each carry a concise GLM-routing pointer to the REQ-GWR-A doctrine file. The six points are: `agent-common-protocol.md` §MCP Fallback Strategy; `settings-management.md` §MCP Configuration; `moai-constitution.md` §URL Verification; `moai-domain-research/SKILL.md`; `output-styles/moai/einstein.md`; and `CLAUDE.md` §10/§12.
- **REQ-GWR-B2 (Ubiquitous — single-source-of-truth)**: The cross-links SHALL NOT duplicate the routing table; they SHALL reference the canonical doctrine file instead.

### REQ-GWR-C — `glm_tools.go` registration refactor

- **REQ-GWR-C1 (Event-driven)**: When a user runs `moai glm tools enable websearch`, the command SHALL register a `web_search_prime` server entry of `type: http` with url `https://api.z.ai/api/mcp/web_search_prime/mcp` and header `Authorization: Bearer ${Z_AI_API_KEY}`.
- **REQ-GWR-C2 (Event-driven)**: When a user runs `moai glm tools enable webreader`, the command SHALL register a `web_reader` server entry of `type: http` with url `https://api.z.ai/api/mcp/web_reader/mcp` and the same Bearer header.
- **REQ-GWR-C3 (Event-driven)**: When a user runs `moai glm tools enable vision`, the command SHALL register the `zai-mcp-server` stdio npx entry (`@z_ai/mcp-server`, env `Z_AI_API_KEY` + `Z_AI_MODE=ZAI`).
- **REQ-GWR-C4 (Event-driven)**: When a user runs `moai glm tools enable all`, the command SHALL register all three servers (vision + web_search_prime + web_reader).
- **REQ-GWR-C5 (Ubiquitous)**: The success message SHALL accurately reflect which server(s) were registered for the given tool-name argument; the disable message SHALL accurately reflect which server(s) were removed.
- **REQ-GWR-C6 (Event-driven)**: When a user runs `moai glm tools disable <tool-name>`, the command SHALL remove only the server(s) corresponding to that tool-name (or all three for `all`), preserving any unrelated MCP entries.
- **REQ-GWR-C7 (Ubiquitous — preservation)**: The refactor SHALL preserve the existing atomic-write, timestamped-backup, idempotency (token-match skip), token-mismatch handling, and Node-version-gate behaviors from SPEC-GLM-MCP-001.
- **REQ-GWR-C8 (Where — capability gate)**: Where a registered server is an HTTP server (web_search_prime, web_reader), the Node-version gate SHALL NOT block registration (the gate is only relevant to the npx vision server). The gate SHALL still apply when the requested tool set includes vision.

### REQ-GWR-D — Dev `.mcp.json` correction

- **REQ-GWR-D1 (Ubiquitous)**: The repo-root `.mcp.json` SHALL register the three z.ai servers separately (`web_search_prime` http, `web_reader` http, `zai-mcp-server` npx) with a corrected `$comment` that accurately describes each server's capability.
- **REQ-GWR-D2 (Ubiquitous)**: The `.mcp.json` `zai-mcp-server` entry SHALL carry `Z_AI_API_KEY` via `${Z_AI_API_KEY}` expansion in addition to `Z_AI_MODE=ZAI` (the current dev entry omits the key).

### REQ-GWR-E — Template mirror + neutrality

- **REQ-GWR-E1 (Ubiquitous)**: Every doctrine/cross-link edit to a file that has a template mirror under `internal/template/templates/.claude/...` SHALL be mirrored there (Template-First).
- **REQ-GWR-E2 (Ubiquitous — neutrality)**: Template-mirrored files SHALL contain NO internal SPEC IDs, internal dates, or commit SHAs; the z.ai vendor reference framed feature-conditionally ("When using moai glm ...") is permitted.
- **REQ-GWR-E3 (While)**: While the template-neutrality CI guard runs (`internal/template/internal_content_leak_test.go` + `.github/workflows/template-neutrality-check.yaml`), it SHALL pass for all newly added/edited template files.

### REQ-GWR-F — Zone registry HARD-clause registration (capability gate)

- **REQ-GWR-F1 (Where)**: Where the doctrine introduces HARD clauses (REQ-GWR-A3, A4), corresponding entries SHALL be added to `.claude/rules/moai/core/zone-registry.md` in the `CONST-V3R5` parallel namespace (next available IDs begin at `CONST-V3R5-040`).

## Exclusions (What NOT to Build)

- **EXCL-1**: This SPEC does NOT change the GLM backend detection mechanism itself (`moai glm` / `moai cg` / `moai cc` env injection). The detection signal is consumed, not modified.
- **EXCL-2**: This SPEC does NOT register the 4th bundled z.ai server (`zread`). Only vision + web search + web reader are in scope.
- **EXCL-3**: This SPEC does NOT add a static template `.mcp.json` under `internal/template/templates/`. No such file exists today; the user-facing registration path is `moai glm tools enable` (REQ-GWR-C), and that is the shipped mechanism.
- **EXCL-4**: This SPEC does NOT modify the built-in Claude `WebSearch` / `WebFetch` / `Read` tools or the Claude-backend behavior. Under a Claude backend (`moai cc`, or the `moai cg` leader pane) the built-in tools remain the canonical path.
- **EXCL-5**: This SPEC does NOT implement automatic interception/rewriting of built-in tool calls at runtime. Routing is enforced by doctrine (agent/orchestrator discipline), not by a runtime shim.
- **EXCL-6**: This SPEC does NOT change quota or billing behavior of the z.ai plan tiers (Lite/Pro/Max). Quota is documented for context only.
- **EXCL-7**: This SPEC does NOT define new function signatures, struct layouts, or test names in spec.md — those are deferred to the Run phase (see plan.md for the cycle_type recommendation and design.md for the structural approach).

## Cross-references

- Predecessor: `SPEC-GLM-MCP-001` (`moai glm tools enable` machinery)
- Detection: `internal/config/defaults.go` (`DefaultGLMBaseURL`), `internal/runtime/cache_control.go` (`LLMMode`)
- Registration: `internal/cli/glm_tools.go`
- Doctrine home (NEW): `.claude/rules/moai/core/glm-web-tooling.md`
- Cross-link targets: `agent-common-protocol.md`, `settings-management.md`, `moai-constitution.md`, `moai-domain-research/SKILL.md`, `einstein.md`, `CLAUDE.md`
- Neutrality guard: `internal/template/internal_content_leak_test.go`, `.github/workflows/template-neutrality-check.yaml`
- z.ai official docs: docs.z.ai/devpack/mcp/{reader,search,vision}-mcp-server
