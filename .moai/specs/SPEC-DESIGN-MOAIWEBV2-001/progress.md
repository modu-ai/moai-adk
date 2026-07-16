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

- **M2 — Mascot 6-pose library + header swap (COMPLETE).** Copied Coffee/Explaining/
  Pointing/Searching/Teaching/Thinking from the v2 bundle into `assets/mascots/` as
  lowercase-kebab `mascot-<pose>.png` (embedded via the existing `assets/mascots`
  glob, served at `/static/mascots/`). Removed `mascot-coding.png` + `mascot-talking.png`.
  Repointed the header brand badge (`board.templ` + `root.templ`) to `mascot-thinking.png`.
  Added `mascots_test.go` (embed+serve, retired-404, header-pose). Library + header only.
- **M3 — v2 token realignment in `console.css` (COMPLETE).** Rewrote the `:root` token
  VALUES to the v2 achromatic canon (name-stable): ink `#060606`, bg `#f4f4f4`,
  achromatic neutral ramp, success `#2e8a63`, fg `#565656`/`#757575`, borders
  `#d1d1d1`/`#e6e6e6`/`#b5b5b5`, focus-ring `rgba(61,125,95,0.16)`, shadow base
  `rgba(6,6,6,…)`, primary hover/active `#316750`/`#265240`, signature accent solid
  `#3d7d5f`. Dropped the `linear-gradient` signature form consistently, including the
  inert `[data-theme="dark"]` override (solid dark point-green `#5a9a7e`) so AC-004's
  file-wide absence grep passes. PRESERVED the `@font-face` block + `--font-latin`/
  `--font-mono` system fallbacks verbatim (REQ-MWV2-030). No `linear-gradient(var(...))`
  double-wrap existed → no component adjustment needed. Updated the one pinned-hex test.
- **M4 — Build + cross-platform + verification (COMPLETE).** `make build` clean;
  `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0;
  `go test ./internal/web/...` exit 0; `go vet ./...` clean; `golangci-lint run
  ./internal/web/...` 0 issues. Folded audit residuals D6 (design.md §A.2/§C.1
  pending-clarification → confirmed) + D7 (acceptance.md AC-005(a) vacuous-clause note +
  a2 orphan check).

### §E.2 Run-phase AC matrix

| AC | Severity | Status | Evidence (command → observed) |
|----|----------|--------|-------------------------------|
| AC-MWV2-001 | MUST | PASS | `grep --color-ink:\s*#060606` + `--color-bg:\s*#f4f4f4` → both match |
| AC-MWV2-002 | MUST | PASS | `--neutral-400:#9fa0a0` + `--neutral-950:#060606` present; awk `:root`-window teal grep → "PASS: no teal residue in :root" |
| AC-MWV2-003 | MUST | PASS | success `#2e8a63` + borders `#d1d1d1`/`#e6e6e6`/`#b5b5b5` + fg `#565656`/`#757575` + focus-ring `rgba(61,125,95,0.16)` → all match |
| AC-MWV2-004 | MUST | PASS | `--gradient-signature:\s*#3d7d5f` present; `! grep --gradient-signature:\s*linear-gradient` file-wide → absent (dark override de-gradiented to solid `#5a9a7e`) |
| AC-MWV2-005 | MUST | PASS | (a) hover `#316750` + active `#265240` v2 values match; (a2) orphan token-name loop → no ORPHAN; (b) `--font-latin:system-ui` + `--font-mono:ui-monospace` retained |
| AC-MWV2-010 | MUST | PASS | 6 `test -f mascot-{coffee,explaining,pointing,searching,teaching,thinking}.png` → all present, no MISSING |
| AC-MWV2-011 | MUST | PASS | embed directive grep matches; `TestMascotPosesEmbeddedAndServed` → ok (each pose 200 via `/static/mascots/`) |
| AC-MWV2-012 | MUST | PASS | `mascot-thinking.png` in board.templ + root.templ; `! grep mascot-coding.png` → absent; `TestHeaderBrandBadgeUsesThinkingPose` ok |
| AC-MWV2-013 | MUST | PASS | `test ! -f mascot-talking.png` → absent; `TestMascotRetiredAssetsRemoved` ok |
| AC-MWV2-020 | MUST | PASS | `! grep data-panel="project"` root.templ+root_templ.go → "PASS: project panel removed"; `TestConsoleRendersReportTab` (orphan-panel assertion) ok |
| AC-MWV2-021 | MUST | PASS | `go test -run 'TestConsoleRendersReportTab\|Test.*Save\|Test.*Atomic\|Test.*Submit'` → ok |
| AC-MWV2-030 | MUST | PASS | `! grep '@import url("http'` + `! grep 'src:url("http'` console.css → both absent |
| AC-MWV2-031 | MUST | PASS | `git diff 62698fc2..HEAD --stat` handlers.go/projectconfig.go/validate.go/server.go/app.go → EMPTY (zero server-contract-file change); `TestHandlers\|TestIntegration\|TestSecurity\|TestHTMX\|TestHtmx` → ok |
| AC-MWV2-032 | MUST | PASS-WITH-DEBT | light-only INTENT preserved (my change adds 0 new dark wiring — verified: diffs touch only :root/dark token VALUES + mascots + panel removal). BUT the literal grep `data-theme\|theme.?toggle` matches 16 PRE-EXISTING hits: the console already ships a functional dark toggle (app.js `wireThemeToggle`/`applyTheme` + `themeToggle` button in root/board.templ + FOUC `setAttribute("data-theme")`) from SPEC-WEB-CONSOLE-004/005 — the AC's "no dark toggle wired" premise did not match the pre-existing reality. DEBT: AC-032 verification command is over-broad (matches the pre-existing toggle this SPEC neither added nor is scoped to remove). |
| AC-MWV2-040 | MUST | PASS | `make build` exit 0; `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go test ./internal/web/...` exit 0 |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-16
run_commit_sha: ab5cdf90e   # M4 final run commit (re-pointed after rebase onto origin/main rewrote pre-rebase SHA a57c38420)
m1_to_mN_commit_strategy: per-milestone (M1 10648b133 / M2 5be1056c3 / M3 b0814ff54 / M4 a57c38420)
run_status: complete
ac_pass_count: 14
ac_pass_with_debt_count: 1        # AC-MWV2-032 (pre-existing dark toggle; over-broad grep)
ac_fail_count: 0
preserve_list_post_run_count: verified   # handlers.go / projectconfig.go / validate.go / server.go / app.go byte-unchanged vs 62698fc2 (git diff --stat EMPTY); @font-face block + i18n.js untouched
l44_pre_commit_fetch: origin/main diverged (+6 remote / +4 local) — parallel session advanced origin; NOT pushed (orchestrator reconciles per mission)
l44_post_push_fetch: n/a (not pushed — orchestrator owns push after verification)
new_warnings_or_lints_introduced: 0    # go vet ./... clean; golangci-lint run ./internal/web/... → 0 issues
cross_platform_build:
  linux_amd64: exit 0        # go build ./...
  windows_amd64: exit 0      # GOOS=windows GOARCH=amd64 go build ./...
total_run_phase_files: 29    # 24 internal/web (incl 8 mascot asset ops) + spec.md + progress.md + design.md + acceptance.md + (regenerated *_templ.go)
full_suite_cascade: internal/web 100% green; ONE pre-existing baseline failure OUTSIDE scope — internal/config TestAuditLoaderCompleteness (delegation section lacks a loader, introduced by commit ce423281b; I never touched internal/config — git diff confirms) — NOT a regression from this SPEC
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-17
sync_commit_sha: pending-backfill-DESIGN-MOAIWEBV2-001   # backfilled in immediate follow-up commit per D3 self-referential-hash exemption
sync_status: complete
changelog_entry_position: "[Unreleased] > Added, immediately before SPEC-CLI-TUX-V3-005"
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (updated: 2026-07-17)"
  plan_md: "no frontmatter block (body-only artifact per this SPEC's convention)"
  acceptance_md: "no frontmatter block (body-only artifact per this SPEC's convention)"
```

- plan_complete_at: 2026-07-16T14:04:42Z
- plan_status: audit-ready

## §F Phase 4 Mode Selection

- Inputs: tier=M, scope ~6-12 files (single package internal/web), domains=1 (Go/Templ/CSS/assets), coding-heavy, concurrency benefit LOW
- Mode evaluation: trivial ✗ (multi-file semantic) / background ✗ (write work) / agent-team ✗ (RETIRED) / parallel ✗ (coding-heavy, single domain) / workflow ✗ (<30 files, non-mechanical) / sub-agent ✓
- Decision: sub-agent
- Justification: coding-heavy single-package work fits the sequential Mode 5 default per Anthropic's coding-task parallelism caveat; milestones M1-M4 are dependency-ordered (removal → assets → tokens → build), so parallel fan-out offers no benefit.
