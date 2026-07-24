# Implementation Plan — SPEC-CLI-TUX-INIT-UPDATE-001

> Ordering note: this plan leads with the highest-change-likelihood decisions (§A.1) so human review focuses there; the milestones (§F) then run in build-dependency order, each tagged with its reversibility.

## §A Context

Presentation-only modernization of `moai init` and `moai update` terminal output, connecting already-existing `internal/tui` primitives that the two commands under-use. **v0.1.2 adds an approved additive scope: restoring the large 6-line "MoAI-ADK" ASCII-art logo** (retired by SPEC-CLI-TUX-V3-004 REQ-TUX4-006) with a vertical coral gradient, rendered above the retained compact identity band on the 3 `PrintBanner` surfaces (root no-args / init / update) — see §A.1 items L1-L4 and §B R6-R8. Development mode is **TDD** (RED-GREEN-REFACTOR per `quality.yaml`). Coverage target is re-scoped per audit D3: **≥90% on the NEW/touched render functions** (per-function measurement), plus **no whole-package regression for `internal/cli` vs its measured baseline of 74.6%** (`internal/tui` baseline is 93%). The whole-package ≥90% target does NOT apply to `internal/cli` — its baseline is already below 90% independent of this presentation-only work. Reuse only the existing `charm.land/*` v2 + `internal/tui` stack.

### §A.1 Highest-change-likelihood decisions (review these FIRST)

These are the decisions most likely to change under review — surface them before the mechanical wiring:

1. **[NEW TYPE INTERFACE — RESOLVED at kickoff] Shape of the canonical glyph source → option (b).** The user-approved kickoff decision is **option (b)**: add an exported raw-rune constant block in the `tui` package as the single glyph SSOT — `tui.GlyphDone = '✓'`, `tui.GlyphRun = '●'`, `tui.GlyphSkip = '○'`, `tui.GlyphErr = '✗'`. `tui.StatusIcon` is refactored to RETURN those constants instead of its own string literals, so there is exactly ONE raw-rune source (not two). The external callers `printer.go` and `uikit` reference the constants directly while KEEPING their own theme-painting — `printer.go` keeps `p.paint(theme.Dim, marker)` with `marker` now sourced from `tui.GlyphSkip`, and `uikit` keeps `SuccessStyle.Render(...)` with the rune sourced from `tui.GlyphDone`. (Rune vs. string const realization — the callers that need a `string` may `string(tui.GlyphDone)` — is a run-phase implementation detail; the invariant is one exported source with zero literal-rune redeclarations.)
   - **Rejected: option (a)** (promote `StatusIcon(kind string) string` to the sole source). `printer` needs the raw marker WITHOUT theme styling, so routing it through the theme-painting `StatusIcon` would fight the printer's own `p.paint` gateway. A plain constant block keeps `printer.go`'s theme-painting intact while still removing the literal-rune duplication.

**Logo decisions (v0.1.2 additive scope — HIGHEST change-likelihood, review FIRST). These are the most visible new element and reverse a prior SPEC's decision, so they lead the review.**

- **L1. [NEW TYPE INTERFACE] Logo primitive + const home → `internal/tui/logo.go`.** Restore the verbatim 6-line `moaiBanner` art and expose it as a self-contained `tui` primitive: `tui.Logo(theme) string` (the per-line gradient-rendered banner) + `tui.CoralRamp(n int) []color.Color` (the stop generator). The ASCII-art constant lives in **`tui`, NOT `uikit`**. Rationale: the theme-SSOT rule (REQ-TUXIU-040/051, no hex outside `internal/tui/`) means the gradient interpolation MUST live in `tui`; homing the const beside the render helper keeps `tui.Logo(th)` fully self-contained (the `uikit` caller just prints `tui.Logo(th)`), and keeps the multi-line raw string out of `uikit/banner.go`. (Rune-vs-string realization of `CoralRamp`'s output and exact per-line render loop are run-phase details; the invariant is one exported logo primitive sourcing colour only from `Theme` tokens.)
  - **Rejected: restore the const in `uikit/banner.go` (its original home).** That would force the gradient interpolation to live in `uikit` (hex/colour math outside `internal/tui/`) OR pass a `[]color.Color` ramp across the package boundary for `uikit` to apply — both leak colour decisions out of the SSOT boundary. Homing everything in `tui/logo.go` keeps the SSOT clean.
- **L2. [NEW TYPE INTERFACE] Gradient ramp = `lipgloss.Blend1D(6, Accent, AccentDeep)`.** The vertical coral ramp is `charm.land/lipgloss/v2`'s `Blend1D` (CIELAB linear interpolation, `blending.go:18`) over the two theme accent tokens: `Accent` (lighter coral) → `AccentDeep` (deeper coral). `lipgloss/v2 v2.0.5` is ALREADY a direct `go.mod` dependency → zero new dependency (REQ-TUXIU-043 holds). Blend1D(6,·) yields 6 distinct stops → per-line top-light→bottom-deep. If a lighter-than-`Accent` top stop is later desired, the theme-SSOT-compliant path is a lightened `Accent` computed in `tui` via the transitively-available `go-colorful` — but Accent→AccentDeep is the default and satisfies the ≥3-distinct-fg AC.
  - **Rejected: hand-rolled RGB interpolation in `tui`.** Blend1D already does perceptually-uniform CIELAB blending and is dependency-free (already imported); a hand-rolled lerp is more code for a worse ramp (Enforce Simplicity — reuse the ecosystem helper).
- **L3. [USER-FACING/UX] Logo placement = stacked ABOVE the retained compact band, in `PrintBanner`.** `PrintBanner` composes `tui.Logo(th)` + `"\n"` + the existing compact `bannerString` and routes both through the Printer gateway (`Printer.Data` → stdout). This single edit covers all 3 surfaces (root no-args / init / update) because they share the one `PrintBanner` entry. **`bannerString` itself is NOT modified** — this is the reversal-minimizing choice: the compact band stays logo-free at the `bannerString` level, so SPEC-CLI-TUX-V3-004's "compact band stays compact" intent survives, and its three retirement tests reconcile by re-targeting from `PrintBanner` to `bannerString` (see §B R6).
- **L4. [USER-FACING/UX — RESOLVED at kickoff] Explicit `moai --help` / `moai help` fang-body injection → approach (a): root-help detector + pre-`fang.Execute` logo print.** The user-approved kickoff decision is **approach (a)**: the restored logo ALSO appears on the explicit `moai --help` / `moai help` path (the most common discovery invocation), wired as a root-help predicate on the `runFang` entry seam (`internal/cli/fang.go` / `internal/cli/root.go`) that inspects ONLY `os.Args[1:]` and prints `tui.Logo(th)` BEFORE `fang.Execute(...)`. The logo is routed through the same Printer gateway / stdout channel as `PrintBanner`, so the stdout-vs-stderr partition (REQ-TUXIU-044) is unchanged. **No fang change, no fang upgrade, no fork — `go.mod` stays unchanged (REQ-TUXIU-043 holds).** Re-verified at kickoff: fang v2.0.1 exposes only `WithVersion` / `WithoutVersion` / `WithoutManpage` / `WithoutCompletions` / `WithColorSchemeFunc` / `WithTheme` / `WithCommit` / `WithErrorHandler` / `WithNotifySignal` — no header/pre-help option exists, so the pre-`fang.Execute` print is the ONLY zero-dependency injection point.
  - **Predicate contract (matched vs NOT matched arg shapes).** The predicate inspects `os.Args[1:]` only:

    | `os.Args[1:]` | Example invocation | Matched (logo printed)? | Why |
    |---------------|--------------------|--------------------------|-----|
    | `[]` (empty) | `moai` | **NO** | Already covered by `rootCmd.Run` (`root.go:31-33` calls `uikit.PrintBanner` before `cmd.Help()`). Matching here would DOUBLE-PRINT the logo. |
    | `["--help"]` | `moai --help` | **YES** | Explicit root help — the primary discovery path. |
    | `["-h"]` | `moai -h` | **YES** | Short form of explicit root help. |
    | `["help"]` (len == 1) | `moai help` | **YES** | Cobra's `help` subcommand with no target = root help. |
    | `["help", <sub>]` | `moai help init` | **NO** | Subcommand help — never carries the logo. |
    | `[<sub>, ...]` (any first arg not in the token set) | `moai init --help`, `moai update` | **NO** | First arg is a subcommand; subcommand help never carries the logo. |

  - **HARD invariant — no-args exclusion.** Excluding the empty arg list is a HARD invariant of this predicate, not an optimization. The no-args surface already prints the logo via `PrintBanner` inside `rootCmd.Run`; a predicate that matches `[]` produces two logos on the single most-visible surface.
  - **Token source.** The existing `trivialCommands` map (`root.go:43-51`, already containing `--help` / `-h` / `help`) is reused as the token source where practical rather than re-declaring the tokens — one vocabulary, not two.
  - **Rejected: approach (b)** (no-args only — accept that the logo shows on `moai` / `moai init` / `moai update` but not on explicit `--help`). Rejected because it leaves the most common discovery path (`moai --help`) logo-free, which is precisely the surface the restoration is for.
  - **Rejected: approach (c)** (upgrade or fork fang to gain a header option). Rejected because it breaches the no-new-dependency envelope (REQ-TUXIU-043) and is out of this SPEC's scope; approach (a) achieves the same visible result with zero dependency movement.

2. **[USER-FACING/UX] Update summary card layout.** Accent `tui.Box` containing the three count pills on one line vs. pills above a thin rule. Zero-count pills omitted (REQ-TUXIU-011). This is the most visible change and the one the approved demo pinned — encode the demo's exact layout.
3. **[USER-FACING/UX] Identity header band + outcome banner wording.** "◆ MoAI-ADK <version> <go-runtime> · claude" (solid `PillPrimary`) and "✓ Updated N files" (solid `PillOk`) + dim note. Wording/pill-kind are cheap to change but visually load-bearing.
4. **[BEHAVIOUR] Spinner-clear contract.** Each finished step MUST clear its spinner line (ANSI erase) and print exactly one ✓ result line; the TTY/non-TTY split in `progress_line.go` must be preserved. This is a behaviour fix (not just cosmetic) and the subtlest correctness risk.
5. **[MECHANICAL] Glyph semantics mapping.** ✓→ok/done (Success), ●→run (Accent bold), ○→skip/pending (Faint). Already the `StatusIcon` mapping; low change-likelihood, listed for completeness.

## §B Known Issues / Risks

- **R1 — live animated printer handles.** The spinner-residue fix touches `printer.go` `stepHandle`/`spinnerHandle` (411-595) which drive live goroutine-free re-render. A wrong erase-sequence order can leave residue OR erase a wanted line. Mitigation: characterization test asserting exactly one result line per step on both TTY and non-TTY paths (RED first).
- **R2 — symbol-source consolidation is cross-package.** Editing `tui`, `printer`, and `uikit` glyph sites risks a regression in commands NOT in scope (doctor/status). Mitigation: the glyph runes are byte-identical to today's; only the DECLARATION site moves. Full `go test ./...` after M1.
- **R3 — hidden hex-literal creep.** New render code in `update.go`/`uikit` must pull colour from `Theme` only. Mitigation: a grep guard AC (REQ-TUXIU-040) run in M4.
- **R4 — (removed per D2).** There is NO JSON/structured output path in `moai init`/`moai update` — verified plan-phase: no `--json`/`--format`/`--output` flag is declared, and the `encoding/json` in `update.go` is settings-file I/O, not command output. There is no JSON branch to locate or preserve, so the earlier "locate the JSON branch during M2" task is dropped. The channel-discipline risk that remains (stdout-vs-stderr partition) is covered by R5 + AC-TUXIU-016.
- **R5 — data/channel drift.** The redesign must not change which files/counts/outcomes are emitted, nor stdout/stderr routing. Mitigation: characterization tests captured BEFORE any edit (M1 pre-flight), asserted unchanged after (M4).
- **R6 — cross-SPEC reversal of SPEC-CLI-TUX-V3-004 REQ-TUX4-006 (B2).** That SPEC retired the logo and added three tests asserting its absence in `PrintBanner` output — `TestCompactBanner_NoASCIILogo` (grep for `██ ╔ ╗ ╚ ╝ ═`), `TestCompactBanner_TwoLineIdentity` (≤2 non-empty lines), `TestCompactBanner_GlyphWhitelist` (all in `internal/cli/uikit/banner_compact_test.go`). Restoring the logo in `PrintBanner` reverses their premise → all three go RED. Mitigation: re-target them to `bannerString` DIRECTLY (which stays logo-free — see §A.1 L3), so `TestCompactBanner_*` assert the COMPACT BAND is still ≤2 lines / logo-free / status-glyph-only, and add a NEW test asserting `PrintBanner` (the composed surface) DOES carry the logo. This preserves each test's ORIGINAL intent (compact band stays compact) at the `bannerString` layer while the logo lives at the `PrintBanner` composition layer. `TestPrintBanner_*` (`banner_test.go`) also assert `PrintBanner` output and need logo-aware updates (RED-first). This reversal is documented in spec.md §A + HISTORY per B2.
- **R7 — status-glyph whitelist vs decorative logo runes.** The logo's block/box-drawing runes (`█ ╗ ╔ ╚ ╝ ═ ║`) fall OUTSIDE the AC-CLI-TUI-017 status-glyph whitelist (✓ ✗ ! · ● ○ ◆ ◇), which `TestCompactBanner_GlyphWhitelist` enforces over the full `PrintBanner` output. Without a carve-out, REQ-TUXIU-004 (whitelist) and REQ-TUXIU-050 (logo) CONTRADICT. Resolution (REQ-TUXIU-056): the whitelist governs SEMANTIC STATUS glyphs only; the logo art is a SEPARATE decorative category exempt from it (the runes predate the whitelist). Mitigation: scope `TestCompactBanner_GlyphWhitelist` to the compact-band portion (or re-target to `bannerString`), so the whitelist assertion no longer sees the logo's decorative runes; the whitelist still binds the checklist/spinner glyphs.
- **R8 — explicit `moai --help` fang injection (RESOLVED at kickoff → approach (a); residual risk = predicate mis-detection).** The load-bearing finding stands: fang v2.0.1 (`charm.land/fang/v2 v2.0.1`) exposes NO header/pre-help option — verified plan-phase and re-verified at kickoff: the only `fang.With*` options are `WithoutCompletions`, `WithoutManpage`, `WithColorSchemeFunc`, `WithTheme`, `WithVersion`, `WithoutVersion`, `WithCommit`, `WithErrorHandler`, `WithNotifySignal`. The explicit `moai --help` / `moai help` path skips `rootCmd.Run` (cobra shortcuts to the help func, which is fang's under `fang.Execute`), so there is no in-fang hook. **Resolution: approach (a)** — a root-help predicate on the `runFang` entry seam prints `tui.Logo(th)` BEFORE `fang.Execute(...)`, through the same Printer/stdout channel as `PrintBanner`. `go.mod` is unchanged (REQ-TUXIU-043 holds); approaches (b) no-args-only and (c) fang upgrade/fork are rejected (§A.1 L4). The fang-capability risk is therefore closed.
  - **Residual risk — predicate mis-detection.** What remains is not a fang risk but a predicate-correctness risk in the two directions the predicate can be wrong: (i) a **subcommand-help invocation wrongly matching** (`moai help init` or `moai init --help` printing a logo it must not), and (ii) the **no-args path double-printing** (the predicate matching an empty `os.Args[1:]`, so `PrintBanner`'s logo inside `rootCmd.Run` and the pre-`fang.Execute` logo both fire on `moai`).
  - **Mitigation:** RED-first unit tests over the predicate covering all 6 arg shapes in the §A.1 L4 predicate table — `[]` / `["--help"]` / `["-h"]` / `["help"]` / `["help", <sub>]` / `[<sub>, ...]` — asserting matched for exactly the three root-help shapes and NOT matched for the empty list and both subcommand-help shapes. The end-to-end double-print guard (no-args logo count == 1) is asserted by AC-TUXIU-024 in M4.

## §C Pre-flight

- [ ] **(FIRST TASK — pre-M1 golden capture, D4)** Capture the characterization baseline BEFORE any edit: run `moai init` and `moai update` and capture **both stdout AND stderr** to golden fixtures for TTY, non-TTY, and `NO_COLOR=1` invocations. Fixture paths: `internal/cli/testdata/tuxiu/{init,update}.{tty,notty,nocolor}.{stdout,stderr}.golden`. (No JSON/structured branch to capture — D2.) These goldens are the reference AC-TUXIU-016 asserts against after M4.
- [ ] Confirm the glyph-source decision (§A.1 item 1) at Implementation Kickoff Approval.
- [x] **(logo, R8) Kickoff decision RESOLVED** — the explicit `moai --help` / `moai help` fang-body logo injection is **IN SCOPE** via approach (a) (root-help predicate + pre-`fang.Execute` print; §A.1 L4, §B R8). No open plan item remains; `go.mod` unchanged. Before M3, re-confirm the `runFang` entry seam and the `trivialCommands` token map still live at `internal/cli/fang.go` / `root.go:43-51` (content-token anchor — line numbers drift).
- [ ] Confirm the update summary card layout matches the approved demo (§A.1 item 2).
- [ ] Enumerate every current glyph-declaration site (already mapped: `tui/status.go`, `tui/progress_line.go`, `printer.go` consts, `uikit/styles.go`, `uikit/status.go`).
- [ ] **(logo, v0.1.2)** Confirm the exact 6-line art against `git show 77893579e^:internal/cli/uikit/banner.go` (the pre-removal `moaiBanner` const) — the restored const MUST be byte-identical to the retired one.
- [ ] **(logo)** Enumerate the reversal test sites requiring reconciliation (already mapped): `TestCompactBanner_NoASCIILogo`, `TestCompactBanner_TwoLineIdentity`, `TestCompactBanner_GlyphWhitelist` (`banner_compact_test.go`) + `TestPrintBanner_OutputFormat`/`_WithVersion`/`_OptionalLeadingV`/`_EmptyVersion` (`banner_test.go`) + `stdout_clean_test.go` (channel invariant). The M1.0 golden captures the PRE-logo baseline; the logo header is EXPECTED NEW output on the 3 surfaces (data-line subset stays invariant — AC-TUXIU-016 logo note).

## §D Constraints

- No new module dependency (`go.mod` unchanged) — REQ-TUXIU-043.
- Theme SSOT: no hex literal outside `internal/tui/` — REQ-TUXIU-040.
- NO_COLOR / non-TTY degrade to plain text; pills → "[label]" — REQ-TUXIU-041.
- MOAI_REDUCED_MOTION static fallback preserved — REQ-TUXIU-042.
- Presentation-only: data + stdout/stderr channel discipline unchanged — REQ-TUXIU-044. (JSON byte-identical is N/A — no machine-readable surface, REQ-TUXIU-045.)
- Glyphs stay in the AC-CLI-TUI-017 whitelist — REQ-TUXIU-004.
- ≥90% coverage on NEW/touched render functions + no `internal/cli` whole-package regression vs the 74.6% baseline — REQ-TUXIU-046 (D3).
- **(logo)** Gradient ramp derives ONLY from `Theme.Accent`/`AccentDeep` via `tui.CoralRamp` (`lipgloss.Blend1D`); no hex literal in `internal/tui/logo.go`; no new hex outside `internal/tui/` — REQ-TUXIU-051/040.
- **(logo)** `bannerString` is NOT modified; the logo stacks in `PrintBanner` only — the compact-band contract is preserved and the reversal is minimized (§A.1 L3, §B R6).
- **(logo)** `go.mod` unchanged — `Blend1D` is from the already-present `charm.land/lipgloss/v2` (REQ-TUXIU-043).
- **(logo)** The logo's decorative block/box-drawing runes are exempt from the status-glyph whitelist (REQ-TUXIU-056) — the whitelist assertion is scoped to status glyphs / the compact band, not the logo art.

## §E Self-Verification

Plan/run/sync audit-ready signals are recorded in `progress.md` (§E.1–§E.4). This agent populated §E.1 (plan-phase) only; run-phase (§E.2/§E.3) is owned by manager-develop and sync-phase (§E.4) by manager-docs.

## §F Milestones (build-dependency order; reversibility-tagged)

### M1 — Shared symbol source + spinner-residue fix (foundation)
Reversibility: **HIGH** (new type interface — §A.1 item 1) + **BEHAVIOUR** (spinner clear — §A.1 item 4). Review first.
- **M1.0 (pre-flight, D4):** capture the golden stdout+stderr baseline (per §C first task) BEFORE any code edit — this is the reference AC-TUXIU-016 asserts against.
- Establish the exported `tui.Glyph*` const block (option (b), §A.1 item 1); refactor `tui.StatusIcon` to RETURN those constants (single source, not two).
- Re-point every external raw-rune site to the const block: `printer.go` markers (280,282), `tui/progress_line.go` sym* (179-201), `uikit/styles.go` (25,28,31), `uikit/status.go` (11,13,15), AND `uikit/render.go:72` (`SuccessStyle.Render("✓")` — added per D5). Runes byte-identical; ZERO raw-rune redeclarations remain outside the const block.
- Fix spinner residue: each finished step in `progress_line.go` (41, 85, 179-201) + `printer.go` (411-595) clears its spinner line and prints exactly one ✓ result line; preserve the TTY/non-TTY split.
- **(logo, v0.1.2 — §A.1 L1/L2) NEW tui primitive `internal/tui/logo.go`:** restore the verbatim 6-line `moaiBanner` const (byte-identical to `77893579e^`), add `tui.CoralRamp(n int) []color.Color` (= `lipgloss.Blend1D(n, th.Accent, th.AccentDeep)`) and `tui.Logo(theme tui.Theme) string` (per-line gradient render, top-light→bottom-deep; NO_COLOR → plain, non-TTY → solid downsample). Colour sourced ONLY from `Theme` tokens — zero hex literal in this file. This is a foundation primitive (same HIGH-reversibility class as the glyph SSOT), so it lands in M1 for review-first.
- Covers: REQ-TUXIU-001..004, 020..022, **050..053, 056**.
- Gate: RED tests for one-clean-line-per-step (TTY + non-TTY); RED tests for `tui.Logo`/`tui.CoralRamp` (6 lines, ≥3 distinct fg on truecolor, plain under NO_COLOR); full `go test ./...` green after re-pointing.

### M2 — update.go presentation wiring
Reversibility: **HIGH** (user-facing UX — §A.1 items 2,3). Review second.
- Card-style classification summary: `tui.Box{Accent:true}` + 3 count pills (PillOk/PillInfo/PillErr), zero-count pills omitted.
- Unified `tui.StatusIcon` glyphs on the 5 deploy steps (update.go:740-839); ●running/✓done/○pending states.
- Block progress bar via `tui.Progress` connected to the step-count signal (update.go:1005 `StepUpdate`) replacing/augmenting "N/M steps complete".
- Identity header band (update.go:700,1081) with solid `PillPrimary` version pill.
- Outcome banner: solid `PillOk` "✓ Updated N files" + dim note (update.go outcome sites 281..1089).
- All rendering is human-readable output; there is no JSON/structured branch to route around (D2). Preserve the stdout-vs-stderr channel partition (printer gateway) unchanged.
- Covers: REQ-TUXIU-010..016, 040, 041, 044. (045 is N/A — no machine-readable surface.)

### M3 — init banner + success card + large logo placement
Reversibility: **HIGH** (the logo placement is the most visible new element and reverses a prior SPEC — §A.1 L3) + **MEDIUM** (init card mirrors M2's established language). Review with M1's foundation.
- `uikit/banner.go` (105,125): ◆ MoAI-ADK identity band + version/go/claude pills via shared primitives.
- `init_warnings.go:88` `buildInitSuccessCard`: shared card + pill visual language.
- init.go:410-411 call sites unchanged in behaviour.
- **(logo, §A.1 L3) `PrintBanner` restore:** compose `tui.Logo(th)` stacked ABOVE the existing compact `bannerString` and route both through `Printer.Data` → stdout — one edit covers all 3 surfaces (root no-args `root.go:32`, `moai init` `init.go:410`, `moai update` `update.go:1253`). **Do NOT modify `bannerString`** (reversal-minimizing — §B R6).
- **(logo reversal reconciliation, §B R6/R7)** Re-target `TestCompactBanner_NoASCIILogo`/`_TwoLineIdentity`/`_GlyphWhitelist` to `bannerString` (compact band stays logo-free / ≤2 lines / status-glyph-only), add a NEW test asserting `PrintBanner` carries the logo, update `TestPrintBanner_*` logo-aware (RED-first). Scope the glyph-whitelist test to the band portion so the logo's decorative runes are exempt (REQ-TUXIU-056).
- **(logo, §A.1 L4 / §B R8 — RESOLVED, in scope)** Explicit `moai --help` / `moai help` fang-body injection via approach (a): wire the root-help predicate + pre-`fang.Execute` logo print on the `runFang` seam in `internal/cli/fang.go` (token map reused from `root.go` `trivialCommands`). The predicate inspects ONLY `os.Args[1:]`, matches the three root-help shapes (`["--help"]` / `["-h"]` / `["help"]` with len == 1) and MUST NOT match the empty arg list (no-args is already covered by `PrintBanner` in `rootCmd.Run` — matching it double-prints) nor any subcommand-help shape (`["help", <sub>]`, `[<sub>, ...]`). The logo routes through the same Printer gateway / stdout channel as `PrintBanner`. **RED-first predicate unit tests** cover all 6 arg shapes of the §A.1 L4 table. No fang change, no `go.mod` change.
- Covers: REQ-TUXIU-030, 031, 040, 041, **054, 055**.

### M4 — Verification and coverage (mechanical; review last)
Reversibility: **LOW** (mechanical verification).
- Characterization: assert data/counts/outcomes byte-identical to the M1 golden stdout+stderr baseline — ANSI-stripped structural diff of the file/count/outcome lines empty AND the stdout-vs-stderr partition unchanged (AC-TUXIU-016). No JSON branch to diff (D2).
- NO_COLOR/monochrome degradation + pill "[label]" fallback (REQ-TUXIU-041).
- MOAI_REDUCED_MOTION static fallback preserved (REQ-TUXIU-042).
- Theme-SSOT grep guard: no NEW hex literal outside `internal/tui/` beyond the 1 pre-existing comment baseline at `wizard/styles.go:20` (REQ-TUXIU-040, D9).
- Glyph-whitelist grep (REQ-TUXIU-004); `go.mod` unchanged (REQ-TUXIU-043).
- Coverage (D3): ≥90% on the NEW/touched render functions via per-function extraction (`go test -coverprofile=cover.out ./internal/cli/... ./internal/tui/... && go tool cover -func=cover.out` filtered to the new render funcs, INCLUDING `tui.Logo`/`tui.CoralRamp`); assert no `internal/cli` whole-package regression vs the 74.6% baseline (REQ-TUXIU-046).
- **(logo)** Logo present on all 3 surfaces — ANSI-stripped grep for `███╗   ███╗` + retained `◆ MoAI-ADK` band (AC-TUXIU-020).
- **(logo, R8/L4)** Root-help predicate end-to-end verification (AC-TUXIU-024): on ANSI-stripped captures, assert the logo signature `███╗   ███╗` **IS present** on `moai --help` and `moai help`; **IS ABSENT** on `moai init --help` and `moai help init`; and appears **exactly ONCE** (occurrence count == 1, not ≥ 1 — this is the double-print guard) on no-args `moai`.
- **(logo)** Per-line gradient ≥3 distinct fg codes top-light→bottom-deep on a truecolor capture (AC-TUXIU-021); NO_COLOR → 0 SGR colour + runes intact, non-TTY → ≤1 distinct fg (AC-TUXIU-022).
- **(logo)** Theme-SSOT guard: `internal/tui/logo.go` has zero hex literals; AC-TUXIU-012 grep still = 1 baseline (no NEW `internal/cli` hex) — AC-TUXIU-023.
- **(logo)** Golden update: capture FRESH init/update goldens WITH the logo as the post-M4 expected PRESENTATION baseline; the file/count/outcome DATA-line subset still diffs empty vs the M1.0 PRE-logo baseline (AC-TUXIU-016 logo note).
- Covers: REQ-TUXIU-004, 040..044, 046, **051, 052, 053, 055, 056** (045 N/A — no machine-readable surface).

## §G Anti-Patterns to Avoid

- **Redefining a glyph rune in a fourth/fifth place** instead of pulling from the canonical source (defeats M1's purpose).
- **Emitting a hex colour literal** in `update.go`/`uikit` instead of a `Theme` token.
- **Rendering into the JSON branch** — new pills/cards belong strictly in the human-readable path.
- **Changing emitted data** ("while I'm here, tweak the file count") — this is presentation-only.
- **Dropping the non-TTY split** — the erase-line fix must keep the plain newline path for non-TTY.
- **Adding a dependency** to draw a box/bar the existing `tui` already draws — likewise for the logo gradient (`Blend1D` is already in `lipgloss/v2`; do NOT add a colour library).
- **(logo) Putting the logo art or its colour math in `uikit`** instead of `internal/tui/logo.go` — leaks hex/colour decisions outside the SSOT boundary (§A.1 L1).
- **(logo) Hand-rolling RGB interpolation** instead of `lipgloss.Blend1D` — more code, worse ramp (§A.1 L2).
- **(logo) Modifying `bannerString` to embed the logo** — breaks the reversal-minimizing design; the compact band must stay logo-free and the logo stacks only in `PrintBanner` (§A.1 L3, §B R6).
- **(logo) Double-printing the logo on no-args `moai`** — letting the root-help predicate match an EMPTY `os.Args[1:]`. The no-args surface already prints the logo via `PrintBanner` inside `rootCmd.Run`; a predicate that also matches `[]` fires twice on the single most-visible surface. The empty-list exclusion is a HARD invariant (§A.1 L4), guarded by the occurrence-count-==-1 assertion in AC-TUXIU-024.
- **(logo) Adding the block/box-drawing runes to the status-glyph whitelist** — they are a SEPARATE decorative category (REQ-TUXIU-056); scope the whitelist test instead of widening the status-glyph vocabulary.

## §H Cross-References

- `spec.md` — GEARS requirements (REQ-TUXIU-001..056; Group F = logo).
- `acceptance.md` — Given-When-Then AC matrix (AC-TUXIU-001..024) + Definition of Done.
- `SPEC-V3R3-CLI-TUI-001` — Theme SSOT (REQ-CLI-TUI-013), NO_COLOR degradation (REQ-CLI-TUI-009), glyph whitelist (AC-CLI-TUI-017).
- `SPEC-CLI-TUX-V3-004` — REQ-TUX4-006 retired the large logo (commit `77893579e`); Group F reverses that (B2 cross-SPEC conflict documented). Retirement tests reconciled in M3.
- `SPEC-CLI-TUX-V3-005` — recent Printer migration / ratchet precedent.
- `charm.land/lipgloss/v2` `Blend1D` (`blending.go:18`) — the CIELAB gradient primitive backing `tui.CoralRamp` (already a `go.mod` dependency).
- CLAUDE.local.md §6 — critical-package coverage context (re-scoped per D3 to per-function ≥90% + no `internal/cli` whole-package regression vs the 74.6% baseline); §2 Template-First (N/A — these are runtime CLI files, not template sources).
