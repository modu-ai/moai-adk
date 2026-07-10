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

### M1b — internal/cli/printer 신규 (TDD RED-GREEN-REFACTOR) + reporter/init 배선

#### RED evidence (spec-test 선작성, 실패 확인)

- `go test ./internal/cli/printer/ -count=1` → exit 1, 11개 테스트 함수 assertion FAIL (`m1b-red.log` verbatim — stub 구현에 대해 ChannelDiscipline/ModeResolution/PlainRendering/JSONRendering/EraseLine/ReducedMotion 전부 RED).
- 테스트 계약 확정 후 GREEN 구현 1-pass green; RED 이후 테스트 수정 2건은 계약 결함 정정(JSON 이벤트 수 8→9, 10 오산) + REFACTOR 단계 커버리지 보강 테스트 3함수 추가.

#### Printer 계약 (design.md §C 스케치 → RED에서 확정)

- `Printer`: Info/Warn/Error/Success/Data/Step + Spinner/Progress 핸들 + Mode(). 3모드 `ModeTTY`/`ModePlain`/`ModeJSON` 구성 시점 1회 해석: 명시 `WithMode` > stderr isatty; NO_COLOR(비어있지 않은 임의 값, `WithNoColor` 오버라이드 가능)는 TTY 해석을 Plain으로 강등 — REQ-CTX-014의 "어느 채널에도 ANSI 0" 충족(erase-line CSI 포함).
- 채널 규율: Data→stdout writer only, 나머지 전부 stderr writer only (모드 무관, REQ-CTX-012).
- 스타일 원천: `internal/tui` 문자열 토큰만(`WithTheme` 주입, 기본 `tui.ResolveOS()`); printer 소스 hex 리터럴 0. 마커 글리프는 `tui.StatusIcon` + tui.Spinner 프레임 전례(`⠋`/`●`), 진행바는 `tui.Progress` 위임(reduced-motion/monochrome 시맨틱 tui 소유).
- TTY 실단말 프로덕션 경로만 `colorprofile.NewWriter(stderr, os.Environ())` 랩(다운샘플링은 writer 소관 — v2 아키텍처); 테스트 버퍼는 raw로 계약 출력 관찰. colorprofile은 M1a에서 이미 direct dep — go.mod/go.sum 변경 0.
- `MOAI_REDUCED_MOTION`(tui 시맨틱 미러): spinner/progress만 정적 라인 강등, Step erase-line은 비대상(REQ-CTX-007 범위 준수).
- 핸들은 terminal-once no-op(panic 금지 — 출력 게이트웨이가 CLI를 죽이면 안 됨; tui.ProgressLine의 panic-on-misuse와 의도적으로 다른 결정, adapter 안전성 우선).

#### ConsoleReporter 재배선 (REQ-CTX-008/015)

- `internal/core/project/reporter.go`: `ConsoleReporter` 삭제 — hex AdaptiveColor 3건(구 43-45) + `fmt.Printf` stdout 직행 6건 제거. `ProgressReporter` 인터페이스 + `NoOpReporter`는 무변경(PhaseExecutor 호출부 0 수정).
- 어댑터는 cli 쪽 `internal/cli/reporter.go` `printerReporter`(design.md §C D-2 — core→cli 역참조 없음). 소비 2 콜사이트 전환: init.go:463대, update.go:943대.
- core 테스트 `TestConsoleReporter` 삭제(타입 소멸에 따른 SPEC-required 정리; 대체 커버리지 = `internal/cli/reporter_test.go` 2함수).

#### init.go 대표 콜사이트 마이그레이션 (REQ-CTX-016)

- Warning stdout 유출 9곳(구 223,227,237,327,334,488,500,509,514) → `p.Warn(...)` (stderr, "! Warning: " 접두). printer는 `printer.New(WithWriters(cmd.OutOrStdout(), cmd.ErrOrStderr()))` — cobra 스트림 기반이라 기존 테스트 캡처 유지.
- 채널 정정(동반): "Initializing MoAI project..." → `p.Info`(stderr); 성공 카드(uikit 유지) + 선행 공백 라인 → `cmd.ErrOrStderr()`; hook installer 2 콜사이트(518/521) writer 인자 → `cmd.ErrOrStderr()` (경고/상태 라인 방출 함수 — 유닛 테스트는 자체 버퍼 주입이라 무영향).
- ratchet 대표 마이그레이션: `clean.go` 직접 `fmt.Print*` 6건 → printer (Info/Warn) — runClean 시그니처에 printer 주입.
- 신규 테스트: `internal/cli/init_channel_test.go` `TestInitWarningChannelStderr` (3-레벨: 배선 positive + full-run 분리 버퍼 negative + 소스 정적 가드).

#### 기존 테스트 픽스처 변경 목록 (AC-CTX-026 사유 기록)

- **변경 0건.** 채널 이동으로 파손된 기존 테스트 파일 없음 — init 계열 기존 테스트는 `SetOut(buf)`+`SetErr(buf)` 동일 버퍼 캡처라 stdout→stderr 이동에 불변. 골든 파일 변경도 0 (M1b는 tui 골든 무접촉).
- 유일한 기존 테스트 수정 = `internal/core/project/coverage_extra_test.go` `TestConsoleReporter` 삭제 (사유: REQ-CTX-015가 타입 자체를 cli 어댑터로 대체; NoOp/SetReporter 테스트는 잔존 PASS).

#### M1b gate evidence (verbatim tails; logs `.moai/state/verify/SPEC-CLI-TUX-V3-001/`)

- 빌드: `go build ./...` + `GOOS=windows` + `GOOS=linux` → `darwin=0 windows=0 linux=0` (`m1b-build.log`)
- 커버리지(AC-CTX-024): `go test -cover ./internal/cli/printer/` → `coverage: 97.3% of statements` (`m1b-cover.log`; ≥90 게이트 통과. 초기 GREEN 85.9% → REFACTOR 테스트 보강 후 97.3%, `m1b-cover-initial.log`)
- AC-CTX-008 hex grep(acceptance 필터 그대로) → **0** (baseline 3 → 0)
- AC-CTX-015: `grep -c 'fmt\.Printf' internal/core/project/reporter.go` → **0** + `go test ./internal/core/project/ -run 'Reporter'` PASS (`m1b-ac-matrix.log`)
- AC-CTX-017 ratchet: `grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' | grep -v '_test.go' | wc -l` → **40** (baseline 46, −6; 잔여 40 = banner 12 [M4 무접촉] + state 11 + migration 8 + worktree 6 + wizard 2 [PRESERVE] + branch_protection 1)
- AC-CTX-011/012/013/014/016 `-run` 필터 테스트 전부 PASS (`m1b-ac-matrix.log` verbatim)
- 패키지 스위트: `go test ./internal/cli/... ./internal/core/... -count=1` → 13 pkg ok (`m1b-pkg-test.log`); `./internal/tui/... ./internal/merge/... ./internal/statusline/...` → 5 pkg ok (`m1b-tui-merge-test.log`) — statusline/wizard/uikit 소스 무수정 유지 (AC-CTX-010)
- lint: `golangci-lint run --timeout=5m` → exit 0, `0 issues.` (`m1b-lint.log`; NEW 0 — 중간 QF1011 1건은 테스트 자체 수정으로 해소)
- vet: `go vet ./...` → exit 0
- 전체 스위트: `go test ./... -count=1` → 93 pkg ok + 1 FAIL = `internal/hook` `TestHookWrapper_TempFileCleanup` (`m1b-full-test.log`) — **pre-existing flaky** (SPEC-HOOK-OFFICIAL-COMPLIANCE-001 progress.md:255에 문서화된 타이밍 민감 부채; 본 세션 isolated rerun `ok 4.694s`, internal/hook 무접촉 `git status` clean)

### M1c — Fang v2 root 통합 + confirm.go 결정 기록 + 회귀 매트릭스 (characterization-first)

#### Characterization 선행 (AC-CTX-020, RED→GREEN 불변)

- 방법: fang 배선 **전에** `internal/cli/fang_characterization_test.go` `TestFangExitCoderCharacterization` 작성 → root 실행 seam(`runFang`)을 raw cobra 상태로 baseline PASS 기록(`m1c-char-baseline.log`), fang 스왑 후 **동일 테스트 무수정 재통과**(`m1c-char-postfang.log`).
- 3-불변 고정: (1) ExitCoder 체인 — 합성 커맨드가 반환한 `*worktree.ExitCodeError{0/1/2/3}` + 로컬 ExitCoder + 평문 error를 `cmd/moai/main.go`의 `errors.As` 언래핑과 동일한 helper로 매핑(0/1/2/3, 평문→1, --help→0, unknown→1); (2) non-TTY+NO_COLOR --help 양 채널 ANSI 부재; (3) trivial fast-path(`isTrivialCommand`) 불변.
- **seam 격리 설계**: 테스트는 전역 `rootCmd`가 아닌 합성 root 트리만 사용 → fang이 `SetHelpFunc`/`SilenceErrors`/`Version`을 뮤테이트해도 ~20개 형제 cli 테스트 파일 오염 없음. 프로덕션 `runFang`도 snapshot+restore(defer)로 전역 rootCmd를 pristine 유지.

#### Fang API 결정 (run-phase 실측, `go doc` + 모듈 소스 fang.go:110-179 / theme.go)

| # | 이슈 | fang v2.0.1 실측 | 결정 |
|---|------|------------------|------|
| 1 | ExitCoder 보존 | `fang.Execute`는 `root.ExecuteContext` 원본 error를 `return err`(fang.go:176) | design.md §E 컨틴전시 **불필요** — 에러 스왈로우 없음. main.go의 `errors.As(ExitCoder)` 그대로 동작 (worktree verify 0/1/2/3 보존) |
| 2 | fang 무조건 error print | `DefaultErrorHandler`가 SilenceErrors 무시하고 항상 styled "ERROR" 박스 출력(fang.go:175) | `WithErrorHandler(moaiErrorHandler)` — ExitCoder 캐리어(worktree verify)면 print 억제(SilenceErrors 의도 보존), 그 외 진짜 에러는 DefaultErrorHandler 위임(REQ-CTX-018 styled errors) |
| 3 | help func override | fang이 `root.SetHelpFunc`(fang.go:140)로 tui `renderRootHelp` 대체 | 프로덕션은 fang help(REQ-CTX-018), 직접 `rootCmd.Execute()` 테스트는 snapshot+restore로 renderRootHelp 유지 → 양립 |
| 4 | --version 이중화 | fang이 `root.Version = buildVersion(opts)` 세팅(fang.go:138) | `WithVersion(version.GetVersion())` 단일 소스 핀. 기존 `SetVersionTemplate` 리터럴 출력 `moai-adk <ver>` 불변 검증 |
| 5 | man 커맨드 추가 | `opts.manpages` 기본 true → `man` 서브커맨드 add(fang.go:143) | `WithoutManpage()` — 현행 무-man 동작 보존 + 반복 in-process Execute() 중복 add 회피 |
| 6 | 테마 hex 유입 | `ColorScheme` 16필드 `color.Color` | `WithColorSchemeFunc(moaiColorScheme)` — 전 필드를 `tui.LightTheme()/DarkTheme()` 토큰 문자열 + `lipgloss.Color`로 파생, `lipgloss.LightDarkFunc`로 배경 감지 선택. **fang.go hex 리터럴 0**(REQ-CTX-022) |

#### 배선 (REQ-CTX-018/019/022)

- `internal/cli/root.go`: `Execute()`의 lazy-init 분기(REQ-CTX-019) **fang 바깥 유지** → `executeRoot(ctx, rootCmd)` → `runFang`. seam은 `cmd.ExecuteContext` 자리만 `fang.Execute(ctx, cmd, fangOptions()...)`로 대체. `initConsole()`는 Execute() 최상단(Windows VT, fang보다 먼저 — acceptance §C edge).
- `internal/cli/fang.go`(신규): `runFang`(snapshot+restore) + `fangOptions()`(WithVersion/WithoutManpage/WithColorSchemeFunc/WithErrorHandler) + `moaiColorScheme`(tui 토큰) + `moaiErrorHandler`(ExitCoder 억제) + `exitCoder` 인터페이스.
- `cmd/moai/main.go`: **무변경**. fang.Execute가 cli.Execute() 내부(root 실행 site)에 있고 원본 error를 반환하므로 main의 ExitCoder 번역기는 그대로 동작 (design.md §E "cmd/moai/main.go → internal/cli.Execute()" 배치 정합).
- 의존성: `go get charm.land/fang/v2@v2.0.1`(최신 실측). fang은 bubbletea/v2·bubbles/v2를 transitively 요구하지 **않음**(go get 산출 확인) → indirect `charmbracelet/ultraviolet` 만 `20251205161215`→`20260205113103` 승격 + charmtone/mango/roff 신규 indirect. `go mod tidy` 후 fang direct 승격.

#### confirm.go 결정 기록 (REQ-CTX-023 / D-3)

- `internal/merge/confirm.go:871` `tea.NewProgram(m)` 직전에 3태그 삽입: `@MX:DEBT`(bubbletea v1 잔류) + `@MX:CEILING`(huh v1이 v1을 그래프에 유지하는 한 유효) + `@MX:UPGRADE`(SPEC-CLI-TUX-V3-003 bubbles v2 list 승격). bubbletea v1 import(`tea "github.com/charmbracelet/bubbletea"`) 무변경. merge 스위트 green.

#### 기존 테스트 픽스처 변경 목록 (AC-CTX-026 사유 기록)

- **골든 변경 0건** — M1c는 tui 골든·testdata 무접촉(seam+fang은 cli 계층). `git diff 27fe9ea37..HEAD --stat -- internal/tui/testdata internal/tui/golden` → 0 (AC-CTX-009).
- **기존 테스트 수정 0건** — snapshot+restore 설계로 fang의 전역 rootCmd 뮤테이션이 help-characterization 테스트(TestCharacterize_Help_*, TestRootCmd_HelpOutput)를 오염하지 않아 형제 테스트 전부 무수정 green. 신규 테스트 파일 1개(fang_characterization_test.go)만 추가.

#### M1c gate evidence (verbatim tails; logs `.moai/state/verify/SPEC-CLI-TUX-V3-001/`)

- 빌드: `go build ./... && GOOS=windows && GOOS=linux` → `darwin=0 windows=0 linux=0`(`m1c-xbuild.log`, AC-CTX-003)
- lint: `golangci-lint run --timeout=5m` → exit 0, `0 issues.`(`m1c-lint.log`; NEW 0)
- vet: `go vet ./...` → exit 0(`m1c-vet.log`)
- 커버리지(AC-CTX-024): `go test -cover ./internal/cli/printer/` → `coverage: 97.3% of statements`(`m1c-printer-cover.log`; M1b 값 불변, printer 무접촉)
- 전체 스위트: `go test ./... -count=1` → **93 pkg ok** + 1 FAIL = `internal/hook` `TestHookWrapper_TempFileCleanup`(`m1c-full-test.log`) — **pre-existing flaky**(SPEC-HOOK-OFFICIAL-COMPLIANCE-001 progress.md:255 문서화; M1c 무접촉 `git status internal/hook` empty; isolated rerun `ok internal/hook 3.600s` `m1c-hook-isolated.log`). 패키지-레벨 FAIL은 internal/hook 1건뿐(`grep -P '^FAIL\t'`).
- ratchet(AC-CTX-017): `grep ... internal/cli | grep -v _test.go | wc -l` → **40**(baseline 46; M1c는 internal/cli에 fmt.Print* 미추가로 40 불변).
- 스모크(프로덕션 fang 경로, 컴파일 바이너리): `--help`→exit 0 / `--version`→exit 0 `moai-adk v3.0.0-rc10`(단일 소스) / unknown cmd→exit 1(fang styled ERROR 박스 stderr) / `worktree verify` 미스냅샷→exit 1 / `worktree snapshot`+`verify` clean tree→JSON `exit_code:0` + process exit 0(stdout JSON fang 무손상).

#### 최종 AC-CTX-001..026 매트릭스 (verbatim 명령 출력은 `.moai/state/verify/SPEC-CLI-TUX-V3-001/m1c-ac-*.log` 인용)

| AC | 상태 | 증거 (실측) |
|----|------|-------------|
| AC-CTX-001 | **PASS-WITH-DEBT** | `go list -m`: lipgloss/v2 v2.0.5 + fang/v2 v2.0.1 resolve(둘 다 사용). bubbletea/v2 + bubbles/v2 → `not a known dependency` — **미추가**(confirm.go=bubbletea v1 잔류 D-3, printer=tui native spinner M1b → 실사용처 없음; `go mod tidy`가 unused direct 제거하므로 blank-import 해킹 없이는 그래프 유입 불가). SPEC 원문 "4개 direct" 대비 편차 — orchestrator 위임 지시("do not add unused direct deps")에 따른 의도적 연기. bubbletea/v2·bubbles/v2 정식 도입은 M3(confirm.go bubbles v2 list 승격) 소관 |
| AC-CTX-002 | PASS | huh v1.0.0 + lipgloss v1.1.0 + bubbletea v1.3.10 공존 + `go build ./...` exit 0 |
| AC-CTX-003 | PASS | darwin/windows/linux build 전부 exit 0 (`m1c-xbuild.log`) |
| AC-CTX-004 | PASS | tui 비테스트 lipgloss v1 import grep → 0 |
| AC-CTX-005 | PASS | `go test ./internal/tui/ -run 'NoColor\|Monochrome\|Detect\|Resolve'` → ok |
| AC-CTX-006 | PASS | `-run 'Theme\|MoaiTheme\|Profile'` → ok |
| AC-CTX-007 | PASS | `-run 'ReducedMotion\|Spinner\|Progress'` → ok |
| AC-CTX-008 | PASS | reporter/cli code-level hex grep → 0 |
| AC-CTX-009 | PASS | golden diff `27fe9ea37..HEAD` → 0 + tui 스위트(3 pkg) ok. M1c 골든 변경 없음 |
| AC-CTX-010 | PASS | 보호 3경로 diff `27fe9ea37..HEAD` → 0 + statusline/wizard/uikit 스위트 ok |
| AC-CTX-011 | PASS | printer `-run 'Interface\|Contract'` → ok |
| AC-CTX-012 | PASS | printer `-run 'ChannelDiscipline'` → ok |
| AC-CTX-013 | PASS | printer `-run 'Mode'` → ok |
| AC-CTX-014 | PASS | printer `-run 'NoANSI\|PlainNoEscape'` → ok |
| AC-CTX-015 | PASS | `grep -c 'fmt\.Printf' reporter.go` → 0 + Reporter 테스트 ok |
| AC-CTX-016 | PASS | init `-run 'InitWarn.*Stderr\|WarningChannel'` → ok |
| AC-CTX-017 | PASS | internal/cli 직접 fmt.Print* → 40 (baseline 46 미만) |
| AC-CTX-018 | PASS | `grep 'fang\.Execute' internal/cli cmd/moai` → 매치(fang.go:43 호출 + fang.go:20 주석) ≥ 1 |
| AC-CTX-019 | PASS | `-run 'Trivial\|LazyInit\|LightDeps'` → ok (lazy-init 분기 fang 바깥 보존) |
| AC-CTX-020 | PASS | `grep 'func TestFangExitCoderCharacterization'` → fang_characterization_test.go:32 + `-run` PASS(internal/cli ok 2.035s). exit 0/1/2/3 보존 |
| AC-CTX-021 | PASS | `NO_COLOR=1 moai --help` ANSI grep → 0, `moai init --help` → 0 (파이프 non-TTY) |
| AC-CTX-022 | PASS | fang 테마 배선 파일 = `internal/cli/fang.go`: hex 0 + tui 토큰 참조 18. `moaiColorScheme`가 tui.LightTheme/DarkTheme 파생 |
| AC-CTX-023 | PASS | confirm.go @MX:DEBT(871)+CEILING(872)+UPGRADE(873) 전부 존재 + merge 스위트 ok + design.md §D 결정 기록 |
| AC-CTX-024 | PASS | `go test -cover ./internal/cli/printer/` → `coverage: 97.3% of statements` (≥ 90) |
| AC-CTX-025 | **PASS-WITH-DEBT** | `go test ./... -count=1` 93 pkg ok + lint 0 issues(NEW 0). 유일 FAIL = internal/hook TestHookWrapper_TempFileCleanup(pre-existing flaky, M1c 무접촉, isolated PASS). |
| AC-CTX-026 | PASS | 본 progress.md §E.2 M1c에 골든-변경-사유(0건) + ratchet 실측치(40) 기재 |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-10
run_commit_sha: 3be403dc4  # M1a b9df38f5b → M1b 5824c812c → M1c 3be403dc4 (main cherry-pick 완료; orchestrator 백필)
run_status: audit-ready
ac_pass_count: 24                  # 26 AC 중 24 full PASS + 2 PASS-WITH-DEBT
ac_fail_count: 0
ac_pass_with_debt: 2               # AC-CTX-001 (bubbletea/v2+bubbles/v2 미추가 연기), AC-CTX-025 (internal/hook pre-existing flaky)
preserve_list_post_run_count: 0    # 보호 경로(statusline/wizard/uikit/tui golden) diff 0
l44_pre_commit_fetch: n/a          # 워크트리 격리 실행 — orchestrator가 cherry-pick 전 pre-spawn sync check 수행
l44_post_push_fetch: n/a           # 워크트리는 push 안 함(orchestrator cherry-pick 위임)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin: pass
  windows: pass
  linux: pass
total_run_phase_files: 6           # M1c: go.mod, go.sum, root.go, confirm.go(+@MX), fang.go(new), fang_characterization_test.go(new)
m1_to_mN_commit_strategy: "M1a(b9df38f5b) + M1b(5824c812c) + M1c(1 commit) — 마일스톤별 분리; M1c는 characterization+fang+@MX+matrix 단일 커밋"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
