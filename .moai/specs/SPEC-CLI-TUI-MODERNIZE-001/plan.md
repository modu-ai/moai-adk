---
id: SPEC-CLI-TUI-MODERNIZE-001
title: "Implementation plan — interactive TUI surface modernization"
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
---

# Implementation Plan — SPEC-CLI-TUI-MODERNIZE-001

> **Ordering note.** §F milestones are ordered by decision-reversibility: the decisions most likely to change (user-facing UX flow, new type interfaces, data-model relocation) lead; mechanical sweeps are deferred to the bottom. Reviewer attention belongs at the top.

---

## §A Context

Three interactive surfaces have drifted apart from the TUX output layer that `SPEC-CLI-TUX-INIT-UPDATE-001` modernized. §A.1-§A.4 of `spec.md` carry the narrative; this plan carries the measured evidence and the execution sequence.

### §A.1 Scope shape

| Milestone | Surface | Layer | Independent? |
|-----------|---------|-------|--------------|
| M1 | `internal/cli/update/preview_tui.go` + `preview_fallback.go` | TUI + TUX | **Yes** — resolves the user-visible defect alone |
| M2 | `internal/merge/confirm.go` + its test files | TUI (dead) | Yes — depends on nothing in M1 |
| M3 | `internal/cli/huh_theme.go` + `internal/cli/wizard/wizard.go` tail | huh form runtime | Depends on M1's token decisions landing first |

---

## §B Measured findings

Every figure below traces to a command executed during plan authoring, on branch `feat/SPEC-AGENT-PARALLEL-OPT-001`. Anything not measured is marked **unmeasured**.

### §B.1 File sizes

Command: `wc -l internal/cli/update/preview_tui.go internal/cli/update/preview_fallback.go internal/merge/confirm.go internal/merge/confirm_test.go internal/cli/huh_theme.go internal/cli/wizard/wizard.go internal/cli/wizard/styles.go`

| File | Lines |
|------|------:|
| `internal/cli/update/preview_tui.go` | 258 |
| `internal/cli/update/preview_fallback.go` | 71 |
| `internal/merge/confirm.go` | 954 |
| `internal/merge/confirm_test.go` | 971 |
| `internal/cli/huh_theme.go` | 83 |
| `internal/cli/wizard/wizard.go` | 589 |
| `internal/cli/wizard/styles.go` | 156 |

A **third** merge test file exists that the task brief did not name: `internal/merge/confirm_coverage_test.go` (15,868 bytes). It also constructs `confirmModel`, `fileListItem`, and `AnalysisFormatter`. M2 must handle it alongside `confirm_test.go`.

### §B.2 Carve-out provenance

Command: `git show --name-only --format="" b1ea545e2 | grep -E "preview_tui|preview_fallback|confirm\.go|wizard\.go|huh_theme"` → **no matches** (grep exit 1). Total files in the merge: **56**.

`git log -1 --format="%h %s" b1ea545e2` → `b1ea545e2 feat(SPEC-CLI-TUX-INIT-UPDATE-001): modernize moai init/update TUI + restore MoAI-ADK logo (#1145)`

Carve-out text confirmed at `.moai/specs/SPEC-CLI-TUX-INIT-UPDATE-001/spec.md` — the `### Out of Scope — the Bubble Tea change-preview flow` heading, and the `exempt` marking of `preview_tui.go` inside REQ-TUXIU-001. That SPEC's frontmatter reads `status: in-progress`, `tier: M`.

### §B.3 `ConfirmMerge` reachability — dead confirmed

Command: `grep -rn "ConfirmMerge" --include="*.go" .`

| Site | Kind |
|------|------|
| `internal/merge/confirm.go:915` | definition |
| `internal/merge/confirm_test.go:478` | **only call site** — a test |
| `internal/cli/update.go:1124`, `:1131`, `:1137` | comment prose |
| `internal/cli/update/preview.go:5`, `:18`, `:59` | comment prose |
| `internal/cli/coverage_improvement_test.go:4286-4290`, `internal/cli/update_skip_sync_test.go:109-112` | comment prose in tests |

Zero production callers. The live path routes `update.go` → `confirmViaPreview` → `update.PreviewClassification` instead.

### §B.4 Keep/delete inventory for `internal/merge/confirm.go` (REQ-TUIM-032)

Reachability method: for each top-level declaration, `grep -rn "\bSYM\b" --include="*.go" .` filtered to exclude `_test.go` files and `confirm.go` itself. A symbol with a hit surviving that filter has a production consumer.

**KEEP — live production types (consumers outside `internal/merge`):**

| Declaration | confirm.go line | Production consumer |
|-------------|----------------:|---------------------|
| `type MergeAnalysis struct` | 26 | `internal/cli/update.go:1110` `toPreviewInputs(analysis merge.MergeAnalysis, …)`; `internal/cli/update/merge/merge.go:230` `AnalyzeMergeChanges(…) mrg.MergeAnalysis` |
| `type FileAnalysis struct` | 35 | `internal/cli/update_tux.go:88` `classifyUpdateCounts(files []merge.FileAnalysis)`; **`internal/cli/update/plan/plan.go:63, 64, 88`** — `AnalyzeFiles(...) []merge.FileAnalysis`, a local slice, and a **named-field composite literal** at `:88-94`; plus the two above via `MergeAnalysis.Files` |

**`plan.go:88` is the field-set-sensitive site** (added at v0.1.2 — the v0.1.1 table omitted it although the census of 6 already counted its hits). It constructs the struct by field name:

```go
files = append(files, merge.FileAnalysis{
    Path:      displayPath,
    Changes:   changeType,
    Strategy:  strategy,
    RiskLevel: risk,
    Note:      "",
})
```

Renaming or removing any of those five fields breaks this site at compile time. It is therefore the strongest mechanical guarantee behind REQ-TUIM-031's "field set unchanged", and AC-TUIM-018 names it explicitly.

Cross-package reference census — `grep -rn "merge\.[A-Z][A-Za-z]*" internal/ cmd/ pkg/ | grep -v _test.go | grep -v "^internal/merge/"` — counted `merge.FileAnalysis` 6, `merge.MergeAnalysis` 3, `merge.ConfirmMerge` 4 (**all 4 comment-only**, verified individually).

**DELETE — reachable only from the dead program or from tests:**

| Declaration | confirm.go line | Sole reachability |
|-------------|----------------:|-------------------|
| `func ConfirmMerge` | 915 | one test call site |
| `type confirmModel` + `Init`/`Update`/`View`/`syncSelectedFiles`/`filterSelectedFiles` | 65, 76, 101, 196, 208, 249 | entered only via `ConfirmMerge` |
| `type fileListItem` + `FilterValue`/`Description`/`Title` | 44, 50, 55, 60 | `confirmModel` list adapter |
| `type AnalysisFormatter` + `formatterStyles` + all `Format*`/`render*`/`format*` methods | 270, 279, 353-760 | constructed at `confirm.go:261` inside `confirmModel.View()`, and in tests |
| `func NewAnalysisFormatter` | 298 | **tests only** |
| `func NewAnalysisFormatterWithSelection` | 309 | `confirm.go:261` (inside `confirmModel.View()`) + tests |
| `func initFormatterStyles` | 319 | formatter constructors |
| `truncateRowField`, `groupFilesByRisk`, `groupByPathPrefix`, `type pathGroup`, `topTwoSegments`, `riskIcon` | 765, 777, 800, 793, 824, 836 | formatter methods only |
| `func validateAnalysis` | 850 | `ConfirmMerge:917` + tests |
| `func sanitizePath` | 884 | `ConfirmMerge:923` + tests |
| `const maxMergeFiles`, `maxPathLength` | 21, 22 | `validateAnalysis` only |

**Security note (finding, not a scope item).** `validateAnalysis` (file-count and path-length DoS limits) and `sanitizePath` (path-traversal sanitization) are reachable only from the dead `ConfirmMerge`. They are therefore **already not running in production** — deleting them removes no live control and is not a regression. But it does surface that the live path (`confirmViaPreview` → `toPreviewInputs`) applies neither. That gap is pre-existing and **out of scope** here; M2 must record it so it is not silently lost.

**Relocation decision (D-1).** After deletion, `confirm.go` would hold only `MergeAnalysis` + `FileAnalysis` — roughly 20 lines under a filename that no longer describes its contents. `internal/merge/types.go` already exists and is the package's type home (`MergeStrategy`, `MergeResult`, `Conflict`, `Engine`, sentinel errors). **Recommendation:** move the two structs into `types.go` and delete `confirm.go` outright. This changes no import path and no exported name — `merge.MergeAnalysis` / `merge.FileAnalysis` resolve identically. Alternative (leave a 20-line `confirm.go`) is acceptable but leaves a misleading filename.

### §B.5 Windows non-TTY guard — already relocated (contradicts the brief's concern)

The brief warned that deleting `confirm.go` might drop the REQ-CFS-007/008 Windows guard. It does not.

| Site | Guard condition | Error string |
|------|-----------------|--------------|
| `internal/merge/confirm.go:940` (dead) | `runtime.GOOS == "windows" && !isatty.IsTerminal(os.Stdin.Fd())` | "merge confirmation UI requires an interactive terminal; rerun with --yes to auto-confirm in a non-TTY environment" |
| `internal/cli/update.go:1145` (**live**) | `!isatty.IsTerminal(os.Stdin.Fd())` — **all platforms** | identical string |

The live guard is strictly **broader**. `grep -rn "REQ-CFS-00[78]" --include="*.go" .` returns exactly two hits: the dead guard and the `update.go:1137` comment that documents the relocation. M2 drops nothing; an AC asserts the live guard survives.

### §B.6 Hex-literal baseline — 4 hits, not 1 (contradicts the brief)

Command: `grep -rnE '"#[0-9a-fA-F]{6}"|#[0-9a-fA-F]{6}' --include="*.go" internal/ cmd/ pkg/ | grep -v "^internal/tui/" | grep -v "_test\.go:"` → **4 lines**.

| Site | Kind |
|------|------|
| `internal/merge/confirm.go:454` | **real code** — `lipgloss.AdaptiveColor{Light: "#111827", Dark: "#E5E7EB"}` |
| `internal/merge/confirm.go:455` | **real code** — `lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#4B5563"}` |
| `internal/cli/wizard/styles.go:17` | comment only |
| `internal/cli/wizard/styles.go:20` | comment only |

The brief named only `styles.go:20`. In fact `internal/merge/confirm.go` carries **four raw hex colour values across two real code lines** — a genuine D1 / AC-CLI-TUI-013 violation living inside the dead code M2 deletes. M2 therefore takes the production count from 2 violating lines to **0**, leaving only the 2 comment-only wizard hits. This is a measurable, non-vacuous AC.

`internal/merge/confirm.go` also **does** import `internal/tui` (line 14), but uses it at only two lines (320-321, `LightTheme()`/`DarkTheme()`), against 41 direct `lipgloss.` usages. It is a hybrid, not a purely parallel system.

### §B.7 Coverage baseline

Command: `go test -cover ./internal/cli/ ./internal/cli/update/ ./internal/merge/ ./internal/cli/wizard/ ./internal/tui/` (exit 0)

| Package | Coverage |
|---------|---------:|
| `internal/cli` | **74.9%** |
| `internal/cli/update` | **70.2%** |
| `internal/merge` | **86.3%** |
| `internal/cli/wizard` | **95.2%** |
| `internal/tui` | **93.6%** |

`internal/cli` sits at 74.9%, **below** the 90% critical-package target the brief cited. An AC demanding "maintain 90%" would be unachievable at entry. AC-TUIM-032 therefore binds **no regression from these measured baselines** instead.

M2 caveat: deleting ~900 lines of well-covered code plus its two test files will move `internal/merge`'s ratio in a direction that is **unmeasured** at plan time. The AC accounts for this (see acceptance.md AC-TUIM-032).

### §B.8 Cross-platform baseline

Command: `GOOS=windows GOARCH=amd64 go build ./...` → **exit 0**. Clean at entry, so any post-change failure is attributable.

### §B.9 The CI blind spot, precisely

Golden fixtures: `ls internal/cli/testdata/tuxiu/` → 6 `.golden` pairs (`init`/`update` × `tty`/`notty`/`nocolor`), plus a `postm4/` subdirectory. Consumed by 4 test functions in `internal/cli/tuxiu_characterization_test.go`. None invokes the interactive program.

`internal/cli/update/preview_test.go` carries **12** test functions that **do** drive the interactive model headlessly (`newPreviewModel` → `tableView()` / `diffView()` / `selectRow()` / `backToTable()`). Reading their bodies: every assertion is `strings.Contains` over class labels and file paths — **content only**. Zero presentation assertions.

So the blind spot is narrower and sharper than "the interactive path is untested": it is **"the interactive path has no presentation regression coverage."** M1's new coverage must assert presentation, not re-assert content.

### §B.10 Library API reality checks (two brief claims corrected)

**`table.WithStyles` exists** — `charm.land/bubbles/v2@v2.1.1/table/table.go`: `type Styles` (106), `func DefaultStyles()` (113), `func (m *Model) SetStyles(s Styles)` (122), `func WithStyles(s Styles) Option` (188). REQ-TUIM-011 is implementable as written. (Version attribution corrected at v0.1.2: `go.mod:6` pins **v2.1.1**, not v2.0.0. Re-verified against the pinned version — the four line numbers are byte-identical and the conclusion is unchanged.)

### §B.10b lipgloss v1 vs v2 — the rendering difference the ACs turn on

`go.mod` carries **both** lipgloss majors, and which one a package imports changes what its output looks like:

| Package | Import | Major |
|---------|--------|-------|
| `internal/tui` | `charm.land/lipgloss/v2` | **v2** |
| `charm.land/bubbles/v2/table` | `charm.land/lipgloss/v2` | **v2** |
| `internal/cli/uikit` | `github.com/charmbracelet/lipgloss` | v1 |

**v2 renders colour as decimal RGB SGR, never as the hex string.** Measured directly against the pinned `charm.land/lipgloss/v2@v2.0.5`:

```
token=#bf6547  rendered="\x1b[38;2;191;101;71mPROBE\x1b[m"  containsHexToken=false
token=#d97757  rendered="\x1b[38;2;217;119;87mPROBE\x1b[m"  containsHexToken=false
token=#c4432b  rendered="\x1b[38;2;196;67;43mPROBE\x1b[m"   containsHexToken=false
```

This matches lipgloss v2's own test fixture (`style_test.go:107-110`: `Color("#5A56E0")` → `"\x1b[38;2;90;86;224mhello\x1b[m"`).

**Consequence:** an assertion of the form `strings.Contains(rendered, th.Accent)` is **false for a correct implementation** — the hex token is consumed by the colour parser and never reaches the output. Token-application tests must compare against a *rendered probe* (`lipgloss.NewStyle().Foreground(lipgloss.Color(tok)).Render("x")`), which stays hex-literal-free and so does not violate constraint D1. See AC-TUIM-004 / -008 / -030b.

**Do not "fix" AC-TUIM-003.** Headless colour-stripping (`Render` detecting a non-TTY and dropping SGR) is a **v1** behaviour. `internal/tui` and `bubbles/v2/table` are v2, which performs no in-`Render` TTY detection, so forcing the axis and diffing light-vs-dark output works headlessly as written.

**`tea.WithAltScreen()` does NOT exist in bubbletea v2.** `grep "^func With" charm.land/bubbletea/v2@v2.0.8/options.go` lists 12 options; `WithAltScreen` is not among them. In v2 the alt-screen is a **field on the returned view**: `tea.View` (`tea.go:84`) carries an `AltScreen` field consumed by the renderer (`cursed_renderer.go:87, 320, 519`). The mechanism is therefore *set the field inside `View()`*, not *pass a program option*. The brief's proposed call is not compilable; §F M1-4 uses the real mechanism.

### §B.10a Mid-run `AltScreen` toggle — verified supported (D-4 precondition)

Decision D-4 rests on `previewModel.View()` returning `AltScreen: false` in the table sub-view and `AltScreen: true` in the diff sub-view — a **mid-program** transition. Verified against `charm.land/bubbletea/v2@v2.0.8/cursed_renderer.go` before recording the decision as settled.

**Supported, and first-class.** The renderer computes the transition per frame by comparing the previous view against the current one (`cursed_renderer.go:320`):

```go
shouldUpdateAltScreen := (s.lastView == nil && view.AltScreen) || (s.lastView != nil && s.lastView.AltScreen != view.AltScreen)
```

On transition it emits the DEC private mode 1049 pair (`cursed_renderer.go:513-525`): `ansi.SetModeAltScreenSaveCursor` entering, `ansi.ResetModeAltScreenSaveCursor` leaving. `enterAltScreen` (line 653) does `SaveCursor` → set mode → `SetFullscreen(true)` → `Erase`; `exitAltScreen` (line 663) does `Erase` → `SetRelativeCursor(true)` → `SetFullscreen(false)` → reset mode → `RestoreCursor`.

**Consequence for D-4's requirement (c).** Mode 1049 is the save-cursor-and-switch-buffer pair: the main screen buffer is untouched while the alt screen is active, and the cursor is restored on exit. "`esc` restores the inline table with prior scrollback intact" is therefore a **guarantee of the terminal mode**, not something the application must implement. No fallback is needed and none is recorded.

**Caveat 1 — keyboard-protocol renegotiation per transition.** `cursed_renderer.go:378-380` re-negotiates the Kitty keyboard protocol whenever `view.AltScreen != s.lastView.AltScreen`, and lines 513-517 disable keyboard enhancements across the switch. The library handles this, but each enter/exit costs a renegotiation round — rapid toggling (holding `enter`/`esc`) produces protocol churn. Not a blocker; noted so it is not diagnosed as a defect later.

**Caveat 2 — the quit path must resolve to the inline view (load-bearing).** The close path (`cursed_renderer.go:167-171`) branches on the **last** view:

```go
if lv.AltScreen {
    enableAltScreen(s, false, true)   // exit alt screen — its content is discarded
} else {
    _, _ = s.scr.WriteString(ansi.EraseScreenBelow)   // inline frame persists in scrollback
}
```

`Update` currently handles `y` / `q` / `ctrl+c` **before** the sub-view check (`preview_tui.go:204-214`), so a quit is reachable from the diff sub-view today. Under D-4 that would make the final frame an alt-screen frame, which the renderer discards on exit — the result summary would vanish and D-4's requirement (d) would silently fail. **The model must set `view = previewTableView` before returning `tea.Quit`.** This is REQ-TUIM-022a and AC-TUIM-014d. It is not obvious from the API and would not be caught by a compile or by any content-level test.

**Consequence for D-4's requirement (d).** Because the inline branch writes `EraseScreenBelow` after moving to the bottom of the frame, the *final inline frame itself* persists in scrollback and everything below it is cleared. So "exactly one result line remains" is achieved by having the final `View()` return only the one-line summary — the renderer repaints in place and the frame shrinks. Requirement (d) is a property of the model's terminal view, not of a separate teardown step.

### §B.11 Theme resolution is not reachable from package `update`

`grep -rn "func resolveTheme"` → `internal/cli/theme.go:11`, package `cli`. The preview lives in package `update` and cannot call it. M1 must resolve independently via the `internal/tui` OS entry points (`IsDarkOS` / `ResolveOS`), or accept an injected theme through `PreviewOptions`. See decision D-2.

### §B.12 The parallel SPEC's actual edit surface

`grep -n "wizard\.go" .moai/specs/SPEC-CLI-WIZARD-RESTRUCTURE-001/plan.md` — declared change sites:

| Ref | Site | Line |
|-----|------|-----:|
| C25 | `RunWithDefaultsModes` signature + `WizardResult` mode fields | ~58-66 |
| — | `buildFormGroups()` group merging | ~162 |
| C18 | `saveAnswer` "harness_profile" case | ~418-419 |
| C17 | `saveBoolAnswer` "advanced_bridge" case | ~427-430 |
| C19 | `saveBoolAnswer` "coverage_exemptions_enabled" case | ~435-436 |

`grep -rn "moaiWizardStyles\|newMoAIWizardTheme\|wizardTokens" .moai/specs/SPEC-CLI-WIZARD-RESTRUCTURE-001/*.md` → **no matches**. The parallel SPEC never touches the theme functions.

Note the brief's stated range ("roughly 147-450") is **too narrow at the low end** — the parallel SPEC also edits ~58-66. The containment boundary must therefore be expressed as a **lower bound on this SPEC's edits**, not as an avoid-range.

**Measured theme-region boundary in `wizard.go`:**

| Line | Declaration |
|-----:|-------------|
| 445 | `func buildConfirmField` — last assembly-region function |
| 479 | last line of `buildConfirmField` body (`}`) |
| 480-482 | doc comment for `wizardIsDark` |
| 483 | `var wizardIsDark = tui.IsDarkOS` |
| 498 | `func newMoAIWizardTheme` |
| 506 | `func moaiWizardStyles` |
| 552 | `type wizardTokenSet` |
| 569 | `func wizardTokens` |
| 589 | EOF |

**Containment rule (D-3):** every hunk this SPEC produces in `internal/cli/wizard/wizard.go` starts at line **≥ 480**. Verifiable via `git diff` hunk-header arithmetic (see acceptance.md AC-TUIM-025).

---

## §C Pre-flight

Before the first edit:

1. `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` — surface any parallel-session divergence.
2. Confirm `SPEC-CLI-WIZARD-RESTRUCTURE-001` run state. If it has landed, re-measure the `wizard.go` theme-region boundary in §B.12 — the line numbers will have shifted, and D-3's `≥ 480` threshold must be re-derived from content anchors (`var wizardIsDark`), not carried forward.
3. Re-run the §B.7 coverage baseline and the §B.6 hex baseline. Both are the denominators of ACs; a stale denominator makes the AC vacuous.
4. `GOOS=windows GOARCH=amd64 go build ./...` — confirm the §B.8 clean entry still holds.
5. Confirm no working-tree changes under `internal/cli/update/`, `internal/merge/`, `internal/cli/wizard/` from a parallel session.

**Commit hygiene.** This repo has live parallel sessions sharing the checkout (see the current branch state). Every commit in this SPEC uses an explicit pathspec — never `git add -A`.

---

## §D Constraints

Carried from `spec.md` §D (D1-D9). Two amplified here:

- **D7 (fallback ANSI-free)** is a *structural* invariant, not a conditional branch. Verified with `grep -c $'\x1b' internal/cli/update/preview_fallback.go` → **0** at baseline. (Corrected at v0.1.2: the previously-recorded `grep -c $'\x1b\['` **errors** in this repo's zsh — the `$'...'` form collapses `\[` to a bare `[`, and `/usr/bin/grep` returns `grep: brackets ([ ]) not balanced`, exit 2. A command that errors cannot substantiate a "→ 0" claim under D9. See acceptance.md §C.1 CMD-ANSI.) The M1 restructure improves *layout* (alignment, grouping, card shape) using plain ASCII/Unicode box characters and whitespace only. Introducing `lipgloss` into this file is prohibited even under a `noColor` guard, because the guarantee's value is that it needs no guard.
- **D8 (wizard.go containment)** — see §B.12 / D-3. This is a merge-safety constraint with a parallel session, not a stylistic preference.

---

## §E Self-verification

Run before declaring any milestone complete. All read-only; issue as a single-turn parallel batch.

```bash
go build ./...
go vet ./...
go test ./internal/cli/... ./internal/merge/... ./internal/tui/...
go test -cover ./internal/cli/ ./internal/cli/update/ ./internal/merge/ ./internal/cli/wizard/ ./internal/tui/
GOOS=windows GOARCH=amd64 go build ./...
grep -rnE '"#[0-9a-fA-F]{6}"|#[0-9a-fA-F]{6}' --include="*.go" internal/ cmd/ pkg/ | grep -v "^internal/tui/" | grep -v "_test\.go:"
grep -cE "'✓'|'✗'|'●'|'○'" internal/cli/update/preview_tui.go
grep -c $'\x1b' internal/cli/update/preview_fallback.go
golangci-lint run --timeout=3m
git diff --stat -- internal/template/templates/    # MUST be empty (REQ-TUIM-056)
```

---

## §F Milestones

### M1 — change-preview interactive TUI redesign

*The user-visible fix. Landable alone (REQ-TUIM-054).*

Ordered within the milestone by decision-reversibility — the two design decisions first, mechanical wiring after.

#### M1-1 — DECISION D-2: how the theme reaches package `update`

Package `update` cannot call `cli.resolveTheme` (§B.11). Two options:

| Option | Shape | Trade-off |
|--------|-------|-----------|
| **A (recommended)** | `preview_tui.go` calls `tui.ResolveOS()` directly at model construction | No signature change; no caller churn; matches how `huh_theme.go` and `wizard.go` each resolve independently |
| B | Add a `Theme *tui.Theme` field to `PreviewOptions`; caller injects | Testable without env manipulation, but changes a public-ish struct and forces the `update.go` call site to change |

Option A plus a package-level indirection var (`previewIsDark = tui.IsDarkOS`, mirroring `huhThemeIsDark` at `huh_theme.go:15` and `wizardIsDark` at `wizard.go:483`) gives testability without a signature change. Recommend A.

#### M1-2 — DECISION D-4 (RESOLVED — HYBRID): scrollback-residue policy

`runPreviewTUI` (`preview_tui.go:251-258`) calls `tea.NewProgram(model)` with no view configuration, so the final frame is left in scrollback and the user sees the run rendered twice. That duplicate render is the symptom the user reported.

**Resolution — hybrid, per sub-view:**

| Sub-view | Screen | Rationale |
|----------|--------|-----------|
| `previewTableView` | **inline** (`AltScreen: false`) | The identity band and classification card stay visible; the user can still see what they are confirming against. |
| `previewDiffView` | **alternate screen** (`AltScreen: true`) | A full-height diff wants the whole viewport, and leaves no scrollback residue of its own. |
| exit | inline, single-line summary | The interactive frame is replaced by one result line naming the outcome and file count. The classification content is not re-rendered. |

The mechanism is the per-view `AltScreen` field (§B.10), which makes this a natural per-sub-view toggle rather than a program-wide mode:

```go
func (m *previewModel) View() tea.View {
    if m.view == previewDiffView {
        v := tea.NewView(m.diffView())
        v.AltScreen = true      // full-viewport diff
        return v
    }
    v := tea.NewView(m.tableView())
    v.AltScreen = false         // inline — preceding output stays visible
    return v
}
```

**Renderer verification (performed, §B.10a):** the mid-run toggle is first-class — `cursed_renderer.go:320` compares `s.lastView.AltScreen != view.AltScreen` per frame and emits the DEC mode 1049 pair. `esc` restoring prior scrollback is a guarantee of mode 1049 itself (save-cursor + separate buffer), not application work. **No fallback is required and none is recorded.**

Two constraints fall out of that verification and are binding on M1-4:

1. **Quit must resolve to the inline view first** (§B.10a caveat 2). The renderer's close path discards the frame when the last view is an alt-screen view. `Update` handles `y` / `q` / `ctrl+c` in the pre-check block (`preview_tui.go:205-214`), so those three are reachable **from the diff sub-view** and would silently discard the exit summary. Set `view = previewTableView` before returning `tea.Quit`. → REQ-TUIM-022a, AC-TUIM-014d.

   **`n` is not diff-reachable.** The diff branch returns at `preview_tui.go:221-223`, so `n` never reaches its quit case (`:230`) from the diff view — it falls through to `viewport.Update`. The fix still applies to `n` for symmetry on the table path, but any AC demanding `n`-from-diff would fail a correct implementation; `n` is covered separately by AC-TUIM-014f. Do **not** make `n` diff-reachable — that is an unspecified behaviour change contradicting the documented keymap at `preview_tui.go:199-200`.
2. **The exit summary is the final `View()`**, not a teardown print. The inline close path writes `EraseScreenBelow` after the last frame, so the last frame persists and everything below is cleared. Returning a one-line view before quitting is what leaves exactly one line.

Keyboard-protocol renegotiation on each toggle (§B.10a caveat 1) is library-handled; noted so it is not later diagnosed as a defect.

#### M1-3 — DECISION D-5 (RESOLVED — SPLIT): presentation-regression mechanism

Two complementary mechanisms, not one. The model already exposes headless accessors (`tableView`, `diffView`, `currentView`, `selectRow`, `backToTable`), so neither needs a real TTY.

| # | Mechanism | Pins | Deliberately does NOT pin |
|---|-----------|------|---------------------------|
| **1** | **Structure golden under `NO_COLOR`** — render over a fixed fixture with the monochrome axis forced, compare against `testdata/*.golden` | Layout, borders, column alignment, row order | Colour. A palette edit in `internal/tui` must not break `internal/cli/update` tests. |
| **2** | **Token-application unit tests** — assert each semantic role resolves to the correct `internal/tui` token, compared against values read from the resolved `Theme` (never hard-coded hex) | Conflict label → `th.Danger`; table border → `th.ChromeBorder`; selected row → `th.Accent`; and the remaining role map from REQ-TUIM-014 | Whole-frame shape |

**The self-trip burden sits on mechanism 2, explicitly.** Mechanism 1 is palette-insensitive *by design* — which means removing the theme wiring would leave its golden unchanged and passing. It provably cannot detect the defect class this SPEC exists to repair. The split exists precisely so that mechanism 2 carries that detection, and AC-TUIM-030c binds the self-trip to mechanism 2's tests by name. Stating this here prevents a future reader from concluding "we have a golden, we are covered."

Reading tokens from the resolved `Theme` rather than hard-coding hex also keeps mechanism 2 inside constraint D1 — a test asserting `"#cc785c"` would itself be a hex literal outside `internal/tui/`.

#### M1-4 — table styling + theme wiring

- Resolve theme per D-2. Build a `table.Styles` from the resolved `tui.Theme` and pass via `table.WithStyles` (`bubbles/v2/table/table.go:188`).
- Apply the semantic class-label roles of REQ-TUIM-014: add→`Success`, update→`Info`, preserve→`Dim`, conflict→`Danger`. Label **strings** stay byte-identical to `ChangeClass.String()` — the existing `preview_test.go` content assertions must keep passing unmodified.
- Apply the D-4 hybrid residue decision: per-sub-view `AltScreen` in `View()`; force `view = previewTableView` before every `tea.Quit` return in `Update` — the pre-check block (`y`, `q`, `ctrl+c`, which are diff-reachable) and the table-only `n` case; return a one-line result summary as the final view.

#### M1-5 — structural primitives

- Replace `buildClassCountHeader` (`preview_tui.go:118-129`) with a `tui.Box` / `tui.Section` card.
- Replace the inline hint literal at `preview_tui.go:143` (`"[enter] view diff  [y] confirm  [n/q] cancel"`) and at `buildDiffViewContent:189` (`"[esc] back to table"`) with `tui.HelpBar([]tui.KeyHint{...})`.
- Any status glyph resolves from `tui.Glyph*` (`internal/tui/glyphs.go`).

#### M1-6 — fallback restructure

- Restructure `renderFallback` into aligned columns and grouped sections with a card-shaped outline.
- **Plain text only.** No `lipgloss` import; no `\x1b[`. The structural guarantee documented at `preview_fallback.go:10-15` survives verbatim.
- Existing fallback tests (`TestPreviewFallbackZeroANSIUnderNoColor`, `TestPreviewFallbackZeroANSIWhenPiped`) must pass unmodified.

#### M1-7 — presentation regression coverage

Implement both D-5 mechanisms.

- **Mechanism 1** — structure goldens under a forced monochrome axis for the table view, the diff view, and the fallback.
- **Mechanism 2** — token-application unit tests for the role map (conflict→`th.Danger`, border→`th.ChromeBorder`, selected row→`th.Accent`, plus the REQ-TUIM-014 class roles), reading token values from the resolved `Theme`.
- Assert the D-4 properties: table view `AltScreen == false`, diff view `AltScreen == true`, `backToTable()` returns to the inline view, and every quit path leaves `currentView() == previewTableView` with a one-line final view.
- Perform the AC-TUIM-030c self-trip against mechanism 2 and record the failing test name.

---

### M2 — dead legacy confirmation TUI removal

*Independent of M1. Destructive — the inventory gate is mandatory.*

#### M2-1 — re-verify reachability before deleting anything

Do **not** trust §B.4. Re-run at execution time:

```bash
grep -rn "ConfirmMerge" --include="*.go" .
grep -rn "merge\.[A-Z][A-Za-z]*" --include="*.go" internal/ cmd/ pkg/ | grep -v "_test\.go:" | grep -v "^internal/merge/"
grep -rn "reflect\." --include="*.go" internal/merge/
grep -rn "//go:build" internal/merge/
```

Additionally confirm no build-tag-gated caller and no reflection-based dispatch. If any symbol's status differs from §B.4, **halt and re-report** rather than proceeding on the plan-time snapshot.

#### M2-2 — surgical deletion

- Delete per the §B.4 DELETE column.
- Preserve `MergeAnalysis` + `FileAnalysis` per the KEEP column, field sets unchanged.
- Apply relocation decision D-1 (recommend: move both structs to `internal/merge/types.go`, delete `confirm.go`).
- Record the §B.4 security note (dead `validateAnalysis` / `sanitizePath`) in the commit body so the pre-existing live-path gap is not lost.

#### M2-3 — test reconciliation

Both `internal/merge/confirm_test.go` (971 lines) and `internal/merge/confirm_coverage_test.go` (15,868 bytes) construct deleted symbols. Delete or retarget both. Any assertion that covers a **surviving** type's behaviour is retargeted, not discarded.

#### M2-4 — dependency + hex sweep

- Confirm `internal/merge` no longer imports `charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, or `github.com/charmbracelet/lipgloss`.
- Confirm the §B.6 hex count drops from 2 violating code lines to 0.
- `go.mod` unchanged — the libraries remain in use elsewhere (REQ-TUIM-055).

---

### M3 — input-form theme unification

*Depends on M1's token decisions. Smallest surface; most merge-sensitive.*

#### M3-1 — containment gate (run FIRST)

Re-derive the theme-region boundary from a content anchor, not the §B.12 line numbers:

```bash
grep -n "var wizardIsDark" internal/cli/wizard/wizard.go
```

Every hunk this SPEC produces in `wizard.go` must start at or after that line. Enforced by acceptance.md AC-TUIM-025.

#### M3-2 — divergence audit (REQ-TUIM-042)

Enumerate, per factory, every style field the sibling sets that it does not. Baseline observation, to be re-derived at execution:

- `moaiHuhStyles` (v1, `huh_theme.go:52-75`) sets: `Title`, `NoteTitle`, `Description`, `ErrorIndicator`, `ErrorMessage`, `SelectSelector`, `NextIndicator`, `PrevIndicator`, `Option`, `MultiSelectSelector`, `SelectedOption`, `UnselectedOption`, `TextInput.Cursor/Placeholder/Prompt`, `FocusedButton`, `BlurredButton`.
- `moaiWizardStyles` (v2, `wizard.go:506-550`) sets all of the above **plus**: `Base` (`BorderForeground`), `Card`, `SelectedPrefix` (`"◆ "`), `UnselectedPrefix` (`"◇ "`), `SelectSelector` string (`"▸ "`), `Next`, and a full `Blurred` derivation with `HiddenBorder()`.

Apparent v1 gaps: border/base styling, the selected/unselected prefixes, the selector string, and the `Blurred` border treatment. Each must be **closed** or **recorded with a reason** (some may be genuinely absent from the huh v1 API — verify against the v1 `FieldStyles` type before concluding a gap is a defect).

#### M3-3 — token alignment

Reconcile both factories to the token set M1 settles on. Keep the two factories separate (REQ-TUIM-040) — the v1/v2 `Theme` type boundary documented at `huh_theme.go:21-30` is real. Do not relocate the wizard theme functions (REQ-TUIM-044): moving ~110 lines would *enlarge* the conflict surface with the parallel SPEC, not shrink it.

Preserve both indirection vars (`huhThemeIsDark`, `wizardIsDark`) per REQ-TUIM-045.

---

## §G Anti-patterns

| # | Anti-pattern | Why it bites |
|---|--------------|--------------|
| G1 | Editing `wizard.go` outside the theme region | Guarantees a merge conflict with `SPEC-CLI-WIZARD-RESTRUCTURE-001` (run pending in a parallel session). D-3 / AC-TUIM-025. |
| G2 | Relocating the wizard theme functions to `styles.go` | Moving ~110 lines makes the diff larger and the conflict *more* likely, not less. REQ-TUIM-044. |
| G3 | Merging the v1 and v2 huh factories | The `Theme` types differ across the library version boundary; the split is documented and deliberate. REQ-TUIM-040. |
| G4 | Introducing `lipgloss` into `preview_fallback.go` "guarded by `noColor`" | Destroys the *structural* ANSI-free guarantee. The guarantee's value is that it needs no guard. D7. |
| G5 | Calling `tea.WithAltScreen()` | Does not exist in bubbletea v2. Use the per-view `View.AltScreen` field. §B.10. |
| G5a | Returning `tea.Quit` while the model is still in the diff sub-view | The renderer discards the last frame when it is an alt-screen frame (`cursed_renderer.go:167-171`), so the exit summary vanishes. Compiles fine; no content-level test catches it. Force `view = previewTableView` first. §B.10a caveat 2. |
| G5e | Making `n` reachable from the diff sub-view "for consistency" | `n` is table-only by design (`preview_tui.go:199-200` keymap, enforced by the diff branch returning at `:221-223`). Adding it is an unspecified behaviour change. AC-TUIM-014d deliberately excludes `n`; AC-TUIM-014f covers it on the table path. |
| G5f | Asserting a raw hex token against rendered output (`strings.Contains(out, th.Accent)`) | lipgloss v2 emits decimal RGB SGR — the hex string never appears, so the assertion is false for a *correct* implementation. Compare against a rendered probe instead. §B.10b. |
| G5h | Asserting a full-CSI prefix (`"\x1b[38;2;191;101;71m"`) against rendered output | v2 merges all SGR parameters into one CSI, so a bold+coloured cell renders `"\x1b[1;38;2;...m"` and the prefix is not a substring. Bubbles styles the selected row bold by default, so this fails on the very element the SPEC prescribes. Assert the parameter run instead — acceptance.md §C.2. |
| G5g | Using a markdown-escaped `\|` inside a `grep -E` pattern | In an ERE that is a *literal pipe*, not alternation, so the command matches nothing and reports `0` unconditionally — including after a regression. Every such command lives in acceptance.md §C.1 as a fenced block with a positive control. |
| G5b | Making the whole program alt-screen "for consistency" | Hides the identity band and classification card the user is being asked to confirm against — the exact context D-4 preserves by keeping the table inline. |
| G5c | Treating the structure golden as sufficient regression coverage | Mechanism 1 is palette-insensitive by design and would stay green with the theme wiring removed. Mechanism 2 carries the detection. D-5. |
| G5d | Hard-coding a hex value in a token-application test | That is itself a hex literal outside `internal/tui/`, violating D1. Read the value from the resolved `Theme`. |
| G6 | Deleting `MergeAnalysis` / `FileAnalysis` with the dead program | Both have production consumers in three files outside `internal/merge`. §B.4. |
| G7 | Deleting only `confirm_test.go` | A third file, `confirm_coverage_test.go`, also constructs the deleted symbols. §B.1. |
| G8 | Asserting "coverage ≥ 90%" for `internal/cli` | Measured baseline is 74.9%. The AC would fail at entry through no fault of the change. §B.7. |
| G9 | Content-only assertions as the new regression coverage | 12 such tests already exist and did not catch this defect class. New coverage must assert **presentation**. §B.9. |
| G10 | Changing a `ChangeClass.String()` label to "improve" it | Breaks the existing content assertions and the deploy-stage coherence contract. REQ-TUIM-014. |
| G11 | Re-deriving classification inside the redesigned view code | A parallel heuristic is exactly what `preview.go`'s single-source design prevents. REQ-TUIM-020. |
| G12 | `git add -A` | Parallel sessions share this checkout. Always commit with an explicit pathspec. §C. |
| G13 | Carrying §B line numbers into execution unverified | The parallel SPEC will shift `wizard.go`; anchor on content tokens instead. §C step 2, M3-1. |

---

## §H Cross-references

- `spec.md` §A.2 — the binding TUX / TUI / huh vocabulary
- `acceptance.md` — 30 acceptance criteria with the measured baselines these milestones are judged against
- `.moai/specs/SPEC-CLI-TUX-INIT-UPDATE-001/spec.md` — the carve-out (§C) and the glyph-SSOT exemption (REQ-TUXIU-001)
- `.moai/specs/SPEC-CLI-WIZARD-RESTRUCTURE-001/plan.md` — the parallel SPEC's declared `wizard.go` edit sites (§B.12)
- `internal/cli/CLAUDE.md`, `CLAUDE.local.md` §6 / §14
- `charm.land/bubbles/v2@v2.1.1/table/table.go` — `Styles` / `WithStyles` / `SetStyles`
- `charm.land/bubbletea/v2@v2.0.8/tea.go:84` — `View.AltScreen`
- `charm.land/bubbletea/v2@v2.0.8/cursed_renderer.go:320, 167-171, 653-672` — mid-run toggle computation, close-path branch, enter/exit sequences

---

## §I Resolved decisions

This SPEC carries **zero** open clarification markers. Both prior open questions were resolved by the user and folded into §F. (This section deliberately avoids restating the marker token itself, so an auditor grepping for it does not match this artifact.)

| # | Decision | Resolution | Landed in |
|---|----------|------------|-----------|
| D-4 | Preview scrollback-residue policy | **HYBRID** — table sub-view inline, diff sub-view on the alternate screen, `esc` restores the inline table with prior scrollback, single-line result summary on exit. Renderer support verified (§B.10a); no fallback needed. | REQ-TUIM-019 / -019a / -022 / -022a; M1-2, M1-4; AC-TUIM-014a-d |
| D-5 | Presentation-regression mechanism | **SPLIT** — a `NO_COLOR` structure golden (palette-insensitive) plus token-application unit tests reading values from the resolved `Theme`. The self-trip burden is carried by the token tests, since the golden provably cannot detect the defect class. | REQ-TUIM-050 / -050a / -050b / -051; M1-3, M1-7; AC-TUIM-030a-c |

Vocabulary (TUX / TUI / huh form runtime) was adopted as authored — `spec.md` §A.2 is unchanged.
