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

### M5 — Local pin + live delivery proofs

**Local pin (MF1 durability)** — `.moai/config/sections/workflow.yaml`
(tracked): NEW `audit:` block, `codex: {model: gpt-5.6-sol, effort: high}`,
`glm: {model: glm-5.3, effort: max}`.

**Live gate harness** — `internal/cli/audit_pin_live_test.go` (NEW,
opt-in `MOAI_AUDIT_PIN_LIVE=1`): AC-AMP-006 through the glmHTTPClient tee
(raw envelope capture) and AC-AMP-007 through the probeRunner codexSession
tap (transmitted-NDJSON capture). SKIP markers per MF6 on absent
credential/binary. Evidence under `.moai/state/verify/t225/`.

#### AC-AMP-006 — GLM reasoning-delivery differential: FAIL under the 2.0 rule (lead decision required)

Four live attempts (all pin-resolved: both runs flow workflow.yaml pin →
resolveGLMAuditModelEffort → callGLMAudit → live z.ai):

| attempt | delivery field | fixed diff | S(max) | S(low) | ratio | verdict(s) |
|---|---|---|---|---|---|---|
| 1 | A: thinking {type:enabled, budget 3072/1024} | 19-line literal | 11533 | 11288 | 1.02 | inconclusive×2 (parser bug, see below) |
| 2 | B: top-level reasoning_effort | 19-line literal | 253 | 189 | 1.34 | fail/fail |
| 3 | B | 80-line literal | 1014 | 547 | 1.85 | fail/fail |
| 4 | B | real M2 diff 63e10bc1b..HEAD (mcp_codex.go) | 2456 | 1662 | 1.48 | fail/fail |

Supporting usage deltas (output_tokens max vs low): attempt 1: 3667/3480
(1.05); attempt 2: 513/367 (1.40); attempt 3: 1275/812 (1.57); attempt 4: —
(see evidence). z.ai's usage envelope carries NO reasoning-token field — S
fell to the thinking-block byte fallback in every attempt.

- **Numeric rule (S(max) ≥ 2.0 × max(S(low),1)) NOT met** — the AC cannot
  count PASS as written. Evidence:
  `.moai/state/verify/t225/ac-amp-006-glm-differential{,-attempt1..4}.md`.
- **The rule's embedded diagnosis ("field ignored") is CONTRADICTED by the
  measurements**: hypothesis A's budget_tokens measured a true null (1.02 —
  ignored, both attempts at ~3.5K output tokens regardless of 3072 vs 1024
  budget), while hypothesis B's top-level reasoning_effort produces a
  REPEATABLE DIRECTIONAL differential (1.34 / 1.85 / 1.48 across three
  difficulty-calibrated targets). The delivery field selected by the live
  evidence is **hypothesis B — top-level reasoning_effort, state names
  verbatim**; it is honored by the endpoint. The endpoint's max-vs-low delta
  on audit-shaped targets does not reach 2x on any instrument tried.
- **This also RESOLVES the prior overlay measurement in the OPPOSITE
  direction** (spec.md §H assumed thinking honored / reasoning_effort
  ignored): live evidence shows reasoning_effort honored (directional),
  thinking.budget_tokens ignored. Correction candidate for the llm.yaml
  overlay note + AC-MTP-032b — out of this SPEC's scope, lead may card it.

**Production bug found + fixed by the live gate (attempt 1)** — with a
delivered reasoning directive, z.ai prefixes the response with THINKING
content blocks; `parseGLMReview` read `Content[0].Text` blindly (empty for a
thinking block) and failed open to inconclusive while a full review sat in
the next block. Fixed: parse the first `type:"text"` block
(`TestGLMAuditParse_SkipsLeadingThinkingBlock`). Attempts 2-4 returned real
verdicts through the fixed parser.

**M3 wire revision (plan.md M3 contingency exercised)** — delivery field
switched A → B per the plan's instruction ("if the differential shows no
delivery, switch the field (hypothesis A ↔ B), re-run"):
`glmMessagesRequest.ReasoningEffort` (top-level, verbatim state, omitted for
invalid/empty values); the thinking object + per-state budget constants were
REMOVED. Unit tests updated to the selected field.

**Command:**

```
MOAI_AUDIT_PIN_LIVE=1 go test ./internal/cli/ -run TestAuditPinLive_GLMDifferential -count=1 -v -timeout 12m
(attempt 4) ... AC-AMP-006 differential: FAIL under the 2.0 rule (ratio 1.48) — NOTE ... (S(max)=2456 S(low)=1662 ratio=1.48)
```

#### AC-AMP-007 — live codex pin confirmation: PASS

One REAL codex_audit turn (adversarial mode, binary
`/Users/goos/.local/bin/codex` codex-cli 0.149.0, cwd = this worktree,
600 s wall) with the codexSession tap recording every transmitted NDJSON
line. The tracked pin resolved audit-scoped (`resolveCodexAuditModelEffort`
→ `{gpt-5.6-sol high}`) and reached BOTH seams on the live wire:

```json
{"method":"thread/start","params":{"cwd":".../worktrees/t225","model":"gpt-5.6-sol"}}
```

- `"model":"gpt-5.6-sol"` — 2 transmitted occurrences (thread/start +
  turn/start); `"effort":"high"` — 1 (turn/start)
- review verdict: pass (informational — the AC gates on transmitted params)
- VERDICT: PASS — the model choice no longer depends on
  `~/.codex/config.toml`; evidence
  `.moai/state/verify/t225/ac-amp-007-codex-pin.md` (+ the full transcript
  `ac-amp-007-codex-transcript.ndjson` under the probe's report dir)
- No tree side effects: `git status` after the run showed only this SPEC's
  own files; codex staged its proposed overlay under /tmp only (readOnly
  sandbox honored)

```
$ MOAI_AUDIT_PIN_LIVE=1 go test ./internal/cli/ -run TestAuditPinLive_CodexPinConfirmation -count=1 -v -timeout 15m
--- PASS: TestAuditPinLive_CodexPinConfirmation (600.02s)
ok  	github.com/modu-ai/moai-adk/internal/cli	600.832s
```

#### M5 close-out verification

```
$ go vet ./internal/cli/
exit 0
$ go test ./internal/cli/ -run "TestGLM|TestCodex|TestMCP|TestAudit|TestMultiAudit|TestConvergence" -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	15.887s   (4 of 5 runs green; 1 unreproduced FAIL — see §E.3 residual-risk)
$ go test ./internal/config/ ./internal/settings/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/config	3.500s
ok  	github.com/modu-ai/moai-adk/internal/settings	0.998s
$ grep -rn "gpt-5.6-sol" internal/template/templates/ | wc -l
0
```

**Changed files (M5):** `.moai/config/sections/workflow.yaml` (local tracked
pin), `internal/cli/mcp_glm.go` (hypothesis B switch + parser fix),
`internal/cli/mcp_glm_audit_pin_test.go` (field switch + parser regression
test), `internal/cli/audit_pin_live_test.go` (NEW — the live gates).

### §E consolidated AC matrix (E1)

| AC | Status | Verification command | Observed output (verbatim tail) |
|---|---|---|---|
| AC-AMP-001 [MUST] | PASS | `go test ./internal/config/ -run TestAuditConfigYAML -count=1` | `ok github.com/modu-ai/moai-adk/internal/config 0.460s` |
| AC-AMP-002 [MUST] | PASS | `go test ./internal/cli/ -run "TestCodexAuditPin" -count=1` (6 tests: both seams + MF2 negative) | `ok github.com/modu-ai/moai-adk/internal/cli 1.569s` |
| AC-AMP-003 [MUST] | PASS | same run (absent/empty/unservable states + C7 anchors `TestCodexSession_*`, `TestMCPAudit_NoDirectFrontmatterRead` green) | `ok ... 9.770s` (M2 broad sweep) |
| AC-AMP-004 [MUST] | PASS | `go test ./internal/cli/ -run "TestGLM" -count=1` (post-hypothesis-B: body carries reasoning_effort; medium omits; absent unchanged; glm_task isolated) | `ok github.com/modu-ai/moai-adk/internal/cli 5.829s` |
| AC-AMP-005 [MUST] | PASS | `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -count=1`; `grep -rn "gpt-5.6-sol" internal/template/templates/` | `ok ... 99.814s` / `0` matches |
| AC-AMP-006 [MUST] | **FAIL (rule unmet; evidence contradicts the rule's diagnosis — lead decision)** | `MOAI_AUDIT_PIN_LIVE=1 go test ./internal/cli/ -run TestAuditPinLive_GLMDifferential -count=1 -v` | attempt 4: `S(max)=2456 S(low)=1662 ratio=1.48`; attempts 1-3: 1.02/1.34/1.85 (see §M5 table) |
| AC-AMP-007 [MUST] | PASS | `MOAI_AUDIT_PIN_LIVE=1 go test ./internal/cli/ -run TestAuditPinLive_CodexPinConfirmation -count=1 -v` | `--- PASS (600.02s)`; wire: `"model":"gpt-5.6-sol"` ×2, `"effort":"high"` ×1 |
| AC-AMP-008 [SHOULD] | PASS | `go test ./internal/settings/ ./internal/web/ -count=1` | `ok ... 0.530s` / `ok ... 3.820s` |

**E2 (cross-platform build):** `go build ./...` exit 0;
`GOOS=windows GOARCH=amd64 go build ./...` exit 0.

**E3 (coverage):** not measured per-package in this run (Gaps — the added
code is resolver/request-body logic under seam tests; the 85% package
threshold is judged by CI's coverage job).

**E4 (subagent boundary):** `grep -rn 'AskUserQuestion' internal/cli/audit_pin.go internal/cli/mcp_glm.go internal/cli/mcp_codex.go internal/cli/audit_pin_live_test.go` → 3 matches, ALL pre-existing documentation comments ("It NEVER invokes AskUserQuestion..."); 0 code references (new code introduces none; the canonical comment-excluding guard shape and the existing `TestGLMAudit_NoAskUserQuestion` suite stayed green in-run).

**E5 (lint):** `go vet ./internal/config/... ./internal/cli/... ./internal/settings/... ./internal/web/...` → exit 0, no findings. golangci-lint not run locally (lane discipline: targeted checks; CI runs the full linter).

**E6 (branch HEAD + push state):** commits M1..M5 on `WT-audit-model-pin`; NO push (the lane owns push/PR after reading this §E, per dispatch).

**E7 (blockers):** one — AC-AMP-006's numeric rule vs measured evidence mismatch requires a lead decision (amend the threshold accepting the directional evidence, or keep the AC unmet). Reported here, not escalated via AskUserQuestion (subagent boundary).

**E8 (RED evidence):** verbatim pre-GREEN failing outputs recorded for M1/M2/M3/M4 in their sections above.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-24
run_commit_sha: pending-backfill-m5   # backfilled in the M5 follow-up commit
run_status: complete-with-one-unresolved-ac
ac_pass_count: 7
ac_fail_count: 1
# AC-AMP-006 FAILs under the SPEC's 2.0 numeric rule while its measured
# evidence CONTRADICTS the rule's embedded diagnosis: the selected delivery
# field (top-level reasoning_effort) IS honored (repeatable directional
# differential 1.34/1.85/1.48), and hypothesis A measured a true null (1.02).
# Closure per §D.4 requires the lead: accept the directional evidence via
# amendment (recalibrated threshold) or keep the AC unmet.
preserve_list_post_run_count: 0   # companion statusline patch files untouched (verified per commit diff)
l44_pre_commit_fetch: n/a-kanban-lane   # no push from this lane (lead owns push/PR)
l44_post_push_fetch: n/a-kanban-lane
new_warnings_or_lints_introduced: 0   # go vet clean on all touched packages
cross_platform_build:
  darwin: pass    # `go build ./...` exit 0 (this host, darwin/arm64)
  windows: pass   # `GOOS=windows GOARCH=amd64 go build ./...` exit 0
  linux: not-run-locally   # CI matrix judges; no OS-specific code added (filepath.Join, no syscall)
total_run_phase_files: 13
m1_to_mN_commit_strategy: per-milestone conventional commits (M1..M5), no push
```

**Residual-risk / Gaps (VCI §3.4-3.5):**

- AC-AMP-006 numeric threshold unmet (see above) — the ONLY unresolved AC.
- One unreproduced test flake: the first post-M5 sweep invocation of
  `go test ./internal/cli/ -run "TestGLM|TestCodex|TestMCP|TestAudit|TestMultiAudit|TestConvergence" -count=1`
  FAILED (7.5 s) with no failing-test line captured; 4 immediate re-runs all
  green (9.5 s / 101 s / 7.8 s / 15.9 s). Not diagnosed; CI full-suite is the
  arbiter.
- Live GLM runs are single-sample-per-attempt (no repetition for variance);
  the directional consistency across 4 attempts is the evidence, not
  within-attempt statistics.
- The M5 evidence files live in `.moai/state/verify/t225/` (gitignored
  runtime state) — cited paths resolve on THIS machine at audit time; the
  numeric tables above preserve the load-bearing figures in the tracked
  record.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase completion>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
