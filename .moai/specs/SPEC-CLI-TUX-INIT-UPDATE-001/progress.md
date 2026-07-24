# Progress — SPEC-CLI-TUX-INIT-UPDATE-001

## §E.1 Plan-phase Audit-Ready Signal

- **SPEC ID:** SPEC-CLI-TUX-INIT-UPDATE-001 — pre-write regex self-check PASS (`^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`); decomposition SPEC ✓ | CLI ✓ | TUX ✓ | INIT ✓ | UPDATE ✓ | 001 ✓.
- **Tier:** M (presentation-only change; 3-core-artifact set spec.md + plan.md + acceptance.md, plus progress.md; ~8-11 files across `internal/cli`, `internal/tui`, `internal/cli/printer`, `internal/cli/uikit`; 4 milestones. Tier M plan-auditor PASS threshold 0.80; design.md/research.md NOT required. Downgraded from Tier L per audit D6.)
- **Artifacts:** spec.md + plan.md + acceptance.md + progress.md created; status `draft`. **Version 0.1.3** (v0.1.2 → v0.1.3 = the R8 kickoff resolution fold; see the Resolved clarification (R8) entry below). Status stays `draft` — the `draft → in-progress` transition is manager-develop's on the M1 commit.
- **Requirements:** REQ-TUXIU-001..056 (Groups A-F; Group F = large logo restoration, added v0.1.2); REQ-045 is **N/A** (no machine-readable surface); REQ-055 is **IN SCOPE** as of v0.1.3 (explicit root-help logo injection — resolved at Implementation Kickoff, no longer capability-gated). Acceptance: AC-TUXIU-001..024 with §D.1 severity, §D.2 traceability, §D.3 edge cases, §E Definition of Done; AC-017 is **N/A**; AC-024 (MUST) binds REQ-055.
- **Audit fold (v0.1.1):** plan-auditor findings D1-D9 applied — D1 glyph-source resolved to option (b); D2 REQ-045/AC-017 N/A; D3 AC-019 coverage re-scoped (per-function ≥90% + no `internal/cli` regression vs the 74.6% baseline, `internal/tui` 93%); D4 AC-016 golden stdout/stderr Verify pinned; D5 AC-004 strengthened + `uikit/render.go:72` added; D6 Tier L→M; D7 AC-007/013a/013b/014 Verify added; D8 REQ-001 `preview_tui.go` exemption; D9 AC-012 hex baseline = 1 comment hit at `wizard/styles.go:20`.
- **Ground truth verified at plan-phase:** all `tui` primitives exist and are exported (Box/ThickBox/Pill/StatusIcon/Spinner/Progress/Stepper/ProgressLine/themes); `BoxOpts.Accent` present; Pill enum incl. PillOk/PillInfo/PillErr/PillPrimary + Solid; Theme = 28 colour tokens; glyph duplication confirmed in 4 sites; printer markers at printer.go:279-282; dev mode = tdd; version v3.0.1.
- **Resolved clarification (plan.md §A.1 item 1, D1):** glyph-source shape = **option (b)** — exported `tui.Glyph*` raw-rune const block (`GlyphDone='✓'`/`GlyphRun='●'`/`GlyphSkip='○'`/`GlyphErr='✗'`) as the single SSOT; `tui.StatusIcon` refactored to return the constants; `printer.go`/`uikit` reference them while keeping their own theme-painting. The NEEDS-CLARIFICATION marker has been removed from plan.md. User-approved at Implementation Kickoff.
- **Logo restoration extension (v0.1.2, user-approved additive scope):** added **Group F** REQ-TUXIU-050..056 + AC-TUXIU-020..023 (Look item 8). Restores the large 6-line "MoAI-ADK" ASCII-art logo retired by SPEC-CLI-TUX-V3-004 REQ-TUX4-006 (commit `77893579e`), with a vertical coral gradient.
  - **Ground truth verified:** the 6-line art is byte-recoverable from `git show 77893579e^:internal/cli/uikit/banner.go`; `charm.land/lipgloss/v2 v2.0.5` (already a direct `go.mod` dep) exposes `Blend1D(steps, stops...)` at `blending.go:18` → **zero new dependency**; ramp endpoints are `Theme.Accent` (light `#bf6547` / dark `#d97757`) → `Theme.AccentDeep` (light `#a84f33` / dark `#b85e3f`), both inside `internal/tui/theme.go`; all 3 target surfaces route through the single `uikit.PrintBanner` entry (`root.go:32`, `init.go:410`, `update.go:1253` — no other caller).
  - **Design decisions (plan.md §A.1 L1-L3):** logo art + gradient helper home = `internal/tui/logo.go` (`tui.Logo` / `tui.CoralRamp`) to keep colour math inside the theme-SSOT boundary; ramp = `lipgloss.Blend1D(6, Accent, AccentDeep)`; placement = stacked ABOVE the retained compact band **in `PrintBanner` only** (`bannerString` unmodified — the reversal-minimizing choice).
  - **Cross-SPEC reversal (B2):** REQ-TUX4-006 is partially reversed. Three retirement tests (`TestCompactBanner_NoASCIILogo`/`_TwoLineIdentity`/`_GlyphWhitelist`, `banner_compact_test.go`) plus `TestPrintBanner_*` (`banner_test.go`) require run-phase reconciliation — re-target to `bannerString` so each test's original "compact band stays compact" intent survives at the band layer.
  - **Contradiction resolved (REQ-TUXIU-056):** the logo's decorative block/box-drawing runes (`█ ╗ ╔ ╚ ╝ ═ ║`) fall outside the AC-CLI-TUI-017 status-glyph whitelist; carved out as a SEPARATE decorative category so REQ-004 and REQ-050 no longer conflict.
  - **Resolved clarification (R8) — v0.1.3:** explicit `moai --help` / `moai help` logo injection = **approach (a)**, IN SCOPE. A root-help predicate on the `runFang` entry seam (`internal/cli/fang.go` / `internal/cli/root.go`) inspects ONLY `os.Args[1:]` and prints `tui.Logo(th)` BEFORE `fang.Execute(...)`, routed through the same Printer gateway / stdout channel as `PrintBanner` (REQ-TUXIU-044 partition unchanged). Predicate shape: MATCHES `["--help"]` / `["-h"]` / `["help"]` (len == 1); does NOT match the empty arg list, `["help", <sub>]` (`moai help init`), or `[<sub>, ...]` (`moai init --help`). **No-double-print invariant (HARD):** the empty-arg exclusion is mandatory because no-args `moai` already prints the logo via `PrintBanner` inside `rootCmd.Run` — a predicate matching `[]` would render two logos on the most-visible surface. The existing `root.go` `trivialCommands` map is reused as the token source. **`go.mod` unchanged confirmed** — re-verified at kickoff that fang v2.0.1 exposes only 9 `With*` options with no header/pre-help hook, so approach (a) needs no fang change, upgrade, or fork (REQ-TUXIU-043 holds). Rejected: (b) no-args-only (leaves the most common discovery path logo-free), (c) fang upgrade/fork (breaches the no-new-dependency envelope). REQ-TUXIU-055 promoted from capability-gated open item to a normal in-scope GEARS requirement; the explicit-`--help` entry removed from spec.md §C Out of Scope (replaced by narrower subcommand-help + fang-modification exclusions); AC-TUXIU-024 added (MUST) with the occurrence-count-`== 1` double-print guard. The NEEDS-CLARIFICATION marker has been removed from plan.md — **zero open items remain**. User-approved at Implementation Kickoff.
  - **Milestone placement:** M1 gains the `tui.Logo`/`tui.CoralRamp` primitive (foundation, HIGH reversibility); M3 gains the `PrintBanner` stacking + reversal-test reconciliation + the (now unconditional, v0.1.3) root-help predicate wiring on the `runFang` seam; M4 gains logo verification, the AC-TUXIU-024 root-help/double-print checks, and the fresh logo-bearing golden capture. **No milestone renumbering.**
  - **Tier:** unchanged at **M** (the logo adds ~1 new file + 1 composition edit + test reconciliation; still within 5-15 files / 300-1000 LOC).
- **Out of Scope:** functional/data changes, JSON output, preview_tui.go re-architecture, new dependency, theme-token changes, other CLI commands, **subcommand help surfaces** (`moai help <sub>` / `moai <sub> --help` stay logo-free), and **fang modification** (no upgrade/fork for a header option). The explicit ROOT-help logo injection is no longer out of scope — it is IN SCOPE as of v0.1.3 (REQ-TUXIU-055 / AC-TUXIU-024).

## §E.2 Run-phase Evidence

**Milestone M1 — Shared symbol source + spinner-residue characterization + logo primitive (foundation).** M1 builds the primitives (glyph SSOT, spinner-residue regression locks, `tui.Logo`/`tui.CoralRamp`); the `PrintBanner` stacking + root-help predicate (M3) and the end-to-end surface / golden verification (M4) are separate milestones. AC rows below are scoped to what M1 actually delivers; surface-level rows (AC-020 3-surface presence, AC-024 root-help predicate) are DEFERRED to M3/M4 by design and are NOT counted as M1 FAILs.

| AC | REQ | M1 status | Verification command | Actual output |
|----|-----|-----------|----------------------|---------------|
| AC-TUXIU-004 | 001/002/003 | PASS | `grep -rnE '"✓"\|"✗"\|"○"\|"●"' internal/cli/printer/ internal/cli/uikit/ \| grep -v _test` | `<<0 matches>>` — zero raw-rune status-glyph literals remain in printer+uikit; all resolve from `tui.Glyph*` |
| AC-TUXIU-004 | 001 | PASS | `go test ./internal/tui/ -run 'TestGlyph\|TestStatusIconResolves'` | `PASS ok internal/tui 0.524s` (const block pinned; StatusIcon returns the constants) |
| AC-TUXIU-006 | 020/021 | PASS | `go test ./internal/tui/ -run TestStepResidue` + `go test ./internal/cli/printer/ -run Residue` | `PASS` — after last `\r\x1b[2K` erase the tail is one clean ✓ line, zero residual ○/⠋ (TTY, step + spinner) |
| AC-TUXIU-007 | 022 | PASS | same suites, non-TTY branch | `PASS` — zero ANSI CSI, zero `\r`, exactly one ✓ result line; TTY/non-TTY split preserved |
| AC-TUXIU-021 | 051/052 | PASS (primitive) | `go test ./internal/tui/ -run TestCoralRamp` + truecolor pty render of `tui.Logo` | `CoralRamp(6)` → 6 distinct stops, `ramp[0]` luminance > `ramp[5]`; truecolor `Logo` render emits 6 distinct `38;2` fg codes (168;79;51 deepest → 191;101;71 lightest), top-light→bottom-deep |
| AC-TUXIU-022 | 053 | PASS (primitive) | `go test ./internal/tui/ -run TestLogo_MonochromePlain` + mono/pipe binary render | mono theme → 0 SGR colour, art runes intact; non-TTY pipe → 0 distinct truecolor fg (≤1) via downsample |
| AC-TUXIU-023 | 051/040 | PASS | `grep -rnE '#[0-9a-fA-F]{6}' internal/tui/logo.go` ; `grep -rnE '#[0-9a-fA-F]{6}' internal/cli/ \| grep -v _test` | logo.go: `<<0 hex>>`; internal/cli: only `wizard/styles.go:20` (the 1 pre-existing comment baseline) — no NEW hex |
| AC-TUXIU-019 | 046 | PASS | `go tool cover -func` on new/touched render funcs | `Logo` 100%, `CoralRamp` 100%, `coralRamp` 100%, `StatusIcon` 100%, `symProgress/symSuccess/symError` 100%, uikit `SymSuccess/SymError/SymWarning/StatusIcon/RenderSuccessCard` 100%; `internal/tui` whole 93.6% (baseline 93.0%), `internal/cli` whole 74.7% (== baseline, no regression) |
| AC-TUXIU-050 (byte-identity) | 050 | PASS | `diff <(git show 77893579e^:…moaiBanner) <(logo.go moaiLogoArt)` | IDENTICAL — same SHA256 `f61c62e1…60856`; restored art byte-identical to the retired const |
| AC-TUXIU-020 | 050/054 | DEFERRED → M3/M4 | (PrintBanner 3-surface stacking) | M1 creates the `tui.Logo` primitive only; the primitive is NOT wired into `PrintBanner` (verified: `TestCompactBanner_NoASCIILogo/_TwoLineIdentity/_GlyphWhitelist` all still PASS — logo not wired early) |
| AC-TUXIU-024 | 055 | DEFERRED → M3/M4 | (root-help predicate) | out of M1 scope; wired in M3 |

**Full-suite regression (R2 cross-package glyph re-point):** `go test ./...` → exit 0, 0 FAIL (all `internal/tui`, `internal/cli`, `internal/cli/printer`, `internal/cli/uikit` packages green). B2 cross-SPEC retirement tests (`TestCompactBanner_*`, `TestPrintBanner_*`) all PASS unchanged.

**Milestone M2 — update.go presentation wiring.** Wired the M1 `tui` primitives (Box/Pill/Progress/StatusIcon) into `runTemplateSyncWithReporter` via a new `internal/cli/update_tux.go` (7 render helpers) + `update_tux_test.go` (15 tests, RED→GREEN). Fixed the visible two-part `○…○` spinner residue by removing the redundant reporter Step-wrapper from the deploy loop (it double-rendered every step on stderr alongside the stdout `tui.ProgressLine`, and called `StepStart` twice for "Restore Settings"); the inline `ProgressLine` is now the single per-step renderer. Empirically verified against a throwaway `moai init` project (never this repo): identity band, classification card, per-step progress bars, and outcome banner all render; stderr no longer carries deploy-step reporter lines.

| AC | REQ | M2 status | Verification command | Actual output |
|----|-----|-----------|----------------------|---------------|
| AC-TUXIU-001 | 010 | PASS | `go test -run TestRenderClassificationSummary_AllThreePills` | PASS — accent `tui.Box` border + pills `+ 1 add` / `~ 23 update` / `! 2 conflict` |
| AC-TUXIU-002a | 011 | PASS | `-run TestRenderClassificationSummary_ZeroConflictOmitted` | PASS — 0-count conflict pill absent; add/update present |
| AC-TUXIU-002b | 011 | PASS | `-run TestRenderClassificationSummary_OnlyUpdatePill` | PASS — only `5 update`; add+conflict omitted |
| AC-TUXIU-003 | 012/013 | PASS | `-run TestDeployStepStateIcon` | PASS — pending/running/done glyph == `tui.StatusIcon(skip/run/ok)` = ○/●/✓ |
| AC-TUXIU-005 | 014 | PASS | `-run TestRenderDeployProgress` + live capture | PASS — `● ██░░░░░░░░ 1/5 steps` … `✓ ██████████ 5/5 steps` |
| AC-TUXIU-008 | 015 | PASS | `-run TestRenderIdentityBand` + live | PASS — `◆ MoAI-ADK <PillPrimary v3.0.0> go1.26.4 · claude` |
| AC-TUXIU-009 | 016 | PASS | `-run TestRenderUpdateOutcome` + live | PASS — solid `PillOk` `✓ Updated N files` + dim `Backup:`/`Recover:` note |
| AC-TUXIU-011 | 011 | PASS | `-run TestRenderClassificationSummary_AllZeroEmpty` | PASS — all-zero → empty card (no noise) |
| AC-TUXIU-012 | 040 | PASS | `grep -rnE '#[0-9a-fA-F]{6}' internal/cli/update.go internal/cli/update_tux.go` | `<<0>>`; `internal/cli` overall still only `wizard/styles.go:20` comment baseline |
| AC-TUXIU-013a | 041 | PASS | `-run 'TestRender.*NoColor'` + `NO_COLOR=1` live | PASS — 0 SGR; pills → `[+ 1 add]` / `[~ 24 update]` / `[✓ Updated 26 files]` |
| AC-TUXIU-013b | 041 | PASS | non-TTY pipe capture of `moai update` | PASS — pills bracketed, `grep -c $'\x1b['[0-9;]*m'` = 0 |
| AC-TUXIU-015 | 043 | PASS | `git diff HEAD -- go.mod go.sum` | `<<0 lines>>` — no new module dependency |
| AC-TUXIU-016 | 044 | PASS-WITH-DEBT → M4 | data-invariance (residue fix removes stderr **presentation** only) | DATA lines preserved on-channel: `· Found 26 files` (stderr), `✓ Updated 26 files` + removed-path list + `Backup:` path (stdout). Final golden diff owned by M4. |
| AC-TUXIU-019 | 046 | PASS | `go tool cover -func` on update_tux.go funcs | 7 new render funcs **100.0%**; `internal/cli` whole-package **74.8%** (≥ 74.7% baseline, no regression) |

**Residue fix evidence (REQ-TUXIU-020/021, task load-bearing):** pre-fix stderr showed the duplicate `○ Restore Settings: Restoring user settings...` line twice + a coarse reporter Step per deploy step; post-fix stderr carries only the pre-deploy phases (`Version Check` / `Loading Templates` / `Loading Manifest`) + the `· Found 26 files to sync` data line — the deploy-step double-render is gone. `go test ./internal/cli/ ./internal/cli/update/... ./internal/cli/printer/... ./internal/tui/...` all green; `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0.

**Parallel-session note:** at M2 run-time the shared checkout carried an unrelated live session's uncommitted edits (catalog.yaml, `SKILL.md`, `manager-{develop,spec}.md`, `gate.yaml`, `tool_policy.go`). These trip `TestCatalogHashParity` / `TestManifestHashFormat` / `TestDeprecatedPaths_NoTemplateCollision` and 2 `tool_policy.go` errcheck lints — NONE are in M2 scope (`internal/cli/update.go` + `internal/cli/update_tux*.go`); only those 3 files were staged for the M2 commit.

**Milestone M3 — init banner + success card + large logo placement (the headline reversal).** Stacked `tui.Logo(th)` ABOVE the compact `bannerString` inside `PrintBanner` — `bannerString` UNMODIFIED (reversal-minimizing, §A.1 L3 / §B R6), so one edit covers all 3 shared-entry surfaces (`root.go:32`, `init.go:410`, `update.go:1274`). Reconciled the SPEC-CLI-TUX-V3-004 REQ-TUX4-006 retirement tests RED-first: re-targeted `TestCompactBanner_NoASCIILogo/_TwoLineIdentity/_GlyphWhitelist` to assert against `bannerString` DIRECTLY (via a new `uikit/export_test.go` `BannerString` test hook — the band stays logo-free / ≤2 lines / status-glyph-only), added `TestPrintBanner_CarriesLogo` (composed surface carries the logo), and made `TestPrintBanner_OutputFormat` logo-aware. Wired the root-help predicate `isRootHelpArgs` on the `runFang` seam (pre-`fang.Execute` `uikit.PrintLogo()`) — matches `moai --help`/`-h`/`help`, excludes the empty arg vector (no-args double-print guard) and subcommand-help. Re-dressed `buildInitSuccessCard` in the shared `tui.Box` + `tui.Pill` language. Regenerated the 3 uikit banner goldens (logo now present). No `go.mod` change, no fang fork.

| AC | REQ | M3 status | Verification command | Actual output |
|----|-----|-----------|----------------------|---------------|
| AC-TUXIU-010 | 030 | PASS | `moai init` shares `PrintBanner` (`init.go:410`) → `bannerString` band+pills | `◆ MoAI-ADK` band + `[v…] [go …] [claude]` pills present (`TestCompactBanner_BrandTagline`/`TestBannerPill_Metadata`) |
| AC-TUXIU-011 | 031 | PASS | `go test -run TestInitSuccessCard_TuiBoxPillLanguage` | PASS — `NO_COLOR` card renders `[3 dirs]` / `[7 files]` `tui.Pill` inside `tui.Box` border; next-actions retained |
| AC-TUXIU-018 | 004 | PASS | `-run TestCompactBanner_GlyphWhitelist` (band-scoped) | PASS — whitelist scoped to `bannerString`; logo block/box runes (`█ ╗ ╔ ╚ ╝ ═`) exempt (REQ-TUXIU-056 carve-out) yet present in `PrintBanner` |
| AC-TUXIU-020 | 050/054 | PASS | scratchpad `moai` no-args ANSI-stripped; shared-entry xref | logo `███╗   ███╗` ABOVE `◆ MoAI-ADK` band (both present); 3 surfaces share `PrintBanner` (`TestPrintBanner_CarriesLogo` in-process) |
| AC-TUXIU-023 | 051/040 | PASS | `grep -rnE '#[0-9a-fA-F]{6}' internal/cli/ | grep -v _test` + `internal/tui/logo.go` | `internal/cli` hex == 1 baseline (`wizard/styles.go:20`); `logo.go` hex `<<0>>` — ramp from `Accent`/`AccentDeep` tokens |
| AC-TUXIU-013a/b | 041 | PASS | scratchpad `NO_COLOR=1 moai --help` SGR count | `<<0>>` SGR codes; logo runes intact (art degrades to plain, no ANSI) |
| AC-TUXIU-024 | 055 | PASS | scratchpad binary, ANSI-stripped `grep -c -F '███╗   ███╗'` | `moai --help`=1, `moai help`=1, `moai init --help`=0, `moai help init`=0, no-args `moai`=**1** (double-print guard); predicate `TestIsRootHelpArgs` 6/6 shapes PASS |

**M3 build/test evidence:** `go test ./...` exit 0 (full suite); `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run ./internal/cli/...` = 0 issues (== baseline). Per-function coverage: `isRootHelpArgs` 100%, `PrintBanner` 100%, `PrintLogo` 100%, `buildInitSuccessCard` 100%, `runFang` 91.7% (the `PrintLogo` root-help branch is exercised by the scratchpad binary, not in-process), `tui.Logo`/`CoralRamp` 100%. `internal/cli/uikit` whole-package 98.8%; `internal/cli` 74.8% (≥ 74.6% baseline, no regression). `git diff -- go.mod go.sum` empty.

**Parallel-session note (M3):** the shared `feat/SPEC-CLI-TUX-INIT-UPDATE-001` branch carries an unrelated live session's uncommitted `CLAUDE.local.md` + untracked `.moai/reports/*.html` + `SPEC-TDD-ANTICHEAT-001/` — NONE staged (explicit-pathspec commit only; M3 files are disjoint from SPEC-CONFIG-AUDIT-REPAIR-001's `uikit`/`fang`/`root`/`init_warnings` scope).

**Milestone M4 — Verification and coverage + logo-bearing golden re-capture (mechanical; final run milestone).** Added the AC-016 standing guard `internal/cli/tuxiu_characterization_test.go` (data-line filter drops the EXPECTED-NEW presentation — logo/band/card/progress-bar — the M2 residue-fix reporter block, and normalizes run-variant backup timestamps + NO_COLOR pill brackets + indentation) comparing the immutable M1.0 PRE-logo baseline against a FRESH post-M4 presentation baseline. Golden strategy = **option (a)**: 12 fresh logo-era goldens captured via the M1 scratchpad harness (`capture.sh` + `pty_capture.py`, throwaway dirs — repo never mutated) and committed ALONGSIDE the M1.0 goldens at `internal/cli/testdata/tuxiu/postm4/`, so the data-invariance guarantee vs the PRE-logo baseline stays a diff of two committed fixtures. Negative-control verified (injected `Updated 27 files` → test FAILs at the drifted line + missing-token; restore → PASS), so the guard is non-vacuous. Logo/gradient/root-help verified from the BUILT BINARY under a forced-truecolor pty (the `tui.downsample` `sync.Once` constraint requires a real binary, not an in-process test).

| AC | REQ | M4 status | Verification command | Actual output |
|----|-----|-----------|----------------------|---------------|
| AC-TUXIU-016 | 044 | PASS | `go test -run 'TestInitUpdateTUXCharacterization\|TestTUXChannelPartition\|TestTUXDataValuesPreserved' ./internal/cli/` | PASS — DATA-line subset byte-identical across all 12 surface×variant×channel pairs; `Updated 26 files` on stdout (never stderr), `Found 26 files to sync` on stderr (never stdout); counts 26/26/13/2/3 preserved. Negative control (27 files) FAILs the guard |
| AC-TUXIU-020 | 050/054 | PASS | binary pty capture, ANSI-stripped `grep -c -F '███╗   ███╗'` + `TestPrintBanner_CarriesLogo` | no-args `moai`: logo=1 ABOVE `◆ MoAI-ADK` band=1; PrintBanner shared entry carries logo (root.go:32 / init.go:410 / update.go:1274 same func) |
| AC-TUXIU-021 | 051/052 | PASS | `COLORTERM=truecolor` pty capture; `grep -oE $'\x1b\\[38;2;[0-9;]*m' \| sort -u` | 6 distinct coral fg stops `217;119;87 → 211;114;82 → 204;109;77 → 197;104;72 → 191;99;67 → 184;94;63`, monotonically deepening (row1 luminance > row6) — top-light→bottom-deep, ≥3 distinct |
| AC-TUXIU-022 | 053 | PASS | `NO_COLOR=1` binary + non-TTY pipe binary | NO_COLOR → 0 SGR colour, `███╗   ███╗` runes intact; non-TTY pipe → 0 distinct fg (≤1) via downsample |
| AC-TUXIU-023 | 051/040 | PASS | `grep -rnE '#[0-9a-fA-F]{6}' internal/tui/logo.go` ; `internal/cli/ \| grep -v _test` | logo.go `<<0 hex>>`; internal/cli only `wizard/styles.go:20` (1 comment baseline) — no NEW hex |
| AC-TUXIU-024 | 055 | PASS | binary pty captures, `grep -c -F '███╗   ███╗'` on 5 surfaces | `moai --help`=1, `moai help`=1, `moai init --help`=0, `moai help init`=0, no-args `moai`=**1** (double-print guard); `TestIsRootHelpArgs` 6/6 shapes |
| AC-TUXIU-013a/b | 041 | PASS | `TestRenderClassificationSummary_NoColorBracketPills` / `_NoColor` + NO_COLOR binary | 0 SGR under NO_COLOR; pills → `[label]` (`[~ 24 update]` `[! 2 conflict]` `[✓ Updated 26 files]`) on init+update+banner |
| AC-TUXIU-014 | 042 | PASS | `go test -run 'TestSpinnerStatic\|TestProgressStatic' ./internal/tui/` | MOAI_REDUCED_MOTION=1 → static `●` spinner + fully-filled bar (0 `⠋`-family, 0 `░`) |
| AC-TUXIU-004/018 | 001/002/003 | PASS | `grep -rnE '"✓"\|"✗"\|"○"\|"●"' internal/cli/printer/ internal/cli/uikit/ \| grep -v _test` | `<<0 raw-rune status-glyph literals>>` — all resolve from `tui.Glyph*`; whitelist scoped to status glyphs (logo decorative runes exempt, REQ-056) |
| AC-TUXIU-012 | 040 | PASS | `grep -rnE '#[0-9a-fA-F]{6}' internal/cli/ \| grep -v _test` | exactly 1 — `wizard/styles.go:20` comment baseline (no NEW hex) |
| AC-TUXIU-015 | 043 | PASS | `git diff HEAD -- go.mod go.sum` | `<<0 lines>>` — no new module dependency |
| AC-TUXIU-019 | 046 | PASS | `go tool cover -func` on new/touched render funcs | isRootHelpArgs/PrintBanner/PrintLogo/bannerString/buildInitSuccessCard/paintToken/deployStepStateIcon/renderIdentityBand/classifyUpdateCounts/renderClassificationSummary/renderDeployProgress/renderUpdateOutcome/Logo/CoralRamp/coralRamp/StatusIcon **100.0%**; runFang 91.7% (root-help branch is binary-exercised) — all ≥90%; `internal/cli` whole 74.9% (≥74.6% baseline, no regression), `internal/tui` 93.6% |

**M4 build/test evidence:** `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run ./internal/cli/... ./internal/tui/...` = 0 issues; `go test ./internal/cli/... ./internal/tui/...` all packages `ok` (0 FAIL — TestCatalogHashParity green; reversal tests `TestCompactBanner_*`/`TestPrintBanner_*` green). Subagent-boundary grep on internal/tui + internal/cli: 0 actual invocations (pre-existing matches are cobra Long doc-strings + the `agentlint` detector + test fixtures — none are calls). `git diff HEAD -- go.mod go.sum` empty.

**Parallel-session note (M4):** the shared branch carries the live SPEC-CONFIG-AUDIT-REPAIR-001 session's uncommitted `CLAUDE.local.md` + `internal/cli/{cc,cg,glm,launcher}.go` + `launcher_test.go` edits + untracked `.moai/reports/*.html` + `SPEC-TDD-ANTICHEAT-001/` — NONE staged. M4 staged ONLY `internal/cli/tuxiu_characterization_test.go` + `internal/cli/testdata/tuxiu/postm4/` + this `progress.md` (explicit-pathspec commit; disjoint from the parallel scope).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-25            # M1+M2+M3+M4 complete — run-phase DONE; ready for sync-phase (manager-docs)
run_commit_sha: 0164ef002              # M4 commit (backfilled here); finalizes the run-phase SHA. run-phase milestone commits: M1 5ee17e9e1, M2 d6d19bbc5, M3 c2fd0bf8c, M4 0164ef002
run_status: run-complete               # M1 foundation + M2 update wiring + M3 init/PrintBanner-logo/root-help + M4 verify/golden-recapture all done
ac_pass_count: 25                      # all gating AC-TUXIU-001..024 PASS (12 MUST + 13 SHOULD); M4 finalized AC-016 (PASS, was PASS-WITH-DEBT at M2) + AC-021/022 truecolor gradient from binary + AC-014/012/023 guards
ac_fail_count: 0
ac_na_count: 1                         # AC-TUXIU-017 (JSON preservation) N/A — no machine-readable surface (D2)
ac_deferred_count: 0
preserve_list_post_run_count: unchanged  # bannerString NOT modified (logo stacks only in PrintBanner); root.go untouched; M4 touched only testdata/tuxiu + a new _test.go (zero source edits); no defect uncovered
l44_pre_commit_fetch: orchestrator-owned  # pre-spawn fetch is orchestrator-side (agent does not fetch/push in isolation); Route B feat branch
l44_post_push_fetch: pending           # orchestrator pushes feat/SPEC-CLI-TUX-INIT-UPDATE-001; PR created by orchestrator via cherry-pick onto a clean branch (parallel-session contamination)
new_warnings_or_lints_introduced: 0    # golangci-lint ./internal/cli/... ./internal/tui/... = 0 issues == pre-edit baseline
cross_platform_build:
  darwin_amd64: pass                   # go build ./... exit 0
  windows_amd64: pass                  # GOOS=windows GOARCH=amd64 go build ./... exit 0
coverage:
  new_render_funcs: ">=90% each"       # 16 funcs 100.0%, runFang 91.7% — all >=90%
  internal_cli_whole: 74.9%            # >= 74.6% baseline (no regression, +0.3%)
  internal_tui_whole: 93.6%            # >= 93.0% baseline
total_run_phase_files: 28              # M1-M3: 4 new tui src/test + 2 residue tests + 6 modified + 12 M1.0 goldens + M2 update_tux.go/_test.go + M3 export_test.go/fang_roothelp_test.go/banner reconcile; M4: +1 tuxiu_characterization_test.go + 12 postm4 goldens + progress.md
m1_to_mN_commit_strategy: per-milestone-commit  # M1 (draft→in-progress) / M2 / M3 / M4 separate commits on the shared feat branch
golden_baseline_captured: true         # M1.0 PRE-logo goldens at internal/cli/testdata/tuxiu/{init,update}.{tty,notty,nocolor}.{stdout,stderr}.golden (immutable data-invariance reference)
golden_postm4_captured: true           # FRESH logo-era goldens at internal/cli/testdata/tuxiu/postm4/ (option a — alongside M1.0); both captured in scratchpad throwaway dirs (repo never mutated); AC-016 diffs the two committed sets
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_

## §F Phase 4 Mode Selection

**Input parameters**

| Signal | Value |
|--------|-------|
| tier | M |
| scope (file count) | ~8-11 files (`internal/tui`, `internal/cli`, `internal/cli/printer`, `internal/cli/uikit`) + 1 new file (`internal/tui/logo.go`) |
| domain count | 1 (Go CLI presentation layer) |
| file language mix | 100% Go (+ golden fixtures under `internal/cli/testdata/tuxiu/`) |
| concurrency benefit | LOW — coding-heavy, 4 milestones in strict build-dependency order (M1 foundation → M2/M3 consumers → M4 verification) |
| Implementation Kickoff Approval | PASSED (user-approved; R8 open item resolved to approach (a) at kickoff) |

**Mode evaluation**

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 `trivial` | no | Multi-file behaviour change (spinner-clear contract) + a new exported primitive — not a typo-class edit |
| 2 `background` | no | Write-heavy implementation, not read-only analysis |
| 3 `agent-team` | no | RETIRED (Phase 4 tombstone) — never selectable |
| 4 `parallel` | no | Single domain (1 < 3) and coding-heavy — Anthropic's coding-task parallelism caveat applies |
| 5 `sub-agent` | **YES** | Default fallback; matches the sequential milestone dependency chain |
| 6 `workflow` | no | ~11 files (< ~30) and the transform is semantic/new-code, not one uniform mechanical rule |

**Decision:** `sub-agent`

**Justification:** This is coding-heavy work in a single domain with a strict build-dependency chain — M1 establishes the `tui.Glyph*` SSOT and the `tui.Logo`/`tui.CoralRamp` primitive that M2 and M3 both consume, and M4 verifies against the M1.0 golden baseline. Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research") makes Mode 4 inappropriate despite the file count, and the scope sits well under Mode 6's ~30-file mechanical-transform threshold with a transformation that is semantic rather than one uniform rule. One sequential `manager-develop` spawn per milestone (`cycle_type=tdd` per `quality.yaml`), with the orchestrator running a read-only parallel verification batch between milestones.

**Boundary case:** none — every threshold is cleared by a wide margin (1 vs 3 domains; ~11 vs ~30 files).
