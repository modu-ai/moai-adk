# Implementation Plan — SPEC-GLM-WEBTOOL-ROUTING-001

## A. Context

This is a Tier M SPEC spanning a single domain (GLM web tooling) with two deliverable shapes: (1) doctrine/cross-link Markdown edits, and (2) a Go refactor of `internal/cli/glm_tools.go` with characterization-preserving tests. It extends SPEC-GLM-MCP-001.

The work is sequenced so the doctrine (the WHAT/WHY) lands first, the registration machinery (the HOW the doctrine becomes actionable) lands second, and the template mirror + neutrality verification closes the loop.

## B. Known Issues / Pre-existing Defects (to characterize before changing)

1. **glm_tools.go tool-name cosmetic bug** — `validateToolName` accepts `vision|websearch|webreader|all`, but `buildZAIMCPEntry(token)` always builds the same single `zai-mcp-server` npx entry regardless of which tool was requested (`internal/cli/glm_tools.go` ~L278-287, called at L417 and L461). So `enable webreader` does NOT register a `web_reader` HTTP server today.
2. **Inaccurate messages** — success message (the "활성화된 도구: Vision / Web Search / Web Reader" block) and disable message ("제거된 도구: Vision, Web Search, Web Reader") claim all three are managed when only the npx vision server is registered.
3. **Dev `.mcp.json` drift** — the `zai-mcp-server` entry uses a `/bin/bash -l -c "exec npx ..."` wrapper (not bare npx), its `$comment` claims "Vision OCR, Web Search, Web Reader", and its `env` block has only `Z_AI_MODE` (the `Z_AI_API_KEY` is absent). Run-phase MUST reconcile the wrapper-vs-bare-npx form with REQ-GWR-D and decide a single canonical form.
4. **22 existing GWT scenarios** — `internal/cli/glm_tools_test.go` (1209 lines, ~49 t.Run/name occurrences) covers `runEnableMCPServer` / `disableMCPServerSafe` (both `@MX:ANCHOR` per glm_tools.go:12, L376, L509). Behavior change to per-tool registration WILL require updating these characterization tests — they currently assert the single-entry shape.

## C. Pre-flight (run-phase entry checks)

- [ ] `go test ./internal/cli/ -run TestGLM` is green on the pre-change commit (baseline; characterization tests pass before any edit).
- [ ] `.claude/rules/moai/core/glm-web-tooling.md` does NOT yet exist (NEW file).
- [ ] zone-registry highest is `CONST-V3R5-039`; new entries begin at `CONST-V3R5-040`.
- [ ] All six cross-link target files + their template mirrors exist (verified in plan-phase).
- [ ] Package-level neutrality `go test ./internal/template/` is green on the pre-change commit (covers BOTH `TestTemplateNeutralityAudit` C1/C2 AND `TestTemplateNoInternalContentLeak` C3/C7 — the narrow `-run TestTemplateNeutralityAudit` form alone does NOT cover internal dates/SHAs).

## D. Constraints

- **Tier M**: M1-M6 milestones below. Single domain — Tier M not L.
- **Template-First**: source edits go in `internal/template/templates/` then `make build`, then local `.claude/` mirror sync. Markdown rule files have mirrors; verify each.
- **Language**: instruction docs in English; Go code comments per project `code_comments: ko`; godoc English.
- **Neutrality**: template-mirrored files carry NO internal SPEC IDs / dates / commit SHAs. z.ai vendor reference is allowed when framed feature-conditionally.
- **Behavior preservation**: glm_tools.go atomic-write, backup, idempotency, token-mismatch, Node-version-gate behaviors MUST survive the refactor.

## E. Self-Verification (manager-develop SHALL run at each milestone)

- After M2 (glm_tools refactor): `go test ./internal/cli/ -run TestGLM -count=1` green; per-tool registration asserted.
- After M4 (template mirror): package-level `go test ./internal/template/ -count=1` green — runs BOTH `TestTemplateNeutralityAudit` (C1/C2) AND `TestTemplateNoInternalContentLeak` (C3 internal dates + C7 commit SHAs, which `TestTemplateNeutralityAudit` explicitly defers to the sibling test). Running only `-run TestTemplateNeutralityAudit` is INSUFFICIENT for the no-internal-ID/date/SHA guarantee. `make build` succeeds; embedded.go regenerated.
- Doctrine grep: each of the six cross-link files contains a pointer string to `glm-web-tooling.md`.
- Routing table grep: `glm-web-tooling.md` contains `mcp__web_search_prime__webSearchPrime`, `mcp__web_reader__webReader`, `mcp__zai-mcp-server__image_analysis`.

## F. Milestones

### M1 — Author the routing doctrine rule file (REQ-GWR-A, REQ-GWR-F)
- Create `internal/template/templates/.claude/rules/moai/core/glm-web-tooling.md` (source) with: GLM-backend detection definition; routing table (WebSearch→webSearchPrime, WebFetch→webReader, Read-on-image→vision); cg-leader exception; ToolSearch preload note; local-file-path vision input note + 8 vision tools; anti-pattern catalogue; cross-references. HARD clauses per REQ-GWR-A3/A4.
- Add `CONST-V3R5-040..` entries to `zone-registry.md` (source + mirror) for the two HARD clauses (REQ-GWR-F1).
- `make build`; sync local `.claude/` mirror.

### M2 — Refactor `glm_tools.go` registration (REQ-GWR-C) [cycle_type: see §G]
- Replace the single-entry `buildZAIMCPEntry` with a per-tool dispatch: `vision`→stdio npx entry; `websearch`→`web_search_prime` http entry; `webreader`→`web_reader` http entry; `all`→all three.
- HTTP entries use `type:http` + url + `Authorization: Bearer ${Z_AI_API_KEY}` header.
- Correct success/disable messages to reflect actually-registered servers (REQ-GWR-C5).
- Preserve atomic-write / backup / idempotency / token-mismatch / Node-gate; apply Node-gate only when the tool set includes vision (REQ-GWR-C8).
- Update the 22 GWT characterization tests for the new per-tool shape; add new scenarios for per-tool registration, `all`, partial enable/disable, idempotency, multi-server disable.

### M3 — Correct dev `.mcp.json` (REQ-GWR-D)
- Register the three servers separately with corrected `$comment`. Add `Z_AI_API_KEY` to the `zai-mcp-server` env. Reconcile the bash-wrapper-vs-bare-npx form (decide canonical form consistent with M2 output).

### M4 — Cross-links into the six reference points (REQ-GWR-B)
- Add a concise pointer to `glm-web-tooling.md` in each of: `agent-common-protocol.md` §MCP Fallback Strategy; `settings-management.md` §MCP Configuration (extend/correct the existing `zai-mcp-server (optional)` line at :32); `moai-constitution.md` §URL Verification; `moai-domain-research/SKILL.md`; `einstein.md`; `CLAUDE.md` §10/§12. Source + each template mirror.
- Do NOT duplicate the routing table — reference only (REQ-GWR-B2).

### M5 — Template mirror + neutrality verification (REQ-GWR-E)
- Confirm every edited file with a mirror is in sync. `make build`. Run the neutrality CI test locally; ensure NO internal SPEC IDs / dates / SHAs leaked into template files.

### M6 — Full-suite verification + dogfood
- `go test ./...` green. `golangci-lint run` clean. CLI smoke: `moai glm tools enable webreader --scope project` registers a `web_reader` http entry (verify written JSON). Routing-doctrine grep checks pass.

## G. cycle_type recommendation (run-phase)

**Recommended: `ddd` (characterization-first).** Rationale: M2 modifies already-tested code (`runEnableMCPServer` / `disableMCPServerSafe`, both `@MX:ANCHOR`, with 22 existing GWT scenarios) with a deliberate behavior change (single-entry → per-tool registration). DDD's PRESERVE phase fits: first capture the current single-entry behavior as characterization tests (some already exist), confirm baseline green, then IMPROVE to per-tool registration while keeping the preserved invariants (atomic-write, backup, idempotency, token-mismatch, Node-gate) green at every step. The intentional behavior change (per-tool dispatch) is then a controlled IMPROVE with its own new specification tests.

If the project's `quality.yaml development_mode` is `tdd`, the acceptable fallback is TDD with characterization preserved: write the new per-tool RED tests first, but DO NOT delete the existing invariant-preserving tests — keep them as the regression net. State this choice explicitly to manager-develop.

The M1/M3/M4/M5 milestones are Markdown/JSON doctrine edits (no cycle_type — direct authoring + grep/neutrality verification).

## H. Anti-Patterns (run-phase guardrails)

- **AP-1**: Deleting the existing GWT invariant tests instead of updating them — loses the regression net for atomic-write/backup/idempotency.
- **AP-2**: Duplicating the routing table into the six cross-link files — violates single-source-of-truth (REQ-GWR-B2). Cross-reference only.
- **AP-3**: Leaking `SPEC-GLM-WEBTOOL-ROUTING-001` / today's date / a commit SHA into a template-mirrored file — fails neutrality CI (REQ-GWR-E2).
- **AP-4**: Applying the Node-version gate to a websearch/webreader-only enable — those are HTTP servers and need no npx/Node (REQ-GWR-C8).
- **AP-5**: Editing only the local `.claude/` copy and forgetting the `internal/template/templates/` source (Template-First inversion).
- **AP-6**: Hardcoding the z.ai endpoints/headers as Go string literals scattered across functions — extract to `const` per CLAUDE.local.md §14 (hardcoding ban).

## Cross-references

- spec.md (this SPEC) — REQ-GWR-A..F
- acceptance.md — AC-GWR-* matrix + GWT scenarios
- design.md — routing-table design, server-registry design, glm_tools dispatch design
- research.md — verified z.ai facts + codebase facts + reference inventory
- Predecessor: SPEC-GLM-MCP-001
