# design.md — SPEC-DESIGN-MOAIWEBV2-001

> System-design decisions for the moai web console v2 alignment. Each section answers a "how" raised by spec.md. Decisions most likely to change are placed first (orphan-panel disposition, signature accent, mascot placement); the mechanical token substitution is last.

## §A. Orphan `project` Panel Disposition (highest-change decision)

### §A.1 The question

The `project` panel is unreachable (research.md §C) but holds real, tested config. "Clean up" = remove the dead DOM, OR restore reachability by re-adding the lost tab?

### §A.2 Decision (confirmed 2026-07-16)

**Confirmed: Option A — REMOVE the orphan** (via AskUserQuestion 2026-07-16; matches the mission's "absence grep for removed panel route" acceptance signal). The user accepted that `development_mode` / `git_convention` / `quality.*` lose their web-console editing surface (they remain editable via config files / CLI).

Removal scope:
1. Delete `templ fieldsetProject(view)` from `fieldsets.templ` (and its `detectionField` helper IF unused elsewhere — grep first).
2. Remove the `<div class="tabpanel" data-panel="project">@fieldsetProject(view)</div>` block from `root.templ`.
3. Update `schemaform_test.go::TestConsoleRendersReportTab` — its `launch`→`project` boundary span must retarget to `launch`→(next rendered panel, `llm`).
4. Remove or retarget `projectnested_parse_test.go` render assertions (they render `fieldsetProject` directly).
5. Prune `pageView` fields consumed ONLY by `fieldsetProject` (`CurDevelopmentMode`, `CurConvention`, `Cur*` detection/quality fields) — ONLY after grep confirms no other consumer; else leave inert.
6. `sec.project.*` i18n keys: leave inert (acceptable) or prune as housekeeping (optional).

**Option B — RESTORE (historical note, NOT chosen)** — would have added `{ID: "project", LabelKey: "sec.project.title", Baseline: "Project"}` to `consoleTabs()` (`schemaform.go`) + a `data-tab="project"` render assertion (zero field loss). Demoted to a historical note: the user confirmed REMOVE (Option A) on 2026-07-16, accepting the editability-loss consequence.

### §A.3 Rationale

- The mission explicitly frames the panel as "orphan" to "clean up" and hints an "absence grep for removed panel route" AC → removal is the stated intent.
- But the panel's fields are real + tested + submittable → removal is NOT purely cosmetic; it is a functional reduction. The verification-claim-integrity stance requires surfacing that consequence rather than silently removing tested capability → hence the clarification gate.
- Sequenced FIRST (M1) because it is the only structural DOM change and the highest-reversibility decision; resolving it early lets the rendered-HTML boundary tests be updated once.

## §B. Token-Mapping Architecture

### §B.1 Decision: value-only, name-stable substitution

Rewrite the `console.css` `:root` token VALUES to the v2 canon per the research.md §B.1 delta table. Token NAMES are unchanged (identical across both files), so every `var(--token)` consumer resolves without edit. This bounds the blast radius to the `:root` block + any test pinning an exact hex.

### §B.2 What is ported vs preserved

| Ported (values → v2 canon) | Preserved verbatim (do NOT touch) |
|---|---|
| color/neutral ramp, semantic, fg roles, borders, focus-ring, shadow base, primary hover/active | `@font-face` self-hosted woff2 block (Pretendard + Noto CJK), `assets/fonts/`, `--font-*` family fallbacks (system-ui, NOT the v2 canon's `Inter`/`JetBrains Mono` CDN) |

The v2 canon's Google Fonts `@import` and OTF `@font-face` are NOT ported (offline-safe invariant, REQ-MWV2-030). The console keeps its `--font-latin: system-ui` / `--font-mono: ui-monospace` fallbacks rather than the canon's CDN-loaded Inter/JetBrains Mono.

### §B.3 De-tint principle

The core conceptual shift: v2 neutrals/ink/shadows are **pure achromatic (hue 0%)**; the current console is **teal-tinted** (`#09110f` ink, `#1a1f1d`/`#0e1513` neutrals, `rgba(9,17,15,…)` shadows). The realignment de-tints all of these to the achromatic canon.

## §C. Signature Accent (solid vs gradient)

### §C.1 Decision (confirmed 2026-07-16)

**Confirmed: adopt the v2 solid `--gradient-signature: #3d7d5f`** (via AskUserQuestion 2026-07-16) — the v2 canon uses the point-green as a SOLID sole accent, not a gradient. This de-tints the accent (removes the teal `#09110f` gradient stop). The linear-gradient form is dropped consistently, including the inert `[data-theme="dark"]` override (solid dark point-green `#5a9a7e`).

**Migration hazard**: any component doing `background: linear-gradient(var(--gradient-signature))` (double-wrapping) would break when the token becomes a color. M3 scans components consuming `--gradient-signature` and adjusts (a component wanting `background: var(--gradient-signature)` works with either form; a `linear-gradient(var(--gradient-signature))` needs the wrapper removed).

**Confirmed via AskUserQuestion 2026-07-16**: adopt the v2 solid `#3d7d5f`; the linear-gradient is dropped.

## §D. Offline-Safe Preservation (unchanged mechanism)

- `console.css` keeps its self-hosted `@font-face` block (Pretendard woff2 subset weights 400/500/600/700/900 + Noto Sans CJK SC/JP subsets) VERBATIM.
- No `@import url("https://…")` and no CDN `src:` are introduced (REQ-MWV2-030).
- Font-family fallbacks stay system-ui / ui-monospace (not the canon's Inter / JetBrains Mono CDN).

## §E. Light-Only Preservation

- The console is light-only. If the v2 `[data-theme="dark"]` token block is imported alongside the `:root` values, it stays inert dead code (no dark toggle wired — mirrors SPEC-1 §H). Importing it is OPTIONAL; the minimal change is to port only the `:root` light values and skip the dark block entirely (preferred for the console since it has no dark infra to keep in sync).

## §F. Mascot Placement Architecture

### §F.1 Naming decision

Adopt lowercase-kebab filenames matching the existing console convention: `mascot-{coffee,explaining,pointing,searching,teaching,thinking}.png`. Source Capitalized bundle files (`MoAI-Mascot-Coffee.png` …) are copied + renamed. (Mirrors SPEC-1 §I casing bridge.)

### §F.2 Pose → surface map (console)

The console (settings + board UI) has fewer mascot surfaces than the docs-site. Proposed map:

| Surface | Pose | File | Status |
|---|---|---|---|
| Header brand badge (`board.templ` + `root.templ`) | Thinking | `mascot-thinking.png` | wired now (M2) — replaces `mascot-coding.png` |
| Board empty / first-run state | Searching | `mascot-searching.png` | OPTIONAL (placement clarification) |
| Save-success affordance | Coffee | `mascot-coffee.png` | OPTIONAL |
| Onboarding / help hint | Teaching / Explaining / Pointing | resp. files | OPTIONAL |

**Confirmed via AskUserQuestion 2026-07-16 — library + header only**: add all 6 to the library + embed + serve (REQ-MWV2-010/011); wire ONLY the header badge (REQ-MWV2-012). The additional live placements in the table above are NOT executed in this SPEC — the table is retained as a forward-looking placement reference for a future SPEC.

### §F.3 Header pose choice

The v2 set has no "coding" pose. Confirmed via AskUserQuestion 2026-07-16: header badge → `mascot-thinking.png` (closest cognitive/working pose). `mascot-coding.png` is removed (not retained).

### §F.4 Unused asset removal

`mascot-talking.png` (0 references, research.md §D.1) is deleted (REQ-MWV2-013). `mascot-coding.png` is deleted after the header repoint (unless retained per §F.3).

### §F.5 Motion (if live placements added)

If additional live placements are wired, reuse the v2 mascot-only bounce easing (`--easing-bounce: cubic-bezier(0.34,1.56,0.64,1)`, already in the token set) and respect `prefers-reduced-motion: reduce`. No new motion tokens.

## §G. Build & Verification Mechanism

- Every milestone touching `*.templ` / `console.css` / mascots runs `make build` (`templ generate` regenerates `*_templ.go`; assets re-embed; Go compiles).
- `*_templ.go` is NEVER hand-edited — it is generated output.
- Cross-platform: `GOOS=windows GOARCH=amd64 go build ./...` (no syscall in `internal/web`, so this is a formality guard).
- The existing rendered-HTML boundary tests are the DDD characterization net; token/mascot/panel changes are verified against them.

## §H. Decision Summary (most-likely-to-change first)

| # | Decision | Default | Override path |
|---|----------|---------|---------------|
| 1 | Orphan `project` panel | REMOVE (Option A) | RESTORE via `project` tab (Option B) if editability loss unacceptable |
| 2 | Signature accent | v2 solid `#3d7d5f` | retain linear-gradient |
| 3 | Header brand pose | `mascot-thinking.png` | another v2 pose, or retain `mascot-coding.png` |
| 4 | Mascot placement | library + header baseline | live-place all 6 per §F.2 map |
| 5 | Dark token block | skip (port `:root` light only) | import as inert dead code |
| 6 | Token realignment | value-only, name-stable | (none — mechanical) |

Decisions 1-4 were confirmed via AskUserQuestion 2026-07-16 (see plan.md §B); the "default" column above records the confirmed choice for each.

## §I. Cross-References

- `.moai/specs/SPEC-DESIGN-DOCSV2-001/design.md` §A/§F/§H/§I — the docs-site precedent (token unification, mascot placement, casing bridge, light-only).
- `.moai/state/ai-design-system/project/colors_and_type.css` — v2 token canon.
- `internal/web/{assets/console.css, board.templ, root.templ, fieldsets.templ, schemaform.go, assets.go, app.go}` — edit targets.
