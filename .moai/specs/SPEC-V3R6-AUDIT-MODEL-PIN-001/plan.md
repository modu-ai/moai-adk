# plan.md — SPEC-V3R6-AUDIT-MODEL-PIN-001

> Revision 1.1.0 — plan-audit iter 1 findings applied. Design relocated to the
> existing `workflow.audit` block per lead ruling C (MF1); codex_task isolation
> (MF2), real symmetry coverage (MF3), single-reading effort rule (MF4), numeric
> live-gate threshold (MF5), SKIP-on-missing-credential (MF6).

## §A Context

Measured baseline (this tree, branch WT-audit-model-pin):

| Anchor | Observation |
|---|---|
| `.gitignore:192` | `.moai/config/sections/llm.yaml` IGNORED (check-ignore rc=0) — an llm.yaml pin is uncommitable; workflow.yaml tracked (rc=1) |
| `internal/config/audit_models.go:59-64` | `AuditConfig { Model token; Gates AuditGates }` — the workflow.audit block being extended |
| `internal/config/types.go:486` | `Audit AuditConfig \`yaml:"audit"\`` on `WorkflowConfig` (the single correct citation; the :474 figure from the brief was the enclosing struct's earlier line) |
| `internal/cli/mcp_codex.go:191` `codexSSOTModelEffort` | Resolves SSOT sync-auditor cell, filters via `codexServableModel` |
| `internal/cli/mcp_codex.go:158` `codexServableModelPrefixes` | `{gpt-, o1, o3, o4, codex}` — `opus` fails → pair dropped at `:200-202` |
| `internal/cli/mcp_codex.go:565-567` `openCodexSessionOn` | `thread/start` carries `model` only; **shared by codex_task** (`internal/cli/codex_task.go:226` calls `openCodexSessionOn`) — MF2 leak path |
| `internal/cli/mcp_codex.go:840-846` `buildCodexReviewParams` | `turn/start` carries `model` + `effort`; also calls the shared `resolveCodexModelEffort` |
| `internal/cli/mcp_glm.go:127` `resolveGLMAuditModel` | Model id only; delegates to shared `resolveGLMModelForAgent` |
| `internal/cli/mcp_glm.go:144-161` `resolveGLMModelForAgent` | Shared body — also serves `glm_task` (`internal/cli/glm_task.go:125`); non-GLM session falls back to `glmAuditDefaultModel` |
| `internal/cli/mcp_glm.go:102-107` `glmMessagesRequest` | No reasoning/effort field; own HTTP POST → env vars do not apply |
| `internal/cli/mcp_glm.go:217-256` `callGLMAudit` | Single request-assembly seam; `glm_audit` and `audit_multi` (`internal/cli/mcp_convergence.go:460`) both flow through it |
| `internal/cli/glm.go:624` `loadLLMSectionOnly` | The section-only load pattern the new workflow helper mirrors |
| `internal/config/audit_struct_yaml_symmetry_test.go:26-32` | symmetryCases = 7 cases (4 MIG-003 sections + StatuslineConfig lineage) — **AuditConfig NOT covered** (MF3 verified) |
| `internal/settings/schema_sections.go:371-375` | Existing typed fields `workflow.audit.model` / `workflow.audit.gate.*` — the pattern the 4 new fields extend |
| `internal/web/schemaform.go:52` | Dedicated **Audit panel** (`ID: "audit"`) already exists; `isAuditFieldName` (`:164-166`) routes `workflow.audit.*`-prefixed fields onto it automatically |
| `internal/web/assets/i18n.js` | Single 4-locale dictionary (en/ko/ja/zh); governed by `i18n_untranslated_allowlist_test.go` |
| `.moai/config/sections/workflow.yaml` | Tracked, already carries the `audit:` block — the pin lands here (M5) |

## §B Known Issues

1. The codex pin currently "works" only via `~/.codex/config.toml:2-3` local
   defaults — machine-dependent, invisible in the web console (the defect this
   SPEC removes).
2. z.ai reasoning delivery for the audit path is UNVERIFIED (AC-MTP-032b): two
   hypotheses — (A) Anthropic-style `thinking` object honored / (B) top-level
   `reasoning_effort` ignored. Prior overlay measurement supports (A), but not on
   this path. Resolved by the M5 live gate with the numeric rule in AC-AMP-006,
   not by assumption.
3. `callGLMAudit` signature change ripples to `mcp_convergence.go:460` — one
   additional caller to update (compile-enforced).
4. The audit resolvers currently read ONLY llm.yaml (`loadLLMSectionOnly`); the
   pin now lives in workflow.yaml, so a workflow-section load helper is required
   (same pattern, same package) — see M1.

## §C Pre-flight

- [ ] `make build` clean on the base tree before edits
- [ ] Existing tests green: `go test ./internal/cli/ -run 'TestMCP|TestCodex|TestGLM'`
  (focused; full-suite verdict belongs to CI per repo discipline)
- [ ] Confirm the companion statusline patch files are untouched by every
  milestone diff (they are already committed; verify via `git status --short`)

## §D Constraints

- Decision-reversibility ordering: schema first (M1), resolver semantics (M2/M3),
  surfaces (M4), values + live proof (M5). Mechanical mirrors (catalog, templ
  regenerate) close each milestone that touches their inputs.
- `C4` (no direct frontmatter read) and `C7` (byte-identical fallback) guards from
  SPEC-MOAI-MCP-SERVER-001 must keep passing — `TestMCPAudit_NoDirectFrontmatterRead`
  and `TestCodexSession_ResolvedModelReachesTransmittedParams` are the named
  regression anchors.
- Template neutrality: no pinned model id under `internal/template/templates/`.
- MF6: every live gate that needs an absent optional backend reports SKIP with an
  evidence-file marker; SKIP never counts as PASS.

## §E Self-Verification

Run-phase §E evidence lives in progress.md. Verification commands per milestone:
`go build ./... && go vet ./internal/config/... ./internal/cli/... ./internal/settings/... ./internal/web/...`,
focused `go test` per touched package, `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...`
(neutrality), and the M5 live differential (acceptance.md AC-AMP-006).

## §F Milestones

### M1 — Config schema: `AuditConfig` extension + load helper (Priority High)
Data model first — everything downstream keys on it.
- `internal/config/audit_models.go`: extend `AuditConfig` with
  `Codex ModelEffort \`yaml:"codex,omitempty"\`` and
  `GLM ModelEffort \`yaml:"glm,omitempty"\`` (reuse `config.ModelEffort`; no new
  pair type; empty = fallback semantics per REQ-AMP-005, stated in the struct
  comment).
- `internal/cli/glm.go` (or a sibling in the same package, implementer's choice):
  `loadWorkflowAuditSection(projectDir)`-shaped helper following the
  `loadLLMSectionOnly` (`internal/cli/glm.go:624`) pattern — reads
  `.moai/config/sections/workflow.yaml`, unmarshals the `workflow:` wrapper,
  returns `config.AuditConfig`; absent file → zero value (no error).
- `internal/config/audit_struct_yaml_symmetry_test.go`: ADD an `AuditConfig`
  symmetry case covering the new sub-keys (verified: the existing 7 cases exclude
  it — MF3). This is the guard AC-AMP-001 cites; it must FAIL if a struct field
  loses its yaml key or vice versa.
- `internal/template/templates/.moai/config/sections/workflow.yaml`: add empty
  `codex:{model: "", effort: ""}` / `glm:{model: "", effort: ""}` sub-keys under
  the existing `audit:` block + a precedence/vocabulary comment.
- `make build` (catalog hashes) at milestone close.

### M2 — Codex AUDIT-scoped resolver precedence + task isolation (Priority High)
- `internal/cli/mcp_codex.go`: introduce an audit-scoped resolution used ONLY on
  the `codex_audit` path — reads `AuditConfig.Codex` FIRST (non-empty model AND
  `codexServableModel(model)` → return the pinned pair), otherwise falls through
  to the existing `codexSSOTModelEffort` unchanged. Thread it through the two
  seams (`thread/start` `model`; `turn/start` `model`+`effort`) via an injected
  resolver (or equivalent seam) so `openCodexSessionOn` can serve both callers:
  the audit flow passes the audit-scoped resolver; `codex_task`
  (`internal/cli/codex_task.go:226`) keeps the legacy SSOT-only
  `resolveCodexModelEffort` (design decision D2 — MF2 isolation; a config-file
  pin is persistent project state and must not leak into delegation tasks).
- Keep the explicit caller `model` param precedence above the pin (mirrors the
  existing `resolveCodexModelEffort` rule).
- Keep `codexServableModel` on the section value (unservable pin → fall back to
  the SSOT path, never break the review gate).
- Tests (`internal/cli/mcp_codex_test.go`): pin present → audit path's
  `threadParams["model"]` + `out["model"]/["effort"]` carry the pin (seam
  capture); **pin present + `codex_task` turn → transmitted params carry NO pin**
  (the MF2 regression test); pin present but unservable → audit path falls back;
  pin absent → byte-identical legacy behavior (C7 anchor).

### M3 — GLM model+effort resolution and wire delivery (Priority High)
- `internal/cli/mcp_glm.go`: new `resolveGLMAuditModelEffort() config.ModelEffort`
  — reads `AuditConfig.GLM` first (non-empty model → return the pair); falls back
  to `resolveGLMModelForAgent(glmAuditAgentKey)` for the model with empty effort.
  The pin bypasses the `IsGLMBackend` session check (a pin is by-construction a
  GLM id; a wrong id degrades via the existing fail-open z.ai-4xx → inconclusive
  — design decision D3).
- Effort single-reading rule (MF4 / REQ-AMP-006): valid set is EXACTLY
  `{low, high, max}` (`template.GLMReasoningStateNames()`), transmitted verbatim;
  any other non-empty value → omit the reasoning directive (model pin still
  applies). NO collapse on this path — `CollapseClaudeEffortToGLM` is the
  vocabulary reference only.
- `handleGLMAudit`: keep explicit caller `model` param precedence (overrides the
  pin, mirroring codex).
- Request delivery: extend `callGLMAudit` to accept the effort and set the
  reasoning field on `glmMessagesRequest`. Default implementation = hypothesis A
  (Anthropic-style `thinking` object), per the prior overlay measurement; the M5
  live gate is the arbiter — if the differential shows no delivery, switch the
  field (hypothesis B) and re-run.
- Update the one additional caller `internal/cli/mcp_convergence.go:460`
  (compile-enforced). `glm_task` path untouched (REQ-AMP-008).
- Tests (`internal/cli/mcp_glm_test.go`): seam-captured outbound body asserts the
  reasoning field + model when the pin is set; invalid effort (`medium`) → no
  reasoning field but pinned model; pin absent → body unchanged (no reasoning
  field, legacy model resolution); a test asserts `glm_task` resolution ignores
  the pin.

### M4 — Web typed edit path: existing Audit panel (Priority Medium)
- `internal/settings/schema_sections.go`: 4 new fields following the existing
  `workflow.audit.*` typed pattern at `:371-375` —
  `audit.codex.model` (text), `audit.codex.effort` (select, codex effort
  vocabulary), `audit.glm.model` (select from `config.ValidGLMModels()`), 
  `audit.glm.effort` (select from `template.GLMReasoningStateNames()`).
- Panel routing: the fields' `workflow.audit.` prefix means `isAuditFieldName`
  (`internal/web/schemaform.go:164-166`) routes them onto the EXISTING Audit
  panel (`schemaform.go:52`, `ID: "audit"`) — verified wiring; extend that panel,
  NOT the 3rd Party LLM panel. Verify no schemaform registry change is needed
  (expected none; if the generic renderer needs a nudge, keep it inside the audit
  panel).
- `internal/web/fieldsets.templ` + regenerate `fieldsets_templ.go` (templ
  generate) if the audit panel's field loop does not already render typed
  selects generically; cite what was found at implementation time.
- `internal/web/assets/i18n.js`: title/desc keys (`f.workflow.audit.codex.*` /
  `f.workflow.audit.glm.*` naming, following the existing
  `f.workflow.audit.model.opt.` precedent) in ALL FOUR locales; review against
  `i18n_untranslated_allowlist_test.go` governance contract.
- Round-trip test: save via typed path → reload → values persist.

### M5 — Local pin + live delivery proof (Priority High — gate)
- Edit THIS project's TRACKED `.moai/config/sections/workflow.yaml`: under the
  existing `audit:` block set `codex: {model: gpt-5.6-sol, effort: high}` and
  `glm: {model: glm-5.3, effort: max}`. Committed with the SPEC's code (both the
  local edit and the template empty defaults are tracked files — the MF1
  durability requirement).
- Live GLM differential per AC-AMP-006 (numeric rule: S(max) ≥ 2.0 ×
  max(S(low), 1) on the SAME fixed diff); evidence under
  `.moai/state/verify/<session>/` with the SKIP marker protocol when
  `GLM_API_KEY` is absent (MF6).
- Live codex confirmation per AC-AMP-007: one real `codex_audit` /
  `mcp__moai__codex_audit` call whose transmitted params / observed result show
  `gpt-5.6-sol`+`high` reaching codex from the moai side (removing the
  config.toml accident). Codex binary/auth absent → SKIP marker (MF6).
- If the differential shows no delivery: switch the delivery field (M3
  hypothesis B), re-run. Evidence retained either way.

## §G Anti-Patterns

- Do NOT read the pin inside the shared `resolveCodexModelEffort` /
  `resolveGLMModelForAgent` bodies — both serve task paths (MF2; REQ-AMP-008).
- Do NOT fork a second GLM effort vocabulary or collapse values on this path —
  the valid set is the z.ai state names, single reading (MF4).
- Do NOT bypass `codexServableModel` for the pinned value.
- Do NOT write pinned values into the template or Go defaults (neutrality).
- Do NOT treat "config was read" as "effort was delivered" — only the live
  differential with the numeric rule closes AC-AMP-006.
- Do NOT place any new key in `.moai/config/sections/llm.yaml` — the file is
  gitignored and update-wiped (MF1 evidence).

## §H Cross-References

- spec.md §D (REQ-AMP-001..009), §F (Out of Scope — companion statusline patch
  must stay untouched).
- acceptance.md §D (AC matrix) — AC-AMP-001..008 trace to M1..M5 above.
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 — the M5 gate
  exists because a config-read assertion is not delivery evidence.
