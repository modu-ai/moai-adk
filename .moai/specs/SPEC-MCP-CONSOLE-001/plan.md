# plan.md — SPEC-MCP-CONSOLE-001

Milestones are ordered by decision reversibility. The two open design questions and the gating seam — the decisions most likely to change — lead; the i18n and rendering sweep, which is mechanical once the shape is fixed, comes last.

## §A. Context

### A.1 Base

All citations read at base commit `ed70e4354`.

### A.2 Tier judgment (recorded, not assumed)

Classified **Tier M**. Counted affected files: `fieldsets.templ` + `fieldsets_templ.go`, `handlers.go`, `app.go`, `schema_sections.go`, `mcp_server.go`, a new codex-auth view-model file + its test, a new console test, `assets/i18n.js`, and the existing i18n/governance tests. That is 11-13 files — above the Tier S ceiling, below the Tier L threshold of >15. If M1's gating seam turns out to touch more of `internal/cli` than the single registration site, the count crosses 15 and the SPEC should be tiered up rather than squeezed; that is the tier-up signal to watch.

### A.3 Verified seams

| Concern | Location | State at base |
|---|---|---|
| Tool registration | `internal/cli/mcp_server.go:105` `registerMoaiMCPTools(s *server.MCPServer)` | **takes no config**; called unconditionally at `:97` |
| Console routes | `internal/web/app.go:138-165` | `/`, `/save`, `/specs`, profile routes, `glmKeyRevealPath`, `/__shutdown__`, `/static/` |
| Schema seam | `internal/settings/schema_sections.go:291` `closedSeam(...)`, `:328` `s(SectionWorkflow, "workflow", TypeBool, "workflow","branch_guard","enabled")` | bool + closed-enum FieldDefs write typed yaml paths |
| Section IDs | `internal/settings/schema.go:26-55` | 20 sections; **no `SectionMCP`** |
| Audit-selection precedent | `internal/settings/schema_sections.go:338-344` + `internal/web/mcp_audit_surface_test.go:22-40` | `workflow.audit.model`, `workflow.audit.gates.{claude,codex,glm}` surface as schema fields |
| No-fork guards | `internal/web/mcp_audit_surface_test.go:47-95` | `internal/web` must not define `activeAuditBackend` or `ResolveAgentModelEffort` |
| GLM credential path | `internal/web/glmkey.go` (whole file), route at `app.go:156` | out-of-schema field, `computeGLMKeyHint` last-four disclosure, validation, reveal endpoint |
| codex probe | `internal/cli/mcp_codex.go:1143-1175` `handleCodexSetup` | reports installed / binary / version / auth_provider / enable_review_gate / allow_write / node_bridge:false |
| codex auth classify | `internal/cli/mcp_codex.go:1181-1183` `classifyCodexAuth` | runs `codex login status`, pattern-matches, fails open to `unknown` |
| i18n governance | `internal/web/i18n_governance_test.go:29-33` | 4 locales `en/ko/ja/zh`; key-coverage, orphan, endonym, shape checks over `assets/i18n.js` |

## §B. Milestones

### M1 — The gating seam (highest reversibility cost)

REQ-C-2. `registerMoaiMCPTools` takes only the server, so nothing today can read a per-tool preference. This milestone adds the read: registration consults a per-tool enablement map and skips a disabled tool, so a disabled tool never appears in `tools/list`.

Ordered first because it determines the config shape everything downstream renders and writes. It also carries the SPEC's only genuine open question:

**[RESOLVED 2026-08-12 — owner decision: option (a)]** Two candidate homes were considered:

- **(a) `.moai/config/sections/mcp.yaml`** via a new `SectionMCP` and the existing schema/`yamlpatch` seam. Consistent with every other console setting; per-project; readable by `internal/cli` at server start. Costs one new section ID and its i18n key block. **← adopted by owner decision 2026-08-12**
- **(b) The `.mcp.json` entry itself**, e.g. an args or env expression on the `moai` entry. Keeps MCP configuration in the MCP file, but `.mcp.json` is git-tracked and shared, and it is the file SPEC-A just put under 3-way merge — encoding per-tool state there entangles a user preference with a merge target. **← rejected (AP-C-6)**

Adopted (a): per-tool enablement stays out of the git-tracked, merge-target `.mcp.json` and on the same seam every other console setting uses.

**[RESOLVED 2026-08-12 — owner decision: all-enabled]** all-enabled-by-default (consistent with SPEC-A's first-class-default posture, and the only choice that keeps SPEC-B's agent grants functional out of the box) **← adopted**; read-only-enabled / write-capable-disabled was considered and rejected because it would ship agents whose SPEC-B-granted `verify_snapshot` / `goal_arm` / `codex_task` silently fails at runtime — a worse (invisible) failure mode than a too-permissive default that is at least surfaceable via REQ-C-3's write-capable distinction. The four write-capable tools remain individually disableable in the console.

### M2 — Per-tool console surface (REQ-C-1, REQ-C-3)

- Single declaration site for the 17 tools, shared by `registerMoaiMCPTools` and the console, so a new tool cannot be added without appearing in both (REQ-C-1's "cannot silently go unrepresented").
- Render through the schema-driven form, following the `workflow.branch_guard.enabled` bool pattern (`schema_sections.go:328`).
- Distinguish the four write-capable tools (`goal_arm`, `verify_snapshot`, `codex_task`, `codex_job_cancel`) from the 13 read-only ones in the rendered control (REQ-C-3).

### M3 — codex authentication surface (REQ-C-4, REQ-C-5, REQ-C-6)

View-model over the existing probe. No new classification logic.

**§C.2 — a scope finding, stated rather than worked around.** The Epic brief asked for "codex OAuth authentication configuration, driving the `codex_setup` handler". Measured against the handler, `codex_setup` is a **read-only probe**: it runs `exec.LookPath`, `codex --version`, and `codex login status`, and classifies the output (`mcp_codex.go:1144-1183`). It has no login capability, and the auth surface it probes is documented in its own source as undocumented and heuristic. So "driving the `codex_setup` handler" is achievable — display its state, offer its two opt-in toggles — but "OAuth configuration" in the sense of performing a login is not, and adding a browser-OAuth launcher to a local web console is a materially larger security surface than this SPEC's scope. REQ-C-5 therefore scopes the console to reporting state and naming the command; if the owner wants an in-console login flow, that is a separate SPEC with its own threat model.

### M4 — GLM key surface (REQ-C-7)

Reuse `glmkey.go` as-is. The whole file already exists and is tested (`glmkey_test.go`, `glm_tier_test.go`). The work is surfacing it in the MCP console section alongside the codex state, not re-authoring it. The out-of-schema placement (`glmkey.go:11-17`) is preserved verbatim — it is a structural anti-leak guarantee, not a stylistic choice.

### M5 — i18n + secret hygiene sweep (REQ-C-8, REQ-C-9, REQ-C-10) — most mechanical, ordered last

- All new keys in `assets/i18n.js` across en / ko / ja / zh in one change; governance suite green.
- Re-assert the no-fork guards (`mcp_audit_surface_test.go`) still pass.
- Confirm nothing this SPEC adds writes a resolved credential anywhere git-tracked.

## §C. Open questions summary

Both markers lived in M1 and both are now resolved (owner decision, 2026-08-12):

1. `per-tool enablement storage location` — **resolved: option (a) `.moai/config/sections/mcp.yaml` + new `SectionMCP`** (recommendation adopted).
2. `default state for the 17 tools` — **resolved: all-enabled** (recommendation adopted).

A third question was considered and resolved from the code rather than marked: whether the console should host a codex login flow. `handleCodexSetup`'s read-only shape settles it (§C.2 above).

No `[NEEDS CLARIFICATION]` markers remain. The SPEC is ready for Implementation Kickoff Approval (plan-auditor PASS 0.95 on the pre-resolution artifacts; this resolution converts two marked recommendations into adopted decisions without changing requirements or acceptance criteria).

## §D. Anti-patterns

- **AP-C-1 — Shipping a toggle nothing reads.** The named failure this SPEC guards against; REQ-C-2 and AC-C-004 exist for it.
- **AP-C-2 — Re-authoring the GLM credential path.** `glmkey.go` is deliberately out of `settings.AllFields()`; a "consistent" schema field for it would undo the anti-leak guarantee.
- **AP-C-3 — A second audit or model resolver inside `internal/web`.** Guarded by `mcp_audit_surface_test.go:47-95`.
- **AP-C-4 — Two tool lists.** One declaration, consumed by both the server and the console. Two lists drift on the first tool added.
- **AP-C-5 — Adding a locale key to en only.** The governance suite fails, but late; add all four together.
- **AP-C-6 — Encoding per-tool state into `.mcp.json`.** Candidate (b) above; entangles a user preference with the file SPEC-A just made a 3-way-merge target.

## §E. Cross-references

- `SPEC-MCP-DEFAULT-ON-001` (SPEC-A) — makes the server present; owns `.mcp.json` shape and merge safety.
- `SPEC-MCP-AGENT-WIRING-001` (SPEC-B) — grants agents the tools this console governs. Direct dependency.
- `SPEC-GLM-KEY-INPUT-001` — the completed SPEC that authored `glmkey.go`; this SPEC reuses, does not amend it.
- `SPEC-MOAI-MCP-SERVER-001` — the SPEC that built the server and the `codex_setup` probe.
