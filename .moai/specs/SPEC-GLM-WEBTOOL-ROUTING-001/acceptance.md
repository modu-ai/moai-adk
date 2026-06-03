# Acceptance Criteria — SPEC-GLM-WEBTOOL-ROUTING-001

All criteria are independently verifiable via grep (doctrine clauses), `go test` (registration + neutrality), and inspection of written JSON. AC sub-IDs may use a trailing lowercase suffix (e.g. `AC-GWR-003a`) to denote paired sub-criteria of one logical AC — this is an acceptance-criteria convention only and does NOT apply to the SPEC ID.

## D. AC Matrix

| AC ID | REQ | Severity | Verification method |
|-------|-----|----------|---------------------|
| AC-GWR-001 | A1 | MUST | `glm-web-tooling.md` exists at the canonical core-rule path (source + local mirror) |
| AC-GWR-002 | A2 | MUST | grep doctrine for `api.z.ai` detection clause + `moai glm` vs `moai cg` distinction |
| AC-GWR-003 | A3 | MUST | grep doctrine routing table for the three `mcp__*` tool references |
| AC-GWR-004 | A4 | MUST | grep doctrine for the HARD prohibition on built-in WebSearch/WebFetch/Read-on-image |
| AC-GWR-005 | A5 | MUST | grep doctrine for the cg-leader-pane exception clause |
| AC-GWR-006 | A6 | SHOULD | grep doctrine for `ToolSearch(query: "select:` preload note |
| AC-GWR-007 | A7 | MUST | grep doctrine for "local file path" + all eight vision tool names |
| AC-GWR-008 | A8 | SHOULD | doctrine has routing-table + anti-pattern + cross-references sections |
| AC-GWR-009 | B1 | MUST | each of the six cross-link files contains a pointer to `glm-web-tooling.md` |
| AC-GWR-010 | B2 | MUST | no cross-link file duplicates the full routing table (reference only) |
| AC-GWR-011 | C1 | MUST | `enable websearch` registers `web_search_prime` http entry (test) |
| AC-GWR-012 | C2 | MUST | `enable webreader` registers `web_reader` http entry (test) |
| AC-GWR-013 | C3 | MUST | `enable vision` registers `zai-mcp-server` npx entry (test) |
| AC-GWR-014 | C4 | MUST | `enable all` registers all three servers (test) |
| AC-GWR-015 | C5 | MUST | success/disable messages reflect actually-registered servers (test) |
| AC-GWR-016 | C6 | MUST | `disable <tool>` removes only the matching server(s), preserves unrelated MCP entries (test) |
| AC-GWR-017 | C7 | MUST | atomic-write/backup/idempotency/token-mismatch/Node-gate invariants preserved (existing GWT tests still green) |
| AC-GWR-018 | C8 | MUST | websearch/webreader-only enable does NOT block on Node version (test) |
| AC-GWR-019 | D1 | MUST | dev `.mcp.json` has three separate z.ai server entries + corrected `$comment` |
| AC-GWR-020 | D2 | MUST | dev `.mcp.json` `zai-mcp-server` env contains `Z_AI_API_KEY` |
| AC-GWR-021 | E1 | MUST | every edited mirrored file is in sync (source ↔ template) |
| AC-GWR-022 | E2/E3 | MUST | `go test ./internal/template/ -run TestTemplateNeutralityAudit` green; no internal SPEC ID/date/SHA in template files |
| AC-GWR-023 | F1 | SHOULD | `zone-registry.md` has new `CONST-V3R5-040..` entries for the HARD clauses |
| AC-GWR-024 | (whole) | MUST | `go test ./...` green + `golangci-lint run` clean post-change |

## D.1 Severity legend

- **MUST**: blocking; SPEC cannot close `implemented` if any MUST AC fails.
- **SHOULD**: non-blocking; failure recorded as debt with rationale.

## Given-When-Then Scenarios

### Scenario 1 — Doctrine routing under `moai glm` (REQ-GWR-A3, A4)
- **Given** a session whose `ANTHROPIC_BASE_URL` contains `api.z.ai` (started by `moai glm`)
- **When** an agent needs to perform a web search
- **Then** the doctrine REQUIRES routing to `mcp__web_search_prime__webSearchPrime` and FORBIDS the built-in `WebSearch`
- **Verify**: `grep -E "mcp__web_search_prime__webSearchPrime" .claude/rules/moai/core/glm-web-tooling.md` matches; the same file contains a HARD `SHALL NOT` clause naming built-in `WebSearch`/`WebFetch`.

### Scenario 2 — cg-leader exception (REQ-GWR-A5)
- **Given** a `moai cg` session, focused on the leader pane (Claude backend, GLM env stripped)
- **When** the leader performs a web fetch
- **Then** the built-in `WebFetch` is permitted (HARD prohibition does not apply to the leader pane)
- **Verify**: `grep` the doctrine for an explicit "cg leader pane" / "Claude backend" exception clause tied to REQ-GWR-A5.

### Scenario 3 — Per-tool registration: webreader (REQ-GWR-C2)
- **Given** a `Z_AI_API_KEY` is available and Node is not required for an HTTP server
- **When** the user runs `moai glm tools enable webreader --scope project`
- **Then** `.mcp.json` gains a `web_reader` entry of `type:http`, url `https://api.z.ai/api/mcp/web_reader/mcp`, header `Authorization: Bearer ${Z_AI_API_KEY}`, AND no `web_search_prime` / `zai-mcp-server` entry is added
- **Verify**: `go test ./internal/cli/ -run TestGLMTools` scenario asserts the written JSON shape; CLI smoke confirms.

### Scenario 4 — `enable all` registers three servers (REQ-GWR-C4)
- **Given** a valid token and Node >= v22 present
- **When** the user runs `moai glm tools enable all`
- **Then** all three entries (`web_search_prime` http, `web_reader` http, `zai-mcp-server` npx) are registered, and the success message names exactly the three servers actually written
- **Verify**: test asserts all three keys present + message content.

### Scenario 5 — Invariant preservation: idempotency + backup (REQ-GWR-C7)
- **Given** the three servers are already registered with the same token
- **When** the user re-runs `moai glm tools enable all`
- **Then** the command skips (token-match) with the "이미 활성화 — 변경 없음" message and writes no new backup
- **Verify**: existing GWT idempotency + backup tests remain green after the refactor.

### Scenario 6 — Node gate scoped to vision only (REQ-GWR-C8)
- **Given** Node is absent from PATH
- **When** the user runs `moai glm tools enable webreader`
- **Then** registration SUCCEEDS (HTTP server needs no Node); but `enable vision` (or `all`) FAILS the Node gate with the existing version-error message
- **Verify**: two test cases — webreader-no-node PASS, vision-no-node FAIL.

### Scenario 7 — Template neutrality (REQ-GWR-E2/E3)
- **Given** the doctrine + cross-links are mirrored into `internal/template/templates/.claude/...`
- **When** the neutrality CI test runs
- **Then** it passes — no `SPEC-GLM-WEBTOOL-ROUTING-001`, no internal date, no commit SHA leaked into any template file
- **Verify**: `go test ./internal/template/ -run TestTemplateNeutralityAudit -count=1` green.

## D.2 Edge cases

- **EC-1**: `disable webreader` when only `web_reader` is registered → removes it, leaves other MCP entries (context7, chrome-devtools) untouched.
- **EC-2**: `disable all` when a partial set (e.g. only vision) is registered → removes what exists, reports accurately, no error on absent entries.
- **EC-3**: `enable websearch` then `enable webreader` (sequential partial) → both http entries coexist; second enable does not clobber the first.
- **EC-4**: token-mismatch on re-enable → existing token-mismatch path triggers backup + overwrite (preserved behavior).
- **EC-5**: Windows — npx vision server registration path unchanged; HTTP entries are OS-independent.

## Definition of Done

- [ ] All MUST ACs pass; SHOULD AC failures (if any) recorded as debt with rationale.
- [ ] `go test ./...` green; `golangci-lint run` clean.
- [ ] `glm-web-tooling.md` exists (source + mirror + local), is the single source of truth, contains routing table + HARD clauses + anti-patterns + cross-refs.
- [ ] Six cross-link files each point to the doctrine (no table duplication).
- [ ] `glm_tools.go` registers the correct server per tool-name; messages accurate; invariants preserved.
- [ ] Dev `.mcp.json` has three separate corrected entries incl. `Z_AI_API_KEY`.
- [ ] Neutrality CI green; no internal markers in template files.
- [ ] zone-registry `CONST-V3R5-040..` HARD entries added.
- [ ] `make build` succeeds; embedded.go regenerated; local mirror synced.
