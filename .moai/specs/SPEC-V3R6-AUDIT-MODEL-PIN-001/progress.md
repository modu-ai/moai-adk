# progress.md — SPEC-V3R6-AUDIT-MODEL-PIN-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-08-24
- plan-audit verdict: PASS, score 1.0 (iter 2/2; N1-N3 findings folded into
  spec/plan/acceptance revision 1.1.0 — MF1 schema relocation to workflow.audit,
  MF2 codex_task isolation, MF3 real symmetry guard, MF4 single-reading effort
  vocabulary, MF5 numeric live-gate rule, MF6 SKIP semantics)
- Implementation Kickoff Approval: granted by lead kickoff 2026-08-24 (run-phase
  entry operator-approved)

## §E.2 Run-phase Evidence

### M1 — Config schema: AuditConfig extension + load helper + template block

Branch `WT-audit-model-pin`, base HEAD `63e10bc1b`. TDD cycle (RED → GREEN).

**RED evidence (E8)** — `TestAuditConfigYAMLRoundTrip` authored BEFORE the
struct extension; verbatim failing output:

```
$ go test ./internal/config/ -run TestAuditConfigYAML -count=1
internal/config/audit_models_test.go:48:66: populated.Codex undefined (type AuditConfig has no field or method Codex)
internal/config/audit_models_test.go:50:10: back.GLM undefined (type AuditConfig has no field or method GLM)
...
too many errors
FAIL	github.com/modu-ai/moai-adk/internal/config [build failed]
```

`TestLoadWorkflowAuditSection` (internal/cli/audit_pin_test.go) likewise RED
(`undefined: loadWorkflowAuditSection`, build failed) before audit_pin.go.

**GREEN + verification:**

```
$ go test ./internal/config/ -run TestAuditConfigYAML -count=1
ok  	github.com/modu-ai/moai-adk/internal/config	0.460s
$ go test ./internal/cli/ -run TestLoadWorkflowAuditSection -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	1.131s
$ go test ./internal/config/ -count=1            (full package — incl. TestShippedConfigKeysHaveReaders)
ok  	github.com/modu-ai/moai-adk/internal/config	5.837s
$ go vet ./internal/config/... ./internal/cli/...
exit 0 (no findings)
$ MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	99.814s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	1.310s
?  	github.com/modu-ai/moai-adk/internal/template/scripts	[no test files]
$ make build
catalog.yaml updated successfully (12899 bytes) — binary rebuilt
exit 0
```

**Changed files (M1):**

- `internal/config/audit_models.go` — `AuditConfig` + `Codex`/`GLM`
  `config.ModelEffort` pin fields (REQ-AMP-001)
- `internal/config/audit_models_test.go` — NEW:
  `TestAuditConfigYAMLRoundTrip` (the AC-AMP-001 dedicated drift guard:
  marshal→unmarshal field-for-field equality + codex:/glm: key-drop arms) +
  `TestAuditConfigYAMLWorkflowWrapperLoad` (workflow: wrapper load contract)
- `internal/cli/audit_pin.go` — NEW: `loadWorkflowAuditSection(projectDir)`
  section-only loader (absent file → zero value, no error) +
  `workflowAuditPins` fail-open wrapper (N3)
- `internal/cli/audit_pin_test.go` — NEW: 4 subtests (populated verbatim /
  absent file / audit-less file / unparseable → error the caller fails open on)
- `internal/template/templates/.moai/config/sections/workflow.yaml` — NEW
  `audit:` block, EMPTY codex/glm sub-keys + precedence/vocabulary comment
  (REQ-AMP-005; `grep -rn "gpt-5.6-sol" internal/template/templates/` = 0
  matches)
- `internal/config/testdata/shipped_key_inventory.yaml` — 4 new
  `workflow.audit.{codex,glm}.{model,effort}` entries (class W, evidence
  reader), header count corrected to the measured 963

**AC trace:** AC-AMP-001 (round-trip guard + loader verbatim pins) — PASS at
M1 scope; AC-AMP-005 template half — PASS (empty defaults + leak-strict green).

### M2 — Codex audit-scoped resolver precedence + task isolation

**RED evidence (E8)** — `internal/cli/mcp_codex_audit_pin_test.go` authored
BEFORE the resolver/seam change; verbatim failing output:

```
$ go test ./internal/cli/ -run TestCodexAuditPin -count=1
internal/cli/mcp_codex_audit_pin_test.go:135:15: undefined: runCodexAuditReviewRPC
internal/cli/mcp_codex_audit_pin_test.go:163:15: undefined: runCodexAuditReviewRPC
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```

**GREEN + verification:**

```
$ go test ./internal/cli/ -run "TestCodexAuditPin|TestCodexSession|TestMCPAudit|TestLoadWorkflowAuditSection" -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	1.569s
$ go vet ./internal/cli/
exit 0
$ go test ./internal/cli/ -run "TestCodex|TestMCP|TestGLM|TestMultiAudit|TestConvergence|TestAudit" -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	9.770s
```

**Mechanism (design decision D2 — injected resolver, not shared-body read):**

- `resolveCodexAuditModelEffort` (mcp_codex.go) — pin-first: `workflowAuditPins`
  .Codex applies only when Model non-empty AND `codexServableModel` holds;
  explicit caller `model` still outranks the pinned model; everything else
  falls through to the untouched legacy `resolveCodexModelEffort`.
- `openCodexSessionResolved` (new core) accepts the injected resolver;
  `openCodexSessionOn` keeps its signature and passes nil — **codex_task.go is
  UNTOUCHED by M2** (isolation is structural: the task path cannot see the pin).
- `codexSessionHandle.resolveME` carries the resolver to
  `buildCodexReviewParams` (4th param) so BOTH seams resolve through it:
  `thread/start` model + `turn/start` model+effort.
- `runCodexReviewRPCResolved` core; `runCodexReviewRPC` (legacy, gate +
  existing tests) and `runCodexAuditReviewRPC` (pin-aware) variants;
  `var codexReviewRPC = runCodexAuditReviewRPC` — both audit callers
  (handleCodexAudit mcp_codex.go:1227, performCodexAudit mcp_convergence.go:425)
  get the pin; the Stop-hook review gate keeps the legacy resolution.

**Changed files (M2):** `internal/cli/mcp_codex.go` (resolver + seams),
`internal/cli/mcp_codex_audit_pin_test.go` (NEW — 6 tests).

**Test inventory (mcp_codex_audit_pin_test.go):**
`TestCodexAuditPin_ReachesTransmittedParams` (AC-AMP-002 positive: both seams
carry gpt-5.6-sol/high), `TestCodexAuditPin_TaskTurnCarriesNoPin` (MF2/REQ-AMP-008
negative: pin present + codex_task turn → SSOT pair, no pin),
`TestCodexAuditPin_UnservableFallsBackToSSOT` (AC-AMP-003 state c),
`TestCodexAuditPin_EffortAlonePinsNothing` (§D.2 edge),
`TestCodexAuditPin_ExplicitModelOverridesPin`, `TestCodexAuditPin_AbsentPinEqualsLegacyResolution`
(AC-AMP-003 states a/b, C7 anchor; the unmodified TestCodexSession_* /
TestMCPAudit_NoDirectFrontmatterRead suite stayed green in the same run).

**AC trace:** AC-AMP-002 — PASS (both arms incl. MF2 regression);
AC-AMP-003 — PASS (states a/b/c + anchors green).

### M3 — GLM model+effort resolution and wire delivery

**RED evidence (E8)** — `internal/cli/mcp_glm_audit_pin_test.go` authored
BEFORE the resolver/wire change; verbatim failing output:

```
$ go test ./internal/cli/ -run TestGLMAuditPin -count=1
internal/cli/mcp_glm_audit_pin_test.go:97:83: undefined: glmAuditThinkingBudgetMax
internal/cli/mcp_glm_audit_pin_test.go:163:8: undefined: resolveGLMAuditModelEffort
internal/cli/mcp_glm_audit_pin_test.go:178:8: undefined: resolveGLMAuditModelEffort
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```

**GREEN + verification:**

```
$ go test ./internal/cli/ -run "TestGLM|TestResolveGLMAuditModel" -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	24.199s
$ go vet ./internal/cli/
exit 0
$ go test ./internal/cli/ -run "TestCodex|TestMCP|TestGLM|TestMultiAudit|TestConvergence|TestAudit" -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	48.958s
```

**Mechanism:**

- `resolveGLMAuditModelEffort()` (mcp_glm.go) replaces `resolveGLMAuditModel`:
  pin-first (workflow.audit.glm, non-empty model → pair verbatim, bypassing
  the IsGLMBackend session check per design decision D3 — a wrong id degrades
  via the existing z.ai-4xx fail-open); fallback = legacy
  `resolveGLMModelForAgent(glmAuditAgentKey)` model with EMPTY effort.
- `handleGLMAudit`: explicit caller `model` outranks the pinned model; the
  pinned effort survives an explicit model override.
- Wire delivery (hypothesis A — the field the M5 live gate arbitrates):
  `glmMessagesRequest.Thinking *glmThinkingDirective` (json `thinking,omitempty`)
  carrying `{"type":"enabled","budget_tokens":B}`; per-state budgets
  1024/2048/3072 (all < glmAuditMaxTokens 4096; max:low = 3:1 vs the ≥ 2.0
  numeric rule).
- Single-reading rule (REQ-AMP-006): `glmAuditThinkingDirective(effort)`
  accepts EXACTLY {low, high, max} (template.GLMState*), verbatim; any other
  non-empty value → nil directive (omitted) while the model pin still applies;
  NO CollapseClaudeEffortToGLM on this path.
- `callGLMAudit(ctx, key, model, effort, focus, diff, token)` — signature
  extended; both callers updated: handleGLMAudit and performGLMAudit
  (mcp_convergence.go:460 — the audit_multi GLM leg, now pin-aware).
  `glm_task` path untouched (REQ-AMP-008).
- Two pre-existing tests updated to the new resolver surface
  (TestResolveGLMAuditModel_SSOT, mcp_glm_fallback_test.go x2 — same fallback
  semantics asserted through `.Model`).

**Changed files (M3):** `internal/cli/mcp_glm.go` (resolver + request body +
directive builder + budgets), `internal/cli/mcp_convergence.go` (performGLMAudit
caller), `internal/cli/mcp_glm_audit_pin_test.go` (NEW — 6 tests),
`internal/cli/mcp_glm_test.go` + `internal/cli/mcp_glm_fallback_test.go`
(resolver rename, semantics unchanged).

**Test inventory:** `TestGLMAuditPin_ReachesRequestBody` (AC-AMP-004 positive:
model + thinking{enabled,3072} on the captured body),
`TestGLMAuditPin_InvalidEffortOmitsReasoningDirective` (medium → model only),
`TestGLMAuditPin_AbsentPinBodyUnchanged` (no reasoning field, legacy model),
`TestGLMAuditPin_TaskResolutionUnaffected` (REQ-AMP-008: glm_task → SSOT
glm-4.6 under a glm-5.3 pin),
`TestGLMAuditPin_BypassesSessionBackendCheck` (D3),
`TestGLMAuditPin_ExplicitModelParamOverridesPin`.

**AC trace:** AC-AMP-004 — PASS (all three arms + task isolation).

### M4 — Web typed edit path: existing Audit panel

**RED evidence (E8):**

```
$ go test ./internal/settings/ -run TestAuditPinFields -count=1
    audit_pin_fields_test.go:62: field "workflow.audit.codex.model" missing from the schema (the audit pin must be web-editable)
FAIL	github.com/modu-ai/moai-adk/internal/settings	0.670s
```

**GREEN + verification:**

```
$ go vet ./internal/settings/... ./internal/web/...
exit 0
$ go test ./internal/settings/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/settings	0.530s
$ go test ./internal/web/ -count=1        (incl. i18n_untranslated_allowlist + tab layout governance)
ok  	github.com/modu-ai/moai-adk/internal/web	3.820s
```

**Mechanism:**

- `internal/settings/schema_sections.go`: 4 new fields after the audit gates —
  `workflow.audit.codex.model` (TypeText), `workflow.audit.codex.effort`
  (select, v4EffortValues 5-level Claude vocabulary),
  `workflow.audit.glm.model` (select, config.ValidGLMModels SSOT),
  `workflow.audit.glm.effort` (select, template.GLMReasoningStateNames —
  EXACTLY {max, high, low}). All three selects carry `withEmptySubmits` +
  emptyLabelUnset/"opt.unset" (clearing a pin persists "" → resolver falls
  back, §D.2). StoreOnly stays FALSE on the effort fields (runtime-applied,
  contrast with the stored-only llm tier efforts, REQ-WCR-033) — asserted by
  test.
- Panel routing: the `workflow.audit.` prefix means `isAuditFieldName`
  (schemaform.go:164-166) routes all four onto the EXISTING Audit panel
  (schemaform.go:52) — verified by the passing TestTabLayout; no schemaform
  registry change needed.
- **fieldsets.templ regen NOT needed** (cited finding): `grep -n "audit"
  internal/web/fieldsets.templ` = 0 matches — the Audit panel renders FieldDefs
  through the shared generic schema renderer (schemaSelectRow/schemaTextRow),
  so the new typed fields render with no templ change.
- `internal/web/assets/i18n.js`: 20 new keys × 4 locales (en/ko/ja/zh) — 8
  title/desc + 12 option labels. Effort descs mark them runtime-applied; glm
  effort desc states the single-reading rule (other values omit the reasoning
  directive). Governance green (i18n_untranslated_allowlist_test).
- `internal/web/widget_policy_test.go`: `workflow.audit.codex.model` added to
  the AC-WCR-023 freeTextWhitelist — the codex-servable family is an OPEN
  prefix set (gpt-*/o*/codex-*), not an enumerable closed set; the servability
  filter at the resolver is the validation point, not the widget.

**Changed files (M4):** `internal/settings/schema_sections.go`,
`internal/settings/audit_pin_fields_test.go` (NEW — 2 tests),
`internal/web/assets/i18n.js` (+80 lines), `internal/web/widget_policy_test.go`
(whitelist entry + rationale).

**AC trace:** AC-AMP-008 — PASS (typed-path round-trip incl. empty-value
clear; closed sets; 4-locale keys; runtime-applied labeling). SHOULD-severity
AC satisfied.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase completion>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
