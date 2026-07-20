# SPEC-CLI-TUX-V3-001 — Research (plan-phase 실측 기록)

> 본 문서의 §A-§E는 2026-07-10 plan-phase 세션에서 **실제 실행한 명령과 verbatim 출력**의 기록이다 (verification-claim-integrity §2 baseline 귀속). §F는 미검증 항목 — run-phase 착수 시 재검증 의무. 라인 번호 앵커는 drift 가능 — run-phase에서 content-token 기준 재실측.

## §A Charm v2 생태계 — 버전 실측 (검증됨)

명령: `go list -m -versions charm.land/<mod>/v2 | tr ' ' '\n' | tail -3` (2026-07-10 실행)

| 모듈 | 최신 확인 버전 | 비고 |
|---|---|---|
| charm.land/lipgloss/v2 | **v2.0.5** | ...v2.0.3, v2.0.4, v2.0.5 |
| charm.land/bubbletea/v2 | **v2.0.8** | ...v2.0.6, v2.0.7, v2.0.8 |
| charm.land/bubbles/v2 | **v2.1.1** | v2.0.0, v2.1.0, v2.1.1 |
| charm.land/fang/v2 | **v2.0.1** | v2.0.0, v2.0.1 |

보고서-출처 주장(본 세션 미실측, 신뢰 수준: 계획 보고서 인용): 2026-02-23 v2 동시 안정판 출시(6년 만의 breaking change, charm.land vanity 도메인 이동), Cursed Renderer, mode 2026 synchronized output, 선언적 View 구조체(tea.EnterAltScreen 커맨드류 제거), lipgloss v2 배경색 감지 정확도 개선·컬러 프로파일 자동 다운샘플링, fang: styled help/errors·자동 --version·manpages·completions·adaptive 테마·WithColorSchemeFunc. → **§F 재검증 대상**.

## §B 현행 의존성 (go.mod 실측, 검증됨)

직접: cobra v1.10.2 · bubbletea **v1.3.10** · huh **v1.0.0** · lipgloss **v1.1.0** · colorprofile v0.4.3 · x/powernap v0.1.6 · go-isatty v0.0.22 · termenv v0.16.0. 간접: bubbles v1.0.0 · catppuccin/go v0.3.0.

사용처 실측 (`grep -rln`, 비테스트):

- **bubbletea v1**: `internal/merge/confirm.go` 단 1곳 (879 LOC 파일; confirmModel Init/Update/View 모델부 43-224 구간).
- **huh v1**: `internal/cli/update.go`, `internal/cli/profile_setup.go`, `internal/cli/init.go`, `internal/cli/wizard/wizard.go` — 4곳.
- **lipgloss v1**: internal/tui 전체(내부 렌더) + statusline + uikit + wizard + reporter.go 등.

## §C 베이스라인 (ratchet SSOT, 검증됨)

### C.1 직접 fmt.Print* 카운트 — **baseline 46** (2026-07-10)

```bash
grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' | grep -v '_test.go' | wc -l
# → 46
```

파일별 분해 (동일 세션 실측):

| 파일 | 건수 |
|---|---|
| internal/cli/uikit/banner.go | 12 |
| internal/cli/state.go | 11 |
| internal/cli/migration.go | 8 |
| internal/cli/clean.go | 6 |
| internal/cli/worktree/tmux_integration.go | 5 |
| internal/cli/wizard/wizard.go | 2 |
| internal/cli/worktree/new.go | 1 |
| internal/cli/branch_protection.go | 1 |
| **합계** | **46** |

REQ-CTX-017 ratchet: 이 명령·이 수치가 기준. M1 종료 시 46 미만이어야 하며(대표 콜사이트 전환분), 신규 유입 금지.

### C.2 경고 stdout 유출 (검증됨)

`internal/cli/init.go`에서 `fmt.Fprintf(cmd.OutOrStdout(), "...Warning...")` **8곳** 실측: 라인 223, 227, 237(주변), 327, 334, 488, 500, 509, 514 (Warning 문자열 grep 기준; 481은 uikit.WarnStyle 렌더). internal/cli 전체 OutOrStdout+Warning 조합 11곳. → REQ-CTX-016 대상.

### C.3 hex 리터럴 잔존 (검증됨)

`internal/core/project/reporter.go:43-45` — `lipgloss.AdaptiveColor{Light: "#059669", ...}` 등 hex 6쌍(3 스타일). AC-CLI-TUI-013 계약(hex는 internal/tui에만)의 미커버 표면 — Printer 배선 시 제거 (REQ-CTX-008). 참고: internal/cli 쪽 uikit/wizard는 **코드 레벨** hex-free — 단 wizard/styles.go:17,20 **주석**에 금지-색상 문서화용 hex 문자열 잔존(`#DA7756 / #C45A3C`, `#7C3AED / #5B21B6`; wizard 무수정 유지, REQ-CTX-010). 따라서 AC-CTX-008 grep은 comment-line 제외 필터(`grep -vE ':[0-9]+:[[:space:]]*//'`)가 필수이고, 테스트 파일 필터는 파일명 위치 고정형 `_test\.go:`을 써야 함 — 17행은 내용에 "wizard_test.go" 문자열이 있어 naive `grep -v '_test.go'`가 콘텐츠 매칭으로 우연히 삼켰던 라인(2026-07-10 실측: 전체 hex 라인 5 = reporter 3 + wizard 주석 2, comment 제외 baseline 3).

### C.4 internal/tui 규모 (검증됨)

비테스트 15파일, **1,744 LOC** (계획 보고서의 "~1,824"는 근사치 — 실측이 우선). 파일: theme.go(185) · detect.go · status.go · progress_line.go(275) · box.go · table.go · pill.go · form.go · prompt.go · help.go · i18n.go · profile.go · catppuccin.go · term.go · doc.go. testdata 디렉터리 ≈108 엔트리(2026-07-10 `find` 실측: 총 108 · 파일 107; 종전 109는 `.`/`..` 포함 계열 카운트 오차 — run-phase 재실측 의무) + golden 디렉터리 — characterization 앵커.

### C.5 환경변수 시맨틱 (검증됨)

- `NO_COLOR`: detect.go — "any non-empty string is set" 표준; 우선순위 체인 `NO_COLOR > MOAI_THEME(light/dark) > DetectDark > dark-default` (detect.go:54 @MX:NOTE).
- `MOAI_THEME`: light/dark/auto/invalid — auto·invalid·unset은 DetectDark 위임 (detect.go:45-48).
- `MOAI_REDUCED_MOTION`: status.go:37-45,84 — 스피너→정적 점(●), 프로그레스→filled bar.

## §D 실행 경로 실측 (검증됨)

- `cmd/moai/main.go`: `cli.Execute()` 호출 + `ExitCoder` 인터페이스 언래핑 (worktree verify 0/1/2/3 커스텀 exit code).
- `internal/cli/root.go`: `trivialCommands` 맵(line 42) — --version/version/-v/help/--help/-h/completion; `Execute()`(line 61) = `initConsole()` → isTrivialCommand 분기(initLightDeps vs InitDependencies, REQ-PERF-003-A/B) → `rootCmd.Execute()`. rootCmd.Run = `uikit.PrintBanner` + Help.
- `internal/core/project/reporter.go`: ProgressReporter 인터페이스(StepStart/StepUpdate/StepComplete/StepError) + NoOpReporter + ConsoleReporter(fmt.Printf stdout 직행).
- quality.yaml: `development_mode: tdd` (검증됨).

## §E 결정적 발견 — 문자열 토큰 경계 (검증됨; design.md §B의 근거)

1. `internal/tui/theme.go`: Theme 구조체는 hex를 **plain string 필드**로 보유 — "Colour tokens ... as plain strings; lipgloss.Color interprets hex sub-strings automatically" (theme.go:18). Monochrome은 빈 문자열 = no-colour.
2. `internal/statusline/theme.go`: lipgloss **v1** import + `lipgloss.Color(tuipkg.CatppuccinMochaPrimary)` — tui 문자열 상수를 소비자가 스스로 감쌈 (thin wrapper, @MX:NOTE R-07). **statusline은 tui의 타입에 무의존.**
3. `internal/cli/wizard/wizard.go:340-384`: `newMoAIWizardTheme`이 lipgloss v1 스타일을 tui 토큰 색으로 조립 (huh v1 테마). wizard.go:342 — "All colour values are derived from internal/tui.LightTheme / DarkTheme tokens".
4. `internal/cli/uikit/styles.go:9`: AC-CLI-TUI-013 — hex 없음, AdaptiveColor를 tui 토큰으로 구성.

∴ tui의 lipgloss v2 포팅은 **내부 구현 교체**로 성립하며, v1 소비자 3계열은 소스 무수정으로 생존한다. shim 불필요. (design.md §B 결정 D-1의 실측 근거.)

## §F 미검증 — run-phase 재검증 의무 (Known Issue #9)

- **bubbletea/bubbles/fang/lipgloss v2의 API 시그니처 상세**: 본 세션에서는 모듈 버전 존재만 확인. v2 Style/Color 생성자, fang.Execute 옵션(테마 주입·에러 반환 모드), bubbles v2 spinner 계약은 착수 전 Context7(`resolve-library-id` → `query-docs`) 또는 `go doc`으로 확정할 것. **API를 추정으로 코딩 금지.**
- **fang의 에러 체인 보존 여부**: ExitCoder 언래핑 호환성 (design.md §E) — characterization test로 확정.
- **lipgloss v2 골든 차이 규모**: 포팅 전 예측 불가 — M1a-1 characterization run이 기준선.
- **huh v1 ↔ lipgloss v2 간접 충돌 부재**: 이론상 문제없음(별도 모듈)이나 `go mod tidy` 후 그래프 diff로 확증.
- 계획 보고서의 v2 기능 서술(Cursed Renderer 등, §A 후단)은 마케팅 계열 서술 — 구현 의존 결정에 사용 시 개별 재확인.

## §G 재검증 명령 모음 (run-phase pre-flight용)

```bash
go test ./internal/tui/... -count=1                     # characterization anchor
grep -rn 'fmt\.Printf\|fmt\.Println\|fmt\.Print(' internal/cli --include='*.go' | grep -v '_test.go' | wc -l   # ratchet=46 대조
grep -rn 'OutOrStdout' internal/cli/init.go | grep -c Warning                                                    # 경고 유출 8 대조
grep -rn '#[0-9A-Fa-f]\{6\}' internal/core/project --include='*.go' | grep -v _test | wc -l                      # hex 잔존 대조
go list -m -versions charm.land/lipgloss/v2 | tr ' ' '\n' | tail -1                                              # 버전 재확인
```
