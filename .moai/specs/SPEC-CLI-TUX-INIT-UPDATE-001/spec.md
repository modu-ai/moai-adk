---
id: SPEC-CLI-TUX-INIT-UPDATE-001
title: "moai init/update terminal output TUX redesign (init + update presentation modernization)"
version: "0.1.3"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: Medium
phase: "v3.1.0"
module: "internal/cli, internal/tui, internal/cli/printer"
lifecycle: spec-anchored
tags: "cli, tui, tux, presentation, theme, init, update, charm, lipgloss"
related_specs: [SPEC-V3R3-CLI-TUI-001, SPEC-CLI-TUX-V3-004, SPEC-CLI-TUX-V3-005]
tier: M
---

# SPEC-CLI-TUX-INIT-UPDATE-001 — moai init/update Terminal Output TUX Redesign

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-25 | manager-spec | Initial draft. Plan-phase authoring: modernize `moai init` and `moai update` terminal output by connecting under-used `internal/tui` primitives. Presentation-only; no functional/data change. Follow-up to SPEC-V3R3-CLI-TUI-001 (theme system origin). |
| 0.1.1 | 2026-07-25 | manager-spec | Folded plan-auditor findings D1-D9: resolved glyph-source to option (b) — exported `tui.Glyph*` const block as raw-rune SSOT with `StatusIcon` refactored to reference it (D1, NEEDS-CLARIFICATION marker removed); marked REQ-045/AC-017 JSON preservation **N/A** (no machine-readable output surface, verified, D2); re-scoped coverage AC-019 to per-function ≥90% on touched render funcs + no `internal/cli` whole-package regression vs the 74.6% baseline (D3); pinned AC-016 golden stdout/stderr characterization Verify (D4); strengthened glyph-SSOT AC-004 to zero-raw-rune-remaining + added missed `uikit/render.go:72` site (D5); downgraded Tier L→M (D6); added explicit Verify lines to AC-007/013a/013b/014 (D7); reconciled REQ-001 vs `preview_tui.go` §C carve-out (D8); pinned AC-012 hex baseline = 1 comment hit at `wizard/styles.go:20` (D9). |
| 0.1.2 | 2026-07-25 | manager-spec | User-approved additive scope: **restore the large 6-line "MoAI-ADK" ASCII-art logo** (retired by SPEC-CLI-TUX-V3-004 REQ-TUX4-006, commit `77893579e`) with a vertical coral gradient. Added **Group F** (REQ-TUXIU-050..056): logo restored as a `tui` primitive (`tui.Logo`/`tui.CoralRamp` in `internal/tui/logo.go`); gradient ramp derived from `Theme.Accent`→`Theme.AccentDeep` via `lipgloss.Blend1D` (zero new dependency — already in go.mod); per-line top-light→bottom-deep; logo on the 3 PrintBanner surfaces (root no-args / init / update) stacked ABOVE the retained compact band; NO_COLOR/non-TTY degradation; explicit `moai --help` fang-body injection flagged as a plan open-item (`[NEEDS CLARIFICATION]`, plan.md §B R8 — fang v2.0.1 exposes no header option); status-glyph-whitelist carve-out (REQ-056) for the logo's decorative block/box-drawing runes. Added AC-TUXIU-020..023 (Look item 8). Cross-SPEC reversal of REQ-TUX4-006 documented (B2); the 3 SPEC-CLI-TUX-V3-004 retirement tests require run-phase reconciliation. Tier unchanged (M). |
| 0.1.3 | 2026-07-25 | manager-spec | Folded the Implementation Kickoff resolution of the single remaining open item (plan.md §A.1 L4 / §B R8): the explicit `moai --help` / `moai help` fang-body logo injection is **user-approved IN SCOPE via approach (a)** — a root-help predicate on the `runFang` seam inspecting only `os.Args[1:]`, printing `tui.Logo` BEFORE `fang.Execute` through the same Printer/stdout channel as `PrintBanner`. Matched shapes: `["--help"]` / `["-h"]` / `["help"]` (len == 1); NOT matched: the empty arg list (HARD no-double-print invariant — no-args is already covered by `PrintBanner` in `rootCmd.Run`) and every subcommand-help shape (`moai help init`, `moai init --help`). Rejected: (b) no-args-only (leaves the most common discovery path logo-free), (c) fang upgrade/fork (breaches the no-new-dependency envelope, REQ-TUXIU-043). Re-verified at kickoff that fang v2.0.1 exposes no header/pre-help option → **`go.mod` unchanged**. Promoted REQ-TUXIU-055 from a capability-gated open item to a normal in-scope GEARS requirement; removed the explicit-`--help` entry from §C Out of Scope; added AC-TUXIU-024 (MUST) asserting logo presence on `moai --help`/`moai help`, absence on `moai init --help`/`moai help init`, and occurrence count **exactly 1** on no-args `moai`. The NEEDS-CLARIFICATION marker is removed from plan.md — zero open items remain. Tier unchanged (M); status remains `draft` (the `draft → in-progress` transition is manager-develop's on the M1 commit). |

## §A Context

`moai init` and `moai update` render their terminal output through `internal/cli/update.go` (~1861 lines of inline `fmt.Fprintln(out, tui.X(...))` calls), the live animated `internal/cli/printer/printer.go` gateway, and the `internal/cli/uikit` banner/card helpers. The full MoAI terminal design system v2 already exists in `internal/tui/` — `Box`/`ThickBox`, `Pill`, `StatusIcon`/`Spinner`/`Progress`/`Stepper`, `ProgressLine`, and a 28-token `Theme` (Light/Dark/Monochrome) — but the two commands under-use it: the classification summary is plain text, the deploy progress shows only "N/M steps complete" with no bar, checklist glyphs are redefined per-file, and finished steps leave stale two-part spinner residue on screen.

The user reviewed and approved a live demo of the target look. This SPEC connects the already-existing primitives to the two screens. It is a **presentation-only** modernization: the DATA emitted (which files, counts, outcomes) and the stdout/stderr channel discipline (the printer gateway) are unchanged. (`moai init`/`moai update` have **no** machine-readable/JSON output surface — verified: no `--json`/`--format`/`--output` flag is declared in `init.go`/`update.go`, and the `encoding/json` usage in `update.go` is settings-file I/O, not command output — so there is no structured-output path to preserve; see REQ-TUXIU-045 **N/A**.)

**Ground-truth investigation (verified during plan-phase):**
- `internal/tui/` primitives all exist and are exported: `Box`/`ThickBox` (box.go:82,91), `Pill` (pill.go:77), `StatusIcon`/`Spinner`/`Progress`/`Stepper` (status.go:18,51,91,128), `Section` (table.go:112), `ProgressLine` (progress_line.go:85), `LightTheme`/`DarkTheme`/`MonochromeTheme` (theme.go:105,146,183).
- `BoxOpts.Accent bool` exists ("applies the accent colour to the border and background").
- `PillOpts.Kind` enum: `PillInfo`/`PillOk`/`PillWarn`/`PillErr`/`PillPrimary`/`PillNeutral`; `PillOpts.Solid bool`.
- `Theme` carries 28 colour tokens; `Pill`/`Progress`/`StatusIcon` source all colour from it (no hex literals in those files — REQ-CLI-TUI-013 already holds).
- The ✓/✗/○/● glyph set is redeclared across the `tui`/`printer`/`uikit` packages: `tui.StatusIcon` (status.go:18-35), `tui.progress_line.go` (symProgress/symSuccess/symError, 179-201), `printer.go` consts (`stepMarker="○"`, `spinnerFrame="⠋"`, `spinnerStatic="●"` at 280,281,282), `uikit/styles.go` (`SymSuccess`/`SymError`/`SymWarning`, 25,28,31), `uikit/status.go` (11,13,15), and `uikit/render.go:72` (`SuccessStyle.Render("✓")` — this site was missed in the original four-place enumeration; added per D5). The `tui`-internal sites (StatusIcon + progress_line sym*) become references to the new const block; the external raw-rune declarations at `printer.go:280,282` / `uikit/styles.go:25,28` / `uikit/status.go:11,15` / `uikit/render.go:72` are removed and replaced by references (verified locations, plan-phase).
- The spinner residue lives in `progress_line.go` (ANSI erase-line `\r\x1b[2K` at line 41, TTY/non-TTY split) and `printer.go` (eraseLine const at 279, stepHandle at 411-467, spinnerHandle at 475-595).
- `quality.yaml` `development_mode: tdd` → run-phase uses RED-GREEN-REFACTOR.
- System version is `v3.0.1` (matches the demo header band "◆ MoAI-ADK v3.0.1").

**Inherited invariants** (from the completed SPEC-V3R3-CLI-TUI-001, re-encoded below as this SPEC's own requirements): Theme SSOT (REQ-CLI-TUI-013), NO_COLOR/monochrome degradation (REQ-CLI-TUI-009), the AC-CLI-TUI-017 Unicode glyph whitelist (✓ ✗ ! · ● ○ ◆ ◇ — this whitelist governs semantic **status** glyphs; the restored logo's decorative block/box-drawing runes are a separate category, carved out by REQ-TUXIU-056), and MOAI_REDUCED_MOTION static fallback.

**Logo restoration (v0.1.2 additive scope).** The large 6-line "MoAI-ADK" ASCII-art banner (`const moaiBanner`) was retired in commit `77893579e` (SPEC-CLI-TUX-V3-004 REQ-TUX4-006, "compact banner") in favour of the compact `◆ MoAI-ADK` identity band. The user has approved restoring the large logo. This is a deliberate **partial reversal of REQ-TUX4-006** (documented per the cross-SPEC policy-conflict pre-scan discipline). The restoration is additive-presentation: the logo is rendered ABOVE the retained compact band (both stack), and the retained band keeps its own contract.

**Ground truth verified for the logo (plan-phase):**
- The exact 6-line art matches the pre-removal `moaiBanner` const (`git show 77893579e^:internal/cli/uikit/banner.go`) verbatim — block runes `█` (U+2588) + box-drawing `╗ ╔ ╚ ╝ ═ ║`.
- The gradient primitive already exists in the ecosystem: `charm.land/lipgloss/v2` exposes `Blend1D(steps int, stops ...color.Color) []color.Color` (`blending.go:18`) — a CIELAB linear ramp. `charm.land/lipgloss/v2 v2.0.5` is already a direct `go.mod` dependency, so the ramp adds **no new module dependency** (REQ-TUXIU-043 holds).
- Ramp endpoints are theme accent tokens: `Theme.Accent` (light `#bf6547` / dark `#d97757`, the lighter coral) → `Theme.AccentDeep` (light `#a84f33` / dark `#b85e3f`, the deeper coral). `Blend1D(6, Accent, AccentDeep)` yields top-light→bottom-deep, matching the approved direction. Both tokens live in `internal/tui/theme.go` — no hex literal leaves `internal/tui/`.
- The 3 target surfaces route through exactly ONE entry — `uikit.PrintBanner(version)` — called at `root.go:32` (root, no-args `moai`), `init.go:410` (`moai init`), and `update.go:1253` (`moai update`). No other `PrintBanner` caller exists, so a single stacking edit at `PrintBanner` covers all three surfaces.
- Root-help injection is asymmetric: the **no-args `moai`** path fires `rootCmd.Run` (`root.go:31-34`) which calls `uikit.PrintBanner(...)` BEFORE `cmd.Help()`, so the logo naturally prints above the fang-rendered help body — no fang change required. The **explicit `moai --help` / `moai help`** path does NOT run `rootCmd.Run` (cobra shortcuts to the help func, which is fang's under `fang.Execute`), so a logo there would need a fang injection point. fang v2.0.1 exposes NO header/pre-help option (verified: only `WithoutCompletions`/`WithoutManpage`/`WithColorSchemeFunc`/`WithTheme`/`WithVersion`/`WithoutVersion`/`WithCommit`/`WithErrorHandler`/`WithNotifySignal`). The explicit-`--help` injection is therefore wired OUTSIDE fang: **resolved at Implementation Kickoff to approach (a)** — a root-help predicate on the `runFang` entry seam (`internal/cli/fang.go` / `internal/cli/root.go`) inspecting only `os.Args[1:]`, printing the logo BEFORE `fang.Execute(...)` through the same Printer gateway / stdout channel as `PrintBanner`. The predicate matches `["--help"]` / `["-h"]` / `["help"]` (len == 1) and deliberately does NOT match the empty arg list (no-args is already covered by `rootCmd.Run` → double-print hazard) nor any subcommand-help shape (`moai help init`, `moai init --help`). No fang change and no `go.mod` change (REQ-TUXIU-043 holds). See REQ-TUXIU-055 + plan.md §A.1 L4 / §B R8.
- **Cross-SPEC reversal impact (B2):** SPEC-CLI-TUX-V3-004 added three tests that assert the logo's absence in `uikit.PrintBanner` output: `TestCompactBanner_NoASCIILogo`, `TestCompactBanner_TwoLineIdentity` (≤2 non-empty lines), and `TestCompactBanner_GlyphWhitelist` (all in `internal/cli/uikit/banner_compact_test.go`). Restoring the logo in `PrintBanner` reverses their premise; they require run-phase reconciliation (re-target to `bannerString` directly, or scope to the compact-band portion). The design that MINIMIZES the reversal keeps `bannerString` itself logo-free and stacks the logo only at the `PrintBanner` composition layer — so the "compact band stays compact" intent survives at the `bannerString` level. `TestPrintBanner_*` (`banner_test.go`) also assert `PrintBanner` output and need logo-aware updates (RED-first).

## §B Requirements (GEARS)

### Group A — Single symbol source (glyph SSOT)

- **REQ-TUXIU-001** (Ubiquitous): The `tui` package **shall** expose one canonical source for the status-glyph set — ✓ (done/ok), ● (run/in-progress), ○ (skip/pending), ✗ (error) — realized as an exported raw-rune constant block (`tui.GlyphDone`/`GlyphRun`/`GlyphSkip`/`GlyphErr`; see plan §A.1), and every render site that emits one of these glyphs **shall** resolve it from that source. The interactive change-preview flow `internal/cli/update/preview_tui.go` is **exempt** (per §C Out of Scope — the Bubble Tea change-preview flow); it currently declares no ✓/✗/○/● literal, so this exemption leaves no glyph render site unreconciled.
- **REQ-TUXIU-002** (Ubiquitous): The `printer` package **shall** source its step marker and spinner-static marker from the canonical `tui` glyph source rather than redeclaring literal glyph constants.
- **REQ-TUXIU-003** (Ubiquitous): The `uikit` package **shall** source its success/error/warning glyphs from the canonical `tui` glyph source rather than redeclaring them.
- **REQ-TUXIU-004** (Ubiquitous): The consolidated glyph set **shall** remain within the existing AC-CLI-TUI-017 Unicode whitelist (✓ ✗ ! · ● ○ ◆ ◇); no new emoji-range codepoint **shall** be introduced.

### Group B — Update-screen presentation

- **REQ-TUXIU-010** (Event-driven): **When** `moai update` has computed the merge classification, the command **shall** render the summary as an accent-bordered card (`tui.Box` with `Accent: true`) carrying three semantic count pills — `PillOk` "+ N add", `PillInfo` "~ N update", `PillErr` "! N conflict".
- **REQ-TUXIU-011** (State-driven): **While** a classification count is zero, the command **shall** omit that count's pill from the summary card, so a clean run shows no "! 0 conflict" noise.
- **REQ-TUXIU-012** (Ubiquitous): The update checklist steps **shall** use a single `tui.StatusIcon` glyph set — ✓ (Success colour), ● (Accent colour, bold), ○ (Faint colour) — with no per-step glyph redefinition.
- **REQ-TUXIU-013** (State-driven): **While** a deploy step is in progress the command **shall** render its line with ● (Accent, bold); **when** the step completes the command **shall** render ✓ (Success); **while** the step is pending the command **shall** render ○ (Faint).
- **REQ-TUXIU-014** (Ubiquitous): The deploy-step progress **shall** render a block progress bar via `tui.Progress` (██████░░░░ style) reflecting completed/total steps, connected to the existing step-count signal that today renders only "N/M steps complete" text.
- **REQ-TUXIU-015** (Ubiquitous): The command **shall** render an identity header band "◆ MoAI-ADK <version> <go-runtime> · claude" with the version rendered as a solid brand pill (`tui.Pill`, `PillPrimary`, `Solid: true`).
- **REQ-TUXIU-016** (Event-driven): **When** `moai update` completes, the command **shall** render the outcome as a solid success pill ("✓ Updated N files") followed by a dim detail note.

### Group C — Spinner-residue removal

- **REQ-TUXIU-020** (Event-driven): **When** a step's spinner finishes, the command **shall** clear the spinner line via the ANSI erase-line sequence and print exactly one clean result line (✓ + message).
- **REQ-TUXIU-021** (Unwanted behavior): The command **shall not** leave a stale or duplicated two-part spinner line ("○ Validating template…  ○ Removing …") on screen after a step completes.
- **REQ-TUXIU-022** (State-driven): **While** output is not a TTY, the progress line **shall** degrade to a single newline-terminated result line without ANSI erase sequences (preserving the existing TTY/non-TTY split).

### Group D — Init-screen presentation

- **REQ-TUXIU-030** (Ubiquitous): The init banner (`internal/cli/uikit/banner.go`) **shall** render the same ◆ MoAI-ADK identity band and version/go/claude pills as the update header band, sourced from the shared `tui` primitives.
- **REQ-TUXIU-031** (Event-driven): **When** `moai init` succeeds, the init success card (`internal/cli/init_warnings.go` `buildInitSuccessCard`) **shall** use the shared card + pill visual language (`tui.Box`, `tui.Pill`).

### Group E — Invariants and constraints

- **REQ-TUXIU-040** (Ubiquitous — Theme SSOT, inherits REQ-CLI-TUI-013): No hex colour literal **shall** appear outside `internal/tui/`; every colour decision **shall** flow through the 28-token `Theme`.
- **REQ-TUXIU-041** (State-driven — inherits REQ-CLI-TUI-009): **While** `NO_COLOR` is set OR output is not a TTY, the command output **shall** degrade cleanly to plain text (lipgloss downsample), and pills **shall** degrade to "[label]".
- **REQ-TUXIU-042** (State-driven): **While** `MOAI_REDUCED_MOTION` is set, the existing static-fallback behaviour (static ● spinner, fully-filled progress bar) **shall** be preserved.
- **REQ-TUXIU-043** (Ubiquitous — no new dependency): The redesign **shall** reuse only the existing `charm.land/*` v2 stack and `internal/tui`; no new module dependency **shall** be added to `go.mod`.
- **REQ-TUXIU-044** (Ubiquitous — characterization/behaviour preservation): The DATA emitted (which files, counts, outcomes) and the stdout/stderr channel discipline (the printer gateway) **shall** be unchanged; this is a presentation-only redesign.
- **REQ-TUXIU-045** — **N/A** (no machine-readable output surface exists for init/update — nothing to preserve). `moai init` and `moai update` expose no `--json` / `--format` / `--output` flag and no structured-output mode (verified plan-phase: no such flag declared in `init.go`/`update.go`; the `encoding/json` in `update.go` is settings-file I/O, not command output). There is no machine-readable output path to keep byte-identical, so this requirement is not applicable. The stdout/stderr channel-discipline preservation that *does* apply is owned by REQ-TUXIU-044.
- **REQ-TUXIU-046** (Ubiquitous — coverage): The NEW/touched render functions introduced by this SPEC in `internal/cli` and `internal/tui` **shall** reach ≥90% statement coverage (per-function measurement), AND the redesign **shall not** regress `internal/cli` whole-package coverage below its measured baseline (74.6%; `internal/tui` baseline is 93%). The whole-package ≥90% target is NOT required for `internal/cli` — its measured baseline is already below 90% independent of this presentation-only SPEC.

### Group F — Large logo restoration

- **REQ-TUXIU-050** (Ubiquitous): The `tui` package **shall** restore the large 6-line "MoAI-ADK" ASCII-art logo (retired by SPEC-CLI-TUX-V3-004 REQ-TUX4-006, commit `77893579e`) as a first-class primitive — the verbatim 6-line art constant homed in `internal/tui/logo.go`, exposed via `tui.Logo(theme tui.Theme) string`. (Const-home decision: the art lives in `tui`, NOT `uikit` — plan.md §A.1 item L1 records the rationale.)
- **REQ-TUXIU-051** (Ubiquitous — Theme SSOT, inherits REQ-CLI-TUI-013): The vertical coral gradient ramp **shall** be derived exclusively from the tui `Theme` accent tokens (`Accent` → `AccentDeep`, plus interpolated stops) via a NEW `tui` helper `tui.CoralRamp(n int) []color.Color` (backed by `lipgloss.Blend1D`); no hex colour literal **shall** appear outside `internal/tui/`.
- **REQ-TUXIU-052** (Ubiquitous): The logo **shall** render with a per-line vertical gradient across the 6 rows — top light → bottom deep — each row painted with one successive ramp stop (row *i* ← `tui.CoralRamp(6)[i]`).
- **REQ-TUXIU-053** (State-driven — inherits REQ-CLI-TUI-009): **While** `NO_COLOR` is set the logo **shall** degrade to plain text (no ANSI colour); **while** output is a non-TTY pipe the logo **shall** degrade to a single solid accent colour (lipgloss downsample). The logo is static art — no animation is introduced, so `MOAI_REDUCED_MOTION` has no logo-specific effect (the existing spinner/progress reduced-motion fallback of REQ-TUXIU-042 is unchanged).
- **REQ-TUXIU-054** (Ubiquitous): The logo **shall** appear on the three `uikit.PrintBanner` surfaces — `moai` (root, no-args), `moai init`, and `moai update` — rendered ABOVE the existing compact identity band (`bannerString`); the compact band **shall** be retained and stacked directly below the logo (NOT replaced). (All three surfaces route through the single `PrintBanner` entry: `root.go:32`, `init.go:410`, `update.go:1253`; the no-args `moai` path prints the logo above the fang help body because `rootCmd.Run` calls `PrintBanner` before `cmd.Help()`.)
- **REQ-TUXIU-055** (Event-driven + Unwanted behavior — explicit root-help surface, IN SCOPE): **When** the user invokes an explicit root-help command — `moai --help`, `moai -h`, or `moai help` with no subcommand argument — the CLI **shall** render the restored logo above the fang-rendered help body, printed BEFORE `fang.Execute(...)` on the `runFang` entry seam and routed through the same Printer gateway / stdout channel as `uikit.PrintBanner` (so the stdout-vs-stderr partition of REQ-TUXIU-044 is unchanged). The root-help predicate **shall** derive its verdict solely from `os.Args[1:]`, and its matched/not-matched behaviour is exhaustively:

  | `os.Args[1:]` | Example | Logo printed by this requirement? |
  |---------------|---------|-----------------------------------|
  | `[]` | `moai` | **NO** — already covered by REQ-TUXIU-054 |
  | `["--help"]` | `moai --help` | **YES** |
  | `["-h"]` | `moai -h` | **YES** |
  | `["help"]` (len == 1) | `moai help` | **YES** |
  | `["help", <sub>]` | `moai help init` | **NO** — subcommand help |
  | `[<sub>, ...]` | `moai init --help` | **NO** — subcommand help |

  The predicate **shall not** match an empty `os.Args[1:]`, and the no-args `moai` invocation **shall** render the logo exactly once: the no-args path already prints it via `uikit.PrintBanner` inside `rootCmd.Run` (REQ-TUXIU-054), so a predicate matching the empty argument vector would emit the logo twice on the most-visible surface. This requirement is satisfied without any fang option, fang upgrade, or fork — `go.mod` remains unchanged (REQ-TUXIU-043). Design rationale, the rejected alternatives, and the predicate's residual mis-detection risk are recorded in plan.md §A.1 L4 / §B R8.
- **REQ-TUXIU-056** (Ubiquitous — status-glyph-whitelist carve-out): The restored logo's decorative block/box-drawing runes (`█ ╗ ╔ ╚ ╝ ═ ║`) are a SEPARATE decorative-art category, EXEMPT from the AC-CLI-TUI-017 status-glyph whitelist (✓ ✗ ! · ● ○ ◆ ◇) which governs only semantic STATUS markers. These runes predate the whitelist (they were in the original `moaiBanner`); restoring them re-introduces the decorative category and **shall not** be treated as a whitelist violation. The status-glyph whitelist (REQ-TUXIU-004) continues to bind the checklist/spinner glyphs unchanged; no emoji-range codepoint is introduced by the logo.

## §C Exclusions — Out of Scope

The following are explicitly **out of scope** for this SPEC. Each is expressed as an `### Out of Scope` sub-heading to satisfy the exclusions contract.

### Out of Scope — functional and data behaviour changes
- Changing which files `moai init`/`moai update` write, deploy, skip, or classify.
- Changing merge classification logic, conflict detection, or file-count arithmetic.
- Altering the stdout/stderr channel routing owned by the printer gateway.

### Out of Scope — machine-readable output
- `moai init`/`moai update` expose **no** machine-readable/JSON output surface (verified: no `--json`/`--format`/`--output` flag). This SPEC introduces none — adding a structured-output mode is out of scope (see REQ-TUXIU-045 N/A).

### Out of Scope — the Bubble Tea change-preview flow
- `internal/cli/update/preview_tui.go` (the interactive classification table + "[enter] view diff / [y] confirm" prompt) beyond, at most, a light glyph/pill touch to match the shared symbol source. No re-architecture of its Bubble Tea model.

### Out of Scope — new TUI library or dependency
- Introducing any new terminal-UI library, colour library, or module dependency. Reuse of the existing `charm.land/*` v2 + `internal/tui` stack only.

### Out of Scope — theme system changes
- Adding new `Theme` tokens, re-introducing dark-theme selection logic, or altering the 28-token palette. Colours are consumed as-is (the logo ramp reuses the existing `Accent`/`AccentDeep` tokens — no new token).

### Out of Scope — subcommand help surfaces
- The logo on any SUBCOMMAND help path — `moai help <subcommand>` (e.g. `moai help init`) and `moai <subcommand> --help` (e.g. `moai init --help`). Only the explicit ROOT-help invocations (`moai --help` / `moai -h` / `moai help`) carry the logo (REQ-TUXIU-055, resolved in scope at Implementation Kickoff); subcommand help stays logo-free, and the root-help predicate must not match those shapes.

### Out of Scope — fang modification
- Modifying, upgrading, or forking `charm.land/fang/v2` to obtain a header/pre-help injection option. fang v2.0.1 exposes no such option (verified); REQ-TUXIU-055 is satisfied entirely OUTSIDE fang by printing before `fang.Execute(...)`, so `go.mod` stays unchanged (REQ-TUXIU-043).

### Out of Scope — other CLI commands
- `moai doctor`, `moai status`, `moai loop`, `moai glm`, and every command other than `init` and `update`. Shared-primitive edits must not regress their existing rendering, but they are not redesign targets.

## §D Acceptance Criteria

Full Given-When-Then scenarios, severity, and traceability live in `acceptance.md`. The AC matrix covers the 8 approved look items (card summary, unified glyphs, block progress bar, spinner-residue removal, identity header band, outcome banner, init banner+card, **large logo**) and the hard constraints (Theme SSOT, NO_COLOR/monochrome, reduced motion, no new dependency, characterization/data preservation, coverage, plus the logo gradient/degradation/SSOT-guard checks and the root-help-surface / no-double-print check for REQ-TUXIU-055).

## §E References

- `plan.md` — milestone breakdown, key reversible decisions, technical approach, risks.
- `acceptance.md` — Given-When-Then scenarios, edge cases, Definition of Done.
- `progress.md` — plan/run/sync audit-ready signals.
- Origin: `SPEC-V3R3-CLI-TUI-001` (terminal design system v2 + Theme SSOT).
- Recent: `SPEC-CLI-TUX-V3-005` (Printer migration / ratchet).
- Reversal source: `SPEC-CLI-TUX-V3-004` REQ-TUX4-006 (commit `77893579e`) retired the large logo; Group F (v0.1.2) restores it with a coral gradient. The three retirement tests (`TestCompactBanner_NoASCIILogo`/`_TwoLineIdentity`/`_GlyphWhitelist`) are reconciled in run-phase.
