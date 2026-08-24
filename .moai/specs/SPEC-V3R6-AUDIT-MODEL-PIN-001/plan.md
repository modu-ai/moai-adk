# plan.md — SPEC-V3R6-AUDIT-MODEL-PIN-001

## §A Context

Measured baseline (this tree, branch WT-audit-model-pin):

| Anchor | Observation |
|---|---|
| `internal/cli/mcp_codex.go:191` `codexSSOTModelEffort` | Resolves SSOT sync-auditor cell, filters via `codexServableModel` |
| `internal/cli/mcp_codex.go:158` `codexServableModelPrefixes` | `{gpt-, o1, o3, o4, codex}` — `opus` fails → pair dropped at `:200-202` |
| `internal/cli/mcp_codex.go:565-567` `openCodexSession` | `thread/start` carries `model` only (effort not legal on thread/start) |
| `internal/cli/mcp_codex.go:840-846` `buildCodexReviewParams` | `turn/start` carries `model` + `effort` |
| `internal/cli/mcp_glm.go:127` `resolveGLMAuditModel` | Model id only; delegates to shared `resolveGLMModelForAgent` |
| `internal/cli/mcp_glm.go:144-161` `resolveGLMModelForAgent` | Shared body — also serves `glm_task` (`internal/cli/glm_task.go:125`); non-GLM session falls back to `glmAuditDefaultModel` |
| `internal/cli/mcp_glm.go:102-107` `glmMessagesRequest` | No reasoning/effort field; own HTTP POST → env vars do not apply |
| `internal/cli/mcp_glm.go:217-256` `callGLMAudit` | Single request-assembly seam; both `glm_audit` and `audit_multi` (`internal/cli/mcp_convergence.go:460`) flow through it |
| `internal/config/types.go:236-289` `LLMConfig` | Target struct for the new `audit` sub-key |
| `internal/config/types.go:486` `WorkflowConfig.Audit` | Pre-existing `workflow.audit` (different section — no collision, do not touch) |
| `internal/settings/schema_sections.go` `llmFields()` | Typed web-field pattern for the llm section |
| `internal/web/assets/i18n.js` | Single 4-locale dictionary (en/ko/ja/zh); governed by `i18n_untranslated_allowlist_test.go` |
| `.moai/config/sections/llm.yaml` | Does NOT exist locally — the pin file is created new |

## §B Known Issues

1. The codex pin currently "works" only via `~/.codex/config.toml:2-3` local
   defaults — machine-dependent, invisible in the web console (the defect this
   SPEC removes).
2. z.ai reasoning delivery for the audit path is UNVERIFIED (AC-MTP-032b): two
   hypotheses — (A) Anthropic-style `thinking` object honored / (B) top-level
   `reasoning_effort` ignored. Prior overlay measurement supports (A), but not on
   this path. Resolved by the M5 live gate, not by assumption.
3. `callGLMAudit` signature change ripples to `mcp_convergence.go:460` — one
   additional caller to update (compile-enforced).

## §C Pre-flight

- [ ] `make build` clean on the base tree before edits
- [ ] Existing tests green: `go test ./internal/cli/ -run 'TestMCP|TestCodex|TestGLM'`
  (focused; full-suite verdict belongs to CI per repo discipline)
- [ ] Confirm the companion statusline patch files are staged/kept as-is (out of
  SPEC scope; verify untouched by every milestone diff)

## §D Constraints

- Decision-reversibility ordering: schema first (M1), resolver semantics (M2/M3),
  surfaces (M4), values + live proof (M5). Mechanical mirrors (catalog, templ
  regenerate) close each milestone that touches their inputs.
- `C4` (no direct frontmatter read) and `C7` (byte-identical fallback) guards from
  SPEC-MOAI-MCP-SERVER-001 must keep passing — the existing tests
  `TestMCPAudit_noDirectFrontmatterRead` and
  `TestCodexSession_ResolvedModelReachesTransmittedParams` are the named
  regression anchors.
- Template neutrality: no pinned model id under `internal/template/templates/`.

## §E Self-Verification

Run-phase §E evidence lives in progress.md. Verification commands per milestone:
`go build ./... && go vet ./internal/config/... ./internal/cli/... ./internal/settings/... ./internal/web/...`,
focused `go test` per touched package, `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...`
(neutrality), and the M5 live differential (see acceptance.md AC-AMP-006).

## §F Milestones

### M1 — Config schema: `llm.audit` (Priority High)
Data model first — everything downstream keys on it.
- `internal/config/types.go`: new `LLMAuditConfig struct { Codex ModelEffort \`yaml:"codex"\`; GLM ModelEffort \`yaml:"glm"\` }`
  + `LLMConfig.Audit LLMAuditConfig \`yaml:"audit"\`` (reuse `ModelEffort`; no new pair type).
- `internal/config/defaults.go`: `NewDefaultLLMConfig` leaves `Audit` at zero value
  (explicit comment: empty = fallback semantics, REQ-AMP-005).
- `internal/template/templates/.moai/config/sections/llm.yaml`: add `audit:` block
  with empty `codex:{model:"",effort:""}` / `glm:{model:"",effort:""}` + explanatory
  comment (precedence + backend-native effort vocabularies).
- Symmetry: verify `CONFIG_STRUCT_YAML_MISMATCH` guard passes
  (`go test ./internal/config/ -run TestStructYAMLSymmetry`).
- `make build` (catalog hashes) at milestone close.

### M2 — Codex resolver precedence (Priority High)
- `internal/cli/mcp_codex.go`: in `codexSSOTModelEffort` (or a thin wrapper above
  it — implementer's choice, keep the C4 comment contract), read the loaded
  `llm.Audit.Codex` FIRST: non-empty model AND `codexServableModel(model)` →
  return `{model, effort}` from the section. Otherwise fall through to the
  existing SSOT resolution unchanged.
- Rationale for keeping the servability filter on the section value (design
  decision D2): a config-file pin is persistent project state, unlike a per-call
  explicit `model` (which bypasses the filter by design); an unservable pinned
  model must degrade to the existing path, not break the review gate.
- Tests (`internal/cli/mcp_codex_test.go`): audit section present →
  `threadParams["model"]` + `out["model"]/out["effort"]` carry the pin (seam
  capture); section present but unservable model → falls back to SSOT path;
  section absent → byte-identical legacy behavior (C7 anchor).

### M3 — GLM model+effort resolution and wire delivery (Priority High)
- `internal/cli/mcp_glm.go`: new `resolveGLMAuditModelEffort() config.ModelEffort`
  — reads `llm.Audit.GLM` first (non-empty model → return pair verbatim; the pin
  is by construction a GLM id, so the `IsGLMBackend` session check does not apply
  — design decision D3); falls back to `resolveGLMModelForAgent(glmAuditAgentKey)`
  for the model with empty effort.
- `handleGLMAudit`: keep explicit caller `model` param precedence (overrides the
  section pin, mirroring codex `resolveCodexModelEffort`).
- Request delivery: extend `callGLMAudit` to accept the effort and set the
  reasoning field on `glmMessagesRequest`. Default implementation = hypothesis A
  (Anthropic-style `thinking` object), per the prior overlay measurement; the M5
  live gate is the arbiter — if the differential shows no delivery, switch the
  field (hypothesis B) and re-run. Effort value passes through as the z.ai state
  name (`max`); if a Claude-vocabulary value arrives, collapse via
  `CollapseClaudeEffortToGLM` (reuse, do not fork).
- Update the one additional caller `internal/cli/mcp_convergence.go:460`
  (compile-enforced).
- `glm_task` path untouched (REQ-AMP-008) — add a regression test asserting its
  resolution ignores `audit.glm`.
- Tests (`internal/cli/mcp_glm_test.go`): seam-captured outbound body asserts the
  reasoning field + model when the section is set; absent section → request body
  unchanged (no reasoning field, legacy model resolution).

### M4 — Web typed edit path (Priority Medium)
- `internal/settings/schema_sections.go` `llmFields()`: 4 new fields
  (`audit.codex.model` text, `audit.codex.effort` select — codex effort
  vocabulary; `audit.glm.model` text or select from `config.ValidGLMModels()`,
  `audit.glm.effort` select from `template.GLMReasoningStateNames()`), following
  the existing `typedField(SectionLLM, ...)` pattern. Effort selects are
  RUNTIME-APPLIED (unlike the stored-only tier efforts) — the field help must say
  so, mirroring REQ-WCR-033's labeling discipline.
- `internal/web/fieldsets.templ` + regenerate `fieldsets_templ.go` (templ
  generate); render the 4 fields in the 3rd Party LLM panel.
- `internal/web/assets/i18n.js`: title/desc keys for the 4 fields + panel note in
  ALL FOUR locales (en/ko/ja/zh); review against
  `i18n_untranslated_allowlist_test.go` governance contract.
- Round-trip test: save via typed path → reload → values persist.

### M5 — Local pin + live delivery proof (Priority High — gate)
- Create THIS project's `.moai/config/sections/llm.yaml` (new file):
  `audit.codex = {model: gpt-5.6-sol, effort: high}`,
  `audit.glm = {model: glm-5.3, effort: max}`. Committed to this repository
  (repo-local audit policy; outside template-neutrality scope).
- Live GLM differential per AC-AMP-006: same diff target, effort `max` vs `low`,
  observed z.ai response evidence (usage/reasoning-token delta or equivalent
  backend-observed signal) recorded under `.moai/state/verify/<session>/`.
- Live codex confirmation per AC-AMP-007: one real `codex_audit` (or MCP
  `mcp__moai__codex_audit`) call whose transmitted params / observed result show
  `gpt-5.6-sol`+`high` reaching codex (removing the config.toml accident).
- If the differential shows no delivery: switch the delivery field (M3
  hypothesis B), re-run. Evidence retained either way.

## §G Anti-Patterns

- Do NOT fork a second GLM effort vocabulary — reuse
  `GLMReasoningStateNames`/`CollapseClaudeEffortToGLM`.
- Do NOT apply the audit pin inside `resolveGLMModelForAgent` (would leak into
  `glm_task`) — apply it at the audit entry only.
- Do NOT bypass `codexServableModel` for the section value (D2 rationale above).
- Do NOT write pinned values into the template or Go defaults (neutrality).
- Do NOT treat "config was read" as "effort was delivered" — only the live
  differential closes AC-AMP-006.

## §H Cross-References

- spec.md §D (REQ-AMP-001..009), §F (Out of Scope — companion statusline patch
  must stay untouched).
- acceptance.md §D (AC matrix) — AC-AMP-001..007 trace to M1..M5 above.
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 — the M5 gate
  exists because a config-read assertion is not delivery evidence.
