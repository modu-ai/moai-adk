---
id: SPEC-MCP-CONSOLE-001
title: "moai web MCP console — per-tool settings surface, codex auth configuration, GLM key reuse, 4-locale coverage"
version: "0.1.0"
status: draft
created: 2026-08-12
updated: 2026-08-12
author: manager-spec
priority: P2
phase: "v3.2 target"
module: "internal/web, internal/settings, internal/cli/mcp_server.go"
lifecycle: spec-anchored
tier: M
tags: "mcp, web-console, settings-schema, codex-auth, glm-key, i18n, secret-hygiene"
depends_on: [SPEC-MCP-AGENT-WIRING-001]
---

# SPEC-MCP-CONSOLE-001 — moai web MCP console (Epic SPEC-C)

## HISTORY

- 2026-08-12 (plan-phase, iter-1) — Initial Tier M authoring. Third and final SPEC of the moai-MCP integration Epic (A → B → C). Adds a per-tool MCP settings surface to the `moai web` console, a codex authentication configuration surface driven by the existing `codex_setup` probe, and a GLM API-key surface reusing the established `internal/web/glmkey.go` path — all with 4-locale coverage and no secret in any git-tracked file. Two scope findings surfaced during reconnaissance are recorded in `plan.md` §C rather than papered over: the MCP server has no per-tool gating seam today, and `codex_setup` reports auth state but cannot perform a login.

## §A. User Story

**As a** MoAI-ADK user who now has the `moai` MCP server provisioned by default (SPEC-A) and agents wired to its tools (SPEC-B),

**I want** to see and control that surface from `moai web` — which of the 17 tools are enabled, whether codex is authenticated, whether a GLM key is configured —

**so that** I manage the MCP surface through the same console I already use for the rest of my project settings, in my own language, without ever hand-editing a JSON file or pasting a credential into a tracked one.

## §B. Scope Summary

Three surfaces in the `moai web` console:

1. **Per-tool MCP settings** — the 17 tools registered by `registerMoaiMCPTools` (`internal/cli/mcp_server.go:105`), each individually enableable, expressed through the existing schema-driven form + `yamlpatch` seam that `workflow.branch_guard.enabled` and the audit-selection fields already use (`internal/settings/schema_sections.go:328`, `:338-344`).
2. **codex authentication configuration** — surfacing the state `codex_setup` already probes (`internal/cli/mcp_codex.go:1143-1175`: installed, binary path, version, auth provider, review-gate and write-mode opt-ins) and offering the opt-in toggles it reports.
3. **GLM API-key configuration** — reusing `internal/web/glmkey.go` verbatim: the out-of-schema credential field, the bounded last-four disclosure (`computeGLMKeyHint`), the parse/validate rules, and the `glmKeyRevealPath` endpoint.

All three carry 4-locale coverage per the existing i18n governance suite, and no surface writes a credential into a git-tracked file.

### Out of Scope — provisioning defaults

- Whether the `moai` MCP server is provisioned at all → **SPEC-MCP-DEFAULT-ON-001** (SPEC-A).

### Out of Scope — agent tool grants

- Which agents may call which tools → **SPEC-MCP-AGENT-WIRING-001** (SPEC-B). The console controls whether a tool is available at all; it does not edit agent frontmatter.

### Out of Scope — performing a codex login

- The console does not run an OAuth flow. `codex_setup` classifies auth state by running `codex login status` and pattern-matching (`internal/cli/mcp_codex.go:1181-1183`); it has no login capability, and shelling a browser-based OAuth flow out of a local web console is a materially different security surface. The console **reports** auth state and links the user to the `codex login` command. See `plan.md` §C.2.

### Out of Scope — third-party MCP entry management

- `moai mcp add|remove|list` (SPEC-TREND-MCP-001 REQ-TMC-005) stays the CLI surface for third-party entries. This console governs the `moai` server's own tool surface.

### Out of Scope — the live WebSocket board

- The real-time `moai web` board remains its own deferred SPEC and is untouched.

## §C. Requirements (GEARS)

> Domain prefix `REQ-C-N`. Citations read at base commit `ed70e4354`.

### M1 — Per-tool settings surface

**REQ-C-1** (Ubiquitous) The console shall present each of the 17 MCP tools as an individually-controllable setting, rendered through the existing schema-driven form seam rather than a bespoke form. The tool list shall be derived from a single declaration so that a tool added to `registerMoaiMCPTools` cannot silently go unrepresented in the console.

**REQ-C-2** (Ubiquitous) The MCP server shall consult the per-tool setting at registration time, so a disabled tool is not registered and does not appear in `tools/list`. This requires a **new gating seam**: `registerMoaiMCPTools(s *server.MCPServer)` (`internal/cli/mcp_server.go:105`, called at `:97`) takes no configuration at base, so a console setting written today would be read by nothing. Without this requirement REQ-C-1 delivers a control that controls nothing.

**REQ-C-3** (State-driven) **While** a write-capable tool (`goal_arm`, `verify_snapshot`, `codex_task`, `codex_job_cancel`) is presented in the console, its control shall be visually and textually distinguished from the read-only tools, so the blast-radius difference is legible at the point of decision rather than only in a SPEC.

### M2 — codex authentication surface

**REQ-C-4** (Event-driven) **When** the user opens the MCP console section, the console shall display the codex state the `codex_setup` probe reports — `installed`, `binary`, `version`, `auth_provider` (ChatGPT / apiKey / provider / unknown), `enable_review_gate`, `allow_write` — reading through the same probe rather than reimplementing the classification.

**REQ-C-5** (Capability gate) **Where** codex is installed but unauthenticated, the console shall surface the remediation as an instruction to run the codex login command, and shall not attempt to perform the login itself.

**REQ-C-6** (Ubiquitous) The codex opt-in toggles (`review_gate.enabled`, the write-mode opt-in) shall be written through the same typed-config seam that the existing fail-closed readers consume (`readCodexReviewGateEnabled`, `readCodexTaskAllowWrite`, `internal/cli/mcp_codex.go:1156-1157`), so the console and the gates read one source of truth.

### M3 — GLM key + i18n + secret hygiene

**REQ-C-7** (Ubiquitous) The GLM API-key surface shall reuse `internal/web/glmkey.go` unchanged: the out-of-schema credential field (`glmAPIKeyFormField`), the bounded disclosure (`computeGLMKeyHint` — configured-boolean plus final four characters only, and no disclosure at all for a key of four characters or fewer), the validation rules, and the `glmKeyRevealPath` endpoint. No second credential path is authored.

**REQ-C-8** (Unwanted) No surface introduced by this SPEC shall write a credential into a git-tracked file. `.mcp.json` **is** git-tracked; the `${VAR}`-literal convention exists precisely for this, and any env value that must reach the runtime is expressed as a `${VAR}` literal expanded at load, never as a resolved value.

**REQ-C-9** (Ubiquitous) Every user-facing string added by this SPEC shall carry all four locales (en / ko / ja / zh) in `internal/web/assets/i18n.js`, satisfying the existing governance suite's key-coverage, orphan, and identity checks (`internal/web/i18n_governance_test.go`).

**REQ-C-10** (Unwanted) The console shall not fork any interpreter it consumes. The audit-backend resolver and `template.ResolveAgentModelEffort` remain defined once outside `internal/web`, as the existing guards assert (`internal/web/mcp_audit_surface_test.go:47-95`).

## §D. Constraints

- **C-C-1** — Reuse the schema + `yamlpatch` seam; do not author a bespoke settings path for MCP.
- **C-C-2** — The GLM credential stays out of `settings.AllFields()`, preserving the structural anti-leak guarantee documented at `internal/web/glmkey.go:11-17`.
- **C-C-3** — `.templ` sources are edited and regenerated; the generated `_templ.go` is committed alongside.
- **C-C-4** — No new HTTP route accepts a credential in a URL or a GET query.
- **C-C-5** — The 17-tool list has exactly one declaration site shared by the server and the console (REQ-C-1).

## §E. Risks

- **R-C-1 — A control that controls nothing.** The highest risk of this SPEC is shipping M1 without REQ-C-2's gating seam: a console toggle that writes a setting no code reads. REQ-C-2 exists to make that failure impossible to ship silently, and AC-C-004 tests the effect end-to-end rather than testing that the setting persisted.
- **R-C-2 — Credential leak through a generic form loop.** A bulk schema-walking read or a diagnostics view that picks up a credential field is the failure `glmkey.go` was designed against. Mitigated by C-C-2 and REQ-C-7's reuse-don't-reauthor rule.
- **R-C-3 — Disabling a tool an agent depends on.** SPEC-B grants agents specific tools; disabling one in the console makes an agent's declared tool inert. Accepted as user intent; REQ-C-3's write-capable distinction is what makes the consequential cases legible.
- **R-C-4 — i18n drift.** Four locales and a governance suite that fails on any gap. Mitigated by REQ-C-9 and by adding all four in the same change.

## §F. Exclusions

### Out of Scope — codex OAuth execution

- No browser-based login flow is launched from the console. State is reported; login is the user's command to run (REQ-C-5, `plan.md` §C.2).

### Out of Scope — third-party MCP servers

- `context7`, `chrome-devtools`, `playwright`, `ast-grep`, and the z.ai servers keep their existing CLI surfaces (`moai mcp add`, `moai glm tools enable`). This console governs the `moai` server only.

### Out of Scope — MCP tool behavior

- No handler logic, schema, or `internal/` integration point of any of the 17 tools is changed. Only whether a tool is registered.

### Out of Scope — the live WebSocket board

- Unchanged and out of scope, as in its own deferred SPEC.
