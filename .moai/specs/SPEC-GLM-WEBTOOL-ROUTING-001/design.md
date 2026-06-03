# Design — SPEC-GLM-WEBTOOL-ROUTING-001

> Design is the bridge between the WHAT/WHY in spec.md and the implementation in the Run phase. It records structural decisions; it is NOT itself implementation.

## B.1 Decision: doctrine-driven routing, not a runtime shim

**Options considered:**
- **Option A (chosen)** — express routing as a HARD doctrine clause that agents/orchestrator follow. No runtime interception.
- Option B — a runtime middleware that rewrites built-in `WebSearch`/`WebFetch`/`Read` calls to the MCP equivalents under GLM.

**Rationale for A**: Option B would require hooking the Claude Code tool dispatch layer (not a MoAI-owned surface) and is brittle across Claude Code versions. MoAI's enforcement model is doctrine + agent discipline (the same model used for the AskUserQuestion boundary, the parallel-execution rule, etc.). Option A is consistent, version-independent, and testable by grep. EXCL-5 records that the runtime shim is out of scope.

## B.2 Routing table (the doctrine core)

| Need | Built-in tool (FORBIDDEN under GLM) | z.ai MCP replacement | Server | Transport |
|------|-------------------------------------|----------------------|--------|-----------|
| Web search | `WebSearch` | `mcp__web_search_prime__webSearchPrime` | `web_search_prime` | HTTP (remote) |
| Web fetch | `WebFetch` | `mcp__web_reader__webReader` | `web_reader` | HTTP (remote) |
| Image read | `Read` (on image file) | `mcp__zai-mcp-server__image_analysis` (+ 7 other vision tools) | `zai-mcp-server` | stdio npx (local) |

The cg-leader pane (Claude backend) is exempt: built-in tools are the canonical path there.

## B.3 Server naming decision (namespacing constraint)

Claude Code namespaces an MCP tool as `mcp__<server-name>__<toolName>`. The maintainer already uses `mcp__web_reader__webReader`. z.ai's own docs name the server `web-reader` (hyphen), which would yield `mcp__web-reader__webReader`. Since the server name is user-chosen at registration time, we register with UNDERSCORE names (`web_reader`, `web_search_prime`) to produce the canonical underscore tool references. `zai-mcp-server` keeps its hyphenated name (the maintainer's existing reference `mcp__zai-mcp-server__*` already uses it).

## B.4 glm_tools.go dispatch design (REQ-GWR-C)

Current shape (single entry):
- `buildZAIMCPEntry(token)` → one npx `zai-mcp-server` map, used for every tool-name.

Target shape (per-tool dispatch):
- A builder that, given a tool-name, returns the set of `{serverKey: entry}` pairs to register:
  - `vision` → `{ "zai-mcp-server": <npx stdio entry> }`
  - `websearch` → `{ "web_search_prime": <http entry> }`
  - `webreader` → `{ "web_reader": <http entry> }`
  - `all` → union of the three
- HTTP entry shape (verified pattern): `{ "type": "http", "url": "<endpoint>", "headers": { "Authorization": "Bearer ${Z_AI_API_KEY}" } }`. The `${Z_AI_API_KEY}` is expanded by Claude Code in `.mcp.json` headers/URLs — so the literal `${Z_AI_API_KEY}` string is written, NOT the resolved token. (Contrast with the npx entry, which uses `env: { Z_AI_API_KEY: <token> }` where the resolved token is written.)
- The existing `runEnableMCPServer` / `runEnableMCPServerScoped` / `disableMCPServerSafe` entry points (both `@MX:ANCHOR`) keep their signatures where possible; the per-tool decision is layered inside. The 22-scenario test net pins these anchors — signature changes ripple to every test, so prefer additive dispatch over signature churn.

**Run-phase author note**: extract the three endpoint URLs + the Bearer header template to package-level `const` (CLAUDE.local.md §14 hardcoding ban). Suggested: `zaiWebSearchPrimeURL`, `zaiWebReaderURL`, plus the existing `zaiNPXPackage`.

## B.5 Node-version gate scoping (REQ-GWR-C8)

Today the Node gate (`detectNodeFn`, min major v22) runs unconditionally on enable. After the refactor, the gate is only meaningful for the npx vision server. Design: evaluate the requested tool set first; run the Node gate ONLY if the set includes `vision` (i.e. `vision` or `all`). `websearch` / `webreader` enable proceeds without the gate. The gate's existing error messages and `errNodeNotFound` sentinel are reused unchanged.

## B.6 Message accuracy (REQ-GWR-C5)

Replace the hardcoded "Vision / Web Search / Web Reader" three-line success block with a block generated from the actually-registered server set. Same for the disable "제거된 도구" line. Design keeps the existing Korean user-facing strings (project `documentation: ko`) but makes them data-driven off the registered set.

## B.7 Dev `.mcp.json` reconciliation (REQ-GWR-D)

The dev file currently wraps npx in `/bin/bash -l -c "exec npx ..."` (a login-shell wrapper used by the other dev servers context7/chrome-devtools for PATH reasons) and omits `Z_AI_API_KEY`. Design decision for run-phase: the dev `.mcp.json` MAY keep the bash-wrapper form for the npx vision server (consistent with sibling dev entries) but MUST add `Z_AI_API_KEY` via `${Z_AI_API_KEY}` and MUST add the two HTTP server entries. The `glm_tools.go` machinery (REQ-GWR-C, the user-facing path) uses the bare-npx form for the vision entry — the two forms are allowed to differ because `.mcp.json` is dev-local (CLAUDE.local.md local-only file) and the machinery is the shipped contract. Run-phase should note this divergence explicitly rather than force-unify.

## B.8 Cross-link insertion design (REQ-GWR-B)

Each of the six files gets a single sentence/bullet pointer, not a copy of the table:
- `agent-common-protocol.md` §MCP Fallback Strategy — add a note: under a GLM backend, prefer the z.ai MCP tools per `glm-web-tooling.md`.
- `settings-management.md` §MCP Configuration — extend the existing `zai-mcp-server (optional)` line (:32) to list the three servers and point to the doctrine.
- `moai-constitution.md` §URL Verification — note that WebFetch verification under GLM uses `mcp__web_reader__webReader`.
- `moai-domain-research/SKILL.md` — note GLM-backend search uses webSearchPrime.
- `einstein.md` — note the Context7→WebFetch fallback chain swaps WebFetch for webReader under GLM.
- `CLAUDE.md` §10/§12 (both the local root and `internal/template/templates/CLAUDE.md` mirror) — orchestrator-facing pointer.

## B.9 Zone-registry HARD entries (REQ-GWR-F)

Two `CONST-V3R5` entries (next free: `CONST-V3R5-040`, `041`) for REQ-GWR-A3 (mandate) and REQ-GWR-A4 (prohibition). Format follows the existing parallel-namespace convention in `zone-registry.md`.

## B.10 Risks

- **R1**: The 22 GWT tests assume single-entry shape; the per-tool refactor touches every one. Mitigation: DDD characterization-first (plan.md §G) — preserve invariants, change the registration shape deliberately.
- **R2**: `${Z_AI_API_KEY}` expansion semantics differ between HTTP-header context and npx-env context. Mitigation: HTTP entries write the literal `${Z_AI_API_KEY}` (Claude Code expands); npx entry writes the resolved token (as today). Design B.4 makes this explicit.
- **R3**: Neutrality CI may flag the z.ai vendor strings. Mitigation: frame every template mention feature-conditionally ("When using moai glm ...") per REQ-GWR-E2; the neutrality test forbids internal markers, not vendor names.
