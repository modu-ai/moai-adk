---
description: "GLM-backend web-tooling routing SSOT. Delivered on-demand by the GLM guardrail hook (SessionStart) when a GLM backend is detected; loaded into context here only when this rule file is edited."
paths: "**/glm-web-tooling.md"
---

# GLM-Backend Web Tooling Routing — Canonical Rule

This file is the **single source of truth** for how MoAI agents and the orchestrator perform web search, web fetch, and image reading when the session runs on the GLM backend (`moai glm`) or the GLM teammate panes of `moai cg`.

> **Why this rule exists**: Under a GLM backend the built-in Claude Code `WebSearch` / `WebFetch` tools route through the z.ai Anthropic-compatible gateway, which intermittently returns HTTP 529 (overload). Reading an image file with the built-in `Read` tool likewise hits a known base64-encoding failure (HTTP 422) under GLM. z.ai ships dedicated MCP servers that run server-side and bypass these failure modes. Without this doctrine, agents silently fall back to the failing built-in tools and research/fetch/vision operations break.

Cross-referenced by: `agent-common-protocol.md` §MCP Fallback Strategy, `settings-management.md` §MCP Configuration, `moai-constitution.md` §URL Verification, `output-styles/moai/einstein.md`, `CLAUDE.md` §10/§12.

---

## GLM-Backend Detection

[ZONE:Evolvable] [HARD] A session is **GLM-backed** when the `ANTHROPIC_BASE_URL` environment variable contains the substring `api.z.ai`. The canonical default value is `DefaultGLMBaseURL = "https://api.z.ai/api/anthropic"` (defined in `internal/config/defaults.go`). Equivalently, the runtime mode is `LLMModeGLM`.

Three launch modes must be distinguished:

| Launcher | Backend scope | Does this rule apply? |
|----------|---------------|-----------------------|
| `moai glm` | **Whole session** — orchestrator AND all teammates run on GLM | YES — applies to every agent and the orchestrator |
| `moai cg` | **Hybrid** — Claude leader pane + GLM teammate panes. The leader pane has its GLM env stripped and runs the Claude backend; only the GLM teammate panes hit z.ai routing | YES for GLM teammate panes; NO for the leader pane (see cg-leader exception below) |
| `moai cc` | Claude backend (no GLM env) | NO — built-in tools are the canonical path |

---

## HARD Routing Table

[ZONE:Evolvable] [HARD] While a session is GLM-backed, the built-in `WebSearch` / `WebFetch` tools and `Read`-on-an-image-file are **PROHIBITED** because they route through the 529-prone `api.z.ai/api/anthropic` gateway and the base64→422 image path. Each built-in tool MUST be replaced by the corresponding z.ai MCP tool:

| Built-in tool (PROHIBITED under GLM) | REQUIRED z.ai MCP replacement | Server | Transport |
|--------------------------------------|-------------------------------|--------|-----------|
| `WebSearch` | `mcp__web_search_prime__webSearchPrime` | `web_search_prime` | HTTP (remote, Bearer auth) |
| `WebFetch` | `mcp__web_reader__webReader` | `web_reader` | HTTP (remote, Bearer auth) |
| `Read` on an image file | `mcp__zai-mcp-server__analyze_image` (+ 7 sibling vision tools) | `zai-mcp-server` | stdio npx (local, GLM-4.6V) |

[ZONE:Evolvable] [HARD] While a session is GLM-backed, MoAI agents and the orchestrator SHALL NOT invoke the built-in `WebSearch` or `WebFetch`, nor `Read` on an image file. They SHALL route web search to `mcp__web_search_prime__webSearchPrime`, web fetch to `mcp__web_reader__webReader`, and image reading to a `mcp__zai-mcp-server__*` vision tool (default `analyze_image`).

### cg-leader exception

[ZONE:Evolvable] [HARD] Where the current pane is the `moai cg` **leader** pane (Claude backend, GLM env stripped), the HARD prohibition above does NOT apply — the built-in `WebSearch` / `WebFetch` / `Read` work normally there and are the canonical path. The HARD rule binds only `moai glm` whole-session contexts and `moai cg` GLM-teammate contexts.

---

## ToolSearch Preload

The z.ai MCP tools are **deferred** (their schema is not loaded at session start) unless the server entry carries `alwaysLoad: true`. Before first use, an agent MUST preload the tool schema:

```
ToolSearch(query: "select:mcp__web_search_prime__webSearchPrime")
ToolSearch(query: "select:mcp__web_reader__webReader")
ToolSearch(query: "select:mcp__zai-mcp-server__analyze_image")
```

Multiple tools may be selected in one call: `ToolSearch(query: "select:mcp__web_search_prime__webSearchPrime,mcp__web_reader__webReader")`.

---

## Image Input Mechanism

[ZONE:Evolvable] [HARD] The z.ai vision tools take a **LOCAL FILE PATH** as input, NOT a base64-encoded image. Pasting an image into the client bypasses the MCP and triggers the base64→422 failure path under GLM. Always pass the on-disk path of the image file to the vision tool.

The `zai-mcp-server` (GLM-4.6V) exposes eight vision tools — pick the most specific one for the task:

| Tool | Use for |
|------|---------|
| `analyze_image` | General image understanding (default) |
| `extract_text_from_screenshot` | OCR / text extraction from a screenshot |
| `diagnose_error_screenshot` | Reading an error message captured as a screenshot |
| `understand_technical_diagram` | Architecture / flow / sequence diagrams |
| `analyze_data_visualization` | Charts, graphs, dashboards |
| `ui_to_artifact` | Converting a UI mockup into code/markup |
| `ui_diff_check` | Comparing two UI states |
| `analyze_video` | Understanding a video file |

---

## Registration

The three servers are registered per-tool via:

```bash
moai glm tools enable vision      # registers zai-mcp-server (npx, GLM-4.6V vision)
moai glm tools enable websearch   # registers web_search_prime (HTTP)
moai glm tools enable webreader   # registers web_reader (HTTP)
moai glm tools enable all         # registers all three
```

Each tool-name argument registers the correct server: `vision` → `zai-mcp-server` (stdio npx), `websearch` → `web_search_prime` (HTTP), `webreader` → `web_reader` (HTTP). On `moai glm` launch the servers are auto-enabled.

---

## Anti-Patterns

- **AP-GWT-001 — Built-in WebSearch under GLM**: Calling `WebSearch` in a `moai glm` session or a `moai cg` GLM teammate pane. Routes through the 529-prone gateway. Use `mcp__web_search_prime__webSearchPrime`.
- **AP-GWT-002 — Built-in WebFetch under GLM**: Calling `WebFetch` under a GLM backend. Use `mcp__web_reader__webReader`.
- **AP-GWT-003 — Read-on-image under GLM**: Calling `Read` on an image file under a GLM backend. Triggers the base64→422 path. Use a `mcp__zai-mcp-server__*` vision tool with a local file path.
- **AP-GWT-004 — Base64 image input**: Pasting an image / passing base64 to a vision tool instead of a local file path. The MCP expects a path.
- **AP-GWT-005 — Applying the GLM prohibition to the cg leader pane**: Forcing the MCP tools on the `moai cg` leader pane, which runs the Claude backend and may use the built-in tools.
- **AP-GWT-006 — Skipping ToolSearch preload**: Invoking a deferred z.ai MCP tool without a `ToolSearch(query: "select:...")` preload (unless the server entry has `alwaysLoad: true`).

---

## Cross-References

- `agent-common-protocol.md` §MCP Fallback Strategy — general MCP fallback behavior
- `settings-management.md` §MCP Configuration — the three z.ai server entries and `alwaysLoad` semantics
- `moai-constitution.md` §URL Verification — URL verification under GLM uses `mcp__web_reader__webReader`
- `CLAUDE.md` §10 Web Search Protocol / §12 MCP Servers — orchestrator-facing routing pointer
- z.ai official docs: docs.z.ai/devpack/mcp (reader / search / vision MCP servers)

---

Version: 1.0.0
Classification: Canonical Reference — do not duplicate the routing table; cross-reference this file instead.
