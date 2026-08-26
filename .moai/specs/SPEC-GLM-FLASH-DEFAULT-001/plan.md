# plan.md — SPEC-GLM-FLASH-DEFAULT-001

## §A Context

Tier M implementation plan for card t289. All file anchors verified against tree 410da655f (origin/main tip) during plan phase — see spec.md §5. Milestones are ordered by decision-reversibility: the default-switch and set-shape decisions (most likely to change, user-visible) come first; mechanical mirror/docs steps last.

**The three operator-binding decisions encoded:**

1. **Flash max-only branch point** — the model-aware branch lands in the effort overlay (`internal/template/glm_effort_overlay.go`), the runtime SSOT, NOT in the yaml comment. Recommended shape: thread the resolved session model into the collapse chain (a model-parameterized variant of `CollapseClaudeEffortToGLM`, consumed by `ResolveGLMReasoning` / `SessionGLMReasoningStateForEffort`), returning the existing `glmReasoningMax` state whenever the model is `glm-5.3-flash` — including for Claude effort `low`. The existing model-unaware signature may remain as the glm-5.3-family delegate. Exact function shape is a run-phase design point; the behavioral requirement (REQ-004) is fixed.
2. **Tier-slot switch scope** — ALL slots switch: `DefaultGLMHigh/Medium/Low/Fable` AND legacy `DefaultGLMHaiku/Sonnet/Opus` aliases → `glm-5.3-flash`; template llm.yaml twin follows. glm-5.3 STAYS selectable: introduce a named constant for it (e.g. `DefaultGLM53`) and list it explicitly in `ValidGLMModels()` alongside the new flash constant, because the set is derived from the defaults constants and a bare retarget would drop glm-5.3 from the offered set (F-3).
3. **Web surface location** — `internal/settings/schema_sections.go:211-229` (tier-slot selects) and `:389-399` (audit GLM-model select) derive their options from `config.ValidGLMModels()`, so the Go side needs ONLY the closed-set change (M1). The hand-authored part is labels: `internal/web/assets/i18n.js` keys `f.llm.glm.models.opt.glm-5.3-flash` and `f.workflow.audit.glm.model.opt.glm-5.3-flash` in all four locale blocks (en/ko/ja/zh). `internal/web/glm_tier_test.go` asserts the exact 4-model set and must be updated to the 5-model set.

## §B Known Issues

- **i18n fallback risk**: a missing option-label key can render the raw key or empty label in the widget — label entries are mandatory, not cosmetic (REQ-007).
- **Longest-substring ambiguity**: `"glm-5.3-flash"` currently matches the `"glm-5.3"` table entry (F-2). Adding the explicit entry changes resolution order behavior only in the direct-hit sense; a table test must pin both the flash entry and the continued glm-5.3 entry.
- **Test-set coupling**: `glm_tier_test.go` (web) hard-asserts {glm-5.3, glm-5.1, glm-4.7, glm-4.5-air}; other tests may assert `DefaultGLM*` literal values — run-phase must grep for `"glm-5.3"` in `*_test.go` under internal/config, internal/web, internal/cli, internal/template before editing.

## §C Pre-flight

- [ ] Confirm worktree branch `WT-glm-flash-default` and clean status.
- [ ] Re-verify the anchors in spec.md §5 (they cite 410da655f; re-pin if HEAD moved).
- [ ] Grep test coupling: `grep -rn '"glm-5.3"' internal --include='*_test.go'` and inventory assertions to update.
- [ ] Overlay call-site inventory (display-vs-wire divergence guard): beyond the launcher, the overlay chain's consumers are `internal/web/agentfm.go:312` (`ResolveGLMReasoning`, display), `internal/cli/model.go:115` (`ResolveGLMReasoning`, display), `internal/cli/glm.go:428` (`SessionGLMReasoningStateForEffort`, wire). Run-phase M2 must thread the resolved model into all three, or state explicitly why a launcher/wire-only threading suffices — otherwise the display surfaces can show `low` for flash while the wire pins max.

## §D Constraints

- English in code/comments/templates; Conventional Commits.
- Template-First: template llm.yaml edit → `make build` → mirror verification. catalog.yaml unchanged. No SPEC-IDs/dates/SHAs in template content.
- Preserved: explicit glm-5.3 selection behavior; profile matrix untouched; overlay totality (unrecognized effort → max).
- No new request-params surface (spec.md §4 finding).
- Docs scope (grep-verified GLM-list locations): README.md / README.ko.md / README.ja.md / README.zh.md; docs-site per-locale `advanced/{config-sections,profile-matrix,statusline}.md` and `multi-llm/{_index,cg-mode}.md` (en/ko/ja/zh). Scope updates to lines that name the default model or enumerate the model set; do not rewrite adjacent content. 4-locale same-PR obligation applies.
- Testing discipline: affected packages only (`go test ./internal/<pkg>/...`), never the local full suite.

## §E Self-Verification (run-phase populates §E.2/§E.3 in progress.md)

Run-phase must produce, at minimum: affected-package test output including the new overlay and context-window table tests; grep evidence that no `"glm-5.3"` literal default remains in defaults.go tier constants; `make build` output; boot-smoke env observation (M6); template-neutrality check result.

## §F Milestones (decision-reversibility order)

### M1 — Registration + default switch (High priority; highest change likelihood)
- `internal/config/defaults.go`: add flash + glm-5.3 named constants; retarget `DefaultGLMHigh/Medium/Low/Fable` + legacy `DefaultGLMHaiku/Sonnet/Opus` to flash; update the surrounding comment blocks (English, no SPEC-ID).
- `internal/config/closed_sets.go`: `ValidGLMModels()` = {flash, glm-5.3, glm-5.1, glm-4.7, glm-4.5-air}; update doc comment.
- Template twin: `internal/template/templates/.moai/config/sections/llm.yaml` `llm.glm.models.*` literals + comments → flash.
- Update coupled tests (inventory from §C).
- Verification: `go test ./internal/config/... ./internal/settings/...` (schema derives options; existing web set test fails → update here).

### M2 — Overlay flash max-only (High priority)
- `internal/template/glm_effort_overlay.go`: model-aware flash branch per §A decision 1; update the collapse doc comment (add flash row: every effort → max).
- Unit tests: flash × {low, medium, high, xhigh, max, unrecognized} → max; glm-5.3 × low → low (mirror-image regression, non-flash unchanged).
- Template llm.yaml overlay comment block: add the flash rule (documentation twin).
- Verification: `go test ./internal/template/...`.

### M3 — Context window entry (Medium priority)
- `internal/statusline/memory.go`: explicit `"glm-5.3-flash": 1_000_000` entry + comment (divergence guard, F-2).
- Table test: flash → 1M direct; glm-5.3 → 1M retained. Registration-time guidance (not a testable deny — longest-substring semantics): an unregistered `glm-5.3-*` variant inherits 1M by substring; any future divergent `glm-5.3-*` id MUST add its own explicit table entry at registration time.
- Verification: `go test ./internal/statusline/...`.

### M4 — Web console surface (Medium priority)
- `internal/web/assets/i18n.js`: `f.llm.glm.models.opt.glm-5.3-flash` + `f.workflow.audit.glm.model.opt.glm-5.3-flash` in en/ko/ja/zh blocks (flash label should mention 1M context per sibling style).
- `internal/web/glm_tier_test.go`: expected set → 5 models.
- Verification: `go test ./internal/web/...`.

### M5 — Docs surfaces (Low priority)
- README 4-locale + docs-site pages listed in §D: default model naming updated (flash), glm-5.3 noted as selectable. 4-locale parity; no emoji; existing style matched.

### M6 — Build, smoke, template discipline (Low priority; mechanical, last)
- `make build`; template mirror verification (template ↔ embedded); template-neutrality self-check.
- Boot smoke (pinned observation — NO live z.ai API dependency; env-level observation only, no interactive prompt, no API round-trip): assert via the env-injection map that a `moai glm` session env resolves the flash model. Add a Go test in `internal/cli` that exercises the injection path (`buildTmuxInjectVars` with default slots, and the `setGLMEnv` process-env path applied to a scratch environment in a `t.TempDir()`-isolated test) and asserts the four `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL` values equal `glm-5.3-flash` and `EnvStatuslineContextSize` / `EnvClaudeCodeAutoCompactWindow` resolve 1000000. Command: `go test ./internal/cli/... -run TestGLMFlashDefaultEnvInjection -v`; expected output: `ok internal/cli ...` with the test quoting the four env values verbatim. Record the verbatim output in progress.md §E.2. The tmux-absent fallback (§D.3, glm.go `setGLMEnv`) is covered by the same test's process-env leg.

## §G Anti-Patterns

- Do NOT restate the model list as literals in a second place (closed set stays derived from constants; glm-5.3 gets its own constant rather than a duplicate string).
- Do NOT implement the flash rule in the yaml comment only (overlay is the runtime SSOT).
- Do NOT build a sampling-params surface (spec.md §4 finding).
- Do NOT touch the profile matrix or glm-5.2 handling.
- Do NOT run the full test suite locally; affected packages only, CI is the full-suite judge.

## §H Cross-References

- spec.md §3 REQ-001..REQ-010; acceptance.md AC-001..AC-013.
- Siblings: SPEC-GLM-EFFORT-MAX-001 (collapse ceiling), SPEC-GLM-MODEL-ALLOWLIST-001 (closed set discipline).
- CLAUDE.local.md §2 (Template-First), §25 (template-neutrality), §4.1 (lane-local verification).
