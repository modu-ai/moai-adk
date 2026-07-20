# acceptance.md — SPEC-DESIGN-MOAIWEBV2-001

> Machine-verifiable acceptance criteria. Every AC is observable via a grep, a file-existence check, a rendered-HTML boundary test, or a build/test exit code. Commands assume repo root `/Users/goos/MoAI/moai-adk-go`.

## §D. Acceptance Criteria Matrix

| AC | REQ | Severity | Verification (machine-checkable) |
|----|-----|----------|----------------------------------|
| AC-MWV2-001 | REQ-MWV2-003 | MUST | `--color-ink: #060606` AND `--color-bg: #f4f4f4` present in `console.css` `:root` |
| AC-MWV2-002 | REQ-MWV2-002 | MUST | achromatic neutral ramp present; teal literals absent from `:root` |
| AC-MWV2-003 | REQ-MWV2-001 | MUST | shared v2 token values (success/borders/fg/focus-ring/shadow) match canon |
| AC-MWV2-004 | REQ-MWV2-005 | MUST | signature accent = solid `#3d7d5f` (confirmed §C), one consistent treatment |
| AC-MWV2-005 | REQ-MWV2-004 | MUST | token superset/subset handling — v2-only tokens now present in `console.css` AND console-only tokens retained |
| AC-MWV2-010 | REQ-MWV2-010/014 | MUST | 6 lowercase-kebab mascot pose files exist under `internal/web/assets/mascots/` |
| AC-MWV2-011 | REQ-MWV2-011 | MUST | `//go:embed assets/mascots` present; poses served at `/static/mascots/mascot-*.png` |
| AC-MWV2-012 | REQ-MWV2-012 | MUST | header brand badge references `mascot-thinking.png` (confirmed); `mascot-coding.png` absent |
| AC-MWV2-013 | REQ-MWV2-013 | MUST | `mascot-talking.png` absent after cleanup |
| AC-MWV2-020 | REQ-MWV2-020/021 | MUST | rendered console HTML has NO orphan panel — `data-panel="project"` absent (removed) |
| AC-MWV2-021 | REQ-MWV2-022 | MUST | atomic-Save contract preserved (no dependent in-DOM field silently dropped) |
| AC-MWV2-030 | REQ-MWV2-030 | MUST | zero remote font/style fetch in `console.css` (absence grep) |
| AC-MWV2-031 | REQ-MWV2-031 | MUST | zero server-contract change (no route/handler/form-field/seam diff) |
| AC-MWV2-032 | REQ-MWV2-032 | MUST | light-only preserved — no dark-theme toggle wired in `app.js` / `*.templ` |
| AC-MWV2-040 | REQ-MWV2-040/041 | MUST | `make build` clean; `go build ./...` + Windows cross-build + `go test ./internal/web/...` exit 0 |

## §D.1 Verification Commands

### AC-MWV2-001 — v2 ink + bg
```bash
grep -E -- '--color-ink:\s*#060606' internal/web/assets/console.css && \
grep -E -- '--color-bg:\s*#f4f4f4' internal/web/assets/console.css
# Expected: both match (exit 0)
```

### AC-MWV2-002 — achromatic neutral ramp; no teal residue
```bash
# v2 canon neutral values present:
grep -E -- '--neutral-400:\s*#9fa0a0' internal/web/assets/console.css && \
grep -E -- '--neutral-950:\s*#060606' internal/web/assets/console.css
# Earlier teal-tinted literals absent from the :root light token block.
# D5: the teal-absence grep is anchored to the :root block (awk window from
# ':root {' to the first '}') so it excludes any [data-theme="dark"] block,
# which legitimately carries darker literals. A whole-file grep would be a
# false-positive surface.
awk '/^:root[[:space:]]*\{/{r=1} r{print} /^\}/{if(r)exit}' internal/web/assets/console.css \
  | grep -E -- '--(color-ink|neutral-[0-9]+):[[:space:]]*#(09110f|1a1f1d|0e1513)' \
  && echo "FAIL: teal residue in :root" || echo "PASS: no teal residue in :root"
# Expected: canon matches present (first two greps); "PASS: no teal residue in :root"
```

### AC-MWV2-003 — shared v2 token values
```bash
grep -E -- '--color-success:\s*#2e8a63' internal/web/assets/console.css && \
grep -E -- '--border-1:\s*#d1d1d1' internal/web/assets/console.css && \
grep -E -- '--border-2:\s*#e6e6e6' internal/web/assets/console.css && \
grep -E -- '--border-strong:\s*#b5b5b5' internal/web/assets/console.css && \
grep -E -- '--fg-2:\s*#565656' internal/web/assets/console.css && \
grep -E -- '--fg-3:\s*#757575' internal/web/assets/console.css && \
grep -E -- '--border-focus-ring:\s*rgba\(61,\s*125,\s*95,\s*0\.16\)' internal/web/assets/console.css
# Expected: all match
```

### AC-MWV2-004 — signature accent = solid v2 point-green (confirmed)
```bash
grep -E -- '--gradient-signature:\s*#3d7d5f' internal/web/assets/console.css
# And the earlier gradient form is gone from the token definition:
! grep -E -- '--gradient-signature:\s*linear-gradient' internal/web/assets/console.css
# Expected: solid form present; linear-gradient form absent
```

### AC-MWV2-005 — token superset/subset handling (adopt v2-only + preserve console-only)
```bash
# D7 note: the realignment is value-only + name-stable (design.md §B.1) — the console
# and the v2 canon share the SAME token NAME set, so clause-1 ("adopt v2-only tokens")
# is VACUOUSLY satisfied: no v2 token NAME is orphaned/missing from console.css. The
# substantive (a) check is therefore that the previously-existing tokens now carry
# their v2 VALUES (the value delta the token-mapping table introduced), NOT that a
# brand-new token name appears. --color-primary-hover / --color-primary-active existed
# in the earlier console generation with non-canon values; they now hold the v2 values:
grep -E -- '--color-primary-hover:\s*#316750' internal/web/assets/console.css && \
grep -E -- '--color-primary-active:\s*#265240' internal/web/assets/console.css
# (a2) "no v2 token name orphaned" — every v2 :root token name resolves in console.css
#      (name-stable invariant; spot-check the accent/border/fg families):
for tok in color-ink color-bg color-success border-1 border-focus-ring \
           gradient-signature neutral-400 neutral-950 fg-2 fg-3; do
  grep -Eq -- "--$tok:" internal/web/assets/console.css || echo "ORPHAN v2 token name: --$tok"
done
# (b) console-only tokens with no v2 equivalent are preserved (the offline-safe
#     font fallbacks are console-only and MUST survive the realignment):
grep -E -- '--font-latin:\s*system-ui' internal/web/assets/console.css && \
grep -E -- '--font-mono:\s*ui-monospace' internal/web/assets/console.css
# Expected: (a) both value greps match; (a2) no ORPHAN line; (b) both console-only
# font fallbacks retained — v2 values adopted AND console-only tokens preserved.
```

### AC-MWV2-010 — 6-pose library (lowercase-kebab)
```bash
for p in coffee explaining pointing searching teaching thinking; do
  test -f "internal/web/assets/mascots/mascot-$p.png" || echo "MISSING mascot-$p.png"
done
# Expected: no MISSING output (all 6 present)
```

### AC-MWV2-011 — embed + serve
```bash
grep -- 'go:embed assets/console.css assets/app.js assets/i18n.js assets/htmx.min.js assets/fonts assets/mascots' internal/web/assets.go
# Runtime serve check (server started on loopback):
#   curl -sf http://127.0.0.1:<port>/static/mascots/mascot-thinking.png -o /dev/null
# Expected: embed directive present; each pose returns 200 (a web-serving test asserts the /static/mascots route)
```

### AC-MWV2-012 — header brand badge pose (confirmed = thinking)
```bash
grep -rn '/static/mascots/mascot-thinking.png' internal/web/board.templ internal/web/root.templ
# And the old reference is gone (mascot-coding.png removed, not retained):
! grep -rn '/static/mascots/mascot-coding.png' internal/web/board.templ internal/web/root.templ
# Expected: mascot-thinking.png referenced in both files; coding ref absent
```

### AC-MWV2-013 — talking asset removed
```bash
test ! -f internal/web/assets/mascots/mascot-talking.png
# Expected: file absent (exit 0)
```

### AC-MWV2-020 — orphan `project` panel removed (rendered-HTML boundary)
```bash
# The project panel is REMOVED (confirmed §B): no data-panel="project" is rendered.
grep -rn 'data-panel="project"' internal/web/root.templ internal/web/root_templ.go && \
  echo "FAIL: project panel still rendered" || echo "PASS: project panel removed"
# A Go rendered-HTML test asserts every remaining data-panel="<id>" has a matching data-tab="<id>":
go test -run 'TestConsole.*Panel|TestConsole.*Tab|TestNoOrphanPanel' ./internal/web/... 2>&1 | tail -5
# Expected: "PASS: project panel removed"; test ok (no orphan data-panel)
```
Given the console page is rendered, When the tab nav and panels are enumerated, Then the `project` panel is absent entirely (removed per §B) and no remaining `data-panel="<id>"` lacks a navigable `data-tab="<id>"`.

### AC-MWV2-021 — atomic-Save contract preserved
```bash
go test -run 'TestConsoleRendersReportTab|Test.*Save|Test.*Atomic|Test.*Submit' ./internal/web/... 2>&1 | tail -5
# Expected: ok — remaining panels still submit; no dependent field dropped without relocation/retirement
```

### AC-MWV2-030 — offline-safe (no remote fetch reintroduced)
```bash
! grep -E -- "@import\s+url\(\"https?:" internal/web/assets/console.css && \
! grep -E -- 'src:\s*url\("https?:' internal/web/assets/console.css
# Expected: no remote @import, no CDN font src (exit 0)
```

### AC-MWV2-031 — zero server-contract change
```bash
git diff --name-only origin/main -- internal/web/handlers.go internal/web/projectconfig.go internal/web/validate.go internal/web/server.go internal/web/app.go
# Expected: no handler/route/seam file appears with a CONTRACT change.
# (Cosmetic changes are acceptable ONLY if no route, form-field name, or seam signature changed;
#  the existing handlers_test.go / integration_test.go / security_test.go suites remain green.)
go test -run 'TestHandlers|TestIntegration|TestSecurity|TestHTMX' ./internal/web/... 2>&1 | tail -5
# Expected: ok (contract tests green)
```

### AC-MWV2-032 — light-only preserved (no dark toggle wired)
```bash
# No dark-theme toggle is wired in the interactive JS or the Templ views.
# (A [data-theme="dark"] token BLOCK may exist in console.css as inert dead code,
#  but no control sets/toggles data-theme, and no theme-toggle button renders.)
! grep -rniE 'data-theme|toggle.?dark|theme.?toggle|setAttribute\(["'\'']data-theme' \
    internal/web/assets/app.js internal/web/*.templ
# Expected: no match (no dark-toggle wiring in app.js or any *.templ)
```

### AC-MWV2-040 — build + cross-platform + test
```bash
make build && echo "make build exit=$?"
go build ./... && echo "build exit=$?"
GOOS=windows GOARCH=amd64 go build ./... && echo "win build exit=$?"
go test ./internal/web/... 2>&1 | tail -5
# Expected: all exit 0; templ generate regenerates *_templ.go cleanly
```

## §D.2 Definition of Done

- [ ] All MUST ACs pass with observed command output cited.
- [ ] `console.css` `:root` values match the v2 canon for every shared token; achromatic ramp; no teal residue in color tokens.
- [ ] 6 mascot poses present (lowercase-kebab), embedded, served; `mascot-talking.png` removed; header badge repointed to `mascot-thinking.png` (`mascot-coding.png` removed).
- [ ] Orphan `project` panel REMOVED (confirmed §B) — no `data-panel="project"` rendered; no remaining `data-panel` without a matching `data-tab`.
- [ ] Signature accent = solid `#3d7d5f`; light-only preserved (no dark toggle wired); token superset/subset handled (v2-only adopted, console-only preserved).
- [ ] Offline-safe font layer preserved verbatim; zero remote fetch reintroduced.
- [ ] Zero server-contract change; all handler/integration/security/htmx tests green.
- [ ] `make build` clean; `go build ./...` + Windows cross-build + `go test ./internal/web/...` exit 0; full `go test ./...` green (cascade check).
- [ ] All 4 plan-phase decisions confirmed via AskUserQuestion 2026-07-16 (recorded in plan.md §B); no open clarification remains.

## §D.3 Edge Cases

- **A restyle test pins an exact hex** (e.g. `restyle_test.go` asserts `#09110f`) — the test's expected literal MUST be updated to the v2 value in M3; a stale pin is a FAIL, not a token defect.
- **Signature-accent consumer expects a gradient** — now solid (§C), a component doing `linear-gradient(var(--gradient-signature))` would double-wrap; scan and fix in M3.
- **Removing `fieldsetProject` orphans a `pageView` field** — a `pageView` field consumed ONLY by `fieldsetProject` becomes dead after removal; remove it ONLY after grep confirms no other consumer (else leave it).
- **`sec.project.*` i18n keys** — after the panel removal these keys are inert; leaving them is acceptable (not a FAIL), pruning is optional housekeeping.
