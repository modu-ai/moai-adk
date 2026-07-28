---
id: SPEC-CLI-TUI-MODERNIZE-001
title: "Progress — interactive TUI surface modernization"
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

# Progress — SPEC-CLI-TUI-MODERNIZE-001

## §E.1 Plan-phase Audit-Ready Signal

- **Artifacts (Tier L, 6-file set)**: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md` under `.moai/specs/SPEC-CLI-TUI-MODERNIZE-001/`. `design.md` + `research.md` added at v0.1.2 (plan-audit D4 — the v0.1.1 4-file set was short of the Tier L requirement).
- **Counts**: 41 GEARS requirements (Groups A-E) / 46 acceptance criteria (Groups A-F, 16 MUST-FIX). Verified: `grep -c '^- \*\*REQ-TUIM-' spec.md` → `41`.
- **Tier**: L. **Milestones**: M1 (preview redesign, independently landable), M2 (dead confirm removal), M3 (huh theme unification).
- **SPEC ID self-check**: `decomposition: SPEC ✓ | CLI ✓ | TUI ✓ | MODERNIZE ✓ | 001 ✓ → PASS` (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`, executed).
- **Vocabulary decision**: TUX / TUI / huh form runtime — user-confirmed, bound in `spec.md` §A.2.
- **Measured baselines** (plan.md §B, all executed this session): file sizes; `ConfirmMerge` production callers = 0; hex literals outside `internal/tui/` = 4 (2 real code in `internal/merge/confirm.go:454-455`, 2 comment-only); coverage `internal/cli` 74.9% / `internal/cli/update` 70.2% / `internal/merge` 86.3% / `internal/cli/wizard` 95.2% / `internal/tui` 93.6%; `GOOS=windows` build exit 0; `wizard.go` theme anchor at line 483.
- **Open clarifications**: **0**. Both prior markers were resolved by the user at v0.1.1 and recorded in plan.md §I — D-4 = HYBRID residue policy (table inline / diff alternate-screen / `esc` restores / single-line exit summary); D-5 = SPLIT verification (`NO_COLOR` structure golden + token-application unit tests, self-trip bound to the token tests).
- **Renderer verification (D-4 precondition, executed)**: the mid-run `AltScreen` toggle is first-class in bubbletea v2 — `cursed_renderer.go:320` compares `s.lastView.AltScreen != view.AltScreen` per frame and emits the DEC mode 1049 pair (`:513-525`). `esc`-restores-scrollback is a mode-1049 guarantee, so **no fallback was needed and none is recorded**. Verification additionally surfaced a load-bearing constraint: the close path (`:167-171`) discards the final frame when the last view is an alt-screen view, so a quit reachable from the diff sub-view would silently drop the exit summary → REQ-TUIM-022a + AC-TUIM-014d (MUST-FIX).
- **plan-audit iteration 1**: FAIL 0.74 (Tier L threshold 0.85) — 4 MUST-FIX, 9 SHOULD-FIX. All 13 repaired at v0.1.2; every numeric baseline and the full M2 DELETE inventory were independently reproduced by the auditor and are unchanged. Repairs were confined to the acceptance-criteria layer plus the two missing Tier L artifacts.
- **v0.1.2 repair evidence (commands executed this session)**: lipgloss v2 renders `#d97757` as `\x1b[38;2;217;119;87m`, `containsHexToken=false` (D1); escaped-`\|` ERE returns exit 1 against a positive control while the unescaped form matches (D3/S1); `grep -c $'\x1b\['` errors with `brackets ([ ]) not balanced` exit 2, corrected form returns `0` (S5); `internal/cli/update/plan/plan.go:63,64,88` confirmed as a fourth `merge.FileAnalysis` consumer with a named-field composite literal (S3); `go.mod:6` pins `bubbles/v2 v2.1.1` (S8); REQ count 41 (S9).
- **plan-audit iteration 2 (delta)**: CONDITIONAL **0.90** — Tier L threshold 0.85 cleared. 12 of 13 v0.1.2 repairs confirmed genuinely fixed; all 7 must-pass criteria PASS; AC-TUIM-030a/030c confirmed carried through byte-for-byte. Three residual findings (NEW-1 MUST-FIX, NEW-2 SHOULD-FIX, NEW-3 NICE-TO-HAVE) plus 3 cosmetics, all repaired at v0.1.3.
- **v0.1.3 NEW-1 evidence (executed before applying the fix, per the auditor's calibration note)**: combined-attribute render against pinned `charm.land/lipgloss/v2@v2.0.5` — `fgOnly = "\x1b[38;2;191;101;71mx\x1b[m"`, `bold+fg = "\x1b[1;38;2;191;101;71mx\x1b[m"`. The v0.1.2 full-CSI prefix is a substring of `fgOnly` (true) but **not** of `bold+fg` (false); the parameter run `38;2;191;101;71` is a substring of all three. The prediction was confirmed, so the fix was applied: SGR-parameter-substring method, recorded with evidence in acceptance.md §C.2.
- **Status**: `draft`. Awaiting Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

Branch `feat/SPEC-CLI-TUI-MODERNIZE-001`, base `d59936c71`. Commits: `217f0cc60` (M1), `1b42e1a0f` (M2), `283877ed0` (M3). Verbatim command logs under `.moai/state/verify/tuim/`.

### Re-measured denominators (plan.md §C pre-flight, executed this session)

| Baseline | plan.md §B value | Re-measured at `d59936c71` | Note |
|---|---|---|---|
| `internal/cli` coverage | 74.9% | **74.8%** | 0.1pp lower on this base revision; the re-measured figure is the denominator used below |
| `internal/cli/update` coverage | 70.2% | 70.2% | matches |
| `internal/merge` coverage | 86.3% | 86.3% | matches |
| `internal/cli/wizard` coverage | 95.2% | 95.2% | matches |
| `internal/tui` coverage | 93.6% | 93.6% | matches |
| CMD-HEX hits | 4 | 4 (2 executable at `confirm.go:454-455`, 2 comment-only in `wizard/styles.go`) | matches |
| CMD-GLYPH / CMD-ANSI | 0 / 0 | 0 / 0 (both exit 1 = zero matches) | matches |
| `GOOS=windows` build | exit 0 | exit 0 | matches |
| `wizardIsDark` anchor | line 483 | line 483 | matches |
| M2 reachability census | FileAnalysis 6 / MergeAnalysis 3 / ConfirmMerge 4 (all comment-only) | identical; no reflection dispatch, no build-tag-gated caller | matches plan.md §B.4 with **zero divergence** |

### AC PASS/FAIL matrix — 46 criteria

MUST-FIX rows are marked ★.

| AC | Sev | Status | Actual output |
|---|---|---|---|
| AC-TUIM-001 ★ | MUST | PASS | `grep -n "modu-ai/moai-adk/internal/tui" preview_tui.go` → `12:` (baseline 0 hits) |
| AC-TUIM-002 ★ | MUST | PASS | `grep -nE "table\.WithStyles\|SetStyles"` → `156: table.WithStyles(previewTableStyles(th)),` (baseline 0) |
| AC-TUIM-003 | SHOULD | PASS | `TestPreviewTableLightAndDarkAxesDiffer` PASS; trips under the self-trip mutation |
| AC-TUIM-004 ★ | MUST | PASS | `TestPreviewTableAppliesAccentToken` PASS via §C.2 method; params `38;2;191;101;71` present |
| AC-TUIM-005 | SHOULD | PASS | `grep -nE "tui\.(Box\|Section\|ThickBox)"` → `293: return tui.Box(...)`; `TestPreviewSummaryRendersThroughStructuralPrimitive` PASS |
| AC-TUIM-006 ★ | MUST | PASS | `grep -n "tui.HelpBar"` → `303`, `312`; `grep -nF "[enter] view diff"` exit 1 (absent); `grep -nF "[esc] back to table"` exit 1 (absent) |
| AC-TUIM-007 | INFO | PASS | `git diff d59936c71..HEAD -- preview_test.go` empty — the four content assertions pass with unedited bodies |
| AC-TUIM-008 | SHOULD | PASS | `TestPreviewClassLabelsCarryDistinctSemanticRoles` PASS — four distinct SGR runs (Success/Info/Dim/Danger) |
| AC-TUIM-009 | SHOULD | PASS | `TestPreviewViewsEmitZeroANSIUnderMonochrome` PASS for tableView, diffView, resultLine |
| AC-TUIM-010 | SHOULD | PASS | CMD-GLYPH → `0`, exit 1; positive control → `2`, exit 0 |
| AC-TUIM-011 ★ | MUST | PASS | CMD-ANSI → `0`, exit 1; positive control → `1`, exit 0. Import block is `fmt, io, os, strings, unicode/utf8` — no lipgloss, no internal/tui |
| AC-TUIM-012 | INFO | PASS | `TestPreviewFallbackZeroANSIUnderNoColor` + `...WhenPiped` PASS with unedited bodies |
| AC-TUIM-013 | SHOULD | PASS | `testdata/preview_fallback.golden` + `TestPreviewFallbackClassColumnHasUniformWidth` PASS (all four rows start the path column at one offset) |
| AC-TUIM-014a | SHOULD | PASS | `TestPreviewTableSubViewRendersInline` PASS — `View().AltScreen == false` |
| AC-TUIM-014b | SHOULD | PASS | `TestPreviewDiffSubViewRendersOnAlternateScreen` PASS — `View().AltScreen == true` |
| AC-TUIM-014c | SHOULD | PASS | `TestPreviewEscRestoresInlineTableSubView` PASS |
| AC-TUIM-014d ★ | MUST | PASS | `TestPreviewDiffReachableQuitKeysResolveToInline` PASS for `y` / `q` / `ctrl+c`, each entered **from the diff sub-view**; each asserts inline view + single-line frame + no card/file-row content |
| AC-TUIM-014e | SHOULD | PASS | `preview_tui.go:36` `# Residue policy — HYBRID…`; `:58` `CONSTRAINT — a quit MUST resolve to the inline sub-view first` with its renderer reason |
| AC-TUIM-014f | SHOULD | PASS | `TestPreviewTableCancelKeyResolvesToInline` PASS — `n` from the table view, `confirmed == false` |
| AC-TUIM-015 | INFO | PASS | `TestPreviewTableClassificationMatchesClassify` PASS unmodified; `grep -nE "^func (Classify\|classifyAll)" preview_tui.go` exit 1 (no re-implementation) |
| AC-TUIM-016 | INFO | PASS | accessor grep → `5` |
| AC-TUIM-017 ★ | MUST | PASS | `grep -rn "ConfirmMerge\|confirmModel\|fileListItem\|AnalysisFormatter"` → comment prose only (update.go:1124/1131/1137/1139, preview.go:5/18/59/69, three test comments); zero executable-code hits |
| AC-TUIM-018 ★ | MUST | PASS | `go build ./...` exit 0 AND all four sites: `update.go` 3 hits, `update_tux.go` 1, `update/merge/merge.go` 3, `update/plan/plan.go` lines 63/64/**88** (named-field literal intact) |
| AC-TUIM-019 | SHOULD | PASS | `grep -rn "charm.land/bubbletea\|charm.land/bubbles\|charmbracelet/lipgloss" internal/merge/*.go` exit 1 — zero hits including test files |
| AC-TUIM-020 ★ | MUST | PASS | CMD-HEX → exactly **2** hits, both comment-only (`wizard/styles.go:17,20`); baseline 4. Positive control matched line 1 |
| AC-TUIM-021 ★ | MUST | PASS | `grep -n "isatty.IsTerminal(os.Stdin.Fd())" update.go` → `152`, `1145`; `git diff d59936c71..HEAD -- update.go` empty ⇒ error string byte-identical |
| AC-TUIM-022 | SHOULD | PASS | `go vet ./...` exit 0, `go test ./internal/merge/...` PASS; both confirm test files handled (deleted, with two tests retargeted) |
| AC-TUIM-023 | SHOULD | PASS | `find internal/cli/testdata/tuxiu -name '*.golden' \| wc -l` → **24**; `git diff d59936c71..HEAD --stat -- internal/cli/testdata/` empty ⇒ no fixture regenerated |
| AC-TUIM-024 | SHOULD | PASS | Reachability re-run at execution time (see the denominator table above); result matched plan.md §B.4 exactly |
| AC-TUIM-025 ★ | MUST | PASS | `git diff d59936c71..HEAD -- internal/cli/wizard/wizard.go` **empty — zero hunks**. Anchor re-derived: `grep -n "var wizardIsDark"` → 483 |
| AC-TUIM-026 | SHOULD | PASS | `grep -c "func moaiHuhStyles" huh_theme.go` → 1; `grep -c "func moaiWizardStyles" wizard.go` → 1 |
| AC-TUIM-027 | SHOULD | PASS | `grep -c "moaiWizardStyles" wizard/styles.go` → 0; all four wizard theme symbols remain in wizard.go |
| AC-TUIM-028 | SHOULD | PASS | Divergence audit table recorded below and at the `moaiHuhStyles` doc comment; all 6 gaps **closed**, 0 reasoned-absent |
| AC-TUIM-029 | SHOULD | PASS | `grep -c "var huhThemeIsDark"` → 1; `grep -c "var wizardIsDark"` → 1 |
| AC-TUIM-030a | SHOULD | PASS | 3 goldens under `internal/cli/update/testdata/`; palette-insensitivity **proved**: light `Accent` → `#00ff00` and `Danger` → `#ff00ff`, all 4 golden tests still PASS, reverted (`git status` clean) |
| AC-TUIM-030b ★ | MUST | PASS | `TestPreviewSemanticRolesResolveToThemeTokens` PASS over 6 roles (conflict→Danger, border→ChromeBorder, selected→Accent, add→Success, update→Info, preserve→Dim); `grep -nE '#[0-9a-fA-F]{6}'` over the new test files exit 1 |
| AC-TUIM-030c ★ | MUST | PASS | Self-trip performed — see §E.2 Self-trip below. 4 mechanism-2 tests FAILED, including `TestPreviewSemanticRolesResolveToThemeTokens` by name |
| AC-TUIM-031 ★ | MUST | PASS | `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (also verified `GOOS=linux` exit 0) |
| AC-TUIM-032 | INFO | PASS | See the coverage table below — no package regressed; two improved |
| AC-TUIM-033 ★ | MUST | PASS | `git diff d59936c71..HEAD -- go.mod go.sum` empty |
| AC-TUIM-034 ★ | MUST | PASS | `git diff d59936c71..HEAD --stat -- internal/template/templates/` empty |
| AC-TUIM-035 | SHOULD | PASS | Detached worktree at `217f0cc60`: `go build ./...` ok, `go test ./...` FAIL count **0** with M2 and M3 absent |
| AC-TUIM-036 | SHOULD | PASS | `golangci-lint run --timeout=3m` → **0 issues**, exit 0 |
| AC-TUIM-037 | SHOULD | PASS | `spec.md` §A.2 carries all three vocabulary rows (lines 52-54); §B Group B/C/D headings carry layer labels (lines 86, 104, 115) |
| AC-TUIM-038 | SHOULD | PASS | `test -d internal/tui` ok; `git diff d59936c71..HEAD --stat -- internal/tui/` empty (no rename, no edit) |
| AC-TUIM-039 | SHOULD | PASS | `TestPreviewInheritsCanonicalAxisPrecedence` PASS over all 6 cases including (a) `NO_COLOR` beats `MOAI_THEME=dark` and (e) `MOAI_THEME=purple` → dark without querying the terminal |

**Result: 46/46 PASS. 16/16 MUST-FIX PASS. 0 FAIL, 0 PASS-WITH-DEBT.**

### Self-trip (AC-TUIM-030c) — verbatim

Mutation applied to `preview_tui.go`: removed the `table.WithStyles(previewTableStyles(th))` option and reverted the row construction to the unpainted `table.Row{c.Class.String(), c.RelPath}`. Command: `go test ./internal/cli/update/ -run 'TestPreview'`. Log: `.moai/state/verify/tuim/01-selftrip-030c.log`.

```
--- FAIL: TestPreviewStructureGoldenTable (0.00s)
--- FAIL: TestPreviewTableAppliesAccentToken (0.00s)
--- FAIL: TestPreviewClassLabelsCarryDistinctSemanticRoles (0.00s)
--- FAIL: TestPreviewSemanticRolesResolveToThemeTokens (0.00s)
--- FAIL: TestPreviewTableLightAndDarkAxesDiffer (0.00s)
--- FAIL: TestPreviewViewsEmitZeroANSIUnderMonochrome (0.00s)
```

Named mechanism-2 failures (AC-TUIM-030b binds these by name):

```
--- FAIL: TestPreviewSemanticRolesResolveToThemeTokens (0.00s)
    preview_presentation_test.go:151: conflict label → Danger: SGR parameter run "38;2;177;67;47" absent from tableView output (AC-TUIM-030b)
    preview_presentation_test.go:151: table border → ChromeBorder: SGR parameter run "38;2;189;186;178" absent from tableView output (AC-TUIM-030b)
    preview_presentation_test.go:151: selected row → Accent: SGR parameter run "38;2;191;101;71" absent from tableView output (AC-TUIM-030b)
    preview_presentation_test.go:151: add label → Success: SGR parameter run "38;2;61;139;110" absent from tableView output (AC-TUIM-030b)
    preview_presentation_test.go:151: update label → Info: SGR parameter run "38;2;31;122;125" absent from tableView output (AC-TUIM-030b)
    preview_presentation_test.go:151: preserve label → Dim: SGR parameter run "38;2;91;98;95" absent from tableView output (AC-TUIM-030b)
```

The trip output additionally shows the defect's own signature — the bubbles default palette re-emerging in the rendered row: `\x1b[1;38;5;212m add   templates/new_file.yaml`. Mutation reverted; `go test ./internal/cli/update/` PASS and `grep -c "table.WithStyles(previewTableStyles(th))"` → 1.

`TestPreviewStructureGoldenTable` (mechanism 1) also failed, because removing `WithStyles` additionally removed the monochrome header rule — a *structural* side effect of this particular mutation, not palette detection. The AC is satisfied by the mechanism-2 failures; mechanism 1's palette-insensitivity is proved separately by AC-TUIM-030a.

### Palette-insensitivity proof (AC-TUIM-030a) — verbatim

`internal/tui/theme.go` light `Accent` `#bf6547` → `#00ff00` and `Danger` `#b1432f` → `#ff00ff`, then `go test -count=1 ./internal/cli/update/ -run TestPreviewStructureGolden -v`:

```
--- PASS: TestPreviewStructureGoldenTable (0.00s)
--- PASS: TestPreviewStructureGoldenDiff (0.00s)
--- PASS: TestPreviewStructureGoldenFallback (0.00s)
--- PASS: TestPreviewStructureGoldensAreANSIFree (0.00s)
```

Reverted; `git status --porcelain internal/tui/` → 0 lines.

### M3 divergence audit (AC-TUIM-028 / REQ-TUIM-042)

Fields set by one factory and not the other, with disposition:

| Field | huh v1 (`moaiHuhStyles`) before | huh v2 (`moaiWizardStyles`) | Disposition |
|---|---|---|---|
| `Base` border | unset | `BorderForeground(Rule)` | **closed** — v1 now sets `BorderForeground(th.Rule)` |
| `Card` | unset | mirrors `Base` | **closed** — v1 mirrors `Base` |
| `SelectSelector` string | library default `"> "` | `"▸ "` | **closed** |
| `SelectedPrefix` | library default `"[•] "` | `"◆ "` + `Success` | **closed** |
| `UnselectedPrefix` | library default `"[ ] "` | `"◇ "` + `Muted`/`Dim` | **closed** |
| `NoteTitle` margin | unset | `MarginBottom(1)` | **closed** |
| `Next` | unset | mirrors `FocusedButton` | **closed** |
| `Blurred.Base` hidden border | inherited from `ThemeBase` | explicit `HiddenBorder()` | already equivalent — no gap |
| `Group.Title` / `Group.Description` | set | set | already equivalent — no gap |

Reverse direction (fields v1 sets that v2 does not): **none found**, so `wizard.go` required no edit — which is also why AC-TUIM-025 passes with zero hunks.

Border-token reconciliation (the axis design.md §C left open): both factories use `Theme.Rule` for the field card. `ChromeBorder` is the preview's *bubbles-table* chrome token and has no huh analogue; the preview's own card primitive (`tui.Box`) draws its non-accent border from `Rule`. `Rule` is therefore the preview-aligned choice for a form card. Recorded at the `moaiHuhStyles` doc comment so it is not re-litigated.

### Coverage (AC-TUIM-032)

`go test -count=1 -cover ./internal/cli/ ./internal/cli/update/ ./internal/merge/ ./internal/cli/wizard/ ./internal/tui/` → exit 0. Log: `.moai/state/verify/tuim/11-coverage.log`.

| Package | Re-measured baseline | After | Delta |
|---|---:|---:|---|
| `internal/cli` | 74.8% | **74.8%** | flat |
| `internal/cli/update` | 70.2% | **88.9%** | +18.7pp (new presentation tests) |
| `internal/merge` | 86.3% | **90.8%** | +4.5pp — the M2 deletion removed dead code whose uncovered branches outweighed its covered ones, so the carve-out's anticipated drop did not occur |
| `internal/cli/wizard` | 95.2% | **95.2%** | flat |
| `internal/tui` | 93.6% | **93.6%** | flat |

No package regressed.

### Deviations from the plan

1. **§C.2 method, one extra `TrimSuffix`.** The snippet's two-line extraction leaves a trailing `m` (`"38;2;191;101;71m"`), while its own comment declares the value to be `"38;2;191;101;71"`. `sgrParams` trims that terminator so the extracted value equals the documented parameter run. Both forms pass here; the trimmed form is the more robust one (it survives a foreground combined with a background, where the colour parameters are no longer CSI-terminal).
2. **M2 test reconciliation went beyond "delete both files."** Two tests in the deleted files covered *surviving* code and were retargeted per M2-3 rather than discarded: `TestMergeAnalysis_Creation` → `internal/merge/types_test.go` (plus a new named-field-literal guard mirroring `plan.go:88`), and `TestToMapInterface` → `internal/merge/strategies_test.go`, where `toMapInterface` actually lives. Deleting the latter with its file would have silently dropped coverage of live strategy code.
3. **Diff body left unpainted.** design.md §C maps "Body text → `Body` → diff body". The diff *header*, the empty-diff notice, and the key hints are painted; the diff body itself is not. It is user data whose own `+`/`-` markers carry the meaning, and wrapping it in one foreground token inside a scrolling viewport risks the content assertions that pin the diff text while buying nothing. Recorded at the `buildDiffViewContent` doc comment.
4. **Monochrome table keeps its header rule.** `previewTableStyles`'s monochrome branch retains `BorderBottom(true)` without a border colour, so the structure golden pins the table's border geometry as AC-TUIM-030a asks. Box-drawing runes are structure, not colour, so AC-TUIM-009 is unaffected.

### Pre-existing findings (not introduced, not repaired)

- `gofmt -l internal/cli/update/` lists `class_test.go`, `deploy/deploy_test.go`, `plan/plan_test.go`, `preview_test.go`. All four are unformatted **at base revision `d59936c71`** (verified by piping `git show d59936c71:<path>` through `gofmt -l`). None of the four files created or modified by this SPEC appears in the list. Left untouched per scope discipline.
- The live confirmation path (`confirmViaPreview` → `toPreviewInputs`) applies neither a file-count/path-length limit nor path-traversal sanitization. Those controls existed only inside the dead `ConfirmMerge` and were therefore already not running in production. Pre-existing, out of scope, recorded so the M2 deletion does not erase the only trace of it.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-25
run_commit_sha: "283877ed0"          # M3 / final; M1 217f0cc60, M2 1b42e1a0f
run_base_sha: "d59936c71"
run_branch: feat/SPEC-CLI-TUI-MODERNIZE-001
run_status: COMPLETE
ac_pass_count: 46
ac_fail_count: 0
ac_pass_with_debt_count: 0
must_fix_pass_count: 16
must_fix_total: 16
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a            # isolated worktree branched from origin/main d59936c71; no shared-checkout commit
l44_post_push_fetch: n/a             # NOT pushed — push and PR are the orchestrator's decision
new_warnings_or_lints_introduced: 0  # golangci-lint 0 issues; go vet exit 0
cross_platform_build:
  darwin_amd64: exit 0
  windows_amd64: exit 0
  linux_amd64: exit 0
coverage:
  internal_cli: {baseline: 74.8, after: 74.8}
  internal_cli_update: {baseline: 70.2, after: 88.9}
  internal_merge: {baseline: 86.3, after: 90.8}
  internal_cli_wizard: {baseline: 95.2, after: 95.2}
  internal_tui: {baseline: 93.6, after: 93.6}
total_run_phase_files: 13            # 4 source, 4 test, 3 golden, 1 deleted-source, 3 deleted-test → net: see below
files_added: 6                       # preview_presentation_test.go, preview_golden_test.go, merge/types_test.go, 3 goldens
files_modified: 5                    # preview_tui.go, preview_fallback.go, merge/types.go, merge/strategies_test.go, cli/huh_theme.go + huh_theme_test.go
files_deleted: 3                     # merge/confirm.go, confirm_test.go, confirm_coverage_test.go
m1_to_mN_commit_strategy: one commit per milestone; SPEC artifacts + draft→in-progress transition rode the M1 commit
pushed: false
pr_opened: false
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
