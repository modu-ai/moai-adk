---
id: SPEC-CLI-TUI-MODERNIZE-001
title: "Research — interactive TUI surface modernization"
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

# Research — SPEC-CLI-TUI-MODERNIZE-001

Measured substrate gathered during plan authoring and plan-audit repair. Every entry names the command that produced it. Figures not produced by an executed command are marked **unmeasured**.

The numeric baselines in `plan.md` §B were independently reproduced by plan-auditor — all 7 line counts, all 5 coverage figures, the 4-hit hex split, the 6/3/4 cross-package census, the Windows-guard relocation, and the alt-screen mechanism. Nothing in §A-§E below was contradicted.

---

## §A Library version pins

`go.mod` (lines 6, 7, 10):

| Module | Pin |
|--------|-----|
| `charm.land/bubbles/v2` | **v2.1.1** |
| `charm.land/bubbletea/v2` | **v2.0.8** |
| `charm.land/lipgloss/v2` | **v2.0.5** |

Both lipgloss majors are present in the build: `github.com/charmbracelet/lipgloss` (v1) is imported by `internal/cli/uikit`, while `charm.land/lipgloss/v2` is imported by `internal/tui` and by `bubbles/v2/table`. This split is load-bearing — see §D.

Provenance note: `plan.md` §B.10 originally attributed the table-API verification to `bubbles/v2@v2.0.0`. Corrected to v2.1.1 at v0.1.2; re-verified against the pinned version with byte-identical line numbers.

---

## §B bubbles v2 table styling API

`charm.land/bubbles/v2@v2.1.1/table/table.go`:

| Symbol | Line |
|--------|-----:|
| `type Styles struct` | 106 |
| `func DefaultStyles() Styles` | 113 |
| `func (m *Model) SetStyles(s Styles)` | 122 |
| `func WithStyles(s Styles) Option` | 188 |

Both the constructor option and the post-construction setter exist, so REQ-TUIM-011 is implementable either way.

**Current state:** `preview_tui.go:87-93` calls `table.New` with `WithColumns` / `WithRows` / `WithFocused` / `WithHeight` / `WithWidth` — and no `WithStyles`. The table therefore renders `DefaultStyles()`, which is the component library's own palette, not MoAI's. This is the mechanical root of "the update screen looks dated".

---

## §C bubbletea v2 alt-screen

### §C.1 It is a view field, not a program option

`grep "^func With" charm.land/bubbletea/v2@v2.0.8/options.go` lists 12 `ProgramOption` constructors. **`WithAltScreen` is not among them.** Alt-screen is instead a field on the value returned from `View()`:

`charm.land/bubbletea/v2@v2.0.8/tea.go:76` `func NewView(s string) View`, `:84` `type View struct` — carrying `AltScreen`, consumed by the renderer at `cursed_renderer.go:87, 320, 519`.

This invalidates the common v1-era idiom `tea.NewProgram(m, tea.WithAltScreen())`, which does not compile against v2.

### §C.2 Mid-run toggling is first-class

`cursed_renderer.go:320` computes the transition per frame by comparing the previous view to the current one:

```go
shouldUpdateAltScreen := (s.lastView == nil && view.AltScreen) || (s.lastView != nil && s.lastView.AltScreen != view.AltScreen)
```

On transition it emits the DEC private mode 1049 pair (`:513-525`): `ansi.SetModeAltScreenSaveCursor` entering, `ansi.ResetModeAltScreenSaveCursor` leaving.

- `enterAltScreen` (`:653`): `SaveCursor` → set mode → `SetFullscreen(true)` → `Erase`
- `exitAltScreen` (`:663`): `Erase` → `SetRelativeCursor(true)` → `SetFullscreen(false)` → reset mode → `RestoreCursor`

Because mode 1049 saves the cursor and switches to a separate screen buffer, the main screen is untouched for the duration. **Scrollback preservation across an alt-screen excursion is a terminal-mode guarantee, not application work.** No application-level fallback is required.

### §C.3 Two caveats

**Keyboard-protocol renegotiation per transition.** `cursed_renderer.go:378-380` re-negotiates the Kitty keyboard protocol whenever `view.AltScreen` changes, and `:513-517` disables keyboard enhancements across the switch. Library-handled, but rapid toggling produces protocol churn. Recorded so it is not later mis-diagnosed as a defect.

**The close path discards an alt-screen final frame.** `cursed_renderer.go:167-171`:

```go
if lv.AltScreen {
    enableAltScreen(s, false, true)                    // alt buffer torn down; its content is gone
} else {
    _, _ = s.scr.WriteString(ansi.EraseScreenBelow)    // inline frame persists; below it is cleared
}
```

Consequences: (a) an exit summary written while in the diff sub-view is lost; (b) an inline final frame persists, so a one-line final view leaves exactly one line. Both are designed around in `design.md` §B.3/§B.4.

---

## §D lipgloss v1 vs v2 rendering — the assertion trap

**v2 renders colour as decimal RGB SGR; the hex token never appears in output.** Measured against the pinned `charm.land/lipgloss/v2@v2.0.5` in a standalone probe module:

```
token=#bf6547  rendered="\x1b[38;2;191;101;71mPROBE\x1b[m"  containsHexToken=false
token=#d97757  rendered="\x1b[38;2;217;119;87mPROBE\x1b[m"  containsHexToken=false
token=#c4432b  rendered="\x1b[38;2;196;67;43mPROBE\x1b[m"   containsHexToken=false
```

Consistent with lipgloss v2's own fixture (`style_test.go:107-110`): `Color("#5A56E0")` → `"\x1b[38;2;90;86;224mhello\x1b[m"`.

Token values read from `internal/tui/theme.go`:

| Token | Light | Dark |
|-------|-------|------|
| `Accent` | `#bf6547` (:118) | `#d97757` (:159) |
| `Danger` | `#b1432f` (:126) | `#ed7d6b` (:167) |
| `Success` | `#3d8b6e` (:122) | `#5bbf9a` (:163) |
| `Info` | `#1f7a7d` (:128) | `#5cc7c9` (:169) |
| `ChromeBorder` | `#bdbab2` (:108) | `#1c2624` (:149) |
| `Dim` | `#5b625f` (:114) | `#9aa3a0` (:155) |

**Implication for acceptance criteria:** `strings.Contains(rendered, th.Accent)` is **false for a correct implementation**. Token-application tests must compare against a rendered probe. This was a MUST-FIX plan-audit finding (D1) and is the reason AC-TUIM-004 / -008 / -030b state their method explicitly.

### §D.1 SGR parameters are merged into a single CSI

A second, independent trap on the same surface. `ansi.Style.String()` joins every parameter into one CSI rather than emitting one escape per attribute:

```go
func (s Style) String() string {
	if len(s) == 0 { return ResetStyle }
	return "\x1b[" + strings.Join(s, ";") + "m"
}
```

So an attribute combined with a colour is prepended **inside** the sequence. Executed against the pinned v2.0.5 with `#bf6547`:

```
fgOnly            = "\x1b[38;2;191;101;71mx\x1b[m"
bold+fg           = "\x1b[1;38;2;191;101;71mx\x1b[m"
fg+bold           = "\x1b[1;38;2;191;101;71mx\x1b[m"

full-CSI prefix "\x1b[38;2;191;101;71m"
  substring of fgOnly ? true
  substring of bold+fg? false
parameter run "38;2;191;101;71"
  substring of fgOnly ? true   bold+fg? true   fg+bold? true
```

Two consequences:

1. **A full-CSI-prefix assertion is false for a correct implementation** whenever the styled element also carries an attribute. This is not hypothetical here: bubbles' `DefaultStyles` sets `Selected: ...Bold(true).Foreground(...)` (`bubbles/v2@v2.1.1/table/table.go:115-116`), and `design.md` §D.1 keeps bold on the selected row. The fix is to assert the **SGR parameter run** — `acceptance.md` §C.2.
2. Builder order is irrelevant — `Bold().Foreground()` and `Foreground().Bold()` render identically — so the parameter run is robust to how the style is constructed.

This was delta-audit finding NEW-1: the v0.1.2 D1 repair removed the hex-literal defect but introduced a same-class false negative one layer down.

**Scope of the v1 colour-stripping behaviour.** Headless `Render` dropping SGR is a **v1** behaviour. `internal/tui` and `bubbles/v2/table` are v2, which performs no in-`Render` TTY detection — so forcing the theme axis and diffing light-vs-dark output works headlessly (AC-TUIM-003 is sound as written). Only `internal/cli/uikit` sits on v1, and it is out of this SPEC's scope.

---

## §E Theme axis precedence

`internal/tui/detect.go` `Resolve` documents and implements the chain:

1. `NO_COLOR` non-empty → `MonochromeTheme`
2. `MOAI_THEME="light"` → `LightTheme`
3. `MOAI_THEME="dark"` → `DarkTheme`
4. `MOAI_THEME="auto"` or unset → `env.DetectDark()`
5. `DetectDark()==false` → `LightTheme`
6. default → `DarkTheme` (safe default)

An invalid `MOAI_THEME` value short-circuits to `DarkTheme` **without** querying the terminal.

Both huh factories resolve through this chain via a package-level indirection var — `huhThemeIsDark = tui.IsDarkOS` (`huh_theme.go:15`) and `wizardIsDark = tui.IsDarkOS` (`wizard.go:483`) — so tests can force an axis without mutating process environment. M1 follows the same pattern.

**Testing consequence:** forcing the axis through the indirection var **bypasses** the chain, so it cannot verify precedence. AC-TUIM-039 exercises the chain through the real resolution path with `t.Setenv` (non-parallel per `CLAUDE.local.md` §6); AC-TUIM-003 keeps using the indirection var for the cheaper light-vs-dark difference check.

---

## §F M2 reachability census

Method: for each top-level declaration in `internal/merge/confirm.go`, `grep -rn "\bSYM\b" --include="*.go" .` filtered to exclude `_test.go` and `confirm.go` itself. A surviving hit means a production consumer.

### §F.1 `ConfirmMerge` — dead

`grep -rn "ConfirmMerge" --include="*.go" .`:

| Site | Kind |
|------|------|
| `internal/merge/confirm.go:915` | definition |
| `internal/merge/confirm_test.go:478` | **only call site** — a test |
| `internal/cli/update.go:1124, 1131, 1137` | comment prose |
| `internal/cli/update/preview.go:5, 18, 59` | comment prose |
| `internal/cli/coverage_improvement_test.go:4286-4290`, `internal/cli/update_skip_sync_test.go:109-112` | comment prose in tests |

Zero production callers. The live path routes `update.go` → `confirmViaPreview` → `update.PreviewClassification`.

### §F.2 Cross-package census

`grep -rn "merge\.[A-Z][A-Za-z]*" internal/ cmd/ pkg/ | grep -v _test.go | grep -v "^internal/merge/"` → `merge.FileAnalysis` 6, `merge.MergeAnalysis` 3, `merge.ConfirmMerge` 4 (all 4 comment-only, verified individually).

### §F.3 Live types — four production consumer sites

| Type | Site |
|------|------|
| `MergeAnalysis` | `internal/cli/update.go:1110` `toPreviewInputs(analysis merge.MergeAnalysis, …)` |
| `MergeAnalysis` | `internal/cli/update/merge/merge.go:230` `AnalyzeMergeChanges(…) mrg.MergeAnalysis` |
| `FileAnalysis` | `internal/cli/update_tux.go:88` `classifyUpdateCounts(files []merge.FileAnalysis)` |
| `FileAnalysis` | `internal/cli/update/plan/plan.go:63, 64, 88` — signature, local slice, and a **named-field composite literal** at `:88-94` |

The `plan.go:88` literal names `Path` / `Changes` / `Strategy` / `RiskLevel` / `Note`, making it the site most sensitive to REQ-TUIM-031's field-set guarantee. It was present in the §F.2 census from the start but was omitted from `plan.md` §B.4's KEEP table until v0.1.2 (plan-audit finding S3).

### §F.4 Windows non-TTY guard — already relocated

| Site | Condition | Error string |
|------|-----------|--------------|
| `internal/merge/confirm.go:940` (dead) | `runtime.GOOS == "windows" && !isatty.IsTerminal(os.Stdin.Fd())` | identical |
| `internal/cli/update.go:1145` (**live**) | `!isatty.IsTerminal(os.Stdin.Fd())` — all platforms | identical |

The live guard is strictly broader. `grep -rn "REQ-CFS-00[78]" --include="*.go" .` returns exactly two hits: the dead guard and the `update.go:1137` comment documenting the relocation. M2 drops no behaviour.

### §F.5 Security controls reachable only from dead code

`validateAnalysis` (file-count and path-length DoS limits, `confirm.go:850`) and `sanitizePath` (path-traversal sanitization, `:884`) are called only from `ConfirmMerge` (`:917`, `:923`) and from tests. They are therefore **already not running in production**, so deleting them is not a regression.

They do surface a pre-existing observation: the live path (`confirmViaPreview` → `toPreviewInputs`) applies neither control. That gap predates this SPEC and is **out of scope**; it is recorded here so the removal does not erase the only trace of it.

---

## §G Existing test coverage and the blind spot

`internal/cli/testdata/tuxiu/` holds **12** top-level `.golden` files — 6 scenarios (`init`/`update` × `tty`/`notty`/`nocolor`) × 2 streams (stdout/stderr) — plus **12** more under `postm4/`, **24** total (`find internal/cli/testdata/tuxiu -name '*.golden' | wc -l` → `24`). They are consumed by 4 test functions in `internal/cli/tuxiu_characterization_test.go`. None invokes the interactive program.

`internal/cli/update/preview_test.go` carries 12 test functions that **do** drive the model headlessly (`newPreviewModel` → `tableView()` / `diffView()` / `selectRow()` / `backToTable()`). Reading their bodies: every assertion is `strings.Contains` over class labels and file paths — **content only**, zero presentation assertions.

So the blind spot is not "the interactive path is untested" but **"the interactive path has no presentation regression coverage"**. That distinction determines the shape of the new coverage: re-asserting content would reproduce the gap.

Coverage baselines, `go test -cover` (exit 0): `internal/cli` 74.9%, `internal/cli/update` 70.2%, `internal/merge` 86.3%, `internal/cli/wizard` 95.2%, `internal/tui` 93.6%. `internal/cli` sits below the 90% critical-package target at entry, so ACs bind no-regression rather than an absolute floor.

---

## §H Shell and tooling hazards affecting verification commands

Recorded because three plan-audit findings (D3, S1, S5) were about commands that looked correct but could not have produced their recorded output.

### §H.1 Escaped pipe inside an ERE

A markdown-escaped `\|` in a `grep -E` pattern is a **literal pipe**, not alternation. Positive control:

```
file: x := "#AABBCC"
grep -rnE '"#[0-9a-fA-F]{6}"\|#[0-9a-fA-F]{6}'  → exit 1 (no match)   ← broken
grep -rnE '"#[0-9a-fA-F]{6}"|#[0-9a-fA-F]{6}'   → exit 0, matches     ← correct
```

Same for the glyph pattern (control `a := '✓'` / `b := '●'`): escaped → exit 1; unescaped → 2 matches.

Both commands report `0` on a healthy tree, so a broken regex is indistinguishable from a clean result. Every such command now lives in `acceptance.md` §C.1 as a fenced block **paired with a positive control**.

### §H.2 `$'\x1b\['` is not a valid bracket expression

In this repo's zsh, the `$'...'` form collapses `\[` to a bare `[`, leaving an unterminated bracket expression:

```
/usr/bin/grep -c $'\x1b\[' internal/cli/update/preview_fallback.go
  → grep: brackets ([ ]) not balanced      exit 2
/usr/bin/grep -c $'\x1b' internal/cli/update/preview_fallback.go
  → 0                                       exit 1 (correct: exit 1 == zero matches)
```

The ANSI-free property is true; the originally-recorded command could not have shown it.

### §H.3 Environment notes

- The default `grep` in this shell is a **ugrep 7.5.0** wrapper function, not GNU/BSD grep. Both it and `/usr/bin/grep` reproduce §H.1 and §H.2 identically.
- `moai spec lint` was run via the installed binary `/Users/goos/go/bin/moai` because `go run ./cmd/moai` failed to build on an unrelated in-flight edit from a parallel session (`internal/template/profile_matrix.go:341: undefined: maps`). Re-run from a clean build before the audit gate.

---

## §I Cross-references

- `plan.md` §B — the same measurements in execution-sequence context
- `design.md` §B — the state model built on §C; §C — the token map built on §D/§E
- `acceptance.md` §C.1 — the corrected commands from §H
