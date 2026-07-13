# SPEC-CLI-TUX-V3-002 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-07-13

## §E.2 Run-phase Evidence

### M2a — huh v2 spike verdict (REQ-TUX2-005)

**huh v2 spike verdict: SUCCESS** — the YOffset scroll defect is RESOLVED under the
huh v2.0.3 + bubbletea v2 pair. M2c takes the unified multi-group form path
(REQ-TUX2-006); plan B (REQ-TUX2-007) is NOT taken.

Reproduction evidence (spike executed in `/tmp/huh-v2-spike`, isolated module
`spike`, `charm.land/huh/v2@v2.0.3` + `charm.land/bubbletea/v2@v2.0.2`):

1. **Source-level**: the v1 defect mechanism — `updateViewportHeight()` with the
   unconditional `s.viewport.YOffset = s.selected` reset (huh v1.0.0
   `field_select.go:543` and `:203`, forced into effect by the OptionsFunc
   `s.height = defaultHeight` path at `:235-236`) — is REMOVED in v2.0.3.
   Replacement: `ensureVisible()` (v2.0.3 `field_select.go:656-668`) scrolls the
   viewport "the minimum amount so that the region [offset, offset+height) is
   within the visible area" — cursor-out-of-view clamping only, never an
   unconditional snap-to-top.
2. **Reproduction-level**: a programmatic multi-group form (2 groups, 9 fields —
   input + selects + confirm, matching the wizard's 7-9 visible-question
   envelope) driven with `tea.KeyPressMsg`/`tea.WindowSizeMsg` at 80x40 AND
   80x12. After `KeyDown` inside a 3-option select (cursor high→medium), the
   option ABOVE the cursor ("high") **remains visible in the rendered frame in
   both terminal sizes**. The v1 defect (options above cursor hidden) did not
   reproduce. Frames archived at `/tmp/huh-v2-spike/frames.txt` (run exit=0).
3. **API compatibility probe** (all wizard usage patterns compile + behave):
   `TitleFunc`/`DescriptionFunc(fn, binding)` OK; `Validate`-as-save OK (saved
   answers harvested: `map[development_mode:tdd model_policy:medium
   plan_type:subscription project_name:spike]`); `Group.WithHideFunc` OK;
   `huh.ErrUserAborted` OK (v2 `form.go:55`); `WithAccessible` OK;
   `huh.NewOption` OK. **API delta**: `WithTheme` now takes a `huh.Theme`
   interface (`Theme(isDark bool) *Styles`) — adapt via `huh.ThemeFunc`; the v2
   `Styles`/`FieldStyles` field set matches the v1 `Theme.Focused/Blurred`
   fields the wizard theme sets; lipgloss v2 drops `AdaptiveColor` — the
   `isDark` parameter supplies the light/dark axis (maps directly onto
   `tui.LightTheme()`/`tui.DarkTheme()` tokens).
4. **Known v2 behavior (not the defect)**: on height-constrained terminals the
   group viewport anchors the FOCUSED FIELD to the viewport top
   (`group.go buildView()` → `SetYOffset(focused-field offset)`), so fields
   above the focused field scroll out when the page does not fit. This is
   standard scroll-into-view behavior (viewport clamps to 0 when content fits),
   not the v1 defect.

### M2a — bubbletea/bubbles v2 adoption (REQ-TUX2-012)

Pinned `charm.land/bubbletea/v2 v2.0.8` + `charm.land/bubbles/v2 v2.1.1` as
direct dependencies (I-3 animated-spinner prerequisite; import-usage lands in
M2d printer backend). huh major follows the verdict above: v2 (M2c).

### M2c — unified multi-group wizard (REQ-TUX2-006/008/009)

Verdict=SUCCESS path taken (single multi-group form). Plan B NOT taken → the
AC-TUX2-006 plan-B leg is N/A.

- `internal/cli/wizard` migrated huh v1.0.0 → `charm.land/huh/v2 v2.0.3`. One
  `huh.Form` per wizard run: consecutive unconditional questions sharing a
  `Question.Group` label merge into one multi-field page ("Project" 4 fields /
  "Git" / "Options" / "Advanced"); each conditional question is its own group
  whose `WithHideFunc` wraps the original `Condition` (huh evaluates hide funcs
  lazily at navigation time; field Blur runs Validate→saveAnswer first, so
  same-form conditions work — verified by `TestUnifiedForm_ConditionalGroupsAppear`).
- Stepper (REQ-TUX2-008): `wizardTotalSteps = 6` constant REMOVED (non-test
  grep 0). Denominator single dynamic source = `stepperDenominator` →
  `TotalVisibleQuestions`; rendered by a skip-focus `huh.NewNote` TitleFunc
  bound to the result struct (hashstructure deep-hash re-eval). Measured
  denominators: manual 5 / personal+github 8 / personal+gitlab 9 / standard 12
  — the audit's "7~9 variation" envelope is covered by the 8/9 states plus the
  5/12 mode extremes.
- Behavior preservation (REQ-TUX2-009): question set / defaults / validation /
  `WizardResult` schema unchanged (additive `Question.Group` label only).
  All-conditions-false path preserved (returns without running the form).
  Locale rendering preserved via `TitleFunc/DescriptionFunc(fn, locale)`
  (same binding pattern as v1). Existing wizard suite green.
- **Changed tests (individually documented per REQ-TUX2-009 escape)**:
  1. `TestCharacterize_WizardTotalSteps` → superseded by
     `TestStepperTotal_DynamicDenominator` (pre-authorized, plan.md §B #11 —
     the test asserted the removed constant; replacement asserts the dynamic
     denominator).
  2. Theme characterization tests (`SelectedPrefix`/`UnselectedPrefix`/
     `SelectSelectorPrefix`/`BlurredInheritsFromFocused`/`UsesDeepTealForTitle`)
     now read `moaiWizardStyles(isDark)` — the resolved style set behind the
     huh v2 `Theme` interface (v1 exposed a struct; v2 exposes
     `Theme(isDark bool) *Styles`). Assertions (glyphs ◆/◇/▸, bold, coral)
     unchanged; light+dark axes both asserted.
  3. Added `TestCharacterize_WizardTokens_PrimaryIsCoral` (v2 token-path
     parity with the legacy `wizardColors` AdaptiveColor path, which remains
     for the exported `Styles`/`NewStyles` v1-lipgloss surface).
- huh major state after M2c: wizard on huh/v2 (direct); `github.com/charmbracelet/huh
  v1.0.0` RETAINED for `update.go` (M3 PRESERVE scope) + `profile_setup.go` +
  the init profile-setup confirm — dual-major coexistence per plan.md §B #7
  (M1 REQ-CTX-002 precedent). bubbletea/v2 now direct (wizard tests);
  bubbles/v2 becomes direct with the M2d printer backend.

### M2b — deferred self-update order (REQ-TUX2-001..004)

TDD RED→GREEN. `runInit` no longer performs any network call before the first
wizard interaction: the pre-wizard `runBinaryUpdateStep` + re-exec block was
removed; `init_update_notice.go` starts a CHECK-ONLY goroutine after wizard
completion / first phase output, flushes a stderr notice with the
`moai update` hint at exit under a bounded grace (1s default, injectable),
never installs, never re-execs (static grep guard pins init.go + the notice
impl free of re-exec references). Injectable seams: `deferredUpdateEnabled`,
`deferredUpdateCheck`, `isInteractiveStdin`, `runWizardFn`. Skip semantics
characterized through the deferred path (templates-only flag /
`EnvSkipBinaryUpdate` / dev-build). Wizard-cancel path returns before the
check starts (zero network side effects on cancel). Explicit trade-off per
plan.md §G: no automatic re-run — the notice names `moai update` for template
refresh.

### M2d — animated spinner + warning collector + completion card (REQ-TUX2-010..014/016)

- Printer Spinner handle: bubbles v2 `spinner.MiniDot` frame-cycling animator
  (frame 0 == legacy static glyph). Animation gated on motion-allowed TTY +
  real-terminal stderr (or the `WithAnimatedHandles` test override); Done/Fail
  signals AND joins the goroutine before the final line — zero frames after
  the terminal event, `-race` clean. Printer interface method set unchanged.
- Fallback matrix (REQ-TUX2-011): non-TTY plain / NO_COLOR-forced plain /
  MOAI_REDUCED_MOTION → zero `\x1b[` + max one static glyph (pinned by
  `TestPlainFallback_NoANSINoAnimation` + existing REQ-CTX-007 tests).
- init wires `newSpinnerReporter` → PhaseExecutor template deployment renders
  through the animated Spinner handle. **Scope note**: `ProgressReporter`
  (internal/core/project) has no per-file deploy events, and core/template
  packages are outside this SPEC's commit envelope (spec.md §E) — the live
  `k/N files` readout capability is delivered and pinned at the handle level
  (`TestProgressLive_FileCountUpdates`, `TestSpinnerAnimated_UpdateCarriesLiveLabel`);
  per-file wiring into PhaseExecutor is deferred to a core-scoped follow-up.
- Warning collector: `warnCollector` wraps the init Printer; every `Warn`
  recorded + streamed; executor result warnings `Collect()`-ed (summary-only);
  consolidated stderr summary panel emitted exactly once at exit (defer —
  success or failure); zero warnings → no panel; stdout never carries warning
  text.
- Completion card: `cd <project>` → `moai cc` → `/moai plan` next-action
  sequence + one-line stderr-summary pointer when warnings exist.

### M2e — redirect hint + regression matrix (REQ-TUX2-015/017/018)

- Existing-project re-init without `--force` now returns the error with a
  `moai update` redirect hint (detection at the init.go executor-error site;
  `internal/core/project/validator.go` untouched — outside commit envelope).
- Ratchet (REQ-TUX2-018): pre-flight baseline **40** (re-measured 2026-07-14,
  matches the 2026-07-13 SPEC figure) → post-run **38** (< baseline; the two
  `fmt.Println(tui.Stepper(...))` wizard call sites migrated into the unified
  form's Note-rendered stepper).
- Cross-platform: `GOOS=windows` + `GOOS=linux` builds exit 0.
- `NO_COLOR=1 go run ./cmd/moai init --help 2>&1 | grep -c $'\x1b'` → `0`.

### AC matrix (verbatim command outcomes; logs at `.moai/state/verify/tux-v3-002/`)

| AC | Command (actually run) | Outcome |
|----|------------------------|---------|
| AC-TUX2-001 | `go test ./internal/cli/ -run 'InitNetworkOrder\|NoNetworkBeforeWizard' -count=1 -v` + existence grep | PASS — `--- PASS: TestInitNoNetworkBeforeWizard (0.13s)`; grep count 1 |
| AC-TUX2-002 | `go test ./internal/cli/ -run 'DeferredUpdateNotice' -count=1 -v` | PASS — 4/4 (`StderrHintOnly`, `NoReexecReference`, `NotAvailableSilent`×subtests, exit 0) |
| AC-TUX2-003 | `go test ./internal/cli/ -run 'SkipBinaryUpdate' -count=1 -v` | PASS — 7/7 incl. `TestSkipBinaryUpdate_DeferredCheckSkipped` (templates-only/env/dev-build) |
| AC-TUX2-004 | `go test ./internal/cli/ -run 'InitNonInteractive.*Update\|NonTTYUpdateCheck' -count=1 -v` | PASS — 2/2 (`NonTTYUpdateCheckNonBlocking`, `InitNonInteractiveDeferredUpdateNotice`) |
| AC-TUX2-005 | `grep -n 'huh v2 spike verdict' .moai/specs/SPEC-CLI-TUX-V3-002/progress.md` | PASS — 2 matches (lines 10/12); verdict committed in M2a `681574852`, before M2c `d5f04baa4` |
| AC-TUX2-006 | `go test ./internal/cli/wizard/ -run 'UnifiedForm\|MultiGroup' -count=1 -v` | PASS — 3/3 (verdict=success leg; plan-B leg N/A per §E.2 M2a verdict) |
| AC-TUX2-007 | non-test `wizardTotalSteps` grep + `go test ./internal/cli/wizard/ -run 'StepperTotal\|VisibleQuestions' -count=1 -v` | PASS — grep `0`; 3/3 tests (denominators 5/8/9/12 dynamic) |
| AC-TUX2-008 | `go test ./internal/cli/wizard/ -count=1` | PASS — `ok … internal/cli/wizard 2.810s` (locale/standard/advanced suite green; changed tests documented in §E.2 M2c) |
| AC-TUX2-009 | `go list -m charm.land/bubbletea/v2 charm.land/bubbles/v2` + non-test import grep + `go mod why charm.land/bubbletea/v2` | PASS — v2.0.8 + v2.1.1 resolve; non-test import·usage 1 (`printer` → `bubbles/v2/spinner`); `go mod why` chain `printer → bubbles/v2/spinner → bubbletea/v2`; both direct in go.mod |
| AC-TUX2-010 | `go test ./internal/cli/printer/ -run 'SpinnerAnimated\|ProgressLive' -count=1 -v` + printer bubbles/v2 grep | PASS — 5/5; grep 1 (non-test import + use) |
| AC-TUX2-011 | `go test ./internal/cli/printer/ -run 'ReducedMotion\|NoANSI\|PlainFallback' -count=1 -v` | PASS — 3/3 (zero `\x1b[`, single static frame) |
| AC-TUX2-012 | `go test ./internal/cli/ -run 'WarningCollector\|WarningSummary' -count=1 -v` | PASS — 2/2 (N-message panel exactly once, no duplicate emission) |
| AC-TUX2-013 | `go test ./internal/cli/ -run 'InitStdoutClean\|WarningChannel' -count=1 -v` | PASS — 2/2 (stdout warning-free) |
| AC-TUX2-014 | `go test ./internal/cli/ -run 'CompletionCard\|NextActions' -count=1 -v` | PASS — 1/1 (`moai cc` + `/moai plan` + warning pointer) |
| AC-TUX2-015 | `go test ./internal/cli/ -run 'ExistingProject.*Hint\|UpdateRedirect' -count=1 -v` | PASS — 1/1 (`moai update` hint + `--force` retained) |
| AC-TUX2-016 | `go test ./... -count=1` + `golangci-lint run --timeout=5m` + `GOOS=windows`/`GOOS=linux` builds | PASS — 100/100 `ok`, `0 issues.` (NEW 0 vs pre-flight `0 issues.` baseline), both builds exit 0 |
| AC-TUX2-017 | `grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' \| grep -v '_test.go' \| wc -l` | PASS — **38** < baseline 40 |
| AC-TUX2-018 | `NO_COLOR=1 go run ./cmd/moai init --help 2>&1 \| grep -c $'\x1b'` | PASS — `0` (grep exit 1, no matches) |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-14
run_commit_sha: pending-backfill-m2e
run_status: complete
ac_pass_count: 18
ac_fail_count: 0
preserve_list_post_run_count: 0   # statusline / merge/confirm.go / update.go sync pipeline / uikit/banner.go / other SPEC dirs untouched
l44_pre_commit_fetch: not-applicable   # push deferred to orchestrator per delegation contract
l44_post_push_fetch: not-applicable
new_warnings_or_lints_introduced: 0    # golangci-lint 0 issues (baseline 0)
cross_platform_build:
  darwin: exit 0
  windows_amd64: exit 0
  linux_amd64: exit 0
total_run_phase_files: 14   # 9 source/test files + go.mod + go.sum + spec.md frontmatter + progress.md (+ verify logs, gitignored)
m1_to_mN_commit_strategy: per-milestone commits M2a..M2e on main (Route A Hybrid Trunk; push deferred to orchestrator)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
