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

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-25            # M1+M2+M3 complete; M4 (final verify/golden re-capture) pending
run_commit_sha: pending-backfill-M3    # M3 commit SHA backfilled in a follow-up chore commit; full run-phase SHA finalizes at M-final
run_status: M3-complete                # M1 foundation + M2 update wiring + M3 init/PrintBanner-logo/root-help done; M4 (verify) pending
ac_pass_count: 26                      # M1(8) + M2(14) + M3: 010,020,024,018 newly-passing (+4); 011/013a/b re-confirmed on the init surface, 023 logo-ramp SSOT held
ac_fail_count: 0
ac_deferred_count: 0                   # AC-020 (3-surface presence) + AC-024 (root-help predicate) now PASS in M3; none deferred
preserve_list_post_run_count: unchanged  # bannerString NOT modified (logo stacks only in PrintBanner); root.go untouched; Spinner/Stepper/term/form literals left per AC-004 carve-out
l44_pre_commit_fetch: not-run          # orchestrator-owned pre-spawn fetch (agent does not fetch/push in isolation); PR-route branch
l44_post_push_fetch: pending           # push to feat/SPEC-CLI-TUX-INIT-UPDATE-001 (Route B; enforce_admins main-direct disabled)
new_warnings_or_lints_introduced: 0    # golangci-lint 0 issues == pre-edit baseline 0
cross_platform_build:
  darwin_amd64: pass                   # go build ./... exit 0
  windows_amd64: pass                  # GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 15              # 4 new src/test (glyphs.go, glyphs_test.go, logo.go, logo_test.go) + 1 (progress_line_residue_test.go) + 1 (step_residue_test.go) + 5 modified (status.go, progress_line.go, printer.go, uikit/styles.go, uikit/status.go, uikit/render.go = 6 modified) + 12 golden fixtures; scoped to internal/tui + internal/cli/{printer,uikit} + testdata/tuxiu
m1_to_mN_commit_strategy: single-M1-commit  # M1 delivered in one commit carrying draft→in-progress; M2/M3/M4 are separate future commits/spawns
golden_baseline_captured: true         # M1.0 pre-logo goldens at internal/cli/testdata/tuxiu/{init,update}.{tty,notty,nocolor}.{stdout,stderr}.golden — executed in scratchpad throwaway dirs (repo never mutated)
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
