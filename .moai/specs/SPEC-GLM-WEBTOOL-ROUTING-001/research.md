# Research — SPEC-GLM-WEBTOOL-ROUTING-001

> All facts below were verified against official `docs.z.ai` documentation and direct reads of the codebase during plan-phase. This file is the research basis; it is not re-researched at run-phase.

## 1. z.ai official GLM Coding Plan MCP servers (3 in scope)

### 1.1 web_reader (remote HTTP MCP) — replaces built-in WebFetch
- Endpoint: `https://api.z.ai/api/mcp/web_reader/mcp`
- Tool name: `webReader`
- Auth: header `Authorization: Bearer ${Z_AI_API_KEY}`
- Source: https://docs.z.ai/devpack/mcp/reader-mcp-server
- Namespacing note: register the server as `web_reader` (underscore) to produce `mcp__web_reader__webReader`. z.ai docs name it `web-reader` (hyphen) → would yield `mcp__web-reader__webReader`. Server name is user-chosen; use underscore.

### 1.2 web_search_prime (remote HTTP MCP) — replaces built-in WebSearch
- Endpoint: `https://api.z.ai/api/mcp/web_search_prime/mcp`
- Tool name: `webSearchPrime`
- Auth: same Bearer header
- Source: https://docs.z.ai/devpack/mcp/search-mcp-server

### 1.3 zai-mcp-server (Vision, local stdio MCP) — replaces built-in Read on images
- Transport: LOCAL stdio via npx `@z_ai/mcp-server`
- Model: GLM-4.6V. Requires `@z_ai/mcp-server >= 0.1.2`, Node >= v22.
- Env: `Z_AI_API_KEY` + `Z_AI_MODE=ZAI` (international api.z.ai) or `Z_AI_MODE=ZHIPU` (China open.bigmodel.cn).
- Eight tools: `image_analysis`, `video_analysis`, `extract_text_from_screenshot`, `diagnose_error_screenshot`, `understand_technical_diagram`, `analyze_data_visualization`, `ui_to_artifact`, `ui_diff_check`.
- Input: a LOCAL FILE PATH (not base64) — this is why it bypasses the built-in Read base64→422 failure under GLM.
- Source: https://docs.z.ai/devpack/mcp/vision-mcp-server
- Per official docs the npx `@z_ai/mcp-server` package is VISION-ONLY; web_reader + web_search_prime are SEPARATE remote HTTP servers.
- Quota tiers: Lite 100 / Pro 1000 / Max 4000, shared across search + reader.

### 1.4 Out of scope
- `zread` (4th bundled server) — NOT in scope (EXCL-2). Only vision + web search + web reader.

### 1.5 Canonical .mcp.json snippet (verified pattern)
```json
{ "mcpServers": {
  "web_reader":       { "type":"http", "url":"https://api.z.ai/api/mcp/web_reader/mcp",       "headers": {"Authorization":"Bearer ${Z_AI_API_KEY}"} },
  "web_search_prime": { "type":"http", "url":"https://api.z.ai/api/mcp/web_search_prime/mcp", "headers": {"Authorization":"Bearer ${Z_AI_API_KEY}"} },
  "zai-mcp-server":   { "command":"npx", "args":["-y","@z_ai/mcp-server@latest"], "env": {"Z_AI_API_KEY":"${Z_AI_API_KEY}","Z_AI_MODE":"ZAI"} }
}}
```
`${Z_AI_API_KEY}` is expanded by Claude Code in `.mcp.json` headers/URLs.

## 2. GLM mode mechanism (codebase, verified)

- GLM detection runtime signal: `os.Getenv("ANTHROPIC_BASE_URL")` contains `api.z.ai`.
- Constant `DefaultGLMBaseURL = "https://api.z.ai/api/anthropic"` at `internal/config/defaults.go` (verified L41 region; alongside `DefaultGLMEnvVar = "GLM_API_KEY"` and GLM model tier constants).
- LLMMode constants `LLMModeClaude` / `LLMModeGLM` at `internal/runtime/cache_control.go` (verified L16-24). `LLMModeGLM = "glm"`.
- `moai glm` (`internal/cli/glm.go`): injects `ANTHROPIC_BASE_URL` + `ANTHROPIC_AUTH_TOKEN` + `Z_AI_API_KEY=apiKey` into PROCESS env (inherited by syscall.Exec). WHOLE session is GLM (leader + teammates).
- `moai cg` (`internal/cli/cg.go` → `applyCGMode` in `internal/cli/launcher.go`): hybrid. Injects GLM env into the TMUX SESSION only; REMOVES GLM env from the leader's process env + settings.local.json. Result: leader pane = Claude (built-in WebSearch/WebFetch WORK), new tmux panes (GLM teammates) = GLM (hit 529). `team_mode='cg'` persisted to llm.yaml.
- `moai cc` (`internal/cli/cc.go` → `applyCCMode`): strips all GLM env. Claude only.
- Auto-enable: `glm_tools.go autoEnableMCPServer()` (verified ~L470) attempts MCP registration on GLM launch unless `MOAI_GLM_NO_AUTO_TOOLS=1`.

## 3. Current registration state (the DRIFT to fix)

- Dev `.mcp.json` (repo root) registers `context7`, `chrome-devtools`, and a single `zai-mcp-server` (via `/bin/bash -l -c "exec npx -y @z_ai/mcp-server@latest"`) with `env: { Z_AI_MODE: "ZAI" }` only (NO `Z_AI_API_KEY`) and `$comment` "Z.AI MCP: Vision OCR, Web Search, Web Reader" — INACCURATE (npx is vision-only).
- `internal/template/templates/.mcp.json` does NOT exist (verified ABSENT). `.mcp.json` is NOT a static shipped template. User-facing registration path is `moai glm tools enable [vision|websearch|webreader|all]` (`internal/cli/glm_tools.go`, predecessor SPEC-GLM-MCP-001).
- `glm_tools.go` BUG (verified): `validateToolName` accepts the four names, but `buildZAIMCPEntry(token)` (verified L278-287) ALWAYS builds the SAME single npx `zai-mcp-server` entry (env `Z_AI_API_KEY` + `Z_AI_MODE=ZAI`), used at both L417 and L461. So `enable webreader` does NOT register a `web_reader` HTTP server — the tool-name argument is cosmetic.
- Success message (verified ~L192-203) hardcodes "Vision / Web Search / Web Reader" three lines; disable message (verified ~L238) hardcodes "제거된 도구: Vision, Web Search, Web Reader". Both claim three when only the npx vision server is registered.
- Registration writes to `~/.claude.json` (user scope, default) or `.mcp.json` (project scope, `--scope project`) via `resolveConfigPath` (verified).
- `runEnableMCPServer` / `disableMCPServerSafe` are `@MX:ANCHOR` (verified glm_tools.go:12, L376, L509). 22 GWT scenarios live in `glm_tools_test.go` (1209 lines, ~49 t.Run/name occurrences).
- Predecessor naming convention: requirements `REQ-GMC-*`, scenarios `GWT-1..22`, Node min major `nodeMinMajorVersion` (v22), `errNodeNotFound` sentinel.

## 4. MoAI WebSearch/WebFetch reference inventory (6 highest-leverage cross-link points)

All six verified to EXIST locally AND to have a template mirror (except CLAUDE.md, whose root copy is local-only per CLAUDE.local.md and whose ship mirror is `internal/template/templates/CLAUDE.md`):

1. `.claude/rules/moai/core/agent-common-protocol.md` §MCP Fallback Strategy — best home for the GLM routing cross-link.
2. `.claude/rules/moai/core/settings-management.md` §MCP Configuration — verified the single existing `zai-mcp-server (optional)` line at :32 (already references `moai glm tools enable [vision|websearch|webreader|all]`); extend + correct the server list.
3. `.claude/rules/moai/core/moai-constitution.md` §URL Verification — WebFetch verify + Sources.
4. `.claude/skills/moai-domain-research/SKILL.md` — research skill (parallel WebSearch+Context7 pattern).
5. `.claude/output-styles/moai/einstein.md` — Context7→WebFetch fallback chain.
6. `CLAUDE.md` §10 Web Search Protocol / §12 MCP Servers — orchestrator-facing. Root `CLAUDE.md` is LOCAL-only; `internal/template/templates/CLAUDE.md` is the shipped mirror — both need the cross-link.

Each cross-linked file has a TEMPLATE MIRROR under `internal/template/templates/.claude/...` that must stay in sync (Template-First). All six mirrors verified present.

## 5. Zone registry state

- `.claude/rules/moai/core/zone-registry.md` uses the `CONST-V3R5-NNN` parallel namespace.
- Highest existing entry: `CONST-V3R5-039` (verified). Next free IDs for new HARD clauses: `CONST-V3R5-040`, `041`.

## 6. Neutrality guard

- `internal/template/internal_content_leak_test.go` (verified present, ~16.8 KB) + `.github/workflows/template-neutrality-check.yaml` (verified present).
- Forbidden in template files: internal SPEC IDs, REQ/AC tokens, audit citations, internal dates, commit SHAs, macOS-bias paths, CLAUDE.local references.
- Allowed: generic prose, mechanism descriptions, vendor references framed feature-conditionally. z.ai is an allowed vendor reference.

## 7. SPEC ID validation record

- Proposed ID `SPEC-GLM-WEBTOOL-ROUTING-001` decomposition: `SPEC | GLM | WEBTOOL | ROUTING | 001` — all middle segments match `[A-Z][A-Z0-9]*`, last segment is exactly 3 digits. PASS against `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`. No collision (verified: only `SPEC-GLM-001` and `SPEC-GLM-MCP-001` exist). ID used as specified.
