# acceptance.md — SPEC-GLM-FLASH-DEFAULT-001

## §D Acceptance Criteria (Given-When-Then)

### AC-001 — Closed set registration (REQ-001)
Given the tree with M1 applied, when `config.ValidGLMModels()` is called, then it returns exactly `["glm-5.3-flash", "glm-5.3", "glm-5.1", "glm-4.7", "glm-4.5-air"]` (verified by `go test ./internal/config/...` output quoting the set).

### AC-002 — Tier-slot defaults switched (REQ-002)
Given the Go defaults after M1, when the default config is loaded, then `DefaultGLMHigh`, `DefaultGLMMedium`, `DefaultGLMLow`, `DefaultGLMFable`, `DefaultGLMHaiku`, `DefaultGLMSonnet`, and `DefaultGLMOpus` all equal `"glm-5.3-flash"` (test output or verbatim constant block cited).

### AC-003 — Template twin switched (REQ-002)
Given the template after M1, when `internal/template/templates/.moai/config/sections/llm.yaml` is read, then all four `llm.glm.models.*` values are `"glm-5.3-flash"` and `make build` exits 0.

### AC-004 — glm-5.3 preserved (REQ-003)
Given an llm.yaml that names `"glm-5.3"` in a tier slot, when the config loads and the session launches, then the slot resolves `glm-5.3` with unchanged behavior (collapse low→low; 1M context) — evidenced by an existing or added config-load test plus the glm-5.3 regression row in AC-006.

### AC-005 — Flash overlay: low collapses to max (REQ-004)
Given the overlay after M2, when the resolved model is `glm-5.3-flash` and Claude effort is `low`, then the resolved reasoning state is `{thinking enabled, reasoning_effort: "max"}` — not `low`.

### AC-006 — Non-flash collapse unchanged (REQ-004, mirror-image regression)
Given the overlay after M2, when the resolved model is `glm-5.3` (or any non-flash model), then Claude effort `low` still resolves to `reasoning_effort: low` and efforts above low still resolve to `max`, matching pre-SPEC behavior.

### AC-007 — Overlay totality retained (REQ-004)
Given the overlay after M2, when an unrecognized effort string is collapsed for any model including flash, then the result is the `max` state (no panic, no under-reasoning).

### AC-008 — Context window direct entry (REQ-006)
Given `glmContextWindows` after M3, when `matchContextWindow` resolves `"glm-5.3-flash"`, then it returns 1000000 via the direct table entry; and `"glm-5.3"` still returns 1000000 (table test output cited).

### AC-009 — Web tier widget offers flash (REQ-007)
Given the settings schema after M1/M4, when the tier-slot select fields are rendered, then their option values include `glm-5.3-flash` and the field type is select (updated `glm_tier_test.go` passes).

### AC-010 — Web labels in four locales (REQ-007)
Given `internal/web/assets/i18n.js` after M4, when the i18n keys `f.llm.glm.models.opt.glm-5.3-flash` and `f.workflow.audit.glm.model.opt.glm-5.3-flash` are looked up, then non-empty localized labels exist in each of the en, ko, ja, and zh locale blocks (grep count = 8).

### AC-011 — Docs updated (REQ-009)
Given the docs surfaces listed in plan.md §D, when the GLM default model is referenced in README (4 locales) and the scoped docs-site pages, then the text names `glm-5.3-flash` as the default and glm-5.3 as selectable; 4-locale parity holds.

### AC-012 — Boot smoke (REQ-010)
Given the built binary after M6, when `moai glm` is launched with default config, then the injected `ANTHROPIC_DEFAULT_OPUS_MODEL` (and sibling slot vars) equal `glm-5.3-flash` and the statusline context size resolves 1000000 — observed at the env-injection level via the `buildTmuxInjectVars` / `setGLMEnv` map (the M6-pinned Go-test assertion; no live z.ai API dependency — env-level observation only), env values quoted verbatim.

### AC-013 — Overlay documentation twin (REQ-005)
Given the template llm.yaml after M2, when the effort-mapping comment block (`internal/template/templates/.moai/config/sections/llm.yaml`) is read, then it states the flash max-only rule — a grep-able statement naming `glm-5.3-flash` and that every Claude effort level (including `low`) resolves to `reasoning_effort: max` — alongside the existing collapse table (`grep -n 'glm-5.3-flash' internal/template/templates/.moai/config/sections/llm.yaml` returns ≥1 hit inside the comment block).

## §D.1 Severity

- **Must-pass**: AC-001, AC-002, AC-003, AC-005, AC-006, AC-007, AC-009, AC-012 (default switch + flash wire correctness + regression guards).
- **Should-pass**: AC-004, AC-008, AC-010, AC-011, AC-013.

## §D.2 Traceability

| AC | REQ | Milestone |
|----|-----|-----------|
| AC-001 | REQ-001 | M1 |
| AC-002 | REQ-002 | M1 |
| AC-003 | REQ-002, REQ-008 | M1/M6 |
| AC-004 | REQ-003 | M1/M2 |
| AC-005 | REQ-004 | M2 |
| AC-006 | REQ-004 | M2 |
| AC-007 | REQ-004 | M2 |
| AC-008 | REQ-006 | M3 |
| AC-009 | REQ-007 | M1/M4 |
| AC-010 | REQ-007 | M4 |
| AC-011 | REQ-009 | M5 |
| AC-012 | REQ-010 | M6 |
| AC-013 | REQ-005 | M2 |

## §D.3 Edge cases

- Unrecognized effort string under flash (AC-007 totality).
- Existing user llm.yaml pinning glm-5.3 or glm-5.2 in a slot (must keep loading; neither id leaves the loadable space).
- Longest-substring table interplay: an unregistered `glm-5.3-*` variant (e.g. a future `glm-5.3-flash-lite`) inherits 1M via longest-substring matching through the `"glm-5.3"` entry — registration-time guidance: any future divergent `glm-5.3-*` id MUST add its own explicit table entry at registration time (otherwise it inherits 1M via substring). The M3 table test pins flash→1M direct and glm-5.3→1M retained.
- i18n key absent in one locale (AC-010 counts all 8 keys).
- tmux absent (boot smoke must still verify the process-env injection path, glm.go:354-357).

## §D.4 Quality gate / Definition of Done

- Affected-package tests green (`internal/config`, `internal/settings`, `internal/template`, `internal/statusline`, `internal/web`), output cited in progress.md §E.2.
- `make build` exit 0; catalog.yaml untouched; template-neutrality clean (no SPEC-IDs/dates/SHAs in template content).
- Full-suite verdict delegated to CI on the PR head.
- All must-pass ACs PASS with verbatim evidence; §E.3 audit-ready signal populated by manager-develop.
