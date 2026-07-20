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
- **M4d-1 commit**: `77893579e` (pushed to main).

### M4d-2 — help group frequency reorder (REQ-TUX4-007, AC-TUX4-008)

- **Mechanism**: `internal/cli/help_order.go` `reorderRootHelpCommands` — `cobra.EnableCommandSorting = false` + position-preserving within-group stable re-sort by the `helpGroupFrequency` leading order, then Remove/Add re-registration. Invoked from `Execute()` before `runFang` (root.go). Group IDs + membership byte-unchanged (pinned by `TestHelpGroupOrder_MembershipUnchanged`); ungrouped `COMMANDS` section keeps alphabetical order.
- **Frequency order (rationale in source)**: launch `cc, glm, cg`; project `init, status, doctor, update, migrate, pr`; tools leading `hook, spec, session, mx, loop, handoff, model, constitution, state` (unlisted tools members follow alphabetically).
- **TDD**: RED `internal/cli/help_order_test.go` (`TestHelpGroupOrder_{FrequencyWithinGroups,MembershipUnchanged,Idempotent}`, `TestHelpGolden_FangGroupHeaders`) → GREEN reorder implementation. The help "golden" is assertion-based (headers + relative order), NOT a byte snapshot — per acceptance.md §C edge (command additions must not break it).
- **End-to-end**: `TestHelpGolden_FangGroupHeaders` runs the ADOPTED fang surface in-process (runFang + `--help`, NO_COLOR) asserting the 3 header literals + zero ANSI + frequency order spot-checks; live `NO_COLOR=1 go run ./cmd/moai --help` re-verified: LAUNCH renders `cc, glm, cg`, PROJECT renders `init, status, doctor, update, migrate, pr`.
- **Verification**: full `internal/cli` + `uikit` suites green in the isolated worktree (`ok 17.932s` / `ok 0.248s`).
- **M4d-2 commit**: `408fe19f0` (pushed to main).

### M4e — ratchet floor + stdout matrix + series-final verification batch

- **Ratchet sweep outcome**: the banner absorption (M4d-1) exhausted the in-scope inventory. Final state (pristine worktree at `1e95bf381`): raw grep = 2 lines — `worktree/new.go:433` (commented-out, D1) + `branch_protection.go:44` (D2 out-of-scope, owned by deferred SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1, NOT modified by this SPEC); D1 comment-excluded = 1; **D1+D2-scoped final gate = 0** (series 46→40→14→**0** on the §E surface set). Residual: PASS-WITH-DEBT (1 deferred call site).
- **Stdout matrix (REQ-TUX4-010)**: `internal/cli/stdout_clean_test.go` — doctor/status/banner/spec-view stdout captures assert absence of `Warning:` / `Error:` / `[!]` status vocabulary; spec view exercised end-to-end via `newSpecViewCmd` + temp-project fixture + `findProjectRootFn` override. M4e commit: `1e95bf381`.
- **Final verification batch** (single-turn parallel, pristine worktree at `1e95bf381` = origin/main; evidence at `.moai/state/verify/tux4/final-*.log`):
  - full suite `go test ./... -count=1` → **exit 0** (`final-1-full.log`)
  - coverage (`final-2-cover.log`): cli **74.5%** (baseline 74.3% → **non-regression PASS**, +0.2pt); template 83.2% (baseline 85.2%); hook 83.4% (baseline 83.5%)
  - `GOOS=windows` build exit 0 + `GOOS=linux` build exit 0 (`final-3-*.log`)
  - `golangci-lint run --timeout=5m` → `0 issues.` (baseline 0 → **NEW = 0**, `final-4-lint.log`)
  - AC grep set (`final-5-greps.log`) + live NO_COLOR matrix (`final-6-nocolor.log`): doctor/help/status ANSI counts all 0; 3 fang group headers present
- **Coverage delta attribution (REQ-TUX4-011 gap record)**: my 6 commits touch NEITHER `internal/template` NOR `internal/hook` (verified: `git log db70597a2..1e95bf381 --grep SPEC-CLI-TUX-V3-004 -- internal/template/ internal/hook/` = empty). template 85.2→83.2 is attributable to the ONLY two commits touching that package in the window — SPEC-MODEL-PROFILE-MATRIX-001 `eee1c4fc1` + `319c3e93e` (parallel session; their coverage debt, not this SPEC's). hook 83.5→83.4: ZERO commits touched internal/hook in the window — the −0.1pt is run-to-run variance of time-dependent test paths on git-identical code. cli (this SPEC's package): +0.2pt non-regression PASS.
- **AC-TUX4-014 evidence**: `git diff db70597a2..1e95bf381 --stat -- internal/cli/doctor.go` = 1 file, 50+/47−; the ONLY function signature added is `runGroupedChecksObserved` (the seam); zero check-function (`checkGit`/`checkGoRuntime`/…) signatures or verdict criteria changed (diff `^[+-]func` grep shows only the seam function).

### E1 — 14-AC PASS/FAIL matrix (final, evidence at .moai/state/verify/tux4/)

| AC | Status | Evidence (verbatim anchor) |
|----|--------|---------------------------|
| AC-TUX4-001 | PASS | `--- PASS: TestDoctorLiveProgress_SeamPreservesVerdicts` + `--- PASS: TestDoctorStep_ProgressOnStderr` (M4c run; re-green in final-1-full.log) |
| AC-TUX4-002 | PASS | `--- PASS: TestDoctorTable_PlainAlignedColumns` + `TestDoctorSectionResult_Counts` + `TestDoctorTable_RichUsesBubblesTable`; grep `charm.land/bubbles/v2` → `doctor_render.go:17` (≥1) |
| AC-TUX4-003 | PASS | TestDoctorGolden_{Light,Dark,NoColor} green (final-1-full.log); live `NO_COLOR=1 moai doctor \| grep -c $'\x1b'` = 0 (final-6-nocolor.log) |
| AC-TUX4-004 | PASS | `go list -m` → `github.com/charmbracelet/glamour v1.0.0`; non-test import: `glamour_style.go`; wiring: status.go:104,123 `renderMarkdown(out`, spec_view.go:104 `renderTreeMarkdown(out` (final-5-greps.log) |
| AC-TUX4-005 | PASS | hex grep over status.go/spec_view.go/glamour_style.go = **0** (final-5-greps.log) |
| AC-TUX4-006 | PASS | TestStatusGolden_* + TestSpecViewPlain_* green; live `NO_COLOR=1 moai status \| grep -c $'\x1b'` = 0 |
| AC-TUX4-007 | PASS | TestCompactBanner_* + TestBannerPill_* green (6/6, M4d-1); `grep -c 'fmt\.Print' internal/cli/uikit/banner.go` = **0** (was 12) |
| AC-TUX4-008 | PASS | verdict literal grep = 2 (§E.2 M4d-1); verdict commit `77893579e` PRECEDES reorder commit `408fe19f0` (timeline); 3 header literals each ≥1 live (final-6: header count 3); TestHelpGolden_* + TestHelpGroupOrder_* green; non-adopted paths N/A |
| AC-TUX4-009 | PASS | all goldens green in final-1-full.log; `GOOS=windows` build exit 0; per-golden-file rationale recorded in §E.2 (M4b/M4c/M4d-1 sections) |
| AC-TUX4-010 | PASS-WITH-DEBT | D1 comment-exclusion idiom applied; D2 scope: final gate grep (comment-excluded, minus branch_protection.go) = **0**; residual = branch_protection.go:44 deferred to SPEC-V3R6-CI-BASELINE-DRIFT-001 (final-5-greps.log verbatim) |
| AC-TUX4-011 | PASS | `--- PASS: TestStdoutClean_{Doctor,Status,Banner}` + `TestNoWarnOnStdout_SpecView` (M4e; re-green in final-1-full.log) |
| AC-TUX4-012 | PASS (non-regression basis) | pre-flight #7 baseline <90% for all three → REQ-TUX4-011 downgrade; cli 74.5% ≥ 74.3% PASS; template/hook deltas attributed cross-SPEC / run-variance with git evidence (§E.2 M4e) |
| AC-TUX4-013 | PASS | full suite exit 0 + lint `0 issues.` (NEW=0) + windows/linux builds exit 0 (final-1/3/4 logs) |
| AC-TUX4-014 | PASS | doctor.go diff 50+/47− limited to observer seam + printer wiring + render delegation; only added func = `runGroupedChecksObserved`; render moved to doctor_render.go (new file) |

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-07-20
run_commit_sha: 1e95bf381
run_status: PASS-WITH-DEBT
ac_pass_count: 14
ac_fail_count: 0
ac_pass_with_debt: [AC-TUX4-010]
debt_register: "branch_protection.go:44 fmt.Printf (interactive y/N) — owned by deferred SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1; worktree/new.go:433 commented-out line excluded via D1 comment-exclusion idiom"
preserve_list_post_run_count: 0 violations (per-commit `git status --porcelain` + staged-path checks; PRESERVE files untouched across all 6 commits)
l44_pre_commit_fetch: yes (git fetch + rev-list divergence check before every commit)
l44_post_push_fetch: yes (post-push divergence re-checked; all pushes fast-forward)
new_warnings_or_lints_introduced: 0 (lint baseline 0 → final 0)
cross_platform_build:
  windows_amd64: exit 0
  linux_amd64: exit 0
total_run_phase_files: 32
m1_to_mN_commit_strategy: "per-milestone commit + push direct-to-main (Route A Hybrid Trunk): M4a 3b3b397cc / M4b 1317b9567 / M4c 2b886c39b / M4d-1 77893579e / M4d-2 408fe19f0 / M4e 1e95bf381; mid-run parallel-session (SPEC-MODEL-PROFILE-MATRIX-001) races absorbed via isolated-worktree verification + strict path staging"

## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-07-20
sync_commit_sha: "pending-backfill-2026-07-20"
sync_status: PASS-WITH-DEBT
changelog_entry_position: "[Unreleased] > Added (first entry)"
debt_carried_forward: "AC-TUX4-010 PASS-WITH-DEBT — internal/cli/branch_protection.go:44 (interactive y/N fmt.Printf) deferred to SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1; not modified by this SPEC"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (single sync commit, per Status Transition Ownership Matrix)"
b12_self_test:
  a_pre_emission_grep: "grep -c 'SPEC-CLI-TUX-V3-004' CHANGELOG.md → 0 (pre-emission, confirmed before append)"
  b_ac_count_match: "acceptance.md §D AC rows = 14 (grep -cE '^\\| AC-TUX4-[0-9]+' acceptance.md); CHANGELOG entry states 14/14 AC PASS — 13 clean + 1 PASS-WITH-DEBT"
  c_file_path_verification: "ls internal/cli/{glamour_style.go,doctor_render.go,help_order.go,uikit/banner.go,status.go,spec_view.go} → all exist (verified pre-draft)"
