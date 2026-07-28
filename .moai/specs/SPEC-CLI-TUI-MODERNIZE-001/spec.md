---
id: SPEC-CLI-TUI-MODERNIZE-001
title: "interactive TUI surface modernization — Bubble Tea preview redesign + dead confirm removal + huh theme unification"
version: "0.1.3"
status: in-progress
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: "internal/cli/update, internal/merge, internal/cli, internal/cli/wizard"
lifecycle: spec-anchored
tags: "cli, tui, tux, bubbletea, huh, theme, preview, dead-code, refactor"
tier: L
related_specs: [SPEC-CLI-TUX-INIT-UPDATE-001, SPEC-CLI-TUX-V3-003, SPEC-V3R3-CLI-TUI-001, SPEC-CLI-WIZARD-RESTRUCTURE-001]
---

# SPEC-CLI-TUI-MODERNIZE-001 — interactive TUI surface modernization

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-25 | manager-spec | Initial plan-phase authoring. Scope confirmed by user as "interactive-surface-wide unification" (three milestones M1/M2/M3). Vocabulary decision (TUX vs TUI vs huh form runtime) confirmed by user and bound in §A.2. |
| 0.1.3 | 2026-07-25 | manager-spec | Delta re-audit repair (CONDITIONAL 0.90 — threshold cleared). NEW-1 (MUST-FIX): the v0.1.2 rendered-probe method asserted a **full-CSI prefix**, which lipgloss v2 breaks by merging all SGR parameters into one CSI — a bold+coloured cell renders `\x1b[1;38;2;...m`, so the foreground-only prefix is not a substring. Executed a combined-attribute render to confirm before fixing; replaced with the **SGR-parameter-substring** method, now recorded with its evidence in acceptance.md §C.2 and applied to AC-TUIM-004/-008/-030b. NEW-2: §G closure gate still demanded four diff-reachable quit keys — restated to three plus AC-TUIM-014f. NEW-3: AC-TUIM-039 gained case (e), the invalid-`MOAI_THEME` short-circuit. Cosmetics: AC-TUIM-023 golden count corrected to a measured 12 (+12 under `postm4/`); two `(Event-detected)` labels corrected to `(Event-driven)`; AC row order 014e/014f. |
| 0.1.2 | 2026-07-25 | manager-spec | plan-audit repair (FAIL 0.74 → repair). D1: three ACs asserted a hex substring that lipgloss v2 never emits (it renders decimal RGB SGR) — assertion method restated. D2: AC-TUIM-014d demanded a quit key unreachable from the diff sub-view — scoped correctly. D3/S1: two AC greps used a markdown-escaped `\|` inside an ERE (a literal pipe, matching nothing) — moved to fenced blocks with corrected regexes, each positive-controlled. D4: Tier L artifact set completed with `design.md` + `research.md`. S4: REQ-TUIM-002 / -034 qualified to *executable code* to stop contradicting AC-TUIM-020. S6/S7: added ACs for REQ-TUIM-001/-003 and for the `MOAI_THEME` precedence chain. S9: requirement count corrected 40 → 41. |
| 0.1.1 | 2026-07-25 | manager-spec | Both open clarifications resolved by user; markers removed. **D-4 = HYBRID residue policy** (table inline, diff on the alternate screen, `esc` restores the inline table, single-line result summary on exit) — REQ-TUIM-019 rewritten, REQ-TUIM-022 added for the exit frame. **D-5 = SPLIT verification** (a `NO_COLOR` structure golden plus token-application unit tests, with the self-trip burden carried by the token tests) — REQ-TUIM-050/051 rewritten. Renderer verification confirmed the mid-run `AltScreen` toggle is first-class in bubbletea v2 and surfaced a quit-path constraint (plan.md §B.10a). |

---

## §A Context

### §A.1 Origin

`moai update` still presents a visually dated confirmation screen even though `moai init` and the `moai update` **output** path were modernized and shipped. The gap is not a deployment failure — it is a scope carve-out that was honoured exactly as written.

`SPEC-CLI-TUX-INIT-UPDATE-001` (status `in-progress`, title "moai init/update terminal **output TUX** redesign") declares an explicit exclusion:

> `### Out of Scope — the Bubble Tea change-preview flow`
> `internal/cli/update/preview_tui.go` (the interactive classification table + "[enter] view diff / [y] confirm" prompt) beyond, at most, a light glyph/pill touch to match the shared symbol source. No re-architecture of its Bubble Tea model.

and marks the same file `exempt` from its glyph-SSOT requirement (REQ-TUXIU-001). The shipping merge commit `b1ea545e2` ("feat(SPEC-CLI-TUX-INIT-UPDATE-001): modernize moai init/update TUI + restore MoAI-ADK logo (#1145)") touched 56 files; `preview_tui.go` and `preview_fallback.go` are not among them.

This SPEC picks up precisely the re-architecture that carve-out deferred, and extends it across the remaining interactive surfaces so the three of them stop diverging.

### §A.2 Vocabulary (USER-CONFIRMED — binding for this SPEC)

This SPEC binds a three-layer vocabulary. Every requirement below names the layer it governs.

| Term | Layer | Definition | Precedent |
|------|-------|------------|-----------|
| **TUX** (Terminal User eXperience) | output-rendering | `internal/tui` + `internal/cli/uikit`. Every function returns a `string` that is simply printed. No state, no key-input loop; output persists in scrollback. | `SPEC-CLI-TUX-INIT-UPDATE-001`, titled "terminal **output TUX** redesign" |
| **TUI** (Terminal User Interface) | interactive program | Bubble Tea `Model`/`Update`/`View` state machines with a key-input loop that repaint frames. | `SPEC-V3R3-CLI-TUI-001` |
| **huh form runtime** | input forms | The `huh` form library. A third surface, sitting on the interactive side of the TUX/TUI split but distinct from a hand-written Bubble Tea program. | `internal/cli/huh_theme.go`, `internal/cli/wizard/wizard.go` |

**Documented naming caveat.** The Go package named `internal/tui` in fact implements the **TUX** layer, not the TUI layer. This is a known, accepted naming artifact. It is recorded here so readers are not misled; it is **not** repaired by this SPEC (see §C).

### §A.3 The three divergent interactive surfaces

| Surface | File | Layer | Current state |
|---------|------|-------|---------------|
| Change preview | `internal/cli/update/preview_tui.go` (258 lines) | TUI | Imports **no** `internal/tui`. Builds `table.New(...)` with no `table.WithStyles`, so it renders the bubbles default (unstyled) table. Header is plain text; key hints are bare inline strings. |
| Change preview fallback | `internal/cli/update/preview_fallback.go` (71 lines) | TUX | Plain-text summary. Structurally ANSI-free **by design** (documented at lines 10-15, asserted by AC-TUX3-010). Layout is unstructured. |
| Legacy merge confirm | `internal/merge/confirm.go` (954 lines) | TUI | Dead: `ConfirmMerge` (defined line 915) has **zero** production callers. Carries a parallel styling system: 41 direct `lipgloss.` usages including 4 raw hex colour literals. |
| huh v1 theme | `internal/cli/huh_theme.go` (83 lines) | huh form runtime | Correctly maps `internal/tui` tokens. Consumed by `init.go`, `profile_setup.go`, `update.go`. |
| huh v2 theme | `internal/cli/wizard/wizard.go` tail (lines ~480-589 of 589) | huh form runtime | Correctly maps `internal/tui` tokens via a separate `wizardTokenSet` indirection. Deliberately separate from v1 — different `Theme` types across the library version boundary (documented at `huh_theme.go:21-30`). |

### §A.4 Why CI stayed green

The existing golden characterization set (`internal/cli/tuxiu_characterization_test.go` over `internal/cli/testdata/tuxiu/*.golden`) covers only the `--yes` / non-TTY paths — six golden pairs across `init`/`update` × `tty`/`notty`/`nocolor`. The interactive program is never invoked there.

The interactive model **is** exercised — `internal/cli/update/preview_test.go` carries 12 test functions that drive `newPreviewModel` and its accessors headlessly. But **every assertion is content-only**: `strings.Contains` over class labels and file paths. Not one assertion touches presentation — no theme token, no border, no styled-table structure. A redesign that skipped this file therefore could not fail any test.

The blind spot is not "the interactive path is untested". It is "the interactive path has no **presentation** regression coverage". This SPEC closes that specific gap.

---

## §B Requirements (GEARS)

### Group A — Vocabulary and layer discipline

- **REQ-TUIM-001** (Ubiquitous): The SPEC artifact set **shall** bind the three-layer vocabulary defined in §A.2 (TUX / TUI / huh form runtime), and every requirement in this SPEC **shall** name the layer it governs.
- **REQ-TUIM-002** (Ubiquitous): The `internal/tui` package **shall** remain the single source of colour tokens for all three layers, and no **executable line** in a production file outside `internal/tui/` **shall** contain a raw hex colour literal. Comment-only occurrences are outside this requirement's scope — two exist at baseline in `internal/cli/wizard/styles.go` and are documentation of a retired palette, not colour decisions.
- **REQ-TUIM-003** (Ubiquitous): The SPEC **shall** record the "package `internal/tui` implements the TUX layer" naming caveat as a documented artifact, and the package **shall not** be renamed under this SPEC.

### Group B — M1: change-preview interactive TUI redesign (TUI + TUX layers)

- **REQ-TUIM-010** (Ubiquitous): The change-preview TUI **shall** resolve its colour tokens from `internal/tui` rather than rendering with component library defaults.
- **REQ-TUIM-011** (Ubiquitous): The preview table **shall** be constructed with an explicit `table.Styles` value derived from the resolved `tui.Theme`, replacing the current default-styled `table.New(...)` construction.
- **REQ-TUIM-012** (Event-driven): **When** the preview model renders the table view, the model **shall** render the per-class count summary through an `internal/tui` structural primitive rather than the current bare `Classification summary:` text block.
- **REQ-TUIM-013** (Event-driven): **When** the preview model renders any sub-view, the model **shall** render its key hints through `tui.HelpBar` rather than the current inline literal hint strings.
- **REQ-TUIM-014** (Ubiquitous): Every class label the preview renders **shall** carry a semantic colour role drawn from the resolved theme, and the four label **strings** **shall** remain textually identical to their current `ChangeClass.String()` values.
- **REQ-TUIM-015** (Ubiquitous): The preview **shall** resolve its light/dark axis through the `internal/tui` OS-resolution entry points, thereby inheriting the canonical precedence chain `NO_COLOR` > `MOAI_THEME` (`light`/`dark`) > `DetectDark()` > dark-default (`internal/tui/detect.go` `Resolve`) already used by the two huh theme factories. The preview **shall not** re-implement or partially re-order that chain.
- **REQ-TUIM-016** (State-driven): **While** `NO_COLOR` is set, the preview TUI **shall** emit no ANSI colour sequence.
- **REQ-TUIM-017** (Ubiquitous): Any status glyph the preview emits **shall** resolve from the canonical `tui.Glyph*` constants, and the preview **shall not** introduce a rune outside the existing status-glyph whitelist.
- **REQ-TUIM-018** (Ubiquitous): The non-interactive fallback **shall** be restructured into an aligned, grouped, card-shaped plain-text layout, and **shall not** emit an ANSI escape sequence under any environment.
- **REQ-TUIM-019** (State-driven): **While** the preview model is in its table sub-view, the model **shall** render inline in the normal output stream so the identity band and classification card the user is confirming against stay visible; **while** the model is in its diff sub-view, the model **shall** render on the terminal's alternate screen.
- **REQ-TUIM-019a** (Event-driven): **When** the user leaves the diff sub-view, the model **shall** restore the inline table sub-view with the prior scrollback intact.
- **REQ-TUIM-020** (Ubiquitous): The redesigned preview **shall** derive classification solely from the existing single-source-of-truth entry point, and **shall not** introduce a parallel classification heuristic.
- **REQ-TUIM-021** (Ubiquitous): The preview model's headless contract-test accessors **shall** remain available so every presentation property is verifiable without a real terminal.
- **REQ-TUIM-022** (Event-driven): **When** the preview program exits, the interactive frame **shall** be replaced by a single-line result summary naming the outcome and the file count, and the classification content **shall not** appear a second time in scrollback.
- **REQ-TUIM-022a** (Event-driven): **When** a quit request arrives while the model is in its diff sub-view, the model **shall** return to the inline table sub-view before quitting, so the exit summary is written to the main screen rather than to an alternate screen that is about to be discarded.

### Group C — M2: dead legacy confirmation TUI removal (TUI layer)

- **REQ-TUIM-030** (Ubiquitous): The legacy interactive merge-confirmation program in `internal/merge/confirm.go` — the entry point, its Bubble Tea model, its list-item adapter, and the formatter rendering stack together with the helpers reachable only from them — **shall** be removed.
- **REQ-TUIM-031** (Ubiquitous): The removal **shall** be surgical: every declaration with a production consumer outside the removed program **shall** be preserved with its field set and exported name unchanged.
- **REQ-TUIM-032** (Ubiquitous): The plan artifact **shall** carry an explicit keep/delete inventory enumerating every top-level declaration in `internal/merge/confirm.go` together with the reachability evidence that classifies it.
- **REQ-TUIM-033** (Event-driven): **When** the removal lands, the `internal/merge` package **shall** carry zero dependency on the Bubble Tea and bubbles component libraries and zero direct `lipgloss` usage.
- **REQ-TUIM-034** (Ubiquitous): The removal **shall** eliminate the raw hex colour literals currently present on executable lines of `internal/merge/confirm.go`, bringing the count of **executable** production lines carrying a raw hex colour literal outside `internal/tui/` to zero. The two comment-only occurrences in `internal/cli/wizard/styles.go` are out of scope and are expected to remain (see AC-TUIM-020).
- **REQ-TUIM-035** (Ubiquitous): The non-TTY confirmation guard governing the **live** confirmation path **shall** remain in force after the removal, with its behaviour unchanged.
- **REQ-TUIM-036** (Ubiquitous): Test files that exist solely to exercise the removed program **shall** be removed or retargeted such that no surviving test references a deleted symbol.
- **REQ-TUIM-037** (Ubiquitous): The removal **shall not** change which files `moai update` deploys, skips, backs up, or classifies.

### Group D — M3: input-form theme unification (huh form runtime layer)

- **REQ-TUIM-040** (Ubiquitous): The huh v1 theme factory and the huh v2 theme factory **shall** remain two separate factories; the library version boundary **shall not** be merged into a single factory.
- **REQ-TUIM-041** (Ubiquitous): Both huh theme factories **shall** apply the same `internal/tui` token-to-role assignment, and every divergence surfaced by the audit **shall** be reconciled toward the token set the redesigned preview TUI uses.
- **REQ-TUIM-042** (Ubiquitous): The audit **shall** enumerate, per factory, every style field the sibling factory sets that the factory under audit does not, and each such gap **shall** be either closed or recorded with a stated reason.
- **REQ-TUIM-043** (Ubiquitous): This SPEC's edits to `internal/cli/wizard/wizard.go` **shall** be confined to the theme region at the file tail, and **shall not** touch the question/group assembly region owned by SPEC-CLI-WIZARD-RESTRUCTURE-001.
- **REQ-TUIM-044** (Ubiquitous): The wizard theme functions **shall not** be relocated to another file.
- **REQ-TUIM-045** (Ubiquitous): Both factories **shall** continue to resolve the light/dark axis through their existing package-level indirection variables, so tests can force either axis without mutating the process environment.

### Group E — Cross-cutting invariants

- **REQ-TUIM-050** (Ubiquitous): The SPEC **shall** add presentation-level regression coverage for the interactive preview surface through **two complementary mechanisms** — a structure golden captured under `NO_COLOR`, and token-application unit tests — and both mechanisms **shall** execute headlessly without requiring a real terminal.
- **REQ-TUIM-050a** (Ubiquitous): The structure golden **shall** pin layout, borders, column alignment, and row order, and **shall** remain stable across `internal/tui` palette changes so a palette edit does not break `internal/cli/update` tests.
- **REQ-TUIM-050b** (Ubiquitous): The token-application tests **shall** assert that each semantic role resolves to the correct `internal/tui` theme token, comparing against the token values read from the resolved `Theme` rather than against hard-coded hex strings.
- **REQ-TUIM-051** (Event-driven): **When** the theme wiring is removed from the preview, the token-application tests **shall** fail. The structure golden alone cannot detect that removal — it is deliberately palette-insensitive — so the token-application mechanism carries the whole regression-detection burden for the defect class this SPEC exists to repair.
- **REQ-TUIM-052** (Ubiquitous): Every milestone **shall** preserve a clean cross-platform build for the project's supported target set.
- **REQ-TUIM-053** (Ubiquitous): Statement coverage for each affected package **shall not** regress below its measured pre-change baseline.
- **REQ-TUIM-054** (Ubiquitous): M1 **shall** be landable independently of M2 and M3, since M1 alone resolves the user-visible defect.
- **REQ-TUIM-055** (Ubiquitous): The SPEC **shall not** introduce a new module dependency; only the terminal-UI stack already present in `go.mod` **shall** be used.
- **REQ-TUIM-056** (Ubiquitous): The SPEC **shall not** modify any file under `internal/template/templates/`, so the Template-First rule and the template-neutrality guard do not apply.

**Requirement count: 41** (Group A 3, Group B 15, Group C 8, Group D 6, Group E **9**). Sub-lettered IDs (`REQ-TUIM-019a`, `-022a`, `-050a`, `-050b`) are paired sub-requirements of their parent, counted individually. Verified: `grep -c '^- \*\*REQ-TUIM-' spec.md` → `41`.

---

## §C Exclusions

The following are explicitly **out of scope** for this SPEC. Each is expressed as an `### Out of Scope` sub-heading to satisfy the exclusions contract.

### Out of Scope — renaming the internal/tui package
- Renaming `internal/tui` to reflect that it implements the TUX layer. The naming caveat is documented in §A.2, not repaired. The blast radius (every importing package, every test, every golden fixture) is disproportionate to the clarity gain.
- Renaming `internal/cli/uikit`, or redistributing responsibilities between `tui` and `uikit`.

### Out of Scope — the wizard's question structure
- The wizard's question set, page layout, group assembly, and default values. These are owned by `SPEC-CLI-WIZARD-RESTRUCTURE-001` (status `draft`, run pending in a parallel session).
- Any edit to the question/group assembly region of `internal/cli/wizard/wizard.go`. This SPEC touches only the theme region at the file tail (REQ-TUIM-043).

### Out of Scope — internal/tui public API changes
- Adding, removing, or altering the signature of any exported `internal/tui` or `internal/cli/uikit` primitive. This SPEC is a **consumer** of TUX primitives.
- Adding new `Theme` tokens or altering the existing token palette. Should the redesign surface a genuine primitive gap, extending the API is permitted only with an explicit written justification recorded in the plan artifact.

### Out of Scope — the already-redesigned output path
- The `moai init` and `moai update` **output** rendering already delivered by `SPEC-CLI-TUX-INIT-UPDATE-001`. Its golden fixtures are a baseline this SPEC must not disturb, not a surface this SPEC rewrites.
- The root-help logo, the identity band, the classification summary card, and the checklist steps on the output path.

### Out of Scope — functional and data behaviour
- Which files `moai update` deploys, skips, backs up, or classifies.
- Merge classification logic, conflict detection, file-count arithmetic, and the user-owned-namespace predicate.
- The three-way merge engine (`internal/merge/strategies.go`, `differ.go`, `three_way.go`, `evolvable_zone.go`), which is untouched by the M2 removal.

### Out of Scope — new dependencies and machine-readable output
- Introducing any new terminal-UI library, colour library, or module dependency.
- Adding a `--json` / `--format` structured-output mode to `moai update`.

### Out of Scope — template distribution
- Any file under `internal/template/templates/`. This SPEC changes only `internal/` Go source, so the Template-First rule and the template-neutrality CI guard have no bearing on it.

---

## §D Constraints

| # | Constraint | Source |
|---|-----------|--------|
| D1 | No raw hex colour literal outside `internal/tui/` | `CLAUDE.local.md` §14; `AC-CLI-TUI-013` |
| D2 | Env-var names come from `internal/config/envkeys.go`; no inline `os.Getenv` string literals | `CLAUDE.local.md` §14 |
| D3 | Test temp directories via `t.TempDir()`; no writes to the project root | `CLAUDE.local.md` §6 |
| D4 | CLI code must not call `AskUserQuestion` or `mcp__askuser__*` | `internal/cli/CLAUDE.md` |
| D5 | stdout = machine-readable output; stderr = human progress and errors; never mixed | `internal/cli/CLAUDE.md` |
| D6 | Cross-platform: the supported target set must build cleanly | `internal/cli/CLAUDE.md` |
| D7 | The fallback path's structural ANSI-free property is an invariant to preserve, not a behaviour to renegotiate | `preview_fallback.go:10-15`; `AC-TUX3-010` |
| D8 | Edits to `internal/cli/wizard/wizard.go` confined to the theme region at the file tail | parallel-session merge-conflict avoidance; REQ-TUIM-043 |
| D9 | Every recorded measurement traces to a command actually executed | `.claude/rules/moai/core/verification-claim-integrity.md` §2 |

---

## §E Success Criteria

1. Running `moai update` interactively presents a preview whose visual language is indistinguishable in family from the already-modernized `moai init` / `moai update` output path, and the run appears in scrollback exactly once.
2. `internal/merge` contains no interactive terminal program, no component-library dependency, and no raw hex colour literal — while every type its production consumers depend on survives untouched.
3. The two huh theme factories produce a matching token-to-role mapping, with every remaining divergence deliberate and documented.
4. A presentation regression in the preview surface fails a test, headlessly.
5. `SPEC-CLI-WIZARD-RESTRUCTURE-001` merges without a conflict in `internal/cli/wizard/wizard.go`.

---

## §F Cross-References

- `.moai/specs/SPEC-CLI-TUX-INIT-UPDATE-001/spec.md` — the carve-out this SPEC picks up (§C "Out of Scope — the Bubble Tea change-preview flow"; REQ-TUXIU-001 exemption)
- `.moai/specs/SPEC-CLI-TUX-V3-003/` — origin of `preview_tui.go` / `preview_fallback.go` and of AC-TUX3-008/009/010/014
- `.moai/specs/SPEC-V3R3-CLI-TUI-001/` — origin of the `internal/tui` token system and AC-CLI-TUI-013
- `.moai/specs/SPEC-CLI-WIZARD-RESTRUCTURE-001/` — the parallel SPEC whose `wizard.go` edit region this SPEC must avoid
- `design.md` (sibling) — interactive-surface design: the hybrid inline/alt-screen state model, the TUX-token → interactive-role map, view composition, and the exit-frame contract
- `research.md` (sibling) — the measured substrate: bubbletea v2 `View.AltScreen` + renderer findings, the bubbles v2 table styling API, the lipgloss v1-vs-v2 rendering difference, and the M2 reachability census
- `internal/cli/CLAUDE.md` — module conventions
- `CLAUDE.local.md` §6, §14 — test isolation and hardcoding prevention
