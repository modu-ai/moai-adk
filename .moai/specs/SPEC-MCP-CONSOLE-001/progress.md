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

### M4 — GLM key surface (REQ-C-7)

**What landed:**

- MODIFIED `internal/web/fieldsets.templ` (+ regenerated `fieldsets_templ.go`) — NEW `glmKeyStateBlock` templ component rendered inside the MCP console section (`fieldsetMCP`) after `codexAuthBlock`. The block surfaces the GLM key STATE only: configured/not-configured boolean + the bounded trailing-four hint. It consumes `view.GLMKeyConfigured` / `view.GLMKeyHint` (already populated by the pre-existing `populateGLMKeyHint` → `computeGLMKeyHint` path authored by SPEC-GLM-KEY-INPUT-001). No second credential path, no submit, no reveal — those live in the 3rd Party LLM section (`fieldsetGLMKey`); M4 is state-only (REQ-C-7).
- MODIFIED `internal/web/assets/i18n.js` — 5 new keys × 4 locales (en/ko/ja/zh) for the GLM key state block (`sec.mcp.glm_key.title`, `f.mcp.glm_key.{status,configured,hint,not_configured}`).
- NEW `internal/web/mcp_glmkey_surface_test.go` — AC-C-010/011 tests.

**Key finding (SPEC SSOT over task assumption):** the task description assumed the GLM key classifier lives in `internal/cli/` and needs a probe extraction mirroring M3's `ProbeCodexSetup`. The code and plan.md §M4 say otherwise: `glmkey.go` is in `internal/web/` (not `internal/cli/`), and `computeGLMKeyHint()` already lives there, authored by SPEC-GLM-KEY-INPUT-001. plan.md §M4 ("Reuse glmkey.go as-is") and AC-C-010 (`git diff ed70e4354 -- internal/web/glmkey.go` shows no logic change) govern. No probe extraction, no new DI seam — M4 surfaces the existing view-model state in the MCP section.

**TDD cycle:** RED (`undefined: glmKeyStateBlock` — 6 new tests fail to compile) → GREEN (templ component + i18n keys + generated code; all 6 M4 tests PASS) → fix (no-fork guard test refined to exclude generated `_templ.go` comment leakage; `glmcred.Load(` confined to glmkey.go).

**Reuse proof (AC-C-010):** `git diff ed70e4354 -- internal/web/glmkey.go` → 0 lines changed (byte-identical). `TestGLMKeyField_AbsentFromSchema` still PASSes (field absent from `settings.AllFields()`). `TestGLMKeyHint_TrailingFourOnly` + `TestGLMKeyHint_ShortKeyDisclosesNothing` still PASS (AC-C-011 disclosure bounded — both branches).

### AC PASS/FAIL matrix (M4 scope)

| AC | Status | Evidence |
|----|--------|----------|
| AC-C-010 (GLM credential path reused unchanged) | PASS | `git diff ed70e4354 -- internal/web/glmkey.go` → 0 lines; `TestGLMKeyField_AbsentFromSchema` PASS; `TestAC_C_010_GLMKeyStateSurfacedInMCPSection` + `_RenderedViaRoute` + `_NoSecondCredentialPathInWeb` — `glmcred.Load(` confined to glmkey.go |
| AC-C-011 (disclosure stays bounded) | PASS | `TestGLMKeyHint_TrailingFourOnly` + `TestGLMKeyHint_ShortKeyDisclosesNothing` (unchanged) PASS; `TestAC_C_011_HintShownWhenConfiguredAndLong` + `_NoCharactersForShortKey` + `_NotConfiguredState` PASS |

### Out-of-M4 (deferred to M5)

AC-C-012 (no credential in git-tracked file — full secret-hygiene sweep), AC-C-013 (4-locale governance re-assertion — the 5 M4 keys already pass), AC-C-014 (no-forked interpreter guards re-asserted) — M5 scope. (Note: the 5 new M4 i18n keys already pass the governance suite `TestI18n`, and the no-fork guards `mcp_audit_surface_test.go` are untouched.)

### M5 — i18n + secret-hygiene sweep + no-fork re-assertion (REQ-C-8, REQ-C-9, REQ-C-10)

**REQ/AC label correction.** The delegation prompt's REQ mapping (REQ-C-8=i18n, REQ-C-9=secret-hygiene, REQ-C-10=no-fork) does NOT match the acceptance.md §D.1 traceability table, which is authoritative: REQ-C-8 → AC-C-012 (secret-hygiene), REQ-C-9 → AC-C-013 (i18n), REQ-C-10 → AC-C-014 (no-fork). M5 proceeded against the ACs (the SSOT), which are unaffected by the label mix-up.

**What landed:**

- NEW `internal/web/mcp_secret_hygiene_test.go` — the consolidated AC-C-012 cross-surface secret-hygiene sweep (4 tests). Characterization test (DDD PRESERVE) over an invariant that already holds: M3/M4 were designed around the no-credential-in-view-model rule and C-C-2 keeps the GLM credential out of `AllFields()`. The sweep exists so a future change that smuggles a credential field into a view model, re-introduces the credential into the schema, or interpolates a credential reader directly inside a template fails here rather than leaking at render time.
  - `TestAC_C_012_NoCredentialFieldInCodexStateView` — reflect-walks `CodexStateView`; no field matches a credential-name fragment (token/secret/credential/password/rawkey/fullkey/apikey). Proves M3's view model carries state/enum/hint only.
  - `TestAC_C_012_GLMViewModelIsBoundedPairOnly` — reflect-walks `pageView`'s GLM-prefixed fields; asserts they are exactly the bounded pair `{GLMKeyConfigured bool, GLMKeyHint string}` (allowlist, not just denylist — a smuggled third GLM field fails even if its name dodges the fragment denylist).
  - `TestAC_C_012_GLMCredentialAbsentFromAllFields` — `settings.AllFields()` contains no `glm_api_key` / `apikey`-named field; the C-C-2 structural anti-leak guarantee, consolidated at the M5 cross-surface level (TestGLMKeyField_AbsentFromSchema remains as the SPEC-GLM-KEY-INPUT-001 anchor).
  - `TestAC_C_012_NoTemplateInterpolatesCredentialReader` — no `.templ` source references the `glmcred` package; templates must consume the bounded view model, never read the credential directly (closes the template-surface bypass of the no-second-path guard).

**i18n (AC-C-013) — no key fills needed.** M2 (17 controls), M3 (10 codex keys), M4 (5 GLM keys) each added all 4 locales (en/ko/ja/zh) in the same change, avoiding AP-C-5. The full 14-test governance suite (`TestI18nKeySetParity`, `TestI18nKeyCoverageForward`/`Reverse`, `TestI18nUntranslatedValues`, `TestI18nAllowlistNoOrphans`, `TestI18nEndonymInvariants`, ...) PASS — 4-locale coverage was already complete at M5 entry. M5 re-asserts it (the verification command in acceptance.md `go test ./internal/web/... -run 'TestI18n'` → PASS).

**No-fork guards (AC-C-014) — still PASS.** M2/M3/M4 were designed around the guards at `mcp_audit_surface_test.go:47-95`. Re-asserted at M5: `TestWebConsole_AuditNoForkedInterpreter`, `TestWebConsole_ResolveAgentModelEffortSSOTShared`, `TestSchemaSurfaces_AuditSelection` all PASS. The codex no-second-classifier guard (`grep -rn 'classifyCodexAuth\|codex login status' internal/web/ --include=*.go | grep -v _test.go`) returns 0 matches.

**TDD discipline (E9).** No RED applies: M5 is a pure-sweep characterization of an invariant that already holds (DDD PRESERVE, not TDD RED-GREEN). The 4 new AC-C-012 tests PASS on first run because M3/M4 never leaked a credential — the tests characterize that correct existing behavior so a future regression fails loudly. No new behavior was introduced that could have a failing-RED state.

### AC PASS/FAIL matrix (M5 scope)

| AC | Status | Evidence |
|----|--------|----------|
| AC-C-012 (no credential in git-tracked file) | PASS | `TestAC_C_012_*` (4/4 PASS); `grep -rEc '(sk-\|ghp_\|Bearer [A-Za-z0-9])' .mcp.json internal/template/templates/.mcp.json` → `0` / `0`; `TestMCPNeutralityTemplateShape` PASS; `glmcred.Load(` confined to glmkey.go; codex no-second-classifier grep 0 matches |
| AC-C-013 (4-locale coverage complete) | PASS | all 14 `TestI18n*` governance tests PASS (key-set parity en=ko=ja=zh; no orphan; endonym invariants hold); no M5 key fills needed (M2/M3/M4 added all 4 locales together) |
| AC-C-014 (no forked interpreter) | PASS | `TestWebConsole_AuditNoForkedInterpreter` + `TestWebConsole_ResolveAgentModelEffortSSOTShared` + `TestSchemaSurfaces_AuditSelection` PASS; guards at `mcp_audit_surface_test.go:47-95` untouched |

### Run-phase regression (M1-M4 ACs re-asserted at M5)

AC-C-001..011 all still PASS: M1 (`TestAC_C_004_*` 3/3, `TestAC_C_005_*`, `TestMoaiMCPTools_*` 4/4, `TestMCPConsoleRendersAllTools`/`_ToolCountMatchesCatalog`/`_WriteCapableTextDistinction`), M3 (`TestAC_C_006..009_*`), M4 (`TestAC_C_010..011_*`, `TestGLMKeyHint_*`, `TestGLMKeyField_AbsentFromSchema`). No regression.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-08-13
run_commit_sha: pending-backfill-m5
run_status: ready
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: clean (worktree HEAD 9136c9345 → M5 commit; no parallel session race — worktree-isolated)
l44_post_push_fetch: deferred (Route B — push is manager-git's sync-phase job)
new_warnings_or_lints_introduced: 0 (golangci-lint 0 issues; baseline 0 → M5 0)
cross_platform_build.linux: PASS (go build ./... exit 0)
cross_platform_build.windows: PASS (GOOS=windows GOARCH=amd64 go build ./... exit 0)
total_run_phase_files: NEW internal/web/mcp_secret_hygiene_test.go (1 file); 0 production-code changes (M5 is test-only)
m1_to_mN_commit_strategy: per-milestone commits (M1..M4 already committed at 9136c9345); M5 = single M5 commit on plan/spec-mcp-console

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
