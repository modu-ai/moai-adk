# plan.md — SPEC-DESIGN-MOAIWEBV2-001

> Implementation plan for the moai web console v2 design alignment. Milestones are ordered by **decision-reversibility** (highest-change-likelihood first): the orphan-panel disposition (UX/data-flow decision, blocked on a clarification) and mascot placement (UX flow + new assets) lead; the mechanical token-value substitution and the build/verify pass are deferred to the bottom.

## §A. Context

- **SPEC**: `SPEC-DESIGN-MOAIWEBV2-001`, Tier M, `internal/web/` only.
- **Branch**: current `release/v3.0.0` (Route A Hybrid Trunk main-direct, Tier M — commits pushed directly; no per-phase PR).
- **Development mode**: DDD (`cycle_type=ddd`) — brownfield restyle of an existing, tested package (`internal/web` has ~30 test files; behavior-preserving change). Characterization posture: the existing rendered-HTML boundary tests (`schemaform_test.go`, `restyle_test.go`, `board_test.go`, `projectnested_parse_test.go`) are the preservation net.
- **PRESERVE targets**: all server-side handlers (`handlers.go`, `projectconfig.go`, `validate.go`, `schemaform.go` parse path), the offline-safe font layer (`console.css` `@font-face` block + `assets/fonts/`), the atomic-Save form contract, the i18n dictionary (`assets/i18n.js`).
- **EXTEND targets**: `console.css` `:root` token values, `internal/web/assets/mascots/` (add 6 poses), `board.templ` + `root.templ` mascot references, and the orphan `project` panel wiring (`root.templ` + `schemaform.go consoleTabs()` + `fieldsets.templ fieldsetProject`).

## §B. Resolved Decisions (confirmed via AskUserQuestion 2026-07-16)

All 4 plan-phase decisions were resolved by the user via AskUserQuestion on 2026-07-16. They are recorded here as the binding execution contract; no open clarification remains.

- **Project-panel disposition — REMOVE (Option A).** Confirmed via AskUserQuestion 2026-07-16. The orphan `project` panel is deleted: remove `fieldsetProject`, the `data-panel="project"` block in `root.templ`, and update `schemaform_test.go` / `projectnested_parse_test.go`. The `development_mode` / `git_convention` / `quality.*` fields lose their web-console editing surface but REMAIN editable via yaml config files / CLI — the user accepted this consequence. Prune `pageView` fields consumed only by `fieldsetProject` after grep confirms no other consumer (scope discipline).

- **Signature accent — v2 solid `#3d7d5f` (drop the linear-gradient).** Confirmed via AskUserQuestion 2026-07-16. `--gradient-signature` is set to the solid v2 point-green `#3d7d5f`, replacing the current `linear-gradient(135deg, #3d7d5f 0%, #09110f 100%)`. M3 scans components consuming `--gradient-signature` for `linear-gradient(var(--gradient-signature))` double-wrapping and adjusts.

- **Header brand pose — `mascot-thinking.png`.** Confirmed via AskUserQuestion 2026-07-16. The header brand badge (`board.templ` + `root.templ`) is repointed from `/static/mascots/mascot-coding.png` to `/static/mascots/mascot-thinking.png`. `mascot-coding.png` is removed (not retained as a 7th asset).

- **Mascot placement — library + header only.** Confirmed via AskUserQuestion 2026-07-16. All 6 v2 poses are embedded in `assets/mascots/` and served (REQ-MWV2-010/011); the ONLY live UI swap is the header brand badge → `mascot-thinking.png` (REQ-MWV2-012). No additional live placements are wired (the design.md §F optional placement map is NOT executed in this SPEC).

## §C. Milestones

Ordered highest-change-likelihood → mechanical.

### M1 — Orphan `project` panel removal (UX/data-flow decision)

Highest-reversibility. Decision confirmed: REMOVE (§B).

- Delete `templ fieldsetProject` from `fieldsets.templ` (and its `detectionField` helper IF unused elsewhere — grep first); remove the `<div class="tabpanel" data-panel="project">@fieldsetProject(view)</div>` block from `root.templ`; update `schemaform_test.go` (the `data-panel="project"` boundary marker in `TestConsoleRendersReportTab` must move to the `launch`→next-panel boundary, `llm`) and remove/retarget `projectnested_parse_test.go` render assertions; prune now-dead `pageView` fields feeding only `fieldsetProject` IF they are unused elsewhere (verify each with grep before deletion — scope discipline). Leave `sec.project.*` i18n keys inert or prune as housekeeping.
- `make build`; `go test ./internal/web/...` green.

### M2 — Mascot 6-pose library + placement wiring (UX flow + new assets)

- Copy the 6 v2 poses from `.moai/state/ai-design-system/project/assets/characters/MoAI-Mascot-{Coffee,Explaining,Pointing,Searching,Teaching,Thinking}.png` into `internal/web/assets/mascots/` renamed lowercase-kebab: `mascot-{coffee,explaining,pointing,searching,teaching,thinking}.png` (REQ-MWV2-010/014).
- Remove the unused `mascot-talking.png` (0 references — REQ-MWV2-013).
- Repoint the header brand badge mascot `src` in `board.templ` + `root.templ` from `/static/mascots/mascot-coding.png` to `/static/mascots/mascot-thinking.png` (confirmed §B); update the two doc comments; remove `mascot-coding.png` (not retained) (REQ-MWV2-012).
- No additional live placements are wired (confirmed §B: library + header only). The design.md §F placement map is documented but NOT executed in this SPEC.
- The `//go:embed assets/mascots` directive already globs the dir — no `assets.go` edit needed unless a pose is placed outside `assets/mascots/` (REQ-MWV2-011).
- `make build`; verify `/static/mascots/mascot-*.png` served; `go test ./internal/web/...` green.

### M3 — v2 token realignment in `console.css` (mechanical value substitution)

- Rewrite the `console.css` `:root` block token VALUES to match the v2 canon (`colors_and_type.css`) per the design.md §B token-mapping table: ink `#060606`, bg/neutral-50 `#f4f4f4`, achromatic neutral ramp (`#e6e6e6`/`#d1d1d1`/`#b5b5b5`/`#9fa0a0`/`#757575`/`#565656`/`#3a3a3a`/`#242424`/`#141414`/`#060606`), semantic `--color-success #2e8a63`, fg roles (`#565656`/`#757575`), borders (`#d1d1d1`/`#e6e6e6`/`#b5b5b5`), focus-ring `rgba(61,125,95,0.16)`, shadow base `rgba(6,6,6,…)`, signature accent per §B decision, primary-hover/active (`#316750`/`#265240`).
- PRESERVE the `@font-face` self-hosted woff2 block VERBATIM (REQ-MWV2-030) — do NOT reintroduce the v2 canon's Google Fonts `@import` or OTF `@font-face`.
- Keep the token NAMES stable (only values change) so component CSS consuming the tokens continues to resolve.
- `make build` (console.css re-embedded); `go test ./internal/web/...` green (restyle tests assert token presence/values — update expected literals where a test pins an exact hex).

### M4 — Build, cross-platform, and verification pass (mechanical)

- `make build` (final templ generate + embed + compile).
- `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `go test ./internal/web/...` exit 0; full `go test ./...` for cascade safety.
- Absence grep: no remote font/style fetch reintroduced; no orphan panel; 6 mascot files present.
- Commit + push per Route A (Conventional Commits, `🗿 MoAI` trailer).

## §D. Technical Approach

- **DDD behavior-preservation**: the console has a dense rendered-HTML test suite. Treat those tests as the characterization net — run them after each milestone. Token-value edits and mascot-path edits are behavior-preserving at the server-contract level (REQ-MWV2-031); only pinned-hex test literals and the panel-boundary assertion need updating.
- **Token realignment is value-only, name-stable**: components reference `var(--token)`; changing values (not names) keeps every consumer resolving. This bounds M3 blast radius to the `:root` block + any test that pins an exact hex.
- **Mascot embed is glob-based**: `//go:embed assets/mascots` already captures new files; adding poses requires no `assets.go` change (REQ-MWV2-011).
- **Orphan cleanup is the only structural change**: M1 changes rendered DOM structure (removes or re-nav-links a panel); it is sequenced first so its clarification is resolved before mechanical work, and so the rendered-HTML boundary tests are updated once.

## §E. Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Removing `fieldsetProject` drops editability of dev-mode/git-convention/quality config from the web console | Accepted by the user (§B, confirmed 2026-07-16); these settings remain editable via yaml config files / CLI. |
| A restyle test pins an exact hex (e.g. `#09110f`) → M3 breaks it | Enumerate pinned-hex assertions via grep before M3; update expected literals to the v2 value in the same milestone. |
| Reintroducing a remote font fetch by copying the v2 canon's `@import` / OTF `@font-face` wholesale | REQ-MWV2-030 unwanted-behavior AC + absence grep for `@import url("http` and CDN `src`; PRESERVE the existing `@font-face` block verbatim, port ONLY `:root` values. |
| Signature-accent change (now solid, §B) breaks a component expecting a gradient value | M3 scans components consuming `--gradient-signature` for `linear-gradient(var(--gradient-signature))` double-wrap misuse and adjusts. |
| `data-panel="project"` boundary marker in `schemaform_test.go TestConsoleRendersReportTab` breaks when the panel is removed | M1 updates that test's boundary index to the `launch`→next-panel span. |
| `make build` not run → stale `*_templ.go` committed | REQ-MWV2-040 gate; M4 final `make build` + `go build ./...` verification. |
| Windows build tag regression (unlikely — no syscall) | M4 `GOOS=windows` cross-build check. |

## §F. Phase 4 Mode Selection (pre-populated placeholder)

Mode selection is recorded by the orchestrator in `progress.md §F` before the first run-phase `Agent()` spawn. Expected: **Mode 5 (sub-agent, sequential)** — coding-heavy single-package work, ~6-12 files, single domain (frontend/web). Not multi-domain-research (Mode 4) and not ≥30-file mechanical (Mode 6).

## §G. Anti-Patterns to Avoid

- Editing `*_templ.go` by hand (regenerated by `templ generate` — edit the `.templ` source).
- Copying the v2 `colors_and_type.css` `@font-face` / `@import` wholesale (reintroduces remote fetch — port `:root` values only).
- Touching `docs-site/`, `internal/template/templates/`, or `.claude/settings*.json` (out of scope).
- Deleting a `pageView` field feeding `fieldsetProject` without grepping for other consumers first (scope-discipline / dead-code hazard).
- Approximate token values ("close to v2") — the AC greps for exact v2 hex literals.

## §H. Cross-References

- `.moai/specs/SPEC-DESIGN-DOCSV2-001/design.md` §A (token unification) + §F (mascot placement) — the SPEC-1 precedent this SPEC mirrors for the console.
- `internal/web/assets.go` — embed directive (mascot glob).
- `internal/web/schemaform.go` `consoleTabs()` — the tab SSOT the orphan cleanup edits.
