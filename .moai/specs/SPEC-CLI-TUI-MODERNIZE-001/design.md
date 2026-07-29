---
id: SPEC-CLI-TUI-MODERNIZE-001
title: "Design — interactive TUI surface modernization"
version: "0.1.3"
status: draft
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

# Design — SPEC-CLI-TUI-MODERNIZE-001

Scope: the **interactive** surface. The output-rendering (TUX) surface was designed and shipped by `SPEC-CLI-TUX-INIT-UPDATE-001`; this document designs the interactive layer to match it, and is deliberately silent on anything that SPEC already settled.

---

## §A Layer model

Three layers, per the vocabulary bound in `spec.md` §A.2:

```
┌─ TUX (output-rendering) ────────────────────────────────┐
│ internal/tui  ·  internal/cli/uikit                     │
│ pure string producers; no state, no key loop            │
│ Theme (28 tokens) + Box/Section/Pill/KV/HelpBar/Glyph*  │
└───────────────┬─────────────────────────────────────────┘
                │ consumed by
   ┌────────────┴────────────┬──────────────────────────┐
   ▼                         ▼                          ▼
┌─ TUI ───────────┐  ┌─ huh form runtime ─┐  ┌─ TUX call sites ─┐
│ preview_tui.go  │  │ huh_theme.go  (v1) │  │ update_tux.go    │
│ Bubble Tea M/U/V│  │ wizard.go tail(v2) │  │ init/update out  │
│ ← M1            │  │ ← M3               │  │ (already shipped)│
└─────────────────┘  └────────────────────┘  └──────────────────┘

RETIRED by M2: internal/merge/confirm.go — a second, parallel TUI
that carried its own lipgloss styling instead of consuming TUX.
```

The design intent is one-directional: **TUX produces tokens and primitives; every interactive surface consumes them.** M2 exists because `confirm.go` violated that direction — it re-derived colour decisions locally (41 direct `lipgloss.` usages, 4 raw hex values) rather than consuming the token source.

---

## §B Hybrid inline / alt-screen state model (D-4)

### §B.1 State machine

```
                    ┌──────────────────────────────────┐
                    │  previewTableView                │
   program start ──▶│  AltScreen: false  (INLINE)      │
                    │  renders below the identity band │
                    │  and classification card         │
                    └──┬────────────────────────┬──────┘
                       │ enter                  │ y / n / q / ctrl+c
                       ▼                        ▼
      ┌────────────────────────────┐   ┌──────────────────────┐
      │ previewDiffView            │   │ exit frame           │
      │ AltScreen: true (ALT)      │   │ AltScreen: false     │
      │ full-viewport diff         │   │ one-line summary     │
      └──┬──────────────────┬──────┘   └──────────────────────┘
         │ esc              │ y / q / ctrl+c
         │                  │  (MUST first set view = previewTableView)
         └──────────────────┴──▶ back to table / exit frame
```

**Why hybrid rather than all-inline or all-alt.** All-inline leaves the diff competing with prior scrollback for vertical space and re-renders the whole run. All-alt hides the identity band and classification card *while the user is being asked to confirm against them* — it removes the very context the confirmation depends on. The split gives each sub-view what it needs: the table needs context, the diff needs room.

### §B.2 Mechanism

Alt-screen in bubbletea v2 is a **field on the returned view**, not a program option — so the toggle is naturally per-sub-view:

```go
func (m *previewModel) View() tea.View {
    if m.view == previewDiffView {
        v := tea.NewView(m.diffView())
        v.AltScreen = true
        return v
    }
    v := tea.NewView(m.tableView())
    v.AltScreen = false
    return v
}
```

The renderer diffs `lastView.AltScreen` against the current view each frame and emits the DEC private mode 1049 pair on transition. See `research.md` §C for the verification.

### §B.3 Two constraints the mechanism imposes

**C1 — `esc` restoration is free.** Mode 1049 saves the cursor and switches to a separate buffer, so the main screen is untouched while the diff is open. "Prior scrollback intact" needs no application code. Design implication: `backToTable()` stays a pure state transition; it must not attempt any manual redraw.

**C2 — quit must resolve to the inline view first.** The renderer's close path branches on the *last* view: an alt-screen last frame is discarded on exit. Since `y`/`q`/`ctrl+c` are handled before the sub-view check, a quit from the diff view would discard the exit summary. Every quit path therefore sets `view = previewTableView` before returning `tea.Quit`.

`n` is not affected — the diff branch returns before reaching its case, so `n` is table-only by design. This asymmetry is intentional and must be preserved (see `plan.md` §G G5e).

### §B.4 Exit-frame contract

The exit summary is **the final `View()`**, not a post-program print. The inline close path writes `EraseScreenBelow` after the last frame, so:

- the last frame persists in scrollback;
- everything below it is cleared;
- therefore a final view containing exactly one line leaves exactly one line.

Shape: `<status glyph> <outcome> · <N> files` — one line, theme-painted, resolving its glyph from `tui.Glyph*` and its colour from `Success` (confirmed) or `Dim` (cancelled). It carries **no** class counts and **no** file rows; re-rendering those is the duplicate-render symptom this design removes.

---

## §C TUX-token → interactive-role map

The single mapping table all three interactive surfaces conform to. M3's audit reconciles the two huh factories against this column set.

| Interactive role | TUX token | Preview TUI (M1) | huh v1 (M3) | huh v2 (M3) |
|------------------|-----------|------------------|-------------|-------------|
| Primary accent / focus | `Accent` | selected table row, section title | `Title`, `NoteTitle`, selectors, indicators | `Primary` → same set |
| Secondary / prompt | `Info` | — | `TextInput.Prompt` | `Secondary` |
| Positive outcome | `Success` | `add` class label, confirmed exit line | `SelectedOption` | `SelectedOption`, `SelectedPrefix` |
| Negative / error | `Danger` | `conflict` class label | `ErrorIndicator`, `ErrorMessage` | `Error` |
| Informational | `Info` | `update` class label | — | — |
| De-emphasized | `Dim` | `preserve` class label, cancelled exit line, key hints | `TextInput.Placeholder` | `Muted`, `UnselectedPrefix` |
| Body text | `Body` | diff body | `Description` | `Body` |
| Primary text | `Fg` | table cells | `Option`, `UnselectedOption` | `Text` |
| Structural border | `ChromeBorder` | table border, summary card border | *(gap — v1 sets no `Base` border)* | `Border` ← `Rule` |
| Panel fill | `Panel` | — | `BlurredButton` background | `ButtonBlurredBg` |

Two observations this table makes visible, both feeding M3-2:

1. The **`ChromeBorder` row is where the two huh factories diverge** — v2 styles `Base`/`Card` borders, v1 does not. That is the principal audit gap.
2. v2 additionally sets prefix strings (`SelectedPrefix` `◆ `, `UnselectedPrefix` `◇ `, `SelectSelector` `▸ `) that v1 leaves at library defaults. Whether v1's API exposes equivalents must be checked against the v1 `FieldStyles` type before a gap is called a defect.

Note the table uses `Rule` → `Border` for the wizard because that is the existing `wizardTokens` mapping; the preview uses `ChromeBorder` directly. Reconciling these two border tokens is an M3-2 decision, not a pre-settled one.

---

## §D View composition

### §D.1 Table sub-view

```
┌─ tui.Box(Accent) ────────────────────────────────┐
│  Classification summary                          │   ← tui.Section title
│    add: 3   update: 12   preserve: 2   conflict:1│   ← semantic-coloured labels
└──────────────────────────────────────────────────┘
  ┌──────────────────────────────────────────────┐
  │ Class          │ File                        │   ← bubbles table,
  │ update         │ .claude/settings.json       │     table.WithStyles from Theme
  │ preserve       │ .claude/skills/hns-my/...   │     border = ChromeBorder
  └──────────────────────────────────────────────┘     selected row = Accent
  enter view diff · y confirm · n/q cancel            ← tui.HelpBar
```

Replaces: the bare `Classification summary:\n` text block (`buildClassCountHeader`), the unstyled `table.New(...)`, and the inline hint literal at `preview_tui.go:143`.

### §D.2 Diff sub-view

Full-viewport, alt-screen. Header names the file; body is the diff; footer is a `tui.HelpBar` carrying `esc back to table`. The empty-diff placeholder notice is preserved verbatim — it is existing behaviour, not a presentation decision.

### §D.3 Fallback (TUX layer, not TUI)

Structurally ANSI-free — **no `lipgloss` import, no `internal/tui` import, no escape byte**, unconditionally. The redesign is layout-only: aligned columns, grouped sections, an ASCII-drawn card outline. Its value is that the guarantee needs no runtime guard, so it must not be re-expressed as `if !noColor { ... }`.

---

## §E Design decisions and their alternatives

| # | Decision | Chosen | Rejected | Reason |
|---|----------|--------|----------|--------|
| 1 | Residue policy | Hybrid per-sub-view | All-alt-screen | Hides the classification card during confirmation |
| 2 | | | All-inline | Duplicate render — the reported symptom |
| 3 | Theme reach into pkg `update` | `tui.ResolveOS()` + package-level indirection var | `Theme` field on `PreviewOptions` | Avoids a signature change and a caller edit; matches how both huh factories already resolve independently |
| 4 | Regression coverage | Structure golden + token tests (split) | Colour-bearing golden alone | Couples `internal/cli/update` tests to every `internal/tui` palette edit |
| 4a | Token-assertion method | SGR **parameter run** substring | Raw hex token | lipgloss v2 emits decimal RGB; the hex never appears |
| 4b | | | Full-CSI prefix | v2 merges attributes into one CSI, so a bold selected row breaks the prefix (`research.md` §D.1) |
| 5 | | | Structure golden alone | Palette-insensitive by design — cannot detect the defect class |
| 6 | Dead-code removal shape | Surgical: delete the program, relocate the two live types to `types.go` | Delete `confirm.go` wholesale | `MergeAnalysis` / `FileAnalysis` have four production consumer sites |
| 7 | | | Leave a 20-line `confirm.go` | Filename would no longer describe contents |
| 8 | huh factory unification | Keep two factories, align token mapping | Merge into one | The v1/v2 `Theme` types differ across the library major boundary |
| 9 | Wizard theme location | Leave in `wizard.go` tail | Move to `styles.go` | Moving ~110 lines enlarges the conflict surface with the parallel SPEC |

---

## §F Cross-references

- `spec.md` §A.2 — the TUX / TUI / huh vocabulary this design is organized around
- `plan.md` §B.10a — renderer verification behind §B.2/§B.3; §F M1 — execution sequence
- `research.md` §C — bubbletea v2 alt-screen findings; §D — lipgloss v1/v2 rendering difference
- `acceptance.md` §C Group B — the ACs binding §B.3 and §B.4
