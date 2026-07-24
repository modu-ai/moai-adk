# Acceptance Criteria — SPEC-CLI-TUX-INIT-UPDATE-001

Development mode: **TDD** (write the failing test first, then implement). All rendering AC assert observable string/ANSI output; the preservation AC (AC-TUXIU-016) asserts byte-identical data against the M1 golden stdout+stderr characterization baseline (there is no JSON/structured output surface — D2).

## §D AC Matrix — Given-When-Then

### Look item 1 — Card-style classification summary

**AC-TUXIU-001** (REQ-TUXIU-010)
- **Given** `moai update` has classified merge changes as add:1 update:23 conflict:2,
- **When** the summary renders on a colour TTY,
- **Then** the output contains an accent-bordered `tui.Box` (Accent:true) whose body carries three pills — `PillOk` "+ 1 add", `PillInfo` "~ 23 update", `PillErr` "! 2 conflict".

**AC-TUXIU-002a** (REQ-TUXIU-011)
- **Given** a clean update with conflict count 0,
- **When** the summary card renders,
- **Then** no "! 0 conflict" pill appears (zero-count pill omitted).

**AC-TUXIU-002b** (REQ-TUXIU-011)
- **Given** an update with add:0 update:5 conflict:0,
- **When** the summary card renders,
- **Then** only the "~ 5 update" pill appears (both zero-count pills omitted).

### Look item 2 — Unified status glyphs

**AC-TUXIU-003** (REQ-TUXIU-012, REQ-TUXIU-013)
- **Given** the update deploy checklist,
- **When** a step is pending / running / done,
- **Then** its glyph is exactly ○ (Faint) / ● (Accent, bold) / ✓ (Success) respectively, resolved from `tui.StatusIcon` — never a per-step redefined glyph.

**AC-TUXIU-004** (REQ-TUXIU-001, REQ-TUXIU-002, REQ-TUXIU-003)
- **Given** the consolidated glyph source (the exported `tui.Glyph*` const block),
- **When** the codebase is grepped for literal glyph declarations,
- **Then** the ✓/✗/○/● status glyphs resolve from exactly ONE canonical `tui` source, and `printer.go` / `uikit` carry ONLY references to it — ZERO raw-rune status-glyph declarations remain in those two packages. The pre-existing raw-rune sites `printer.go:280,282`, `uikit/styles.go:25,28`, `uikit/status.go:11,15`, and `uikit/render.go:72` MUST be **removed** (not merely "no NEW declaration added").
- **Verify:** `grep -rnE '"✓"|"✗"|"○"|"●"' internal/cli/printer/ internal/cli/uikit/ | grep -v _test` returns **no output** (zero raw-rune string literals — references use `tui.Glyph*` and do not contain the quoted rune). Scope is printer+uikit per D5; the `tui.Spinner`/`tui.Stepper` internal frame literals in `internal/tui/status.go` are a separate rendering concern outside this grep.

### Look item 3 — Block progress bar

**AC-TUXIU-005** (REQ-TUXIU-014)
- **Given** the deploy runs 5 steps,
- **When** 3 of 5 have completed,
- **Then** the output contains a `tui.Progress` block bar (██████░░░░ style) reflecting 3/5, in addition to or replacing the "3/5 steps complete" text.

### Look item 4 — Spinner-residue removal

**AC-TUXIU-006** (REQ-TUXIU-020, REQ-TUXIU-021)
- **Given** a deploy step with an in-flight spinner on a TTY,
- **When** the step finishes,
- **Then** the spinner line is cleared (ANSI erase-line) and exactly ONE clean ✓ result line is printed — no stale two-part spinner residue remains.
- **Verify (RED first):** a characterization test drives a step to completion and asserts the captured output contains exactly one result line for that step and zero residual `stepMarker`/`spinnerFrame` fragments.

**AC-TUXIU-007** (REQ-TUXIU-022)
- **Given** output is not a TTY,
- **When** a step finishes,
- **Then** a single newline-terminated result line is printed with NO ANSI erase sequence (the TTY/non-TTY split is preserved).
- **Verify:** capture the non-TTY step output and assert it contains ZERO ANSI CSI sequences — `grep -c $'\x1b\\[' <non-tty-capture>` returns `0` (no erase-line `\r\x1b[2K`, no colour CSI); exactly one newline-terminated result line per step.

### Look item 5 — Identity header band

**AC-TUXIU-008** (REQ-TUXIU-015)
- **Given** `moai update` starts,
- **When** the header renders,
- **Then** it shows "◆ MoAI-ADK <version> <go-runtime> · claude" with the version rendered as a solid `PillPrimary` pill.

### Look item 6 — Outcome banner

**AC-TUXIU-009** (REQ-TUXIU-016)
- **Given** `moai update` updated 24 files,
- **When** the run completes,
- **Then** the outcome renders as a solid `PillOk` "✓ Updated 24 files" pill followed by a dim detail note.

### Look item 7 — init banner + success card

**AC-TUXIU-010** (REQ-TUXIU-030)
- **Given** `moai init` starts,
- **When** the banner renders,
- **Then** it shows the same ◆ MoAI-ADK identity band + version/go/claude pills as the update header band, sourced from the shared `tui` primitives.

**AC-TUXIU-011** (REQ-TUXIU-031)
- **Given** `moai init` succeeds,
- **When** the success card renders,
- **Then** `buildInitSuccessCard` output uses the shared `tui.Box` + `tui.Pill` visual language.

### Look item 8 — Large logo restoration

**AC-TUXIU-020** (REQ-TUXIU-050, REQ-TUXIU-054) — logo present on all 3 surfaces
- **Given** each of the three surfaces `moai` (root, no-args), `moai init`, and `moai update`,
- **When** the surface renders on a colour TTY,
- **Then** the ANSI-stripped output contains the restored 6-line ASCII-art logo (greppable first-line marker `███╗   ███╗`) rendered ABOVE the compact `◆ MoAI-ADK` identity band (both present — the band is retained, stacked below the logo).
- **Verify:** for each of the 3 surfaces, strip ANSI from the capture and assert `grep -F '███╗   ███╗'` returns ≥1 (logo present) AND `grep -F '◆ MoAI-ADK'` returns ≥1 (compact band retained). The `PrintBanner` entry is shared, so a passing `moai update` capture plus the shared-entry cross-reference (`root.go:32`, `init.go:410`, `update.go:1253`) covers all three; the no-args `moai` capture additionally confirms the logo prints above the fang help body.

**AC-TUXIU-021** (REQ-TUXIU-051, REQ-TUXIU-052) — per-line coral gradient (top light → bottom deep)
- **Given** a truecolor TTY capture of the logo (the 6 rows),
- **When** the rows render,
- **Then** at least 3 DISTINCT coral foreground SGR colour codes appear across the 6 lines, ordered top-light → bottom-deep (the ramp is `tui.CoralRamp(6)` = `lipgloss.Blend1D(6, Accent, AccentDeep)`, which yields 6 distinct CIELAB-interpolated stops).
- **Verify:** on a forced-truecolor capture, `grep -oE $'\x1b\\[38;2;[0-9;]*m' <capture> | sort -u | wc -l` returns ≥ 3 (Blend1D(6,·) produces 6 distinct 24-bit fg codes); AND assert directionality — the row-1 fg RGB is lighter (higher perceptual luminance) than the row-6 fg RGB (row 1 = `Accent`, the lighter coral; row 6 = `AccentDeep`, the deeper coral).

**AC-TUXIU-022** (REQ-TUXIU-053) — NO_COLOR → plain, non-TTY → solid downsample
- **Given** `NO_COLOR=1`,
- **When** the logo renders,
- **Then** the art prints as plain text with ZERO ANSI colour sequences, and the art runes are still present.
- **Verify:** with `NO_COLOR=1`, capture the logo and assert (a) `grep -c $'\x1b\\[[0-9;]*m' <capture>` returns `0` (no SGR colour), and (b) `grep -F '███╗   ███╗'` returns ≥1 (art runes intact). Separately, a non-TTY pipe capture (colour allowed but downsampled) shows the logo as a single solid accent colour, not a 6-way gradient — assert the distinct-fg-code count on the non-TTY pipe capture is ≤ 1.

**AC-TUXIU-023** (REQ-TUXIU-051, REQ-TUXIU-040) — theme-SSOT guard still holds with the ramp in `tui`
- **Given** the logo colour ramp derives from `Theme.Accent`/`Theme.AccentDeep` via `tui.CoralRamp`,
- **When** the tree is grepped for hex colour literals,
- **Then** the AC-TUXIU-012 guard is unaffected — no NEW `#RRGGBB` literal appears outside `internal/tui/` (the logo colour code lives in `internal/tui/logo.go`, INSIDE the SSOT boundary), and `internal/tui/logo.go` itself contains zero hex literals (ramp colours come from tokens, not literals).
- **Verify:** `grep -rnE '#[0-9a-fA-F]{6}' internal/cli/ | grep -v _test` still returns exactly the 1 pre-existing comment baseline (`wizard/styles.go:20`) — the logo added no `internal/cli` hex literal — AND `grep -rnE '#[0-9a-fA-F]{6}' internal/tui/logo.go` returns no output (the new logo file sources colour from `Accent`/`AccentDeep` tokens, not literals).

**AC-TUXIU-024** (REQ-TUXIU-055) — explicit root-help surface carries the logo; subcommand help does not; no-args prints it exactly once
- **Given** the root-help predicate wired on the `runFang` seam (pre-`fang.Execute` logo print),
- **When** each of the five invocations `moai --help`, `moai help`, `moai init --help`, `moai help init`, and no-args `moai` renders,
- **Then** (a) the ANSI-stripped output of `moai --help` AND of `moai help` each contains the logo signature `███╗   ███╗`; (b) the ANSI-stripped output of `moai init --help` AND of `moai help init` contains it ZERO times (subcommand help is logo-free); and (c) the ANSI-stripped output of no-args `moai` contains it **exactly once** — occurrence count `== 1`, NOT `>= 1` (the double-print guard: the no-args path is already covered by `PrintBanner` in `rootCmd.Run`, so a predicate that also matched the empty arg vector would emit two logos).
- **Verify:** for each invocation, capture output, strip ANSI, and count the signature with `grep -c -F '███╗   ███╗'` (line-count is sufficient — the signature occupies the logo's first row, one line per render):
  - `moai --help` → `>= 1`
  - `moai help` → `>= 1`
  - `moai init --help` → `0`
  - `moai help init` → `0`
  - `moai` (no args) → **exactly `1`** — assert equality, not `>=`; a value of `2` is the double-print regression this AC exists to catch.
- **Predicate unit tests (RED first):** the `os.Args[1:]` predicate is additionally unit-tested over all 6 arg shapes of plan.md §A.1 L4 — `[]` / `["--help"]` / `["-h"]` / `["help"]` / `["help","init"]` / `["init","--help"]` — asserting matched for exactly the three root-help shapes and NOT matched for the empty list and both subcommand-help shapes.

### Constraint 1 — Theme SSOT

**AC-TUXIU-012** (REQ-TUXIU-040)
- **Given** the redesigned render code,
- **When** the tree is grepped for hex colour literals,
- **Then** no NEW `#RRGGBB` literal appears outside `internal/tui/` beyond the 1 pre-existing baseline hit.
- **Verify:** `grep -rnE '#[0-9a-fA-F]{6}' internal/cli/ | grep -v _test` returns exactly the 1 pre-existing baseline hit — `internal/cli/wizard/styles.go:20`, which is INSIDE a comment (`// Secondary: info color (replaces purple #7C3AED / #5B21B6)`), not an active colour literal. The guard is "no NEW hex literal beyond that 1 comment baseline" (verified plan-phase: the grep line count is 1).

### Constraint 2 — NO_COLOR / monochrome

**AC-TUXIU-013a** (REQ-TUXIU-041)
- **Given** `NO_COLOR=1`,
- **When** `moai update` renders,
- **Then** output is plain text with no ANSI colour, and every pill degrades to "[label]" form.
- **Verify:** with `NO_COLOR=1`, capture `moai update` output and assert (a) zero ANSI SGR colour sequences — `grep -c $'\x1b\\[[0-9;]*m' <capture>` returns `0` — and (b) each count pill appears in bracketed `[label]` form (e.g. `[+ 1 add]`, `[~ 23 update]`) rather than a coloured pill.

**AC-TUXIU-013b** (REQ-TUXIU-041)
- **Given** output is redirected to a non-TTY pipe,
- **When** `moai update` renders,
- **Then** output downsamples to plain text (lipgloss downsample) with pills as "[label]".
- **Verify:** pipe `moai update` to a non-TTY sink and assert the count pills render as bracketed `[label]` forms with no ANSI colour (same `[label]` assertion as AC-013a, triggered by the non-TTY path rather than `NO_COLOR`).

### Constraint 3 — Reduced motion

**AC-TUXIU-014** (REQ-TUXIU-042)
- **Given** `MOAI_REDUCED_MOTION=1`,
- **When** a spinner and a progress bar render,
- **Then** the spinner is the static ● and the progress bar is fully filled (existing static-fallback behaviour preserved).
- **Verify:** with `MOAI_REDUCED_MOTION=1`, assert the spinner frame is the static `●` (not an animated `⠋`-family braille frame) and the progress bar renders fully filled — `grep` the capture for the static `●` and assert ZERO `⠋`-family spinner frames and ZERO unfilled `░` cells.

### Constraint 4 — No new dependency

**AC-TUXIU-015** (REQ-TUXIU-043)
- **Given** the completed implementation,
- **When** `go.mod` is diffed against the pre-SPEC baseline,
- **Then** no new module dependency line is added (only `charm.land/*` v2 + `internal/tui` are used).

### Constraint 5 — Characterization / data + JSON preservation

**AC-TUXIU-016** (REQ-TUXIU-044) — **top-risk preservation gate**
- **Given** the M1 golden fixtures capturing `moai init` and `moai update` **stdout AND stderr** (TTY + non-TTY + NO_COLOR), captured BEFORE any M1 edit,
- **When** the redesign is complete (after M4),
- **Then** (a) an ANSI-stripped structural diff of the file/count/outcome lines against the golden is empty, AND (b) the stdout-vs-stderr line partition (printer gateway discipline) is unchanged.
- **Verify:** the characterization test `TestInitUpdateTUXCharacterization` (in `internal/cli/tuxiu_characterization_test.go`) replays `moai init` and `moai update`, strips ANSI, and asserts the file/count/outcome lines match the goldens at `internal/cli/testdata/tuxiu/{init,update}.{tty,notty,nocolor}.{stdout,stderr}.golden` AND that each captured line appears on the SAME channel (stdout vs stderr) as in the golden.
- **Logo note (v0.1.2):** the restored logo header lines (the 6 ASCII-art rows) are EXPECTED NEW presentation output on the init/update surfaces — they carry NO file/count/outcome DATA, so they are OUTSIDE the data-line diff of this AC and do NOT invalidate it. The M1.0 pre-flight golden captures the PRE-logo baseline; its file/count/outcome DATA-line subset is the invariant. Post-M4, a fresh golden set WITH the logo is captured as the new expected PRESENTATION baseline, but the DATA-line subset comparison against the M1.0 baseline remains empty (the logo adds presentation, not data). The logo rides `Printer.Data` → stdout (same channel as the compact band), so the stdout-vs-stderr partition (b) is preserved.

**AC-TUXIU-017** (REQ-TUXIU-045) — **N/A**
- `moai init`/`moai update` have **no** machine-readable/JSON output surface (verified: no `--json`/`--format`/`--output` flag; the `encoding/json` in `update.go` is settings-file I/O — nothing to preserve). This AC is **not applicable** and is NOT counted as a MUST gate. The stdout/stderr channel preservation that matters is asserted by AC-TUXIU-016.

### Constraint 6 — Glyph whitelist + coverage

**AC-TUXIU-018** (REQ-TUXIU-004)
- **Given** the consolidated glyph set,
- **When** the render code is scanned,
- **Then** every status glyph is within the AC-CLI-TUI-017 whitelist (✓ ✗ ! · ● ○ ◆ ◇); no new emoji-range codepoint appears.

**AC-TUXIU-019** (REQ-TUXIU-046)
- **Given** the NEW/touched render functions this SPEC adds in `internal/cli` and `internal/tui`,
- **When** per-function coverage is extracted,
- **Then** each new/touched render function reaches ≥90% statement coverage, AND `internal/cli` whole-package coverage does not regress below its measured 74.6% baseline (`internal/tui` baseline 93%).
- **Verify:** `go test -coverprofile=/tmp/tuxiu-cover.out ./internal/cli/... ./internal/tui/... && go tool cover -func=/tmp/tuxiu-cover.out | grep -E '<new-render-func-names>'` — cite each new render func's per-function `N.N%` line (must be ≥90.0%); separately cite the `internal/cli` package total (must be ≥ 74.6%, no regression). The whole-package ≥90% number is NOT the gate for `internal/cli` — its baseline is already below 90%, independent of this presentation-only SPEC.

## §D.1 Severity Classification

| Severity | AC | Rationale |
|----------|----|-----------|
| MUST (blocking) | AC-TUXIU-006, 007, 016, 012, 015, 013a, 013b, 014, 019, 022, 023, 024 | Behaviour/data preservation, spinner-residue fix, SSOT, degradation, coverage — regression here breaks correctness or the demo. AC-016 is the top-risk preservation gate (data + stdout/stderr channel). AC-022 (logo NO_COLOR/non-TTY degradation) and AC-023 (logo theme-SSOT guard) are correctness gates in the same class as AC-013a/b and AC-012. AC-024 is MUST because of its double-print component: the no-args occurrence-count-`== 1` assertion is a correctness invariant (a mis-scoped predicate renders two logos on the most-visible surface), and the subcommand-help absence assertion guards predicate over-matching — both outrank the SHOULD-class "logo appears on `--help`" look aspect bundled into the same AC. |
| SHOULD | AC-TUXIU-001, 002a, 002b, 003, 004, 005, 008, 009, 010, 011, 018, 020, 021 | The approved look items — visually load-bearing but not correctness gates. AC-020 (logo present on 3 surfaces) and AC-021 (per-line gradient) are the approved logo look. |
| N/A | AC-TUXIU-017 | No machine-readable/JSON output surface exists for init/update — nothing to preserve (D2). |

## §D.2 Traceability (AC → REQ)

| AC | REQ | Look item / constraint |
|----|-----|------------------------|
| AC-TUXIU-001 | REQ-TUXIU-010 | Card summary |
| AC-TUXIU-002a/b | REQ-TUXIU-011 | Zero-count pill omission |
| AC-TUXIU-003 | REQ-TUXIU-012/013 | Unified glyphs |
| AC-TUXIU-004 | REQ-TUXIU-001/002/003 | Glyph SSOT |
| AC-TUXIU-005 | REQ-TUXIU-014 | Block progress bar |
| AC-TUXIU-006 | REQ-TUXIU-020/021 | Spinner-residue removal |
| AC-TUXIU-007 | REQ-TUXIU-022 | Non-TTY split |
| AC-TUXIU-008 | REQ-TUXIU-015 | Identity header band |
| AC-TUXIU-009 | REQ-TUXIU-016 | Outcome banner |
| AC-TUXIU-010 | REQ-TUXIU-030 | Init banner |
| AC-TUXIU-011 | REQ-TUXIU-031 | Init success card |
| AC-TUXIU-012 | REQ-TUXIU-040 | Theme SSOT |
| AC-TUXIU-013a/b | REQ-TUXIU-041 | NO_COLOR/monochrome |
| AC-TUXIU-014 | REQ-TUXIU-042 | Reduced motion |
| AC-TUXIU-015 | REQ-TUXIU-043 | No new dependency |
| AC-TUXIU-016 | REQ-TUXIU-044 | Data + stdout/stderr channel preservation (top-risk gate) |
| AC-TUXIU-017 (N/A) | REQ-TUXIU-045 (N/A) | JSON untouched — no machine-readable surface (D2) |
| AC-TUXIU-018 | REQ-TUXIU-004 | Glyph whitelist |
| AC-TUXIU-019 | REQ-TUXIU-046 | Coverage |
| AC-TUXIU-020 | REQ-TUXIU-050/054 | Logo present on 3 surfaces (stacked above compact band) |
| AC-TUXIU-021 | REQ-TUXIU-051/052 | Per-line coral gradient (top light → bottom deep) |
| AC-TUXIU-022 | REQ-TUXIU-053 | Logo NO_COLOR → plain / non-TTY → solid |
| AC-TUXIU-023 | REQ-TUXIU-051/040 | Logo ramp theme-SSOT guard (no hex outside tui) |
| AC-TUXIU-024 | REQ-TUXIU-055 | Explicit root-help surface (`moai --help` / `moai help`) carries the logo; subcommand help does not; no-args prints it exactly once (double-print guard) |

> **REQ coverage note:** REQ-TUXIU-055 (explicit root-help logo injection) is bound by **AC-TUXIU-024** (MUST) — the open item was resolved IN SCOPE at Implementation Kickoff via approach (a) (root-help predicate + pre-`fang.Execute` print; plan.md §A.1 L4 / §B R8), so the 4th surface is no longer capability-gated. REQ-TUXIU-056 (status-glyph-whitelist carve-out for the logo's decorative runes) is verified negatively by AC-TUXIU-018 remaining scoped to the status-glyph vocabulary (the carve-out is what keeps AC-018 passing once the logo's block/box-drawing runes appear) — no separate positive AC is required.

## §D.3 Edge Cases

- Update with all three counts zero (no changes): summary card omits all pills, or renders a "no changes" note — confirm the demo's empty state.
- Very large file counts (N ≥ 1000): pill/label width does not break the card border.
- `NO_COLOR=1` AND `MOAI_REDUCED_MOTION=1` simultaneously: plain text + static fallbacks compose without conflict.
- Non-TTY pipe: no ANSI, plain-text downsample with `[label]` pills (there is no JSON/structured mode — D2).
- A step that fails mid-deploy: the ✗ (error) glyph renders and the spinner still clears (no residue on the error path).
- Logo under `NO_COLOR=1`: plain-text art (no ANSI), runes intact — the 6-way gradient collapses to monochrome (AC-TUXIU-022).
- Logo on a narrow terminal (< 60 cols): the 6-line art may wrap; confirm the wrap does not corrupt the stacked compact band below it (presentation-only — no data impact).
- Logo + `MOAI_REDUCED_MOTION=1`: no interaction — the logo is static art (not animated), so reduced-motion neither fills nor freezes it; the spinner/progress reduced-motion fallback (AC-TUXIU-014) is unaffected.
- No-args `moai` vs explicit `moai --help`: BOTH show the logo, but via two DIFFERENT code paths — no-args through `PrintBanner` inside `rootCmd.Run` (REQ-TUXIU-054), explicit root-help through the pre-`fang.Execute` predicate print (REQ-TUXIU-055). The hazard is the overlap: if the predicate also matched the empty arg vector, no-args would render two logos. AC-TUXIU-024 asserts the no-args occurrence count is exactly 1.
- Subcommand help (`moai help init`, `moai init --help`): logo-free by design — the predicate must not over-match. Both shapes are asserted at count 0 by AC-TUXIU-024.
- `moai help` with an UNKNOWN subcommand (e.g. `moai help nosuchcmd`): `os.Args[1:]` is `["help","nosuchcmd"]` (len 2), so the predicate does NOT match and no logo is printed — cobra's unknown-topic error renders alone. This falls out of the len-==-1 rule; no special case is needed.
- Global flags before help (e.g. `moai --verbose --help`): `os.Args[1]` is `--verbose`, not a root-help token, so the predicate does NOT match and no logo prints. This is accepted behaviour for this SPEC — the predicate is deliberately first-token-only (simple and over-match-safe), and flag-permuted help is not one of the named surfaces.

## §E Definition of Done

- [ ] All MUST-severity AC pass with cited command output.
- [ ] All 8 approved look items render as the approved demo (SHOULD AC) — including the restored large logo.
- [ ] Logo present on all 3 `PrintBanner` surfaces (root no-args / init / update), stacked above the retained compact band (AC-TUXIU-020).
- [ ] Logo per-line coral gradient renders ≥3 distinct fg codes top-light→bottom-deep on truecolor (AC-TUXIU-021); degrades to plain under NO_COLOR / solid under non-TTY (AC-TUXIU-022).
- [ ] Logo ramp derives from `tui` `Accent`/`AccentDeep` tokens via `tui.CoralRamp`; no hex literal in `internal/tui/logo.go` or new hex outside `internal/tui/` (AC-TUXIU-023).
- [ ] Cross-SPEC reversal reconciled: `TestCompactBanner_NoASCIILogo` / `_TwoLineIdentity` / `_GlyphWhitelist` re-targeted to `bannerString` (compact band stays logo-free) or scoped to the band portion; `TestPrintBanner_*` updated logo-aware (RED-first). No unrelated test regressed.
- [ ] Explicit root-help logo injection (REQ-TUXIU-055) wired via the root-help predicate + pre-`fang.Execute` print: logo present on `moai --help` / `moai help`, absent on `moai init --help` / `moai help init`, and rendered exactly once on no-args `moai` (AC-TUXIU-024). Predicate unit-tested over all 6 arg shapes (RED-first). `go.mod` unchanged — no fang upgrade or fork.
- [ ] `go test ./...` green (no regression in out-of-scope commands).
- [ ] Per-function coverage ≥ 90% on the new/touched render funcs (via `go tool cover -func`); `internal/cli` whole-package ≥ 74.6% baseline (no regression). — D3
- [ ] `golangci-lint run` clean.
- [ ] `go.mod` unchanged (no new dependency).
- [ ] Theme-SSOT grep guard passes (no NEW hex literal outside `internal/tui/` beyond the 1 comment baseline at `wizard/styles.go:20`). — D9
- [ ] AC-TUXIU-016 golden stdout+stderr characterization passes (data + channel partition unchanged). — D4
- [ ] `make build` succeeds (binary rebuilt if any embedded template touched — none expected).
