# Progress — SPEC-WEB-CONSOLE-004

## Plan Phase

- **Tier**: M (3 artifacts: spec.md + plan.md + acceptance.md)
- **cycle_type (run-phase)**: tdd (per `development_mode: tdd`)
- **status**: draft
- **Cohort**: web-console-v3 visual-restyle member (004); siblings 001 (모태) / 002 S1 (completed) / 003 S2a (in-progress)
- **Primary source**: `.moai/design/web-console-handoff/IMPLEMENTATION-HANDOFF.md` (formalized into SPEC)
- **REQ count**: 12 (REQ-WC4-001 .. REQ-WC4-012)
- **AC count**: 13 (AC-WC4-001 .. AC-WC4-013)
- **MUST-PASS invariant**: zero server-contract change (spec.md §3, REQ-WC4-009)

### Plan-phase artifacts created

- spec.md — 12 GEARS REQs + 13 AC index + §4 Exclusions (E.1 web-i18n/langpick→S3, E.2 CJK font→S3, E.3 nested config→S2b, E.4 .review aside, E.5 anti-patterns)
- plan.md — Tier M justification + §A Context + §B Known Issues + §C Pre-flight + §D Constraints + §E Self-Verification + §F Milestones (M1 token/font → M2 layout/components [server-contract gate] → M3 dark-mode → M4 inline-SVG icons → M5 a11y/closure) + §G Anti-Patterns + §H Cross-References
- acceptance.md — full Given-When-Then for all 13 AC + §D edge cases (EC-1..EC-10) + §E Definition of Done + §F forward-looking S3/S2b checks

plan_complete_at: 2026-06-03
plan_status: audit-ready

## Run Phase

- status transition: draft -> in-progress (M1 commit)
- cycle_type: tdd (RED -> GREEN per milestone)
- runtime materialized an L1 worktree (`.claude/worktrees/agent-*`); milestones committed in the worktree, orchestrator cherry-picks to the feature branch.

### Milestones implemented (M1 -> M5)

- **M1 — Token + font layer (offline-safe foundation)**: ported the 모두의AI token layer (colors_and_type.css `:root` + `[data-theme="dark"]` + base typography) into a new embedded `internal/web/assets/console.css`; removed the Google-Fonts `@import`; replaced Inter/JetBrains-Mono with system/mono fallbacks. Generated a self-hosted Pretendard **woff2 subset** (Latin + brand Hangul 모두의AI, 5 used weights 400/500/600/700/900, ~18KB each ~ 100KB total) from the official orioncactus/pretendard v1.3.9 OFL-1.1 release via pyftsubset; shipped under `internal/web/assets/fonts/` with `OFL.txt`. Extended the `assets.go` go:embed directive to `assets/console.css assets/app.js assets/page.html.tmpl assets/fonts`. Removed the old `style.css`.
- **M2 — Layout + component port (server-contract gate)**: restyled `page.html.tmpl` (appbar + pagehead + profilebar + banner + 5 fieldsets + actions) with the console.css chrome; extended the langSelect/optSelect define blocks with title/key/desc field-chrome params (structure preserved). Added a `BindAddr` view-model field to `pageView` + an `app.bindAddr` accessor wired to `Server.displayBindAddr()` in `NewServer` (REQ-WC4-005). Mapped `.BannerKind` ok/error -> banner--success/banner--error template-locally (server kind values unchanged). Dropped all non-canonical design options; option lists remain `{{range}}`-driven.
- **M3 — Dark mode + theme toggle**: appbar sun/moon `#themeToggle` + client `localStorage("moai-console-theme")` persistence in app.js; FOUC-prevention inline `<head>` snippet applying the persisted theme before first paint; prefers-reduced-motion guard preserved. No server round-trip / no theme persistence field.
- **M4 — Inline-SVG icon subset**: ~12 inline-SVG icons via a template `icon` helper. No lucide CDN `<script>`, no data-lucide runtime markup, no icon-library runtime JS. lucide ISC license acknowledged in the template comment.
- **M5 — Accessibility + closure**: non-color error cues (icon + border + text), focus-visible outlines, prefers-reduced-motion guard, aria-label on the theme toggle, aria-invalid + aria-describedby on errored fields. Full closure gate green.

### §E.2 — Run-phase Evidence (AC PASS/FAIL matrix)

| AC | REQ | Status | Verification | Actual output |
|----|-----|--------|--------------|---------------|
| AC-WC4-001 | REQ-WC4-001 | PASS | `TestConsoleCSSEmbedded` + offline grep | brand tokens present; offline grep -> 0 matches |
| AC-WC4-002 | REQ-WC4-002 | PASS | `TestPretendardFontSubsetEmbedded` + `TestFontServedFromStatic` | 5 woff2 subsets + OFL.txt embedded; @font-face src relative `/static/fonts/...`; served 200 |
| AC-WC4-003 | REQ-WC4-003 | PASS | `TestComponentChromePresent` | section/field-title/key-chip/select-wrap/seg/btn--primary markers render; langSelect+optSelect define blocks present + used |
| AC-WC4-004 | REQ-WC4-004 | PASS | `TestAppbarRendered` | appbar + brand badge SVG + 모두의AI + loopback + themeToggle; NO langpick/data-i18n |
| AC-WC4-005 | REQ-WC4-005 | PASS | `TestLoopbackIndicatorShowsRealBindAddr` | injected 127.0.0.1:7777 rendered from `{{.BindAddr}}`; template has NO literal 127.0.0.1:3041 |
| AC-WC4-006 | REQ-WC4-006 | PASS | `TestDarkModeAndThemeToggle` | `[data-theme="dark"]` block + themeToggle + FOUC head init + reduced-motion guard; localStorage-only, no server theme field |
| AC-WC4-007 | REQ-WC4-007 | PASS | `TestInlineSVGIconsNoCDN` + grep | inline `<svg>` >= 5; 0 unpkg/lucide@/data-lucide/lucide.min.js |
| AC-WC4-008 | REQ-WC4-008 | PASS | `TestNoNonCanonicalOptions` (D1 structural) | NO es/fr/de, NO haiku[1m], structural `segment_[a-z]+-[a-z]` = 0; 15 canonical snake_case keys render via `{{range .AllSegments}}` |
| AC-WC4-009a | REQ-WC4-009 | PASS | `TestNameAttributesPreserved` + `TestProfileSwitchNameAttrPreserved` | all canonical name= (incl. 15 segment_<key> + __profile/__profile_select) render; method/action + hidden __profile + server-side .FieldErrors present |
| AC-WC4-009b | REQ-WC4-009 | PASS | `TestGoldenPath_ReadWriteRoundTrip` + `git diff validate.go` | 001/002/003 invariant + DO_NOT_TOUCH sentinel tests green; validate.go byte-unchanged; no 0.0.0.0/auth/yaml.Marshal added |
| AC-WC4-010 | REQ-WC4-010 | PASS | `TestBannerKindMapping` | ok->banner--success, error->banner--error; handlers.go still sets BannerKind ok/error |
| AC-WC4-011 | REQ-WC4-011 | PASS | `TestAccessibilityCues` | focus-visible + reduced-motion + aria-label(themeToggle) + aria-invalid/aria-describedby(errored field) + non-color icon+border+text error cue |
| AC-WC4-012 | REQ-WC4-012 | PASS | `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` + go:embed grep | both builds exit 0; go:embed enumerates console.css app.js page.html.tmpl fonts |
| AC-WC4-013 | all | PASS | `go test ./internal/web/... ./internal/cli/... ./internal/config/...` | all packages green |

### §E.3 — Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-03
run_commit_sha: <backfill-after-cherry-pick>
run_status: implemented
cycle_type: tdd
ac_pass_count: 13
ac_fail_count: 0
validate_go_byte_unchanged: true
offline_regression_grep_matches: 0
non_canonical_option_grep_matches: 0
cross_platform_build:
  darwin: exit-0
  windows_amd64: exit-0
web_test_coverage: "90.9%"
lint_status: "0 issues (golangci-lint ./internal/web/...)"
race_detector: pass
new_warnings_or_lints_introduced: false
total_run_phase_files: 15
m1_to_mN_commit_strategy: "single feature-branch commit (M1-M5 holistic, worktree -> cherry-pick)"
```
