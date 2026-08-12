# progress.md — SPEC-MCP-CONSOLE-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-12
tier: M
artifacts: spec.md, plan.md, acceptance.md (+ this progress.md)
depends_on: SPEC-MCP-AGENT-WIRING-001
open_clarifications: 0 — both resolved 2026-08-12 via owner decision (D1: `.moai/config/sections/mcp.yaml` + new `SectionMCP`; D2: all-enabled). Ready for Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

### M1 — The gating seam (REQ-C-2)

**What landed:**

- NEW `internal/mcp/catalog.go` — single shared declaration of the 17-tool surface (`ToolDef{Name, WriteCapable}` + `MoaiMCPTools()`). Consumed by both the server (registration) and the schema (per-tool fields). @MX:ANCHOR tagged.
- MODIFIED `internal/settings/schema.go` — new `SectionMCP` section ID, added to `AllSections()`.
- MODIFIED `internal/settings/schema_sections.go` — `mcpFields()` generates one bool per catalog tool at path `mcp.tools.<name>.enabled`, derived from the shared catalog (no second tool list).
- MODIFIED `internal/settings/sectionroute.go` — `"mcp": RouteSeam` (seam-writable).
- MODIFIED `internal/settings/sectionwrite.go` — `"mcp"` root key registered in `sectionRootKeys`.
- MODIFIED `internal/cli/mcp_server.go` — `registerMoaiMCPTools(s, projectDir)` reads per-tool enablement once at construction via `readMCPToolEnablement`; each `add(name, tool, handler)` is gated on `enabled[name]`. Default all-enabled (owner decision). Fail-OPEN reader (inverse of codex gates' fail-CLOSED posture).
- NEW `internal/settings/testdata/sections/mcp.yaml` — fixture for the seam round-trip test.
- MODIFIED `internal/web/schema_label_test.go` + `schema_render_test.go` — SectionMCP exempted from web i18n/render parity until M2 renders the console section (precedent: statusline, git_strategy, M3-removed sections).

**TDD cycle:** RED (AC-C-004 disabled-tool-absent FAILS, AC-C-005 schema-fields-exist FAILS) → GREEN (all 5 M1 tests PASS) → REFACTOR (added malformed-yaml fail-open edge test).

**Single declaration (REQ-C-1 / AP-C-4):** the catalog is the one list. `TestMoaiMCPServer_RegistrationMatchesCatalog` asserts set-equality between tools/list and the catalog. `mcpFields()` derives the schema from the same catalog. A tool added to registration without a catalog entry fails the build.

### AC PASS/FAIL matrix (M1 scope)

| AC | Status | Evidence |
|----|--------|----------|
| AC-C-004 (disabled tool absent from tools/list) | PASS | `TestAC_C_004_DisabledToolAbsentFromToolsList` + `_ReadonlyTool` + `_NoConfig_AllEnabled` — 3/3 PASS |
| AC-C-005 (setting round-trips through schema seam) | PASS | `TestAC_C_005_PerToolFieldsInSchema` — 17/17 fields present in `AllFields()`; seam route + root key registered |
| (guard) AC-C-001/002 single-declaration | PASS | `TestMoaiMCPServer_RegistrationMatchesCatalog` + `internal/mcp/catalog_test.go` — catalog == registration, 17 tools, 4 write-capable |

### Out-of-M1 (deferred to M2-M5)

AC-C-001/002/003 (console rendering), AC-C-006..009 (codex auth), AC-C-010/011 (GLM key), AC-C-012/013/014 (secret hygiene + i18n + no-fork) — all M2-M5 scope, not touched by M1.

### M3 — codex authentication surface (REQ-C-4, REQ-C-5, REQ-C-6)

**What landed:**

- MODIFIED `internal/cli/mcp_codex.go` — extracted `ProbeCodexSetup(ctx)` returning a typed `CodexSetupResult` struct; `handleCodexSetup` now delegates to it. The classification logic (`classifyCodexAuth`) is untouched — consumed, not forked (AC-C-006). Both the MCP tool handler and the web console share one probe entry point.
- NEW `internal/web/codex_state.go` — `CodexStateView` view model + display-layer auth-provider constants + `defaultCodexStateProbe` (fail-open zero view for bare test apps).
- MODIFIED `internal/web/app.go` — added `codexStateProbe func(ctx) CodexStateView` DI field; wired default in `newApp`.
- MODIFIED `internal/web/server.go` — added `Config.CodexStateProbe` field; `NewServer` wires it into the app when provided.
- MODIFIED `internal/web/handlers.go` — `pageView` gains `CodexState`; `buildIndexView` calls the injected probe.
- MODIFIED `internal/web/schemaform.go` — `codexAuthProviderLabel` display helper; `partitionWorkflowFields` excludes codex toggle fields (rendered in MCP section, not workflow tabs); `isCodexToggleFieldName` predicate.
- MODIFIED `internal/web/fieldsets.templ` (+ regenerated `fieldsets_templ.go`) — `codexAuthBlock` renders the probe state (installed/binary/version/auth_provider) + login remediation (AC-C-007) + not-installed state (AC-C-008) + two opt-in toggles via the schema seam.
- MODIFIED `internal/settings/schema_sections.go` — two new SectionWorkflow seam fields: `workflow.codex.review_gate.enabled` and `workflow.codex.task.allow_write` (AC-C-009 — same path the fail-closed readers consume).
- MODIFIED `internal/cli/web.go` — wires `Config.CodexStateProbe` to a wrapper around `ProbeCodexSetup` (the DI adapter that lets web consume the probe without importing internal/cli).
- MODIFIED `internal/web/assets/i18n.js` — 10 new keys × 4 locales (en/ko/ja/zh) for codex auth display + toggle field labels.
- NEW `internal/web/mcp_codex_surface_test.go` — AC-C-006/007/008 tests.
- NEW `internal/cli/mcp_codex_consoletest_test.go` — AC-C-009 toggle round-trip test.

**TDD cycle:** RED (AC-C-006 probe-state-missing FAIL, AC-C-006 grep-guard FAIL on comment literal, AC-C-007 login-command-missing FAIL, AC-C-008 not-installed-missing FAIL, AC-C-009 yamlpatch-nil-parent FAIL) → GREEN (all 7 M3 tests PASS) → fix (partition exclusion for double-render, lint nil-context).

### AC PASS/FAIL matrix (M3 scope)

| AC | Status | Evidence |
|----|--------|----------|
| AC-C-006 (probe state displayed, not recomputed) | PASS | `TestAC_C_006_ProbeStateDisplayed` + `TestAC_C_006_NoSecondClassifierInWeb`; grep guard: `grep -rn 'classifyCodexAuth\|codex login status' internal/web/ --include=*.go \| grep -v _test.go` → 0 matches |
| AC-C-007 (unauthenticated → names command, no login) | PASS | `TestAC_C_007_UnauthenticatedNamesCommand` + `TestAC_C_007_NoAuthRouteInApp` |
| AC-C-008 (codex absent → graceful) | PASS | `TestAC_C_008_CodexAbsentGraceful` + `TestAC_C_008_DefaultProbeIsFailOpen` |
| AC-C-009 (toggles write the seam the gates read) | PASS | `TestAC_C_009_TogglesWriteTheSeamThatGatesRead` + `TestAC_C_009_ToggleOffRoundTrips` — ApplySchemaEdits → readCodexReviewGateEnabled/readCodexTaskAllowWrite round-trip |

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Input parameters:
- tier: M
- scope: ~11-13 files (fieldsets.templ + fieldsets_templ.go, handlers.go, app.go, schema_sections.go, mcp_server.go, new codex-auth view-model + test, new console test, assets/i18n.js, existing governance tests)
- domain count: 3 (internal/web console, internal/cli MCP server, internal/settings schema)
- file language mix: Go (.templ + .go) + JS (i18n.js) + markdown (SPEC)
- concurrency benefit: LOW (coding-heavy, per Anthropic's coding-task parallelism caveat)

Mode evaluation:
- Mode 1 trivial: not selected — multi-file semantic change, new gating seam
- Mode 2 background: not selected — write-capable implementation, not read-only
- Mode 3 agent-team: RETIRED (never selected)
- Mode 4 parallel: not selected — coding-heavy, not research-heavy (parallelism caveat)
- Mode 5 sub-agent: **selected** — coding-heavy Tier M, sequential per-milestone manager-develop delegation
- Mode 6 workflow: not selected — not high-volume mechanical (semantic, new-code, multi-rule)

Decision: sub-agent

Justification: coding-heavy Tier M implementation; per Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research"), sequential per-milestone manager-develop delegation (Mode 5) is the correct default. Tier M Section A-E delegation template applies. Progression mode: semi-autonomous (per-milestone checkpoint; owner may upgrade to autonomous ac_converge at any milestone).

Implementation Kickoff Approval: PASSED — owner approved run-phase entry 2026-08-12 (source_session_id 68fdc108). Plan-audit skip-eligibility note: plan-auditor PASS 0.95 (≥ Tier M 0.80), but plan.md was edited post-verdict to resolve the two clarification markers, so the plan-artifact hash changed — Phase 1 Plan Audit Gate will re-execute on /moai run (the marker resolution converted marked recommendations into adopted decisions; no REQ/AC changed, so the re-run is expected to re-PASS).
