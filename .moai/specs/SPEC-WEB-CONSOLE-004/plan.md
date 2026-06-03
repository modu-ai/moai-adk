# Implementation Plan — SPEC-WEB-CONSOLE-004

> 모두의AI design system application to the `moai web` console (visual restyle, zero server-contract change). Tier **M**. Run-phase `cycle_type=tdd` per `development_mode: tdd`. Plan-phase authored in main checkout (no worktree) on branch `docs/glm-webtool-routing-m1-m5`.

## §A — Context

### A.1 Position & tier

- **Cohort**: `web-console-v3` visual-restyle member (004). Siblings: 001 (모태), 002/S1 (port+parity, completed), 003/S2a (flat project-config, in-progress).
- **Tier M justification**: multi-file cross-asset change — `page.html.tmpl` (template restyle + appbar + field chrome) + an embedded token/component CSS layer + Pretendard woff2 subset font assets + inline-SVG icon subset + `app.js` (theme toggle) + `handlers.go` (one new `BindAddr` view-model field) + `assets.go` (`go:embed` directive extension) + tests. LOC is moderate (mostly CSS/template/asset, modest Go), but the change spans templates + assets + a Go view-model field + go:embed wiring + a new font/icon delivery pipeline, exceeding Tier S's "<5 files / <300 LOC" envelope. The cross-asset coordination (offline font/icon pipeline + server-contract regression surface) warrants the 3-artifact Tier M set (spec.md + plan.md + acceptance.md). Not Tier L: no constitutional change, no new persistence model, no nested-config redesign, < ~15 files.

### A.2 SPEC artifacts

- `.moai/specs/SPEC-WEB-CONSOLE-004/spec.md` — 12 GEARS REQs (REQ-WC4-001..012) + 13 AC index + §4 Exclusions (E.1–E.5).
- `.moai/specs/SPEC-WEB-CONSOLE-004/plan.md` — this file.
- `.moai/specs/SPEC-WEB-CONSOLE-004/acceptance.md` — full Given-When-Then AC enumeration.
- `.moai/specs/SPEC-WEB-CONSOLE-004/progress.md` — plan-complete signal + run-phase evidence (created at plan close).

### A.3 PRIMARY SOURCE

`.moai/design/web-console-handoff/IMPLEMENTATION-HANDOFF.md` is the research + design basis. The run-phase implementer reads it FULLY before coding — it carries the design→Go mapping (§3), the 11-item server-contract list (§4), the canonical discrepancy table (§5.1), the cohort decomposition (§6), the font (§7) / icon (§8) offline strategy with the recommended options (woff2 subset / inline SVG), the M1–M5 step plan (§9), and the acceptance basis (§10). The design CSS/markup/screenshots under `from-claude-design/` are the visual SSOT.

### A.4 Existing infrastructure — PRESERVE vs EXTEND

| Item | Action | Detail |
|------|--------|--------|
| `internal/web/server.go` `loopbackHost`/`listenerAddr()` | PRESERVE + READ | bind posture unchanged; `listenerAddr()` (or equivalent) is the source for the `BindAddr` view-model field |
| `internal/web/app.go` `hostCheckMiddleware` + `routes()` | PRESERVE | no-auth loopback posture + `/static/` serving unchanged |
| `internal/web/validate.go` (canonical lists + validators) | PRESERVE (byte-unchanged) | 004 is visual-only; touches NO validator |
| `internal/web/handlers.go` `newPageView()` | EXTEND (minimal) | add ONE `BindAddr` field to `pageView` + populate it in `newPageView()`; everything else unchanged |
| `internal/web/assets/page.html.tmpl` | RESTYLE | appbar + pagehead + profilebar + banner + 5-fieldset + actions chrome; preserve `name=`/`{{range}}`/`.FieldErrors`/helpers |
| `internal/web/assets/style.css` | REPLACE/EXTEND | embedded 모두의AI token + component layer |
| `internal/web/assets/app.js` | EXTEND | theme-toggle (client `localStorage`); preserve existing segment-visibility logic |
| `internal/web/assets.go` `go:embed` directive | EXTEND | enumerate new font / CSS / SVG / JS assets |
| `internal/web/assets/fonts/` (NEW) | CREATE | Pretendard woff2 subset + OFL-1.1 license |

### A.5 PRESERVE list (do NOT modify outside SPEC scope)

- `internal/web/validate.go` — canonical lists + `validatePrefs` + `validateProjectConfig` (byte-unchanged).
- `internal/web/server.go` bind/host logic; `internal/web/app.go` middleware/auth posture.
- The S1/S2a form field set (no field added/removed/re-scoped).
- All parallel-session in-flight files; the 3 parallel-session commits (divergence `0 3`); the modified-but-unrelated working-tree files (`.claude/settings.json`, `.moai/config/sections/{statusline,user}.yaml`, the deleted `SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001/progress.md`); `.moai/design/web-console-handoff/` (leave untracked); `.moai/docs/harness-delivery-strategy.md`.

---

## §B — Known Issues (auto-injected for run-phase delegation)

Filtered to the categories relevant to this `internal/web` visual-restyle SPEC:

- **B3 (Subagent Boundary)** — `internal/web` is product code; no `AskUserQuestion` in the web layer (it never had any). Not a concern here, but the implementer MUST NOT add user-interaction surfaces.
- **B4 (Frontmatter schema)** — N/A to run-phase (spec.md frontmatter already canonical: `created`/`updated`/`tags`).
- **B5 (CI 3-tier)** — spec-lint, golangci-lint, Test(per OS) fail independently. The visual restyle should not affect Go lint (mostly template/CSS/asset); the one Go change (`BindAddr` field) must compile clean and cross-platform.
- **B8 (Working tree hygiene)** — `git add` ONLY `internal/web/**` run-phase changes; do NOT stage the parallel-session files, the design handoff dir, or unrelated modified config. Runtime-managed files untouched.
- **B10 (Untouched paths PRESERVE)** — §A.5 PRESERVE list. Parallel `manager-develop`/sibling-SPEC instances may be in flight on other paths; stay inside `internal/web/`.
- **B11 (AskUserQuestion forbidden)** — subagent returns a structured blocker report on any blocker; never prompts the user.
- **B1 (cross-platform build)** — the one Go change (`BindAddr` view-model field) is platform-neutral, but the run-phase self-verification MUST include `GOOS=windows GOARCH=amd64 go build ./...` (REQ-WC4-012 / AC-WC4-012). The font/icon assets are static bytes — platform-neutral by construction.

### B-extra: restyle-specific risks

- **R1 — Offline regression (highest)**: the design's two CDN dependencies (Google-Fonts `@import` in `colors_and_type.css:26` + lucide unpkg `<script>`) MUST both be removed. AC-WC4-001 / AC-WC4-007 grep the served CSS/HTML for `fonts.googleapis.com` / `unpkg.com` / any external font/style/script `https://` URL = 0 matches. A single leaked CDN reference fails the zero-network invariant.
- **R2 — Non-canonical option leak**: porting the design's hardcoded `<option>` lists would introduce `es/fr/de`, `haiku[1m]`, or kebab segment keys → server rejects on POST. AC-WC4-008 greps the rendered/template form for these forbidden tokens = 0; option lists must stay `{{range}}`-driven.
- **R3 — `name=` loss**: the design markup is `id=`-only. Dropping a `name=` while restyling breaks `r.PostFormValue`. AC-WC4-009a greps every canonical `name="…"` survives.
- **R4 — Banner class drift**: server sets `.BannerKind` = `"ok"`/`"error"`; the design uses `banner--success`/`banner--error`. Map in the template; do NOT change the server-set kind values (REQ-WC4-010).
- **R5 — Font subset pipeline**: woff2 subsetting (`pyftsubset`/`fonttools`) needs Latin+Hangul + the used weights only. The OFL-1.1 license file MUST ship alongside. Keep the embedded font footprint small (subset, not the 9MB OTF set).
- **R6 — go:embed glob**: extending the `assets.go` `go:embed` directive to cover a `fonts/` subdir + woff2 + SVG must keep the build green on all platforms (`embed` uses forward-slash paths — already cross-platform-safe).

---

## §C — Pre-flight Checklist (run-phase, before any code change)

```bash
# 1. Branch + baseline
git branch --show-current          # docs/glm-webtool-routing-m1-m5 (or run-phase feat branch)
git rev-parse HEAD

# 2. Cross-platform build baseline (must already be green before restyle)
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. Web package test baseline (NEW vs pre-existing distinction)
go test ./internal/web/... 2>&1 | tail -10

# 4. Lint baseline
golangci-lint run ./internal/web/... --timeout=2m 2>&1 | tail -5

# 5. Read the primary source FULLY + the design assets
#    .moai/design/web-console-handoff/IMPLEMENTATION-HANDOFF.md (all 10 sections)
#    from-claude-design/assets/{colors_and_type.css, console.css}
#    from-claude-design/MoAI-Web-Console.html

# 6. Confirm the validators are the byte-unchanged baseline
git diff --stat internal/web/validate.go    # expect NO change at run-phase end
```

---

## §D — Constraints (DO NOT VIOLATE)

- **PRESERVE** the §A.5 list verbatim. `internal/web/validate.go` byte-unchanged.
- **Server contract** (REQ-WC4-009): `name=` on every input/select; `{{range}}` server-rendered options; `.FieldErrors` server-side render; `form method/action` + hidden `__profile`; `langSelect`/`optSelect` helper structure; loopback bind; no-auth + Host-check.
- **Zero network**: no Google-Fonts `@import`, no lucide CDN `<script>`, no external font/style/script `https://` URL in served assets. Fonts = self-host woff2 subset via `go:embed`; icons = inline SVG.
- **No non-canonical options**: no `es/fr/de`, no `haiku[1m]`, no kebab segment keys; option lists stay `{{range}}`-driven from the canonical view-model.
- **No field-set change**: visual/layout ONLY — no settings field added/removed/re-scoped; no `.review` aside; no interface-i18n / language picker / `data-i18n` / `i18n.js`.
- **No new persistence**: 004 writes no config; no `yaml.Marshal`/`os.WriteFile` in `internal/web`; theme persistence client-side `localStorage` only.
- **Licensing**: OFL-1.1 (Pretendard) file shipped with the font subset; lucide ISC acknowledged with the icon subset.
- **Git**: Conventional Commits (`feat(SPEC-WEB-CONSOLE-004): M{N} …`); `Authored-By-Agent: manager-develop` + `🗿 MoAI` trailer; `git add` specific `internal/web/**` paths only; NO `--no-verify`, NO `--amend`, NO force-push to main.
- **No template mirroring / make build** — web assets are package-local `go:embed`, not `internal/template/templates/` deployed assets.

---

## §E — Self-Verification (run-phase deliverables)

The run-phase implementer reports the AC PASS/FAIL matrix (acceptance.md SSOT) + the following verification batch (parallel single-turn):

```bash
# Functional / build
go test ./internal/web/... ./internal/cli/... ./internal/config/...   # AC-WC4-013 closure gate
go build ./...                                                        # AC-WC4-012
GOOS=windows GOARCH=amd64 go build ./...                              # AC-WC4-012

# Offline / zero-network regression (AC-WC4-001 / AC-WC4-007) — expect 0 matches
grep -rn 'fonts.googleapis.com\|unpkg.com\|cdnjs\|jsdelivr\|https://fonts\|@import url("http' internal/web/assets/

# Non-canonical option regression (AC-WC4-008) — expect 0 matches
grep -rn 'value="es"\|value="fr"\|value="de"\|haiku\[1m\]\|segment_git-branch\|segment_session-cost' internal/web/assets/page.html.tmpl

# name= survival (AC-WC4-009a) — expect each canonical name present
grep -c 'name="user_name"\|name="conversation_lang"\|name="permission_mode"\|name="model"\|name="__profile"\|name="segment_' internal/web/assets/page.html.tmpl

# Validators byte-unchanged (AC-WC4-009b / E.5.1)
git diff --exit-code internal/web/validate.go && echo "validate.go unchanged"

# go:embed enumerates new assets (AC-WC4-012)
grep -n 'go:embed' internal/web/assets.go

# Font + license present (AC-WC4-002)
ls internal/web/assets/fonts/   # woff2 subset + OFL/LICENSE
```

E-deliverables: AC binary matrix, cross-platform build result, web/cli/config test result, the offline + non-canonical + name= regression grep outputs, `validate.go` unchanged confirmation, commit SHAs + push status, blocker report (if any).

---

## §F — Milestones (priority-ordered; no time estimates)

Per the handoff §9 step plan. Each milestone names its preservation verification. `cycle_type=tdd` — for the testable surface (template render assertions, server-contract regression greps, the `BindAddr` view-model field), write the failing test first, then implement.

### M1 — Token + font layer (offline-safe foundation)

- Port the 모두의AI token layer (`colors_and_type.css` `:root` + `[data-theme="dark"]` + base typography) into an embedded console CSS sheet under `internal/web/assets/`. **Remove** the Google-Fonts `@import`.
- Add the Pretendard woff2 **subset** (Latin+Hangul, used weights only) under `internal/web/assets/fonts/` + the **OFL-1.1 license** file. Reference via relative `@font-face src` (no `https://`). Inter/JetBrains-Mono `@import` removed (system/mono fallback for the base look).
- Extend the `internal/web/assets.go` `go:embed` directive to enumerate the new CSS + font assets.
- **Preservation check**: `go build ./...` + `GOOS=windows … build` green; served CSS has 0 external font/style URL (AC-WC4-001/002); `go:embed` covers the new assets (AC-WC4-012).
- **REQ coverage**: REQ-WC4-001, REQ-WC4-002, (partial) REQ-WC4-012.

### M2 — Layout + component port (server-contract preservation — the critical gate)

- Restyle `page.html.tmpl`: appbar (M4 icons stubbed) + pagehead + profilebar + banner + the 5 fieldsets + actions, applying the `console.css` component chrome (section cards, field chrome with title + `<code>` key chip + description, styled selects, segment cards, primary button).
- Wire field chrome through the `langSelect`/`optSelect` define blocks (extend their signature if a title/key/desc param is needed — structure preserved, not replaced).
- Add the `BindAddr` view-model field to `pageView` + populate it in `newPageView()` (from `Server.listenerAddr()` / bind knowledge) for the loopback indicator (REQ-WC4-005).
- Map `.BannerKind` (`"ok"`/`"error"`) → `banner--success`/`banner--error` chrome in the template (server-set kind unchanged) (REQ-WC4-010).
- **PRESERVE** every `name=` attr, every `{{range}}` server-rendered option, `.FieldErrors` server-side render, `form method/action` + hidden `__profile` + `{{if .ShowProfileSwitch}}`. **DROP** the design's non-canonical options (don't copy hardcoded `<option>`s).
- **Preservation check (gate)**: `go test ./internal/web/...` green; `validatePrefs`/`validateProjectConfig` byte-unchanged; POST round-trip + per-field error render intact; name= regression grep passes; non-canonical option grep = 0 (AC-WC4-003/005/008/009a/009b/010).
- **REQ coverage**: REQ-WC4-003, REQ-WC4-004 (chrome, icons stubbed), REQ-WC4-005, REQ-WC4-008, REQ-WC4-009, REQ-WC4-010.

### M3 — Dark mode + theme toggle

- Wire the appbar theme-toggle button (sun/moon) + client-side `localStorage` persistence in `app.js`; add the FOUC-prevention inline `<head>` snippet applying the persisted theme before first paint. Preserve the `prefers-reduced-motion` guard. No server round-trip / no theme persistence field.
- **Preservation check**: light/dark toggle works; `[data-theme]` override block present; no server-side theme write; existing segment-visibility JS still works (AC-WC4-006).
- **REQ coverage**: REQ-WC4-006.

### M4 — Inline-SVG icon subset

- Replace the lucide CDN `<script>` with ~14 **inline SVG** icons (or a small embedded SVG sprite) in the template/appbar: user-round / languages / rocket / panel-bottom / folder-git-2 / chevron-down / sun / moon / save / check / alert-circle / x + banner-state icons. Acknowledge the lucide ISC license.
- **Preservation check**: icons render offline (inline `<svg>` present); 0 `unpkg.com`/lucide CDN `<script>` in served HTML; no runtime icon-library JS (AC-WC4-007).
- **REQ coverage**: REQ-WC4-007.

### M5 — Accessibility + closure verification

- Verify/strengthen non-color error cues (icon+border+text), `focus-visible` outlines, `prefers-reduced-motion` guard, ARIA attrs (`aria-label` on theme toggle + icon-only controls; `aria-invalid`/`aria-describedby` on errored fields where already present).
- Run the full closure gate `go test ./internal/web/... ./internal/cli/... ./internal/config/...` + the §E regression batch.
- **REQ coverage**: REQ-WC4-011, REQ-WC4-012 (final), closure (AC-WC4-013).

> **i18n (handoff §9 M5/S3) is NOT a 004 milestone.** Web interface i18n + appbar language picker + `i18n.js` + full CJK webfont are cohort S3 (§4 E.1/E.2). 004 stops at the base brand look with Pretendard Latin+Hangul.

### Milestone ordering rationale

M1 (token/font) is the visual-parity prerequisite. M2 is the server-contract preservation gate — if the contract breaks here, the form stops working, so it is the highest-risk milestone and gates M3/M4/M5. M3/M4/M5 layer onto M2 incrementally. Accessibility (M5) verifies the end state.

---

## §G — Anti-Patterns (do NOT do)

- Copying the design's hardcoded `<option>` lists (introduces `es/fr/de` / `haiku[1m]` / kebab segment keys → server reject). Keep `{{range}}` server-render.
- Leaving any Google-Fonts `@import` or lucide CDN `<script>` (breaks offline). Self-host + inline SVG.
- Adopting client-side validation as the source of truth (server `.FieldErrors` is SSOT; `name=` mandatory).
- Hardcoding `127.0.0.1:3041` in the loopback indicator (must be the real bound address via `BindAddr`).
- Changing `.BannerKind` server values to `"success"`/`"error"` (map in template; server contract stays `"ok"`/`"error"`).
- Touching `validate.go` / the field set / adding settings fields (004 is visual-only).
- Porting the `.review` "State preview" aside (design-tool scaffold, not product).
- Adding i18n / language picker / `data-i18n` / `i18n.js` / Noto-CJK (S3 scope).
- Shipping the 9MB OTF font set instead of a woff2 subset (binary bloat — subset Latin+Hangul + used weights).
- Mirroring assets into `internal/template/templates/` or running `make build` (web assets are package-local `go:embed`).

---

## §H — Cross-References

- `.moai/design/web-console-handoff/IMPLEMENTATION-HANDOFF.md` — PRIMARY design + research basis.
- `.moai/specs/SPEC-WEB-CONSOLE-001/` — original loopback console (invariant source).
- `.moai/specs/SPEC-WEB-CONSOLE-002/` — port 3041 + web↔TUI parity (S1 invariant).
- `.moai/specs/SPEC-WEB-CONSOLE-003/` — flat project-config parity (S2a; same `internal/web` module, sibling scope fence).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field frontmatter schema (canonical).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Tier M Section A-E delegation template.
