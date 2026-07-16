# progress.md — SPEC-DESIGN-MOAIWEBV2-001

> Lifecycle progress tracking. §E.* headings are parser-load-bearing (era.go classifies via literal `§E.2`/`§E.3`/`§E.4` tokens). Plan-phase populates §E.1 only; §E.2/§E.3 are owned by manager-develop (run-phase), §E.4 by manager-docs (sync-phase).

## §E.1 Plan-phase Audit-Ready Signal

- `plan_status: audit-ready`
- `plan_complete_at: 2026-07-16`
- Tier: M
- Artifacts: spec.md, plan.md, acceptance.md, research.md, design.md, progress.md (full 5-artifact set + progress, exceeding the Tier M 3-file minimum per mission mandate).
- SPEC ID self-check: `SPEC-DESIGN-MOAIWEBV2-001` → decomposition: SPEC ✓ | DESIGN ✓ | MOAIWEBV2 ✓ | 001 ✓ → PASS (regex `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`).
- Clarifications (4) — ALL resolved via AskUserQuestion 2026-07-16 (recorded in plan.md §B / research.md §G):
  1. project-panel disposition → **REMOVE** (settings stay editable via yaml/CLI)
  2. signature accent → **v2 solid `#3d7d5f`**
  3. header brand pose → **`mascot-thinking.png`**
  4. mascot placement → **library + header only**
- plan-auditor: iter-1 FAIL (0.83; sole blocker MP-7 clarification gate; Traceability 0.55). Defect-delta fixes applied 2026-07-16: all 4 markers resolved (grep-0); D2 → AC-MWV2-005 (REQ-MWV2-004 superset/subset); D3 → AC-MWV2-032 (REQ-MWV2-032 light-only); D4 → REQ-MWV2-041 label Event-driven; D5 → AC-MWV2-002 teal grep anchored to `:root`. Awaiting iter-2 re-audit.

## §E.2 Run-phase Evidence

Development mode: DDD (cycle_type=ddd) — brownfield restyle of a tested package; the
existing rendered-HTML boundary tests are the preservation net. Milestones M1-M4
dependency-ordered (removal → mascots → tokens → build/verify).

### Milestone log

- **M1 — Orphan `project` panel removal (COMPLETE).** Removed `fieldsetProject` +
  `detectionField` from `fieldsets.templ`; removed the `data-panel="project"` block
  from `root.templ`. Server-side parse/validate/persist seam (`parseProjectNestedForm`
  in `projectconfig.go`, wired in `handlers.go`) + the `pageView.Cur*` fields are
  PRESERVED (REQ-MWV2-031) — `development_mode` / `git_convention` / `quality.*` stay
  editable via yaml config / CLI, now inert in the view-model (not deletable-safe: the
  server contract + its tests still reference them). Blast radius: the reject-echo +
  DOM-field-presence render assertions across `schemaform_test.go`,
  `projectnested_parse_test.go`, `projectnested_test.go`, `projectnested_error_test.go`,
  `projectconfig_handler_test.go`, `m5b_verify_test.go`, `restyle_test.go`,
  `i18n_test.go`, `schema_render_test.go` were updated to the retirement (server 400 +
  atomic no-write assertions preserved; pure-render project tests removed).
  `go test ./internal/web/...` green.

### §E.2 Run-phase AC matrix (finalized at M4)

_<AC PASS/FAIL matrix populated at M4 verification>_

## §E.3 Run-phase Audit-Ready Signal

_<pending — finalized at M4 verification>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_

- plan_complete_at: 2026-07-16T14:04:42Z
- plan_status: audit-ready

## §F Phase 4 Mode Selection

- Inputs: tier=M, scope ~6-12 files (single package internal/web), domains=1 (Go/Templ/CSS/assets), coding-heavy, concurrency benefit LOW
- Mode evaluation: trivial ✗ (multi-file semantic) / background ✗ (write work) / agent-team ✗ (RETIRED) / parallel ✗ (coding-heavy, single domain) / workflow ✗ (<30 files, non-mechanical) / sub-agent ✓
- Decision: sub-agent
- Justification: coding-heavy single-package work fits the sequential Mode 5 default per Anthropic's coding-task parallelism caveat; milestones M1-M4 are dependency-ordered (removal → assets → tokens → build), so parallel fan-out offers no benefit.
