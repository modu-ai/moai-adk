---
id: SPEC-CLI-TUI-MODERNIZE-001
title: "Acceptance criteria — interactive TUI surface modernization"
version: "0.1.3"
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

# Acceptance Criteria — SPEC-CLI-TUI-MODERNIZE-001

## §A Verification principles

1. **Presentation, not content.** 12 content-only tests already exist in `internal/cli/update/preview_test.go` and did not catch this defect class (plan.md §B.9). Any AC whose only assertion is "the label string is present" is vacuous here and is rejected.
2. **Measured denominators.** Every baseline figure below traces to a command executed during plan authoring. Each must be **re-measured at run-phase entry** before it is used as an AC denominator; a stale denominator makes the AC vacuous.
3. **Headless.** No AC may require a real TTY. The model exposes `tableView` / `diffView` / `currentView` / `selectRow` / `backToTable`, and `Update` accepts synthetic `tea.KeyPressMsg` values.
4. **Deletion is proved by reachability, not by a green build.** A green build after deleting a symbol proves only that nothing *compiles* against it. ACs in Group C assert the reachability evidence directly.

---

## §B Given-When-Then scenarios

### Scenario 1 — the reported defect is resolved

> **Given** a project whose templates have pending updates, and a terminal attached to stdin,
> **When** the user runs `moai update` without `--yes` and reaches the change-preview prompt,
> **Then** the preview renders with `internal/tui` theme tokens applied to its table, a structural card around the per-class count summary, and a `tui.HelpBar` key hint row — visually the same design family as the already-modernized `moai init` / `moai update` output path.

Headless proxy: `newPreviewModel(...).tableView()` output satisfies AC-TUIM-001 through AC-TUIM-006.

### Scenario 2 — the CI blind spot is closed

> **Given** the redesigned preview and its new presentation coverage,
> **When** a future change removes the theme wiring from `preview_tui.go` (simulating exactly the regression this SPEC repairs),
> **Then** at least one test in `internal/cli/update/` fails.

This is a **self-trip** requirement: run-phase must actually perform the mutation, observe the failure, and revert. A passing test suite is not evidence that the test would catch the regression.

### Scenario 3 — the non-interactive path is untouched

> **Given** `moai update --yes`, or a non-TTY stdin,
> **When** the update runs,
> **Then** the plain-text fallback renders with zero ANSI escape sequences, and the six `internal/cli/testdata/tuxiu/*.golden` fixtures compare equal without regeneration.

### Scenario 4 — the dead program is gone without collateral damage

> **Given** the M2 removal has landed,
> **When** the full test suite and the cross-platform build run,
> **Then** both pass, `internal/merge` carries no terminal-UI library dependency, and `merge.MergeAnalysis` / `merge.FileAnalysis` remain resolvable from their three production consumers.

### Scenario 5 — the parallel SPEC merges cleanly

> **Given** `SPEC-CLI-WIZARD-RESTRUCTURE-001` rewriting the question/group assembly region of `internal/cli/wizard/wizard.go`,
> **When** this SPEC's branch is merged with it,
> **Then** no conflict arises, because every hunk this SPEC produced in that file begins at or after the `var wizardIsDark` anchor line.

---

## §C Acceptance criteria

### Group A — M1 presentation (REQ-TUIM-010 … 017)

| ID | Criterion | Verification |
|----|-----------|--------------|
| **AC-TUIM-001** | `internal/cli/update/preview_tui.go` imports `internal/tui`. | `grep -n "modu-ai/moai-adk/internal/tui" internal/cli/update/preview_tui.go` → ≥1 hit. Baseline: **0 hits** (measured). |
| **AC-TUIM-002** | The preview table is constructed with an explicit style value, not the component default. | `grep -nE "table\.WithStyles|SetStyles" internal/cli/update/preview_tui.go` → ≥1 hit. Baseline: **0 hits** (measured). |
| **AC-TUIM-003** | Under a forced light axis and a forced dark axis, `tableView()` output differs. | Go test: force each axis via the package-level indirection var, render, assert `light != dark`. Proves the axis is actually consumed rather than resolved and discarded. |
| **AC-TUIM-004** | The resolved theme's `Accent` token reaches `tableView()` output **in rendered form**. | Go test using the **SGR-parameter-substring** method (§C.2). Extract the parameter run from a probe and assert it appears in the output:<br><br>`probe := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent)).Render("x")`<br>`params := strings.TrimSuffix(strings.TrimPrefix(probe, "\x1b["), "x\x1b[m")` → `"38;2;191;101;71"`<br>`assert strings.Contains(rendered, params)`<br><br>**Two forms are wrong and must not be substituted in** — see §C.2 for the measured evidence: (1) `strings.Contains(rendered, th.Accent)` — the hex never appears, lipgloss v2 emits decimal RGB; (2) the full-CSI prefix `"\x1b[38;2;191;101;71m"` — lipgloss v2 merges all SGR parameters into one CSI, so a bold+coloured cell renders `"\x1b[1;38;2;191;101;71m"` and the prefix is **not** a substring. The parameter run survives both. Still hex-literal-free, so D1 and AC-TUIM-020 hold. |
| **AC-TUIM-005** | The per-class count summary is rendered through an `internal/tui` structural primitive. | `grep -nE "tui\.(Box|Section|ThickBox)" internal/cli/update/preview_tui.go` → ≥1 hit; plus a Go test asserting a box-border rune is present in `tableView()` output under a colour-enabled axis. |
| **AC-TUIM-006** | Key hints render through `tui.HelpBar`; no inline hint literal survives. | `grep -n "tui.HelpBar" internal/cli/update/preview_tui.go` → ≥1 hit **AND** `grep -nF "[enter] view diff" internal/cli/update/preview_tui.go` → 0 hits. Baseline: the literal is present at line 143. |
| **AC-TUIM-007** | The four class labels are byte-identical to their current `ChangeClass.String()` values. | The four existing content assertions in `preview_test.go` (`TestPreviewTableRendersPerClassCounts`, `TestPreservedLabelInTUITable`) pass **unmodified**. Any edit to those test bodies fails this AC. |
| **AC-TUIM-008** | Each class label carries a distinct semantic colour role. | Go test: render under a colour-enabled axis; derive the **SGR parameter run** per role via the §C.2 method (as in AC-TUIM-004), and assert the four labels each carry a **different** such run. Comparison is against the parameter run — never the raw hex token, and never a full-CSI prefix (both fail for a correct implementation; §C.2). |
| **AC-TUIM-009** | Under `NO_COLOR`, `tableView()` and `diffView()` emit zero ANSI sequences. | Go test: force the monochrome axis, assert `!strings.Contains(out, "\x1b[")` for both views. |
| **AC-TUIM-010** | Any status glyph in the preview resolves from `tui.Glyph*`, and no raw glyph rune literal is introduced. | Command **CMD-GLYPH** in §C.1, plus its positive control. Both outputs must be cited. Sole coverage for REQ-TUIM-017. |

### Group B — M1 fallback and residue (REQ-TUIM-018 … 021)

| ID | Criterion | Verification |
|----|-----------|--------------|
| **AC-TUIM-011** | `preview_fallback.go` contains no ANSI escape sequence and imports no colour library. | Command **CMD-ANSI** in §C.1 → output `0`. **AND** `grep -n "lipgloss" internal/cli/update/preview_fallback.go` → 0 hits, `grep -n "internal/tui" internal/cli/update/preview_fallback.go` → 0 hits (two separate greps; no alternation needed). Carries forward AC-TUX3-010. |
| **AC-TUIM-012** | The existing fallback ANSI tests pass unmodified. | `TestPreviewFallbackZeroANSIUnderNoColor` and `TestPreviewFallbackZeroANSIWhenPiped` pass with unedited bodies. |
| **AC-TUIM-013** | The fallback's layout is restructured: file rows are column-aligned and the summary is visually grouped. | Golden comparison of `renderFallback(...)` over a fixed fixture against a new `testdata` golden, plus a structural assertion that the class column has a uniform width across rows. |
| **AC-TUIM-014a** | The table sub-view renders **inline**. | Go test: `m.View()` while `m.currentView() == previewTableView` returns a `tea.View` with `AltScreen == false`. |
| **AC-TUIM-014b** | The diff sub-view renders on the **alternate screen**. | Go test: after `m.selectRow()`, `m.View()` returns a `tea.View` with `AltScreen == true`. |
| **AC-TUIM-014c** | `esc` restores the inline table sub-view. | Go test: drive `Update` with a synthetic `tea.KeyPressMsg` for `esc` from the diff sub-view; assert `currentView() == previewTableView` **and** the subsequent `View()` carries `AltScreen == false`. Prior-scrollback preservation is a guarantee of DEC mode 1049 (plan.md §B.10a) and is not separately asserted. |
| **AC-TUIM-014d** | Every **diff-reachable** quit key resolves to the inline sub-view before quitting, and the final view is a single result line. | Go test, one case per key in **`y` / `q` / `ctrl+c`**, each entered **from the diff sub-view**: assert `currentView() == previewTableView` after the quit message, `View().AltScreen == false`, the final view is one line, and it contains neither the class-count summary nor any file row. The from-diff entry is what exercises the discard hazard (plan.md §B.10a caveat 2).<br><br>**`n` is deliberately excluded here** — `Update` (`preview_tui.go:202-236`) handles `ctrl+c`/`q`/`y` before the sub-view check and *returns inside* the diff branch, so `n` never reaches a quit case from the diff view; it falls through to `viewport.Update`. Requiring `n`-from-diff would fail a correct implementation. `n` is covered by AC-TUIM-014f. Do **not** make `n` diff-reachable — that is an unspecified behaviour change contradicting the documented keymap at `preview_tui.go:199-200`. |
| **AC-TUIM-014e** | The chosen residue policy is documented at its implementation site. | A doc comment in `preview_tui.go` names the hybrid policy, the per-sub-view `AltScreen` mechanism, and the quit-must-resolve-to-inline constraint with its reason. |
| **AC-TUIM-014f** | The table-view-only quit key `n` also resolves to a single-line inline exit. | Go test, one case, entered **from the table sub-view**: send `n`; assert `currentView() == previewTableView`, `View().AltScreen == false`, the final view is one line carrying the cancelled outcome, and `confirmed == false`. |
| **AC-TUIM-015** | Classification still derives solely from the single-source entry point. | `TestPreviewTableClassificationMatchesClassify` passes unmodified **AND** `grep -nE "func Classify|classifyAll" internal/cli/update/preview_tui.go` shows no re-implementation. |
| **AC-TUIM-016** | The five headless contract accessors remain present with unchanged signatures. | `grep -nE "func \(m \*previewModel\) (tableView|diffView|currentView|selectRow|backToTable)" internal/cli/update/preview_tui.go` → 5 hits. |

### Group C — M2 surgical deletion (REQ-TUIM-030 … 037)

| ID | Criterion | Verification |
|----|-----------|--------------|
| **AC-TUIM-017** | `ConfirmMerge`, `confirmModel`, `fileListItem`, and the `AnalysisFormatter` stack are absent from the repository. | `grep -rn "ConfirmMerge\|confirmModel\|fileListItem\|AnalysisFormatter" --include="*.go" .` → hits only in comment prose, or 0. Any hit in executable code fails. |
| **AC-TUIM-018** | `merge.MergeAnalysis` and `merge.FileAnalysis` remain resolvable, with unchanged field sets, from **all four** production consumer sites. | `go build ./...` passes **AND** every site below still references the type: <br>1. `grep -n "merge.MergeAnalysis" internal/cli/update.go` (≥1)<br>2. `grep -n "merge.FileAnalysis" internal/cli/update_tux.go` (≥1)<br>3. `grep -n "mrg.MergeAnalysis" internal/cli/update/merge/merge.go` (≥1)<br>4. `grep -n "merge.FileAnalysis" internal/cli/update/plan/plan.go` (≥1 — expect lines 63, 64, 88)<br><br>Site 4 is the **field-set-sensitive** one: `plan.go:88-94` is a named-field composite literal (`Path`, `Changes`, `Strategy`, `RiskLevel`, `Note`). Renaming or removing any of those five fields breaks it at compile time, which is exactly the REQ-TUIM-031 "field set unchanged" guarantee. A green build alone does **not** satisfy this AC. |
| **AC-TUIM-019** | `internal/merge` imports no terminal-UI library. | `grep -rn "charm.land/bubbletea\|charm.land/bubbles\|charmbracelet/lipgloss" internal/merge/*.go` → 0 hits (excluding `_test.go` only if the tests genuinely need none; state which). Baseline: `confirm.go` imports all three. |
| **AC-TUIM-020** | Raw hex colour literals on **executable** production lines outside `internal/tui/` drop to zero. | Command **CMD-HEX** in §C.1, plus its positive control. Expected after M2: exactly **2** hits, both comment-only in `internal/cli/wizard/styles.go` (lines 17, 20). Baseline: **4** hits, of which 2 are executable code at `internal/merge/confirm.go:454-455`. The 2 comment-only hits are expected to remain — REQ-TUIM-002 / -034 are scoped to executable lines. |
| **AC-TUIM-021** | The live non-TTY confirmation guard survives with unchanged behaviour. | `grep -n "isatty.IsTerminal(os.Stdin.Fd())" internal/cli/update.go` → ≥1 hit inside `confirmViaPreview` **AND** the error string is byte-identical to its pre-change value. |
| **AC-TUIM-022** | No surviving test references a deleted symbol. | `go vet ./...` and `go test ./internal/merge/...` both pass. Both `confirm_test.go` and `confirm_coverage_test.go` are handled. |
| **AC-TUIM-023** | The M2 removal changes no deploy/skip/backup/classify behaviour. | All **12** `internal/cli/testdata/tuxiu/*.golden` fixtures — 6 scenarios (`init`/`update` × `tty`/`notty`/`nocolor`) × 2 streams (stdout/stderr) — compare equal **without regeneration**, as do the **12** under `postm4/` (24 total, measured via `find internal/cli/testdata/tuxiu -name '*.golden' | wc -l`). `go test ./internal/cli/...` passes. |
| **AC-TUIM-024** | The plan's keep/delete inventory was re-verified at execution time, not carried from plan authoring. | The run-phase evidence block cites the re-run reachability commands (plan.md M2-1) with their verbatim output, and states whether the result matched plan.md §B.4. |

### Group D — M3 form-theme unification (REQ-TUIM-040 … 045)

| ID | Criterion | Verification |
|----|-----------|--------------|
| **AC-TUIM-025** | Every hunk this SPEC produces in `internal/cli/wizard/wizard.go` begins at or after the `var wizardIsDark` anchor line. | `git diff <base>..HEAD -- internal/cli/wizard/wizard.go` — parse each `@@ -a,b +c,d @@` header; assert every `c` ≥ the line number returned by `grep -n "var wizardIsDark" internal/cli/wizard/wizard.go` on the base revision (measured baseline: line **483**). Zero hunks is also a pass. |
| **AC-TUIM-026** | Two separate huh theme factories remain; neither is merged into the other. | `grep -n "func moaiHuhStyles" internal/cli/huh_theme.go` → 1 hit **AND** `grep -n "func moaiWizardStyles" internal/cli/wizard/wizard.go` → 1 hit. |
| **AC-TUIM-027** | The wizard theme functions were not relocated. | `newMoAIWizardTheme`, `moaiWizardStyles`, `wizardTokenSet`, and `wizardTokens` all remain in `internal/cli/wizard/wizard.go`; `grep -n "moaiWizardStyles" internal/cli/wizard/styles.go` → 0 hits. |
| **AC-TUIM-028** | The per-factory divergence audit is recorded, with every gap either closed or reasoned. | The run-phase evidence block carries a table naming each style field set by one factory and not the other, with a disposition (closed / reasoned-absent / absent-from-v1-API) per row. |
| **AC-TUIM-029** | Both light/dark indirection vars survive so tests can force either axis without touching the process environment. | `grep -n "var huhThemeIsDark" internal/cli/huh_theme.go` → 1 hit **AND** `grep -n "var wizardIsDark" internal/cli/wizard/wizard.go` → 1 hit. |

### Group E — cross-cutting (REQ-TUIM-050 … 056)

| ID | Criterion | Verification |
|----|-----------|--------------|
| **AC-TUIM-030a** | **Mechanism 1** — a structure golden captured under a forced monochrome axis exists for the table view, the diff view, and the fallback, and is palette-insensitive. | New `testdata` goldens plus their comparison tests exist in `internal/cli/update/` and pass headlessly. **Palette-insensitivity is proved, not assumed:** temporarily alter a colour token value in `internal/tui`, re-run `go test ./internal/cli/update/`, observe the structure-golden tests still PASS, revert. Cite the verbatim output. |
| **AC-TUIM-030b** | **Mechanism 2** — token-application unit tests assert each semantic role resolves to the correct `internal/tui` token, comparing against values read from the resolved `Theme` via the §C.2 SGR-parameter-substring method. | Tests cover at minimum: conflict label → `th.Danger`, table border → `th.ChromeBorder`, selected row → `th.Accent`, plus the four REQ-TUIM-014 class roles. `grep -nE '#[0-9a-fA-F]{6}' <new test files>` → **0 hits** — a hard-coded hex in these tests would itself violate D1.<br><br>The selected-row case is the one that *requires* §C.2: bubbles' own baseline styles it `Bold(true).Foreground(...)` (`table.go:115-116`) and design.md §D.1 keeps that, so it renders as a merged CSI. A full-CSI-prefix assertion would fail there against a correct implementation. |
| **AC-TUIM-030c** | The token-application tests actually trip when the theme wiring is removed. | **Self-trip, mandatory, bound to mechanism 2.** Remove the theme wiring from `preview_tui.go`; run `go test ./internal/cli/update/`; observe ≥1 FAIL **from a mechanism-2 test named in AC-TUIM-030b**; revert. The run-phase evidence block cites the failing test name and verbatim output. A FAIL originating only from a mechanism-1 golden does **not** satisfy this AC — mechanism 1 is palette-insensitive by design (AC-TUIM-030a) and provably cannot detect this defect class, which is exactly why the split exists. Without this self-trip, AC-TUIM-030a and 030b are unfalsifiable. |
| **AC-TUIM-031** | Cross-platform build stays clean. | `GOOS=windows GOARCH=amd64 go build ./...` → exit 0. Baseline: exit 0 (measured). |
| **AC-TUIM-032** | Statement coverage does not regress below the measured baseline for each affected package. | `go test -cover ./internal/cli/ ./internal/cli/update/ ./internal/merge/ ./internal/cli/wizard/ ./internal/tui/`. Baselines measured at plan authoring: `internal/cli` **74.9%**, `internal/cli/update` **70.2%**, `internal/merge` **86.3%**, `internal/cli/wizard` **95.2%**, `internal/tui` **93.6%**. **`internal/merge` carve-out:** M2 deletes ~900 lines of well-covered code plus two test files; the resulting ratio shift is **unmeasured** at plan time. For that package the criterion is: report the post-change figure and state whether the delta is attributable to the deletion. A drop attributable solely to removing well-covered dead code is an accepted outcome, not a failure. |
| **AC-TUIM-033** | No new module dependency. | `git diff -- go.mod go.sum` → empty. |
| **AC-TUIM-034** | No template file is touched. | `git diff --stat -- internal/template/templates/` → empty. |
| **AC-TUIM-035** | M1 is independently landable. | M1's commits build and pass tests with M2 and M3 absent. Demonstrated by the milestone commit sequence (M1 commits precede any M2/M3 commit and each is green). |
| **AC-TUIM-036** | Lint stays clean. | `golangci-lint run --timeout=3m` → exit 0, or the pre-existing finding set unchanged (state the baseline count if non-zero). |

### Group F — vocabulary and axis-precedence (REQ-TUIM-001, -003, -015)

| ID | Criterion | Verification |
|----|-----------|--------------|
| **AC-TUIM-037** | The three-layer vocabulary is bound in the SPEC and each requirement group names the layer it governs. | `spec.md` §A.2 contains the TUX / TUI / huh-form-runtime table with all three rows, and each of Groups B, C, D in §B carries a layer label in its heading. Document review. Covers REQ-TUIM-001. |
| **AC-TUIM-038** | The `internal/tui` package is not renamed, and the naming caveat is recorded. | `test -d internal/tui` succeeds **AND** `git diff --stat <base>..HEAD -- internal/tui/` shows no file rename (no `=>` in the stat output) **AND** `spec.md` §A.2 carries the "package `internal/tui` implements the TUX layer" caveat paragraph. Covers REQ-TUIM-003. |
| **AC-TUIM-039** | The preview inherits the canonical axis-precedence chain rather than re-implementing it. | Go test with `t.Setenv`, exercising the chain at `internal/tui/detect.go` `Resolve` **through the preview's own resolution path** (not through the test indirection var, which bypasses it):<br>(a) `NO_COLOR=1` + `MOAI_THEME=dark` → monochrome wins;<br>(b) `MOAI_THEME=light` (no `NO_COLOR`) → light;<br>(c) `MOAI_THEME=dark` → dark;<br>(d) `MOAI_THEME=auto` and `MOAI_THEME` unset → defers to detection;<br>(e) `MOAI_THEME=<invalid>` (e.g. `purple`) → `DarkTheme` **without** querying the terminal — the `default:` short-circuit branch in `Resolve`.<br>Case (a) is the ordering assertion — it fails if the chain is re-ordered. Case (e) is the branch REQ-TUIM-015's "shall not re-implement or partially re-order" reaches but which (a)-(d) leave untested. Covers REQ-TUIM-015, which AC-TUIM-003 (forces the axis via the indirection var, bypassing the chain) and AC-TUIM-009 (`NO_COLOR` only) together leave unverified. Per `CLAUDE.local.md` §6, use `t.Setenv` in a non-parallel test. |

**Acceptance-criteria count: 46** (Group A 10, Group B 11, Group C 8, Group D 5, Group E 9, Group F 3). Lowercase-suffixed IDs (`AC-TUIM-014a`..`f`, `AC-TUIM-030a`..`c`) are paired sub-criteria of one logical AC and are counted individually.

---

## §C.1 Command appendix

Commands whose regexes cannot survive a markdown table cell. Each is given in a fenced block so **no escaping is applied** — a markdown-escaped `\|` inside an ERE is a *literal pipe*, not alternation, and silently matches nothing.

Each command is paired with a **positive control**, because every one of them reports `0` on a healthy tree: without a control, a broken regex is indistinguishable from a clean result. Cite both outputs.

### CMD-HEX — raw hex colour literals outside `internal/tui/` (AC-TUIM-020)

```bash
grep -rnE '"#[0-9a-fA-F]{6}"|#[0-9a-fA-F]{6}' --include="*.go" internal/ cmd/ pkg/ \
  | grep -v "^internal/tui/" \
  | grep -v "_test\.go:"
```

Positive control (regex proof):

```bash
printf 'x := "#AABBCC"\nplain line\n' > /tmp/moai-ac-ctl.txt
grep -nE '"#[0-9a-fA-F]{6}"|#[0-9a-fA-F]{6}' /tmp/moai-ac-ctl.txt   # MUST match line 1
```

### CMD-GLYPH — raw status-glyph rune literals in the preview (AC-TUIM-010)

```bash
grep -cE "'✓'|'✗'|'●'|'○'" internal/cli/update/preview_tui.go
```

Positive control (regex proof):

```bash
printf "a := '✓'\nb := '●'\nplain\n" > /tmp/moai-ac-glyph.txt
grep -cE "'✓'|'✗'|'●'|'○'" /tmp/moai-ac-glyph.txt                   # MUST report 2
```

### CMD-ANSI — ANSI escape sequences in the fallback (AC-TUIM-011)

```bash
grep -c $'\x1b' internal/cli/update/preview_fallback.go
```

Expect output `0`. Note `grep -c` exits **1** when the count is zero — that is the success case here, not an error.

Do **not** use `grep -c $'\x1b\['`: in this repo's zsh the `$'...'` form collapses `\[` to a bare `[`, leaving an unterminated bracket expression. Measured: `/usr/bin/grep` → `grep: brackets ([ ]) not balanced`, exit 2. A command that errors cannot substantiate a "→ 0" claim (constraint D9).

---

## §C.2 Token-application assertion method (SGR parameter substring)

The canonical method for AC-TUIM-004, AC-TUIM-008, and AC-TUIM-030b. Two intuitive forms are **false for a correct implementation**; this section records the measurement that rules each out, so a future editor does not "simplify" back into either.

### The method

```go
probe  := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent)).Render("x")
params := strings.TrimSuffix(strings.TrimPrefix(probe, "\x1b["), "x\x1b[m")
// params == "38;2;191;101;71"
if !strings.Contains(rendered, params) { t.Errorf(...) }
```

Assert on the **SGR parameter run**, never on the whole escape sequence.

### Why not the hex token

`th.Accent` is a hex string, but lipgloss v2 converts it to decimal RGB before emission — the hex never reaches the output.

### Why not the full-CSI prefix

lipgloss v2 joins **all** SGR parameters into a single CSI (`github.com/charmbracelet/x/ansi` `Style.String()`). Any attribute combined with the colour is prepended *inside* the sequence, so a foreground-only prefix stops being a substring. This is not hypothetical: bubbles' `DefaultStyles` sets `Selected: ...Bold(true).Foreground(...)` (`table.go:115-116`), and design.md §D.1 keeps bold on the selected row.

### Measured evidence

Executed against the pinned `charm.land/lipgloss/v2@v2.0.5` with `th.Accent` = `#bf6547` (light):

```
fgOnly            = "\x1b[38;2;191;101;71mx\x1b[m"
bold+fg           = "\x1b[1;38;2;191;101;71mx\x1b[m"
fg+bold           = "\x1b[1;38;2;191;101;71mx\x1b[m"

probePrefix (full-CSI form) = "\x1b[38;2;191;101;71m"
  substring of fgOnly ? true
  substring of bold+fg? false      ← fails on a correct implementation

params (this method) = "38;2;191;101;71"
  substring of fgOnly ? true
  substring of bold+fg? true
  substring of fg+bold? true
```

Note `bold+fg` and `fg+bold` render identically, so the parameter run is also robust to the order attributes are applied in the style builder.

### Interaction with AC-TUIM-030c

The self-trip is unaffected by this choice: removing the theme wiring drops the SGR entirely, so both the full-CSI form and the parameter form would trip. The defect this section fixes is a **false negative on a correct implementation**, not a missed regression.

---

## §D Severity classification

| Severity | ACs | Meaning |
|----------|-----|---------|
| **MUST-FIX (blocking)** | AC-TUIM-001, 002, 004, 006, 011, 014d, 017, 018, 020, 021, 025, 030b, 030c, 031, 033, 034 | A failure here means the SPEC did not do its job, broke a live surface, or produced an unfalsifiable claim. AC-TUIM-014d is blocking because the quit-path discard hazard is invisible to compilation and to every content-level test. |
| **SHOULD-FIX** | AC-TUIM-003, 005, 008, 009, 010, 013, 014a, 014b, 014c, 014e, 014f, 019, 022, 023, 024, 026, 027, 028, 029, 030a, 035, 036, 038, 039 | A failure is a real gap; run-phase must either fix it or return a blocker report with the reason. AC-TUIM-010 is promoted from INFORMATIONAL: it is the sole coverage for REQ-TUIM-017, so treating it as advisory left that requirement unverified. |
| **INFORMATIONAL** | AC-TUIM-007, 012, 015, 016, 032, 037 | No-regression guards, document-review checks, and reported measurements. A change here is not automatically a failure but must be explained. |

---

## §E Traceability

| Requirement group | Requirements | Acceptance criteria |
|-------------------|--------------|---------------------|
| A — vocabulary / layer discipline | REQ-TUIM-001..003 | REQ-001 → AC-TUIM-037; REQ-002 → AC-TUIM-020; REQ-003 → AC-TUIM-038 |
| B — M1 preview redesign | REQ-TUIM-010..022a | AC-TUIM-001..016 (incl. AC-TUIM-014a..f). REQ-TUIM-015 additionally → AC-TUIM-039; REQ-TUIM-017 → AC-TUIM-010 (sole coverage) |
| C — M2 dead removal | REQ-TUIM-030..037 | AC-TUIM-017..024 |
| D — M3 form-theme unification | REQ-TUIM-040..045 | AC-TUIM-025..029 |
| E — cross-cutting | REQ-TUIM-050..056 (incl. -050a, -050b) | AC-TUIM-030a..c, 031..036 |

---

## §F Indirect verification notes

Three properties cannot be asserted directly and use a stated proxy:

1. **"Looks modern"** is not machine-checkable. The proxy is *mechanism presence plus effect*: the theme import (AC-TUIM-001), an explicit style value (AC-TUIM-002), the token reaching the output (AC-TUIM-004), and the axis actually changing the output (AC-TUIM-003). Presence alone would be a dead-prose pass; the effect assertions are what make the set non-vacuous.

1a. **"Prior scrollback intact after `esc`"** is a property of the terminal, not of the program — DEC private mode 1049 preserves the main screen buffer while the alternate screen is active (plan.md §B.10a). Asserting it directly would require a PTY harness, which §A.3 forbids. The proxy is AC-TUIM-014c: assert the model returns to the inline view and the emitted view carries `AltScreen == false`, which is the condition under which the terminal's own guarantee applies.

1b. **"The run appears in scrollback exactly once"** is likewise not directly observable headlessly. The proxy is AC-TUIM-014d: the final view is one line and contains no classification content, combined with the renderer's documented inline close path (`EraseScreenBelow` after the last frame). The two together are sufficient for the single-appearance property.
2. **"Dead code"** is proved by reachability evidence (AC-TUIM-017 + AC-TUIM-024), not by a green build. A green build after deletion proves only that nothing compiles against the symbol.
3. **"Merges cleanly with the parallel SPEC"** cannot be verified until that SPEC lands. The proxy is the hunk-range containment assertion (AC-TUIM-025), which is a sufficient condition for the file-level conflict this constraint exists to prevent.

---

## §G Closure gates

Before this SPEC may transition out of run-phase:

- [ ] All 16 MUST-FIX acceptance criteria PASS with cited verbatim command output.
- [ ] Every §C.1 command was run **together with its positive control**, and both outputs are cited. A `0` result without a passing control is not evidence.
- [ ] No test asserts a raw hex token against rendered output (D1 — lipgloss v2 emits decimal RGB SGR, so such an assertion is false for a correct implementation).
- [ ] Zero open clarification markers remain (both resolved at v0.1.1; see plan.md §I Resolved decisions).
- [ ] The AC-TUIM-030c self-trip was actually performed, observed, and reverted — with the failing test name cited, and that test confirmed to be a **mechanism-2** (token-application) test, not a structure golden.
- [ ] The AC-TUIM-030a palette-insensitivity check was performed (token altered, goldens still green, reverted).
- [ ] AC-TUIM-014d was exercised **from the diff sub-view** for the three diff-reachable quit keys (`y` / `q` / `ctrl+c`), and AC-TUIM-014f covered `n` **from the table view**. `n` is not diff-reachable by design — demanding it from the diff view would reinstate the defect D2 removed.
- [ ] No token-application test asserts a full-CSI prefix; all use the §C.2 SGR-parameter-substring method.
- [ ] The AC-TUIM-024 reachability re-verification was run at execution time, and any divergence from plan.md §B.4 is reported.
- [ ] Every coverage and count figure in the completion report is attributed to a command run in that session, per `.claude/rules/moai/core/verification-claim-integrity.md` §2.
- [ ] `go test ./...` passes.

## §H Definition of Done

1. `moai update` presents an interactive preview in the same visual family as the modernized output path: the table inline over the visible classification card, the diff on the alternate screen, and one result line left in scrollback.
2. `internal/merge` holds no interactive program, no terminal-UI dependency, and no raw hex literal — with every live type preserved.
3. Both huh theme factories share one token-to-role mapping; remaining divergences are documented.
4. A presentation regression in the preview surface fails a test, headlessly — demonstrated, not asserted.
5. `internal/cli/wizard/wizard.go` carries no hunk that could conflict with `SPEC-CLI-WIZARD-RESTRUCTURE-001`.
