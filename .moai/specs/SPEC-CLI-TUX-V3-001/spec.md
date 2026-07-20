---
id: SPEC-CLI-TUX-V3-001
title: "Charm v2 Migration + Unified CLI Printer (MoAI Terminal Design Language v3 — M1)"
version: "0.1.0"
status: completed
created: 2026-07-10
updated: 2026-07-10
author: manager-spec
priority: P0
phase: "v3.0.0 target"
module: "internal/tui + internal/cli/printer + internal/cli + internal/merge"
lifecycle: spec-anchored
tags: "cli, tux, charm-v2, lipgloss, bubbletea, bubbles, fang, printer, design-language, tier-l"
era: V3R6
tier: L
---

# SPEC-CLI-TUX-V3-001 — Charm v2 마이그레이션 + 통합 CLI Printer (MoAI Terminal Design Language v3, M1)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-10 | manager-spec | Initial draft — CLI TUX 현대화 계획 보고서(`.moai/reports/moai-cli-tux-modernization-plan-20260710.html`) M1 마일스톤으로부터 작성 |

## §A Context

moai CLI의 시각 계층은 전부 Charm v1 세대(lipgloss v1.1.0 · bubbletea v1.3.10 · huh v1.0.0 · bubbles v1.0.0 indirect)에 묶여 있고, 출력 채널은 `fmt.Printf` 직접 호출 / `cmd.OutOrStdout()` / `ConsoleReporter` 3계열로 분산되어 있다. 특히 `internal/cli/init.go`는 경고 8곳을 **stdout**으로 유출하며(자체 컨벤션 "stdout=데이터, stderr=상태" 위반, `internal/cli/CLAUDE.md` Output streams 규약), 비동기 스피너가 없어 긴 단계가 무반응처럼 보인다.

2026-02-23 Charm이 Bubble Tea·Lip Gloss·Bubbles v2 안정판을 동시 출시했고(import 경로가 `charm.land/*/v2` vanity 도메인으로 이동), Fang v2가 cobra help/에러/completion 표면을 한 줄로 현대화한다. 본 SPEC은 CLI TUX 현대화 4-마일스톤 중 **M1(기반 교체)** 만을 다룬다: (1) Charm v2 의존성 도입, (2) `internal/tui` 디자인 시스템 커널(실측 1,744 LOC, 15개 비테스트 파일)의 lipgloss v2 behavior-preserving 포팅, (3) 신규 `internal/cli/printer` 통합 Printer 추상화, (4) `fang.Execute` root 진입 교체, (5) `internal/merge/confirm.go` bubbletea v1 프로그램 처리 결정.

핵심 설계 제약: **huh v1.0.0은 lipgloss v1에 의존**하므로 v1/v2 혼재가 불가피하다. 다행히 `internal/tui`의 공개 계약은 이미 **plain hex 문자열 토큰**(Theme 구조체 string 필드 + `Catppuccin*` 상수)이어서 lipgloss 메이저 버전에 중립적이다 — statusline·wizard·uikit이 전부 문자열만 소비함을 실측 확인했다(research.md §E). 이 문자열 토큰 경계를 v1/v2 interop 경계로 승격하는 것이 본 SPEC의 중심 설계 결정이다(design.md §B).

찾은 사실의 전체 실측 기록과 재검증 명령은 research.md가 SSOT다. 모든 앵커(라인 번호·카운트)는 run-phase 착수 시 재실측한다.

## §B Requirements (GEARS)

### B.1 의존성 마이그레이션 (Deps)

- **REQ-CTX-001**: The go.mod dependency set shall include `charm.land/lipgloss/v2` (>= v2.0.5), `charm.land/bubbletea/v2` (>= v2.0.8), `charm.land/bubbles/v2` (>= v2.1.1), and `charm.land/fang/v2` (>= v2.0.1) as direct dependencies.
- **REQ-CTX-002**: While `huh` v1.0.0 remains a direct dependency, the module graph shall retain `github.com/charmbracelet/lipgloss` v1 (and `github.com/charmbracelet/bubbletea` v1 per the REQ-CTX-023 deferral) as coexisting major versions without build conflict; full v1 removal is a stretch goal, not a requirement of this SPEC.
- **REQ-CTX-003**: The repository shall build cleanly on darwin, linux, and windows targets after the migration (`go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` + `GOOS=linux GOARCH=amd64 go build ./...`).

### B.2 internal/tui — Lip Gloss v2 포팅 (behavior-preserving)

- **REQ-CTX-004**: The `internal/tui` package shall render through `charm.land/lipgloss/v2` internally while preserving its public string-token contract — the `Theme` struct hex-string fields, the `Catppuccin*` string constants, and all exported function signatures consumed by downstream packages.
- **REQ-CTX-005**: While `NO_COLOR` is set to any non-empty value, the tui theme resolution shall return the monochrome theme, preserving the existing priority chain NO_COLOR > MOAI_THEME(light/dark) > DetectDark > dark-default (`internal/tui/detect.go`).
- **REQ-CTX-006**: When `MOAI_THEME` carries `light`, `dark`, `auto`, or an invalid value, the theme resolution shall behave identically to the pre-migration implementation (light → LightTheme, dark → DarkTheme, auto/invalid/unset → DetectDark delegation).
- **REQ-CTX-007**: While `MOAI_REDUCED_MOTION` is set to a non-empty value, `Spinner` and `Progress` renderers shall produce the existing static variants (static dot / filled bar) instead of animated frames.
- **REQ-CTX-008**: The migrated surfaces shall not introduce raw hex color literals outside `internal/tui` (AC-CLI-TUI-013 계약 승계), and the `internal/core/project/reporter.go` hex `AdaptiveColor` literals shall be removed in favor of tui tokens during the Printer rewiring.
- **REQ-CTX-009**: The existing `internal/tui` test suite shall pass unmodified after the port, OR where rendered-output goldens legitimately change under lipgloss v2, each golden update shall be individually documented in progress.md §E.2 with a before/after rationale.
- **REQ-CTX-010**: Downstream lipgloss-v1 consumers of tui string tokens — `internal/statusline/theme.go` (thin wrapper), `internal/cli/wizard` (huh v1 theme), `internal/cli/uikit` — shall continue to compile and pass their existing tests without modification.

### B.3 internal/cli/printer — 통합 Printer 추상화 (신규, TDD)

- **REQ-CTX-011**: The new `internal/cli/printer` package shall provide a `Printer` interface exposing `Info`, `Warn`, `Error`, `Success`, `Data`, and `Step` methods, plus spinner/progress handle constructors for long-running step feedback.
- **REQ-CTX-012**: The Printer shall route `Data` output to stdout and `Info`/`Warn`/`Error`/`Success`/`Step` output to stderr (HARD channel discipline per `internal/cli/CLAUDE.md` Output streams).
- **REQ-CTX-013**: The Printer shall support three render modes — `tty-rich` (TTY + color), `plain` (non-TTY/CI), and `json` (future `--plain`/`--json` global flag consumption) — selected at construction time.
- **REQ-CTX-014**: While the process runs without a TTY or with `NO_COLOR` set, the Printer shall emit no ANSI escape sequences on either channel.
- **REQ-CTX-015**: The `ConsoleReporter` (`internal/core/project/reporter.go`) shall route its `StepStart`/`StepUpdate`/`StepComplete`/`StepError` output through the Printer abstraction (adapter or direct rewiring), eliminating its direct `fmt.Printf`-to-stdout calls.
- **REQ-CTX-016**: When the migrated `internal/cli/init.go` warning paths emit a warning, the warning text shall be written to stderr, not stdout (representative call-site migration; the full 34k-LOC sweep is out of scope per §C).
- **REQ-CTX-017**: The count of direct `fmt.Printf`/`fmt.Println`/`fmt.Print(` calls in non-test `internal/cli` sources shall not exceed the recorded baseline of 46 (ratchet; baseline command and per-file breakdown recorded in research.md §C), and the post-M1 count shall be strictly lower than the baseline as representative sites migrate to the Printer.

### B.4 Fang 통합 (root 실행 경로)

- **REQ-CTX-018**: The root execution path (`cmd/moai` + `internal/cli/root.go` `Execute()`) shall route through `fang.Execute(ctx, rootCmd)` (charm.land/fang/v2), providing styled help, styled errors, `--version`, and completions.
- **REQ-CTX-019**: While a trivial fast-path command runs (`version`/`help`/`completion` per the `trivialCommands` map, root.go), the lazy dependency-initialization semantics (REQ-PERF-003-A: `initLightDeps` instead of `InitDependencies`) shall be preserved.
- **REQ-CTX-020**: The fang integration shall preserve existing process exit-code behavior, including the `ExitCoder` custom-code chain (`moai worktree verify` 0/1/2/3) surfaced through `cmd/moai/main.go`.
- **REQ-CTX-021**: When `NO_COLOR=1` is set and output is piped (non-TTY), `moai --help` and `moai init --help` shall render without any ANSI escape sequences.
- **REQ-CTX-022**: Where the fang theme is configured, its colors shall be constructed from `internal/tui` tokens (no new hex literals outside internal/tui).

### B.5 internal/merge/confirm.go 처리 결정

- **REQ-CTX-023**: The `internal/merge/confirm.go` Bubble Tea program shall remain on bubbletea v1 in this SPEC, with the deferral decision recorded in design.md §D, an `@MX:DEBT` annotation (with `@MX:CEILING` and `@MX:UPGRADE` sub-lines pointing at SPEC-CLI-TUX-V3-003's bubbles-v2 list promotion) placed at the `tea.NewProgram` call site, and its existing test suite passing unchanged. (근거: M3/SPEC-003이 동일 UI를 bubbles v2 list로 전면 승격 예정 — M1 단순 포팅은 폐기 작업이 됨; design.md §D 참조.)

### B.6 품질 게이트

- **REQ-CTX-024**: The new `internal/cli/printer` package shall achieve >= 90% statement coverage (critical-package threshold per CLAUDE.local.md §6).
- **REQ-CTX-025**: The full repository test suite (`go test ./... -count=1`) and `golangci-lint run` shall pass with no NEW findings relative to the pre-migration baseline.

## §C Out of Scope (Non-Goals)

본 SPEC은 CLI TUX 현대화 계획의 M1만 다룬다. 아래 항목은 명시적으로 out of scope이며 후속 SPEC(002-004)의 소관이다.

### Out of Scope — init 위저드 재설계 + 셀프업데이트 순서 교정 (M2 / SPEC-CLI-TUX-V3-002)
- 셀프업데이트 체크를 위저드 이후 비동기로 이동(I-1), huh 단일 멀티필드 폼 통합(I-2), 라이브 진행률(I-3), 경고 수집기(I-4), 재실행 안내(I-5)는 전부 M2 소관.
- M1의 init.go 개입은 "대표 콜사이트의 Printer 전환 + 경고 stderr 라우팅"에 한정된다.

### Out of Scope — update.go 분해 + 변경 프리뷰 TUI (M3 / SPEC-CLI-TUX-V3-003)
- update.go 서브패키지 분해(U-1), Bubble Tea 변경 프리뷰 테이블(U-2), outcome 카드 통일(U-3), namespace 보호 가시화(U-4)는 M3 소관.
- `internal/merge/confirm.go`의 bubbles v2 list 승격도 M3 소관 — 본 SPEC은 v1 잔류 결정 + @MX:DEBT 기록만 수행한다(REQ-CTX-023).

### Out of Scope — doctor/status/spec 폴리시 + glamour (M4 / SPEC-CLI-TUX-V3-004)
- doctor 라이브 대시보드, glamour 마크다운 렌더, 배너 경량화, help 그룹 재정리는 M4 소관.

### Out of Scope — huh v2 업그레이드 (위저드는 M1에서 huh v1 유지)
- init 위저드·profile confirm·update 폼은 M1 동안 huh v1.0.0에 잔류한다. huh v2 승격(스크롤 버그 해소 스파이크 포함)은 M2 선행 스파이크 소관.
- 단, tui가 제공하는 huh v1 테마 경로(`internal/cli/wizard/wizard.go` `newMoAIWizardTheme`)는 M1 이후에도 동작해야 한다 — 이것이 v1/v2 혼재 interop 리스크이며 design.md §B가 경계를 정의한다.

### Out of Scope — statusline 시각 언어
- `internal/statusline`은 본 SPEC 무접촉(병렬 세션이 renderer.go 작업 중 — plan.md §D PRESERVE). tui와의 결합은 문자열 상수 소비뿐이므로 컴파일·테스트 무변경이 AC로 검증된다(REQ-CTX-010).

### Out of Scope — internal/cli 전체 콜사이트 일괄 전환
- 34,284 LOC 전체의 fmt.Printf 일괄 전환은 명시적으로 제외. 본 SPEC은 대표 집합(init.go 성공/경고 경로 + ConsoleReporter)만 전환하고, grep 기반 ratchet(REQ-CTX-017)으로 신규 유입만 차단한다. 잔여 콜사이트(state.go 11곳, migration.go 8곳, clean.go 6곳 등)는 후속 SPEC들이 표면별로 흡수한다.

## §D Non-Functional Constraints

- **Behavior preservation 기본**: tui 포팅(B.2)은 DDD식 characterization 선행 — 렌더 출력이 바뀌는 지점은 golden 갱신 사유를 개별 문서화(REQ-CTX-009).
- **Windows/CI 안전**: 모든 AC 검증은 non-TTY 폴백과 `GOOS=windows` 빌드를 포함(REQ-CTX-003, REQ-CTX-014, REQ-CTX-021).
- **성능**: root 실행 trivial fast-path(REQ-PERF-003-A)의 lazy-init가 fang 래핑 후에도 유지(REQ-CTX-019) — `moai --version` 기동에 무거운 의존성 그래프 초기화가 재유입되면 회귀.
- **커버리지**: 신규 printer 패키지 ≥ 90%(REQ-CTX-024); 기존 cli/template/hook 커버리지 목표(≥90%) 비회귀.

## §E Dependencies & Interop

- **depends_on**: 없음 — 본 SPEC이 CLI TUX v3 시리즈의 진입점이다. SPEC-CLI-TUX-V3-002/003/004가 본 SPEC에 의존한다(계획 보고서 §8).
- **v1/v2 혼재 경계**: `internal/tui`의 문자열 토큰 계약이 유일한 cross-major 경계다. lipgloss v1 잔류 소비자: huh v1(wizard/profile/update 폼), statusline, uikit, merge/confirm.go(bubbletea v1). 상세: design.md §B.
- **병렬 작업 격리**: working tree의 `internal/statusline/*` 및 `CLAUDE.local.md` 수정분은 타 세션 소유 — 본 SPEC의 커밋 범위에서 제외(plan.md §D).
