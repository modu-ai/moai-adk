# SPEC-CLI-TUX-V3-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- 2026-07-10: plan-phase artifact set authored (spec.md / plan.md / acceptance.md / design.md / research.md / progress.md — Tier L 6-artifact) by manager-spec, from CLI TUX modernization plan `.moai/reports/moai-cli-tux-modernization-plan-20260710.html` Milestone M1. Status: draft. 25 REQ (REQ-CTX-001..025) / 26 AC (AC-CTX-001..026).
- **Implementation Kickoff Approval evidence**: 사용자가 2026-07-10 명시적 `/goal` 지시("plan 생성 후 run > sync까지 모두 완료")로 run-phase 진입을 사전 승인함 — plan→run HUMAN GATE 승인 근거로 본 항목을 기록한다. (plan-audit PASS는 별도 선행 조건으로 유지.)
- Key design decisions: D-1 문자열 토큰 경계(huh v1 / lipgloss v2 interop, design.md §B), D-3 merge/confirm.go bubbletea v1 잔류 + @MX:DEBT 연기(design.md §D), D-4 fang.Execute 래핑 시 lazy-init 분기 보존(design.md §E).
- Ratchet baseline frozen: internal/cli 직접 fmt.Print* = 46 (research.md §C.1, 2026-07-10).

plan_complete_at: 2026-07-10
plan_status: audit-ready

## §E.2 Run-phase Evidence

### M1a — internal/tui Lip Gloss v2 port (characterization-first)

preflight_head: 27fe9ea37ee1ec3b6fe448bac9516cfa82fe9609

AC-CTX-009 / AC-CTX-010 range diffs anchor on this sha (`git diff 27fe9ea37..HEAD`).

#### v2 API spike (research.md §F obligation — actual API inspected via `go doc charm.land/lipgloss/v2` + module source, 2026-07-10)

| # | v1 API | v2 API | Port decision |
|---|--------|--------|---------------|
| 1 | import `github.com/charmbracelet/lipgloss` | import `charm.land/lipgloss/v2` (package name stays `lipgloss`) | mechanical import swap, 11 non-test files |
| 2 | `lipgloss.Color` is a string TYPE (`Color("#hex")` = conversion) | `Color(s string) color.Color` is a FUNCTION returning stdlib `color.Color` | call sites `lipgloss.Color(str)` compile unchanged; type annotations break — unexported `borderColor` return type changed `lipgloss.Color` → `color.Color` (box.go) |
| 3 | `Color("")` → empty color → no ANSI | `Color("")` → `NoColor{}` sentinel; v2 style.go:353 guard `fg != noColor` skips the attribute | MonochromeTheme empty-token semantics preserved verbatim — no change needed |
| 4 | `HasDarkBackground() bool` (implicit global renderer) | `HasDarkBackground(in, out term.File) bool` — explicit handles, OSC 11 terminal query, returns true (dark) on error | `detect.go OSEnv.DetectDark()` → `lipgloss.HasDarkBackground(os.Stdin, os.Stdout)`. Observable resolution order unchanged: NO_COLOR > MOAI_THEME > DetectDark > dark-default (query fires only in auto/unset mode); error→dark matches the existing safe-dark default |
| 5 | Global default renderer: `Style.Render` auto-downsampled/stripped ANSI per os.Stdout TTY+env detection (why v1 goldens are ESC-free under `go test`) | Global renderer REMOVED — `Style.Render` always emits full-fidelity ANSI; downsampling is the writer's job (`lipgloss.Writer` / `colorprofile.Writer`) | Added unexported `downsample()` + lazy `outputProfile()` (sync.Once, `colorprofile.Detect(os.Stdout, os.Environ())`) in profile.go; every public render function's return is re-encoded. NoTTY→`ansi.Strip` (verified in colorprofile v0.4.3 writer.go:42-44) reproduces v1 test-time stripping; TrueColor passthrough; ANSI/ANSI256 downsampled. No lipgloss/colorprofile type leaks (string-in/string-out, D-1 boundary intact) |
| 6 | `Style.Width(n)` = block width INCLUDING padding, EXCLUDING border columns | `Style.Width(n)` = TOTAL block width INCLUDING borders — empirically verified: `Width(34)` renders 34 total cols (v1 rendered 36) | box.go `renderBox` + term.go `Term`: `innerW := max(W-2-4, 1) + 2` (v1 inner width + 2 border cols). Goldens byte-identical — no golden regeneration |
| 7 | Borders: `NormalBorder/RoundedBorder/ThickBorder` | unchanged in v2 | no change |
| 8 | deps | new indirects: `charmbracelet/ultraviolet`, `x/termios`, `x/windows` | `go mod tidy` clean; no existing pin bumped (huh v1.0.0 / lipgloss v1.1.0 / bubbletea v1.3.10 / bubbles v1.0.0 / catppuccin v0.3.0 unchanged — REQ-CTX-002 coexistence verified) |

bubbletea/v2, bubbles/v2, fang/v2: NOT added in M1a (M1c scope; no transitive need arose).

#### Characterization anchor (AC-CTX-009 golden gate)

- Baseline (pre-port @ 27fe9ea37): `go test ./internal/tui/... -count=1` → 3 packages `ok` (`.moai/state/verify/SPEC-CLI-TUX-V3-001/m1a-baseline-tui-test.log`)
- Post-port: same command → 3 packages `ok` (`m1a-post-port-tui-test.log`)
- **Golden updates: NONE** — `git status --porcelain internal/tui/testdata internal/tui/golden` → 0 entries. Test files unmodified. The width-formula compensation (spike delta #6) preserved v1 geometry, so no golden changed; AC-CTX-026's golden-rationale requirement is satisfied by this "no changes" record.

#### Background-detection decision (spawn-prompt documentation obligation)

lipgloss v2 requires stdin/stdout handles for the background query. Decision: keep the env-var-first chain untouched (`Resolve`: NO_COLOR > MOAI_THEME(light/dark) > DetectDark > dark-default) and pass `os.Stdin, os.Stdout` at the single production call site (`OSEnv.DetectDark`). Safe fallback = v2's documented error→true(dark), identical to the v1 safe-dark default (REQ-CLI-TUI-010). Tests inject `staticEnv`/`profileEnv` doubles, so no terminal query occurs under `go test`.

#### M1a gate evidence (verbatim tails; logs under `.moai/state/verify/SPEC-CLI-TUX-V3-001/`)

- `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0; `GOOS=linux GOARCH=amd64 go build ./...` → exit 0 (`m1a-build.log`)
- `go test ./internal/tui/... ./internal/statusline/... ./internal/cli/... ./internal/merge/... -count=1` → exit 0, 14 packages `ok` (`m1a-consumer-test.log`) — statusline/wizard/uikit source-unmodified (AC-CTX-010)
- `go vet ./...` → exit 0 (`m1a-vet.log`)
- `golangci-lint run --timeout=5m` → exit 0, `0 issues.` (`m1a-lint.log`; pre-port baseline also `0 issues.` → NEW findings 0)
- AC-CTX-004 grep: `grep -rn '"github.com/charmbracelet/lipgloss"' internal/tui --include='*.go' | grep -v _test.go` → 0 matches
- `go list -m charm.land/lipgloss/v2` → `charm.land/lipgloss/v2 v2.0.5`; `go list -m huh/lipgloss/bubbletea (v1)` → `v1.0.0 / v1.1.0 / v1.3.10` (AC-CTX-002)

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
