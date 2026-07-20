# SPEC-CLI-TUX-V3-004 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-07-13

## §F Phase 4 Mode Selection

- Inputs: tier=M, scope ~10-15 files (internal/cli + uikit), domains=1 (Go CLI rendering), language=Go, concurrency benefit LOW (coding-heavy).
- Evaluation: trivial no (multi-milestone) / background no (write-capable) / agent-team RETIRED / parallel no (single domain, coding-heavy) / workflow no (<30 files, non-uniform transforms) / sub-agent selected.
- Decision: sub-agent
- Justification: Coding-heavy single-domain Tier M SPEC per Anthropic coding-task parallelism caveat — sequential manager-develop (cycle_type=tdd) per milestone is the default and correct envelope. Plan-audit iter-1 PASS-WITH-DEBT 0.87 (2026-07-20); Implementation Kickoff Approval obtained (user selected run-phase entry).

## §E.2 Run-phase Evidence

### Pre-flight (plan.md §C #1-#7, executed 2026-07-20)

- **#1 baseline**: branch `main`, spawn-time HEAD `db70597a2f07cb83e3188413d6dcbc5d3eea5e94`; re-based mid-run to `319c3e93e`/`dd0a6ce72` as the parallel session (SPEC-MODEL-PROFILE-MATRIX-001 M2/M3 + README restructure) landed commits. Divergence re-checked before every commit (`git rev-list --count --left-right origin/main...HEAD` = `0 0`).
- **#2 depends_on (strict completed)**: SPEC-CLI-TUX-V3-002 `status: completed`, SPEC-CLI-TUX-V3-003 `status: completed` — PASS.
- **#3 cross-platform build + lint baseline**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=5m` → `0 issues.` (lint baseline = 0; any finding at M4e is NEW). Logs: `.moai/state/verify/tux4/preflight-{build-native,build-win,lint}.log`.
- **#4 golden green anchor**: `go test ./internal/cli/ -run 'Golden' -count=1` → exit 0 (`[no tests to run]` — no existing test name contains "Golden"; existing characterization tests are `TestDoctor_Current_*`/`TestStatus_Current_*`, verified green in the full-package baseline). `go test ./internal/cli/uikit/ -count=1` → ok (exit 0).
- **#5 ratchet inventory re-measure (audit-debt D1/D2 resolution)**: raw grep = **14** (plan.md "전체 40" was stale). Breakdown: `internal/cli/uikit/banner.go` ×12 (in-scope, REQ-TUX4-006), `internal/cli/branch_protection.go:44` ×1 (out-of-scope, see D2), `internal/cli/worktree/new.go:433` ×1 (commented-out line, see D1). Inventory files: `.moai/state/verify/tux4/preflight-ratchet-{raw,nocomment}.txt`.
  - **D1 resolution**: the series-final ratchet grep (AC-TUX4-010) appends the comment-exclusion idiom `| grep -v '^[^:]*:[0-9]*:[[:space:]]*//'` (same as AC-TUX4-005). The out-of-scope comment in `worktree/new.go:433` is NOT deleted to satisfy a naive grep. Comment-excluded count at pre-flight = **13**.
  - **D2 resolution**: `internal/cli/branch_protection.go:44` (interactive y/N `fmt.Printf`) is owned by deferred SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 and is OUTSIDE this SPEC's §E commit scope. The final ratchet-0 gate is scoped to the §E surface set (doctor/status/spec/root/uikit + pre-flight #5 inventory minus branch_protection.go). Residual: **PASS-WITH-DEBT** — 1 remaining call site, deferred to its owning SPEC. branch_protection.go is not modified by this SPEC.
- **#6 glamour version + dependency graph**: latest glamour = `v1.0.0`. Its go.mod requires `github.com/charmbracelet/lipgloss v1.1.1-0.20250404203927-76690c660834` (**lipgloss v1 line**). The repo graph ALREADY carries the v1/v2 coexistence (`github.com/charmbracelet/lipgloss v1.1.0` direct + `charm.land/lipgloss/v2 v2.0.5`), so glamour introduces NO new major — it bumps the existing v1 entry to v1.1.1-pre. No graph re-expansion; pinned `glamour@v1.0.0`.
- **#7 coverage baseline (REQ-TUX4-011 downgrade decision)**: `internal/cli` **74.3%**, `internal/template` **85.2%**, `internal/hook` **83.5%** (sub-packages 79.5-100%). **All three below 90%** → per REQ-TUX4-011's built-in relief, the gate degrades to **strict non-regression** for all three packages, gap recorded here: cli −15.7pt, template −4.8pt, hook −6.5pt vs the 90% target. Log: `.moai/state/verify/tux4/preflight-cover.log`.

### M4a — glamour introduction + style decision (REQ-TUX4-004, AC-TUX4-005)

- **Dependency**: `go get github.com/charmbracelet/glamour@v1.0.0` + import in `internal/cli/glamour_style.go` + `go mod tidy` → direct requirement in go.mod. go.mod/go.sum diff verified pure (glamour-induced entries only) before staging — shared-file discipline with the parallel session.
- **Style file path (AC-TUX4-005 convention)**: `internal/cli/glamour_style.go` — matches the acceptance.md path convention exactly; no substituted verification command needed.
- **Token mapping decision** (no hex — all colours reference `tui.Theme` fields): Document/Text/Item/Enumeration/CodeBlock/Table → `Body`; Heading/H1-H3 → `Accent` (bold); Strong → `Fg` (bold); Emph → `Body` (italic); Link/LinkText/inline Code → `Info` (link underlined); BlockQuote → `Dim`; HorizontalRule → `Rule`.
- **Render routing (REQ-TUX4-005)**: rich glamour path only when NO_COLOR unset AND destination writer is a character-device `*os.File`; all other cases (pipes, CI, golden-test buffers) are byte-stable plain markdown passthrough. Fixed word-wrap width 100 (`glamourWordWrap`) — no live terminal-width probing, keeping golden line-wrap stable (acceptance §C width edge).
- **Shared render symbols (AC-TUX4-004 wiring record)**: `renderMarkdown(w, md)` is the glamour-mediated gateway; `glamourRender(md, theme)` is the always-rich body; `markdownRichEnabled(noColor, tty)` is the routing predicate. M4b wires `status.go` and `spec_view.go` through these symbols.
- **TDD**: RED `internal/cli/glamour_style_test.go` (token-mapping, light/dark divergence, routing matrix, passthrough byte-stability, rich-path ANSI presence) → GREEN `internal/cli/glamour_style.go`.
- **M4a commit**: `3b3b397cc` (pushed to main). Verification performed in an isolated scratch worktree at HEAD because the parallel session (SPEC-MODEL-PROFILE-MATRIX-001) held the shared working tree in a transiently broken mid-refactor state (`TierProfileEntry`/`applyUpdateTierProfile` undefined in THEIR uncommitted files); my staged paths verified independent of that WIP.

### M4b — status/spec glamour render (REQ-TUX4-004/005, AC-TUX4-004~006)

- **Wiring (AC-TUX4-004 symbol record)**: `internal/cli/status.go` `runStatus` composes its data as markdown and calls `renderMarkdown(out, md)`; `internal/cli/spec_view.go` `viewAcceptanceCriteria` calls `renderTreeMarkdown(out, header, tree)`. Both symbols live in `internal/cli/glamour_style.go` and route through `glamourRender` (glamour TermRenderer) on the rich path — shared-helper placement, so the AC grep uses the symbols `renderMarkdown` / `renderTreeMarkdown` (not a file-restricted grep).
- **status surface change (render-layer swap)**: tui.Box/Section/KV/Pill composition → markdown document (`# Project Status`, `## Project`, `## Configuration`, `**Status**` line). Every data field preserved: project name, ADK version, config path, SPEC count, section-file count, initialized/not-initialized status (§D behavior-preserving on the data surface). Non-TTY/NO_COLOR = plain markdown passthrough (REQ-TUX4-005); TTY = glamour rich render.
- **spec view surface**: non-TTY output byte-stable vs pre-M4b (`header\n\ntree` — no fences, no ANSI, verified by `TestSpecViewPlain_TreePassthrough`); TTY wraps the tree in a text fence under glamour so glyphs stay monospace. `printTree` kept as thin wrapper over new writer-based `fprintTree` (existing test compatibility).
- **TDD**: RED `internal/cli/status_specview_render_test.go` (`TestStatusGolden_MarkdownSurface`, `TestStatusGolden_NoColorByteIdentical`, `TestStatusGolden_NotInitialized`, `TestSpecViewPlain_TreePassthrough`, `TestSpecViewPlain_CommandUsesGlamourGateway`) → GREEN wiring above.
- **Golden changes (per-file rationale — REQ-TUX4-008 discipline)**:
  - `internal/cli/testdata/status-light.golden`: REGENERATED — render-layer swap Box→markdown per REQ-TUX4-004; light theme affects only the TTY glamour path, so the non-TTY capture is plain markdown.
  - `internal/cli/testdata/status-dark.golden`: REGENERATED — same rationale; dark theme invisible off-TTY.
  - `internal/cli/testdata/status-nocolor.golden`: REGENERATED — same rationale; NO_COLOR forces the same plain passthrough. All three goldens are now byte-identical (md5 `76c6e564…`) — expected: theme env vars cannot change non-TTY bytes under REQ-TUX4-005, and `TestStatusGolden_NoColorByteIdentical` pins this invariant.
- **Verification**: full `internal/cli` + `internal/cli/uikit` suites green in the isolated worktree (`ok 18.896s` / `ok 0.415s`); live `NO_COLOR=1 go run ./cmd/moai status | grep -c $'\x1b'` = 0 (AC-TUX4-006 second half).
- **M4b commit**: `1317b9567` (pushed to main).

### M4c — doctor live progress + result table (REQ-TUX4-001~003, AC-TUX4-001~003, AC-TUX4-014)

- **Progress reporter seam (AC-TUX4-014)**: `checkObserver` type + `runGroupedChecksObserved(verbose, filter, obs)` in doctor.go; `runGroupedChecks` delegates with nil observer (legacy callers unchanged). The observer wraps each check thunk exactly once and returns the result unchanged — verdict functions (`checkGit`, `checkGoRuntime`, …) are byte-identical, pinned by `TestDoctorLiveProgress_SeamPreservesVerdicts` (observer-vs-baseline verdict equality).
- **Live progress (REQ-TUX4-001)**: `runDoctor` builds `printer.New(WithWriters(stdout, stderr))` and emits one `Step` per check on stderr — TTY: in-place erase-line updates; non-TTY/NO_COLOR: printer ModePlain line-per-event, zero ANSI. Fail status → `StepHandle.Fail`.
- **Result table (REQ-TUX4-002)**: new `internal/cli/doctor_render.go` — per-section pass/fail table + per-section counts (`N ok, N warn, N fail`) + overall Pass/Warn/Fail pills inside the System Diagnostics box. Rich path: bubbles v2 table (`charm.land/bubbles/v2/table`) styled from tui tokens (AC-TUX4-002 grep target: `doctor_render.go:17`); plain path: aligned plain-text columns (STATUS/CHECK/MESSAGE). Discovery: the bubbles v2 table viewport defaults to width 0 and renders no rows — `WithWidth` is load-bearing (commented in source).
- **Render separation**: the render block moved out of doctor.go into doctor_render.go so the doctor.go diff is limited to the seam + printer wiring + render delegation (AC-TUX4-014 proof surface for `git diff -- internal/cli/doctor.go`).
- **TDD**: RED `internal/cli/doctor_render_test.go` (`TestDoctorLiveProgress_SeamPreservesVerdicts`, `TestDoctorStep_ProgressOnStderr`, `TestDoctorTable_PlainAlignedColumns`, `TestDoctorSectionResult_Counts`, `TestDoctorTable_RichUsesBubblesTable`) → GREEN seam + renderer. RED interim state observed: rich-table test failed on the width-0 viewport before the `WithWidth` fix.
- **Golden test renames (AC-TUX4-003 run-pattern alignment)**: `TestDoctor_Current_Light`→`TestDoctorGolden_Light`, `TestDoctor_Current_Dark`→`TestDoctorGolden_Dark`, `TestDoctor_NoColor`→`TestDoctorGolden_NoColor` (now also asserting zero ANSI on BOTH channels per REQ-TUX4-003). `captureDoctorCmd` split into (stdout, stderr) so goldens pin the stdout result surface while stderr progress is asserted separately.
- **Golden changes (per-file rationale)**:
  - `internal/cli/testdata/doctor-light.golden`: REGENERATED — CheckLine rows → per-section aligned table + section counts (REQ-TUX4-002 render swap); data set (check names/messages/statuses) unchanged.
  - `internal/cli/testdata/doctor-dark.golden`: REGENERATED — same rationale (light/dark identical off-TTY).
  - `internal/cli/testdata/doctor-nocolor.golden`: REGENERATED — same rationale + zero-ANSI assertion added in the test body.
- **Verification**: contract tests + full `internal/cli`/`uikit` suites green in the isolated worktree (`ok 17.509s` / `ok 0.784s`); live `NO_COLOR=1 go run ./cmd/moai doctor | grep -c $'\x1b'` = 0; bubbles-v2 reachability grep = `doctor_render.go:17` (≥1).
- **M4c commit**: `2b886c39b` (pushed to main).

### M4d-1 — compact banner (REQ-TUX4-006, AC-TUX4-007) + help render path verdict (REQ-TUX4-007, AC-TUX4-008 precedence record)

- **Banner**: large ASCII logo + "Version:" label retired. New surface = 2 lines: `◆ MoAI-ADK <tagline>` identity (accent bold + body dim, ◆ from the §D glyph whitelist) + pill row `[vX] [go X] [claude X]`. `PrintBanner`/`PrintWelcomeMessage` keep their signatures (call sites in init.go/update.go are PRESERVE-frozen under the parallel-session envelope) and route through `printer.New().Data(...)` — the Printer gateway absorption; `grep -c 'fmt\.Print' internal/cli/uikit/banner.go` = **0** (was 12).
- **TDD**: RED `internal/cli/uikit/banner_compact_test.go` (`TestCompactBanner_{TwoLineIdentity,NoASCIILogo,GlyphWhitelist,BrandTagline}`, `TestBannerPill_{Metadata,ClaudeFallback}`) → GREEN `bannerString`/`welcomeString` + Printer routing. Legacy `TestPrintBanner_OutputFormat` expectation updated ("Version" label → `v1.2.3` pill; SPEC-sanctioned redesign).
- **Golden changes (per-file rationale)**: `uikit/testdata/banner-current-{light,dark,nocolor}.golden` REGENERATED — ASCII logo → compact 2-line identity (REQ-TUX4-006); `uikit/testdata/welcome-current-{light,dark,nocolor}.golden` REGENERATED — same content routed through Printer.Data (trailing blank-line shape normalized; text unchanged).
- **help render path verdict (recorded BEFORE any reorder/golden commit — AC-TUX4-008(1))**: **(a) keep-fang**. Evidence + rationale:
  - Live surface re-verified 2026-07-20: fang v2 renders cobra Groups as uppercase headers — actual literals `LAUNCH COMMANDS:` / `PROJECT COMMANDS:` / `TOOLS:` (+ ungrouped `COMMANDS` section). Matches spec.md §A.4 재실측.
  - `renderRootHelpTUI` (help.go:101) is confirmed shadowed in production: `runFang` installs fang's own helpFunc (`fang.go:130-140` writes via `colorprofile.NewWriter(c.OutOrStdout())`), so the custom 4-group surface renders only for in-process cobra `Execute()` test paths. Reviving it would trade a maintained, token-styled fang surface for a hand-rolled one + 4-group→3-group reconciliation — no UX gain, higher risk.
  - Option (c) fang-customize: fang v2.0.1's public options are Version/Manpage/ColorScheme/ErrorHandler (+ completion) — **no help-layout/group-ordering API exists**; plan-B not required because ordering is controllable at the cobra layer: fang's help iterates `c.Commands()` verbatim (`help.go:420`), and cobra's alphabetical sort is gated by `cobra.EnableCommandSorting`. Reorder point = cobra command registration order with sorting disabled.
  - Adopted-surface header literals for AC-TUX4-008(2): `LAUNCH COMMANDS:`, `PROJECT COMMANDS:`, `TOOLS:`. Non-adopted paths (revive / customize) header checks: **N/A** (-002 M2a spike pattern mirror).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
